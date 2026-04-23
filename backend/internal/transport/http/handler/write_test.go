package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/middleware"
)

type writeTaskRepo struct{ items map[string]*task.Task }

func newWriteTaskRepo() *writeTaskRepo { return &writeTaskRepo{items: map[string]*task.Task{}} }

func (r *writeTaskRepo) Insert(_ context.Context, t *task.Task) error {
	cp := *t
	r.items[t.TaskID] = &cp
	return nil
}

func (r *writeTaskRepo) FindByID(_ context.Context, taskID string) (*task.Task, error) {
	item, ok := r.items[taskID]
	if !ok {
		return nil, task.ErrTaskNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *writeTaskRepo) UpdateStatus(_ context.Context, taskID string, expected, next task.Status, at time.Time) error {
	return nil
}

func (r *writeTaskRepo) UpdateAssignment(_ context.Context, taskID string, executor shared.Actor, contractID string, at time.Time) error {
	return nil
}

func (r *writeTaskRepo) TouchActivity(_ context.Context, taskID string, at time.Time) error {
	item, ok := r.items[taskID]
	if !ok {
		return task.ErrTaskNotFound
	}
	item.LastActivityAt = &at
	item.UpdatedAt = at
	return nil
}

func (r *writeTaskRepo) List(_ context.Context, f task.Filter) ([]*task.Task, int64, error) {
	return nil, 0, nil
}

func (r *writeTaskRepo) ListByExecutor(_ context.Context, executorID string, statuses []task.Status, page, pageSize int) ([]*task.Task, int64, error) {
	return nil, 0, nil
}

type writeBidRepo struct{ items map[string]*bid.Bid }

func newWriteBidRepo() *writeBidRepo { return &writeBidRepo{items: map[string]*bid.Bid{}} }

func (r *writeBidRepo) Insert(_ context.Context, b *bid.Bid) error {
	cp := *b
	r.items[b.BidID] = &cp
	return nil
}

func (r *writeBidRepo) FindByID(_ context.Context, bidID string) (*bid.Bid, error) {
	item, ok := r.items[bidID]
	if !ok {
		return nil, bid.ErrBidNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *writeBidRepo) ListByTask(_ context.Context, taskID string, page, pageSize int) ([]*bid.Bid, int64, error) {
	return nil, 0, nil
}

func (r *writeBidRepo) ListByExecutor(_ context.Context, executorID string, page, pageSize int) ([]*bid.Bid, int64, error) {
	return nil, 0, nil
}

func (r *writeBidRepo) MarkAwarded(_ context.Context, bidID string) error { return nil }
func (r *writeBidRepo) InvalidateOthers(_ context.Context, taskID string, exceptBidID string) error {
	return nil
}

type writeAccountRepo struct{ items map[string]*account.Account }

func (r *writeAccountRepo) Insert(context.Context, *account.Account) error { return nil }
func (r *writeAccountRepo) FindByID(_ context.Context, accountID string) (*account.Account, error) {
	item, ok := r.items[accountID]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	cp := *item
	return &cp, nil
}
func (r *writeAccountRepo) FindByNodeID(context.Context, string) (*account.Account, error) {
	return nil, account.ErrAccountNotFound
}
func (r *writeAccountRepo) List(context.Context, account.Filter) ([]*account.Account, int64, error) {
	return nil, 0, nil
}
func (r *writeAccountRepo) ListAgentsByOwner(context.Context, string, int, int) ([]*account.Account, int64, error) {
	return nil, 0, nil
}

type writeEventRepo struct{}

func (writeEventRepo) Insert(context.Context, *event.DomainEvent) error { return nil }
func (writeEventRepo) ListByAggregate(context.Context, string, string, int, int) ([]*event.DomainEvent, int64, error) {
	return nil, 0, nil
}

func TestWrite_CreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	taskRepo := newWriteTaskRepo()
	accountRepo := &writeAccountRepo{items: map[string]*account.Account{
		"acct_human_001": {AccountID: "acct_human_001", Type: account.TypeHuman, Status: account.StatusActive, DisplayName: "Alice"},
	}}
	svc := appcmd.NewService(appcmd.Options{
		Tasks:      taskRepo,
		Bids:       newWriteBidRepo(),
		DomainEvts: writeEventRepo{},
		Now:        func() time.Time { return time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.POST("/api/tasks", w.CreateTask)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{"taskId":"task_001","title":"Build landing page","category":"coding","reward":{"mode":"fixed","amount":300,"currency":"USD"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerAccountID, "acct_human_001")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	got, err := taskRepo.FindByID(context.Background(), "task_001")
	if err != nil {
		t.Fatalf("FindByID err = %v", err)
	}
	if got.Requester.ID != "acct_human_001" || got.Requester.Kind != shared.ActorKindUser {
		t.Fatalf("requester = %+v", got.Requester)
	}
}

func TestWrite_CreateBid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	taskRepo := newWriteTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:    "task_001",
		Title:     "Build landing page",
		Category:  "coding",
		Status:    task.StatusOpen,
		Requester: shared.Actor{ID: "acct_human_001", Kind: shared.ActorKindUser},
		Reward:    task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	bidRepo := newWriteBidRepo()
	accountRepo := &writeAccountRepo{items: map[string]*account.Account{
		"acct_human_002": {AccountID: "acct_human_002", Type: account.TypeHuman, Status: account.StatusActive, DisplayName: "Bob"},
	}}
	svc := appcmd.NewService(appcmd.Options{
		Tasks:      taskRepo,
		Bids:       bidRepo,
		DomainEvts: writeEventRepo{},
		Now:        func() time.Time { return now.Add(time.Minute) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.POST("/api/tasks/:taskId/bids", w.CreateBid)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/task_001/bids", bytes.NewBufferString(`{"bidId":"bid_001","price":260,"currency":"USD","proposal":"Can deliver within 24 hours"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerAccountID, "acct_human_002")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			BidID string `json:"bidId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Success || body.Data.BidID != "bid_001" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
	got, err := bidRepo.FindByID(context.Background(), "bid_001")
	if err != nil {
		t.Fatalf("FindByID err = %v", err)
	}
	if got.Executor.ID != "acct_human_002" || got.Executor.Kind != shared.ActorKindUser {
		t.Fatalf("executor = %+v", got.Executor)
	}
}
