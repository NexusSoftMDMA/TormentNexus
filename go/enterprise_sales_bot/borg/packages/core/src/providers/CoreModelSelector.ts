import { ModelSelector, type ModelSelectionRequest, type SelectedModel } from '@tormentnexus/ai';

import { ProviderBalanceService } from './ProviderBalanceService.js';
import { ProviderRegistry } from './ProviderRegistry.js';
import { NormalizedQuotaService } from './NormalizedQuotaService.js';
import {
    type FallbackCandidateSnapshot,
    type ProviderModelDefinition,
    type ProviderRoutingStrategy,
    type ProviderTaskType,
    type RoutingSelectionRequest,
} from './types.js';

const DEFAULT_TASK_STRATEGIES: Record<ProviderTaskType, ProviderRoutingStrategy> = {
    coding: 'cheapest',
    planning: 'best',
    research: 'best',
    general: 'round-robin',
    worker: 'best',
    supervisor: 'best',
};

const DEFAULT_PROVIDER_LIMITS: Record<string, number> = {
    google: 5,
    anthropic: 10,
    openai: 10,
    deepseek: 5,
};

/** Maximum number of fallback events retained in the in-process ring buffer. */
const MAX_FALLBACK_EVENTS = 50;

/**
 * A single provider-selection decision recorded whenever `selectModel` does NOT
 * honor the caller's preferred provider (or uses a budget/emergency override).
 * These snapshots power the "Recent Fallback Decisions" card on the billing dashboard.
 */
export interface ProviderFallbackEvent {
    /** Monotonic event ID (counter, not UUID) for stable list keys in the UI. */
    id: number;
    /** Unix milliseconds when the decision was made. */
    timestamp: number;
    /** Provider requested by the caller, or undefined if none was specified. */
    requestedProvider?: string;
    /** Provider that was actually selected. */
    selectedProvider: string;
    /** Model ID that was selected. */
    selectedModelId: string;
    /** Task type that drove the routing decision. */
    taskType: ProviderTaskType;
    /** Routing strategy that was active at decision time. */
    strategy: ProviderRoutingStrategy;
    /**
     * Human-readable reason string from the selector (e.g. `TASK_TYPE_CODING`,
     * `BUDGET_EXCEEDED_FORCED_LOCAL`, `EMERGENCY_FALLBACK`).
     */
    reason: string;
    /**
     * Structured cause code for programmatic triage:
     * - `preference_honored`: selected provider matches requested provider (this event is
     *   recorded only for completeness in the first-selection case).
     * - `fallback_provider`: caller preferred a different provider but it was unavailable.
     * - `budget_forced_local`: daily budget exceeded, forced route to local provider.
     * - `emergency_fallback`: no executable candidates were available at all.
     */
    causeCode: 'preference_honored' | 'fallback_provider' | 'budget_forced_local' | 'emergency_fallback';
}

interface SelectorOptions {
    registry?: ProviderRegistry;
    quotaService?: NormalizedQuotaService;
    balanceService?: ProviderBalanceService;
    routingStrategy?: ProviderRoutingStrategy;
    taskStrategies?: Partial<Record<ProviderTaskType, ProviderRoutingStrategy>>;
}

export class CoreModelSelector extends ModelSelector {
    private readonly registry: ProviderRegistry;
    private readonly quotaTracker: NormalizedQuotaService;
    private readonly taskStrategies: Record<ProviderTaskType, ProviderRoutingStrategy>;
    private defaultRoutingStrategy: ProviderRoutingStrategy;
    private readonly candidateCooldowns = new Map<string, number>();
    private readonly roundRobinOffsets = new Map<ProviderTaskType, number>();
    /** Append-only ring buffer of provider fallback events (capped to MAX_FALLBACK_EVENTS). */
    private readonly fallbackEventBuffer: ProviderFallbackEvent[] = [];
    private fallbackEventCounter = 0;

    constructor(options: SelectorOptions = {}) {
        super();
        this.registry = options.registry ?? new ProviderRegistry();
        this.quotaTracker = options.quotaService ?? new NormalizedQuotaService(this.registry, options.balanceService);
        this.quotaTracker.setConfig({
            dailyBudgetUsd: 5,
            monthlyBudgetUsd: 100,
            providerLimits: DEFAULT_PROVIDER_LIMITS,
        });
        this.defaultRoutingStrategy = options.routingStrategy ?? 'best';
        this.taskStrategies = {
            ...DEFAULT_TASK_STRATEGIES,
            ...(options.taskStrategies ?? {}),
        };
    }

    public getQuotaService(): NormalizedQuotaService {
        this.quotaTracker.refreshAuthStates();
        return this.quotaTracker;
    }

    public setRoutingStrategy(strategy: ProviderRoutingStrategy) {
        this.defaultRoutingStrategy = strategy;
    }

    public getRoutingStrategy(): ProviderRoutingStrategy {
        return this.defaultRoutingStrategy;
    }

    public setTaskRoutingStrategy(taskType: ProviderTaskType, strategy: ProviderRoutingStrategy) {
        this.taskStrategies[taskType] = strategy;
    }

    public resetTaskRoutingStrategy(taskType: ProviderTaskType) {
        this.taskStrategies[taskType] = DEFAULT_TASK_STRATEGIES[taskType];
    }

    public getTaskRoutingRules(): Record<ProviderTaskType, ProviderRoutingStrategy> {
        return { ...this.taskStrategies };
    }

    public async getProviderSnapshots() {
        await this.quotaTracker.refreshProviderBalances();
        return this.quotaTracker.getAllQuotas();
    }

    public getAvailableModels(): FallbackCandidateSnapshot[] {
        return this.registry.buildSnapshots('general').map((candidate) => ({
            ...candidate,
            reason: candidate.recommended ? 'recommended' : 'available',
        }));
    }

    public getFallbackChain(request: Partial<RoutingSelectionRequest> = {}): FallbackCandidateSnapshot[] {
        const taskType = this.resolveTaskType(request);
        const strategy = request.routingStrategy ?? this.taskStrategies[taskType] ?? this.defaultRoutingStrategy;
        return this.rankCandidates(taskType, strategy, request.provider, new Set(request.exclude ?? []), false).map((candidate) => ({
            id: candidate.id,
            provider: candidate.provider,
            model: candidate.id,
            name: candidate.name,
            inputPrice: candidate.inputPrice,
            outputPrice: candidate.outputPrice,
            contextWindow: candidate.contextWindow,
            tier: candidate.tier,
            recommended: this.registry.isRecommendedForTask(candidate, taskType),
            reason: this.buildReason(candidate, taskType, strategy, request.provider),
        }));
    }

    public async selectModel(request: ModelSelectionRequest & Partial<RoutingSelectionRequest> = {}): Promise<SelectedModel> {
        const taskType = this.resolveTaskType(request);
        const strategy = request.routingStrategy ?? this.taskStrategies[taskType] ?? this.defaultRoutingStrategy;

        if (this.quotaTracker.isBudgetExceeded()) {
            const forcedLocal = this.rankCandidates(taskType, 'cheapest', undefined, new Set(), false)
                .find((candidate) => candidate.provider === 'lmstudio' || candidate.provider === 'ollama');
            if (forcedLocal) {
                const result: SelectedModel = {
                    provider: forcedLocal.provider,
                    modelId: forcedLocal.id,
                    reason: 'BUDGET_EXCEEDED_FORCED_LOCAL',
                };
                this.recordFallbackEvent(result, taskType, strategy, request.provider, 'budget_forced_local');
                return result;
            }
        }

        const candidates = this.rankCandidates(taskType, strategy, request.provider, new Set(request.exclude ?? []), true);
        const selected = candidates[0];

        if (!selected) {
            // Cascade fallback: try any other executable models in the catalog first
            const allAvailable = this.registry.listExecutableModels();
            const activeFallback = allAvailable.find(c => c.provider === 'openrouter' || c.provider === 'google');
            if (activeFallback) {
                const result: SelectedModel = {
                    provider: activeFallback.provider,
                    modelId: activeFallback.id,
                    reason: 'EMERGENCY_CATALOG_CASCADE',
                };
                this.recordFallbackEvent(result, taskType, strategy, request.provider, 'fallback_provider');
                return result;
            }

            const fallbackModelId = this.registry.getProvider('lmstudio')?.defaultModel || 'local';
            const result: SelectedModel = {
                provider: 'lmstudio',
                modelId: fallbackModelId,
                reason: 'EMERGENCY_FALLBACK',
            };
            this.recordFallbackEvent(result, taskType, strategy, request.provider, 'emergency_fallback');
            return result;
        }

        this.quotaTracker.markProviderHealthy(selected.provider);
        const reason = request.provider === selected.provider
            ? 'PROVIDER_PREFERENCE'
            : this.buildReason(selected, taskType, strategy, request.provider);
        const result: SelectedModel = {
            provider: selected.provider,
            modelId: selected.id,
            reason,
        };
        // Record non-trivial events: actual fallbacks (requested provider not honored) and
        // first-selection summaries when no provider preference was expressed.
        const causeCode = request.provider && request.provider !== selected.provider
            ? 'fallback_provider'
            : 'preference_honored';
        if (causeCode === 'fallback_provider') {
            this.recordFallbackEvent(result, taskType, strategy, request.provider, causeCode);
        }
        return result;
    }

    /**
     * Returns the last N fallback events in reverse-chronological order.
     * Events only include decisions where the preferred provider was not honored,
     * plus budget/emergency overrides.
     */
    public getFallbackHistory(limit = 20): ProviderFallbackEvent[] {
        return this.fallbackEventBuffer.slice(-Math.max(1, limit)).reverse();
    }

    /** Clears the in-process fallback event ring buffer. */
    public clearFallbackHistory(): void {
        this.fallbackEventBuffer.length = 0;
        this.fallbackEventCounter = 0;
    }

    /** Appends a fallback event to the ring buffer, evicting oldest when full. */
    private recordFallbackEvent(
        result: SelectedModel,
        taskType: ProviderTaskType,
        strategy: ProviderRoutingStrategy,
        requestedProvider: string | undefined,
        causeCode: ProviderFallbackEvent['causeCode'],
    ): void {
        const event: ProviderFallbackEvent = {
            id: ++this.fallbackEventCounter,
            timestamp: Date.now(),
            requestedProvider: requestedProvider ?? undefined,
            selectedProvider: result.provider,
            selectedModelId: result.modelId,
            taskType,
            strategy,
            reason: result.reason,
            causeCode,
        };
        this.fallbackEventBuffer.push(event);
        // Evict oldest entries when the buffer exceeds the cap.
        if (this.fallbackEventBuffer.length > MAX_FALLBACK_EVENTS) {
            this.fallbackEventBuffer.splice(0, this.fallbackEventBuffer.length - MAX_FALLBACK_EVENTS);
        }
    }

    /**
     * Returns a snapshot of currently blocked or cooling-down provider/model entries.
     *
     * Overrides the base class which reads from internal `modelStates` (never set here).
     * Instead, we combine:
     * 1. Per-(provider, model) cooldowns from `reportFailure` transient backoff.
     * 2. Provider-wide quota states (`rate_limited`, `quota_exhausted`, `cooldown`) from
     *    `NormalizedQuotaService`.
     *
     * The billing dashboard "Blocked / Cooling-Down Models" card consumes this via
     * `trpc.billing.getDepletedModels`.
     */
    public getDepletedModels(): Array<{
        key: string;
        provider: string;
        modelId: string;
        depletedAt: number;
        retryAfter: number;
        isPermanent: boolean;
        coolsDownAt: string;
    }> {
        const now = Date.now();
        const result: ReturnType<CoreModelSelector['getDepletedModels']> = [];
        const coveredProviders = new Set<string>();

        // Per-model transient cooldowns from failed individual calls.
        for (const [key, cooldownUntil] of this.candidateCooldowns) {
            if (cooldownUntil <= now) continue; // already expired, will be cleaned up on next isCandidateReady check
            const separatorIdx = key.indexOf(':');
            const provider = separatorIdx !== -1 ? key.slice(0, separatorIdx) : key;
            const modelId = separatorIdx !== -1 ? key.slice(separatorIdx + 1) : '';
            coveredProviders.add(provider);
            result.push({
                key,
                provider,
                modelId,
                depletedAt: cooldownUntil - 60_000, // approximate start time (60s backoff)
                retryAfter: cooldownUntil,
                isPermanent: false,
                coolsDownAt: new Date(cooldownUntil).toLocaleTimeString(),
            });
        }

        // Provider-wide quota states (rate_limited, quota_exhausted, cooldown).
        for (const snapshot of this.quotaTracker.getAllQuotas()) {
            const { availability, provider, retryAfter } = snapshot;
            if (availability === 'available' || availability === 'missing_auth') continue;
            // Avoid duplicating a provider already represented by a per-model entry.
            if (coveredProviders.has(provider)) continue;
            const retryAfterMs = retryAfter ? new Date(retryAfter).getTime() : Infinity;
            const isPermanent = retryAfterMs === Infinity;
            result.push({
                key: `${provider}:quota`,
                provider,
                modelId: '*',
                depletedAt: now,
                retryAfter: retryAfterMs,
                isPermanent,
                coolsDownAt: isPermanent
                    ? 'session end'
                    : retryAfter
                        ? new Date(retryAfter).toLocaleTimeString()
                        : 'unknown',
            });
        }

        return result;
    }

    public reportFailure(provider: string, modelId: string, cause?: unknown) {
        const cooldownUntil = Date.now() + 60_000;
        this.candidateCooldowns.set(this.getCooldownKey(provider, modelId), cooldownUntil);

        const message = this.getFailureMessage(cause);
        const status = this.getFailureStatus(cause);

        // Auth revocation — credential is present but rejected by the provider.
        // Mark the provider as revoked and remove it from routing until the key is rotated.
        if (status === 401 || status === 403) {
            this.quotaTracker.markAuthRevoked(provider, message || `Provider rejected credential (HTTP ${status}).`);
            return;
        }

        if (status === 429 || message.includes('rate limit')) {
            this.quotaTracker.markRateLimited(provider, cooldownUntil, undefined, message || 'Provider rate limited.');
            return;
        }

        if (message.includes('quota') || message.includes('insufficient_quota') || message.includes('resource_exhausted')) {
            this.quotaTracker.markQuotaExceeded(provider, cooldownUntil, message || 'Provider quota exhausted.');
            return;
        }

        this.quotaTracker.markRateLimited(provider, cooldownUntil, undefined, message || 'Provider temporarily unavailable.');
    }

    private rankCandidates(
        taskType: ProviderTaskType,
        strategy: ProviderRoutingStrategy,
        preferredProvider: string | undefined,
        excluded: Set<string>,
        advanceRoundRobin: boolean,
    ): ProviderModelDefinition[] {
        const available = this.registry.listExecutableModels().filter((candidate) => {
            const candidateKey = this.getCooldownKey(candidate.provider, candidate.id);
            if (excluded.has(candidateKey)) {
                return false;
            }
            if (!this.isCandidateReady(candidate.provider, candidate.id)) {
                return false;
            }
            return true;
        });

        const withPreferred = [...available].sort((left, right) => {
            const preferredLeft = preferredProvider && left.provider === preferredProvider ? 1 : 0;
            const preferredRight = preferredProvider && right.provider === preferredProvider ? 1 : 0;
            if (preferredLeft !== preferredRight) {
                return preferredRight - preferredLeft;
            }
            return 0;
        });

        if (strategy === 'round-robin') {
            const ordered = withPreferred.sort((left, right) => this.compareByTaskFit(left, right, taskType) || left.provider.localeCompare(right.provider) || left.id.localeCompare(right.id));
            const offset = this.roundRobinOffsets.get(taskType) ?? 0;
            const rotated = ordered.length === 0 ? [] : ordered.slice(offset).concat(ordered.slice(0, offset));
            if (advanceRoundRobin && ordered.length > 0) {
                this.roundRobinOffsets.set(taskType, (offset + 1) % ordered.length);
            }
            return rotated;
        }

        const comparator = strategy === 'cheapest'
            ? (left: ProviderModelDefinition, right: ProviderModelDefinition) => this.compareByTaskFit(left, right, taskType) || this.compareByPrice(left, right) || this.compareByQuality(right, left) || left.provider.localeCompare(right.provider) || left.id.localeCompare(right.id)
            : (left: ProviderModelDefinition, right: ProviderModelDefinition) => this.compareByTaskFit(left, right, taskType) || this.compareByQuality(left, right) || this.compareByContext(left, right) || this.compareByPrice(left, right) || left.provider.localeCompare(right.provider) || left.id.localeCompare(right.id);

        return withPreferred.sort(comparator);
    }

    private resolveTaskType(request: Partial<RoutingSelectionRequest>): ProviderTaskType {
        if (request.routingTaskType) {
            return request.routingTaskType;
        }
        if (request.taskType === 'supervisor') {
            return 'planning';
        }
        if (request.taskComplexity === 'high') {
            return 'planning';
        }
        if (request.taskComplexity === 'low') {
            return 'coding';
        }
        return 'general';
    }

    private isCandidateReady(provider: string, modelId: string): boolean {
        const candidateKey = this.getCooldownKey(provider, modelId);
        const cooldown = this.candidateCooldowns.get(candidateKey);
        if (cooldown && cooldown > Date.now()) {
            return false;
        }
        if (cooldown && cooldown <= Date.now()) {
            this.candidateCooldowns.delete(candidateKey);
            this.quotaTracker.markProviderHealthy(provider);
        }
        return this.quotaTracker.isProviderReady(provider);
    }

    private compareByTaskFit(left: ProviderModelDefinition, right: ProviderModelDefinition, taskType: ProviderTaskType): number {
        const leftFit = this.registry.isRecommendedForTask(left, taskType) ? 1 : 0;
        const rightFit = this.registry.isRecommendedForTask(right, taskType) ? 1 : 0;
        return rightFit - leftFit;
    }

    private compareByPrice(left: ProviderModelDefinition, right: ProviderModelDefinition): number {
        const leftPrice = (left.inputPrice ?? 0) + (left.outputPrice ?? 0);
        const rightPrice = (right.inputPrice ?? 0) + (right.outputPrice ?? 0);
        return leftPrice - rightPrice;
    }

    private compareByQuality(left: ProviderModelDefinition, right: ProviderModelDefinition): number {
        return (right.qualityScore ?? 0) - (left.qualityScore ?? 0);
    }

    private compareByContext(left: ProviderModelDefinition, right: ProviderModelDefinition): number {
        return (right.contextWindow ?? 0) - (left.contextWindow ?? 0);
    }

    private buildReason(candidate: ProviderModelDefinition, taskType: ProviderTaskType, strategy: ProviderRoutingStrategy, preferredProvider?: string): string {
        if (preferredProvider && candidate.provider === preferredProvider) {
            return 'PROVIDER_PREFERENCE';
        }
        if (this.registry.isRecommendedForTask(candidate, taskType)) {
            return `TASK_TYPE_${taskType.toUpperCase()}`;
        }
        return strategy.toUpperCase().replace('-', '_');
    }

    private getCooldownKey(provider: string, modelId: string): string {
        return `${provider}:${modelId}`;
    }

    private getFailureStatus(cause: unknown): number | undefined {
        if (cause && typeof cause === 'object' && 'status' in cause && typeof (cause as { status?: unknown }).status === 'number') {
            return (cause as { status: number }).status;
        }
        return undefined;
    }

    private getFailureMessage(cause: unknown): string {
        if (cause instanceof Error) {
            return cause.message.toLowerCase();
        }
        if (typeof cause === 'string') {
            return cause.toLowerCase();
        }
        if (cause && typeof cause === 'object' && 'message' in cause && typeof (cause as { message?: unknown }).message === 'string') {
            return (cause as { message: string }).message.toLowerCase();
        }
        return '';
    }
}
