/**
 * @file api-keys.repo.ts
 * @module packages/core/src/db/repositories/api-keys.repo
 *
 * WHAT:
 * Repository for managing API Keys.
 *
 * WHY:
 * Handles secure creation (generation) and management of API Keys.
 * Implements key generation using `nanoid`.
 */

import {
    ApiKeyCreateInput,
    ApiKeyType,
    ApiKeyUpdateInput,
} from "../../types/mcp-admin/index.js";
import { and, desc, eq, isNull, or } from "drizzle-orm";
import { customAlphabet } from "nanoid";

import { db } from "../index.js";
import { apiKeysTable } from "../mcp-admin-schema.js";
import { randomUUID } from "node:crypto";

type ApiKeyRow = typeof apiKeysTable.$inferSelect;
type ApiKeyInsert = typeof apiKeysTable.$inferInsert;

const nanoid = customAlphabet(
    "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
    64,
);

export class ApiKeysRepository {
    /**
     * Generate a new API key with the specified format: sk_mt_{64-char-nanoid}
     */
    private generateApiKey(): string {
        const keyPart = nanoid();
        const key = `sk_mt_${keyPart}`;

        return key;
    }

    async create(input: ApiKeyCreateInput): Promise<{
        uuid: string;
        name: string;
        key: string;
        user_id: string | null;
        created_at: Date;
        is_active: boolean;
        type: ApiKeyType;
    }> {
        const key = this.generateApiKey();
        const payload: ApiKeyInsert = {
            uuid: randomUUID(),
            name: input.name,
            key,
            user_id: input.user_id ?? null,
            is_active: input.is_active ?? true,
        };

        const [createdApiKey] = await db
            .insert(apiKeysTable)
            .values(payload)
            .returning({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key, // Original code returns 'key' here too
                user_id: apiKeysTable.user_id,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
            });

        if (!createdApiKey) {
            throw new Error("Failed to create API key");
        }

        return {
            ...createdApiKey,
            type: "MCP",
            key, // Return the actual key (redundant if returning above, but safe)
        };
    }

    async findByUserId(userId: string): Promise<Array<Pick<ApiKeyRow, "uuid" | "name" | "key" | "created_at" | "is_active">>> {
        return await db
            .select({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
            })
            .from(apiKeysTable)
            .where(eq(apiKeysTable.user_id, userId))
            .orderBy(desc(apiKeysTable.created_at));
    }

    // Find all API keys (both public and user-owned)
    async findAll(): Promise<Array<Pick<ApiKeyRow, "uuid" | "name" | "key" | "created_at" | "is_active" | "user_id">>> {
        return await db
            .select({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
                user_id: apiKeysTable.user_id,
            })
            .from(apiKeysTable)
            .orderBy(desc(apiKeysTable.created_at));
    }

    // Find public API keys (no user ownership)
    async findPublicApiKeys(): Promise<Array<Pick<ApiKeyRow, "uuid" | "name" | "key" | "created_at" | "is_active" | "user_id">>> {
        return await db
            .select({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
                user_id: apiKeysTable.user_id,
            })
            .from(apiKeysTable)
            .where(isNull(apiKeysTable.user_id))
            .orderBy(desc(apiKeysTable.created_at));
    }

    // Find API keys accessible to a specific user (public + user's own keys)
    async findAccessibleToUser(userId: string): Promise<Array<Pick<ApiKeyRow, "uuid" | "name" | "key" | "created_at" | "is_active" | "user_id">>> {
        return await db
            .select({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
                user_id: apiKeysTable.user_id,
            })
            .from(apiKeysTable)
            .where(
                or(
                    isNull(apiKeysTable.user_id), // Public API keys
                    eq(apiKeysTable.user_id, userId), // User's own API keys
                ),
            )
            .orderBy(desc(apiKeysTable.created_at));
    }

    async findByUuid(uuid: string, userId: string): Promise<Pick<ApiKeyRow, "uuid" | "name" | "key" | "created_at" | "is_active" | "user_id"> | undefined> {
        const [apiKey] = await db
            .select({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
                user_id: apiKeysTable.user_id,
            })
            .from(apiKeysTable)
            .where(
                and(eq(apiKeysTable.uuid, uuid), eq(apiKeysTable.user_id, userId)),
            );

        return apiKey;
    }

    // Find API key by UUID with access control (user can access their own keys + public keys)
    async findByUuidWithAccess(uuid: string, userId?: string): Promise<Pick<ApiKeyRow, "uuid" | "name" | "key" | "created_at" | "is_active" | "user_id"> | undefined> {
        const [apiKey] = await db
            .select({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
                user_id: apiKeysTable.user_id,
            })
            .from(apiKeysTable)
            .where(
                and(
                    eq(apiKeysTable.uuid, uuid),
                    userId
                        ? or(
                            isNull(apiKeysTable.user_id), // Public API keys
                            eq(apiKeysTable.user_id, userId), // User's own API keys
                        )
                        : isNull(apiKeysTable.user_id), // Only public if no user context
                ),
            );

        return apiKey;
    }

    async validateApiKey(key: string): Promise<{
        valid: boolean;
        user_id?: string | null;
        key_uuid?: string;
        type?: ApiKeyType;
    }> {
        const [apiKey] = await db
            .select({
                uuid: apiKeysTable.uuid,
                user_id: apiKeysTable.user_id,
                is_active: apiKeysTable.is_active,
            })
            .from(apiKeysTable)
            .where(eq(apiKeysTable.key, key));

        if (!apiKey) {
            return { valid: false };
        }

        // Check if key is active
        if (!apiKey.is_active) {
            return { valid: false };
        }

        return {
            valid: true,
            user_id: apiKey.user_id,
            key_uuid: apiKey.uuid,
            type: "MCP", // Hardcoded default as schema removed it
        };
    }

    async update(uuid: string, userId: string, input: ApiKeyUpdateInput): Promise<Pick<ApiKeyRow, "uuid" | "name" | "key" | "created_at" | "is_active">> {
        const [updatedApiKey] = await db
            .update(apiKeysTable)
            .set({
                ...(input.name && { name: input.name }),
                // ...(input.type && { type: input.type }), // Removed from schema
                ...(input.is_active !== undefined && { is_active: input.is_active }),
            })
            .where(
                and(
                    eq(apiKeysTable.uuid, uuid),
                    or(eq(apiKeysTable.user_id, userId), isNull(apiKeysTable.user_id)),
                ),
            )
            .returning({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
                key: apiKeysTable.key,
                created_at: apiKeysTable.created_at,
                is_active: apiKeysTable.is_active,
            });

        if (!updatedApiKey) {
            throw new Error("Failed to update API key or API key not found");
        }

        return updatedApiKey;
    }

    async delete(uuid: string, userId: string) {
        const [deletedApiKey] = await db
            .delete(apiKeysTable)
            .where(
                and(
                    eq(apiKeysTable.uuid, uuid),
                    or(eq(apiKeysTable.user_id, userId), isNull(apiKeysTable.user_id)),
                ),
            )
            .returning({
                uuid: apiKeysTable.uuid,
                name: apiKeysTable.name,
            });

        if (!deletedApiKey) {
            throw new Error("Failed to delete API key or API key not found");
        }

        return deletedApiKey;
    }
}

export const apiKeysRepository = new ApiKeysRepository();
