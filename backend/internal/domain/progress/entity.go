package progress

import (
	"time"
)

type Report struct {
	ProgressID string     `bson:"progressId"           json:"progressId"`
	TaskID     string     `bson:"taskId"               json:"taskId"`
	ContractID string     `bson:"contractId,omitempty" json:"contractId,omitempty"`
	Stage      string     `bson:"stage,omitempty"      json:"stage,omitempty"`
	Percent    *float64   `bson:"percent,omitempty"    json:"percent,omitempty"`
	Summary    string     `bson:"summary"              json:"summary"`
	ReportedAt time.Time  `bson:"reportedAt"           json:"reportedAt"`
}
