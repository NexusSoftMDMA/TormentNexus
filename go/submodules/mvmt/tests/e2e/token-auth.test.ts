import { createHash } from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';
import { Container, dockerAvailable, parseMcpResponse, runMvmtContainer } from './harness.js';

const HAS_DOCKER = dockerAvailable();
const describeIfDocker = HAS_DOCKER ? describe : describe.skip;
const PROTOCOL_VERSION = '2025-03-26';

describeIfDocker('mvmt e2e: token + auth', () => {
  let container: Container;
  let workdir: string;
  let notesHost: string;
  let workspaceHost: string;
  let sharedFileHost: string;

  beforeAll(async () => {
    workdir = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-e2e-'));
    notesHost = path.join(workdir, 'notes');
    workspaceHost = path.join(workdir, 'workspace');
    sharedFileHost = path.join(workdir, 'shared.bin');
    fs.mkdirSync(notesHost, { recursive: true });
    fs.mkdirSync(workspaceHost, { recursive: true });
    fs.writeFileSync(path.join(notesHost, 'launch.md'), '# Launch\nNotes about the launch plan.\n');
    fs.writeFileSync(path.join(workspaceHost, 'plan.md'), '# Plan\nProject planning workspace.\n');
    fs.writeFileSync(sharedFileHost, Buffer.from('shared file from docker e2e\n'));

    container = await runMvmtContainer({
      binds: [
        { host: notesHost, container: '/data/notes' },
        { host: workspaceHost, container: '/data/workspace' },
        { host: sharedFileHost, container: '/data/shared.bin', readonly: true },
      ],
    });

    const addNotes = await container.exec([
      'mounts', 'add', 'notes', '/data/notes', '--mount-path', '/notes', '--read-only',
    ]);
    expect(addNotes.exitCode, addNotes.stderr).toBe(0);
    const addWorkspace = await container.exec([
      'mounts', 'add', 'workspace', '/data/workspace', '--mount-path', '/workspace', '--write',
    ]);
    expect(addWorkspace.exitCode, addWorkspace.stderr).toBe(0);

    await container.startServer();
  }, 180_000);

  afterAll(async () => {
    if (container) await container.stop();
    if (workdir) fs.rmSync(workdir, { recursive: true, force: true });
  }, 30_000);

  async function createToken(id: string, args: string[]): Promise<string> {
    const result = await container.exec(['token', 'add', id, ...args]);
    expect(result.exitCode, result.stderr).toBe(0);
    const match = result.stdout.match(/Token:\s+(mvmt_t_[A-Za-z0-9_-]+)/);
    if (!match) throw new Error(`could not parse token from output:\n${result.stdout}`);
    return match[1];
  }

  async function openMcpSession(token: string): Promise<string> {
    const res = await container.http({
      method: 'POST',
      path: '/mcp',
      token,
      protocolVersion: PROTOCOL_VERSION,
      body: {
        jsonrpc: '2.0', id: 1, method: 'initialize',
        params: { protocolVersion: PROTOCOL_VERSION, capabilities: {}, clientInfo: { name: 'e2e', version: '0.0.0' } },
      },
    });
    expect(res.status).toBe(200);
    const sessionId = res.headers['mcp-session-id'];
    expect(sessionId).toBeTruthy();
    return sessionId;
  }

  async function callTool(token: string, sessionId: string, id: number, name: string, args: object): Promise<any> {
    const res = await container.http({
      method: 'POST',
      path: '/mcp',
      token,
      sessionId,
      protocolVersion: PROTOCOL_VERSION,
      body: { jsonrpc: '2.0', id, method: 'tools/call', params: { name, arguments: args } },
    });
    expect(res.status).toBe(200);
    return parseMcpResponse(res.body);
  }

  function wasRejected(response: any): boolean {
    if (response.error !== undefined) return true;
    if (response.result?.isError === true) return true;
    return false;
  }

  it('reads in-scope files with a read-scoped token', async () => {
    const token = await createToken('reader', ['--read', '/notes']);
    const sessionId = await openMcpSession(token);
    const read = await callTool(token, sessionId, 2, 'read', { path: '/notes/launch.md' });
    const payload = JSON.parse(read.result.content[0].text);
    expect(payload).toMatchObject({ path: '/notes/launch.md' });
    expect(payload.content).toContain('Launch');
  });

  it('rejects out-of-scope reads', async () => {
    const token = await createToken('scoped-notes', ['--read', '/notes']);
    const sessionId = await openMcpSession(token);
    const read = await callTool(token, sessionId, 2, 'read', { path: '/workspace/plan.md' });
    expect(wasRejected(read)).toBe(true);
  });

  it('rejects writes from a read-only token', async () => {
    const token = await createToken('reader-no-write', ['--read', '/workspace']);
    const sessionId = await openMcpSession(token);
    const write = await callTool(token, sessionId, 2, 'write', {
      path: '/workspace/new.md', content: 'should be rejected',
    });
    expect(wasRejected(write)).toBe(true);
    expect(fs.existsSync(path.join(workspaceHost, 'new.md'))).toBe(false);
  });

  it('allows writes within a write-scoped path and persists them on the host', async () => {
    const token = await createToken('writer', ['--write', '/workspace']);
    const sessionId = await openMcpSession(token);
    const write = await callTool(token, sessionId, 2, 'write', {
      path: '/workspace/from-e2e.md', content: 'hello from e2e',
    });
    expect(write.result.isError).not.toBe(true);
    expect(fs.readFileSync(path.join(workspaceHost, 'from-e2e.md'), 'utf-8')).toBe('hello from e2e');
  });

  it('rejects path-traversal attempts that escape the mount', async () => {
    const token = await createToken('traverser', ['--read', '/notes']);
    const sessionId = await openMcpSession(token);
    const read = await callTool(token, sessionId, 2, 'read', { path: '/notes/../../etc/passwd' });
    expect(wasRejected(read)).toBe(true);
  });

  it('returns 401 with WWW-Authenticate when no bearer header is sent', async () => {
    const res = await container.http({
      method: 'POST',
      path: '/mcp',
      protocolVersion: PROTOCOL_VERSION,
      body: { jsonrpc: '2.0', id: 1, method: 'initialize', params: {} },
    });
    expect(res.status).toBe(401);
    expect(res.headers['www-authenticate']).toMatch(/resource_metadata=.*oauth-protected-resource/);
  });

  it('returns 401 for a bearer token that does not match any configured client', async () => {
    const res = await container.http({
      method: 'POST',
      path: '/mcp',
      token: 'mvmt_t_definitely_not_real',
      protocolVersion: PROTOCOL_VERSION,
      body: { jsonrpc: '2.0', id: 1, method: 'initialize', params: {} },
    });
    expect(res.status).toBe(401);
  });

  it('creates a file lease and downloads it by checksum', async () => {
    const lease = await container.exec(['lease', 'create', '/data/shared.bin', '--label', 'Shared file']);
    expect(lease.exitCode, lease.stderr).toBe(0);
    expect(lease.stdout).toContain('(24h default)');
    const match = lease.stdout.match(/URL:\s+(http:\/\/127\.0\.0\.1:4141\/lease\/\S+)/);
    if (!match) throw new Error(`could not parse lease URL from output:\n${lease.stdout}`);
    const url = new URL(match[1]);

    const downloaded = await container.run([
      'curl', '-fsSL', `${url.origin}${url.pathname}/files/shared.bin?token=${url.searchParams.get('token')}`,
    ]);
    expect(downloaded.exitCode, downloaded.stderr).toBe(0);
    const expectedHash = createHash('sha256').update(fs.readFileSync(sharedFileHost)).digest('hex');
    const actualHash = createHash('sha256').update(downloaded.stdout).digest('hex');
    expect(actualHash).toBe(expectedHash);
  });

  it('creates a folder lease and downloads a file by checksum', async () => {
    const lease = await container.exec(['lease', 'create', '/data/notes', '--label', 'E2E notes']);
    expect(lease.exitCode, lease.stderr).toBe(0);
    expect(lease.stdout).toContain('(24h default)');
    const match = lease.stdout.match(/URL:\s+(http:\/\/127\.0\.0\.1:4141\/lease\/\S+)/);
    if (!match) throw new Error(`could not parse lease URL from output:\n${lease.stdout}`);
    const url = new URL(match[1]);

    const listing = await container.run(['curl', '-fsSL', match[1]]);
    expect(listing.exitCode, listing.stderr).toBe(0);
    expect(listing.stdout).toContain('E2E notes');
    expect(listing.stdout).toContain('launch.md');

    const downloaded = await container.run([
      'curl', '-fsSL', `${url.origin}${url.pathname}/files/launch.md?token=${url.searchParams.get('token')}`,
    ]);
    expect(downloaded.exitCode, downloaded.stderr).toBe(0);
    const expectedHash = createHash('sha256').update(fs.readFileSync(path.join(notesHost, 'launch.md'))).digest('hex');
    const actualHash = createHash('sha256').update(downloaded.stdout).digest('hex');
    expect(actualHash).toBe(expectedHash);
  });

  it('creates an upload-only folder lease and receives a file', async () => {
    const lease = await container.exec(['lease', 'create', '/data/workspace', '--label', 'Phone uploads', '--mode', 'upload']);
    expect(lease.exitCode, lease.stderr).toBe(0);
    expect(lease.stdout).toContain('Mode: upload only');
    const match = lease.stdout.match(/URL:\s+(http:\/\/127\.0\.0\.1:4141\/lease\/\S+)/);
    if (!match) throw new Error(`could not parse lease URL from output:\n${lease.stdout}`);
    const url = new URL(match[1]);
    const uploadUrl = `${url.origin}${url.pathname}/files/e2e-upload.txt?token=${url.searchParams.get('token')}`;

    const uploaded = await container.run([
      'curl', '-fsSL', '-X', 'PUT', '-H', 'Content-Type: application/octet-stream',
      '--data-binary', 'uploaded from docker e2e', uploadUrl,
    ]);
    expect(uploaded.exitCode, uploaded.stderr).toBe(0);

    const expectedHash = createHash('sha256').update('uploaded from docker e2e').digest('hex');
    const actualHash = createHash('sha256').update(fs.readFileSync(path.join(workspaceHost, 'e2e-upload.txt'))).digest('hex');
    expect(actualHash).toBe(expectedHash);
  });
});
