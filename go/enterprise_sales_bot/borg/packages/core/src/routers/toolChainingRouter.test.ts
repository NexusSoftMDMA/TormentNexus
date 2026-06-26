import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/tool-chains.repo.js', () => ({
    toolChainsRepository: {
        getAliasById: vi.fn(),
        upsertAlias: vi.fn(),
        getAliases: vi.fn(),
        deleteAlias: vi.fn(),
        createChain: vi.fn(),
        getAllChains: vi.fn(),
        getChainById: vi.fn(),
        deleteChain: vi.fn(),
    },
}));

const { toolChainingRouter } = await import('./toolChainingRouter.js');
const { toolChainsRepository } = await import('../db/repositories/tool-chains.repo.js');

function createCaller() {
    return toolChainingRouter.createCaller({} as never);
}

describe('toolChainingRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(toolChainsRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for listAliases when SQLite is unavailable', async () => {
        repositoryMocks.getAliases.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.listAliases()).rejects.toMatchObject({
            message: 'Tool alias store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for listChains when SQLite is unavailable', async () => {
        repositoryMocks.getAllChains.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.listChains()).rejects.toMatchObject({
            message: 'Tool chain store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
