package clawhire

import (
	"strings"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Reward struct {
	Mode     string  `json:"mode"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type AcceptanceSpec struct {
	Mode  string   `json:"mode"`
	Rules []string `json:"rules,omitempty"`
}

type SettlementTerms struct {
	Trigger string `json:"trigger,omitempty"`
}

type Claim struct {
	Type     string  `json:"type"`
	Amount   float64 `json:"amount,omitempty"`
	Currency string  `json:"currency,omitempty"`
}

type Evidence struct {
	Type  string   `json:"type,omitempty"`
	Items []string `json:"items,omitempty"`
}

type PostTaskPayload struct {
	TaskID          string           `json:"taskId"`
	Requester       shared.Actor     `json:"requester"`
	Reviewer        *shared.Actor    `json:"reviewer,omitempty"`
	Title           string           `json:"title"`
	Description     string           `json:"description,omitempty"`
	Category        string           `json:"category"`
	Reward          Reward           `json:"reward"`
	AcceptanceSpec  *AcceptanceSpec  `json:"acceptanceSpec,omitempty"`
	SettlementTerms *SettlementTerms `json:"settlementTerms,omitempty"`
	Deadline        *time.Time       `json:"deadline,omitempty"`
}

type PlaceBidPayload struct {
	TaskID   string       `json:"taskId"`
	BidID    string       `json:"bidId"`
	Executor shared.Actor `json:"executor"`
	Price    float64      `json:"price"`
	Currency string       `json:"currency"`
	Proposal string       `json:"proposal,omitempty"`
}

type AwardTaskPayload struct {
	TaskID       string       `json:"taskId"`
	ContractID   string       `json:"contractId"`
	Executor     shared.Actor `json:"executor"`
	AgreedReward Reward       `json:"agreedReward"`
}

type StartTaskPayload struct {
	TaskID     string        `json:"taskId"`
	ContractID string        `json:"contractId,omitempty"`
	Executor   *shared.Actor `json:"executor,omitempty"`
	StartedAt  *time.Time    `json:"startedAt,omitempty"`
}

type ReportProgressPayload struct {
	TaskID     string            `json:"taskId"`
	ProgressID string            `json:"progressId"`
	ContractID string            `json:"contractId,omitempty"`
	Executor   shared.Actor      `json:"executor"`
	Stage      string            `json:"stage,omitempty"`
	Percent    *float64          `json:"percent,omitempty"`
	Summary    string            `json:"summary"`
	Artifacts  []shared.Artifact `json:"artifacts,omitempty"`
	ReportedAt *time.Time        `json:"reportedAt,omitempty"`
}

type CompleteMilestonePayload struct {
	TaskID      string            `json:"taskId"`
	ContractID  string            `json:"contractId,omitempty"`
	MilestoneID string            `json:"milestoneId"`
	Executor    *shared.Actor     `json:"executor,omitempty"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary,omitempty"`
	Artifacts   []shared.Artifact `json:"artifacts,omitempty"`
	Claim       *Claim            `json:"claim,omitempty"`
	ReportedAt  *time.Time        `json:"reportedAt,omitempty"`
}

type CreateSubmissionPayload struct {
	TaskID       string            `json:"taskId"`
	SubmissionID string            `json:"submissionId"`
	ContractID   string            `json:"contractId,omitempty"`
	Executor     shared.Actor      `json:"executor"`
	Artifacts    []shared.Artifact `json:"artifacts,omitempty"`
	Summary      string            `json:"summary"`
	Evidence     *Evidence         `json:"evidence,omitempty"`
}

type AcceptSubmissionPayload struct {
	TaskID       string       `json:"taskId"`
	SubmissionID string       `json:"submissionId"`
	AcceptedBy   shared.Actor `json:"acceptedBy"`
	AcceptedAt   *time.Time   `json:"acceptedAt,omitempty"`
}

type RejectSubmissionPayload struct {
	TaskID       string       `json:"taskId"`
	SubmissionID string       `json:"submissionId"`
	RejectedBy   shared.Actor `json:"rejectedBy"`
	Reason       string       `json:"reason"`
	RejectedAt   *time.Time   `json:"rejectedAt,omitempty"`
}

type RecordSettlementPayload struct {
	TaskID       string       `json:"taskId"`
	ContractID   string       `json:"contractId,omitempty"`
	SettlementID string       `json:"settlementId"`
	Payee        shared.Actor `json:"payee"`
	Amount       float64      `json:"amount"`
	Currency     string       `json:"currency"`
	Status       string       `json:"status,omitempty"`
	Channel      string       `json:"channel,omitempty"`
	ExternalRef  string       `json:"externalRef,omitempty"`
	RecordedAt   *time.Time   `json:"recordedAt,omitempty"`
}

type CancelTaskPayload struct {
	TaskID      string        `json:"taskId"`
	CancelledBy *shared.Actor `json:"cancelledBy,omitempty"`
	Reason      string        `json:"reason,omitempty"`
	CancelledAt *time.Time    `json:"cancelledAt,omitempty"`
}

type DisputeTaskPayload struct {
	TaskID     string        `json:"taskId"`
	DisputedBy *shared.Actor `json:"disputedBy,omitempty"`
	Reason     string        `json:"reason,omitempty"`
	DisputedAt *time.Time    `json:"disputedAt,omitempty"`
}

func NormalizeAcceptanceMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "" {
		return "manual"
	}
	return mode
}

func NormalizeSettlementStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return "recorded"
	}
	return status
}
