package vault

import (
	"database/sql"
	"time"
)

type MemoryType string

const (
	MemoryRaw       MemoryType = "raw"
	MemoryHeuristic MemoryType = "heuristic"
)

type SessionStatus string

const (
	StatusActive    SessionStatus = "active"
	StatusCompleted SessionStatus = "completed"
	StatusAborted   SessionStatus = "aborted"
)

type Session struct {
	ID          string        `json:"id"`
	AgentName   string        `json:"agentName"`
	Project     string        `json:"project"`
	Status      SessionStatus `json:"status"`
	Prompt      string        `json:"prompt"`
	SystemCtx   string        `json:"systemCtx"`
	CreatedAt   time.Time     `json:"createdAt"`
	CompletedAt *time.Time    `json:"completedAt,omitempty"`
}

type ScratchEntry struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"sessionId"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	ToolName   string    `json:"toolName,omitempty"`
	ToolArgs   string    `json:"toolArgs,omitempty"`
	ToolResult string    `json:"toolResult,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type VaultEntry struct {
	ID           string     `json:"id"`
	MemoryType   MemoryType `json:"memoryType"`
	SessionID    string     `json:"sessionId"`
	AgentName    string     `json:"agentName"`
	Project      string     `json:"project"`
	Content      string     `json:"content"`
	Tags         string     `json:"tags"`
	Importance   float64    `json:"importance"`
	SourceModels string     `json:"sourceModels"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type ContextHit struct {
	Entry VaultEntry `json:"entry"`
	Score float64    `json:"score"`
	Rank  int        `json:"rank"`
}

const ddlL1 = `
CREATE TABLE IF NOT EXISTS l1_sessions (
    id           TEXT PRIMARY KEY,
    agent_name   TEXT    NOT NULL,
    project      TEXT    DEFAULT '',
    status       TEXT    DEFAULT 'active',
    prompt       TEXT    DEFAULT '',
    system_ctx   TEXT    DEFAULT '',
    created_at   DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    completed_at DATETIME DEFAULT NULL
);
CREATE TABLE IF NOT EXISTS l1_scratch (
    id          TEXT PRIMARY KEY,
    session_id  TEXT    NOT NULL REFERENCES l1_sessions(id) ON DELETE CASCADE,
    role        TEXT    NOT NULL,
    content     TEXT    DEFAULT '',
    tool_name   TEXT    DEFAULT '',
    tool_args   TEXT    DEFAULT '',
    tool_result TEXT    DEFAULT '',
    created_at  DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_l1_scratch_session ON l1_scratch(session_id);
`

const ddlL2 = `
CREATE TABLE IF NOT EXISTS l2_entries (
    id            TEXT PRIMARY KEY,
    memory_type   TEXT    NOT NULL CHECK(memory_type IN ('raw', 'heuristic')),
    session_id    TEXT    NOT NULL,
    agent_name    TEXT    DEFAULT '',
    project       TEXT    DEFAULT '',
    content       TEXT    NOT NULL,
    tags          TEXT    DEFAULT '',
    importance    REAL    DEFAULT 0.5,
    source_models TEXT    DEFAULT '',
    created_at    DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_l2_type       ON l2_entries(memory_type);
CREATE INDEX IF NOT EXISTS idx_l2_session    ON l2_entries(session_id);
CREATE INDEX IF NOT EXISTS idx_l2_project    ON l2_entries(project);
CREATE INDEX IF NOT EXISTS idx_l2_importance ON l2_entries(importance);
CREATE TABLE IF NOT EXISTS l2_embeddings (
    entry_id   TEXT    NOT NULL REFERENCES l2_entries(id) ON DELETE CASCADE,
    model_name TEXT    NOT NULL DEFAULT 'all-MiniLM-L6-v2',
    dimension  INTEGER NOT NULL DEFAULT 384,
    vector     BLOB    NOT NULL,
    created_at DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    PRIMARY KEY (entry_id, model_name)
);
`

func InitL1Schema(db *sql.DB) error {
	_, err := db.Exec(ddlL1)
	return err
}

func InitL2Schema(db *sql.DB) error {
	_, err := db.Exec(ddlL2)
	return err
}
