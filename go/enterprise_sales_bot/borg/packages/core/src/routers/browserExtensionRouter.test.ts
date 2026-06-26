import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/browser-data.repo.js', () => ({
    browserDataRepository: {
        saveMemory: vi.fn(),
        getAllMemories: vi.fn(),
        getMemoryById: vi.fn(),
        deleteMemory: vi.fn(),
    },
}));

const { browserExtensionRouter } = await import('./browserExtensionRouter.js');
const { browserDataRepository } = await import('../db/repositories/browser-data.repo.js');

function createCaller() {
    return browserExtensionRouter.createCaller({} as never);
}

describe('browserExtensionRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(browserDataRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for listMemories when SQLite is unavailable', async () => {
        repositoryMocks.getAllMemories.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.listMemories({
            limit: 10,
            offset: 0,
        })).rejects.toMatchObject({
            message: 'Browser memory store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for deleteMemory when SQLite is unavailable', async () => {
        repositoryMocks.deleteMemory.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.deleteMemory({ id: 'mem-1' })).rejects.toMatchObject({
            message: 'Browser memory store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
