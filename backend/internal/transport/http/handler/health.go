package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
)

type Health struct {
	mongo *mgo.Client
}

func NewHealth(m *mgo.Client) *Health {
	return &Health{mongo: m}
}

func (h *Health) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Health) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.mongo.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"checks": gin.H{"mongo": err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{"mongo": "ok"},
	})
}
