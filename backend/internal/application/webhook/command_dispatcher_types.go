package webhook

import (
	"time"

	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/milestone"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/progress"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
)

type CommandDispatcher struct {
	tasks       task.Repository
	bids        bid.Repository
	contracts   contract.Repository
	progress    progress.Repository
	milestones  milestone.Repository
	submissions submission.Repository
	reviews     review.Repository
	settlements settlement.Repository
	accounts    account.Repository
	domainEvts  event.DomainEventRepository
	sm          task.StateMachine
	now         Now
	commands    *appcmd.Service
	handlers    map[string]commandFunc
}

type CommandDispatcherOptions struct {
	Tasks       task.Repository
	Bids        bid.Repository
	Contracts   contract.Repository
	Progress    progress.Repository
	Milestones  milestone.Repository
	Submissions submission.Repository
	Reviews     review.Repository
	Settlements settlement.Repository
	Accounts    account.Repository
	DomainEvts  event.DomainEventRepository
	StateMach   task.StateMachine
	Commands    *appcmd.Service
	Now         Now
}

func NewCommandDispatcher(opt CommandDispatcherOptions) *CommandDispatcher {
	now := opt.Now
	if now == nil {
		now = time.Now
	}
	sm := opt.StateMach
	if sm == nil {
		sm = task.NewStateMachine()
	}

	commands := opt.Commands
	if commands == nil {
		commands = appcmd.NewService(appcmd.Options{
			Tasks:      opt.Tasks,
			Bids:       opt.Bids,
			DomainEvts: opt.DomainEvts,
			StateMach:  sm,
			Now:        appcmd.Now(now),
		})
	}

	d := &CommandDispatcher{
		tasks:       opt.Tasks,
		bids:        opt.Bids,
		contracts:   opt.Contracts,
		progress:    opt.Progress,
		milestones:  opt.Milestones,
		submissions: opt.Submissions,
		reviews:     opt.Reviews,
		settlements: opt.Settlements,
		accounts:    opt.Accounts,
		domainEvts:  opt.DomainEvts,
		sm:          sm,
		now:         now,
		commands:    commands,
	}
	d.handlers = map[string]commandFunc{
		clawhire.TypeTaskPosted:         d.handleTaskPosted,
		clawhire.TypeBidPlaced:          d.handleBidPlaced,
		clawhire.TypeTaskAwarded:        d.handleTaskAwarded,
		clawhire.TypeTaskStarted:        d.handleTaskStarted,
		clawhire.TypeProgressReported:   d.handleProgressReported,
		clawhire.TypeMilestoneCompleted: d.handleMilestoneCompleted,
		clawhire.TypeSubmissionCreated:  d.handleSubmissionCreated,
		clawhire.TypeSubmissionAccepted: d.handleSubmissionAccepted,
		clawhire.TypeSubmissionRejected: d.handleSubmissionRejected,
		clawhire.TypeSettlementRecorded: d.handleSettlementRecorded,
		clawhire.TypeTaskCancelled:      d.handleTaskCancelled,
		clawhire.TypeTaskDisputed:       d.handleTaskDisputed,
	}
	return d
}
