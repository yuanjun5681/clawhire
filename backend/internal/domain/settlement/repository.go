package settlement

import (
	"context"
	"errors"
)

var ErrSettlementNotFound = errors.New("settlement not found")

type Repository interface {
	Insert(ctx context.Context, s *Settlement) error
	FindByID(ctx context.Context, settlementID string) (*Settlement, error)
	ListByTask(ctx context.Context, taskID string) ([]*Settlement, error)
	ListByPayee(ctx context.Context, payeeID string, page, pageSize int) ([]*Settlement, int64, error)
}
