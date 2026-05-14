import { eq, and, inArray } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { db } from "../index.js";
import { toolSetsTable, toolSetItemsTable } from "../mcp-admin-schema.js";
import { ToolSet } from "../../types/metamcp/tool-sets.zod.js";

type ToolSetRow = typeof toolSetsTable.$inferSelect;
type ToolSetInsert = typeof toolSetsTable.$inferInsert;
type ToolSetItemRow = typeof toolSetItemsTable.$inferSelect;
type ToolSetItemInsert = typeof toolSetItemsTable.$inferInsert;

export class ToolSetsRepository {
    async findAll(): Promise<ToolSet[]> {
        const sets = await db.select().from(toolSetsTable);
        if (sets.length === 0) return [];

        return this.hydrateBatch(sets);
    }

    async findByUuid(uuid: string): Promise<ToolSet | undefined> {
        const [set] = await db.select().from(toolSetsTable).where(eq(toolSetsTable.uuid, uuid));
        if (!set) return undefined;
        const [hydrated] = await this.hydrateBatch([set]);
        return hydrated;
    }

    async create(input: { name: string; description?: string | null; tools: string[]; user_id?: string | null }): Promise<ToolSet> {
        const uuid = randomUUID();
        const payload: ToolSetInsert = {
            uuid,
            name: input.name,
            description: input.description ?? null,
            user_id: input.user_id ?? null,
        };

        const [set] = await db.insert(toolSetsTable).values(payload).returning();

        if (input.tools && input.tools.length > 0) {
            await this.addTools(uuid, input.tools);
        }

        const [hydrated] = await this.hydrateBatch([set]);
        return hydrated;
    }

    async update(input: { uuid: string; name?: string; description?: string | null; tools?: string[]; user_id?: string | null }): Promise<ToolSet | undefined> {
        const [existing] = await db.select().from(toolSetsTable).where(eq(toolSetsTable.uuid, input.uuid));
        if (!existing) {
            return undefined;
        }

        const [updatedSet] = await db
            .update(toolSetsTable)
            .set({
                name: input.name ?? existing.name,
                description: input.description === undefined ? existing.description : input.description,
                user_id: input.user_id === undefined ? existing.user_id : input.user_id,
            })
            .where(eq(toolSetsTable.uuid, input.uuid))
            .returning();

        if (input.tools) {
            await db.delete(toolSetItemsTable).where(eq(toolSetItemsTable.tool_set_uuid, input.uuid));
            await this.addTools(input.uuid, Array.from(new Set(input.tools)));
        }

        const [hydrated] = await this.hydrateBatch([updatedSet]);
        return hydrated;
    }

    async deleteByUuid(uuid: string): Promise<void> {
        await db.delete(toolSetsTable).where(eq(toolSetsTable.uuid, uuid));
    }

    private async addTools(toolSetUuid: string, toolUuids: string[]) {
        if (toolUuids.length === 0) return;
        const items: ToolSetItemInsert[] = toolUuids.map((toolUuid) => ({
            uuid: randomUUID(),
            tool_set_uuid: toolSetUuid,
            tool_uuid: toolUuid,
        }));

        await db.insert(toolSetItemsTable).values(items);
    }

    private async hydrateBatch(sets: ToolSetRow[]): Promise<ToolSet[]> {
        if (sets.length === 0) return [];

        const setUuids = sets.map((s) => s.uuid);
        const allItems = await db
            .select()
            .from(toolSetItemsTable)
            .where(inArray(toolSetItemsTable.tool_set_uuid, setUuids));

        const itemsBySet = allItems.reduce((acc, item) => {
            if (!acc[item.tool_set_uuid]) {
                acc[item.tool_set_uuid] = [];
            }
            acc[item.tool_set_uuid].push(item.tool_uuid);
            return acc;
        }, {} as Record<string, string[]>);

        return sets.map((s) => ({
            uuid: s.uuid,
            name: s.name,
            description: s.description,
            tools: itemsBySet[s.uuid] || [],
        }));
    }
}

export const toolSetsRepository = new ToolSetsRepository();
