-- Fix request_logs foreign keys: add ON DELETE SET NULL so server/user
-- deletions are not blocked by log entries.  SQLite requires recreating
-- the table to alter column constraints.

CREATE TABLE IF NOT EXISTS request_logs_new (
    id            TEXT PRIMARY KEY,
    timestamp     DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_id       TEXT REFERENCES users(id) ON DELETE SET NULL,
    server_id     TEXT REFERENCES servers(id) ON DELETE SET NULL,
    method        TEXT NOT NULL,
    endpoint_name TEXT,
    duration_ms   INTEGER,
    status        TEXT,
    error_msg     TEXT
);

INSERT INTO request_logs_new SELECT * FROM request_logs;
DROP TABLE request_logs;
ALTER TABLE request_logs_new RENAME TO request_logs;

CREATE INDEX IF NOT EXISTS idx_request_logs_timestamp ON request_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_request_logs_server ON request_logs(server_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_user ON request_logs(user_id);
