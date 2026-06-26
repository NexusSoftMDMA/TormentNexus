import { TRPCError } from '@trpc/server';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { buildFallbackChainResponse } from './billingRouter.js';

const llmServiceMock = {
    getCostStats: vi.fn(() => ({ estimatedCostUSD: 0 })),
    modelSelector: {
        getQuotaService: vi.fn(() => ({ getUsageByModel: vi.fn(() => []) })),
        getFallbackHistory: vi.fn(),
        clearFallbackHistory: vi.fn(),
        getDepletedModels: vi.fn(),
    },
};

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getLLMService: () => llmServiceMock,
    };
});

describe('buildFallbackChainResponse', () => {
    it('forwards the selected task type to the selector and preserves ranked entries', () => {
        const getFallbackChain = vi.fn().mockReturnValue([
            {
                provider: 'anthropic',
                model: 'claude-sonnet-4-20250514',
                reason: 'TASK_TYPE_PLANNING',
            },
            {
                provider: 'openai',
                model: 'gpt-4o',
                reason: 'BEST',
            },
        ]);

        const response = buildFallbackChainResponse({ getFallbackChain }, 'planning');

        expect(getFallbackChain).toHaveBeenCalledWith({ routingTaskType: 'planning' });
        expect(response).toEqual({
            selectedTaskType: 'planning',
            chain: [
                {
                    priority: 1,
                    provider: 'anthropic',
                    model: 'claude-sonnet-4-20250514',
                    reason: 'TASK_TYPE_PLANNING',
                },
                {
                    priority: 2,
                    provider: 'openai',
                    model: 'gpt-4o',
                    reason: 'BEST',
                },
            ],
        });
    });

    it('falls back to the default chain when no task type is supplied', () => {
        const getFallbackChain = vi.fn().mockReturnValue([]);

        const response = buildFallbackChainResponse({ getFallbackChain });

        expect(getFallbackChain).toHaveBeenCalledWith(undefined);
        expect(response).toEqual({
            selectedTaskType: null,
            chain: [],
        });
    });
});

describe('billingRouter truthfulness', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('returns a concise TRPC error when fallback history is unavailable', async () => {
        llmServiceMock.modelSelector.getFallbackHistory.mockImplementation(() => {
            throw new Error('fallback buffer offline');
        });

        const { billingRouter } = await import('./billingRouter.js');
        const caller = billingRouter.createCaller({});

        await expect(caller.getFallbackHistory({ limit: 20 })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Fallback history is unavailable: fallback buffer offline',
        });
    });

    it('returns a concise TRPC error when depleted model state is unavailable', async () => {
        llmServiceMock.modelSelector.getDepletedModels.mockImplementation(() => {
            throw new Error('depletion state offline');
        });

        const { billingRouter } = await import('./billingRouter.js');
        const caller = billingRouter.createCaller({});

        await expect(caller.getDepletedModels()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Depleted model status is unavailable: depletion state offline',
        });
    });

    it('returns a concise TRPC error when clearing fallback history fails', async () => {
        llmServiceMock.modelSelector.clearFallbackHistory.mockImplementation(() => {
            throw new Error('fallback clear offline');
        });

        const { billingRouter } = await import('./billingRouter.js');
        const caller = billingRouter.createCaller({});

        await expect(caller.clearFallbackHistory()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Fallback history clearing is unavailable: fallback clear offline',
        });
    });
});
