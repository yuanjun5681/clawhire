package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

type Meta struct {
	Page     int   `json:"page,omitempty"`
	PageSize int   `json:"pageSize,omitempty"`
	Total    int64 `json:"total,omitempty"`
}

type Success struct {
	Success bool  `json:"success"`
	Data    any   `json:"data,omitempty"`
	Meta    *Meta `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Failure struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Success{Success: true, Data: data})
}

func OKMeta(c *gin.Context, data any, meta *Meta) {
	c.JSON(http.StatusOK, Success{Success: true, Data: data, Meta: meta})
}

func Fail(c *gin.Context, status int, code apierr.Code, msg string) {
	c.AbortWithStatusJSON(status, Failure{
		Success: false,
		Error:   ErrorBody{Code: string(code), Message: msg},
	})
}

func FailErr(c *gin.Context, err error) {
	if e, ok := apierr.As(err); ok {
		Fail(c, e.HTTPStatus, e.Code, e.Message)
		return
	}
	Fail(c, http.StatusInternalServerError, apierr.CodeInternalError, "internal server error")
}
