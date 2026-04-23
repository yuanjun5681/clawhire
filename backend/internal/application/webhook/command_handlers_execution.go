package webhook

import (
	"context"
	"strings"

	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/milestone"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/progress"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func (d *CommandDispatcher) handleBidPlaced(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.PlaceBidPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	_, err := d.commands.PlaceBid(ctx, appcmd.PlaceBidCommand{
		Payload: payload,
		Event: &appcmd.EventMeta{
			ID:   DeriveEventKey(env),
			Type: env.Type,
		},
	})
	return err
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
