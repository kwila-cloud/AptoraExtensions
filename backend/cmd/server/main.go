package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kwila-cloud/aptora-extensions/backend/internal/server"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	srv := server.NewServer(logger)
	ctx := context.Background()
	if err := srv.Run(ctx, ":8080"); err != nil {
		logger.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
