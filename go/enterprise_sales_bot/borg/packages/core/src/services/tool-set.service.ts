import { eq, sql } from "drizzle-orm";
import { randomUUID } from "node:crypto";

import { db } from "../db/index.js";
import { toolSetItemsTable, toolSetsTable, toolsTable } from "../db/mcp-admin-schema.js";

export interface ToolSet {
    uuid: string;
    name: string;
    description: string | null;
    tools: string[];
}

export class ToolSetService {
    async listToolSets(userId?: string | null): Promise<ToolSet[]> {
        const conditions = [];
        if (userId) {
            conditions.push(eq(toolSetsTable.user_id, userId));
        } else {
            conditions.push(sql`${toolSetsTable.user_id} IS NULL`);
        }

        const sets = await db
            .select()
            .from(toolSetsTable)
            .where(sql.join(conditions, sql` AND `));

        // Fetch items for each set (N+1 query, but N is small)
        // Could use a join, but this is simpler for now
        // Join with toolsTable to get names
        const results: ToolSet[] = [];
        for (const set of sets) {
            const items = await db
                .select({ name: toolsTable.name })
                .from(toolSetItemsTable)
                .innerJoin(toolsTable, eq(toolSetItemsTable.tool_uuid, toolsTable.uuid))
                .where(eq(toolSetItemsTable.tool_set_uuid, set.uuid));

            results.push({
                uuid: set.uuid,
                name: set.name,
                description: set.description,
                tools: items.map(i => i.name),
            });
        }

        return results;
    }

    async getToolSet(name: string, userId?: string | null): Promise<ToolSet | undefined> {
        const conditions = [eq(toolSetsTable.name, name)];
        if (userId) {
            conditions.push(eq(toolSetsTable.user_id, userId));
        } else {
            conditions.push(sql`${toolSetsTable.user_id} IS NULL`);
        }

        const [set] = await db
            .select()
            .from(toolSetsTable)
            .where(sql.join(conditions, sql` AND `))
            .limit(1);

        if (!set) return undefined;

        const items = await db
            .select({ name: toolsTable.name })
            .from(toolSetItemsTable)
            .innerJoin(toolsTable, eq(toolSetItemsTable.tool_uuid, toolsTable.uuid))
            .where(eq(toolSetItemsTable.tool_set_uuid, set.uuid));

        return {
            uuid: set.uuid,
            name: set.name,
            description: set.description,
            tools: items.map(i => i.name),
        };
    }

    async createToolSet(
        name: string,
        tools: string[],
        description?: string,
        userId?: string | null
    ): Promise<ToolSet> {
        // Transactional creation
        return await db.transaction(async (tx) => {
            const [set] = await tx
                .insert(toolSetsTable)
                .values({
                    uuid: randomUUID(),
                    name,
                    description,
                    user_id: userId,
                })
                .onConflictDoUpdate({
                    target: [toolSetsTable.name, toolSetsTable.user_id],
                    set: { description: sql`excluded.description` } // Update desc if exists
                })
                .returning();

            // Clear existing items if updating
            await tx
                .delete(toolSetItemsTable)
                .where(eq(toolSetItemsTable.tool_set_uuid, set.uuid));

            // Insert new items
            // check if tools are UUIDs or Names? Assuming UUIDs for relation
            // If they are names, we should resolve them first, but that adds complexity.
            // For now, assuming the input `tools` array contains UUIDs as is standard for relational updates.
            // However, the interface `ToolSet` returns names. Ideally we should standardize on UUIDs for API.
            // For this fix, I will assume the caller provides UUIDs.

            if (tools.length > 0) {
                await tx.insert(toolSetItemsTable).values(
                    tools.map(toolUuid => ({
                        uuid: randomUUID(),
                        tool_set_uuid: set.uuid,
                        tool_uuid: toolUuid,
                    }))
                );
            }

            return {
                uuid: set.uuid,
                name: set.name,
                description: set.description,
                tools: tools,
            };
        });
    }
}

export const toolSetService = new ToolSetService();
