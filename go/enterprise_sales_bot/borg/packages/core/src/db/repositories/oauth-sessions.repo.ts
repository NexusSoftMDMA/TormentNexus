/**
 * @file oauth-sessions.repo.ts
 * @module packages/core/src/db/repositories/oauth-sessions.repo
 *
 * WHAT:
 * Repository for OAuth Sessions (User-Server connection state).
 *
 * WHY:
 * Manages the state of a user's OAuth connection to an external MCP Server (e.g. Google Drive).
 * Stores tokens securely (encrypted at rest ideally, but here just storage).
 */

import {
    OAuthSessionCreateInput,
    // OAuthSessionUpdateInput, // Unused
} from "../../types/mcp-admin/index.js";
import { desc, eq } from "drizzle-orm";
import { randomUUID } from "node:crypto";

import { db } from "../index.js";
import { oauthSessionsTable } from "../mcp-admin-schema.js";

type OAuthSessionRow = typeof oauthSessionsTable.$inferSelect;
type OAuthSessionInsert = typeof oauthSessionsTable.$inferInsert;

export class OAuthSessionsRepository {
    async upsert(input: OAuthSessionCreateInput): Promise<OAuthSessionRow> {
        // Check if session exists
        const [existingSession] = await db
            .select()
            .from(oauthSessionsTable)
            .where(eq(oauthSessionsTable.mcp_server_uuid, input.mcp_server_uuid));

        if (existingSession) {
            // Update
            const [updatedSession] = await db
                .update(oauthSessionsTable)
                .set({
                    // Merge input fields, keep existing if undefined in input (partial update logic handled by service usually, but here strict)
                    // @ts-expect-error -- type mismatch from inferred schema types
                    client_information:
                        input.client_information ?? existingSession.client_information,
                    // @ts-expect-error -- type mismatch from inferred schema types
                    tokens: input.tokens ?? existingSession.tokens,
                    code_verifier: input.code_verifier ?? existingSession.code_verifier,
                    updated_at: new Date(),
                })
                .where(eq(oauthSessionsTable.uuid, existingSession.uuid))
                .returning();

            return updatedSession;
        } else {
            // Create
            const payload: OAuthSessionInsert = {
                uuid: randomUUID(),
                mcp_server_uuid: input.mcp_server_uuid,
                // @ts-expect-error -- type mismatch from inferred schema types
                tokens: input.tokens ?? null,
                code_verifier: input.code_verifier ?? null,
                ...(input.client_information ? { client_information: input.client_information } : {}),
            };

            const [createdSession] = await db
                .insert(oauthSessionsTable)
                .values(payload)
                .returning();

            return createdSession;
        }
    }

    async findByMcpServerUuid(mcpServerUuid: string): Promise<OAuthSessionRow | undefined> {
        const [session] = await db
            .select()
            .from(oauthSessionsTable)
            .where(eq(oauthSessionsTable.mcp_server_uuid, mcpServerUuid));

        return session;
    }

    // Find all sessions (maintenance/admin)
    async findAll(): Promise<OAuthSessionRow[]> {
        return await db
            .select()
            .from(oauthSessionsTable)
            .orderBy(desc(oauthSessionsTable.updated_at));
    }

    async delete(uuid: string): Promise<OAuthSessionRow | undefined> {
        const [deletedSession] = await db
            .delete(oauthSessionsTable)
            .where(eq(oauthSessionsTable.uuid, uuid))
            .returning();

        return deletedSession;
    }
}

export const oauthSessionsRepository = new OAuthSessionsRepository();
