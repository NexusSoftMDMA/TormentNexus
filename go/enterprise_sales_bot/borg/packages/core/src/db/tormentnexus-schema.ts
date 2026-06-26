/**
 * @file tormentnexus-schema.ts
 * @description SQLite schema definition for TormentNexus integration into TormentNexus.
 * @module packages/core/src/db/tormentnexus-schema
 *
 * WHAT:
 * This file defines the database schema for all TormentNexus functionality (Servers, Tools, Namespaces, etc.)
 * adapted for SQLite 3. It mirrors the structure of the original PostgreSQL schema (`tormentnexus-schema.pg.ts`)
 * but uses SQLite-compatible column types.
 *
 * WHY:
 * TormentNexus runs primarily as a local application where SQLite (`better-sqlite3`) is the standard embedded database.
 * The original TormentNexus uses PostgreSQL (`pg`). To support full feature parity locally without requiring a running
 * Postgres instance, we verify and adapt the schema to SQLite.
 *
 * HOW:
 * - `uuid` columns mapped to `text` (strings).
 * - `jsonb` columns mapped to `text` with `{ mode: 'json' }`.
 * - `timestamp` columns mapped to `integer` with `{ mode: 'timestamp' }` (Unix millis).
 * - `boolean` columns mapped to `integer` with `{ mode: 'boolean' }`.
 * - Enums mapped to simple text checks or handled in application logic (Zod).
 * - Foreign keys enforced via `references()`.
 */

import { sql } from "drizzle-orm";
import {
    integer,
    sqliteTable,
    text,
    index,
    unique,
    AnySQLiteColumn,
} from "drizzle-orm/sqlite-core";
import { type OAuthClientInformation, type OAuthTokens } from "@modelcontextprotocol/sdk/shared/auth.js";

// -- ENUMS (Defined as const arrays for Zod/App compatibility, stored as TEXT) --

export const McpServerTypeEnum = ["STDIO", "SSE", "STREAMABLE_HTTP"] as const;
export const McpServerStatusEnum = ["ACTIVE", "INACTIVE", "ERROR", "CONNECTING"] as const;
export const McpServerErrorStatusEnum = ["NONE", "CONNECTION_FAILED", "TIMEOUT", "INTERNAL_ERROR"] as const;
export const DockerSessionStatusEnum = ["PENDING", "RUNNING", "STOPPED", "ERROR", "NOT_FOUND"] as const;

// -- TABLES --

/**
 * Table: mcp_servers
 * Stores configuration for upstream MCP servers (the ones TormentNexus connects TO).
 */
export const mcpServersTable = sqliteTable(
    "mcp_servers",
    {
        uuid: text("uuid").primaryKey(), // Generated via crypto.randomUUID() in app logic
        name: text("name").notNull(),
        description: text("description"),
        type: text("type", { enum: McpServerTypeEnum }).notNull().default("STDIO"),
        command: text("command"),
        args: text("args", { mode: "json" }).$type<string[]>().notNull().default(sql`'[]'`), // SQLite doesn't support arrays, use JSON
        env: text("env", { mode: "json" })
            .$type<{ [key: string]: string }>()
            .notNull()
            .default(sql`'{}'`),
        url: text("url"),
        error_status: text("error_status", { enum: McpServerErrorStatusEnum })
            .notNull()
            .default("NONE"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        bearerToken: text("bearer_token"),
        headers: text("headers", { mode: "json" })
            .$type<{ [key: string]: string }>()
            .notNull()
            .default(sql`'{}'`),
        always_on: integer("always_on", { mode: "boolean" })
            .notNull()
            .default(false),
        user_id: text("user_id").notNull(), // Foreign key to usersTable (if auth exists) or 'system'
        source_published_server_uuid: text("source_published_server_uuid"), // FK to published_mcp_servers.uuid (nullable)
    },
    (table) => ({
        nameIdx: index("mcp_servers_name_idx").on(table.name),
        userIdIdx: index("mcp_servers_user_id_idx").on(table.user_id),
        nameUserUnique: unique("mcp_servers_name_user_unique").on(table.name, table.user_id),
    })
);

/**
 * Table: oauth_sessions
 * Manages OAuth 2.0 flows for servers that require it.
 */
export const oauthSessionsTable = sqliteTable(
    "oauth_sessions",
    {
        uuid: text("uuid").primaryKey(),
        mcp_server_uuid: text("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        client_information: text("client_information", { mode: "json" })
            .$type<OAuthClientInformation>()
            .notNull()
            .default(sql`'{}'`),
        tokens: text("tokens", { mode: "json" }).$type<OAuthTokens>(),
        code_verifier: text("code_verifier"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        serverIdx: index("oauth_sessions_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        uniquePerServer: unique("oauth_sessions_unique_per_server_idx").on(table.mcp_server_uuid),
    })
);

/**
 * Table: tools
 * Caches discovered tools from MCP servers to avoid constant re-discovery.
 */
export const toolsTable = sqliteTable(
    "tools",
    {
        uuid: text("uuid").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        // Schema is stored as JSON text
        toolSchema: text("tool_schema", { mode: "json" })
            .$type<{
                type: "object";
                properties?: Record<string, any>;
                required?: string[];
            }>()
            .notNull(),
        is_deferred: integer("is_deferred", { mode: "boolean" })
            .notNull()
            .default(false),
        always_on: integer("always_on", { mode: "boolean" })
            .notNull()
            .default(false),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        mcp_server_uuid: text("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
    },
    (table) => ({
        serverIdx: index("tools_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        uniqueNamePerServer: unique("tools_unique_tool_name_per_server_idx").on(
            table.mcp_server_uuid,
            table.name
        ),
    })
);

// -- Better-Auth Tables (Simplified for SQLite) --

export const usersTable = sqliteTable("users", {
    id: text("id").primaryKey(),
    name: text("name").notNull(),
    email: text("email").notNull().unique(),
    emailVerified: integer("email_verified", { mode: "boolean" }).notNull().default(false),
    image: text("image"),
    createdAt: integer("created_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updatedAt: integer("updated_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
});

export const sessionsTable = sqliteTable("sessions", {
    id: text("id").primaryKey(),
    expiresAt: integer("expires_at", { mode: "timestamp" }).notNull(),
    token: text("token").notNull().unique(),
    createdAt: integer("created_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updatedAt: integer("updated_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    ipAddress: text("ip_address"),
    userAgent: text("user_agent"),
    userId: text("user_id")
        .notNull()
        .references(() => usersTable.id, { onDelete: "cascade" }),
});

export const accountsTable = sqliteTable("accounts", {
    id: text("id").primaryKey(),
    accountId: text("account_id").notNull(),
    providerId: text("provider_id").notNull(),
    userId: text("user_id")
        .notNull()
        .references(() => usersTable.id, { onDelete: "cascade" }),
    accessToken: text("access_token"),
    refreshToken: text("refresh_token"),
    idToken: text("id_token"),
    accessTokenExpiresAt: integer("access_token_expires_at", { mode: "timestamp" }),
    refreshTokenExpiresAt: integer("refresh_token_expires_at", { mode: "timestamp" }),
    scope: text("scope"),
    password: text("password"),
    createdAt: integer("created_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updatedAt: integer("updated_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
});

export const verificationsTable = sqliteTable("verifications", {
    id: text("id").primaryKey(),
    identifier: text("identifier").notNull(),
    value: text("value").notNull(),
    expiresAt: integer("expires_at", { mode: "timestamp" }).notNull(),
    createdAt: integer("created_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updatedAt: integer("updated_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
});

// -- CORE TORMENTNEXUS TABLES --

/**
 * Table: namespaces
 * Logical groupings for tools (e.g. "coding", "search", "personal").
 */
export const namespacesTable = sqliteTable(
    "namespaces",
    {
        uuid: text("uuid").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => ({
        userIdIdx: index("namespaces_user_id_idx").on(table.user_id),
        nameUserUnique: unique("namespaces_name_user_unique_idx").on(table.name, table.user_id),
    })
);

/**
 * Table: endpoints
 * Defines routing endpoints (HTTP/SSE) that expose Namespaces to the outside world.
 */
export const endpointsTable = sqliteTable(
    "endpoints",
    {
        uuid: text("uuid").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        namespace_uuid: text("namespace_uuid")
            .notNull()
            .references(() => namespacesTable.uuid, { onDelete: "cascade" }),
        enable_api_key_auth: integer("enable_api_key_auth", { mode: "boolean" }).notNull().default(true),
        enable_oauth: integer("enable_oauth", { mode: "boolean" }).notNull().default(false),
        enable_max_rate: integer("enable_max_rate", { mode: "boolean" }).notNull().default(false),
        enable_client_max_rate: integer("enable_client_max_rate", { mode: "boolean" })
            .notNull()
            .default(false),
        max_rate: integer("max_rate"),
        max_rate_seconds: integer("max_rate_seconds"),
        client_max_rate: integer("client_max_rate"),
        client_max_rate_seconds: integer("client_max_rate_seconds"),
        client_max_rate_strategy: text("client_max_rate_strategy"),
        client_max_rate_strategy_key: text("client_max_rate_strategy_key"),
        use_query_param_auth: integer("use_query_param_auth", { mode: "boolean" })
            .notNull()
            .default(false),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => ({
        namespaceIdx: index("endpoints_namespace_uuid_idx").on(table.namespace_uuid),
        userIdIdx: index("endpoints_user_id_idx").on(table.user_id),
        nameUnique: unique("endpoints_name_unique").on(table.name),
    })
);

/**
 * Table: namespace_server_mappings
 * M2M Link: Namespaces <-> MCPServers.
 */
export const namespaceServerMappingsTable = sqliteTable(
    "namespace_server_mappings",
    {
        uuid: text("uuid").primaryKey(),
        namespace_uuid: text("namespace_uuid")
            .notNull()
            .references(() => namespacesTable.uuid, { onDelete: "cascade" }),
        mcp_server_uuid: text("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        status: text("status", { enum: McpServerStatusEnum }) // ACTIVE/INACTIVE
            .notNull()
            .default("ACTIVE"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        namespaceIdx: index("nsm_namespace_uuid_idx").on(table.namespace_uuid),
        serverIdx: index("nsm_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        uniqueMapping: unique("nsm_unique_idx").on(
            table.namespace_uuid,
            table.mcp_server_uuid
        ),
    })
);

/**
 * Table: namespace_tool_mappings
 * M2M Link: Namespaces <-> Tools (Overrides & Status).
 * Allows a namespace to "rename" a tool or disable it specifically for that namespace.
 */
export const namespaceToolMappingsTable = sqliteTable(
    "namespace_tool_mappings",
    {
        uuid: text("uuid").primaryKey(),
        namespace_uuid: text("namespace_uuid")
            .notNull()
            .references(() => namespacesTable.uuid, { onDelete: "cascade" }),
        tool_uuid: text("tool_uuid")
            .notNull()
            .references(() => toolsTable.uuid, { onDelete: "cascade" }),
        mcp_server_uuid: text("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        status: text("status", { enum: McpServerStatusEnum })
            .notNull()
            .default("ACTIVE"),
        override_name: text("override_name"),
        override_title: text("override_title"),
        override_description: text("override_description"),
        override_annotations: text("override_annotations", { mode: "json" })
            .$type<Record<string, unknown> | null>()
            .default(null),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        namespaceIdx: index("ntm_namespace_uuid_idx").on(table.namespace_uuid),
        toolIdx: index("ntm_tool_uuid_idx").on(table.tool_uuid),
        uniqueMapping: unique("ntm_unique_idx").on(
            table.namespace_uuid,
            table.tool_uuid
        ),
    })
);

/**
 * Table: api_keys
 * API Keys for accessing Endpoints.
 */
export const apiKeysTable = sqliteTable(
    "api_keys",
    {
        uuid: text("uuid").primaryKey(),
        name: text("name").notNull(),
        key: text("key").notNull().unique(),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        is_active: integer("is_active", { mode: "boolean" }).notNull().default(true),
    },
    (table) => ({
        userIdIdx: index("api_keys_user_id_idx").on(table.user_id),
        keyIdx: index("api_keys_key_idx").on(table.key),
        namePerUserUnique: unique("api_keys_name_per_user_idx").on(table.user_id, table.name),
    })
);

/**
 * Table: config
 * Global application configuration settings.
 */
export const configTable = sqliteTable("config", {
    id: text("id").primaryKey(),
    value: text("value").notNull(),
    description: text("description"),
    created_at: integer("created_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updated_at: integer("updated_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
});

/**
 * Table: workspace_secrets
 * Environment variables and sensitive OAuth tokens/keys.
 */
export const workspaceSecretsTable = sqliteTable("workspace_secrets", {
    key: text("key").primaryKey(),
    value: text("value").notNull(),
    created_at: integer("created_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updated_at: integer("updated_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
});

export const importedSessionsTable = sqliteTable(
    "imported_sessions",
    {
        uuid: text("uuid").primaryKey(),
        source_tool: text("source_tool").notNull(),
        source_path: text("source_path").notNull(),
        source_size: integer("source_size"),
        source_mtime: integer("source_mtime"),
        external_session_id: text("external_session_id"),
        title: text("title"),
        session_format: text("session_format").notNull().default("generic"),
        transcript: text("transcript").notNull(),
        excerpt: text("excerpt"),
        working_directory: text("working_directory"),
        transcript_hash: text("transcript_hash").notNull(),
        transcript_archive_path: text("transcript_archive_path"),
        transcript_metadata_archive_path: text("transcript_metadata_archive_path"),
        transcript_archive_format: text("transcript_archive_format"),
        transcript_stored_bytes: integer("transcript_stored_bytes"),
        normalized_session: text("normalized_session", { mode: "json" })
            .$type<Record<string, unknown>>()
            .notNull(),
        metadata: text("metadata", { mode: "json" })
            .$type<Record<string, unknown>>()
            .notNull()
            .default(sql`'{}'`),
        discovered_at: integer("discovered_at", { mode: "timestamp" }).notNull(),
        imported_at: integer("imported_at", { mode: "timestamp" }).notNull(),
        last_modified_at: integer("last_modified_at", { mode: "timestamp" }),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        transcriptHashIdx: index("imported_sessions_transcript_hash_idx").on(table.transcript_hash),
        sourceToolIdx: index("imported_sessions_source_tool_idx").on(table.source_tool),
        sourcePathIdx: index("imported_sessions_source_path_idx").on(table.source_path),
        uniqueTranscript: unique("imported_sessions_transcript_hash_unique").on(table.transcript_hash),
    })
);

export const importedSessionMemoriesTable = sqliteTable(
    "imported_session_memories",
    {
        uuid: text("uuid").primaryKey(),
        imported_session_uuid: text("imported_session_uuid")
            .notNull()
            .references(() => importedSessionsTable.uuid, { onDelete: "cascade" }),
        memory_index: integer("memory_index").notNull(),
        kind: text("kind").notNull().default("memory"),
        content: text("content").notNull(),
        tags: text("tags", { mode: "json" })
            .$type<string[]>()
            .notNull()
            .default(sql`'[]'`),
        source: text("source").notNull().default("heuristic"),
        metadata: text("metadata", { mode: "json" })
            .$type<Record<string, unknown>>()
            .notNull()
            .default(sql`'{}'`),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        sessionIdx: index("imported_session_memories_session_idx").on(table.imported_session_uuid),
        kindIdx: index("imported_session_memories_kind_idx").on(table.kind),
        uniqueMemoryPerSession: unique("imported_session_memories_session_index_unique").on(
            table.imported_session_uuid,
            table.memory_index
        ),
    })
);

// -- OAUTH PROVIDER TABLES --

export const oauthClientsTable = sqliteTable("oauth_clients", {
    client_id: text("client_id").primaryKey(),
    client_secret: text("client_secret"),
    client_name: text("client_name").notNull(),
    redirect_uris: text("redirect_uris", { mode: "json" })
        .$type<string[]>()
        .notNull()
        .default(sql`'[]'`),
    grant_types: text("grant_types", { mode: "json" })
        .$type<string[]>()
        .notNull()
        .default(sql`'["authorization_code","refresh_token"]'`),
    response_types: text("response_types", { mode: "json" })
        .$type<string[]>()
        .notNull()
        .default(sql`'["code"]'`),
    token_endpoint_auth_method: text("token_endpoint_auth_method")
        .notNull()
        .default("none"),
    scope: text("scope").default("admin"),
    client_uri: text("client_uri"),
    logo_uri: text("logo_uri"),
    contacts: text("contacts", { mode: "json" }).$type<string[]>(),
    tos_uri: text("tos_uri"),
    policy_uri: text("policy_uri"),
    software_id: text("software_id"),
    software_version: text("software_version"),
    created_at: integer("created_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
    updated_at: integer("updated_at", { mode: "timestamp" })
        .notNull()
        .default(sql`(strftime('%s', 'now'))`),
});

export const oauthAuthorizationCodesTable = sqliteTable(
    "oauth_authorization_codes",
    {
        code: text("code").primaryKey(),
        client_id: text("client_id")
            .notNull()
            .references(() => oauthClientsTable.client_id, { onDelete: "cascade" }),
        redirect_uri: text("redirect_uri").notNull(),
        scope: text("scope").notNull().default("admin"),
        user_id: text("user_id")
            .notNull()
            .references(() => usersTable.id, { onDelete: "cascade" }),
        code_challenge: text("code_challenge"),
        code_challenge_method: text("code_challenge_method"),
        expires_at: integer("expires_at", { mode: "timestamp" }).notNull(),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        ctxClientId: index("oac_client_id_idx").on(table.client_id),
        ctxUserId: index("oac_user_id_idx").on(table.user_id),
    })
);

export const oauthAccessTokensTable = sqliteTable(
    "oauth_access_tokens",
    {
        access_token: text("access_token").primaryKey(),
        client_id: text("client_id")
            .notNull()
            .references(() => oauthClientsTable.client_id, { onDelete: "cascade" }),
        user_id: text("user_id")
            .notNull()
            .references(() => usersTable.id, { onDelete: "cascade" }),
        scope: text("scope").notNull().default("admin"),
        expires_at: integer("expires_at", { mode: "timestamp" }).notNull(),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        clientIdx: index("oat_client_id_idx").on(table.client_id),
        userIdx: index("oat_user_id_idx").on(table.user_id),
    })
);

/**
 * Table: docker_sessions
 * Manages Docker-in-Docker containers for sandboxed MCP servers.
 */
export const dockerSessionsTable = sqliteTable(
    "docker_sessions",
    {
        uuid: text("uuid").primaryKey(),
        mcp_server_uuid: text("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        container_id: text("container_id").notNull(),
        container_name: text("container_name"),
        url: text("url"),
        status: text("status", { enum: DockerSessionStatusEnum }).notNull().default("PENDING"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        started_at: integer("started_at", { mode: "timestamp" }),
        stopped_at: integer("stopped_at", { mode: "timestamp" }),
        error_message: text("error_message"),
        retry_count: integer("retry_count").notNull().default(0),
        last_retry_at: integer("last_retry_at", { mode: "timestamp" }),
        max_retries: integer("max_retries").notNull().default(3),
    },
    (table) => ({
        serverIdx: index("ds_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        uniqueServer: unique("ds_unique_server_idx").on(table.mcp_server_uuid),
    })
);

/**
 * Table: policies
 * Access control policies.
 */
export const policiesTable = sqliteTable(
    "policies",
    {
        uuid: text("uuid").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        rules: text("rules", { mode: "json" })
            .$type<Record<string, unknown>>()
            .notNull()
            .default(sql`'{}'`),
        createdAt: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updatedAt: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        uniqueName: unique("policies_name_unique_idx").on(table.name),
    })
);

/**
 * Table: tool_call_logs
 * Observability logs for tool execution.
 */
export const toolCallLogsTable = sqliteTable(
    "tool_call_logs",
    {
        uuid: text("uuid").primaryKey(),
        tool_name: text("tool_name").notNull(),
        mcp_server_uuid: text("mcp_server_uuid")
            .references(() => mcpServersTable.uuid, { onDelete: "set null" }),
        namespace_uuid: text("namespace_uuid")
            .references(() => namespacesTable.uuid, { onDelete: "set null" }),
        endpoint_uuid: text("endpoint_uuid")
            .references(() => endpointsTable.uuid, { onDelete: "set null" }),
        args: text("args", { mode: "json" }).$type<Record<string, unknown>>(),
        result: text("result", { mode: "json" }).$type<Record<string, unknown>>(),
        error: text("error"),
        duration_ms: integer("duration_ms"),
        session_id: text("session_id"),
        parent_call_uuid: text("parent_call_uuid"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        toolNameIdx: index("tcl_tool_name_idx").on(table.tool_name),
        serverIdx: index("tcl_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        createdAtIdx: index("tcl_created_at_idx").on(table.created_at),
    })
);

/**
 * Table: tool_sets
 * Reusable groupings of tools.
 */
export const toolSetsTable = sqliteTable(
    "tool_sets",
    {
        uuid: text("uuid").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => ({
        nameUserUnique: unique("tool_sets_name_user_unique_idx").on(table.name, table.user_id),
    })
);

/**
 * Table: tool_set_items
 * Joint table for ToolSets <-> Tools.
 */
export const toolSetItemsTable = sqliteTable(
    "tool_set_items",
    {
        uuid: text("uuid").primaryKey(),
        tool_set_uuid: text("tool_set_uuid")
            .notNull()
            .references(() => toolSetsTable.uuid, { onDelete: "cascade" }),
        tool_uuid: text("tool_uuid")
            .notNull()
            .references(() => toolsTable.uuid, { onDelete: "cascade" }),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        setIdx: index("tsi_tool_set_uuid_idx").on(table.tool_set_uuid),
        toolIdx: index("tsi_tool_uuid_idx").on(table.tool_uuid),
        uniqueItem: unique("tsi_unique_idx").on(table.tool_set_uuid, table.tool_uuid),
    })
);

/**
 * Table: saved_scripts
 * User-defined scripts (JavaScript/Python) for tasks.
 */
export const savedScriptsTable = sqliteTable(
    "saved_scripts",
    {
        uuid: text("uuid").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        code: text("code").notNull(),
        language: text("language").notNull().default("javascript"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => ({
        nameUserUnique: unique("saved_scripts_name_user_unique_idx").on(table.name, table.user_id),
    })
);

/**
 * Table: workflows
 * User-defined interactive canvas pipelines in the dashboard.
 */
export const workflowsTable = sqliteTable(
    "workflows",
    {
        id: text("id").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        nodes_json: text("nodes_json", { mode: "json" })
            .$type<any[]>()
            .notNull()
            .default(sql`'[]'`),
        edges_json: text("edges_json", { mode: "json" })
            .$type<any[]>()
            .notNull()
            .default(sql`'[]'`),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    }
);

// ============================================================
// MCP REGISTRY INTELLIGENCE — Published Catalog Tables
// ============================================================
// These tables implement the P0 "MCP Registry Intelligence" feature:
// ingesting publicly known MCP servers from external registries
// (Glama.ai, Smithery.ai, curated GitHub lists, etc.), normalizing
// them into a canonical catalog, and tracking validation results.
//
// Key design principles (from AGENTS.md):
//   - Archivist: catalog integrity — canonical_id, provenance, dedup
//   - Configurator: high-confidence recipes or human review
//   - Verifier: only certify what we have safely observed
//   - Guardian: isolation boundaries for validation runs

/** Enum: status lifecycle for a published server entry */
export const PublishedServerStatusEnum = [
    "discovered",   // Seen in at least one external registry
    "normalized",   // canonical_id assigned, fields cleaned
    "probeable",    // Recipe exists and has enough info to attempt connection
    "validated",    // At least one validation run passed (tools/list succeeded)
    "certified",    // Manually reviewed by operator or passed full smoke test
    "broken",       // Validation consistently fails
    "archived",     // Removed from tracked sources / deprecated
] as const;

/** Enum: transport type of the published server */
export const PublishedServerTransportEnum = ["stdio", "sse", "streamable_http", "unknown"] as const;

/** Enum: primary install mechanism */
export const PublishedServerInstallMethodEnum = ["npx", "uvx", "docker", "binary", "url", "unknown"] as const;

/** Enum: authentication model */
export const PublishedServerAuthModelEnum = ["none", "api_key", "oauth", "bearer", "unknown"] as const;

/** Enum: validation run execution mode */
export const ValidationRunModeEnum = ["transport_probe", "tools_list", "smoke_test", "full_validation"] as const;

/** Enum: validation run outcome */
export const ValidationRunOutcomeEnum = ["pending", "passed", "failed", "error", "timeout", "skipped"] as const;
export const LinkBacklogResearchStatusEnum = ["pending", "running", "done", "failed", "skipped"] as const;

/**
 * Table: published_mcp_servers
 * The canonical catalog of publicly known MCP servers.
 * Each row represents a de-duplicated, normalized server entry aggregated
 * from one or more external registries.
 *
 * canonical_id format: "<source-org>/<repo-slug>" or "<registry>/<id>"
 * confidence: 0–100, reflects richness + validation evidence
 * status: follows the PublishedServerStatusEnum lifecycle
 */
export const publishedMcpServersTable = sqliteTable(
    "published_mcp_servers",
    {
        uuid: text("uuid").primaryKey(),
        // Deduplicated stable identifier — e.g. "modelcontextprotocol/servers/filesystem" or "glama/abc123"
        canonical_id: text("canonical_id").notNull().unique(),
        display_name: text("display_name").notNull(),
        description: text("description"),
        author: text("author"),
        repository_url: text("repository_url"),
        homepage_url: text("homepage_url"),
        icon_url: text("icon_url"),
        transport: text("transport", { enum: PublishedServerTransportEnum })
            .notNull()
            .default("unknown"),
        install_method: text("install_method", { enum: PublishedServerInstallMethodEnum })
            .notNull()
            .default("unknown"),
        auth_model: text("auth_model", { enum: PublishedServerAuthModelEnum })
            .notNull()
            .default("unknown"),
        // Lifecycle status — controlled by Archivist
        status: text("status", { enum: PublishedServerStatusEnum })
            .notNull()
            .default("discovered"),
        // Confidence score 0–100: increases with each validation pass, decreases on failures
        confidence: integer("confidence").notNull().default(0),
        tags: text("tags", { mode: "json" })
            .$type<string[]>()
            .notNull()
            .default(sql`'[]'`),
        categories: text("categories", { mode: "json" })
            .$type<string[]>()
            .notNull()
            .default(sql`'[]'`),
        stars: integer("stars"), // GitHub stars if available
        last_seen_at: integer("last_seen_at", { mode: "timestamp" }),
        last_verified_at: integer("last_verified_at", { mode: "timestamp" }),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        canonicalIdx: index("pms_canonical_id_idx").on(table.canonical_id),
        statusIdx: index("pms_status_idx").on(table.status),
        transportIdx: index("pms_transport_idx").on(table.transport),
        installIdx: index("pms_install_method_idx").on(table.install_method),
        updatedIdx: index("pms_updated_at_idx").on(table.updated_at),
    })
);

/**
 * Table: published_mcp_server_sources
 * Provenance tracking — one row per (server × registry) pair.
 * Records where each server was found, when, and the raw payload.
 * Enables accurate deduplication and registry change detection.
 */
export const publishedMcpServerSourcesTable = sqliteTable(
    "published_mcp_server_sources",
    {
        uuid: text("uuid").primaryKey(),
        server_uuid: text("server_uuid")
            .notNull()
            .references(() => publishedMcpServersTable.uuid, { onDelete: "cascade" }),
        // Friendly name of the source registry, e.g. "glama.ai", "smithery.ai", "awesome-mcp-servers"
        source_name: text("source_name").notNull(),
        // URL of the specific page/API endpoint where this entry was found
        source_url: text("source_url"),
        // The raw JSON payload returned by the source (stored for audit/re-processing)
        raw_payload: text("raw_payload", { mode: "json" })
            .$type<Record<string, unknown>>(),
        first_seen_at: integer("first_seen_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        last_seen_at: integer("last_seen_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        serverIdx: index("pmss_server_uuid_idx").on(table.server_uuid),
        sourceNameIdx: index("pmss_source_name_idx").on(table.source_name),
        // One row per server × source_name pair
        uniqueServerSource: unique("pmss_unique_server_source_idx").on(table.server_uuid, table.source_name),
    })
);

/**
 * Table: published_mcp_config_recipes
 * Configuration templates generated by the Configurator agent.
 * A recipe is an mcp.json-compatible config snippet with placeholders
 * for required secrets/env vars. Confidence 0–100 indicates how reliable
 * the recipe is (100 = validated by a successful validation run).
 */
export const publishedMcpConfigRecipesTable = sqliteTable(
    "published_mcp_config_recipes",
    {
        uuid: text("uuid").primaryKey(),
        server_uuid: text("server_uuid")
            .notNull()
            .references(() => publishedMcpServersTable.uuid, { onDelete: "cascade" }),
        // Monotonically increasing version per server (recipe can be improved over time)
        recipe_version: integer("recipe_version").notNull().default(1),
        // The mcp.json-compatible config object (command/args/env/url/type etc.)
        template: text("template", { mode: "json" })
            .$type<Record<string, unknown>>()
            .notNull(),
        // List of env var names that must be provided by the operator, e.g. ["GITHUB_TOKEN"]
        required_secrets: text("required_secrets", { mode: "json" })
            .$type<string[]>()
            .notNull()
            .default(sql`'[]'`),
        // Any additional env vars with suggested default values
        required_env: text("required_env", { mode: "json" })
            .$type<Record<string, string>>()
            .notNull()
            .default(sql`'{}'`),
        // Confidence 0–100: how reliable this recipe is
        confidence: integer("confidence").notNull().default(0),
        // Human-readable explanation of the recipe and its confidence
        explanation: text("explanation"),
        // Whether this is the currently active recipe for this server
        is_active: integer("is_active", { mode: "boolean" }).notNull().default(true),
        // Which agent generated this recipe (Configurator by default)
        generated_by: text("generated_by").notNull().default("Configurator"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        serverIdx: index("pmcr_server_uuid_idx").on(table.server_uuid),
        activeIdx: index("pmcr_is_active_idx").on(table.server_uuid, table.is_active),
    })
);

/**
 * Table: published_mcp_validation_runs
 * Recorded state of each automated validation attempt against a published server.
 * These are performed by the Verifier agent in a sandboxed environment.
 *
 * run_mode hierarchy (cost/risk ascending):
 *   transport_probe → tools_list → smoke_test → full_validation
 */
export const publishedMcpValidationRunsTable = sqliteTable(
    "published_mcp_validation_runs",
    {
        uuid: text("uuid").primaryKey(),
        server_uuid: text("server_uuid")
            .notNull()
            .references(() => publishedMcpServersTable.uuid, { onDelete: "cascade" }),
        run_mode: text("run_mode", { enum: ValidationRunModeEnum }).notNull(),
        started_at: integer("started_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        finished_at: integer("finished_at", { mode: "timestamp" }),
        outcome: text("outcome", { enum: ValidationRunOutcomeEnum })
            .notNull()
            .default("pending"),
        // Categorical failure reason for grouping/alerting
        failure_class: text("failure_class"),
        // Number of tools discovered (populated when tools_list succeeds)
        tool_count: integer("tool_count"),
        // Structured summary of findings (tools list excerpt, capability flags, errors)
        findings_summary: text("findings_summary", { mode: "json" })
            .$type<Record<string, unknown>>(),
        // Agent role responsible for this run
        performed_by: text("performed_by").notNull().default("Verifier"),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        serverIdx: index("pmvr_server_uuid_idx").on(table.server_uuid),
        outcomeIdx: index("pmvr_outcome_idx").on(table.outcome),
        createdIdx: index("pmvr_created_at_idx").on(table.created_at),
    })
);

/**
 * Table: links_backlog
 * Canonical TormentNexus link backlog, initially fed by BobbyBookmarks and designed to
 * become one of the data sources for the universal integrated MCP directory.
 */
export const linksBacklogTable = sqliteTable(
    "links_backlog",
    {
        uuid: text("uuid").primaryKey(),
        url: text("url").notNull(),
        normalized_url: text("normalized_url").notNull().unique(),
        title: text("title"),
        description: text("description"),
        tags: text("tags", { mode: "json" })
            .$type<string[]>()
            .notNull()
            .default(sql`'[]'`),
        source: text("source").notNull().default("manual"),
        is_duplicate: integer("is_duplicate", { mode: "boolean" }).notNull().default(false),
        duplicate_of: text("duplicate_of"),
        research_status: text("research_status", { enum: LinkBacklogResearchStatusEnum })
            .notNull()
            .default("pending"),
        http_status: integer("http_status"),
        page_title: text("page_title"),
        page_description: text("page_description"),
        favicon_url: text("favicon_url"),
        researched_at: integer("researched_at", { mode: "timestamp" }),
        cluster_id: text("cluster_id"),
        bobbybookmarks_bookmark_id: integer("bobbybookmarks_bookmark_id"),
        import_session_id: integer("import_session_id"),
        raw_payload: text("raw_payload", { mode: "json" }).$type<Record<string, unknown> | null>(),
        synced_at: integer("synced_at", { mode: "timestamp" }),
        created_at: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updated_at: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        normalizedIdx: index("links_backlog_normalized_url_idx").on(table.normalized_url),
        sourceIdx: index("links_backlog_source_idx").on(table.source),
        researchStatusIdx: index("links_backlog_research_status_idx").on(table.research_status),
        clusterIdx: index("links_backlog_cluster_id_idx").on(table.cluster_id),
        syncedAtIdx: index("links_backlog_synced_at_idx").on(table.synced_at),
    })
);
