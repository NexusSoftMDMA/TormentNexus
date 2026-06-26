package store_test

import (
	"testing"

	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func TestProfileCRUD(t *testing.T) {
	db := testutil.OpenTestDB(t)
	profiles := store.NewProfileStore(db)

	t.Run("Create", func(t *testing.T) {
		p, err := profiles.Create("read-only", "Read-only access profile")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if p.ID == "" {
			t.Error("ID should be generated")
		}
		if p.Name != "read-only" {
			t.Errorf("Name = %q, want %q", p.Name, "read-only")
		}
		if p.Description != "Read-only access profile" {
			t.Errorf("Description = %q, want %q", p.Description, "Read-only access profile")
		}
	})

	t.Run("Get", func(t *testing.T) {
		created, _ := profiles.Create("getme", "desc")
		got, err := profiles.Get(created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Name != "getme" {
			t.Errorf("Name = %q, want %q", got.Name, "getme")
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		created, _ := profiles.Create("by-name", "desc")
		got, err := profiles.GetByName("by-name")
		if err != nil {
			t.Fatalf("GetByName() error = %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("ID = %q, want %q", got.ID, created.ID)
		}
	})

	t.Run("List sorted", func(t *testing.T) {
		// Already have "read-only", "getme", "by-name" from above
		list, err := profiles.List()
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(list) < 3 {
			t.Fatalf("List() returned %d profiles, want at least 3", len(list))
		}
		// Verify sorted by name
		for i := 1; i < len(list); i++ {
			if list[i].Name < list[i-1].Name {
				t.Errorf("List() not sorted: %q comes after %q", list[i].Name, list[i-1].Name)
			}
		}
	})

	t.Run("Update", func(t *testing.T) {
		created, _ := profiles.Create("old-name", "old desc")
		if err := profiles.Update(created.ID, "new-name", "new desc"); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		got, _ := profiles.Get(created.ID)
		if got.Name != "new-name" {
			t.Errorf("Name = %q, want %q", got.Name, "new-name")
		}
		if got.Description != "new desc" {
			t.Errorf("Description = %q, want %q", got.Description, "new desc")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		created, _ := profiles.Create("deleteme", "")
		if err := profiles.Delete(created.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		// Get should return error (sql.ErrNoRows wrapped)
		_, err := profiles.Get(created.ID)
		if err == nil {
			t.Error("Get() after Delete should return error")
		}
	})
}

func TestProfilePermissionCRUD(t *testing.T) {
	db := testutil.OpenTestDB(t)
	profiles := store.NewProfileStore(db)

	// Insert a server for FK constraint
	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	profile, err := profiles.Create("perm-test", "")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	t.Run("SetPermission", func(t *testing.T) {
		if err := profiles.SetPermission(profile.ID, "srv-1", "tool", "get_users"); err != nil {
			t.Fatalf("SetPermission() error = %v", err)
		}
	})

	t.Run("GetPermissions", func(t *testing.T) {
		perms, err := profiles.GetPermissions(profile.ID)
		if err != nil {
			t.Fatalf("GetPermissions() error = %v", err)
		}
		if len(perms) != 1 {
			t.Fatalf("GetPermissions() returned %d, want 1", len(perms))
		}
		if perms[0].EndpointName != "get_users" {
			t.Errorf("EndpointName = %q, want %q", perms[0].EndpointName, "get_users")
		}
	})

	t.Run("CheckPermission granted", func(t *testing.T) {
		ok, err := profiles.CheckPermission(profile.ID, "srv-1", "tool", "get_users")
		if err != nil {
			t.Fatalf("CheckPermission() error = %v", err)
		}
		if !ok {
			t.Error("CheckPermission() should return true for granted permission")
		}
	})

	t.Run("CheckPermission denied", func(t *testing.T) {
		ok, err := profiles.CheckPermission(profile.ID, "srv-1", "tool", "delete_users")
		if err != nil {
			t.Fatalf("CheckPermission() error = %v", err)
		}
		if ok {
			t.Error("CheckPermission() should return false for non-existent permission")
		}
	})

	t.Run("RemovePermission", func(t *testing.T) {
		if err := profiles.RemovePermission(profile.ID, "srv-1", "tool", "get_users"); err != nil {
			t.Fatalf("RemovePermission() error = %v", err)
		}
		perms, _ := profiles.GetPermissions(profile.ID)
		if len(perms) != 0 {
			t.Errorf("GetPermissions() after remove returned %d, want 0", len(perms))
		}
	})
}

func TestProfileBulkSetPermissions(t *testing.T) {
	db := testutil.OpenTestDB(t)
	profiles := store.NewProfileStore(db)

	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	profile, err := profiles.Create("bulk-test", "")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Set 3 permissions initially
	initial := []store.ProfilePermission{
		{ProfileID: profile.ID, ServerID: "srv-1", EndpointType: "tool", EndpointName: "a"},
		{ProfileID: profile.ID, ServerID: "srv-1", EndpointType: "tool", EndpointName: "b"},
		{ProfileID: profile.ID, ServerID: "srv-1", EndpointType: "tool", EndpointName: "c"},
	}
	if err := profiles.BulkSetPermissions(profile.ID, "srv-1", initial); err != nil {
		t.Fatalf("BulkSetPermissions() initial error = %v", err)
	}

	perms, _ := profiles.GetPermissions(profile.ID)
	if len(perms) != 3 {
		t.Fatalf("expected 3 permissions after initial bulk set, got %d", len(perms))
	}

	// Replace with 2 permissions
	replacement := []store.ProfilePermission{
		{ProfileID: profile.ID, ServerID: "srv-1", EndpointType: "tool", EndpointName: "x"},
		{ProfileID: profile.ID, ServerID: "srv-1", EndpointType: "tool", EndpointName: "y"},
	}
	if err := profiles.BulkSetPermissions(profile.ID, "srv-1", replacement); err != nil {
		t.Fatalf("BulkSetPermissions() replacement error = %v", err)
	}

	perms, _ = profiles.GetPermissions(profile.ID)
	if len(perms) != 2 {
		t.Errorf("expected 2 permissions after replacement, got %d", len(perms))
	}
}

func TestProfileSeedFromTier(t *testing.T) {
	db := testutil.OpenTestDB(t)
	profiles := store.NewProfileStore(db)

	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	// Insert access tiers for the server
	if _, err := db.Exec(`INSERT INTO endpoint_access_tiers (server_id, endpoint_type, endpoint_name, access_tier) VALUES ('srv-1', 'tool', 'list_items', 'read')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO endpoint_access_tiers (server_id, endpoint_type, endpoint_name, access_tier) VALUES ('srv-1', 'tool', 'create_item', 'write')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO endpoint_access_tiers (server_id, endpoint_type, endpoint_name, access_tier) VALUES ('srv-1', 'tool', 'delete_all', 'admin')`); err != nil {
		t.Fatal(err)
	}

	t.Run("read tier", func(t *testing.T) {
		p, _ := profiles.Create("read-tier", "")
		if err := profiles.SeedFromTier(p.ID, "srv-1", "read"); err != nil {
			t.Fatalf("SeedFromTier(read) error = %v", err)
		}
		perms, _ := profiles.GetPermissions(p.ID)
		if len(perms) != 1 {
			t.Errorf("read tier: expected 1 permission, got %d", len(perms))
		}
	})

	t.Run("write tier", func(t *testing.T) {
		p, _ := profiles.Create("write-tier", "")
		if err := profiles.SeedFromTier(p.ID, "srv-1", "write"); err != nil {
			t.Fatalf("SeedFromTier(write) error = %v", err)
		}
		perms, _ := profiles.GetPermissions(p.ID)
		if len(perms) != 2 {
			t.Errorf("write tier: expected 2 permissions, got %d", len(perms))
		}
	})

	t.Run("admin tier", func(t *testing.T) {
		p, _ := profiles.Create("admin-tier", "")
		if err := profiles.SeedFromTier(p.ID, "srv-1", "admin"); err != nil {
			t.Fatalf("SeedFromTier(admin) error = %v", err)
		}
		perms, _ := profiles.GetPermissions(p.ID)
		if len(perms) != 3 {
			t.Errorf("admin tier: expected 3 permissions, got %d", len(perms))
		}
	})
}

func TestProfileCascadeDelete(t *testing.T) {
	db := testutil.OpenTestDB(t)
	profiles := store.NewProfileStore(db)

	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	profile, err := profiles.Create("cascade-test", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := profiles.SetPermission(profile.ID, "srv-1", "tool", "a"); err != nil {
		t.Fatal(err)
	}
	if err := profiles.SetPermission(profile.ID, "srv-1", "tool", "b"); err != nil {
		t.Fatal(err)
	}

	// Verify permissions exist
	perms, _ := profiles.GetPermissions(profile.ID)
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions before delete, got %d", len(perms))
	}

	if err := profiles.Delete(profile.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify permissions were cascade-deleted
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM profile_permissions WHERE profile_id = ?", profile.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("permissions should be cascade-deleted, got %d remaining", count)
	}
}

func TestProfilePermissionCount(t *testing.T) {
	db := testutil.OpenTestDB(t)
	profiles := store.NewProfileStore(db)

	if _, err := db.Exec(`INSERT INTO servers (id, name, display_name, server_type, config, status, created_at, updated_at)
		VALUES ('srv-1', 'test', 'Test', 'stdio', '{}', 'stopped', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatal(err)
	}

	profile, err := profiles.Create("count-test", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := profiles.SetPermission(profile.ID, "srv-1", "tool", "a"); err != nil {
		t.Fatal(err)
	}
	if err := profiles.SetPermission(profile.ID, "srv-1", "tool", "b"); err != nil {
		t.Fatal(err)
	}
	if err := profiles.SetPermission(profile.ID, "srv-1", "resource", "c"); err != nil {
		t.Fatal(err)
	}

	count, err := profiles.PermissionCount(profile.ID)
	if err != nil {
		t.Fatalf("PermissionCount() error = %v", err)
	}
	if count != 3 {
		t.Errorf("PermissionCount() = %d, want 3", count)
	}
}

func TestProfileAPIKeyCount(t *testing.T) {
	db := testutil.OpenTestDB(t)
	profiles := store.NewProfileStore(db)
	users := store.NewUserStore(db)

	user, err := users.Create("testuser", "pass", "admin")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}

	profile, err := profiles.Create("test-profile", "")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Create an API key linked to this profile
	_, _, err = users.CreateAPIKey(user.ID, "test key", &profile.ID)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	count, err := profiles.APIKeyCount(profile.ID)
	if err != nil {
		t.Fatalf("APIKeyCount() error = %v", err)
	}
	if count != 1 {
		t.Errorf("APIKeyCount() = %d, want 1", count)
	}

	// Profile with no keys should return 0
	other, _ := profiles.Create("no-keys", "")
	count, err = profiles.APIKeyCount(other.ID)
	if err != nil {
		t.Fatalf("APIKeyCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("APIKeyCount() = %d, want 0", count)
	}
}
