package task

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidTransition 在以下情况返回：
//   - action 未定义
//   - 当前状态不在 action 的前置状态集合内
//   - 终态被尝试变更
var ErrInvalidTransition = errors.New("invalid state transition")

// ErrUnknownAction 表示传入的 action 未注册到状态机。
var ErrUnknownAction = errors.New("unknown action")

// transitionRule 描述单个 action 的迁移规则。
// 若 changes=false，动作不改变任务主状态（如 place_bid、report_progress、complete_milestone）。
type transitionRule struct {
	allowedFrom []Status
	nextStatus  Status
	changes     bool
}

// transitions 来源于 docs/state_machine_design.md §五。
var transitions = map[Action]transitionRule{
	// post_task 特殊处理：没有前置状态；初始状态由 InitialStatusForReward 决定。
	ActionPostTask: {allowedFrom: nil, changes: false},

	ActionPlaceBid: {
		allowedFrom: []Status{StatusOpen, StatusBidding},
		changes:     false,
	},
	ActionAwardTask: {
		allowedFrom: []Status{StatusOpen, StatusBidding},
		nextStatus:  StatusAwarded,
		changes:     true,
	},
	ActionStartTask: {
		allowedFrom: []Status{StatusAwarded, StatusRejected},
		nextStatus:  StatusInProgress,
		changes:     true,
	},
	ActionReportProgress: {
		allowedFrom: []Status{StatusInProgress},
		changes:     false,
	},
	ActionCompleteMilestone: {
		allowedFrom: []Status{StatusInProgress},
		changes:     false,
	},
	ActionCreateSubmission: {
		allowedFrom: []Status{StatusInProgress},
		nextStatus:  StatusSubmitted,
		changes:     true,
	},
	ActionAcceptSubmission: {
		allowedFrom: []Status{StatusSubmitted},
		nextStatus:  StatusAccepted,
		changes:     true,
	},
	ActionRejectSubmission: {
		allowedFrom: []Status{StatusSubmitted},
		nextStatus:  StatusRejected,
		changes:     true,
	},
	ActionRecordSettlement: {
		allowedFrom: []Status{StatusAccepted},
		nextStatus:  StatusSettled,
		changes:     true,
	},
	ActionCancelTask: {
		allowedFrom: []Status{StatusOpen, StatusBidding, StatusAwarded},
		nextStatus:  StatusCancelled,
		changes:     true,
	},
	ActionDisputeTask: {
		allowedFrom: []Status{StatusOpen, StatusBidding, StatusAwarded, StatusInProgress, StatusSubmitted},
		nextStatus:  StatusDisputed,
		changes:     true,
	},
}

// StateMachine 是无状态的纯函数集合。通过接口抽象便于在测试中替换。
type StateMachine interface {
	CanTransit(current Status, action Action) error
	Transit(current Status, action Action) (next Status, changed bool, err error)
}

type stateMachine struct{}

func NewStateMachine() StateMachine { return stateMachine{} }

func (stateMachine) CanTransit(current Status, action Action) error {
	rule, ok := transitions[action]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownAction, action)
	}

	// post_task 无前置，单独处理。
	if action == ActionPostTask {
		return nil
	}

	if current.IsTerminal() {
		return fmt.Errorf("%w: action %s not allowed on terminal state %s", ErrInvalidTransition, action, current)
	}

	for _, s := range rule.allowedFrom {
		if s == current {
			return nil
		}
	}
	return fmt.Errorf("%w: action %s not allowed from %s", ErrInvalidTransition, action, current)
}

// Transit 返回下一状态及是否发生变更；若 action 不改变主状态，返回 (current, false, nil)。
// 注意：post_task 不走 Transit，调用方应直接使用 InitialStatusForReward 初始化。
func (sm stateMachine) Transit(current Status, action Action) (Status, bool, error) {
	if action == ActionPostTask {
		return "", false, fmt.Errorf("%w: post_task must use InitialStatusForReward", ErrInvalidTransition)
	}
	if err := sm.CanTransit(current, action); err != nil {
		return "", false, err
	}
	rule := transitions[action]
	if !rule.changes {
		return current, false, nil
	}
	return rule.nextStatus, true, nil
}

// InitialStatusForReward 决定 clawhire.task.posted 的初始状态。
// 约定：
//   - reward.mode = "bid"       → BIDDING（需要报价）
//   - reward.mode = "fixed"     → OPEN
//   - reward.mode = "milestone" → OPEN
//   - 其他                      → OPEN
func InitialStatusForReward(rewardMode string) Status {
	if strings.EqualFold(strings.TrimSpace(rewardMode), "bid") {
		return StatusBidding
	}
	return StatusOpen
}
