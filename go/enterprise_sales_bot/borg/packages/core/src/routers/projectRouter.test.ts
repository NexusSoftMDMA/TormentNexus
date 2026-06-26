import { TRPCError } from '@trpc/server';
import { afterEach, describe, expect, it, vi } from 'vitest';

const accessMock = vi.fn();
const readFileMock = vi.fn();
const readdirMock = vi.fn();

vi.mock('fs/promises', () => ({
    default: {
        access: accessMock,
        readFile: readFileMock,
        readdir: readdirMock,
    },
    access: accessMock,
    readFile: readFileMock,
    readdir: readdirMock,
}));

describe('projectRouter truthfulness', () => {
    afterEach(() => {
        vi.restoreAllMocks();
        accessMock.mockReset();
        readFileMock.mockReset();
        readdirMock.mockReset();
    });

    it('throws a concise TRPC error when project context cannot be read', async () => {
        accessMock.mockResolvedValueOnce(undefined);
        readFileMock.mockRejectedValueOnce(new Error('permission denied'));

        const { projectRouter } = await import('./projectRouter.js');
        const caller = projectRouter.createCaller({});

        await expect(caller.getContext()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Project context is unavailable: permission denied',
        });
    });

    it('throws a concise TRPC error when handoffs cannot be listed', async () => {
        accessMock.mockResolvedValueOnce(undefined);
        readdirMock.mockRejectedValueOnce(new Error('directory unreadable'));

        const { projectRouter } = await import('./projectRouter.js');
        const caller = projectRouter.createCaller({});

        await expect(caller.getHandoffs()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Project handoffs are unavailable: directory unreadable',
        });
    });
});
