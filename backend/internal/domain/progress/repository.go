package progress

import "context"

type Repository interface {
	Insert(ctx context.Context, r *Report) error
	ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*Report, int64, error)
}
