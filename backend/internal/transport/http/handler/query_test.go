package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/milestone"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/progress"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
)

type queryTaskRepo struct {
	items map[string]*task.Task
}

func (r *queryTaskRepo) Insert(context.Context, *task.Task) error { return nil }
func (r *queryTaskRepo) UpdateStatus(context.Context, string, task.Status, task.Status, time.Time) error {
	return nil
}
func (r *queryTaskRepo) UpdateAssignment(context.Context, string, shared.Actor, string, time.Time) error {
	return nil
}
func (r *queryTaskRepo) TouchActivity(context.Context, string, time.Time) error { return nil }
func (r *queryTaskRepo) FindByID(_ context.Context, taskID string) (*task.Task, error) {
	item, ok := r.items[taskID]
	if !ok {
		return nil, task.ErrTaskNotFound
	}
	cp := *item
	return &cp, nil
}
func (r *queryTaskRepo) List(_ context.Context, f task.Filter) ([]*task.Task, int64, error) {
	var list []*task.Task
	for _, item := range r.items {
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
}
func (r *queryTaskRepo) ListByExecutor(_ context.Context, executorID string, statuses []task.Status, page, pageSize int) ([]*task.Task, int64, error) {
	var list []*task.Task
	for _, item := range r.items {
		if item.AssignedExecutor != nil && item.AssignedExecutor.ID == executorID {
			cp := *item
			list = append(list, &cp)
		}
	}
	return list, int64(len(list)), nil
}

type queryAccountRepo struct {
	items []*account.Account
}

func (r *queryAccountRepo) Insert(context.Context, *account.Account) error { return nil }
func (r *queryAccountRepo) FindByID(context.Context, string) (*account.Account, error) {
	return nil, account.ErrAccountNotFound
}
func (r *queryAccountRepo) FindByNodeID(context.Context, string) (*account.Account, error) {
	return nil, account.ErrAccountNotFound
}
func (r *queryAccountRepo) List(context.Context, account.Filter) ([]*account.Account, int64, error) {
	return r.items, int64(len(r.items)), nil
}
func (r *queryAccountRepo) ListAgentsByOwner(context.Context, string, int, int) ([]*account.Account, int64, error) {
	return r.items, int64(len(r.items)), nil
}

type emptyBidRepo struct{}

func (emptyBidRepo) Insert(context.Context, *bid.Bid) error { return nil }
func (emptyBidRepo) FindByID(context.Context, string) (*bid.Bid, error) {
	return nil, bid.ErrBidNotFound
}
func (emptyBidRepo) ListByTask(context.Context, string, int, int) ([]*bid.Bid, int64, error) {
	return nil, 0, nil
}
func (emptyBidRepo) ListByExecutor(context.Context, string, int, int) ([]*bid.Bid, int64, error) {
	return nil, 0, nil
}
func (emptyBidRepo) MarkAwarded(context.Context, string) error              { return nil }
func (emptyBidRepo) InvalidateOthers(context.Context, string, string) error { return nil }

type emptyProgressRepo struct{}

func (emptyProgressRepo) Insert(context.Context, *progress.Report) error { return nil }
func (emptyProgressRepo) ListByTask(context.Context, string, int, int) ([]*progress.Report, int64, error) {
	return nil, 0, nil
}

type emptyMilestoneRepo struct{}

func (emptyMilestoneRepo) Upsert(context.Context, *milestone.Milestone) error { return nil }
func (emptyMilestoneRepo) FindByID(context.Context, string) (*milestone.Milestone, error) {
	return nil, nil
}
func (emptyMilestoneRepo) ListByTask(context.Context, string) ([]*milestone.Milestone, error) {
	return nil, nil
}

type emptySubmissionRepo struct{}

func (emptySubmissionRepo) Insert(context.Context, *submission.Submission) error { return nil }
func (emptySubmissionRepo) FindByID(context.Context, string) (*submission.Submission, error) {
	return nil, submission.ErrSubmissionNotFound
}
func (emptySubmissionRepo) UpdateStatus(context.Context, string, submission.Status) error { return nil }
func (emptySubmissionRepo) ListByTask(context.Context, string, int, int) ([]*submission.Submission, int64, error) {
	return nil, 0, nil
}
func (emptySubmissionRepo) LatestByTask(context.Context, string) (*submission.Submission, error) {
	return nil, submission.ErrSubmissionNotFound
}

type reviewRepoStub struct {
	items []*review.Review
}

func (r reviewRepoStub) Insert(context.Context, *review.Review) error { return nil }
func (r reviewRepoStub) ListByTask(context.Context, string, int, int) ([]*review.Review, int64, error) {
	return r.items, int64(len(r.items)), nil
}
func (r reviewRepoStub) ListBySubmission(context.Context, string) ([]*review.Review, error) {
	return r.items, nil
}

type settlementRepoStub struct {
	items []*settlement.Settlement
}

func (r settlementRepoStub) Insert(context.Context, *settlement.Settlement) error { return nil }
func (r settlementRepoStub) FindByID(context.Context, string) (*settlement.Settlement, error) {
	return nil, settlement.ErrSettlementNotFound
}
func (r settlementRepoStub) ListByTask(context.Context, string) ([]*settlement.Settlement, error) {
	return r.items, nil
}
func (r settlementRepoStub) ListByPayee(context.Context, string, int, int) ([]*settlement.Settlement, int64, error) {
	return r.items, int64(len(r.items)), nil
}

func TestQuery_ListTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	taskRepo := &queryTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID:         "task_001",
			Title:          "Build landing page",
			Category:       "coding",
			Status:         task.StatusOpen,
			Requester:      shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
			Reward:         task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
			LastActivityAt: &now,
		},
	}}
	e := gin.New()
	q := NewQuery(
		taskRepo,
		emptyBidRepo{},
		emptyProgressRepo{},
		emptyMilestoneRepo{},
		emptySubmissionRepo{},
		reviewRepoStub{},
		settlementRepoStub{},
		&queryAccountRepo{},
	)
	e.GET("/api/tasks", q.ListTasks)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks?page=1&pageSize=20", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Success bool `json:"success"`
		Data    []struct {
			TaskID string `json:"taskId"`
		} `json:"data"`
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Success || len(body.Data) != 1 || body.Data[0].TaskID != "task_001" || body.Meta.Total != 1 {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestQuery_GetTask_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	e := gin.New()
	q := NewQuery(
		&queryTaskRepo{items: map[string]*task.Task{}},
		emptyBidRepo{},
		emptyProgressRepo{},
		emptyMilestoneRepo{},
		emptySubmissionRepo{},
		reviewRepoStub{},
		settlementRepoStub{},
		&queryAccountRepo{},
	)
	e.GET("/api/tasks/:taskId", q.GetTask)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/missing", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestQuery_ExecutorHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	acceptedAt := time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)
	settledAt := time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
	taskRepo := &queryTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID:    "task_001",
			Title:     "Build landing page",
			Category:  "coding",
			Status:    task.StatusSettled,
			Reward:    task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
			Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
			AssignedExecutor: &shared.Actor{
				ID:   "agent_007",
				Kind: shared.ActorKindAgent,
			},
		},
	}}
	e := gin.New()
	q := NewQuery(
		taskRepo,
		emptyBidRepo{},
		emptyProgressRepo{},
		emptyMilestoneRepo{},
		emptySubmissionRepo{},
		reviewRepoStub{items: []*review.Review{{TaskID: "task_001", Decision: review.DecisionAccepted, ReviewedAt: acceptedAt}}},
		settlementRepoStub{items: []*settlement.Settlement{{TaskID: "task_001", RecordedAt: settledAt}}},
		&queryAccountRepo{},
	)
	e.GET("/api/executors/:executorId/history", q.ExecutorHistory)

	req := httptest.NewRequest(http.MethodGet, "/api/executors/agent_007/history", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body struct {
		Success bool `json:"success"`
		Data    []struct {
			TaskID     string     `json:"taskId"`
			AcceptedAt *time.Time `json:"acceptedAt"`
			SettledAt  *time.Time `json:"settledAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Success || len(body.Data) != 1 || body.Data[0].TaskID != "task_001" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
	if body.Data[0].AcceptedAt == nil || !body.Data[0].AcceptedAt.Equal(acceptedAt) {
		t.Fatalf("unexpected acceptedAt: %+v", body.Data[0].AcceptedAt)
	}
	if body.Data[0].SettledAt == nil || !body.Data[0].SettledAt.Equal(settledAt) {
		t.Fatalf("unexpected settledAt: %+v", body.Data[0].SettledAt)
	}
}
