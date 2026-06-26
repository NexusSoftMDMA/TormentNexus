-- Middleware pipeline configuration
CREATE TABLE IF NOT EXISTS middleware_configs (
    id          TEXT PRIMARY KEY,
    server_id   TEXT REFERENCES servers(id) ON DELETE CASCADE,  -- NULL = global default
    middleware  TEXT NOT NULL,                                   -- middleware name (e.g. 'sanitizer', 'sizer', 'alerter')
    enabled     BOOLEAN DEFAULT TRUE,
    config      TEXT NOT NULL DEFAULT '{}',                     -- JSON config
    priority    INTEGER DEFAULT 100,                            -- execution order (lower = first)
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(server_id, middleware)
);

-- Index for fast lookups by server
CREATE INDEX IF NOT EXISTS idx_middleware_configs_server ON middleware_configs(server_id);

-- Middleware event log for alerter and audit trail
CREATE TABLE IF NOT EXISTS middleware_events (
    id          TEXT PRIMARY KEY,
    timestamp   DATETIME DEFAULT CURRENT_TIMESTAMP,
    server_id   TEXT REFERENCES servers(id) ON DELETE CASCADE,
    middleware  TEXT NOT NULL,
    event_type  TEXT NOT NULL,      -- 'redacted', 'blocked', 'truncated', 'alert'
    summary     TEXT NOT NULL,      -- human-readable description
    request_method TEXT,            -- e.g. 'tools/call'
    endpoint_name  TEXT,
    user_id     TEXT
);

CREATE INDEX IF NOT EXISTS idx_middleware_events_server ON middleware_events(server_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_middleware_events_time ON middleware_events(timestamp DESC);
