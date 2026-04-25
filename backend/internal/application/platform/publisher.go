// Package platform 处理 ClawHire 向外部平台（TrustMesh 等）同步任务事件的逻辑。
package platform

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/infrastructure/clawsynapse"
)

// synapseClient 是对 ClawSynapse 节点发布能力的最小接口，便于测试 mock。
type synapseClient interface {
	Publish(ctx context.Context, req clawsynapse.PublishRequest) (*clawsynapse.PublishResult, error)
}

// SyncPublisher 将 ClawHire 任务事件同步给外部平台。
// 若执行方未绑定平台账号，静默跳过，不阻塞主业务。
type SyncPublisher struct {
	connections account.PlatformConnectionRepository
	synapse     synapseClient
	log         *logrus.Logger
}

func NewSyncPublisher(
	connections account.PlatformConnectionRepository,
	synapse *clawsynapse.Client,
	log *logrus.Logger,
) *SyncPublisher {
	return &SyncPublisher{
		connections: connections,
		synapse:     synapse,
		log:         log,
	}
}

// --- 事件 payload 定义 ---

type taskAwardedMessage struct {
	TaskID       string     `json:"taskId"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	Category     string     `json:"category"`
	ContractID   string     `json:"contractId"`
	Reward       taskReward `json:"agreedReward"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	RequesterID  string     `json:"requesterId"`
}

type taskReward struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type submissionAcceptedMessage struct {
	TaskID       string     `json:"taskId"`
	SubmissionID string     `json:"submissionId"`
	ContractID   string     `json:"contractId,omitempty"`
	AcceptedAt   *time.Time `json:"acceptedAt,omitempty"`
}

type submissionRejectedMessage struct {
	TaskID       string     `json:"taskId"`
	SubmissionID string     `json:"submissionId"`
	Reason       string     `json:"reason,omitempty"`
	RejectedAt   *time.Time `json:"rejectedAt,omitempty"`
}

// --- 公开方法 ---

// NotifyTaskAwarded 通知执行方（TrustMesh）任务已被指派。
func (p *SyncPublisher) NotifyTaskAwarded(ctx context.Context, t *task.Task, contractID, executorLocalID string) {
	conns, err := p.connections.FindByLocalUser(ctx, executorLocalID, "")
	if err != nil || len(conns) == 0 {
		return
	}
	msg := taskAwardedMessage{
		TaskID:      t.TaskID,
		Title:       t.Title,
		Description: t.Description,
		Category:    t.Category,
		ContractID:  contractID,
		Reward:      taskReward{Amount: t.Reward.Amount, Currency: t.Reward.Currency},
		Deadline:    t.Deadline,
		RequesterID: t.Requester.ID,
	}
	for _, conn := range conns {
		p.publish(ctx, conn, "clawhire.task.awarded", executorLocalID, conn.RemoteUserID, msg)
	}
}

// NotifySubmissionAccepted 通知执行方提交物已被验收通过。
func (p *SyncPublisher) NotifySubmissionAccepted(ctx context.Context, taskID, submissionID, contractID, executorLocalID string, acceptedAt *time.Time) {
	conns, err := p.connections.FindByLocalUser(ctx, executorLocalID, "")
	if err != nil || len(conns) == 0 {
		return
	}
	msg := submissionAcceptedMessage{
		TaskID:       taskID,
		SubmissionID: submissionID,
		ContractID:   contractID,
		AcceptedAt:   acceptedAt,
	}
	for _, conn := range conns {
		p.publish(ctx, conn, "clawhire.submission.accepted", executorLocalID, conn.RemoteUserID, msg)
	}
}

// NotifySubmissionRejected 通知执行方提交物被驳回。
func (p *SyncPublisher) NotifySubmissionRejected(ctx context.Context, taskID, submissionID, reason, executorLocalID string, rejectedAt *time.Time) {
	conns, err := p.connections.FindByLocalUser(ctx, executorLocalID, "")
	if err != nil || len(conns) == 0 {
		return
	}
	msg := submissionRejectedMessage{
		TaskID:       taskID,
		SubmissionID: submissionID,
		Reason:       reason,
		RejectedAt:   rejectedAt,
	}
	for _, conn := range conns {
		p.publish(ctx, conn, "clawhire.submission.rejected", executorLocalID, conn.RemoteUserID, msg)
	}
}

// --- 内部发布 ---

func (p *SyncPublisher) publish(
	ctx context.Context,
	conn *account.PlatformConnection,
	eventType string,
	clawhireAccountID string,
	remoteUserID string,
	payload interface{},
) {
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		p.log.WithError(err).Errorf("sync publisher: marshal %s payload", eventType)
		return
	}
	req := clawsynapse.PublishRequest{
		TargetNode: conn.PlatformNodeID,
		Type:       eventType,
		Message:    string(msgBytes),
		Metadata: map[string]interface{}{
			"clawhireAccountId": clawhireAccountID,
			"remoteUserId":      remoteUserID,
			"platform":          conn.Platform,
		},
	}
	if _, err := p.synapse.Publish(ctx, req); err != nil {
		p.log.WithError(err).Errorf("sync publisher: publish %s to node %s", eventType, conn.PlatformNodeID)
		return
	}
	p.log.WithFields(map[string]interface{}{
		"eventType":      eventType,
		"targetNode":     conn.PlatformNodeID,
		"remoteUserId":   remoteUserID,
	}).Info("sync publisher: event published")
}
