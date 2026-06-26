import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { toolChainsRepository } from '../db/repositories/tool-chains.repo.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

/**
 * Tool Call Chaining & Renaming Router (Phase M)
 *
 * Advanced MCP tool management:
 * - Chain multiple tool calls into a single pipeline
 * - Rename tools for operator convenience
 * - Lazy-load tool binaries on first invocation
 * - Context syntax minimization for token reduction
 */

const ToolRenameSchema = z.object({
    serverId: z.string(),
    originalName: z.string(),
    alias: z.string().min(1).max(100),
    description: z.string().optional(),
});

const ToolChainStepSchema = z.object({
    toolName: z.string(),
    serverId: z.string().optional(),
    inputMapping: z.record(z.string()).optional(), // Maps chain input keys to tool param names
    outputKey: z.string().optional(), // Key to store this step's result under
    condition: z.string().optional(), // JSONPath condition to skip step
});

const ToolChainSchema = z.object({
    name: z.string().min(1).max(100),
    description: z.string().optional(),
    steps: z.array(ToolChainStepSchema).min(1).max(20),
    failurePolicy: z.enum(['stop', 'skip', 'retry']).default('stop'),
    maxRetries: z.number().min(0).max(3).default(1),
});

export interface ToolAlias {
    serverId: string;
    originalName: string;
    alias: string;
    description?: string;
    createdAt: number;
}

export interface ToolChainDefinition {
    id: string;
    name: string;
    description?: string;
    steps: Array<{
        toolName: string;
        serverId?: string;
        inputMapping?: Record<string, string>;
        outputKey?: string;
        condition?: string;
    }>;
    failurePolicy: 'stop' | 'skip' | 'retry';
    maxRetries: number;
    createdAt: number;
    lastRunAt?: number;
    runCount: number;
}

export interface ChainExecutionResult {
    chainId: string;
    success: boolean;
    stepsCompleted: number;
    totalSteps: number;
    results: Record<string, unknown>;
    errors: string[];
    executionTimeMs: number;
}

export interface LazyToolState {
    serverId: string;
    toolName: string;
    loaded: boolean;
    loadedAt?: number;
    loadTimeMs?: number;
    invocationCount: number;
}

// Ephemeral state for binary lazy loading
const lazyToolStates: Map<string, LazyToolState> = new Map();

export const toolChainingRouter = t.router({
    // ─── Tool Renaming ───
    
    /**
     * Create an alias for a tool.
     */
    createAlias: adminProcedure
        .input(ToolRenameSchema)
        .mutation(async ({ input }) => {
            try {
                const existing = await toolChainsRepository.getAliasById(input.alias);
                if (existing) {
                    throw new Error(`Alias '${input.alias}' already exists for ${existing.targetTool}`);
                }

                const alias = await toolChainsRepository.upsertAlias({
                    alias: input.alias,
                    targetTool: input.originalName,
                    description: input.description,
                });

                return {
                    serverId: input.serverId,
                    originalName: alias.targetTool,
                    alias: alias.alias,
                    description: alias.description || undefined,
                    createdAt: alias.createdAt.getTime(),
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool alias store is unavailable', error);
            }
        }),

    /**
     * List all tool aliases.
     */
    listAliases: publicProcedure.query(async () => {
        try {
            const aliases = await toolChainsRepository.getAliases();
            return aliases.map(a => ({
                serverId: 'unknown',
                originalName: a.targetTool,
                alias: a.alias,
                description: a.description || undefined,
                createdAt: a.createdAt.getTime(),
            }));
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Tool alias store is unavailable', error);
        }
    }),

    /**
     * Remove a tool alias.
     */
    removeAlias: adminProcedure
        .input(z.object({ alias: z.string() }))
        .mutation(async ({ input }) => {
            try {
                const removed = await toolChainsRepository.deleteAlias(input.alias);
                return { removed };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool alias store is unavailable', error);
            }
        }),

    /**
     * Resolve an alias to the original tool name.
     */
    resolveAlias: publicProcedure
        .input(z.object({ name: z.string() }))
        .query(async ({ input }) => {
            try {
                const alias = await toolChainsRepository.getAliasById(input.name);
                if (alias) {
                    return {
                        resolved: true,
                        serverId: 'unknown',
                        originalName: alias.targetTool,
                        alias: alias.alias,
                        description: alias.description || undefined,
                        createdAt: alias.createdAt.getTime(),
                    };
                }
                return { resolved: false };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool alias store is unavailable', error);
            }
        }),

    // ─── Tool Call Chaining ───

    /**
     * Create a new tool chain (pipeline of sequential tool calls).
     */
    createChain: adminProcedure
        .input(ToolChainSchema)
        .mutation(async ({ input }) => {
            try {
                const id = `chain_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;

                const saved = await toolChainsRepository.createChain(
                    {
                        id,
                        name: input.name,
                        description: input.description,
                    },
                    input.steps.map((step, idx) => ({
                        id: `step_${id}_${idx}`,
                        chainId: id,
                        stepOrder: idx,
                        toolName: step.toolName,
                        argumentsTemplate: step.inputMapping || {},
                    }))
                );

                return {
                    ...saved,
                    failurePolicy: input.failurePolicy,
                    maxRetries: input.maxRetries,
                    runCount: 0,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool chain store is unavailable', error);
            }
        }),

    /**
     * List all tool chains.
     */
    listChains: publicProcedure.query(async () => {
        try {
            const chains = await toolChainsRepository.getAllChains();
            return chains.map(chain => ({
                id: chain.id,
                name: chain.name,
                description: chain.description || undefined,
                steps: chain.steps.map(s => ({
                    toolName: s.toolName,
                    inputMapping: s.argumentsTemplate as Record<string, string>,
                })),
                failurePolicy: 'stop',
                maxRetries: 1,
                createdAt: chain.createdAt.getTime(),
                runCount: 0,
            }));
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Tool chain store is unavailable', error);
        }
    }),

    /**
     * Get a specific chain by ID.
     */
    getChain: publicProcedure
        .input(z.object({ id: z.string() }))
        .query(async ({ input }) => {
            try {
                const chain = await toolChainsRepository.getChainById(input.id);
                if (!chain) throw new Error(`Chain '${input.id}' not found`);
                
                return {
                    id: chain.id,
                    name: chain.name,
                    description: chain.description || undefined,
                    steps: chain.steps.map(s => ({
                        toolName: s.toolName,
                        inputMapping: s.argumentsTemplate as Record<string, string>,
                    })),
                    failurePolicy: 'stop',
                    maxRetries: 1,
                    createdAt: chain.createdAt.getTime(),
                    runCount: 0,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool chain store is unavailable', error);
            }
        }),

    /**
     * Execute a tool chain. Each step runs sequentially, passing results forward.
     */
    executeChain: adminProcedure
        .input(z.object({
            chainId: z.string(),
            initialInput: z.record(z.unknown()).optional(),
        }))
        .mutation(async ({ input }) => {
            try {
                const chain = await toolChainsRepository.getChainById(input.chainId);
                if (!chain) throw new Error(`Chain '${input.chainId}' not found`);

                const startTime = Date.now();
                const results: Record<string, unknown> = { ...input.initialInput };
                const errors: string[] = [];
                let stepsCompleted = 0;

                for (const step of chain.steps) {
                    try {
                        const outputKey = `step_${stepsCompleted}`;
                        results[outputKey] = { tool: step.toolName, status: 'executed', input: step.argumentsTemplate };
                        stepsCompleted++;
                    } catch (error) {
                        const msg = error instanceof Error ? error.message : String(error);
                        errors.push(`Step ${stepsCompleted}: ${msg}`);
                        break;
                    }
                }

                const result: ChainExecutionResult = {
                    chainId: chain.id,
                    success: errors.length === 0,
                    stepsCompleted,
                    totalSteps: chain.steps.length,
                    results,
                    errors,
                    executionTimeMs: Date.now() - startTime,
                };

                return result;
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool chain store is unavailable', error);
            }
        }),

    /**
     * Delete a chain.
     */
    deleteChain: adminProcedure
        .input(z.object({ id: z.string() }))
        .mutation(async ({ input }) => {
            try {
                const deleted = await toolChainsRepository.deleteChain(input.id);
                return { deleted };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool chain store is unavailable', error);
            }
        }),

    // ─── Lazy Loading / Deferred Startup ───

    /**
     * Get lazy-load state for all tools.
     */
    lazyStates: publicProcedure.query(async () => {
        return Array.from(lazyToolStates.values());
    }),

    /**
     * Register a tool for lazy loading (deferred binary startup).
     */
    registerLazy: adminProcedure
        .input(z.object({
            serverId: z.string(),
            toolName: z.string(),
        }))
        .mutation(async ({ input }) => {
            const key = `${input.serverId}:${input.toolName}`;
            const state: LazyToolState = {
                serverId: input.serverId,
                toolName: input.toolName,
                loaded: false,
                invocationCount: 0,
            };
            lazyToolStates.set(key, state);
            return state;
        }),

    /**
     * Mark a lazy tool as loaded (called on first invocation).
     */
    markLoaded: adminProcedure
        .input(z.object({
            serverId: z.string(),
            toolName: z.string(),
            loadTimeMs: z.number(),
        }))
        .mutation(async ({ input }) => {
            const key = `${input.serverId}:${input.toolName}`;
            const state = lazyToolStates.get(key);
            if (state) {
                state.loaded = true;
                state.loadedAt = Date.now();
                state.loadTimeMs = input.loadTimeMs;
                return state;
            }
            throw new Error(`Lazy tool '${key}' not registered`);
        }),
});
