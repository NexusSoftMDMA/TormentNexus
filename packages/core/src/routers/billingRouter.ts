import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { adminProcedure, t, publicProcedure, getLLMService } from '../lib/trpc-core.js';
import type { ProviderRoutingStrategy, ProviderTaskType } from '../providers/types.js';

interface ProviderQuotaRuntime {
    provider: string;
    name: string;
    configured: boolean;
    authenticated: boolean;
    authMethod: string;
    /** @see ProviderAuthTruth */
    authTruth?: string;
    tier: string;
    limit: number | null;
    used: number;
    remaining: number | null;
    resetDate: string | null;
    rateLimitRpm: number | null;
    availability: string;
    lastError?: string;
    windows?: Array<{
        key: string;
        label: string;
        used: number;
        limit: number | null;
        remaining: number | null;
        resetDate: string | null;
        unit: string;
    }>;
    source?: string;
    connectionId?: string | null;
    /** @see QuotaDataConfidence */
    quotaConfidence?: string;
    /** ISO-8601 timestamp of last real-time fetch from provider, or null. */
    quotaRefreshedAt?: string | null;
}

interface BillingSelectorRuntime {
    getAvailableModels?: () => BillingModelRuntime[];
    getFallbackChain?: (options?: { routingTaskType?: ProviderTaskType }) => FallbackEntryRuntime[];
    getProviderSnapshots?: () => ProviderQuotaRuntime[] | Promise<ProviderQuotaRuntime[]>;
    getRoutingStrategy?: () => ProviderRoutingStrategy;
    setRoutingStrategy?: (strategy: ProviderRoutingStrategy) => void;
    getTaskRoutingRules?: () => Record<ProviderTaskType, ProviderRoutingStrategy>;
    setTaskRoutingStrategy?: (taskType: ProviderTaskType, strategy: ProviderRoutingStrategy) => void;
    resetTaskRoutingStrategy?: (taskType: ProviderTaskType) => void;
    getFallbackHistory?: (limit?: number) => ProviderFallbackEventRuntime[];
    clearFallbackHistory?: () => void;
}

interface ProviderFallbackEventRuntime {
    id: number;
    timestamp: number;
    requestedProvider?: string;
    selectedProvider: string;
    selectedModelId: string;
    taskType: ProviderTaskType;
    strategy: ProviderRoutingStrategy;
    reason: string;
    causeCode: 'preference_honored' | 'fallback_provider' | 'budget_forced_local' | 'emergency_fallback';
}

interface BillingModelRuntime {
    id: string;
    provider: string;
    name?: string;
    inputPrice?: number | null;
    outputPrice?: number | null;
    contextWindow?: number | null;
    tier?: string;
    recommended?: boolean;
}

interface FallbackEntryRuntime {
    provider: string;
    model?: string;
    reason?: string;
}

export const TASK_TYPE_VALUES = ['coding', 'planning', 'research', 'general', 'worker', 'supervisor'] as const satisfies readonly ProviderTaskType[];
const TASK_TYPES: ProviderTaskType[] = [...TASK_TYPE_VALUES];

function buildBillingUnavailableError(label: string, error: unknown): TRPCError {
    const message = error instanceof Error ? error.message : String(error);
    return new TRPCError({
        code: 'INTERNAL_SERVER_ERROR',
        message: `${label} is unavailable: ${message}`,
    });
}

export function buildFallbackChainResponse(
    selector: BillingSelectorRuntime,
    taskType?: ProviderTaskType,
) {
    const chain = (selector.getFallbackChain?.(taskType ? { routingTaskType: taskType } : undefined) ?? []) as FallbackEntryRuntime[];
    return {
        selectedTaskType: taskType ?? null,
        chain: chain.map((entry, index: number) => ({
            priority: index + 1,
            provider: entry.provider,
            model: entry.model,
            reason: entry.reason ?? 'configured',
        })),
    };
}

export const billingRouter = t.router({
    /** Get current billing status — keys, usage, and provider breakdown */
    getStatus: publicProcedure.query(async () => {
        const keys = {
            openai: !!process.env.OPENAI_API_KEY,
            anthropic: !!process.env.ANTHROPIC_API_KEY,
            gemini: !!process.env.GEMINI_API_KEY,
            mistral: !!process.env.MISTRAL_API_KEY,
            deepseek: !!process.env.DEEPSEEK_API_KEY,
            xai: !!process.env.XAI_API_KEY,
            openrouter: !!process.env.OPENROUTER_API_KEY,
            groq: !!process.env.GROQ_API_KEY,
        };

        try {
        const llm = getLLMService();
        const stats = llm.getCostStats();
        const quota = llm.modelSelector.getQuotaService();

        let breakdown = quota.getUsageByModel();
        if (breakdown.length === 0) {
            breakdown = [{ provider: 'No Usage Yet', cost: 0, requests: 0 }];
        }

        const usage = {
            currentMonth: stats.estimatedCostUSD,
            limit: 100.00,
            breakdown,
        };

        return { keys, usage };
        } catch {
            return { keys, usage: { currentMonth: 0, limit: 100.00, breakdown: [{ provider: 'No Usage Yet', cost: 0, requests: 0 }] } };
        }
    }),

    /** Get quota information per provider — limits, remaining, reset dates */
    getProviderQuotas: publicProcedure.query(async () => {
        const llm = getLLMService();
        const selector = llm.modelSelector as typeof llm.modelSelector & BillingSelectorRuntime & {
            getQuotaService?: () => {
                getQuota?: (providerId: string) => {
                    authTruth?: string;
                    tier?: string;
                    limit?: number | null;
                    used?: number;
                    remaining?: number | null;
                    resetDate?: string | null;
                    rateLimitRpm?: number | null;
                    quotaConfidence?: string;
                    quotaRefreshedAt?: string | null;
                };
            };
        };

        const normalized = await selector.getProviderSnapshots?.();
        if (normalized && normalized.length > 0) {
            return normalized.map((provider) => ({
                provider: provider.provider,
                name: provider.name,
                configured: provider.configured,
                authenticated: provider.authenticated,
                authMethod: provider.authMethod,
                authTruth: provider.authTruth ?? null,
                tier: provider.tier,
                limit: provider.limit,
                used: provider.used,
                remaining: provider.remaining,
                resetDate: provider.resetDate,
                rateLimitRpm: provider.rateLimitRpm,
                availability: provider.availability,
                lastError: provider.lastError ?? null,
                windows: provider.windows ?? [],
                source: provider.source ?? null,
                connectionId: provider.connectionId ?? null,
                quotaConfidence: provider.quotaConfidence ?? 'estimated',
                quotaRefreshedAt: provider.quotaRefreshedAt ?? null,
            }));
        }

        const quota = selector.getQuotaService?.();
        const providers = [
            { id: 'openai', name: 'OpenAI', envKeys: ['OPENAI_API_KEY'] },
            { id: 'anthropic', name: 'Anthropic', envKeys: ['ANTHROPIC_API_KEY'] },
            { id: 'gemini', name: 'Google Gemini', envKeys: ['GEMINI_API_KEY', 'GOOGLE_API_KEY'] },
            { id: 'deepseek', name: 'DeepSeek', envKeys: ['DEEPSEEK_API_KEY'] },
            { id: 'xai', name: 'xAI (Grok)', envKeys: ['XAI_API_KEY'] },
            { id: 'mistral', name: 'Mistral', envKeys: ['MISTRAL_API_KEY'] },
            { id: 'openrouter', name: 'OpenRouter', envKeys: ['OPENROUTER_API_KEY'] },
            { id: 'groq', name: 'Groq', envKeys: ['GROQ_API_KEY'] },
        ];

        return providers.map((provider) => {
            const hasKey = provider.envKeys.some((envKey) => Boolean(process.env[envKey]));
            const quotaInfo = quota?.getQuota?.(provider.id);
            return {
                provider: provider.id,
                name: provider.name,
                configured: hasKey,
                authenticated: hasKey,
                authMethod: hasKey ? 'api_key' : 'none',
                authTruth: quotaInfo?.authTruth ?? (hasKey ? 'authenticated' : 'not_configured'),
                tier: quotaInfo?.tier ?? 'unknown',
                limit: quotaInfo?.limit ?? null,
                used: quotaInfo?.used ?? 0,
                remaining: quotaInfo?.remaining ?? null,
                resetDate: quotaInfo?.resetDate ?? null,
                rateLimitRpm: quotaInfo?.rateLimitRpm ?? null,
                availability: hasKey ? 'available' : 'missing_config',
                lastError: null,
                quotaConfidence: quotaInfo?.quotaConfidence ?? 'estimated',
                quotaRefreshedAt: quotaInfo?.quotaRefreshedAt ?? null,
            };
        });
    }),

    /** Get cost history over time for charts */
    getCostHistory: publicProcedure.input(z.object({
        days: z.number().min(1).max(90).default(30),
    }).optional()).query(async ({ input }) => {
        const llm = getLLMService();
        const stats = llm.getCostStats();

        // Build daily breakdown from metrics if available
        const days = input?.days ?? 30;
        const history: { date: string; cost: number; requests: number }[] = [];
        const statsAny = stats;

        // If there's a cost history, use it; otherwise generate from current stats
        if (statsAny.dailyHistory) {
            return { history: statsAny.dailyHistory.slice(-days) };
        }

        // Fallback: single entry for today
        const today = new Date().toISOString().split('T')[0];
        history.push({
            date: today,
            cost: stats.estimatedCostUSD ?? 0,
            requests: statsAny.totalRequests ?? 0,
        });

        return { history };
    }),

    /** Get model-level pricing and efficiency data */
    getModelPricing: publicProcedure.query(async () => {
        const llm = getLLMService();
        const selector = llm.modelSelector as typeof llm.modelSelector & BillingSelectorRuntime;

        // Get available models with their pricing info
        const models = (selector.getAvailableModels?.() ?? []) as BillingModelRuntime[];
        return {
            models: models.map((m) => ({
                id: m.id,
                provider: m.provider,
                name: m.name ?? m.id,
                inputPricePer1k: m.inputPrice ?? null,
                outputPricePer1k: m.outputPrice ?? null,
                contextWindow: m.contextWindow ?? null,
                tier: m.tier ?? 'standard',
                recommended: m.recommended ?? false,
            })),
        };
    }),

    /** Get current fallback chain configuration */
    getFallbackChain: publicProcedure
        .input(z.object({
            taskType: z.enum(TASK_TYPE_VALUES).optional(),
        }).optional())
        .query(async ({ input }) => {
        const llm = getLLMService();
        const selector = llm.modelSelector as typeof llm.modelSelector & BillingSelectorRuntime;
        return buildFallbackChainResponse(selector, input?.taskType);
    }),

    /** Get task-specific routing rules and top-ranked fallback previews */
    getTaskRoutingRules: publicProcedure.query(async () => {
        const llm = getLLMService();
        const selector = llm.modelSelector as typeof llm.modelSelector & BillingSelectorRuntime;
        const taskRules = selector.getTaskRoutingRules?.() ?? {
            coding: 'cheapest',
            planning: 'best',
            research: 'best',
            general: 'round-robin',
            worker: 'cheapest',
            supervisor: 'best',
        } satisfies Record<ProviderTaskType, ProviderRoutingStrategy>;

        const defaultStrategy = selector.getRoutingStrategy?.() ?? 'best';

        return {
            defaultStrategy,
            rules: TASK_TYPES.map((taskType) => ({
                taskType,
                strategy: taskRules[taskType] ?? defaultStrategy,
                fallbackPreview: (selector.getFallbackChain?.({ routingTaskType: taskType }) ?? [])
                    .slice(0, 3)
                    .map((entry) => ({
                        provider: entry.provider,
                        model: entry.model,
                        reason: entry.reason ?? 'configured',
                    })),
            })),
        };
    }),

    setRoutingStrategy: adminProcedure
        .input(z.object({
            strategy: z.enum(['cheapest', 'best', 'round-robin']),
        }))
        .mutation(async ({ input }) => {
            const llm = getLLMService();
            const selector = llm.modelSelector as typeof llm.modelSelector & BillingSelectorRuntime;

            selector.setRoutingStrategy?.(input.strategy);

            return {
                ok: true,
                strategy: selector.getRoutingStrategy?.() ?? input.strategy,
            };
        }),

    setTaskRoutingRule: adminProcedure
        .input(z.object({
            taskType: z.enum(TASK_TYPE_VALUES),
            strategy: z.enum(['cheapest', 'best', 'round-robin']).nullable(),
        }))
        .mutation(async ({ input }) => {
            const llm = getLLMService();
            const selector = llm.modelSelector as typeof llm.modelSelector & BillingSelectorRuntime;

            if (input.strategy) {
                selector.setTaskRoutingStrategy?.(input.taskType, input.strategy);
            } else {
                selector.resetTaskRoutingStrategy?.(input.taskType);
            }

            const rules = selector.getTaskRoutingRules?.() ?? {
                coding: 'cheapest',
                planning: 'best',
                research: 'best',
                general: 'round-robin',
                worker: 'cheapest',
                supervisor: 'best',
            } satisfies Record<ProviderTaskType, ProviderRoutingStrategy>;

            return {
                ok: true,
                taskType: input.taskType,
                strategy: rules[input.taskType],
            };
        }),

    /** Get models currently depleted or on cooldown in the ModelSelector session state. */
    getDepletedModels: publicProcedure.query(async () => {
        const llm = getLLMService();
        try {
            return llm.modelSelector.getDepletedModels?.() ?? [];
        } catch (error) {
            throw buildBillingUnavailableError('Depleted model status', error);
        }
    }),

    /**
     * Get recent provider fallback decisions from the in-process ring buffer.
     *
     * Each entry represents a routing decision where the preferred provider was not
     * honored (due to quota exhaustion, rate limits, budget cap, or no candidates
     * available at all).  Returns entries in reverse-chronological order.
     *
     * The billing dashboard "Recent Fallback Decisions" card consumes this to let
     * operators understand why TormentNexus substituted a different provider/model than
     * configured.
     */
    getFallbackHistory: publicProcedure
        .input(z.object({ limit: z.number().min(1).max(50).optional() }).optional())
        .query(async ({ input }) => {
            const llm = getLLMService();
            try {
                const selector = llm.modelSelector as unknown as {
                    getFallbackHistory?: (limit: number) => Array<{
                        id: number;
                        timestamp: number;
                        requestedProvider?: string;
                        selectedProvider: string;
                        selectedModelId: string;
                        taskType: string;
                        strategy: string;
                        reason: string;
                        causeCode: string;
                    }>;
                };
                return selector.getFallbackHistory?.(input?.limit ?? 20) ?? [];
            } catch (error) {
                throw buildBillingUnavailableError('Fallback history', error);
            }
        }),

    /** Clear in-memory provider fallback history ring buffer. */
    clearFallbackHistory: adminProcedure
        .mutation(async () => {
            const llm = getLLMService();
            try {
                const selector = llm.modelSelector as unknown as {
                    clearFallbackHistory?: () => void;
                };
                selector.clearFallbackHistory?.();
                return { ok: true };
            } catch (error) {
                throw buildBillingUnavailableError('Fallback history clearing', error);
            }
        }),
});

