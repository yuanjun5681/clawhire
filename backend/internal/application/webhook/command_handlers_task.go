package webhook

import (
	"context"
	"strings"

	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
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
	_, err := d.commands.PostTask(ctx, appcmd.PostTaskCommand{
		Payload: payload,
		Event: &appcmd.EventMeta{
			ID:   DeriveEventKey(env),
			Type: env.Type,
		},
	})
	return err
}

func (d *CommandDispatcher) handleTaskAwarded(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.AwardTaskPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	return d.commands.AwardTask(ctx, appcmd.AwardTaskCommand{
		Payload: payload,
		Event: &appcmd.EventMeta{
			ID:   DeriveEventKey(env),
			Type: env.Type,
		},
	})
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
