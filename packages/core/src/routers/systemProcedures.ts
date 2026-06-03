import { z } from 'zod';
import {
    t,
    publicProcedure,
    adminProcedure,
    getMcpServer,
    getProjectTracker,
    getLspService,
    getMcpAggregator,
    getAgentMemoryService,
    getBrowserService,
    getSessionSupervisor,
    getMcpConfigService,
} from '../lib/trpc-core.js';
import { mcpServersRepository, toolsRepository } from '../db/repositories/index.js';
import { buildStartupStatusSnapshot } from './startupStatus.js';
import { detectLocalExecutionEnvironment } from '../services/execution-environment.js';
import { getCachedToolInventory } from '../mcp/cachedToolInventory.js';
import { readTormentNexusStoreStatus } from './memoryRouter.tormentnexus.js';
import { summarizeCachedInventory } from './startupInventorySummary.js';
import type { MemoryPipelineSummary } from '../services/memory/MemoryManager.js';
import { mcpServerPool } from '../services/mcp-server-pool.service.js';

const EXECUTION_ENV_CACHE_TTL_MS = Number(process.env.TORMENTNEXUS_EXECUTION_ENV_CACHE_TTL_MS ?? 30_000);

let executionEnvironmentCache:
    | {
        value: Awaited<ReturnType<typeof detectLocalExecutionEnvironment>> | null;
        expiresAt: number;
      }
    | null = null;
let executionEnvironmentInFlight: Promise<Awaited<ReturnType<typeof detectLocalExecutionEnvironment>> | null> | null = null;

async function getCachedExecutionEnvironment() {
    if (EXECUTION_ENV_CACHE_TTL_MS <= 0) {
        return detectLocalExecutionEnvironment().catch(() => null);
    }

    const now = Date.now();
    if (executionEnvironmentCache && executionEnvironmentCache.expiresAt > now) {
        return executionEnvironmentCache.value;
    }

    if (executionEnvironmentInFlight) {
        return executionEnvironmentInFlight;
    }

    executionEnvironmentInFlight = detectLocalExecutionEnvironment()
        .then((environment) => {
            executionEnvironmentCache = {
                value: environment,
                expiresAt: Date.now() + EXECUTION_ENV_CACHE_TTL_MS,
            };
            return environment;
        })
        .catch(() => {
            executionEnvironmentCache = {
                value: null,
                expiresAt: Date.now() + EXECUTION_ENV_CACHE_TTL_MS,
            };
            return null;
        })
        .finally(() => {
            executionEnvironmentInFlight = null;
        });

    return executionEnvironmentInFlight;
}

export const systemProcedures = {
    health: publicProcedure.query(() => {
        return { status: 'running', service: '@tormentnexus/core' };
    }),
    startupStatus: publicProcedure.query(async () => {
        const mcpServer = getMcpServer();
        const aggregator = getMcpAggregator();
        const agentMemory = getAgentMemoryService();
        const browserService = getBrowserService();
        const sessionSupervisor = getSessionSupervisor();
        const mcpConfigService = getMcpConfigService();

        const memoryManager = (mcpServer as { memoryManager?: { getPipelineSummary?: () => MemoryPipelineSummary } }).memoryManager;
        const memoryPipelineSummary: MemoryPipelineSummary | null = memoryManager?.getPipelineSummary?.() ?? null;

        const [runtimeServers, sessionCount, browserStatus, persistedServers, persistedTools, executionEnvironment, cachedInventory, tormentnexusStoreStatus] = await Promise.all([
            aggregator?.listServers?.().catch(() => []) ?? [],
            Promise.resolve(sessionSupervisor?.listSessions?.().length ?? 0),
            Promise.resolve(browserService?.getStatus?.() ?? { active: false, pageCount: 0, pageIds: [] }),
            mcpServersRepository.findAll().catch(() => []),
            toolsRepository.findAll().catch(() => []),
            getCachedExecutionEnvironment(),
            getCachedToolInventory().catch(() => ({ servers: [], tools: [], toolCounts: new Map(), source: 'empty' as const, snapshotUpdatedAt: null })),
            readTormentNexusStoreStatus(process.cwd(), memoryPipelineSummary).catch(() => null),
        ]);

        const liveServerCount = runtimeServers.filter((server) => server.status === 'connected').length;
        const residentLiveServerCount = runtimeServers.filter(
            (server) => server.status === 'connected' && Boolean(server.advertisedAlwaysOn),
        ).length;
        const warmingServerCount = runtimeServers.filter((server) => server.warmupStatus === 'scheduled' || server.warmupStatus === 'warming').length;
        const failedWarmupServerCount = runtimeServers.filter((server) => server.warmupStatus === 'failed').length;

        // @ts-expect-error -- type mismatch from inferred schema types
        const cachedInventorySummary = summarizeCachedInventory(cachedInventory);

        const persistedServerCount = cachedInventorySummary.source === 'empty'
            ? persistedServers.length
            : cachedInventorySummary.serverCount;
        const alwaysOnServerUuids = new Set(
            persistedServers
                .filter((server) => Boolean(server.always_on))
                .map((server) => server.uuid),
        );
        const persistedToolCount = cachedInventorySummary.source === 'empty'
            ? persistedTools.length
            : cachedInventorySummary.toolCount;
        const persistedAlwaysOnServerCount = cachedInventorySummary.source === 'empty'
            ? alwaysOnServerUuids.size
            : cachedInventorySummary.alwaysOnServerCount;
        const persistedAlwaysOnToolCount = cachedInventorySummary.source === 'empty'
            ? persistedTools.filter((tool) => Boolean(tool.always_on) || alwaysOnServerUuids.has(tool.mcp_server_uuid)).length
            : cachedInventorySummary.alwaysOnToolCount;

        return buildStartupStatusSnapshot({
            mcpServer,
            aggregator,
            agentMemory,
            browserService,
            browserStatus,
            sessionSupervisor,
            sessionCount,
            mcpConfigService,
            liveServerCount,
            residentLiveServerCount,
            warmingServerCount,
            failedWarmupServerCount,
            persistedServerCount,
            persistedToolCount,
            persistedAlwaysOnServerCount,
            persistedAlwaysOnToolCount,
    lazyMode: mcpServerPool.getLifecycleModes().lazySessionMode,
            inventorySource: cachedInventorySummary.source,
            inventorySnapshotUpdatedAt: cachedInventorySummary.snapshotUpdatedAt,
            executionEnvironment: executionEnvironment?.summary ?? null,
            tormentnexus: tormentnexusStoreStatus
                ? {
                    enabled: Boolean(tormentnexusStoreStatus.runtimePipeline.tormentnexusEnabled),
                    storePath: tormentnexusStoreStatus.storePath,
                    storeExists: tormentnexusStoreStatus.exists,
                    totalEntries: tormentnexusStoreStatus.totalEntries,
                    sectionCount: tormentnexusStoreStatus.sectionCount,
                    defaultSectionCount: tormentnexusStoreStatus.defaultSectionCount,
                    presentDefaultSectionCount: tormentnexusStoreStatus.presentDefaultSectionCount,
                    missingSections: tormentnexusStoreStatus.missingSections,
                    lastUpdatedAt: tormentnexusStoreStatus.lastUpdatedAt,
                }
                : null,
        });
    }),
    getTaskStatus: publicProcedure
        .input(z.object({ taskId: z.string().optional() }))
        .query(() => {
            const tracker = getProjectTracker();
            if (!tracker) {
                return { taskId: 'offline', status: 'offline', progress: 0, currentTask: 'Offline' };
            }

            const status = tracker.getStatus();
            return {
                taskId: status.currentTask,
                currentTask: status.currentTask,
                status: status.status,
                progress: status.progress
            };
        }),
    indexingStatus: t.procedure.query(() => {
        const lspService = getLspService();
        if (!lspService) return { status: 'offline', filesIndexed: 0, totalFiles: 0 };
        return lspService.getStatus();
    }),
    executeTool: adminProcedure.input(z.object({
        name: z.string(),
        args: z.record(z.unknown())
    })).mutation(async ({ input }) => {
        const result = await getMcpServer().executeTool(input.name, input.args);
        if (result.isError) throw new Error(result.content[0].text);
        return result.content[0].text;
    }),
};
