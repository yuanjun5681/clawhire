package submission

import (
	"context"
	"errors"
)

var ErrSubmissionNotFound = errors.New("submission not found")

type Repository interface {
	Insert(ctx context.Context, s *Submission) error
	FindByID(ctx context.Context, submissionID string) (*Submission, error)
	UpdateStatus(ctx context.Context, submissionID string, status Status) error
	ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*Submission, int64, error)
	LatestByTask(ctx context.Context, taskID string) (*Submission, error)
}
