package command

import (
	"context"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func (s *Service) PostTask(ctx context.Context, cmd PostTaskCommand) (*PostTaskResult, error) {
	payload := cmd.Payload
	if err := validatePostTask(payload); err != nil {
		return nil, err
	}
	if _, err := s.tasks.FindByID(ctx, payload.TaskID); err == nil {
		return nil, apierr.New(apierr.CodeInvalidMessagePayload, "taskId already exists")
	} else if err != task.ErrTaskNotFound {
		return nil, apierr.Wrap(apierr.CodeInternalError, "find task", err)
	}

	now := s.now().UTC()
	reviewer := payload.Reviewer
	if reviewer == nil {
		cp := payload.Requester
		reviewer = &cp
	}
	item := &task.Task{
		TaskID:            payload.TaskID,
		Title:             strings.TrimSpace(payload.Title),
		Description:       strings.TrimSpace(payload.Description),
		Category:          strings.TrimSpace(payload.Category),
		Status:            task.InitialStatusForReward(payload.Reward.Mode),
		Requester:         payload.Requester,
		Reviewer:          reviewer,
		Reward:            task.Reward{Mode: task.RewardMode(strings.TrimSpace(payload.Reward.Mode)), Amount: payload.Reward.Amount, Currency: strings.TrimSpace(payload.Reward.Currency)},
		AcceptanceSpec:    normalizeAcceptanceSpec(payload.AcceptanceSpec),
		SettlementTerms:   normalizeSettlementTerms(payload.SettlementTerms),
		Deadline:          payload.Deadline,
		LastActivityAt:    &now,
		CreatedAt:         now,
		UpdatedAt:         now,
		AssignedExecutor:  nil,
		CurrentContractID: "",
	}
	if err := s.tasks.Insert(ctx, item); err != nil {
		return nil, apierr.Wrap(apierr.CodeInternalError, "insert task", err)
	}
	if err := s.recordDomainEvent(ctx, "task", payload.TaskID, cmd.Event, payload); err != nil {
		return nil, err
	}
	res := &PostTaskResult{TaskID: payload.TaskID}
	if cmd.Event != nil {
		res.EventID = cmd.Event.ID
	}
	return res, nil
}
