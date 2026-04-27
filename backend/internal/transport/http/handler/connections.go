package handler

import (
	"context"
	"net/http"
	"net/url"
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
	trustMeshWebURL         string
	selfNodeIDProvider      selfNodeIDProvider
}

type selfNodeIDProvider interface {
	GetSelfNodeID(ctx context.Context) (string, error)
}

type ConnectionOption func(*Connections)

func WithTrustMeshConnect(webURL string, provider selfNodeIDProvider) ConnectionOption {
	return func(h *Connections) {
		h.trustMeshWebURL = strings.TrimRight(strings.TrimSpace(webURL), "/")
		h.selfNodeIDProvider = provider
	}
}

func NewConnections(repo account.PlatformConnectionRepository, defaultNodes map[string]string, opts ...ConnectionOption) *Connections {
	if defaultNodes == nil {
		defaultNodes = map[string]string{}
	}
	h := &Connections{connections: repo, defaultNodeIDByPlatform: defaultNodes}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type createConnectionRequest struct {
	Platform       string `json:"platform"`
	RemoteUserID   string `json:"remoteUserId"`
	PlatformNodeID string `json:"platformNodeId,omitempty"`
}

type trustMeshConnectURLResponse struct {
	URL string `json:"url"`
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

// TrustMeshConnectURL GET /api/accounts/me/connections/trustmesh/connect-url
func (h *Connections) TrustMeshConnectURL(c *gin.Context) {
	accountID := middleware.CurrentAccountID(c)
	if h.trustMeshWebURL == "" || h.selfNodeIDProvider == nil {
		response.Fail(c, http.StatusServiceUnavailable, apierr.CodeInternalError, "trustmesh connect is not configured")
		return
	}

	clawhireNodeID, err := h.selfNodeIDProvider.GetSelfNodeID(c.Request.Context())
	if err != nil {
		response.Fail(c, http.StatusBadGateway, apierr.CodeInternalError, "failed to resolve clawhire node id")
		return
	}
	clawhireNodeID = strings.TrimSpace(clawhireNodeID)
	if clawhireNodeID == "" {
		response.Fail(c, http.StatusBadGateway, apierr.CodeInternalError, "empty clawhire node id")
		return
	}

	connectURL, err := url.Parse(h.trustMeshWebURL + "/connect")
	if err != nil || connectURL.Scheme == "" || connectURL.Host == "" {
		response.Fail(c, http.StatusServiceUnavailable, apierr.CodeInternalError, "invalid trustmesh web url")
		return
	}
	q := connectURL.Query()
	q.Set("platform", "clawhire")
	q.Set("platform_node_id", clawhireNodeID)
	q.Set("remote_user_id", accountID)
	connectURL.RawQuery = q.Encode()

	response.OK(c, trustMeshConnectURLResponse{URL: connectURL.String()})
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
