package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

// DomainEventRepo 实现 event.DomainEventRepository。
type DomainEventRepo struct {
	coll *mongo.Collection
}

func NewDomainEventRepo(db *mongo.Database) *DomainEventRepo {
	return &DomainEventRepo{coll: db.Collection(mgo.CollDomainEvents)}
}

func (r *DomainEventRepo) Insert(ctx context.Context, e *event.DomainEvent) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	if _, err := r.coll.InsertOne(ctx, e); err != nil {
		return fmt.Errorf("insert domain event: %w", err)
	}
	return nil
}

func (r *DomainEventRepo) ListByAggregate(ctx context.Context, aggType, aggID string, page, pageSize int) ([]*event.DomainEvent, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	filter := bson.M{"aggregateType": aggType, "aggregateId": aggID}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count domain events: %w", err)
	}

	opt := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))
	cur, err := r.coll.Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, fmt.Errorf("find domain events: %w", err)
	}
	defer cur.Close(ctx)

	var list []*event.DomainEvent
	if err := cur.All(ctx, &list); err != nil {
		return nil, 0, fmt.Errorf("decode domain events: %w", err)
	}
	return list, total, nil
}
