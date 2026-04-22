package task

// Action 是 clawhire.* 消息对应的领域动作。
// 仅定义领域层概念；协议层的 messageType 到 Action 的映射放在 internal/protocol/clawhire。
type Action string

const (
	ActionPostTask          Action = "post_task"
	ActionPlaceBid          Action = "place_bid"
	ActionAwardTask         Action = "award_task"
	ActionStartTask         Action = "start_task"
	ActionReportProgress    Action = "report_progress"
	ActionCompleteMilestone Action = "complete_milestone"
	ActionCreateSubmission  Action = "create_submission"
	ActionAcceptSubmission  Action = "accept_submission"
	ActionRejectSubmission  Action = "reject_submission"
	ActionRecordSettlement  Action = "record_settlement"
	ActionCancelTask        Action = "cancel_task"
	ActionDisputeTask       Action = "dispute_task"
)

func (a Action) String() string { return string(a) }
