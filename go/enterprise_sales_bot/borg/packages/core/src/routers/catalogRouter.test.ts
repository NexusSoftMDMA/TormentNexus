import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/published-catalog.repo.js', () => ({
    publishedCatalogRepository: {
        listServers: vi.fn(),
        countServers: vi.fn(),
        findServerByUuid: vi.fn(),
        listRunsForServer: vi.fn(),
        getActiveRecipe: vi.fn(),
        findSourcesByServerUuid: vi.fn(),
        listUuidsByStatus: vi.fn(),
        countRecentlyUpdated: vi.fn(),
    },
}));

vi.mock('../db/repositories/mcp-servers.repo.js', () => ({
    mcpServersRepository: {
        findByName: vi.fn(),
        findAll: vi.fn(),
        create: vi.fn(),
    },
}));

vi.mock('../services/published-catalog-ingestor.js', () => ({
    ingestPublishedCatalog: vi.fn(),
}));

vi.mock('../services/published-catalog-validator.js', () => ({
    validatePublishedServer: vi.fn(),
}));

const { catalogRouter } = await import('./catalogRouter.js');
const { publishedCatalogRepository } = await import('../db/repositories/published-catalog.repo.js');

function createCaller() {
    return catalogRouter.createCaller({} as never);
}

describe('catalogRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(publishedCatalogRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repositoryMocks.listServers.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list({ limit: 10, offset: 0 })).rejects.toMatchObject({
            message: 'Published catalog is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for stats when SQLite is unavailable', async () => {
        repositoryMocks.countServers.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.stats()).rejects.toMatchObject({
            message: 'Published catalog is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
