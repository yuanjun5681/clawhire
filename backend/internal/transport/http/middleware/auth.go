package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	infraauth "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/auth"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/response"
)

const (
	CtxKeyAccountID   = "accountId"
	CtxKeyAccountType = "accountType"
)

// Auth 校验 Authorization: Bearer <jwt> 并注入当前账号到 gin.Context。
// 路由层对免登录接口（/healthz、/webhooks/*、/api/auth/*）不应挂载此中间件。
func Auth(issuer *infraauth.JWTIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(c.GetHeader("Authorization"))
		if raw == "" {
			response.Fail(c, 401, apierr.CodeUnauthorized, "missing authorization header")
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(raw, prefix) {
			response.Fail(c, 401, apierr.CodeUnauthorized, "invalid authorization scheme")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(raw, prefix))
		if token == "" {
			response.Fail(c, 401, apierr.CodeUnauthorized, "missing bearer token")
			return
		}
		claims, err := issuer.Verify(token)
		if err != nil {
			response.Fail(c, 401, apierr.CodeUnauthorized, "invalid or expired token")
			return
		}
		c.Set(CtxKeyAccountID, claims.AccountID)
		c.Set(CtxKeyAccountType, claims.AccountType)
		c.Next()
	}
}

// CurrentAccountID 从 context 读出已认证的账号 ID。若未设置返回空串。
func CurrentAccountID(c *gin.Context) string {
	v, _ := c.Get(CtxKeyAccountID)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
