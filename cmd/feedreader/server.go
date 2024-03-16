package feedreader

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gorm.io/gorm"
)

type Server struct {
	ctx    context.Context
	logger *slog.Logger
	db     *gorm.DB
	mux    *http.ServeMux
	srv    *http.Server
}

func NewServer(ctx context.Context, cfg Config) *Server {
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return &Server{
		ctx:    ctx,
		logger: cfg.logger,
		db:     cfg.db,
		mux:    mux,
		srv:    srv,
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

	s.prepareRoutes()
	s.logger.Info("starting server", "addr", s.srv.Addr)
	err := s.srv.ListenAndServe()
	<-done
	return err
}

func (s *Server) Stop() error {
	s.logger.Info("shutting down server")
	return s.srv.Shutdown(s.ctx)
}
