package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func (s *Service) AwardTask(ctx context.Context, cmd AwardTaskCommand) error {
	payload := cmd.Payload
	if err := validateAwardTask(payload); err != nil {
		return err
	}
	t, err := s.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return toAPIError("find task", err)
	}
	next, _, err := s.sm.Transit(t.Status, task.ActionAwardTask)
	if err != nil {
		return apierr.Wrap(apierr.CodeInvalidState, "award task not allowed", err)
	}
	if current, err := s.contracts.FindActiveByTask(ctx, payload.TaskID); err == nil {
		return apierr.New(apierr.CodeInvalidState, fmt.Sprintf("task already has active contract %s", current.ContractID))
	} else if err != contract.ErrContractNotFound {
		return toAPIError("find active contract", err)
	}

	now := s.now().UTC()
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
	if err := s.contracts.Insert(ctx, item); err != nil {
		if err == contract.ErrActiveContractExists {
			return apierr.New(apierr.CodeInvalidState, "task already has active contract")
		}
		return apierr.Wrap(apierr.CodeInternalError, "insert contract", err)
	}
	if err := s.tasks.UpdateStatus(ctx, payload.TaskID, t.Status, next, now); err != nil {
		return toAPIError("update task status", err)
	}
	if err := s.tasks.UpdateAssignment(ctx, payload.TaskID, payload.Executor, payload.ContractID, now); err != nil {
		return toAPIError("update assignment", err)
	}
	if bids, _, err := s.bids.ListByTask(ctx, payload.TaskID, 1, 200); err == nil {
		awardedBidID := ""
		for _, item := range bids {
			if item.Executor.ID == payload.Executor.ID && item.Status == bid.StatusActive {
				awardedBidID = item.BidID
				break
			}
		}
		if awardedBidID != "" {
			if err := s.bids.MarkAwarded(ctx, awardedBidID); err != nil {
				return toAPIError("mark bid awarded", err)
			}
		}
		if err := s.bids.InvalidateOthers(ctx, payload.TaskID, awardedBidID); err != nil {
			return apierr.Wrap(apierr.CodeInternalError, "invalidate other bids", err)
		}
	}
	if err := s.recordDomainEvent(ctx, "task", payload.TaskID, cmd.Event, payload); err != nil {
		return err
	}
	if s.syncPub != nil {
		s.syncPub.NotifyTaskAwarded(ctx, t, payload.ContractID, payload.Executor.ID)
	}
	return nil
}
