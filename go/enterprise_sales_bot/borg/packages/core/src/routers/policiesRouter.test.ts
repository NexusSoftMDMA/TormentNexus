import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/index.js', () => ({
    policiesRepository: {
        findAll: vi.fn(),
        findByUuid: vi.fn(),
        create: vi.fn(),
        update: vi.fn(),
        delete: vi.fn(),
    },
}));

const { policiesRouter } = await import('./policiesRouter.js');
const { policiesRepository } = await import('../db/repositories/index.js');

function createCaller() {
    return policiesRouter.createCaller({} as never);
}

describe('policiesRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(policiesRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repositoryMocks.findAll.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list()).rejects.toMatchObject({
            message: 'Policies are unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for create when SQLite is unavailable', async () => {
        repositoryMocks.create.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.create({
            name: 'default-policy',
            rules: { allow: ['tool:*'] },
        })).rejects.toMatchObject({
            message: 'Policies are unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
