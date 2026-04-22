package webhook

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

type fakeRawRepo struct {
	mu    sync.Mutex
	items map[string]*event.RawEvent
}

func newFakeRawRepo() *fakeRawRepo {
	return &fakeRawRepo{items: make(map[string]*event.RawEvent)}
}

func (r *fakeRawRepo) Insert(_ context.Context, e *event.RawEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[e.EventKey]; ok {
		return event.ErrDuplicateEvent
	}
	cp := *e
	r.items[e.EventKey] = &cp
	return nil
}

func (r *fakeRawRepo) FindByEventKey(_ context.Context, k string) (*event.RawEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.items[k]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *e
	return &cp, nil
}

func (r *fakeRawRepo) MarkProcessed(_ context.Context, k string, st event.ProcessStatus, at time.Time, msg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.items[k]
	if !ok {
		return errors.New("not found")
	}
	e.ProcessStatus = st
	at2 := at
	e.ProcessedAt = &at2
	e.ErrorMessage = msg
	return nil
}

type fixedDispatcher struct {
	status event.ProcessStatus
	err    error
}

func (d fixedDispatcher) Dispatch(_ context.Context, _ *clawsynapse.Envelope) (event.ProcessStatus, error) {
	return d.status, d.err
}

func fixedNow(t time.Time) Now { return func() time.Time { return t } }

func TestIngest_RejectsNonClawHireType(t *testing.T) {
	svc := NewService(Options{RawRepo: newFakeRawRepo(), Dispatcher: NoopDispatcher{}})
	_, err := svc.Ingest(context.Background(), &clawsynapse.Envelope{
		Type:    "other.task.posted",
		Message: "{}",
	}, []byte("{}"), nil)
	ae, ok := apierr.As(err)
	if !ok || ae.Code != apierr.CodeUnsupportedMessageType {
		t.Fatalf("expected UNSUPPORTED_MESSAGE_TYPE, got %v", err)
	}
}

func TestIngest_Success_WritesRawAndMarksProcessed(t *testing.T) {
	repo := newFakeRawRepo()
	at := time.Date(2026, 4, 22, 1, 2, 3, 0, time.UTC)
	svc := NewService(Options{
		RawRepo:    repo,
		Dispatcher: fixedDispatcher{status: event.ProcessStatusSucceeded},
		Now:        fixedNow(at),
	})

	env := &clawsynapse.Envelope{
		Type:       "clawhire.task.posted",
		SessionKey: "sess-1",
		Message:    `{"taskId":"task_001"}`,
		Metadata:   map[string]interface{}{"eventId": "evt_001"},
	}
	res, err := svc.Ingest(context.Background(), env, []byte(`{"x":1}`), map[string]string{"X-Request-ID": "r1"})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if !res.Accepted || res.Duplicate || res.EventKey != "evt_001" || res.Status != event.ProcessStatusSucceeded {
		t.Fatalf("unexpected result: %+v", res)
	}

	got, _ := repo.FindByEventKey(context.Background(), "evt_001")
	if got == nil || got.ProcessStatus != event.ProcessStatusSucceeded {
		t.Fatalf("expected persisted record with succeeded status, got %+v", got)
	}
	if got.Headers["X-Request-ID"] != "r1" {
		t.Fatalf("expected header archived, got %+v", got.Headers)
	}
	if got.ProcessedAt == nil || !got.ProcessedAt.Equal(at) {
		t.Fatalf("expected processedAt = %v, got %v", at, got.ProcessedAt)
	}
}

func TestIngest_Duplicate_ReturnsDuplicateError(t *testing.T) {
	repo := newFakeRawRepo()
	svc := NewService(Options{RawRepo: repo, Dispatcher: NoopDispatcher{}})
	env := &clawsynapse.Envelope{
		Type:       "clawhire.task.posted",
		SessionKey: "sess-1",
		Message:    `{"taskId":"task_001"}`,
		Metadata:   map[string]interface{}{"eventId": "evt_dup"},
	}
	if _, err := svc.Ingest(context.Background(), env, []byte(`{}`), nil); err != nil {
		t.Fatalf("first ingest: %v", err)
	}

	res, err := svc.Ingest(context.Background(), env, []byte(`{}`), nil)
	ae, ok := apierr.As(err)
	if !ok || ae.Code != apierr.CodeDuplicateEvent {
		t.Fatalf("expected DUPLICATE_EVENT, got %v", err)
	}
	if res == nil || !res.Duplicate || res.EventKey != "evt_dup" {
		t.Fatalf("expected duplicate result, got %+v", res)
	}
}

func TestIngest_DispatcherError_MarksFailed(t *testing.T) {
	repo := newFakeRawRepo()
	svc := NewService(Options{
		RawRepo:    repo,
		Dispatcher: fixedDispatcher{err: errors.New("boom")},
	})
	env := &clawsynapse.Envelope{
		Type:       "clawhire.task.posted",
		SessionKey: "sess-2",
		Message:    `{"taskId":"task_002"}`,
		Metadata:   map[string]interface{}{"eventId": "evt_fail"},
	}
	res, err := svc.Ingest(context.Background(), env, []byte(`{}`), nil)
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if res.Status != event.ProcessStatusFailed {
		t.Fatalf("expected failed, got %s", res.Status)
	}
	got, _ := repo.FindByEventKey(context.Background(), "evt_fail")
	if got.ProcessStatus != event.ProcessStatusFailed || got.ErrorMessage != "boom" {
		t.Fatalf("expected persisted failure record, got %+v", got)
	}
}

func TestDeriveEventKey_Priorities(t *testing.T) {
	// 1) metadata.eventId 优先
	env1 := &clawsynapse.Envelope{
		Type: "clawhire.task.posted", SessionKey: "s", Message: "{}",
		Metadata: map[string]interface{}{"eventId": "E1", "taskId": "T1"},
	}
	if got := DeriveEventKey(env1); got != "E1" {
		t.Fatalf("expect E1, got %s", got)
	}
	// 2) session+type+taskId
	env2 := &clawsynapse.Envelope{
		Type: "clawhire.task.posted", SessionKey: "s", Message: "{}",
		Metadata: map[string]interface{}{"taskId": "T2"},
	}
	if got := DeriveEventKey(env2); got != "cs:s:clawhire.task.posted:T2" {
		t.Fatalf("unexpected key: %s", got)
	}
	// 3) hash fallback，形如 "h:<hex>"
	env3 := &clawsynapse.Envelope{Type: "clawhire.task.posted", Message: "{}"}
	got3 := DeriveEventKey(env3)
	if len(got3) < 3 || got3[:2] != "h:" {
		t.Fatalf("expected hash fallback, got %s", got3)
	}
	// 同样输入产生相同 hash
	if DeriveEventKey(env3) != got3 {
		t.Fatalf("hash fallback must be deterministic")
	}
}
