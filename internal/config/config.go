package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// Config holds runtime configuration read from the environment.
type Config struct {
	Addr          string
	DBPath        string
	LogLevel      slog.Level
	BaseURL       string
	RetentionDays int
	SecretKey     string
	CheckWorkers  int
}

// Load reads configuration from environment variables, applying defaults.
func Load() (Config, error) {
	c := Config{
		Addr:          env("BB_ADDR", ":8080"),
		DBPath:        env("BB_DB_PATH", "./data/big-brother.db"),
		BaseURL:       env("BB_BASE_URL", "http://localhost:8080"),
		RetentionDays: 30,
		SecretKey:     os.Getenv("BB_SECRET_KEY"),
		CheckWorkers:  16,
	}

	if v := os.Getenv("BB_RETENTION_DAYS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return c, fmt.Errorf("BB_RETENTION_DAYS must be a positive integer, got %q", v)
		}
		c.RetentionDays = n
	}

	if v := os.Getenv("BB_CHECK_WORKERS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return c, fmt.Errorf("BB_CHECK_WORKERS must be a positive integer, got %q", v)
		}
		c.CheckWorkers = n
	}

	level, err := parseLevel(env("BB_LOG_LEVEL", "info"))
	if err != nil {
		return c, err
	}
	c.LogLevel = level

	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid BB_LOG_LEVEL %q", s)
	}
}
