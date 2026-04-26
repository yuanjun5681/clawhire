package command

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

type RecordSettlementResult struct {
	TaskID       string `json:"taskId"`
	SettlementID string `json:"settlementId"`
	EventID      string `json:"eventId,omitempty"`
}

func (s *Service) RecordSettlement(ctx context.Context, cmd RecordSettlementCommand) (*RecordSettlementResult, error) {
	payload := cmd.Payload
	if strings.TrimSpace(payload.TaskID) == "" {
		return nil, apierr.New(apierr.CodeInvalidMessagePayload, "taskId is required")
	}

	t, err := s.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return nil, toAPIError("find task", err)
	}
	next, _, err := s.sm.Transit(t.Status, task.ActionRecordSettlement)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInvalidState, "record settlement not allowed", err)
	}

	normalized, err := s.normalizeSettlementPayload(ctx, payload, t)
	if err != nil {
		return nil, err
	}
	if err := validateSettlement(normalized); err != nil {
		return nil, err
	}

	status := settlement.Status(clawhire.NormalizeSettlementStatus(normalized.Status))
	if !status.Valid() {
		return nil, apierr.New(apierr.CodeInvalidMessagePayload, "invalid settlement status")
	}

	at := s.now().UTC()
	if normalized.RecordedAt != nil && !normalized.RecordedAt.IsZero() {
		at = normalized.RecordedAt.UTC()
	}

	item := &settlement.Settlement{
		SettlementID: normalized.SettlementID,
		TaskID:       normalized.TaskID,
		ContractID:   firstNonEmpty(normalized.ContractID, t.CurrentContractID),
		Payee:        normalized.Payee,
		Amount:       normalized.Amount,
		Currency:     strings.TrimSpace(normalized.Currency),
		Status:       status,
		Channel:      strings.TrimSpace(normalized.Channel),
		ExternalRef:  strings.TrimSpace(normalized.ExternalRef),
		RecordedAt:   at,
	}
	if err := s.settlements.Insert(ctx, item); err != nil {
		return nil, apierr.Wrap(apierr.CodeInternalError, "insert settlement", err)
	}
	if err := s.tasks.UpdateStatus(ctx, normalized.TaskID, t.Status, next, at); err != nil {
		return nil, toAPIError("update task status", err)
	}
	if err := s.recordDomainEvent(ctx, "task", normalized.TaskID, cmd.Event, normalized); err != nil {
		return nil, err
	}
	return &RecordSettlementResult{
		TaskID:       normalized.TaskID,
		SettlementID: normalized.SettlementID,
		EventID:      eventID(cmd.Event),
	}, nil
}

func (s *Service) normalizeSettlementPayload(ctx context.Context, payload clawhire.RecordSettlementPayload, t *task.Task) (clawhire.RecordSettlementPayload, error) {
	out := payload
	out.TaskID = strings.TrimSpace(out.TaskID)
	out.ContractID = firstNonEmpty(out.ContractID, t.CurrentContractID)
	out.SettlementID = strings.TrimSpace(out.SettlementID)
	if out.SettlementID == "" {
		out.SettlementID = "settlement_" + uuid.New().String()
	}

	if strings.TrimSpace(out.Payee.ID) == "" {
		if t.AssignedExecutor == nil {
			return out, apierr.New(apierr.CodeInvalidMessagePayload, "payee is required")
		}
		out.Payee = *t.AssignedExecutor
	}

	if out.Amount <= 0 || strings.TrimSpace(out.Currency) == "" {
		money := shared.Money{
			Amount:   t.Reward.Amount,
			Currency: string(t.Reward.Currency),
		}
		if out.ContractID != "" && s.contracts != nil {
			if c, err := s.contracts.FindByID(ctx, out.ContractID); err == nil {
				money = c.AgreedReward
			}
		}
		if out.Amount <= 0 {
			out.Amount = money.Amount
		}
		if strings.TrimSpace(out.Currency) == "" {
			out.Currency = money.Currency
		}
	}
	return out, nil
}
