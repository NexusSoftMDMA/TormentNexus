import { execFile, spawn, type ChildProcessWithoutNullStreams } from 'child_process';
import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { promisify } from 'util';
import { describe, expect, it } from 'vitest';
import { parseConfig, readConfig, saveConfig } from '../src/config/loader.js';

const execFileAsync = promisify(execFile);
const root = path.resolve(import.meta.dirname, '..');
const cliArgs = ['--import', 'tsx', 'bin/mvmt.ts', '--no-update-check'];
const serverOutput = new WeakMap<ChildProcessWithoutNullStreams, string>();

// Every test here spawns the real CLI binary as a child process — often
// several spawns per test — so the 5s default is too tight under CI
// load. Scope the higher timeout to this exec-heavy suite rather than
// loosening it globally for the fast unit-test files.
describe('CLI usability', { timeout: 20_000 }, () => {
  it('shows examples in top-level help', async () => {
    const { stdout } = await runCli(['--help']);

    expect(stdout).toContain('Share local files and folders through dashboard links, leases, and MCP');
    expect(stdout).toContain('Examples:');
    expect(stdout).toContain('mvmt serve -i');
    expect(stdout).toContain('start the local dashboard and control prompt');
    expect(stdout).toContain('mvmt serve --path ~/Documents');
    expect(stdout).toContain('temporarily expose one read-only source for this run');
    expect(stdout).toContain('mvmt lease create ~/Taxes ~/Receipts --label "Sarah - tax docs"');
    expect(stdout).toContain('create one 24h shared link for multiple paths');
    expect(stdout).toContain('mvmt users add owner --admin');
    expect(stdout).toContain('create a local dashboard login that can manage sources');
  });

  it('suggests close command names for typos', async () => {
    const result = await runCliAllowFailure(['mountz']);

    expect(result.code).toBe(1);
    expect(result.output).toContain("unknown command 'mountz'");
    expect(result.output).toContain('Did you mean mounts?');
  });

  it('prints mounts as JSON for scripts', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    try {
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: '/tmp/notes' }],
      }));

      const { stdout } = await runCli(['mounts', '--config', configPath, '--json']);
      expect(JSON.parse(stdout)).toMatchObject({
        mounts: [
          {
            name: 'notes',
            path: '/notes',
            root: '/tmp/notes',
            enabled: true,
            writeAccess: false,
          },
        ],
      });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('creates privileged dashboard users non-interactively', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    try {
      const env = { ...process.env, HOME: tmp };
      const created = await execFileAsync(
        process.execPath,
        [...cliArgs, 'users', 'add', 'sarah', '--password', 'correct horse battery staple', '--json'],
        { cwd: root, env },
      );
      expect(JSON.parse(created.stdout)).toMatchObject({
        user: { username: 'sarah' },
      });

      const listed = await execFileAsync(
        process.execPath,
        [...cliArgs, 'users', '--json'],
        { cwd: root, env },
      );
      expect(JSON.parse(listed.stdout)).toMatchObject({
        users: [expect.objectContaining({ username: 'sarah', disabled: false })],
      });

      const rotated = await execFileAsync(
        process.execPath,
        [...cliArgs, 'users', 'password', 'sarah', '--password', 'new dashboard password', '--json'],
        { cwd: root, env },
      );
      expect(JSON.parse(rotated.stdout)).toMatchObject({
        user: { username: 'sarah' },
      });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('removes privileged dashboard users non-interactively', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    try {
      const env = { ...process.env, HOME: tmp };
      await execFileAsync(
        process.execPath,
        [...cliArgs, 'users', 'add', 'sarah', '--password', 'correct horse battery staple', '--json'],
        { cwd: root, env },
      );

      const removed = await execFileAsync(
        process.execPath,
        [...cliArgs, 'users', 'remove', 'sarah', '--yes', '--json'],
        { cwd: root, env },
      );
      expect(JSON.parse(removed.stdout)).toMatchObject({
        removed: { username: 'sarah' },
      });

      const empty = await execFileAsync(
        process.execPath,
        [...cliArgs, 'users', '--json'],
        { cwd: root, env },
      );
      expect(JSON.parse(empty.stdout)).toEqual({ users: [] });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('removes a mount non-interactively with --yes', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    try {
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: '/tmp/notes' }],
      }));

      const { stdout } = await runCli(['mounts', 'remove', 'notes', '--config', configPath, '--yes']);
      expect(stdout).toContain('Mount notes removed');
      expect(stdout).toContain('Running mvmt unloads mount changes on the next request.');

      const { stdout: json } = await runCli(['mounts', '--config', configPath, '--json']);
      expect(JSON.parse(json)).toEqual({ mounts: [] });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('gives a next step when reindex has no mounts', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    try {
      await saveConfig(configPath, parseConfig({ version: 1 }));

      const result = await runCliAllowFailure(['reindex', '--config', configPath]);
      expect(result.code).toBe(1);
      expect(result.output).toContain('No mounts configured.');
      expect(result.output).toContain('mvmt mounts add <name> <folder>');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('starts non-interactive HTTP serve with no mounts configured', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const port = String(45_000 + Math.floor(Math.random() * 1_000));
    let child: ChildProcessWithoutNullStreams | undefined;
    try {
      await saveConfig(configPath, parseConfig({ version: 1 }));

      child = await startCliServer(['serve', '--config', configPath, '--port', port]);
      const response = await fetch(`http://127.0.0.1:${port}/.well-known/oauth-authorization-server`);
      expect(response.status).toBe(200);
    } finally {
      await stopCliServer(child);
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('creates an empty config instead of opening guided setup for non-interactive HTTP serve', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'missing-config.yaml');
    const port = String(46_000 + Math.floor(Math.random() * 1_000));
    let child: ChildProcessWithoutNullStreams | undefined;
    try {
      child = await startCliServer(['serve', '--config', configPath, '--port', port]);

      expect(readConfig(configPath)).toMatchObject({ version: 1, mounts: [] });
      expect(readServerOutput(child)).not.toContain('ExitPromptError');
    } finally {
      await stopCliServer(child);
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('rejects non-interactive mount roots that do not exist', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const missingRoot = path.join(tmp, 'missing');
    try {
      await saveConfig(configPath, parseConfig({ version: 1 }));

      const result = await runCliAllowFailure([
        'mounts',
        'add',
        'missing',
        missingRoot,
        '--config',
        configPath,
      ]);
      expect(result.code).toBe(1);
      expect(result.output).toContain(`File or folder not found: ${missingRoot}`);
      expect(result.output).not.toContain('Error:');
      expect(result.output).not.toContain('at async');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('explains mount base permission after add and edit', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'docs');
    try {
      await fs.mkdir(mountRoot);

      const add = await runCli([
        'mounts',
        'add',
        'docs',
        mountRoot,
        '--config',
        configPath,
        '--mount-path',
        '/docs',
        '--read-only',
      ]);
      expect(add.stdout).toContain('Base permission: read-only. Tokens cannot exceed it.');
      expect(add.stdout).toContain('Running mvmt loads mount changes on the next request.');

      const edit = await runCli([
        'mounts',
        'edit',
        'docs',
        '--config',
        configPath,
        '--write',
      ]);
      expect(edit.stdout).toContain('Base permission: read/write. Tokens cannot exceed it.');

      const updated = readConfig(configPath);
      expect(updated.mounts[0].writeAccess).toBe(true);
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('adds a single-file mount non-interactively', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const filePath = path.join(tmp, 'report.txt');
    try {
      await fs.writeFile(filePath, 'report', 'utf-8');

      const add = await runCli([
        'mounts',
        'add',
        'report',
        filePath,
        '--config',
        configPath,
        '--mount-path',
        '/report.txt',
        '--read-only',
      ]);
      expect(add.stdout).toContain('Mount report saved');

      const updated = readConfig(configPath);
      expect(updated.mounts[0]).toMatchObject({
        name: 'report',
        root: filePath,
        path: '/report.txt',
      });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('can disable tunnel access without removing saved tunnel details', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    try {
      await saveConfig(configPath, parseConfig({
        version: 1,
        server: {
          access: 'tunnel',
          tunnel: {
            provider: 'custom',
            command: 'cloudflared tunnel --config ~/.cloudflared/mvmt.yml run',
            url: 'https://mvmt.example.com',
          },
        },
      }));

      const { stdout } = await runCli(['tunnel', 'disable', '--config', configPath]);
      expect(stdout).toContain('Tunnel access disabled');

      const updated = readConfig(configPath);
      expect(updated.server.access).toBe('local');
      expect(updated.server.tunnel?.url).toBe('https://mvmt.example.com');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('configures relay tunnel access non-interactively', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    try {
      await saveConfig(configPath, parseConfig({ version: 1 }));

      const { stdout } = await runCli([
        'tunnel',
        'config',
        '--config',
        configPath,
        '--relay-url',
        'ws://127.0.0.1:8080/connect',
        '--relay-workspace',
        'demo',
        '--relay-token',
        'agent-secret',
        '--public-url',
        'https://demo.example.com',
      ]);
      expect(stdout).toContain('Tunnel config saved');

      const updated = readConfig(configPath);
      expect(updated.server.access).toBe('tunnel');
      expect(updated.server.tunnel).toMatchObject({
        provider: 'relay',
        relayUrl: 'ws://127.0.0.1:8080/connect',
        workspaceSlug: 'demo',
        agentToken: 'agent-secret',
        url: 'https://demo.example.com',
      });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('creates and lists scoped API tokens non-interactively', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'notes');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: mountRoot }],
      }));

      const { stdout } = await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--scope',
        'notes:read',
        '--name',
        'Codex CLI',
        '--description',
        'Local Codex token',
        '--expires',
        '7d',
      ]);
      expect(stdout).toContain('Token created.');
      expect(stdout).toContain('Name:    codex');
      expect(stdout).toContain('Scope:   notes:read');
      expect(stdout).toContain('Token:   mvmt_t_');
      expect(stdout).toContain('Expires:');
      expect(stdout).toContain('HTTP MCP endpoint');
      expect(stdout).toContain('URL:    http://127.0.0.1:4141/mcp');
      expect(stdout).toContain('Header: Authorization: Bearer mvmt_t_');
      expect(stdout).toContain('Use these values in any HTTP MCP client');
      expect(stdout).toContain('For OAuth clients, paste this token into the mvmt approval page');
      expect(stdout).not.toContain('claude mcp add');

      const { stdout: list } = await runCli(['token', '--config', configPath]);
      expect(list).toContain('NAME');
      expect(list).toContain('SCOPE');
      expect(list).toContain('codex');
      expect(list).toContain('notes:read');
      expect(list).toContain('(any)');

      const { stdout: json } = await runCli(['token', '--config', configPath, '--json']);
      expect(JSON.parse(json)).toMatchObject({
        tokens: [
          {
            name: 'codex',
            scope: 'notes:read',
            client: null,
            createdAt: expect.any(String),
            lastUsedAt: null,
            expiresAt: expect.any(String),
          },
        ],
      });
      await fs.writeFile(path.join(tmp, 'audit.log'), `${JSON.stringify({
        ts: new Date(Date.now() - 12 * 60 * 1000).toISOString(),
        event: 'token.use',
        connectorId: 'mvmt',
        tool: 'search',
        clientId: 'codex',
        name: 'codex',
        argKeys: [],
        argPreview: '{}',
        isError: false,
        durationMs: 1,
      })}\n`, 'utf-8');
      const { stdout: listWithUse } = await runCli(['token', '--config', configPath]);
      expect(listWithUse).toContain('12 minutes ago');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('publishes and unpublishes a scoped API token', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'notes');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: mountRoot }],
      }));
      await runCli(['token', 'add', 'codex', '--config', configPath, '--scope', 'notes:read']);

      const { stdout: published } = await runCli(['token', 'publish', 'codex', '--config', configPath]);
      expect(published).toContain("Token 'codex' published");
      expect(readConfig(configPath).clients?.[0]).toMatchObject({ id: 'codex', published: true });

      const { stdout: unpublished } = await runCli(['tokens', 'unpublish', 'codex', '--config', configPath]);
      expect(unpublished).toContain("Token 'codex' unpublished");
      expect(readConfig(configPath).clients?.[0]).toMatchObject({ id: 'codex', published: false });

      const { stdout: republished } = await runCli(['tokens', 'publish', 'codex', '--config', configPath]);
      expect(republished).toContain("Token 'codex' published");
      expect(readConfig(configPath).clients?.[0]).toMatchObject({ id: 'codex', published: true });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('labels client-bound API-token endpoints clearly', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'notes');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: mountRoot }],
      }));

      const { stdout } = await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--read',
        '/notes',
        '--client',
        'codex',
      ]);

      expect(stdout).toContain('HTTP MCP endpoint (client: codex):');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('shows a safe empty state for token listing', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    try {
      await saveConfig(configPath, parseConfig({ version: 1 }));

      const { stdout } = await runCli(['token', '--config', configPath]);
      expect(stdout).toContain('No tokens configured.');
      expect(stdout).toContain('Create one with: mvmt token add <name>');

      const { stdout: json } = await runCli(['token', '--config', configPath, '--json']);
      expect(JSON.parse(json)).toEqual({ tokens: [] });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('rejects API-token write access that exceeds the mount base permission', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'notes');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: mountRoot }],
      }));

      const result = await runCliAllowFailure([
        'tokens',
        'add',
        'codex',
        '--config',
        configPath,
        '--write',
        '/notes',
      ]);
      expect(result.code).toBe(1);
      expect(result.output).toContain('Mount /notes is read-only');
      expect(result.output).not.toContain('Error:');
      expect(result.output).not.toContain('at async');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('edits a scoped API token non-interactively', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'workspace');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: mountRoot, writeAccess: true }],
      }));

      await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--read',
        '/workspace',
      ]);
      const { stdout } = await runCli([
        'token',
        'edit',
        'codex',
        '--config',
        configPath,
        '--scope',
        'workspace:write',
        '--description',
        'Updated token',
        '--expires',
        'never',
      ]);
      expect(stdout).toContain('Token updated.');
      expect(stdout).toContain('Existing token value was not changed.');
      expect(stdout).toContain('Permission edit applies to existing API tokens and OAuth grants on the next request.');

      const { stdout: list } = await runCli(['token', '--config', configPath]);
      expect(list).toContain('workspace:write');
      expect(list).toContain('never');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('explains client-binding edits require OAuth reauthorization', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'workspace');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: mountRoot, writeAccess: true }],
      }));

      await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--read',
        '/workspace',
      ]);
      const { stdout } = await runCli([
        'token',
        'edit',
        'codex',
        '--config',
        configPath,
        '--client',
        'claude',
        '--scope',
        'workspace:write',
      ]);

      expect(stdout).toContain('Existing OAuth grants must reauthorize because this edit changed client binding.');
      expect(stdout).toContain('Permission edit applies immediately to API tokens, and to OAuth grants after reauthorization.');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('explains that narrowing a scoped API token does not require OAuth reauthorization', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'workspace');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: mountRoot, writeAccess: true }],
      }));

      await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--write',
        '/workspace',
      ]);
      const { stdout } = await runCli([
        'token',
        'edit',
        'codex',
        '--config',
        configPath,
        '--scope',
        'workspace:read',
      ]);

      expect(stdout).toContain('Token updated.');
      expect(stdout).toContain('Existing token value was not changed.');
      expect(stdout).toContain('Permission edit applies to existing API tokens and OAuth grants on the next request.');

      const { stdout: list } = await runCli(['token', '--config', configPath]);
      expect(list).toContain('workspace:read');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('can explicitly edit a scoped API token down to no permissions', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'notes');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: mountRoot }],
      }));

      await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--read',
        '/notes',
      ]);
      const { stdout } = await runCli([
        'token',
        'edit',
        'codex',
        '--config',
        configPath,
        '--no-permissions',
      ]);
      expect(stdout).toContain('Token updated.');
      expect(stdout).toContain('Scope:   (none)');

      const { stdout: list } = await runCli(['token', '--config', configPath]);
      expect(list).toContain('(none)');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('rejects --no-permissions when a replacement scope is also provided', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'notes');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: mountRoot }],
      }));

      await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--read',
        '/notes',
      ]);
      const result = await runCliAllowFailure([
        'token',
        'edit',
        'codex',
        '--config',
        configPath,
        '--no-permissions',
        '--scope',
        'notes:read',
      ]);

      expect(result.code).toBe(1);
      expect(result.output).toContain('Use either --no-permissions or --scope/--read/--write, not both.');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('rotates a scoped API token non-interactively and prints the replacement once', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-cli-usability-'));
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'notes');
    try {
      await fs.mkdir(mountRoot);
      await saveConfig(configPath, parseConfig({
        version: 1,
        mounts: [{ name: 'notes', type: 'local_folder', path: '/notes', root: mountRoot }],
      }));

      await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--read',
        '/notes',
      ]);
      const { stdout } = await runCli(['token', 'rotate', 'codex', '--config', configPath, '--yes']);

      expect(stdout).toContain('Token rotated.');
      expect(stdout).toContain('Token:   mvmt_t_');
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });
});

async function runCli(args: string[]): Promise<{ stdout: string; stderr: string }> {
  return execFileAsync(process.execPath, [...cliArgs, ...args], { cwd: root });
}

async function runCliAllowFailure(args: string[]): Promise<{ code: number; output: string }> {
  try {
    const { stdout, stderr } = await runCli(args);
    return { code: 0, output: `${stdout}${stderr}` };
  } catch (err) {
    const failure = err as { code?: number; stdout?: string; stderr?: string };
    return {
      code: typeof failure.code === 'number' ? failure.code : 1,
      output: `${failure.stdout ?? ''}${failure.stderr ?? ''}`,
    };
  }
}

async function startCliServer(args: string[]): Promise<ChildProcessWithoutNullStreams> {
  const child = spawn(process.execPath, [...cliArgs, ...args], {
    cwd: root,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  serverOutput.set(child, '');
  const appendOutput = (chunk: Buffer) => {
    serverOutput.set(child, `${serverOutput.get(child) ?? ''}${chunk.toString()}`);
  };
  child.stdout.on('data', appendOutput);
  child.stderr.on('data', appendOutput);

  const port = args[args.indexOf('--port') + 1];
  if (!port) throw new Error('startCliServer requires --port');

  for (let attempt = 0; attempt < 80; attempt += 1) {
    if (child.exitCode !== null) {
      throw new Error(readServerOutput(child) || `mvmt serve exited with ${child.exitCode}`);
    }
    try {
      const response = await fetch(`http://127.0.0.1:${port}/.well-known/oauth-authorization-server`);
      if (response.status === 200) return child;
    } catch {
      // Server is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }

  await stopCliServer(child);
  throw new Error(`mvmt serve did not become reachable:\n${readServerOutput(child)}`);
}

function readServerOutput(child: ChildProcessWithoutNullStreams): string {
  return serverOutput.get(child) ?? '';
}

async function stopCliServer(child: ChildProcessWithoutNullStreams | undefined): Promise<void> {
  if (!child || child.exitCode !== null) return;
  child.kill('SIGTERM');
  await Promise.race([
    new Promise<void>((resolve) => child.once('close', () => resolve())),
    new Promise((resolve) => setTimeout(resolve, 2_000)),
  ]);
  if (child.exitCode === null) child.kill('SIGKILL');
}
