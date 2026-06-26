/**
 * @file namespaces.repo.ts
 * @module packages/core/src/db/repositories/namespaces.repo
 *
 * WHAT:
 * Repository for managing Namespaces.
 *
 * WHY:
 * Handles logical grouping of MCP servers/tools.
 * Manages mappings between Namespaces, Servers, and Tools.
 *
 * HOW:
 * - Uses inferred Drizzle typing instead of cast/suppress patterns.
 * - Preserves mapping refresh semantics during namespace update.
 */

import {
    DatabaseNamespace,
    DatabaseNamespaceTool,
    DatabaseNamespaceWithServers,
    NamespaceCreateInput,
    NamespaceUpdateInput,
} from "../../types/mcp-admin/index.js";
import { and, desc, eq, inArray, isNull, or } from "drizzle-orm";

import { db } from "../index.js";
import {
    mcpServersTable,
    namespaceServerMappingsTable,
    namespacesTable,
    namespaceToolMappingsTable,
    toolsTable,
} from "../mcp-admin-schema.js";
import { randomUUID } from "node:crypto";
import { namespaceMappingsRepository } from "./namespace-mappings.repo.js";

type NamespaceInsert = typeof namespacesTable.$inferInsert;
type NamespaceServerMappingInsert = typeof namespaceServerMappingsTable.$inferInsert;
type NamespaceToolMappingInsert = typeof namespaceToolMappingsTable.$inferInsert;

const namespaceSelect = {
    uuid: namespacesTable.uuid,
    name: namespacesTable.name,
    description: namespacesTable.description,
    created_at: namespacesTable.created_at,
    updated_at: namespacesTable.updated_at,
    user_id: namespacesTable.user_id,
} as const;

export class NamespacesRepository {
    async create(input: NamespaceCreateInput): Promise<DatabaseNamespace> {
        const payload: NamespaceInsert = {
            uuid: randomUUID(),
            name: input.name,
            description: input.description ?? null,
            user_id: input.user_id ?? null,
        };

        const [createdNamespace] = await db
            .insert(namespacesTable)
            .values(payload)
            .returning(namespaceSelect);

        if (!createdNamespace) {
            throw new Error("Failed to create namespace");
        }

        // If mcp server UUIDs are provided, create the mappings with default ACTIVE status.
        if (input.mcpServerUuids && input.mcpServerUuids.length > 0) {
            const mappings: NamespaceServerMappingInsert[] = input.mcpServerUuids.map((serverUuid) => ({
                uuid: randomUUID(),
                namespace_uuid: createdNamespace.uuid,
                mcp_server_uuid: serverUuid,
                status: "ACTIVE",
            }));

            await db.insert(namespaceServerMappingsTable).values(mappings);

            // Also create namespace-tool mappings for all tools of the selected servers.
            const serverTools = await db
                .select({
                    uuid: toolsTable.uuid,
                    mcp_server_uuid: toolsTable.mcp_server_uuid,
                })
                .from(toolsTable)
                .where(inArray(toolsTable.mcp_server_uuid, input.mcpServerUuids));

            if (serverTools.length > 0) {
                const toolMappings: NamespaceToolMappingInsert[] = serverTools.map((tool) => ({
                    uuid: randomUUID(),
                    namespace_uuid: createdNamespace.uuid,
                    tool_uuid: tool.uuid,
                    mcp_server_uuid: tool.mcp_server_uuid,
                    status: "ACTIVE",
                }));

                await db.insert(namespaceToolMappingsTable).values(toolMappings);
            }
        }

        return createdNamespace;
    }

    async findAll(): Promise<DatabaseNamespace[]> {
        return await db
            .select(namespaceSelect)
            .from(namespacesTable)
            .orderBy(desc(namespacesTable.created_at));
    }

    // Find namespaces accessible to a specific user (public + user's own namespaces)
    async findAllAccessibleToUser(userId: string): Promise<DatabaseNamespace[]> {
        return await db
            .select(namespaceSelect)
            .from(namespacesTable)
            .where(
                or(
                    isNull(namespacesTable.user_id),
                    eq(namespacesTable.user_id, userId),
                ),
            )
            .orderBy(desc(namespacesTable.created_at));
    }

    // Find only public namespaces (no user ownership)
    async findPublicNamespaces(): Promise<DatabaseNamespace[]> {
        return await db
            .select(namespaceSelect)
            .from(namespacesTable)
            .where(isNull(namespacesTable.user_id))
            .orderBy(desc(namespacesTable.created_at));
    }

    // Find namespaces owned by a specific user
    async findByUserId(userId: string): Promise<DatabaseNamespace[]> {
        return await db
            .select(namespaceSelect)
            .from(namespacesTable)
            .where(eq(namespacesTable.user_id, userId))
            .orderBy(desc(namespacesTable.created_at));
    }

    async findByUuid(uuid: string): Promise<DatabaseNamespace | undefined> {
        const [namespace] = await db
            .select(namespaceSelect)
            .from(namespacesTable)
            .where(eq(namespacesTable.uuid, uuid));

        return namespace;
    }

    // Find namespace by name within user scope (for uniqueness checks)
    async findByNameAndUserId(
        name: string,
        userId: string | null,
    ): Promise<DatabaseNamespace | undefined> {
        const [namespace] = await db
            .select(namespaceSelect)
            .from(namespacesTable)
            .where(
                and(
                    eq(namespacesTable.name, name),
                    userId
                        ? eq(namespacesTable.user_id, userId)
                        : isNull(namespacesTable.user_id),
                ),
            )
            .limit(1);

        return namespace;
    }

    async findByUuidWithServers(
        uuid: string,
    ): Promise<DatabaseNamespaceWithServers | null> {
        // First, get the namespace.
        const namespace = await this.findByUuid(uuid);

        if (!namespace) {
            return null;
        }

        // Then, get servers associated with this namespace.
        const serversData = await db
            .select({
                uuid: mcpServersTable.uuid,
                name: mcpServersTable.name,
                description: mcpServersTable.description,
                type: mcpServersTable.type,
                command: mcpServersTable.command,
                args: mcpServersTable.args,
                url: mcpServersTable.url,
                env: mcpServersTable.env,
                bearerToken: mcpServersTable.bearerToken,
                headers: mcpServersTable.headers,
                error_status: mcpServersTable.error_status,
                created_at: mcpServersTable.created_at,
                user_id: mcpServersTable.user_id,
                status: namespaceServerMappingsTable.status,
            })
            .from(mcpServersTable)
            .innerJoin(
                namespaceServerMappingsTable,
                eq(mcpServersTable.uuid, namespaceServerMappingsTable.mcp_server_uuid),
            )
            .where(eq(namespaceServerMappingsTable.namespace_uuid, uuid));

        const servers = serversData.map((server) => ({
            ...server,
            args: server.args || [],
            env: server.env || {},
            headers: server.headers || {},
        }));

        return {
            ...namespace,
            servers,
        } as DatabaseNamespaceWithServers;
    }

    async findToolsByNamespaceUuid(
        namespaceUuid: string,
    ): Promise<DatabaseNamespaceTool[]> {
        const toolsData = await db
            .select({
                // Tool fields
                uuid: toolsTable.uuid,
                name: toolsTable.name,
                description: toolsTable.description,
                toolSchema: toolsTable.toolSchema,
                created_at: toolsTable.created_at,
                updated_at: toolsTable.updated_at,
                mcp_server_uuid: toolsTable.mcp_server_uuid,
                // Server fields
                serverName: mcpServersTable.name,
                serverUuid: mcpServersTable.uuid,
                // Namespace mapping fields
                status: namespaceToolMappingsTable.status,
                overrideName: namespaceToolMappingsTable.override_name,
                overrideTitle: namespaceToolMappingsTable.override_title,
                overrideDescription: namespaceToolMappingsTable.override_description,
                overrideAnnotations: namespaceToolMappingsTable.override_annotations,
            })
            .from(toolsTable)
            .innerJoin(
                namespaceToolMappingsTable,
                eq(toolsTable.uuid, namespaceToolMappingsTable.tool_uuid),
            )
            .innerJoin(
                mcpServersTable,
                eq(toolsTable.mcp_server_uuid, mcpServersTable.uuid),
            )
            .where(eq(namespaceToolMappingsTable.namespace_uuid, namespaceUuid))
            .orderBy(desc(toolsTable.created_at));

        return toolsData as DatabaseNamespaceTool[];
    }

    async deleteByUuid(uuid: string): Promise<DatabaseNamespace | undefined> {
        const [deletedNamespace] = await db
            .delete(namespacesTable)
            .where(eq(namespacesTable.uuid, uuid))
            .returning(namespaceSelect);

        return deletedNamespace;
    }

    async update(input: NamespaceUpdateInput): Promise<DatabaseNamespace> {
        // Update the namespace.
        const [updatedNamespace] = await db
            .update(namespacesTable)
            .set({
                name: input.name,
                description: input.description ?? null,
                user_id: input.user_id ?? null,
                updated_at: new Date(),
            })
            .where(eq(namespacesTable.uuid, input.uuid))
            .returning(namespaceSelect);

        if (!updatedNamespace) {
            throw new Error("Namespace not found");
        }

        // If mcpServerUuids are provided, update the mappings.
        if (input.mcpServerUuids) {
            // Get existing tool mappings to preserve their status.
            const existingToolMappings =
                await namespaceMappingsRepository.findToolMappingsByNamespace(
                    input.uuid,
                );
            const existingToolStatusMap = new Map<
                string,
                NonNullable<NamespaceToolMappingInsert["status"]>
            >();

            // Create a map of existing tool statuses by tool_uuid.
            existingToolMappings.forEach((mapping) => {
                existingToolStatusMap.set(mapping.tool_uuid, mapping.status);
            });

            // Delete existing server mappings.
            await db
                .delete(namespaceServerMappingsTable)
                .where(eq(namespaceServerMappingsTable.namespace_uuid, input.uuid));

            // Delete existing tool mappings.
            await db
                .delete(namespaceToolMappingsTable)
                .where(eq(namespaceToolMappingsTable.namespace_uuid, input.uuid));

            // Create new server mappings if any servers are specified.
            if (input.mcpServerUuids.length > 0) {
                const serverMappings: NamespaceServerMappingInsert[] = input.mcpServerUuids.map((serverUuid) => ({
                    uuid: randomUUID(),
                    namespace_uuid: input.uuid,
                    mcp_server_uuid: serverUuid,
                    status: "ACTIVE",
                }));

                await db.insert(namespaceServerMappingsTable).values(serverMappings);

                // Also create namespace-tool mappings for all tools of the selected servers.
                const serverTools = await db
                    .select({
                        uuid: toolsTable.uuid,
                        mcp_server_uuid: toolsTable.mcp_server_uuid,
                    })
                    .from(toolsTable)
                    .where(inArray(toolsTable.mcp_server_uuid, input.mcpServerUuids));

                if (serverTools.length > 0) {
                    const toolMappings: NamespaceToolMappingInsert[] = serverTools.map((tool) => ({
                        uuid: randomUUID(),
                        namespace_uuid: input.uuid,
                        tool_uuid: tool.uuid,
                        mcp_server_uuid: tool.mcp_server_uuid,
                        // Preserve existing status if tool was previously mapped, otherwise default to ACTIVE.
                        status:
                            existingToolStatusMap.get(tool.uuid) || "ACTIVE",
                    }));

                    await db.insert(namespaceToolMappingsTable).values(toolMappings);
                }
            }
        }

        return updatedNamespace;
    }
}

export const namespacesRepository = new NamespacesRepository();
