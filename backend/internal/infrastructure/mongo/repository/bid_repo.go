package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/bid"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type BidRepo struct {
	coll *mongo.Collection
}

func NewBidRepo(db *mongo.Database) *BidRepo {
	return &BidRepo{coll: db.Collection(mgo.CollBids)}
}

func (r *BidRepo) Insert(ctx context.Context, b *bid.Bid) error {
	if _, err := r.coll.InsertOne(ctx, b); err != nil {
		return fmt.Errorf("insert bid: %w", err)
	}
	return nil
}

func (r *BidRepo) FindByID(ctx context.Context, bidID string) (*bid.Bid, error) {
	var out bid.Bid
	err := r.coll.FindOne(ctx, bson.M{"bidId": bidID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, bid.ErrBidNotFound
		}
		return nil, fmt.Errorf("find bid: %w", err)
	}
	return &out, nil
}

func (r *BidRepo) ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*bid.Bid, int64, error) {
	filter := bson.M{"taskId": taskID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count bids by task: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(page, pageSize, bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find bids by task: %w", err)
	}
	defer cur.Close(ctx)

	var list []*bid.Bid
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode bids by task: %w", err)
	}
	return list, total, nil
}

func (r *BidRepo) ListByExecutor(ctx context.Context, executorID string, page, pageSize int) ([]*bid.Bid, int64, error) {
	filter := bson.M{"executor.id": executorID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count bids by executor: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(page, pageSize, bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find bids by executor: %w", err)
	}
	defer cur.Close(ctx)

	var list []*bid.Bid
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode bids by executor: %w", err)
	}
	return list, total, nil
}

func (r *BidRepo) MarkAwarded(ctx context.Context, bidID string) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"bidId": bidID},
		bson.M{"$set": bson.M{"status": bid.StatusAwarded}},
	)
	if err != nil {
		return fmt.Errorf("mark bid awarded: %w", err)
	}
	if res.MatchedCount == 0 {
		return bid.ErrBidNotFound
	}
	return nil
}

func (r *BidRepo) InvalidateOthers(ctx context.Context, taskID string, exceptBidID string) error {
	filter := bson.M{
		"taskId": taskID,
		"status": bid.StatusActive,
	}
	if exceptBidID != "" {
		filter["bidId"] = bson.M{"$ne": exceptBidID}
	}
	if _, err := r.coll.UpdateMany(ctx, filter, bson.M{"$set": bson.M{"status": bid.StatusRejected}}); err != nil {
		return fmt.Errorf("invalidate bids: %w", err)
	}
	return nil
}
