package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

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
	employeeStr := r.URL.Query().Get("employee")

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
		SELECT i."Tran No" as Number, i."Tran Date", i."Sales Rep", i."Tran Subtotal"
		FROM aptCDV_VW_APT_InvSalCredEstList i
		WHERE i."Tran Date" >= @p1 AND i."Tran Date" <= @p2`
	args := []interface{}{startDate, endDate}

	if employeeStr != "" {
		query += ` AND i."Sales Rep" = @p3`
		args = append(args, employeeStr)
	}

	// Sort by date ascending by default
	query += ` ORDER BY i."Tran Date" ASC, i."Tran No" ASC`

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// First, check count to enforce 500 limit
	countQuery := `
		SELECT COUNT(*) 
		FROM aptCDV_VW_APT_InvSalCredEstList i
		WHERE i."Tran Date" >= @p1 AND i."Tran Date" <= @p2`
	if employeeStr != "" {
		countQuery += ` AND i."Sales Rep" = @p3`
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
		Number       int     `json:"number"`
		Date         string  `json:"date"`
		EmployeeName string  `json:"employee_name"`
		Subtotal     float64 `json:"subtotal"`
		// TODO: add total_cost
		// TODO: add gross_profit (subtotal - total_cost)
		// TODO: add gross_profit_percentage gross_profit / subtotal
		// TODO: add is_write_off
	}

	s.logger.Info("hello!")
	invoices := []Invoice{}
	for rows.Next() {
		var inv Invoice
		var date time.Time
		if err := rows.Scan(&inv.Number, &date, &inv.EmployeeName, &inv.Subtotal); err != nil {
			s.logger.Error("failed to scan invoice row", slog.Any("error", err))
			continue
		}
		inv.Date = date.Format("2006-01-02")
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
