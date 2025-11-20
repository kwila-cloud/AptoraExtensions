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
	logger     *slog.Logger
	router     chi.Router
	httpServer *http.Server
}

func NewServer(logger *slog.Logger) *Server {
	r := chi.NewRouter()
	registerRoutes(r)

	return &Server{
		logger: logger,
		router: r,
	}
}

func registerRoutes(r chi.Router) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}

func (s *Server) Run(ctx context.Context, addr string) error {
	if s.httpServer != nil {
		return errors.New("server already running")
	}

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("http server listening", slog.String("addr", addr))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}
