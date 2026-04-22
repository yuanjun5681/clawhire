package task

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type RewardMode string

const (
	RewardModeFixed     RewardMode = "fixed"
	RewardModeBid       RewardMode = "bid"
	RewardModeMilestone RewardMode = "milestone"
)

type Reward struct {
	Mode     RewardMode `bson:"mode"     json:"mode"`
	Amount   float64    `bson:"amount"   json:"amount"`
	Currency string     `bson:"currency" json:"currency"`
}

type AcceptanceMode string

const (
	AcceptanceModeManual AcceptanceMode = "manual"
	AcceptanceModeSchema AcceptanceMode = "schema"
	AcceptanceModeTest   AcceptanceMode = "test"
	AcceptanceModeHybrid AcceptanceMode = "hybrid"
)

type AcceptanceSpec struct {
	Mode  AcceptanceMode `bson:"mode"            json:"mode"`
	Rules []string       `bson:"rules,omitempty" json:"rules,omitempty"`
}

type SettlementTerms struct {
	Trigger string `bson:"trigger" json:"trigger"`
}

// Task 是任务主聚合根，对应 MongoDB 集合 tasks。
type Task struct {
	TaskID            string           `bson:"taskId"                         json:"taskId"`
	Title             string           `bson:"title"                          json:"title"`
	Description       string           `bson:"description,omitempty"          json:"description,omitempty"`
	Category          string           `bson:"category"                       json:"category"`
	Status            Status           `bson:"status"                         json:"status"`
	Requester         shared.Actor     `bson:"requester"                      json:"requester"`
	Reviewer          *shared.Actor    `bson:"reviewer,omitempty"             json:"reviewer,omitempty"`
	Reward            Reward           `bson:"reward"                         json:"reward"`
	AcceptanceSpec    AcceptanceSpec   `bson:"acceptanceSpec"                 json:"acceptanceSpec"`
	SettlementTerms   *SettlementTerms `bson:"settlementTerms,omitempty"      json:"settlementTerms,omitempty"`
	Deadline          *time.Time       `bson:"deadline,omitempty"             json:"deadline,omitempty"`
	AssignedExecutor  *shared.Actor    `bson:"assignedExecutor,omitempty"     json:"assignedExecutor,omitempty"`
	CurrentContractID string           `bson:"currentContractId,omitempty"    json:"currentContractId,omitempty"`
	LastActivityAt    *time.Time       `bson:"lastActivityAt,omitempty"       json:"lastActivityAt,omitempty"`
	CreatedAt         time.Time        `bson:"createdAt"                      json:"createdAt"`
	UpdatedAt         time.Time        `bson:"updatedAt"                      json:"updatedAt"`
}
