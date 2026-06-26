import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/index.js', () => ({
    toolSetsRepository: {
        findAll: vi.fn(),
        findByUuid: vi.fn(),
        create: vi.fn(),
        update: vi.fn(),
        deleteByUuid: vi.fn(),
    },
}));

const { toolSetsRouter } = await import('./toolSetsRouter.js');
const { toolSetsRepository } = await import('../db/repositories/index.js');

function createCaller() {
    return toolSetsRouter.createCaller({} as never);
}

describe('toolSetsRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(toolSetsRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repositoryMocks.findAll.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list()).rejects.toMatchObject({
            message: 'Tool sets are unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for create when SQLite is unavailable', async () => {
        repositoryMocks.create.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.create({
            name: 'default',
            tools: ['search_tools'],
        })).rejects.toMatchObject({
            message: 'Tool sets are unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
