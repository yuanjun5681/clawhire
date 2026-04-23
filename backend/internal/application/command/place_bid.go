package command

import (
	"context"
	"strings"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func (s *Service) PlaceBid(ctx context.Context, cmd PlaceBidCommand) (*PlaceBidResult, error) {
	payload := cmd.Payload
	if err := validatePlaceBid(payload); err != nil {
		return nil, err
	}
	t, err := s.tasks.FindByID(ctx, payload.TaskID)
	if err != nil {
		return nil, toAPIError("find task", err)
	}
	if err := s.sm.CanTransit(t.Status, task.ActionPlaceBid); err != nil {
		return nil, apierr.Wrap(apierr.CodeInvalidState, "place bid not allowed", err)
	}
	item := &bid.Bid{
		BidID:     payload.BidID,
		TaskID:    payload.TaskID,
		Executor:  payload.Executor,
		Price:     payload.Price,
		Currency:  strings.TrimSpace(payload.Currency),
		Proposal:  strings.TrimSpace(payload.Proposal),
		Status:    bid.StatusActive,
		CreatedAt: s.now().UTC(),
	}
	if err := s.bids.Insert(ctx, item); err != nil {
		return nil, apierr.Wrap(apierr.CodeInternalError, "insert bid", err)
	}
	if err := s.tasks.TouchActivity(ctx, payload.TaskID, s.now().UTC()); err != nil {
		return nil, toAPIError("touch task", err)
	}
	if err := s.recordDomainEvent(ctx, "task", payload.TaskID, cmd.Event, payload); err != nil {
		return nil, err
	}
	res := &PlaceBidResult{TaskID: payload.TaskID, BidID: payload.BidID}
	if cmd.Event != nil {
		res.EventID = cmd.Event.ID
	}
	return res, nil
}
