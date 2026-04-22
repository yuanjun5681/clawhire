package submission

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Status string

const (
	StatusSubmitted Status = "submitted"
	StatusAccepted  Status = "accepted"
	StatusRejected  Status = "rejected"
)

type Evidence struct {
	Type  string   `bson:"type,omitempty"  json:"type,omitempty"`
	Items []string `bson:"items,omitempty" json:"items,omitempty"`
}

type Submission struct {
	SubmissionID string            `bson:"submissionId"         json:"submissionId"`
	TaskID       string            `bson:"taskId"               json:"taskId"`
	ContractID   string            `bson:"contractId,omitempty" json:"contractId,omitempty"`
	Executor     shared.Actor      `bson:"executor"             json:"executor"`
	Summary      string            `bson:"summary"              json:"summary"`
	Artifacts    []shared.Artifact `bson:"artifacts,omitempty"  json:"artifacts,omitempty"`
	Evidence     *Evidence         `bson:"evidence,omitempty"   json:"evidence,omitempty"`
	Status       Status            `bson:"status"               json:"status"`
	SubmittedAt  time.Time         `bson:"submittedAt"          json:"submittedAt"`
}
