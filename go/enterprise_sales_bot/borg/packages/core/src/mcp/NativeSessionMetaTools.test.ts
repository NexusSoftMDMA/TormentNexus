import { describe, expect, it, vi } from 'vitest';
import type { Tool } from '@modelcontextprotocol/sdk/types.js';

import { NativeSessionMetaTools } from './NativeSessionMetaTools.js';
import { SessionToolWorkingSet } from './SessionToolWorkingSet.js';

function parseTextJson<T>(result: unknown): T {
    const text = (result as { content?: Array<{ type?: string; text?: string }> })
        ?.content?.find((item) => item.type === 'text')?.text ?? 'null';
    return JSON.parse(text) as T;
}

describe('NativeSessionMetaTools search auto-load', () => {
    it('auto-loads a high-confidence top search result into the working set', async () => {
        const workingSet = new SessionToolWorkingSet();
        const metaTools = new NativeSessionMetaTools(workingSet);
        const tools: Tool[] = [
            {
                name: 'browser__open_tab',
                description: 'Open a browser tab',
                inputSchema: { type: 'object', properties: {} },
            },
            {
                name: 'browser__close_tab',
                description: 'Close a browser tab',
                inputSchema: { type: 'object', properties: {} },
            },
        ];

        metaTools.refreshCatalog(tools);

        const result = await metaTools.handleToolCall('search_tools', { query: 'browser__open_tab' });
        const payload = parseTextJson<Array<{ name: string; loaded: boolean; autoLoaded?: boolean; matchReason: string }>>(result);

        expect(payload[0]).toMatchObject({
            name: 'browser__open_tab',
            loaded: true,
            autoLoaded: true,
        });
        expect(payload[0]?.matchReason).toContain('auto-loaded after exact tool name match');
        expect(workingSet.isLoaded('browser__open_tab')).toBe(true);
    });

    it('does not auto-load ambiguous search results', async () => {
        const workingSet = new SessionToolWorkingSet();
        const metaTools = new NativeSessionMetaTools(workingSet);

        metaTools.refreshCatalog([
            {
                name: 'browser__open_tab',
                description: 'Open a browser tab',
                inputSchema: { type: 'object', properties: {} },
            },
            {
                name: 'browser__open_window',
                description: 'Open a browser window',
                inputSchema: { type: 'object', properties: {} },
            },
        ]);

        const result = await metaTools.handleToolCall('search_tools', { query: 'open' });
        const payload = parseTextJson<Array<{ name: string; loaded: boolean; autoLoaded?: boolean }>>(result);

        expect(payload.some((tool) => tool.autoLoaded)).toBe(false);
        expect(workingSet.getLoadedToolNames()).toEqual([]);
    });

    it('returns JIT tool context through the native meta-tool bridge', async () => {
        const metaTools = new NativeSessionMetaTools(new SessionToolWorkingSet(), {
            toolContextResolver: ({ toolName }) => ({
                toolName,
                query: `${toolName} src/app/page.tsx`,
                matchedPaths: ['src/app/page.tsx'],
                observationCount: 1,
                summaryCount: 1,
                prompt: `JIT tool context for ${toolName}:\n- Prior note`,
            }),
        });

        const result = await metaTools.handleToolCall('get_tool_context', {
            name: 'read_file',
            arguments: { filePath: 'src/app/page.tsx' },
        });
        const payload = parseTextJson<{
            toolName: string;
            observationCount: number;
            summaryCount: number;
            prompt: string;
        }>(result);

        expect(payload).toMatchObject({
            toolName: 'read_file',
            observationCount: 1,
            summaryCount: 1,
        });
        expect(payload.prompt).toContain('Prior note');
    });

    it('refreshes loaded-tool LRU order when a loaded tool is used again', () => {
        vi.useFakeTimers();
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 3,
            maxHydratedSchemas: 2,
        });
        const metaTools = new NativeSessionMetaTools(workingSet);

        metaTools.refreshCatalog([
            { name: 'browser__alpha', description: 'Alpha tool', inputSchema: { type: 'object', properties: {} } },
            { name: 'browser__beta', description: 'Beta tool', inputSchema: { type: 'object', properties: {} } },
            { name: 'browser__gamma', description: 'Gamma tool', inputSchema: { type: 'object', properties: {} } },
            { name: 'browser__delta', description: 'Delta tool', inputSchema: { type: 'object', properties: {} } },
        ]);

        vi.setSystemTime(new Date(0));
        metaTools.loadToolIntoSession('browser__alpha');

        vi.advanceTimersByTime(10);
        metaTools.loadToolIntoSession('browser__beta');

        vi.advanceTimersByTime(10);
        metaTools.loadToolIntoSession('browser__gamma');

        vi.advanceTimersByTime(10);
        expect(metaTools.touchLoadedTool('browser__alpha')).toBe(true);

        vi.advanceTimersByTime(10);
        const { evicted } = metaTools.loadToolIntoSession('browser__delta');

        expect(evicted).toEqual(['browser__beta']);
        expect(workingSet.getLoadedToolNames().sort()).toEqual(['browser__alpha', 'browser__delta', 'browser__gamma']);
        vi.useRealTimers();
    });

    it('keeps always-loaded tools visible and protected from full unloads', async () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 1,
        });
        const metaTools = new NativeSessionMetaTools(workingSet);

        metaTools.refreshCatalog([
            { name: 'browser__pinned', description: 'Pinned browser tool', inputSchema: { type: 'object', properties: {} } },
            { name: 'browser__alpha', description: 'Alpha tool', inputSchema: { type: 'object', properties: {} } },
            { name: 'browser__beta', description: 'Beta tool', inputSchema: { type: 'object', properties: {} } },
        ]);

        metaTools.setAlwaysLoadedTools(['browser__pinned']);
        metaTools.loadToolIntoSession('browser__alpha');
        metaTools.loadToolIntoSession('browser__beta');

        expect(metaTools.getLoadedToolNames()).toEqual(['browser__pinned', 'browser__alpha', 'browser__beta']);

        const unloadResult = await metaTools.handleToolCall('unload_tool', { name: 'browser__pinned' });
        const unloadText = (unloadResult?.content?.find((item) => item.type === 'text')?.text) ?? '';

        expect(unloadText).toContain("was not loaded");
        expect(metaTools.getLoadedToolNames()).toEqual(['browser__pinned', 'browser__alpha', 'browser__beta']);

        const visibleNames = metaTools.getVisibleLoadedTools().map((tool) => tool.name);
        expect(visibleNames).toContain('browser__pinned');
    });

    it('lists all catalog tools with always-on and working-set metadata', async () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 1,
        });
        const metaTools = new NativeSessionMetaTools(workingSet);

        metaTools.refreshCatalog([
            {
                name: 'browser__pinned',
                description: 'Pinned browser tool',
                inputSchema: { type: 'object', properties: { url: { type: 'string' } } },
                alwaysOn: true,
                server: 'browser',
            } as Tool,
            {
                name: 'filesystem__read_file',
                description: 'Read a file from disk',
                inputSchema: { type: 'object', properties: { path: { type: 'string' } } },
                server: 'filesystem',
            } as Tool,
        ]);

        metaTools.setAlwaysLoadedTools(['browser__pinned']);
        metaTools.loadToolIntoSession('filesystem__read_file');

        const result = await metaTools.handleToolCall('list_all_tools', {});
        const payload = parseTextJson<{
            summary: { total: number; loaded: number; alwaysOn: number };
            tools: Array<{ name: string; alwaysOn: boolean; loaded: boolean; inputSchema: { type: string } }>;
        }>(result);

        expect(payload.summary.total).toBe(2);
        expect(payload.summary.loaded).toBe(2);
        expect(payload.summary.alwaysOn).toBe(1);
        expect(payload.tools.find((tool) => tool.name === 'browser__pinned')).toMatchObject({
            alwaysOn: true,
            loaded: true,
            inputSchema: { type: 'object' },
        });
        expect(payload.tools.find((tool) => tool.name === 'filesystem__read_file')).toMatchObject({
            alwaysOn: false,
            loaded: true,
            inputSchema: { type: 'object' },
        });
    });

    it('supports runtime capacity updates through set_capacity', async () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 4,
            maxHydratedSchemas: 2,
        });
        const metaTools = new NativeSessionMetaTools(workingSet);

        const result = await metaTools.handleToolCall('set_capacity', {
            maxLoadedTools: 6,
            maxHydratedSchemas: 3,
            idleEvictionThresholdMs: 60_000,
        });
        const text = result?.content?.find((item) => item.type === 'text')?.text ?? '';

        expect(text).toContain('maxLoadedTools=6');
        expect(text).toContain('maxHydratedSchemas=3');
        expect(text).toContain('idleEvictionThresholdMs=60000');
        expect(workingSet.getLimits()).toMatchObject({
            maxLoadedTools: 6,
            maxHydratedSchemas: 3,
            idleEvictionThresholdMs: 60_000,
        });
    });

    it('normalizes set_capacity inputs using working-set reconfigure semantics', async () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 4,
            maxHydratedSchemas: 2,
        });
        const metaTools = new NativeSessionMetaTools(workingSet);

        await metaTools.handleToolCall('set_capacity', {
            maxLoadedTools: 0.2,
            maxHydratedSchemas: 0,
            idleEvictionThresholdMs: -15,
        });

        expect(workingSet.getLimits()).toMatchObject({
            maxLoadedTools: 1,
            maxHydratedSchemas: 1,
            idleEvictionThresholdMs: 0,
        });
    });

    it('returns and clears working-set eviction history', async () => {
        const workingSet = new SessionToolWorkingSet({
            maxLoadedTools: 2,
            maxHydratedSchemas: 1,
        });
        const metaTools = new NativeSessionMetaTools(workingSet);

        metaTools.refreshCatalog([
            { name: 'browser__one', description: 'One', inputSchema: { type: 'object', properties: {} } },
            { name: 'browser__two', description: 'Two', inputSchema: { type: 'object', properties: {} } },
            { name: 'browser__three', description: 'Three', inputSchema: { type: 'object', properties: {} } },
        ]);

        metaTools.loadToolIntoSession('browser__one');
        metaTools.loadToolIntoSession('browser__two');
        metaTools.loadToolIntoSession('browser__three');

        const historyResult = await metaTools.handleToolCall('get_eviction_history', {});
        const history = parseTextJson<Array<{ toolName: string; tier: string }>>(historyResult);

        expect(history.length).toBeGreaterThan(0);
        expect(history[0]?.toolName).toBe('browser__one');
        expect(history[0]?.tier).toBe('loaded');

        const clearResult = await metaTools.handleToolCall('clear_eviction_history', {});
        const clearText = clearResult?.content?.find((item) => item.type === 'text')?.text ?? '';
        expect(clearText).toContain('Cleared');

        const afterClear = parseTextJson<Array<unknown>>(await metaTools.handleToolCall('get_eviction_history', {}));
        expect(afterClear).toEqual([]);
    });
});
