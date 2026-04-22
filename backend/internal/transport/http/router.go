package http

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/handler"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/middleware"
)

type Deps struct {
	Log             *logrus.Logger
	Health          *handler.Health
	ClawSynapseHook *handler.ClawSynapseWebhook
	Query           *handler.Query
}

func RegisterRoutes(e *gin.Engine, d Deps) {
	e.Use(middleware.RequestID())
	e.Use(middleware.Recovery(d.Log))
	e.Use(middleware.Logging(d.Log))

	e.GET("/healthz", d.Health.Live)
	e.GET("/readyz", d.Health.Ready)

	api := e.Group("/api")
	if d.Query != nil {
		api.GET("/tasks", d.Query.ListTasks)
		api.GET("/tasks/:taskId", d.Query.GetTask)
		api.GET("/tasks/:taskId/bids", d.Query.ListTaskBids)
		api.GET("/tasks/:taskId/progress", d.Query.ListTaskProgress)
		api.GET("/tasks/:taskId/milestones", d.Query.ListTaskMilestones)
		api.GET("/tasks/:taskId/submissions", d.Query.ListTaskSubmissions)
		api.GET("/tasks/:taskId/reviews", d.Query.ListTaskReviews)
		api.GET("/tasks/:taskId/settlements", d.Query.ListTaskSettlements)
		api.GET("/accounts", d.Query.ListAccounts)
		api.GET("/executors/:executorId/history", d.Query.ExecutorHistory)
	}

	webhooks := e.Group("/webhooks")
	if d.ClawSynapseHook != nil {
		webhooks.POST("/clawsynapse", d.ClawSynapseHook.Handle)
	}
}
