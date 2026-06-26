/**
 * Provider state transition tests — Task 028
 *
 * Verifies that authTruth and quotaConfidence values are correctly
 * computed, propagated, and mutated across the lifecycle of a provider:
 *   env-only detection → live balance refresh → 401/403 revocation → recovery
 */

import { describe, expect, it, vi } from 'vitest';

import { CoreModelSelector } from './CoreModelSelector.js';
import { NormalizedQuotaService } from './NormalizedQuotaService.js';
import { ProviderBalanceService } from './ProviderBalanceService.js';
import { ProviderRegistry } from './ProviderRegistry.js';
import type { ProviderDefinition, ProviderQuotaSnapshot } from './types.js';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeApiKeyProvider(id: string, envKey: string): ProviderDefinition {
    return {
        id,
        name: id,
        authMethod: 'api_key',
        envKeys: [envKey],
        executable: true,
        defaultModel: `${id}/model`,
        models: [
            {
                id: `${id}/model`,
                provider: id,
                name: `${id} model`,
                inputPrice: 0.001,
                outputPrice: 0.002,
                contextWindow: 128_000,
                tier: 'standard',
                capabilities: ['coding'],
                executable: true,
                qualityScore: 7,
            },
        ],
    };
}

function makeOauthProvider(id: string, tokenKey: string): ProviderDefinition {
    return {
        id,
        name: id,
        authMethod: 'oauth',
        oauthEnvKeys: [tokenKey],
        executable: false,
        defaultModel: `${id}/model`,
        models: [
            {
                id: `${id}/model`,
                provider: id,
                name: `${id} model`,
                inputPrice: null,
                outputPrice: null,
                contextWindow: null,
                tier: 'managed',
                capabilities: ['coding'],
                qualityScore: 6,
            },
        ],
    };
}

/**
 * Build a NormalizedQuotaService with a specific env so auth states are
 * set correctly after the constructor (which calls refreshAuthStates).
 *
 * NOTE: The NormalizedQuotaService constructor calls setConfig() which calls
 * refreshAuthStates(process.env). To have the test env visible we call
 * refreshAuthStates(testEnv) AFTER the constructor finishes.
 */
function makeQuotaService(
    provider: ProviderDefinition,
    testEnv: NodeJS.ProcessEnv = {},
    balanceService?: ProviderBalanceService,
): NormalizedQuotaService {
    const registry = new ProviderRegistry([provider]);
    const svc = new NormalizedQuotaService(registry, balanceService);
    // Re-apply the test env after the constructor's refreshAuthStates(process.env).
    svc.refreshAuthStates(testEnv);
    return svc;
}

// ---------------------------------------------------------------------------
// ProviderRegistry.resolveAuthState — authTruth via env
// ---------------------------------------------------------------------------

describe('ProviderRegistry authTruth detection', () => {
    it('returns not_configured when api_key provider has no env key', () => {
        const registry = new ProviderRegistry([makeApiKeyProvider('acme', 'ACME_API_KEY')]);
        const state = registry.resolveAuthState('acme', {});

        expect(state.authTruth).toBe('not_configured');
        expect(state.authenticated).toBe(false);
    });

    it('returns authenticated when api_key provider has a non-empty env key', () => {
        const registry = new ProviderRegistry([makeApiKeyProvider('acme', 'ACME_API_KEY')]);
        const state = registry.resolveAuthState('acme', { ACME_API_KEY: 'sk-abc123' });

        expect(state.authTruth).toBe('authenticated');
        expect(state.authenticated).toBe(true);
    });

    it('returns expired when OAuth provider has a token but its EXPIRES_AT is in the past', () => {
        const registry = new ProviderRegistry([makeOauthProvider('goog', 'GOOG_OAUTH_TOKEN')]);
        const pastTimestamp = new Date(Date.now() - 60_000).toISOString();
        const state = registry.resolveAuthState('goog', {
            GOOG_OAUTH_TOKEN: 'ya29.token',
            GOOG_OAUTH_TOKEN_EXPIRES_AT: pastTimestamp,
        });

        expect(state.authTruth).toBe('expired');
    });

    it('returns authenticated when OAuth provider has a token with a future EXPIRES_AT', () => {
        const registry = new ProviderRegistry([makeOauthProvider('goog', 'GOOG_OAUTH_TOKEN')]);
        const futureTimestamp = new Date(Date.now() + 3_600_000).toISOString();
        const state = registry.resolveAuthState('goog', {
            GOOG_OAUTH_TOKEN: 'ya29.token',
            GOOG_OAUTH_TOKEN_EXPIRES_AT: futureTimestamp,
        });

        expect(state.authTruth).toBe('authenticated');
    });

    it('returns authenticated when OAuth token has no expiry env var (unknown expiry)', () => {
        const registry = new ProviderRegistry([makeOauthProvider('goog', 'GOOG_OAUTH_TOKEN')]);
        const state = registry.resolveAuthState('goog', { GOOG_OAUTH_TOKEN: 'ya29.token' });

        expect(state.authTruth).toBe('authenticated');
    });
});

// ---------------------------------------------------------------------------
// NormalizedQuotaService — authTruth mutation and quotaConfidence propagation
// ---------------------------------------------------------------------------

describe('NormalizedQuotaService authTruth mutation', () => {
    const PROVIDER = makeApiKeyProvider('acme', 'ACME_API_KEY');
    const AUTHED_ENV = { ACME_API_KEY: 'sk-test' } as NodeJS.ProcessEnv;

    it('authTruth becomes revoked after markAuthRevoked is called', () => {
        const svc = makeQuotaService(PROVIDER, AUTHED_ENV);
        svc.markAuthRevoked('acme', 'Test revocation');

        const snapshot = svc.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.authTruth).toBe('revoked');
        expect(snapshot?.authenticated).toBe(false);
        expect(snapshot?.availability).toBe('missing_auth');
    });

    it('preserves revoked authTruth across refreshAuthStates calls when credential is still present', () => {
        const svc = makeQuotaService(PROVIDER, AUTHED_ENV);
        svc.markAuthRevoked('acme', 'Revoked by 401');

        // Refresh with env that still has the key — revoked state must persist.
        svc.refreshAuthStates(AUTHED_ENV);

        const snapshot = svc.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.authTruth).toBe('revoked');
    });

    it('markProviderHealthy restores authTruth to authenticated after revocation', () => {
        const svc = makeQuotaService(PROVIDER, AUTHED_ENV);
        svc.markAuthRevoked('acme', 'Test revocation');
        svc.markProviderHealthy('acme');

        const snapshot = svc.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.authTruth).toBe('authenticated');
    });

    it('full recovery: markProviderHealthy + refreshAuthStates makes provider isProviderReady', () => {
        const svc = makeQuotaService(PROVIDER, AUTHED_ENV);
        svc.markAuthRevoked('acme', 'Test revocation');
        // Step 1: restore authTruth (clears the "revoked" flag)
        svc.markProviderHealthy('acme');
        // Step 2: re-read env to restore authenticated=true + availability=available
        svc.refreshAuthStates(AUTHED_ENV);

        expect(svc.isProviderReady('acme')).toBe(true);
    });

    it('quotaConfidence is estimated after refreshAuthStates (no live balance data)', () => {
        const svc = makeQuotaService(PROVIDER, AUTHED_ENV);
        const snapshot = svc.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.quotaConfidence).toBe('estimated');
        expect(snapshot?.quotaRefreshedAt).toBeNull();
    });

    it('quotaConfidence becomes real-time after refreshProviderBalances for providers in the balance catalog', async () => {
        // Inject a mock balance service that returns a real-time snapshot for 'acme'.
        const mockSnapshot: ProviderQuotaSnapshot = {
            provider: 'acme',
            name: 'acme',
            authMethod: 'api_key',
            configured: true,
            authenticated: true,
            detail: 'api_key credential detected.',
            authTruth: 'authenticated',
            tier: 'standard',
            used: 0,
            limit: 100,
            remaining: 100,
            resetDate: null,
            rateLimitRpm: null,
            availability: 'available',
            retryAfter: null,
            quotaConfidence: 'real-time',
            quotaRefreshedAt: null,
        };
        const mockBalanceService = {
            fetchSnapshots: vi.fn().mockResolvedValue([mockSnapshot]),
        } as unknown as ProviderBalanceService;

        const svc = makeQuotaService(PROVIDER, AUTHED_ENV, mockBalanceService);
        await svc.refreshProviderBalances();

        const snapshot = svc.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.quotaConfidence).toBe('real-time');
        expect(typeof snapshot?.quotaRefreshedAt).toBe('string');
    });

    it('quotaConfidence remains estimated for providers not in the balance service catalog', async () => {
        const emptyBalanceService = {
            fetchSnapshots: vi.fn().mockResolvedValue([]),
        } as unknown as ProviderBalanceService;

        const svc = makeQuotaService(PROVIDER, AUTHED_ENV, emptyBalanceService);
        await svc.refreshProviderBalances();

        const snapshot = svc.getAllQuotas().find((q) => q.provider === 'acme');
        // Provider not in balance catalog — quota data stays at estimated from refreshAuthStates.
        expect(snapshot?.quotaConfidence).toBe('estimated');
    });
});

// ---------------------------------------------------------------------------
// CoreModelSelector — 401/403 triggers revocation; 429 does not
// ---------------------------------------------------------------------------

describe('CoreModelSelector.reportFailure auth revocation', () => {
    const PROVIDER = makeApiKeyProvider('acme', 'ACME_API_KEY');
    const AUTHED_ENV = { ACME_API_KEY: 'sk-test' } as NodeJS.ProcessEnv;

    function buildSelector() {
        const registry = new ProviderRegistry([PROVIDER]);
        const quotaService = new NormalizedQuotaService(registry);
        const selector = new CoreModelSelector({ registry, quotaService });
        // Re-apply test env AFTER selector constructor's setConfig() call.
        quotaService.refreshAuthStates(AUTHED_ENV);
        return { selector, quotaService };
    }

    it('reportFailure with status 401 sets authTruth to revoked', () => {
        const { selector, quotaService } = buildSelector();
        selector.reportFailure('acme', 'acme/model', { status: 401, message: 'Unauthorized' });

        const snapshot = quotaService.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.authTruth).toBe('revoked');
        expect(snapshot?.availability).toBe('missing_auth');
    });

    it('reportFailure with status 403 sets authTruth to revoked', () => {
        const { selector, quotaService } = buildSelector();
        selector.reportFailure('acme', 'acme/model', { status: 403, message: 'Forbidden' });

        const snapshot = quotaService.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.authTruth).toBe('revoked');
        expect(snapshot?.availability).toBe('missing_auth');
    });

    it('reportFailure with status 429 does NOT set authTruth to revoked (rate limit only)', () => {
        const { selector, quotaService } = buildSelector();
        selector.reportFailure('acme', 'acme/model', { status: 429, message: 'Rate limit exceeded' });

        const snapshot = quotaService.getAllQuotas().find((q) => q.provider === 'acme');
        // Should remain authenticated — 429 is a transient rate limit, not a credential failure.
        expect(snapshot?.authTruth).not.toBe('revoked');
        expect(snapshot?.authTruth).toBe('authenticated');
    });

    it('revoked provider is excluded from routing pool (isProviderReady returns false)', () => {
        const { selector, quotaService } = buildSelector();
        selector.reportFailure('acme', 'acme/model', { status: 401, message: 'Unauthorized' });

        expect(quotaService.isProviderReady('acme')).toBe(false);
    });

    it('full recovery: markProviderHealthy + refreshAuthStates restores provider to ready state', () => {
        const { selector, quotaService } = buildSelector();
        selector.reportFailure('acme', 'acme/model', { status: 401, message: 'Unauthorized' });

        // Step 1: markProviderHealthy clears the 'revoked' authTruth flag.
        quotaService.markProviderHealthy('acme');
        // Step 2: refreshAuthStates re-reads env → restores authenticated=true + availability=available.
        quotaService.refreshAuthStates(AUTHED_ENV);

        const snapshot = quotaService.getAllQuotas().find((q) => q.provider === 'acme');
        expect(snapshot?.authTruth).toBe('authenticated');
        expect(quotaService.isProviderReady('acme')).toBe(true);
    });
});
