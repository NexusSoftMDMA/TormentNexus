import { db } from "../index.js";
import { eq, desc } from "drizzle-orm";
import {
    webMemoriesTable,
    browserHistoryTable,
    browserConsoleLogsTable,
} from "../mcp-browser-schema.js";

// === Web Memories ===

export type WebMemory = typeof webMemoriesTable.$inferSelect;
export type NewWebMemory = typeof webMemoriesTable.$inferInsert;

// === Browser History ===

export type BrowserHistoryEntry = typeof browserHistoryTable.$inferSelect;
export type NewBrowserHistoryEntry = typeof browserHistoryTable.$inferInsert;

// === Console Logs ===

export type ConsoleLogEntry = typeof browserConsoleLogsTable.$inferSelect;
export type NewConsoleLogEntry = typeof browserConsoleLogsTable.$inferInsert;

export const browserDataRepository = {
    // --- Memories ---
    async saveMemory(memory: NewWebMemory): Promise<WebMemory> {
        const [saved] = await db.insert(webMemoriesTable).values(memory).returning();
        return saved;
    },

    async getAllMemories(): Promise<WebMemory[]> {
        return db.select().from(webMemoriesTable).orderBy(desc(webMemoriesTable.savedAt));
    },

    async getMemoryById(id: string): Promise<WebMemory | null> {
        const [mem] = await db.select().from(webMemoriesTable).where(eq(webMemoriesTable.id, id));
        return mem || null;
    },

    async deleteMemory(id: string): Promise<boolean> {
        const res = await db.delete(webMemoriesTable).where(eq(webMemoriesTable.id, id)).returning();
        return res.length > 0;
    },

    // --- History ---
    async addHistoryEntry(entry: NewBrowserHistoryEntry): Promise<BrowserHistoryEntry> {
        // Technically should upsert if `url` exists on same `domain` today, but simple insert is fine for now
        const [saved] = await db.insert(browserHistoryTable).values(entry).returning();
        return saved;
    },

    async getHistory(options?: { limit?: number; offset?: number }): Promise<BrowserHistoryEntry[]> {
        const query = db.select().from(browserHistoryTable).orderBy(desc(browserHistoryTable.visitedAt));
        if (options?.limit) query.limit(options.limit);
        if (options?.offset) query.offset(options.offset);
        return query;
    },

    async clearHistory(): Promise<void> {
        await db.delete(browserHistoryTable);
    },

    // --- Console Logs ---
    async addConsoleLog(log: NewConsoleLogEntry): Promise<ConsoleLogEntry> {
        const [saved] = await db.insert(browserConsoleLogsTable).values(log).returning();
        return saved;
    },

    async getConsoleLogs(options?: { limit?: number; offset?: number; level?: string }): Promise<ConsoleLogEntry[]> {
        let query = db.select().from(browserConsoleLogsTable).$dynamic();
        
        if (options?.level) {
            query = query.where(eq(browserConsoleLogsTable.level, options.level));
        }
        
        query = query.orderBy(desc(browserConsoleLogsTable.timestamp));
        
        if (options?.limit) query.limit(options.limit);
        if (options?.offset) query.offset(options.offset);
        
        return query;
    },

    async clearConsoleLogs(): Promise<void> {
        await db.delete(browserConsoleLogsTable);
    }
};
