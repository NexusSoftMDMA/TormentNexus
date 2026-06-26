package store_test

import (
	"testing"

	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func TestUserCreate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	user, err := users.Create("alice", "password123", "user")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.ID == "" {
		t.Error("user ID should be generated")
	}
	if user.Username != "alice" {
		t.Errorf("Username = %q, want %q", user.Username, "alice")
	}
	if user.Role != "user" {
		t.Errorf("Role = %q, want %q", user.Role, "user")
	}
	if user.AccessLevel != "write" {
		t.Errorf("AccessLevel = %q, want %q (default)", user.AccessLevel, "write")
	}
	if user.PasswordHash == "" {
		t.Error("PasswordHash should be set")
	}
	if user.PasswordHash == "password123" {
		t.Error("PasswordHash should not be plaintext")
	}
}

func TestUserCreateDuplicate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	_, err := users.Create("alice", "pass1", "user")
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = users.Create("alice", "pass2", "user")
	if err == nil {
		t.Error("duplicate Create() should return error")
	}
}

func TestCreateWithAccessLevel(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	t.Run("admin role forces admin access level", func(t *testing.T) {
		user, err := users.CreateWithAccessLevel("admin1", "pass", "admin", "read", nil)
		if err != nil {
			t.Fatalf("CreateWithAccessLevel() error = %v", err)
		}
		if user.AccessLevel != "admin" {
			t.Errorf("AccessLevel = %q, want %q (forced by admin role)", user.AccessLevel, "admin")
		}
	})

	t.Run("explicit access level for non-admin", func(t *testing.T) {
		user, err := users.CreateWithAccessLevel("reader", "pass", "user", "read", nil)
		if err != nil {
			t.Fatalf("CreateWithAccessLevel() error = %v", err)
		}
		if user.AccessLevel != "read" {
			t.Errorf("AccessLevel = %q, want %q", user.AccessLevel, "read")
		}
	})
}

func TestAuthenticate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	_, err := users.Create("bob", "correct-password", "user")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	t.Run("valid password", func(t *testing.T) {
		user, err := users.Authenticate("bob", "correct-password")
		if err != nil {
			t.Fatalf("Authenticate() error = %v", err)
		}
		if user == nil {
			t.Fatal("Authenticate() returned nil for valid credentials")
		}
		if user.Username != "bob" {
			t.Errorf("Username = %q, want %q", user.Username, "bob")
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		user, err := users.Authenticate("bob", "wrong-password")
		if err != nil {
			t.Fatalf("Authenticate() error = %v", err)
		}
		if user != nil {
			t.Error("Authenticate() should return nil for invalid password")
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		user, err := users.Authenticate("nonexistent", "password")
		if err != nil {
			t.Fatalf("Authenticate() error = %v", err)
		}
		if user != nil {
			t.Error("Authenticate() should return nil for nonexistent user")
		}
	})
}

func TestUserGetAndGetByUsername(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	created, _ := users.Create("charlie", "pass", "user")

	t.Run("Get found", func(t *testing.T) {
		user, err := users.Get(created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if user == nil {
			t.Fatal("Get() returned nil")
		}
		if user.Username != "charlie" {
			t.Errorf("Username = %q, want %q", user.Username, "charlie")
		}
	})

	t.Run("Get not found", func(t *testing.T) {
		user, err := users.Get("nonexistent-id")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if user != nil {
			t.Error("Get() should return nil for nonexistent ID")
		}
	})

	t.Run("GetByUsername found", func(t *testing.T) {
		user, err := users.GetByUsername("charlie")
		if err != nil {
			t.Fatalf("GetByUsername() error = %v", err)
		}
		if user == nil {
			t.Fatal("GetByUsername() returned nil")
		}
		if user.ID != created.ID {
			t.Errorf("ID = %q, want %q", user.ID, created.ID)
		}
	})

	t.Run("GetByUsername not found", func(t *testing.T) {
		user, err := users.GetByUsername("nonexistent")
		if err != nil {
			t.Fatalf("GetByUsername() error = %v", err)
		}
		if user != nil {
			t.Error("GetByUsername() should return nil for nonexistent username")
		}
	})
}

func TestUserList(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	if _, err := users.Create("user1", "pass", "user"); err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create("user2", "pass", "admin"); err != nil {
		t.Fatal(err)
	}

	list, err := users.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("List() returned %d users, want 2", len(list))
	}
}

func TestUserDeleteCascadesAPIKeys(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	user, err := users.Create("deleteme", "pass", "user")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := users.CreateAPIKey(user.ID, "my-key", nil); err != nil {
		t.Fatal(err)
	}

	// Verify key exists
	keys, _ := users.ListAPIKeys(user.ID)
	if len(keys) != 1 {
		t.Fatalf("expected 1 API key before delete, got %d", len(keys))
	}

	if err := users.Delete(user.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// User should be gone
	found, _ := users.Get(user.ID)
	if found != nil {
		t.Error("user should be deleted")
	}

	// API keys should be cascade-deleted
	var keyCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE user_id = ?", user.ID).Scan(&keyCount); err != nil {
		t.Fatal(err)
	}
	if keyCount != 0 {
		t.Errorf("API keys should be cascade-deleted, got %d remaining", keyCount)
	}
}

func TestEnsureAdmin(t *testing.T) {
	t.Run("creates admin when none exist", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		users := store.NewUserStore(db)

		if err := users.EnsureAdmin("admin-pass"); err != nil {
			t.Fatalf("EnsureAdmin() error = %v", err)
		}

		admin, err := users.GetByUsername("admin")
		if err != nil {
			t.Fatalf("GetByUsername() error = %v", err)
		}
		if admin == nil {
			t.Fatal("admin user should have been created")
		}
		if admin.Role != "admin" {
			t.Errorf("Role = %q, want %q", admin.Role, "admin")
		}
		if admin.AccessLevel != "admin" {
			t.Errorf("AccessLevel = %q, want %q", admin.AccessLevel, "admin")
		}
	})

	t.Run("idempotent when users exist", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		users := store.NewUserStore(db)

		if _, err := users.Create("existing", "pass", "user"); err != nil {
			t.Fatal(err)
		}

		if err := users.EnsureAdmin("admin-pass"); err != nil {
			t.Fatalf("EnsureAdmin() error = %v", err)
		}

		// Should not create another admin
		list, _ := users.List()
		if len(list) != 1 {
			t.Errorf("user count = %d, want 1 (no new admin created)", len(list))
		}
	})
}

func TestAPIKeyRoundTrip(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	user, _ := users.Create("apiuser", "pass", "user")

	rawKey, ak, err := users.CreateAPIKey(user.ID, "test-key", nil)
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	if rawKey == "" {
		t.Error("rawKey should not be empty")
	}
	if ak.ID == "" {
		t.Error("APIKey ID should be generated")
	}
	if ak.Name != "test-key" {
		t.Errorf("Name = %q, want %q", ak.Name, "test-key")
	}

	// Validate the raw key
	validated, err := users.ValidateAPIKey(rawKey)
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}
	if validated == nil {
		t.Fatal("ValidateAPIKey() returned nil for valid key")
	}
	if validated.ID != user.ID {
		t.Errorf("validated user ID = %q, want %q", validated.ID, user.ID)
	}
}

func TestValidateAPIKeyInvalid(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	user, err := users.ValidateAPIKey("nonexistent-key")
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}
	if user != nil {
		t.Error("ValidateAPIKey() should return nil for nonexistent key")
	}
}

func TestValidateAPIKeyRevoked(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)

	user, _ := users.Create("revokeuser", "pass", "user")
	rawKey, ak, _ := users.CreateAPIKey(user.ID, "to-revoke", nil)

	if err := users.RevokeAPIKey(ak.ID); err != nil {
		t.Fatalf("RevokeAPIKey() error = %v", err)
	}

	// Validate should return nil for revoked key
	validated, err := users.ValidateAPIKey(rawKey)
	if err != nil {
		t.Fatalf("ValidateAPIKey() error = %v", err)
	}
	if validated != nil {
		t.Error("ValidateAPIKey() should return nil for revoked key")
	}
}
