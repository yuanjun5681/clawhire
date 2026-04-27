package account

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrConnectionNotFound = errors.New("platform connection not found")
var ErrConnectionExists = errors.New("platform connection already exists")

// PlatformConnection 记录 ClawHire 账号与外部平台账号的绑定关系。
type PlatformConnection struct {
	ID             bson.ObjectID `bson:"_id"            json:"id"`
	Platform       string        `bson:"platform"       json:"platform"`
	PlatformNodeID string        `bson:"platformNodeId" json:"platformNodeId"`
	LocalUserID    string        `bson:"localUserId"    json:"localUserId"`
	RemoteUserID   string        `bson:"remoteUserId"   json:"remoteUserId"`
	LinkedAt       time.Time     `bson:"linkedAt"       json:"linkedAt"`
}

type PlatformConnectionRepository interface {
	Insert(ctx context.Context, conn *PlatformConnection) error
	// UpsertByLocalUserAndNode 写入或更新本地账号与外部平台节点的绑定关系。
	UpsertByLocalUserAndNode(ctx context.Context, conn *PlatformConnection) error
	// FindByLocalUser 返回指定账号在某平台上的所有连接（platform="" 则返回全部）。
	FindByLocalUser(ctx context.Context, localUserID, platform string) ([]*PlatformConnection, error)
	// FindByRemote 入站事件反查本地账号：通过来源节点 + 对方 userId 定位。
	FindByRemote(ctx context.Context, platformNodeID, remoteUserID string) (*PlatformConnection, error)
	// DeleteByLocalUserAndNode 解除绑定：localUserId + platformNodeId 组合唯一。
	DeleteByLocalUserAndNode(ctx context.Context, localUserID, platformNodeID string) error
}
