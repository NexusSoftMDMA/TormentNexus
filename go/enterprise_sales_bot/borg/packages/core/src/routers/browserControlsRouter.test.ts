import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/browser-data.repo.js', () => ({
    browserDataRepository: {
        addHistoryEntry: vi.fn(),
        getHistory: vi.fn(),
        clearHistory: vi.fn(),
        addConsoleLog: vi.fn(),
        getConsoleLogs: vi.fn(),
        clearConsoleLogs: vi.fn(),
    },
}));

const { browserControlsRouter } = await import('./browserControlsRouter.js');
const { browserDataRepository } = await import('../db/repositories/browser-data.repo.js');

function createCaller() {
    return browserControlsRouter.createCaller({} as never);
}

describe('browserControlsRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(browserDataRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for queryHistory when SQLite is unavailable', async () => {
        repositoryMocks.getHistory.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.queryHistory({
            limit: 10,
        })).rejects.toMatchObject({
            message: 'Browser data store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for stats when SQLite is unavailable', async () => {
        repositoryMocks.getHistory.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.stats()).rejects.toMatchObject({
            message: 'Browser data store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
