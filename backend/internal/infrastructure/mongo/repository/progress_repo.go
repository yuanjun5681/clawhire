package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/progress"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type ProgressRepo struct {
	coll *mongo.Collection
}

func NewProgressRepo(db *mongo.Database) *ProgressRepo {
	return &ProgressRepo{coll: db.Collection(mgo.CollProgress)}
}

func (r *ProgressRepo) Insert(ctx context.Context, report *progress.Report) error {
	if _, err := r.coll.InsertOne(ctx, report); err != nil {
		return fmt.Errorf("insert progress: %w", err)
	}
	return nil
}

func (r *ProgressRepo) ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*progress.Report, int64, error) {
	filter := bson.M{"taskId": taskID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count progress: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(page, pageSize, bson.D{{Key: "reportedAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find progress: %w", err)
	}
	defer cur.Close(ctx)

	var list []*progress.Report
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode progress: %w", err)
	}
	return list, total, nil
}
