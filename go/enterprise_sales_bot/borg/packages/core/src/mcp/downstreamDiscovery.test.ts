import { beforeEach, describe, expect, it, vi } from 'vitest';

const getMcpServersMock = vi.fn();
const getSessionMock = vi.fn();

vi.mock('../services/mcp-config-discovery.service.js', () => ({
    getMcpServers: (...args: unknown[]) => getMcpServersMock(...args),
}));

vi.mock('../services/mcp-server-pool.service.js', () => ({
    mcpServerPool: {
        getSession: (...args: unknown[]) => getSessionMock(...args),
    },
}));

vi.mock('../services/tool-name-parser.service.js', () => ({
    parseToolName: (name: string) => {
        const [serverName, originalToolName] = name.split('__');
        if (!serverName || !originalToolName) {
            return null;
        }
        return { serverName, originalToolName };
    },
}));

vi.mock('../services/common-utils.js', () => ({
    sanitizeName: (value: string) => value,
}));

import { listDownstreamPrompts } from './downstreamDiscovery.js';

describe('listDownstreamPrompts', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('returns promptly with empty prompts when downstream prompt listing hangs', async () => {
        getMcpServersMock.mockResolvedValue({
            alpha: {
                name: 'alpha',
                command: 'node',
                args: ['server.js'],
            },
        });

        getSessionMock.mockResolvedValue({
            client: {
                getServerVersion: () => ({ name: 'alpha' }),
                getServerCapabilities: () => ({ prompts: {} }),
                request: () => new Promise(() => {
                    // Intentionally never resolves to simulate a hung downstream MCP.
                }),
            },
        });

        const startedAt = Date.now();
        const result = await listDownstreamPrompts({
            context: {
                namespaceUuid: 'test-namespace',
                sessionId: 'test-session',
            },
            promptToClient: {},
        });
        const elapsedMs = Date.now() - startedAt;

        expect(result.prompts).toEqual([]);
        expect(elapsedMs).toBeLessThan(6_000);
    });

    it('still returns prompts from healthy downstream servers when another times out', async () => {
        getMcpServersMock.mockResolvedValue({
            healthy: {
                name: 'healthy',
                command: 'node',
                args: ['healthy.js'],
            },
            hung: {
                name: 'hung',
                command: 'node',
                args: ['hung.js'],
            },
        });

        getSessionMock.mockImplementation(async (_sessionId: string, uuid: string) => {
            if (uuid === 'healthy') {
                return {
                    client: {
                        getServerVersion: () => ({ name: 'healthy' }),
                        getServerCapabilities: () => ({ prompts: {} }),
                        request: async () => ({
                            prompts: [{ name: 'review-code', description: 'Review code quality' }],
                        }),
                    },
                };
            }

            return {
                client: {
                    getServerVersion: () => ({ name: 'hung' }),
                    getServerCapabilities: () => ({ prompts: {} }),
                    request: () => new Promise(() => {
                        // Intentionally never resolves to simulate hung prompt listing.
                    }),
                },
            };
        });

        const promptToClient: Record<string, unknown> = {};
        const result = await listDownstreamPrompts({
            context: {
                namespaceUuid: 'test-namespace',
                sessionId: 'test-session',
            },
            promptToClient: promptToClient as any,
        });

        expect(result.prompts).toHaveLength(1);
        expect(result.prompts[0]?.name).toBe('healthy__review-code');
        expect(promptToClient['healthy__review-code']).toBeTruthy();
    });
});
