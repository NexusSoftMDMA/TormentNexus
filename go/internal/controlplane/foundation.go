package controlplane

import (
	"context"
	"time"
)

// --- Component 1: The Resilient LLM Client (Waterfall Routing) ---

type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
}

type LLMResponse struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens int `json:"prompt_tokens"`
		CompTokens   int `json:"completion_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type LLMClient interface {
	// Generate handles automatic fallback routing: NIM -> OpenRouter -> Local
	Generate(ctx context.Context, req LLMRequest) (*LLMResponse, error)
}

// --- Component 2: Dual-Tier Memory Architecture (L1/L2) ---

type MemoryType string

const (
	MemoryWorking   MemoryType = "working"   // Active L1 context or high-heat L2
	MemoryLongTerm  MemoryType = "long_term" // Persistent L2 knowledge
	MemoryArchive   MemoryType = "archive"   // L3 Cold storage
)

// L2Relation represents GraphRAG semantic/structural edges between memories.
type L2Relation struct {
	SourceID     string  `json:"source_id"`
	TargetID     string  `json:"target_id"`
	RelationType string  `json:"relation_type"`
	Weight       float64 `json:"weight"`
}

// L1Scratchpad is ephemeral, fast memory tied to an active goroutine/session
type L1Scratchpad struct {
	SessionID      string            `json:"session_id"`
	Prompt         string            `json:"prompt"`
	ToolOutputs    map[string]string `json:"tool_outputs"`
	ChainOfThought []string          `json:"chain_of_thought"`
	CreatedAt      time.Time         `json:"created_at"`
}

// L2VaultRecord is permanent storage in the on-disk SQLite vector database
type L2VaultRecord struct {
	ID             string     `json:"id"`
	SessionID      string     `json:"session_id"`
	Type           MemoryType `json:"memory_type"`
	Kind           string     `json:"memory_kind"`
	Category       string     `json:"category"`
	Tags           string     `json:"tags,omitempty"`
	SourceURL      string     `json:"source_url,omitempty"`
	Content        string     `json:"content"`
	Importance     float64    `json:"importance"`
	HeatScore      float64    `json:"heat_score"` // 0-100
	Embedding      []float32  `json:"-"`          // sqlite-vec target
	LastAccessedAt time.Time  `json:"last_accessed_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type MemoryVault interface {
	Commit(ctx context.Context, entry L2VaultRecord) error
	SemanticSearch(ctx context.Context, query string, limit int) ([]L2VaultRecord, error)
}

// --- Component 3: SQLite Schema (sqlite-vec) ---

const VectorSchemaSQL = `
-- Enable sqlite-vec extension (must be loaded by the driver)

-- MCP Directory for Layer 1 Tool Routing
CREATE TABLE IF NOT EXISTS mcp_directory (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    parameters TEXT NOT NULL, -- JSON string
    server_name TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Standard Table for hyper-fast tool search embeddings
CREATE TABLE IF NOT EXISTS vec_mcp_directory (
    id TEXT PRIMARY KEY,
    embedding BLOB NOT NULL
);

-- L2 Vault for Semantic Global Memory
CREATE TABLE IF NOT EXISTS l2_vault (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    memory_type TEXT NOT NULL CHECK(memory_type IN ('working', 'long_term', 'archive')),
    memory_kind TEXT NOT NULL DEFAULT 'fact',
    category TEXT NOT NULL DEFAULT 'general',
    tags TEXT,
    source_url TEXT,
    content TEXT NOT NULL,
    importance REAL DEFAULT 0.5,
    heat_score REAL DEFAULT 50.0,
    last_accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Standard Table for context matching embeddings
CREATE TABLE IF NOT EXISTS vec_l2_vault (
    id TEXT PRIMARY KEY,
    embedding BLOB NOT NULL
);

-- Relational edges between L2 memories (GraphRAG support)
CREATE TABLE IF NOT EXISTS l2_relations (
    source_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    relation_type TEXT NOT NULL,
    weight REAL DEFAULT 1.0,
    PRIMARY KEY (source_id, target_id, relation_type),
    FOREIGN KEY (source_id) REFERENCES l2_vault(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES l2_vault(id) ON DELETE CASCADE
);
`

func Now() time.Time {
	return time.Now().UTC()
}
