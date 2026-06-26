-- OAuth clients registered via Dynamic Client Registration (RFC 7591).
-- Persisted so clients survive server restarts.
CREATE TABLE IF NOT EXISTS oauth_clients (
    client_id            TEXT PRIMARY KEY,
    client_secret_hash   TEXT NOT NULL DEFAULT '',
    client_name          TEXT NOT NULL DEFAULT '',
    redirect_uris        TEXT NOT NULL DEFAULT '[]',
    token_endpoint_auth_method TEXT NOT NULL DEFAULT 'none',
    created_at           DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- OAuth refresh tokens for token rotation.
CREATE TABLE IF NOT EXISTS oauth_refresh_tokens (
    token_hash  TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id   TEXT NOT NULL,
    scope       TEXT NOT NULL DEFAULT 'mcp',
    resource    TEXT NOT NULL DEFAULT '',
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_user ON oauth_refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_refresh_expires ON oauth_refresh_tokens(expires_at);
