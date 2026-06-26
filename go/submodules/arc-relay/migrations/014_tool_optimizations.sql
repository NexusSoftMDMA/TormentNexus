-- Tool context optimization: stores LLM-optimized tool definitions per server
CREATE TABLE IF NOT EXISTS tool_optimizations (
    id              TEXT PRIMARY KEY,
    server_id       TEXT NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    tools_hash      TEXT NOT NULL,          -- SHA-256 of original tools/list payload
    original_chars  INTEGER NOT NULL,       -- total chars of original tool definitions
    optimized_chars INTEGER NOT NULL,       -- total chars after optimization
    optimized_tools TEXT NOT NULL,          -- JSON array of optimized {name, description, inputSchema}
    prompt_version  TEXT NOT NULL,          -- version hash of the system prompt used
    model           TEXT NOT NULL,          -- LLM model used for optimization
    status          TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'running', 'ready', 'stale', 'error')),
    error_msg       TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(server_id)
);

-- Per-server toggle: whether to serve optimized tools via the proxy
ALTER TABLE servers ADD COLUMN optimize_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_tool_optimizations_server ON tool_optimizations(server_id);
