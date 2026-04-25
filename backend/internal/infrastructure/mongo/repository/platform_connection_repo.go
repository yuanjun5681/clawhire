package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type PlatformConnectionRepo struct {
	coll *mongo.Collection
}

func NewPlatformConnectionRepo(db *mongo.Database) *PlatformConnectionRepo {
	return &PlatformConnectionRepo{coll: db.Collection(mgo.CollPlatformConnections)}
}

func (r *PlatformConnectionRepo) Insert(ctx context.Context, conn *account.PlatformConnection) error {
	conn.LinkedAt = time.Now().UTC()
	_, err := r.coll.InsertOne(ctx, conn)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return account.ErrConnectionExists
		}
		return fmt.Errorf("insert platform connection: %w", err)
	}
	return nil
}

func (r *PlatformConnectionRepo) FindByLocalUser(ctx context.Context, localUserID, platform string) ([]*account.PlatformConnection, error) {
	filter := bson.M{"localUserId": localUserID}
	if platform != "" {
		filter["platform"] = platform
	}
	cur, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find platform connections: %w", err)
	}
	defer cur.Close(ctx)
	var list []*account.PlatformConnection
	if err := cur.All(ctx, &list); err != nil {
		return nil, fmt.Errorf("decode platform connections: %w", err)
	}
	return list, nil
}

func (r *PlatformConnectionRepo) FindByRemote(ctx context.Context, platformNodeID, remoteUserID string) (*account.PlatformConnection, error) {
	var out account.PlatformConnection
	err := r.coll.FindOne(ctx, bson.M{
		"platformNodeId": platformNodeID,
		"remoteUserId":   remoteUserID,
	}).Decode(&out)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, account.ErrConnectionNotFound
		}
		return nil, fmt.Errorf("find platform connection by remote: %w", err)
	}
	return &out, nil
}

func (r *PlatformConnectionRepo) DeleteByLocalUserAndNode(ctx context.Context, localUserID, platformNodeID string) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{
		"localUserId":    localUserID,
		"platformNodeId": platformNodeID,
	})
	if err != nil {
		return fmt.Errorf("delete platform connection: %w", err)
	}
	if res.DeletedCount == 0 {
		return account.ErrConnectionNotFound
	}
	return nil
}
