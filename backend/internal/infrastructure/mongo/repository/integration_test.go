package repository

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

func mongoTestClient(t *testing.T) *mgo.Client {
	t.Helper()

	uri := strings.TrimSpace(os.Getenv("MONGODB_URI_TEST"))
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017"
	}
	dbName := "clawhire_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mgo.NewClient(ctx, uri, dbName)
	if err != nil {
		t.Skipf("skip mongo integration test: %v", err)
	}
	if err := mgo.EnsureIndexes(ctx, client.DB()); err != nil {
		_ = client.Close(ctx)
		t.Fatalf("ensure indexes: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_ = client.DB().Drop(cleanupCtx)
		_ = client.Close(cleanupCtx)
	})
	return client
}

func TestTaskRepo_RealMongoLifecycle(t *testing.T) {
	client := mongoTestClient(t)
	repo := NewTaskRepo(client.DB())
	ctx := context.Background()
	now := time.Date(2026, 4, 23, 20, 0, 0, 0, time.UTC)

	item := &task.Task{
		TaskID:    "task_001",
		Title:     "Build landing page",
		Category:  "coding",
		Status:    task.StatusOpen,
		Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		Reward:    task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repo.Insert(ctx, item); err != nil {
		t.Fatalf("insert task: %v", err)
	}

	got, err := repo.FindByID(ctx, "task_001")
	if err != nil {
		t.Fatalf("find task: %v", err)
	}
	if got.Status != task.StatusOpen {
		t.Fatalf("status = %s, want %s", got.Status, task.StatusOpen)
	}

	if err := repo.UpdateAssignment(ctx, "task_001", shared.Actor{ID: "agent_007", Kind: shared.ActorKindAgent}, "contract_001", now.Add(time.Minute)); err != nil {
		t.Fatalf("update assignment: %v", err)
	}
	if err := repo.UpdateStatus(ctx, "task_001", task.StatusOpen, task.StatusAwarded, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("update status: %v", err)
	}

	got, err = repo.FindByID(ctx, "task_001")
	if err != nil {
		t.Fatalf("find task after update: %v", err)
	}
	if got.AssignedExecutor == nil || got.AssignedExecutor.ID != "agent_007" {
		t.Fatalf("unexpected assigned executor: %+v", got.AssignedExecutor)
	}
	if got.Status != task.StatusAwarded {
		t.Fatalf("status after update = %s, want %s", got.Status, task.StatusAwarded)
	}

	list, total, err := repo.ListByExecutor(ctx, "agent_007", []task.Status{task.StatusAwarded}, 1, 20)
	if err != nil {
		t.Fatalf("list by executor: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("unexpected list result: total=%d len=%d", total, len(list))
	}
}

func TestBidRepo_RealMongoAwardAndInvalidate(t *testing.T) {
	client := mongoTestClient(t)
	repo := NewBidRepo(client.DB())
	ctx := context.Background()
	now := time.Date(2026, 4, 23, 21, 0, 0, 0, time.UTC)

	first := &bid.Bid{
		BidID:     "bid_001",
		TaskID:    "task_001",
		Executor:  shared.Actor{ID: "agent_001", Kind: shared.ActorKindAgent},
		Price:     100,
		Currency:  "USD",
		Status:    bid.StatusActive,
		CreatedAt: now,
	}
	second := &bid.Bid{
		BidID:     "bid_002",
		TaskID:    "task_001",
		Executor:  shared.Actor{ID: "agent_002", Kind: shared.ActorKindAgent},
		Price:     120,
		Currency:  "USD",
		Status:    bid.StatusActive,
		CreatedAt: now.Add(time.Minute),
	}
	if err := repo.Insert(ctx, first); err != nil {
		t.Fatalf("insert first bid: %v", err)
	}
	if err := repo.Insert(ctx, second); err != nil {
		t.Fatalf("insert second bid: %v", err)
	}
	if err := repo.MarkAwarded(ctx, "bid_002"); err != nil {
		t.Fatalf("mark awarded: %v", err)
	}
	if err := repo.InvalidateOthers(ctx, "task_001", "bid_002"); err != nil {
		t.Fatalf("invalidate others: %v", err)
	}

	gotFirst, _ := repo.FindByID(ctx, "bid_001")
	gotSecond, _ := repo.FindByID(ctx, "bid_002")
	if gotFirst.Status != bid.StatusRejected {
		t.Fatalf("first bid status = %s, want %s", gotFirst.Status, bid.StatusRejected)
	}
	if gotSecond.Status != bid.StatusAwarded {
		t.Fatalf("second bid status = %s, want %s", gotSecond.Status, bid.StatusAwarded)
	}
}

func TestContractRepo_RealMongoAllowsOnlyOneActivePerTask(t *testing.T) {
	client := mongoTestClient(t)
	repo := NewContractRepo(client.DB())
	ctx := context.Background()
	now := time.Date(2026, 4, 23, 21, 30, 0, 0, time.UTC)

	first := &contract.Contract{
		ContractID:   "contract_001",
		TaskID:       "task_001",
		Requester:    shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		Executor:     shared.Actor{ID: "agent_001", Kind: shared.ActorKindAgent},
		AgreedReward: shared.Money{Amount: 100, Currency: "USD"},
		Status:       contract.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	second := &contract.Contract{
		ContractID:   "contract_002",
		TaskID:       "task_001",
		Requester:    shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		Executor:     shared.Actor{ID: "agent_002", Kind: shared.ActorKindAgent},
		AgreedReward: shared.Money{Amount: 120, Currency: "USD"},
		Status:       contract.StatusActive,
		CreatedAt:    now.Add(time.Minute),
		UpdatedAt:    now.Add(time.Minute),
	}

	if err := repo.Insert(ctx, first); err != nil {
		t.Fatalf("insert first contract: %v", err)
	}
	if err := repo.Insert(ctx, second); !errors.Is(err, contract.ErrActiveContractExists) {
		t.Fatalf("expected active contract exists, got %v", err)
	}
}

func TestSubmissionReviewSettlementAccountAndEventRepos_RealMongo(t *testing.T) {
	client := mongoTestClient(t)
	ctx := context.Background()
	now := time.Date(2026, 4, 23, 22, 0, 0, 0, time.UTC)

	subRepo := NewSubmissionRepo(client.DB())
	revRepo := NewReviewRepo(client.DB())
	settleRepo := NewSettlementRepo(client.DB())
	accountRepo := NewAccountRepo(client.DB())
	rawRepo := NewRawEventRepo(client.DB())
	domainRepo := NewDomainEventRepo(client.DB())

	sub := &submission.Submission{
		SubmissionID: "submission_001",
		TaskID:       "task_001",
		Executor:     shared.Actor{ID: "agent_007", Kind: shared.ActorKindAgent},
		Summary:      "done",
		Artifacts:    []shared.Artifact{{Type: shared.ArtifactTypeURL, Value: "https://example.com"}},
		Status:       submission.StatusSubmitted,
		SubmittedAt:  now,
	}
	if err := subRepo.Insert(ctx, sub); err != nil {
		t.Fatalf("insert submission: %v", err)
	}
	if err := subRepo.UpdateStatus(ctx, "submission_001", submission.StatusAccepted); err != nil {
		t.Fatalf("update submission: %v", err)
	}
	latest, err := subRepo.LatestByTask(ctx, "task_001")
	if err != nil {
		t.Fatalf("latest submission: %v", err)
	}
	if latest.Status != submission.StatusAccepted {
		t.Fatalf("latest submission status = %s, want %s", latest.Status, submission.StatusAccepted)
	}

	if err := revRepo.Insert(ctx, &review.Review{
		ReviewID:     "review_001",
		TaskID:       "task_001",
		SubmissionID: "submission_001",
		Reviewer:     shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		Decision:     review.DecisionAccepted,
		ReviewedAt:   now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("insert review: %v", err)
	}
	reviews, total, err := revRepo.ListByTask(ctx, "task_001", 1, 20)
	if err != nil || total != 1 || len(reviews) != 1 {
		t.Fatalf("list reviews: total=%d len=%d err=%v", total, len(reviews), err)
	}

	if err := settleRepo.Insert(ctx, &settlement.Settlement{
		SettlementID: "settlement_001",
		TaskID:       "task_001",
		Payee:        shared.Actor{ID: "agent_007", Kind: shared.ActorKindAgent},
		Amount:       260,
		Currency:     "USD",
		Status:       settlement.StatusRecorded,
		RecordedAt:   now.Add(2 * time.Minute),
	}); err != nil {
		t.Fatalf("insert settlement: %v", err)
	}
	settlements, err := settleRepo.ListByTask(ctx, "task_001")
	if err != nil || len(settlements) != 1 {
		t.Fatalf("list settlements: len=%d err=%v", len(settlements), err)
	}

	if err := accountRepo.Insert(ctx, &account.Account{
		AccountID:   "agent_007",
		Type:        account.TypeAgent,
		DisplayName: "BuilderBot",
		Status:      account.StatusActive,
		NodeID:      "node-007",
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("insert account: %v", err)
	}
	foundAccount, err := accountRepo.FindByNodeID(ctx, "node-007")
	if err != nil || foundAccount.AccountID != "agent_007" {
		t.Fatalf("find account by node id: %+v err=%v", foundAccount, err)
	}

	if err := rawRepo.Insert(ctx, &event.RawEvent{
		EventKey:      "evt_001",
		Source:        "clawsynapse",
		MessageType:   "clawhire.task.posted",
		Payload:       map[string]interface{}{"taskId": "task_001"},
		ReceivedAt:    now,
		ProcessStatus: event.ProcessStatusPending,
	}); err != nil {
		t.Fatalf("insert raw event: %v", err)
	}
	if err := rawRepo.MarkProcessed(ctx, "evt_001", event.ProcessStatusSucceeded, now.Add(3*time.Minute), ""); err != nil {
		t.Fatalf("mark raw processed: %v", err)
	}
	rawItem, err := rawRepo.FindByEventKey(ctx, "evt_001")
	if err != nil || rawItem.ProcessStatus != event.ProcessStatusSucceeded {
		t.Fatalf("find raw event: %+v err=%v", rawItem, err)
	}

	if err := domainRepo.Insert(ctx, &event.DomainEvent{
		EventID:       "evt_001",
		AggregateType: "task",
		AggregateID:   "task_001",
		EventType:     "clawhire.task.posted",
		Data:          map[string]interface{}{"taskId": "task_001"},
		CreatedAt:     now,
	}); err != nil {
		t.Fatalf("insert domain event: %v", err)
	}
	domainEvents, total, err := domainRepo.ListByAggregate(ctx, "task", "task_001", 1, 20)
	if err != nil || total != 1 || len(domainEvents) != 1 {
		t.Fatalf("list domain events: total=%d len=%d err=%v", total, len(domainEvents), err)
	}
}
