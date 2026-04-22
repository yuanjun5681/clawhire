package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/review"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type ReviewRepo struct {
	coll *mongo.Collection
}

func NewReviewRepo(db *mongo.Database) *ReviewRepo {
	return &ReviewRepo{coll: db.Collection(mgo.CollReviews)}
}

func (r *ReviewRepo) Insert(ctx context.Context, rev *review.Review) error {
	if _, err := r.coll.InsertOne(ctx, rev); err != nil {
		return fmt.Errorf("insert review: %w", err)
	}
	return nil
}

func (r *ReviewRepo) ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*review.Review, int64, error) {
	filter := bson.M{"taskId": taskID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(page, pageSize, bson.D{{Key: "reviewedAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find reviews: %w", err)
	}
	defer cur.Close(ctx)

	var list []*review.Review
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode reviews: %w", err)
	}
	return list, total, nil
}

func (r *ReviewRepo) ListBySubmission(ctx context.Context, submissionID string) ([]*review.Review, error) {
	cur, err := r.coll.Find(ctx, bson.M{"submissionId": submissionID}, options.Find().SetSort(bson.D{{Key: "reviewedAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("find reviews by submission: %w", err)
	}
	defer cur.Close(ctx)

	var list []*review.Review
	if err := cur.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode reviews by submission: %w", err)
	}
	return list, nil
}
