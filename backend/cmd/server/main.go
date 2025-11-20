package main

import (
	"log/slog"
	"os"

	"github.com/kwila-cloud/aptora-extensions/backend/internal/server"
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	srv := server.NewServer(logger)
	if err := srv.Start(); err != nil {
		logger.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	}
}
