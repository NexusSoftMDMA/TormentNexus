import { describe, expect, it, vi } from 'vitest';

import {
    TORMENTNEXUS_CORE_LOADER_STATUS_TOOL,
    buildCachedLoaderCatalog,
    callLoaderTool,
    createEmptyLoaderRuntimeState,
} from './stdioLoader.js';

describe('buildCachedLoaderCatalog', () => {
    it('surfaces cached tools from enabled downstream servers', () => {
        const catalog = buildCachedLoaderCatalog({
            mcpServers: {
                alpha: {
                    _meta: {
                        status: 'ready',
                        toolCount: 1,
                        discoveredAt: '2026-03-17T10:00:00.000Z',
                        tools: [
                            {
                                name: 'search_docs',
                                description: 'Search cached docs',
                                inputSchema: {
                                    type: 'object',
                                    properties: {
                                        query: { type: 'string' },
                                    },
                                },
                            },
                        ],
                    },
                },
                disabledServer: {
                    disabled: true,
                    _meta: {
                        status: 'ready',
                        toolCount: 1,
                        tools: [
                            { name: 'should_not_show' },
                        ],
                    },
                },
            },
        });

        expect(catalog.source).toBe('config');
        expect(catalog.enabledServerCount).toBe(1);
        expect(catalog.cachedToolCount).toBe(1);
        expect(catalog.tools.map((tool) => tool.name)).toEqual([
            TORMENTNEXUS_CORE_LOADER_STATUS_TOOL,
            'alpha__search_docs',
        ]);
    });
});

describe('callLoaderTool', () => {
    it('returns loader status without touching the core proxy', async () => {
        const runtimeState = createEmptyLoaderRuntimeState();
        const proxyToolCall = vi.fn();

        const result = await callLoaderTool(TORMENTNEXUS_CORE_LOADER_STATUS_TOOL, {}, runtimeState, {
            isCoreBridgeHealthy: vi.fn().mockResolvedValue(false) as any,
            proxyToolCall,
        });

        expect(proxyToolCall).not.toHaveBeenCalled();
        expect(result.isError).toBeFalsy();
        expect(result.content[0]).toMatchObject({
            type: 'text',
            text: expect.stringContaining('tormentnexus-core-stdio-loader'),
        });
    });

    it('returns a warming response when the background core is still unavailable', async () => {
        const runtimeState = createEmptyLoaderRuntimeState();

        const result = await callLoaderTool('alpha__search_docs', { query: 'mcp' }, runtimeState, {
            isCoreBridgeHealthy: vi.fn().mockResolvedValue(false) as any,
            ensureBackgroundCoreRunning: vi.fn().mockResolvedValue({
                status: 'spawned',
                pid: 4242,
                cliEntryPath: 'C:/tormentnexus/packages/cli/dist/index.js',
            }) as any,
            waitForCoreBridge: vi.fn().mockResolvedValue(false) as any,
            proxyToolCall: vi.fn(),
            now: () => '2026-03-17T12:00:00.000Z',
        });

        expect(result.isError).toBe(true);
        expect(result.content[0]).toMatchObject({
            type: 'text',
            text: expect.stringContaining('still warming'),
        });
        expect(result.content[0]).toMatchObject({
            type: 'text',
            text: expect.stringContaining('PID 4242'),
        });
    });

    it('proxies the tool call once the background core becomes healthy', async () => {
        const runtimeState = createEmptyLoaderRuntimeState();
        const proxyToolCall = vi.fn().mockResolvedValue({
            content: [{ type: 'text', text: 'ok' }],
        });

        const result = await callLoaderTool('alpha__search_docs', { query: 'mcp' }, runtimeState, {
            isCoreBridgeHealthy: vi.fn().mockResolvedValue(false) as any,
            ensureBackgroundCoreRunning: vi.fn().mockResolvedValue({
                status: 'already-running',
            }) as any,
            waitForCoreBridge: vi.fn().mockResolvedValue(true) as any,
            proxyToolCall,
        });

        expect(proxyToolCall).toHaveBeenCalledWith('alpha__search_docs', { query: 'mcp' });
        expect(result.content[0]).toMatchObject({
            type: 'text',
            text: 'ok',
        });
    });
});
