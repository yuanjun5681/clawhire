package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/response"
)

func Recovery(log *logrus.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		rid, _ := c.Get(CtxKeyRequestID)
		log.WithFields(logrus.Fields{
			"requestId": rid,
			"panic":     err,
			"path":      c.Request.URL.Path,
		}).Error("panic recovered")
		response.Fail(c, http.StatusInternalServerError, apierr.CodeInternalError, "internal server error")
	})
}
