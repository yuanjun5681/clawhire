package http

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	infraauth "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/auth"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/handler"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/middleware"
)

type Deps struct {
	Log             *logrus.Logger
	Health          *handler.Health
	ClawSynapseHook *handler.ClawSynapseWebhook
	Query           *handler.Query
	Write           *handler.Write
	Auth            *handler.Auth
	Connections     *handler.Connections
	JWTIssuer       *infraauth.JWTIssuer
}

func RegisterRoutes(e *gin.Engine, d Deps) {
	e.Use(middleware.RequestID())
	e.Use(middleware.Recovery(d.Log))
	e.Use(middleware.Logging(d.Log))

	e.GET("/healthz", d.Health.Live)
	e.GET("/readyz", d.Health.Ready)

	api := e.Group("/api")

	// 公开的认证接口（注册 / 登录）
	if d.Auth != nil {
		authGroup := api.Group("/auth")
		authGroup.POST("/register", d.Auth.Register)
		authGroup.POST("/login", d.Auth.Login)
	}

	// 以下业务接口均需 Bearer token
	authed := api.Group("")
	if d.JWTIssuer != nil {
		authed.Use(middleware.Auth(d.JWTIssuer))
	}
	if d.Write != nil {
		authed.POST("/tasks", d.Write.CreateTask)
		authed.POST("/tasks/:taskId/bids", d.Write.CreateBid)
		authed.POST("/tasks/:taskId/award", d.Write.AwardTask)
		authed.POST("/tasks/:taskId/submissions", d.Write.CreateSubmission)
		authed.POST("/tasks/:taskId/accept", d.Write.AcceptSubmission)
		authed.POST("/tasks/:taskId/reject", d.Write.RejectSubmission)
		authed.POST("/tasks/:taskId/settlements", d.Write.RecordSettlement)
	}
	if d.Connections != nil {
		authed.GET("/accounts/me/connections", d.Connections.ListConnections)
		authed.POST("/accounts/me/connections", d.Connections.CreateConnection)
		authed.DELETE("/accounts/me/connections/:platform", d.Connections.DeleteConnection)
	}
	if d.Query != nil {
		authed.GET("/tasks", d.Query.ListTasks)
		authed.GET("/tasks/:taskId", d.Query.GetTask)
		authed.GET("/tasks/:taskId/bids", d.Query.ListTaskBids)
		authed.GET("/tasks/:taskId/progress", d.Query.ListTaskProgress)
		authed.GET("/tasks/:taskId/milestones", d.Query.ListTaskMilestones)
		authed.GET("/tasks/:taskId/submissions", d.Query.ListTaskSubmissions)
		authed.GET("/tasks/:taskId/reviews", d.Query.ListTaskReviews)
		authed.GET("/tasks/:taskId/settlements", d.Query.ListTaskSettlements)
		authed.GET("/accounts", d.Query.ListAccounts)
		authed.GET("/accounts/:accountId", d.Query.GetAccount)
		authed.GET("/accounts/:accountId/agents", d.Query.ListAccountAgents)
		authed.GET("/executors/:executorId/history", d.Query.ExecutorHistory)
	}

	webhooks := e.Group("/webhooks")
	if d.ClawSynapseHook != nil {
		webhooks.POST("/clawsynapse", d.ClawSynapseHook.Handle)
	}
}
