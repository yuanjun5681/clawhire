package milestone

import "context"

type Repository interface {
	Upsert(ctx context.Context, m *Milestone) error
	FindByID(ctx context.Context, milestoneID string) (*Milestone, error)
	ListByTask(ctx context.Context, taskID string) ([]*Milestone, error)
}
