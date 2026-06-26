package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/comma-compliance/arc-relay/internal/config"
	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

// newTestHandlersWithInvites extends newTestHandlers with an InviteStore.
func newTestHandlersWithInvites(t *testing.T) (*Handlers, *store.SessionStore, *store.User, *store.InviteStore) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)
	invites := store.NewInviteStore(db)

	user, err := users.Create("testadmin", "pass", "admin")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}

	cfg := &config.Config{}
	cfg.Auth.SessionSecret = "test-secret"

	h := &Handlers{
		cfg:          cfg,
		users:        users,
		sessionStore: sessions,
		inviteStore:  invites,
		csrfSecret:   []byte(cfg.Auth.SessionSecret),
	}
	return h, sessions, user, invites
}

func TestInviteExchange_ValidToken(t *testing.T) {
	h, _, user, invites := newTestHandlersWithInvites(t)

	rawToken, _, err := invites.CreateAccountInvite("user", "write", nil, user.ID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("creating invite token: %v", err)
	}

	body, _ := json.Marshal(map[string]string{
		"token":    rawToken,
		"username": "newuser",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.handleInviteExchange(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp["api_key"] == "" {
		t.Error("response missing api_key")
	}

	// Verify user was created with correct role
	created, _ := h.users.GetByUsername("newuser")
	if created == nil {
		t.Fatal("user 'newuser' was not created")
	}
	if created.Role != "user" {
		t.Errorf("created user role = %q, want %q", created.Role, "user")
	}
}

func TestInviteExchange_UsernameConflict(t *testing.T) {
	h, _, user, invites := newTestHandlersWithInvites(t)

	rawToken, _, _ := invites.CreateAccountInvite("user", "write", nil, user.ID, time.Now().Add(time.Hour))

	// Try to create a user with the same username as the admin
	body, _ := json.Marshal(map[string]string{
		"token":    rawToken,
		"username": "testadmin",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.handleInviteExchange(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	// Token should still be available (rollback preserved it)
	peeked, _ := invites.Peek(rawToken)
	if peeked == nil {
		t.Error("invite token was consumed despite username conflict - should have been rolled back")
	}
}

func TestInviteExchange_WeakPassword(t *testing.T) {
	h, _, user, invites := newTestHandlersWithInvites(t)

	rawToken, _, _ := invites.CreateAccountInvite("user", "write", nil, user.ID, time.Now().Add(time.Hour))

	body, _ := json.Marshal(map[string]string{
		"token":    rawToken,
		"username": "shortpw",
		"password": "short",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.handleInviteExchange(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestInviteExchange_MissingCredentials(t *testing.T) {
	h, _, user, invites := newTestHandlersWithInvites(t)

	rawToken, _, _ := invites.CreateAccountInvite("user", "write", nil, user.ID, time.Now().Add(time.Hour))

	// Token only, no username/password
	body, _ := json.Marshal(map[string]string{"token": rawToken})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.handleInviteExchange(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestInviteExchange_InvalidToken(t *testing.T) {
	h, _, _, _ := newTestHandlersWithInvites(t)

	body, _ := json.Marshal(map[string]string{
		"token":    "bogus-invalid-token",
		"username": "someone",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.handleInviteExchange(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestInviteExchange_EmptyBody(t *testing.T) {
	h, _, _, _ := newTestHandlersWithInvites(t)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/invite", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.handleInviteExchange(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestInviteExchange_GETReturns405(t *testing.T) {
	h, _, _, _ := newTestHandlersWithInvites(t)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/invite", nil)
	rec := httptest.NewRecorder()

	h.handleInviteExchange(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusMethodNotAllowed, rec.Body.String())
	}
}
