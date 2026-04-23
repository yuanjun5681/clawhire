package webhook

import (
	"context"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

type commandFunc func(context.Context, *clawsynapse.Envelope) error

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
