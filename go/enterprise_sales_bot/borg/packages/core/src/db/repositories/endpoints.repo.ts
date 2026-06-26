/**
 * @file endpoints.repo.ts
 * @module packages/core/src/db/repositories/endpoints.repo
 *
 * WHAT:
 * Repository for managing MCP Endpoints (Gateways).
 *
 * WHY:
 * Handles CRUD for Endpoints, including user access control and namespace linking.
 *
 * HOW:
 * - Uses shared select projections to keep endpoint result shapes consistent.
 * - Uses inferred Drizzle insert/select types to avoid unsafe cast-based queries.
 */

import {
    DatabaseEndpoint,
    DatabaseEndpointWithNamespace,
    EndpointCreateInput,
    EndpointUpdateInput,
} from "../../types/mcp-admin/index.js";
import { and, desc, eq, isNull, or } from "drizzle-orm";
import { randomUUID } from "node:crypto";

import { db } from "../index.js";
import { endpointsTable, namespacesTable } from "../mcp-admin-schema.js";

type EndpointRow = typeof endpointsTable.$inferSelect;
type EndpointInsert = typeof endpointsTable.$inferInsert;

const endpointSelect = {
    uuid: endpointsTable.uuid,
    name: endpointsTable.name,
    description: endpointsTable.description,
    namespace_uuid: endpointsTable.namespace_uuid,
    enable_api_key_auth: endpointsTable.enable_api_key_auth,
    enable_oauth: endpointsTable.enable_oauth,
    enable_max_rate: endpointsTable.enable_max_rate,
    enable_client_max_rate: endpointsTable.enable_client_max_rate,
    max_rate: endpointsTable.max_rate,
    client_max_rate: endpointsTable.client_max_rate,
    max_rate_seconds: endpointsTable.max_rate_seconds,
    client_max_rate_seconds: endpointsTable.client_max_rate_seconds,
    client_max_rate_strategy: endpointsTable.client_max_rate_strategy,
    client_max_rate_strategy_key: endpointsTable.client_max_rate_strategy_key,
    use_query_param_auth: endpointsTable.use_query_param_auth,
    created_at: endpointsTable.created_at,
    updated_at: endpointsTable.updated_at,
    user_id: endpointsTable.user_id,
} as const;

const endpointWithNamespaceSelect = {
    ...endpointSelect,
    namespace: {
        uuid: namespacesTable.uuid,
        name: namespacesTable.name,
        description: namespacesTable.description,
        created_at: namespacesTable.created_at,
        updated_at: namespacesTable.updated_at,
        user_id: namespacesTable.user_id,
    },
} as const;

export class EndpointsRepository {
    async create(input: EndpointCreateInput): Promise<DatabaseEndpoint> {
        const payload: EndpointInsert = {
            uuid: randomUUID(),
            name: input.name,
            description: input.description ?? null,
            namespace_uuid: input.namespace_uuid,
            enable_api_key_auth: input.enable_api_key_auth ?? true,
            enable_oauth: input.enable_oauth ?? false,
            enable_max_rate: input.enable_max_rate,
            enable_client_max_rate: input.enable_client_max_rate,
            max_rate: input.max_rate ?? null,
            max_rate_seconds: input.max_rate_seconds ?? null,
            client_max_rate: input.client_max_rate ?? null,
            client_max_rate_seconds: input.client_max_rate_seconds ?? null,
            client_max_rate_strategy: input.client_max_rate_strategy ?? null,
            client_max_rate_strategy_key: input.client_max_rate_strategy_key ?? null,
            use_query_param_auth: input.use_query_param_auth ?? false,
            user_id: input.user_id ?? null,
        };

        const [endpoint] = await db.insert(endpointsTable).values(payload).returning(endpointSelect);

        if (!endpoint) {
            throw new Error("Failed to create endpoint");
        }

        return endpoint;
    }

    async findAll(): Promise<DatabaseEndpoint[]> {
        return await db
            .select(endpointSelect)
            .from(endpointsTable)
            .orderBy(desc(endpointsTable.created_at));
    }

    // Find endpoints accessible to a specific user (public + user's own endpoints)
    async findAllAccessibleToUser(userId: string): Promise<DatabaseEndpoint[]> {
        return await db
            .select(endpointSelect)
            .from(endpointsTable)
            .where(or(isNull(endpointsTable.user_id), eq(endpointsTable.user_id, userId)))
            .orderBy(desc(endpointsTable.created_at));
    }

    // Find endpoints accessible to a specific user with namespace data (public + user's own endpoints)
    async findAllAccessibleToUserWithNamespaces(
        userId: string,
    ): Promise<DatabaseEndpointWithNamespace[]> {
        const endpointsData = await db
            .select(endpointWithNamespaceSelect)
            .from(endpointsTable)
            .innerJoin(
                namespacesTable,
                eq(endpointsTable.namespace_uuid, namespacesTable.uuid),
            )
            .where(or(isNull(endpointsTable.user_id), eq(endpointsTable.user_id, userId)))
            .orderBy(desc(endpointsTable.created_at));

        return endpointsData as DatabaseEndpointWithNamespace[];
    }

    // Find only public endpoints (no user ownership)
    async findPublicEndpoints(): Promise<DatabaseEndpoint[]> {
        return await db
            .select(endpointSelect)
            .from(endpointsTable)
            .where(isNull(endpointsTable.user_id))
            .orderBy(desc(endpointsTable.created_at));
    }

    // Find endpoints owned by a specific user
    async findByUserId(userId: string): Promise<DatabaseEndpoint[]> {
        return await db
            .select(endpointSelect)
            .from(endpointsTable)
            .where(eq(endpointsTable.user_id, userId))
            .orderBy(desc(endpointsTable.created_at));
    }

    async findAllWithNamespaces(): Promise<DatabaseEndpointWithNamespace[]> {
        const endpointsData = await db
            .select(endpointWithNamespaceSelect)
            .from(endpointsTable)
            .innerJoin(
                namespacesTable,
                eq(endpointsTable.namespace_uuid, namespacesTable.uuid),
            )
            .orderBy(desc(endpointsTable.created_at));

        return endpointsData as DatabaseEndpointWithNamespace[];
    }

    async findByUuid(uuid: string): Promise<DatabaseEndpoint | undefined> {
        const [endpoint] = await db
            .select(endpointSelect)
            .from(endpointsTable)
            .where(eq(endpointsTable.uuid, uuid));

        return endpoint;
    }

    async findByUuidWithNamespace(
        uuid: string,
    ): Promise<DatabaseEndpointWithNamespace | undefined> {
        const [endpointData] = await db
            .select(endpointWithNamespaceSelect)
            .from(endpointsTable)
            .innerJoin(
                namespacesTable,
                eq(endpointsTable.namespace_uuid, namespacesTable.uuid),
            )
            .where(eq(endpointsTable.uuid, uuid));

        return endpointData as DatabaseEndpointWithNamespace | undefined;
    }

    async findByName(name: string): Promise<DatabaseEndpoint | undefined> {
        const [endpoint] = await db
            .select(endpointSelect)
            .from(endpointsTable)
            .where(eq(endpointsTable.name, name));

        return endpoint;
    }

    // Find endpoint by name within user scope (for uniqueness checks)
    async findByNameAndUserId(
        name: string,
        userId: string | null,
    ): Promise<DatabaseEndpoint | undefined> {
        const [endpoint] = await db
            .select(endpointSelect)
            .from(endpointsTable)
            .where(
                and(
                    eq(endpointsTable.name, name),
                    userId
                        ? eq(endpointsTable.user_id, userId)
                        : isNull(endpointsTable.user_id),
                ),
            )
            .limit(1);

        return endpoint;
    }

    async deleteByUuid(uuid: string): Promise<DatabaseEndpoint | undefined> {
        const [deletedEndpoint] = await db
            .delete(endpointsTable)
            .where(eq(endpointsTable.uuid, uuid))
            .returning(endpointSelect);

        return deletedEndpoint;
    }

    async update(input: EndpointUpdateInput): Promise<DatabaseEndpoint> {
        const { uuid, ...updates } = input;
        const payload: Partial<EndpointInsert> = {
            ...updates,
            description: updates.description ?? null,
            max_rate: updates.max_rate ?? null,
            max_rate_seconds: updates.max_rate_seconds ?? null,
            client_max_rate: updates.client_max_rate ?? null,
            client_max_rate_seconds: updates.client_max_rate_seconds ?? null,
            client_max_rate_strategy: updates.client_max_rate_strategy ?? null,
            client_max_rate_strategy_key: updates.client_max_rate_strategy_key ?? null,
            user_id: updates.user_id ?? null,
            updated_at: new Date(),
        };

        const [updatedEndpoint] = await db
            .update(endpointsTable)
            .set(payload)
            .where(eq(endpointsTable.uuid, uuid))
            .returning(endpointSelect);

        if (!updatedEndpoint) {
            throw new Error("Failed to update endpoint");
        }

        return updatedEndpoint;
    }
}

// Export the repository instance
export const endpointsRepository = new EndpointsRepository();
