import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { toolRegistry, RegisteredTool } from '../services/ToolRegistry.js';
import { detectCliHarnesses } from '../services/cli-harness-detection.js';
import { detectLocalExecutionEnvironment } from '../services/execution-environment.js';
import { detectInstallSurfaceArtifacts } from '../services/install-surface-detection.js';
import { toolsRepository, mcpServersRepository } from '../db/repositories/index.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';
import {
    ToolCreateInputSchema,
    ToolUpsertInputSchema
} from '../types/tormentnexus/tools.zod.js';

export const toolsRouter = t.router({
    list: publicProcedure.query(async () => {
        try {
            const [tools, servers] = await Promise.all([
                toolsRepository.findAll(),
                mcpServersRepository.findAll()
            ]);
            const serverMap = new Map(servers.map(s => [s.uuid, s.name]));
            
            return tools.map(t => ({
                uuid: t.name,
                name: t.name,
                description: t.description,
                server: serverMap.get(t.mcp_server_uuid) || 'unknown',
                inputSchema: t.toolSchema,
                isDeferred: false,
                schemaParamCount: Object.keys((t.toolSchema as any)?.properties || {}).length,
                mcpServerUuid: t.mcp_server_uuid,
                always_on: t.always_on,
            }));
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Tool catalog is unavailable', error);
        }
    }),

    listByServer: publicProcedure
        .input(z.object({ mcpServerUuid: z.string() }))
        .query(async ({ input }) => {
            try {
                // Need to resolve if it's the name or UUID. For safety, get the server first.
                let serverUuid = input.mcpServerUuid;
                const allServers = await mcpServersRepository.findAll();
                const server = allServers.find(s => s.uuid === input.mcpServerUuid || s.name === input.mcpServerUuid);
                if (server) {
                    serverUuid = server.uuid;
                }

                const tools = await toolsRepository.findByMcpServerUuid(serverUuid);
                
                return tools.map(t => ({
                    uuid: t.name,
                    name: t.name,
                    description: t.description,
                    server: server?.name || 'unknown',
                    inputSchema: t.toolSchema,
                    isDeferred: false,
                    schemaParamCount: Object.keys((t.toolSchema as any)?.properties || {}).length,
                    mcpServerUuid: t.mcp_server_uuid,
                    always_on: t.always_on,
                }));
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool catalog is unavailable', error);
            }
        }),

    search: publicProcedure
        .input(z.object({
            query: z.string().min(1),
            limit: z.number().min(1).max(100).default(30),
        }))
        .query(async ({ input }) => {
            try {
                const [allTools, servers] = await Promise.all([
                    toolsRepository.findAll(),
                    mcpServersRepository.findAll()
                ]);
                
                const serverMap = new Map(servers.map(s => [s.uuid, s.name]));
                const q = input.query.toLowerCase();

                const filtered = allTools.filter(t => {
                    const name = String(t.name ?? '').toLowerCase();
                    const description = String(t.description ?? '').toLowerCase();
                    const serverName = String(serverMap.get(t.mcp_server_uuid) ?? '').toLowerCase();
                    return name.includes(q) || description.includes(q) || serverName.includes(q);
                });
                
                return filtered.slice(0, input.limit).map(t => ({
                    uuid: t.name,
                    name: t.name,
                    description: t.description,
                    server: serverMap.get(t.mcp_server_uuid) || 'unknown',
                    inputSchema: t.toolSchema,
                    isDeferred: false,
                    schemaParamCount: Object.keys((t.toolSchema as any)?.properties || {}).length,
                    mcpServerUuid: t.mcp_server_uuid,
                    always_on: t.always_on,
                }));
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool catalog is unavailable', error);
            }
        }),

    detectCliHarnesses: publicProcedure.query(async () => {
        return detectCliHarnesses();
    }),

    detectExecutionEnvironment: publicProcedure.query(async () => {
        return detectLocalExecutionEnvironment();
    }),

    detectInstallSurfaces: publicProcedure.query(async () => {
        return detectInstallSurfaceArtifacts();
    }),

    get: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .query(async ({ input }) => {
            try {
                // Treat input.uuid as tool name
                const allTools = await toolsRepository.findAll();
                const tool = allTools.find(t => t.name === input.uuid);
                if (!tool) return null;
                
                const server = await mcpServersRepository.findByUuid(tool.mcp_server_uuid);
                
                return {
                    uuid: tool.name,
                    name: tool.name,
                    description: tool.description,
                    server: server?.name || 'unknown',
                    inputSchema: tool.toolSchema,
                    isDeferred: false,
                    schemaParamCount: Object.keys((tool.toolSchema as any)?.properties || {}).length,
                    mcpServerUuid: tool.mcp_server_uuid,
                    always_on: tool.always_on,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool catalog is unavailable', error);
            }
        }),

    create: adminProcedure
        .input(ToolCreateInputSchema)
        .mutation(async ({ input }) => {
            try {
                await toolsRepository.create(input);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool catalog is unavailable', error);
            }
        }),

    upsertBatch: adminProcedure
        .input(ToolUpsertInputSchema)
        .mutation(async ({ input }) => {
            try {
                await toolsRepository.bulkUpsert(input);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool catalog is unavailable', error);
            }
        }),

    delete: adminProcedure
        .input(z.object({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            toolRegistry.deleteTool(input.uuid);
            return { success: true };
        }),

    setAlwaysOn: adminProcedure
        .input(z.object({ uuid: z.string(), alwaysOn: z.boolean() }))
        .mutation(async ({ input }) => {
            try {
                const tool = await toolsRepository.setAlwaysOn(input.uuid, input.alwaysOn);
                if (!tool) {
                    throw new Error(`Tool with UUID ${input.uuid} not found.`);
                }
                return { success: true, tool };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool catalog is unavailable', error);
            }
        }),
});
