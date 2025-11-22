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
	Encrypt              string // "disable", "true", or "false" - controls TLS encryption
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
		"server=%s;port=%s;database=%s;user id=%s;password=%s;encrypt=%s;ApplicationIntent=ReadOnly",
		cfg.Host, cfg.Port, cfg.AptoraDBName, cfg.AptoraDBUser, cfg.AptoraDBPassword, cfg.Encrypt,
	)
	extensionsConnStr := fmt.Sprintf(
		"server=%s;port=%s;database=%s;user id=%s;password=%s;encrypt=%s",
		cfg.Host, cfg.Port, cfg.ExtensionsDBName, cfg.ExtensionsDBUser, cfg.ExtensionsDBPassword, cfg.Encrypt,
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

	// Additional safety: verify read-only mode on Aptora connection
	if err := m.verifyReadOnly(aptoraDB); err != nil {
		aptoraDB.Close()
		m.setUnhealthy(fmt.Sprintf("failed to verify read-only mode for Aptora database: %v", err))
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

	// Initialize Extensions database schema and health check
	if err := m.initializeExtensionsSchema(extensionsDB, cfg.ExtensionsDBName); err != nil {
		aptoraDB.Close()
		extensionsDB.Close()
		m.setUnhealthy(fmt.Sprintf("failed to initialize Extensions schema: %v", err))
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

// initializeExtensionsSchema creates the health_check table if it doesn't exist
// and inserts a test row. It verifies the database name to ensure we only
// run schema initialization on the Extensions database (never on Aptora).
func (m *Manager) initializeExtensionsSchema(db *sql.DB, expectedDBName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify we're connected to the Extensions database
	var actualDBName string
	if err := db.QueryRowContext(ctx, "SELECT DB_NAME()").Scan(&actualDBName); err != nil {
		return fmt.Errorf("failed to verify database name: %w", err)
	}

	if actualDBName != expectedDBName {
		return fmt.Errorf("safety check failed: expected Extensions database %q but connected to %q - refusing to initialize schema", expectedDBName, actualDBName)
	}

	// Create health_check table if not exists
	createTableSQL := `
	IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='health_check' AND xtype='U')
	CREATE TABLE health_check (
		id INT IDENTITY(1,1) PRIMARY KEY,
		timestamp DATETIME2 NOT NULL
	)`

	if _, err := db.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to create health_check table: %w", err)
	}

	// Insert test row
	insertSQL := `INSERT INTO health_check (timestamp) VALUES (SYSDATETIME())`
	if _, err := db.ExecContext(ctx, insertSQL); err != nil {
		return fmt.Errorf("failed to insert health check row: %w", err)
	}

	m.logger.Info("initialized Extensions database schema", slog.String("database", actualDBName))

	return nil
}

// verifyReadOnly attempts to perform a write operation to verify the connection
// is truly read-only. This provides defense-in-depth protection.
func (m *Manager) verifyReadOnly(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to create a permanent table - this should fail on read-only connection.
	// NOTE: We must use a permanent table (not #temp) because SQL Server allows
	// all users to create temp tables in tempdb regardless of permissions.
	testSQL := `CREATE TABLE aptora_extensions_readonly_test (id INT)`
	_, err := db.ExecContext(ctx, testSQL)

	if err != nil {
		// Good! Write operation failed, connection is read-only
		m.logger.Info("verified read-only access to Aptora database")
		return nil
	}

	// Bad! Write succeeded when it shouldn't have - clean up and error
	_, _ = db.ExecContext(ctx, `DROP TABLE aptora_extensions_readonly_test`)
	return fmt.Errorf("connection is NOT read-only - write operations are allowed")
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
