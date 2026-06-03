import { beforeEach, describe, expect, it, vi } from 'vitest';

const sqliteUnavailable = new Error(
    'SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)',
);

const configServiceMock = {
    getMcpMaxAttempts: vi.fn(),
};

const repositoryMock = {
    findByUuid: vi.fn(),
    updateServerErrorStatus: vi.fn(),
};

vi.mock('./config.service.js', () => ({
    configService: configServiceMock,
}));

vi.mock('../db/repositories/index.js', () => ({
    mcpServersRepository: repositoryMock,
}));

describe('ServerErrorTracker degraded logging', () => {
    beforeEach(() => {
        vi.resetModules();
        vi.clearAllMocks();
    });

    it('falls back cleanly when config-backed max attempts are unavailable', async () => {
        configServiceMock.getMcpMaxAttempts.mockRejectedValue(sqliteUnavailable);
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

        const { serverErrorTracker } = await import('./server-error-tracker.service.js');
        const maxAttempts = await serverErrorTracker.getServerMaxAttempts('server-1');

        expect(maxAttempts).toBe(1);
        expect(warnSpy).toHaveBeenCalledWith(
            '[ServerErrorTracker] Failed to get MCP max attempts from config, using fallback: SQLite runtime is unavailable for this run.',
        );
    });

    it('rethrows SQLite-unavailable server error state lookups instead of reporting false healthy state', async () => {
        repositoryMock.findByUuid.mockRejectedValue(sqliteUnavailable);
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

        const { serverErrorTracker } = await import('./server-error-tracker.service.js');
        await expect(serverErrorTracker.isServerInErrorState('server-2')).rejects.toThrow(sqliteUnavailable);
        expect(warnSpy).not.toHaveBeenCalled();
    });

    it('resets server error state with a concise warning when persistence is unavailable', async () => {
        repositoryMock.updateServerErrorStatus.mockRejectedValue(sqliteUnavailable);
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        const infoSpy = vi.spyOn(console, 'info').mockImplementation(() => {});

        const { serverErrorTracker } = await import('./server-error-tracker.service.js');
        await serverErrorTracker.resetServerErrorState('server-3');

        expect(warnSpy).toHaveBeenCalledWith(
            '[ServerErrorTracker] Error resetting error state for server-3: SQLite runtime is unavailable for this run.',
        );
        expect(infoSpy).not.toHaveBeenCalledWith('Reset error state for server server-3');
    });
});
