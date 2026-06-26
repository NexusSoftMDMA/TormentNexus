import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/index.js', () => ({
    logsRepository: {
        findAll: vi.fn(),
        clear: vi.fn(),
    },
}));

const { logsRouter } = await import('./logsRouter.js');
const { logsRepository } = await import('../db/repositories/index.js');

function createCaller() {
    return logsRouter.createCaller({} as never);
}

describe('logsRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(logsRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repositoryMocks.findAll.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list({ limit: 10 })).rejects.toMatchObject({
            message: 'Logs are unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for clear when SQLite is unavailable', async () => {
        repositoryMocks.clear.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.clear()).rejects.toMatchObject({
            message: 'Logs are unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for summary when SQLite is unavailable', async () => {
        repositoryMocks.findAll.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.summary({ limit: 1000 })).rejects.toMatchObject({
            message: 'Logs are unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for summary when logs loading fails for non-SQLite reasons', async () => {
        repositoryMocks.findAll.mockRejectedValue(new Error('log index offline'));

        const caller = createCaller();

        await expect(caller.summary({ limit: 1000 })).rejects.toMatchObject({
            message: 'Logs are unavailable: log index offline',
        });
    });
});
