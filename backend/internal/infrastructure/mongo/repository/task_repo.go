package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/shared"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/task"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type TaskRepo struct {
	coll *mongo.Collection
}

func NewTaskRepo(db *mongo.Database) *TaskRepo {
	return &TaskRepo{coll: db.Collection(mgo.CollTasks)}
}

func (r *TaskRepo) Insert(ctx context.Context, t *task.Task) error {
	if _, err := r.coll.InsertOne(ctx, t); err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (r *TaskRepo) FindByID(ctx context.Context, taskID string) (*task.Task, error) {
	var out task.Task
	err := r.coll.FindOne(ctx, bson.M{"taskId": taskID}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, task.ErrTaskNotFound
		}
		return nil, fmt.Errorf("find task: %w", err)
	}
	return &out, nil
}

func (r *TaskRepo) UpdateStatus(ctx context.Context, taskID string, expected, next task.Status, at time.Time) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"taskId": taskID, "status": expected},
		bson.M{"$set": bson.M{
			"status":         next,
			"updatedAt":      at,
			"lastActivityAt": at,
		}},
	)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	if res.MatchedCount > 0 {
		return nil
	}
	if _, err := r.FindByID(ctx, taskID); err != nil {
		return err
	}
	return task.ErrStatusConflict
}

func (r *TaskRepo) UpdateAssignment(ctx context.Context, taskID string, executor shared.Actor, contractID string, at time.Time) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"taskId": taskID},
		bson.M{"$set": bson.M{
			"assignedExecutor":  executor,
			"currentContractId": contractID,
			"updatedAt":         at,
			"lastActivityAt":    at,
		}},
	)
	if err != nil {
		return fmt.Errorf("update task assignment: %w", err)
	}
	if res.MatchedCount == 0 {
		return task.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepo) TouchActivity(ctx context.Context, taskID string, at time.Time) error {
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"taskId": taskID},
		bson.M{"$set": bson.M{
			"updatedAt":      at,
			"lastActivityAt": at,
		}},
	)
	if err != nil {
		return fmt.Errorf("touch task activity: %w", err)
	}
	if res.MatchedCount == 0 {
		return task.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepo) List(ctx context.Context, f task.Filter) ([]*task.Task, int64, error) {
	filter := bson.M{}
	if len(f.Status) > 0 {
		statuses := make([]string, 0, len(f.Status))
		for _, s := range f.Status {
			statuses = append(statuses, string(s))
		}
		filter["status"] = bson.M{"$in": statuses}
	}
	if f.Category != "" {
		filter["category"] = f.Category
	}
	if f.RequesterID != "" {
		filter["requester.id"] = f.RequesterID
	}
	if f.ExecutorID != "" {
		filter["assignedExecutor.id"] = f.ExecutorID
	}
	if f.ReviewerID != "" {
		filter["reviewer.id"] = f.ReviewerID
	}
	if f.Keyword != "" {
		rx := keywordRegex(f.Keyword)
		filter["$or"] = bson.A{
			bson.M{"title": rx},
			bson.M{"description": rx},
		}
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(f.Page, f.PageSize, bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find tasks: %w", err)
	}
	defer cur.Close(ctx)

	var list []*task.Task
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode tasks: %w", err)
	}
	return list, total, nil
}

func (r *TaskRepo) ListByExecutor(ctx context.Context, executorID string, statuses []task.Status, page, pageSize int) ([]*task.Task, int64, error) {
	filter := bson.M{"assignedExecutor.id": executorID}
	if len(statuses) > 0 {
		vals := make([]string, 0, len(statuses))
		for _, s := range statuses {
			vals = append(vals, string(s))
		}
		filter["status"] = bson.M{"$in": vals}
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count executor tasks: %w", err)
	}
	cur, err := r.coll.Find(ctx, filter, findOptions(page, pageSize, bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, 0, fmt.Errorf("find executor tasks: %w", err)
	}
	defer cur.Close(ctx)

	var list []*task.Task
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode executor tasks: %w", err)
	}
	return list, total, nil
}
