import { TRPCError } from '@trpc/server';
import { afterEach, describe, expect, it, vi } from 'vitest';

const aggregatorMock = {
    listServers: vi.fn(async () => []),
    listAggregatedTools: vi.fn(async () => []),
    getInitializationStatus: vi.fn(() => ({ initialized: true, inProgress: false, connectedClientCount: 0, configuredServerCount: 0 })),
    getTrafficEvents: vi.fn(async () => []),
};

const serverMock = {
    executeTool: vi.fn(async () => ({ content: [] })),
};

vi.mock('../mcp/cachedToolInventory.js', () => ({
    getCachedToolInventory: vi.fn(),
}));

vi.mock('../lib/trpc-core.js', async () => {
    const actual = await vi.importActual<typeof import('../lib/trpc-core.js')>('../lib/trpc-core.js');
    return {
        ...actual,
        getMcpAggregator: () => aggregatorMock,
        getMcpServer: () => serverMock,
    };
});

describe('mcpRouter degraded inventory handling', () => {
    afterEach(() => {
        vi.restoreAllMocks();
        aggregatorMock.listServers.mockResolvedValue([]);
        aggregatorMock.listAggregatedTools.mockResolvedValue([]);
        aggregatorMock.getTrafficEvents.mockResolvedValue([]);
        serverMock.executeTool.mockResolvedValue({ content: [] });
    });

    it('returns a concise TRPC error when listServers cannot read persisted inventory', async () => {
        const { getCachedToolInventory } = await import('../mcp/cachedToolInventory.js');
        vi.mocked(getCachedToolInventory).mockRejectedValueOnce(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.listServers()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP inventory is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('returns a concise TRPC error when getStatus cannot read persisted inventory', async () => {
        const { getCachedToolInventory } = await import('../mcp/cachedToolInventory.js');
        vi.mocked(getCachedToolInventory).mockRejectedValueOnce(
            new Error('SQLite runtime is unavailable for TormentNexus DB-backed features (Could not locate the bindings file. Tried: better-sqlite3.node)'),
        );

        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.getStatus()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP inventory is unavailable: SQLite runtime is unavailable for this run.',
        });
    });

    it('returns a concise TRPC error when getStatus cannot read live runtime status', async () => {
        aggregatorMock.listServers.mockRejectedValueOnce(new Error('aggregator unavailable'));
        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.getStatus()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP status is unavailable: aggregator unavailable',
        });
    });

    it('returns a concise TRPC error when traffic cannot be read', async () => {
        aggregatorMock.getTrafficEvents.mockRejectedValueOnce(new Error('traffic stream offline'));
        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.traffic()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP traffic is unavailable: traffic stream offline',
        });
    });

    it('returns a concise TRPC error when eviction history cannot be read', async () => {
        serverMock.executeTool.mockRejectedValueOnce(new Error('eviction tracker offline'));
        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.getWorkingSetEvictionHistory()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP eviction history is unavailable: eviction tracker offline',
        });
    });

    it('returns a concise TRPC error when eviction history cannot be cleared', async () => {
        serverMock.executeTool.mockRejectedValueOnce(new Error('eviction tracker offline'));
        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.clearWorkingSetEvictionHistory()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'Failed to clear eviction history: eviction tracker offline',
        });
    });

    it('returns a concise TRPC error when listServers cannot read runtime inventory for non-SQLite failures', async () => {
        const { getCachedToolInventory } = await import('../mcp/cachedToolInventory.js');
        vi.mocked(getCachedToolInventory).mockRejectedValueOnce(new Error('cache file unreadable'));

        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.listServers()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP inventory is unavailable: cache file unreadable',
        });
    });

    it('returns a concise TRPC error when listTools cannot read runtime inventory for non-SQLite failures', async () => {
        const { getCachedToolInventory } = await import('../mcp/cachedToolInventory.js');
        vi.mocked(getCachedToolInventory).mockRejectedValueOnce(new Error('tool cache unreadable'));

        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.listTools()).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP inventory is unavailable: tool cache unreadable',
        });
    });

    it('returns a concise TRPC error when searchTools cannot query the live aggregator', async () => {
        const { getCachedToolInventory } = await import('../mcp/cachedToolInventory.js');
        vi.mocked(getCachedToolInventory).mockResolvedValueOnce({ servers: [], tools: [], toolCounts: new Map() });
        aggregatorMock.searchTools = vi.fn(async () => {
            throw new Error('search backend offline');
        });

        const { mcpRouter } = await import('./mcpRouter.js');
        const caller = mcpRouter.createCaller({});

        await expect(caller.searchTools({ query: 'browser tool' })).rejects.toMatchObject<Partial<TRPCError>>({
            message: 'MCP tool search is unavailable: search backend offline',
        });
    });
});
