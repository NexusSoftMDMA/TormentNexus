import { afterEach, describe, expect, it, vi } from 'vitest';

import { isMcpServerDebugEnabled, mcpServerDebugLog } from './mcpServerDebug.js';

describe('mcpServerDebug', () => {
    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('stays quiet by default for normal dev boot', () => {
        expect(isMcpServerDebugEnabled({})).toBe(false);
        expect(isMcpServerDebugEnabled({ TORMENTNEXUS_MCP_SERVER_DEBUG: '0' })).toBe(false);
    });

    it('enables logging via the dedicated TormentNexus flag or DEBUG namespace', () => {
        expect(isMcpServerDebugEnabled({ TORMENTNEXUS_MCP_SERVER_DEBUG: '1' })).toBe(true);
        expect(isMcpServerDebugEnabled({ TORMENTNEXUS_MCP_SERVER_DEBUG: 'true' })).toBe(true);
        expect(isMcpServerDebugEnabled({ DEBUG: 'other,tormentnexus:mcp-server' })).toBe(true);
    });

    it('only writes to console when debug is enabled', () => {
        const logSpy = vi.spyOn(console, 'log').mockImplementation(() => undefined);

        mcpServerDebugLog('[DEBUG] hidden by default', {});
        expect(logSpy).not.toHaveBeenCalled();

        mcpServerDebugLog('[DEBUG] visible when enabled', { TORMENTNEXUS_MCP_SERVER_DEBUG: '1' });
        expect(logSpy).toHaveBeenCalledWith('[DEBUG] visible when enabled');
    });
});