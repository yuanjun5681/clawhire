package progress

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Report struct {
	ProgressID string            `bson:"progressId"           json:"progressId"`
	TaskID     string            `bson:"taskId"               json:"taskId"`
	ContractID string            `bson:"contractId,omitempty" json:"contractId,omitempty"`
	Executor   shared.Actor      `bson:"executor"             json:"executor"`
	Stage      string            `bson:"stage,omitempty"      json:"stage,omitempty"`
	Percent    *float64          `bson:"percent,omitempty"    json:"percent,omitempty"`
	Summary    string            `bson:"summary"              json:"summary"`
	Artifacts  []shared.Artifact `bson:"artifacts,omitempty"  json:"artifacts,omitempty"`
	ReportedAt time.Time         `bson:"reportedAt"           json:"reportedAt"`
}
