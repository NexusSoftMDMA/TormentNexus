import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/index.js', () => ({
    mcpServersRepository: {
        findByUuid: vi.fn(),
    },
    oauthRepository: {
        createClient: vi.fn(),
        findClientById: vi.fn(),
    },
    oauthSessionsRepository: {
        upsert: vi.fn(),
        findByMcpServerUuid: vi.fn(),
    },
}));

const { oauthRouter } = await import('./oauthRouter.js');
const { oauthRepository, oauthSessionsRepository } = await import('../db/repositories/index.js');

function createCaller() {
    return oauthRouter.createCaller({} as never);
}

describe('oauthRouter degraded SQLite handling', () => {
    const oauthRepoMocks = vi.mocked(oauthRepository);
    const sessionRepoMocks = vi.mocked(oauthSessionsRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for client creation when SQLite is unavailable', async () => {
        oauthRepoMocks.createClient.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.clients.create({
            client_id: 'client-1',
            client_secret: null,
            client_name: 'Client 1',
            redirect_uris: ['https://example.com/callback'],
            grant_types: ['authorization_code'],
            response_types: ['code'],
            token_endpoint_auth_method: 'client_secret_post',
            scope: null,
            created_at: new Date(),
        })).rejects.toMatchObject({
            message: 'OAuth store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for session lookup when SQLite is unavailable', async () => {
        sessionRepoMocks.findByMcpServerUuid.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.sessions.getByServer({
            mcpServerUuid: '550e8400-e29b-41d4-a716-446655440000',
        })).rejects.toMatchObject({
            message: 'OAuth store is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
