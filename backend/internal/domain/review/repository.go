package review

import "context"

type Repository interface {
	Insert(ctx context.Context, r *Review) error
	ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*Review, int64, error)
	ListBySubmission(ctx context.Context, submissionID string) ([]*Review, error)
}
