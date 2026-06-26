import { beforeEach, describe, expect, it, vi } from 'vitest';
import { TRPCError } from '@trpc/server';

const serverMock = {
    llmService: {
        getCostStats: vi.fn(),
        modelSelector: {
            getQuotaService: vi.fn(),
        },
    },
    metricsService: {
        getStats: vi.fn(),
        track: vi.fn(),
        startMonitoring: vi.fn(),
        stopMonitoring: vi.fn(),
    },
};

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getMcpServer: () => serverMock,
    };
});

describe('metricsRouter provider breakdown truthfulness', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('returns an explicit degraded payload instead of a synthetic startup provider', async () => {
        serverMock.llmService.getCostStats.mockImplementation(() => {
            throw new Error('quota service offline');
        });

        const { metricsRouter } = await import('./metricsRouter.js');
        const caller = metricsRouter.createCaller({});

        await expect(caller.getProviderBreakdown()).resolves.toEqual({
            totalCost: null,
            totalRequests: null,
            averageLatency: null,
            providers: [],
            error: 'Provider metrics unavailable: quota service offline',
        });
    });

    it('returns a concise TRPC error when routing history is unavailable', async () => {
        serverMock.llmService.getRoutingHistory = vi.fn(() => {
            throw new Error('routing buffer unavailable');
        });

        const { metricsRouter } = await import('./metricsRouter.js');
        const caller = metricsRouter.createCaller({});

        await expect(caller.getRoutingHistory({ limit: 20 })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Routing history is unavailable: routing buffer unavailable',
        });
    });
});
