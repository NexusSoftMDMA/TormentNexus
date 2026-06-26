import { afterEach, describe, expect, it, vi } from 'vitest';
import { parseConfig } from '../src/config/loader.js';
import { tunnelLegacyAccessWarning } from '../src/cli/tunnel-safety.js';
import { applyTunnelConfig, printTunnelEnabledWithNoTokens } from '../src/cli/tunnel.js';

const EXISTING_TOKEN_VERIFIER = 'scrypt:v1:AAAAAAAAAAAAAAAAAAAAAA:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

afterEach(() => {
  vi.restoreAllMocks();
});

describe('tunnelLegacyAccessWarning', () => {
  it('warns when tunnel access has no scoped API tokens', () => {
    const config = parseConfig({
      version: 1,
      server: {
        access: 'tunnel',
        tunnel: {
          provider: 'cloudflare-quick',
          command: 'cloudflared tunnel --url http://127.0.0.1:{port}',
        },
      },
    });

    expect(tunnelLegacyAccessWarning(config, {})).toContain('No API tokens are configured');
  });

  it('does not warn with API-token policy or explicit unsafe override', () => {
    const config = parseConfig({
      version: 1,
      server: {
        access: 'tunnel',
        tunnel: {
          provider: 'cloudflare-quick',
          command: 'cloudflared tunnel --url http://127.0.0.1:{port}',
        },
      },
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/tmp/workspace' }],
      clients: [
        {
          id: 'codex',
          name: 'Codex',
          auth: { type: 'token', tokenHash: EXISTING_TOKEN_VERIFIER },
          permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
        },
      ],
    });

    expect(tunnelLegacyAccessWarning(config, {})).toBeUndefined();

    const legacy = parseConfig({
      version: 1,
      server: config.server,
    });
    expect(tunnelLegacyAccessWarning(legacy, { MVMT_ALLOW_LEGACY_TUNNEL: '1' })).toBeUndefined();
  });
});

describe('applyTunnelConfig', () => {
  it('enables tunnel details and warns when no API tokens exist', () => {
    const config = parseConfig({ version: 1 });
    const tunnel = {
      provider: 'custom' as const,
      command: 'cloudflared tunnel --config ~/.cloudflared/mvmt.yml run',
      url: 'https://mvmt.example.com',
    };

    const applied = applyTunnelConfig(config, tunnel, {});

    expect(applied.enabled).toBe(true);
    expect(applied.warning).toContain('No API tokens are configured');
    expect(applied.config.server.access).toBe('tunnel');
    expect(applied.config.server.tunnel).toEqual(tunnel);
  });

  it('enables tunnel access when scoped clients are configured', () => {
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/tmp/workspace' }],
      clients: [
        {
          id: 'codex',
          name: 'Codex',
          auth: { type: 'token', tokenHash: EXISTING_TOKEN_VERIFIER },
          permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
        },
      ],
    });
    const tunnel = {
      provider: 'custom' as const,
      command: 'cloudflared tunnel --config ~/.cloudflared/mvmt.yml run',
      url: 'https://mvmt.example.com',
    };

    const applied = applyTunnelConfig(config, tunnel, {});

    expect(applied.enabled).toBe(true);
    expect(applied.warning).toBeUndefined();
    expect(applied.config.server.access).toBe('tunnel');
    expect(applied.config.server.tunnel).toEqual(tunnel);
  });

  it('prints concrete next steps when tunnel has no API tokens', () => {
    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);

    printTunnelEnabledWithNoTokens('No API tokens are configured.');

    const lines = output.mock.calls.map((call) => String(call[0])).join('\n');
    expect(lines).toContain('No API tokens are configured');
    expect(lines).toContain('mvmt token add');
  });
});
