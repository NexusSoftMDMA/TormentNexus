import {
    type ProviderAuthMethod,
    type ProviderAvailability,
    type ProviderBalanceConnection,
    type ProviderBalanceConnectionSource,
    type ProviderBalanceProvider,
    type ProviderBalanceTokenRefresher,
    type ProviderQuotaProvider,
    type ProviderQuotaSnapshot,
    type ProviderQuotaUnit,
    type ProviderQuotaWindowSnapshot,
} from './types.js';

interface ProviderBalanceSpec {
    provider: ProviderBalanceProvider;
    billingProvider: string;
    name: string;
    defaultAuthMethod: Extract<ProviderAuthMethod, 'oauth' | 'pat'>;
    exposeWhenMissing: boolean;
}

interface ProviderBalanceHooks {
    connectionSource: ProviderBalanceConnectionSource;
    tokenRefresher?: ProviderBalanceTokenRefresher;
}

type JsonRecord = Record<string, unknown>;

const PROVIDER_BALANCE_SPECS: ProviderBalanceSpec[] = [
    {
        provider: 'github',
        billingProvider: 'github-copilot',
        name: 'GitHub Copilot',
        defaultAuthMethod: 'oauth',
        exposeWhenMissing: true,
    },
    {
        provider: 'claude',
        billingProvider: 'anthropic',
        name: 'Anthropic',
        defaultAuthMethod: 'oauth',
        exposeWhenMissing: false,
    },
    {
        provider: 'codex',
        billingProvider: 'openai',
        name: 'OpenAI',
        defaultAuthMethod: 'oauth',
        exposeWhenMissing: false,
    },
    {
        provider: 'antigravity',
        billingProvider: 'antigravity',
        name: 'Antigravity',
        defaultAuthMethod: 'oauth',
        exposeWhenMissing: true,
    },
    {
        provider: 'kiro',
        billingProvider: 'kiro',
        name: 'Kiro',
        defaultAuthMethod: 'oauth',
        exposeWhenMissing: true,
    },
    {
        provider: 'kimi-coding',
        billingProvider: 'kimi-coding',
        name: 'Kimi Coding',
        defaultAuthMethod: 'oauth',
        exposeWhenMissing: true,
    },
];

const NULL_CONNECTION_SOURCE: ProviderBalanceConnectionSource = {
    async getConnection() {
        return null;
    },
};

let providerBalanceHooks: ProviderBalanceHooks = {
    connectionSource: NULL_CONNECTION_SOURCE,
};

export function configureProviderBalanceHooks(hooks: Partial<ProviderBalanceHooks>) {
    providerBalanceHooks = {
        ...providerBalanceHooks,
        ...hooks,
        connectionSource: hooks.connectionSource ?? providerBalanceHooks.connectionSource,
    };
}

function asRecord(value: unknown): JsonRecord {
    return value && typeof value === 'object' && !Array.isArray(value) ? (value as JsonRecord) : {};
}

function asNumber(value: unknown, fallback = 0): number {
    if (typeof value === 'number' && Number.isFinite(value)) {
        return value;
    }
    if (typeof value === 'string' && value.trim().length > 0) {
        const parsed = Number(value);
        return Number.isFinite(parsed) ? parsed : fallback;
    }
    return fallback;
}

function asString(value: unknown): string | null {
    return typeof value === 'string' && value.trim().length > 0 ? value : null;
}

function roundTo(value: number, digits = 2): number {
    const factor = 10 ** digits;
    return Math.round(value * factor) / factor;
}

function parseResetDate(value: unknown): string | null {
    if (value == null) {
        return null;
    }
    if (typeof value === 'number') {
        const millis = value > 10_000_000_000 ? value : value * 1000;
        return new Date(millis).toISOString();
    }
    if (typeof value === 'string' && value.trim().length > 0) {
        return new Date(value).toISOString();
    }
    return null;
}

function getRemainingRatio(window: ProviderQuotaWindowSnapshot): number | null {
    if (typeof window.remaining === 'number' && typeof window.limit === 'number' && window.limit > 0) {
        return window.remaining / window.limit;
    }
    return null;
}

function summarizeWindows(windows: ProviderQuotaWindowSnapshot[]) {
    if (windows.length === 0) {
        return {
            used: 0,
            limit: null,
            remaining: null,
            resetDate: null,
        };
    }

    const constrained = [...windows].sort((left, right) => {
        const leftRatio = getRemainingRatio(left);
        const rightRatio = getRemainingRatio(right);

        if (leftRatio == null && rightRatio == null) {
            return 0;
        }
        if (leftRatio == null) {
            return 1;
        }
        if (rightRatio == null) {
            return -1;
        }
        return leftRatio - rightRatio;
    })[0] ?? windows[0];

    return {
        used: constrained.used,
        limit: constrained.limit,
        remaining: constrained.remaining,
        resetDate: constrained.resetDate,
    };
}

function deriveAvailability(error: unknown): ProviderAvailability {
    const message = error instanceof Error ? error.message.toLowerCase() : String(error).toLowerCase();
    if (message.includes('quota') || message.includes('usage limit') || message.includes('exhaust')) {
        return 'quota_exhausted';
    }
    if (message.includes('rate limit') || message.includes('429')) {
        return 'rate_limited';
    }
    return 'cooldown';
}

function createSnapshot(
    spec: ProviderBalanceSpec,
    connection: ProviderBalanceConnection,
    tier: string,
    windows: ProviderQuotaWindowSnapshot[],
    detail = 'Live provider balance snapshot',
): ProviderQuotaSnapshot {
    const summary = summarizeWindows(windows);
    return {
        provider: spec.billingProvider,
        name: spec.name,
        authMethod: connection.authMethod,
        configured: true,
        authenticated: true,
        authTruth: 'authenticated',
        detail,
        used: summary.used,
        limit: summary.limit,
        remaining: summary.remaining,
        resetDate: summary.resetDate,
        rateLimitRpm: null,
        tier,
        availability: 'available',
        windows,
        source: 'balance',
        connectionId: connection.id,
        quotaConfidence: 'real-time',
        quotaRefreshedAt: new Date().toISOString(),
    };
}

function createMissingSnapshot(spec: ProviderBalanceSpec): ProviderQuotaSnapshot {
    return {
        provider: spec.billingProvider,
        name: spec.name,
        authMethod: spec.defaultAuthMethod,
        configured: false,
        authenticated: false,
        authTruth: 'not_configured',
        detail: `No ${spec.provider} balance connection configured.`,
        used: 0,
        limit: null,
        remaining: null,
        resetDate: null,
        rateLimitRpm: null,
        tier: 'unknown',
        availability: 'missing_auth',
        source: 'balance',
        connectionId: null,
        windows: [],
        quotaConfidence: 'unknown',
        quotaRefreshedAt: null,
    };
}

function createErrorSnapshot(spec: ProviderBalanceSpec, connection: ProviderBalanceConnection, error: unknown): ProviderQuotaSnapshot {
    const message = error instanceof Error ? error.message : String(error);
    return {
        provider: spec.billingProvider,
        name: spec.name,
        authMethod: connection.authMethod,
        configured: true,
        authenticated: true,
        authTruth: 'authenticated',
        detail: 'Balance lookup failed.',
        used: 0,
        limit: null,
        remaining: null,
        resetDate: null,
        rateLimitRpm: null,
        tier: 'unknown',
        availability: deriveAvailability(error),
        lastError: message,
        source: 'balance',
        connectionId: connection.id,
        windows: [],
        quotaConfidence: 'unknown',
        quotaRefreshedAt: null,
    };
}

function githubQuotaWindow(key: string, label: string, used: number, limit: number, remaining: number | null, resetDate: string | null): ProviderQuotaWindowSnapshot {
    return {
        key,
        label,
        used,
        limit,
        remaining,
        resetDate,
        unit: 'requests',
    };
}

function percentageWindow(key: string, label: string, usedPercent: number, resetDate: string | null): ProviderQuotaWindowSnapshot {
    const used = roundTo(usedPercent);
    const remaining = roundTo(Math.max(0, 100 - used));
    return {
        key,
        label,
        used,
        limit: 100,
        remaining,
        resetDate,
        unit: 'percent',
    };
}

function requestsWindow(
    key: string,
    label: string,
    used: number,
    limit: number,
    remaining: number | null,
    resetDate: string | null,
    unit: ProviderQuotaUnit = 'requests',
): ProviderQuotaWindowSnapshot {
    return {
        key,
        label,
        used: roundTo(used),
        limit: roundTo(limit),
        remaining: remaining == null ? null : roundTo(remaining),
        resetDate,
        unit,
    };
}

function mapAntigravityPlan(data: JsonRecord): string {
    const allowedTiers = Array.isArray(data.allowedTiers) ? data.allowedTiers : [];
    const defaultTier = allowedTiers.find((tier) => asRecord(tier).isDefault === true);
    const rawTier = asString(asRecord(defaultTier).id)
        ?? asString(asRecord(data.currentTier).id)
        ?? asString(asRecord(data.currentTier).name)
        ?? 'Free';
    const upper = rawTier.toUpperCase();
    if (upper.includes('ULTRA')) return 'Ultra';
    if (upper.includes('PRO')) return 'Pro';
    if (upper.includes('ENTERPRISE')) return 'Enterprise';
    if (upper.includes('BUSINESS') || upper.includes('STANDARD')) return 'Business';
    return 'Free';
}

function mapKimiPlan(level: string | null): string {
    switch (level) {
        case 'LEVEL_BASIC':
            return 'Moderato';
        case 'LEVEL_INTERMEDIATE':
            return 'Allegretto';
        case 'LEVEL_ADVANCED':
            return 'Allegro';
        case 'LEVEL_STANDARD':
            return 'Vivace';
        default:
            return 'Kimi Coding';
    }
}

async function fetchJson(url: string, init: RequestInit): Promise<JsonRecord> {
    const response = await fetch(url, init);
    if (!response.ok) {
        const body = await response.text();
        throw new Error(`HTTP ${response.status}: ${body.slice(0, 240)}`);
    }
    return asRecord(await response.json());
}

class GitHubQuotaProvider implements ProviderQuotaProvider {
    public readonly provider = 'github';
    public readonly billingProvider = 'github-copilot';
    public readonly name = 'GitHub Copilot';

    public async fetchQuotaSnapshot(connection: ProviderBalanceConnection): Promise<ProviderQuotaSnapshot> {
        const data = await fetchJson('https://api.github.com/copilot_internal/user', {
            headers: {
                Authorization: `token ${connection.accessToken}`,
                Accept: 'application/json',
                'X-GitHub-Api-Version': '2022-11-28',
                'User-Agent': 'GitHubCopilotChat/0.26.7',
                'Editor-Version': 'vscode/1.100.0',
                'Editor-Plugin-Version': 'copilot-chat/0.26.7',
            },
        });

        const plan = asString(data.copilot_plan) ?? asString(data.access_type_sku) ?? 'Copilot';
        const resetDate = parseResetDate(data.quota_reset_date ?? data.limited_user_reset_date);
        const windows: ProviderQuotaWindowSnapshot[] = [];
        const quotaSnapshots = asRecord(data.quota_snapshots);

        if (Object.keys(quotaSnapshots).length > 0) {
            for (const [key, value] of Object.entries(quotaSnapshots)) {
                const snapshot = asRecord(value);
                const limit = asNumber(snapshot.entitlement, 0);
                const remaining = snapshot.unlimited === true ? null : asNumber(snapshot.remaining, 0);
                windows.push(githubQuotaWindow(
                    key,
                    key.replace(/_/g, ' '),
                    remaining == null ? 0 : Math.max(limit - remaining, 0),
                    limit,
                    remaining,
                    resetDate,
                ));
            }
        } else {
            const monthly = asRecord(data.monthly_quotas);
            const used = asRecord(data.limited_user_quotas);
            for (const [key, value] of Object.entries(monthly)) {
                const limit = asNumber(value, 0);
                const usedValue = asNumber(used[key], 0);
                windows.push(githubQuotaWindow(key, key.replace(/_/g, ' '), usedValue, limit, Math.max(limit - usedValue, 0), resetDate));
            }
        }

        return createSnapshot(PROVIDER_BALANCE_SPECS[0], connection, plan, windows);
    }
}

class ClaudeQuotaProvider implements ProviderQuotaProvider {
    public readonly provider = 'claude';
    public readonly billingProvider = 'anthropic';
    public readonly name = 'Anthropic';

    public async fetchQuotaSnapshot(connection: ProviderBalanceConnection): Promise<ProviderQuotaSnapshot> {
        const data = await fetchJson('https://api.anthropic.com/api/oauth/usage', {
            headers: {
                Authorization: `Bearer ${connection.accessToken}`,
                'anthropic-beta': 'oauth-2025-04-20',
                'anthropic-version': '2023-06-01',
            },
        });

        const windows: ProviderQuotaWindowSnapshot[] = [];
        const standardWindows: Array<[string, string]> = [
            ['five_hour', 'Session (5h)'],
            ['seven_day', 'Weekly (7d)'],
        ];

        for (const [key, label] of standardWindows) {
            const value = asRecord(data[key]);
            if (typeof value.utilization === 'number') {
                windows.push(percentageWindow(key, label, asNumber(value.utilization, 0), parseResetDate(value.resets_at)));
            }
        }

        for (const [key, value] of Object.entries(data)) {
            if (!key.startsWith('seven_day_') || key === 'seven_day') {
                continue;
            }
            const window = asRecord(value);
            if (typeof window.utilization === 'number') {
                const modelName = key.replace('seven_day_', '').replace(/_/g, ' ');
                windows.push(percentageWindow(key, `Weekly ${modelName} (7d)`, asNumber(window.utilization, 0), parseResetDate(window.resets_at)));
            }
        }

        return createSnapshot(PROVIDER_BALANCE_SPECS[1], connection, 'Claude Code', windows);
    }
}

class CodexQuotaProvider implements ProviderQuotaProvider {
    public readonly provider = 'codex';
    public readonly billingProvider = 'openai';
    public readonly name = 'OpenAI';

    public async fetchQuotaSnapshot(connection: ProviderBalanceConnection): Promise<ProviderQuotaSnapshot> {
        const headers: Record<string, string> = {
            Authorization: `Bearer ${connection.accessToken}`,
            Accept: 'application/json',
        };
        const workspaceId = asString(connection.metadata?.workspaceId);
        if (workspaceId) {
            headers['chatgpt-account-id'] = workspaceId;
        }

        const data = await fetchJson('https://chatgpt.com/backend-api/wham/usage', { headers });
        const windows: ProviderQuotaWindowSnapshot[] = [];
        const primary = asRecord(data.rate_limit).primary_window ? asRecord(asRecord(data.rate_limit).primary_window) : asRecord(data.primary_window);
        const secondary = asRecord(data.rate_limit).secondary_window ? asRecord(asRecord(data.rate_limit).secondary_window) : asRecord(data.secondary_window);
        const codeReview = asRecord(asRecord(data.code_review_rate_limit).primary_window);

        if (Object.keys(primary).length > 0) {
            windows.push(percentageWindow('session', 'Session (5h)', asNumber(primary.used_percent ?? primary.usedPercent, 0), parseResetDate(primary.reset_at ?? primary.resetAt)));
        }
        if (Object.keys(secondary).length > 0) {
            windows.push(percentageWindow('weekly', 'Weekly', asNumber(secondary.used_percent ?? secondary.usedPercent, 0), parseResetDate(secondary.reset_at ?? secondary.resetAt)));
        }
        if (Object.keys(codeReview).length > 0) {
            windows.push(percentageWindow('code_review', 'Code review', asNumber(codeReview.used_percent ?? codeReview.usedPercent, 0), parseResetDate(codeReview.reset_at ?? codeReview.resetAt)));
        }

        return createSnapshot(PROVIDER_BALANCE_SPECS[2], connection, asString(data.plan_type) ?? 'Codex', windows);
    }
}

class AntigravityQuotaProvider implements ProviderQuotaProvider {
    public readonly provider = 'antigravity';
    public readonly billingProvider = 'antigravity';
    public readonly name = 'Antigravity';

    public async fetchQuotaSnapshot(connection: ProviderBalanceConnection): Promise<ProviderQuotaSnapshot> {
        const subscription = await fetchJson('https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist', {
            method: 'POST',
            headers: {
                Authorization: `Bearer ${connection.accessToken}`,
                'Content-Type': 'application/json',
                'User-Agent': 'google-api-nodejs-client/9.15.1',
                'X-Goog-Api-Client': 'google-cloud-sdk vscode_cloudshelleditor/0.1',
                'Client-Metadata': '{"ideType":"IDE_UNSPECIFIED","platform":"PLATFORM_UNSPECIFIED","pluginType":"GEMINI"}',
            },
            body: JSON.stringify({
                metadata: {
                    ideType: 'IDE_UNSPECIFIED',
                    platform: 'PLATFORM_UNSPECIFIED',
                    pluginType: 'GEMINI',
                },
            }),
        });

        const projectId = asString(subscription.cloudaicompanionProject);
        const data = await fetchJson('https://cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels', {
            method: 'POST',
            headers: {
                Authorization: `Bearer ${connection.accessToken}`,
                'Content-Type': 'application/json',
                'User-Agent': 'antigravity/1.11.3 Darwin/arm64',
            },
            body: JSON.stringify(projectId ? { project: projectId } : {}),
        });

        const models = asRecord(data.models);
        const windows: ProviderQuotaWindowSnapshot[] = [];
        for (const [key, value] of Object.entries(models)) {
            const model = asRecord(value);
            if (model.isInternal === true) {
                continue;
            }
            const quotaInfo = asRecord(model.quotaInfo);
            if (Object.keys(quotaInfo).length === 0) {
                continue;
            }
            const remainingPercent = roundTo(asNumber(quotaInfo.remainingFraction, 0) * 100);
            windows.push({
                key,
                label: asString(model.displayName) ?? key,
                used: roundTo(100 - remainingPercent),
                limit: 100,
                remaining: remainingPercent,
                resetDate: parseResetDate(quotaInfo.resetTime),
                unit: 'percent',
            });
        }

        return createSnapshot(PROVIDER_BALANCE_SPECS[3], connection, mapAntigravityPlan(subscription), windows);
    }
}

class KiroQuotaProvider implements ProviderQuotaProvider {
    public readonly provider = 'kiro';
    public readonly billingProvider = 'kiro';
    public readonly name = 'Kiro';

    public async fetchQuotaSnapshot(connection: ProviderBalanceConnection): Promise<ProviderQuotaSnapshot> {
        const profileArn = asString(connection.metadata?.profileArn);
        if (!profileArn) {
            throw new Error('Kiro balance lookup requires profileArn metadata.');
        }

        const data = await fetchJson('https://codewhisperer.us-east-1.amazonaws.com', {
            method: 'POST',
            headers: {
                Authorization: `Bearer ${connection.accessToken}`,
                'Content-Type': 'application/x-amz-json-1.0',
                'x-amz-target': 'AmazonCodeWhispererService.GetUsageLimits',
                Accept: 'application/json',
            },
            body: JSON.stringify({
                origin: 'AI_EDITOR',
                profileArn,
                resourceType: 'AGENTIC_REQUEST',
            }),
        });

        const windows: ProviderQuotaWindowSnapshot[] = [];
        const usageBreakdownList = Array.isArray(data.usageBreakdownList) ? data.usageBreakdownList : [];
        const resetDate = parseResetDate(data.nextDateReset ?? data.resetDate);

        for (const value of usageBreakdownList) {
            const item = asRecord(value);
            const key = asString(item.resourceType)?.toLowerCase() ?? 'usage';
            const used = asNumber(item.currentUsageWithPrecision, 0);
            const limit = asNumber(item.usageLimitWithPrecision, 0);
            windows.push(requestsWindow(key, key.replace(/_/g, ' '), used, limit, Math.max(limit - used, 0), resetDate));

            const freeTrial = asRecord(item.freeTrialInfo);
            if (Object.keys(freeTrial).length > 0) {
                const freeUsed = asNumber(freeTrial.currentUsageWithPrecision, 0);
                const freeLimit = asNumber(freeTrial.usageLimitWithPrecision, 0);
                windows.push(requestsWindow(`${key}_free_trial`, `${key.replace(/_/g, ' ')} free trial`, freeUsed, freeLimit, Math.max(freeLimit - freeUsed, 0), resetDate));
            }
        }

        return createSnapshot(PROVIDER_BALANCE_SPECS[4], connection, asString(asRecord(data.subscriptionInfo).subscriptionTitle) ?? 'Kiro', windows);
    }
}

class KimiCodingQuotaProvider implements ProviderQuotaProvider {
    public readonly provider = 'kimi-coding';
    public readonly billingProvider = 'kimi-coding';
    public readonly name = 'Kimi Coding';

    public async fetchQuotaSnapshot(connection: ProviderBalanceConnection): Promise<ProviderQuotaSnapshot> {
        const data = await fetchJson('https://api.kimi.com/coding/v1/usages', {
            headers: {
                Authorization: `Bearer ${connection.accessToken}`,
                'Content-Type': 'application/json',
                'X-Msh-Platform': 'hypercode',
                'X-Msh-Version': '0.1.0',
                'X-Msh-Device-Model': 'hypercode-core',
                'X-Msh-Device-Id': connection.id,
            },
        });

        const windows: ProviderQuotaWindowSnapshot[] = [];
        const usage = asRecord(data.usage);
        const limit = asNumber(usage.limit ?? usage.Limit, 0);
        const used = asNumber(usage.used ?? usage.Used, 0);
        const remaining = asNumber(usage.remaining ?? usage.Remaining, 0);
        if (limit > 0) {
            windows.push(requestsWindow('weekly', 'Weekly', used, limit, remaining, parseResetDate(usage.resetTime ?? usage.reset_at ?? usage.resetAt)));
        }

        const limits = Array.isArray(data.limits) ? data.limits : [];
        for (const limitValue of limits) {
            const item = asRecord(limitValue);
            const detail = asRecord(item.detail);
            const rateLimit = asNumber(detail.limit ?? detail.Limit, 0);
            const rateRemaining = asNumber(detail.remaining ?? detail.Remaining, 0);
            if (rateLimit > 0) {
                windows.push(requestsWindow('rate_limit', 'Rate limit', rateLimit - rateRemaining, rateLimit, rateRemaining, parseResetDate(detail.resetTime ?? detail.reset_at ?? detail.resetAt)));
            }
        }

        const membership = asRecord(asRecord(data.user).membership);
        return createSnapshot(PROVIDER_BALANCE_SPECS[5], connection, mapKimiPlan(asString(membership.level)), windows);
    }
}

export class ProviderBalanceService {
    private readonly connectionSource: ProviderBalanceConnectionSource;
    private readonly tokenRefresher?: ProviderBalanceTokenRefresher;
    private readonly providers = new Map<ProviderBalanceProvider, ProviderQuotaProvider>();

    constructor(hooks: Partial<ProviderBalanceHooks> = {}) {
        this.connectionSource = hooks.connectionSource ?? providerBalanceHooks.connectionSource;
        this.tokenRefresher = hooks.tokenRefresher ?? providerBalanceHooks.tokenRefresher;

        const providerImplementations: ProviderQuotaProvider[] = [
            new GitHubQuotaProvider(),
            new ClaudeQuotaProvider(),
            new CodexQuotaProvider(),
            new AntigravityQuotaProvider(),
            new KiroQuotaProvider(),
            new KimiCodingQuotaProvider(),
        ];

        for (const provider of providerImplementations) {
            this.providers.set(provider.provider, provider);
        }
    }

    public getSupportedProviders() {
        return [...PROVIDER_BALANCE_SPECS];
    }

    public async fetchSnapshots(): Promise<ProviderQuotaSnapshot[]> {
        const snapshots: ProviderQuotaSnapshot[] = [];

        for (const spec of PROVIDER_BALANCE_SPECS) {
            const connection = await this.connectionSource.getConnection(spec.provider);
            if (!connection) {
                if (spec.exposeWhenMissing) {
                    snapshots.push(createMissingSnapshot(spec));
                }
                continue;
            }

            const provider = this.providers.get(spec.provider);
            if (!provider) {
                continue;
            }

            try {
                const refreshed = this.tokenRefresher
                    ? await this.tokenRefresher.refreshConnection(connection)
                    : connection;
                snapshots.push(await provider.fetchQuotaSnapshot(refreshed));
            } catch (error) {
                snapshots.push(createErrorSnapshot(spec, connection, error));
            }
        }

        return snapshots;
    }
}