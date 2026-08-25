package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Abh19avM/recess/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies SQL migration files in sequence to the PostgreSQL database.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	slog.Info("checking and applying PostgreSQL database schema migrations")

	// Create schema_migrations table if not exists
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 1. Try embedded migrations first
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err == nil && len(entries) > 0 {
		var sqlFiles []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
				sqlFiles = append(sqlFiles, entry.Name())
			}
		}
		sort.Strings(sqlFiles)

		for _, version := range sqlFiles {
			var applied bool
			err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
			if err != nil {
				return fmt.Errorf("failed to check migration status for %s: %w", version, err)
			}
			if applied {
				continue
			}

			slog.Info("applying embedded migration", "version", version)
			content, err := fs.ReadFile(migrations.FS, version)
			if err != nil {
				return fmt.Errorf("failed to read embedded migration %s: %w", version, err)
			}

			tx, err := pool.Begin(ctx)
			if err != nil {
				return fmt.Errorf("failed to start migration transaction: %w", err)
			}

			if _, err := tx.Exec(ctx, string(content)); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("failed to execute migration %s: %w", version, err)
			}

			if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)", version, time.Now()); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("failed to record applied migration %s: %w", version, err)
			}

			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("failed to commit migration transaction for %s: %w", version, err)
			}
			slog.Info("embedded migration applied successfully", "version", version)
		}
		return nil
	}

	// Fallback to disk directory if embedded not available
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory %s: %w", migrationsDir, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		version := file.Name()

		// Check if migration is already applied
		var applied bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}

		if applied {
			continue
		}

		slog.Info("applying migration", "version", version)
		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start migration transaction: %w", err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)", version, time.Now()); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to record applied migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration transaction for %s: %w", version, err)
		}

		slog.Info("migration applied successfully", "version", version)
	}

	return nil
}
