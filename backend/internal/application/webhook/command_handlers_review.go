package webhook

import (
	"context"
	"strings"

	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
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
	return d.commands.AcceptSubmission(ctx, appcmd.AcceptSubmissionCommand{
		Payload: payload,
		Event: &appcmd.EventMeta{
			ID:   DeriveEventKey(env),
			Type: env.Type,
		},
	})
}

func (d *CommandDispatcher) handleSubmissionRejected(ctx context.Context, env *clawsynapse.Envelope) error {
	var payload clawhire.RejectSubmissionPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	return d.commands.RejectSubmission(ctx, appcmd.RejectSubmissionCommand{
		Payload: payload,
		Event: &appcmd.EventMeta{
			ID:   DeriveEventKey(env),
			Type: env.Type,
		},
	})
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
