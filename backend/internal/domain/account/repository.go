package account

import (
	"context"
	"errors"
)

var ErrAccountNotFound = errors.New("account not found")

type Filter struct {
	Type           *Type
	Status         *Status
	OwnerAccountID string
	NodeID         string
	Keyword        string
	Page           int
	PageSize       int
}

type Repository interface {
	Insert(ctx context.Context, a *Account) error
	FindByID(ctx context.Context, accountID string) (*Account, error)
	FindByNodeID(ctx context.Context, nodeID string) (*Account, error)
	List(ctx context.Context, f Filter) ([]*Account, int64, error)
	ListAgentsByOwner(ctx context.Context, ownerAccountID string, page, pageSize int) ([]*Account, int64, error)
}
