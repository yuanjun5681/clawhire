package contract

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
	StatusDisputed  Status = "disputed"
)

type Contract struct {
	ContractID   string       `bson:"contractId"          json:"contractId"`
	TaskID       string       `bson:"taskId"              json:"taskId"`
	Requester    shared.Actor `bson:"requester"           json:"requester"`
	Executor     shared.Actor `bson:"executor"            json:"executor"`
	AgreedReward shared.Money `bson:"agreedReward"        json:"agreedReward"`
	Status       Status       `bson:"status"              json:"status"`
	StartedAt    *time.Time   `bson:"startedAt,omitempty" json:"startedAt,omitempty"`
	CreatedAt    time.Time    `bson:"createdAt"           json:"createdAt"`
	UpdatedAt    time.Time    `bson:"updatedAt"           json:"updatedAt"`
}
