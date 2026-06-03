import { describe, expect, it, vi } from 'vitest';

import { ensureBackgroundCoreRunning, isCoreBridgeHealthy, waitForCoreBridge } from './backgroundCoreBootstrap.js';

describe('isCoreBridgeHealthy', () => {
    it('returns true when the health probe succeeds', async () => {
        expect(await isCoreBridgeHealthy('http://127.0.0.1:3001/health', vi.fn().mockResolvedValue({ ok: true }) as any)).toBe(true);
    });

    it('returns false when the health probe throws', async () => {
        expect(await isCoreBridgeHealthy('http://127.0.0.1:3001/health', vi.fn().mockRejectedValue(new Error('offline')) as any)).toBe(false);
    });
});

describe('waitForCoreBridge', () => {
    it('polls until the bridge becomes healthy', async () => {
        const fetchImpl = vi
            .fn()
            .mockResolvedValueOnce({ ok: false })
            .mockResolvedValueOnce({ ok: false })
            .mockResolvedValueOnce({ ok: true });

        expect(await waitForCoreBridge(
            {
                timeoutMs: 2_000,
                pollIntervalMs: 1,
            },
            {
                fetchImpl: fetchImpl as any,
                waitImpl: async () => undefined,
            },
        )).toBe(true);
        expect(fetchImpl).toHaveBeenCalledTimes(3);
    });
});

describe('ensureBackgroundCoreRunning', () => {
    it('does not spawn a new process when the bridge is already healthy', async () => {
        const spawnImpl = vi.fn();

        await expect(ensureBackgroundCoreRunning(
            {
                cliEntryPath: 'C:/tormentnexus/packages/cli/dist/index.js',
            },
            {
                fetchImpl: vi.fn().mockResolvedValue({ ok: true }) as any,
                spawnImpl: spawnImpl as any,
            },
        )).resolves.toEqual({ status: 'already-running' });

        expect(spawnImpl).not.toHaveBeenCalled();
    });

    it('spawns TormentNexus CLI start in detached mode when the bridge is offline', async () => {
        const unref = vi.fn();
        const spawnImpl = vi.fn().mockReturnValue({ pid: 4321, unref });

        await expect(ensureBackgroundCoreRunning(
            {
                cliEntryPath: 'C:/tormentnexus/packages/cli/dist/index.js',
                waitForReady: false,
                host: '127.0.0.1',
            },
            {
                fetchImpl: vi.fn().mockResolvedValue({ ok: false }) as any,
                spawnImpl: spawnImpl as any,
            },
        )).resolves.toEqual({
            status: 'spawned',
            pid: 4321,
            cliEntryPath: 'C:/tormentnexus/packages/cli/dist/index.js',
        });

        expect(spawnImpl).toHaveBeenCalledWith(process.execPath, [
            'C:/tormentnexus/packages/cli/dist/index.js',
            'start',
            '--host',
            '127.0.0.1',
        ], {
            detached: true,
            stdio: 'ignore',
            windowsHide: true,
        });
        expect(unref).toHaveBeenCalled();
    });

    it('reports launch-unavailable when the CLI entrypoint cannot be resolved', async () => {
        const log = vi.fn();

        await expect(ensureBackgroundCoreRunning(
            {
                cliEntryPath: null,
                log,
            },
            {
                fetchImpl: vi.fn().mockResolvedValue({ ok: false }) as any,
            },
        )).resolves.toEqual({ status: 'launch-unavailable' });

        expect(log).toHaveBeenCalled();
    });

    it('returns warming when the background start is still coming online after the timeout window', async () => {
        const spawnImpl = vi.fn().mockReturnValue({ pid: 999, unref: vi.fn() });
        const fetchImpl = vi.fn().mockResolvedValue({ ok: false });

        await expect(ensureBackgroundCoreRunning(
            {
                cliEntryPath: 'C:/tormentnexus/packages/cli/dist/index.js',
                startupTimeoutMs: 5,
                pollIntervalMs: 1,
            },
            {
                fetchImpl: fetchImpl as any,
                spawnImpl: spawnImpl as any,
                waitImpl: async () => undefined,
            },
        )).resolves.toEqual({
            status: 'warming',
            pid: 999,
            cliEntryPath: 'C:/tormentnexus/packages/cli/dist/index.js',
        });
    });
});