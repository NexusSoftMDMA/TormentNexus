/**
 * @file docker-sessions.repo.ts
 * @module packages/core/src/db/repositories/docker-sessions.repo
 *
 * WHAT:
 * Repository for Docker Sessions.
 *
 * WHY:
 * Tracks running Docker containers for MCP servers created dynamically.
 */

// Schema definition for Create input (since not in Zod types explicitly yet? check imports)
// Assuming we match the table insert structure.
import { eq } from "drizzle-orm"; // Removed desc

import { db } from "../index.js";
import { dockerSessionsTable } from "../mcp-admin-schema.js";
import { randomUUID } from "node:crypto";

type DockerSessionRow = typeof dockerSessionsTable.$inferSelect;
type DockerSessionInsert = typeof dockerSessionsTable.$inferInsert;

export class DockerSessionsRepository {
    async create(input: Omit<DockerSessionInsert, "uuid">): Promise<DockerSessionRow> {
        const [session] = await db
            .insert(dockerSessionsTable)
            .values({
                ...input,
                uuid: randomUUID(),
            })
            .returning();

        return session;
    }

    async findByMcpServerUuid(mcpServerUuid: string): Promise<DockerSessionRow | undefined> {
        const [session] = await db
            .select()
            .from(dockerSessionsTable)
            .where(eq(dockerSessionsTable.mcp_server_uuid, mcpServerUuid));

        return session;
    }

    async updateStatus(containerId: string, status: DockerSessionRow["status"]): Promise<DockerSessionRow | undefined> {
        const [updatedSession] = await db
            .update(dockerSessionsTable)
            .set({
                status,
                updated_at: new Date(),
            })
            .where(eq(dockerSessionsTable.container_id, containerId))
            .returning();

        return updatedSession;
    }

    async delete(uuid: string): Promise<DockerSessionRow | undefined> {
        const [deletedSession] = await db
            .delete(dockerSessionsTable)
            .where(eq(dockerSessionsTable.uuid, uuid))
            .returning();

        return deletedSession;
    }
}

export const dockerSessionsRepository = new DockerSessionsRepository();
