package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type AccountRepo struct {
	coll *mongo.Collection
}

func NewAccountRepo(db *mongo.Database) *AccountRepo {
	return &AccountRepo{coll: db.Collection(mgo.CollAccounts)}
}

func (r *AccountRepo) Insert(ctx context.Context, a *account.Account) error {
	if _, err := r.coll.InsertOne(ctx, a); err != nil {
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

func (r *AccountRepo) FindByID(ctx context.Context, accountID string) (*account.Account, error) {
	var out account.Account
	err := r.coll.FindOne(ctx, bson.M{"accountId": accountID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, account.ErrAccountNotFound
		}
		return nil, fmt.Errorf("find account: %w", err)
	}
	return &out, nil
}

func (r *AccountRepo) FindByNodeID(ctx context.Context, nodeID string) (*account.Account, error) {
	var out account.Account
	err := r.coll.FindOne(ctx, bson.M{"nodeId": nodeID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, account.ErrAccountNotFound
		}
		return nil, fmt.Errorf("find account by nodeId: %w", err)
	}
	return &out, nil
}

func (r *AccountRepo) List(ctx context.Context, f account.Filter) ([]*account.Account, int64, error) {
	filter := bson.M{}
	if f.Type != nil {
		filter["type"] = *f.Type
	}
	if f.Status != nil {
		filter["status"] = *f.Status
	}
	if f.OwnerAccountID != "" {
		filter["ownerAccountId"] = f.OwnerAccountID
	}
	if f.NodeID != "" {
		filter["nodeId"] = f.NodeID
	}
	if f.Keyword != "" {
		filter["displayName"] = keywordRegex(f.Keyword)
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count accounts: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(f.Page, f.PageSize, bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find accounts: %w", err)
	}
	defer cur.Close(ctx)

	var list []*account.Account
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode accounts: %w", err)
	}
	return list, total, nil
}

func (r *AccountRepo) ListAgentsByOwner(ctx context.Context, ownerAccountID string, page, pageSize int) ([]*account.Account, int64, error) {
	accType := account.TypeAgent
	return r.List(ctx, account.Filter{
		Type:           &accType,
		OwnerAccountID: ownerAccountID,
		Page:           page,
		PageSize:       pageSize,
	})
}
