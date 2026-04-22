package task

import (
	"errors"
	"testing"
)

func TestStateMachine_CanTransit_Valid(t *testing.T) {
	sm := NewStateMachine()

	cases := []struct {
		name    string
		current Status
		action  Action
	}{
		{"post_task", "", ActionPostTask},
		{"place_bid_on_open", StatusOpen, ActionPlaceBid},
		{"place_bid_on_bidding", StatusBidding, ActionPlaceBid},
		{"award_from_open", StatusOpen, ActionAwardTask},
		{"award_from_bidding", StatusBidding, ActionAwardTask},
		{"start_from_awarded", StatusAwarded, ActionStartTask},
		{"start_from_rejected", StatusRejected, ActionStartTask},
		{"progress_in_progress", StatusInProgress, ActionReportProgress},
		{"milestone_in_progress", StatusInProgress, ActionCompleteMilestone},
		{"submission_in_progress", StatusInProgress, ActionCreateSubmission},
		{"accept_submitted", StatusSubmitted, ActionAcceptSubmission},
		{"reject_submitted", StatusSubmitted, ActionRejectSubmission},
		{"settlement_accepted", StatusAccepted, ActionRecordSettlement},
		{"cancel_from_open", StatusOpen, ActionCancelTask},
		{"cancel_from_bidding", StatusBidding, ActionCancelTask},
		{"cancel_from_awarded", StatusAwarded, ActionCancelTask},
		{"dispute_in_progress", StatusInProgress, ActionDisputeTask},
		{"dispute_submitted", StatusSubmitted, ActionDisputeTask},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := sm.CanTransit(c.current, c.action); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestStateMachine_CanTransit_Invalid(t *testing.T) {
	sm := NewStateMachine()

	cases := []struct {
		name    string
		current Status
		action  Action
	}{
		{"bid_on_awarded", StatusAwarded, ActionPlaceBid},
		{"bid_on_submitted", StatusSubmitted, ActionPlaceBid},
		{"progress_on_open", StatusOpen, ActionReportProgress},
		{"submission_on_open", StatusOpen, ActionCreateSubmission},
		{"accept_on_in_progress", StatusInProgress, ActionAcceptSubmission},
		{"accept_on_rejected", StatusRejected, ActionAcceptSubmission},
		{"settlement_on_submitted", StatusSubmitted, ActionRecordSettlement},
		{"settlement_on_rejected", StatusRejected, ActionRecordSettlement},
		{"cancel_on_in_progress", StatusInProgress, ActionCancelTask},
		{"dispute_on_settled", StatusSettled, ActionDisputeTask},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := sm.CanTransit(c.current, c.action)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, ErrInvalidTransition) {
				t.Fatalf("expected ErrInvalidTransition, got %v", err)
			}
		})
	}
}

func TestStateMachine_CanTransit_TerminalProtected(t *testing.T) {
	sm := NewStateMachine()
	terminals := []Status{StatusSettled, StatusCancelled, StatusExpired}
	actions := []Action{
		ActionPlaceBid, ActionAwardTask, ActionStartTask, ActionReportProgress,
		ActionCompleteMilestone, ActionCreateSubmission, ActionAcceptSubmission,
		ActionRejectSubmission, ActionRecordSettlement, ActionCancelTask, ActionDisputeTask,
	}

	for _, s := range terminals {
		for _, a := range actions {
			if err := sm.CanTransit(s, a); err == nil {
				t.Fatalf("terminal %s should reject action %s", s, a)
			}
		}
	}
}

func TestStateMachine_CanTransit_UnknownAction(t *testing.T) {
	sm := NewStateMachine()
	err := sm.CanTransit(StatusOpen, Action("no_such_action"))
	if err == nil || !errors.Is(err, ErrUnknownAction) {
		t.Fatalf("expected ErrUnknownAction, got %v", err)
	}
}

func TestStateMachine_Transit_ChangesAndStays(t *testing.T) {
	sm := NewStateMachine()

	cases := []struct {
		name     string
		current  Status
		action   Action
		want     Status
		changed  bool
	}{
		{"award_to_awarded", StatusOpen, ActionAwardTask, StatusAwarded, true},
		{"start_to_in_progress", StatusAwarded, ActionStartTask, StatusInProgress, true},
		{"reject_returns_to_rejected", StatusSubmitted, ActionRejectSubmission, StatusRejected, true},
		{"rework_from_rejected", StatusRejected, ActionStartTask, StatusInProgress, true},
		{"submit_to_submitted", StatusInProgress, ActionCreateSubmission, StatusSubmitted, true},
		{"accept_to_accepted", StatusSubmitted, ActionAcceptSubmission, StatusAccepted, true},
		{"settlement_to_settled", StatusAccepted, ActionRecordSettlement, StatusSettled, true},
		{"cancel_to_cancelled", StatusBidding, ActionCancelTask, StatusCancelled, true},
		{"dispute_from_in_progress", StatusInProgress, ActionDisputeTask, StatusDisputed, true},

		// 不改变主状态的动作
		{"place_bid_stays", StatusOpen, ActionPlaceBid, StatusOpen, false},
		{"report_progress_stays", StatusInProgress, ActionReportProgress, StatusInProgress, false},
		{"complete_milestone_stays", StatusInProgress, ActionCompleteMilestone, StatusInProgress, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			next, changed, err := sm.Transit(c.current, c.action)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if changed != c.changed {
				t.Fatalf("changed: want %v got %v", c.changed, changed)
			}
			if next != c.want {
				t.Fatalf("next: want %s got %s", c.want, next)
			}
		})
	}
}

func TestStateMachine_Transit_PostTaskDisallowed(t *testing.T) {
	sm := NewStateMachine()
	_, _, err := sm.Transit("", ActionPostTask)
	if err == nil {
		t.Fatal("post_task should not go through Transit")
	}
}

func TestInitialStatusForReward(t *testing.T) {
	cases := []struct {
		mode string
		want Status
	}{
		{"bid", StatusBidding},
		{"BID", StatusBidding},
		{" bid ", StatusBidding},
		{"fixed", StatusOpen},
		{"milestone", StatusOpen},
		{"", StatusOpen},
		{"unknown", StatusOpen},
	}
	for _, c := range cases {
		t.Run(c.mode, func(t *testing.T) {
			got := InitialStatusForReward(c.mode)
			if got != c.want {
				t.Fatalf("InitialStatusForReward(%q) = %s, want %s", c.mode, got, c.want)
			}
		})
	}
}
