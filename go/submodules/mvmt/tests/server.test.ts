import { createHash } from 'node:crypto';
import fs from 'node:fs';
import { request as httpRequest } from 'node:http';
import { createServer } from 'node:net';
import os from 'node:os';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import type { Request } from 'express';
import type { ClientConfig } from '../src/config/schema.js';
import {
  buildOriginCheck,
  isBenignDuplicateSseConflict,
  isStandaloneSseRequest,
  MVMT_SERVER_INSTRUCTIONS,
  startHttpServer,
} from '../src/server/index.js';
import { parseConfig, readConfig, saveConfig } from '../src/config/loader.js';
import { TextContextIndex } from '../src/context/text-index.js';
import { OAuthStore } from '../src/server/oauth.js';
import { ToolRouter } from '../src/apps/mcp/router.js';
import { addLeaseResources, createLease, listLeases, revokeLease, setLeasePublished } from '../src/core/leases/store.js';
import { findLeaseSecret, leaseSecretsPathForLeaseStore } from '../src/core/leases/secrets.js';
import { createPrivilegedUser, removePrivilegedUser } from '../src/apps/dashboard/users.js';
import { createAuditLogger } from '../src/utils/audit.js';
import { hashApiToken } from '../src/utils/api-token-hash.js';
import { ensureSigningKey, generateSessionToken, rotateSigningKey } from '../src/utils/token.js';

function req(origin?: string): Request {
  return { headers: origin === undefined ? {} : { origin } } as unknown as Request;
}

function requestWithHost(
  port: number,
  requestPath: string,
  host: string,
  options: { method?: string; headers?: Record<string, string>; body?: string } = {},
): Promise<{ status: number; body: string; json: () => unknown }> {
  const body = options.body ?? '';
  return new Promise((resolve, reject) => {
    const req = httpRequest({
      hostname: '127.0.0.1',
      port,
      path: requestPath,
      method: options.method ?? 'GET',
      headers: {
        ...(options.headers ?? {}),
        Host: host,
        ...(body ? { 'Content-Length': Buffer.byteLength(body).toString() } : {}),
      },
    }, (res) => {
      const chunks: Buffer[] = [];
      res.on('data', (chunk: Buffer | string) => {
        chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
      });
      res.on('end', () => {
        const text = Buffer.concat(chunks).toString('utf-8');
        resolve({
          status: res.statusCode ?? 0,
          body: text,
          json: () => JSON.parse(text),
        });
      });
    });
    req.on('error', reject);
    if (body) req.write(body);
    req.end();
  });
}

describe('buildOriginCheck', () => {
  const check = buildOriginCheck([]);

  it('allows requests with no Origin (non-browser MCP clients)', () => {
    expect(check(req())).toBe(true);
  });

  it('allows localhost and 127.0.0.1 origins on any port', () => {
    expect(check(req('http://localhost'))).toBe(true);
    expect(check(req('http://localhost:4141'))).toBe(true);
    expect(check(req('http://127.0.0.1:4141'))).toBe(true);
    expect(check(req('http://[::1]:4141'))).toBe(true);
  });

  it('rejects arbitrary remote origins (DNS rebinding)', () => {
    expect(check(req('http://evil.example.com'))).toBe(false);
    expect(check(req('https://claude.ai'))).toBe(false);
  });

  it('allows explicitly configured extra origins', () => {
    const withClaude = buildOriginCheck(['https://claude.ai']);
    expect(withClaude(req('https://claude.ai'))).toBe(true);
    expect(withClaude(req('https://Claude.AI'))).toBe(true);
    expect(withClaude(req('https://evil.example.com'))).toBe(false);
  });

  it('rejects malformed Origin headers', () => {
    expect(check(req('not a url'))).toBe(false);
  });
});

describe('SSE request helpers', () => {
  it('identifies standalone SSE GET requests', () => {
    expect(
      isStandaloneSseRequest({
        method: 'GET',
        headers: { accept: 'application/json, text/event-stream' },
      } as unknown as Request),
    ).toBe(true);
  });

  it('does not classify non-SSE GET requests as standalone SSE requests', () => {
    expect(
      isStandaloneSseRequest({
        method: 'GET',
        headers: { accept: 'application/json' },
      } as unknown as Request),
    ).toBe(false);
  });

  it('treats duplicate standalone SSE conflicts as benign reconnect noise', () => {
    expect(isBenignDuplicateSseConflict(new Error('Conflict: Only one SSE stream is allowed per session'))).toBe(true);
    expect(isBenignDuplicateSseConflict(new Error('Conflict: Stream already has an active connection'))).toBe(false);
  });
});

describe('file lease downloads', () => {
  it('streams a single-file lease through a query-token URL', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const filePath = path.join(tmp, 'payload.bin');
    const payload = Buffer.from([0, 1, 2, 3, 4, 255]);
    fs.writeFileSync(filePath, payload);
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'payload', type: 'local_folder', path: '/payload.bin', root: filePath }],
    });
    const lease = createLease(leaseStorePath, {
      label: 'Payload',
      path: '/payload.bin',
      resources: [{ path: '/payload.bin', sourcePath: '/payload.bin', type: 'file' }],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}/files/payload.bin?token=${lease.token}`);
      expect(response.status).toBe(200);
      expect(response.headers.get('content-disposition')).toContain('payload.bin');
      expect(response.headers.get('accept-ranges')).toBe('bytes');
      const downloaded = Buffer.from(await response.arrayBuffer());
      expect(createHash('sha256').update(downloaded).digest('hex')).toBe(createHash('sha256').update(payload).digest('hex'));
      expect(listLeases(leaseStorePath)[0].downloadCount).toBe(1);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('supports byte ranges for file lease downloads', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const filePath = path.join(tmp, 'payload.bin');
    fs.writeFileSync(filePath, Buffer.from([10, 11, 12, 13, 14]));
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'payload', type: 'local_folder', path: '/payload.bin', root: filePath }],
    });
    const lease = createLease(leaseStorePath, {
      label: 'Payload',
      path: '/payload.bin',
      resources: [{ path: '/payload.bin', sourcePath: '/payload.bin', type: 'file' }],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}/files/payload.bin?token=${lease.token}`, {
        headers: { Range: 'bytes=1-3' },
      });
      expect(response.status).toBe(206);
      expect(response.headers.get('content-range')).toBe('bytes 1-3/5');
      expect([...Buffer.from(await response.arrayBuffer())]).toEqual([11, 12, 13]);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('enforces one-time download leases without consuming the limit on HEAD', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const filePath = path.join(tmp, 'invite.txt');
    fs.writeFileSync(filePath, 'relay secret');
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'invite', type: 'local_folder', path: '/invite.txt', root: filePath }],
    });
    const lease = createLease(leaseStorePath, {
      label: 'Invite',
      path: '/invite.txt',
      resources: [{ path: '/invite.txt', sourcePath: '/invite.txt', type: 'file' }],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
      maxDownloads: 1,
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    const url = `http://127.0.0.1:${server.port}/lease/${lease.record.id}/files/invite.txt?token=${lease.token}`;
    try {
      const head = await fetch(url, { method: 'HEAD' });
      expect(head.status).toBe(200);
      expect(listLeases(leaseStorePath)[0].downloadCount).toBe(0);

      const first = await fetch(url);
      expect(first.status).toBe(200);
      expect(await first.text()).toBe('relay secret');
      expect(listLeases(leaseStorePath)[0].downloadCount).toBe(1);

      const second = await fetch(url);
      expect(second.status).toBe(410);
      expect((await second.json() as { error: string }).error).toBe('lease_download_limit_reached');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects invalid and expired lease tokens', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const filePath = path.join(tmp, 'payload.bin');
    fs.writeFileSync(filePath, 'payload', 'utf-8');
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'payload', type: 'local_folder', path: '/payload.bin', root: filePath }],
    });
    const active = createLease(leaseStorePath, {
      label: 'Active',
      path: '/payload.bin',
      resources: [{ path: '/payload.bin', sourcePath: '/payload.bin', type: 'file' }],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const expired = createLease(leaseStorePath, {
      label: 'Expired',
      path: '/payload.bin',
      resources: [{ path: '/payload.bin', sourcePath: '/payload.bin', type: 'file' }],
      expiresAt: new Date(Date.now() - 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const invalid = await fetch(`http://127.0.0.1:${server.port}/lease/${active.record.id}/files/payload.bin?token=wrong`);
      expect(invalid.status).toBe(401);
      const gone = await fetch(`http://127.0.0.1:${server.port}/lease/${expired.record.id}/files/payload.bin?token=${expired.token}`);
      expect(gone.status).toBe(410);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });
});

// The exposure boundary: a capability-only grant (explicitly
// published:false) is local-only. It still works for apps reaching
// mvmt over 127.0.0.1; a public-tunnel request for it is rejected.
// Grants without an explicit published value are grandfathered as
// published so existing tokens and leases keep working.
describe('grant exposure boundary', () => {
  async function withLeaseServer(
    run: (ctx: { port: number; leaseStorePath: string; leaseId: string; token: string }) => Promise<void>,
  ): Promise<void> {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-exposure-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const filePath = path.join(tmp, 'payload.bin');
    fs.writeFileSync(filePath, Buffer.from([1, 2, 3]));
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'payload', type: 'local_folder', path: '/payload.bin', root: filePath }],
    });
    const lease = createLease(leaseStorePath, {
      label: 'Payload',
      path: '/payload.bin',
      resources: [{ path: '/payload.bin', sourcePath: '/payload.bin', type: 'file' }],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      await run({ port: server.port, leaseStorePath, leaseId: lease.record.id, token: lease.token });
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  }

  const leaseUrl = (port: number, id: string, token: string) =>
    `http://127.0.0.1:${port}/lease/${id}/files/payload.bin?token=${token}`;

  const publicHostHeaders = { Host: 'public.example.test' };

  it('rejects a relay-forwarded request for a capability-only lease', async () => {
    await withLeaseServer(async ({ port, leaseStorePath, leaseId, token }) => {
      setLeasePublished(leaseStorePath, leaseId, false);
      const viaRelay = await fetch(leaseUrl(port, leaseId, token), {
        headers: { 'X-MVMT-Transport': 'relay' },
      });
      expect(viaRelay.status).toBe(403);
      expect((await viaRelay.json() as { error: string }).error).toBe('lease_not_published');
    });
  });

  it('rejects a public-host request for a capability-only lease', async () => {
    await withLeaseServer(async ({ port, leaseStorePath, leaseId, token }) => {
      setLeasePublished(leaseStorePath, leaseId, false);
      const url = new URL(leaseUrl(port, leaseId, token));
      const viaPublicTunnel = await requestWithHost(port, `${url.pathname}${url.search}`, publicHostHeaders.Host);
      expect(viaPublicTunnel.status).toBe(403);
      expect((viaPublicTunnel.json() as { error: string }).error).toBe('lease_not_published');
    });
  });

  it('allows a localhost request for a capability-only lease', async () => {
    await withLeaseServer(async ({ port, leaseStorePath, leaseId, token }) => {
      setLeasePublished(leaseStorePath, leaseId, false);
      const local = await fetch(leaseUrl(port, leaseId, token));
      expect(local.status).toBe(200);
    });
  });

  it('allows a relay-forwarded request for a published lease', async () => {
    await withLeaseServer(async ({ port, leaseStorePath, leaseId, token }) => {
      setLeasePublished(leaseStorePath, leaseId, true);
      const viaRelay = await fetch(leaseUrl(port, leaseId, token), {
        headers: { 'X-MVMT-Transport': 'relay' },
      });
      expect(viaRelay.status).toBe(200);
    });
  });

  it('grandfathers a lease with no explicit published value as published', async () => {
    await withLeaseServer(async ({ port, leaseId, token }) => {
      const viaRelay = await fetch(leaseUrl(port, leaseId, token), {
        headers: { 'X-MVMT-Transport': 'relay' },
      });
      expect(viaRelay.status).toBe(200);
    });
  });

  async function withMcpServer(
    published: boolean | undefined,
    run: (ctx: { port: number; token: string }) => Promise<void>,
  ): Promise<void> {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-exposure-mcp-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const root = path.join(tmp, 'workspace');
    fs.mkdirSync(root, { recursive: true });
    fs.writeFileSync(path.join(root, 'note.md'), 'hello');
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root }],
      clients: [
        {
          id: 'codex',
          name: 'Codex CLI',
          auth: { type: 'token', tokenHash: hashApiToken('codex-token') },
          permissions: [{ path: '/workspace/**', actions: ['read'] }],
          ...(published === undefined ? {} : { published }),
        },
      ],
    });
    const router = new ToolRouter();
    await router.initialize();
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      clients: config.clients,
    });
    try {
      await run({ port: server.port, token: 'codex-token' });
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  }

  const mcpInitialize = async (
    port: number,
    token: string,
    options: { viaRelay?: boolean; host?: string } = {},
  ): Promise<{ status: number; json: () => Promise<unknown> }> => {
    const body = JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: { protocolVersion: '2025-03-26', capabilities: {}, clientInfo: { name: 't', version: '0' } },
    });
    const headers = {
      Authorization: `Bearer ${token}`,
      Accept: 'application/json, text/event-stream',
      'Content-Type': 'application/json',
      ...(options.viaRelay ? { 'X-MVMT-Transport': 'relay' } : {}),
    };
    if (options.host) {
      const response = await requestWithHost(port, '/mcp', options.host, { method: 'POST', headers, body });
      return { status: response.status, json: async () => response.json() };
    }
    return fetch(`http://127.0.0.1:${port}/mcp`, {
      method: 'POST',
      headers,
      body,
    });
  };

  it('rejects a relay-forwarded /mcp request for a capability-only token', async () => {
    await withMcpServer(false, async ({ port, token }) => {
      const viaRelay = await mcpInitialize(port, token, { viaRelay: true });
      expect(viaRelay.status).toBe(403);
      expect((await viaRelay.json() as { error: string }).error).toBe('grant_not_published');
    });
  });

  it('rejects a public-host /mcp request for a capability-only token', async () => {
    await withMcpServer(false, async ({ port, token }) => {
      const viaPublicTunnel = await mcpInitialize(port, token, { host: 'public.example.test' });
      expect(viaPublicTunnel.status).toBe(403);
      expect((await viaPublicTunnel.json() as { error: string }).error).toBe('grant_not_published');
    });
  });

  it('allows a localhost /mcp request for a capability-only token', async () => {
    await withMcpServer(false, async ({ port, token }) => {
      const local = await mcpInitialize(port, token);
      expect(local.status).toBe(200);
    });
  });

  it('allows a relay-forwarded /mcp request for a published token', async () => {
    await withMcpServer(true, async ({ port, token }) => {
      const viaRelay = await mcpInitialize(port, token, { viaRelay: true });
      expect(viaRelay.status).toBe(200);
    });
  });

  it('grandfathers a token with no explicit published value as published', async () => {
    await withMcpServer(undefined, async ({ port, token }) => {
      const viaRelay = await mcpInitialize(port, token, { viaRelay: true });
      expect(viaRelay.status).toBe(200);
    });
  });
});

describe('folder lease access', () => {
  it('lists a leased folder and streams arbitrary files by checksum', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const root = path.join(tmp, 'taxes');
    fs.mkdirSync(path.join(root, 'nested'), { recursive: true });
    fs.mkdirSync(path.join(root, '.ssh'), { recursive: true });
    const payload = Buffer.from([0, 1, 2, 3, 4, 255]);
    fs.writeFileSync(path.join(root, 'w2.pdf'), payload);
    fs.writeFileSync(path.join(root, 'nested', 'note.txt'), 'nested note', 'utf-8');
    fs.writeFileSync(path.join(root, '.ssh', 'id_ed25519'), 'secret', 'utf-8');
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'taxes', type: 'local_folder', path: '/taxes', root }],
    });
    const lease = createLease(leaseStorePath, {
      label: 'Sarah - tax docs',
      path: '/taxes',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const listing = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}/files?token=${lease.token}`);
      expect(listing.status).toBe(200);
      const body = await listing.json() as { entries: { name: string; type: string }[] };
      expect(body.entries).toEqual([
        expect.objectContaining({ name: 'nested', type: 'directory' }),
        expect.objectContaining({ name: 'w2.pdf', type: 'file' }),
      ]);
      expect(body.entries.some((entry) => entry.name === '.ssh')).toBe(false);

      const html = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}?token=${lease.token}`);
      expect(html.status).toBe(200);
      expect(await html.text()).toContain('Folder lease');

      const download = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}/files/w2.pdf?token=${lease.token}`);
      expect(download.status).toBe(200);
      expect(download.headers.get('content-disposition')).toContain('w2.pdf');
      const downloaded = Buffer.from(await download.arrayBuffer());
      expect(createHash('sha256').update(downloaded).digest('hex')).toBe(createHash('sha256').update(payload).digest('hex'));
      expect(listLeases(leaseStorePath)[0].downloadCount).toBe(1);
      expect(listLeases(leaseStorePath)[0].lastUsedAt).toBeTruthy();
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('escapes browser lease pages and does not reflect unsafe query tokens', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const readRoot = path.join(tmp, 'read');
    const uploadRoot = path.join(tmp, 'upload');
    fs.mkdirSync(readRoot, { recursive: true });
    fs.mkdirSync(uploadRoot, { recursive: true });
    fs.writeFileSync(path.join(readRoot, 'report & <draft>.txt'), 'draft', 'utf-8');
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'read', type: 'local_folder', path: '/read', root: readRoot },
        { name: 'upload', type: 'local_folder', path: '/upload', root: uploadRoot, writeAccess: true },
      ],
    });
    const label = '<img src=x onerror=alert(1)> Sarah';
    const readLease = createLease(leaseStorePath, {
      label,
      path: '/read',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const uploadLease = createLease(leaseStorePath, {
      label,
      path: '/upload',
      permissions: ['upload'],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const unsafeQuery = encodeURIComponent('</script><img src=x onerror=alert(1)>');
      const readPage = await fetch(`http://127.0.0.1:${server.port}/lease/${readLease.record.id}?token=${unsafeQuery}`, {
        headers: { Authorization: `Bearer     ${readLease.token}` },
      });
      expect(readPage.status).toBe(200);
      const readHtml = await readPage.text();
      expect(readHtml).toContain('Folder lease');
      expect(readHtml).toContain("sessionStorage.setItem(tokenKey, token)");
      expect(readHtml).toContain("history.replaceState(null, '', clean.pathname + clean.search)");
      expect(readHtml).not.toContain('&lt;img src=x onerror=alert(1)&gt; Sarah');
      expect(readHtml).not.toContain('report &amp; &lt;draft&gt;.txt');
      expect(readHtml).not.toContain('<img src=x onerror=alert(1)>');
      expect(readHtml).not.toContain('</script><img');
      for (const body of inlineScriptBodies(readHtml)) {
        expect(() => new Function(body)).not.toThrow();
      }

      const uploadPage = await fetch(`http://127.0.0.1:${server.port}/lease/${uploadLease.record.id}?token=${unsafeQuery}`, {
        headers: { Authorization: `Bearer     ${uploadLease.token}` },
      });
      expect(uploadPage.status).toBe(200);
      const uploadHtml = await uploadPage.text();
      expect(uploadHtml).toContain('Upload-only lease');
      expect(uploadHtml).toContain("sessionStorage.setItem(tokenKey, token)");
      expect(uploadHtml).toContain('is too large for this public relay');
      expect(uploadHtml).toContain('Upload failed before reaching mvmt');
      expect(uploadHtml).toContain('id="upload-progress"');
      expect(uploadHtml).toContain('id="upload-progress-bar"');
      expect(uploadHtml).toContain('function showUploadProgress(message, percent)');
      expect(uploadHtml).toContain('new XMLHttpRequest()');
      expect(uploadHtml).toContain("showUploadProgress('Uploading ' + file.name + ' (' + percent + '%)', percent)");
      expect(uploadHtml).not.toContain('&lt;img src=x onerror=alert(1)&gt; Sarah');
      expect(uploadHtml).not.toContain('</script><img');
      expect(uploadHtml).not.toContain('token = "</script>');
      for (const body of inlineScriptBodies(uploadHtml)) {
        expect(() => new Function(body)).not.toThrow();
      }
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('serves multiple lease resources through one browser link and one MCP token', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const taxesRoot = path.join(tmp, 'taxes');
    const receiptsRoot = path.join(tmp, 'receipts');
    const privateRoot = path.join(tmp, 'private');
    fs.mkdirSync(taxesRoot, { recursive: true });
    fs.mkdirSync(receiptsRoot, { recursive: true });
    fs.mkdirSync(privateRoot, { recursive: true });
    fs.writeFileSync(path.join(taxesRoot, 'w2.md'), '# W2\nalpha taxes', 'utf-8');
    fs.writeFileSync(path.join(receiptsRoot, 'meal.md'), '# Meal\nalpha receipt', 'utf-8');
    fs.writeFileSync(path.join(privateRoot, 'secret.md'), '# Secret\nalpha private', 'utf-8');
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'taxes', type: 'local_folder', path: '/taxes', root: taxesRoot },
        { name: 'receipts', type: 'local_folder', path: '/receipts', root: receiptsRoot },
        { name: 'private', type: 'local_folder', path: '/private', root: privateRoot },
      ],
    });
    const lease = createLease(leaseStorePath, {
      label: 'Sarah files',
      resources: [
        { path: '/taxes', sourcePath: '/taxes', type: 'folder' },
      ],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const index = new TextContextIndex({ mounts: config.mounts, indexPath: path.join(tmp, 'index.json') });
    await index.rebuild();
    const router = new ToolRouter(undefined, { contextIndex: index });
    await router.initialize();
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const rootListing = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}/files?token=${lease.token}`);
      expect(rootListing.status).toBe(200);
      const rootBody = await rootListing.json() as { entries: { path: string; type: string }[] };
      expect(rootBody.entries).toEqual([expect.objectContaining({ path: '/w2.md', type: 'file' })]);

      const download = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}/files/w2.md?token=${lease.token}`);
      expect(download.status).toBe(200);
      expect(await download.text()).toContain('alpha taxes');

      const sessionId = await initializeMcpSession(server.port, lease.token);
      const deniedBeforeAdd = await mcpJsonRequest(server.port, lease.token, sessionId, 2, 'tools/call', {
        name: 'read',
        arguments: { path: '/receipts/meal.md' },
      });
      expect(deniedBeforeAdd.result.isError).toBe(true);

      const updated = addLeaseResources(leaseStorePath, lease.record.id, [
        { path: '/receipts', sourcePath: '/receipts', type: 'folder' },
      ]);
      expect(updated?.tokenHash).toBe(lease.record.tokenHash);

      const updatedListing = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}/files?token=${lease.token}`);
      expect(updatedListing.status).toBe(200);
      const updatedBody = await updatedListing.json() as { entries: { path: string; type: string }[] };
      expect(updatedBody.entries).toEqual([
        expect.objectContaining({ path: '/receipts', type: 'directory' }),
        expect.objectContaining({ path: '/taxes', type: 'directory' }),
      ]);

      const listRoot = await mcpJsonRequest(server.port, lease.token, sessionId, 3, 'tools/call', {
        name: 'list',
        arguments: {},
      });
      const listPayload = JSON.parse(listRoot.result.content[0].text);
      expect(listPayload.entries.map((entry: { path: string }) => entry.path).sort()).toEqual(['/receipts', '/taxes']);

      const allowedAfterAdd = await mcpJsonRequest(server.port, lease.token, sessionId, 4, 'tools/call', {
        name: 'read',
        arguments: { path: '/receipts/meal.md' },
      });
      expect(allowedAfterAdd.result.isError).not.toBe(true);

      const denied = await mcpJsonRequest(server.port, lease.token, sessionId, 5, 'tools/call', {
        name: 'read',
        arguments: { path: '/private/secret.md' },
      });
      expect(denied.result.isError).toBe(true);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('supports upload-only leases without browse, download, or overwrite access', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const root = path.join(tmp, 'dropbox');
    fs.mkdirSync(path.join(root, 'nested'), { recursive: true });
    fs.writeFileSync(path.join(root, 'existing.txt'), 'already here', 'utf-8');
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'dropbox', type: 'local_folder', path: '/dropbox', root, writeAccess: true }],
    });
    const uploadLease = createLease(leaseStorePath, {
      label: 'Phone uploads',
      path: '/dropbox',
      permissions: ['upload'],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const readLease = createLease(leaseStorePath, {
      label: 'Read only',
      path: '/dropbox',
      permissions: ['read'],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const readUploadLease = createLease(leaseStorePath, {
      label: 'Read plus uploads',
      path: '/dropbox',
      permissions: ['read', 'upload'],
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const page = await fetch(`http://127.0.0.1:${server.port}/lease/${uploadLease.record.id}?token=${uploadLease.token}`);
      expect(page.status).toBe(200);
      expect(await page.text()).toContain('Upload-only lease');

      const listing = await fetch(`http://127.0.0.1:${server.port}/lease/${uploadLease.record.id}/files?token=${uploadLease.token}`);
      expect(listing.status).toBe(403);

      const readUploadPage = await fetch(`http://127.0.0.1:${server.port}/lease/${readUploadLease.record.id}?token=${readUploadLease.token}`);
      expect(readUploadPage.status).toBe(200);
      const readUploadHtml = await readUploadPage.text();
      expect(readUploadHtml).toContain('Upload to this folder');
      expect(readUploadHtml.indexOf('function uploadUrl(fileName)')).toBeLessThan(readUploadHtml.indexOf('async function loadListing()'));
      expect(readUploadHtml).toContain("uploadInput.addEventListener('change', () => uploadFiles(uploadInput.files).catch");
      expect(readUploadHtml).toContain('is too large for this public relay');
      expect(readUploadHtml).toContain('Upload failed before reaching mvmt');
      expect(readUploadHtml).toContain('id="upload-progress"');
      expect(readUploadHtml).toContain('id="upload-progress-bar"');
      expect(readUploadHtml).toContain('function showUploadProgress(message, percent)');
      expect(readUploadHtml).toContain('new XMLHttpRequest()');
      const readUploadListing = await fetch(`http://127.0.0.1:${server.port}/lease/${readUploadLease.record.id}/files?token=${readUploadLease.token}`);
      expect(((await readUploadListing.json()) as { canUpload: boolean }).canUpload).toBe(true);

      const uploaded = await fetch(`http://127.0.0.1:${server.port}/lease/${uploadLease.record.id}/files/nested/phone.txt?token=${uploadLease.token}`, {
        method: 'PUT',
        body: 'from phone',
      });
      expect(uploaded.status).toBe(201);
      expect(fs.readFileSync(path.join(root, 'nested', 'phone.txt'), 'utf-8')).toBe('from phone');
      expect(listLeases(leaseStorePath)[0].uploadCount).toBe(1);

      // Upload-only leases never overwrite. A name collision is suffixed.
      const overwrite = await fetch(`http://127.0.0.1:${server.port}/lease/${uploadLease.record.id}/files/existing.txt?token=${uploadLease.token}`, {
        method: 'PUT',
        body: 'replace',
      });
      expect(overwrite.status).toBe(201);
      expect(((await overwrite.json()) as { filename: string }).filename).toBe('existing (2).txt');
      expect(fs.readFileSync(path.join(root, 'existing.txt'), 'utf-8')).toBe('already here');
      expect(fs.readFileSync(path.join(root, 'existing (2).txt'), 'utf-8')).toBe('replace');

      const traversal = await fetch(`http://127.0.0.1:${server.port}/lease/${uploadLease.record.id}/files/%2e%2e/escape.txt?token=${uploadLease.token}`, {
        method: 'PUT',
        body: 'escape',
      });
      expect(traversal.status).toBe(404);

      const readOnlyPut = await fetch(`http://127.0.0.1:${server.port}/lease/${readLease.record.id}/files/nope.txt?token=${readLease.token}`, {
        method: 'PUT',
        body: 'nope',
      });
      expect(readOnlyPut.status).toBe(403);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('enforces lease token, expiry, revocation, range, and path boundaries', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const root = path.join(tmp, 'leased');
    fs.mkdirSync(root, { recursive: true });
    fs.writeFileSync(path.join(root, 'payload.bin'), Buffer.from([10, 11, 12, 13, 14]));
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'leased', type: 'local_folder', path: '/leased', root }],
    });
    const active = createLease(leaseStorePath, {
      label: 'Active',
      path: '/leased',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const expired = createLease(leaseStorePath, {
      label: 'Expired',
      path: '/leased',
      expiresAt: new Date(Date.now() - 60_000).toISOString(),
    });
    const revoked = createLease(leaseStorePath, {
      label: 'Revoked',
      path: '/leased',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    revokeLease(leaseStorePath, revoked.record.id);
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const invalid = await fetch(`http://127.0.0.1:${server.port}/lease/${active.record.id}/files?token=wrong`);
      expect(invalid.status).toBe(401);

      const gone = await fetch(`http://127.0.0.1:${server.port}/lease/${expired.record.id}/files?token=${expired.token}`);
      expect(gone.status).toBe(410);

      const revokedResponse = await fetch(`http://127.0.0.1:${server.port}/lease/${revoked.record.id}/files?token=${revoked.token}`);
      expect(revokedResponse.status).toBe(410);

      const range = await fetch(`http://127.0.0.1:${server.port}/lease/${active.record.id}/files/payload.bin?token=${active.token}`, {
        headers: { Range: 'bytes=1-3' },
      });
      expect(range.status).toBe(206);
      expect(range.headers.get('content-range')).toBe('bytes 1-3/5');
      expect([...Buffer.from(await range.arrayBuffer())]).toEqual([11, 12, 13]);

      const traversal = await fetch(`http://127.0.0.1:${server.port}/lease/${active.record.id}/files/%2e%2e/payload.bin?token=${active.token}`);
      expect(traversal.status).toBe(404);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });
});

describe('dashboard access', () => {
  it('serves mounted files through the cookie-authenticated FS API with range, write, remove, and audit', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-fs-api-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const photosRoot = path.join(tmp, 'photos');
    fs.mkdirSync(photosRoot, { recursive: true });
    const imageBytes = Buffer.from([0, 1, 2, 3, 4, 5, 6]);
    fs.writeFileSync(path.join(photosRoot, 'photo.jpg'), imageBytes);
    createPrivilegedUser(usersPath, { username: 'photo-admin', password: 'correct horse battery staple' });
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'photos', type: 'local_folder', path: '/photos', root: photosRoot, writeAccess: true },
      ],
    });
    const auditEntries: unknown[] = [];
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      privilegedUsersPath: usersPath,
      audit: {
        record: (entry) => auditEntries.push(entry),
        recordHttp: () => undefined,
      },
    });
    try {
      const blocked = await fetch(`http://127.0.0.1:${server.port}/api/fs/sources`);
      expect(blocked.status).toBe(401);

      const cookie = await loginDashboard(server.port, 'photo-admin', 'correct horse battery staple');
      const sources = await fetch(`http://127.0.0.1:${server.port}/api/fs/sources`, {
        headers: { Cookie: cookie },
      });
      expect(sources.status).toBe(200);
      expect(await sources.json()).toMatchObject({
        sources: [
          expect.objectContaining({ name: 'photos', path: '/photos', type: 'directory', writeAccess: true }),
        ],
      });

      const listing = await fetch(`http://127.0.0.1:${server.port}/api/fs/list?path=${encodeURIComponent('/photos')}`, {
        headers: { Cookie: cookie },
      });
      expect(listing.status).toBe(200);
      expect(await listing.json()).toMatchObject({
        path: '/photos',
        entries: [
          expect.objectContaining({ name: 'photo.jpg', path: '/photos/photo.jpg', type: 'file', size: imageBytes.length }),
        ],
      });

      const cold = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/photos/photo.jpg')}`, {
        headers: { Cookie: cookie },
      });
      expect(cold.status).toBe(200);
      const etag = cold.headers.get('etag');
      expect(etag).toMatch(/^W\/"[0-9a-f]{24}"$/);
      expect(cold.headers.get('cache-control')).toBe('private, max-age=300, must-revalidate');
      expect(cold.headers.get('vary')).toBe('Range');
      expect(cold.headers.get('last-modified')).toBeTruthy();
      expect([...Buffer.from(await cold.arrayBuffer())]).toEqual([...imageBytes]);

      const cached = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/photos/photo.jpg')}`, {
        headers: { Cookie: cookie, 'If-None-Match': etag ?? '' },
      });
      expect(cached.status).toBe(304);
      expect(cached.headers.get('etag')).toBe(etag);
      expect(cached.headers.get('vary')).toBe('Range');
      expect(await cached.text()).toBe('');

      const cachedByDate = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/photos/photo.jpg')}`, {
        headers: { Cookie: cookie, 'If-Modified-Since': cold.headers.get('last-modified') ?? '' },
      });
      expect(cachedByDate.status).toBe(304);
      expect(cachedByDate.headers.get('etag')).toBe(etag);
      expect(await cachedByDate.text()).toBe('');

      const updatedImageBytes = Buffer.from([9, 8, 7, 6, 5, 4, 3, 2]);
      fs.writeFileSync(path.join(photosRoot, 'photo.jpg'), updatedImageBytes);
      const refreshed = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/photos/photo.jpg')}`, {
        headers: { Cookie: cookie, 'If-None-Match': etag ?? '' },
      });
      expect(refreshed.status).toBe(200);
      expect(refreshed.headers.get('etag')).not.toBe(etag);
      expect([...Buffer.from(await refreshed.arrayBuffer())]).toEqual([...updatedImageBytes]);

      const range = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/photos/photo.jpg')}`, {
        headers: { Cookie: cookie, Range: 'bytes=2-4' },
      });
      expect(range.status).toBe(206);
      expect(range.headers.get('cache-control')).toBe('private, max-age=300, must-revalidate');
      expect(range.headers.get('vary')).toBe('Range');
      expect(range.headers.get('etag')).toBe(refreshed.headers.get('etag'));
      expect(range.headers.get('content-range')).toBe('bytes 2-4/8');
      expect(range.headers.get('content-type')).toBe('image/jpeg');
      expect([...Buffer.from(await range.arrayBuffer())]).toEqual([7, 6, 5]);

      const writeBody = '{"saved":true}';
      const write = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/photos/upload.json')}`, {
        method: 'PUT',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: writeBody,
      });
      expect(write.status).toBe(200);
      expect(fs.readFileSync(path.join(photosRoot, 'upload.json'), 'utf-8')).toBe(writeBody);

      const remove = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/photos/upload.json')}`, {
        method: 'DELETE',
        headers: { Cookie: cookie },
      });
      expect(remove.status).toBe(200);
      expect(fs.existsSync(path.join(photosRoot, 'upload.json'))).toBe(false);

      expect(auditEntries).toEqual(
        expect.arrayContaining([
          expect.objectContaining({ tool: 'fs.read', clientId: 'photo-admin', isError: false }),
          expect.objectContaining({ tool: 'fs.read', clientId: 'photo-admin', isError: false, deniedReason: 'cache_hit' }),
          expect.objectContaining({ tool: 'fs.write', clientId: 'photo-admin', isError: false }),
          expect.objectContaining({ tool: 'fs.remove', clientId: 'photo-admin', isError: false }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('caps cookie-authenticated FS uploads before writing to disk', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-fs-api-cap-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const docsRoot = path.join(tmp, 'docs');
    fs.mkdirSync(docsRoot, { recursive: true });
    createPrivilegedUser(usersPath, { username: 'writer', password: 'correct horse battery staple' });
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'docs', type: 'local_folder', path: '/docs', root: docsRoot, writeAccess: true },
      ],
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      privilegedUsersPath: usersPath,
      maxFsUploadBytes: 4,
    });
    try {
      const cookie = await loginDashboard(server.port, 'writer', 'correct horse battery staple');
      const response = await fetch(`http://127.0.0.1:${server.port}/api/fs/file?path=${encodeURIComponent('/docs/too-large.bin')}`, {
        method: 'PUT',
        headers: { Cookie: cookie, 'Content-Type': 'application/octet-stream' },
        body: Buffer.from([0, 1, 2, 3, 4]),
      });
      expect(response.status).toBe(413);
      expect(await response.json()).toEqual({ error: 'fs_upload_too_large' });
      expect(fs.existsSync(path.join(docsRoot, 'too-large.bin'))).toBe(false);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('exposes unavailable mounts in /api/fs/list to local admins only', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-fs-unavailable-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const aliveRoot = path.join(tmp, 'alive');
    fs.mkdirSync(aliveRoot, { recursive: true });
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });
    createPrivilegedUser(usersPath, { username: 'viewer', password: 'correct horse battery staple', admin: false });
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'alive', type: 'local_folder', path: '/alive', root: aliveRoot },
        { name: 'broken', type: 'local_folder', path: '/broken', root: path.join(tmp, 'does-not-exist') },
      ],
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      privilegedUsersPath: usersPath,
    });
    try {
      const adminCookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const adminListing = await fetch(`http://127.0.0.1:${server.port}/api/fs/list?path=${encodeURIComponent('/')}`, {
        headers: { Cookie: adminCookie },
      });
      expect(adminListing.status).toBe(200);
      const adminBody = await adminListing.json() as { entries: Array<{ path: string; unavailable?: boolean }> };
      expect(adminBody.entries.find((entry) => entry.path === '/alive')).toMatchObject({ path: '/alive' });
      expect(adminBody.entries.find((entry) => entry.path === '/alive')?.unavailable).toBeUndefined();
      expect(adminBody.entries.find((entry) => entry.path === '/broken')).toMatchObject({ path: '/broken', unavailable: true });

      const viewerCookie = await loginDashboard(server.port, 'viewer', 'correct horse battery staple');
      const viewerListing = await fetch(`http://127.0.0.1:${server.port}/api/fs/list?path=${encodeURIComponent('/')}`, {
        headers: { Cookie: viewerCookie },
      });
      expect(viewerListing.status).toBe(200);
      const viewerBody = await viewerListing.json() as { entries: Array<{ path: string; unavailable?: boolean }> };
      expect(viewerBody.entries.map((entry) => entry.path)).toEqual(['/alive']);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('renders the dashboard without referencing the retired /dashboard/api/files endpoint', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-html-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/dashboard`);
      expect(response.status).toBe(200);
      const body = await response.text();
      expect(body).not.toContain('/dashboard/api/files');
      expect(body).toContain('/api/fs/list');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('lists installed apps and serves them under /apps/:id behind dashboard auth', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-apps-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'curious', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const blockedList = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/apps`);
      expect(blockedList.status).toBe(401);
      const blockedApp = await fetch(`http://127.0.0.1:${server.port}/apps/file-inspector`);
      expect(blockedApp.status).toBe(401);

      const cookie = await loginDashboard(server.port, 'curious', 'correct horse battery staple');

      const list = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/apps`, { headers: { Cookie: cookie } });
      expect(list.status).toBe(200);
      const listBody = await list.json() as { apps: Array<{ id: string; label: string; description: string }> };
      expect(listBody.apps).toEqual(expect.arrayContaining([
        expect.objectContaining({ id: 'file-inspector', label: 'File Inspector' }),
        expect.objectContaining({ id: 'photos', label: 'Photos' }),
      ]));

      const served = await fetch(`http://127.0.0.1:${server.port}/apps/file-inspector`, { headers: { Cookie: cookie } });
      expect(served.status).toBe(200);
      expect(served.headers.get('content-type')).toMatch(/text\/html/);
      const html = await served.text();
      expect(html).toContain('File Inspector');
      expect(html).toContain('/api/fs/sources');
      expect(html).toContain('/api/fs/list');
      expect(html).toContain('/api/fs/stat');
      expect(html).toContain('ENTRY_BATCH_SIZE = 200');
      expect(html).toContain('Show more entries');
      expect(html).toContain('Refresh stat');

      const photos = await fetch(`http://127.0.0.1:${server.port}/apps/photos`, { headers: { Cookie: cookie } });
      expect(photos.status).toBe(200);
      expect(photos.headers.get('content-type')).toMatch(/text\/html/);
      const photosHtml = await photos.text();
      expect(photosHtml).toContain('Photos');
      expect(photosHtml).toContain('/api/fs/sources');
      expect(photosHtml).toContain('/api/fs/list');
      expect(photosHtml).toContain('/api/fs/file?path=');
      expect(photosHtml).toContain('PHOTO_BATCH_SIZE = 48');
      expect(photosHtml).toContain('Show more photos');
      expect(photosHtml).not.toContain('/api/fs/write');

      const missing = await fetch(`http://127.0.0.1:${server.port}/apps/unknown-app`, { headers: { Cookie: cookie } });
      expect(missing.status).toBe(404);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('renders the dashboard Apps tab nav, panel, and loader script', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-apps-tab-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/dashboard`);
      expect(response.status).toBe(200);
      const body = await response.text();
      expect(body).toContain('data-view="apps"');
      expect(body).toContain('id="view-apps-panel"');
      expect(body).toContain('id="apps-grid"');
      expect(body).toContain("api('/dashboard/api/apps')");
      expect(body).toContain('Apps failed to load. Click Apps again to retry.');
      expect(body).toContain('state.appsLoaded = false;');
      expect(body).toContain("appBasePath() + '/apps/'");
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('keeps the file-inspector SPA relay-prefix-safe', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-file-inspector-prefix-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'curious', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'curious', 'correct horse battery staple');
      const served = await fetch(`http://127.0.0.1:${server.port}/apps/file-inspector`, { headers: { Cookie: cookie } });
      expect(served.status).toBe(200);
      const html = await served.text();

      // Local path: no workspace prefix to add.
      expect(appBaseFromHtml(html, '/apps/file-inspector')).toBe('');
      expect(appBaseFromHtml(html, '/apps/file-inspector/')).toBe('');
      // Relay path: SPA must preserve the workspace prefix when assembling
      // /api/fs/* and back-to-dashboard URLs, or remote calls land at the
      // relay root instead of the agent.
      expect(appBaseFromHtml(html, '/t/demo/apps/file-inspector')).toBe('/t/demo');
      expect(appBaseFromHtml(html, '/t/demo/apps/file-inspector/')).toBe('/t/demo');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('keeps the Photos SPA relay-prefix-safe', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-photos-prefix-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'curious', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'curious', 'correct horse battery staple');
      const served = await fetch(`http://127.0.0.1:${server.port}/apps/photos`, { headers: { Cookie: cookie } });
      expect(served.status).toBe(200);
      const html = await served.text();

      expect(appBaseFromHtml(html, '/apps/photos')).toBe('');
      expect(appBaseFromHtml(html, '/apps/photos/')).toBe('');
      expect(appBaseFromHtml(html, '/t/demo/apps/photos')).toBe('/t/demo');
      expect(appBaseFromHtml(html, '/t/demo/apps/photos/')).toBe('/t/demo');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('logs in privileged users, creates read/upload/two-way leases, and suffixes collision uploads', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const readRoot = path.join(tmp, 'read');
    const writeRoot = path.join(tmp, 'write');
    fs.mkdirSync(readRoot, { recursive: true });
    fs.mkdirSync(writeRoot, { recursive: true });
    fs.writeFileSync(path.join(readRoot, 'note.txt'), 'read note', 'utf-8');
    createPrivilegedUser(usersPath, { username: 'sarah', password: 'correct horse battery staple' });
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'read', type: 'local_folder', path: '/read', root: readRoot },
        { name: 'write', type: 'local_folder', path: '/write', root: writeRoot, writeAccess: true },
      ],
    });
    const requestLogs: Array<{ kind: string; path: string; status: number; clientId?: string }> = [];
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
      requestLog: (entry) => requestLogs.push(entry),
    });
    try {
      const cookie = await loginDashboard(server.port, 'sarah', 'correct horse battery staple');

      // Read lease on a file: only 'read' mode is valid.
      const readLease = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/read/note.txt', label: 'Read note', mode: 'read', expires: '1h', maxDownloads: 1 }),
      });
      expect(readLease.status).toBe(201);
      const readBody = await readLease.json() as { lease: { id: string; url: string; permissions: string[]; maxDownloads?: number } };
      expect(readBody.lease.permissions).toEqual(['read']);
      expect(readBody.lease.maxDownloads).toBe(1);
      const listedLeases = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        headers: { Cookie: cookie },
      });
      expect(listedLeases.status).toBe(200);
      const listedBody = await listedLeases.json() as { leases: { id: string; url?: string }[] };
      expect(listedBody.leases.find((lease) => lease.id === readBody.lease.id)?.url).toBe(readBody.lease.url);

      const loggedLeases = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        headers: { Cookie: cookie },
      });
      expect(loggedLeases.status).toBe(200);
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'dashboard.leases',
            path: '/dashboard/api/leases',
            status: 200,
            clientId: 'sarah',
          }),
        ]),
      );

      // 'write' mode is no longer accepted by the dashboard.
      const writeBlocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/write', label: 'Bad write', mode: 'write' }),
      });
      expect(writeBlocked.status).toBe(400);
      expect((await writeBlocked.json() as { error: string }).error).toBe('invalid_mode');

      // Upload mode on a file target is rejected — needs a folder.
      const uploadOnFile = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/read/note.txt', label: 'Bad upload', mode: 'upload' }),
      });
      expect(uploadOnFile.status).toBe(400);
      expect((await uploadOnFile.json() as { error: string }).error).toBe('mode_requires_folder');

      // Upload mode on a read-only mount is rejected.
      const uploadOnReadOnly = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/read', label: 'Bad upload', mode: 'upload' }),
      });
      expect(uploadOnReadOnly.status).toBe(400);
      expect((await uploadOnReadOnly.json() as { error: string }).error).toBe('mount_read_only');

      const uploadWithDownloadLimit = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/write', label: 'Bad limit', mode: 'upload', maxDownloads: 1 }),
      });
      expect(uploadWithDownloadLimit.status).toBe(400);
      expect((await uploadWithDownloadLimit.json() as { error: string }).error).toBe('download_limit_requires_read');

      const unlimitedDownloads = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/read/note.txt', label: 'Unlimited note', mode: 'read', maxDownloads: -1 }),
      });
      expect(unlimitedDownloads.status).toBe(201);
      const unlimitedBody = await unlimitedDownloads.json() as { lease: { maxDownloads?: number } };
      expect(unlimitedBody.lease.maxDownloads).toBeUndefined();

      // Two-way lease (read + upload) on a writable folder.
      const twoWay = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/write', label: 'Inbox', mode: 'two-way' }),
      });
      expect(twoWay.status).toBe(201);
      const twoWayBody = await twoWay.json() as { lease: { id: string; url: string; permissions: string[] } };
      expect(twoWayBody.lease.permissions).toEqual(['read', 'upload']);

      const twoWayUrl = new URL(twoWayBody.lease.url);
      const token = twoWayUrl.searchParams.get('token');

      // First upload of report.txt → stored as-is.
      const firstUpload = await fetch(`${twoWayUrl.origin}${twoWayUrl.pathname}/files/report.txt?token=${token}`, {
        method: 'PUT',
        body: 'first',
      });
      expect(firstUpload.status).toBe(201);
      const firstBody = await firstUpload.json() as { path: string; filename: string };
      expect(firstBody.filename).toBe('report.txt');
      expect(firstBody.path).toBe('/report.txt');

      // Second upload of report.txt → suffixed to report (2).txt; the original file is preserved.
      const secondUpload = await fetch(`${twoWayUrl.origin}${twoWayUrl.pathname}/files/report.txt?token=${token}`, {
        method: 'PUT',
        body: 'second',
      });
      expect(secondUpload.status).toBe(201);
      const secondBody = await secondUpload.json() as { path: string; filename: string };
      expect(secondBody.filename).toBe('report (2).txt');
      expect(secondBody.path).toBe('/report (2).txt');
      expect(fs.readFileSync(path.join(writeRoot, 'report.txt'), 'utf-8')).toBe('first');
      expect(fs.readFileSync(path.join(writeRoot, 'report (2).txt'), 'utf-8')).toBe('second');

      // Third upload bumps to (3).
      const thirdUpload = await fetch(`${twoWayUrl.origin}${twoWayUrl.pathname}/files/report.txt?token=${token}`, {
        method: 'PUT',
        body: 'third',
      });
      expect(thirdUpload.status).toBe(201);
      expect(((await thirdUpload.json()) as { filename: string }).filename).toBe('report (3).txt');

      // Two-way leases cannot DELETE existing files.
      const deleteBlocked = await fetch(`${twoWayUrl.origin}${twoWayUrl.pathname}/files/report.txt?token=${token}`, {
        method: 'DELETE',
      });
      expect(deleteBlocked.status).toBe(403);
      expect(fs.existsSync(path.join(writeRoot, 'report.txt'))).toBe(true);

      // Revoke the read lease.
      const revoke = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${readBody.lease.id}/revoke`, {
        method: 'POST',
        headers: { Cookie: cookie },
      });
      expect(revoke.status).toBe(200);
      expect(listLeases(leaseStorePath).find((lease) => lease.id === readBody.lease.id)?.revokedAt).toBeTruthy();
      expect(findLeaseSecret(leaseSecretsPathForLeaseStore(leaseStorePath), readBody.lease.id)).toBeUndefined();
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('adds, edits, and removes mounts through dashboard APIs', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-mounts-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'photos');
    const otherRoot = path.join(tmp, 'reports');
    fs.mkdirSync(mountRoot, { recursive: true });
    fs.mkdirSync(otherRoot, { recursive: true });
    fs.writeFileSync(path.join(mountRoot, 'one.txt'), 'one', 'utf-8');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });
    await saveConfig(configPath, parseConfig({ version: 1 }));

    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      configPath,
      leaseMounts: () => readConfig(configPath).mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
      localPathPicker: async (kind) => kind === 'folder' ? mountRoot : path.join(mountRoot, 'one.txt'),
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');

      const initialMounts = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, { headers: { Cookie: cookie } });
      const initialBody = await initialMounts.json() as { canManage: boolean; mounts: unknown[] };
      expect(initialBody.canManage).toBe(true);
      expect(initialBody.mounts).toEqual([]);

      const missingRoot = path.join(tmp, 'missing');
      const invalidAdd = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: missingRoot, path: '/missing' }),
      });
      expect(invalidAdd.status).toBe(400);
      expect(await invalidAdd.json()).toMatchObject({ error: 'invalid_root' });
      expect(readConfig(configPath).mounts).toEqual([]);

      const relativeAdd = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: 'relative/path', path: '/relative' }),
      });
      expect(relativeAdd.status).toBe(400);
      expect(await relativeAdd.json()).toMatchObject({ error: 'invalid_root' });
      expect(readConfig(configPath).mounts).toEqual([]);

      const nulAdd = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: `${mountRoot}\0suffix`, path: '/nul' }),
      });
      expect(nulAdd.status).toBe(400);
      expect(await nulAdd.json()).toMatchObject({ error: 'invalid_root' });
      expect(readConfig(configPath).mounts).toEqual([]);

      const pickedFolder = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/local-path-picker`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ kind: 'folder' }),
      });
      expect(pickedFolder.status).toBe(200);
      expect(await pickedFolder.json()).toEqual({ path: mountRoot });

      const add = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: mountRoot, path: '/photos', writeAccess: false }),
      });
      expect(add.status).toBe(201);
      const addBody = await add.json() as { mount: { name: string; path: string; writeAccess: boolean } };
      expect(addBody.mount.path).toBe('/photos');
      expect(addBody.mount.writeAccess).toBe(false);

      const persistedAfterAdd = readConfig(configPath);
      expect(persistedAfterAdd.mounts).toHaveLength(1);
      expect(persistedAfterAdd.mounts[0]?.path).toBe('/photos');

      const filesListing = await fetch(`http://127.0.0.1:${server.port}/api/fs/list?path=/photos`, {
        headers: { Cookie: cookie },
      });
      expect(filesListing.status).toBe(200);
      const filesBody = await filesListing.json() as { entries: { name: string }[] };
      expect(filesBody.entries.map((entry) => entry.name)).toContain('one.txt');

      const duplicate = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: otherRoot, path: '/photos' }),
      });
      expect(duplicate.status).toBe(400);

      const edit = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts/${encodeURIComponent(addBody.mount.name)}`, {
        method: 'PATCH',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ writeAccess: true, description: 'family photos' }),
      });
      expect(edit.status).toBe(200);
      const editBody = await edit.json() as { mount: { writeAccess: boolean; description?: string } };
      expect(editBody.mount.writeAccess).toBe(true);
      expect(editBody.mount.description).toBe('family photos');

      const invalidEdit = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts/${encodeURIComponent(addBody.mount.name)}`, {
        method: 'PATCH',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: missingRoot }),
      });
      expect(invalidEdit.status).toBe(400);
      expect(await invalidEdit.json()).toMatchObject({ error: 'invalid_root' });

      const remove = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts/${encodeURIComponent(addBody.mount.name)}`, {
        method: 'DELETE',
        headers: { Cookie: cookie },
      });
      expect(remove.status).toBe(200);
      expect(readConfig(configPath).mounts).toEqual([]);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('shows stale root sources to local admins so they can remove them', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-stale-mount-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const configPath = path.join(tmp, 'config.yaml');
    const liveRoot = path.join(tmp, 'live');
    const missingRoot = path.join(tmp, 'missing');
    fs.mkdirSync(liveRoot, { recursive: true });
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });
    await saveConfig(configPath, parseConfig({
      version: 1,
      mounts: [
        { name: 'live', type: 'local_folder', path: '/live', root: liveRoot },
        { name: 'gone', type: 'local_folder', path: '/gone', root: missingRoot },
      ],
    }));

    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      configPath,
      leaseMounts: () => readConfig(configPath).mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const listing = await fetch(`http://127.0.0.1:${server.port}/api/fs/list?path=/`, {
        headers: { Cookie: cookie },
      });
      expect(listing.status).toBe(200);
      const body = await listing.json() as { entries: { path: string; unavailable?: boolean }[] };
      expect(body.entries).toEqual(expect.arrayContaining([
        expect.objectContaining({ path: '/live' }),
        expect.objectContaining({ path: '/gone', unavailable: true }),
      ]));
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('keeps source management local-only for remote dashboard sessions', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-remote-local-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'photos');
    fs.mkdirSync(mountRoot, { recursive: true });
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });
    await saveConfig(configPath, parseConfig({
      version: 1,
      mounts: [{ name: 'photos', type: 'local_folder', path: '/photos', root: mountRoot, writeAccess: true }],
    }));

    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      configPath,
      leaseMounts: () => readConfig(configPath).mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const remoteHeaders = { Cookie: cookie, 'X-MVMT-Transport': 'relay' };

      const me = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/me`, { headers: remoteHeaders });
      expect(me.status).toBe(200);
      expect(((await me.json()) as { localOwner: boolean }).localOwner).toBe(false);

      const mounts = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, { headers: remoteHeaders });
      expect(mounts.status).toBe(200);
      const mountsBody = await mounts.json() as { canManage: boolean; mounts: { root?: string }[] };
      expect(mountsBody.canManage).toBe(false);
      expect(mountsBody.mounts[0]?.root).toBeUndefined();

      const status = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/status`, { headers: remoteHeaders });
      expect(status.status).toBe(403);
      expect((await status.json()) as { error: string }).toEqual({ error: 'local_dashboard_required' });

      const blocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { ...remoteHeaders, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: mountRoot, path: '/blocked' }),
      });
      expect(blocked.status).toBe(403);
      expect((await blocked.json()) as { error: string }).toEqual({ error: 'local_dashboard_required' });
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rotates a lease token, invalidating the old URL and issuing a new one', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-rotate-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const readRoot = path.join(tmp, 'read');
    fs.mkdirSync(readRoot, { recursive: true });
    fs.writeFileSync(path.join(readRoot, 'note.txt'), 'hello', 'utf-8');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple' });

    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'read', type: 'local_folder', path: '/read', root: readRoot }],
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const create = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/read/note.txt', label: 'Note', mode: 'read', expires: '1h' }),
      });
      expect(create.status).toBe(201);
      const createBody = await create.json() as { lease: { id: string; url: string } };
      const originalUrl = new URL(createBody.lease.url);
      const originalToken = originalUrl.searchParams.get('token');
      expect(originalToken).toBeTruthy();
      expect(findLeaseSecret(leaseSecretsPathForLeaseStore(leaseStorePath), createBody.lease.id)?.token).toBe(originalToken);

      const rotate = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${encodeURIComponent(createBody.lease.id)}/rotate`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: '{}',
      });
      expect(rotate.status).toBe(200);
      const rotated = await rotate.json() as { lease: { url: string } };
      const newUrl = new URL(rotated.lease.url);
      const newToken = newUrl.searchParams.get('token');
      expect(newToken).toBeTruthy();
      expect(newToken).not.toBe(originalToken);
      expect(findLeaseSecret(leaseSecretsPathForLeaseStore(leaseStorePath), createBody.lease.id)?.token).toBe(newToken);

      const oldFetch = await fetch(`${originalUrl.origin}/lease/${createBody.lease.id}/files?token=${originalToken}`);
      expect(oldFetch.status).toBe(401);

      const newFetch = await fetch(`${newUrl.origin}/lease/${createBody.lease.id}/files?token=${newToken}`);
      expect(newFetch.status).toBe(200);

      const revoked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${encodeURIComponent(createBody.lease.id)}/revoke`, {
        method: 'POST',
        headers: { Cookie: cookie },
      });
      expect(revoked.status).toBe(200);
      const rotateRevoked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${encodeURIComponent(createBody.lease.id)}/rotate`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: '{}',
      });
      expect(rotateRevoked.status).toBe(410);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  // Creating a share link in the dashboard is the explicit publish
  // gesture, so the lease is minted published. The publish endpoint then
  // toggles public-tunnel exposure without revoking the lease.
  it('mints dashboard leases published and toggles public-tunnel exposure via the publish endpoint', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-publish-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const root = path.join(tmp, 'read');
    fs.mkdirSync(root, { recursive: true });
    fs.writeFileSync(path.join(root, 'note.txt'), 'hello');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple' });
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'read', type: 'local_folder', path: '/read', root }],
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const create = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/read/note.txt', label: 'Note', mode: 'read', expires: '1h' }),
      });
      expect(create.status).toBe(201);
      const createBody = await create.json() as { lease: { id: string; url: string; published?: boolean } };
      // Dashboard-minted leases are explicitly published.
      expect(createBody.lease.published).toBe(true);
      expect(listLeases(leaseStorePath)[0]!.published).toBe(true);

      const leaseId = createBody.lease.id;
      const token = new URL(createBody.lease.url).searchParams.get('token');
      const relayHeaders = { 'X-MVMT-Transport': 'relay' };

      // Published: a relay-forwarded request works.
      const beforeUnpublish = await fetch(
        `http://127.0.0.1:${server.port}/lease/${leaseId}/files?token=${token}`,
        { headers: relayHeaders },
      );
      expect(beforeUnpublish.status).toBe(200);

      const unpublish = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${leaseId}/publish`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ published: false }),
      });
      expect(unpublish.status).toBe(200);
      expect((await unpublish.json() as { lease: { published: boolean } }).lease.published).toBe(false);

      const malformedPublish = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${leaseId}/publish`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      });
      expect(malformedPublish.status).toBe(400);
      expect((await malformedPublish.json() as { error: string }).error).toBe('published_required');
      expect(listLeases(leaseStorePath)[0]!.published).toBe(false);

      // Capability-only now: relay is rejected, localhost still works.
      const afterUnpublishRelay = await fetch(
        `http://127.0.0.1:${server.port}/lease/${leaseId}/files?token=${token}`,
        { headers: relayHeaders },
      );
      expect(afterUnpublishRelay.status).toBe(403);
      const afterUnpublishLocal = await fetch(
        `http://127.0.0.1:${server.port}/lease/${leaseId}/files?token=${token}`,
      );
      expect(afterUnpublishLocal.status).toBe(200);

      // Re-publish restores public-tunnel reachability.
      const republish = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${leaseId}/publish`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ published: true }),
      });
      expect(republish.status).toBe(200);
      const afterRepublish = await fetch(
        `http://127.0.0.1:${server.port}/lease/${leaseId}/files?token=${token}`,
        { headers: relayHeaders },
      );
      expect(afterRepublish.status).toBe(200);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  // The relay routes /t/{slug}/** to the agent. If we built share URLs
  // from baseUrlFor(req) we'd drop the workspace prefix and the public
  // link would 404 on the relay's catch-all error page. Make sure both
  // creation and rotation keep the configured public URL intact.
  it('keeps the configured public URL prefix in lease share URLs', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-share-prefix-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const root = path.join(tmp, 'docs');
    fs.mkdirSync(root, { recursive: true });
    fs.writeFileSync(path.join(root, 'note.txt'), 'hello');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });

    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'docs', type: 'local_folder', path: '/docs', root }],
    });

    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
      resolvePublicBaseUrl: () => 'https://relay.example.com/t/demo',
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const create = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/docs/note.txt', label: 'Note', mode: 'read', expires: '1h' }),
      });
      expect(create.status).toBe(201);
      const createdBody = await create.json() as { lease: { id: string; url: string } };
      const createdUrl = new URL(createdBody.lease.url);
      expect(createdUrl.origin).toBe('https://relay.example.com');
      expect(createdUrl.pathname).toBe(`/t/demo/lease/${encodeURIComponent(createdBody.lease.id)}`);
      expect(createdUrl.searchParams.get('token')).toBeTruthy();

      const rotate = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases/${encodeURIComponent(createdBody.lease.id)}/rotate`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: '{}',
      });
      expect(rotate.status).toBe(200);
      const rotatedBody = await rotate.json() as { lease: { url: string } };
      const rotatedUrl = new URL(rotatedBody.lease.url);
      expect(rotatedUrl.pathname).toBe(`/t/demo/lease/${encodeURIComponent(createdBody.lease.id)}`);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  // The lease browser page extracts leaseId from location.pathname and
  // builds /lease/{id}/files URLs from it. When the URL is served behind
  // a relay prefix like /t/demo/lease/{id}, pathParts[1] is 'demo' (not
  // the lease id) and fetches strip the prefix. Anchor on 'lease' so the
  // page works in both shapes.
  it('lease pages anchor leaseId and base prefix on the lease path segment', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-lease-page-prefix-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const root = path.join(tmp, 'docs');
    fs.mkdirSync(root, { recursive: true });
    fs.writeFileSync(path.join(root, 'note.txt'), 'hi');
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'docs', type: 'local_folder', path: '/docs', root }],
    });
    const lease = createLease(leaseStorePath, {
      label: 'Docs',
      path: '/docs',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/lease/${lease.record.id}?token=${lease.token}`);
      expect(response.status).toBe(200);
      const html = await response.text();
      expect(html).toContain("const leaseIdx = pathParts.indexOf('lease');");
      expect(html).toContain("const basePrefix = leaseIdx > 0 ? '/' + pathParts.slice(0, leaseIdx).join('/') : '';");
      expect(html).toContain("new URL(basePrefix + '/lease/' + encodeURIComponent(leaseId)");
      // No remaining absolute-path URL constructions that ignore the prefix.
      expect(html).not.toContain("new URL('/lease/' + encodeURIComponent(leaseId)");
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('refuses mount mutation when no configPath is wired', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-no-config-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });

    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const mounts = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, { headers: { Cookie: cookie } });
      const mountsBody = await mounts.json() as { canManage: boolean };
      expect(mountsBody.canManage).toBe(false);

      const blocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: tmp, path: '/blocked' }),
      });
      expect(blocked.status).toBe(403);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('refuses mount mutation for non-admin dashboard users', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-non-admin-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const configPath = path.join(tmp, 'config.yaml');
    const mountRoot = path.join(tmp, 'photos');
    fs.mkdirSync(mountRoot, { recursive: true });
    createPrivilegedUser(usersPath, { username: 'member', password: 'correct horse battery staple' });
    await saveConfig(configPath, parseConfig({
      version: 1,
      mounts: [{ name: 'photos', type: 'local_folder', path: '/photos', root: mountRoot }],
    }));

    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      configPath,
      leaseMounts: () => readConfig(configPath).mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'member', 'correct horse battery staple');

      const me = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/me`, { headers: { Cookie: cookie } });
      expect(((await me.json()) as { user: { admin: boolean } }).user.admin).toBe(false);

      const mounts = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, { headers: { Cookie: cookie } });
      const mountsBody = await mounts.json() as { canManage: boolean; mounts: { root?: string }[] };
      expect(mountsBody.canManage).toBe(false);
      expect(mountsBody.mounts[0]?.root).toBeUndefined();

      const addBlocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: mountRoot }),
      });
      expect(addBlocked.status).toBe(403);

      const patchBlocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts/photos`, {
        method: 'PATCH',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ writeAccess: true }),
      });
      expect(patchBlocked.status).toBe(403);

      const deleteBlocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts/photos`, {
        method: 'DELETE',
        headers: { Cookie: cookie },
      });
      expect(deleteBlocked.status).toBe(403);

      // Non-admin can still create leases for already-mounted paths.
      const leaseOk = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/photos', label: 'Read photos', mode: 'read', expires: '1h' }),
      });
      expect(leaseOk.status).toBe(201);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('auto-mount-name uses basename only and suffixes on collision', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-name-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const configPath = path.join(tmp, 'config.yaml');
    const docs1 = path.join(tmp, 'parent-a', 'docs');
    const docs2 = path.join(tmp, 'parent-b', 'docs');
    fs.mkdirSync(docs1, { recursive: true });
    fs.mkdirSync(docs2, { recursive: true });
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });
    await saveConfig(configPath, parseConfig({ version: 1 }));

    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      configPath,
      leaseMounts: () => readConfig(configPath).mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');

      const first = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: docs1, path: '/docs-a' }),
      });
      expect(first.status).toBe(201);
      expect(((await first.json()) as { mount: { name: string } }).mount.name).toBe('docs');

      const second = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/mounts`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ root: docs2, path: '/docs-b' }),
      });
      expect(second.status).toBe(201);
      expect(((await second.json()) as { mount: { name: string } }).mount.name).toBe('docs-2');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('persists dashboard activity to the audit log', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-audit-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const auditPath = path.join(tmp, 'audit.log');
    const mountRoot = path.join(tmp, 'docs');
    fs.mkdirSync(mountRoot, { recursive: true });
    fs.writeFileSync(path.join(mountRoot, 'note.txt'), 'hello', 'utf-8');
    createPrivilegedUser(usersPath, { username: 'sarah', password: 'correct horse battery staple' });

    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'docs', type: 'local_folder', path: '/docs', root: mountRoot }],
    });
    const audit = createAuditLogger(auditPath);
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: config.mounts,
      leaseStorePath,
      privilegedUsersPath: usersPath,
      requestLog: (entry) => audit.recordHttp(entry),
    });
    try {
      // A failed login still produces an audit entry.
      await fetch(`http://127.0.0.1:${server.port}/dashboard/api/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'sarah', password: 'wrong' }),
      });

      const cookie = await loginDashboard(server.port, 'sarah', 'correct horse battery staple');
      await fetch(`http://127.0.0.1:${server.port}/api/fs/list?path=/`, { headers: { Cookie: cookie } });
      const leaseResp = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/leases`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: '/docs/note.txt', label: 'Note', mode: 'read', expires: '1h' }),
      });
      const leaseBody = await leaseResp.json() as { lease: { id: string; url: string } };
      const leaseUrl = new URL(leaseBody.lease.url);
      await fetch(`${leaseUrl.origin}/lease/${leaseBody.lease.id}/files?token=${leaseUrl.searchParams.get('token')}`);

      const lines = fs.readFileSync(auditPath, 'utf-8').split('\n').filter(Boolean).map((line) => JSON.parse(line) as Record<string, unknown>);
      const kinds = lines.map((entry) => entry.kind);
      expect(kinds).toContain('dashboard.login'); // includes both invalid_credentials and ok rows
      expect(kinds).toContain('fs.list');
      expect(kinds).toContain('dashboard.leases');
      expect(kinds).toContain('lease.request');
      // every entry is tagged http and carries source ip
      for (const entry of lines) {
        expect(entry.type).toBe('http');
        expect(typeof entry.ip).toBe('string');
        expect(entry.ip).toMatch(/^(127\.0\.0\.1|::1)$/);
      }
      // a failed login lands as 401 with detail invalid_credentials.
      const failedLogin = lines.find((entry) => entry.kind === 'dashboard.login' && entry.detail === 'invalid_credentials');
      expect(failedLogin?.status).toBe(401);
      // a successful lease fetch lands as 200 with the lease id in detail.
      const leaseFetch = lines.find((entry) => entry.kind === 'lease.request' && entry.status === 200);
      expect(leaseFetch).toBeDefined();
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('serves a dashboard HTML page with breadcrumbs, lease tabs, and mount management controls', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-html-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const leaseStorePath = path.join(tmp, '.leases.json');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      leaseStorePath,
      privilegedUsersPath: usersPath,
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/dashboard`);
      expect(response.status).toBe(200);
      expect(response.headers.get('content-type')).toMatch(/text\/html/);
      const html = await response.text();
      expect(html).toContain('id="crumbs"');
      expect(html).toContain('data-test="lease-tabs"');
      expect(html).toContain('id="add-mount"');
      expect(html).toContain('id="mount-pick-folder"');
      expect(html).toContain('id="mount-pick-file"');
      expect(html).toContain('data-view="files"');
      expect(html).toContain('id="files-grid"');
      expect(html).toContain('id="share-selected"');
      expect(html).toContain('id="settings-nav"');
      expect(html).toContain('Shared links');
      expect(html).toContain('id="mount-modal"');
      expect(html).toContain('id="lease-modal"');
      expect(html).toContain('id="context-menu"');
      expect(html).toContain('id="context-rename-source"');
      expect(html).toContain('id="context-remove-source"');
      expect(html).toContain('id="properties-modal"');
      expect(html).toContain('id="lease-mode"');
      expect(html).toContain('id="copy-url"');
      expect(html).toContain('<section id="login" class="panel login-card">');
      expect(html).toContain('@media (max-width:640px)');
      expect(html).toContain("setAttribute('data-label', 'Actions')");
      expect(html).toContain('<option value="upload">Upload only</option>');
      expect(html).toContain('<option value="two-way">Read and upload</option>');
      expect(html).toContain('function scrubDashboardUrl()');
      expect(html).toContain("history.replaceState(null, '', location.pathname)");
      expect(html).toContain("var DASHBOARD_API_PREFIX = '/dashboard/api/'");
      expect(html).toContain("var APP_API_PREFIX = '/api/'");
      expect(html).toContain('function dashboardRequestUrl(url)');
      expect(html).toContain("return dashboardBasePath() + '/api/' + url.slice(DASHBOARD_API_PREFIX.length)");
      expect(html).toContain('function appBasePath()');
      expect(html).toContain('return appBasePath() + url;');
      expect(dashboardRequestUrlFromHtml(html, '/dashboard', '/api/fs/list?path=%2F')).toBe('/api/fs/list?path=%2F');
      expect(dashboardRequestUrlFromHtml(html, '/t/demo/dashboard', '/api/fs/list?path=%2F')).toBe('/t/demo/api/fs/list?path=%2F');
      expect(dashboardRequestUrlFromHtml(html, '/t/demo/dashboard', '/dashboard/api/leases')).toBe('/t/demo/dashboard/api/leases');
      expect(html).toContain("api('/dashboard/api/local-path-picker'");
      expect(html).toContain('Choose a local file or folder first.');
      expect(html).toContain('Remove source');
      expect(html).toContain('Rename source');
      expect(html).toContain('Replace link');
      expect(html).toContain('You can copy this link again from the dashboard');
      expect(html).toContain("'/dashboard/api/leases/' + encodeURIComponent(id) + '/publish'");
      expect(html).toContain("publishBtn.textContent = isPublished ? 'Unpublish' : 'Publish'");
      expect(html).toContain('id="mcp-modal"');
      expect(html).toContain('data-view="mcp"');
      expect(html).toContain('id="view-mcp-panel"');
      expect(html).toContain('id="new-grant"');
      expect(html).toContain("api('/dashboard/api/grants'");
      expect(html).toContain('function submitMcpGrant()');
      expect(html).toContain('function renderGrantRow(grant)');
      // The old context-menu MCP entry point is gone — replaced by the tab.
      expect(html).not.toContain('id="context-mcp"');
      expect(html).not.toContain('id="add-mount-settings"');
      expect(html).not.toContain('id="mounts-wrap"');
      expect(html).not.toContain('...(opts.headers || {})');
      expect(html).not.toContain('60_000');

      // The dashboard's <script> body is built inside a backtick template
      // literal, so any backslash that isn't doubled gets eaten before it
      // reaches the browser. We previously shipped `replace(/\/+$/, '')`,
      // which rendered as `replace(//+$/, '')` — a parse error that
      // killed the entire inline script, broke the JS submit interceptor,
      // and forced the login form to submit natively without a working
      // bootstrap. Parse the served script to catch this whole class of
      // bug at the test boundary instead of in someone's browser.
      const scripts = inlineScriptBodies(html);
      expect(scripts.length).toBeGreaterThan(0);
      for (const body of scripts) {
        expect(() => new Function(body)).not.toThrow();
      }
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('redirects dashboard URLs with query params before serving the login UI', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-query-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/dashboard?username=pnee&password=secret`, {
        redirect: 'manual',
      });
      expect(response.status).toBe(303);
      expect(response.headers.get('location')).toBe('dashboard');
      expect(await response.text()).toBe('');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('accepts dashboard login posts forwarded by relay transport', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-relay-login-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'sarah', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const blocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Origin: 'https://mvmt-relay.fly.dev' },
        body: JSON.stringify({ username: 'sarah', password: 'correct horse battery staple' }),
      });
      expect(blocked.status).toBe(403);

      const allowed = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Origin: 'https://mvmt-relay.fly.dev',
          'X-MVMT-Transport': 'relay',
        },
        body: JSON.stringify({ username: 'sarah', password: 'correct horse battery staple' }),
      });
      expect(allowed.status).toBe(200);
      expect(allowed.headers.get('set-cookie')).toContain('mvmt_dashboard=');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('supports safe form POST fallback for dashboard login', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-form-login-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'sarah', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/dashboard`, {
        method: 'POST',
        redirect: 'manual',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          Origin: `http://127.0.0.1:${server.port}`,
        },
        body: new URLSearchParams({ username: 'sarah', password: 'correct horse battery staple' }),
      });
      expect(response.status).toBe(303);
      expect(response.headers.get('location')).toBe('dashboard');
      expect(response.headers.get('set-cookie')).toContain('mvmt_dashboard=');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  // The mvmt agent is reachable both locally over HTTP and via the relay
  // over HTTPS. Marking the session cookie Secure unconditionally (e.g.
  // because the configured public URL is HTTPS) silently drops it on the
  // local HTTP browser, which makes login look broken: the POST succeeds
  // but every follow-up request comes in unauthenticated.
  it('sets Secure on the session cookie based on the request Origin, not the configured public URL', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-secure-cookie-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'sarah', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
      resolvePublicBaseUrl: () => 'https://mvmt-relay.fly.dev/t/demo',
    });
    try {
      const local = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Origin: `http://127.0.0.1:${server.port}`,
        },
        body: JSON.stringify({ username: 'sarah', password: 'correct horse battery staple' }),
      });
      expect(local.status).toBe(200);
      const localCookie = local.headers.get('set-cookie') ?? '';
      expect(localCookie).toContain('mvmt_dashboard=');
      expect(localCookie).not.toContain('Secure');

      const relayed = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Origin: 'https://mvmt-relay.fly.dev',
          'X-MVMT-Transport': 'relay',
        },
        body: JSON.stringify({ username: 'sarah', password: 'correct horse battery staple' }),
      });
      expect(relayed.status).toBe(200);
      expect(relayed.headers.get('set-cookie')).toContain('Secure');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects existing dashboard sessions after the user is removed', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-remove-user-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'sarah', password: 'correct horse battery staple' });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'sarah', 'correct horse battery staple');
      const before = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/me`, {
        headers: { Cookie: cookie },
      });
      expect(before.status).toBe(200);

      removePrivilegedUser(usersPath, 'sarah');

      const after = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/me`, {
        headers: { Cookie: cookie },
      });
      expect(after.status).toBe(401);
      expect(after.headers.get('set-cookie')).toContain('Max-Age=0');

      const loginAgain = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'sarah', password: 'correct horse battery staple' }),
      });
      expect(loginAgain.status).toBe(401);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  // The dashboard MCP-access tab mints clients[] API-token grants. A
  // grant is either "all mounts" (one /** scope that tracks the mount
  // list) or a per-mount selection. Creation is admin + config-path
  // gated; the resulting grant honours the publish exposure boundary.
  function mvmtGrantFixture(): { tmp: string; configPath: string; usersPath: string; tokenPath: string; leaseStorePath: string } {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-grant-test-'));
    const configPath = path.join(tmp, 'config.yaml');
    const usersPath = path.join(tmp, '.privileged-users.json');
    const workspace = path.join(tmp, 'workspace');
    const readonly = path.join(tmp, 'readonly');
    fs.mkdirSync(workspace, { recursive: true });
    fs.mkdirSync(readonly, { recursive: true });
    fs.writeFileSync(path.join(workspace, 'note.md'), 'hello');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });
    createPrivilegedUser(usersPath, { username: 'member', password: 'correct horse battery staple' });
    return {
      tmp,
      configPath,
      usersPath,
      tokenPath: path.join(tmp, '.session-token'),
      leaseStorePath: path.join(tmp, '.leases.json'),
    };
  }

  it('mints all-mounts and per-mount MCP grants from the dashboard and lists them', async () => {
    const fx = mvmtGrantFixture();
    await saveConfig(fx.configPath, parseConfig({
      version: 1,
      mounts: [
        { name: 'workspace', type: 'local_folder', path: '/workspace', root: path.join(fx.tmp, 'workspace'), writeAccess: true },
        { name: 'readonly', type: 'local_folder', path: '/readonly', root: path.join(fx.tmp, 'readonly') },
      ],
    }));
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath: fx.tokenPath,
      configPath: fx.configPath,
      leaseMounts: () => readConfig(fx.configPath).mounts,
      clients: () => readConfig(fx.configPath).clients,
      leaseStorePath: fx.leaseStorePath,
      privilegedUsersPath: fx.usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const post = (body: unknown) => fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      // All-mounts grant → one /** scope that tracks the mount list.
      const all = await post({ label: 'Claude - everything', expires: 'never', published: true, allMounts: true });
      expect(all.status).toBe(201);
      const allBody = await all.json() as { grant: { id: string }; token: string; endpoint: string };
      expect(allBody.token).toMatch(/^mvmt_t_/);
      expect(allBody.endpoint).toMatch(/\/mcp$/);
      const allClient = readConfig(fx.configPath).clients?.find((c) => c.id === allBody.grant.id);
      expect(allClient?.permissions).toEqual([{ path: '/**', actions: ['search', 'read', 'write'] }]);
      expect(allClient?.published).toBe(true);

      // Per-mount grant → one resolved scope entry per selected mount.
      const custom = await post({
        label: 'Claude - workspace only',
        published: false,
        scopes: [{ path: '/workspace', mode: 'write' }, { path: '/readonly', mode: 'read' }],
      });
      expect(custom.status).toBe(201);
      const customBody = await custom.json() as { grant: { id: string } };
      const customClient = readConfig(fx.configPath).clients?.find((c) => c.id === customBody.grant.id);
      expect(customClient?.permissions).toEqual([
        { path: '/workspace/**', actions: ['search', 'read', 'write'] },
        { path: '/readonly/**', actions: ['search', 'read'] },
      ]);
      expect(customClient?.published).toBe(false);

      // GET lists both, with scope summaries and reach.
      const list = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants`, { headers: { Cookie: cookie } });
      expect(list.status).toBe(200);
      const listBody = await list.json() as { canManage: boolean; grants: { id: string; scope: string; published: boolean }[] };
      expect(listBody.canManage).toBe(true);
      const allRow = listBody.grants.find((g) => g.id === allBody.grant.id);
      const customRow = listBody.grants.find((g) => g.id === customBody.grant.id);
      expect(allRow).toMatchObject({ scope: 'All mounts', published: true });
      expect(customRow).toMatchObject({ scope: '2 mounts', published: false });

      const remoteList = await requestWithHost(server.port, '/dashboard/api/grants', 'public.example.test', {
        headers: { Cookie: cookie },
      });
      expect(remoteList.status).toBe(200);
      expect((remoteList.json() as { canManage: boolean }).canManage).toBe(false);
    } finally {
      await server.close();
      fs.rmSync(fx.tmp, { recursive: true, force: true });
    }
  });

  it('rejects invalid MCP grant requests and never writes a partial grant', async () => {
    const fx = mvmtGrantFixture();
    await saveConfig(fx.configPath, parseConfig({
      version: 1,
      mounts: [{ name: 'readonly', type: 'local_folder', path: '/readonly', root: path.join(fx.tmp, 'readonly') }],
    }));
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath: fx.tokenPath,
      configPath: fx.configPath,
      leaseMounts: () => readConfig(fx.configPath).mounts,
      clients: () => readConfig(fx.configPath).clients,
      leaseStorePath: fx.leaseStorePath,
      privilegedUsersPath: fx.usersPath,
    });
    try {
      const adminCookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const post = (cookie: string, body: unknown) => fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      const noLabel = await post(adminCookie, { allMounts: true });
      expect(noLabel.status).toBe(400);
      expect((await noLabel.json() as { error: string }).error).toBe('label_required');

      const noScope = await post(adminCookie, { label: 'x' });
      expect(noScope.status).toBe(400);
      expect((await noScope.json() as { error: string }).error).toBe('scope_required');

      const outside = await post(adminCookie, { label: 'x', scopes: [{ path: '/not-a-mount', mode: 'read' }] });
      expect(outside.status).toBe(400);

      const writeReadOnly = await post(adminCookie, { label: 'x', scopes: [{ path: '/readonly', mode: 'write' }] });
      expect(writeReadOnly.status).toBe(400);
      expect((await writeReadOnly.json() as { error: string }).error).toBe('mount_read_only');

      const memberCookie = await loginDashboard(server.port, 'member', 'correct horse battery staple');
      const nonAdmin = await post(memberCookie, { label: 'x', allMounts: true });
      expect(nonAdmin.status).toBe(403);

      // A non-admin can still see the (empty) list — no silent dead-end.
      const memberList = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants`, { headers: { Cookie: memberCookie } });
      expect(memberList.status).toBe(200);
      expect((await memberList.json() as { canManage: boolean }).canManage).toBe(false);

      expect(readConfig(fx.configPath).clients ?? []).toEqual([]);
    } finally {
      await server.close();
      fs.rmSync(fx.tmp, { recursive: true, force: true });
    }
  });

  it('publishes, unpublishes, and revokes an MCP grant from the dashboard', async () => {
    const fx = mvmtGrantFixture();
    await saveConfig(fx.configPath, parseConfig({
      version: 1,
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: path.join(fx.tmp, 'workspace'), writeAccess: true }],
    }));
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath: fx.tokenPath,
      configPath: fx.configPath,
      leaseMounts: () => readConfig(fx.configPath).mounts,
      clients: () => readConfig(fx.configPath).clients,
      leaseStorePath: fx.leaseStorePath,
      privilegedUsersPath: fx.usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const created = await (await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ label: 'Claude', published: false, allMounts: true }),
      })).json() as { grant: { id: string }; token: string };
      const id = created.grant.id;

      const publish = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants/${id}/publish`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ published: true }),
      });
      expect(publish.status).toBe(200);
      expect(readConfig(fx.configPath).clients?.[0]?.published).toBe(true);

      const unpublish = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants/${id}/publish`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ published: false }),
      });
      expect(unpublish.status).toBe(200);
      expect(readConfig(fx.configPath).clients?.[0]?.published).toBe(false);

      const malformedPublish = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants/${id}/publish`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({}),
      });
      expect(malformedPublish.status).toBe(400);
      expect((await malformedPublish.json() as { error: string }).error).toBe('published_required');
      expect(readConfig(fx.configPath).clients?.[0]?.published).toBe(false);

      const revoke = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants/${id}`, {
        method: 'DELETE',
        headers: { Cookie: cookie },
      });
      expect(revoke.status).toBe(200);
      expect(readConfig(fx.configPath).clients ?? []).toEqual([]);

      const revokeMissing = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants/grant-missing`, {
        method: 'DELETE',
        headers: { Cookie: cookie },
      });
      expect(revokeMissing.status).toBe(404);
    } finally {
      await server.close();
      fs.rmSync(fx.tmp, { recursive: true, force: true });
    }
  });

  it('refuses dashboard MCP grant creation when no configPath is wired', async () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-dashboard-grant-noconfig-test-'));
    const tokenPath = path.join(tmp, '.session-token');
    const usersPath = path.join(tmp, '.privileged-users.json');
    createPrivilegedUser(usersPath, { username: 'admin', password: 'correct horse battery staple', admin: true });
    const server = await startHttpServer(new ToolRouter(), {
      port: 0,
      tokenPath,
      leaseMounts: [],
      privilegedUsersPath: usersPath,
    });
    try {
      const cookie = await loginDashboard(server.port, 'admin', 'correct horse battery staple');
      const blocked = await fetch(`http://127.0.0.1:${server.port}/dashboard/api/grants`, {
        method: 'POST',
        headers: { Cookie: cookie, 'Content-Type': 'application/json' },
        body: JSON.stringify({ label: 'x', allMounts: true }),
      });
      expect(blocked.status).toBe(403);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });
});

describe('startHttpServer lifecycle', () => {
  it('sends agent instructions during MCP initialization', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const response = await fetch(`http://127.0.0.1:${server.port}/mcp`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${sessionToken}`,
          Accept: 'application/json, text/event-stream',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 1,
          method: 'initialize',
          params: {
            protocolVersion: '2025-03-26',
            capabilities: {},
            clientInfo: { name: 'mvmt-instructions-test', version: '0.0.0' },
          },
        }),
      });

      expect(response.status).toBe(200);
      const body = parseMcpResponse(await response.text());
      expect(body.result.instructions).toBe(MVMT_SERVER_INSTRUCTIONS);
      expect(body.result.instructions).toContain('For content questions, call search first');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('serves OAuth authorization metadata for the root and MCP resource path', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      for (const pathSuffix of [
        '/.well-known/oauth-authorization-server',
        '/.well-known/oauth-authorization-server/mcp',
      ]) {
        const response = await fetch(`http://127.0.0.1:${server.port}${pathSuffix}`);
        expect(response.status).toBe(200);
        const metadata = await response.json();
        expect(metadata.registration_endpoint).toBe(`http://127.0.0.1:${server.port}/register`);
        expect(metadata.authorization_endpoint).toBe(`http://127.0.0.1:${server.port}/authorize`);
        expect(metadata.token_endpoint).toBe(`http://127.0.0.1:${server.port}/token`);
        expect(metadata.authorization_response_iss_parameter_supported).toBeUndefined();
        expect(metadata.grant_types_supported).toEqual(['authorization_code', 'refresh_token']);
        expect(metadata.scopes_supported).toEqual(['mcp', 'offline_access']);
      }
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('serves protected resource metadata with the scopes ChatGPT requests', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      for (const pathSuffix of [
        '/.well-known/oauth-protected-resource',
        '/.well-known/oauth-protected-resource/mcp',
      ]) {
        const response = await fetch(`http://127.0.0.1:${server.port}${pathSuffix}`);
        expect(response.status).toBe(200);
        const metadata = await response.json();
        expect(metadata.resource).toBe(`http://127.0.0.1:${server.port}/mcp`);
        expect(metadata.authorization_servers).toEqual([`http://127.0.0.1:${server.port}`]);
        expect(metadata.scopes_supported).toEqual(['mcp', 'offline_access']);
      }
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('emits sanitized request logs for OAuth discovery, registration, and auth failures', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; path: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const resource = `http://127.0.0.1:${server.port}/mcp`;
      await fetch(`http://127.0.0.1:${server.port}/.well-known/oauth-authorization-server/mcp`);
      await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${sessionToken}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: 'test-client',
          redirect_uris: ['https://client.example/callback'],
        }),
      });
      await fetch(
        `http://127.0.0.1:${server.port}/authorize?response_type=code&client_id=test-client&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&resource=${encodeURIComponent(resource)}&code_challenge=secret-challenge&code_challenge_method=S256`,
      );
      await fetch(`http://127.0.0.1:${server.port}/mcp`);

      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'oauth.discovery',
            path: '/.well-known/oauth-authorization-server/mcp',
            status: 200,
          }),
          expect.objectContaining({
            kind: 'oauth.register',
            path: '/register',
            status: 201,
            clientId: 'test-client',
          }),
          expect.objectContaining({
            kind: 'oauth.authorize',
            path: '/authorize',
            status: 200,
            clientId: 'test-client',
          }),
          expect.objectContaining({
            kind: 'mcp.auth',
            path: '/mcp',
            status: 401,
            detail: 'missing_bearer',
          }),
        ]),
      );
      expect(JSON.stringify(requestLogs)).not.toContain('secret-challenge');
      expect(JSON.stringify(requestLogs)).not.toContain('client.example/callback');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('logs invalid bearer tokens distinctly from missing ones', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; path: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/mcp`, {
        method: 'POST',
        headers: { Authorization: 'Bearer nope', 'Content-Type': 'application/json' },
        body: JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'tools/list', params: {} }),
      });
      expect(response.status).toBe(401);
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'mcp.auth',
            path: '/mcp',
            status: 401,
            detail: 'invalid_bearer',
          }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('logs the authenticated MCP client for successful requests', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; path: string; status: number; clientId?: string; ip?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: [
        {
          id: 'claude',
          name: 'Claude Desktop',
          auth: { type: 'token', tokenHash: hashApiToken('claude-token') },
          rawToolsEnabled: true,
          permissions: [],
        },
      ],
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      await initializeMcpSession(server.port, 'claude-token');

      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'mcp.request',
            path: '/mcp',
            status: 200,
            clientId: 'claude',
            ip: '127.0.0.1',
          }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('logs relay-forwarded MCP requests with the original remote address when provided', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; path: string; status: number; clientId?: string; ip?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: [
        {
          id: 'claude',
          name: 'Claude Desktop',
          auth: { type: 'token', tokenHash: hashApiToken('claude-token') },
          rawToolsEnabled: true,
          permissions: [],
          published: true,
        },
      ],
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/mcp`, {
        method: 'POST',
        headers: {
          Authorization: 'Bearer claude-token',
          Accept: 'application/json, text/event-stream',
          'Content-Type': 'application/json',
          'X-MVMT-Transport': 'relay',
          'Fly-Client-IP': '203.0.113.10',
          'X-Forwarded-For': '198.51.100.20, 10.0.0.5',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 1,
          method: 'initialize',
          params: {
            protocolVersion: '2025-03-26',
            capabilities: {},
            clientInfo: { name: 'mvmt-relay-log-test', version: '0.0.0' },
          },
        }),
      });
      expect(response.status).toBe(200);
      await response.text();

      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'mcp.request',
            path: '/mcp',
            status: 200,
            clientId: 'claude',
            ip: '203.0.113.10',
          }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('blocks cross-origin browser requests to OAuth endpoints', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; path: string; status: number; detail?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const origin = 'https://evil.example.com';
      const responses = await Promise.all([
        fetch(`http://127.0.0.1:${server.port}/register`, {
          method: 'POST',
          headers: { Origin: origin, 'Content-Type': 'application/json' },
          body: JSON.stringify({ redirect_uris: ['https://client.example/callback'] }),
        }),
        fetch(`http://127.0.0.1:${server.port}/authorize`, {
          headers: { Origin: origin },
        }),
        fetch(`http://127.0.0.1:${server.port}/authorize`, {
          method: 'POST',
          headers: { Origin: origin, 'Content-Type': 'application/x-www-form-urlencoded' },
          body: new URLSearchParams({}),
        }),
        fetch(`http://127.0.0.1:${server.port}/token`, {
          method: 'POST',
          headers: { Origin: origin, 'Content-Type': 'application/x-www-form-urlencoded' },
          body: new URLSearchParams({ grant_type: 'authorization_code' }),
        }),
      ]);

      expect(responses.map((response) => response.status)).toEqual([403, 403, 403, 403]);
      expect(requestLogs.filter((entry) => entry.kind === 'oauth.origin')).toHaveLength(4);
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({ path: '/register', status: 403, detail: 'origin_not_allowed' }),
          expect.objectContaining({ path: '/authorize', status: 403, detail: 'origin_not_allowed' }),
          expect.objectContaining({ path: '/token', status: 403, detail: 'origin_not_allowed' }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('allows OAuth browser requests from the public tunnel origin', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      resolvePublicBaseUrl: () => 'https://mvmt.example.com',
    });

    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { Origin: 'https://mvmt.example.com', 'Content-Type': 'application/json' },
        body: JSON.stringify({ redirect_uris: ['https://client.example/callback'] }),
      });
      expect(response.status).toBe(201);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('exchanges an authorization code when the OAuth resource parameter is echoed', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const resource = `http://127.0.0.1:${server.port}/mcp`;
      const redirectUri = 'https://chatgpt.com/connector/oauth/test-callback';
      const { verifier, challenge } = s256Pair();
      await registerClient(server.port, 'chatgpt', [redirectUri], sessionToken);
      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'chatgpt',
          redirect_uri: redirectUri,
          resource,
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });
      expect(authorize.status).toBe(302);
      const location = authorize.headers.get('location');
      expect(location).toBeTruthy();
      const redirect = new URL(location!);
      const code = redirect.searchParams.get('code');
      expect(code).toBeTruthy();
      expect(redirect.searchParams.get('iss')).toBeNull();

      const token = await fetch(`http://127.0.0.1:${server.port}/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          code: code!,
          client_id: 'chatgpt',
          redirect_uri: redirectUri,
          resource,
          code_verifier: verifier,
        }),
      });
      expect(token.status).toBe(200);
      const body = await token.json();
      expect(body.token_type).toBe('Bearer');
      expect(body.access_token).toBeTypeOf('string');
      expect(body.refresh_token).toBeTypeOf('string');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('exchanges an authorization code when the token request omits a resource already bound at authorize time', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const resource = `http://127.0.0.1:${server.port}/mcp`;
      const redirectUri = 'https://claude.ai/api/mcp/auth_callback';
      const { verifier, challenge } = s256Pair();
      await registerClient(server.port, 'claude', [redirectUri], sessionToken);
      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'claude',
          redirect_uri: redirectUri,
          resource,
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });
      expect(authorize.status).toBe(302);
      const code = new URL(authorize.headers.get('location')!).searchParams.get('code');
      expect(code).toBeTruthy();

      const token = await fetch(`http://127.0.0.1:${server.port}/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          code: code!,
          client_id: 'claude',
          redirect_uri: redirectUri,
          code_verifier: verifier,
        }),
      });
      expect(token.status).toBe(200);
      const body = await token.json();
      expect(body.token_type).toBe('Bearer');
      expect(body.access_token).toBeTypeOf('string');
      expect(body.refresh_token).toBeTypeOf('string');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('exchanges a refresh token for a new access token', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const resource = `http://127.0.0.1:${server.port}/mcp`;
      const redirectUri = 'https://chatgpt.com/connector/oauth/test-callback';
      const { verifier, challenge } = s256Pair();
      await registerClient(server.port, 'chatgpt', [redirectUri], sessionToken);
      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'chatgpt',
          redirect_uri: redirectUri,
          resource,
          scope: 'mcp offline_access',
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });
      const code = new URL(authorize.headers.get('location')!).searchParams.get('code');
      expect(code).toBeTruthy();

      const token = await fetch(`http://127.0.0.1:${server.port}/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          code: code!,
          client_id: 'chatgpt',
          redirect_uri: redirectUri,
          resource,
          code_verifier: verifier,
        }),
      });
      expect(token.status).toBe(200);
      const firstGrant = await token.json();
      expect(firstGrant.refresh_token).toBeTypeOf('string');

      const refresh = await fetch(`http://127.0.0.1:${server.port}/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'refresh_token',
          client_id: 'chatgpt',
          refresh_token: firstGrant.refresh_token,
        }),
      });
      expect(refresh.status).toBe(200);
      const refreshedGrant = await refresh.json();
      expect(refreshedGrant.access_token).toBeTypeOf('string');
      expect(refreshedGrant.refresh_token).toBeTypeOf('string');
      expect(refreshedGrant.access_token).not.toBe(firstGrant.access_token);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('omits the optional issuer parameter from the authorization redirect', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const resource = `http://127.0.0.1:${server.port}/mcp`;
      const redirectUri = 'https://claude.ai/api/mcp/auth_callback';
      const { challenge } = s256Pair();
      await registerClient(server.port, 'claude', [redirectUri], sessionToken);
      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'claude',
          redirect_uri: redirectUri,
          resource,
          code_challenge: challenge,
          code_challenge_method: 'S256',
          state: 'test-state',
          session_token: sessionToken,
        }),
      });
      expect(authorize.status).toBe(302);
      const redirect = new URL(authorize.headers.get('location')!);
      expect(redirect.searchParams.get('code')).toBeTruthy();
      expect(redirect.searchParams.get('state')).toBe('test-state');
      expect(redirect.searchParams.has('iss')).toBe(false);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects /authorize for unregistered redirect_uri', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const { challenge } = s256Pair();
      await registerClient(server.port, 'claude', ['https://claude.ai/registered/cb'], sessionToken);

      const response = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'claude',
          redirect_uri: 'https://attacker.example/cb',
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });
      expect(response.status).toBe(400);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('defaults a missing GET /authorize resource to the canonical MCP resource', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const signingKeyPath = path.join(tmp, '.mvmt', '.signing-key');
    const requestLogs: Array<{ kind: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      signingKeyPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const redirectUri = 'https://claude.ai/api/mcp/auth_callback';
      const { verifier, challenge } = s256Pair();
      const resource = `http://127.0.0.1:${server.port}/mcp`;
      await registerClient(server.port, 'claude', [redirectUri], sessionToken);

      const promptUrl = new URL(`http://127.0.0.1:${server.port}/authorize`);
      promptUrl.search = new URLSearchParams({
        response_type: 'code',
        client_id: 'claude',
        redirect_uri: redirectUri,
        state: 'resource-state',
        code_challenge: challenge,
        code_challenge_method: 'S256',
      }).toString();
      const prompt = await fetch(promptUrl);
      expect(prompt.status).toBe(200);
      const promptBody = await prompt.text();
      expect(promptBody).toContain(`name="resource" value="${resource}"`);

      const approve = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'claude',
          redirect_uri: redirectUri,
          resource,
          state: 'resource-state',
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });
      expect(approve.status).toBe(302);
      const approveRedirect = new URL(approve.headers.get('location')!);
      const code = approveRedirect.searchParams.get('code');
      expect(code).toBeTruthy();

      const accessToken = await exchangeAuthorizationCodeForToken({
        port: server.port,
        clientId: 'claude',
        redirectUri,
        code: code!,
        verifier,
      });
      expectAccessTokenAudience(accessToken, signingKeyPath, resource);
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'oauth.authorize',
            status: 200,
            detail: expect.stringContaining('resource_defaulted=true'),
          }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('defaults a missing POST /authorize resource to the canonical MCP resource', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const signingKeyPath = path.join(tmp, '.mvmt', '.signing-key');
    const requestLogs: Array<{ kind: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      signingKeyPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const redirectUri = 'https://claude.ai/api/mcp/auth_callback';
      const { verifier, challenge } = s256Pair();
      const resource = `http://127.0.0.1:${server.port}/mcp`;
      await registerClient(server.port, 'claude', [redirectUri], sessionToken);

      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'claude',
          redirect_uri: redirectUri,
          state: 'resource-state',
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });
      expect(authorize.status).toBe(302);
      const authorizeRedirect = new URL(authorize.headers.get('location')!);
      const code = authorizeRedirect.searchParams.get('code');
      expect(code).toBeTruthy();

      const accessToken = await exchangeAuthorizationCodeForToken({
        port: server.port,
        clientId: 'claude',
        redirectUri,
        code: code!,
        verifier,
      });
      expectAccessTokenAudience(accessToken, signingKeyPath, resource);
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'oauth.authorize',
            status: 302,
            detail: expect.stringContaining('resource_defaulted=true'),
          }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('redirects explicit /authorize resource mismatches to the registered redirect_uri', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const requestLogs: Array<{ kind: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const redirectUri = 'https://claude.ai/api/mcp/auth_callback';
      const { challenge } = s256Pair();
      await registerClient(server.port, 'claude', [redirectUri], sessionToken);

      const wrongUrl = new URL(`http://127.0.0.1:${server.port}/authorize`);
      wrongUrl.search = new URLSearchParams({
        response_type: 'code',
        client_id: 'claude',
        redirect_uri: redirectUri,
        resource: 'https://other.example.com/mcp',
        state: 'get-state',
        code_challenge: challenge,
        code_challenge_method: 'S256',
      }).toString();
      const wrong = await fetch(wrongUrl, { redirect: 'manual' });
      expect(wrong.status).toBe(302);
      const wrongRedirect = new URL(wrong.headers.get('location')!);
      expect(`${wrongRedirect.origin}${wrongRedirect.pathname}`).toBe(redirectUri);
      expect(wrongRedirect.searchParams.get('error')).toBe('invalid_target');
      expect(wrongRedirect.searchParams.get('error_description')).toBe('Invalid resource');
      expect(wrongRedirect.searchParams.get('state')).toBe('get-state');
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({ kind: 'oauth.authorize', status: 302, detail: 'invalid_resource' }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('accepts resource URLs with normalized host casing and trailing slash', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      resolvePublicBaseUrl: () => 'https://mvmt.example.com',
    });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const redirectUri = 'https://claude.ai/api/mcp/auth_callback';
      const { challenge } = s256Pair();
      await registerClient(server.port, 'claude', [redirectUri], sessionToken);

      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'claude',
          redirect_uri: redirectUri,
          resource: 'https://MVMT.EXAMPLE.COM/mcp/',
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });

      expect(authorize.status).toBe(302);
      const redirect = new URL(authorize.headers.get('location')!);
      expect(redirect.searchParams.get('code')).toEqual(expect.any(String));
      expect(redirect.searchParams.has('error')).toBe(false);

      for (const resource of ['https://user@mvmt.example.com/mcp', 'https://mvmt.example.com/mcp#fragment']) {
        const rejected = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
          method: 'POST',
          redirect: 'manual',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: new URLSearchParams({
            response_type: 'code',
            client_id: 'claude',
            redirect_uri: redirectUri,
            resource,
            code_challenge: challenge,
            code_challenge_method: 'S256',
            session_token: sessionToken,
          }),
        });
        expect(rejected.status).toBe(302);
        const rejectedRedirect = new URL(rejected.headers.get('location')!);
        expect(rejectedRedirect.searchParams.get('error')).toBe('invalid_target');
        expect(rejectedRedirect.searchParams.get('error_description')).toBe('Invalid resource');
      }
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects /register when redirect_uris are missing', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ client_name: 'no uris' }),
      });
      expect(response.status).toBe(400);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('guards caller-supplied OAuth client_id registration', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const named = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: 'chatgpt',
          redirect_uris: ['https://chatgpt.com/connector/oauth/test-callback'],
        }),
      });
      expect(named.status).toBe(401);

      const generated = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          redirect_uris: ['https://chatgpt.com/connector/oauth/test-callback'],
        }),
      });
      expect(generated.status).toBe(201);
      const generatedBody = await generated.json();
      expect(generatedBody.client_id).toEqual(expect.stringMatching(/^mvmt-[0-9a-f-]{36}$/));

      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      await registerClient(server.port, 'chatgpt', ['https://chatgpt.com/connector/oauth/test-callback'], sessionToken);

      const duplicate = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${sessionToken}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: 'chatgpt',
          redirect_uris: ['https://attacker.example/callback'],
        }),
      });
      expect(duplicate.status).toBe(409);
      const duplicateBody = await duplicate.json();
      expect(duplicateBody.error).toBe('invalid_client_metadata');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('limits OAuth dynamic client registration fan-out', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          redirect_uris: Array.from({ length: 11 }, (_, index) => `https://client.example/cb/${index}`),
        }),
      });
      expect(response.status).toBe(429);
      const body = await response.json();
      expect(body.error).toBe('invalid_client_metadata');
      expect(body.error_description).toContain('redirect_uris exceeds');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('returns 500 from /register when the client registry cannot be persisted', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      fs.chmodSync(path.dirname(tokenPath), 0o500);
      const response = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${sessionToken}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: 'persist-fail-client',
          redirect_uris: ['https://persist-fail.example/cb'],
        }),
      });
      expect(response.status).toBe(500);
      const body = await response.json();
      expect(body.error).toBe('server_error');
    } finally {
      fs.chmodSync(path.dirname(tokenPath), 0o700);
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('returns an RFC 7591-style client information response from /register', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const response = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${sessionToken}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: 'chatgpt',
          client_name: 'ChatGPT Connector',
          scope: 'mcp',
          redirect_uris: ['https://chatgpt.com/connector/oauth/test-callback'],
        }),
      });
      expect(response.status).toBe(201);
      const body = await response.json();
      expect(body).toMatchObject({
        client_id: 'chatgpt',
        client_name: 'ChatGPT Connector',
        scope: 'mcp',
        token_endpoint_auth_method: 'none',
        grant_types: ['authorization_code', 'refresh_token'],
        response_types: ['code'],
        redirect_uris: ['https://chatgpt.com/connector/oauth/test-callback'],
      });
      expect(body.client_id_issued_at).toEqual(expect.any(Number));
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('uses resolvePublicBaseUrl and ignores X-Forwarded-Host in OAuth metadata', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      resolvePublicBaseUrl: () => 'https://mvmt.example.com',
    });

    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/.well-known/oauth-authorization-server`, {
        headers: { 'X-Forwarded-Host': 'attacker.example.com', 'X-Forwarded-Proto': 'https' },
      });
      const metadata = await response.json();
      expect(metadata.issuer).toBe('https://mvmt.example.com');
      expect(metadata.authorization_endpoint).toBe('https://mvmt.example.com/authorize');
      expect(JSON.stringify(metadata)).not.toContain('attacker');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  // When mvmt is reached through a relay path prefix (e.g. /t/demo), the
  // OAuth discovery chain must keep that prefix end to end. getBaseUrl()
  // strips the path, which would point WWW-Authenticate, the metadata
  // docs, and the resource/issuer at the relay root — where they 404 and
  // the client (Claude) reports "couldn't reach the MCP server".
  it('keeps the relay workspace prefix across the whole OAuth discovery chain', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-oauth-prefix-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      resolvePublicBaseUrl: () => 'https://mvmt-relay.fly.dev/t/demo',
    });
    const base = 'https://mvmt-relay.fly.dev/t/demo';
    try {
      // Step 1: the unauthenticated /mcp 401 points at the prefixed metadata.
      const unauth = await fetch(`http://127.0.0.1:${server.port}/mcp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: '{}',
      });
      expect(unauth.status).toBe(401);
      expect(unauth.headers.get('www-authenticate'))
        .toContain(`resource_metadata="${base}/.well-known/oauth-protected-resource"`);

      // Step 2: protected-resource metadata advertises the prefixed resource.
      const resourceMeta = await (await fetch(
        `http://127.0.0.1:${server.port}/.well-known/oauth-protected-resource`,
      )).json() as { resource: string; authorization_servers: string[] };
      expect(resourceMeta.resource).toBe(`${base}/mcp`);
      expect(resourceMeta.authorization_servers).toEqual([base]);

      // Step 3: authorization-server metadata advertises prefixed endpoints.
      const authMeta = await (await fetch(
        `http://127.0.0.1:${server.port}/.well-known/oauth-authorization-server`,
      )).json() as Record<string, string>;
      expect(authMeta.issuer).toBe(base);
      expect(authMeta.authorization_endpoint).toBe(`${base}/authorize`);
      expect(authMeta.token_endpoint).toBe(`${base}/token`);
      expect(authMeta.registration_endpoint).toBe(`${base}/register`);

      // Step 4: the approval page form must POST relatively. An absolute
      // action="/authorize" would submit to the relay root and 404.
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${sessionToken}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({ client_id: 'test-client', redirect_uris: ['https://client.example/callback'] }),
      });
      const approvalPage = await (await fetch(
        `http://127.0.0.1:${server.port}/authorize?response_type=code&client_id=test-client`
        + `&redirect_uri=https%3A%2F%2Fclient.example%2Fcallback&resource=${encodeURIComponent(`${base}/mcp`)}`
        + `&code_challenge=secret-challenge&code_challenge_method=S256`,
      )).text();
      expect(approvalPage).toContain('<form method="POST" action="authorize">');
      expect(approvalPage).not.toContain('action="/authorize"');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('returns a close handle that releases the listening port', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });
    const port = server.port;

    try {
      const token = fs.readFileSync(tokenPath, 'utf-8').trim();
      const response = await fetch(`http://127.0.0.1:${port}/health`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      expect(response.status).toBe(200);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }

    await expect(canListenOn(port)).resolves.toBe(true);
    await expect(server.close()).resolves.toBeUndefined();
  });

  it('reuses an existing session token across server restarts', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const originalToken = generateSessionToken(tokenPath);
    const originalStat = fs.statSync(tokenPath);
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const tokenAfterStart = fs.readFileSync(tokenPath, 'utf-8').trim();
      const statAfterStart = fs.statSync(tokenPath);
      expect(tokenAfterStart).toBe(originalToken);
      expect(statAfterStart.mtimeMs).toBe(originalStat.mtimeMs);

      const response = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${originalToken}` },
      });
      expect(response.status).toBe(200);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('keeps OAuth access tokens valid across restarts when the advertised resource is unchanged', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const resourceBaseUrl = 'https://mvmt.example.com';
    const firstServer = await startHttpServer(router, {
      port: 0,
      tokenPath,
      resolvePublicBaseUrl: () => resourceBaseUrl,
    });
    const port = firstServer.port;

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const accessToken = await exchangeAccessToken(port, sessionToken, resourceBaseUrl);

      const beforeRestart = await fetch(`http://127.0.0.1:${port}/health`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      expect(beforeRestart.status).toBe(200);

      await firstServer.close();

      const secondServer = await startHttpServer(router, {
        port: 0,
        tokenPath,
        resolvePublicBaseUrl: () => resourceBaseUrl,
      });
      try {
        const afterRestart = await fetch(`http://127.0.0.1:${secondServer.port}/health`, {
          headers: { Authorization: `Bearer ${accessToken}` },
        });
        expect(afterRestart.status).toBe(200);
      } finally {
        await secondServer.close();
      }
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('recovers stale MCP session IDs after a server restart', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const firstServer = await startHttpServer(router, { port: 0, tokenPath });
    const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();

    try {
      const initialize = await fetch(`http://127.0.0.1:${firstServer.port}/mcp`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${sessionToken}`,
          Accept: 'application/json, text/event-stream',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: 1,
          method: 'initialize',
          params: {
            protocolVersion: '2025-03-26',
            capabilities: {},
            clientInfo: { name: 'mvmt-test', version: '0.0.0' },
          },
        }),
      });
      expect(initialize.status).toBe(200);
      const staleSessionId = initialize.headers.get('mcp-session-id');
      expect(staleSessionId).toBeTruthy();
      await initialize.text();

      await firstServer.close();

      const secondServer = await startHttpServer(router, { port: 0, tokenPath });
      try {
        const listTools = await fetch(`http://127.0.0.1:${secondServer.port}/mcp`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${sessionToken}`,
            Accept: 'application/json, text/event-stream',
            'Content-Type': 'application/json',
            'Mcp-Protocol-Version': '2025-03-26',
            'Mcp-Session-Id': staleSessionId!,
          },
          body: JSON.stringify({ jsonrpc: '2.0', id: 2, method: 'tools/list', params: {} }),
        });

        expect(listTools.status).toBe(200);
        await expect(listTools.text()).resolves.toContain('"tools"');
      } finally {
        await secondServer.close();
      }
    } finally {
      await firstServer.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects OAuth access tokens when the advertised resource changes across restart', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const firstServer = await startHttpServer(router, {
      port: 0,
      tokenPath,
      resolvePublicBaseUrl: () => 'https://mvmt.example.com',
    });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const accessToken = await exchangeAccessToken(firstServer.port, sessionToken, 'https://mvmt.example.com');

      await firstServer.close();

      const secondServer = await startHttpServer(router, {
        port: 0,
        tokenPath,
        resolvePublicBaseUrl: () => 'https://other.example.com',
      });
      try {
        const afterRestart = await fetch(`http://127.0.0.1:${secondServer.port}/health`, {
          headers: { Authorization: `Bearer ${accessToken}` },
        });
        expect(afterRestart.status).toBe(401);
      } finally {
        await secondServer.close();
      }
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('accepts legacy audience-less OAuth access tokens during the compatibility window', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const signingKeyPath = path.join(tmp, '.mvmt', '.signing-key');
    const signingKey = ensureSigningKey(signingKeyPath);
    const legacyStore = new OAuthStore({ signingKey });
    const legacyToken = legacyStore.issueAccessToken({ clientId: 'claude' }).token;
    const server = await startHttpServer(router, { port: 0, tokenPath, signingKeyPath });

    try {
      const response = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${legacyToken}` },
      });
      expect(response.status).toBe(200);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects unknown OAuth clients with a quarantine error once clients[] is configured', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const signingKeyPath = path.join(tmp, '.mvmt', '.signing-key');
    const requestLogs: Array<{ kind: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      signingKeyPath,
      clients: [
        {
          id: 'codex',
          name: 'Codex CLI',
          auth: { type: 'token', tokenHash: hashApiToken('codex-local-token') },
          rawToolsEnabled: true,
          permissions: [],
        },
      ],
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const oauthStore = new OAuthStore({ signingKey: ensureSigningKey(signingKeyPath) });
      const accessToken = oauthStore.issueAccessToken({
        clientId: 'unknown-dcr-client',
        audience: `http://127.0.0.1:${server.port}/mcp`,
      }).token;
      const response = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      expect(response.status).toBe(403);
      expect(await response.json()).toEqual({
        error: 'oauth_client_quarantined',
        error_description: 'OAuth client_id is not mapped to a configured mvmt client; admin must approve',
      });
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'health.auth',
            status: 403,
            clientId: 'quarantine:unknown-dcr-client',
            detail: 'quarantined oauth_client_id=unknown-dcr-client',
          }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('authorizes OAuth sessions with a scoped API token selected by the user', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const signingKeyPath = path.join(tmp, '.mvmt', '.signing-key');
    const requestLogs: Array<{ kind: string; status: number; detail?: string; clientId?: string }> = [];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      signingKeyPath,
      clients: [
        {
          id: 'codex',
          name: 'Codex CLI',
          auth: { type: 'token', tokenHash: hashApiToken('codex-api-token') },
          rawToolsEnabled: false,
          permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
        },
      ],
      requestLog: (entry) => requestLogs.push(entry),
    });

    try {
      const redirectUri = 'https://codex.example/callback';
      const registration = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ redirect_uris: [redirectUri] }),
      });
      expect(registration.status).toBe(201);
      const { client_id: oauthClientId } = await registration.json() as { client_id: string };
      const { verifier, challenge } = s256Pair();
      const resource = `http://127.0.0.1:${server.port}/mcp`;

      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: oauthClientId,
          redirect_uri: redirectUri,
          resource,
          code_challenge: challenge,
          code_challenge_method: 'S256',
          api_token: 'codex-api-token',
        }),
      });
      expect(authorize.status).toBe(302);
      const code = new URL(authorize.headers.get('location')!).searchParams.get('code');
      expect(code).toBeTruthy();

      const token = await fetch(`http://127.0.0.1:${server.port}/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          code: code!,
          client_id: oauthClientId,
          redirect_uri: redirectUri,
          resource,
          code_verifier: verifier,
        }),
      });
      expect(token.status).toBe(200);
      const grant = await token.json() as { access_token: string };

      const health = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${grant.access_token}` },
      });
      expect(health.status).toBe(200);
      expect(requestLogs).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            kind: 'oauth.authorize',
            status: 302,
            clientId: oauthClientId,
            detail: expect.stringContaining('authorized_client=codex'),
          }),
          expect.objectContaining({
            kind: 'health.request',
            status: 200,
          }),
        ]),
      );
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('invalidates existing OAuth grants when the selected API token is rotated', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const signingKeyPath = path.join(tmp, '.mvmt', '.signing-key');
    let clients: ClientConfig[] = [
      {
        id: 'codex',
        name: 'Codex CLI',
        credentialVersion: 1,
        auth: { type: 'token', tokenHash: hashApiToken('codex-api-token') },
        rawToolsEnabled: false,
        permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
      },
    ];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      signingKeyPath,
      clients: () => clients,
    });

    try {
      const redirectUri = 'https://codex.example/callback';
      const registration = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ redirect_uris: [redirectUri] }),
      });
      expect(registration.status).toBe(201);
      const { client_id: oauthClientId } = await registration.json() as { client_id: string };
      const { verifier, challenge } = s256Pair();
      const resource = `http://127.0.0.1:${server.port}/mcp`;

      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: oauthClientId,
          redirect_uri: redirectUri,
          resource,
          code_challenge: challenge,
          code_challenge_method: 'S256',
          api_token: 'codex-api-token',
        }),
      });
      expect(authorize.status).toBe(302);
      const code = new URL(authorize.headers.get('location')!).searchParams.get('code');
      expect(code).toBeTruthy();

      const token = await fetch(`http://127.0.0.1:${server.port}/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'authorization_code',
          code: code!,
          client_id: oauthClientId,
          redirect_uri: redirectUri,
          resource,
          code_verifier: verifier,
        }),
      });
      expect(token.status).toBe(200);
      const grant = await token.json() as { access_token: string; refresh_token: string };

      const beforeRotate = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${grant.access_token}` },
      });
      expect(beforeRotate.status).toBe(200);

      const current = clients[0]!;
      clients = [
        {
          ...current,
          credentialVersion: 2,
          auth: { type: 'token', tokenHash: hashApiToken('rotated-api-token') },
        },
      ];

      const afterRotate = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${grant.access_token}` },
      });
      expect(afterRotate.status).toBe(401);

      const refresh = await fetch(`http://127.0.0.1:${server.port}/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          grant_type: 'refresh_token',
          refresh_token: grant.refresh_token,
          client_id: oauthClientId,
        }),
      });
      const refreshBody = await refresh.json() as { error?: string };
      expect(refresh.status).toBe(400);
      expect(refreshBody.error).toBe('invalid_grant');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('keeps existing OAuth grants after scoped API token permissions are edited', async () => {
    const { index, tmp: indexTmp } = await createTextIndexServerFixture();
    const router = new ToolRouter(undefined, { contextIndex: index });
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    let clients: ClientConfig[] = [
      {
        id: 'codex',
        name: 'Codex CLI',
        credentialVersion: 1,
        auth: { type: 'token', tokenHash: hashApiToken('codex-api-token') },
        rawToolsEnabled: false,
        permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
      },
    ];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: () => clients,
    });

    try {
      const accessToken = await exchangeAccessTokenWithApiToken(server.port, 'codex-api-token');
      const sessionId = await initializeMcpSession(server.port, accessToken);
      const beforeEdit = await mcpJsonRequest(server.port, accessToken, sessionId, 2, 'tools/list', {});
      expect(beforeEdit.result.tools.map((tool: { name: string }) => tool.name)).toEqual([
        'search',
        'list',
        'read',
      ]);

      const current = clients[0]!;
      clients = [
        {
          ...current,
          permissions: [{ path: '/workspace/**', actions: ['search', 'read', 'write'] }],
        },
      ];

      const afterEdit = await mcpJsonRequest(server.port, accessToken, sessionId, 3, 'tools/list', {});
      expect(afterEdit.result.tools.map((tool: { name: string }) => tool.name)).toEqual([
        'search',
        'list',
        'read',
        'write',
        'remove',
      ]);

      const write = await mcpJsonRequest(server.port, accessToken, sessionId, 4, 'tools/call', {
        name: 'write',
        arguments: { path: '/workspace/after-oauth-edit.md', content: 'after edit' },
      });
      expect(write.result.isError).not.toBe(true);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
      fs.rmSync(indexTmp, { recursive: true, force: true });
    }
  });

  it('rejects OAuth approval with an invalid scoped API token', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: [
        {
          id: 'codex',
          name: 'Codex CLI',
          auth: { type: 'token', tokenHash: hashApiToken('codex-api-token') },
          rawToolsEnabled: false,
          permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
        },
      ],
    });

    try {
      const redirectUri = 'https://codex.example/callback';
      const registration = await fetch(`http://127.0.0.1:${server.port}/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ redirect_uris: [redirectUri] }),
      });
      const { client_id: oauthClientId } = await registration.json() as { client_id: string };
      const { challenge } = s256Pair();
      const authorize = await fetch(`http://127.0.0.1:${server.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: oauthClientId,
          redirect_uri: redirectUri,
          resource: `http://127.0.0.1:${server.port}/mcp`,
          code_challenge: challenge,
          code_challenge_method: 'S256',
          api_token: 'wrong-token',
        }),
      });

      expect(authorize.status).toBe(401);
      const body = await authorize.text();
      expect(body).toContain('Invalid API token. Try again.');
      expect(body).not.toContain('name="code"');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rejects the session token on /mcp once clients[] is configured', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: [
        {
          id: 'codex',
          name: 'Codex CLI',
          auth: { type: 'token', tokenHash: hashApiToken('codex-local-token') },
          rawToolsEnabled: true,
          permissions: [],
        },
      ],
    });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const response = await fetch(`http://127.0.0.1:${server.port}/mcp`, {
        method: 'GET',
        headers: { Authorization: `Bearer ${sessionToken}` },
      });
      expect(response.status).toBe(401);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('preserves legacy session-token access when clients[] is absent', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const response = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${sessionToken}` },
      });
      expect(response.status).toBe(200);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('enforces client binding for direct bearer-token requests', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: [
        {
          id: 'codex',
          name: 'Codex CLI',
          clientBinding: 'codex-cli',
          auth: { type: 'token', tokenHash: hashApiToken('codex-local-token') },
          rawToolsEnabled: true,
          permissions: [],
        },
      ],
    });

    try {
      const rejected = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: 'Bearer codex-local-token', 'User-Agent': 'curl/8.0' },
      });
      expect(rejected.status).toBe(401);

      const accepted = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: 'Bearer codex-local-token', 'User-Agent': 'Codex-CLI/1.0' },
      });
      expect(accepted.status).toBe(200);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('filters MCP tools/list and denies tool calls by resolved client permissions', async () => {
    const { index, tmp: indexTmp } = await createTextIndexServerFixture();
    const router = new ToolRouter(undefined, { contextIndex: index });
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: [
        {
          id: 'searcher',
          name: 'Search-only client',
          auth: { type: 'token', tokenHash: hashApiToken('search-token') },
          rawToolsEnabled: true,
          permissions: [{ path: '/workspace/**', actions: ['search'] }],
        },
      ],
    });

    try {
      const sessionId = await initializeMcpSession(server.port, 'search-token');
      const listTools = await mcpJsonRequest(server.port, 'search-token', sessionId, 2, 'tools/list', {});
      expect(listTools.result.tools.map((tool: { name: string }) => tool.name)).toEqual([
        'search',
      ]);

      const denied = await mcpJsonRequest(server.port, 'search-token', sessionId, 3, 'tools/call', {
        name: 'read',
        arguments: { path: '/workspace/note.md' },
      });
      expect(denied.result.isError).toBe(true);
      expect(denied.result.content[0].text).toContain('missing_permission path=/workspace/note.md action=read');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
      fs.rmSync(indexTmp, { recursive: true, force: true });
    }
  });

  it('applies edited token permissions to existing MCP sessions on the next request', async () => {
    const { index, tmp: indexTmp } = await createTextIndexServerFixture();
    const router = new ToolRouter(undefined, { contextIndex: index });
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    let clients: ClientConfig[] = [
      {
        id: 'writer',
        name: 'Writer client',
        auth: { type: 'token', tokenHash: hashApiToken('writer-token') },
        rawToolsEnabled: true,
        permissions: [{ path: '/workspace/**', actions: ['search', 'read', 'write'] }],
      },
    ];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: () => clients,
    });

    try {
      const sessionId = await initializeMcpSession(server.port, 'writer-token');
      const initialWrite = await mcpJsonRequest(server.port, 'writer-token', sessionId, 2, 'tools/call', {
        name: 'write',
        arguments: { path: '/workspace/new.md', content: 'before edit' },
      });
      expect(initialWrite.result.isError).not.toBe(true);

      const current = clients[0]!;
      clients = [
        {
          ...current,
          permissions: [{ path: '/workspace/**', actions: ['search', 'read'] }],
        },
      ];

      const listTools = await mcpJsonRequest(server.port, 'writer-token', sessionId, 3, 'tools/list', {});
      expect(listTools.result.tools.map((tool: { name: string }) => tool.name)).toEqual([
        'search',
        'list',
        'read',
      ]);

      const deniedWrite = await mcpJsonRequest(server.port, 'writer-token', sessionId, 4, 'tools/call', {
        name: 'write',
        arguments: { path: '/workspace/after-edit.md', content: 'after edit' },
      });
      expect(deniedWrite.result.isError).toBe(true);
      expect(deniedWrite.result.content[0].text).toContain('missing_permission path=/workspace/after-edit.md action=write');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
      fs.rmSync(indexTmp, { recursive: true, force: true });
    }
  });

  it('denies writes after a token changes from path-specific write to all read', async () => {
    const { index, tmp: indexTmp } = await createTextIndexServerFixture();
    const router = new ToolRouter(undefined, { contextIndex: index });
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    let clients: ClientConfig[] = [
      {
        id: 'claude',
        name: 'Claude',
        auth: { type: 'token', tokenHash: hashApiToken('claude-token') },
        rawToolsEnabled: true,
        permissions: [{ path: '/workspace/**', actions: ['search', 'read', 'write'] }],
      },
    ];
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: () => clients,
    });

    try {
      const sessionId = await initializeMcpSession(server.port, 'claude-token');
      const initialWrite = await mcpJsonRequest(server.port, 'claude-token', sessionId, 2, 'tools/call', {
        name: 'write',
        arguments: { path: '/workspace/before-edit.md', content: 'before edit' },
      });
      expect(initialWrite.result.isError).not.toBe(true);

      const current = clients[0]!;
      clients = [
        {
          ...current,
          permissions: [{ path: '/**', actions: ['search', 'read'] }],
        },
      ];

      const listTools = await mcpJsonRequest(server.port, 'claude-token', sessionId, 3, 'tools/list', {});
      expect(listTools.result.tools.map((tool: { name: string }) => tool.name)).toEqual([
        'search',
        'list',
        'read',
      ]);

      const listMounts = await mcpJsonRequest(server.port, 'claude-token', sessionId, 4, 'tools/call', {
        name: 'list',
        arguments: {},
      });
      expect(JSON.parse(listMounts.result.content[0].text).entries).toEqual([
        expect.objectContaining({ path: '/workspace', write_access: false }),
      ]);

      const deniedWrite = await mcpJsonRequest(server.port, 'claude-token', sessionId, 5, 'tools/call', {
        name: 'write',
        arguments: { path: '/workspace/after-edit.md', content: 'after edit' },
      });
      expect(deniedWrite.result.isError).toBe(true);
      expect(deniedWrite.result.content[0].text).toContain('missing_permission path=/workspace/after-edit.md action=write');
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
      fs.rmSync(indexTmp, { recursive: true, force: true });
    }
  });

  it('serves mount tools over MCP for clients without raw tool access', async () => {
    const { index, tmp: indexTmp } = await createTextIndexServerFixture({
      mountName: 'notes',
      mountPath: '/notes',
      files: { 'projects/launch.md': '# Launch\nShip it.' },
    });
    const router = new ToolRouter(undefined, { contextIndex: index });
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      clients: [
        {
          id: 'chatgpt',
          name: 'ChatGPT',
          auth: { type: 'token', tokenHash: hashApiToken('chatgpt-token') },
          rawToolsEnabled: false,
          permissions: [{ path: '/notes/**', actions: ['search', 'read'] }],
        },
      ],
    });

    try {
      const sessionId = await initializeMcpSession(server.port, 'chatgpt-token');
      const listTools = await mcpJsonRequest(server.port, 'chatgpt-token', sessionId, 2, 'tools/list', {});
      expect(listTools.result.tools.map((tool: { name: string }) => tool.name)).toEqual([
        'search',
        'list',
        'read',
      ]);

      const search = await mcpJsonRequest(server.port, 'chatgpt-token', sessionId, 3, 'tools/call', {
        name: 'search',
        arguments: { query: 'launch' },
      });
      expect(JSON.parse(search.result.content[0].text).results).toEqual([
        expect.objectContaining({ mount: 'notes', path: '/notes/projects/launch.md' }),
      ]);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
      fs.rmSync(indexTmp, { recursive: true, force: true });
    }
  });

  it('revokes outstanding OAuth access tokens the moment the signing key file is rewritten', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const signingKeyPath = path.join(tmp, '.mvmt', '.signing-key');
    const server = await startHttpServer(router, { port: 0, tokenPath, signingKeyPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const accessToken = await exchangeAccessToken(server.port, sessionToken);

      const beforeRotate = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      expect(beforeRotate.status).toBe(200);

      // Simulate internal session-token rotation writing a new signing key. The
      // running server must pick it up without restart.
      rotateSigningKey(signingKeyPath);

      const afterRotate = await fetch(`http://127.0.0.1:${server.port}/health`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      expect(afterRotate.status).toBe(401);
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('rate-limits auth-surface routes and returns 429 once the bucket is exhausted', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const server = await startHttpServer(router, {
      port: 0,
      tokenPath,
      rateLimits: { auth: { windowMs: 60_000, max: 2 } },
    });

    try {
      const hit = () =>
        fetch(`http://127.0.0.1:${server.port}/register`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ redirect_uris: ['https://rl.example/cb'] }),
        });

      const first = await hit();
      const second = await hit();
      const third = await hit();

      expect(first.status).toBe(201);
      expect(second.status).toBe(201);
      expect(third.status).toBe(429);
      expect(third.headers.get('retry-after')).toBeTruthy();
    } finally {
      await server.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  it('persists OAuth client registrations across server restarts', async () => {
    const router = new ToolRouter();
    await router.initialize();
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-test-'));
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const firstServer = await startHttpServer(router, { port: 0, tokenPath });

    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const registration = await fetch(`http://127.0.0.1:${firstServer.port}/register`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${sessionToken}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_id: 'persisted-client',
          redirect_uris: ['https://persisted.example/cb'],
        }),
      });
      expect(registration.status).toBe(201);
    } finally {
      await firstServer.close();
    }

    const secondServer = await startHttpServer(router, { port: 0, tokenPath });
    try {
      const sessionToken = fs.readFileSync(tokenPath, 'utf-8').trim();
      const resource = `http://127.0.0.1:${secondServer.port}/mcp`;
      const { challenge } = s256Pair();
      const authorize = await fetch(`http://127.0.0.1:${secondServer.port}/authorize`, {
        method: 'POST',
        redirect: 'manual',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          response_type: 'code',
          client_id: 'persisted-client',
          redirect_uri: 'https://persisted.example/cb',
          resource,
          code_challenge: challenge,
          code_challenge_method: 'S256',
          session_token: sessionToken,
        }),
      });
      // A 302 proves the server trusted the registered redirect_uri
      // even though /register was never called on this fresh instance.
      expect(authorize.status).toBe(302);
    } finally {
      await secondServer.close();
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });
});

interface TextIndexServerFixtureOptions {
  mountName?: string;
  mountPath?: string;
  files?: Record<string, string>;
}

async function createTextIndexServerFixture(
  options: TextIndexServerFixtureOptions = {},
): Promise<{ index: TextContextIndex; tmp: string }> {
  const mountName = options.mountName ?? 'workspace';
  const mountPath = options.mountPath ?? '/workspace';
  const files = options.files ?? { 'note.md': 'alpha note' };
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-server-index-'));
  const root = path.join(tmp, 'root');
  fs.mkdirSync(root);

  for (const [relativePath, content] of Object.entries(files)) {
    const filePath = path.join(root, relativePath);
    await fs.promises.mkdir(path.dirname(filePath), { recursive: true });
    await fs.promises.writeFile(filePath, content, 'utf-8');
  }

  const config = parseConfig({
    version: 1,
    mounts: [{ name: mountName, type: 'local_folder', path: mountPath, root, writeAccess: true }],
  });
  const index = new TextContextIndex({
    mounts: config.mounts,
    indexPath: path.join(tmp, 'index.json'),
  });
  await index.rebuild();
  return { index, tmp };
}

function canListenOn(port: number): Promise<boolean> {
  return new Promise((resolve) => {
    const server = createServer();
    server.once('error', () => resolve(false));
    server.listen(port, '127.0.0.1', () => {
      server.close(() => resolve(true));
    });
  });
}

function s256Pair(): { verifier: string; challenge: string } {
  const verifier = 'test-verifier-' + Math.random().toString(36).slice(2);
  const challenge = createHash('sha256').update(verifier).digest('base64url');
  return { verifier, challenge };
}

async function registerClient(port: number, clientId: string, redirectUris: string[], sessionToken: string): Promise<void> {
  const response = await fetch(`http://127.0.0.1:${port}/register`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${sessionToken}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ client_id: clientId, redirect_uris: redirectUris }),
  });
  expect(response.status).toBe(201);
}

async function exchangeAccessToken(port: number, sessionToken: string, resourceBaseUrl?: string): Promise<string> {
  const redirectUri = 'https://codex.example/callback';
  const { verifier, challenge } = s256Pair();
  const resource = `${resourceBaseUrl ?? `http://127.0.0.1:${port}`}/mcp`;
  await registerClient(port, 'codex', [redirectUri], sessionToken);
  const authorize = await fetch(`http://127.0.0.1:${port}/authorize`, {
    method: 'POST',
    redirect: 'manual',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      response_type: 'code',
      client_id: 'codex',
      redirect_uri: redirectUri,
      resource,
      code_challenge: challenge,
      code_challenge_method: 'S256',
      session_token: sessionToken,
    }),
  });
  const location = authorize.headers.get('location');
  const code = location ? new URL(location).searchParams.get('code') : undefined;
  expect(authorize.status).toBe(302);
  expect(code).toBeTruthy();

  const token = await fetch(`http://127.0.0.1:${port}/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code: code!,
      client_id: 'codex',
      redirect_uri: redirectUri,
      resource,
      code_verifier: verifier,
    }),
  });
  const body = await token.json();
  expect(token.status).toBe(200);
  expect(body.access_token).toBeTypeOf('string');
  return body.access_token as string;
}

async function exchangeAccessTokenWithApiToken(port: number, apiToken: string): Promise<string> {
  const redirectUri = 'https://codex.example/callback';
  const { verifier, challenge } = s256Pair();
  const resource = `http://127.0.0.1:${port}/mcp`;
  const registration = await fetch(`http://127.0.0.1:${port}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ redirect_uris: [redirectUri] }),
  });
  expect(registration.status).toBe(201);
  const { client_id: clientId } = await registration.json() as { client_id: string };
  const authorize = await fetch(`http://127.0.0.1:${port}/authorize`, {
    method: 'POST',
    redirect: 'manual',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      response_type: 'code',
      client_id: clientId,
      redirect_uri: redirectUri,
      resource,
      code_challenge: challenge,
      code_challenge_method: 'S256',
      api_token: apiToken,
    }),
  });
  expect(authorize.status).toBe(302);
  const code = new URL(authorize.headers.get('location')!).searchParams.get('code');
  expect(code).toBeTruthy();

  const token = await fetch(`http://127.0.0.1:${port}/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code: code!,
      client_id: clientId,
      redirect_uri: redirectUri,
      resource,
      code_verifier: verifier,
    }),
  });
  const body = await token.json();
  expect(token.status).toBe(200);
  expect(body.access_token).toBeTypeOf('string');
  return body.access_token as string;
}

async function exchangeAuthorizationCodeForToken(input: {
  port: number;
  clientId: string;
  redirectUri: string;
  code: string;
  verifier: string;
  resource?: string;
}): Promise<string> {
  const token = await fetch(`http://127.0.0.1:${input.port}/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code: input.code,
      client_id: input.clientId,
      redirect_uri: input.redirectUri,
      ...(input.resource ? { resource: input.resource } : {}),
      code_verifier: input.verifier,
    }),
  });
  const body = await token.json();
  expect(token.status).toBe(200);
  expect(body.access_token).toBeTypeOf('string');
  return body.access_token as string;
}

function expectAccessTokenAudience(accessToken: string, signingKeyPath: string, expectedAudience: string): void {
  const validator = new OAuthStore({ signingKey: ensureSigningKey(signingKeyPath) });
  const validated = validator.validateAccessToken(`Bearer ${accessToken}`, {
    expectedAudience,
    allowLegacyNoAudience: false,
  });
  expect(validated?.audience).toBe(expectedAudience);
}

async function initializeMcpSession(port: number, token: string): Promise<string> {
  const response = await fetch(`http://127.0.0.1:${port}/mcp`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'application/json, text/event-stream',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'initialize',
      params: {
        protocolVersion: '2025-03-26',
        capabilities: {},
        clientInfo: { name: 'mvmt-policy-test', version: '0.0.0' },
      },
    }),
  });
  expect(response.status).toBe(200);
  const sessionId = response.headers.get('mcp-session-id');
  expect(sessionId).toBeTruthy();
  await response.text();
  return sessionId!;
}

async function loginDashboard(port: number, username: string, password: string): Promise<string> {
  const response = await fetch(`http://127.0.0.1:${port}/dashboard/api/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  expect(response.status).toBe(200);
  const setCookie = response.headers.get('set-cookie');
  expect(setCookie).toContain('Path=/;');
  const cookie = setCookie?.split(';')[0];
  expect(cookie).toBeTruthy();
  return cookie!;
}

async function mcpJsonRequest(
  port: number,
  token: string,
  sessionId: string,
  id: number,
  method: string,
  params: Record<string, unknown>,
): Promise<any> {
  const response = await fetch(`http://127.0.0.1:${port}/mcp`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'application/json, text/event-stream',
      'Content-Type': 'application/json',
      'Mcp-Protocol-Version': '2025-03-26',
      'Mcp-Session-Id': sessionId,
    },
    body: JSON.stringify({ jsonrpc: '2.0', id, method, params }),
  });
  expect(response.status).toBe(200);
  return parseMcpResponse(await response.text());
}

function parseMcpResponse(text: string): any {
  if (text.trimStart().startsWith('{')) return JSON.parse(text);
  const dataLine = text.split('\n').find((line) => line.startsWith('data: '));
  if (!dataLine) throw new Error(`Could not parse MCP response: ${text}`);
  return JSON.parse(dataLine.slice('data: '.length));
}

function inlineScriptBodies(html: string): string[] {
  const bodies: string[] = [];
  const lowerHtml = html.toLowerCase();
  let cursor = 0;

  while (cursor < html.length) {
    const openStart = findScriptOpenTag(lowerHtml, cursor);
    if (openStart === -1) break;
    const openEnd = html.indexOf('>', openStart);
    if (openEnd === -1) break;

    const closeStart = findScriptCloseTag(lowerHtml, openEnd + 1);
    if (closeStart === -1) break;
    bodies.push(html.slice(openEnd + 1, closeStart));

    const closeEnd = html.indexOf('>', closeStart);
    if (closeEnd === -1) break;
    cursor = closeEnd + 1;
  }

  return bodies;
}

function appBaseFromHtml(html: string, pathname: string): string {
  const script = inlineScriptBodies(html).find((body) => body.includes('function appBase'));
  expect(script).toBeTruthy();
  const fn = jsFunctionDeclaration(script ?? '', 'appBase');
  const sandbox = {
    location: { pathname, search: '', hash: '' },
    result: '',
  };
  Function(
    'sandbox',
    `with (sandbox) {
      ${fn}
      sandbox.result = appBase();
    }`,
  )(sandbox);
  return sandbox.result;
}

function dashboardRequestUrlFromHtml(html: string, pathname: string, url: string): string {
  const script = inlineScriptBodies(html).find((body) => body.includes('function dashboardRequestUrl'));
  expect(script).toBeTruthy();
  const snippet = [
    jsVariableDeclaration(script ?? '', 'DASHBOARD_API_PREFIX'),
    jsVariableDeclaration(script ?? '', 'APP_API_PREFIX'),
    jsFunctionDeclaration(script ?? '', 'dashboardBasePath'),
    jsFunctionDeclaration(script ?? '', 'appBasePath'),
    jsFunctionDeclaration(script ?? '', 'dashboardRequestUrl'),
  ].join('\n');
  const sandbox = {
    location: { pathname, search: '', hash: '' },
    result: '',
  };
  Function(
    'sandbox',
    `with (sandbox) {
      ${snippet}
      sandbox.result = dashboardRequestUrl(${JSON.stringify(url)});
    }`,
  )(sandbox);
  return sandbox.result;
}

function jsVariableDeclaration(script: string, name: string): string {
  const match = script.match(new RegExp(`var ${name} = [^;]+;`));
  if (!match) throw new Error(`Missing dashboard script variable: ${name}`);
  return match[0];
}

function jsFunctionDeclaration(script: string, name: string): string {
  const start = script.indexOf(`function ${name}(`);
  if (start === -1) throw new Error(`Missing dashboard script function: ${name}`);
  const braceStart = script.indexOf('{', start);
  if (braceStart === -1) throw new Error(`Missing dashboard script function body: ${name}`);

  // Deliberately simple: these dashboard URL helpers do not contain braces
  // inside strings or comments, and this test should fail if that changes.
  let depth = 0;
  for (let index = braceStart; index < script.length; index += 1) {
    const char = script[index];
    if (char === '{') depth += 1;
    if (char === '}') {
      depth -= 1;
      if (depth === 0) return script.slice(start, index + 1);
    }
  }

  throw new Error(`Unterminated dashboard script function: ${name}`);
}

function findScriptOpenTag(lowerHtml: string, from: number): number {
  let cursor = from;
  while (cursor < lowerHtml.length) {
    const candidate = lowerHtml.indexOf('<script', cursor);
    if (candidate === -1) return -1;
    if (isScriptNameBoundary(lowerHtml[candidate + '<script'.length])) return candidate;
    cursor = candidate + 1;
  }
  return -1;
}

function findScriptCloseTag(lowerHtml: string, from: number): number {
  let cursor = from;
  while (cursor < lowerHtml.length) {
    const candidate = lowerHtml.indexOf('</', cursor);
    if (candidate === -1) return -1;

    let nameStart = candidate + 2;
    while (isHtmlSpace(lowerHtml[nameStart])) nameStart += 1;
    if (
      lowerHtml.slice(nameStart, nameStart + 'script'.length) === 'script' &&
      isScriptNameBoundary(lowerHtml[nameStart + 'script'.length])
    ) {
      return candidate;
    }

    cursor = candidate + 2;
  }
  return -1;
}

function isScriptNameBoundary(char: string | undefined): boolean {
  return char === undefined || char === '>' || char === '/' || isHtmlSpace(char);
}

function isHtmlSpace(char: string | undefined): boolean {
  return char === ' ' || char === '\t' || char === '\n' || char === '\r' || char === '\f';
}
