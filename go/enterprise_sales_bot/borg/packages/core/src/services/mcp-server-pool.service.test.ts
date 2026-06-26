import { beforeEach, describe, expect, it, vi } from 'vitest';

const sqliteUnavailable = new Error(
    'SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)',
);

const configServiceMock = {
    getSessionLifetime: vi.fn(),
};

vi.mock('./config.service.js', () => ({
    configService: configServiceMock,
}));

vi.mock('./auto-reconnect.service.js', () => ({
    autoReconnectService: {
        scheduleReconnect: vi.fn(),
        clearReconnectTimeout: vi.fn(),
    },
}));

vi.mock('./mcp-client.service.js', () => ({
    connecttormentnexusClient: vi.fn(),
}));

vi.mock('./server-error-tracker.service.js', () => ({
    serverErrorTracker: {
        resetServerErrorState: vi.fn(),
    },
}));

describe('McpServerPool degraded cleanup logging', () => {
    beforeEach(() => {
        vi.resetModules();
        vi.clearAllMocks();
        vi.spyOn(globalThis, 'setInterval').mockImplementation(() => 0 as unknown as NodeJS.Timeout);
    });

    it('logs the SQLite-unavailable cleanup warning only once per outage', async () => {
        configServiceMock.getSessionLifetime.mockRejectedValue(sqliteUnavailable);
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

        const { mcpServerPool } = await import('./mcp-server-pool.service.js');

        await (mcpServerPool as any).cleanupExpiredSessions();
        await (mcpServerPool as any).cleanupExpiredSessions();

        expect(warnSpy).toHaveBeenCalledTimes(1);
        expect(warnSpy).toHaveBeenCalledWith(
            '[McpServerPool] Skipping automatic session cleanup: SQLite runtime is unavailable for this run.',
        );
    });

    it('logs the cleanup warning again after SQLite access recovers and fails later', async () => {
        configServiceMock.getSessionLifetime
            .mockRejectedValueOnce(sqliteUnavailable)
            .mockResolvedValueOnce(null)
            .mockRejectedValueOnce(sqliteUnavailable);
        const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});

        const { mcpServerPool } = await import('./mcp-server-pool.service.js');

        await (mcpServerPool as any).cleanupExpiredSessions();
        await (mcpServerPool as any).cleanupExpiredSessions();
        await (mcpServerPool as any).cleanupExpiredSessions();

        expect(warnSpy).toHaveBeenCalledTimes(2);
        expect(warnSpy.mock.calls[0]?.[0]).toBe(
            '[McpServerPool] Skipping automatic session cleanup: SQLite runtime is unavailable for this run.',
        );
        expect(warnSpy.mock.calls[1]?.[0]).toBe(
            '[McpServerPool] Skipping automatic session cleanup: SQLite runtime is unavailable for this run.',
        );
    });
});
