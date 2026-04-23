package command

import (
	"context"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func (s *Service) CreateSubmission(ctx context.Context, cmd CreateSubmissionCommand) error {
	payload := cmd.Payload
	if err := validateSubmission(payload); err != nil {
		return err
	}
	t, err := s.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := s.sm.Transit(t.Status, task.ActionCreateSubmission)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "create submission not allowed", err)
	}
	submittedAt := s.now().UTC()
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
	if err := s.submissions.Insert(ctx, item); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert submission", err)
	}
	if err := s.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, submittedAt); err != nil {
		return toAPIError("update task status", err)
	}
	return s.recordDomainEvent(ctx, "task", payload.TaskID, cmd.Event, payload)
}

func (s *Service) AcceptSubmission(ctx context.Context, cmd AcceptSubmissionCommand) error {
	payload := cmd.Payload
	if err := validateAcceptSubmission(payload); err != nil {
		return err
	}
	t, err := s.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := s.sm.Transit(t.Status, task.ActionAcceptSubmission)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "accept submission not allowed", err)
	}
	subItem, err := s.submissions.FindByID(ctx, payload.SubmissionID)
	if err != nil {
		return toAPIError("find submission", err)
	}
	if subItem.TaskID != payload.TaskID {
		return apierr.New(apierr.CodeInvalidMessagePayload, "submission does not belong to task")
	}
	at := s.now().UTC()
	if payload.AcceptedAt != nil && !payload.AcceptedAt.IsZero() {
		at = payload.AcceptedAt.UTC()
	}
	if err := s.submissions.UpdateStatus(ctx, payload.SubmissionID, submission.StatusAccepted); err != nil {
		return toAPIError("update submission", err)
	}
	if err := s.reviews.Insert(ctx, &review.Review{
		ReviewID:     reviewID(cmd.Event, payload.TaskID, payload.SubmissionID, review.DecisionAccepted),
		TaskID:       payload.TaskID,
		SubmissionID: payload.SubmissionID,
		Reviewer:     payload.AcceptedBy,
		Decision:     review.DecisionAccepted,
		ReviewedAt:   at,
	}); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert review", err)
	}
	if err := s.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	if t.CurrentContractID != "" {
		if err := s.contracts.UpdateStatus(ctx, t.CurrentContractID, contract.StatusCompleted, at); err != nil {
			return toAPIError("complete contract", err)
		}
	}
	return s.recordDomainEvent(ctx, "task", payload.TaskID, cmd.Event, payload)
}

func (s *Service) RejectSubmission(ctx context.Context, cmd RejectSubmissionCommand) error {
	payload := cmd.Payload
	if err := validateRejectSubmission(payload); err != nil {
		return err
	}
	t, err := s.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := s.sm.Transit(t.Status, task.ActionRejectSubmission)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "reject submission not allowed", err)
	}
	subItem, err := s.submissions.FindByID(ctx, payload.SubmissionID)
	if err != nil {
		return toAPIError("find submission", err)
	}
	if subItem.TaskID != payload.TaskID {
		return apierr.New(apierr.CodeInvalidMessagePayload, "submission does not belong to task")
	}
	at := s.now().UTC()
	if payload.RejectedAt != nil && !payload.RejectedAt.IsZero() {
		at = payload.RejectedAt.UTC()
	}
	if err := s.submissions.UpdateStatus(ctx, payload.SubmissionID, submission.StatusRejected); err != nil {
		return toAPIError("update submission", err)
	}
	if err := s.reviews.Insert(ctx, &review.Review{
		ReviewID:     reviewID(cmd.Event, payload.TaskID, payload.SubmissionID, review.DecisionRejected),
		TaskID:       payload.TaskID,
		SubmissionID: payload.SubmissionID,
		Reviewer:     payload.RejectedBy,
		Decision:     review.DecisionRejected,
		Reason:       strings.TrimSpace(payload.Reason),
		ReviewedAt:   at,
	}); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert review", err)
	}
	if err := s.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, at); err != nil {
		return toAPIError("update task status", err)
	}
	return s.recordDomainEvent(ctx, "task", payload.TaskID, cmd.Event, payload)
}
