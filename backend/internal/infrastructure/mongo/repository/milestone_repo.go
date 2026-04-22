package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/milestone"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type MilestoneRepo struct {
	coll *mongo.Collection
}

func NewMilestoneRepo(db *mongo.Database) *MilestoneRepo {
	return &MilestoneRepo{coll: db.Collection(mgo.CollMilestones)}
}

func (r *MilestoneRepo) Upsert(ctx context.Context, m *milestone.Milestone) error {
	_, err := r.coll.ReplaceOne(
		ctx,
		bson.M{"milestoneId": m.MilestoneID},
		m,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert milestone: %w", err)
	}
	return nil
}

func (r *MilestoneRepo) FindByID(ctx context.Context, milestoneID string) (*milestone.Milestone, error) {
	var out milestone.Milestone
	err := r.coll.FindOne(ctx, bson.M{"milestoneId": milestoneID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find milestone: %w", err)
	}
	return &out, nil
}

func (r *MilestoneRepo) ListByTask(ctx context.Context, taskID string) ([]*milestone.Milestone, error) {
	cur, err := r.coll.Find(ctx, bson.M{"taskId": taskID}, options.Find().SetSort(bson.D{{Key: "reportedAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("find milestones: %w", err)
	}
	defer cur.Close(ctx)

	var list []*milestone.Milestone
	if err := cur.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode milestones: %w", err)
	}
	return list, nil
}
