package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	logger *slog.Logger
	router chi.Router
	server *http.Server
}

func NewServer(logger *slog.Logger) *Server {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	return &Server{
		logger: logger,
		router: r,
	}
}

func (s *Server) Start() error {
	if s.server != nil {
		return errors.New("server already started")
	}

	s.server = &http.Server{
		Addr:              ":8080",
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.logger.Info("starting HTTP server", slog.String("addr", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}
