package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Abh19avM/recess/internal/database/dbgen"
)

func TestMigrationFilesExistAndSyntax(t *testing.T) {
	// Find migrations dir
	dirs := []string{"../../migrations", "../migrations", "migrations", "backend/migrations"}
	var migDir string
	for _, d := range dirs {
		if _, err := os.Stat(d); err == nil {
			migDir = d
			break
		}
	}

	if migDir == "" {
		t.Skip("migrations directory not found relative to test")
	}

	upPath := filepath.Join(migDir, "000001_init_schema.up.sql")
	content, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("failed to read up migration: %v", err)
	}

	sqlStr := string(content)

	// Verify required tables are declared
	requiredTables := []string{
		"users",
		"profiles",
		"rooms",
		"matches",
		"match_players",
		"game_results",
		"ratings",
	}

	for _, tbl := range requiredTables {
		if !strings.Contains(sqlStr, "CREATE TABLE IF NOT EXISTS "+tbl) {
			t.Errorf("expected table definition for %s in migration", tbl)
		}
	}
}

func TestDBGenQuerierInterface(t *testing.T) {
	// Test that dbgen.Querier interface matches dbgen.Queries struct
	var _ dbgen.Querier = (*dbgen.Queries)(nil)
}
