package logger

import (
	"log/slog"
	"os"
)

// InitLogger sets up the global structured logger (JSON format)
func InitLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// set as global default
	slog.SetDefault(logger)
}
