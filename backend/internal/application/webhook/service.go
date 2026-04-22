// Package webhook 实现 Webhook 入站用例：
//
//	签名校验 → 解析 → 生成 eventKey → raw_events 幂等落库 → 分派 → 更新状态
//
// 签名校验本身由 transport 层完成（因为需要拿到原始字节），
// 本层接收的是已经校验通过的 Envelope 以及原始 payload 用于审计存档。
package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

// Now 便于测试注入时间；生产使用 time.Now。
type Now func() time.Time

// Service 处理已经完成签名校验的 Webhook 请求。
type Service struct {
	rawRepo    event.RawEventRepository
	dispatcher Dispatcher
	now        Now
}

// Options 组装 Service 依赖。
type Options struct {
	RawRepo    event.RawEventRepository
	Dispatcher Dispatcher
	Now        Now
}

// NewService 创建 Webhook Service。
func NewService(opt Options) *Service {
	now := opt.Now
	if now == nil {
		now = time.Now
	}
	d := opt.Dispatcher
	if d == nil {
		d = NoopDispatcher{}
	}
	return &Service{
		rawRepo:    opt.RawRepo,
		dispatcher: d,
		now:        now,
	}
}

// Result 返回给 transport 层用于构造 HTTP 响应。
type Result struct {
	Accepted    bool                `json:"accepted"`
	EventKey    string              `json:"eventKey"`
	MessageType string              `json:"messageType"`
	Status      event.ProcessStatus `json:"status"`
	Duplicate   bool                `json:"duplicate,omitempty"`
}

// Ingest 处理一次 Webhook 入站请求。
//
// - 校验 type 是否属于 clawhire.*
// - 生成 eventKey，原始事件落库（已存在返回 DUPLICATE_EVENT）
// - 调用 dispatcher，成功/失败/跳过分别记录处理状态
//
// headers 会被原样存档到 raw_events.headers，用于审计。
func (s *Service) Ingest(ctx context.Context, env *clawsynapse.Envelope, rawBody []byte, headers map[string]string) (*Result, error) {
	if env == nil {
		return nil, apierr.New(apierr.CodeInvalidRequest, "empty envelope")
	}
	if !clawhire.IsClawHireType(env.Type) {
		return nil, apierr.New(apierr.CodeUnsupportedMessageType, "only clawhire.* messages are accepted")
	}
	if !clawhire.IsKnown(env.Type) {
		return nil, apierr.New(apierr.CodeUnsupportedMessageType, "unsupported clawhire message type")
	}

	eventKey := DeriveEventKey(env)
	if eventKey == "" {
		return nil, apierr.New(apierr.CodeInvalidRequest, "failed to derive event key")
	}

	now := s.now().UTC()
	payload := decodePayload(rawBody, env)

	raw := &event.RawEvent{
		EventKey:      eventKey,
		Source:        clawsynapse.Source,
		MessageType:   env.Type,
		Payload:       payload,
		Headers:       headers,
		ReceivedAt:    now,
		ProcessStatus: event.ProcessStatusPending,
	}

	if err := s.rawRepo.Insert(ctx, raw); err != nil {
		if errors.Is(err, event.ErrDuplicateEvent) {
			// 幂等命中：返回 409 + 原处理状态，便于发送方判断是否需要重试。
			prev, findErr := s.rawRepo.FindByEventKey(ctx, eventKey)
			status := event.ProcessStatusPending
			if findErr == nil && prev != nil {
				status = prev.ProcessStatus
			}
			return &Result{
					Accepted:    false,
					EventKey:    eventKey,
					MessageType: env.Type,
					Status:      status,
					Duplicate:   true,
				},
				apierr.New(apierr.CodeDuplicateEvent, "event already accepted")
		}
		return nil, apierr.Wrap(apierr.CodeInternalError, "persist raw event", err)
	}

	// 分派到具体处理器；失败不影响 raw_events 已落库的事实，仅记录 errorMessage。
	status, dispErr := s.dispatcher.Dispatch(ctx, env)
	processedAt := s.now().UTC()
	errMsg := ""
	if dispErr != nil {
		status = event.ProcessStatusFailed
		errMsg = dispErr.Error()
	}
	if status == "" {
		status = event.ProcessStatusSucceeded
	}

	if err := s.rawRepo.MarkProcessed(ctx, eventKey, status, processedAt, errMsg); err != nil {
		// 处理状态写失败不阻塞入站协议，但需要告警。
		return &Result{
			Accepted:    true,
			EventKey:    eventKey,
			MessageType: env.Type,
			Status:      status,
		}, apierr.Wrap(apierr.CodeInternalError, "mark processed", err)
	}

	return &Result{
		Accepted:    true,
		EventKey:    eventKey,
		MessageType: env.Type,
		Status:      status,
	}, nil
}

// decodePayload 尝试把原始 body 解析成 map 存档；失败则退化成 {"_raw": body}。
func decodePayload(rawBody []byte, env *clawsynapse.Envelope) map[string]interface{} {
	if len(rawBody) > 0 {
		var m map[string]interface{}
		if err := json.Unmarshal(rawBody, &m); err == nil {
			return m
		}
	}
	// fallback：直接把 envelope 字段摊平存进去。
	return map[string]interface{}{
		"nodeId":     env.NodeID,
		"type":       env.Type,
		"from":       env.From,
		"sessionKey": env.SessionKey,
		"message":    env.Message,
		"metadata":   env.Metadata,
	}
}
