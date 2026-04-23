package command

import (
	"context"
	"testing"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
)

type fakeTaskRepo struct {
	items map[string]*task.Task
}

func newFakeTaskRepo() *fakeTaskRepo { return &fakeTaskRepo{items: map[string]*task.Task{}} }

func (r *fakeTaskRepo) Insert(_ context.Context, t *task.Task) error {
	cp := *t
	r.items[t.TaskID] = &cp
	return nil
}

func (r *fakeTaskRepo) FindByID(_ context.Context, taskID string) (*task.Task, error) {
	item, ok := r.items[taskID]
	if !ok {
		return nil, task.ErrTaskNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *fakeTaskRepo) UpdateStatus(_ context.Context, taskID string, expected, next task.Status, at time.Time) error {
	item, ok := r.items[taskID]
	if !ok {
		return task.ErrTaskNotFound
	}
	if item.Status != expected {
		return task.ErrStatusConflict
	}
	item.Status = next
	item.UpdatedAt = at
	item.LastActivityAt = &at
	return nil
}

func (r *fakeTaskRepo) UpdateAssignment(_ context.Context, taskID string, executor shared.Actor, contractID string, at time.Time) error {
	return nil
}

func (r *fakeTaskRepo) TouchActivity(_ context.Context, taskID string, at time.Time) error {
	item, ok := r.items[taskID]
	if !ok {
		return task.ErrTaskNotFound
	}
	item.UpdatedAt = at
	item.LastActivityAt = &at
	return nil
}

func (r *fakeTaskRepo) List(_ context.Context, f task.Filter) ([]*task.Task, int64, error) {
	return nil, 0, nil
}

func (r *fakeTaskRepo) ListByExecutor(_ context.Context, executorID string, statuses []task.Status, page, pageSize int) ([]*task.Task, int64, error) {
	return nil, 0, nil
}

type fakeBidRepo struct {
	items map[string]*bid.Bid
}

func newFakeBidRepo() *fakeBidRepo { return &fakeBidRepo{items: map[string]*bid.Bid{}} }

func (r *fakeBidRepo) Insert(_ context.Context, b *bid.Bid) error {
	cp := *b
	r.items[b.BidID] = &cp
	return nil
}

func (r *fakeBidRepo) FindByID(_ context.Context, bidID string) (*bid.Bid, error) {
	item, ok := r.items[bidID]
	if !ok {
		return nil, bid.ErrBidNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *fakeBidRepo) ListByTask(_ context.Context, taskID string, page, pageSize int) ([]*bid.Bid, int64, error) {
	return nil, 0, nil
}

func (r *fakeBidRepo) ListByExecutor(_ context.Context, executorID string, page, pageSize int) ([]*bid.Bid, int64, error) {
	return nil, 0, nil
}

func (r *fakeBidRepo) MarkAwarded(_ context.Context, bidID string) error { return nil }
func (r *fakeBidRepo) InvalidateOthers(_ context.Context, taskID string, exceptBidID string) error {
	return nil
}

type fakeDomainEventRepo struct {
	items []*event.DomainEvent
}

func (r *fakeDomainEventRepo) Insert(_ context.Context, e *event.DomainEvent) error {
	cp := *e
	r.items = append(r.items, &cp)
	return nil
}

func (r *fakeDomainEventRepo) ListByAggregate(_ context.Context, aggType, aggID string, page, pageSize int) ([]*event.DomainEvent, int64, error) {
	return nil, 0, nil
}

func TestService_PostTaskDefaultsReviewerAndRecordsEvent(t *testing.T) {
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	eventRepo := &fakeDomainEventRepo{}
	svc := NewService(Options{
		Tasks:      taskRepo,
		Bids:       newFakeBidRepo(),
		DomainEvts: eventRepo,
		Now:        func() time.Time { return now },
	})

	res, err := svc.PostTask(context.Background(), PostTaskCommand{
		Payload: clawhire.PostTaskPayload{
			TaskID:    "task_001",
			Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
			Title:     "Build landing page",
			Category:  "coding",
			Reward:    clawhire.Reward{Mode: "fixed", Amount: 300, Currency: "USD"},
		},
		Event: &EventMeta{ID: "evt_001", Type: "clawhire.task.posted"},
	})
	if err != nil {
		t.Fatalf("PostTask err = %v", err)
	}
	if res.TaskID != "task_001" || res.EventID != "evt_001" {
		t.Fatalf("unexpected result: %+v", res)
	}
	got, _ := taskRepo.FindByID(context.Background(), "task_001")
	if got.Reviewer == nil || got.Reviewer.ID != "user_001" {
		t.Fatalf("reviewer = %+v", got.Reviewer)
	}
	if len(eventRepo.items) != 1 || eventRepo.items[0].EventID != "evt_001" {
		t.Fatalf("domain events = %+v", eventRepo.items)
	}
}

func TestService_PlaceBidInsertsBid(t *testing.T) {
	now := time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:    "task_001",
		Title:     "Build landing page",
		Category:  "coding",
		Status:    task.StatusOpen,
		Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		Reward:    task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	bidRepo := newFakeBidRepo()
	svc := NewService(Options{
		Tasks: taskRepo,
		Bids:  bidRepo,
		Now:   func() time.Time { return now.Add(time.Minute) },
	})

	res, err := svc.PlaceBid(context.Background(), PlaceBidCommand{
		Payload: clawhire.PlaceBidPayload{
			TaskID:   "task_001",
			BidID:    "bid_001",
			Executor: shared.Actor{ID: "user_002", Kind: shared.ActorKindUser},
			Price:    260,
			Currency: "USD",
		},
	})
	if err != nil {
		t.Fatalf("PlaceBid err = %v", err)
	}
	if res.BidID != "bid_001" {
		t.Fatalf("unexpected result: %+v", res)
	}
	got, _ := bidRepo.FindByID(context.Background(), "bid_001")
	if got.Executor.ID != "user_002" {
		t.Fatalf("unexpected bid: %+v", got)
	}
}
