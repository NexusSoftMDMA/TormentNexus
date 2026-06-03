import { z } from 'zod';
import { TRPCError } from '@trpc/server';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { mcpServersRepository } from '../db/repositories/index.js';
import fs from 'node:fs/promises';
import path from 'node:path';
import {
    McpServerCreateInputSchema,
    McpServerUpdateInputSchema
} from '../types/mcp-admin/index.js';
import { loadTormentNexusMcpConfig } from '../mcp/mcpJsonConfig.js';
import { clientConfigSyncService, SUPPORTED_MCP_CLIENTS } from '../mcp/clientConfigSync.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

const MASTER_INDEX_PATH = path.join(process.cwd(), 'TORMENTNEXUS_MASTER_INDEX.jsonc');

const stripJsonComments = (content: string) =>
    content.replace(/\\"|"(?:\\"|[^"])*"|(\/\/.*|\/\*[\s\S]*?\*\/)/g, (m, g) => g ? '' : m);

function getContextUserId(ctx: unknown): string | undefined {
    if (!ctx || typeof ctx !== 'object') {
        return undefined;
    }

    const session = (ctx as { session?: unknown }).session;
    if (!session || typeof session !== 'object') {
        return undefined;
    }

    const user = (session as { user?: unknown }).user;
    if (!user || typeof user !== 'object') {
        return undefined;
    }

    const userId = (user as { id?: unknown }).id;
    return typeof userId === 'string' && userId.trim().length > 0
        ? userId
        : undefined;
}

export const mcpServersRouter = t.router({
    list: publicProcedure.query(async ({ ctx }) => {
        try {
            const userId = getContextUserId(ctx);
            const [servers, config] = await Promise.all([
                mcpServersRepository.findAll(userId),
                loadTormentNexusMcpConfig(),
            ]);

            return servers.map((server) => ({
                ...server,
                _meta: config.mcpServers?.[server.name]?._meta ?? null,
            }));
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
        }
    }),

    get: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .query(async ({ input }) => {
            try {
                const [server, config] = await Promise.all([
                    mcpServersRepository.findByUuid(input.uuid),
                    loadTormentNexusMcpConfig(),
                ]);

                if (!server) {
                    return undefined;
                }

                return {
                    ...server,
                    _meta: config.mcpServers?.[server.name]?._meta ?? null,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
            }
        }),

    create: adminProcedure
        .input(McpServerCreateInputSchema)
        .mutation(async ({ input }) => {
            try {
                const { metadataStrategy, ...serverInput } = input;
                return await mcpServersRepository.create(serverInput, { metadataStrategy });
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
            }
        }),

    update: adminProcedure
        .input(McpServerUpdateInputSchema)
        .mutation(async ({ input }) => {
            try {
                if (!input.uuid) throw new Error("UUID required for update");
                const { metadataStrategy, ...serverInput } = input;
                return await mcpServersRepository.update(serverInput, { metadataStrategy });
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
            }
        }),

    reloadMetadata: adminProcedure
        .input(z.object({
            uuid: z.string(),
            mode: z.enum(['auto', 'binary', 'cache']).default('binary'),
        }))
        .mutation(async ({ input }) => {
            try {
                return await mcpServersRepository.reloadMetadata(input.uuid, input.mode);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
            }
        }),

    clearMetadataCache: adminProcedure
        .input(z.object({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            try {
                return await mcpServersRepository.clearMetadataCache(input.uuid);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
            }
        }),

    delete: adminProcedure
        .input(z.object({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            try {
                return await mcpServersRepository.deleteByUuid(input.uuid);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
            }
        }),

    bulkImport: adminProcedure
        .input(z.array(McpServerCreateInputSchema))
        .mutation(async ({ input }) => {
            try {
                return await mcpServersRepository.bulkCreate(input);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('MCP server registry is unavailable', error);
            }
        }),

    syncTargets: publicProcedure.query(async () => {
        return await clientConfigSyncService.listTargets();
    }),

    exportClientConfig: publicProcedure
        .input(z.object({
            client: z.enum(SUPPORTED_MCP_CLIENTS),
            path: z.string().optional(),
        }))
        .query(async ({ input }) => {
            return await clientConfigSyncService.previewSync(input.client, input.path);
        }),

    syncClientConfig: adminProcedure
        .input(z.object({
            client: z.enum(SUPPORTED_MCP_CLIENTS),
            path: z.string().optional(),
        }))
        .mutation(async ({ input }) => {
            return await clientConfigSyncService.syncClientConfig(input.client, input.path);
        }),

    registrySnapshot: publicProcedure.query(async () => {
        try {
            const content = await fs.readFile(MASTER_INDEX_PATH, 'utf-8');
            const parsed = JSON.parse(stripJsonComments(content)) as {
                categories?: Record<string, Array<Record<string, unknown>>>;
            };

            const categories = parsed.categories ?? {};
            const collected: Array<{
                id: string;
                name: string;
                url: string;
                category: string;
                description: string;
                tags: string[];
            }> = [];

            for (const [category, items] of Object.entries(categories)) {
                if (!Array.isArray(items)) continue;

                for (const raw of items) {
                    const url = String(raw?.url ?? '');
                    if (!url) continue;

                    const kind = String(raw?.kind ?? '').toLowerCase();
                    const tags = Array.isArray(raw?.tags)
                        ? raw.tags.map((t) => String(t).toLowerCase())
                        : [];

                    const isMcpLike =
                        category.toLowerCase().includes('mcp') ||
                        kind.includes('mcp') ||
                        tags.some((t) => t.includes('mcp')) ||
                        url.toLowerCase().includes('modelcontextprotocol') ||
                        url.toLowerCase().includes('mcp');

                    if (!isMcpLike) continue;

                    const id = String(raw?.id ?? url);
                    const name = String(raw?.name ?? raw?.id ?? url);
                    const description = String(raw?.description ?? raw?.summary ?? 'No description available.');

                    collected.push({
                        id,
                        name,
                        url,
                        category,
                        description,
                        tags,
                    });
                }
            }

            const deduped = new Map<string, (typeof collected)[number]>();
            for (const item of collected) {
                if (!deduped.has(item.url)) {
                    deduped.set(item.url, item);
                }
            }

            return Array.from(deduped.values()).slice(0, 300);
        } catch (error) {
            const errorCode = (error as NodeJS.ErrnoException | undefined)?.code;
            if (errorCode === 'ENOENT') {
                return [];
            }

            const detail = error instanceof SyntaxError
                ? 'TORMENTNEXUS_MASTER_INDEX.jsonc contains invalid JSON.'
                : error instanceof Error
                    ? error.message
                    : String(error);

            throw new TRPCError({
                code: 'INTERNAL_SERVER_ERROR',
                message: `Registry snapshot is unavailable: ${detail}`,
            });
        }
    }),

});
