import { beforeEach, describe, expect, it, vi } from 'vitest';

const sqliteUnavailable = new Error(
    'SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)',
);

const repositoryMock = {
    findByUuid: vi.fn(),
    findAll: vi.fn(),
};

vi.mock('../db/repositories/index.js', () => ({
    mcpServersRepository: repositoryMock,
}));

vi.mock('./config.service.js', () => ({
    configService: {
        getMcpTimeout: vi.fn(async () => 10_000),
        getConfig: vi.fn(async () => null),
    },
}));

vi.mock('./auto-reconnect.service.js', () => ({
    autoReconnectService: {
        scheduleReconnection: vi.fn(),
    },
}));

vi.mock('./mcp-client.service.js', () => ({
    connecttormentnexusClient: vi.fn(),
}));

vi.mock('./server-error-tracker.service.js', () => ({
    serverErrorTracker: {
        isServerInErrorState: vi.fn(async () => false),
    },
}));

vi.mock('./utils.service.js', () => ({
    convertDbServerToParams: vi.fn(),
}));

describe('ServerHealthService degraded persistence', () => {
    beforeEach(() => {
        vi.resetModules();
        vi.clearAllMocks();
    });

    it('returns a concise degraded result when server lookup is unavailable', async () => {
        repositoryMock.findByUuid.mockRejectedValue(sqliteUnavailable);

        const { serverHealthService } = await import('./server-health.service.js');
        const result = await serverHealthService.checkServerHealth('server-lookup');

        expect(result).toMatchObject({
            serverUuid: 'server-lookup',
            success: false,
            errorMessage: 'Server health persistence is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('returns a concise degraded batch result when listing all servers is unavailable', async () => {
        repositoryMock.findAll.mockRejectedValue(sqliteUnavailable);

        const { serverHealthService } = await import('./server-health.service.js');
        const results = await serverHealthService.checkMultipleServers();

        expect(results).toEqual([
            expect.objectContaining({
                serverUuid: 'catalog',
                success: false,
                errorMessage: 'Server health persistence is unavailable: SQLite runtime is unavailable for this run.',
            }),
        ]);
    });
});
