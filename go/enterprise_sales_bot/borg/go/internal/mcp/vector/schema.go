package vector

import (
	"database/sql"
	"time"
)

type ToolRecord struct {
	ID          string    `json:"id"`
	ServerName  string    `json:"serverName"`
	ToolName    string    `json:"toolName"`
	Description string    `json:"description"`
	SchemaJSON  string    `json:"schemaJson"`
	Category    string    `json:"category"`
	Tags        string    `json:"tags"`
	Source      string    `json:"source"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SearchResult struct {
	Tool    ToolRecord `json:"tool"`
	Score   float64    `json:"score"`
	Rank    int        `json:"rank"`
	Boosted bool       `json:"boosted"`
}

type SearchQuery struct {
	QueryText  string    `json:"queryText"`
	QueryVec   []float32 `json:"queryVec"`
	TopK       int       `json:"topK"`
	MinScore   float64   `json:"minScore"`
	Categories []string  `json:"categories"`
	Tags       []string  `json:"tags"`
}

const ddlTools = `
CREATE TABLE IF NOT EXISTS tools (
    id          TEXT PRIMARY KEY,
    server_name TEXT    NOT NULL,
    tool_name   TEXT    NOT NULL,
    description TEXT    DEFAULT '',
    schema_json TEXT    DEFAULT '{}',
    category    TEXT    DEFAULT '',
    tags        TEXT    DEFAULT '',
    source      TEXT    DEFAULT 'builtin',
    version     TEXT    DEFAULT '1.0.0',
    created_at  DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    updated_at  DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_tools_category ON tools(category);
CREATE INDEX IF NOT EXISTS idx_tools_server   ON tools(server_name);
CREATE INDEX IF NOT EXISTS idx_tools_source   ON tools(source);
`

const ddlEmbeddings = `
CREATE TABLE IF NOT EXISTS tool_embeddings (
    tool_id     TEXT    NOT NULL REFERENCES tools(id) ON DELETE CASCADE,
    model_name  TEXT    NOT NULL DEFAULT 'all-MiniLM-L6-v2',
    dimension   INTEGER NOT NULL DEFAULT 384,
    vector      BLOB    NOT NULL,
    content_src TEXT    NOT NULL DEFAULT 'description',
    created_at  DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    PRIMARY KEY (tool_id, model_name)
);
`

const ddlUsage = `
CREATE TABLE IF NOT EXISTS tool_usage (
    tool_id      TEXT PRIMARY KEY REFERENCES tools(id) ON DELETE CASCADE,
    select_count INTEGER  DEFAULT 0,
    success_rate REAL     DEFAULT 0.0,
    last_used_at DATETIME DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
`

func InitSchema(db *sql.DB) error {
	for _, ddl := range []string{ddlTools, ddlEmbeddings, ddlUsage} {
		if _, err := db.Exec(ddl); err != nil {
			return err
		}
	}
	return nil
}
