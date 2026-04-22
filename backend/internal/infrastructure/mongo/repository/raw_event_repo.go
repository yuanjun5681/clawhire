// Package repository 提供 domain 层仓储接口的 MongoDB 实现。
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

// RawEventRepo 实现 event.RawEventRepository。
type RawEventRepo struct {
	coll *mongo.Collection
}

// NewRawEventRepo 绑定到 raw_events 集合。
func NewRawEventRepo(db *mongo.Database) *RawEventRepo {
	return &RawEventRepo{coll: db.Collection(mgo.CollRawEvents)}
}

// Insert 遇 eventKey 冲突时返回 event.ErrDuplicateEvent。
func (r *RawEventRepo) Insert(ctx context.Context, e *event.RawEvent) error {
	if e.ReceivedAt.IsZero() {
		e.ReceivedAt = time.Now().UTC()
	}
	if e.ProcessStatus == "" {
		e.ProcessStatus = event.ProcessStatusPending
	}
	if _, err := r.coll.InsertOne(ctx, e); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return event.ErrDuplicateEvent
		}
		return fmt.Errorf("insert raw event: %w", err)
	}
	return nil
}

// FindByEventKey 返回对应记录；不存在返回 nil, nil。
func (r *RawEventRepo) FindByEventKey(ctx context.Context, eventKey string) (*event.RawEvent, error) {
	var out event.RawEvent
	err := r.coll.FindOne(ctx, bson.M{"eventKey": eventKey}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("find raw event: %w", err)
	}
	return &out, nil
}

// MarkProcessed 更新处理状态与错误信息（幂等）。
func (r *RawEventRepo) MarkProcessed(ctx context.Context, eventKey string, status event.ProcessStatus, at time.Time, errMsg string) error {
	update := bson.M{
		"processStatus": status,
		"processedAt":   at,
	}
	if errMsg != "" {
		update["errorMessage"] = errMsg
	} else {
		update["errorMessage"] = ""
	}
	res, err := r.coll.UpdateOne(ctx, bson.M{"eventKey": eventKey}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("mark raw event processed: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("mark raw event processed: eventKey=%s not found", eventKey)
	}
	return nil
}
