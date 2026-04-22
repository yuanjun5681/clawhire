package bid

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusWithdrawn Status = "withdrawn"
	StatusRejected  Status = "rejected"
	StatusAwarded   Status = "awarded"
)

type Bid struct {
	BidID     string       `bson:"bidId"            json:"bidId"`
	TaskID    string       `bson:"taskId"           json:"taskId"`
	Executor  shared.Actor `bson:"executor"         json:"executor"`
	Price     float64      `bson:"price"            json:"price"`
	Currency  string       `bson:"currency"         json:"currency"`
	Proposal  string       `bson:"proposal,omitempty" json:"proposal,omitempty"`
	Status    Status       `bson:"status"           json:"status"`
	CreatedAt time.Time    `bson:"createdAt"        json:"createdAt"`
}
