package store

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// InviteToken represents a one-time account-template invite for onboarding.
// Invites are not tied to a pre-existing user; the recipient chooses their
// own username and password when redeeming.
type InviteToken struct {
	ID             string     `json:"id"`
	TokenHash      string     `json:"-"`
	Role           string     `json:"role"`
	AccessLevel    string     `json:"access_level"`
	ProfileID      *string    `json:"profile_id,omitempty"`
	CreatedBy      string     `json:"created_by"`
	ExpiresAt      time.Time  `json:"expires_at"`
	UsedAt         *time.Time `json:"used_at,omitempty"`
	RedeemedUserID *string    `json:"redeemed_user_id,omitempty"`
	Status         string     `json:"status"`
	// Populated on read via JOIN, not stored on the token itself.
	RedeemedUsername string `json:"redeemed_username,omitempty"`
	CreatedByName    string `json:"created_by_name,omitempty"`
}

// InviteStore manages invite tokens.
type InviteStore struct {
	db *DB
}

func NewInviteStore(db *DB) *InviteStore {
	return &InviteStore{db: db}
}

// DB returns the underlying database for transaction management.
func (s *InviteStore) DB() *DB {
	return s.db
}

func hashInviteToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// CreateAccountInvite generates a new invite token with account-template settings.
// Returns the raw token (shown once to the admin).
func (s *InviteStore) CreateAccountInvite(role, accessLevel string, profileID *string, createdBy string, expiresAt time.Time) (string, *InviteToken, error) {
	rawToken := uuid.New().String()
	tokenHash := hashInviteToken(rawToken)

	if role == "admin" {
		accessLevel = "admin"
	}
	if accessLevel == "" {
		accessLevel = "write"
	}

	t := &InviteToken{
		ID:          uuid.New().String(),
		TokenHash:   tokenHash,
		Role:        role,
		AccessLevel: accessLevel,
		ProfileID:   profileID,
		CreatedBy:   createdBy,
		ExpiresAt:   expiresAt,
		Status:      "pending",
	}

	_, err := s.db.Exec(`
		INSERT INTO invite_tokens (id, token_hash, role, access_level, profile_id, created_by, expires_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.TokenHash, t.Role, t.AccessLevel, t.ProfileID, t.CreatedBy, t.ExpiresAt, t.Status,
	)
	if err != nil {
		return "", nil, fmt.Errorf("creating account invite token: %w", err)
	}
	return rawToken, t, nil
}

// Peek checks whether a raw token is valid (pending and not expired) without consuming it.
// Used to render the browser invite form before the user submits.
func (s *InviteStore) Peek(rawToken string) (*InviteToken, error) {
	tokenHash := hashInviteToken(rawToken)
	now := time.Now()

	t := &InviteToken{}
	var storedHash string
	var profileID sql.NullString
	err := s.db.QueryRow(`
		SELECT id, token_hash, role, access_level, profile_id, created_by, expires_at, status
		FROM invite_tokens
		WHERE token_hash = ? AND status = 'pending' AND expires_at > ?`,
		tokenHash, now,
	).Scan(&t.ID, &storedHash, &t.Role, &t.AccessLevel, &profileID, &t.CreatedBy, &t.ExpiresAt, &t.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("peeking invite token: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(storedHash)) != 1 {
		return nil, nil
	}

	if profileID.Valid {
		t.ProfileID = &profileID.String
	}
	return t, nil
}

// ValidateAndConsumeTx atomically marks a token as used within the given transaction.
// Returns nil if the token is invalid, expired, or already used.
// The caller is responsible for committing or rolling back the transaction.
func (s *InviteStore) ValidateAndConsumeTx(tx *sql.Tx, rawToken string) (*InviteToken, error) {
	tokenHash := hashInviteToken(rawToken)
	now := time.Now()

	result, err := tx.Exec(`
		UPDATE invite_tokens SET status = 'used', used_at = ?
		WHERE token_hash = ? AND status = 'pending' AND expires_at > ?`,
		now, tokenHash, now,
	)
	if err != nil {
		return nil, fmt.Errorf("consuming invite token: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("checking invite token update: %w", err)
	}
	if affected == 0 {
		return nil, nil
	}

	t := &InviteToken{}
	var storedHash string
	var profileID sql.NullString
	err = tx.QueryRow(`
		SELECT id, token_hash, role, access_level, profile_id, created_by, expires_at, used_at, status
		FROM invite_tokens WHERE token_hash = ?`, tokenHash,
	).Scan(&t.ID, &storedHash, &t.Role, &t.AccessLevel, &profileID, &t.CreatedBy, &t.ExpiresAt, &t.UsedAt, &t.Status)
	if err != nil {
		return nil, fmt.Errorf("reading consumed invite token: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(storedHash)) != 1 {
		return nil, nil
	}

	if profileID.Valid {
		t.ProfileID = &profileID.String
	}
	return t, nil
}

// SetRedeemedUserTx records which user redeemed the invite (within the caller's transaction).
func (s *InviteStore) SetRedeemedUserTx(tx *sql.Tx, tokenID, userID string) error {
	_, err := tx.Exec(`UPDATE invite_tokens SET redeemed_user_id = ? WHERE id = ?`, userID, tokenID)
	return err
}

// ListAll returns all invite tokens (admin view) ordered by most recent first.
func (s *InviteStore) ListAll() ([]*InviteToken, error) {
	rows, err := s.db.Query(`
		SELECT it.id, it.role, it.access_level, it.profile_id, it.created_by,
		       COALESCE(cb.username, ''), it.expires_at, it.used_at,
		       it.redeemed_user_id, COALESCE(ru.username, ''), it.status
		FROM invite_tokens it
		LEFT JOIN users cb ON it.created_by = cb.id
		LEFT JOIN users ru ON it.redeemed_user_id = ru.id
		ORDER BY it.expires_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing all invite tokens: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanInviteTokens(rows)
}

// ListPending returns pending (unclaimed) invite tokens for admin view.
func (s *InviteStore) ListPending() ([]*InviteToken, error) {
	rows, err := s.db.Query(`
		SELECT it.id, it.role, it.access_level, it.profile_id, it.created_by,
		       COALESCE(cb.username, ''), it.expires_at, it.used_at,
		       it.redeemed_user_id, COALESCE(ru.username, ''), it.status
		FROM invite_tokens it
		LEFT JOIN users cb ON it.created_by = cb.id
		LEFT JOIN users ru ON it.redeemed_user_id = ru.id
		WHERE it.status = 'pending' AND it.expires_at > ?
		ORDER BY it.expires_at DESC`, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("listing pending invite tokens: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanInviteTokens(rows)
}

func scanInviteTokens(rows *sql.Rows) ([]*InviteToken, error) {
	var tokens []*InviteToken
	for rows.Next() {
		t := &InviteToken{}
		var profileID, redeemedUserID sql.NullString
		if err := rows.Scan(
			&t.ID, &t.Role, &t.AccessLevel, &profileID, &t.CreatedBy,
			&t.CreatedByName, &t.ExpiresAt, &t.UsedAt,
			&redeemedUserID, &t.RedeemedUsername, &t.Status,
		); err != nil {
			return nil, fmt.Errorf("scanning invite token: %w", err)
		}
		if profileID.Valid {
			t.ProfileID = &profileID.String
		}
		if redeemedUserID.Valid {
			t.RedeemedUserID = &redeemedUserID.String
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// Delete removes an invite token.
func (s *InviteStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM invite_tokens WHERE id = ?", id)
	return err
}

// CleanupExpired marks expired pending tokens.
func (s *InviteStore) CleanupExpired() error {
	_, err := s.db.Exec(`
		UPDATE invite_tokens SET status = 'expired'
		WHERE status = 'pending' AND expires_at < ?`, time.Now())
	return err
}
