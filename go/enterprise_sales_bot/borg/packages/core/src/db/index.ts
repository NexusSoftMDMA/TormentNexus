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

    try {
        const tableInfo = database.pragma("table_info(imported_sessions)") as Array<{ name: string }>;
        if (tableInfo.length > 0) {
            const hasSourceSize = tableInfo.some((col) => col.name === "source_size");
            if (!hasSourceSize) {
                database.exec(`ALTER TABLE imported_sessions ADD COLUMN source_size INTEGER;`);
                console.info("[DB Migration] Added 'source_size' column to imported_sessions table.");
            }
            const hasSourceMtime = tableInfo.some((col) => col.name === "source_mtime");
            if (!hasSourceMtime) {
                database.exec(`ALTER TABLE imported_sessions ADD COLUMN source_mtime INTEGER;`);
                console.info("[DB Migration] Added 'source_mtime' column to imported_sessions table.");
            }
        }
    } catch (err) {
        console.warn("[DB Migration] Failed to alter imported_sessions table:", err);
    }
    // Add missing tables for tormentnexus-schema
    try {
        database.exec(`
            CREATE TABLE IF NOT EXISTS workspace_secrets (
                key TEXT PRIMARY KEY,
                value TEXT NOT NULL,
                created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
                updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
            );
            CREATE TABLE IF NOT EXISTS published_mcp_servers (
                uuid TEXT PRIMARY KEY,
                canonical_id TEXT NOT NULL UNIQUE,
                display_name TEXT NOT NULL,
                description TEXT,
                author TEXT,
                repository_url TEXT,
                homepage_url TEXT,
                icon_url TEXT,
                transport TEXT NOT NULL DEFAULT 'unknown',
                install_method TEXT NOT NULL DEFAULT 'unknown',
                auth_model TEXT NOT NULL DEFAULT 'unknown',
                status TEXT NOT NULL DEFAULT 'discovered',
                confidence INTEGER NOT NULL DEFAULT 0,
                tags TEXT NOT NULL DEFAULT '[]',
                categories TEXT NOT NULL DEFAULT '[]',
                stars INTEGER,
                last_seen_at INTEGER,
                last_verified_at INTEGER,
                created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
                updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
            );
            CREATE TABLE IF NOT EXISTS council_debates (
                uuid TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                topic TEXT NOT NULL,
                participants TEXT NOT NULL DEFAULT '[]',
                rounds TEXT NOT NULL DEFAULT '[]',
                consensus TEXT,
                status TEXT NOT NULL DEFAULT 'pending',
                created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
                updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
            );
            CREATE TABLE IF NOT EXISTS links_backlog (
                uuid TEXT PRIMARY KEY,
                url TEXT NOT NULL,
                normalized_url TEXT NOT NULL UNIQUE,
                title TEXT,
                description TEXT,
                tags TEXT NOT NULL DEFAULT '[]',
                source TEXT NOT NULL DEFAULT 'manual',
                is_duplicate INTEGER NOT NULL DEFAULT 0,
                duplicate_of TEXT,
                research_status TEXT NOT NULL DEFAULT 'pending',
                http_status INTEGER,
                page_title TEXT,
                page_description TEXT,
                discovered_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
                last_attempted_at INTEGER,
                completed_at INTEGER,
                error_message TEXT,
                metadata TEXT NOT NULL DEFAULT '{}'
            );
            CREATE INDEX IF NOT EXISTS published_mcp_servers_canonical_id_idx ON published_mcp_servers(canonical_id);
            CREATE INDEX IF NOT EXISTS council_debates_session_id_idx ON council_debates(session_id);
            CREATE INDEX IF NOT EXISTS links_backlog_normalized_url_idx ON links_backlog(normalized_url);
            CREATE INDEX IF NOT EXISTS links_backlog_source_idx ON links_backlog(source);
            CREATE INDEX IF NOT EXISTS links_backlog_research_status_idx ON links_backlog(research_status);
        `);
        console.info("[DB Migration] Added missing tables: workspace_secrets, published_mcp_servers, council_debates, links_backlog");
    } catch (err) {
        console.warn("[DB Migration] Failed to create missing tables:", err);
    }
    // Add missing columns to links_backlog
    try {
        const lbCols = database.pragma("table_info(links_backlog)") as Array<{ name: string }>;
        const lbExisting = new Set(lbCols.map((col) => col.name));
        const lbMigrations: Array<[string, string]> = [
            ["tags", "ALTER TABLE links_backlog ADD COLUMN tags TEXT NOT NULL DEFAULT '[]';"],
            ["is_duplicate", "ALTER TABLE links_backlog ADD COLUMN is_duplicate INTEGER NOT NULL DEFAULT 0;"],
            ["duplicate_of", "ALTER TABLE links_backlog ADD COLUMN duplicate_of TEXT;"],
            ["http_status", "ALTER TABLE links_backlog ADD COLUMN http_status INTEGER;"],
            ["page_title", "ALTER TABLE links_backlog ADD COLUMN page_title TEXT;"],
            ["page_description", "ALTER TABLE links_backlog ADD COLUMN page_description TEXT;"],
        ["favicon_url", "ALTER TABLE links_backlog ADD COLUMN favicon_url TEXT;"],
        ["researched_at", "ALTER TABLE links_backlog ADD COLUMN researched_at INTEGER;"],
        ["cluster_id", "ALTER TABLE links_backlog ADD COLUMN cluster_id TEXT;"],
        ["bobbybookmarks_bookmark_id", "ALTER TABLE links_backlog ADD COLUMN bobbybookmarks_bookmark_id INTEGER;"],
        ["import_session_id", "ALTER TABLE links_backlog ADD COLUMN import_session_id INTEGER;"],
        ["raw_payload", "ALTER TABLE links_backlog ADD COLUMN raw_payload TEXT;"],
        ["synced_at", "ALTER TABLE links_backlog ADD COLUMN synced_at INTEGER;"],
        ["created_at", "ALTER TABLE links_backlog ADD COLUMN created_at INTEGER;"],
        ["updated_at", "ALTER TABLE links_backlog ADD COLUMN updated_at INTEGER;"],
        ];
        for (const [colName, sql] of lbMigrations) {
            if (!lbExisting.has(colName)) {
                database.exec(sql);
            }
        }
        if (lbMigrations.some(([colName]) => !lbExisting.has(colName))) {
            console.info("[DB Migration] Added missing columns to links_backlog table.");
        }
    } catch (err) {
        console.warn("[DB Migration] Failed to alter links_backlog table:", err);
    }

    // Add favicon_url column to published_mcp_servers if missing
    try {
        const pmsCols = database.pragma("table_info(published_mcp_servers)") as Array<{ name: string }>;
        const hasFaviconUrl = pmsCols.some((col) => col.name === "favicon_url");
        if (!hasFaviconUrl) {
            database.exec(`ALTER TABLE published_mcp_servers ADD COLUMN favicon_url TEXT;`);
            console.info("[DB Migration] Added 'favicon_url' column to published_mcp_servers table.");
        }
    } catch (err) {
        console.warn("[DB Migration] Failed to alter published_mcp_servers table:", err);
    }

    // Add source_published_server_uuid column to mcp_servers if missing
    try {
        const tableInfo = database.pragma("table_info(mcp_servers)") as Array<{ name: string }>;
        const hasSourcePublishedServerUuid = tableInfo.some((col) => col.name === "source_published_server_uuid");
        if (!hasSourcePublishedServerUuid) {
            database.exec(`ALTER TABLE mcp_servers ADD COLUMN source_published_server_uuid TEXT;`);
            console.info("[DB Migration] Added 'source_published_server_uuid' column to mcp_servers table.");
        }
    } catch (err) {
        console.warn("[DB Migration] Failed to alter mcp_servers table:", err);
    }

    // Add remaining schema tables (browser, tool-chains, workflows, council, published catalog)
    try {
        const allTables = new Set(
            (database.pragma("table_list") as Array<{ name: string }>).map((t) => t.name)
        );
        const missingTableMigrations: Array<string> = [];
        if (!allTables.has("browser_history")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS browser_history (
                id TEXT PRIMARY KEY, url TEXT NOT NULL, title TEXT NOT NULL,
                domain TEXT NOT NULL, visited_at INTEGER NOT NULL, visit_count INTEGER NOT NULL DEFAULT 1
            ); CREATE INDEX IF NOT EXISTS bh_domain_idx ON browser_history(domain);
            CREATE INDEX IF NOT EXISTS bh_visited_at_idx ON browser_history(visited_at);`
        );
        if (!allTables.has("browser_console_logs")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS browser_console_logs (
                id TEXT PRIMARY KEY, level TEXT NOT NULL, message TEXT NOT NULL,
                source TEXT NOT NULL, url TEXT, line_number INTEGER, timestamp INTEGER NOT NULL
            ); CREATE INDEX IF NOT EXISTS bcl_level_idx ON browser_console_logs(level);
            CREATE INDEX IF NOT EXISTS bcl_timestamp_idx ON browser_console_logs(timestamp);`
        );
        if (!allTables.has("web_memories")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS web_memories (
                id TEXT PRIMARY KEY, url TEXT NOT NULL, normalized_url TEXT NOT NULL,
                title TEXT NOT NULL, content TEXT NOT NULL, selected_text TEXT,
                tags TEXT NOT NULL DEFAULT '[]', favicon TEXT, source TEXT NOT NULL,
                content_hash TEXT NOT NULL, saved_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
            ); CREATE INDEX IF NOT EXISTS wm_url_idx ON web_memories(url);
            CREATE INDEX IF NOT EXISTS wm_normalized_url_idx ON web_memories(normalized_url);
            CREATE INDEX IF NOT EXISTS wm_saved_at_idx ON web_memories(saved_at);`
        );
        if (!allTables.has("tool_chains")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS tool_chains (
                id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT,
                trigger_pattern TEXT, created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
            ); CREATE INDEX IF NOT EXISTS tc_name_idx ON tool_chains(name);`
        );
        if (!allTables.has("tool_chain_steps")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS tool_chain_steps (
                id TEXT PRIMARY KEY, chain_id TEXT NOT NULL, step_order INTEGER NOT NULL,
                tool_name TEXT NOT NULL, arguments_template TEXT NOT NULL DEFAULT '{}',
                timeout_ms INTEGER, failure_policy TEXT NOT NULL DEFAULT 'abort',
                retry_count INTEGER NOT NULL DEFAULT 0
            ); CREATE INDEX IF NOT EXISTS tcs_chain_id_idx ON tool_chain_steps(chain_id);`
        );
        if (!allTables.has("tool_aliases")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS tool_aliases (
                alias TEXT PRIMARY KEY, target_tool TEXT NOT NULL, description TEXT,
                default_arguments TEXT NOT NULL DEFAULT '{}',
                created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
            ); CREATE INDEX IF NOT EXISTS ta_target_tool_idx ON tool_aliases(target_tool);`
        );
        if (!allTables.has("workflows")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS workflows (
                id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT,
                nodes_json TEXT NOT NULL DEFAULT '[]', edges_json TEXT NOT NULL DEFAULT '[]',
                created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
                updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now')), user_id TEXT
            );`
        );
        if (!allTables.has("council_workspaces")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS council_workspaces (
                id TEXT PRIMARY KEY, name TEXT NOT NULL, path TEXT NOT NULL,
                status TEXT NOT NULL DEFAULT 'active', config TEXT NOT NULL DEFAULT '{}',
                description TEXT, created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
                updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
            ); CREATE INDEX IF NOT EXISTS idx_council_workspaces_status ON council_workspaces(status);`
        );
        if (!allTables.has("published_mcp_server_sources")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS published_mcp_server_sources (
                uuid TEXT PRIMARY KEY, server_uuid TEXT NOT NULL, source_name TEXT NOT NULL,
                source_url TEXT, raw_payload TEXT,
                first_seen_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
                last_seen_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
            ); CREATE INDEX IF NOT EXISTS pmss_server_uuid_idx ON published_mcp_server_sources(server_uuid);
            CREATE INDEX IF NOT EXISTS pmss_source_name_idx ON published_mcp_server_sources(source_name);
            CREATE UNIQUE INDEX IF NOT EXISTS pmss_unique_server_source_idx ON published_mcp_server_sources(server_uuid, source_name);`
        );
        if (!allTables.has("published_mcp_config_recipes")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS published_mcp_config_recipes (
                uuid TEXT PRIMARY KEY, server_uuid TEXT NOT NULL, recipe_version INTEGER NOT NULL DEFAULT 1,
                template TEXT NOT NULL, required_secrets TEXT NOT NULL DEFAULT '[]',
                required_env TEXT NOT NULL DEFAULT '{}', confidence INTEGER NOT NULL DEFAULT 0,
                explanation TEXT, is_active INTEGER NOT NULL DEFAULT 1,
                generated_by TEXT NOT NULL DEFAULT 'Configurator',
                created_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
                updated_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
            );`
        );
        if (!allTables.has("published_mcp_validation_runs")) missingTableMigrations.push(
            `CREATE TABLE IF NOT EXISTS published_mcp_validation_runs (
                uuid TEXT PRIMARY KEY, server_uuid TEXT NOT NULL, run_mode TEXT NOT NULL,
                started_at INTEGER NOT NULL DEFAULT (strftime('%s','now')),
                finished_at INTEGER, outcome TEXT NOT NULL DEFAULT 'pending',
                failure_class TEXT, tool_count INTEGER, findings_summary TEXT,
                performed_by TEXT NOT NULL DEFAULT 'Verifier',
                created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
            ); CREATE INDEX IF NOT EXISTS pmvr_server_uuid_idx ON published_mcp_validation_runs(server_uuid);
            CREATE INDEX IF NOT EXISTS pmvr_outcome_idx ON published_mcp_validation_runs(outcome);`
        );
        if (missingTableMigrations.length > 0) {
            for (const sql of missingTableMigrations) {
                database.exec(sql);
            }
            console.info(`[DB Migration] Created ${missingTableMigrations.length} additional missing tables.`);
        }
    } catch (err) {
        console.warn("[DB Migration] Failed to create additional tables:", err);
    }
}

// Default to SQLite local file
const dbPath = process.env.DATABASE_URL || "tormentnexus.db";

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
