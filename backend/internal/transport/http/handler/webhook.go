package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/yuanjun5681/clawhire/backend/internal/application/webhook"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/response"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/middleware"
)

// ClawSynapseWebhook 处理 POST /webhooks/clawsynapse。
type ClawSynapseWebhook struct {
	svc *webhook.Service
	log *logrus.Logger
}

// NewClawSynapseWebhook 构造 handler。
func NewClawSynapseWebhook(svc *webhook.Service, log *logrus.Logger) *ClawSynapseWebhook {
	return &ClawSynapseWebhook{svc: svc, log: log}
}

// Handle 是 Gin 入口。
//
//  1. 读取原始 body（用于原文存档）
//  2. 反序列化 ClawSynapse Envelope
//  3. 交给 webhook.Service 处理
//  4. 按 apierr/response 规范返回
func (h *ClawSynapseWebhook) Handle(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "read request body failed")
		return
	}
	// 还原 body，便于后续 middleware / log 使用
	c.Request.Body = io.NopCloser(bytes.NewReader(rawBody))

	var env clawsynapse.Envelope
	if err := json.Unmarshal(rawBody, &env); err != nil {
		response.Fail(c, http.StatusBadRequest, apierr.CodeInvalidRequest, "invalid JSON body")
		return
	}

	headers := collectAuditHeaders(c)
	res, err := h.svc.Ingest(c.Request.Context(), &env, rawBody, headers)

	logEntry := h.logEntry(c).WithFields(logrus.Fields{
		"messageType": env.Type,
	})
	if res != nil {
		logEntry = logEntry.WithField("eventKey", res.EventKey)
	}

	if err != nil {
		if ae, ok := apierr.As(err); ok {
			if ae.Code == apierr.CodeDuplicateEvent && res != nil {
				// 409 + 原处理状态
				c.JSON(http.StatusConflict, response.Success{
					Success: false,
					Data:    res,
				})
				logEntry.Info("webhook duplicate event")
				return
			}
			logEntry.WithError(err).Warn("webhook ingest rejected")
			response.Fail(c, ae.HTTPStatus, ae.Code, ae.Message)
			return
		}
		logEntry.WithError(err).Error("webhook ingest failed")
		response.FailErr(c, err)
		return
	}

	logEntry.WithField("status", res.Status).Info("webhook accepted")
	response.OK(c, res)
}

// logEntry 带上 requestId 便于链路追踪。
func (h *ClawSynapseWebhook) logEntry(c *gin.Context) *logrus.Entry {
	rid, _ := c.Get(middleware.CtxKeyRequestID)
	return h.log.WithField("requestId", rid)
}

// collectAuditHeaders 选取需要存档的请求头。
//
// 只保留对审计有价值、不含敏感业务数据的头：
//   - X-Request-ID
//   - X-ClawSynapse-* （若上游后续扩展头字段，可一并归档）
//   - Content-Type
func collectAuditHeaders(c *gin.Context) map[string]string {
	out := map[string]string{}
	for k, v := range c.Request.Header {
		if len(v) == 0 {
			continue
		}
		kl := strings.ToLower(k)
		switch {
		case kl == "content-type":
		case kl == "x-request-id":
		case strings.HasPrefix(kl, "x-clawsynapse-"):
		default:
			continue
		}
		out[k] = v[0]
	}
	return out
}
