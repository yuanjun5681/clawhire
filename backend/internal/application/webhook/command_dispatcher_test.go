package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

type fakeTaskRepo struct {
	items map[string]*task.Task
}

func newFakeTaskRepo() *fakeTaskRepo {
	return &fakeTaskRepo{items: map[string]*task.Task{}}
}

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
	var list []*task.Task
	for _, item := range r.items {
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
}

func (r *fakeTaskRepo) ListByExecutor(_ context.Context, executorID string, statuses []task.Status, page, pageSize int) ([]*task.Task, int64, error) {
	var list []*task.Task
	for _, item := range r.items {
		if item.AssignedExecutor == nil || item.AssignedExecutor.ID != executorID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
}

type fakeBidRepo struct {
	items map[string]*bid.Bid
}

func newFakeBidRepo() *fakeBidRepo {
	return &fakeBidRepo{items: map[string]*bid.Bid{}}
}

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
	var list []*bid.Bid
	for _, item := range r.items {
		if item.TaskID != taskID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
}

func (r *fakeBidRepo) ListByExecutor(_ context.Context, executorID string, page, pageSize int) ([]*bid.Bid, int64, error) {
	var list []*bid.Bid
	for _, item := range r.items {
		if item.Executor.ID != executorID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
}

func (r *fakeBidRepo) MarkAwarded(_ context.Context, bidID string) error {
	item, ok := r.items[bidID]
	if !ok {
		return bid.ErrBidNotFound
	}
	item.Status = bid.StatusAwarded
	return nil
}

func (r *fakeBidRepo) InvalidateOthers(_ context.Context, taskID string, exceptBidID string) error {
	for _, item := range r.items {
		if item.TaskID == taskID && item.BidID != exceptBidID && item.Status == bid.StatusActive {
			item.Status = bid.StatusRejected
		}
	}
	return nil
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
	item, ok := r.items[contractID]
	if !ok {
		return contract.ErrContractNotFound
	}
	item.StartedAt = &at
	item.UpdatedAt = at
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
	var list []*submission.Submission
	for _, item := range r.items {
		if item.TaskID != taskID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
}

func (r *fakeSubmissionRepo) LatestByTask(_ context.Context, taskID string) (*submission.Submission, error) {
	var latest *submission.Submission
	for _, item := range r.items {
		if item.TaskID != taskID {
			continue
		}
		if latest == nil || item.SubmittedAt.After(latest.SubmittedAt) {
			cp := *item
			latest = &cp
		}
	}
	if latest == nil {
		return nil, submission.ErrSubmissionNotFound
	}
	return latest, nil
}

type fakeReviewRepo struct {
	items map[string]*review.Review
}

func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{items: map[string]*review.Review{}}
}

func (r *fakeReviewRepo) Insert(_ context.Context, item *review.Review) error {
	cp := *item
	r.items[item.ReviewID] = &cp
	return nil
}

func (r *fakeReviewRepo) ListByTask(_ context.Context, taskID string, page, pageSize int) ([]*review.Review, int64, error) {
	var list []*review.Review
	for _, item := range r.items {
		if item.TaskID != taskID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
}

func (r *fakeReviewRepo) ListBySubmission(_ context.Context, submissionID string) ([]*review.Review, error) {
	var list []*review.Review
	for _, item := range r.items {
		if item.SubmissionID != submissionID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, nil
}

type fakeSettlementRepo struct {
	items map[string]*settlement.Settlement
}

func newFakeSettlementRepo() *fakeSettlementRepo {
	return &fakeSettlementRepo{items: map[string]*settlement.Settlement{}}
}

func (r *fakeSettlementRepo) Insert(_ context.Context, item *settlement.Settlement) error {
	cp := *item
	r.items[item.SettlementID] = &cp
	return nil
}

func (r *fakeSettlementRepo) FindByID(_ context.Context, settlementID string) (*settlement.Settlement, error) {
	item, ok := r.items[settlementID]
	if !ok {
		return nil, settlement.ErrSettlementNotFound
	}
	cp := *item
	return &cp, nil
}

func (r *fakeSettlementRepo) ListByTask(_ context.Context, taskID string) ([]*settlement.Settlement, error) {
	var list []*settlement.Settlement
	for _, item := range r.items {
		if item.TaskID != taskID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, nil
}

func (r *fakeSettlementRepo) ListByPayee(_ context.Context, payeeID string, page, pageSize int) ([]*settlement.Settlement, int64, error) {
	var list []*settlement.Settlement
	for _, item := range r.items {
		if item.Payee.ID != payeeID {
			continue
		}
		cp := *item
		list = append(list, &cp)
	}
	return list, int64(len(list)), nil
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
	var list []*event.DomainEvent
	for _, item := range r.items {
		if item.AggregateType == aggType && item.AggregateID == aggID {
			cp := *item
			list = append(list, &cp)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	return list, int64(len(list)), nil
}

func TestCommandDispatcher_MinimalLifecycle(t *testing.T) {
	taskRepo := newFakeTaskRepo()
	bidRepo := newFakeBidRepo()
	contractRepo := newFakeContractRepo()
	submissionRepo := newFakeSubmissionRepo()
	reviewRepo := newFakeReviewRepo()
	settlementRepo := newFakeSettlementRepo()
	domainEventRepo := &fakeDomainEventRepo{}
	rawRepo := newFakeRawRepo()

	now := fixedNow(time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC))
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:       taskRepo,
		Bids:        bidRepo,
		Contracts:   contractRepo,
		Submissions: submissionRepo,
		Reviews:     reviewRepo,
		Settlements: settlementRepo,
		DomainEvts:  domainEventRepo,
		Now:         now,
	})
	svc := NewService(Options{
		RawRepo:    rawRepo,
		Dispatcher: dispatcher,
		Now:        now,
	})

	events := []struct {
		eventID string
		msgType string
		data    map[string]interface{}
	}{
		{
			eventID: "evt_posted",
			msgType: "clawhire.task.posted",
			data: map[string]interface{}{
				"taskId":   "task_001",
				"title":    "Build landing page",
				"category": "coding",
				"requester": map[string]interface{}{
					"id":   "user_001",
					"kind": "user",
				},
				"reward": map[string]interface{}{
					"mode":     "fixed",
					"amount":   300,
					"currency": "USD",
				},
			},
		},
		{
			eventID: "evt_awarded",
			msgType: "clawhire.task.awarded",
			data: map[string]interface{}{
				"taskId":     "task_001",
				"contractId": "contract_001",
				"executor": map[string]interface{}{
					"id":   "agent_007",
					"kind": "agent",
				},
				"agreedReward": map[string]interface{}{
					"amount":   260,
					"currency": "USD",
				},
			},
		},
		{
			eventID: "evt_started",
			msgType: "clawhire.task.started",
			data: map[string]interface{}{
				"taskId":     "task_001",
				"contractId": "contract_001",
			},
		},
		{
			eventID: "evt_submitted",
			msgType: "clawhire.submission.created",
			data: map[string]interface{}{
				"taskId":       "task_001",
				"submissionId": "submission_001",
				"executor": map[string]interface{}{
					"id":   "agent_007",
					"kind": "agent",
				},
				"summary": "Landing page delivered",
				"artifacts": []map[string]interface{}{
					{
						"type":  "url",
						"value": "https://example.com/result/123",
					},
				},
			},
		},
		{
			eventID: "evt_accepted",
			msgType: "clawhire.submission.accepted",
			data: map[string]interface{}{
				"taskId":       "task_001",
				"submissionId": "submission_001",
				"acceptedBy": map[string]interface{}{
					"id":   "user_001",
					"kind": "user",
				},
			},
		},
		{
			eventID: "evt_settled",
			msgType: "clawhire.settlement.recorded",
			data: map[string]interface{}{
				"taskId":       "task_001",
				"contractId":   "contract_001",
				"settlementId": "settlement_001",
				"payee": map[string]interface{}{
					"id":   "agent_007",
					"kind": "agent",
				},
				"amount":   260,
				"currency": "USD",
				"status":   "recorded",
			},
		},
	}

	for _, item := range events {
		if err := ingestEnvelope(t, svc, item.eventID, item.msgType, item.data); err != nil {
			t.Fatalf("ingest %s: %v", item.msgType, err)
		}
	}

	gotTask, err := taskRepo.FindByID(context.Background(), "task_001")
	if err != nil {
		t.Fatalf("find task: %v", err)
	}
	if gotTask.Status != task.StatusSettled {
		t.Fatalf("task status = %s, want %s", gotTask.Status, task.StatusSettled)
	}
	if gotTask.CurrentContractID != "contract_001" || gotTask.AssignedExecutor == nil || gotTask.AssignedExecutor.ID != "agent_007" {
		t.Fatalf("unexpected assignment: %+v", gotTask)
	}

	gotContract, err := contractRepo.FindByID(context.Background(), "contract_001")
	if err != nil {
		t.Fatalf("find contract: %v", err)
	}
	if gotContract.StartedAt == nil {
		t.Fatalf("contract should be started")
	}
	if gotContract.Status != contract.StatusCompleted {
		t.Fatalf("contract status = %s, want %s", gotContract.Status, contract.StatusCompleted)
	}

	gotSubmission, err := submissionRepo.FindByID(context.Background(), "submission_001")
	if err != nil {
		t.Fatalf("find submission: %v", err)
	}
	if gotSubmission.Status != submission.StatusAccepted {
		t.Fatalf("submission status = %s, want %s", gotSubmission.Status, submission.StatusAccepted)
	}

	reviews, _, err := reviewRepo.ListByTask(context.Background(), "task_001", 1, 20)
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	if len(reviews) != 1 || reviews[0].Decision != review.DecisionAccepted {
		t.Fatalf("unexpected reviews: %+v", reviews)
	}

	settlements, err := settlementRepo.ListByTask(context.Background(), "task_001")
	if err != nil {
		t.Fatalf("list settlements: %v", err)
	}
	if len(settlements) != 1 || settlements[0].Status != settlement.StatusRecorded {
		t.Fatalf("unexpected settlements: %+v", settlements)
	}

	if len(domainEventRepo.items) != len(events) {
		t.Fatalf("domain event count = %d, want %d", len(domainEventRepo.items), len(events))
	}
	for _, item := range events {
		raw, err := rawRepo.FindByEventKey(context.Background(), item.eventID)
		if err != nil {
			t.Fatalf("find raw event %s: %v", item.eventID, err)
		}
		if raw == nil || raw.ProcessStatus != event.ProcessStatusSucceeded {
			t.Fatalf("raw event %s not marked succeeded: %+v", item.eventID, raw)
		}
	}
}

func TestCommandDispatcher_PostTaskUsesEnvelopeEventForDomainEvent(t *testing.T) {
	taskRepo := newFakeTaskRepo()
	domainEventRepo := &fakeDomainEventRepo{}
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:      taskRepo,
		DomainEvts: domainEventRepo,
		Now:        fixedNow(time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)),
	})
	env := &clawsynapse.Envelope{
		Type: "clawhire.task.posted",
		Message: mustJSON(map[string]interface{}{
			"taskId":   "task_001",
			"title":    "Build landing page",
			"category": "coding",
			"requester": map[string]interface{}{
				"id":   "user_001",
				"kind": "user",
			},
			"reward": map[string]interface{}{
				"mode":     "fixed",
				"amount":   300,
				"currency": "USD",
			},
		}),
		Metadata: map[string]interface{}{"eventId": "evt_post_defaults"},
	}
	if _, err := dispatcher.Dispatch(context.Background(), env); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if _, err := taskRepo.FindByID(context.Background(), "task_001"); err != nil {
		t.Fatalf("find task: %v", err)
	}
	if len(domainEventRepo.items) != 1 {
		t.Fatalf("domain event count = %d, want 1", len(domainEventRepo.items))
	}
	got := domainEventRepo.items[0]
	if got.EventID != "evt_post_defaults" || got.EventType != "clawhire.task.posted" || got.AggregateID != "task_001" {
		t.Fatalf("unexpected domain event: %+v", got)
	}
}

func ingestEnvelope(t *testing.T, svc *Service, eventID, msgType string, data map[string]interface{}) error {
	t.Helper()

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	env := &clawsynapse.Envelope{
		NodeID:     "synapse-node-a",
		Type:       msgType,
		From:       "agent://sender",
		SessionKey: "sess-1",
		Message:    string(payload),
		Metadata: map[string]interface{}{
			"eventId": eventID,
			"taskId":  taskIDFromData(data),
		},
	}
	rawBody, err := json.Marshal(env)
	if err != nil {
		return err
	}
	_, err = svc.Ingest(context.Background(), env, rawBody, map[string]string{"Content-Type": "application/json"})
	return err
}

func taskIDFromData(data map[string]interface{}) string {
	v, _ := data["taskId"].(string)
	return v
}

func TestCommandDispatcher_UnknownKnownTypeGuard(t *testing.T) {
	svc := NewService(Options{
		RawRepo:    newFakeRawRepo(),
		Dispatcher: NewCommandDispatcher(CommandDispatcherOptions{}),
		Now:        fixedNow(time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)),
	})
	env := &clawsynapse.Envelope{
		Type:       "clawhire.unknown.action",
		SessionKey: "sess-x",
		Message:    "{}",
		Metadata:   map[string]interface{}{"eventId": "evt_unknown"},
	}
	_, err := svc.Ingest(context.Background(), env, []byte(`{}`), nil)
	if err == nil {
		t.Fatal("expected unsupported message type error")
	}
	ae, ok := apierr.As(err)
	if !ok || ae.Code != apierr.CodeUnsupportedMessageType {
		t.Fatalf("expected unsupported message type, got %v", err)
	}
}

func TestCommandDispatcher_AwardTaskUsesEnvelopeEventForDomainEvent(t *testing.T) {
	taskRepo := &fakeTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID:    "task_001",
			Status:    task.StatusOpen,
			Requester: shared.Actor{ID: "user_001", Kind: shared.ActorKindUser},
		},
	}}
	bidRepo := newFakeBidRepo()
	contractRepo := newFakeContractRepo()
	domainEventRepo := &fakeDomainEventRepo{}
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:      taskRepo,
		Bids:       bidRepo,
		Contracts:  contractRepo,
		DomainEvts: domainEventRepo,
		Now:        fixedNow(time.Date(2026, 4, 23, 13, 30, 0, 0, time.UTC)),
	})
	env := &clawsynapse.Envelope{
		Type: "clawhire.task.awarded",
		Message: mustJSON(map[string]interface{}{
			"taskId":     "task_001",
			"contractId": "contract_002",
			"executor": map[string]interface{}{
				"id":   "agent_007",
				"kind": "agent",
			},
			"agreedReward": map[string]interface{}{
				"amount":   260,
				"currency": "USD",
			},
		}),
		Metadata: map[string]interface{}{"eventId": "evt_award"},
	}
	if _, err := dispatcher.Dispatch(context.Background(), env); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if _, err := contractRepo.FindByID(context.Background(), "contract_002"); err != nil {
		t.Fatalf("find contract: %v", err)
	}
	if len(domainEventRepo.items) != 1 {
		t.Fatalf("domain event count = %d, want 1", len(domainEventRepo.items))
	}
	got := domainEventRepo.items[0]
	if got.EventID != "evt_award" || got.EventType != "clawhire.task.awarded" || got.AggregateID != "task_001" {
		t.Fatalf("unexpected domain event: %+v", got)
	}
}

func TestCommandDispatcher_BidPlacedUsesEnvelopeEventForDomainEvent(t *testing.T) {
	taskRepo := &fakeTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID: "task_001",
			Status: task.StatusOpen,
		},
	}}
	bidRepo := newFakeBidRepo()
	domainEventRepo := &fakeDomainEventRepo{}
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:      taskRepo,
		Bids:       bidRepo,
		DomainEvts: domainEventRepo,
		Now:        fixedNow(time.Date(2026, 4, 23, 14, 0, 0, 0, time.UTC)),
	})
	env := &clawsynapse.Envelope{
		Type: "clawhire.bid.placed",
		Message: mustJSON(map[string]interface{}{
			"taskId":   "task_001",
			"bidId":    "bid_001",
			"price":    100,
			"currency": "USD",
			"executor": map[string]interface{}{
				"id":   "agent_007",
				"kind": "agent",
			},
		}),
		Metadata: map[string]interface{}{"eventId": "evt_bid_invalid"},
	}
	if _, err := dispatcher.Dispatch(context.Background(), env); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if _, err := bidRepo.FindByID(context.Background(), "bid_001"); err != nil {
		t.Fatalf("find bid: %v", err)
	}
	if len(domainEventRepo.items) != 1 {
		t.Fatalf("domain event count = %d, want 1", len(domainEventRepo.items))
	}
	got := domainEventRepo.items[0]
	if got.EventID != "evt_bid_invalid" || got.EventType != "clawhire.bid.placed" || got.AggregateID != "task_001" {
		t.Fatalf("unexpected domain event: %+v", got)
	}
}

func TestCommandDispatcher_StartTaskFromRejectedState(t *testing.T) {
	now := fixedNow(time.Date(2026, 4, 23, 15, 0, 0, 0, time.UTC))
	taskRepo := &fakeTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID:            "task_001",
			Status:            task.StatusRejected,
			CurrentContractID: "contract_001",
		},
	}}
	contractRepo := newFakeContractRepo()
	contractRepo.items["contract_001"] = &contract.Contract{
		ContractID: "contract_001",
		TaskID:     "task_001",
		Status:     contract.StatusActive,
	}
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:     taskRepo,
		Contracts: contractRepo,
		Now:       now,
	})
	startEnv := &clawsynapse.Envelope{
		Type: "clawhire.task.started",
		Message: mustJSON(map[string]interface{}{
			"taskId":     "task_001",
			"contractId": "contract_001",
		}),
		Metadata: map[string]interface{}{"eventId": "evt_restart"},
	}
	if _, err := dispatcher.Dispatch(context.Background(), startEnv); err != nil {
		t.Fatalf("restart: %v", err)
	}
	gotTask, _ := taskRepo.FindByID(context.Background(), "task_001")
	if gotTask.Status != task.StatusInProgress {
		t.Fatalf("task status after restart = %s, want %s", gotTask.Status, task.StatusInProgress)
	}
}

func TestCommandDispatcher_CancelAwardedTaskCancelsContract(t *testing.T) {
	now := fixedNow(time.Date(2026, 4, 23, 16, 0, 0, 0, time.UTC))
	taskRepo := &fakeTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID:            "task_001",
			Status:            task.StatusAwarded,
			CurrentContractID: "contract_001",
		},
	}}
	contractRepo := newFakeContractRepo()
	contractRepo.items["contract_001"] = &contract.Contract{
		ContractID: "contract_001",
		TaskID:     "task_001",
		Status:     contract.StatusActive,
	}
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:     taskRepo,
		Contracts: contractRepo,
		Now:       now,
	})
	env := &clawsynapse.Envelope{
		Type: "clawhire.task.cancelled",
		Message: mustJSON(map[string]interface{}{
			"taskId": "task_001",
			"reason": "request withdrawn",
		}),
		Metadata: map[string]interface{}{"eventId": "evt_cancel"},
	}
	if _, err := dispatcher.Dispatch(context.Background(), env); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	gotTask, _ := taskRepo.FindByID(context.Background(), "task_001")
	if gotTask.Status != task.StatusCancelled {
		t.Fatalf("task status = %s, want %s", gotTask.Status, task.StatusCancelled)
	}
	gotContract, _ := contractRepo.FindByID(context.Background(), "contract_001")
	if gotContract.Status != contract.StatusCancelled {
		t.Fatalf("contract status = %s, want %s", gotContract.Status, contract.StatusCancelled)
	}
}

func TestCommandDispatcher_DisputeInProgressTaskDisputesContract(t *testing.T) {
	now := fixedNow(time.Date(2026, 4, 23, 17, 0, 0, 0, time.UTC))
	taskRepo := &fakeTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID:            "task_001",
			Status:            task.StatusInProgress,
			CurrentContractID: "contract_001",
		},
	}}
	contractRepo := newFakeContractRepo()
	contractRepo.items["contract_001"] = &contract.Contract{
		ContractID: "contract_001",
		TaskID:     "task_001",
		Status:     contract.StatusActive,
	}
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:     taskRepo,
		Contracts: contractRepo,
		Now:       now,
	})
	env := &clawsynapse.Envelope{
		Type: "clawhire.task.disputed",
		Message: mustJSON(map[string]interface{}{
			"taskId": "task_001",
			"reason": "quality issue",
		}),
		Metadata: map[string]interface{}{"eventId": "evt_dispute"},
	}
	if _, err := dispatcher.Dispatch(context.Background(), env); err != nil {
		t.Fatalf("dispute: %v", err)
	}
	gotTask, _ := taskRepo.FindByID(context.Background(), "task_001")
	if gotTask.Status != task.StatusDisputed {
		t.Fatalf("task status = %s, want %s", gotTask.Status, task.StatusDisputed)
	}
	gotContract, _ := contractRepo.FindByID(context.Background(), "contract_001")
	if gotContract.Status != contract.StatusDisputed {
		t.Fatalf("contract status = %s, want %s", gotContract.Status, contract.StatusDisputed)
	}
}

func TestCommandDispatcher_InvalidSettlementStatusFails(t *testing.T) {
	now := fixedNow(time.Date(2026, 4, 23, 18, 0, 0, 0, time.UTC))
	taskRepo := &fakeTaskRepo{items: map[string]*task.Task{
		"task_001": {
			TaskID:            "task_001",
			Status:            task.StatusAccepted,
			CurrentContractID: "contract_001",
		},
	}}
	dispatcher := NewCommandDispatcher(CommandDispatcherOptions{
		Tasks:       taskRepo,
		Settlements: newFakeSettlementRepo(),
		Now:         now,
	})
	env := &clawsynapse.Envelope{
		Type: "clawhire.settlement.recorded",
		Message: mustJSON(map[string]interface{}{
			"taskId":       "task_001",
			"settlementId": "settlement_001",
			"amount":       260,
			"currency":     "USD",
			"status":       "weird_status",
			"payee": map[string]interface{}{
				"id":   "agent_007",
				"kind": "agent",
			},
		}),
		Metadata: map[string]interface{}{"eventId": "evt_invalid_settlement"},
	}
	_, err := dispatcher.Dispatch(context.Background(), env)
	ae, ok := apierr.As(err)
	if err == nil || !ok || ae.Code != apierr.CodeInvalidMessagePayload {
		t.Fatalf("expected invalid payload, got %v", err)
	}
}

func mustJSON(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustJSON: %v", err))
	}
	return string(raw)
}
