import type { BillingRoutingStrategy, BillingTaskRoutingRuleSummary } from './billing-portal-data';

type UnknownRecord = Record<string, unknown>;

const TASK_TYPES: Set<BillingTaskRoutingRuleSummary['taskType']> = new Set([
    'coding',
    'planning',
    'research',
    'general',
    'worker',
    'supervisor',
]);

const ROUTING_STRATEGIES: Set<BillingRoutingStrategy> = new Set([
    'best',
    'cheapest',
    'round-robin',
]);

function asRecord(value: unknown): UnknownRecord | null {
    return value && typeof value === 'object' && !Array.isArray(value)
        ? (value as UnknownRecord)
        : null;
}

function toNumber(value: unknown, fallback = 0): number {
    return typeof value === 'number' && Number.isFinite(value)
        ? value
        : fallback;
}

function toStringValue(value: unknown, fallback = ''): string {
    return typeof value === 'string' && value.trim().length > 0
        ? value.trim()
        : fallback;
}

function toBoolean(value: unknown, fallback = false): boolean {
    return typeof value === 'boolean'
        ? value
        : fallback;
}

function toTaskType(value: unknown, fallback: BillingTaskRoutingRuleSummary['taskType'] = 'general'): BillingTaskRoutingRuleSummary['taskType'] {
    return typeof value === 'string' && TASK_TYPES.has(value as BillingTaskRoutingRuleSummary['taskType'])
        ? (value as BillingTaskRoutingRuleSummary['taskType'])
        : fallback;
}

function toRoutingStrategy(value: unknown, fallback: BillingRoutingStrategy = 'best'): BillingRoutingStrategy {
    return typeof value === 'string' && ROUTING_STRATEGIES.has(value as BillingRoutingStrategy)
        ? (value as BillingRoutingStrategy)
        : fallback;
}

export interface BillingUsageBreakdownRow {
    provider: string;
    requests: number;
    cost: number;
}

export interface BillingUsageSummary {
    currentMonth: number;
    limit: number;
    breakdown: BillingUsageBreakdownRow[];
}

export function getBillingUsageSummary(status: unknown): BillingUsageSummary {
    const statusRecord = asRecord(status);
    const usage = asRecord(statusRecord?.usage);
    const breakdown = Array.isArray(usage?.breakdown)
        ? usage.breakdown
            .map((entry, index) => {
                const row = asRecord(entry);
                if (!row) {
                    return null;
                }

                return {
                    provider: toStringValue(row.provider, `provider-${index + 1}`),
                    requests: toNumber(row.requests, 0),
                    cost: toNumber(row.cost, 0),
                } satisfies BillingUsageBreakdownRow;
            })
            .filter((entry): entry is BillingUsageBreakdownRow => Boolean(entry))
        : [];

    return {
        currentMonth: toNumber(usage?.currentMonth, 0),
        limit: toNumber(usage?.limit, 0),
        breakdown,
    };
}

export interface BillingFallbackLink {
    priority: number;
    provider: string;
    model?: string;
    reason: string;
}

export function getFallbackTaskType(
    fallback: unknown,
    defaultTaskType: BillingTaskRoutingRuleSummary['taskType'],
): BillingTaskRoutingRuleSummary['taskType'] {
    return toTaskType(asRecord(fallback)?.selectedTaskType, defaultTaskType);
}

export function normalizeFallbackChain(fallback: unknown): BillingFallbackLink[] {
    const fallbackRecord = asRecord(fallback);
    const chain = Array.isArray(fallbackRecord?.chain) ? fallbackRecord.chain : [];

    return chain.reduce<BillingFallbackLink[]>((acc, entry, index) => {
            const row = asRecord(entry);
            if (!row) {
                return acc;
            }

            acc.push({
                priority: Math.max(1, Math.floor(toNumber(row.priority, index + 1))),
                provider: toStringValue(row.provider, `provider-${index + 1}`),
                model: toStringValue(row.model) || undefined,
                reason: toStringValue(row.reason, 'ranked fallback'),
            });

            return acc;
        }, []);
}

export function getDefaultRoutingStrategy(taskRouting: unknown): BillingRoutingStrategy {
    return toRoutingStrategy(asRecord(taskRouting)?.defaultStrategy, 'best');
}

export function normalizeTaskRoutingRules(taskRouting: unknown): BillingTaskRoutingRuleSummary[] {
    const routingRecord = asRecord(taskRouting);
    const rules = Array.isArray(routingRecord?.rules) ? routingRecord.rules : [];

    return rules.reduce<BillingTaskRoutingRuleSummary[]>((acc, entry) => {
            const row = asRecord(entry);
            if (!row) {
                return acc;
            }

            const fallbackPreview = Array.isArray(row.fallbackPreview)
                ? row.fallbackPreview.reduce<Array<{ provider: string; model?: string; reason?: string }>>((previewAcc, candidate, index) => {
                        const candidateRecord = asRecord(candidate);
                        if (!candidateRecord) {
                            return previewAcc;
                        }

                        previewAcc.push({
                            provider: toStringValue(candidateRecord.provider, `provider-${index + 1}`),
                            model: toStringValue(candidateRecord.model) || undefined,
                            reason: toStringValue(candidateRecord.reason) || undefined,
                        });

                        return previewAcc;
                    }, [])
                : [];

            acc.push({
                taskType: toTaskType(row.taskType, 'general'),
                strategy: toRoutingStrategy(row.strategy, 'best'),
                fallbackPreview,
            });

            return acc;
        }, []);
}

export interface BillingQuotaTableRow {
    provider: string;
    name: string;
    configured: boolean;
    authenticated: boolean;
    authMethod: string;
    tier: string;
    limit: number | null;
    used: number;
    rateLimitRpm: number | null;
    availability: string;
    lastError: string | null;
}

export function normalizeBillingQuotaRows(quotas: unknown): BillingQuotaTableRow[] {
    if (!Array.isArray(quotas)) {
        return [];
    }

    return quotas
        .map((entry, index) => {
            const row = asRecord(entry);
            if (!row) {
                return null;
            }

            const provider = toStringValue(row.provider, `provider-${index + 1}`);
            return {
                provider,
                name: toStringValue(row.name, provider),
                configured: toBoolean(row.configured, false),
                authenticated: toBoolean(row.authenticated, false),
                authMethod: toStringValue(row.authMethod, 'none'),
                tier: toStringValue(row.tier, 'standard'),
                limit: row.limit === null ? null : toNumber(row.limit, 0),
                used: toNumber(row.used, 0),
                rateLimitRpm: row.rateLimitRpm === null ? null : toNumber(row.rateLimitRpm, 0),
                availability: toStringValue(row.availability, 'unknown'),
                lastError: toStringValue(row.lastError) || null,
            } satisfies BillingQuotaTableRow;
        })
        .filter((entry): entry is BillingQuotaTableRow => Boolean(entry));
}

export interface BillingPricingModelRow {
    id: string;
    contextWindow: number | null;
    inputPrice: number | null;
    inputPricePer1k: number;
    outputPricePer1k: number;
    recommended: boolean;
}

export function normalizeBillingPricingModels(pricing: unknown): BillingPricingModelRow[] {
    const pricingRecord = asRecord(pricing);
    const models = Array.isArray(pricingRecord?.models) ? pricingRecord.models : [];

    return models
        .map((entry, index) => {
            const row = asRecord(entry);
            if (!row) {
                return null;
            }

            return {
                id: toStringValue(row.id, `model-${index + 1}`),
                contextWindow: row.contextWindow === null ? null : toNumber(row.contextWindow, 0),
                inputPrice: row.inputPrice === null ? null : toNumber(row.inputPrice, 0),
                inputPricePer1k: toNumber(row.inputPricePer1k, 0),
                outputPricePer1k: toNumber(row.outputPricePer1k, 0),
                recommended: toBoolean(row.recommended, false),
            } satisfies BillingPricingModelRow;
        })
        .filter((entry): entry is BillingPricingModelRow => Boolean(entry));
}
