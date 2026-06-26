package store

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// OAuthClient represents a registered OAuth client.
type OAuthClient struct {
	ClientID                string   `json:"client_id"`
	ClientSecretHash        string   `json:"-"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	CreatedAt               time.Time
}

// HasRedirectURI checks if the given URI is registered for this client.
func (c *OAuthClient) HasRedirectURI(uri string) bool {
	for _, registered := range c.RedirectURIs {
		if registered == uri {
			return true
		}
	}
	return false
}

// OAuthClientStore manages OAuth client registrations in SQLite.
type OAuthClientStore struct {
	db *DB
}

func NewOAuthClientStore(db *DB) *OAuthClientStore {
	return &OAuthClientStore{db: db}
}

// Create registers a new OAuth client.
func (s *OAuthClientStore) Create(clientID, clientSecretHash, clientName, authMethod string, redirectURIs []string) error {
	urisJSON, _ := json.Marshal(redirectURIs)
	_, err := s.db.Exec(`
		INSERT INTO oauth_clients (client_id, client_secret_hash, client_name, redirect_uris, token_endpoint_auth_method)
		VALUES (?, ?, ?, ?, ?)`,
		clientID, clientSecretHash, clientName, string(urisJSON), authMethod,
	)
	if err != nil {
		return fmt.Errorf("creating oauth client: %w", err)
	}
	return nil
}

// Get retrieves a client by ID, or nil if not found.
func (s *OAuthClientStore) Get(clientID string) *OAuthClient {
	var client OAuthClient
	var urisJSON string
	err := s.db.QueryRow(`
		SELECT client_id, client_secret_hash, client_name, redirect_uris, token_endpoint_auth_method, created_at
		FROM oauth_clients WHERE client_id = ?`, clientID,
	).Scan(&client.ClientID, &client.ClientSecretHash, &client.ClientName, &urisJSON, &client.TokenEndpointAuthMethod, &client.CreatedAt)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return nil
	}
	_ = json.Unmarshal([]byte(urisJSON), &client.RedirectURIs)
	return &client
}

// ValidateSecret checks a client's secret. Returns true for public clients (no secret).
func (s *OAuthClientStore) ValidateSecret(clientID, secret string) bool {
	client := s.Get(clientID)
	if client == nil {
		return false
	}
	if client.ClientSecretHash == "" {
		return true // public client
	}
	h := sha256.Sum256([]byte(secret))
	provided := hex.EncodeToString(h[:])
	return subtle.ConstantTimeCompare([]byte(provided), []byte(client.ClientSecretHash)) == 1
}

// OAuthRefreshTokenStore manages OAuth refresh tokens in SQLite.
type OAuthRefreshTokenStore struct {
	db *DB
}

func NewOAuthRefreshTokenStore(db *DB) *OAuthRefreshTokenStore {
	return &OAuthRefreshTokenStore{db: db}
}

// Create stores a new refresh token hash.
func (s *OAuthRefreshTokenStore) Create(tokenHash, userID, clientID, scope, resource string, expiresAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO oauth_refresh_tokens (token_hash, user_id, client_id, scope, resource, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		tokenHash, userID, clientID, scope, resource, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating refresh token: %w", err)
	}
	return nil
}

// Consume validates and deletes a refresh token (single-use rotation).
// Returns the token metadata, or nil if invalid/expired.
func (s *OAuthRefreshTokenStore) Consume(rawToken string) *OAuthRefreshTokenData {
	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	var data OAuthRefreshTokenData
	err := s.db.QueryRow(`
		SELECT token_hash, user_id, client_id, scope, resource, expires_at
		FROM oauth_refresh_tokens WHERE token_hash = ?`, tokenHash,
	).Scan(&data.TokenHash, &data.UserID, &data.ClientID, &data.Scope, &data.Resource, &data.ExpiresAt)
	if err != nil {
		return nil
	}

	if time.Now().After(data.ExpiresAt) {
		_, _ = s.db.Exec(`DELETE FROM oauth_refresh_tokens WHERE token_hash = ?`, tokenHash)
		return nil
	}

	// Delete (single-use)
	_, _ = s.db.Exec(`DELETE FROM oauth_refresh_tokens WHERE token_hash = ?`, tokenHash)
	return &data
}

// Cleanup removes expired refresh tokens.
func (s *OAuthRefreshTokenStore) Cleanup() {
	_, _ = s.db.Exec(`DELETE FROM oauth_refresh_tokens WHERE expires_at < ?`, time.Now())
}

// OAuthRefreshTokenData holds the metadata from a consumed refresh token.
type OAuthRefreshTokenData struct {
	TokenHash string
	UserID    string
	ClientID  string
	Scope     string
	Resource  string
	ExpiresAt time.Time
}
