package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// Manager handles connections to both the Aptora database (read-only)
// and the Extensions database (read-write).
type Manager struct {
	logger *slog.Logger

	aptoraDB     *sql.DB
	extensionsDB *sql.DB

	mu      sync.RWMutex
	healthy bool
	errMsg  string
}

// Config contains the database connection parameters.
type Config struct {
	Host                 string
	Port                 string
	AptoraDBName         string
	AptoraDBUser         string
	AptoraDBPassword     string
	ExtensionsDBName     string
	ExtensionsDBUser     string
	ExtensionsDBPassword string
}

// NewManager creates a new database manager and starts attempting connections.
func NewManager(logger *slog.Logger, cfg Config) *Manager {
	m := &Manager{
		logger:  logger,
		healthy: false,
	}

	// Start connection attempts in background
	go m.connectLoop(cfg)

	return m
}

// connectLoop attempts to connect to both databases, retrying every 30 seconds on failure.
func (m *Manager) connectLoop(cfg Config) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Try immediately on startup
	m.tryConnect(cfg)

	// Retry every 30 seconds if unhealthy
	for range ticker.C {
		m.mu.RLock()
		healthy := m.healthy
		m.mu.RUnlock()

		if !healthy {
			m.tryConnect(cfg)
		}
	}
}

// tryConnect attempts to establish connections to both databases.
func (m *Manager) tryConnect(cfg Config) {
	m.logger.Info("attempting to connect to databases")

	aptoraConnStr := fmt.Sprintf(
		"server=%s;port=%s;database=%s;user id=%s;password=%s;encrypt=disable",
		cfg.Host, cfg.Port, cfg.AptoraDBName, cfg.AptoraDBUser, cfg.AptoraDBPassword,
	)
	extensionsConnStr := fmt.Sprintf(
		"server=%s;port=%s;database=%s;user id=%s;password=%s;encrypt=disable",
		cfg.Host, cfg.Port, cfg.ExtensionsDBName, cfg.ExtensionsDBUser, cfg.ExtensionsDBPassword,
	)

	aptoraDB, err := sql.Open("sqlserver", aptoraConnStr)
	if err != nil {
		m.setUnhealthy(fmt.Sprintf("failed to open Aptora database: %v", err))
		return
	}

	aptoraDB.SetMaxOpenConns(10)
	aptoraDB.SetMaxIdleConns(5)
	aptoraDB.SetConnMaxLifetime(5 * time.Minute)

	if err := aptoraDB.Ping(); err != nil {
		aptoraDB.Close()
		m.setUnhealthy(fmt.Sprintf("failed to ping Aptora database: %v", err))
		return
	}

	extensionsDB, err := sql.Open("sqlserver", extensionsConnStr)
	if err != nil {
		aptoraDB.Close()
		m.setUnhealthy(fmt.Sprintf("failed to open Extensions database: %v", err))
		return
	}

	extensionsDB.SetMaxOpenConns(10)
	extensionsDB.SetMaxIdleConns(5)
	extensionsDB.SetConnMaxLifetime(5 * time.Minute)

	if err := extensionsDB.Ping(); err != nil {
		aptoraDB.Close()
		extensionsDB.Close()
		m.setUnhealthy(fmt.Sprintf("failed to ping Extensions database: %v", err))
		return
	}

	// Initialize schema and health check
	if err := m.initializeSchema(extensionsDB); err != nil {
		aptoraDB.Close()
		extensionsDB.Close()
		m.setUnhealthy(fmt.Sprintf("failed to initialize schema: %v", err))
		return
	}

	// Close old connections if they exist
	m.mu.Lock()
	if m.aptoraDB != nil {
		m.aptoraDB.Close()
	}
	if m.extensionsDB != nil {
		m.extensionsDB.Close()
	}

	m.aptoraDB = aptoraDB
	m.extensionsDB = extensionsDB
	m.healthy = true
	m.errMsg = ""
	m.mu.Unlock()

	m.logger.Info("successfully connected to databases")
}

// initializeSchema creates the health_check table if it doesn't exist
// and inserts a test row.
func (m *Manager) initializeSchema(db *sql.DB) error {
	// Create table if not exists
	createTableSQL := `
	IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='health_check' AND xtype='U')
	CREATE TABLE health_check (
		id INT IDENTITY(1,1) PRIMARY KEY,
		timestamp DATETIME2 NOT NULL
	)`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to create health_check table: %w", err)
	}

	// Insert test row
	insertSQL := `INSERT INTO health_check (timestamp) VALUES (SYSDATETIME())`
	if _, err := db.ExecContext(ctx, insertSQL); err != nil {
		return fmt.Errorf("failed to insert health check row: %w", err)
	}

	return nil
}

// setUnhealthy marks the manager as unhealthy with an error message.
func (m *Manager) setUnhealthy(errMsg string) {
	m.mu.Lock()
	m.healthy = false
	m.errMsg = errMsg
	m.mu.Unlock()

	m.logger.Error("database connection failed", slog.String("error", errMsg))
}

// IsHealthy returns true if both database connections are established.
func (m *Manager) IsHealthy() (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.healthy, m.errMsg
}

// AptoraDB returns the Aptora database connection (read-only).
// Returns nil if not connected.
func (m *Manager) AptoraDB() *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.aptoraDB
}

// ExtensionsDB returns the Extensions database connection (read-write).
// Returns nil if not connected.
func (m *Manager) ExtensionsDB() *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.extensionsDB
}

// Close closes both database connections.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	if m.aptoraDB != nil {
		if err := m.aptoraDB.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if m.extensionsDB != nil {
		if err := m.extensionsDB.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}

	return nil
}
