package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"redis-lite/internal/server"
	"redis-lite/pkg/aof"
	"redis-lite/pkg/cfg"
	"redis-lite/pkg/core"
	"redis-lite/pkg/database"
	"strings"
	"syscall"
)

func main() {
	config := cfg.NewConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	db := database.NewStore()

	aofHandler, err := aof.NewAof(config)
	if err != nil {
		panic("failed to create AOF")
	}

	slog.Info("Restoring data from AOF...")
	aofHandler.Read(func(cmd string) {
		cmd = strings.TrimSpace(cmd)
		args := strings.Fields(cmd)
		if len(args) == 0 {
			return
		}
		core.Eval(db, args)
	})
	slog.Info("Data restoration complete.")

	jntr := database.NewJanitor(config)
	go jntr.Run(db)

	srv := server.NewServer(config.Host, config.Port, db, aofHandler)

	if err := srv.Run(ctx); err != nil {
		slog.ErrorContext(ctx, "server exited properly", "error", err)
	}
}
