package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Clear relevant env vars to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("ENV")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("LOG_LEVEL")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected env development, got %s", cfg.Environment)
	}
	if cfg.DatabaseURL == "" {
		t.Errorf("expected non-empty database URL")
	}
	if cfg.RedisURL == "" {
		t.Errorf("expected non-empty redis URL")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected log level info, got %s", cfg.LogLevel)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("expected 15s read timeout, got %v", cfg.ReadTimeout)
	}
}

func TestConfigLoadCustomEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("DATABASE_URL", "postgres://custom:pass@localhost:5432/custom_db")
	t.Setenv("REDIS_URL", "redis://custom:6379/1")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("READ_TIMEOUT", "30s")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected env production, got %s", cfg.Environment)
	}
	if cfg.DatabaseURL != "postgres://custom:pass@localhost:5432/custom_db" {
		t.Errorf("expected custom database URL, got %s", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://custom:6379/1" {
		t.Errorf("expected custom redis URL, got %s", cfg.RedisURL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.LogLevel)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("expected 30s read timeout, got %v", cfg.ReadTimeout)
	}

	logger := cfg.SetupLogger()
	if logger == nil {
		t.Fatal("expected logger instance, got nil")
	}
}
