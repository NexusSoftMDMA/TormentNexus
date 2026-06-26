import { sql } from "drizzle-orm";
import {
    integer,
    sqliteTable,
    text,
    index,
} from "drizzle-orm/sqlite-core";

/**
 * Table: tool_chains
 * Defines multi-step execution chains for grouping tool calls
 */
export const toolChainsTable = sqliteTable(
    "tool_chains",
    {
        id: text("id").primaryKey(),
        name: text("name").notNull(),
        description: text("description"),
        triggerPattern: text("trigger_pattern"),
        createdAt: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        nameIdx: index("tc_name_idx").on(table.name),
    })
);

/**
 * Table: tool_chain_steps
 * Ordered steps within a tool chain
 */
export const toolChainStepsTable = sqliteTable(
    "tool_chain_steps",
    {
        id: text("id").primaryKey(),
        chainId: text("chain_id")
            .notNull()
            .references(() => toolChainsTable.id, { onDelete: "cascade" }),
        stepOrder: integer("step_order").notNull(),
        toolName: text("tool_name").notNull(),
        argumentsTemplate: text("arguments_template", { mode: "json" })
            .$type<Record<string, unknown>>()
            .notNull()
            .default(sql`'{}'`),
        timeoutMs: integer("timeout_ms"),
        failurePolicy: text("failure_policy").notNull().default("abort"),
        retryCount: integer("retry_count").notNull().default(0),
    },
    (table) => ({
        chainIdx: index("tcs_chain_id_idx").on(table.chainId),
    })
);

/**
 * Table: tool_aliases
 * Custom aliases for existing tools (e.g. mapping "read_file" to "fs_read")
 */
export const toolAliasesTable = sqliteTable(
    "tool_aliases",
    {
        alias: text("alias").primaryKey(),
        targetTool: text("target_tool").notNull(),
        description: text("description"),
        defaultArguments: text("default_arguments", { mode: "json" })
            .$type<Record<string, unknown>>()
            .notNull()
            .default(sql`'{}'`),
        createdAt: integer("created_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        targetIdx: index("ta_target_tool_idx").on(table.targetTool),
    })
);
