import { TRPCError } from '@trpc/server';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const trackerMock = {
    isServerInErrorState: vi.fn(),
    getServerAttempts: vi.fn(() => 0),
    getServerMaxAttempts: vi.fn(() => Promise.resolve(1)),
    resetServerErrorState: vi.fn(() => Promise.resolve()),
};

vi.mock('../services/server-error-tracker.service.js', () => ({
    serverErrorTracker: trackerMock,
}));

describe('serverHealthRouter degraded status handling', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        trackerMock.getServerAttempts.mockReturnValue(0);
        trackerMock.getServerMaxAttempts.mockResolvedValue(1);
    });

    it('returns a concise TRPC error when persisted server error state is unavailable', async () => {
        trackerMock.isServerInErrorState.mockRejectedValueOnce(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const { serverHealthRouter } = await import('./serverHealthRouter.js');
        const caller = serverHealthRouter.createCaller({});

        await expect(caller.check({ serverUuid: 'server-1' })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Server health status is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
