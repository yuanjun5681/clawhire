package repository

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/submission"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type SubmissionRepo struct {
	coll *mongo.Collection
}

func NewSubmissionRepo(db *mongo.Database) *SubmissionRepo {
	return &SubmissionRepo{coll: db.Collection(mgo.CollSubmissions)}
}

func (r *SubmissionRepo) Insert(ctx context.Context, s *submission.Submission) error {
	if _, err := r.coll.InsertOne(ctx, s); err != nil {
		return fmt.Errorf("insert submission: %w", err)
	}
	return nil
}

func (r *SubmissionRepo) FindByID(ctx context.Context, submissionID string) (*submission.Submission, error) {
	var out submission.Submission
	err := r.coll.FindOne(ctx, bson.M{"submissionId": submissionID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, submission.ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("find submission: %w", err)
	}
	return &out, nil
}

func (r *SubmissionRepo) UpdateStatus(ctx context.Context, submissionID string, status submission.Status) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"submissionId": submissionID},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return fmt.Errorf("update submission status: %w", err)
	}
	if res.MatchedCount == 0 {
		return submission.ErrSubmissionNotFound
	}
	return nil
}

func (r *SubmissionRepo) ListByTask(ctx context.Context, taskID string, page, pageSize int) ([]*submission.Submission, int64, error) {
	filter := bson.M{"taskId": taskID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count submissions: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(page, pageSize, bson.D{{Key: "submittedAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find submissions: %w", err)
	}
	defer cur.Close(ctx)

	var list []*submission.Submission
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode submissions: %w", err)
	}
	return list, total, nil
}

func (r *SubmissionRepo) LatestByTask(ctx context.Context, taskID string) (*submission.Submission, error) {
	var out submission.Submission
	err := r.coll.FindOne(
		ctx,
		bson.M{"taskId": taskID},
		options.FindOne().SetSort(bson.D{{Key: "submittedAt", Value: -1}}),
	).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, submission.ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("find latest submission: %w", err)
	}
	return &out, nil
}
