import { describe, expect, it } from 'vitest';
import fssync from 'fs';
import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { expandHome, loadConfig, parseConfig, readConfig, saveConfig, withConfigLock } from '../src/config/loader.js';
import {
  DEFAULT_MOUNT_EXCLUDE_PATTERNS,
  DEFAULT_MOUNT_PROTECT_PATTERNS,
  MvmtConfig,
  resolveProxySourceId,
} from '../src/config/schema.js';

const SCRYPT_TOKEN_VERIFIER = 'scrypt:v1:AAAAAAAAAAAAAAAAAAAAAA:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';
const SCRYPT_TOKEN_VERIFIER_2 = 'scrypt:v1:BBBBBBBBBBBBBBBBBBBBBB:BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB';

describe('saveConfig', () => {
  it('writes a config file that loadConfig can read back', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-save-config-'));
    const configPath = path.join(dir, 'config.yaml');
    const config = parseConfig({ version: 1, server: { port: 4242 } });

    await saveConfig(configPath, config);
    const loaded = loadConfig(configPath);
    expect(loaded.server.port).toBe(4242);

    await fs.rm(dir, { recursive: true, force: true });
  });

  it('creates the parent directory if it does not exist', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-save-dir-'));
    const nestedPath = path.join(dir, 'nested', 'config.yaml');
    const config = parseConfig({ version: 1 });

    await saveConfig(nestedPath, config);
    expect(fssync.existsSync(nestedPath)).toBe(true);

    await fs.rm(dir, { recursive: true, force: true });
  });

  it('preserves valid config structure (round-trip test)', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-round-trip-'));
    const configPath = path.join(dir, 'config.yaml');
    const config: MvmtConfig = parseConfig({
      version: 1,
      server: { port: 5000 },
      proxy: [{ name: 'test-proxy', transport: 'http', url: 'http://localhost:8080' }]
    });

    await saveConfig(configPath, config);
    const raw = readConfig(configPath);
    const loaded = parseConfig(raw);

    expect(loaded).toEqual(config);

    await fs.rm(dir, { recursive: true, force: true });
  });

  it('on non-Windows, writes with mode 0o600', async () => {
    if (process.platform === 'win32') return;

    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-chmod-test-'));
    const configPath = path.join(dir, 'config.yaml');
    const config = parseConfig({ version: 1 });

    await saveConfig(configPath, config);
    const stat = await fs.stat(configPath);
    expect(stat.mode & 0o777).toBe(0o600);

    await fs.rm(dir, { recursive: true, force: true });
  });

  it('serializes config mutations with the config lock', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-config-lock-'));
    const configPath = path.join(dir, 'config.yaml');
    let active = 0;
    let maxActive = 0;

    try {
      await Promise.all([
        withConfigLock(configPath, async () => {
          active += 1;
          maxActive = Math.max(maxActive, active);
          await new Promise((resolve) => setTimeout(resolve, 50));
          active -= 1;
        }),
        withConfigLock(configPath, async () => {
          active += 1;
          maxActive = Math.max(maxActive, active);
          active -= 1;
        }),
      ]);

      expect(maxActive).toBe(1);
    } finally {
      await fs.rm(dir, { recursive: true, force: true });
    }
  });
});

describe('parseConfig', () => {
  it('applies schema defaults', () => {
    const config = parseConfig({ version: 1 });

    expect(config.server.port).toBe(4141);
    expect(config.server.access).toBe('local');
    expect(config.proxy).toEqual([]);
    expect(config.mounts).toEqual([]);
  });

  it('parses local folder mounts for the prototype text index', () => {
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '~/code/mvmt', writeAccess: true }],
    });

    expect(config.mounts[0]).toMatchObject({
      name: 'workspace',
      type: 'local_folder',
      path: '/workspace',
      root: '~/code/mvmt',
      description: '',
      guidance: '',
      exclude: [...DEFAULT_MOUNT_EXCLUDE_PATTERNS],
      protect: [...DEFAULT_MOUNT_PROTECT_PATTERNS],
      writeAccess: true,
      enabled: true,
    });
  });

  it('parses mount descriptions and guidance', () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        {
          name: 'notes',
          type: 'local_folder',
          path: '/notes',
          root: '~/notes',
          description: 'Personal notes vault.',
          guidance: 'Use for reference. Do not write unless asked.',
        },
      ],
    });

    expect(config.mounts[0]).toMatchObject({
      description: 'Personal notes vault.',
      guidance: 'Use for reference. Do not write unless asked.',
    });
  });

  it('rejects duplicate source ids across mounts and proxy sources', () => {
    expect(() =>
      parseConfig({
        version: 1,
        proxy: [{ id: 'workspace', name: 'filesystem', command: 'npx' }],
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/workspace' }],
      }),
    ).toThrow(/duplicate sourceId "workspace"/);
  });

  it('rejects duplicate mount paths', () => {
    expect(() =>
      parseConfig({
        version: 1,
        mounts: [
          { name: 'notes', type: 'local_folder', path: '/data', root: '/notes' },
          { name: 'archive', type: 'local_folder', path: '/data', root: '/archive' },
        ],
      }),
    ).toThrow(/duplicate mount path "\/data"/);
  });

  it('parses tunnel server access config', () => {
    const config = parseConfig({
      version: 1,
      server: {
        access: 'tunnel',
        tunnel: {
          provider: 'cloudflare-quick',
          command: 'cloudflared tunnel --url http://127.0.0.1:{port}',
          url: 'https://demo.trycloudflare.com',
        },
      },
    });

    expect(config.server.tunnel).toMatchObject({
      provider: 'cloudflare-quick',
      command: 'cloudflared tunnel --url http://127.0.0.1:{port}',
      url: 'https://demo.trycloudflare.com',
    });
  });

  it('parses relay tunnel config', () => {
    const config = parseConfig({
      version: 1,
      server: {
        access: 'tunnel',
        tunnel: {
          provider: 'relay',
          relayUrl: 'wss://relay.example.com/connect',
          workspaceSlug: 'demo-workspace',
          agentToken: 'agent-secret',
          url: 'https://demo.example.com',
        },
      },
    });

    expect(config.server.tunnel).toMatchObject({
      provider: 'relay',
      relayUrl: 'wss://relay.example.com/connect',
      workspaceSlug: 'demo-workspace',
      agentToken: 'agent-secret',
      url: 'https://demo.example.com',
    });
  });

  it('defaults stdio proxy transport', () => {
    const config = parseConfig({
      version: 1,
      proxy: [{ name: 'filesystem', command: 'npx' }],
    });

    expect(config.proxy[0]).toMatchObject({
      name: 'filesystem',
      transport: 'stdio',
      command: 'npx',
      args: [],
      env: {},
      enabled: true,
    });
  });

  it('parses proxy write access policy', () => {
    const config = parseConfig({
      version: 1,
      proxy: [{ name: 'filesystem', command: 'npx', writeAccess: true }],
    });

    expect(config.proxy[0].writeAccess).toBe(true);
  });

  it('tolerates legacy proxy source metadata', () => {
    const config = parseConfig({
      version: 1,
      proxy: [{ name: 'filesystem', source: 'manual', command: 'npx' }],
    });

    expect(config.proxy[0]).toMatchObject({
      name: 'filesystem',
      source: 'manual',
      command: 'npx',
    });
  });

  it('rejects invalid ports', () => {
    expect(() => parseConfig({ version: 1, server: { port: 70000 } })).toThrow('Invalid config');
  });

  it('rejects tunnel access without a tunnel command', () => {
    expect(() => parseConfig({ version: 1, server: { access: 'tunnel' } })).toThrow(
      'tunnel access requires "tunnel" config',
    );
  });

  it('rejects transport configs without required command or url', () => {
    expect(() => parseConfig({ version: 1, proxy: [{ name: 'bad', transport: 'stdio' }] })).toThrow(
      'stdio transport requires "command"',
    );
    expect(() => parseConfig({ version: 1, proxy: [{ name: 'bad', transport: 'http' }] })).toThrow(
      'http transport requires "url"',
    );
  });
});

describe('expandHome', () => {
  it('expands tilde paths', () => {
    expect(expandHome('~/notes')).not.toContain('~');
  });

  it('leaves normal paths unchanged', () => {
    expect(expandHome('/tmp/notes')).toBe('/tmp/notes');
  });
});

describe('loadConfig', () => {
  it('loads YAML config from an override path', async () => {
    const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-config-'));
    const configPath = path.join(dir, 'config.yaml');
    await fs.writeFile(
      configPath,
      ['version: 1', 'server:', '  port: 4142', 'proxy: []'].join('\n'),
      'utf-8',
    );

    expect(loadConfig(configPath)).toMatchObject({
      version: 1,
      server: { port: 4142 },
      proxy: [],
    });
  });
});

describe('client policy schema', () => {
  it('treats clients and semanticTools as optional (no behavior change for existing configs)', () => {
    const config = parseConfig({ version: 1 });
    expect(config.clients).toBeUndefined();
    expect(config.semanticTools).toBeUndefined();
  });

  it('parses a token-auth client with path permissions', () => {
    const config = parseConfig({
      version: 1,
      proxy: [{ name: 'workspace', command: 'npx' }],
      mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: '/vault' }],
      clients: [
        {
          id: 'codex',
          name: 'Codex CLI',
          createdAt: '2026-04-29T12:00:00.000Z',
          lastUsedAt: '2026-04-29T12:30:00.000Z',
          credentialVersion: 3,
          clientBinding: 'codex',
          auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
          rawToolsEnabled: true,
          permissions: [
            { path: '/workspace/**', actions: ['search', 'read', 'write'] },
            { path: '/notes/**', actions: ['search', 'read'] },
          ],
        },
      ],
    });

    expect(config.clients).toHaveLength(1);
    expect(config.clients?.[0]).toMatchObject({
      id: 'codex',
      name: 'Codex CLI',
      createdAt: '2026-04-29T12:00:00.000Z',
      lastUsedAt: '2026-04-29T12:30:00.000Z',
      credentialVersion: 3,
      clientBinding: 'codex',
      auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
      rawToolsEnabled: true,
    });
  });

  it('rejects legacy SHA-256 token verifiers', () => {
    expect(() =>
      parseConfig({
        version: 1,
        clients: [
          {
            id: 'codex',
            name: 'Codex CLI',
            auth: { type: 'token', tokenHash: 'a'.repeat(64) },
          },
        ],
      }),
    ).toThrow(/scrypt verifier/);
  });

  it('parses an oauth-auth client with mapped client ids', () => {
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: '/vault' }],
      clients: [
        {
          id: 'chatgpt',
          name: 'ChatGPT',
          auth: { type: 'oauth', oauthClientIds: ['chatgpt-mvmt', 'chatgpt-mvmt-v2'] },
          permissions: [{ path: '/notes/**', actions: ['search', 'read'] }],
        },
      ],
    });

    expect(config.clients?.[0].auth).toEqual({
      type: 'oauth',
      oauthClientIds: ['chatgpt-mvmt', 'chatgpt-mvmt-v2'],
    });
    expect(config.clients?.[0].rawToolsEnabled).toBe(false);
  });

  it('rejects invalid client permission paths', () => {
    expect(() =>
      parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/workspace' }],
        clients: [
          {
            id: 'codex',
            name: 'Codex',
            auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
            permissions: [{ path: 'relative/path', actions: ['read'] }],
          },
        ],
      }),
    ).toThrow(/permission path must be absolute/);

    expect(() =>
      parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/workspace' }],
        clients: [
          {
            id: 'codex',
            name: 'Codex',
            auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
            permissions: [{ path: '/workspace/**/secrets', actions: ['read'] }],
          },
        ],
      }),
    ).toThrow(/permission path may only use \/\*\* as a trailing subtree wildcard/);

    expect(() =>
      parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/workspace' }],
        clients: [
          {
            id: 'codex',
            name: 'Codex',
            auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
            permissions: [{ path: '/workspace/*.md', actions: ['read'] }],
          },
        ],
      }),
    ).toThrow(/permission path may only use \/\*\* as a trailing subtree wildcard/);
  });

  it('rejects empty client permission actions', () => {
    expect(() =>
      parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/workspace' }],
        clients: [
          {
            id: 'codex',
            name: 'Codex',
            auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
            permissions: [{ path: '/workspace/**', actions: [] }],
          },
        ],
      }),
    ).toThrow(/permission actions must include at least one action/);
  });

  it('rejects permission paths that do not target a configured mount or proxy source', () => {
    expect(() =>
      parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/workspace' }],
        clients: [
          {
            id: 'codex',
            name: 'Codex',
            auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
            permissions: [{ path: '/notes/**', actions: ['read'] }],
          },
        ],
      }),
    ).toThrow(/permission path "\/notes\/\*\*" does not target any configured mount or proxy source/);
  });

  it('allows global and nested mount permissions', () => {
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'project', type: 'local_folder', path: '/workspace/project', root: '/workspace/project' }],
      clients: [
        {
          id: 'codex',
          name: 'Codex',
          auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
          permissions: [
            { path: '/workspace/project/src/**', actions: ['read'] },
            { path: '/**', actions: ['search'] },
          ],
        },
      ],
    });

    expect(config.clients?.[0].permissions).toEqual([
      { path: '/workspace/project/src/**', actions: ['read'] },
      { path: '/**', actions: ['search'] },
    ]);
  });

  it('rejects wildcard mount paths', () => {
    expect(() =>
      parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace/*', root: '/workspace' }],
      }),
    ).toThrow(/mount path must be a literal virtual path/);
  });

  it('rejects duplicate client ids', () => {
    expect(() =>
      parseConfig({
        version: 1,
        clients: [
          { id: 'codex', name: 'A', auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER } },
          { id: 'codex', name: 'B', auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER_2 } },
        ],
      }),
    ).toThrow(/duplicate client id "codex"/);
  });

  it('rejects duplicate tokenHash across clients (config order would otherwise be a security decision)', () => {
    expect(() =>
      parseConfig({
        version: 1,
        clients: [
          { id: 'codex', name: 'Codex', auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER } },
          { id: 'cursor', name: 'Cursor', auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER } },
        ],
      }),
    ).toThrow(/duplicate tokenHash/);
  });

  it('rejects duplicate oauthClientIds across OAuth clients', () => {
    expect(() =>
      parseConfig({
        version: 1,
        clients: [
          {
            id: 'chatgpt',
            name: 'ChatGPT',
            auth: { type: 'oauth', oauthClientIds: ['mvmt-shared'] },
          },
          {
            id: 'claude',
            name: 'Claude',
            auth: { type: 'oauth', oauthClientIds: ['mvmt-shared', 'claude-only'] },
          },
        ],
      }),
    ).toThrow(/oauth client_id "mvmt-shared" is already mapped/);
  });

  it('rejects an OAuth client_id duplicated within the same client', () => {
    expect(() =>
      parseConfig({
        version: 1,
        clients: [
          {
            id: 'chatgpt',
            name: 'ChatGPT',
            auth: { type: 'oauth', oauthClientIds: ['mvmt-shared', 'mvmt-shared'] },
          },
        ],
      }),
    ).toThrow(/oauth client_id "mvmt-shared" is already mapped/);
  });

  it('rejects malformed client id', () => {
    expect(() =>
      parseConfig({
        version: 1,
        clients: [
          {
            id: 'Has Spaces',
            name: 'bad',
            auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
          },
        ],
      }),
    ).toThrow(/lowercase alphanum/);
  });

  it('rejects malformed tokenHash', () => {
    expect(() =>
      parseConfig({
        version: 1,
        clients: [
          {
            id: 'codex',
            name: 'Codex',
            auth: { type: 'token', tokenHash: 'plaintext-token' },
          },
        ],
      }),
    ).toThrow(/scrypt verifier/);
  });

  it('parses semanticTools and validates source references', () => {
    const config = parseConfig({
      version: 1,
      proxy: [{ name: 'workspace', command: 'npx' }],
      mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: '/vault' }],
      semanticTools: {
        searchPersonalContext: { enabled: true, sourceIds: ['workspace', 'notes'] },
        readContextItem: { sourceIds: ['workspace'] },
      },
    });

    expect(config.semanticTools?.searchPersonalContext?.sourceIds).toEqual(['workspace', 'notes']);
    expect(config.semanticTools?.readContextItem?.enabled).toBe(true);
  });

  it('rejects unknown sourceId in semanticTools', () => {
    expect(() =>
      parseConfig({
        version: 1,
        semanticTools: {
          searchPersonalContext: { sourceIds: ['nonexistent'] },
        },
      }),
    ).toThrow(/unknown sourceId "nonexistent"/);
  });

  it('uses proxy.id when set, falling back to name for source resolution', () => {
    const config = parseConfig({
      version: 1,
      proxy: [
        { id: 'workspace', name: 'filesystem', command: 'npx' },
      ],
      clients: [
        {
          id: 'codex',
          name: 'Codex',
          auth: { type: 'token', tokenHash: SCRYPT_TOKEN_VERIFIER },
          permissions: [{ path: '/workspace/**', actions: ['read'] }],
        },
      ],
    });

    expect(resolveProxySourceId(config.proxy[0])).toBe('workspace');
    expect(config.clients?.[0].permissions[0].path).toBe('/workspace/**');
  });

  it('falls back to proxy.name when id is omitted', () => {
    const config = parseConfig({
      version: 1,
      proxy: [{ name: 'search', command: '/usr/bin/search-server' }],
    });
    expect(resolveProxySourceId(config.proxy[0])).toBe('search');
  });
});
