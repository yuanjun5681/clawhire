package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/milestone"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/progress"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

type commandFunc func(context.Context, *clawsynapse.Envelope) error

type CommandDispatcher struct {
	tasks       task.Repository
	bids        bid.Repository
	contracts   contract.Repository
	progress    progress.Repository
	milestones  milestone.Repository
	submissions submission.Repository
	reviews     review.Repository
	settlements settlement.Repository
	accounts    account.Repository
	domainEvts  event.DomainEventRepository
	sm          task.StateMachine
	now         Now
	handlers    map[string]commandFunc
}

type CommandDispatcherOptions struct {
	Tasks       task.Repository
	Bids        bid.Repository
	Contracts   contract.Repository
	Progress    progress.Repository
	Milestones  milestone.Repository
	Submissions submission.Repository
	Reviews     review.Repository
	Settlements settlement.Repository
	Accounts    account.Repository
	DomainEvts  event.DomainEventRepository
	StateMach   task.StateMachine
	Now         Now
}

func NewCommandDispatcher(opt CommandDispatcherOptions) *CommandDispatcher {
	now := opt.Now
	if now == nil {
		now = time.Now
	}
	sm := opt.StateMach
	if sm == nil {
		sm = task.NewStateMachine()
	}

	d := &CommandDispatcher{
		tasks:       opt.Tasks,
		bids:        opt.Bids,
		contracts:   opt.Contracts,
		progress:    opt.Progress,
		milestones:  opt.Milestones,
		submissions: opt.Submissions,
		reviews:     opt.Reviews,
		settlements: opt.Settlements,
		accounts:    opt.Accounts,
		domainEvts:  opt.DomainEvts,
		sm:          sm,
		now:         now,
	}
	d.handlers = map[string]commandFunc{
		clawhire.TypeTaskPosted:         d.handleTaskPosted,
		clawhire.TypeBidPlaced:          d.handleBidPlaced,
		clawhire.TypeTaskAwarded:        d.handleTaskAwarded,
		clawhire.TypeTaskStarted:        d.handleTaskStarted,
		clawhire.TypeProgressReported:   d.handleProgressReported,
		clawhire.TypeMilestoneCompleted: d.handleMilestoneCompleted,
		clawhire.TypeSubmissionCreated:  d.handleSubmissionCreated,
		clawhire.TypeSubmissionAccepted: d.handleSubmissionAccepted,
		clawhire.TypeSubmissionRejected: d.handleSubmissionRejected,
		clawhire.TypeSettlementRecorded: d.handleSettlementRecorded,
		clawhire.TypeTaskCancelled:      d.handleTaskCancelled,
		clawhire.TypeTaskDisputed:       d.handleTaskDisputed,
	}
	return d
}

func (d *CommandDispatcher) Dispatch(ctx context.Context, env *clawsynapse.Envelope) (event.ProcessStatus, error) {
	if env == nil {
		return event.ProcessStatusFailed, apierr.New(apierr.CodeInvalidRequest, "empty envelope")
	}
	handler, ok := d.handlers[env.Type]
	if !ok {
		return event.ProcessStatusFailed, apierr.New(apierr.CodeUnsupportedMessageType, "unsupported clawhire message type")
	}
	if err := handler(ctx, env); err != nil {
		return event.ProcessStatusFailed, err
	}
	return event.ProcessStatusSucceeded, nil
}

func (d *CommandDispatcher) handleTaskPosted(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.PostTaskPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validatePostTask(payload); err != nil {
		return err
	}
	if _, err := d.tasks.FindByID(ctx, payload.TaskID); err == nil {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId already exists")
	} else if err != task.ErrTaskNotFound {
		return apierr.Wrap(apierr.CodeInternalError, "find task", err)
	}
	now := d.now().UTC()
	reviewer := payload.Reviewer
	if reviewer == nil {
		cp := payload.Requester
		reviewer = &cp
	}
	item := &task.Task{
		TaskID:            payload.TaskID,
		Title:             strings.TrimSpace(payload.Title),
		Description:       strings.TrimSpace(payload.Description),
		Category:          strings.TrimSpace(payload.Category),
		Status:            task.InitialStatusForReward(payload.Reward.Mode),
		Requester:         payload.Requester,
		Reviewer:          reviewer,
		Reward:            task.Reward{Mode: task.RewardMode(strings.TrimSpace(payload.Reward.Mode)), Amount: payload.Reward.Amount, Currency: strings.TrimSpace(payload.Reward.Currency)},
		AcceptanceSpec:    normalizeAcceptanceSpec(payload.AcceptanceSpec),
		SettlementTerms:   normalizeSettlementTerms(payload.SettlementTerms),
		Deadline:          payload.Deadline,
		LastActivityAt:    &now,
		CreatedAt:         now,
		UpdatedAt:         now,
		AssignedExecutor:  nil,
		CurrentContractID: "",
	}
	if err := d.tasks.Insert(ctx, item); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert task", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleBidPlaced(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.PlaceBidPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validatePlaceBid(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	if err := d.sm.CanTransit(t.Status, task.ActionPlaceBid); err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "place bid not allowed", err)
	}
	item := &bid.Bid{
		BidID:     payload.BidID,
		TaskID:    payload.TaskID,
		Executor:  payload.Executor,
		Price:     payload.Price,
		Currency:  strings.TrimSpace(payload.Currency),
		Proposal:  strings.TrimSpace(payload.Proposal),
		Status:    bid.StatusActive,
		CreatedAt: d.now().UTC(),
	}
	if err := d.bids.Insert(ctx, item); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert bid", err)
	}
	if err := d.tasks.TouchActivity(ctx, payload.TaskID, d.now().UTC()); err != nil {
		return toAPIError("touch task", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleTaskAwarded(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.AwardTaskPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateAwardTask(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionAwardTask)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "award task not allowed", err)
	}
	if current, err := d.contracts.FindActiveByTask(ctx, payload.TaskID); err == nil {
		return apierr.New(apierr.CodeInvalidState, fmt.Sprintf("task already has active contract %s", current.ContractID))
	} else if err != contract.ErrContractNotFound {
		return toAPIError("find active contract", err)
	}
	now := d.now().UTC()
	item := &contract.Contract{
		ContractID:   payload.ContractID,
		TaskID:       payload.TaskID,
		Requester:    t.Requester,
		Executor:     payload.Executor,
		AgreedReward: shared.Money{Amount: payload.AgreedReward.Amount, Currency: strings.TrimSpace(payload.AgreedReward.Currency)},
		Status:       contract.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := d.contracts.Insert(ctx, item); err != nil {
		if err == contract.ErrActiveContractExists {
			return apierr.New(apierr.CodeInvalidState, "task already has active contract")
		}
		return apierr.Wrap(apierr.CodeInternalError, "insert contract", err)
	}
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, now); err != nil {
		return toAPIError("update task status", err)
	}
	if err := d.tasks.UpdateAssignment(ctx, payload.TaskID, payload.Executor, payload.ContractID, now); err != nil {
		return toAPIError("update assignment", err)
	}
	if bids, _, err := d.bids.ListByTask(ctx, payload.TaskID, 1, 200); err == nil {
		awardedBidID := ""
		for _, item := range bids {
			if item.Executor.ID == payload.Executor.ID && item.Status == bid.StatusActive {
				awardedBidID = item.BidID
				break
			}
		}
		if awardedBidID != "" {
			if err := d.bids.MarkAwarded(ctx, awardedBidID); err != nil {
				return toAPIError("mark bid awarded", err)
			}
		}
		if err := d.bids.InvalidateOthers(ctx, payload.TaskID, awardedBidID); err != nil {
			return apierr.Wrap(apierr.CodeInternalError, "invalidate other bids", err)
		}
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleTaskStarted(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.StartTaskPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateTaskStart(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionStartTask)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "start task not allowed", err)
	}
	at := chooseTime(payload.StartedAt, d.now)
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	contractID := strings.TrimSpace(payload.ContractID)
	if contractID == "" {
		contractID = t.CurrentContractID
	}
	if contractID == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "missing contractId for task.start")
	}
	if err := d.contracts.MarkStarted(ctx, contractID, at); err != nil {
		return toAPIError("mark contract started", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleProgressReported(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.ReportProgressPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateProgress(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	if err := d.sm.CanTransit(t.Status, task.ActionReportProgress); err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "report progress not allowed", err)
	}
	reportedAt := chooseTime(payload.ReportedAt, d.now)
	item := &progress.Report{
		ProgressID: payload.ProgressID,
		TaskID:     payload.TaskID,
		ContractID: firstNonEmpty(payload.ContractID, t.CurrentContractID),
		Executor:   payload.Executor,
		Stage:      strings.TrimSpace(payload.Stage),
		Percent:    payload.Percent,
		Summary:    strings.TrimSpace(payload.Summary),
		Artifacts:  payload.Artifacts,
		ReportedAt: reportedAt,
	}
	if err := d.progress.Insert(ctx, item); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert progress", err)
	}
	if err := d.tasks.TouchActivity(ctx, payload.TaskID, reportedAt); err != nil {
		return toAPIError("touch task", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleMilestoneCompleted(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.CompleteMilestonePayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateMilestone(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	if err := d.sm.CanTransit(t.Status, task.ActionCompleteMilestone); err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "complete milestone not allowed", err)
	}
	reportedAt := chooseTime(payload.ReportedAt, d.now)
	item := &milestone.Milestone{
		MilestoneID: payload.MilestoneID,
		TaskID:      payload.TaskID,
		ContractID:  firstNonEmpty(payload.ContractID, t.CurrentContractID),
		Title:       strings.TrimSpace(payload.Title),
		Summary:     strings.TrimSpace(payload.Summary),
		Status:      milestone.StatusDeclared,
		Claim:       normalizeClaim(payload.Claim),
		Artifacts:   payload.Artifacts,
		ReportedAt:  &reportedAt,
	}
	if err := d.milestones.Upsert(ctx, item); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "upsert milestone", err)
	}
	if err := d.tasks.TouchActivity(ctx, payload.TaskID, reportedAt); err != nil {
		return toAPIError("touch task", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleSubmissionCreated(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.CreateSubmissionPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateSubmission(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionCreateSubmission)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "create submission not allowed", err)
	}
	submittedAt := d.now().UTC()
	item := &submission.Submission{
		SubmissionID: payload.SubmissionID,
		TaskID:       payload.TaskID,
		ContractID:   firstNonEmpty(payload.ContractID, t.CurrentContractID),
		Executor:     payload.Executor,
		Summary:      strings.TrimSpace(payload.Summary),
		Artifacts:    payload.Artifacts,
		Evidence:     normalizeEvidence(payload.Evidence),
		Status:       submission.StatusSubmitted,
		SubmittedAt:  submittedAt,
	}
	if err := d.submissions.Insert(ctx, item); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert submission", err)
	}
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, submittedAt); err != nil {
		return toAPIError("update task status", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleSubmissionAccepted(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.AcceptSubmissionPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateAcceptSubmission(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionAcceptSubmission)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "accept submission not allowed", err)
	}
	subItem, err := d.submissions.FindByID(ctx, payload.SubmissionID)
	if err != nil {
		return toAPIError("find submission", err)
	}
	if subItem.TaskID != payload.TaskID {
		return apierr.New(apierr.CodeInvalidMessagePayload, "submission does not belong to task")
	}
	at := chooseTime(payload.AcceptedAt, d.now)
	if err := d.submissions.UpdateStatus(ctx, payload.SubmissionID, submission.StatusAccepted); err != nil {
		return toAPIError("update submission", err)
	}
	if err := d.reviews.Insert(ctx, &review.Review{
		ReviewID:     "review:" + DeriveEventKey(env),
		TaskID:       payload.TaskID,
		SubmissionID: payload.SubmissionID,
		Reviewer:     payload.AcceptedBy,
		Decision:     review.DecisionAccepted,
		ReviewedAt:   at,
	}); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert review", err)
	}
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	if t.CurrentContractID != "" {
		if err := d.contracts.UpdateStatus(ctx, t.CurrentContractID, contract.StatusCompleted, at); err != nil {
			return toAPIError("complete contract", err)
		}
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleSubmissionRejected(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.RejectSubmissionPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateRejectSubmission(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionRejectSubmission)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "reject submission not allowed", err)
	}
	subItem, err := d.submissions.FindByID(ctx, payload.SubmissionID)
	if err != nil {
		return toAPIError("find submission", err)
	}
	if subItem.TaskID != payload.TaskID {
		return apierr.New(apierr.CodeInvalidMessagePayload, "submission does not belong to task")
	}
	at := chooseTime(payload.RejectedAt, d.now)
	if err := d.submissions.UpdateStatus(ctx, payload.SubmissionID, submission.StatusRejected); err != nil {
		return toAPIError("update submission", err)
	}
	if err := d.reviews.Insert(ctx, &review.Review{
		ReviewID:     "review:" + DeriveEventKey(env),
		TaskID:       payload.TaskID,
		SubmissionID: payload.SubmissionID,
		Reviewer:     payload.RejectedBy,
		Decision:     review.DecisionRejected,
		Reason:       strings.TrimSpace(payload.Reason),
		ReviewedAt:   at,
	}); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert review", err)
	}
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleSettlementRecorded(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.RecordSettlementPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateSettlement(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionRecordSettlement)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "record settlement not allowed", err)
	}
	at := chooseTime(payload.RecordedAt, d.now)
	status := settlement.Status(clawhire.NormalizeSettlementStatus(payload.Status))
	if !status.Valid() {
		return apierr.New(apierr.CodeInvalidMessagePayload, "invalid settlement status")
	}
	item := &settlement.Settlement{
		SettlementID: payload.SettlementID,
		TaskID:       payload.TaskID,
		ContractID:   firstNonEmpty(payload.ContractID, t.CurrentContractID),
		Payee:        payload.Payee,
		Amount:       payload.Amount,
		Currency:     strings.TrimSpace(payload.Currency),
		Status:       status,
		Channel:      strings.TrimSpace(payload.Channel),
		ExternalRef:  strings.TrimSpace(payload.ExternalRef),
		RecordedAt:   at,
	}
	if err := d.settlements.Insert(ctx, item); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert settlement", err)
	}
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleTaskCancelled(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.CancelTaskPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateCancelTask(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionCancelTask)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "cancel task not allowed", err)
	}
	at := chooseTime(payload.CancelledAt, d.now)
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	if t.CurrentContractID != "" {
		if err := d.contracts.UpdateStatus(ctx, t.CurrentContractID, contract.StatusCancelled, at); err != nil {
			return toAPIError("cancel contract", err)
		}
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func (d *CommandDispatcher) handleTaskDisputed(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.DisputeTaskPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	if err := validateDisputeTask(payload); err != nil {
		return err
	}
	t, err := d.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := d.sm.Transit(t.Status, task.ActionDisputeTask)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "dispute task not allowed", err)
	}
	at := chooseTime(payload.DisputedAt, d.now)
	if err := d.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	if t.CurrentContractID != "" {
		if err := d.contracts.UpdateStatus(ctx, t.CurrentContractID, contract.StatusDisputed, at); err != nil {
			return toAPIError("dispute contract", err)
		}
	}
	return d.recordDomainEvent(ctx, env, "task", payload.TaskID, env.Type, payload)
}

func decodeMessage(env *clawsynapse.Envelope, out interface{}) error {
	if err := env.DecodeMessage(out); err != nil {
		return apierr.Wrap(apierr.CodeInvalidMessagePayload, "decode message payload", err)
	}
	return nil
}

func normalizeAcceptanceSpec(spec *clawhire.AcceptanceSpec) task.AcceptanceSpec {
	if spec == nil {
		return task.AcceptanceSpec{Mode: task.AcceptanceModeManual}
	}
	return task.AcceptanceSpec{
		Mode:  task.AcceptanceMode(clawhire.NormalizeAcceptanceMode(spec.Mode)),
		Rules: spec.Rules,
	}
}

func normalizeSettlementTerms(terms *clawhire.SettlementTerms) *task.SettlementTerms {
	if terms == nil {
		return &task.SettlementTerms{Trigger: "on_acceptance"}
	}
	trigger := strings.TrimSpace(terms.Trigger)
	if trigger == "" {
		trigger = "on_acceptance"
	}
	return &task.SettlementTerms{Trigger: trigger}
}

func normalizeClaim(claim *clawhire.Claim) *milestone.Claim {
	if claim == nil {
		return nil
	}
	return &milestone.Claim{
		Type:     strings.TrimSpace(claim.Type),
		Amount:   claim.Amount,
		Currency: strings.TrimSpace(claim.Currency),
	}
}

func normalizeEvidence(ev *clawhire.Evidence) *submission.Evidence {
	if ev == nil {
		return nil
	}
	return &submission.Evidence{
		Type:  strings.TrimSpace(ev.Type),
		Items: ev.Items,
	}
}

func chooseTime(ts *time.Time, now Now) time.Time {
	if ts != nil && !ts.IsZero() {
		return ts.UTC()
	}
	return now().UTC()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func payloadMap(payload interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *CommandDispatcher) recordDomainEvent(ctx context.Context, env *clawsynapse.Envelope, aggregateType, aggregateID, eventType string, payload interface{}) error {
	if d.domainEvts == nil {
		return nil
	}
	data, err := payloadMap(payload)
	if err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "marshal domain event", err)
	}
	if err := d.domainEvts.Insert(ctx, &event.DomainEvent{
		EventID:       DeriveEventKey(env),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Data:          data,
		CreatedAt:     d.now().UTC(),
	}); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert domain event", err)
	}
	return nil
}

func toAPIError(op string, err error) error {
	switch err {
	case nil:
		return nil
	case task.ErrTaskNotFound, submission.ErrSubmissionNotFound, bid.ErrBidNotFound, contract.ErrContractNotFound, settlement.ErrSettlementNotFound, account.ErrAccountNotFound:
		return apierr.Wrap(apierr.CodeNotFound, op, err)
	case task.ErrStatusConflict:
		return apierr.Wrap(apierr.CodeInvalidState, op, err)
	default:
		return apierr.Wrap(apierr.CodeInternalError, op, err)
	}
}

func validateActor(actor shared.Actor, field string) error {
	if strings.TrimSpace(actor.ID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, fmt.Sprintf("%s.id is required", field))
	}
	if !actor.Kind.Valid() {
		return apierr.New(apierr.CodeInvalidMessagePayload, fmt.Sprintf("%s.kind is invalid", field))
	}
	return nil
}

func validatePostTask(p clawhire.PostTaskPayload) error {
	if strings.TrimSpace(p.TaskID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId is required")
	}
	if strings.TrimSpace(p.Title) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "title is required")
	}
	if strings.TrimSpace(p.Category) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "category is required")
	}
	if err := validateActor(p.Requester, "requester"); err != nil {
		return err
	}
	if p.Reviewer != nil {
		if err := validateActor(*p.Reviewer, "reviewer"); err != nil {
			return err
		}
	}
	if strings.TrimSpace(p.Reward.Mode) == "" || strings.TrimSpace(p.Reward.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "reward.mode and reward.currency are required")
	}
	return nil
}

func validatePlaceBid(p clawhire.PlaceBidPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.BidID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and bidId are required")
	}
	if err := validateActor(p.Executor, "executor"); err != nil {
		return err
	}
	if strings.TrimSpace(p.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "currency is required")
	}
	return nil
}

func validateAwardTask(p clawhire.AwardTaskPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.ContractID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and contractId are required")
	}
	if err := validateActor(p.Executor, "executor"); err != nil {
		return err
	}
	if strings.TrimSpace(p.AgreedReward.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "agreedReward.currency is required")
	}
	return nil
}

func validateTaskStart(p clawhire.StartTaskPayload) error {
	if strings.TrimSpace(p.TaskID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId is required")
	}
	return nil
}

func validateProgress(p clawhire.ReportProgressPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.ProgressID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and progressId are required")
	}
	if strings.TrimSpace(p.Summary) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "summary is required")
	}
	return validateActor(p.Executor, "executor")
}

func validateMilestone(p clawhire.CompleteMilestonePayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.MilestoneID) == "" || strings.TrimSpace(p.Title) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId, milestoneId and title are required")
	}
	return nil
}

func validateSubmission(p clawhire.CreateSubmissionPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.SubmissionID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and submissionId are required")
	}
	if strings.TrimSpace(p.Summary) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "summary is required")
	}
	if len(p.Artifacts) == 0 {
		return apierr.New(apierr.CodeInvalidMessagePayload, "artifacts are required")
	}
	return validateActor(p.Executor, "executor")
}

func validateAcceptSubmission(p clawhire.AcceptSubmissionPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.SubmissionID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and submissionId are required")
	}
	return validateActor(p.AcceptedBy, "acceptedBy")
}

func validateRejectSubmission(p clawhire.RejectSubmissionPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.SubmissionID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and submissionId are required")
	}
	if strings.TrimSpace(p.Reason) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "reason is required")
	}
	return validateActor(p.RejectedBy, "rejectedBy")
}

func validateSettlement(p clawhire.RecordSettlementPayload) error {
	if strings.TrimSpace(p.TaskID) == "" || strings.TrimSpace(p.SettlementID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId and settlementId are required")
	}
	if strings.TrimSpace(p.Currency) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "currency is required")
	}
	return validateActor(p.Payee, "payee")
}

func validateCancelTask(p clawhire.CancelTaskPayload) error {
	if strings.TrimSpace(p.TaskID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId is required")
	}
	return nil
}

func validateDisputeTask(p clawhire.DisputeTaskPayload) error {
	if strings.TrimSpace(p.TaskID) == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "taskId is required")
	}
	return nil
}
