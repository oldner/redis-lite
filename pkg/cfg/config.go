package cfg

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ServerType string

const (
	TCP ServerType = "tcp"
)

type Config struct {
	Host            string
	Port            string
	ServerType      ServerType
	JanitorInterval time.Duration
	AofPath         string
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found")
	}
	return &Config{
		Host:            getEnv("HOST", "localhost"),
		Port:            getEnv("PORT", "6379"),
		ServerType:      ServerType(getEnv("SERVER", "tcp")),
		JanitorInterval: getEnvDuration("JANITOR_INTERVAL", time.Minute),
		AofPath:         getEnv("AOF_PATH", "aof"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
