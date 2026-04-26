package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/response"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/middleware"
)

type Connections struct {
	connections             account.PlatformConnectionRepository
	defaultNodeIDByPlatform map[string]string
}

func NewConnections(repo account.PlatformConnectionRepository, defaultNodes map[string]string) *Connections {
	if defaultNodes == nil {
		defaultNodes = map[string]string{}
	}
	return &Connections{connections: repo, defaultNodeIDByPlatform: defaultNodes}
}

type createConnectionRequest struct {
	Platform       string `json:"platform"`
	RemoteUserID   string `json:"remoteUserId"`
	PlatformNodeID string `json:"platformNodeId,omitempty"`
}

// ListConnections GET /api/accounts/me/connections
func (h *Connections) ListConnections(c *gin.Context) {
	accountID := middleware.CurrentAccountID(c)
	platform := strings.TrimSpace(c.Query("platform"))

	list, err := h.connections.FindByLocalUser(c.Request.Context(), accountID, platform)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, apierr.CodeInternalError, "list connections failed")
		return
	}
	if list == nil {
		list = []*account.PlatformConnection{}
	}
	response.OK(c, list)
}

// CreateConnection POST /api/accounts/me/connections
func (h *Connections) CreateConnection(c *gin.Context) {
	accountID := middleware.CurrentAccountID(c)

	var req createConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "invalid request body")
		return
	}
	req.Platform = strings.TrimSpace(req.Platform)
	req.RemoteUserID = strings.TrimSpace(req.RemoteUserID)
	req.PlatformNodeID = strings.TrimSpace(req.PlatformNodeID)

	if req.Platform == "" {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "platform is required")
		return
	}
	if req.RemoteUserID == "" {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "remoteUserId is required")
		return
	}

	nodeID := req.PlatformNodeID
	if nodeID == "" {
		nodeID = h.defaultNodeIDByPlatform[req.Platform]
	}
	if nodeID == "" {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "platformNodeId is required (no default configured for platform: "+req.Platform+")")
		return
	}

	conn := &account.PlatformConnection{
		ID:             bson.NewObjectID(),
		Platform:       req.Platform,
		PlatformNodeID: nodeID,
		LocalUserID:    accountID,
		RemoteUserID:   req.RemoteUserID,
		LinkedAt:       time.Now().UTC(),
	}
	if err := h.connections.Insert(c.Request.Context(), conn); err != nil {
		if err == account.ErrConnectionExists {
			response.Fail(c, http.StatusConflict, apierr.CodeConflict, "connection already exists for this platform node")
			return
		}
		response.Fail(c, http.StatusInternalServerError, apierr.CodeInternalError, "create connection failed")
		return
	}
	response.Created(c, conn)
}

// DeleteConnection DELETE /api/accounts/me/connections/:platform?platformNodeId=xxx
func (h *Connections) DeleteConnection(c *gin.Context) {
	accountID := middleware.CurrentAccountID(c)
	platform := strings.TrimSpace(c.Param("platform"))

	platformNodeID := strings.TrimSpace(c.Query("platformNodeId"))
	if platformNodeID == "" {
		platformNodeID = h.defaultNodeIDByPlatform[platform]
	}
	if platformNodeID == "" {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "platformNodeId query param is required (no default configured for platform: "+platform+")")
		return
	}

	if err := h.connections.DeleteByLocalUserAndNode(c.Request.Context(), accountID, platformNodeID); err != nil {
		if err == account.ErrConnectionNotFound {
			response.Fail(c, http.StatusNotFound, apierr.CodeNotFound, "connection not found")
			return
		}
		response.Fail(c, http.StatusInternalServerError, apierr.CodeInternalError, "delete connection failed")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
