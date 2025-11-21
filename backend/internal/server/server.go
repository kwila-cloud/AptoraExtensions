package server

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kwila-cloud/aptora-extensions/backend/internal/database"
)

//go:embed all:built-frontend
var staticFiles embed.FS

type Server struct {
	logger     *slog.Logger
	router     chi.Router
	httpServer *http.Server
	devMode    bool
	db         *database.Manager
}

func NewServer(logger *slog.Logger, devMode bool, db *database.Manager) *Server {
	s := &Server{
		logger:  logger,
		router:  chi.NewRouter(),
		devMode: devMode,
		db:      db,
	}
	s.registerRoutes()
	return s
}

// requestLogger logs each HTTP request with structured logging
func (s *Server) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(ww, r)

		// Log request details
		duration := time.Since(start)
		s.logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.Int("status", ww.statusCode),
			slog.Int("bytes", ww.bytesWritten),
			slog.Duration("duration", duration),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func (s *Server) registerRoutes() {
	// Add request logging middleware
	s.router.Use(s.requestLogger)

	// API routes
	s.router.Get("/health", s.handleHealth)
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/employees", s.handleEmployees)
		r.Get("/invoices", s.handleInvoices)
		// Catch-all for unmatched API routes - return 404
		r.NotFound(s.handleAPINotFound)
	})

	// Frontend routes (SPA catch-all)
	if s.devMode {
		s.router.HandleFunc("/*", s.proxyToVite)
	} else {
		s.router.HandleFunc("/*", s.serveAssets)
	}
}

func (s *Server) handleAPINotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	resp := map[string]string{"error": "API endpoint not found"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode error response", slog.Any("error", err))
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	healthy, errMsg := s.db.IsHealthy()
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		resp := map[string]string{
			"status": "unhealthy",
			"error":  errMsg,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode health response", slog.Any("error", err))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := map[string]string{"status": "healthy"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode health response", slog.Any("error", err))
	}
}

func (s *Server) handleEmployees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := s.db.AptoraDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		resp := map[string]string{"error": "database not available"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT id, Name FROM Employees WHERE DateReleased IS NULL")
	if err != nil {
		s.logger.Error("failed to query employees", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "failed to query employees"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}
	defer rows.Close()

	type Employee struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	employees := []Employee{}
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.ID, &emp.Name); err != nil {
			s.logger.Error("failed to scan employee row", slog.Any("error", err))
			continue
		}
		employees = append(employees, emp)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("error iterating employee rows", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "failed to read employees"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}

	resp := map[string][]Employee{"employees": employees}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode employees response", slog.Any("error", err))
	}
}

func (s *Server) handleInvoices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db := s.db.AptoraDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		resp := map[string]string{"error": "database not available"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}

	// Parse and validate query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	employeeIDStr := r.URL.Query().Get("employee_id")

	if startDate == "" || endDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "start_date and end_date are required (YYYY-MM-DD format)"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}

	// Build query - join with Employees to get employee name
	query := `
		SELECT i.id, i.Date, i.RepID, e.Name, i.Total 
		FROM Invoices i
		LEFT JOIN Employees e ON i.RepID = e.id
		WHERE i.Date >= @p1 AND i.Date <= @p2`
	args := []interface{}{startDate, endDate}

	if employeeIDStr != "" {
		employeeID, err := strconv.Atoi(employeeIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"error": "employee_id must be a valid integer"}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				s.logger.Error("failed to encode error response", slog.Any("error", err))
			}
			return
		}
		query += " AND i.RepID = @p3"
		args = append(args, employeeID)
	}

	// Sort by date ascending by default
	query += " ORDER BY i.Date ASC, i.id ASC"

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// First, check count to enforce 500 limit
	countQuery := `
		SELECT COUNT(*) 
		FROM Invoices i
		WHERE i.Date >= @p1 AND i.Date <= @p2`
	if employeeIDStr != "" {
		countQuery += " AND i.RepID = @p3"
	}
	var count int
	if err := db.QueryRowContext(ctx, countQuery, args...).Scan(&count); err != nil {
		s.logger.Error("failed to count invoices", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "failed to count invoices"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}

	if count > 500 {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": "query would return more than 500 invoices, please use a narrower filter"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}

	// Execute main query
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("failed to query invoices", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "failed to query invoices"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}
	defer rows.Close()

	type Invoice struct {
		ID           int     `json:"id"`
		Date         string  `json:"date"`
		EmployeeID   int     `json:"employee_id"`
		EmployeeName string  `json:"employee_name"`
		Total        float64 `json:"total"`
	}

	invoices := []Invoice{}
	for rows.Next() {
		var inv Invoice
		var date time.Time
		var employeeName sql.NullString
		if err := rows.Scan(&inv.ID, &date, &inv.EmployeeID, &employeeName, &inv.Total); err != nil {
			s.logger.Error("failed to scan invoice row", slog.Any("error", err))
			continue
		}
		inv.Date = date.Format("2006-01-02")
		if employeeName.Valid {
			inv.EmployeeName = employeeName.String
		}
		invoices = append(invoices, inv)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("error iterating invoice rows", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]string{"error": "failed to read invoices"}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode error response", slog.Any("error", err))
		}
		return
	}

	resp := map[string][]Invoice{"invoices": invoices}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("failed to encode invoices response", slog.Any("error", err))
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
