import { TRPCError } from '@trpc/server';
import { afterEach, describe, expect, it, vi } from 'vitest';

const submoduleServiceMock = {
    listSubmodules: vi.fn(async () => []),
    updateAll: vi.fn(),
    installDependencies: vi.fn(),
    buildSubmodule: vi.fn(),
    enableSubmodule: vi.fn(),
    detectCapabilities: vi.fn(),
};

const runtimeState = {
    serviceAvailable: true,
};

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getSubmoduleService: () => (runtimeState.serviceAvailable ? submoduleServiceMock : undefined),
    };
});

describe('submoduleRouter truthfulness', () => {
    afterEach(() => {
        vi.restoreAllMocks();
        runtimeState.serviceAvailable = true;
        submoduleServiceMock.listSubmodules.mockReset();
        submoduleServiceMock.listSubmodules.mockResolvedValue([]);
    });

    it('throws a concise TRPC error when the submodule service is unavailable', async () => {
        runtimeState.serviceAvailable = false;

        const { submoduleRouter } = await import('./submoduleRouter.js');
        const caller = submoduleRouter.createCaller({});

        await expect(caller.list()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Submodule inventory is unavailable: SubmoduleService not available',
        });
    });

    it('throws a concise TRPC error when listing submodules fails', async () => {
        submoduleServiceMock.listSubmodules.mockRejectedValueOnce(new Error('git config parse failed'));

        const { submoduleRouter } = await import('./submoduleRouter.js');
        const caller = submoduleRouter.createCaller({});

        await expect(caller.list()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Submodule inventory is unavailable: git config parse failed',
        });
    });
});
