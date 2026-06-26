import { TRPCError } from '@trpc/server';
import { afterEach, describe, expect, it, vi } from 'vitest';

const readFileMock = vi.fn();

vi.mock('fs/promises', () => ({
    default: {
        readFile: readFileMock,
    },
    readFile: readFileMock,
}));

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getGitService: () => ({
            getLog: vi.fn(),
            getStatus: vi.fn(),
            revert: vi.fn(),
        }),
    };
});

describe('gitRouter getModules truthfulness', () => {
    afterEach(() => {
        vi.restoreAllMocks();
        readFileMock.mockReset();
    });

    it('returns empty when .gitmodules is missing', async () => {
        readFileMock.mockRejectedValueOnce(Object.assign(new Error('missing'), { code: 'ENOENT' }));

        const { gitRouter } = await import('./gitRouter.js');
        const caller = gitRouter.createCaller({});

        await expect(caller.getModules()).resolves.toEqual([]);
    });

    it('throws a concise TRPC error when .gitmodules cannot be parsed', async () => {
        readFileMock.mockResolvedValueOnce('[submodule "broken"]\npath = packages/core\n');

        const { gitRouter } = await import('./gitRouter.js');
        const caller = gitRouter.createCaller({});

        await expect(caller.getModules()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Git modules are unavailable: Malformed .gitmodules file',
        });
    });
});
