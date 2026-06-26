package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/comma-compliance/arc-relay/internal/config"
	"github.com/comma-compliance/arc-relay/internal/middleware"
	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

// newTestArchiveHandoffStore builds a store without the background
// cleanup goroutine so tests are deterministic and do not leak timers.
func newTestArchiveHandoffStore() *archiveHandoffStore {
	return &archiveHandoffStore{
		requests: make(map[string]*archiveHandoffRequest),
	}
}

func TestArchiveHandoffStore_CreateAndConsume(t *testing.T) {
	s := newTestArchiveHandoffStore()

	nonce, err := s.create("user-1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if nonce == "" {
		t.Fatal("nonce should not be empty")
	}

	if !s.consume(nonce, "user-1") {
		t.Fatal("consume with matching user should succeed")
	}

	// Second consume should fail - nonces are single-use.
	if s.consume(nonce, "user-1") {
		t.Fatal("second consume should fail - nonce is single-use")
	}
}

func TestArchiveHandoffStore_CrossUserRejected(t *testing.T) {
	s := newTestArchiveHandoffStore()

	nonce, err := s.create("user-1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if s.consume(nonce, "user-2") {
		t.Fatal("consume from different user should fail")
	}
	// And the nonce must be burned by the failed attempt so it cannot
	// be replayed by the legitimate user either. This protects against
	// an attacker trying nonces they observed in someone else's URL
	// bar - the real user's subsequent handoff will fail, forcing them
	// to restart the flow rather than silently succeeding with a
	// possibly-tampered fragment.
	if s.consume(nonce, "user-1") {
		t.Fatal("nonce should be burned after a failed cross-user consume")
	}
}

func TestArchiveHandoffStore_UnknownNonce(t *testing.T) {
	s := newTestArchiveHandoffStore()

	if s.consume("does-not-exist", "user-1") {
		t.Fatal("consume of unknown nonce should fail")
	}
}

func TestArchiveHandoffStore_ExpiredNonce(t *testing.T) {
	s := newTestArchiveHandoffStore()

	// Manually insert an already-expired request so we do not have to
	// sleep the whole TTL during tests.
	s.requests["expired"] = &archiveHandoffRequest{
		Nonce:     "expired",
		UserID:    "user-1",
		CreatedAt: time.Now().Add(-archiveHandoffTTL - time.Minute),
		ExpiresAt: time.Now().Add(-time.Minute),
	}

	if s.consume("expired", "user-1") {
		t.Fatal("expired nonce should not be consumable")
	}
}

func TestDecorateArchiveConfigForDisplay_NoKey(t *testing.T) {
	raw := []byte(`{"url":"https://example.com","auth_type":"bearer","include":"both"}`)
	got := decorateArchiveConfigForDisplay(raw)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := parsed["nacl_key_id"]; ok {
		t.Error("nacl_key_id should not be present when no recipient key is configured")
	}
}

func TestDecorateArchiveConfigForDisplay_WithKey(t *testing.T) {
	// 32-byte all-zeros key is valid for the decode path; the
	// fingerprint it produces is deterministic.
	// base64(32 zero bytes) = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	zeroKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	raw := []byte(`{"url":"https://example.com","nacl_recipient_key":"` + zeroKey + `"}`)
	got := decorateArchiveConfigForDisplay(raw)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	kid, ok := parsed["nacl_key_id"].(string)
	if !ok || kid == "" {
		t.Errorf("nacl_key_id should be present and non-empty, got %v", parsed["nacl_key_id"])
	}
	// Store blob must be untouched: the original key stays in the blob
	// and nothing else is added beyond the computed kid.
	if parsed["nacl_recipient_key"] != zeroKey {
		t.Error("decorate should not mutate the recipient key")
	}
}

func TestDecorateArchiveConfigForDisplay_InvalidJSONPassthrough(t *testing.T) {
	raw := []byte(`not json at all`)
	got := decorateArchiveConfigForDisplay(raw)
	if got != string(raw) {
		t.Errorf("unparseable input should pass through unchanged, got %q", got)
	}
}

func TestDecorateArchiveConfigForDisplay_Empty(t *testing.T) {
	if got := decorateArchiveConfigForDisplay(nil); got != "" {
		t.Errorf("empty input should return empty, got %q", got)
	}
	if got := decorateArchiveConfigForDisplay([]byte{}); got != "" {
		t.Errorf("empty input should return empty, got %q", got)
	}
}

// newHandoffTestHandlers builds a Handlers wired with just enough
// state to exercise the archive handoff endpoints end-to-end.
func newHandoffTestHandlers(t *testing.T) (*Handlers, *store.User) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	sessions := store.NewSessionStore(db)
	mwStore := store.NewMiddlewareStore(db)

	admin, err := users.Create("admin", "pass", "admin")
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}

	cfg := &config.Config{}
	cfg.Auth.SessionSecret = "test-handoff-secret"

	h := &Handlers{
		cfg:             cfg,
		users:           users,
		sessionStore:    sessions,
		middlewareStore: mwStore,
		archiveHandoff:  newTestArchiveHandoffStore(),
		csrfSecret:      []byte(cfg.Auth.SessionSecret),
	}
	return h, admin
}

// callHandoffComplete invokes handleArchiveHandoffComplete directly
// with a context that carries the given user. We bypass the full
// CSRF middleware chain because CSRF is already covered by
// csrf_test.go; here we are testing the handler's own semantics.
func callHandoffComplete(h *Handlers, user *store.User, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/archive/handoff/complete", bytes.NewReader(b))
	req = req.WithContext(setUser(context.Background(), user))
	rec := httptest.NewRecorder()
	h.handleArchiveHandoffComplete(rec, req)
	return rec
}

func TestHandleArchiveHandoffComplete_RejectsMissingState(t *testing.T) {
	h, admin := newHandoffTestHandlers(t)
	rec := callHandoffComplete(h, admin, map[string]string{
		"archive_url":   "https://example.com/ingest",
		"archive_token": "token",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleArchiveHandoffComplete_RejectsUnknownState(t *testing.T) {
	h, admin := newHandoffTestHandlers(t)
	rec := callHandoffComplete(h, admin, map[string]string{
		"state":         "never-issued",
		"archive_url":   "https://example.com/ingest",
		"archive_token": "token",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleArchiveHandoffComplete_RejectsCrossUserState(t *testing.T) {
	h, admin := newHandoffTestHandlers(t)

	// Mint a nonce for the admin, then try to redeem it while
	// impersonating a different user. This is the attack the cross-
	// user binding is meant to stop.
	nonce, err := h.archiveHandoff.create(admin.ID)
	if err != nil {
		t.Fatalf("create nonce: %v", err)
	}

	otherUser := &store.User{ID: "other-id", Username: "other", Role: "admin"}
	rec := callHandoffComplete(h, otherUser, map[string]string{
		"state":         nonce,
		"archive_url":   "https://example.com/ingest",
		"archive_token": "token",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for cross-user nonce", rec.Code)
	}

	// And the nonce must be burned, so even the legitimate admin
	// cannot use it afterwards. Forces the admin to restart the flow
	// if someone tampered with their session.
	rec2 := callHandoffComplete(h, admin, map[string]string{
		"state":         nonce,
		"archive_url":   "https://example.com/ingest",
		"archive_token": "token",
	})
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("nonce should be burned after failed cross-user attempt, got status %d", rec2.Code)
	}
}

func TestHandleArchiveHandoffComplete_HappyPath(t *testing.T) {
	h, admin := newHandoffTestHandlers(t)

	nonce, err := h.archiveHandoff.create(admin.ID)
	if err != nil {
		t.Fatalf("create nonce: %v", err)
	}

	rec := callHandoffComplete(h, admin, map[string]string{
		"state":         nonce,
		"archive_url":   "https://example.com/ingest",
		"archive_token": "bearer-token",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	// Saved config should reflect what we posted.
	stored, err := h.middlewareStore.GetGlobal("archive")
	if err != nil {
		t.Fatalf("GetGlobal: %v", err)
	}
	if stored == nil {
		t.Fatal("expected a stored archive config after successful handoff")
	}
	var saved middleware.ArchiveConfig
	if err := json.Unmarshal(stored.Config, &saved); err != nil {
		t.Fatalf("unmarshal stored config: %v", err)
	}
	if saved.URL != "https://example.com/ingest" {
		t.Errorf("saved URL = %q, want %q", saved.URL, "https://example.com/ingest")
	}
	if saved.AuthType != "bearer" || saved.AuthValue != "bearer-token" {
		t.Errorf("saved auth = (%q, %q), want (bearer, bearer-token)", saved.AuthType, saved.AuthValue)
	}

	// Response must echo auth_value so the JS form does not end up
	// with a stale empty field and silently clobber the token later.
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	configMap, ok := resp["config"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing config object")
	}
	if av, _ := configMap["auth_value"].(string); av != "bearer-token" {
		t.Errorf("response config.auth_value = %q, want %q", av, "bearer-token")
	}
}

func TestHandleArchiveHandoffComplete_PreservesTokenOnKeyOnlyRotation(t *testing.T) {
	h, admin := newHandoffTestHandlers(t)

	// Seed an existing archive config with a bearer token the way
	// handleGlobalMiddleware would have persisted it.
	seed := middleware.ArchiveConfig{
		URL:          "https://old.example.com/ingest",
		AuthType:     "bearer",
		AuthValue:    "old-token",
		Include:      "both",
		APIKeyHeader: "X-API-Key",
	}
	raw, _ := json.Marshal(seed)
	if err := h.middlewareStore.UpsertGlobal(&store.MiddlewareConfig{
		Middleware: "archive",
		Config:     raw,
		Priority:   40,
	}); err != nil {
		t.Fatalf("seed global archive config: %v", err)
	}

	nonce, err := h.archiveHandoff.create(admin.ID)
	if err != nil {
		t.Fatalf("create nonce: %v", err)
	}

	// Re-run handoff with only a new URL - no archive_token, no key.
	// The existing token must survive the round-trip.
	rec := callHandoffComplete(h, admin, map[string]string{
		"state":       nonce,
		"archive_url": "https://new.example.com/ingest",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	stored, _ := h.middlewareStore.GetGlobal("archive")
	var saved middleware.ArchiveConfig
	if err := json.Unmarshal(stored.Config, &saved); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if saved.URL != "https://new.example.com/ingest" {
		t.Errorf("URL not updated: %q", saved.URL)
	}
	if saved.AuthValue != "old-token" {
		t.Errorf("token clobbered: got %q, want %q", saved.AuthValue, "old-token")
	}
	if saved.AuthType != "bearer" {
		t.Errorf("auth_type clobbered: got %q", saved.AuthType)
	}
}

func TestHandleArchiveHandoffComplete_RejectsExpiredState(t *testing.T) {
	h, admin := newHandoffTestHandlers(t)

	// Insert an already-expired nonce directly so we do not have to
	// sleep the 10-minute TTL in a test.
	nonce := "expired-nonce"
	h.archiveHandoff.mu.Lock()
	h.archiveHandoff.requests[nonce] = &archiveHandoffRequest{
		Nonce:     nonce,
		UserID:    admin.ID,
		CreatedAt: time.Now().Add(-archiveHandoffTTL - time.Minute),
		ExpiresAt: time.Now().Add(-time.Minute),
	}
	h.archiveHandoff.mu.Unlock()

	rec := callHandoffComplete(h, admin, map[string]string{
		"state":         nonce,
		"archive_url":   "https://example.com/ingest",
		"archive_token": "token",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expired nonce should return 400, got %d", rec.Code)
	}
}

func TestHandleArchiveHandoffBegin_AdminGated(t *testing.T) {
	h, _ := newHandoffTestHandlers(t)

	nonAdmin := &store.User{ID: "u2", Username: "reader", Role: "viewer"}
	req := httptest.NewRequest(http.MethodPost, "/api/archive/handoff/begin", nil)
	req = req.WithContext(setUser(context.Background(), nonAdmin))
	rec := httptest.NewRecorder()
	h.handleArchiveHandoffBegin(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("non-admin should get 403, got %d", rec.Code)
	}
}

func TestHandleArchiveHandoffBegin_HappyPath(t *testing.T) {
	h, admin := newHandoffTestHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/api/archive/handoff/begin", nil)
	req = req.WithContext(setUser(context.Background(), admin))
	rec := httptest.NewRecorder()
	h.handleArchiveHandoffBegin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	state, ok := resp["state"].(string)
	if !ok || state == "" {
		t.Errorf("state missing or empty: %v", resp["state"])
	}
	expiresIn, ok := resp["expires_in"].(float64)
	if !ok || expiresIn <= 0 {
		t.Errorf("expires_in missing or non-positive: %v", resp["expires_in"])
	}

	// The minted nonce must actually be usable by the same admin.
	if !h.archiveHandoff.consume(state, admin.ID) {
		t.Error("just-issued state should be consumable by the same admin")
	}
}
