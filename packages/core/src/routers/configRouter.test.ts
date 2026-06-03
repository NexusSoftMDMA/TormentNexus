import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/index.js', () => ({
    configRepo: {
        findAll: vi.fn(),
        get: vi.fn(),
        set: vi.fn(),
        delete: vi.fn(),
    },
}));

vi.mock('../services/config.service.js', () => ({
    configService: {
        getMcpTimeout: vi.fn(),
        setMcpTimeout: vi.fn(),
        getMcpMaxAttempts: vi.fn(),
        setMcpMaxAttempts: vi.fn(),
        getMcpMaxTotalTimeout: vi.fn(),
        setMcpMaxTotalTimeout: vi.fn(),
        getMcpResetTimeoutOnProgress: vi.fn(),
        setMcpResetTimeoutOnProgress: vi.fn(),
        getSessionLifetime: vi.fn(),
        setSessionLifetime: vi.fn(),
        isSignupDisabled: vi.fn(),
        setSignupDisabled: vi.fn(),
        isSsoSignupDisabled: vi.fn(),
        setSsoSignupDisabled: vi.fn(),
        isBasicAuthDisabled: vi.fn(),
        setBasicAuthDisabled: vi.fn(),
        getAuthProviders: vi.fn(),
    },
}));

vi.mock('../services/config/JsonConfigProvider.js', () => ({
    jsonConfigProvider: {
        loadAlwaysVisibleTools: vi.fn(),
        saveAlwaysVisibleTools: vi.fn(),
    },
}));

vi.mock('../mcp/mcpJsonConfig.js', () => ({
    getTormentNexusConfigDir: vi.fn(() => 'C:\\Users\\hyper\\.tormentnexus'),
    writeTormentNexusMcpConfig: vi.fn(),
}));

const { configRouter } = await import('./configRouter.js');
const { configRepo } = await import('../db/repositories/index.js');
const { configService } = await import('../services/config.service.js');
const { writeTormentNexusMcpConfig } = await import('../mcp/mcpJsonConfig.js');

function createCaller() {
    return configRouter.createCaller({} as never);
}

describe('configRouter degraded SQLite handling', () => {
    const repoMocks = vi.mocked(configRepo);
    const serviceMocks = vi.mocked(configService);
    const writeConfigMock = vi.mocked(writeTormentNexusMcpConfig);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repoMocks.findAll.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list()).rejects.toMatchObject({
            message: 'Configuration store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for getMcpTimeout when SQLite is unavailable', async () => {
        serviceMocks.getMcpTimeout.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.getMcpTimeout()).rejects.toMatchObject({
            message: 'Configuration store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('resets only matching config keys for a section', async () => {
        repoMocks.findAll.mockResolvedValue({
            'server.port': '4000',
            'server.host': '127.0.0.1',
            'mcp.timeout': '60000',
        });

        const caller = createCaller();
        const result = await caller.reset({ section: 'server' });

        expect(repoMocks.delete).toHaveBeenCalledTimes(2);
        expect(repoMocks.delete).toHaveBeenNthCalledWith(1, 'server.port');
        expect(repoMocks.delete).toHaveBeenNthCalledWith(2, 'server.host');
        expect(result).toEqual({
            success: true,
            section: 'server',
            removed: 2,
            keys: ['server.port', 'server.host'],
        });
    });

    it('initializes local config files', async () => {
        const caller = createCaller();
        const result = await caller.init({ scope: 'local' });

        expect(writeConfigMock).toHaveBeenCalledWith({ mcpServers: {} }, expect.stringMatching(/\.tormentnexus$/));
        expect(result).toEqual({
            success: true,
            scope: 'local',
            path: expect.stringMatching(/\.tormentnexus$/),
        });
    });
});
