import { describe, expect, it, vi, beforeEach } from 'vitest';
import { initializeConnectors, formatConnectorError } from '../src/cli/connector-loader.js';

describe('connector-loader', () => {
  let logger: any;

  beforeEach(() => {
    logger = { info: vi.fn(), warn: vi.fn(), error: vi.fn() };
    vi.clearAllMocks();
  });

  describe('formatConnectorError', () => {
    it('returns raw message for normal errors', () => {
      expect(formatConnectorError(new Error('raw error'))).toBe('raw error');
    });

    it('extracts validation paths from JSON error arrays', () => {
      const jsonError = JSON.stringify([
        { path: ['tools', 0, 'name'], message: 'invalid' },
        { path: ['capabilities'], message: 'missing' }
      ]);
      expect(formatConnectorError(new Error(jsonError))).toContain('upstream server returned invalid MCP schema at tools.0.name, capabilities');
    });
  });

  describe('initializeConnectors', () => {
    it('ignores legacy proxy connectors', async () => {
      const config = { proxy: [{ name: 'test', enabled: true }] } as any;

      const loaded = await initializeConnectors(config, false, logger);
      expect(loaded).toHaveLength(0);
      expect(logger.warn).toHaveBeenCalledWith(expect.stringContaining('Legacy proxy connectors are ignored'));
    });

    it('does not warn when no legacy proxy connectors are enabled', async () => {
      const config = { proxy: [{ name: 'test', enabled: false }] } as any;

      const loaded = await initializeConnectors(config, false, logger);
      expect(loaded).toHaveLength(0);
      expect(logger.warn).not.toHaveBeenCalled();
    });
  });
});
