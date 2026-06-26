import { TRPCError } from '@trpc/server';
import { afterEach, describe, expect, it, vi } from 'vitest';

const runtimeState = {
    agentMemoryService: undefined as Record<string, unknown> | undefined,
};

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getMemoryManager: () => ({
            saveContext: vi.fn(),
            listContexts: vi.fn(async () => []),
        }),
        getAgentMemoryService: () => runtimeState.agentMemoryService as any,
    };
});

describe('memoryRouter truthfulness', () => {
    afterEach(() => {
        vi.restoreAllMocks();
        runtimeState.agentMemoryService = undefined;
    });

    it('throws when recent observations are unavailable because the agent memory service is missing', async () => {
        const { memoryRouter } = await import('./memoryRouter.js');
        const caller = memoryRouter.createCaller({});

        await expect(caller.getRecentObservations({ limit: 5, namespace: 'project' })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Recent observations is unavailable: Agent memory service not initialized',
        });
    });

    it('throws when cross-session links are unavailable because the service lacks support', async () => {
        runtimeState.agentMemoryService = {
            getStats: vi.fn(),
            search: vi.fn(async () => []),
            add: vi.fn(async () => undefined),
        };

        const { memoryRouter } = await import('./memoryRouter.js');
        const caller = memoryRouter.createCaller({});

        await expect(caller.getCrossSessionMemoryLinks({ sessionId: 'session-1', memoryId: 'mem-1', limit: 4 })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Cross-session memory links is unavailable: Agent memory service does not support this operation',
        });
    });
});
