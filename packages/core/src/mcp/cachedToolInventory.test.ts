import { describe, expect, it, vi } from 'vitest';
import { getCachedToolInventory } from './cachedToolInventory.js';
import * as mcpJsonConfig from './mcpJsonConfig.js';
import * as dbRepos from '../db/repositories/index.js';

vi.mock('./mcpJsonConfig.js', async (importOriginal) => {
    const actual = await importOriginal<typeof import('./mcpJsonConfig.js')>();
    return {
        ...actual,
        loadTormentNexusMcpConfig: vi.fn(),
    };
});

vi.mock('../db/repositories/index.js', () => ({
    mcpServersRepository: {
        findAll: vi.fn(),
    },
    toolsRepository: {
        findAll: vi.fn(),
    }
}));

describe('getCachedToolInventory', () => {
    it('merges servers and tools from both mcp.jsonc and SQLite successfully', async () => {
        vi.mocked(mcpJsonConfig.loadTormentNexusMcpConfig).mockResolvedValue({
            mcpServers: {
                manual_server: {
                    command: 'node',
                    args: ['manual.js'],
                    _meta: {
                        status: 'ready',
                        toolCount: 1,
                        tools: [
                            {
                                name: 'manual_tool',
                                description: 'A manual tool',
                                alwaysOn: true,
                            } as any
                        ]
                    } as any
                }
            }
        });

        vi.mocked(dbRepos.mcpServersRepository.findAll).mockResolvedValue([
            {
                uuid: 'db-123',
                name: 'db_server',
                always_on: 1,
            } as any
        ]);

        vi.mocked(dbRepos.toolsRepository.findAll).mockResolvedValue([
            {
                mcp_server_uuid: 'db-123',
                name: 'db_tool',
                description: 'A DB tool',
                always_on: 1,
            } as any
        ]);

        const inventory = await getCachedToolInventory();
        
        expect(inventory.databaseAvailable).toBe(true);
        expect(inventory.fallbackUsed).toBe(false);
        expect(inventory.servers).toHaveLength(2); // 1 manual, 1 db
        expect(inventory.tools).toHaveLength(2);   // 1 manual, 1 db
        
        const toolNames = inventory.tools.map(t => t.name).sort();
        expect(toolNames).toEqual(['db_server__db_tool', 'manual_server__manual_tool']);
    });

    it('gracefully handles discovery failures (SQLite unavailable) and falls back to config', async () => {
        vi.mocked(mcpJsonConfig.loadTormentNexusMcpConfig).mockResolvedValue({
            mcpServers: {
                manual_server: {
                    command: 'node',
                    args: ['manual.js'],
                    _meta: {
                        status: 'ready',
                        toolCount: 1,
                        tools: [
                            {
                                name: 'manual_tool',
                                description: 'A manual tool',
                                alwaysOn: true,
                            } as any
                        ]
                    } as any
                }
            }
        });

        // Simulate SQLite throw
        vi.mocked(dbRepos.mcpServersRepository.findAll).mockRejectedValue(new Error('SqliteError: disk I/O error'));

        const inventory = await getCachedToolInventory();
        
        expect(inventory.databaseAvailable).toBe(false);
        expect(inventory.fallbackUsed).toBe(true);
        expect(inventory.servers).toHaveLength(1);
        expect(inventory.tools).toHaveLength(1);
        expect(inventory.databaseError).toContain('disk I/O error');
    });

    it('returns empty cache if both sources are empty', async () => {
        vi.mocked(mcpJsonConfig.loadTormentNexusMcpConfig).mockResolvedValue({ mcpServers: {} });
        vi.mocked(dbRepos.mcpServersRepository.findAll).mockResolvedValue([]);
        vi.mocked(dbRepos.toolsRepository.findAll).mockResolvedValue([]);

        const inventory = await getCachedToolInventory();
        
        expect(inventory.source).toBe('empty');
        expect(inventory.servers).toHaveLength(0);
        expect(inventory.tools).toHaveLength(0);
    });
});
