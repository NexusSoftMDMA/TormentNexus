import { TRPCError } from '@trpc/server';
import { describe, expect, it, vi } from 'vitest';

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getBrowserService: () => null,
        getMcpServer: () => ({
            executeTool: vi.fn(),
        }),
    };
});

describe('browserRouter mutation truthfulness', () => {
    it('throws when closePage is requested without a browser service', async () => {
        const { browserRouter } = await import('./browserRouter.js');
        const caller = browserRouter.createCaller({});

        await expect(caller.closePage({ pageId: 'page-1' })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Browser service unavailable',
        });
    });

    it('throws when closeAll is requested without a browser service', async () => {
        const { browserRouter } = await import('./browserRouter.js');
        const caller = browserRouter.createCaller({});

        await expect(caller.closeAll()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Browser service unavailable',
        });
    });
});
