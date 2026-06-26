import { eq, desc } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { db } from "../index.js";
import { savedScriptsTable } from "../mcp-admin-schema.js";
import { SavedScript } from "../../types/tormentnexus/saved-scripts.zod.js";

type SavedScriptRow = typeof savedScriptsTable.$inferSelect;
type SavedScriptInsert = typeof savedScriptsTable.$inferInsert;

export class SavedScriptsRepository {
    async findAll(): Promise<SavedScript[]> {
        const scripts = await db.select().from(savedScriptsTable).orderBy(desc(savedScriptsTable.created_at));
        return scripts.map((script) => this.mapToDomain(script));
    }

    async findByUuid(uuid: string): Promise<SavedScript | undefined> {
        const [script] = await db.select().from(savedScriptsTable).where(eq(savedScriptsTable.uuid, uuid));
        return script ? this.mapToDomain(script) : undefined;
    }

    async create(input: { name: string; description?: string | null; code: string; userId?: string | null }): Promise<SavedScript> {
        const payload: SavedScriptInsert = {
            uuid: randomUUID(),
            name: input.name,
            description: input.description ?? null,
            code: input.code,
            user_id: input.userId ?? null,
        };

        const [script] = await db.insert(savedScriptsTable).values(payload).returning();
        return this.mapToDomain(script);
    }

    async update(uuid: string, input: { name?: string; description?: string | null; code?: string }): Promise<SavedScript> {
        const [script] = await db.update(savedScriptsTable)
            .set({
                ...(input.name && { name: input.name }),
                ...(input.description !== undefined && { description: input.description }),
                ...(input.code && { code: input.code }),
                updated_at: new Date(),
            })
            .where(eq(savedScriptsTable.uuid, uuid))
            .returning();

        if (!script) throw new Error("Script not found");
        return this.mapToDomain(script);
    }

    async delete(uuid: string): Promise<void> {
        await db.delete(savedScriptsTable).where(eq(savedScriptsTable.uuid, uuid));
    }

    private mapToDomain(dbScript: SavedScriptRow): SavedScript {
        return {
            uuid: dbScript.uuid,
            name: dbScript.name,
            description: dbScript.description,
            code: dbScript.code,
            userId: dbScript.user_id,
            createdAt: new Date(dbScript.created_at),
            updatedAt: new Date(dbScript.updated_at),
        };
    }
}

export const savedScriptsRepository = new SavedScriptsRepository();
