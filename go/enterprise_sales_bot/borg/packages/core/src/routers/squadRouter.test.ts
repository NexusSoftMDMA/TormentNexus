import { TRPCError } from '@trpc/server';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getSquadService: () => undefined,
    };
});

describe('squadRouter indexer truthfulness', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('returns a service-unavailable error when indexer status is requested without SquadService', async () => {
        const { squadRouter } = await import('./squadRouter.js');
        const caller = squadRouter.createCaller({});

        await expect(caller.getIndexerStatus()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Squad service is not initialized.',
        });
    });

    it('returns a service-unavailable error when toggling the indexer without SquadService', async () => {
        const { squadRouter } = await import('./squadRouter.js');
        const caller = squadRouter.createCaller({});

        await expect(caller.toggleIndexer({ enabled: true })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Squad service is not initialized; cannot toggle the indexer.',
        });
    });
});
