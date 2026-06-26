import { describe, expect, it } from 'vitest';

// @ts-ignore
import { getPendingStartupChecks, getWaitingReasons } from '../../../scripts/dev_tabby_ready_helpers.mjs';

describe('dev_tabby_ready startup summaries', () => {
  it('maps pending startup checks to operator-facing labels', () => {
    expect(
      getPendingStartupChecks({
        ready: false,
        checks: {
          mcpAggregator: { ready: false },
          configSync: { ready: true },
          sessionSupervisor: { ready: false },
          browser: { ready: true },
          memory: { ready: false },
          extensionBridge: { ready: false },
        },
      }),
    ).toEqual([
      'cached MCP inventory',
      'resident MCP runtime',
      'session restore',
      'memory initialization',
      'extension bridge listener',
    ]);
  });

  it('falls back to query-level waiting reasons when no detailed startup checks are available', () => {
    expect(
      getWaitingReasons({
        web: { port: 3010 },
        coreBridge: { ok: true },
        startupStatus: { ok: true, data: { ready: false } },
        mcpStatus: { ok: false },
        memoryStatus: { ok: true },
        browserStatus: { ok: false },
        sessionStatus: { ok: false },
        extension: { ready: true, missing: [] },
      }),
    ).toEqual([
      'MCP telemetry query',
      'browser telemetry query',
      'session telemetry query',
    ]);
  });

  it('combines infrastructure and detailed startup reasons', () => {
    expect(
      getWaitingReasons({
        web: null,
        coreBridge: { ok: false },
        startupStatus: {
          ok: true,
          data: {
            ready: false,
            checks: {
              configSync: { ready: false },
              extensionBridge: { ready: false },
            },
          },
        },
        mcpStatus: { ok: true },
        memoryStatus: { ok: true },
        browserStatus: { ok: true },
        sessionStatus: { ok: true },
        extension: { ready: false, missing: ['background.js', 'popup.js'] },
      }),
    ).toEqual([
      'dashboard web server',
      'core extension bridge (/api/mesh/stream)',
      'MCP config sync',
      'extension bridge listener',
      'extension dist files (background.js, popup.js)',
    ]);
  });

  it('surfaces official Chromium and Firefox browser-extension gaps separately', () => {
    expect(
      getWaitingReasons({
        web: { port: 3000 },
        coreBridge: { ok: true },
        startupStatus: { ok: true, data: { ready: true } },
        mcpStatus: { ok: true },
        memoryStatus: { ok: true },
        browserStatus: { ok: true },
        sessionStatus: { ok: true },
        extensions: [
          {
            id: 'browser-extension-chromium',
            label: 'browser extension Chromium bundle',
            ready: false,
            missingFiles: ['background.js'],
          },
          {
            id: 'browser-extension-firefox',
            label: 'browser extension Firefox bundle',
            ready: false,
            missingFiles: ['manifest.json'],
          },
        ],
        extension: { ready: false, missing: [] },
      }),
    ).toEqual([
      'browser extension Chromium bundle (background.js)',
      'browser extension Firefox bundle (manifest.json)',
    ]);
  });
});