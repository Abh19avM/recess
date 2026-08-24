package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Abh19avM/recess/internal/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database manages the PostgreSQL connection pool.
type Database struct {
	Pool *pgxpool.Pool
}

// New initializes a connection pool to the PostgreSQL database with production-ready connection pool settings.
func New(ctx context.Context, databaseURL string) (*Database, error) {
	if databaseURL == "" || databaseURL == "in-memory" {
		return nil, fmt.Errorf("database URL not configured (in-memory mode)")
	}
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 0
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute
	if config.ConnConfig.ConnectTimeout == 0 {
		config.ConnConfig.ConnectTimeout = 2 * time.Second
	}

	connectCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pgx connection pool: %w", err)
	}

	// Verify initial connectivity
	if err := pool.Ping(connectCtx); err != nil {
		slog.Warn("initial database ping failed (will retry on readiness checks)", "error", err)
	} else {
		slog.Info("connected to PostgreSQL successfully")
		// Attempt to run migrations if migrations dir is available
		for _, dir := range []string{"migrations", "../migrations", "../../migrations", "backend/migrations"} {
			if err := RunMigrations(connectCtx, pool, dir); err == nil {
				break
			}
		}
	}

	return &Database{Pool: pool}, nil
}

// Ping checks if the PostgreSQL database is reachable and returns latency.
func (d *Database) Ping(ctx context.Context) (time.Duration, error) {
	if d == nil || d.Pool == nil {
		return 0, fmt.Errorf("database connection pool not initialized")
	}
	defer metrics.ObserveDB("ping", "system", time.Now())
	start := time.Now()
	err := d.Pool.Ping(ctx)
	latency := time.Since(start)
	return latency, err
}

// Close safely drains and closes all connections in the pool.
func (d *Database) Close() {
	if d != nil && d.Pool != nil {
		slog.Info("closing PostgreSQL connection pool")
		d.Pool.Close()
	}
}
