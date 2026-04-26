package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/application/platform"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

type Now func() time.Time

type EventMeta struct {
	ID   string
	Type string
}

type Service struct {
	tasks       task.Repository
	bids        bid.Repository
	contracts   contract.Repository
	submissions submission.Repository
	reviews     review.Repository
	settlements settlement.Repository
	domainEvts  event.DomainEventRepository
	sm          task.StateMachine
	now         Now
	syncPub     *platform.SyncPublisher // nil 时跳过跨平台同步
}

type Options struct {
	Tasks       task.Repository
	Bids        bid.Repository
	Contracts   contract.Repository
	Submissions submission.Repository
	Reviews     review.Repository
	Settlements settlement.Repository
	DomainEvts  event.DomainEventRepository
	StateMach   task.StateMachine
	Now         Now
	SyncPub     *platform.SyncPublisher
}

type PostTaskCommand struct {
	Payload clawhire.PostTaskPayload
	Event   *EventMeta
}

type PlaceBidCommand struct {
	Payload clawhire.PlaceBidPayload
	Event   *EventMeta
}

type PostTaskResult struct {
	TaskID  string `json:"taskId"`
	EventID string `json:"eventId,omitempty"`
}

type PlaceBidResult struct {
	TaskID  string `json:"taskId"`
	BidID   string `json:"bidId"`
	EventID string `json:"eventId,omitempty"`
}

type AwardTaskCommand struct {
	Payload clawhire.AwardTaskPayload
	Event   *EventMeta
}

type CreateSubmissionCommand struct {
	Payload clawhire.CreateSubmissionPayload
	Event   *EventMeta
}

type AcceptSubmissionCommand struct {
	Payload clawhire.AcceptSubmissionPayload
	Event   *EventMeta
}

type RejectSubmissionCommand struct {
	Payload clawhire.RejectSubmissionPayload
	Event   *EventMeta
}

type RecordSettlementCommand struct {
	Payload clawhire.RecordSettlementPayload
	Event   *EventMeta
}

func NewService(opt Options) *Service {
	now := opt.Now
	if now == nil {
		now = time.Now
	}
	sm := opt.StateMach
	if sm == nil {
		sm = task.NewStateMachine()
	}
	return &Service{
		tasks:       opt.Tasks,
		bids:        opt.Bids,
		contracts:   opt.Contracts,
		submissions: opt.Submissions,
		reviews:     opt.Reviews,
		settlements: opt.Settlements,
		domainEvts:  opt.DomainEvts,
		sm:          sm,
		now:         now,
		syncPub:     opt.SyncPub,
	}
}

func payloadMap(payload interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) recordDomainEvent(ctx context.Context, aggregateType, aggregateID string, meta *EventMeta, payload interface{}) error {
	if s.domainEvts == nil || meta == nil || meta.ID == "" || meta.Type == "" {
		return nil
	}
	data, err := payloadMap(payload)
	if err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "marshal domain event", err)
	}
	if err := s.domainEvts.Insert(ctx, &event.DomainEvent{
		EventID:       meta.ID,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     meta.Type,
		Data:          data,
		CreatedAt:     s.now().UTC(),
	}); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "insert domain event", err)
	}
	return nil
}
