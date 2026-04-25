// Package clawsynapse 提供访问 ClawSynapse 节点 HTTP API 的客户端。
package clawsynapse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client 是 ClawSynapse 节点 HTTP API 客户端。
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(nodeAPIURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(nodeAPIURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// apiResponse 是 ClawSynapse 统一响应信封。
type apiResponse struct {
	OK      bool            `json:"ok"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// PublishRequest 对应 POST /v1/publish 请求体。
type PublishRequest struct {
	TargetNode string                 `json:"targetNode"`
	Type       string                 `json:"type,omitempty"`
	AgentID    string                 `json:"agentId,omitempty"`
	Message    string                 `json:"message"`
	SessionKey string                 `json:"sessionKey,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PublishResult 对应成功发布后 data 字段。
type PublishResult struct {
	TargetNode string `json:"targetNode"`
	MessageID  string `json:"messageId"`
	SessionKey string `json:"sessionKey"`
}

// Publish 调用 POST /v1/publish 向目标节点发布消息。
func (c *Client) Publish(ctx context.Context, req PublishRequest) (*PublishResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal publish request: %w", err)
	}
	resp, err := c.post(ctx, "/v1/publish", body)
	if err != nil {
		return nil, err
	}
	var result PublishResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("decode publish result: %w", err)
	}
	return &result, nil
}

// healthData 仅解析 /v1/health 中需要的字段。
type healthData struct {
	Self struct {
		NodeID string `json:"nodeId"`
	} `json:"self"`
}

// GetSelfNodeID 通过 GET /v1/health 获取本节点 nodeId。
func (c *Client) GetSelfNodeID(ctx context.Context) (string, error) {
	resp, err := c.get(ctx, "/v1/health")
	if err != nil {
		return "", err
	}
	var data healthData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("decode health data: %w", err)
	}
	if data.Self.NodeID == "" {
		return "", fmt.Errorf("empty nodeId from health response")
	}
	return data.Self.NodeID, nil
}

func (c *Client) get(ctx context.Context, path string) (*apiResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	return c.do(req)
}

func (c *Client) post(ctx context.Context, path string, body []byte) (*apiResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}

func (c *Client) do(req *http.Request) (*apiResponse, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clawsynapse http: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read clawsynapse response: %w", err)
	}
	var resp apiResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode clawsynapse response: %w", err)
	}
	if !resp.OK {
		return nil, fmt.Errorf("clawsynapse error %s: %s", resp.Code, resp.Message)
	}
	return &resp, nil
}
