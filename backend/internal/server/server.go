package server

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

//go:embed all:built-frontend
var staticFiles embed.FS

type Server struct {
	logger     *slog.Logger
	router     chi.Router
	httpServer *http.Server
	devMode    bool
}

func NewServer(logger *slog.Logger, devMode bool) *Server {
	s := &Server{
		logger:  logger,
		router:  chi.NewRouter(),
		devMode: devMode,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			s.logger.Error("failed to write health response", slog.Any("error", err))
		}
	})

	if s.devMode {
		s.router.HandleFunc("/*", s.proxyToVite)
	} else {
		s.router.HandleFunc("/*", s.serveAssets)
	}
}

func (s *Server) serveAssets(w http.ResponseWriter, r *http.Request) {
	// Strip the "built-frontend" prefix from the embedded filesystem
	sub, err := fs.Sub(staticFiles, "built-frontend")
	if err != nil {
		s.logger.Error("failed to create sub filesystem", "err", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If path maps to a real file in the embedded fs, serve it.
	// Otherwise, serve index.html (SPA fallback).
	requestPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if requestPath == "" {
		requestPath = "."
	}

	// Try to open the requested file
	if f, err := sub.Open(requestPath); err == nil {
		stat, err := f.Stat()
		f.Close()

		// If it's a file (not a directory), serve it
		if err == nil && !stat.IsDir() {
			fileServer := http.FileServer(http.FS(sub))
			fileServer.ServeHTTP(w, r)
			return
		}
	}

	// Not found or is a directory — return index.html for client-side routing
	data, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		s.logger.Error("failed to find index.html", slog.Any("error", err))
		http.Error(w, "404 - Page Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(data); err != nil {
		s.logger.Error("failed to write index.html response", slog.Any("error", err))
	}
}

func (s *Server) proxyToVite(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "localhost:5173",
	})

	proxy.ServeHTTP(w, r)
}

// Run starts the HTTP server and blocks until the provided context is cancelled
// or the server exits with an error. Graceful shutdown is handled when
// the context is cancelled.
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
		s.logger.Info("http server listening", slog.String("addr", addr), slog.Bool("dev_mode", s.devMode))
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
