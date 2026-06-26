import { sql } from "drizzle-orm";
import {
    integer,
    sqliteTable,
    text,
    index,
} from "drizzle-orm/sqlite-core";

/**
 * Table: web_memories
 * Stores snapshots of web pages (via browser extension)
 */
export const webMemoriesTable = sqliteTable(
    "web_memories",
    {
        id: text("id").primaryKey(),
        url: text("url").notNull(),
        normalizedUrl: text("normalized_url").notNull(),
        title: text("title").notNull(),
        content: text("content").notNull(),
        selectedText: text("selected_text"),
        tags: text("tags", { mode: "json" })
            .$type<string[]>()
            .notNull()
            .default(sql`'[]'`),
        favicon: text("favicon"),
        source: text("source").notNull(),
        contentHash: text("content_hash").notNull(),
        savedAt: integer("saved_at", { mode: "timestamp" })
            .notNull()
            .default(sql`(strftime('%s', 'now'))`),
    },
    (table) => ({
        urlIdx: index("wm_url_idx").on(table.url),
        normalizedUrlIdx: index("wm_normalized_url_idx").on(table.normalizedUrl),
        savedAtIdx: index("wm_saved_at_idx").on(table.savedAt),
    })
);

/**
 * Table: browser_history
 * Stores visited URLs logged by the browser extension
 */
export const browserHistoryTable = sqliteTable(
    "browser_history",
    {
        id: text("id").primaryKey(),
        url: text("url").notNull(),
        title: text("title").notNull(),
        domain: text("domain").notNull(),
        visitedAt: integer("visited_at", { mode: "timestamp" }).notNull(),
        visitCount: integer("visit_count").notNull().default(1),
    },
    (table) => ({
        domainIdx: index("bh_domain_idx").on(table.domain),
        visitedAtIdx: index("bh_visited_at_idx").on(table.visitedAt),
    })
);

/**
 * Table: browser_console_logs
 * Stores intercepted console logs from the browser extension
 */
export const browserConsoleLogsTable = sqliteTable(
    "browser_console_logs",
    {
        id: text("id").primaryKey(),
        level: text("level").notNull(),
        message: text("message").notNull(),
        source: text("source").notNull(),
        url: text("url"),
        lineNumber: integer("line_number"),
        timestamp: integer("timestamp", { mode: "timestamp" }).notNull(),
    },
    (table) => ({
        levelIdx: index("bcl_level_idx").on(table.level),
        timestampIdx: index("bcl_timestamp_idx").on(table.timestamp),
    })
);
