package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

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

	rows, err := db.QueryContext(ctx, "SELECT id, Name FROM Employees WHERE inactive = 0")
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
