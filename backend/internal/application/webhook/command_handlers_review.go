package webhook

import (
	"context"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

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
