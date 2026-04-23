package webhook

import (
	"context"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

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
