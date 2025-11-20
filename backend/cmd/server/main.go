package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwila-cloud/aptora-extensions/backend/internal/server"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	srv := server.NewServer(logger)

	// Create context that can be cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx, ":8080"); err != nil {
		logger.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
