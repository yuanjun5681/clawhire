package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	appauth "github.com/yuanjun5681/clawhire/backend/internal/application/auth"
	appcmd "github.com/yuanjun5681/clawhire/backend/internal/application/command"
	"github.com/yuanjun5681/clawhire/backend/internal/application/platform"
	"github.com/yuanjun5681/clawhire/backend/internal/application/webhook"
	infraauth "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/auth"
	"github.com/yuanjun5681/clawhire/backend/internal/infrastructure/clawsynapse"
	"github.com/yuanjun5681/clawhire/backend/internal/infrastructure/config"
	"github.com/yuanjun5681/clawhire/backend/internal/infrastructure/logx"
	mgo "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo"
	"github.com/yuanjun5681/clawhire/backend/internal/infrastructure/mongo/repository"
	httpserver "github.com/yuanjun5681/clawhire/backend/internal/transport/http"
	"github.com/yuanjun5681/clawhire/backend/internal/transport/http/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log := logx.New(cfg.LogLevel)
	log.WithFields(map[string]interface{}{
		"appEnv": cfg.AppEnv,
		"port":   cfg.HTTPPort,
	}).Info("clawhire-api starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mc, err := mgo.NewClient(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		log.WithError(err).Fatal("mongo connect failed")
	}
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = mc.Close(shutdownCtx)
	}()

	indexCtx, indexCancel := context.WithTimeout(ctx, 30*time.Second)
	if err := mgo.EnsureIndexes(indexCtx, mc.DB()); err != nil {
		indexCancel()
		log.WithError(err).Fatal("ensure mongo indexes failed")
	}
	indexCancel()
	log.Info("mongo indexes ensured")

	// --- 仓储装配 ---
	rawRepo := repository.NewRawEventRepo(mc.DB())
	domainEventRepo := repository.NewDomainEventRepo(mc.DB())
	platformConnRepo := repository.NewPlatformConnectionRepo(mc.DB())
	taskRepo := repository.NewTaskRepo(mc.DB())
	bidRepo := repository.NewBidRepo(mc.DB())
	contractRepo := repository.NewContractRepo(mc.DB())
	progressRepo := repository.NewProgressRepo(mc.DB())
	milestoneRepo := repository.NewMilestoneRepo(mc.DB())
	submissionRepo := repository.NewSubmissionRepo(mc.DB())
	reviewRepo := repository.NewReviewRepo(mc.DB())
	settlementRepo := repository.NewSettlementRepo(mc.DB())
	accountRepo := repository.NewAccountRepo(mc.DB())

	// --- ClawSynapse 跨平台同步装配（NodeAPIURL 为空时禁用）---
	var syncPub *platform.SyncPublisher
	if cfg.ClawSynapse.NodeAPIURL != "" {
		synapseClient := clawsynapse.NewClient(cfg.ClawSynapse.NodeAPIURL)
		syncPub = platform.NewSyncPublisher(platformConnRepo, synapseClient, log)
		log.WithFields(map[string]interface{}{
			"nodeAPIURL":             cfg.ClawSynapse.NodeAPIURL,
			"defaultTrustMeshNodeID": cfg.ClawSynapse.DefaultTrustMeshNodeID,
		}).Info("clawsynapse sync publisher enabled")
	}

	// --- 应用层装配 ---
	commandSvc := appcmd.NewService(appcmd.Options{
		Tasks:       taskRepo,
		Bids:        bidRepo,
		Contracts:   contractRepo,
		Submissions: submissionRepo,
		Reviews:     reviewRepo,
		DomainEvts:  domainEventRepo,
		SyncPub:     syncPub,
	})
	dispatcher := webhook.NewCommandDispatcher(webhook.CommandDispatcherOptions{
		Tasks:       taskRepo,
		Bids:        bidRepo,
		Contracts:   contractRepo,
		Progress:    progressRepo,
		Milestones:  milestoneRepo,
		Submissions: submissionRepo,
		Reviews:     reviewRepo,
		Settlements: settlementRepo,
		Accounts:    accountRepo,
		DomainEvts:  domainEventRepo,
		Commands:    commandSvc,
	})
	webhookSvc := webhook.NewService(webhook.Options{
		RawRepo:    rawRepo,
		Dispatcher: dispatcher,
	})

	// --- Auth 装配 ---
	jwtIssuer := infraauth.NewJWTIssuer(cfg.Auth.JWTSecret, cfg.Auth.JWTTTL, cfg.Auth.JWTIssuer)
	authSvc := appauth.NewService(appauth.Options{
		Accounts:    accountRepo,
		Issuer:      jwtIssuer,
		BcryptCost:  cfg.Auth.BcryptCost,
		MinPassword: cfg.Auth.MinPassword,
	})

	// --- HTTP 服务器 ---
	srv := httpserver.NewServer(httpserver.Options{
		Port:   cfg.HTTPPort,
		AppEnv: cfg.AppEnv,
		Log:    log,
	})
	defaultNodes := map[string]string{}
	if cfg.ClawSynapse.DefaultTrustMeshNodeID != "" {
		defaultNodes["trustmesh"] = cfg.ClawSynapse.DefaultTrustMeshNodeID
	}

	httpserver.RegisterRoutes(srv.Engine(), httpserver.Deps{
		Log:             log,
		Health:          handler.NewHealth(mc),
		ClawSynapseHook: handler.NewClawSynapseWebhook(webhookSvc, log),
		Write:           handler.NewWrite(commandSvc, accountRepo),
		Auth:            handler.NewAuth(authSvc),
		Connections:     handler.NewConnections(platformConnRepo, defaultNodes),
		JWTIssuer:       jwtIssuer,
		Query: handler.NewQuery(
			taskRepo,
			bidRepo,
			progressRepo,
			milestoneRepo,
			submissionRepo,
			reviewRepo,
			settlementRepo,
			accountRepo,
		),
	})

	go func() {
		if err := srv.Run(); err != nil {
			log.WithError(err).Fatal("http server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutdown signal received")

	shutdownCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("http server shutdown error")
	}
	log.Info("clawhire-api stopped")
}
