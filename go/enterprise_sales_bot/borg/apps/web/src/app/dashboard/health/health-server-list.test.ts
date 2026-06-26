import { describe, expect, it } from 'vitest';

import { getConnectedServerKeys, normalizeHealthServers } from './health-server-list';

describe('health server list helpers', () => {
    it('normalizes malformed server payloads into safe renderable entries', () => {
        const servers = normalizeHealthServers([
            null,
            123,
            {
                configKey: ' local-mcp ',
                uuid: '',
                name: ' ',
                transportType: null,
            },
            {
                uuid: 'server-2',
                configKey: 'ext-mcp',
                name: 'External MCP',
                transportType: 'sse',
            },
        ] as any);

        expect(servers).toEqual([
            {
                uuid: 'local-mcp-3',
                configKey: 'local-mcp',
                name: 'local-mcp',
                transportType: 'unknown',
            },
            {
                uuid: 'server-2',
                configKey: 'ext-mcp',
                name: 'External MCP',
                transportType: 'sse',
            },
        ]);
    });

    it('returns empty list when servers payload is not an array', () => {
        expect(normalizeHealthServers({ bad: true } as any)).toEqual([]);
        expect(normalizeHealthServers(undefined as any)).toEqual([]);
    });

    it('extracts connected server keys only from object-shaped mcp status', () => {
        expect(getConnectedServerKeys({
            servers: {
                alpha: { status: 'ok' },
                beta: { status: 'ok' },
            },
        })).toEqual(['alpha', 'beta']);

        expect(getConnectedServerKeys({ servers: [] } as any)).toEqual([]);
        expect(getConnectedServerKeys({ servers: 'oops' } as any)).toEqual([]);
        expect(getConnectedServerKeys(null as any)).toEqual([]);
    });
});
