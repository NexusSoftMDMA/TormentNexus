package store

import (
	"path/filepath"
	"testing"

	"github.com/comma-compliance/arc-relay/migrations"
)

// openTestDB creates a file-based SQLite database in a temp directory.
// We cannot use testutil here because it would create an import cycle
// (testutil imports store, and this file is in the store package).
// We use a file instead of :memory: because SyncAfterEnumerate uses both
// a transaction and a separate query (GetAllTiers), which require all
// connections to share the same database.
func openTestDB(t *testing.T) *DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path, migrations.FS)
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestTierLevel(t *testing.T) {
	tests := []struct {
		tier string
		want int
	}{
		{"read", 1},
		{"write", 2},
		{"admin", 3},
		{"unknown", 2},
		{"", 2},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			got := tierLevel(tt.tier)
			if got != tt.want {
				t.Errorf("tierLevel(%q) = %d, want %d", tt.tier, got, tt.want)
			}
		})
	}
}

func TestCheckAccess(t *testing.T) {
	db := openTestDB(t)
	access := NewAccessStore(db)

	tests := []struct {
		name     string
		user     string
		endpoint string
		want     bool
	}{
		{"admin >= admin", "admin", "admin", true},
		{"admin >= write", "admin", "write", true},
		{"admin >= read", "admin", "read", true},
		{"write >= write", "write", "write", true},
		{"write >= read", "write", "read", true},
		{"write < admin", "write", "admin", false},
		{"read >= read", "read", "read", true},
		{"read < write", "read", "write", false},
		{"read < admin", "read", "admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := access.CheckAccess(tt.user, tt.endpoint)
			if got != tt.want {
				t.Errorf("CheckAccess(%q, %q) = %v, want %v", tt.user, tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestGetTier(t *testing.T) {
	db := openTestDB(t)
	access := NewAccessStore(db)

	// Create a server for FK constraint
	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	t.Run("default is write", func(t *testing.T) {
		tier := access.GetTier("srv-1", "tool", "nonexistent")
		if tier != "write" {
			t.Errorf("GetTier() = %q, want %q (default)", tier, "write")
		}
	})

	t.Run("returns stored value", func(t *testing.T) {
		if err := access.SetTier("srv-1", "tool", "get_users", "read"); err != nil {
			t.Fatal(err)
		}
		tier := access.GetTier("srv-1", "tool", "get_users")
		if tier != "read" {
			t.Errorf("GetTier() = %q, want %q", tier, "read")
		}
	})
}

func TestSetTier(t *testing.T) {
	db := openTestDB(t)
	access := NewAccessStore(db)

	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	// Insert
	if err := access.SetTier("srv-1", "tool", "my_tool", "read"); err != nil {
		t.Fatalf("SetTier() error = %v", err)
	}

	tiers, _ := access.GetAllTiers("srv-1")
	if len(tiers) != 1 {
		t.Fatalf("expected 1 tier, got %d", len(tiers))
	}
	if tiers[0].AutoClassified {
		t.Error("SetTier should set auto_classified = FALSE")
	}

	// Update (upsert)
	if err := access.SetTier("srv-1", "tool", "my_tool", "admin"); err != nil {
		t.Fatalf("SetTier() update error = %v", err)
	}

	tier := access.GetTier("srv-1", "tool", "my_tool")
	if tier != "admin" {
		t.Errorf("GetTier() after update = %q, want %q", tier, "admin")
	}
}

func TestSyncAfterEnumerate(t *testing.T) {
	db := openTestDB(t)
	access := NewAccessStore(db)

	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	classify := func(epType, name, desc string) string {
		if name == "get_users" {
			return "read"
		}
		if name == "delete_users" {
			return "admin"
		}
		return "write"
	}

	// Initial sync
	endpoints := []EndpointInfo{
		{Type: "tool", Name: "get_users", Description: "Get users"},
		{Type: "tool", Name: "create_user", Description: "Create user"},
		{Type: "tool", Name: "delete_users", Description: "Delete users"},
	}
	access.SyncAfterEnumerate("srv-1", endpoints, classify)

	tiers, _ := access.GetAllTiers("srv-1")
	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers after sync, got %d", len(tiers))
	}

	// Verify auto-classified values
	tier := access.GetTier("srv-1", "tool", "get_users")
	if tier != "read" {
		t.Errorf("get_users tier = %q, want %q", tier, "read")
	}
	tier = access.GetTier("srv-1", "tool", "delete_users")
	if tier != "admin" {
		t.Errorf("delete_users tier = %q, want %q", tier, "admin")
	}

	// Manually override one tier
	if err := access.SetTier("srv-1", "tool", "get_users", "admin"); err != nil {
		t.Fatal(err)
	}

	// Re-sync: should preserve manual override, update auto-classified, remove stale
	newEndpoints := []EndpointInfo{
		{Type: "tool", Name: "get_users", Description: "Get users"},
		{Type: "tool", Name: "create_user", Description: "Create user"},
		// delete_users removed
	}
	access.SyncAfterEnumerate("srv-1", newEndpoints, classify)

	// Manual override preserved
	tier = access.GetTier("srv-1", "tool", "get_users")
	if tier != "admin" {
		t.Errorf("manual override should be preserved: get_users tier = %q, want %q", tier, "admin")
	}

	// Stale entry removed
	tiers, _ = access.GetAllTiers("srv-1")
	for _, tr := range tiers {
		if tr.EndpointName == "delete_users" {
			t.Error("stale endpoint delete_users should have been removed")
		}
	}

	if len(tiers) != 2 {
		t.Errorf("expected 2 tiers after re-sync, got %d", len(tiers))
	}
}
