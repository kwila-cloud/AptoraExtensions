package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwila-cloud/aptora-extensions/backend/internal/config"
	"github.com/kwila-cloud/aptora-extensions/backend/internal/database"
	"github.com/kwila-cloud/aptora-extensions/backend/internal/server"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	devMode := flag.Bool("dev", false, "Enable development mode (proxy to Vite dev server)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Create database manager
	dbCfg := database.Config{
		Host:                 cfg.DBHost,
		Port:                 cfg.DBPort,
		Encrypt:              cfg.DBEncrypt,
		AptoraDBName:         cfg.AptoraDBName,
		AptoraDBUser:         cfg.AptoraDBUser,
		AptoraDBPassword:     cfg.AptoraDBPassword,
		ExtensionsDBName:     cfg.ExtensionsDBName,
		ExtensionsDBUser:     cfg.ExtensionsDBUser,
		ExtensionsDBPassword: cfg.ExtensionsDBPassword,
	}
	db := database.NewManager(logger, dbCfg)
	defer db.Close()

	srv := server.NewServer(logger, *devMode, db)

	// Create context that can be cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := "0.0.0.0:80"
	if *devMode {
		addr = "localhost:8080"
	}

	if err := srv.Run(ctx, addr); err != nil {
		logger.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
