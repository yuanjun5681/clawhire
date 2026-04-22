package review

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Decision string

const (
	DecisionAccepted Decision = "accepted"
	DecisionRejected Decision = "rejected"
)

type Review struct {
	ReviewID     string       `bson:"reviewId"        json:"reviewId"`
	TaskID       string       `bson:"taskId"          json:"taskId"`
	SubmissionID string       `bson:"submissionId"    json:"submissionId"`
	Reviewer     shared.Actor `bson:"reviewer"        json:"reviewer"`
	Decision     Decision     `bson:"decision"        json:"decision"`
	Reason       string       `bson:"reason,omitempty" json:"reason,omitempty"`
	ReviewedAt   time.Time    `bson:"reviewedAt"      json:"reviewedAt"`
}
