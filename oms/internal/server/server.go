package server

import (
	"context"
	"log/slog"
	"net/http"
	"oms/internal/handler"
	"oms/internal/logger"
	"oms/internal/repository"
	"oms/internal/routes"
	"oms/internal/service"
	"time"
)

type Server struct {
	server *http.Server
	logger *slog.Logger
}

func New(port string, db *repository.Repository) *Server {
	// Initialize logger
	l := logger.Init()

	// Initialize service
	s := service.New(l, db)

	// Initialize handler
	h := handler.New(l, s)

	// Initialize router
	router := routes.RegisterRoutes(h)

	// Create server
	srv := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		server: srv,
		logger: l,
	}
}

func (s *Server) Start() error {
	s.logger.Info("starting server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	return s.server.Shutdown(ctx)
}
