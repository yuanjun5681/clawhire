package task

import (
	"context"
	"errors"
	"time"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
)

// ErrTaskNotFound 表示任务不存在。
var ErrTaskNotFound = errors.New("task not found")

// ErrStatusConflict 表示期望的状态与实际不符（CAS 冲突）。
var ErrStatusConflict = errors.New("task status conflict")

// Filter 用于任务大厅查询。
type Filter struct {
	Status      []Status
	Category    string
	RequesterID string
	ExecutorID  string
	Keyword     string
	Page        int
	PageSize    int
}

type Repository interface {
	Insert(ctx context.Context, t *Task) error
	FindByID(ctx context.Context, taskID string) (*Task, error)

	// UpdateStatus 使用 CAS 保证并发安全：仅当当前 status=expected 时才更新为 next。
	UpdateStatus(ctx context.Context, taskID string, expected, next Status, at time.Time) error

	UpdateAssignment(ctx context.Context, taskID string, executor shared.Actor, contractID string, at time.Time) error
	TouchActivity(ctx context.Context, taskID string, at time.Time) error

	List(ctx context.Context, f Filter) ([]*Task, int64, error)
	ListByExecutor(ctx context.Context, executorID string, statuses []Status, page, pageSize int) ([]*Task, int64, error)
}
