package http

import (
	"context"
	"fmt"
	nhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Options struct {
	Port   int
	AppEnv string
	Log    *logrus.Logger
}

type Server struct {
	engine *gin.Engine
	srv    *nhttp.Server
	log    *logrus.Logger
}

func NewServer(opts Options) *Server {
	if opts.AppEnv != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.New()
	s := &Server{engine: e, log: opts.Log}
	s.srv = &nhttp.Server{
		Addr:              fmt.Sprintf(":%d", opts.Port),
		Handler:           e,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s
}

func (s *Server) Engine() *gin.Engine { return s.engine }

func (s *Server) Run() error {
	s.log.WithField("addr", s.srv.Addr).Info("http server listening")
	if err := s.srv.ListenAndServe(); err != nil && err != nhttp.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
