package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

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
