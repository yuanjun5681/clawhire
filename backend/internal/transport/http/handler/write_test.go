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
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
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
	item, ok := r.items[taskID]
	if !ok {
		return task.ErrTaskNotFound
	}
	if item.Status != expected {
		return task.ErrStatusConflict
	}
	item.Status = next
	item.LastActivityAt = &at
	item.UpdatedAt = at
	return nil
}

func (r *writeTaskRepo) UpdateAssignment(_ context.Context, taskID string, executor shared.Actor, contractID string, at time.Time) error {
	item, ok := r.items[taskID]
	if !ok {
		return task.ErrTaskNotFound
	}
	cp := executor
	item.AssignedExecutor = &cp
	item.CurrentContractID = contractID
	item.LastActivityAt = &at
	item.UpdatedAt = at
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

type writeContractRepo struct{ items map[string]*contract.Contract }

func newWriteContractRepo() *writeContractRepo {
	return &writeContractRepo{items: map[string]*contract.Contract{}}
}

func (r *writeContractRepo) Insert(_ context.Context, c *contract.Contract) error {
	cp := *c
	r.items[c.ContractID] = &cp
	return nil
}

func (r *writeContractRepo) FindByID(_ context.Context, contractID string) (*contract.Contract, error) {
	item, ok := r.items[contractID]
	if !ok {
		return nil, contract.ErrContractNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *writeContractRepo) FindActiveByTask(_ context.Context, taskID string) (*contract.Contract, error) {
	for _, item := range r.items {
		if item.TaskID == taskID && item.Status == contract.StatusActive {
			cp := *item
			return &cp, nil
		}
	}
	return nil, contract.ErrContractNotFound
}

func (r *writeContractRepo) UpdateStatus(_ context.Context, contractID string, status contract.Status, at time.Time) error {
	item, ok := r.items[contractID]
	if !ok {
		return contract.ErrContractNotFound
	}
	item.Status = status
	item.UpdatedAt = at
	return nil
}

func (r *writeContractRepo) MarkStarted(_ context.Context, contractID string, at time.Time) error {
	item, ok := r.items[contractID]
	if !ok {
		return contract.ErrContractNotFound
	}
	item.Status = contract.StatusActive
	item.StartedAt = &at
	item.UpdatedAt = at
	return nil
}

type writeSubmissionRepo struct {
	items map[string]*submission.Submission
}

func newWriteSubmissionRepo() *writeSubmissionRepo {
	return &writeSubmissionRepo{items: map[string]*submission.Submission{}}
}

func (r *writeSubmissionRepo) Insert(_ context.Context, s *submission.Submission) error {
	cp := *s
	r.items[s.SubmissionID] = &cp
	return nil
}

func (r *writeSubmissionRepo) FindByID(_ context.Context, submissionID string) (*submission.Submission, error) {
	item, ok := r.items[submissionID]
	if !ok {
		return nil, submission.ErrSubmissionNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *writeSubmissionRepo) UpdateStatus(_ context.Context, submissionID string, status submission.Status) error {
	item, ok := r.items[submissionID]
	if !ok {
		return submission.ErrSubmissionNotFound
	}
	item.Status = status
	return nil
}

func (r *writeSubmissionRepo) ListByTask(_ context.Context, taskID string, page, pageSize int) ([]*submission.Submission, int64, error) {
	return nil, 0, nil
}

func (r *writeSubmissionRepo) LatestByTask(_ context.Context, taskID string) (*submission.Submission, error) {
	return nil, submission.ErrSubmissionNotFound
}

type writeReviewRepo struct{ items map[string]*review.Review }

func newWriteReviewRepo() *writeReviewRepo {
	return &writeReviewRepo{items: map[string]*review.Review{}}
}

func (r *writeReviewRepo) Insert(_ context.Context, item *review.Review) error {
	cp := *item
	r.items[item.ReviewID] = &cp
	return nil
}

func (r *writeReviewRepo) ListByTask(_ context.Context, taskID string, page, pageSize int) ([]*review.Review, int64, error) {
	return nil, 0, nil
}

func (r *writeReviewRepo) ListBySubmission(_ context.Context, submissionID string) ([]*review.Review, error) {
	return nil, nil
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

// testAccountHeader 让测试通过请求头模拟已认证账号；handler 现在从 ctx 拿 accountId，
// 因此用 testAuthStub 中间件把头字段搬到 ctx 里，和真实的 Auth 中间件行为对齐。
const testAccountHeader = "X-Account-ID"

func testAuthStub() gin.HandlerFunc {
	return func(c *gin.Context) {
		if v := c.GetHeader(testAccountHeader); v != "" {
			c.Set(middleware.CtxKeyAccountID, v)
		}
		c.Next()
	}
}

func TestWrite_CreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	taskRepo := newWriteTaskRepo()
	accountRepo := &writeAccountRepo{items: map[string]*account.Account{
		"acct_human_001": {AccountID: "acct_human_001", Type: account.TypeHuman, Status: account.StatusActive, DisplayName: "Alice"},
	}}
	svc := appcmd.NewService(appcmd.Options{
		Tasks:       taskRepo,
		Bids:        newWriteBidRepo(),
		Contracts:   newWriteContractRepo(),
		Submissions: newWriteSubmissionRepo(),
		Reviews:     newWriteReviewRepo(),
		DomainEvts:  writeEventRepo{},
		Now:         func() time.Time { return time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.Use(testAuthStub())
	e.POST("/api/tasks", w.CreateTask)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString(`{"taskId":"task_001","title":"Build landing page","category":"coding","reward":{"mode":"fixed","amount":300,"currency":"USD"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_human_001")
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
		Tasks:       taskRepo,
		Bids:        bidRepo,
		Contracts:   newWriteContractRepo(),
		Submissions: newWriteSubmissionRepo(),
		Reviews:     newWriteReviewRepo(),
		DomainEvts:  writeEventRepo{},
		Now:         func() time.Time { return now.Add(time.Minute) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.Use(testAuthStub())
	e.POST("/api/tasks/:taskId/bids", w.CreateBid)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/task_001/bids", bytes.NewBufferString(`{"bidId":"bid_001","price":260,"currency":"USD","proposal":"Can deliver within 24 hours"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_human_002")
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

func TestWrite_AwardTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)
	taskRepo := newWriteTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:    "task_001",
		Status:    task.StatusOpen,
		Requester: shared.Actor{ID: "acct_human_001", Kind: shared.ActorKindUser},
		Reward:    task.Reward{Mode: task.RewardModeFixed, Amount: 300, Currency: "USD"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	contractRepo := newWriteContractRepo()
	accountRepo := &writeAccountRepo{items: map[string]*account.Account{
		"acct_human_001": {AccountID: "acct_human_001", Type: account.TypeHuman, Status: account.StatusActive, DisplayName: "Alice"},
		"acct_agent_001": {AccountID: "acct_agent_001", Type: account.TypeAgent, Status: account.StatusActive, DisplayName: "BuildBot"},
	}}
	svc := appcmd.NewService(appcmd.Options{
		Tasks:       taskRepo,
		Bids:        newWriteBidRepo(),
		Contracts:   contractRepo,
		Submissions: newWriteSubmissionRepo(),
		Reviews:     newWriteReviewRepo(),
		DomainEvts:  writeEventRepo{},
		Now:         func() time.Time { return now.Add(time.Minute) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.Use(testAuthStub())
	e.POST("/api/tasks/:taskId/award", w.AwardTask)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/task_001/award", bytes.NewBufferString(`{"contractId":"contract_001","executorId":"acct_agent_001","agreedReward":{"amount":260,"currency":"USD"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_human_001")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	gotTask, err := taskRepo.FindByID(context.Background(), "task_001")
	if err != nil {
		t.Fatalf("FindByID err = %v", err)
	}
	if gotTask.AssignedExecutor == nil || gotTask.AssignedExecutor.ID != "acct_agent_001" || gotTask.Status != task.StatusAwarded {
		t.Fatalf("task = %+v", gotTask)
	}
	gotContract, err := contractRepo.FindByID(context.Background(), "contract_001")
	if err != nil {
		t.Fatalf("FindByID contract err = %v", err)
	}
	if gotContract.Executor.ID != "acct_agent_001" || gotContract.Status != contract.StatusActive {
		t.Fatalf("contract = %+v", gotContract)
	}
}

func TestWrite_CreateSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
	taskRepo := newWriteTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:            "task_001",
		Status:            task.StatusInProgress,
		Requester:         shared.Actor{ID: "acct_human_001", Kind: shared.ActorKindUser},
		AssignedExecutor:  &shared.Actor{ID: "acct_human_002", Kind: shared.ActorKindUser},
		CurrentContractID: "contract_001",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	submissionRepo := newWriteSubmissionRepo()
	accountRepo := &writeAccountRepo{items: map[string]*account.Account{
		"acct_human_002": {AccountID: "acct_human_002", Type: account.TypeHuman, Status: account.StatusActive, DisplayName: "Bob"},
	}}
	svc := appcmd.NewService(appcmd.Options{
		Tasks:       taskRepo,
		Bids:        newWriteBidRepo(),
		Contracts:   newWriteContractRepo(),
		Submissions: submissionRepo,
		Reviews:     newWriteReviewRepo(),
		DomainEvts:  writeEventRepo{},
		Now:         func() time.Time { return now.Add(time.Minute) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.Use(testAuthStub())
	e.POST("/api/tasks/:taskId/submissions", w.CreateSubmission)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/task_001/submissions", bytes.NewBufferString(`{"submissionId":"submission_001","summary":"Delivered landing page","artifacts":[{"type":"url","value":"https://example.com/result","label":"Preview"}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_human_002")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(submissionRepo.items) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(submissionRepo.items))
	}
	var got *submission.Submission
	for _, s := range submissionRepo.items {
		got = s
		break
	}
	if got.Status != submission.StatusSubmitted {
		t.Fatalf("submission status = %s, want %s", got.Status, submission.StatusSubmitted)
	}
}

func TestWrite_AcceptSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 4, 23, 13, 0, 0, 0, time.UTC)
	taskRepo := newWriteTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:            "task_001",
		Status:            task.StatusSubmitted,
		Requester:         shared.Actor{ID: "acct_human_001", Kind: shared.ActorKindUser},
		CurrentContractID: "contract_001",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	contractRepo := newWriteContractRepo()
	contractRepo.items["contract_001"] = &contract.Contract{
		ContractID: "contract_001",
		TaskID:     "task_001",
		Status:     contract.StatusActive,
	}
	submissionRepo := newWriteSubmissionRepo()
	submissionRepo.items["submission_001"] = &submission.Submission{
		SubmissionID: "submission_001",
		TaskID:       "task_001",
		Status:       submission.StatusSubmitted,
	}
	reviewRepo := newWriteReviewRepo()
	accountRepo := &writeAccountRepo{items: map[string]*account.Account{
		"acct_human_001": {AccountID: "acct_human_001", Type: account.TypeHuman, Status: account.StatusActive, DisplayName: "Alice"},
	}}
	svc := appcmd.NewService(appcmd.Options{
		Tasks:       taskRepo,
		Bids:        newWriteBidRepo(),
		Contracts:   contractRepo,
		Submissions: submissionRepo,
		Reviews:     reviewRepo,
		DomainEvts:  writeEventRepo{},
		Now:         func() time.Time { return now.Add(time.Minute) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.Use(testAuthStub())
	e.POST("/api/tasks/:taskId/accept", w.AcceptSubmission)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/task_001/accept", bytes.NewBufferString(`{"submissionId":"submission_001"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_human_001")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	gotTask, err := taskRepo.FindByID(context.Background(), "task_001")
	if err != nil {
		t.Fatalf("FindByID err = %v", err)
	}
	if gotTask.Status != task.StatusAccepted {
		t.Fatalf("task status = %s", gotTask.Status)
	}
	gotSub, err := submissionRepo.FindByID(context.Background(), "submission_001")
	if err != nil {
		t.Fatalf("FindByID submission err = %v", err)
	}
	if gotSub.Status != submission.StatusAccepted {
		t.Fatalf("submission status = %s", gotSub.Status)
	}
	gotContract, err := contractRepo.FindByID(context.Background(), "contract_001")
	if err != nil {
		t.Fatalf("FindByID contract err = %v", err)
	}
	if gotContract.Status != contract.StatusCompleted {
		t.Fatalf("contract status = %s", gotContract.Status)
	}
	if len(reviewRepo.items) != 1 {
		t.Fatalf("reviews = %+v", reviewRepo.items)
	}
}

func TestWrite_RejectSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 4, 23, 14, 0, 0, 0, time.UTC)
	taskRepo := newWriteTaskRepo()
	taskRepo.items["task_001"] = &task.Task{
		TaskID:    "task_001",
		Status:    task.StatusSubmitted,
		Requester: shared.Actor{ID: "acct_human_001", Kind: shared.ActorKindUser},
		CreatedAt: now,
		UpdatedAt: now,
	}
	submissionRepo := newWriteSubmissionRepo()
	submissionRepo.items["submission_001"] = &submission.Submission{
		SubmissionID: "submission_001",
		TaskID:       "task_001",
		Status:       submission.StatusSubmitted,
	}
	reviewRepo := newWriteReviewRepo()
	accountRepo := &writeAccountRepo{items: map[string]*account.Account{
		"acct_human_001": {AccountID: "acct_human_001", Type: account.TypeHuman, Status: account.StatusActive, DisplayName: "Alice"},
	}}
	svc := appcmd.NewService(appcmd.Options{
		Tasks:       taskRepo,
		Bids:        newWriteBidRepo(),
		Contracts:   newWriteContractRepo(),
		Submissions: submissionRepo,
		Reviews:     reviewRepo,
		DomainEvts:  writeEventRepo{},
		Now:         func() time.Time { return now.Add(time.Minute) },
	})
	w := NewWrite(svc, accountRepo)

	e := gin.New()
	e.Use(middleware.RequestID())
	e.Use(testAuthStub())
	e.POST("/api/tasks/:taskId/reject", w.RejectSubmission)

	req := httptest.NewRequest(http.MethodPost, "/api/tasks/task_001/reject", bytes.NewBufferString(`{"submissionId":"submission_001","reason":"Missing test report"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_human_001")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	gotTask, err := taskRepo.FindByID(context.Background(), "task_001")
	if err != nil {
		t.Fatalf("FindByID err = %v", err)
	}
	if gotTask.Status != task.StatusRejected {
		t.Fatalf("task status = %s", gotTask.Status)
	}
	gotSub, err := submissionRepo.FindByID(context.Background(), "submission_001")
	if err != nil {
		t.Fatalf("FindByID submission err = %v", err)
	}
	if gotSub.Status != submission.StatusRejected {
		t.Fatalf("submission status = %s", gotSub.Status)
	}
	if len(reviewRepo.items) != 1 {
		t.Fatalf("reviews = %+v", reviewRepo.items)
	}
}
