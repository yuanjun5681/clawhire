package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/settlement"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type SettlementRepo struct {
	coll *mongo.Collection
}

func NewSettlementRepo(db *mongo.Database) *SettlementRepo {
	return &SettlementRepo{coll: db.Collection(mgo.CollSettlements)}
}

func (r *SettlementRepo) Insert(ctx context.Context, s *settlement.Settlement) error {
	if _, err := r.coll.InsertOne(ctx, s); err != nil {
		return fmt.Errorf("insert settlement: %w", err)
	}
	return nil
}

func (r *SettlementRepo) FindByID(ctx context.Context, settlementID string) (*settlement.Settlement, error) {
	var out settlement.Settlement
	err := r.coll.FindOne(ctx, bson.M{"settlementId": settlementID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, settlement.ErrSettlementNotFound
		}
		return nil, fmt.Errorf("find settlement: %w", err)
	}
	return &out, nil
}

func (r *SettlementRepo) ListByTask(ctx context.Context, taskID string) ([]*settlement.Settlement, error) {
	cur, err := r.coll.Find(ctx, bson.M{"taskId": taskID}, options.Find().SetSort(bson.D{{Key: "recordedAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("find settlements by task: %w", err)
	}
	defer cur.Close(ctx)

	var list []*settlement.Settlement
	if err := cur.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode settlements by task: %w", err)
	}
	return list, nil
}

func (r *SettlementRepo) ListByPayee(ctx context.Context, payeeID string, page, pageSize int) ([]*settlement.Settlement, int64, error) {
	filter := bson.M{"payee.id": payeeID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count settlements by payee: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(page, pageSize, bson.D{{Key: "recordedAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find settlements by payee: %w", err)
	}
	defer cur.Close(ctx)

	var list []*settlement.Settlement
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode settlements by payee: %w", err)
	}
	return list, total, nil
}
