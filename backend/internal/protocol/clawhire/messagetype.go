// Package clawhire 定义 clawhire.* 业务消息类型与 payload。
//
// 所有消息都以 "clawhire." 前缀命名，以便与其他业务系统隔离。
// 详细定义见 docs/clawhire_proposal.md 第九节。
package clawhire

import "strings"

// TypePrefix 是所有 clawhire 业务消息的统一前缀。
const TypePrefix = "clawhire."

// IsClawHireType 判断给定 type 是否属于 clawhire.* 族。
func IsClawHireType(t string) bool {
	return strings.HasPrefix(strings.TrimSpace(t), TypePrefix)
}

// 已知消息类型常量。
const (
	TypeTaskPosted          = "clawhire.task.posted"
	TypeBidPlaced           = "clawhire.bid.placed"
	TypeTaskAwarded         = "clawhire.task.awarded"
	TypeTaskStarted         = "clawhire.task.started"
	TypeProgressReported    = "clawhire.progress.reported"
	TypeMilestoneCompleted  = "clawhire.milestone.completed"
	TypeSubmissionCreated   = "clawhire.submission.created"
	TypeSubmissionAccepted  = "clawhire.submission.accepted"
	TypeSubmissionRejected  = "clawhire.submission.rejected"
	TypeSettlementRecorded  = "clawhire.settlement.recorded"
	TypeTaskCancelled       = "clawhire.task.cancelled"
	TypeTaskDisputed        = "clawhire.task.disputed"
)

// KnownTypes 枚举当前 MVP 支持的所有消息类型。
func KnownTypes() []string {
	return []string{
		TypeTaskPosted,
		TypeBidPlaced,
		TypeTaskAwarded,
		TypeTaskStarted,
		TypeProgressReported,
		TypeMilestoneCompleted,
		TypeSubmissionCreated,
		TypeSubmissionAccepted,
		TypeSubmissionRejected,
		TypeSettlementRecorded,
		TypeTaskCancelled,
		TypeTaskDisputed,
	}
}

// IsKnown 判断是否为当前 MVP 支持的消息类型。
func IsKnown(t string) bool {
	for _, k := range KnownTypes() {
		if k == t {
			return true
		}
	}
	return false
}
