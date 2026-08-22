package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration values for the application.
type Config struct {
	Port            string        `json:"port"`
	Environment     string        `json:"environment"`
	DatabaseURL     string        `json:"database_url"`
	RedisURL        string        `json:"redis_url"`
	LogLevel        string        `json:"log_level"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	CORSAllowed     []string      `json:"cors_allowed"`
	JWTSecret       string        `json:"-"`
}

// Load loads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", getEnv("ENV", "development")),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://recess:recess_secret@localhost:5432/recess_db?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		JWTSecret:       getEnv("JWT_SECRET", "recess_jwt_super_secret_signing_key_for_development"),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:     getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		CORSAllowed:     []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:5173"},
	}
}

// SetupLogger initializes the structured slog logger based on environment and log level.
func (c *Config) SetupLogger() *slog.Logger {
	var level slog.Level
	switch c.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if c.Environment == "production" || c.Environment == "staging" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

// Helper functions for reading environment variables with fallback defaults.

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
		// Try parsing as integer seconds
		if sec, err := strconv.Atoi(val); err == nil {
			return time.Duration(sec) * time.Second
		}
		slog.Warn(fmt.Sprintf("Invalid duration for %s=%q, using fallback %v", key, val, fallback))
	}
	return fallback
}
