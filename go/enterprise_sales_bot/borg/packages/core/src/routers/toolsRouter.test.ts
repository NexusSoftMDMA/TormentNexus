import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../db/repositories/index.js', () => ({
    toolsRepository: {
        findAll: vi.fn(),
        findByMcpServerUuid: vi.fn(),
        create: vi.fn(),
        bulkUpsert: vi.fn(),
        setAlwaysOn: vi.fn(),
    },
    mcpServersRepository: {
        findAll: vi.fn(),
        findByUuid: vi.fn(),
    },
}));

vi.mock('../services/ToolRegistry.js', () => ({
    toolRegistry: {
        deleteTool: vi.fn(),
    },
    RegisteredTool: class RegisteredTool {},
}));

const { toolsRouter } = await import('./toolsRouter.js');
const { toolsRepository } = await import('../db/repositories/index.js');

function createCaller() {
    return toolsRouter.createCaller({} as never);
}

describe('toolsRouter degraded SQLite handling', () => {
    const repositoryMocks = vi.mocked(toolsRepository);

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('surfaces a clear error for list when SQLite is unavailable', async () => {
        repositoryMocks.findAll.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.list()).rejects.toMatchObject({
            message: 'Tool catalog is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('surfaces a clear error for create when SQLite is unavailable', async () => {
        repositoryMocks.create.mockRejectedValue(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const caller = createCaller();

        await expect(caller.create({
            name: 'search_tools',
            description: 'Search tools',
            toolSchema: { type: 'object', properties: {} },
            mcp_server_uuid: 'server-1',
        })).rejects.toMatchObject({
            message: 'Tool catalog is unavailable: SQLite runtime is unavailable for this run.',
        });
    });
});
