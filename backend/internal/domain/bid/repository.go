package bid

import (
	"context"
	"errors"
)

var ErrBidNotFound = errors.New("bid not found")

type Repository interface {
	Insert(ctx context.Context, b *Bid) error
	FindByID(ctx context.Context, bidID string) (*Bid, error)
	ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*Bid, int64, error)
	ListByExecutor(ctx context.Context, executorID string, page, pageSize int) ([]*Bid, int64, error)

	// MarkAwarded 将指定 bid 置为 awarded。
	MarkAwarded(ctx context.Context, bidID string) error

	// InvalidateOthers 把同一任务下、除 exceptBidID 外所有 active bid 置为 rejected。
	// 若 exceptBidID 为空，则全部置为 rejected。
	InvalidateOthers(ctx context.Context, taskID string, exceptBidID string) error
}
