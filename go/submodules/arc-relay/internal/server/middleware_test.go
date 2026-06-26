package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/comma-compliance/arc-relay/internal/server"
	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func TestUserFromContext(t *testing.T) {
	t.Run("with user", func(t *testing.T) {
		// We can't set the private contextKey directly from an external test package,
		// so we test via the middleware round-trip instead.
		// UserFromContext with an empty context should return nil.
		user := server.UserFromContext(context.Background())
		if user != nil {
			t.Error("UserFromContext(empty) should return nil")
		}
	})
}

func setupUserStore(t *testing.T) (*store.UserStore, string) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	user, err := users.Create("testuser", "pass", "user")
	if err != nil {
		t.Fatalf("creating test user: %v", err)
	}
	rawKey, _, err := users.CreateAPIKey(user.ID, "test-key", nil)
	if err != nil {
		t.Fatalf("creating API key: %v", err)
	}
	return users, rawKey
}

func TestAPIKeyAuth(t *testing.T) {
	users, validKey := setupUserStore(t)

	// The inner handler verifies the user is in context
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := server.UserFromContext(r.Context())
		if user == nil {
			t.Error("user should be in context after APIKeyAuth")
			http.Error(w, "no user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := server.APIKeyAuth(users, "http://localhost:8080")(inner)

	t.Run("valid key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mcp/test", nil)
		req.Header.Set("Authorization", "Bearer "+validKey)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("missing header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mcp/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("bad prefix", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mcp/test", nil)
		req.Header.Set("Authorization", "Basic "+validKey)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mcp/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-key-value")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func TestAdminOnly(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := server.AdminOnly(inner)

	t.Run("admin passes", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		users := store.NewUserStore(db)
		admin, _ := users.Create("admin", "pass", "admin")

		// Use APIKeyAuth to set user in context, then chain AdminOnly
		rawKey, _, _ := users.CreateAPIKey(admin.ID, "admin-key", nil)
		fullHandler := server.APIKeyAuth(users, "http://localhost:8080")(handler)

		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+rawKey)
		rec := httptest.NewRecorder()
		fullHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		db := testutil.OpenTestDB(t)
		users := store.NewUserStore(db)
		user, _ := users.Create("regular", "pass", "user")

		rawKey, _, _ := users.CreateAPIKey(user.ID, "user-key", nil)
		fullHandler := server.APIKeyAuth(users, "http://localhost:8080")(handler)

		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+rawKey)
		rec := httptest.NewRecorder()
		fullHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("no user forbidden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}
