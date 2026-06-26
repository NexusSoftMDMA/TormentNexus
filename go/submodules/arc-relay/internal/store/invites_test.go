package store_test

import (
	"testing"
	"time"

	"github.com/comma-compliance/arc-relay/internal/store"
	"github.com/comma-compliance/arc-relay/internal/testutil"
)

func TestCreateAccountInvite(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, err := users.Create("admin", "pass", "admin")
	if err != nil {
		t.Fatalf("creating admin: %v", err)
	}

	rawToken, tok, err := invites.CreateAccountInvite("user", "write", nil, admin.ID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateAccountInvite() error = %v", err)
	}
	if rawToken == "" {
		t.Error("rawToken should not be empty")
	}
	if tok.ID == "" {
		t.Error("token ID should be generated")
	}
	if tok.Status != "pending" {
		t.Errorf("Status = %q, want %q", tok.Status, "pending")
	}
	if tok.Role != "user" {
		t.Errorf("Role = %q, want %q", tok.Role, "user")
	}
	if tok.AccessLevel != "write" {
		t.Errorf("AccessLevel = %q, want %q", tok.AccessLevel, "write")
	}
}

func TestPeekDoesNotConsume(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, _ := users.Create("admin", "pass", "admin")
	rawToken, _, _ := invites.CreateAccountInvite("user", "write", nil, admin.ID, time.Now().Add(time.Hour))

	// Peek should return the token details
	peeked, err := invites.Peek(rawToken)
	if err != nil {
		t.Fatalf("Peek() error = %v", err)
	}
	if peeked == nil {
		t.Fatal("Peek() returned nil for valid token")
	}
	if peeked.Status != "pending" {
		t.Errorf("Peek Status = %q, want %q", peeked.Status, "pending")
	}

	// Peek again - should still work (not consumed)
	peeked2, err := invites.Peek(rawToken)
	if err != nil {
		t.Fatalf("second Peek() error = %v", err)
	}
	if peeked2 == nil {
		t.Fatal("second Peek() returned nil - peek should not consume")
	}
}

func TestValidateAndConsumeTx(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, _ := users.Create("admin", "pass", "admin")
	rawToken, _, _ := invites.CreateAccountInvite("user", "write", nil, admin.ID, time.Now().Add(time.Hour))

	tx, err := invites.DB().Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	consumed, err := invites.ValidateAndConsumeTx(tx, rawToken)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("ValidateAndConsumeTx() error = %v", err)
	}
	if consumed == nil {
		_ = tx.Rollback()
		t.Fatal("ValidateAndConsumeTx() returned nil for valid token")
	}
	if consumed.Status != "used" {
		t.Errorf("Status = %q, want %q", consumed.Status, "used")
	}
	if consumed.Role != "user" {
		t.Errorf("Role = %q, want %q", consumed.Role, "user")
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Second consume should fail
	tx2, _ := invites.DB().Begin()
	consumed2, err := invites.ValidateAndConsumeTx(tx2, rawToken)
	if err != nil {
		_ = tx2.Rollback()
		t.Fatalf("second ValidateAndConsumeTx() error = %v", err)
	}
	_ = tx2.Rollback()
	if consumed2 != nil {
		t.Error("second ValidateAndConsumeTx() should return nil - token already consumed")
	}
}

func TestValidateAndConsumeTx_RollbackPreservesToken(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, _ := users.Create("admin", "pass", "admin")
	rawToken, _, _ := invites.CreateAccountInvite("user", "write", nil, admin.ID, time.Now().Add(time.Hour))

	// Consume in a transaction, then rollback
	tx, _ := invites.DB().Begin()
	consumed, _ := invites.ValidateAndConsumeTx(tx, rawToken)
	if consumed == nil {
		_ = tx.Rollback()
		t.Fatal("ValidateAndConsumeTx() returned nil")
	}
	_ = tx.Rollback()

	// Token should still be available after rollback
	peeked, err := invites.Peek(rawToken)
	if err != nil {
		t.Fatalf("Peek() after rollback error = %v", err)
	}
	if peeked == nil {
		t.Fatal("Peek() returned nil after rollback - token should still be pending")
	}
}

func TestExpiredToken(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, _ := users.Create("admin", "pass", "admin")
	rawToken, _, _ := invites.CreateAccountInvite("user", "write", nil, admin.ID, time.Now().Add(-time.Hour))

	peeked, _ := invites.Peek(rawToken)
	if peeked != nil {
		t.Error("Peek() should return nil for expired token")
	}

	tx, _ := invites.DB().Begin()
	consumed, _ := invites.ValidateAndConsumeTx(tx, rawToken)
	_ = tx.Rollback()
	if consumed != nil {
		t.Error("ValidateAndConsumeTx() should return nil for expired token")
	}
}

func TestCleanupExpired(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, _ := users.Create("admin", "pass", "admin")
	_, _, _ = invites.CreateAccountInvite("user", "write", nil, admin.ID, time.Now().Add(-time.Hour))

	if err := invites.CleanupExpired(); err != nil {
		t.Fatalf("CleanupExpired() error = %v", err)
	}

	tokens, err := invites.ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token after cleanup, got %d", len(tokens))
	}
	if tokens[0].Status != "expired" {
		t.Errorf("Status = %q, want %q", tokens[0].Status, "expired")
	}
}

func TestListPending(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, _ := users.Create("admin", "pass", "admin")
	// 2 pending
	_, _, _ = invites.CreateAccountInvite("user", "write", nil, admin.ID, time.Now().Add(time.Hour))
	_, _, _ = invites.CreateAccountInvite("admin", "admin", nil, admin.ID, time.Now().Add(2*time.Hour))
	// 1 expired
	_, _, _ = invites.CreateAccountInvite("user", "read", nil, admin.ID, time.Now().Add(-time.Hour))

	pending, err := invites.ListPending()
	if err != nil {
		t.Fatalf("ListPending() error = %v", err)
	}
	if len(pending) != 2 {
		t.Errorf("ListPending() returned %d, want 2", len(pending))
	}
}

func TestAdminInviteForcesAdminAccessLevel(t *testing.T) {
	db := testutil.OpenTestDB(t)
	users := store.NewUserStore(db)
	invites := store.NewInviteStore(db)

	admin, _ := users.Create("admin", "pass", "admin")

	// Creating an admin invite with "write" access should be forced to "admin"
	_, tok, err := invites.CreateAccountInvite("admin", "write", nil, admin.ID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateAccountInvite() error = %v", err)
	}
	if tok.AccessLevel != "admin" {
		t.Errorf("AccessLevel = %q, want %q for admin role", tok.AccessLevel, "admin")
	}
}
