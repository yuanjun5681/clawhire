package task

type Status string

const (
	StatusOpen       Status = "OPEN"
	StatusBidding    Status = "BIDDING"
	StatusAwarded    Status = "AWARDED"
	StatusInProgress Status = "IN_PROGRESS"
	StatusSubmitted  Status = "SUBMITTED"
	StatusAccepted   Status = "ACCEPTED"
	StatusSettled    Status = "SETTLED"
	StatusRejected   Status = "REJECTED"
	StatusCancelled  Status = "CANCELLED"
	StatusExpired    Status = "EXPIRED"
	StatusDisputed   Status = "DISPUTED"
)

// IsTerminal 终态：不允许任何消息再推进主状态。
func (s Status) IsTerminal() bool {
	switch s {
	case StatusSettled, StatusCancelled, StatusExpired:
		return true
	}
	return false
}

// IsValid 校验状态是否合法枚举。
func (s Status) IsValid() bool {
	switch s {
	case StatusOpen, StatusBidding, StatusAwarded, StatusInProgress,
		StatusSubmitted, StatusAccepted, StatusSettled, StatusRejected,
		StatusCancelled, StatusExpired, StatusDisputed:
		return true
	}
	return false
}
