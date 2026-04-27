package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- fake PlatformConnectionRepository ---

type fakeConnHandlerRepo struct {
	items []*account.PlatformConnection
}

func (r *fakeConnHandlerRepo) Insert(_ context.Context, conn *account.PlatformConnection) error {
	for _, c := range r.items {
		if c.LocalUserID == conn.LocalUserID && c.PlatformNodeID == conn.PlatformNodeID {
			return account.ErrConnectionExists
		}
	}
	cp := *conn
	r.items = append(r.items, &cp)
	return nil
}

func (r *fakeConnHandlerRepo) UpsertByLocalUserAndNode(_ context.Context, conn *account.PlatformConnection) error {
	for _, c := range r.items {
		if c.LocalUserID == conn.LocalUserID && c.PlatformNodeID == conn.PlatformNodeID {
			c.Platform = conn.Platform
			c.RemoteUserID = conn.RemoteUserID
			c.LinkedAt = conn.LinkedAt
			return nil
		}
	}
	cp := *conn
	r.items = append(r.items, &cp)
	return nil
}

func (r *fakeConnHandlerRepo) FindByLocalUser(_ context.Context, localUserID, platform string) ([]*account.PlatformConnection, error) {
	var out []*account.PlatformConnection
	for _, c := range r.items {
		if c.LocalUserID != localUserID {
			continue
		}
		if platform != "" && c.Platform != platform {
			continue
		}
		cp := *c
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeConnHandlerRepo) FindByRemote(_ context.Context, platformNodeID, remoteUserID string) (*account.PlatformConnection, error) {
	return nil, account.ErrConnectionNotFound
}

func (r *fakeConnHandlerRepo) DeleteByLocalUserAndNode(_ context.Context, localUserID, platformNodeID string) error {
	for i, c := range r.items {
		if c.LocalUserID == localUserID && c.PlatformNodeID == platformNodeID {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return account.ErrConnectionNotFound
}

// --- helpers ---

func newConnectionsEngine(repo *fakeConnHandlerRepo, defaults map[string]string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewConnections(repo, defaults)
	e := gin.New()
	e.Use(testAuthStub())
	e.GET("/api/accounts/me/connections", h.ListConnections)
	e.GET("/api/accounts/me/connections/trustmesh/connect-url", h.TrustMeshConnectURL)
	e.POST("/api/accounts/me/connections", h.CreateConnection)
	e.DELETE("/api/accounts/me/connections/:platform", h.DeleteConnection)
	return e
}

func newConnectionsEngineWithOptions(repo *fakeConnHandlerRepo, defaults map[string]string, opts ...ConnectionOption) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewConnections(repo, defaults, opts...)
	e := gin.New()
	e.Use(testAuthStub())
	e.GET("/api/accounts/me/connections/trustmesh/connect-url", h.TrustMeshConnectURL)
	return e
}

type fakeSelfNodeProvider struct {
	nodeID string
	err    error
}

func (p fakeSelfNodeProvider) GetSelfNodeID(_ context.Context) (string, error) {
	return p.nodeID, p.err
}

func defaultNodes() map[string]string {
	return map[string]string{"trustmesh": "node_trustmesh_prod"}
}

// --- tests ---

func TestConnections_List_EmptyReturnsArray(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngine(repo, defaultNodes())

	req := httptest.NewRequest(http.MethodGet, "/api/accounts/me/connections", nil)
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data []interface{} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Data == nil {
		t.Fatal("data should be empty array, not null")
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 items, got %d", len(resp.Data))
	}
}

func TestConnections_TrustMeshConnectURL_ReturnsAuthorizedLink(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngineWithOptions(
		repo,
		defaultNodes(),
		WithTrustMeshConnect("https://trustmesh.example.com/", fakeSelfNodeProvider{nodeID: "node_clawhire"}),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/accounts/me/connections/trustmesh/connect-url", nil)
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data trustMeshConnectURLResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	u, err := url.Parse(resp.Data.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	if u.Scheme != "https" || u.Host != "trustmesh.example.com" || u.Path != "/connect" {
		t.Fatalf("unexpected connect url: %s", resp.Data.URL)
	}
	q := u.Query()
	if q.Get("platform") != "clawhire" {
		t.Errorf("platform = %q", q.Get("platform"))
	}
	if q.Get("platform_node_id") != "node_clawhire" {
		t.Errorf("platform_node_id = %q", q.Get("platform_node_id"))
	}
	if q.Get("remote_user_id") != "acct_alice" {
		t.Errorf("remote_user_id = %q", q.Get("remote_user_id"))
	}
}

func TestConnections_TrustMeshConnectURL_NotConfiguredReturns503(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngineWithOptions(repo, defaultNodes())

	req := httptest.NewRequest(http.MethodGet, "/api/accounts/me/connections/trustmesh/connect-url", nil)
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConnections_Create_UsesDefaultNodeID(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngine(repo, defaultNodes())

	body := `{"platform":"trustmesh","remoteUserId":"usr_xxxx"}`
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/me/connections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(repo.items))
	}
	if repo.items[0].PlatformNodeID != "node_trustmesh_prod" {
		t.Errorf("platformNodeId = %q", repo.items[0].PlatformNodeID)
	}
	if repo.items[0].LocalUserID != "acct_alice" {
		t.Errorf("localUserId = %q", repo.items[0].LocalUserID)
	}
	if repo.items[0].RemoteUserID != "usr_xxxx" {
		t.Errorf("remoteUserId = %q", repo.items[0].RemoteUserID)
	}
}

func TestConnections_Create_ExplicitNodeIDOverridesDefault(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngine(repo, defaultNodes())

	body := `{"platform":"trustmesh","remoteUserId":"usr_yyyy","platformNodeId":"node_trustmesh_staging"}`
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/me/connections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if repo.items[0].PlatformNodeID != "node_trustmesh_staging" {
		t.Errorf("platformNodeId = %q", repo.items[0].PlatformNodeID)
	}
}

func TestConnections_Create_DuplicateReturns409(t *testing.T) {
	repo := &fakeConnHandlerRepo{
		items: []*account.PlatformConnection{
			{
				ID:             bson.NewObjectID(),
				Platform:       "trustmesh",
				PlatformNodeID: "node_trustmesh_prod",
				LocalUserID:    "acct_alice",
				RemoteUserID:   "usr_xxxx",
			},
		},
	}
	e := newConnectionsEngine(repo, defaultNodes())

	body := `{"platform":"trustmesh","remoteUserId":"usr_xxxx"}`
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/me/connections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConnections_Create_MissingPlatformReturns400(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngine(repo, defaultNodes())

	body := `{"remoteUserId":"usr_xxxx"}`
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/me/connections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConnections_Create_UnknownPlatformNoDefaultReturns400(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngine(repo, map[string]string{}) // 无任何默认节点

	body := `{"platform":"unknownplatform","remoteUserId":"usr_xxxx"}`
	req := httptest.NewRequest(http.MethodPost, "/api/accounts/me/connections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConnections_Delete_ExistingReturns200(t *testing.T) {
	repo := &fakeConnHandlerRepo{
		items: []*account.PlatformConnection{
			{
				ID:             bson.NewObjectID(),
				Platform:       "trustmesh",
				PlatformNodeID: "node_trustmesh_prod",
				LocalUserID:    "acct_alice",
				RemoteUserID:   "usr_xxxx",
			},
		},
	}
	e := newConnectionsEngine(repo, defaultNodes())

	req := httptest.NewRequest(http.MethodDelete, "/api/accounts/me/connections/trustmesh", nil)
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if len(repo.items) != 0 {
		t.Fatalf("expected 0 connections after delete, got %d", len(repo.items))
	}
}

func TestConnections_Delete_NotFoundReturns404(t *testing.T) {
	repo := &fakeConnHandlerRepo{}
	e := newConnectionsEngine(repo, defaultNodes())

	req := httptest.NewRequest(http.MethodDelete, "/api/accounts/me/connections/trustmesh", nil)
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestConnections_List_FiltersOtherUsers(t *testing.T) {
	repo := &fakeConnHandlerRepo{
		items: []*account.PlatformConnection{
			{ID: bson.NewObjectID(), Platform: "trustmesh", PlatformNodeID: "node_trustmesh_prod", LocalUserID: "acct_alice", RemoteUserID: "usr_a"},
			{ID: bson.NewObjectID(), Platform: "trustmesh", PlatformNodeID: "node_trustmesh_prod", LocalUserID: "acct_bob", RemoteUserID: "usr_b"},
		},
	}
	e := newConnectionsEngine(repo, defaultNodes())

	req := httptest.NewRequest(http.MethodGet, "/api/accounts/me/connections", nil)
	req.Header.Set(testAccountHeader, "acct_alice")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var resp struct {
		Data []account.PlatformConnection `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 item for alice, got %d", len(resp.Data))
	}
	if resp.Data[0].RemoteUserID != "usr_a" {
		t.Errorf("got wrong connection: remoteUserId = %q", resp.Data[0].RemoteUserID)
	}
}
