package store

import (
	"crypto/subtle"
	"database/sql"
	"fmt"
	"time"
)

// OAuthTokenStore manages OAuth 2.1 access tokens for the MCP proxy.
// These are separate from API keys and only grant access to /mcp/ routes.
type OAuthTokenStore struct {
	db *DB
}

func NewOAuthTokenStore(db *DB) *OAuthTokenStore {
	return &OAuthTokenStore{db: db}
}

// DB returns the underlying database connection for sharing with related stores.
func (s *OAuthTokenStore) DB() *DB {
	return s.db
}

// Create stores a new OAuth token hash.
func (s *OAuthTokenStore) Create(tokenHash, userID, clientID, scope, resource string, expiresAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO oauth_tokens (token_hash, user_id, client_id, scope, resource, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		tokenHash, userID, clientID, scope, resource, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating oauth token: %w", err)
	}
	return nil
}

// Validate checks a raw OAuth token and returns the associated user, or nil.
// Uses the same profile resolution as API keys (user's DefaultProfileID).
func (s *OAuthTokenStore) Validate(rawToken string) (*User, error) {
	tokenHash := hashAPIKey(rawToken) // reuse SHA256 helper

	var storedHash string
	var userID string
	var expiresAt time.Time
	err := s.db.QueryRow(`
		SELECT token_hash, user_id, expires_at FROM oauth_tokens WHERE token_hash = ?`,
		tokenHash,
	).Scan(&storedHash, &userID, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("looking up oauth token: %w", err)
	}

	// Constant-time comparison
	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(storedHash)) != 1 {
		return nil, nil
	}

	// Check expiry
	if time.Now().After(expiresAt) {
		return nil, nil
	}

	// Look up user with profile resolution
	var user User
	var defaultProfileID sql.NullString
	err = s.db.QueryRow(`
		SELECT id, username, role, access_level, default_profile_id
		FROM users WHERE id = ?`, userID,
	).Scan(&user.ID, &user.Username, &user.Role, &user.AccessLevel, &defaultProfileID)
	if err != nil {
		return nil, nil
	}
	if defaultProfileID.Valid {
		user.DefaultProfileID = &defaultProfileID.String
		user.ProfileID = &defaultProfileID.String
	}

	return &user, nil
}

// Cleanup removes all expired OAuth tokens.
func (s *OAuthTokenStore) Cleanup() {
	_, _ = s.db.Exec(`DELETE FROM oauth_tokens WHERE expires_at < ?`, time.Now())
}
