export type ProviderAuthMethod = 'api_key' | 'oauth' | 'pat' | 'none';
export type ProviderRoutingStrategy = 'cheapest' | 'best' | 'round-robin';
export type ProviderTaskType = 'coding' | 'planning' | 'research' | 'general' | 'worker' | 'supervisor';
export type ProviderAvailability = 'available' | 'rate_limited' | 'quota_exhausted' | 'cooldown' | 'missing_auth';
export type ProviderCapability = 'coding' | 'reasoning' | 'vision' | 'tools' | 'long_context';
export type ProviderBalanceProvider = 'github' | 'claude' | 'codex' | 'antigravity' | 'kiro' | 'kimi-coding';
export type ProviderQuotaUnit = 'percent' | 'requests' | 'tokens' | 'credits' | 'unknown';

export interface ProviderModelDefinition {
    id: string;
    provider: string;
    name: string;
    inputPrice: number | null;
    outputPrice: number | null;
    contextWindow: number | null;
    tier: string;
    recommendedFor?: ProviderTaskType[];
    capabilities: ProviderCapability[];
    executable?: boolean;
    qualityScore?: number;
}

export interface ProviderDefinition {
    id: string;
    name: string;
    authMethod: ProviderAuthMethod;
    envKeys?: string[];
    oauthEnvKeys?: string[];
    patEnvKeys?: string[];
    executable?: boolean;
    defaultModel: string;
    models: ProviderModelDefinition[];
    preferredTasks?: ProviderTaskType[];
}

export type ProviderAuthTruth = 'not_configured' | 'authenticated' | 'expired' | 'revoked';

export interface ProviderAuthState {
    provider: string;
    name: string;
    authMethod: ProviderAuthMethod;
    configured: boolean;
    authenticated: boolean;
    detail: string;
    authTruth?: ProviderAuthTruth;
}

export interface ProviderQuotaWindowSnapshot {
    key: string;
    label: string;
    used: number;
    limit: number | null;
    remaining: number | null;
    resetDate: string | null;
    unit: ProviderQuotaUnit;
}

export interface ProviderQuotaSnapshot extends ProviderAuthState {
    used: number;
    limit: number | null;
    remaining: number | null;
    resetDate: string | null;
    rateLimitRpm: number | null;
    tier: string;
    availability: ProviderAvailability;
    lastError?: string;
    retryAfter?: string | null;
    windows?: ProviderQuotaWindowSnapshot[];
    source?: 'runtime' | 'balance';
    connectionId?: string | null;
    quotaConfidence?: 'real-time' | 'cached' | 'estimated' | 'unknown';
    quotaRefreshedAt?: string | null;
}

export interface ProviderBalanceConnection {
    id: string;
    provider: ProviderBalanceProvider;
    authMethod: Extract<ProviderAuthMethod, 'oauth' | 'pat'>;
    accessToken: string;
    refreshToken?: string | null;
    expiresAt?: string | null;
    accountLabel?: string | null;
    metadata?: Record<string, unknown>;
}

export interface ProviderBalanceConnectionSource {
    getConnection(provider: ProviderBalanceProvider): Promise<ProviderBalanceConnection | null>;
}

export interface ProviderBalanceTokenRefresher {
    refreshConnection(connection: ProviderBalanceConnection): Promise<ProviderBalanceConnection>;
}

export interface ProviderQuotaProvider {
    provider: ProviderBalanceProvider;
    billingProvider: string;
    name: string;
    fetchQuotaSnapshot(connection: ProviderBalanceConnection): Promise<ProviderQuotaSnapshot>;
}

export interface RoutingSelectionRequest {
    provider?: string;
    taskComplexity?: 'low' | 'medium' | 'high';
    taskType?: 'worker' | 'supervisor';
    routingTaskType?: ProviderTaskType;
    routingStrategy?: ProviderRoutingStrategy;
    exclude?: string[];
}

export interface FallbackCandidateSnapshot {
    id: string;
    provider: string;
    model: string;
    name: string;
    inputPrice: number | null;
    outputPrice: number | null;
    contextWindow: number | null;
    tier: string;
    recommended: boolean;
    reason?: string;
}
