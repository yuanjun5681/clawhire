package event

import (
	"context"
	"errors"
	"time"
)

// ErrDuplicateEvent 表示 eventKey 已存在（幂等命中）。
var ErrDuplicateEvent = errors.New("duplicate event")

type RawEventRepository interface {
	// Insert 在 eventKey 重复时返回 ErrDuplicateEvent。
	Insert(ctx context.Context, e *RawEvent) error
	FindByEventKey(ctx context.Context, eventKey string) (*RawEvent, error)
	MarkProcessed(ctx context.Context, eventKey string, status ProcessStatus, at time.Time, errMsg string) error
}

type DomainEventRepository interface {
	Insert(ctx context.Context, e *DomainEvent) error
	ListByAggregate(ctx context.Context, aggType, aggID string, page, pageSize int) ([]*DomainEvent, int64, error)
}
