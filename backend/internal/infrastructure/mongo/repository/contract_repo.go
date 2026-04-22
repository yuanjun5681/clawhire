package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/contract"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type ContractRepo struct {
	coll *mongo.Collection
}

func NewContractRepo(db *mongo.Database) *ContractRepo {
	return &ContractRepo{coll: db.Collection(mgo.CollContracts)}
}

func (r *ContractRepo) Insert(ctx context.Context, c *contract.Contract) error {
	if _, err := r.coll.InsertOne(ctx, c); err != nil {
		if mongo.IsDuplicateKeyError(err) && strings.Contains(err.Error(), "uk_active_contract_per_task") {
			return contract.ErrActiveContractExists
		}
		return fmt.Errorf("insert contract: %w", err)
	}
	return nil
}

func (r *ContractRepo) FindByID(ctx context.Context, contractID string) (*contract.Contract, error) {
	var out contract.Contract
	err := r.coll.FindOne(ctx, bson.M{"contractId": contractID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, contract.ErrContractNotFound
		}
		return nil, fmt.Errorf("find contract: %w", err)
	}
	return &out, nil
}

func (r *ContractRepo) FindActiveByTask(ctx context.Context, taskID string) (*contract.Contract, error) {
	var out contract.Contract
	err := r.coll.FindOne(ctx, bson.M{"taskId": taskID, "status": contract.StatusActive}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, contract.ErrContractNotFound
		}
		return nil, fmt.Errorf("find active contract: %w", err)
	}
	return &out, nil
}

func (r *ContractRepo) UpdateStatus(ctx context.Context, contractID string, status contract.Status, at time.Time) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"contractId": contractID},
		bson.M{"$set": bson.M{"status": status, "updatedAt": at}},
	)
	if err != nil {
		return fmt.Errorf("update contract status: %w", err)
	}
	if res.MatchedCount == 0 {
		return contract.ErrContractNotFound
	}
	return nil
}

func (r *ContractRepo) MarkStarted(ctx context.Context, contractID string, at time.Time) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"contractId": contractID},
		bson.M{"$set": bson.M{"startedAt": at, "updatedAt": at}},
	)
	if err != nil {
		return fmt.Errorf("mark contract started: %w", err)
	}
	if res.MatchedCount == 0 {
		return contract.ErrContractNotFound
	}
	return nil
}
