package milestone

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Status string

const (
	StatusPlanned  Status = "planned"
	StatusDeclared Status = "declared"
	StatusAccepted Status = "accepted"
	StatusRejected Status = "rejected"
)

type Claim struct {
	Type     string  `bson:"type"               json:"type"`
	Amount   float64 `bson:"amount,omitempty"   json:"amount,omitempty"`
	Currency string  `bson:"currency,omitempty" json:"currency,omitempty"`
}

type Milestone struct {
	MilestoneID string            `bson:"milestoneId"          json:"milestoneId"`
	TaskID      string            `bson:"taskId"               json:"taskId"`
	ContractID  string            `bson:"contractId,omitempty" json:"contractId,omitempty"`
	Title       string            `bson:"title"                json:"title"`
	Summary     string            `bson:"summary,omitempty"    json:"summary,omitempty"`
	Status      Status            `bson:"status"               json:"status"`
	Claim       *Claim            `bson:"claim,omitempty"      json:"claim,omitempty"`
	Artifacts   []shared.Artifact `bson:"artifacts,omitempty"  json:"artifacts,omitempty"`
	ReportedAt  *time.Time        `bson:"reportedAt,omitempty" json:"reportedAt,omitempty"`
}
