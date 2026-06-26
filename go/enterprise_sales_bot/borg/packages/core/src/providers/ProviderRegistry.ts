import { DEFAULT_OPENROUTER_FREE_MODEL } from '@tormentnexus/ai';

import {
    type FallbackCandidateSnapshot,
    type ProviderAuthState,
    type ProviderAuthTruth,
    type ProviderDefinition,
    type ProviderModelDefinition,
    type ProviderTaskType,
    type ProviderCapability,
} from './types.js';

export const DEFAULT_PROVIDER_CATALOG: ProviderDefinition[] = [
    {
        id: 'google',
        name: 'Google Gemini',
        authMethod: 'api_key',
        envKeys: ['GOOGLE_API_KEY', 'GEMINI_API_KEY'],
        executable: true,
        defaultModel: 'gemini-2.5-flash',
        preferredTasks: ['coding', 'research'],
        models: [
            {
                id: 'gemini-2.5-flash',
                provider: 'google',
                name: 'Gemini 2.5 Flash',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 1_000_000,
                tier: 'free',
                recommendedFor: ['coding', 'worker', 'general', 'research'],
                capabilities: ['coding', 'vision', 'tools', 'long_context'],
                executable: true,
                qualityScore: 8,
            },
            {
                id: 'gemini-2.0-flash',
                provider: 'google',
                name: 'Gemini 2.0 Flash',
                inputPrice: 0.0001,
                outputPrice: 0.0004,
                contextWindow: 1_000_000,
                tier: 'standard',
                recommendedFor: ['coding', 'worker', 'general'],
                capabilities: ['coding', 'vision', 'tools', 'long_context'],
                executable: true,
                qualityScore: 9,
            },
            {
                id: 'gemini-2.0-flash-lite-preview-02-05',
                provider: 'google',
                name: 'Gemini 2.0 Flash Lite',
                inputPrice: 0.000075,
                outputPrice: 0.0003,
                contextWindow: 1_000_000,
                tier: 'free',
                recommendedFor: ['worker', 'general'],
                capabilities: ['coding', 'long_context'],
                executable: true,
                qualityScore: 7,
            },
            {
                id: 'gemini-1.5-pro',
                provider: 'google',
                name: 'Gemini 1.5 Pro',
                inputPrice: 0.0035,
                outputPrice: 0.0105,
                contextWindow: 2_000_000,
                tier: 'premium',
                recommendedFor: ['planning', 'research', 'supervisor'],
                capabilities: ['reasoning', 'vision', 'long_context', 'tools'],
                executable: true,
                qualityScore: 8,
            },
        ],
    },
    {
        id: 'anthropic',
        name: 'Anthropic',
        authMethod: 'api_key',
        envKeys: ['ANTHROPIC_API_KEY'],
        executable: true,
        defaultModel: 'claude-3-7-sonnet-20250219',
        preferredTasks: ['planning', 'research'],
        models: [
            {
                id: 'claude-3-7-sonnet-20250219',
                provider: 'anthropic',
                name: 'Claude 3.7 Sonnet',
                inputPrice: 0.003,
                outputPrice: 0.015,
                contextWindow: 200_000,
                tier: 'premium',
                recommendedFor: ['planning', 'research', 'supervisor', 'coding'],
                capabilities: ['reasoning', 'coding', 'tools', 'long_context'],
                executable: true,
                qualityScore: 10,
            },
            {
                id: 'claude-3-5-haiku-20241022',
                provider: 'anthropic',
                name: 'Claude 3.5 Haiku',
                inputPrice: 0.00025,
                outputPrice: 0.00125,
                contextWindow: 200_000,
                tier: 'standard',
                recommendedFor: ['worker', 'coding'],
                capabilities: ['coding', 'tools'],
                executable: true,
                qualityScore: 7,
            },
        ],
    },
    {
        id: 'openai',
        name: 'OpenAI',
        authMethod: 'api_key',
        envKeys: ['OPENAI_API_KEY'],
        executable: true,
        defaultModel: 'gpt-4o',
        preferredTasks: ['planning', 'coding'],
        models: [
            {
                id: 'gpt-4o',
                provider: 'openai',
                name: 'GPT-4o',
                inputPrice: 0.005,
                outputPrice: 0.015,
                contextWindow: 128_000,
                tier: 'premium',
                recommendedFor: ['planning', 'supervisor'],
                capabilities: ['reasoning', 'coding', 'vision', 'tools'],
                executable: true,
                qualityScore: 9,
            },
            {
                id: 'gpt-4o-mini',
                provider: 'openai',
                name: 'GPT-4o Mini',
                inputPrice: 0.00015,
                outputPrice: 0.0006,
                contextWindow: 128_000,
                tier: 'standard',
                recommendedFor: ['coding', 'general', 'worker'],
                capabilities: ['coding', 'tools'],
                executable: true,
                qualityScore: 6,
            },
        ],
    },
    {
        id: 'deepseek',
        name: 'DeepSeek',
        authMethod: 'api_key',
        envKeys: ['DEEPSEEK_API_KEY'],
        executable: true,
        defaultModel: 'deepseek-chat',
        preferredTasks: ['coding', 'research'],
        models: [
            {
                id: 'deepseek-chat',
                provider: 'deepseek',
                name: 'DeepSeek V3',
                inputPrice: 0.00014,
                outputPrice: 0.00028,
                contextWindow: 64_000,
                tier: 'standard',
                recommendedFor: ['coding', 'general', 'worker'],
                capabilities: ['coding', 'reasoning'],
                executable: true,
                qualityScore: 9,
            },
            {
                id: 'deepseek-reasoner',
                provider: 'deepseek',
                name: 'DeepSeek R1',
                inputPrice: 0.00055,
                outputPrice: 0.00219,
                contextWindow: 64_000,
                tier: 'premium',
                recommendedFor: ['planning', 'research'],
                capabilities: ['reasoning', 'coding'],
                executable: true,
                qualityScore: 10,
            },
        ],
    },
    {
        id: 'openrouter',
        name: 'OpenRouter',
        authMethod: 'api_key',
        envKeys: ['OPENROUTER_API_KEY'],
        executable: true,
        defaultModel: 'openrouter/free',
        preferredTasks: ['coding', 'research', 'general', 'worker'],
        models: [
            {
                id: 'openrouter/free',
                provider: 'openrouter',
                name: 'OpenRouter Free (Auto Router)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: null,
                tier: 'free',
                recommendedFor: ['coding', 'research', 'general', 'worker'],
                capabilities: ['reasoning', 'coding', 'tools'],
                executable: true,
                qualityScore: 9,
            },
            {
                id: 'google/gemini-2.0-flash-lite-preview-02-05:free',
                provider: 'openrouter',
                name: 'Gemini 2.0 Flash Lite (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 1048576,
                tier: 'free',
                recommendedFor: ['worker'],
                capabilities: ['tools', 'long_context'],
                executable: true,
                qualityScore: 7,
            },
            {
                id: 'google/gemini-2.0-flash-exp:free',
                provider: 'openrouter',
                name: 'Gemini 2.0 Flash (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 1048576,
                tier: 'free',
                recommendedFor: ['worker'],
                capabilities: ['tools', 'long_context'],
                executable: true,
                qualityScore: 8,
            },
            {
                id: 'meta-llama/llama-3.3-70b-instruct:free',
                provider: 'openrouter',
                name: 'Llama 3.3 70B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 131072,
                tier: 'free',
                recommendedFor: ['worker'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 7,
            },
            {
                id: 'openrouter/nousresearch/hermes-3-llama-3.1-405b:free',
                provider: 'openrouter',
                name: 'Hermes 3 Llama 3.1 405B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 131072,
                tier: 'free',
                recommendedFor: ['worker', 'general', 'planning'],
                capabilities: ['reasoning', 'coding'],
                executable: true,
                qualityScore: 8,
            },
            {
                id: 'openrouter/liquid/lfm-2.5-1.2b-instruct:free',
                provider: 'openrouter',
                name: 'LFM 2.5 1.2B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 32000,
                tier: 'free',
                recommendedFor: ['worker'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 6,
            },
            {
                id: 'openrouter/qwen/qwen3-coder:free',
                provider: 'openrouter',
                name: 'Qwen 3 Coder (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 32000,
                tier: 'free',
                recommendedFor: ['worker', 'coding'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 7,
            },
            {
                id: 'openrouter/deepseek/deepseek-v4-flash:free',
                provider: 'openrouter',
                name: 'DeepSeek V4 Flash (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 64000,
                tier: 'free',
                recommendedFor: ['worker', 'coding', 'general'],
                capabilities: ['coding', 'reasoning'],
                executable: true,
                qualityScore: 8,
            },
            {
                id: 'openrouter/google/gemma-2-9b-it:free',
                provider: 'openrouter',
                name: 'Gemma 2 9B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 8192,
                tier: 'free',
                recommendedFor: ['worker', 'general'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 6,
            },
            {
                id: 'openrouter/meta-llama/llama-3-8b-instruct:free',
                provider: 'openrouter',
                name: 'Llama 3 8B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 8192,
                tier: 'free',
                recommendedFor: ['worker'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 6,
            },
            {
                id: 'openrouter/microsoft/phi-3-medium-128k-instruct:free',
                provider: 'openrouter',
                name: 'Phi 3 Medium 128K (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 128000,
                tier: 'free',
                recommendedFor: ['worker', 'general'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 6,
            },
            {
                id: 'openrouter/mistralai/mistral-7b-instruct:free',
                provider: 'openrouter',
                name: 'Mistral 7B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 32000,
                tier: 'free',
                recommendedFor: ['worker'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 6,
            },
            {
                id: 'openrouter/openai/gpt-oss-120b:free',
                provider: 'openrouter',
                name: 'GPT OSS 120B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 131072,
                tier: 'free',
                recommendedFor: ['worker', 'general'],
                capabilities: ['coding', 'reasoning'],
                executable: true,
                qualityScore: 7,
            },
            {
                id: 'openrouter/nvidia/nemotron-3-super-120b-a12b:free',
                provider: 'openrouter',
                name: 'Nemotron 3 Super 120B (Free)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 1000000,
                tier: 'free',
                recommendedFor: ['worker', 'general'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 7,
            },
            {
                id: 'openrouter/auto',
                provider: 'openrouter',
                name: 'OpenRouter Auto (Best Overall)',
                inputPrice: null,
                outputPrice: null,
                contextWindow: null,
                tier: 'meta',
                capabilities: ['reasoning', 'coding', 'vision', 'tools'],
                executable: true,
                qualityScore: 5,
            },
            {
                id: 'openrouter/best-available-coding',
                provider: 'openrouter',
                name: 'Best Available Coding (Meta)',
                inputPrice: null,
                outputPrice: null,
                contextWindow: null,
                tier: 'meta',
                capabilities: ['coding', 'tools'],
                executable: true,
                qualityScore: 10,
            },
        ],
    },
    {
        id: 'copilot',
        name: 'GitHub Copilot',
        authMethod: 'pat',
        patEnvKeys: ['COPILOT_PAT', 'GITHUB_TOKEN'],
        executable: false,
        defaultModel: 'copilot/gpt-4.1',
        models: [
            {
                id: 'copilot/gpt-4.1',
                provider: 'copilot',
                name: 'Copilot GPT-4.1',
                inputPrice: null,
                outputPrice: null,
                contextWindow: null,
                tier: 'managed',
                capabilities: ['coding', 'tools'],
                qualityScore: 7,
            },
        ],
    },
    {
        id: 'google-oauth',
        name: 'Google OAuth',
        authMethod: 'oauth',
        oauthEnvKeys: ['GOOGLE_OAUTH_ACCESS_TOKEN'],
        executable: false,
        defaultModel: 'google-oauth/gemini',
        models: [
            {
                id: 'google-oauth/gemini',
                provider: 'google-oauth',
                name: 'Google OAuth Gemini',
                inputPrice: null,
                outputPrice: null,
                contextWindow: null,
                tier: 'managed',
                capabilities: ['reasoning', 'vision'],
                qualityScore: 6,
            },
        ],
    },
    {
        id: 'lmstudio',
        name: 'LM Studio',
        authMethod: 'none',
        executable: true,
        defaultModel: 'C:/Users/hyper/.lmstudio/models/HauhauCS/Gemma-4-E2B-Uncensored-HauhauCS-Aggressive/Gemma-4-E2B-Uncensored-HauhauCS-Aggressive-Q2_K_P.gguf gemma-4-e2b-uncensored-hauhaucs-aggressive',
        preferredTasks: ['general', 'worker'],
        models: [
            {
                id: 'C:/Users/hyper/.lmstudio/models/HauhauCS/Gemma-4-E2B-Uncensored-HauhauCS-Aggressive/Gemma-4-E2B-Uncensored-HauhauCS-Aggressive-Q2_K_P.gguf gemma-4-e2b-uncensored-hauhaucs-aggressive',
                provider: 'lmstudio',
                name: 'Gemma 4 Aggressive (Utility)',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 32000,
                tier: 'local',
                recommendedFor: ['worker', 'general'],
                capabilities: ['coding', 'reasoning'],
                executable: true,
                qualityScore: 10,
            },
            {
                id: 'local',
                provider: 'lmstudio',
                name: 'LM Studio Local',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: null,
                tier: 'local',
                recommendedFor: ['general'],
                capabilities: ['coding', 'reasoning'],
                executable: true,
                qualityScore: 4,
            },
        ],
    },
    {
        id: 'ollama',
        name: 'Ollama',
        authMethod: 'none',
        executable: true,
        defaultModel: 'gemma:2b',
        preferredTasks: ['general'],
        models: [
            {
                id: 'gemma:2b',
                provider: 'ollama',
                name: 'Gemma 2B',
                inputPrice: 0,
                outputPrice: 0,
                contextWindow: 8_192,
                tier: 'local',
                recommendedFor: ['general'],
                capabilities: ['coding'],
                executable: true,
                qualityScore: 3,
            },
        ],
    },
];

export class ProviderRegistry {
    private readonly catalog: ProviderDefinition[];
    private readonly modelIndex: Map<string, ProviderModelDefinition>;
    private refreshInterval: NodeJS.Timeout | null = null;

    constructor(catalog: ProviderDefinition[] = DEFAULT_PROVIDER_CATALOG) {
        this.catalog = catalog;
        this.modelIndex = new Map(
            catalog.flatMap((provider) => provider.models.map((model) => [model.id, model] as const)),
        );
        this.startPeriodicRefresh();
    }

    public async refreshFreeModels(): Promise<void> {
        try {
            if (typeof fetch !== 'function') {
                console.warn('[ProviderRegistry] fetch is not available; skipping free models refresh');
                return;
            }
            const response = await fetch('https://openrouter.ai/api/v1/models');
            if (!response.ok) {
                console.warn(`[ProviderRegistry] Failed to fetch OpenRouter models: HTTP ${response.status}`);
                return;
            }
            const payload = await response.json() as { data?: any[] };
            if (!payload || !Array.isArray(payload.data)) {
                console.warn('[ProviderRegistry] Invalid OpenRouter models response payload');
                return;
            }

            const openrouterProvider = this.catalog.find(p => p.id === 'openrouter');
            if (!openrouterProvider) {
                console.warn('[ProviderRegistry] OpenRouter provider definition not found in catalog');
                return;
            }

            const incomingFreeModels = payload.data.filter((model: any) => {
                const isFreeId = typeof model.id === 'string' && model.id.endsWith(':free');
                const pricing = model.pricing;
                const isFreePricing = pricing &&
                    parseFloat(pricing.prompt || '0') === 0 &&
                    parseFloat(pricing.completion || '0') === 0;
                return isFreeId || isFreePricing;
            });

            const newModels: ProviderModelDefinition[] = incomingFreeModels.map((model: any) => {
                const rawId = model.id;
                const existing = openrouterProvider.models.find(m => 
                    m.id === rawId || 
                    m.id === `openrouter/${rawId}` || 
                    `openrouter/${m.id}` === rawId
                );

                const finalId = existing ? existing.id : (rawId.startsWith('openrouter/') ? rawId : `openrouter/${rawId}`);

                const capabilities: ProviderCapability[] = ['coding'];
                const nameLower = (model.name || '').toLowerCase();
                const descLower = (model.description || '').toLowerCase();
                if (nameLower.includes('reasoning') || descLower.includes('reasoning') || nameLower.includes('thinking') || descLower.includes('thinking') || nameLower.includes('opus') || nameLower.includes('llama-3.3') || nameLower.includes('hermes-3')) {
                    capabilities.push('reasoning');
                }
                if (descLower.includes('vision') || descLower.includes('multimodal')) {
                    capabilities.push('vision');
                }
                if (descLower.includes('tool') || descLower.includes('function call')) {
                    capabilities.push('tools');
                }

                const contextWindow = typeof model.context_length === 'number' ? model.context_length : 8192;

                return {
                    id: finalId,
                    provider: 'openrouter',
                    name: model.name || rawId,
                    inputPrice: 0,
                    outputPrice: 0,
                    contextWindow,
                    tier: 'free',
                    recommendedFor: ['worker', 'general'] as ProviderTaskType[],
                    capabilities,
                    executable: true,
                    qualityScore: existing?.qualityScore ?? 7
                };
            });

            const nonFreeOrMeta = openrouterProvider.models.filter(m => m.tier !== 'free');
            
            const uniqueNewModelsMap = new Map<string, ProviderModelDefinition>();
            for (const m of newModels) {
                uniqueNewModelsMap.set(m.id, m);
            }
            const uniqueNewModels = Array.from(uniqueNewModelsMap.values());

            openrouterProvider.models = [...uniqueNewModels, ...nonFreeOrMeta];

            this.modelIndex.clear();
            for (const provider of this.catalog) {
                for (const model of provider.models) {
                    this.modelIndex.set(model.id, model);
                }
            }

            console.log(`[ProviderRegistry] Successfully refreshed ${uniqueNewModels.length} free OpenRouter models.`);
        } catch (error) {
            console.error('[ProviderRegistry] Error refreshing free models:', error);
        }
    }

    public startPeriodicRefresh(intervalMs: number = 6 * 60 * 60 * 1000): void {
        this.stopPeriodicRefresh();
        
        this.refreshFreeModels().catch(err => {
            console.error('[ProviderRegistry] Initial free model refresh failed:', err);
        });

        const timer = setInterval(() => {
            this.refreshFreeModels().catch(err => {
                console.error('[ProviderRegistry] Periodic free model refresh failed:', err);
            });
        }, intervalMs);

        if (typeof timer.unref === 'function') {
            timer.unref();
        }
        this.refreshInterval = timer;
    }

    public stopPeriodicRefresh(): void {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    public getCatalog(): ProviderDefinition[] {
        return this.catalog;
    }

    public getProvider(providerId: string): ProviderDefinition | undefined {
        return this.catalog.find((provider) => provider.id === providerId);
    }

    public getModel(modelId: string): ProviderModelDefinition | undefined {
        return this.modelIndex.get(modelId);
    }

    public resolveAuthState(providerId: string, env: NodeJS.ProcessEnv = process.env): ProviderAuthState {
        const provider = this.getProvider(providerId);
        if (!provider) {
            throw new Error(`Unknown provider: ${providerId}`);
        }

        const configured = this.hasAuthCredential(provider, env);
        const authenticated = provider.authMethod === 'none' ? true : configured;
        const authTruth = this.computeAuthTruth(provider, env);

        return {
            provider: provider.id,
            name: provider.name,
            authMethod: provider.authMethod,
            configured,
            authenticated,
            detail: this.getAuthDetail(provider, configured),
            authTruth,
        };
    }

    public getAuthStates(env: NodeJS.ProcessEnv = process.env): ProviderAuthState[] {
        return this.catalog.map((provider) => this.resolveAuthState(provider.id, env));
    }

    public getConfiguredProviders(env: NodeJS.ProcessEnv = process.env): ProviderAuthState[] {
        return this.getAuthStates(env).filter((state) => state.authenticated);
    }

    public listExecutableModels(env: NodeJS.ProcessEnv = process.env): ProviderModelDefinition[] {
        return this.catalog.flatMap((provider) => {
            const auth = this.resolveAuthState(provider.id, env);
            if ((provider.executable ?? false) && auth.authenticated) {
                return provider.models.filter((model) => model.executable !== false);
            }
            return [];
        });
    }

    public buildSnapshots(taskType: ProviderTaskType, env: NodeJS.ProcessEnv = process.env): FallbackCandidateSnapshot[] {
        return this.listExecutableModels(env).map((model) => ({
            id: model.id,
            provider: model.provider,
            model: model.id,
            name: model.name,
            inputPrice: model.inputPrice,
            outputPrice: model.outputPrice,
            contextWindow: model.contextWindow,
            tier: model.tier,
            recommended: this.isRecommendedForTask(model, taskType),
            reason: this.isRecommendedForTask(model, taskType) ? `${taskType}-optimized` : 'fallback',
        }));
    }

    public isRecommendedForTask(model: ProviderModelDefinition, taskType: ProviderTaskType): boolean {
        return model.recommendedFor?.includes(taskType) ?? false;
    }

    private hasAuthCredential(provider: ProviderDefinition, env: NodeJS.ProcessEnv): boolean {
        if (provider.authMethod === 'none') {
            return true;
        }

        const keys = [
            ...(provider.envKeys ?? []),
            ...(provider.oauthEnvKeys ?? []),
            ...(provider.patEnvKeys ?? []),
        ];

        return keys.some((key) => {
            const value = env[key];
            return typeof value === 'string' && value.trim().length > 0;
        });
    }

    private getAuthDetail(provider: ProviderDefinition, configured: boolean): string {
        if (provider.authMethod === 'none') {
            return 'Local provider does not require authentication.';
        }

        if (configured) {
            return `${provider.authMethod} credential detected.`;
        }

        const expectedKey = provider.envKeys?.[0] ?? provider.oauthEnvKeys?.[0] ?? provider.patEnvKeys?.[0] ?? 'credential';
        return `Missing ${expectedKey}.`;
    }

    /**
     * Computes the nuanced `ProviderAuthTruth` from environment credentials.
     *
     * - `none` providers are always `'authenticated'` (local, no key needed).
     * - For OAuth providers we additionally check `*_TOKEN_EXPIRES_AT` env vars to
     *   detect a configured-but-expired token without making a live API call.
     * - `'revoked'` cannot be determined from env alone; it is set externally via
     *   `NormalizedQuotaService.markAuthRevoked()` when a live 401/403 is observed.
     */
    private computeAuthTruth(provider: ProviderDefinition, env: NodeJS.ProcessEnv): ProviderAuthTruth {
        if (provider.authMethod === 'none') {
            return 'authenticated';
        }

        const hasCredential = this.hasAuthCredential(provider, env);
        if (!hasCredential) {
            return 'not_configured';
        }

        // For OAuth providers, detect expired tokens via *_TOKEN_EXPIRES_AT env vars.
        if (provider.authMethod === 'oauth') {
            const expiryKeys = (provider.oauthEnvKeys ?? []).map((key) => `${key}_EXPIRES_AT`);
            for (const expiryKey of expiryKeys) {
                const expiryValue = env[expiryKey];
                if (typeof expiryValue === 'string' && expiryValue.trim().length > 0) {
                    const expiryMs = Date.parse(expiryValue);
                    if (!Number.isNaN(expiryMs) && expiryMs < Date.now()) {
                        return 'expired';
                    }
                }
            }
        }

        return 'authenticated';
    }
}
