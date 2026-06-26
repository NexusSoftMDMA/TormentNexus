package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/comma-compliance/arc-relay/internal/config"
	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

// newTestHandlers creates a minimal Handlers for CSRF testing.
func newTestHandlers(t *testing.T) (*Handlers, *store.SessionStore, *store.User) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)

	user, err := users.Create("testuser", "pass", "admin")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}

	cfg := &config.Config{}
	cfg.Auth.SessionSecret = "test-csrf-secret"

	h := &Handlers{
		cfg:          cfg,
		users:        users,
		sessionStore: sessions,
		csrfSecret:   []byte(cfg.Auth.SessionSecret),
	}
	return h, sessions, user
}

func TestRequireAuth_CSRF(t *testing.T) {
	h, sessions, user := newTestHandlers(t)

	sessionID := "test-session-id"
	if err := sessions.Create(sessionID, user.ID, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("creating session: %v", err)
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := h.requireAuth(inner)

	t.Run("POST without CSRF token returns 403", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/servers/1/start", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
		if !strings.Contains(rec.Body.String(), "CSRF") {
			t.Errorf("body should mention CSRF, got: %s", rec.Body.String())
		}
	})

	t.Run("POST with valid CSRF form field succeeds", func(t *testing.T) {
		token := h.csrfToken(sessionID)
		form := url.Values{"csrf_token": {token}}
		req := httptest.NewRequest("POST", "/servers/1/start", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("POST with valid CSRF header succeeds", func(t *testing.T) {
		token := h.csrfToken(sessionID)
		req := httptest.NewRequest("POST", "/servers/1/start", nil)
		req.Header.Set("X-CSRF-Token", token)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("POST with wrong CSRF token returns 403", func(t *testing.T) {
		form := url.Values{"csrf_token": {"invalid-token"}}
		req := httptest.NewRequest("POST", "/servers/1/start", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("GET without CSRF token succeeds", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/servers/1", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("DELETE without CSRF token returns 403", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/servers/1", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionID})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}
