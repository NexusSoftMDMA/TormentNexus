/**
 * tormentnexus Database Schema
 * Migrated from TormentNexus PostgreSQL/Drizzle to SQLite/better-sqlite3
 *
 * Tables:
 * - mcp_servers: MCP server configurations (STDIO, SSE, StreamableHTTP)
 * - tools: Tool definitions with optional embeddings
 * - namespaces: Server groupings for organization
 * - endpoints: Public routing endpoints
 * - namespace_server_mappings: M2M namespace-server relationships
 * - namespace_tool_mappings: Tool status and overrides per namespace
 * - api_keys: Authentication tokens
 * - policies: Access control rules (allow/deny)
 * - tool_call_logs: Traffic inspection (McpShark)
 * - saved_scripts: User-saved code snippets
 * - tool_sets: Curated tool collections
 */
// ============================================
// SQL Schema Definitions
// ============================================
export const SQL_SCHEMA = `
-- MCP Servers
CREATE TABLE IF NOT EXISTS mcp_servers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK (type IN ('stdio', 'sse', 'streamable_http')),
    command TEXT,
    args TEXT, -- JSON array
    env TEXT, -- JSON object
    url TEXT,
    bearer_token TEXT,
    headers TEXT, -- JSON object
    description TEXT,
    icon TEXT,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_mcp_servers_name ON mcp_servers(name);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_type ON mcp_servers(type);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_enabled ON mcp_servers(enabled);

-- Tools
CREATE TABLE IF NOT EXISTS tools (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    input_schema TEXT, -- JSON object
    mcp_server_id TEXT NOT NULL,
    embedding TEXT, -- JSON array of floats
    category TEXT,
    tags TEXT, -- JSON array
    usage_count INTEGER NOT NULL DEFAULT 0,
    last_used_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE,
    UNIQUE(name, mcp_server_id)
);

CREATE INDEX IF NOT EXISTS idx_tools_name ON tools(name);
CREATE INDEX IF NOT EXISTS idx_tools_mcp_server_id ON tools(mcp_server_id);
CREATE INDEX IF NOT EXISTS idx_tools_category ON tools(category);
CREATE INDEX IF NOT EXISTS idx_tools_usage_count ON tools(usage_count DESC);

-- Namespaces
CREATE TABLE IF NOT EXISTS namespaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    icon TEXT,
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_namespaces_name ON namespaces(name);

-- Endpoints
CREATE TABLE IF NOT EXISTS endpoints (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    namespace_id TEXT NOT NULL,
    api_key_id TEXT,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_endpoints_path ON endpoints(path);
CREATE INDEX IF NOT EXISTS idx_endpoints_namespace_id ON endpoints(namespace_id);

-- Namespace Server Mappings
CREATE TABLE IF NOT EXISTS namespace_server_mappings (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    mcp_server_id TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE,
    UNIQUE(namespace_id, mcp_server_id)
);

CREATE INDEX IF NOT EXISTS idx_ns_server_map_namespace ON namespace_server_mappings(namespace_id);
CREATE INDEX IF NOT EXISTS idx_ns_server_map_server ON namespace_server_mappings(mcp_server_id);

-- Namespace Tool Mappings
CREATE TABLE IF NOT EXISTS namespace_tool_mappings (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    tool_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'hidden')),
    alias_name TEXT,
    override_description TEXT,
    override_annotations TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE,
    UNIQUE(namespace_id, tool_id)
);


CREATE INDEX IF NOT EXISTS idx_ns_tool_map_namespace ON namespace_tool_mappings(namespace_id);
CREATE INDEX IF NOT EXISTS idx_ns_tool_map_tool ON namespace_tool_mappings(tool_id);
CREATE INDEX IF NOT EXISTS idx_ns_tool_map_status ON namespace_tool_mappings(status);
CREATE INDEX IF NOT EXISTS idx_ns_tool_map_alias ON namespace_tool_mappings(alias_name);


-- API Keys
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    key_prefix TEXT NOT NULL,
    scopes TEXT NOT NULL, -- JSON array
    expires_at INTEGER,
    last_used_at INTEGER,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix);

-- Policies
CREATE TABLE IF NOT EXISTS policies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 100,
    rules TEXT NOT NULL, -- JSON array of PolicyRule
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_policies_priority ON policies(priority);
CREATE INDEX IF NOT EXISTS idx_policies_enabled ON policies(enabled);

-- Tool Call Logs
CREATE TABLE IF NOT EXISTS tool_call_logs (
    id TEXT PRIMARY KEY,
    timestamp INTEGER NOT NULL,
    tool_name TEXT NOT NULL,
    mcp_server_id TEXT,
    namespace_id TEXT,
    endpoint_id TEXT,
    api_key_id TEXT,
    request_args TEXT, -- JSON object
    response_data TEXT, -- JSON (can be large)
    error_message TEXT,
    duration_ms INTEGER NOT NULL,
    success INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tool_call_logs_timestamp ON tool_call_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_tool_call_logs_tool_name ON tool_call_logs(tool_name);
CREATE INDEX IF NOT EXISTS idx_tool_call_logs_mcp_server_id ON tool_call_logs(mcp_server_id);
CREATE INDEX IF NOT EXISTS idx_tool_call_logs_success ON tool_call_logs(success);

-- Tool Annotations
CREATE TABLE IF NOT EXISTS tool_annotations (
    id TEXT PRIMARY KEY,
    namespace_id TEXT,
    mcp_server_name TEXT NOT NULL,
    tool_name TEXT NOT NULL,
    data TEXT NOT NULL,
    updated_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    UNIQUE(namespace_id, mcp_server_name, tool_name)
);

CREATE INDEX IF NOT EXISTS idx_tool_annotations_namespace ON tool_annotations(namespace_id);
CREATE INDEX IF NOT EXISTS idx_tool_annotations_server_tool ON tool_annotations(mcp_server_name, tool_name);

-- Saved Scripts
CREATE TABLE IF NOT EXISTS saved_scripts (

    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    language TEXT NOT NULL CHECK (language IN ('typescript', 'javascript', 'python')),
    code TEXT NOT NULL,
    tags TEXT, -- JSON array
    is_favorite INTEGER NOT NULL DEFAULT 0,
    run_count INTEGER NOT NULL DEFAULT 0,
    last_run_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_saved_scripts_name ON saved_scripts(name);
CREATE INDEX IF NOT EXISTS idx_saved_scripts_language ON saved_scripts(language);
CREATE INDEX IF NOT EXISTS idx_saved_scripts_is_favorite ON saved_scripts(is_favorite);

-- Tool Sets
CREATE TABLE IF NOT EXISTS tool_sets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    tool_ids TEXT NOT NULL, -- JSON array
    icon TEXT,
    is_public INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tool_sets_name ON tool_sets(name);
CREATE INDEX IF NOT EXISTS idx_tool_sets_is_public ON tool_sets(is_public);

-- Default namespace (inserted if not exists)
INSERT OR IGNORE INTO namespaces (id, name, description, is_default, created_at, updated_at)
VALUES ('default', 'Default', 'Default namespace for all servers', 1, strftime('%s', 'now') * 1000, strftime('%s', 'now') * 1000);
`;
// ============================================
// Helper Functions
// ============================================
export function generateId() {
    return `${Date.now().toString(36)}-${Math.random().toString(36).substring(2, 9)}`;
}
export function toDbRow(obj) {
    const row = {};
    for (const [key, value] of Object.entries(obj)) {
        // Convert camelCase to snake_case
        const snakeKey = key.replace(/[A-Z]/g, letter => `_${letter.toLowerCase()}`);
        // Serialize arrays and objects to JSON
        if (Array.isArray(value) || (typeof value === 'object' && value !== null)) {
            row[snakeKey] = JSON.stringify(value);
        }
        else if (typeof value === 'boolean') {
            row[snakeKey] = value ? 1 : 0;
        }
        else {
            row[snakeKey] = value;
        }
    }
    return row;
}
export function fromDbRow(row, jsonFields = []) {
    const obj = {};
    for (const [key, value] of Object.entries(row)) {
        // Convert snake_case to camelCase
        const camelKey = key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
        // Parse JSON fields
        if (jsonFields.includes(camelKey) && typeof value === 'string') {
            try {
                obj[camelKey] = JSON.parse(value);
            }
            catch {
                obj[camelKey] = value;
            }
        }
        else if (typeof value === 'number' && (camelKey.startsWith('is') || camelKey === 'enabled' || camelKey === 'success')) {
            obj[camelKey] = value === 1;
        }
        else {
            obj[camelKey] = value;
        }
    }
    return obj;
}
