-- Durable archive queue for retry and compliance delivery guarantees
CREATE TABLE IF NOT EXISTS archive_queue (
    id              TEXT PRIMARY KEY,
    server_id       TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    payload         TEXT NOT NULL,
    url             TEXT NOT NULL,
    auth_type       TEXT NOT NULL DEFAULT 'none',
    auth_value      TEXT NOT NULL DEFAULT '',
    api_key_header  TEXT NOT NULL DEFAULT 'X-API-Key',
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending, hold
    attempts        INTEGER NOT NULL DEFAULT 0,
    next_attempt_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_attempt_at DATETIME,
    last_error      TEXT
);

CREATE INDEX IF NOT EXISTS idx_archive_queue_status_next
    ON archive_queue(status, next_attempt_at);
CREATE INDEX IF NOT EXISTS idx_archive_queue_server_status
    ON archive_queue(server_id, status, created_at);
