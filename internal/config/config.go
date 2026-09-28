package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
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

	OIDCIssuer          string
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	AllowedEmailDomains []string
	BootstrapAdminEmail string
	SessionTTL          time.Duration
	CookieSecure        bool
}

// Load reads configuration from environment variables, applying defaults, and
// validates it fail-closed.
func Load() (Config, error) {
	c := Config{
		Addr:                env("BB_ADDR", ":8080"),
		DBPath:              env("BB_DB_PATH", "./data/big-brother.db"),
		BaseURL:             env("BB_BASE_URL", "http://localhost:8080"),
		RetentionDays:       30,
		SecretKey:           os.Getenv("BB_SECRET_KEY"),
		CheckWorkers:        16,
		OIDCIssuer:          env("BB_OIDC_ISSUER", "https://accounts.google.com"),
		GoogleClientID:      os.Getenv("BB_GOOGLE_CLIENT_ID"),
		GoogleClientSecret:  os.Getenv("BB_GOOGLE_CLIENT_SECRET"),
		BootstrapAdminEmail: strings.ToLower(strings.TrimSpace(os.Getenv("BB_BOOTSTRAP_ADMIN_EMAIL"))),
		AllowedEmailDomains: splitCSV(os.Getenv("BB_ALLOWED_EMAIL_DOMAINS")),
		SessionTTL:          168 * time.Hour,
		CookieSecure:        true,
	}
	c.GoogleRedirectURL = env("BB_GOOGLE_REDIRECT_URL", strings.TrimRight(c.BaseURL, "/")+"/auth/callback")

	if v, err := positiveInt("BB_RETENTION_DAYS", 30); err != nil {
		return c, err
	} else {
		c.RetentionDays = v
	}
	if v, err := positiveInt("BB_CHECK_WORKERS", 16); err != nil {
		return c, err
	} else {
		c.CheckWorkers = v
	}

	level, err := parseLevel(env("BB_LOG_LEVEL", "info"))
	if err != nil {
		return c, err
	}
	c.LogLevel = level

	if v := os.Getenv("BB_SESSION_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return c, fmt.Errorf("BB_SESSION_TTL must be a positive duration, got %q", v)
		}
		c.SessionTTL = d
	}
	if v := os.Getenv("BB_COOKIE_SECURE"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return c, fmt.Errorf("BB_COOKIE_SECURE must be a boolean, got %q", v)
		}
		c.CookieSecure = b
	}

	if c.GoogleClientID != "" {
		if c.GoogleClientSecret == "" {
			return c, fmt.Errorf("BB_GOOGLE_CLIENT_SECRET is required when BB_GOOGLE_CLIENT_ID is set")
		}
		if len(c.AllowedEmailDomains) == 0 {
			return c, fmt.Errorf("BB_ALLOWED_EMAIL_DOMAINS is required when BB_GOOGLE_CLIENT_ID is set")
		}
		if c.GoogleRedirectURL == "" {
			return c, fmt.Errorf("BB_GOOGLE_REDIRECT_URL could not be derived")
		}
	}

	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func positiveInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", key, v)
	}
	return n, nil
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		if p := strings.ToLower(strings.TrimSpace(part)); p != "" {
			out = append(out, p)
		}
	}
	return out
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
