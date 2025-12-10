package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"redis-lite/internal/server"
	"redis-lite/pkg/cfg"
	"redis-lite/pkg/database"
	"syscall"
)

func main() {
	config := cfg.NewConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db := database.NewStore()

	srv := server.NewServer(config.Host, config.Port, db)

	if err := srv.Run(ctx); err != nil {
		slog.ErrorContext(ctx, "server exited properly", err)
	}
}
