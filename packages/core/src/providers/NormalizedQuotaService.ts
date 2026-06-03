import { QuotaService, type QuotaConfig } from '@tormentnexus/ai';

import { ProviderBalanceService } from './ProviderBalanceService.js';
import { ProviderRegistry } from './ProviderRegistry.js';
import { type ProviderAvailability, type ProviderQuotaSnapshot } from './types.js';

interface UsageRecord {
    provider: string;
    modelId: string;
    costUsd: number;
    inputTokens: number;
    outputTokens: number;
    timestamp: number;
}

const DEFAULT_RETRY_MS = 60_000;

export class NormalizedQuotaService extends QuotaService {
    private readonly registry: ProviderRegistry;
    private readonly balanceService: ProviderBalanceService;
    private readonly providerUsage: UsageRecord[] = [];
    protected configState: QuotaConfig = {
        dailyBudgetUsd: 5,
        monthlyBudgetUsd: 100,
        providerLimits: {},
    };
    private readonly snapshots = new Map<string, ProviderQuotaSnapshot>();
    private lastEnv: NodeJS.ProcessEnv = process.env;

    constructor(registry: ProviderRegistry, balanceService: ProviderBalanceService = new ProviderBalanceService()) {
        super();
        this.registry = registry;
        this.balanceService = balanceService;
        this.refreshAuthStates(this.lastEnv);
    }

    public setConfig(config: QuotaConfig) {
        this.configState = {
            ...this.configState,
            ...config,
            providerLimits: {
                ...(this.configState.providerLimits ?? {}),
                ...(config.providerLimits ?? {}),
            },
        };
        this.refreshAuthStates(this.lastEnv);
    }

    private readonly revokedProviders = new Set<string>();

    public markAuthRevoked(provider: string, reason?: string) {
        this.revokedProviders.add(provider);
        const snapshot = this.snapshots.get(provider);
        if (snapshot) {
            this.snapshots.set(provider, {
                ...snapshot,
                authTruth: 'revoked',
                authenticated: false,
                availability: 'missing_auth',
                lastError: reason ?? 'Authentication revoked',
            });
        }
    }

    public refreshAuthStates(env: NodeJS.ProcessEnv = process.env) {
        this.lastEnv = env;
        for (const authState of this.registry.getAuthStates(env)) {
            const existing = this.snapshots.get(authState.provider);
            const limit = this.configState.providerLimits?.[authState.provider] ?? existing?.limit ?? null;
            const used = existing?.used ?? 0;

            const isRevoked = this.revokedProviders.has(authState.provider);
            const authTruth = isRevoked ? 'revoked' : authState.authTruth;
            const authenticated = isRevoked ? false : authState.authenticated;
            const availability = isRevoked 
                ? 'missing_auth'
                : authenticated ? (existing?.availability ?? 'available') : 'missing_auth';

            const quotaConfidence = existing?.quotaConfidence ?? 'estimated';
            const quotaRefreshedAt = existing?.quotaRefreshedAt ?? null;

            this.snapshots.set(authState.provider, {
                ...authState,
                authTruth,
                authenticated,
                used,
                limit,
                remaining: typeof limit === 'number' ? Math.max(limit - used, 0) : null,
                resetDate: existing?.resetDate ?? null,
                rateLimitRpm: existing?.rateLimitRpm ?? null,
                tier: existing?.tier ?? this.registry.getProvider(authState.provider)?.models[0]?.tier ?? 'standard',
                availability,
                lastError: existing?.lastError,
                retryAfter: existing?.retryAfter ?? null,
                quotaConfidence,
                quotaRefreshedAt,
            });
        }
    }

    public trackUsage(modelId: string, inputTokens: number, outputTokens: number) {
        const model = this.registry.getModel(modelId);
        const provider = model?.provider ?? 'unknown';
        const inputPrice = model?.inputPrice ?? 0;
        const outputPrice = model?.outputPrice ?? 0;
        const costUsd = (inputTokens / 1000 * inputPrice) + (outputTokens / 1000 * outputPrice);

        this.providerUsage.push({
            provider,
            modelId,
            costUsd,
            inputTokens,
            outputTokens,
            timestamp: Date.now(),
        });

        const snapshot = this.snapshots.get(provider);
        if (snapshot) {
            const nextUsed = snapshot.used + costUsd;
            const limit = snapshot.limit;
            const threshold = this.configState.preEmptiveSwitchThreshold ?? 0.95;

            let availability = snapshot.availability;
            if (typeof limit === 'number' && limit > 0) {
                if (nextUsed >= limit) {
                    availability = 'quota_exhausted';
                } else if (nextUsed / limit >= threshold) {
                    availability = 'cooldown';
                }
            }

            this.snapshots.set(provider, {
                ...snapshot,
                used: nextUsed,
                remaining: typeof limit === 'number' ? Math.max(limit - nextUsed, 0) : null,
                availability,
            });
        }
    }

    public getSessionTotal(): number {
        return this.providerUsage.reduce((total, entry) => total + entry.costUsd, 0);
    }

    public getDailyTotal(): number {
        const today = new Date().setHours(0, 0, 0, 0);
        return this.providerUsage
            .filter((entry) => entry.timestamp >= today)
            .reduce((total, entry) => total + entry.costUsd, 0);
    }

    public getUsageByModel() {
        const byProvider = new Map<string, { cost: number; requests: number }>();
        for (const entry of this.providerUsage) {
            const current = byProvider.get(entry.provider) ?? { cost: 0, requests: 0 };
            current.cost += entry.costUsd;
            current.requests += 1;
            byProvider.set(entry.provider, current);
        }

        return Array.from(byProvider.entries()).map(([provider, stats]) => ({
            provider,
            cost: stats.cost,
            requests: stats.requests,
        }));
    }

    public isBudgetExceeded(): boolean {
        return this.getDailyTotal() >= this.configState.dailyBudgetUsd;
    }

    public getNearQuotaWarnings(): Array<{ provider: string; usedPercent: number }> {
        const warnings: Array<{ provider: string; usedPercent: number }> = [];
        const threshold = this.configState.preEmptiveSwitchThreshold ?? 0.95;
        for (const [provider, snapshot] of this.snapshots.entries()) {
            if (typeof snapshot.limit === 'number' && snapshot.limit > 0) {
                const usedPercent = snapshot.used / snapshot.limit;
                if (usedPercent >= threshold && usedPercent < 1.0) {
                    warnings.push({
                        provider,
                        usedPercent: Math.round(usedPercent * 100),
                    });
                }
            }
        }
        return warnings;
    }

    public getReport() {
        return {
            session: this.getSessionTotal(),
            daily: this.getDailyTotal(),
            config: this.configState,
            isExceeded: this.isBudgetExceeded(),
        };
    }

    public async refreshProviderBalances() {
        const liveSnapshots = await this.balanceService.fetchSnapshots();
        for (const snapshot of liveSnapshots) {
            const existing = this.snapshots.get(snapshot.provider);

            let quotaRefreshedAt = snapshot.quotaRefreshedAt ?? existing?.quotaRefreshedAt ?? null;
            if ((quotaRefreshedAt as any) instanceof Date) {
                quotaRefreshedAt = (quotaRefreshedAt as any).toISOString();
            } else if (typeof quotaRefreshedAt === 'number') {
                quotaRefreshedAt = new Date(quotaRefreshedAt).toISOString();
            } else if (quotaRefreshedAt === null) {
                quotaRefreshedAt = new Date().toISOString();
            }

            this.snapshots.set(snapshot.provider, {
                ...existing,
                ...snapshot,
                quotaConfidence: 'real-time',
                quotaRefreshedAt,
                windows: snapshot.windows ?? existing?.windows,
                source: snapshot.source ?? existing?.source ?? 'balance',
                connectionId: snapshot.connectionId ?? existing?.connectionId ?? null,
            });
        }
    }

    public getQuota(provider: string): ProviderQuotaSnapshot | undefined {
        this.refreshProviderAvailability(provider);
        return this.snapshots.get(provider);
    }

    public getAllQuotas(): ProviderQuotaSnapshot[] {
        for (const provider of this.snapshots.keys()) {
            this.refreshProviderAvailability(provider);
        }
        return Array.from(this.snapshots.values());
    }

    public markRateLimited(provider: string, retryAfter: number | Date = Date.now() + DEFAULT_RETRY_MS, rateLimitRpm?: number, message = 'Provider returned a rate-limit response.') {
        this.updateAvailability(provider, 'rate_limited', retryAfter, message, rateLimitRpm);
    }

    public markQuotaExceeded(provider: string, resetDate: number | Date = Date.now() + DEFAULT_RETRY_MS, message = 'Provider quota exhausted.') {
        this.updateAvailability(provider, 'quota_exhausted', resetDate, message);
    }

    public markProviderHealthy(provider: string) {
        const snapshot = this.snapshots.get(provider);
        if (!snapshot) {
            return;
        }

        this.revokedProviders.delete(provider);
        const standardAuthState = this.registry.getAuthStates(this.lastEnv).find(s => s.provider === provider);

        this.snapshots.set(provider, {
            ...snapshot,
            authTruth: standardAuthState?.authTruth ?? 'authenticated',
            availability: standardAuthState?.authenticated ? 'available' : 'missing_auth',
            retryAfter: null,
            lastError: undefined,
        });
    }

    public isProviderReady(provider: string): boolean {
        this.refreshProviderAvailability(provider);
        const snapshot = this.snapshots.get(provider);
        return snapshot?.authenticated === true && snapshot.availability === 'available';
    }

    private updateAvailability(
        provider: string,
        availability: Extract<ProviderAvailability, 'rate_limited' | 'quota_exhausted' | 'cooldown'>,
        retryAfter: number | Date,
        message: string,
        rateLimitRpm?: number,
    ) {
        const snapshot = this.snapshots.get(provider);
        if (!snapshot) {
            return;
        }

        const retryDate = typeof retryAfter === 'number' ? new Date(retryAfter) : retryAfter;
        this.snapshots.set(provider, {
            ...snapshot,
            availability,
            retryAfter: retryDate.toISOString(),
            resetDate: availability === 'quota_exhausted' ? retryDate.toISOString() : snapshot.resetDate,
            rateLimitRpm: rateLimitRpm ?? snapshot.rateLimitRpm,
            lastError: message,
        });
    }

    private refreshProviderAvailability(provider: string) {
        const snapshot = this.snapshots.get(provider);
        if (!snapshot?.retryAfter) {
            return;
        }

        if (Date.now() >= new Date(snapshot.retryAfter).getTime()) {
            this.markProviderHealthy(provider);
        }
    }
}
