import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/links-backlog.repo.js', () => ({
    linksBacklogRepository: {
        listLinks: vi.fn(),
        countLinks: vi.fn(),
        getStats: vi.fn(),
        findByUuid: vi.fn(),
    },
}));

vi.mock('../services/bobby-bookmarks-adapter.js', () => ({
    BobbyBookmarksBacklogAdapter: class BobbyBookmarksBacklogAdapter {
        sync() {
            return Promise.resolve({
                source: 'bobbybookmarks',
                fetched: 0,
                upserted: 0,
                pages: 0,
                errors: [],
                baseUrl: 'https://example.test',
            });
        }
    },
}));

const { linksBacklogRouter } = await import('./linksBacklogRouter.js');
const { linksBacklogRepository } = await import('../db/repositories/links-backlog.repo.js');

function createCaller() {
    return linksBacklogRouter.createCaller({} as never);
}

describe('linksBacklogRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(linksBacklogRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repositoryMocks.listLinks.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list({ limit: 10, offset: 0 })).rejects.toMatchObject({
            message: 'Links backlog is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for stats when SQLite is unavailable', async () => {
        repositoryMocks.getStats.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.stats()).rejects.toMatchObject({
            message: 'Links backlog is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
