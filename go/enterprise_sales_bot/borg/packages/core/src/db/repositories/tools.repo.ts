/**
 * @file tools.repo.ts
 * @module packages/core/src/db/repositories/tools.repo
 *
 * WHAT:
 * Repository for managing MCP Tools.
 *
 * WHY:
 * Handles CRUD operations for Tools discovered from servers.
 * Supports bulk upserting tools during server synchronization.
 *
 * HOW:
 * - Upserts tools using `onConflictDoUpdate`.
 * - Handles bulk deletions of obsolete tools.
 * - Exposes a feature-flagged post-upsert hook for optional AI enhancement pipelines.
 */

import {
    DatabaseTool,
    ToolCreateInput,
    ToolUpsertInput,
} from "../../types/mcp-admin/index.js";
import { and, eq, notInArray, sql } from "drizzle-orm";

// import { descriptionEnhancerService } from "../../lib/ai/description-enhancer.service";
// import { toolSearchService } from "../../lib/ai/tool-search.service";

import { db } from "../index.js";
import { toolsTable } from "../mcp-admin-schema.js";
import { randomUUID } from "node:crypto";

type ToolRow = typeof toolsTable.$inferSelect;
type ToolInsert = typeof toolsTable.$inferInsert;

function runPostUpsertHooks(tools: DatabaseTool[]): void {
    // Optional future integration point for AI description enhancement / embeddings.
    // Disabled by default to preserve current behavior and avoid introducing hidden async side effects.
    if (process.env.ENABLE_TOOL_AI_POST_PROCESSING !== "true") {
        return;
    }

    if (tools.length === 0) {
        return;
    }

    console.info(
        `[ToolsRepository] ENABLE_TOOL_AI_POST_PROCESSING=true; ${tools.length} tools queued for post-processing hook.`,
    );
}

export class ToolsRepository {
    async findByMcpServerUuid(mcpServerUuid: string): Promise<DatabaseTool[]> {
        return await db
            .select()
            .from(toolsTable)
            .where(eq(toolsTable.mcp_server_uuid, mcpServerUuid))
            .orderBy(toolsTable.name);
    }

    async findAll(): Promise<DatabaseTool[]> {
        return await db.select().from(toolsTable).orderBy(toolsTable.name);
    }

    async create(input: ToolCreateInput): Promise<DatabaseTool> {
        const payload: ToolInsert = {
            uuid: randomUUID(),
            name: input.name,
            description: input.description ?? null,
            // @ts-expect-error -- type mismatch from inferred schema types
            toolSchema: input.toolSchema,
            mcp_server_uuid: input.mcp_server_uuid,
        };

        const [createdTool] = await db.insert(toolsTable).values(payload).returning();

        return createdTool;
    }

    async bulkUpsert(input: ToolUpsertInput): Promise<DatabaseTool[]> {
        if (!input.tools || input.tools.length === 0) {
            return [];
        }

        // Format tools for database insertion
        const toolsToInsert: ToolInsert[] = input.tools.map((tool) => ({
            uuid: randomUUID(),
            name: tool.name,
            description: tool.description || "",
            // Cast to any to satisfy type checker if needed, but schema matches
            toolSchema: {
                type: "object" as const,
                ...tool.inputSchema,
            },
            mcp_server_uuid: input.mcpServerUuid,
        }));

        // Batch insert all tools with upsert
        // Note: Drizzle's `returning` behavior depends on the driver. better-sqlite3 supports it.
        const results = await db
            .insert(toolsTable)
            .values(toolsToInsert)
            .onConflictDoUpdate({
                target: [toolsTable.mcp_server_uuid, toolsTable.name],
                set: {
                    description: sql`excluded.description`,
                    toolSchema: sql`excluded.tool_schema`,
                    updated_at: new Date(),
                },
            })
            .returning();

        runPostUpsertHooks(results);

        return results;
    }

    async findByUuid(uuid: string): Promise<DatabaseTool | undefined> {
        const [tool] = await db
            .select()
            .from(toolsTable)
            .where(eq(toolsTable.uuid, uuid))
            .limit(1);

        return tool;
    }

    async deleteByUuid(uuid: string): Promise<DatabaseTool | undefined> {
        const [deletedTool] = await db
            .delete(toolsTable)
            .where(eq(toolsTable.uuid, uuid))
            .returning();

        return deletedTool;
    }

    /**
     * Delete tools that are no longer present in the current tool list
     * @param mcpServerUuid - UUID of the MCP server
     * @param currentToolNames - Array of tool names that currently exist in the MCP server
     * @returns Array of deleted tools
     */
    async deleteObsoleteTools(
        mcpServerUuid: string,
        currentToolNames: string[],
    ): Promise<DatabaseTool[]> {
        if (currentToolNames.length === 0) {
            // If no tools are provided, delete all tools for this server
            return await db
                .delete(toolsTable)
                .where(eq(toolsTable.mcp_server_uuid, mcpServerUuid))
                .returning();
        }

        // Delete tools that are in DB but not in current tool list
        return await db
            .delete(toolsTable)
            .where(
                and(
                    eq(toolsTable.mcp_server_uuid, mcpServerUuid),
                    notInArray(toolsTable.name, currentToolNames),
                ),
            )
            .returning();
    }

    /**
     * Sync tools for a server: upsert current tools and delete obsolete ones
     * @param input - Tool upsert input containing tools and server UUID
     * @returns Object with upserted and deleted tools
     */
    async syncTools(input: ToolUpsertInput): Promise<{
        upserted: DatabaseTool[];
        deleted: DatabaseTool[];
    }> {
        const currentToolNames = input.tools.map((tool) => tool.name);

        // First, delete obsolete tools
        const deleted = await this.deleteObsoleteTools(
            input.mcpServerUuid,
            currentToolNames,
        );

        // Then, upsert current tools
        let upserted: DatabaseTool[] = [];
        if (input.tools.length > 0) {
            upserted = await this.bulkUpsert(input);
        }

        return { upserted, deleted };
    }
    async setAlwaysOn(uuid: string, alwaysOn: boolean): Promise<DatabaseTool | undefined> {
        const [updatedTool] = await db
            .update(toolsTable)
            .set({ always_on: alwaysOn, updated_at: new Date() })
            .where(eq(toolsTable.uuid, uuid))
            .returning();
        return updatedTool;
    }
}

export const toolsRepository = new ToolsRepository();
