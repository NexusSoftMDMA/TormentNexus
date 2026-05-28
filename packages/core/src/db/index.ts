/**
 * @file index.ts
 * @module packages/core/src/db/index
 *
 * WHAT:
 * Database connection initialization for Drizzle ORM.
 *
 * WHY:
 * Provides a singleton `db` instance used by all repositories.
 * Supports switching between SQLite (dev) and PostgreSQL (prod) via env vars,
 * though predominantly targets SQLite for this local-first architecture.
 */

import { drizzle } from "drizzle-orm/better-sqlite3";
import Database from "better-sqlite3";
import * as schema from "./mcp-admin-schema.js";
import * as dotenv from "dotenv";
import path from "path";
import fs from "fs";

dotenv.config();

function initializeSchema(database: InstanceType<typeof Database>): void {
    database.pragma("foreign_keys = ON");

    database.exec(`
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            email_verified INTEGER NOT NULL DEFAULT 0,
            image TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );

        CREATE TABLE IF NOT EXISTS sessions (
            id TEXT PRIMARY KEY,
            expires_at INTEGER NOT NULL,
            token TEXT NOT NULL UNIQUE,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            ip_address TEXT,
            user_agent TEXT,
            user_id TEXT NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS accounts (
            id TEXT PRIMARY KEY,
            account_id TEXT NOT NULL,
            provider_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            access_token TEXT,
            refresh_token TEXT,
            id_token TEXT,
            access_token_expires_at INTEGER,
            refresh_token_expires_at INTEGER,
            scope TEXT,
            password TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS verifications (
            id TEXT PRIMARY KEY,
            identifier TEXT NOT NULL,
            value TEXT NOT NULL,
            expires_at INTEGER NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );

        CREATE TABLE IF NOT EXISTS mcp_servers (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            type TEXT NOT NULL DEFAULT 'STDIO',
            command TEXT,
            args TEXT NOT NULL DEFAULT '[]',
            env TEXT NOT NULL DEFAULT '{}',
            url TEXT,
            error_status TEXT NOT NULL DEFAULT 'NONE',
            always_on INTEGER NOT NULL DEFAULT 0,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            bearer_token TEXT,
            headers TEXT NOT NULL DEFAULT '{}',
            user_id TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS oauth_sessions (
            uuid TEXT PRIMARY KEY,
            mcp_server_uuid TEXT NOT NULL,
            client_information TEXT NOT NULL DEFAULT '{}',
            tokens TEXT,
            code_verifier TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS tools (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            tool_schema TEXT NOT NULL,
            is_deferred INTEGER NOT NULL DEFAULT 0,
            always_on INTEGER NOT NULL DEFAULT 0,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            mcp_server_uuid TEXT NOT NULL,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS namespaces (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS endpoints (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            namespace_uuid TEXT NOT NULL,
            enable_api_key_auth INTEGER NOT NULL DEFAULT 1,
            enable_oauth INTEGER NOT NULL DEFAULT 0,
            enable_max_rate INTEGER NOT NULL DEFAULT 0,
            enable_client_max_rate INTEGER NOT NULL DEFAULT 0,
            max_rate INTEGER,
            max_rate_seconds INTEGER,
            client_max_rate INTEGER,
            client_max_rate_seconds INTEGER,
            client_max_rate_strategy TEXT,
            client_max_rate_strategy_key TEXT,
            use_query_param_auth INTEGER NOT NULL DEFAULT 0,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS namespace_server_mappings (
            uuid TEXT PRIMARY KEY,
            namespace_uuid TEXT NOT NULL,
            mcp_server_uuid TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'ACTIVE',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE CASCADE,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS namespace_tool_mappings (
            uuid TEXT PRIMARY KEY,
            namespace_uuid TEXT NOT NULL,
            tool_uuid TEXT NOT NULL,
            mcp_server_uuid TEXT NOT NULL,
            status TEXT NOT NULL DEFAULT 'ACTIVE',
            override_name TEXT,
            override_title TEXT,
            override_description TEXT,
            override_annotations TEXT DEFAULT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE CASCADE,
            FOREIGN KEY (tool_uuid) REFERENCES tools(uuid) ON DELETE CASCADE,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS api_keys (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            key TEXT NOT NULL UNIQUE,
            user_id TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            is_active INTEGER NOT NULL DEFAULT 1,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS config (
            id TEXT PRIMARY KEY,
            value TEXT NOT NULL,
            description TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );

        CREATE TABLE IF NOT EXISTS oauth_clients (
            client_id TEXT PRIMARY KEY,
            client_secret TEXT,
            client_name TEXT NOT NULL,
            redirect_uris TEXT NOT NULL DEFAULT '[]',
            grant_types TEXT NOT NULL DEFAULT '["authorization_code","refresh_token"]',
            response_types TEXT NOT NULL DEFAULT '["code"]',
            token_endpoint_auth_method TEXT NOT NULL DEFAULT 'none',
            scope TEXT DEFAULT 'admin',
            client_uri TEXT,
            logo_uri TEXT,
            contacts TEXT,
            tos_uri TEXT,
            policy_uri TEXT,
            software_id TEXT,
            software_version TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );

        CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
            code TEXT PRIMARY KEY,
            client_id TEXT NOT NULL,
            redirect_uri TEXT NOT NULL,
            scope TEXT NOT NULL DEFAULT 'admin',
            user_id TEXT NOT NULL,
            code_challenge TEXT,
            code_challenge_method TEXT,
            expires_at INTEGER NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS oauth_access_tokens (
            access_token TEXT PRIMARY KEY,
            client_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            scope TEXT NOT NULL DEFAULT 'admin',
            expires_at INTEGER NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS docker_sessions (
            uuid TEXT PRIMARY KEY,
            mcp_server_uuid TEXT NOT NULL,
            container_id TEXT NOT NULL,
            container_name TEXT,
            url TEXT,
            status TEXT NOT NULL DEFAULT 'PENDING',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            started_at INTEGER,
            stopped_at INTEGER,
            error_message TEXT,
            retry_count INTEGER NOT NULL DEFAULT 0,
            last_retry_at INTEGER,
            max_retries INTEGER NOT NULL DEFAULT 3,
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS policies (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            rules TEXT NOT NULL DEFAULT '{}',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );

        CREATE TABLE IF NOT EXISTS tool_call_logs (
            uuid TEXT PRIMARY KEY,
            tool_name TEXT NOT NULL,
            mcp_server_uuid TEXT,
            namespace_uuid TEXT,
            endpoint_uuid TEXT,
            args TEXT,
            result TEXT,
            error TEXT,
            duration_ms INTEGER,
            session_id TEXT,
            parent_call_uuid TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (mcp_server_uuid) REFERENCES mcp_servers(uuid) ON DELETE SET NULL,
            FOREIGN KEY (namespace_uuid) REFERENCES namespaces(uuid) ON DELETE SET NULL,
            FOREIGN KEY (endpoint_uuid) REFERENCES endpoints(uuid) ON DELETE SET NULL
        );

        CREATE TABLE IF NOT EXISTS tool_sets (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS tool_set_items (
            uuid TEXT PRIMARY KEY,
            tool_set_uuid TEXT NOT NULL,
            tool_uuid TEXT NOT NULL,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (tool_set_uuid) REFERENCES tool_sets(uuid) ON DELETE CASCADE,
            FOREIGN KEY (tool_uuid) REFERENCES tools(uuid) ON DELETE CASCADE
        );

        CREATE TABLE IF NOT EXISTS saved_scripts (
            uuid TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            code TEXT NOT NULL,
            language TEXT NOT NULL DEFAULT 'javascript',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            user_id TEXT,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );

        CREATE INDEX IF NOT EXISTS mcp_servers_name_idx ON mcp_servers(name);
        CREATE INDEX IF NOT EXISTS mcp_servers_user_id_idx ON mcp_servers(user_id);
        CREATE UNIQUE INDEX IF NOT EXISTS mcp_servers_name_user_unique ON mcp_servers(name, user_id);

        CREATE INDEX IF NOT EXISTS oauth_sessions_mcp_server_uuid_idx ON oauth_sessions(mcp_server_uuid);
        CREATE UNIQUE INDEX IF NOT EXISTS oauth_sessions_unique_per_server_idx ON oauth_sessions(mcp_server_uuid);

        CREATE INDEX IF NOT EXISTS tools_mcp_server_uuid_idx ON tools(mcp_server_uuid);
        CREATE UNIQUE INDEX IF NOT EXISTS tools_unique_tool_name_per_server_idx ON tools(mcp_server_uuid, name);

        CREATE INDEX IF NOT EXISTS namespaces_user_id_idx ON namespaces(user_id);
        CREATE UNIQUE INDEX IF NOT EXISTS namespaces_name_user_unique_idx ON namespaces(name, user_id);

        CREATE INDEX IF NOT EXISTS endpoints_namespace_uuid_idx ON endpoints(namespace_uuid);
        CREATE INDEX IF NOT EXISTS endpoints_user_id_idx ON endpoints(user_id);
        CREATE UNIQUE INDEX IF NOT EXISTS endpoints_name_unique ON endpoints(name);

        CREATE INDEX IF NOT EXISTS nsm_namespace_uuid_idx ON namespace_server_mappings(namespace_uuid);
        CREATE INDEX IF NOT EXISTS nsm_mcp_server_uuid_idx ON namespace_server_mappings(mcp_server_uuid);
        CREATE UNIQUE INDEX IF NOT EXISTS nsm_unique_idx ON namespace_server_mappings(namespace_uuid, mcp_server_uuid);

        CREATE INDEX IF NOT EXISTS ntm_namespace_uuid_idx ON namespace_tool_mappings(namespace_uuid);
        CREATE INDEX IF NOT EXISTS ntm_tool_uuid_idx ON namespace_tool_mappings(tool_uuid);
        CREATE UNIQUE INDEX IF NOT EXISTS ntm_unique_idx ON namespace_tool_mappings(namespace_uuid, tool_uuid);

        CREATE INDEX IF NOT EXISTS api_keys_user_id_idx ON api_keys(user_id);
        CREATE INDEX IF NOT EXISTS api_keys_key_idx ON api_keys(key);
        CREATE UNIQUE INDEX IF NOT EXISTS api_keys_name_per_user_idx ON api_keys(user_id, name);

        CREATE INDEX IF NOT EXISTS oac_client_id_idx ON oauth_authorization_codes(client_id);
        CREATE INDEX IF NOT EXISTS oac_user_id_idx ON oauth_authorization_codes(user_id);

        CREATE INDEX IF NOT EXISTS oat_client_id_idx ON oauth_access_tokens(client_id);
        CREATE INDEX IF NOT EXISTS oat_user_id_idx ON oauth_access_tokens(user_id);

        CREATE INDEX IF NOT EXISTS ds_mcp_server_uuid_idx ON docker_sessions(mcp_server_uuid);
        CREATE UNIQUE INDEX IF NOT EXISTS ds_unique_server_idx ON docker_sessions(mcp_server_uuid);

        CREATE UNIQUE INDEX IF NOT EXISTS policies_name_unique_idx ON policies(name);

        CREATE INDEX IF NOT EXISTS tcl_tool_name_idx ON tool_call_logs(tool_name);
        CREATE INDEX IF NOT EXISTS tcl_mcp_server_uuid_idx ON tool_call_logs(mcp_server_uuid);
        CREATE INDEX IF NOT EXISTS tcl_created_at_idx ON tool_call_logs(created_at);

        CREATE UNIQUE INDEX IF NOT EXISTS tool_sets_name_user_unique_idx ON tool_sets(name, user_id);

        CREATE INDEX IF NOT EXISTS tsi_tool_set_uuid_idx ON tool_set_items(tool_set_uuid);
        CREATE INDEX IF NOT EXISTS tsi_tool_uuid_idx ON tool_set_items(tool_uuid);
        CREATE UNIQUE INDEX IF NOT EXISTS tsi_unique_idx ON tool_set_items(tool_set_uuid, tool_uuid);

        CREATE UNIQUE INDEX IF NOT EXISTS saved_scripts_name_user_unique_idx ON saved_scripts(name, user_id);

        CREATE TABLE IF NOT EXISTS imported_sessions (
            uuid TEXT PRIMARY KEY,
            source_tool TEXT NOT NULL,
            source_path TEXT NOT NULL,
            source_size INTEGER,
            source_mtime INTEGER,
            external_session_id TEXT,
            title TEXT,
            session_format TEXT NOT NULL,
            transcript TEXT NOT NULL,
            excerpt TEXT,
            working_directory TEXT,
            transcript_hash TEXT NOT NULL UNIQUE,
            transcript_archive_path TEXT,
            transcript_metadata_archive_path TEXT,
            transcript_archive_format TEXT,
            transcript_stored_bytes INTEGER,
            normalized_session TEXT NOT NULL DEFAULT '{}',
            metadata TEXT NOT NULL DEFAULT '{}',
            discovered_at INTEGER NOT NULL,
            imported_at INTEGER NOT NULL,
            last_modified_at INTEGER,
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
        );

        CREATE TABLE IF NOT EXISTS imported_session_memories (
            uuid TEXT PRIMARY KEY,
            imported_session_uuid TEXT NOT NULL,
            memory_index INTEGER NOT NULL,
            kind TEXT NOT NULL DEFAULT 'memory',
            content TEXT NOT NULL,
            tags TEXT NOT NULL DEFAULT '[]',
            source TEXT NOT NULL DEFAULT 'heuristic',
            metadata TEXT NOT NULL DEFAULT '{}',
            created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            FOREIGN KEY (imported_session_uuid) REFERENCES imported_sessions(uuid) ON DELETE CASCADE,
            UNIQUE(imported_session_uuid, memory_index)
        );

        CREATE INDEX IF NOT EXISTS imported_sessions_transcript_hash_idx ON imported_sessions(transcript_hash);
        CREATE INDEX IF NOT EXISTS imported_sessions_source_tool_idx ON imported_sessions(source_tool);
        CREATE INDEX IF NOT EXISTS imported_sessions_source_path_idx ON imported_sessions(source_path);
        CREATE UNIQUE INDEX IF NOT EXISTS imported_sessions_transcript_hash_unique ON imported_sessions(transcript_hash);

        CREATE INDEX IF NOT EXISTS imported_session_memories_session_idx ON imported_session_memories(imported_session_uuid);
        CREATE INDEX IF NOT EXISTS imported_session_memories_kind_idx ON imported_session_memories(kind);
    `);

    // Dynamic Migrations for existing databases
    try {
        const tableInfo = database.pragma("table_info(mcp_servers)") as Array<{ name: string }>;
        const hasAlwaysOn = tableInfo.some((col) => col.name === "always_on");
        if (!hasAlwaysOn) {
            database.exec(`ALTER TABLE mcp_servers ADD COLUMN always_on INTEGER NOT NULL DEFAULT 0;`);
            console.info("[DB Migration] Added 'always_on' column to mcp_servers table.");
        }
    } catch (err) {
        console.warn("[DB Migration] Failed to alter mcp_servers table:", err);
    }

    try {
        const tableInfo = database.pragma("table_info(tools)") as Array<{ name: string }>;
        const hasAlwaysOn = tableInfo.some((col) => col.name === "always_on");
        if (!hasAlwaysOn) {
            database.exec(`ALTER TABLE tools ADD COLUMN always_on INTEGER NOT NULL DEFAULT 0;`);
            console.info("[DB Migration] Added 'always_on' column to tools table.");
        }
    } catch (err) {
        console.warn("[DB Migration] Failed to alter tools table:", err);
    }
}

// Default to SQLite local file
const dbPath = process.env.DATABASE_URL || "borg.db";

// Ensure we are using absolute path if it is a local file
const resolvedDbPath = dbPath.startsWith("file:")
    ? dbPath.slice(5)
    : path.resolve(process.cwd(), dbPath);

// Ensure the parent directory exists before opening the database
const dbDir = path.dirname(resolvedDbPath);
if (!fs.existsSync(dbDir)) {
    fs.mkdirSync(dbDir, { recursive: true });
}

const sqlite = new Database(resolvedDbPath);
initializeSchema(sqlite);
export const db = drizzle(sqlite, { schema });

// Export the raw sqlite instance for code that needs direct DB access
export const sqliteInstance: InstanceType<typeof Database> = sqlite;

// Export the schema for convenience
export { schema };
