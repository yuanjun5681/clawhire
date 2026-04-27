package webhook

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawhire"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

func (d *CommandDispatcher) handleConnectionEstablished(ctx context.Context, env *clawsynapse.Envelope) error {
	if d.connections == nil {
		return apierr.New(apierr.CodeInternalError, "platform connection repository is not configured")
	}
	var payload clawhire.ConnectionEstablishedPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	trustMeshNodeID := firstNonEmpty(payload.TrustMeshNodeID, env.From)
	remoteUserID := strings.TrimSpace(payload.RemoteUserID)
	if trustMeshNodeID == "" || remoteUserID == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "trustMeshNodeId and remoteUserId are required")
	}
	linkedAt := chooseTime(payload.LinkedAt, d.now)
	conn := &account.PlatformConnection{
		ID:             bson.NewObjectID(),
		Platform:       "trustmesh",
		PlatformNodeID: trustMeshNodeID,
		LocalUserID:    remoteUserID,
		RemoteUserID:   remoteUserID,
		LinkedAt:       linkedAt,
	}
	if err := d.connections.UpsertByLocalUserAndNode(ctx, conn); err != nil {
		return apierr.Wrap(apierr.CodeInternalError, "upsert platform connection", err)
	}
	return nil
}

func (d *CommandDispatcher) handleConnectionRemoved(ctx context.Context, env *clawsynapse.Envelope) error {
	if d.connections == nil {
		return apierr.New(apierr.CodeInternalError, "platform connection repository is not configured")
	}
	var payload clawhire.ConnectionRemovedPayload
	if err := decodeMessage(env, &payload); err != nil {
		return err
	}
	trustMeshNodeID := firstNonEmpty(payload.TrustMeshNodeID, env.From)
	remoteUserID := strings.TrimSpace(payload.RemoteUserID)
	if trustMeshNodeID == "" || remoteUserID == "" {
		return apierr.New(apierr.CodeInvalidMessagePayload, "trustMeshNodeId and remoteUserId are required")
	}
	if err := d.connections.DeleteByLocalUserAndNode(ctx, remoteUserID, trustMeshNodeID); err != nil {
		if err == account.ErrConnectionNotFound {
			return nil
		}
		return apierr.Wrap(apierr.CodeInternalError, "delete platform connection", err)
	}
	return nil
}
