import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/index.js', () => ({
    apiKeysRepository: {
        findAll: vi.fn(),
        findByUuidWithAccess: vi.fn(),
        create: vi.fn(),
        update: vi.fn(),
        delete: vi.fn(),
        validateApiKey: vi.fn(),
    },
}));

const { apiKeysRouter } = await import('./apiKeysRouter.js');
const { apiKeysRepository } = await import('../db/repositories/index.js');

function createCaller() {
    return apiKeysRouter.createCaller({} as never);
}

describe('apiKeysRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(apiKeysRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repositoryMocks.findAll.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list()).rejects.toMatchObject({
            message: 'API key store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for create when SQLite is unavailable', async () => {
        repositoryMocks.create.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.create({
            name: 'default-key',
            type: 'MCP',
        })).rejects.toMatchObject({
            message: 'API key store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
