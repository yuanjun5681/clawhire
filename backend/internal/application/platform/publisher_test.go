package platform

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/infrastructure/clawsynapse"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- fakes ---

type fakeConnRepo struct {
	byLocal map[string][]*account.PlatformConnection
}

func (r *fakeConnRepo) Insert(_ context.Context, conn *account.PlatformConnection) error {
	r.byLocal[conn.LocalUserID] = append(r.byLocal[conn.LocalUserID], conn)
	return nil
}

func (r *fakeConnRepo) FindByLocalUser(_ context.Context, localUserID, platform string) ([]*account.PlatformConnection, error) {
	conns := r.byLocal[localUserID]
	if platform == "" {
		return conns, nil
	}
	var out []*account.PlatformConnection
	for _, c := range conns {
		if c.Platform == platform {
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *fakeConnRepo) FindByRemote(_ context.Context, platformNodeID, remoteUserID string) (*account.PlatformConnection, error) {
	return nil, account.ErrConnectionNotFound
}

func (r *fakeConnRepo) DeleteByLocalUserAndNode(_ context.Context, localUserID, platformNodeID string) error {
	return account.ErrConnectionNotFound
}

type fakeSynapse struct {
	calls []clawsynapse.PublishRequest
	err   error
}

func (f *fakeSynapse) Publish(_ context.Context, req clawsynapse.PublishRequest) (*clawsynapse.PublishResult, error) {
	f.calls = append(f.calls, req)
	if f.err != nil {
		return nil, f.err
	}
	return &clawsynapse.PublishResult{TargetNode: req.TargetNode, MessageID: "msg_test"}, nil
}

func newConn(localID, remoteID, nodeID string) *account.PlatformConnection {
	return &account.PlatformConnection{
		ID:             bson.NewObjectID(),
		Platform:       "trustmesh",
		PlatformNodeID: nodeID,
		LocalUserID:    localID,
		RemoteUserID:   remoteID,
		LinkedAt:       time.Now(),
	}
}

func newPublisher(repo *fakeConnRepo, syn *fakeSynapse) *SyncPublisher {
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	return &SyncPublisher{connections: repo, synapse: syn, log: log}
}

func sampleTask() *task.Task {
	deadline := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	return &task.Task{
		TaskID:      "task_001",
		Title:       "Build API docs",
		Description: "Write OpenAPI spec",
		Category:    "writing",
		Requester:   shared.Actor{ID: "acct_alice", Kind: shared.ActorKindUser},
		Reward:      task.Reward{Mode: task.RewardModeFixed, Amount: 500, Currency: "USDC"},
		Deadline:    &deadline,
	}
}

// --- tests: NotifyTaskAwarded ---

func TestNotifyTaskAwarded_NoConnection_NothingPublished(t *testing.T) {
	repo := &fakeConnRepo{byLocal: map[string][]*account.PlatformConnection{}}
	syn := &fakeSynapse{}
	pub := newPublisher(repo, syn)

	pub.NotifyTaskAwarded(context.Background(), sampleTask(), "ctr_001", "acct_bob")

	if len(syn.calls) != 0 {
		t.Fatalf("expected no publish calls, got %d", len(syn.calls))
	}
}

func TestNotifyTaskAwarded_WithConnection_PublishesCorrectPayload(t *testing.T) {
	conn := newConn("acct_bob", "usr_xxxx", "node_trustmesh_prod")
	repo := &fakeConnRepo{byLocal: map[string][]*account.PlatformConnection{
		"acct_bob": {conn},
	}}
	syn := &fakeSynapse{}
	pub := newPublisher(repo, syn)
	t_ := sampleTask()

	pub.NotifyTaskAwarded(context.Background(), t_, "ctr_001", "acct_bob")

	if len(syn.calls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(syn.calls))
	}
	call := syn.calls[0]
	if call.Type != "clawhire.task.awarded" {
		t.Errorf("type = %q, want clawhire.task.awarded", call.Type)
	}
	if call.TargetNode != "node_trustmesh_prod" {
		t.Errorf("targetNode = %q, want node_trustmesh_prod", call.TargetNode)
	}
	if call.Metadata["clawhireAccountId"] != "acct_bob" {
		t.Errorf("metadata.clawhireAccountId = %v", call.Metadata["clawhireAccountId"])
	}
	if call.Metadata["remoteUserId"] != "usr_xxxx" {
		t.Errorf("metadata.remoteUserId = %v", call.Metadata["remoteUserId"])
	}
	var msg taskAwardedMessage
	if err := json.Unmarshal([]byte(call.Message), &msg); err != nil {
		t.Fatalf("unmarshal message: %v", err)
	}
	if msg.TaskID != "task_001" {
		t.Errorf("message.taskId = %q", msg.TaskID)
	}
	if msg.ContractID != "ctr_001" {
		t.Errorf("message.contractId = %q", msg.ContractID)
	}
	if msg.Reward.Amount != 500 || msg.Reward.Currency != "USDC" {
		t.Errorf("message.agreedReward = %+v", msg.Reward)
	}
	if msg.RequesterID != "acct_alice" {
		t.Errorf("message.requesterId = %q", msg.RequesterID)
	}
}

func TestNotifyTaskAwarded_MultipleConnections_PublishesAll(t *testing.T) {
	conn1 := newConn("acct_bob", "usr_xxxx", "node_trustmesh_prod")
	conn2 := newConn("acct_bob", "usr_yyyy", "node_trustmesh_staging")
	repo := &fakeConnRepo{byLocal: map[string][]*account.PlatformConnection{
		"acct_bob": {conn1, conn2},
	}}
	syn := &fakeSynapse{}
	pub := newPublisher(repo, syn)

	pub.NotifyTaskAwarded(context.Background(), sampleTask(), "ctr_001", "acct_bob")

	if len(syn.calls) != 2 {
		t.Fatalf("expected 2 publish calls, got %d", len(syn.calls))
	}
}

func TestNotifyTaskAwarded_SynapseError_NoReturnError(t *testing.T) {
	conn := newConn("acct_bob", "usr_xxxx", "node_trustmesh_prod")
	repo := &fakeConnRepo{byLocal: map[string][]*account.PlatformConnection{
		"acct_bob": {conn},
	}}
	syn := &fakeSynapse{err: errors.New("connection refused")}
	pub := newPublisher(repo, syn)

	// 不应 panic，不返回错误（graceful degradation）
	pub.NotifyTaskAwarded(context.Background(), sampleTask(), "ctr_001", "acct_bob")
}

// --- tests: NotifySubmissionAccepted ---

func TestNotifySubmissionAccepted_WithConnection_CorrectPayload(t *testing.T) {
	conn := newConn("acct_bob", "usr_xxxx", "node_trustmesh_prod")
	repo := &fakeConnRepo{byLocal: map[string][]*account.PlatformConnection{
		"acct_bob": {conn},
	}}
	syn := &fakeSynapse{}
	pub := newPublisher(repo, syn)
	at := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)

	pub.NotifySubmissionAccepted(context.Background(), "task_001", "sub_001", "ctr_001", "acct_bob", &at)

	if len(syn.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(syn.calls))
	}
	call := syn.calls[0]
	if call.Type != "clawhire.submission.accepted" {
		t.Errorf("type = %q", call.Type)
	}
	var msg submissionAcceptedMessage
	if err := json.Unmarshal([]byte(call.Message), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msg.TaskID != "task_001" || msg.SubmissionID != "sub_001" || msg.ContractID != "ctr_001" {
		t.Errorf("msg = %+v", msg)
	}
	if msg.AcceptedAt == nil || !msg.AcceptedAt.Equal(at) {
		t.Errorf("acceptedAt = %v, want %v", msg.AcceptedAt, at)
	}
}

func TestNotifySubmissionAccepted_NoConnection_NothingPublished(t *testing.T) {
	repo := &fakeConnRepo{byLocal: map[string][]*account.PlatformConnection{}}
	syn := &fakeSynapse{}
	pub := newPublisher(repo, syn)

	pub.NotifySubmissionAccepted(context.Background(), "task_001", "sub_001", "", "acct_bob", nil)

	if len(syn.calls) != 0 {
		t.Fatalf("expected no calls, got %d", len(syn.calls))
	}
}

// --- tests: NotifySubmissionRejected ---

func TestNotifySubmissionRejected_WithConnection_CorrectPayload(t *testing.T) {
	conn := newConn("acct_bob", "usr_xxxx", "node_trustmesh_prod")
	repo := &fakeConnRepo{byLocal: map[string][]*account.PlatformConnection{
		"acct_bob": {conn},
	}}
	syn := &fakeSynapse{}
	pub := newPublisher(repo, syn)
	at := time.Date(2026, 4, 25, 11, 0, 0, 0, time.UTC)

	pub.NotifySubmissionRejected(context.Background(), "task_001", "sub_001", "missing error codes", "acct_bob", &at)

	if len(syn.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(syn.calls))
	}
	call := syn.calls[0]
	if call.Type != "clawhire.submission.rejected" {
		t.Errorf("type = %q", call.Type)
	}
	var msg submissionRejectedMessage
	if err := json.Unmarshal([]byte(call.Message), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msg.Reason != "missing error codes" {
		t.Errorf("reason = %q", msg.Reason)
	}
}
