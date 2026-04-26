package webhook

import (
	"context"

	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
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
	_, err := d.commands.RecordSettlement(ctx, appcmd.RecordSettlementCommand{
		Payload: payload,
		Event: &appcmd.EventMeta{
			ID:   DeriveEventKey(env),
			Type: env.Type,
		},
	})
	return err
}
