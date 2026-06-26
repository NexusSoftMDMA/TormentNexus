import { z } from 'zod';
import fs from 'node:fs/promises';
import { TRPCError } from '@trpc/server';
import { formatOptionalSqliteFailure } from '../db/sqliteAvailability.js';
import { t, publicProcedure, adminProcedure, getMcpAggregator, getMcpServer } from '../lib/trpc-core.js';
import { getCachedToolInventory } from '../mcp/cachedToolInventory.js';
import { parseNamespacedToolName } from '../mcp/namespaces.js';
import { evaluateAutoLoadCandidate, rankToolSearchCandidates, type ToolSearchProfile } from '../mcp/toolSearchRanking.js';
import { toolSelectionTelemetry } from '../mcp/toolSelectionTelemetry.js';
import { getTormentNexusMcpJsoncPath, loadTormentNexusMcpConfig, stripJsonComments, writeTormentNexusMcpConfig } from '../mcp/mcpJsonConfig.js';
import {
    buildToolPreferenceSettings,
    mergeToolPreferences,
    readToolPreferencesFromSettings,
    type ToolPreferences,
} from './mcp-tool-preferences.js';

function handleTrpcError(action: string, error: unknown): never {
    throw new TRPCError({
        code: 'INTERNAL_SERVER_ERROR',
        message: formatOptionalSqliteFailure(action, error),
    });
}


type McpToolCallResult = {
    content?: Array<{ type?: string; text?: string }>;
};

type ServerProbeTrafficEvent = {
    server: string;
    method: 'tools/list' | 'tools/call';
    paramsSummary: string;
    latencyMs: number;
    success: boolean;
    timestamp: number;
    toolName?: string;
    error?: string;
};

function getToolTextContent(result: unknown): string {
    const toolResult = result as McpToolCallResult | null | undefined;
    return toolResult?.content?.find((item) => item?.type === 'text')?.text ?? '';
}

function parseToolJson<T>(result: unknown, fallback: T): T {
    const text = getToolTextContent(result);
    if (!text) {
        return fallback;
    }

    try {
        return JSON.parse(text) as T;
    } catch {
        return fallback;
    }
}

function parseEvictedToolsFromMessage(message: string): string[] {
    const match = message.match(/Evicted idle tools:\s*(.+?)\.?$/i);
    if (!match?.[1]) {
        return [];
    }

    return match[1]
        .split(',')
        .map((tool) => tool.trim())
        .filter(Boolean);
}

function toSerializablePayload(value: unknown): unknown {
    try {
        return JSON.parse(JSON.stringify(value, (_key, current) => {
            if (typeof current === 'bigint') {
                return current.toString();
            }

            if (current instanceof Error) {
                return {
                    name: current.name,
                    message: current.message,
                    stack: current.stack,
                };
            }

            if (typeof current === 'undefined') {
                return null;
            }

            return current;
        }));
    } catch {
        return String(value);
    }
}

function summarizeProbePayload(value: unknown): string {
    const textContent = getToolTextContent(value);
    if (textContent) {
        return textContent.length > 180 ? `${textContent.slice(0, 177)}...` : textContent;
    }

    if (Array.isArray(value)) {
        return `Array(${value.length})`;
    }

    if (value && typeof value === 'object') {
        const record = value as Record<string, unknown>;
        if (Array.isArray(record.tools)) {
            return `Returned ${record.tools.length} tool${record.tools.length === 1 ? '' : 's'}`;
        }

        const keys = Object.keys(record);
        return keys.length > 0
            ? `Object with keys: ${keys.slice(0, 6).join(', ')}`
            : 'Empty object';
    }

    if (typeof value === 'string') {
        return value.length > 180 ? `${value.slice(0, 177)}...` : value;
    }

    if (value === null || typeof value === 'undefined') {
        return 'No payload returned';
    }

    return String(value);
}

function collectServerProbeTrafficEvents(options: {
    startedAt: number;
    targetKind: 'router' | 'server';
    operation: 'tools/list' | 'tools/call';
    serverName?: string;
    toolName?: string;
}): ServerProbeTrafficEvent[] {
    const aggregator = getMcpAggregator() as { getTrafficEvents?: () => ServerProbeTrafficEvent[] } | null;
    const events = aggregator?.getTrafficEvents?.() ?? [];
    const normalizedToolName = options.toolName
        ? (parseNamespacedToolName(options.toolName)?.toolName ?? options.toolName)
        : undefined;

    return events.filter((event) => {
        if (event.timestamp < options.startedAt) {
            return false;
        }

        if (event.method !== options.operation) {
            return false;
        }

        if (options.targetKind === 'server' && options.serverName && event.server !== options.serverName) {
            return false;
        }

        if (normalizedToolName && event.toolName && event.toolName !== normalizedToolName && event.toolName !== options.toolName) {
            return false;
        }

        return true;
    });
}

function recordAutoLoadTelemetry(toolName: string, message: string): void {
    toolSelectionTelemetry.record({
        type: 'load',
        toolName,
        status: 'success',
        message,
        evictedTools: parseEvictedToolsFromMessage(message),
    });
}

function toLatencyMs(startedAt: number): number {
    return Math.max(0, Date.now() - startedAt);
}

const toolSearchProfileSchema = z.enum(['web-research', 'repo-coding', 'browser-automation', 'local-ops', 'database']);

async function readToolPreferences(): Promise<ToolPreferences> {
    try {
        const config = await loadTormentNexusMcpConfig();
        const settings = config.settings as { toolSelection?: { importantTools?: unknown; alwaysLoadedTools?: unknown; autoLoadMinConfidence?: unknown; maxLoadedTools?: unknown; maxHydratedSchemas?: unknown } } | undefined;
        return readToolPreferencesFromSettings(settings?.toolSelection);
    } catch {
        return { importantTools: [], alwaysLoadedTools: [], autoLoadMinConfidence: 0.85, maxLoadedTools: 16, maxHydratedSchemas: 8 };
    }
}

async function writeToolPreferences(nextPreferences: ToolPreferences): Promise<ToolPreferences> {
    const config = await loadTormentNexusMcpConfig();
    const existingSettings = config.settings && typeof config.settings === 'object'
        ? config.settings as Record<string, unknown>
        : {};

    const normalized = readToolPreferencesFromSettings(nextPreferences);
    const nextSettings = buildToolPreferenceSettings(existingSettings, normalized);

    await writeTormentNexusMcpConfig({
        ...config,
        settings: nextSettings,
    });

    return normalized;
}

async function ensureAlwaysLoadedTools(server: NonNullable<ReturnType<typeof getMcpServer>>, preferences: ToolPreferences): Promise<void> {
    if (typeof server.setAlwaysLoadedTools === 'function') {
        server.setAlwaysLoadedTools(preferences.alwaysLoadedTools);
    }

    if (preferences.alwaysLoadedTools.length === 0) {
        return;
    }

    const workingSet = parseToolJson<{ tools?: Array<{ name?: string }> }>(await server.executeTool('list_loaded_tools', {}), { tools: [] });
    const loadedTools = new Set((workingSet.tools ?? [])
        .map((tool) => typeof tool?.name === 'string' ? tool.name : '')
        .filter(Boolean));

    for (const toolName of preferences.alwaysLoadedTools) {
        if (loadedTools.has(toolName)) {
            continue;
        }

        try {
            await server.executeTool('load_tool', { name: toolName });
            loadedTools.add(toolName);
        } catch {
            // Ignore load failures here so the dashboard can still render the rest of the state.
        }
    }
}

export const mcpRouter = t.router({
    /** List all registered downstream MCP servers */
    listServers: publicProcedure.query(async () => {
        const aggregator = getMcpAggregator();
        try {
            const { servers, toolCounts } = await getCachedToolInventory();
            const runtimeStates = new Map((await aggregator?.listServers?.() ?? []).map((server) => [server.name, server]));

            return servers.map((server) => {
                const runtime = runtimeStates.get(server.name);
                const cachedToolCount = toolCounts.get(server.uuid) ?? 0;
                return {
                    name: server.name ?? 'unknown',
                    displayName: server.displayName ?? server.name ?? 'unknown',
                    tags: server.tags ?? [],
                    status: runtime?.status ?? (server.error_status === 'NONE' ? 'cached' : server.error_status.toLowerCase()),
                    runtimeState: runtime?.status ?? 'stopped',
                    warmupState: runtime?.warmupStatus ?? (server.alwaysOnAdvertised ? 'scheduled' : 'idle'),
                    runtimeConnected: runtime?.status === 'connected',
                    toolCount: runtime?.toolCount ?? cachedToolCount,
                    advertisedToolCount: runtime?.advertisedToolCount ?? cachedToolCount,
                    advertisedSource: runtime?.advertisedSource ?? 'empty',
                    lastConnectedAt: runtime?.lastConnectedAt ? new Date(runtime.lastConnectedAt).toISOString() : null,
                    lastError: runtime?.lastError ?? null,
                    alwaysOn: Boolean(server.alwaysOnAdvertised),
                    config: {
                        command: runtime?.config?.command ?? server.command ?? '',
                        args: runtime?.config?.args ?? server.args ?? [],
                        env: runtime?.config?.env
                            ? Object.keys(runtime.config.env)
                            : Object.keys(server.env ?? {}),
                    },
                };
            });
        } catch (error) {
            handleTrpcError('MCP inventory is unavailable', error);
        }
    }),

    /** List all aggregated tools across servers */
    listTools: publicProcedure.query(async () => {
        try {
            const { tools } = await getCachedToolInventory();
            if (tools.length > 0) {
                return tools;
            }

            const aggregator = getMcpAggregator();
            const liveTools = await aggregator?.listAggregatedTools?.() ?? [];
            return liveTools.map((tool) => ({
                name: tool.name,
                description: tool.description ?? '',
                server: tool.server ?? 'unknown',
                serverDisplayName: tool.serverDisplayName ?? tool.server ?? 'unknown',
                serverTags: tool.serverTags ?? [],
                toolTags: tool.toolTags ?? [],
                semanticGroup: tool.semanticGroup ?? 'general-utility',
                semanticGroupLabel: tool.semanticGroupLabel ?? 'general utility',
                advertisedName: tool.advertisedName ?? tool.name,
                keywords: tool.keywords ?? [],
                alwaysOn: Boolean(tool.alwaysOn),
                inputSchema: tool.inputSchema ?? null,
            }));
        } catch (error) {
            handleTrpcError('MCP inventory is unavailable', error);
        }
    }),

    traffic: publicProcedure.query(async () => {
        const aggregator = getMcpAggregator();
        try {
            return await aggregator?.getTrafficEvents?.() ?? [];
        } catch (error) {
            handleTrpcError('MCP traffic is unavailable', error);
        }
    }),

    searchTools: publicProcedure.input(z.object({
        query: z.string().default(''),
        profile: toolSearchProfileSchema.optional(),
    })).query(async ({ input }) => {
        const searchStartedAt = Date.now();
        try {
            const { tools } = await getCachedToolInventory();
            const preferences = await readToolPreferences();
            const normalizedQuery = input.query.trim().toLowerCase();
            const selectedProfile = input.profile as ToolSearchProfile | undefined;
            const server = getMcpServer();

            if (server) {
                await ensureAlwaysLoadedTools(server, preferences);
                try {
                    const result = await server.executeTool('search_tools', { query: input.query });
                    const rankedResults = parseToolJson<Array<{
                        name: string;
                        description: string;
                        serverName?: string;
                        serverDisplayName?: string;
                        originalName?: string;
                        advertisedName?: string;
                        serverTags?: string[];
                        toolTags?: string[];
                        semanticGroup?: string;
                        semanticGroupLabel?: string;
                        keywords?: string[];
                        alwaysOn?: boolean;
                        loaded?: boolean;
                        hydrated?: boolean;
                        deferred?: boolean;
                        requiresSchemaHydration?: boolean;
                        matchReason?: string;
                        score?: number;
                        rank?: number;
                        autoLoaded?: boolean;
                        mcpServerUuid?: string;
                    }>>(result, []);

                    if (rankedResults.length > 0) {
                        const runtimeMappedResults = rankedResults.map((tool, index) => ({
                            name: tool.name,
                            description: tool.description ?? '',
                            server: tool.serverName ?? 'unknown',
                            serverDisplayName: tool.serverDisplayName ?? tool.serverName ?? 'unknown',
                            serverTags: tool.serverTags ?? [],
                            toolTags: tool.toolTags ?? [],
                            semanticGroup: tool.semanticGroup ?? 'general-utility',
                            semanticGroupLabel: tool.semanticGroupLabel ?? 'general utility',
                            advertisedName: tool.advertisedName ?? tool.name,
                            keywords: tool.keywords ?? [],
                            alwaysOn: Boolean(tool.alwaysOn),
                            originalName: tool.originalName ?? null,
                            loaded: Boolean(tool.loaded),
                            hydrated: Boolean(tool.hydrated),
                            deferred: Boolean(tool.deferred),
                            requiresSchemaHydration: Boolean(tool.requiresSchemaHydration),
                            matchReason: tool.matchReason ?? 'matched available metadata',
                            score: tool.score ?? 0,
                            rank: tool.rank ?? (index + 1),
                            autoLoaded: Boolean(tool.autoLoaded),
                            inputSchema: null,
                        }));

                        const mappedResults = selectedProfile
                            ? rankToolSearchCandidates(
                                runtimeMappedResults.map((tool) => ({
                                    name: tool.name,
                                    description: tool.description,
                                    serverName: tool.server,
                                    serverDisplayName: tool.serverDisplayName,
                                    advertisedName: tool.advertisedName,
                                    originalName: tool.originalName ?? undefined,
                                    serverTags: tool.serverTags,
                                    toolTags: tool.toolTags,
                                    semanticGroup: tool.semanticGroup,
                                    semanticGroupLabel: tool.semanticGroupLabel,
                                    keywords: tool.keywords,
                                    alwaysOn: tool.alwaysOn,
                                    loaded: tool.loaded,
                                    hydrated: tool.hydrated,
                                    deferred: tool.deferred,
                                })),
                                normalizedQuery,
                                runtimeMappedResults.length,
                                selectedProfile,
                            ).map((tool, index) => ({
                                name: tool.name,
                                description: tool.description ?? '',
                                server: tool.serverName ?? 'unknown',
                                serverDisplayName: tool.serverDisplayName ?? tool.serverName ?? 'unknown',
                                serverTags: tool.serverTags ?? [],
                                toolTags: tool.toolTags ?? [],
                                semanticGroup: tool.semanticGroup ?? 'general-utility',
                                semanticGroupLabel: tool.semanticGroupLabel ?? 'general utility',
                                advertisedName: tool.advertisedName ?? tool.name,
                                keywords: tool.keywords ?? [],
                                alwaysOn: Boolean(tool.alwaysOn),
                                originalName: tool.originalName ?? null,
                                loaded: Boolean(tool.loaded),
                                hydrated: Boolean(tool.hydrated),
                                deferred: Boolean(tool.deferred),
                                requiresSchemaHydration: Boolean(tool.requiresSchemaHydration),
                                matchReason: tool.matchReason ?? 'matched available metadata',
                                score: tool.score ?? 0,
                                rank: index + 1,
                                autoLoaded: false,
                                inputSchema: null,
                            }))
                            : runtimeMappedResults;

                        const autoLoadedResult = mappedResults.find((tool) => tool.autoLoaded);
                        if (autoLoadedResult) {
                            recordAutoLoadTelemetry(autoLoadedResult.name, `Tool '${autoLoadedResult.name}' auto-loaded from search.`);
                        }

                        const topScore = typeof mappedResults[0]?.score === 'number' ? mappedResults[0].score : undefined;
                        const secondScore = typeof mappedResults[1]?.score === 'number' ? mappedResults[1].score : 0;
                        const scoreGap = typeof topScore === 'number' ? topScore - secondScore : undefined;

                        toolSelectionTelemetry.record({
                            type: 'search',
                            query: input.query,
                            profile: selectedProfile,
                            source: 'runtime-search',
                            resultCount: mappedResults.length,
                            topResultName: mappedResults[0]?.name,
                            topMatchReason: mappedResults[0]?.matchReason,
                            topScore,
                            secondResultName: mappedResults[1]?.name,
                            secondMatchReason: mappedResults[1]?.matchReason,
                            secondScore,
                            scoreGap,
                            latencyMs: toLatencyMs(searchStartedAt),
                            status: 'success',
                        });

                        return mergeToolPreferences(mappedResults, preferences, tools);
                    }
                } catch {
                    // Fall back to cached ranking when the runtime search tool is unavailable.
                }
            }

            if (tools.length > 0) {
                const rankedCandidateResults = rankToolSearchCandidates(
                    tools.map((tool) => ({
                        name: tool.name,
                        description: tool.description ?? '',
                        serverName: tool.server,
                        serverDisplayName: tool.serverDisplayName,
                        advertisedName: tool.advertisedName,
                        originalName: tool.originalName ?? (tool.name.includes('__') ? tool.name.split('__').slice(1).join('__') : undefined),
                        serverTags: tool.serverTags,
                        toolTags: tool.toolTags,
                        semanticGroup: tool.semanticGroup,
                        semanticGroupLabel: tool.semanticGroupLabel,
                        keywords: tool.keywords,
                        alwaysOn: tool.alwaysOn,
                        deferred: tool.inputSchema == null,
                        hydrated: tool.inputSchema != null,
                    })),
                    normalizedQuery,
                    normalizedQuery ? 10 : tools.length,
                    selectedProfile,
                );

                const autoLoadEvaluation = server
                    ? evaluateAutoLoadCandidate(rankedCandidateResults, normalizedQuery, {
                        minConfidence: preferences.autoLoadMinConfidence,
                    })
                    : {
                        evaluated: false,
                        outcome: 'not-applicable' as const,
                        decision: null,
                        skipReason: 'runtime MCP server unavailable',
                        minConfidence: preferences.autoLoadMinConfidence,
                    };
                const autoLoadDecision = autoLoadEvaluation.decision;
                let autoLoadExecutionStatus: 'success' | 'error' | 'not-attempted' = 'not-attempted';
                let autoLoadExecutionError: string | undefined;

                if (server && autoLoadDecision) {
                    const autoLoadStartedAt = Date.now();
                    try {
                        const loadResult = await server.executeTool('load_tool', { name: autoLoadDecision.toolName });
                        autoLoadExecutionStatus = 'success';
                        toolSelectionTelemetry.record({
                            type: 'load',
                            toolName: autoLoadDecision.toolName,
                            status: 'success',
                            message: getToolTextContent(loadResult),
                            evictedTools: parseEvictedToolsFromMessage(getToolTextContent(loadResult)),
                            latencyMs: toLatencyMs(autoLoadStartedAt),
                            autoLoadReason: autoLoadDecision.reason,
                            autoLoadConfidence: autoLoadDecision.confidence,
                            scoreGap: autoLoadDecision.scoreGap,
                            topScore: autoLoadDecision.topScore,
                        });
                    } catch (error) {
                        autoLoadExecutionStatus = 'error';
                        autoLoadExecutionError = error instanceof Error ? error.message : String(error);
                        toolSelectionTelemetry.record({
                            type: 'load',
                            toolName: autoLoadDecision.toolName,
                            status: 'error',
                            message: autoLoadExecutionError,
                            latencyMs: toLatencyMs(autoLoadStartedAt),
                            autoLoadReason: autoLoadDecision.reason,
                            autoLoadConfidence: autoLoadDecision.confidence,
                            scoreGap: autoLoadDecision.scoreGap,
                            topScore: autoLoadDecision.topScore,
                        });
                    }
                }

                const rankedResults = rankedCandidateResults.map((tool, index) => ({
                    name: tool.name,
                    description: tool.description ?? '',
                    server: tool.serverName ?? 'unknown',
                    serverDisplayName: tool.serverDisplayName ?? tool.serverName ?? 'unknown',
                    serverTags: tool.serverTags ?? [],
                    toolTags: tool.toolTags ?? [],
                    semanticGroup: tool.semanticGroup ?? 'general-utility',
                    semanticGroupLabel: tool.semanticGroupLabel ?? 'general utility',
                    advertisedName: tool.advertisedName ?? tool.name,
                    keywords: tool.keywords ?? [],
                    alwaysOn: Boolean(tool.alwaysOn),
                    originalName: tool.originalName ?? null,
                    loaded: tool.loaded || tool.name === autoLoadDecision?.toolName,
                    hydrated: tool.hydrated,
                    deferred: tool.deferred,
                    requiresSchemaHydration: tool.requiresSchemaHydration,
                    matchReason: tool.name === autoLoadDecision?.toolName
                        ? `${tool.matchReason}; ${autoLoadDecision.reason}`
                        : tool.matchReason,
                    score: tool.score,
                    rank: index + 1,
                    autoLoaded: tool.name === autoLoadDecision?.toolName,
                    inputSchema: null,
                }));

                toolSelectionTelemetry.record({
                    type: 'search',
                    query: input.query,
                    profile: selectedProfile,
                    source: 'cached-ranking',
                    resultCount: rankedResults.length,
                    topResultName: rankedResults[0]?.name,
                    topMatchReason: rankedResults[0]?.matchReason,
                    topScore: rankedResults[0]?.score,
                    secondResultName: rankedResults[1]?.name,
                    secondMatchReason: rankedResults[1]?.matchReason,
                    secondScore: rankedResults[1]?.score,
                    scoreGap: rankedResults.length > 1 ? (rankedResults[0]?.score ?? 0) - (rankedResults[1]?.score ?? 0) : undefined,
                    latencyMs: toLatencyMs(searchStartedAt),
                    autoLoadReason: autoLoadDecision?.reason,
                    autoLoadConfidence: autoLoadDecision?.confidence,
                    autoLoadEvaluated: autoLoadEvaluation.evaluated,
                    autoLoadOutcome: autoLoadEvaluation.outcome,
                    autoLoadSkipReason: autoLoadEvaluation.skipReason,
                    autoLoadMinConfidence: autoLoadEvaluation.minConfidence,
                    autoLoadExecutionStatus,
                    autoLoadExecutionError,
                    status: 'success',
                });

                return mergeToolPreferences(rankedResults, preferences, tools);
            }

            const aggregator = getMcpAggregator();
            const liveTools = await aggregator?.searchTools?.(input.query) ?? [];
            const rankedLiveTools = rankToolSearchCandidates(
                liveTools.map((tool) => ({
                    name: tool.name,
                    description: tool.description ?? '',
                    serverName: tool.server ?? 'unknown',
                    serverDisplayName: tool.serverDisplayName ?? tool.server ?? 'unknown',
                    serverTags: tool.serverTags ?? [],
                    toolTags: tool.toolTags ?? [],
                    semanticGroup: tool.semanticGroup ?? 'general-utility',
                    semanticGroupLabel: tool.semanticGroupLabel ?? 'general utility',
                    advertisedName: tool.advertisedName ?? tool.name,
                    keywords: tool.keywords ?? [],
                    alwaysOn: Boolean(tool.alwaysOn),
                    originalName: null,
                    loaded: false,
                    hydrated: tool.inputSchema != null,
                    deferred: tool.inputSchema == null,
                })),
                normalizedQuery,
                normalizedQuery ? 10 : liveTools.length,
                selectedProfile,
            );
            const mappedLiveTools = rankedLiveTools.map((tool, index) => ({
                name: tool.name,
                description: tool.description ?? '',
                server: tool.serverName ?? 'unknown',
                serverDisplayName: tool.serverDisplayName ?? tool.serverName ?? 'unknown',
                serverTags: tool.serverTags ?? [],
                toolTags: tool.toolTags ?? [],
                semanticGroup: tool.semanticGroup ?? 'general-utility',
                semanticGroupLabel: tool.semanticGroupLabel ?? 'general utility',
                advertisedName: tool.advertisedName ?? tool.name,
                keywords: tool.keywords ?? [],
                alwaysOn: Boolean(tool.alwaysOn),
                originalName: null,
                loaded: false,
                hydrated: tool.hydrated,
                deferred: tool.deferred,
                requiresSchemaHydration: tool.requiresSchemaHydration,
                matchReason: tool.matchReason,
                score: tool.score,
                rank: index + 1,
                inputSchema: null,
            }));

            toolSelectionTelemetry.record({
                type: 'search',
                query: input.query,
                profile: selectedProfile,
                source: 'live-aggregator',
                resultCount: mappedLiveTools.length,
                topResultName: mappedLiveTools[0]?.name,
                topMatchReason: mappedLiveTools[0]?.matchReason,
                topScore: mappedLiveTools[0]?.score,
                secondResultName: mappedLiveTools[1]?.name,
                secondMatchReason: mappedLiveTools[1]?.matchReason,
                secondScore: mappedLiveTools[1]?.score,
                scoreGap: mappedLiveTools.length > 1 ? (mappedLiveTools[0]?.score ?? 0) - (mappedLiveTools[1]?.score ?? 0) : undefined,
                latencyMs: toLatencyMs(searchStartedAt),
                status: 'success',
            });

            return mergeToolPreferences(mappedLiveTools, preferences, tools);
        } catch (error) {
            toolSelectionTelemetry.record({
                type: 'search',
                query: input.query,
                profile: input.profile,
                latencyMs: toLatencyMs(searchStartedAt),
                status: 'error',
                message: 'search failed',
            });
            handleTrpcError('MCP tool search is unavailable', error);
        }
    }),

    getToolSelectionTelemetry: publicProcedure.query(() => toolSelectionTelemetry.list()),

    getToolPreferences: publicProcedure.query(async () => readToolPreferences()),

    setToolPreferences: adminProcedure.input(z.object({
        importantTools: z.array(z.string().min(1)).default([]),
        alwaysLoadedTools: z.array(z.string().min(1)).default([]),
        autoLoadMinConfidence: z.number().min(0.5).max(0.99).default(0.85),
        maxLoadedTools: z.number().int().min(4).max(64).default(16),
        maxHydratedSchemas: z.number().int().min(2).max(32).default(8),
    })).mutation(async ({ input }) => {
        const next = await writeToolPreferences({
            importantTools: input.importantTools,
            alwaysLoadedTools: input.alwaysLoadedTools,
            autoLoadMinConfidence: input.autoLoadMinConfidence,
            maxLoadedTools: input.maxLoadedTools,
            maxHydratedSchemas: input.maxHydratedSchemas,
        });
        const server = getMcpServer();
        if (server) {
            await ensureAlwaysLoadedTools(server, next);
            // Apply the new capacity limits to the live working set immediately.
            try {
                await server.executeTool('set_capacity', {
                    maxLoadedTools: next.maxLoadedTools,
                    maxHydratedSchemas: next.maxHydratedSchemas,
                });
            } catch {
                // Non-fatal: the working set will apply limits on the next load/hydrate call.
            }
        }
        return { ok: true, ...next };
    }),

    /** Return the bounded recent eviction history for the session working set. */
    getWorkingSetEvictionHistory: publicProcedure.query(async () => {
        const server = getMcpServer();
        if (!server) {
            return [];
        }

        try {
            const result = await server.executeTool('get_eviction_history', {});
            return parseToolJson<Array<{ toolName: string; timestamp: number; tier: string }>>(result, []);
        } catch (error) {
            handleTrpcError('MCP eviction history is unavailable', error);
        }
    }),

    /** Clear the bounded recent eviction history for the current session working set. */
    clearWorkingSetEvictionHistory: adminProcedure.mutation(async () => {
        const server = getMcpServer();
        if (!server) {
            return { ok: true, message: 'MCP server unavailable; eviction history already empty.' };
        }

        try {
            const result = await server.executeTool('clear_eviction_history', {});
            return {
                ok: true,
                message: getToolTextContent(result) || 'Eviction history cleared.',
            };
        } catch (error) {
            handleTrpcError('Failed to clear eviction history', error);
        }
    }),

    getJsoncEditor: publicProcedure.query(async () => {
        const jsoncPath = getTormentNexusMcpJsoncPath();
        try {
            const content = await fs.readFile(jsoncPath, 'utf-8');
            return { path: jsoncPath, content };
        } catch (error) {
            const errorCode = (error as NodeJS.ErrnoException).code;
            if (errorCode === 'ENOENT') {
                const fallbackConfig = await loadTormentNexusMcpConfig();
                return {
                    path: jsoncPath,
                    content: `// TormentNexus MCP configuration\n${JSON.stringify(fallbackConfig, null, 2)}\n`,
                };
            }
            throw error;
        }
    }),

    saveJsoncEditor: adminProcedure.input(z.object({
        content: z.string().min(2),
    })).mutation(async ({ input }) => {
        const parsed = JSON.parse(stripJsonComments(input.content)) as Record<string, unknown>;
        await writeTormentNexusMcpConfig(parsed as never);
        return { ok: true };
    }),

    callTool: publicProcedure.input(z.object({
        name: z.string().min(1),
        args: z.record(z.unknown()).optional().default({}),
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        if (!server) {
            throw new Error('MCP Server not initialized');
        }

        const result = await server.executeTool(input.name, input.args);
        return {
            ok: true,
            result,
        };
    }),

    runServerTest: publicProcedure.input(z.object({
        targetKind: z.enum(['router', 'server']),
        serverName: z.string().optional(),
        operation: z.enum(['tools/list', 'tools/call']),
        toolName: z.string().optional(),
        args: z.record(z.unknown()).optional().default({}),
    })).mutation(async ({ input }) => {
        const startedAt = Date.now();
        const requestId = `probe-${startedAt}`;
        const normalizedToolName = input.toolName
            ? (parseNamespacedToolName(input.toolName)?.toolName ?? input.toolName)
            : undefined;

        const requestPayload = {
            jsonrpc: '2.0',
            id: requestId,
            method: input.operation,
            params: input.operation === 'tools/call'
                ? {
                    name: input.targetKind === 'server' ? normalizedToolName : input.toolName,
                    arguments: input.args,
                }
                : {},
            target: input.targetKind === 'router'
                ? 'tormentnexus-router'
                : input.serverName ?? 'unknown-server',
            via: input.targetKind === 'router' ? 'tormentnexus-router' : 'direct-downstream',
        };

        if (input.targetKind === 'server' && !input.serverName) {
            return {
                success: false,
                target: {
                    kind: 'server' as const,
                    displayName: 'Unknown downstream server',
                    serverName: null,
                    via: 'direct-downstream',
                },
                operation: input.operation,
                startedAt,
                endedAt: Date.now(),
                latencyMs: Date.now() - startedAt,
                request: requestPayload,
                response: {
                    summary: 'Downstream probe requires a server name.',
                    payload: { error: 'Downstream probe requires a server name.' },
                },
                trafficEvents: [],
            };
        }

        if (input.operation === 'tools/call' && !input.toolName) {
            return {
                success: false,
                target: {
                    kind: input.targetKind,
                    displayName: input.targetKind === 'router' ? 'TormentNexus router' : input.serverName ?? 'Unknown downstream server',
                    serverName: input.serverName ?? null,
                    via: input.targetKind === 'router' ? 'tormentnexus-router' : 'direct-downstream',
                },
                operation: input.operation,
                startedAt,
                endedAt: Date.now(),
                latencyMs: Date.now() - startedAt,
                request: requestPayload,
                response: {
                    summary: 'Tool call probe requires a tool name.',
                    payload: { error: 'Tool call probe requires a tool name.' },
                },
                trafficEvents: [],
            };
        }

        try {
            let responsePayload: unknown;

            if (input.targetKind === 'router') {
                const aggregator = getMcpAggregator() as {
                    listAggregatedTools?: () => Promise<unknown[]>;
                    executeTool?: (name: string, args: unknown) => Promise<unknown>;
                } | null;

                if (!aggregator) {
                    throw new Error('TormentNexus MCP router is not initialized.');
                }

                if (input.operation === 'tools/list') {
                    const tools = await aggregator.listAggregatedTools?.() ?? [];
                    responsePayload = {
                        toolCount: tools.length,
                        tools: toSerializablePayload(tools),
                    };
                } else {
                    responsePayload = await aggregator.executeTool?.(input.toolName!, input.args);
                }
            } else {
                const aggregator = getMcpAggregator() as {
                    clients?: Map<string, {
                        listTools: () => Promise<unknown[]>;
                        callTool: (toolName: string, args: unknown) => Promise<unknown>;
                    }>;
                } | null;

                const directClient = aggregator?.clients?.get(input.serverName!);
                if (!directClient) {
                    throw new Error(`Downstream server '${input.serverName}' is not currently connected.`);
                }

                if (input.operation === 'tools/list') {
                    const tools = await directClient.listTools();
                    responsePayload = {
                        toolCount: tools.length,
                        tools: toSerializablePayload(tools),
                    };
                } else {
                    responsePayload = await directClient.callTool(normalizedToolName!, input.args);
                }
            }

            const endedAt = Date.now();
            return {
                success: true,
                target: {
                    kind: input.targetKind,
                    displayName: input.targetKind === 'router' ? 'TormentNexus router' : input.serverName!,
                    serverName: input.serverName ?? null,
                    via: input.targetKind === 'router' ? 'tormentnexus-router' : 'direct-downstream',
                },
                operation: input.operation,
                startedAt,
                endedAt,
                latencyMs: endedAt - startedAt,
                request: requestPayload,
                response: {
                    summary: summarizeProbePayload(responsePayload),
                    payload: toSerializablePayload(responsePayload),
                },
                trafficEvents: collectServerProbeTrafficEvents({
                    startedAt,
                    targetKind: input.targetKind,
                    operation: input.operation,
                    serverName: input.serverName,
                    toolName: input.toolName,
                }),
            };
        } catch (error) {
            const endedAt = Date.now();
            const message = error instanceof Error ? error.message : String(error);
            return {
                success: false,
                target: {
                    kind: input.targetKind,
                    displayName: input.targetKind === 'router' ? 'TormentNexus router' : input.serverName ?? 'Unknown downstream server',
                    serverName: input.serverName ?? null,
                    via: input.targetKind === 'router' ? 'tormentnexus-router' : 'direct-downstream',
                },
                operation: input.operation,
                startedAt,
                endedAt,
                latencyMs: endedAt - startedAt,
                request: requestPayload,
                response: {
                    summary: message,
                    payload: { error: message },
                },
                trafficEvents: collectServerProbeTrafficEvents({
                    startedAt,
                    targetKind: input.targetKind,
                    operation: input.operation,
                    serverName: input.serverName,
                    toolName: input.toolName,
                }),
            };
        }
    }),

    clearToolSelectionTelemetry: adminProcedure.mutation(() => {
        toolSelectionTelemetry.clear();
        return { ok: true };
    }),

    getWorkingSet: publicProcedure.query(async () => {
        const server = getMcpServer();
        if (!server) {
            return {
                limits: { maxLoadedTools: 0, maxHydratedSchemas: 0 },
                tools: [],
            };
        }

        try {
            await ensureAlwaysLoadedTools(server, await readToolPreferences());
            const result = await server.executeTool('list_loaded_tools', {});
            return parseToolJson(result, {
                limits: { maxLoadedTools: 0, maxHydratedSchemas: 0 },
                tools: [],
            });
        } catch {
            return {
                limits: { maxLoadedTools: 0, maxHydratedSchemas: 0 },
                tools: [],
            };
        }
    }),

    loadTool: publicProcedure.input(z.object({
        name: z.string().min(1),
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        if (!server) {
            throw new Error('MCP Server not initialized');
        }

        const startedAt = Date.now();
        const result = await server.executeTool('load_tool', { name: input.name });
        const message = getToolTextContent(result);
        const evictedTools = parseEvictedToolsFromMessage(message);
        toolSelectionTelemetry.record({
            type: 'load',
            toolName: input.name,
            status: 'success',
            message,
            evictedTools,
            latencyMs: toLatencyMs(startedAt),
        });
        return {
            ok: true,
            message,
        };
    }),

    unloadTool: publicProcedure.input(z.object({
        name: z.string().min(1),
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        if (!server) {
            throw new Error('MCP Server not initialized');
        }

        const startedAt = Date.now();
        const result = await server.executeTool('unload_tool', { name: input.name });
        const message = getToolTextContent(result);
        toolSelectionTelemetry.record({
            type: 'unload',
            toolName: input.name,
            status: 'success',
            message,
            latencyMs: toLatencyMs(startedAt),
        });
        return {
            ok: true,
            message,
        };
    }),

    getToolSchema: publicProcedure.input(z.object({
        name: z.string().min(1),
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        if (!server) {
            throw new Error('MCP Server not initialized');
        }

        const startedAt = Date.now();
        const result = await server.executeTool('get_tool_schema', { name: input.name });
        const parsed = parseToolJson(result, {
            inputSchema: null,
            evictedHydratedTools: [],
        });
        toolSelectionTelemetry.record({
            type: 'hydrate',
            toolName: input.name,
            status: 'success',
            message: 'schema hydrated',
            evictedTools: Array.isArray(parsed.evictedHydratedTools) ? parsed.evictedHydratedTools : [],
            latencyMs: toLatencyMs(startedAt),
        });
        return parsed;
    }),

    /** Get aggregator status and stats */
    getStatus: publicProcedure.query(async () => {
        const aggregator = getMcpAggregator();

        let cachedInventory;
        try {
            cachedInventory = await getCachedToolInventory();
        } catch (error) {
            handleTrpcError('MCP inventory is unavailable', error);
        }

        let liveServers: any[] = [];
        let liveTools: any[] = [];
        try {
            [liveServers, liveTools] = await Promise.all([
                aggregator?.listServers?.() ?? Promise.resolve([]),
                aggregator?.listAggregatedTools?.() ?? Promise.resolve([]),
            ]);
        } catch (error) {
            handleTrpcError('MCP status is unavailable', error);
        }

        const effectiveServerCount = liveServers.length > 0
            ? liveServers.length
            : cachedInventory.servers.length;
        const effectiveToolCount = liveTools.length > 0
            ? liveTools.length
            : cachedInventory.tools.length;

        return {
            initialized: Boolean(aggregator),
            serverCount: effectiveServerCount,
            toolCount: effectiveToolCount,
            connectedCount: liveServers.filter((s) => s.status === 'connected').length,
        };
    }),

    /** Connect to a configured MCP server */
    connectServer: adminProcedure.input(z.object({
        name: z.string().min(1),
    })).mutation(async ({ input }) => {
        const aggregator = getMcpAggregator();
        if (!aggregator) throw new Error('MCP Aggregator not initialized');

        // Find server config from the JSONC config
        const config = await loadTormentNexusMcpConfig();
        const serverConfig = config.mcpServers?.[input.name];
        if (!serverConfig) {
            throw new Error(`Server '${input.name}' not found in configuration. Use mcp.listServers to see available servers.`);
        }

        await (aggregator as any).connectToServer(input.name, {
            command: serverConfig.command ?? 'npx',
            args: serverConfig.args ?? [],
            env: serverConfig.env ?? {},
            enabled: true,
        });
        return { success: true, name: input.name };
    }),

    /** Disconnect a running MCP server */
    disconnectServer: adminProcedure.input(z.object({
        name: z.string().min(1),
    })).mutation(async ({ input }) => {
        const aggregator = getMcpAggregator();
        if (!aggregator) throw new Error('MCP Aggregator not initialized');

        const client = (aggregator as any).clients?.get(input.name);
        if (client) {
            await client.close();
            (aggregator as any).clients?.delete(input.name);
        }
        return { success: true, name: input.name };
    }),

    /** Add a new downstream MCP server */
    addServer: adminProcedure.input(z.object({
        name: z.string().min(1),
        command: z.string().min(1),
        args: z.array(z.string()).optional().default([]),
        env: z.record(z.string()).optional().default({}),
    })).mutation(async ({ input }) => {
        const aggregator = getMcpAggregator();
        if (!aggregator) throw new Error('MCP Aggregator not initialized');
        if (!aggregator.addServerConfig) {
            throw new Error('MCP Aggregator addServerConfig unavailable');
        }

        await aggregator.addServerConfig(input.name, {
            command: input.command,
            args: input.args,
            env: input.env,
            enabled: true,
        });
        return { success: true, name: input.name };
    }),

    /** Remove a downstream MCP server */
    removeServer: adminProcedure.input(z.object({
        name: z.string(),
    })).mutation(async ({ input }) => {
        const aggregator = getMcpAggregator();
        if (!aggregator) throw new Error('MCP Aggregator not initialized');

        if (!aggregator.removeServerConfig) {
            throw new Error('MCP Aggregator removeServerConfig unavailable');
        }

        await aggregator.removeServerConfig(input.name);

        return { success: true };
    }),

    /**
     * Push a conversation turn into the ConversationalToolInjector sliding window.
     * The injector will use recent turns to predict which non-always-on tools to
     * pre-load into the session working set on the next ListTools call.
     *
     * Typical callers: dashboard chat surface, tRPC bridge clients, Go kernel proxy.
     */
    appendConversationTurn: publicProcedure.input(z.object({
        role: z.enum(['user', 'assistant', 'tool']).default('user'),
        text: z.string().min(1).max(2000),
    })).mutation(({ input }) => {
        const server = getMcpServer();
        if (!server || typeof server.appendConversationTurn !== 'function') {
            return { ok: false, reason: 'MCP server not initialized or injector unavailable' };
        }
        server.appendConversationTurn(input.role, input.text);
        return { ok: true };
    }),

    /**
     * Return the current ConversationalToolInjector sliding window snapshot.
     * Useful for dashboard debugging and the predictive tool injection panel.
     */
    getConversationWindow: publicProcedure.query(() => {
        const server = getMcpServer();
        if (!server || typeof server.getConversationInjector !== 'function') {
            return { turns: [], tokenCount: 0 };
        }
        const injector = server.getConversationInjector();
        return {
            turns: Array.from(injector.getWindow()),
            tokenCount: injector.getWindowTokenCount(),
        };
    }),
});


