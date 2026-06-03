/**
 * @file tormentnexus-schema.pg.ts
 * @description PostgreSQL schema definition for TormentNexus integration into TormentNexus.
 * @module packages/core/src/db/tormentnexus-schema.pg
 *
 * WHAT:
 * This file defines the EXACT PostgreSQL schema used by the original TormentNexus project.
 * It is the "source of truth" for the data model when running TormentNexus with a Postgres backend
 * (e.g. in production or docker-compose environments).
 *
 * WHY:
 * The user requested support for BOTH SQLite (for local development) and PostgreSQL.
 * This file allows the application to be configured to use Postgres by swapping the schema import
 * in the database connection setup. It maintains 1:1 compatibility with TormentNexus's data structures.
 *
 * HOW:
 * - Uses `drizzle-orm/pg-core`.
 * - Definitions are identical to `tormentnexus/apps/backend/src/db/schema.ts`.
 * - Relies on `packages/core/src/types/tormentnexus-types.ts` for Enums (once ported).
 */

import { sql } from "drizzle-orm";
import {
    boolean,
    index,
    integer,
    jsonb,
    pgEnum,
    pgTable,
    text,
    timestamp,
    unique,
    uuid,
} from "drizzle-orm/pg-core";
import { type OAuthClientInformation, type OAuthTokens } from "@modelcontextprotocol/sdk/shared/auth.js";

// -- ENUMS (Standardized for Postgres) --
// These match the Zod Enums defined in `src/types/tormentnexus-types.ts`

export const mcpServerTypeEnum = pgEnum("mcp_server_type", ["STDIO", "SSE", "STREAMABLE_HTTP"]);
export const mcpServerStatusEnum = pgEnum("mcp_server_status", ["ACTIVE", "INACTIVE", "ERROR", "CONNECTING"]);
export const mcpServerErrorStatusEnum = pgEnum("mcp_server_error_status", ["NONE", "CONNECTION_FAILED", "TIMEOUT", "INTERNAL_ERROR"]);
export const dockerSessionStatusEnum = pgEnum("docker_session_status", ["PENDING", "RUNNING", "STOPPED", "ERROR", "NOT_FOUND"]);

// -- TABLES --

export const mcpServersTable = pgTable(
    "mcp_servers",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        description: text("description"),
        type: mcpServerTypeEnum("type").notNull().default("STDIO"),
        command: text("command"),
        args: text("args")
            .array()
            .notNull()
            .default(sql`'{}'::text[]`),
        env: jsonb("env")
            .$type<{ [key: string]: string }>()
            .notNull()
            .default(sql`'{}'::jsonb`),
        url: text("url"),
        error_status: mcpServerErrorStatusEnum("error_status")
            .notNull()
            .default("NONE"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        bearerToken: text("bearer_token"),
        headers: jsonb("headers")
            .$type<{ [key: string]: string }>()
            .notNull()
            .default(sql`'{}'::jsonb`),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => [
        index("mcp_servers_type_idx").on(table.type),
        index("mcp_servers_user_id_idx").on(table.user_id),
        index("mcp_servers_error_status_idx").on(table.error_status),
        unique("mcp_servers_name_user_unique_idx").on(table.name, table.user_id),
        sql`CONSTRAINT mcp_servers_name_regex_check CHECK (name ~ '^[a-zA-Z0-9_-]+$')`,
        sql`CONSTRAINT mcp_servers_url_check CHECK (
        (type = 'SSE' AND url IS NOT NULL AND command IS NULL AND url ~ '^https?://[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*(:[0-9]+)?(/[a-zA-Z0-9-._~:/?#\[\]@!$&''()*+,;=]*)?$') OR
        (type = 'STDIO' AND url IS NULL AND command IS NOT NULL) OR
        (type = 'STREAMABLE_HTTP' AND url IS NOT NULL AND command IS NULL AND url ~ '^https?://[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*(:[0-9]+)?(/[a-zA-Z0-9-._~:/?#\[\]@!$&''()*+,;=]*)?$')
      )`,
    ]
);

export const oauthSessionsTable = pgTable(
    "oauth_sessions",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        mcp_server_uuid: uuid("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        client_information: jsonb("client_information")
            .$type<OAuthClientInformation>()
            .notNull()
            .default(sql`'{}'::jsonb`),
        tokens: jsonb("tokens").$type<OAuthTokens>(),
        code_verifier: text("code_verifier"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updated_at: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        index("oauth_sessions_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        unique("oauth_sessions_unique_per_server_idx").on(table.mcp_server_uuid),
    ]
);

export const toolsTable = pgTable(
    "tools",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        description: text("description"),
        toolSchema: jsonb("tool_schema")
            .$type<{
                type: "object";
                properties?: Record<string, any>;
                required?: string[];
            }>()
            .notNull(),
        is_deferred: boolean("is_deferred").notNull().default(false),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updated_at: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        mcp_server_uuid: uuid("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
    },
    (table) => [
        index("tools_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        unique("tools_unique_tool_name_per_server_idx").on(
            table.mcp_server_uuid,
            table.name
        ),
    ]
);

// Better-auth tables (Standard Postgres)
export const usersTable = pgTable("users", {
    id: text("id").primaryKey(),
    name: text("name").notNull(),
    email: text("email").notNull().unique(),
    emailVerified: boolean("email_verified").notNull().default(false),
    image: text("image"),
    createdAt: timestamp("created_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
    updatedAt: timestamp("updated_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
});

export const sessionsTable = pgTable("sessions", {
    id: text("id").primaryKey(),
    expiresAt: timestamp("expires_at", { withTimezone: true }).notNull(),
    token: text("token").notNull().unique(),
    createdAt: timestamp("created_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
    updatedAt: timestamp("updated_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
    ipAddress: text("ip_address"),
    userAgent: text("user_agent"),
    userId: text("user_id")
        .notNull()
        .references(() => usersTable.id, { onDelete: "cascade" }),
});

export const accountsTable = pgTable("accounts", {
    id: text("id").primaryKey(),
    accountId: text("account_id").notNull(),
    providerId: text("provider_id").notNull(),
    userId: text("user_id")
        .notNull()
        .references(() => usersTable.id, { onDelete: "cascade" }),
    accessToken: text("access_token"),
    refreshToken: text("refresh_token"),
    idToken: text("id_token"),
    accessTokenExpiresAt: timestamp("access_token_expires_at", {
        withTimezone: true,
    }),
    refreshTokenExpiresAt: timestamp("refresh_token_expires_at", {
        withTimezone: true,
    }),
    scope: text("scope"),
    password: text("password"),
    createdAt: timestamp("created_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
    updatedAt: timestamp("updated_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
});

export const verificationsTable = pgTable("verifications", {
    id: text("id").primaryKey(),
    identifier: text("identifier").notNull(),
    value: text("value").notNull(),
    expiresAt: timestamp("expires_at", { withTimezone: true }).notNull(),
    createdAt: timestamp("created_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
    updatedAt: timestamp("updated_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
});

// Namespaces table
export const namespacesTable = pgTable(
    "namespaces",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        description: text("description"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updated_at: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => [
        index("namespaces_user_id_idx").on(table.user_id),
        unique("namespaces_name_user_unique_idx").on(table.name, table.user_id),
        sql`CONSTRAINT namespaces_name_regex_check CHECK (name ~ '^[a-zA-Z0-9_-]+$')`,
    ]
);

// Endpoints table
export const endpointsTable = pgTable(
    "endpoints",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        description: text("description"),
        namespace_uuid: uuid("namespace_uuid")
            .notNull()
            .references(() => namespacesTable.uuid, { onDelete: "cascade" }),
        enable_api_key_auth: boolean("enable_api_key_auth").notNull().default(true),
        enable_oauth: boolean("enable_oauth").notNull().default(false),
        enable_max_rate: boolean("enable_max_rate").notNull().default(false),
        enable_client_max_rate: boolean("enable_client_max_rate")
            .notNull()
            .default(false),
        max_rate: integer("max_rate"),
        max_rate_seconds: integer("max_rate_seconds"),
        client_max_rate: integer("client_max_rate"),
        client_max_rate_seconds: integer("client_max_rate_seconds"),
        client_max_rate_strategy: text("client_max_rate_strategy"),
        client_max_rate_strategy_key: text("client_max_rate_strategy_key"),
        use_query_param_auth: boolean("use_query_param_auth")
            .notNull()
            .default(false),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updated_at: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => [
        index("endpoints_namespace_uuid_idx").on(table.namespace_uuid),
        index("endpoints_user_id_idx").on(table.user_id),
        unique("endpoints_name_unique").on(table.name),
        sql`CONSTRAINT endpoints_name_url_compatible_check CHECK (name ~ '^[a-zA-Z0-9_-]+$')`,
    ]
);

// Namespace Server Mappings
export const namespaceServerMappingsTable = pgTable(
    "namespace_server_mappings",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        namespace_uuid: uuid("namespace_uuid")
            .notNull()
            .references(() => namespacesTable.uuid, { onDelete: "cascade" }),
        mcp_server_uuid: uuid("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        status: mcpServerStatusEnum("status")
            .notNull()
            .default("ACTIVE"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        index("namespace_server_mappings_namespace_uuid_idx").on(table.namespace_uuid),
        index("namespace_server_mappings_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        index("namespace_server_mappings_status_idx").on(table.status),
        unique("namespace_server_mappings_unique_idx").on(
            table.namespace_uuid,
            table.mcp_server_uuid
        ),
    ]
);

// Namespace Tool Mappings
export const namespaceToolMappingsTable = pgTable(
    "namespace_tool_mappings",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        namespace_uuid: uuid("namespace_uuid")
            .notNull()
            .references(() => namespacesTable.uuid, { onDelete: "cascade" }),
        tool_uuid: uuid("tool_uuid")
            .notNull()
            .references(() => toolsTable.uuid, { onDelete: "cascade" }),
        mcp_server_uuid: uuid("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        status: mcpServerStatusEnum("status")
            .notNull()
            .default("ACTIVE"),
        override_name: text("override_name"),
        override_title: text("override_title"),
        override_description: text("override_description"),
        override_annotations: jsonb("override_annotations")
            .$type<Record<string, unknown> | null>()
            .default(sql`NULL`),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        index("namespace_tool_mappings_namespace_uuid_idx").on(table.namespace_uuid),
        index("namespace_tool_mappings_tool_uuid_idx").on(table.tool_uuid),
        index("namespace_tool_mappings_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        index("namespace_tool_mappings_status_idx").on(table.status),
        unique("namespace_tool_mappings_unique_idx").on(
            table.namespace_uuid,
            table.tool_uuid
        ),
    ]
);

// API Keys
export const apiKeysTable = pgTable(
    "api_keys",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        key: text("key").notNull().unique(),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        is_active: boolean("is_active").notNull().default(true),
    },
    (table) => [
        index("api_keys_user_id_idx").on(table.user_id),
        index("api_keys_key_idx").on(table.key),
        index("api_keys_is_active_idx").on(table.is_active),
        unique("api_keys_name_per_user_idx").on(table.user_id, table.name),
    ]
);

// Config
export const configTable = pgTable("config", {
    id: text("id").primaryKey(),
    value: text("value").notNull(),
    description: text("description"),
    created_at: timestamp("created_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
    updated_at: timestamp("updated_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
});

// OAuth Providers
export const oauthClientsTable = pgTable("oauth_clients", {
    client_id: text("client_id").primaryKey(),
    client_secret: text("client_secret"),
    client_name: text("client_name").notNull(),
    redirect_uris: text("redirect_uris")
        .array()
        .notNull()
        .default(sql`'{}'::text[]`),
    grant_types: text("grant_types")
        .array()
        .notNull()
        .default(sql`'{"authorization_code","refresh_token"}'::text[]`),
    response_types: text("response_types")
        .array()
        .notNull()
        .default(sql`'{"code"}'::text[]`),
    token_endpoint_auth_method: text("token_endpoint_auth_method")
        .notNull()
        .default("none"),
    scope: text("scope").default("admin"),
    client_uri: text("client_uri"),
    logo_uri: text("logo_uri"),
    contacts: text("contacts").array(),
    tos_uri: text("tos_uri"),
    policy_uri: text("policy_uri"),
    software_id: text("software_id"),
    software_version: text("software_version"),
    created_at: timestamp("created_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
    updated_at: timestamp("updated_at", { withTimezone: true })
        .notNull()
        .defaultNow(),
});

export const oauthAuthorizationCodesTable = pgTable(
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
        expires_at: timestamp("expires_at", { withTimezone: true }).notNull(),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        index("oauth_authorization_codes_client_id_idx").on(table.client_id),
        index("oauth_authorization_codes_user_id_idx").on(table.user_id),
        index("oauth_authorization_codes_expires_at_idx").on(table.expires_at),
    ]
);

export const oauthAccessTokensTable = pgTable(
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
        expires_at: timestamp("expires_at", { withTimezone: true }).notNull(),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        index("oauth_access_tokens_client_id_idx").on(table.client_id),
        index("oauth_access_tokens_user_id_idx").on(table.user_id),
        index("oauth_access_tokens_expires_at_idx").on(table.expires_at),
    ]
);

export const dockerSessionsTable = pgTable(
    "docker_sessions",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        mcp_server_uuid: uuid("mcp_server_uuid")
            .notNull()
            .references(() => mcpServersTable.uuid, { onDelete: "cascade" }),
        container_id: text("container_id").notNull(),
        container_name: text("container_name"),
        url: text("url"),
        status: dockerSessionStatusEnum("status").notNull().default("PENDING"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updated_at: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        started_at: timestamp("started_at", { withTimezone: true }),
        stopped_at: timestamp("stopped_at", { withTimezone: true }),
        error_message: text("error_message"),
        retry_count: integer("retry_count").notNull().default(0),
        last_retry_at: timestamp("last_retry_at", { withTimezone: true }),
        max_retries: integer("max_retries").notNull().default(3),
    },
    (table) => [
        index("docker_sessions_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        index("docker_sessions_status_idx").on(table.status),
        unique("docker_sessions_unique_server_idx").on(table.mcp_server_uuid),
    ]
);

export const policiesTable = pgTable(
    "policies",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        description: text("description"),
        rules: jsonb("rules").$type<Record<string, unknown>>().notNull().default(sql`'{}'::jsonb`),
        createdAt: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updatedAt: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        unique("policies_name_unique_idx").on(table.name),
    ]
);

export const toolCallLogsTable = pgTable(
    "tool_call_logs",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        tool_name: text("tool_name").notNull(),
        mcp_server_uuid: uuid("mcp_server_uuid")
            .references(() => mcpServersTable.uuid, { onDelete: "set null" }),
        namespace_uuid: uuid("namespace_uuid")
            .references(() => namespacesTable.uuid, { onDelete: "set null" }),
        endpoint_uuid: uuid("endpoint_uuid")
            .references(() => endpointsTable.uuid, { onDelete: "set null" }),
        args: jsonb("args").$type<Record<string, unknown>>(),
        result: jsonb("result").$type<Record<string, unknown>>(),
        error: text("error"),
        duration_ms: integer("duration_ms"),
        session_id: text("session_id"),
        parent_call_uuid: text("parent_call_uuid"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        index("tool_call_logs_tool_name_idx").on(table.tool_name),
        index("tool_call_logs_mcp_server_uuid_idx").on(table.mcp_server_uuid),
        index("tool_call_logs_created_at_idx").on(table.created_at),
    ]
);

export const toolSetsTable = pgTable(
    "tool_sets",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        description: text("description"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updated_at: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => [
        unique("tool_sets_name_user_unique_idx").on(table.name, table.user_id),
    ]
);

export const toolSetItemsTable = pgTable(
    "tool_set_items",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        tool_set_uuid: uuid("tool_set_uuid")
            .notNull()
            .references(() => toolSetsTable.uuid, { onDelete: "cascade" }),
        tool_uuid: uuid("tool_uuid")
            .notNull()
            .references(() => toolsTable.uuid, { onDelete: "cascade" }),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
    },
    (table) => [
        index("tool_set_items_tool_set_uuid_idx").on(table.tool_set_uuid),
        index("tool_set_items_tool_uuid_idx").on(table.tool_uuid),
        unique("tool_set_items_unique_idx").on(table.tool_set_uuid, table.tool_uuid),
    ]
);

export const savedScriptsTable = pgTable(
    "saved_scripts",
    {
        uuid: uuid("uuid").primaryKey().defaultRandom(),
        name: text("name").notNull(),
        description: text("description"),
        code: text("code").notNull(),
        language: text("language").notNull().default("javascript"),
        created_at: timestamp("created_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        updated_at: timestamp("updated_at", { withTimezone: true })
            .notNull()
            .defaultNow(),
        user_id: text("user_id").references(() => usersTable.id, {
            onDelete: "cascade",
        }),
    },
    (table) => [
        unique("saved_scripts_name_user_unique_idx").on(table.name, table.user_id),
    ]
);
