package database

import (
	"context"
	"testing"
)

func TestDatabaseNilPing(t *testing.T) {
	var db *Database
	_, err := db.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error when pinging nil database, got nil")
	}

	db = &Database{Pool: nil}
	_, err = db.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error when pinging database with nil pool, got nil")
	}

	// Calling Close on nil db should not panic
	db.Close()
}

func TestDatabaseInvalidURL(t *testing.T) {
	ctx := context.Background()
	_, err := New(ctx, "invalid-database-url-:::://")
	if err == nil {
		t.Fatal("expected error when initializing with invalid database URL")
	}
}
