package command

import (
	"context"
	"testing"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
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
	item, ok := r.items[taskID]
	if !ok {
		return task.ErrTaskNotFound
	}
	cp := executor
	item.AssignedExecutor = &cp
	item.CurrentContractID = contractID
	item.UpdatedAt = at
	item.LastActivityAt = &at
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

type fakeContractRepo struct {
	items map[string]*contract.Contract
}

func newFakeContractRepo() *fakeContractRepo {
	return &fakeContractRepo{items: map[string]*contract.Contract{}}
}

func (r *fakeContractRepo) Insert(_ context.Context, c *contract.Contract) error {
	cp := *c
	r.items[c.ContractID] = &cp
	return nil
}

func (r *fakeContractRepo) FindByID(_ context.Context, contractID string) (*contract.Contract, error) {
	item, ok := r.items[contractID]
	if !ok {
		return nil, contract.ErrContractNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *fakeContractRepo) FindActiveByTask(_ context.Context, taskID string) (*contract.Contract, error) {
	for _, item := range r.items {
		if item.TaskID == taskID && item.Status == contract.StatusActive {
			cp := *item
			return &cp, nil
		}
	}
	return nil, contract.ErrContractNotFound
}

func (r *fakeContractRepo) UpdateStatus(_ context.Context, contractID string, status contract.Status, at time.Time) error {
	item, ok := r.items[contractID]
	if !ok {
		return contract.ErrContractNotFound
	}
	item.Status = status
	item.UpdatedAt = at
	return nil
}

func (r *fakeContractRepo) MarkStarted(_ context.Context, contractID string, at time.Time) error {
	return nil
}

type fakeSubmissionRepo struct {
	items map[string]*submission.Submission
}

func newFakeSubmissionRepo() *fakeSubmissionRepo {
	return &fakeSubmissionRepo{items: map[string]*submission.Submission{}}
}

func (r *fakeSubmissionRepo) Insert(_ context.Context, s *submission.Submission) error {
	cp := *s
	r.items[s.SubmissionID] = &cp
	return nil
}

func (r *fakeSubmissionRepo) FindByID(_ context.Context, submissionID string) (*submission.Submission, error) {
	item, ok := r.items[submissionID]
	if !ok {
		return nil, submission.ErrSubmissionNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *fakeSubmissionRepo) UpdateStatus(_ context.Context, submissionID string, status submission.Status) error {
	item, ok := r.items[submissionID]
	if !ok {
		return submission.ErrSubmissionNotFound
	}
	item.Status = status
	return nil
}

func (r *fakeSubmissionRepo) ListByTask(_ context.Context, taskID string, page, pageSize int) ([]*submission.Submission, int64, error) {
	return nil, 0, nil
}

func (r *fakeSubmissionRepo) LatestByTask(_ context.Context, taskID string) (*submission.Submission, error) {
	return nil, submission.ErrSubmissionNotFound
}

type fakeReviewRepo struct {
	items map[string]*review.Review
}

func newFakeReviewRepo() *fakeReviewRepo { return &fakeReviewRepo{items: map[string]*review.Review{}} }

func (r *fakeReviewRepo) Insert(_ context.Context, item *review.Review) error {
	cp := *item
	r.items[item.ReviewID] = &cp
	return nil
}

func (r *fakeReviewRepo) ListByTask(_ context.Context, taskID string, page, pageSize int) ([]*review.Review, int64, error) {
	return nil, 0, nil
}

func (r *fakeReviewRepo) ListBySubmission(_ context.Context, submissionID string) ([]*review.Review, error) {
	return nil, nil
}

func TestService_PostTaskDefaultsReviewerAndRecordsEvent(t *testing.T) {
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	eventRepo := &fakeDomainEventRepo{}
	svc := NewService(Options{
		Tasks:       taskRepo,
		Bids:        newFakeBidRepo(),
		Contracts:   newFakeContractRepo(),
		Submissions: newFakeSubmissionRepo(),
		Reviews:     newFakeReviewRepo(),
		DomainEvts:  eventRepo,
		Now:         func() time.Time { return now },
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
		Tasks:       taskRepo,
		Bids:        bidRepo,
		Contracts:   newFakeContractRepo(),
		Submissions: newFakeSubmissionRepo(),
		Reviews:     newFakeReviewRepo(),
		Now:         func() time.Time { return now.Add(time.Minute) },
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

func TestService_PlaceBidOnAwardedTaskFails(t *testing.T) {
	now := time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:    "task_001",
		Title:     "Build landing page",
		Category:  "coding",
		Status:    task.StatusAwarded,
		Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		Reward:    task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	svc := NewService(Options{
		Tasks:       taskRepo,
		Bids:        newFakeBidRepo(),
		Contracts:   newFakeContractRepo(),
		Submissions: newFakeSubmissionRepo(),
		Reviews:     newFakeReviewRepo(),
		Now:         func() time.Time { return now },
	})

	_, err := svc.PlaceBid(context.Background(), PlaceBidCommand{
		Payload: clawhire.PlaceBidPayload{
			TaskID:   "task_001",
			BidID:    "bid_001",
			Executor: shared.Actor{ID: "user_002", Kind: shared.ActorKindUser},
			Price:    260,
			Currency: "USD",
		},
	})
	if err == nil {
		t.Fatal("expected invalid state error")
	}
}

func TestService_AwardTaskCreatesContractAndUpdatesAssignment(t *testing.T) {
	now := time.Date(2026, 4, 23, 13, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:    "task_001",
		Status:    task.StatusOpen,
		Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		Reward:    task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	bidRepo := newFakeBidRepo()
	contractRepo := newFakeContractRepo()
	svc := NewService(Options{
		Tasks:       taskRepo,
		Bids:        bidRepo,
		Contracts:   contractRepo,
		Submissions: newFakeSubmissionRepo(),
		Reviews:     newFakeReviewRepo(),
		Now:         func() time.Time { return now },
	})

	err := svc.AwardTask(context.Background(), AwardTaskCommand{
		Payload: clawhire.AwardTaskPayload{
			TaskID:     "task_001",
			ContractID: "contract_001",
			Executor:   shared.Actor{ID: "agent_007", Kind: shared.ActorKindAgent},
			AgreedReward: clawhire.Reward{
				Amount:   260,
				Currency: "USD",
			},
		},
		Event: &EventMeta{ID: "evt_award", Type: "clawhire.task.awarded"},
	})
	if err != nil {
		t.Fatalf("AwardTask err = %v", err)
	}
	gotTask, _ := taskRepo.FindByID(context.Background(), "task_001")
	if gotTask.Status != task.StatusAwarded || gotTask.AssignedExecutor == nil || gotTask.AssignedExecutor.ID != "agent_007" {
		t.Fatalf("unexpected task: %+v", gotTask)
	}
	gotContract, _ := contractRepo.FindByID(context.Background(), "contract_001")
	if gotContract.Status != contract.StatusActive || gotContract.Executor.ID != "agent_007" {
		t.Fatalf("unexpected contract: %+v", gotContract)
	}
}

func TestService_AwardTaskWithExistingActiveContractFails(t *testing.T) {
	now := time.Date(2026, 4, 23, 13, 30, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:    "task_001",
		Status:    task.StatusOpen,
		Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
	}
	contractRepo := newFakeContractRepo()
	contractRepo.items["contract_existing"] = &contract.Contract{
		ContractID: "contract_existing",
		TaskID:     "task_001",
		Status:     contract.StatusActive,
	}
	svc := NewService(Options{
		Tasks:       taskRepo,
		Bids:        newFakeBidRepo(),
		Contracts:   contractRepo,
		Submissions: newFakeSubmissionRepo(),
		Reviews:     newFakeReviewRepo(),
		Now:         func() time.Time { return now },
	})

	err := svc.AwardTask(context.Background(), AwardTaskCommand{
		Payload: clawhire.AwardTaskPayload{
			TaskID:     "task_001",
			ContractID: "contract_002",
			Executor:   shared.Actor{ID: "agent_007", Kind: shared.ActorKindAgent},
			AgreedReward: clawhire.Reward{
				Amount:   260,
				Currency: "USD",
			},
		},
	})
	if err == nil {
		t.Fatal("expected invalid state error")
	}
}

func TestService_AwardTaskWithoutTaskFails(t *testing.T) {
	now := time.Date(2026, 4, 23, 12, 30, 0, 0, time.UTC)
	svc := NewService(Options{
		Tasks:       newFakeTaskRepo(),
		Bids:        newFakeBidRepo(),
		Contracts:   newFakeContractRepo(),
		Submissions: newFakeSubmissionRepo(),
		Reviews:     newFakeReviewRepo(),
		Now:         func() time.Time { return now },
	})

	err := svc.AwardTask(context.Background(), AwardTaskCommand{
		Payload: clawhire.AwardTaskPayload{
			TaskID:     "missing",
			ContractID: "contract_001",
			Executor:   shared.Actor{ID: "agent_007", Kind: shared.ActorKindAgent},
			AgreedReward: clawhire.Reward{
				Amount:   260,
				Currency: "USD",
			},
		},
	})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestService_CreateSubmissionTransitionsTask(t *testing.T) {
	now := time.Date(2026, 4, 23, 14, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:            "task_001",
		Status:            task.StatusInProgress,
		Requester:         shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		CurrentContractID: "contract_001",
	}
	subRepo := newFakeSubmissionRepo()
	svc := NewService(Options{
		Tasks:       taskRepo,
		Bids:        newFakeBidRepo(),
		Contracts:   newFakeContractRepo(),
		Submissions: subRepo,
		Reviews:     newFakeReviewRepo(),
		Now:         func() time.Time { return now },
	})

	err := svc.CreateSubmission(context.Background(), CreateSubmissionCommand{
		Payload: clawhire.CreateSubmissionPayload{
			TaskID:       "task_001",
			SubmissionID: "submission_001",
			Executor:     shared.Actor{ID: "agent_007", Kind: shared.ActorKindAgent},
			Summary:      "Landing page delivered",
			Artifacts:    []shared.Artifact{{Type: "url", Value: "https://example.com/result"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateSubmission err = %v", err)
	}
	gotTask, _ := taskRepo.FindByID(context.Background(), "task_001")
	if gotTask.Status != task.StatusSubmitted {
		t.Fatalf("task status = %s, want %s", gotTask.Status, task.StatusSubmitted)
	}
	gotSub, _ := subRepo.FindByID(context.Background(), "submission_001")
	if gotSub.Status != submission.StatusSubmitted {
		t.Fatalf("submission status = %s, want %s", gotSub.Status, submission.StatusSubmitted)
	}
}

func TestService_AcceptSubmissionCompletesTaskAndContract(t *testing.T) {
	now := time.Date(2026, 4, 23, 15, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:            "task_001",
		Status:            task.StatusSubmitted,
		CurrentContractID: "contract_001",
	}
	subRepo := newFakeSubmissionRepo()
	subRepo.items["submission_001"] = &submission.Submission{
		SubmissionID: "submission_001",
		TaskID:       "task_001",
		Status:       submission.StatusSubmitted,
		SubmittedAt:  now.Add(-time.Hour),
	}
	contractRepo := newFakeContractRepo()
	contractRepo.items["contract_001"] = &contract.Contract{
		ContractID: "contract_001",
		TaskID:     "task_001",
		Status:     contract.StatusActive,
	}
	reviewRepo := newFakeReviewRepo()
	svc := NewService(Options{
		Tasks:       taskRepo,
		Bids:        newFakeBidRepo(),
		Contracts:   contractRepo,
		Submissions: subRepo,
		Reviews:     reviewRepo,
		Now:         func() time.Time { return now },
	})

	err := svc.AcceptSubmission(context.Background(), AcceptSubmissionCommand{
		Payload: clawhire.AcceptSubmissionPayload{
			TaskID:       "task_001",
			SubmissionID: "submission_001",
			AcceptedBy:   shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		},
		Event: &EventMeta{ID: "evt_accept", Type: "clawhire.submission.accepted"},
	})
	if err != nil {
		t.Fatalf("AcceptSubmission err = %v", err)
	}
	gotTask, _ := taskRepo.FindByID(context.Background(), "task_001")
	if gotTask.Status != task.StatusAccepted {
		t.Fatalf("task status = %s, want %s", gotTask.Status, task.StatusAccepted)
	}
	gotSub, _ := subRepo.FindByID(context.Background(), "submission_001")
	if gotSub.Status != submission.StatusAccepted {
		t.Fatalf("submission status = %s, want %s", gotSub.Status, submission.StatusAccepted)
	}
	gotContract, _ := contractRepo.FindByID(context.Background(), "contract_001")
	if gotContract.Status != contract.StatusCompleted {
		t.Fatalf("contract status = %s, want %s", gotContract.Status, contract.StatusCompleted)
	}
	if len(reviewRepo.items) != 1 {
		t.Fatalf("review count = %d, want 1", len(reviewRepo.items))
	}
}

func TestService_RejectSubmissionMarksRejectedAndCreatesReview(t *testing.T) {
	now := time.Date(2026, 4, 23, 16, 0, 0, 0, time.UTC)
	taskRepo := newFakeTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:            "task_001",
		Status:            task.StatusSubmitted,
		CurrentContractID: "contract_001",
	}
	subRepo := newFakeSubmissionRepo()
	subRepo.items["submission_001"] = &submission.Submission{
		SubmissionID: "submission_001",
		TaskID:       "task_001",
		Status:       submission.StatusSubmitted,
		SubmittedAt:  now.Add(-time.Hour),
	}
	reviewRepo := newFakeReviewRepo()
	svc := NewService(Options{
		Tasks:       taskRepo,
		Bids:        newFakeBidRepo(),
		Contracts:   newFakeContractRepo(),
		Submissions: subRepo,
		Reviews:     reviewRepo,
		Now:         func() time.Time { return now },
	})

	err := svc.RejectSubmission(context.Background(), RejectSubmissionCommand{
		Payload: clawhire.RejectSubmissionPayload{
			TaskID:       "task_001",
			SubmissionID: "submission_001",
			RejectedBy:   shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
			Reason:       "Missing mobile adaptation",
		},
		Event: &EventMeta{ID: "evt_reject", Type: "clawhire.submission.rejected"},
	})
	if err != nil {
		t.Fatalf("RejectSubmission err = %v", err)
	}
	gotTask, _ := taskRepo.FindByID(context.Background(), "task_001")
	if gotTask.Status != task.StatusRejected {
		t.Fatalf("task status = %s, want %s", gotTask.Status, task.StatusRejected)
	}
	gotSub, _ := subRepo.FindByID(context.Background(), "submission_001")
	if gotSub.Status != submission.StatusRejected {
		t.Fatalf("submission status = %s, want %s", gotSub.Status, submission.StatusRejected)
	}
	if len(reviewRepo.items) != 1 {
		t.Fatalf("review count = %d, want 1", len(reviewRepo.items))
	}
}
