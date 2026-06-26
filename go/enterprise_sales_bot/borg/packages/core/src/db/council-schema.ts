/**
 * @file council-schema.ts
 * @description SQLite schema definition for TormentNexus Council (Auto-Orchestrator) integration.
 * @module packages/core/src/db/council-schema
 */

import { sql } from "drizzle-orm";
import {
    integer,
    sqliteTable,
    text,
    index,
    real,
} from "drizzle-orm/sqlite-core";

/**
 * Table: council_debates
 * Stores the history of council debates and decisions.
 */
export const councilDebatesTable = sqliteTable(
    "council_debates",
    {
        id: text("id").primaryKey(),
        title: text("title").notNull(),
        sessionId: text("session_id"),
        workspaceId: text("workspace_id"),
        taskType: text("task_type").notNull().default("general"),
        status: text("status").notNull().default("completed"),
        consensus: real("consensus").notNull().default(0),
        weightedConsensus: real("weighted_consensus").notNull().default(0),
        outcome: text("outcome").notNull(), // 'approved', 'rejected', 'deadlock'
        rounds: integer("rounds").notNull().default(1),
        timestamp: integer("timestamp", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        data: text("data", { mode: "json" })
            .$type<any>()
            .notNull(),
    },
    (table) => ({
        sessionIdIdx: index("idx_council_debates_session").on(table.sessionId),
        timestampIdx: index("idx_council_debates_timestamp").on(table.timestamp),
    })
);

/**
 * Table: council_workspaces
 * Stores configuration for supervised workspaces.
 */
export const councilWorkspacesTable = sqliteTable(
    "council_workspaces",
    {
        id: text("id").primaryKey(),
        name: text("name").notNull(),
        path: text("path").notNull(),
        status: text("status").notNull().default("active"),
        config: text("config", { mode: "json" })
            .$type<any>()
            .notNull()
            .default(sql`'{}'`),
        description: text("description"),
        createdAt: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
        updatedAt: integer("updated_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        statusIdx: index("idx_council_workspaces_status").on(table.status),
    })
);
