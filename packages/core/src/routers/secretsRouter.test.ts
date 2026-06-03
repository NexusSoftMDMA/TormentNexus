import { beforeEach, describe, expect, it, vi } from 'vitest';

const selectFromMock = vi.fn();
const insertOnConflictDoUpdateMock = vi.fn();
const insertValuesMock = vi.fn(() => ({
    onConflictDoUpdate: insertOnConflictDoUpdateMock,
}));
const insertMock = vi.fn(() => ({
    values: insertValuesMock,
}));
const deleteWhereMock = vi.fn();
const deleteMock = vi.fn(() => ({
    where: deleteWhereMock,
}));

vi.mock('../db/index.js', () => ({
    db: {
        select: vi.fn(() => ({
            from: selectFromMock,
        })),
        insert: insertMock,
        delete: deleteMock,
    },
}));

const { secretsRouter } = await import('./secretsRouter.js');

function createCaller() {
    return secretsRouter.createCaller({} as never);
}

describe('secretsRouter degraded SQLite handling', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        selectFromMock.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list()).rejects.toMatchObject({
            message: 'Workspace secrets storage is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for set when SQLite is unavailable', async () => {
        insertOnConflictDoUpdateMock.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.set({
            key: 'OPENAI_API_KEY',
            value: 'secret',
        })).rejects.toMatchObject({
            message: 'Workspace secrets storage is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
