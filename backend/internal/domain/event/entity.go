package event

import "time"

type ProcessStatus string

const (
	ProcessStatusPending   ProcessStatus = "pending"
	ProcessStatusSucceeded ProcessStatus = "succeeded"
	ProcessStatusFailed    ProcessStatus = "failed"
	ProcessStatusSkipped   ProcessStatus = "skipped"
)

// RawEvent 对应 MongoDB 集合 raw_events。
// 存档 Webhook 原始载荷与处理状态，用于审计和回放。
type RawEvent struct {
	EventKey      string                 `bson:"eventKey"              json:"eventKey"`
	Source        string                 `bson:"source"                json:"source"`
	MessageType   string                 `bson:"messageType"           json:"messageType"`
	Payload       map[string]interface{} `bson:"payload"               json:"payload"`
	Headers       map[string]string      `bson:"headers,omitempty"     json:"headers,omitempty"`
	ReceivedAt    time.Time              `bson:"receivedAt"            json:"receivedAt"`
	ProcessedAt   *time.Time             `bson:"processedAt,omitempty" json:"processedAt,omitempty"`
	ProcessStatus ProcessStatus          `bson:"processStatus"         json:"processStatus"`
	ErrorMessage  string                 `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
}

// DomainEvent 对应 MongoDB 集合 domain_events。
// 标准化的业务事件，便于审计与聚合回放。
type DomainEvent struct {
	EventID       string                 `bson:"eventId"       json:"eventId"`
	AggregateType string                 `bson:"aggregateType" json:"aggregateType"`
	AggregateID   string                 `bson:"aggregateId"   json:"aggregateId"`
	EventType     string                 `bson:"eventType"     json:"eventType"`
	Data          map[string]interface{} `bson:"data"          json:"data"`
	CreatedAt     time.Time              `bson:"createdAt"     json:"createdAt"`
}
