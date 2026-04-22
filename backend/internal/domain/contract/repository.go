package contract

import (
	"context"
	"errors"
	"time"
)

var ErrContractNotFound = errors.New("contract not found")
var ErrActiveContractExists = errors.New("active contract already exists")

type Repository interface {
	Insert(ctx context.Context, c *Contract) error
	FindByID(ctx context.Context, contractID string) (*Contract, error)
	FindActiveByTask(ctx context.Context, taskID string) (*Contract, error)
	UpdateStatus(ctx context.Context, contractID string, status Status, at time.Time) error
	MarkStarted(ctx context.Context, contractID string, at time.Time) error
}
