package testutil

import (
	"path/filepath"
	"testing"

	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/migrations"
)

// OpenTestDB creates an in-memory SQLite database with all migrations applied.
func OpenTestDB(t *testing.T) *store.DB {
	t.Helper()
	db, err := store.Open(":memory:", migrations.FS)
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// OpenTestFileDB creates a temp-file SQLite database with all migrations applied.
// Use this instead of OpenTestDB when tests span goroutines or connections
// (e.g., background workers), since :memory: DBs are per-connection in SQLite.
func OpenTestFileDB(t *testing.T) *store.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.Open(dbPath, migrations.FS)
	if err != nil {
		t.Fatalf("opening test file db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
