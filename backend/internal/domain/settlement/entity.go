package settlement

import (
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

type Status string

const (
	StatusRecorded       Status = "recorded"
	StatusPendingPayment Status = "pending_payment"
	StatusPaid           Status = "paid"
	StatusFailed         Status = "failed"
	StatusRefunded       Status = "refunded"
)

func (s Status) Valid() bool {
	switch s {
	case StatusRecorded, StatusPendingPayment, StatusPaid, StatusFailed, StatusRefunded:
		return true
	}
	return false
}

type Settlement struct {
	SettlementID string       `bson:"settlementId"         json:"settlementId"`
	TaskID       string       `bson:"taskId"               json:"taskId"`
	ContractID   string       `bson:"contractId,omitempty" json:"contractId,omitempty"`
	Payee        shared.Actor `bson:"payee"                json:"payee"`
	Amount       float64      `bson:"amount"               json:"amount"`
	Currency     string       `bson:"currency"             json:"currency"`
	Status       Status       `bson:"status"               json:"status"`
	Channel      string       `bson:"channel,omitempty"    json:"channel,omitempty"`
	ExternalRef  string       `bson:"externalRef,omitempty" json:"externalRef,omitempty"`
	RecordedAt   time.Time    `bson:"recordedAt"           json:"recordedAt"`
}
