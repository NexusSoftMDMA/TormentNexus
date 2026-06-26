-- Arc Relay initial schema

CREATE TABLE IF NOT EXISTS servers (
    id           TEXT PRIMARY KEY,
    name         TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    server_type  TEXT NOT NULL CHECK(server_type IN ('stdio', 'http', 'remote')),
    config       TEXT NOT NULL DEFAULT '{}',
    status       TEXT NOT NULL DEFAULT 'stopped' CHECK(status IN ('stopped', 'starting', 'running', 'error')),
    error_msg    TEXT,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin', 'user')),
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_keys (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash   TEXT NOT NULL,
    name       TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used  DATETIME,
    revoked    BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS endpoint_permissions (
    id            TEXT PRIMARY KEY,
    server_id     TEXT NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    endpoint_type TEXT NOT NULL CHECK(endpoint_type IN ('tool', 'resource', 'prompt')),
    endpoint_name TEXT NOT NULL,
    user_id       TEXT REFERENCES users(id) ON DELETE CASCADE,
    allowed       BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE(server_id, endpoint_type, endpoint_name, user_id)
);

CREATE TABLE IF NOT EXISTS request_logs (
    id            TEXT PRIMARY KEY,
    timestamp     DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_id       TEXT REFERENCES users(id),
    server_id     TEXT REFERENCES servers(id),
    method        TEXT NOT NULL,
    endpoint_name TEXT,
    duration_ms   INTEGER,
    status        TEXT,
    error_msg     TEXT
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_endpoint_permissions_server ON endpoint_permissions(server_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_timestamp ON request_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_request_logs_server ON request_logs(server_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_user ON request_logs(user_id);
