// Package clawsynapse 描述 ClawSynapse 节点推送的 Webhook 结构。
//
// ClawHire 依赖 ClawSynapse 作为消息传输层：
//
//	Synapse 节点 -> POST /webhooks/clawsynapse -> ClawHire Webhook Adapter
//
// 该包只关心"外层信封"，内层 message 的反序列化由 clawhire 协议包负责。
package clawsynapse

import (
	"encoding/json"
	"strings"
)

// Source 标识原始事件来源系统（raw_events.source）。
const Source = "clawsynapse"

// Envelope 是 ClawSynapse Webhook 请求的标准载荷。
//
//	{
//	  "nodeId": "...",
//	  "type": "clawhire.task.posted",
//	  "from": "agent://requester-001",
//	  "sessionKey": "...",
//	  "message": "{...json string...}",
//	  "metadata": { "domain":"clawhire", "schemaVersion":"v1", "taskId":"task_001" }
//	}
//
// `message` 字段约定为 JSON 字符串，便于 Synapse 节点透传任意业务数据。
type Envelope struct {
	NodeID     string                 `json:"nodeId"`
	Type       string                 `json:"type"`
	From       string                 `json:"from,omitempty"`
	SessionKey string                 `json:"sessionKey,omitempty"`
	Message    string                 `json:"message"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// MetaString 返回 metadata 中的字符串字段（不存在或非字符串返回空）。
func (e *Envelope) MetaString(key string) string {
	if e == nil || e.Metadata == nil {
		return ""
	}
	v, ok := e.Metadata[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

// DecodeMessage 把 Envelope.Message 反序列化成给定结构体。
// Synapse 约定 message 为 JSON 字符串，这里做一次性解包。
func (e *Envelope) DecodeMessage(out interface{}) error {
	if e == nil {
		return errEmptyEnvelope
	}
	s := strings.TrimSpace(e.Message)
	if s == "" {
		return errEmptyMessage
	}
	return json.Unmarshal([]byte(s), out)
}
