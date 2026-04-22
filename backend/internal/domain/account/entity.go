package account

import "time"

type Type string

const (
	TypeHuman Type = "human"
	TypeAgent Type = "agent"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
	StatusPending  Status = "pending"
)

// Account 对应 MongoDB 集合 accounts。human 与 agent 复用同一结构。
type Account struct {
	AccountID      string                 `bson:"accountId"                  json:"accountId"`
	Type           Type                   `bson:"type"                       json:"type"`
	DisplayName    string                 `bson:"displayName"                json:"displayName"`
	Status         Status                 `bson:"status"                     json:"status"`
	NodeID         string                 `bson:"nodeId,omitempty"           json:"nodeId,omitempty"`
	OwnerAccountID string                 `bson:"ownerAccountId,omitempty"   json:"ownerAccountId,omitempty"`
	Profile        map[string]interface{} `bson:"profile,omitempty"          json:"profile,omitempty"`
	CreatedAt      time.Time              `bson:"createdAt"                  json:"createdAt"`
	UpdatedAt      time.Time              `bson:"updatedAt"                  json:"updatedAt"`
}
