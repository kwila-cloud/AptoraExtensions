package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwila-cloud/aptora-extensions/backend/internal/server"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	devMode := flag.Bool("dev", false, "Enable development mode (proxy to Vite dev server)")
	flag.Parse()

	srv := server.NewServer(logger, *devMode)

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
