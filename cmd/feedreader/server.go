package feedreader

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abatilo/amanuensis/internal/db"
	"gorm.io/gorm"
)

type Server struct {
	ctx        context.Context
	logger     *slog.Logger
	db         *gorm.DB
	mux        *http.ServeMux
	srv        *http.Server
	httpClient *http.Client
}

func NewServer(ctx context.Context, cfg Config) *Server {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	c := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &Server{
		ctx:        ctx,
		logger:     cfg.logger,
		db:         cfg.db,
		mux:        mux,
		srv:        srv,
		httpClient: c,
	}
}

func (s *Server) Start() error {
	done := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		_ = s.Stop()
		close(done)
	}()

	err := s.db.AutoMigrate(
		&db.Feed{},
		&db.Validated{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	s.prepareRoutes()
	s.logger.Info("starting server", "addr", s.srv.Addr)
	err = s.srv.ListenAndServe()
	<-done
	return err
}

func (s *Server) Stop() error {
	s.logger.Info("shutting down server")
	return s.srv.Shutdown(s.ctx)
}
