import { existsSync } from 'node:fs';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { resolveUpstreamBases } from '../../../../lib/trpc-upstream';
import { GET, POST } from './route';

function resolveRepoRoot(): string {
  const candidates = [
    process.cwd(),
    path.resolve(process.cwd(), '..'),
    path.resolve(process.cwd(), '..', '..'),
  ];

  for (const candidate of candidates) {
    if (existsSync(path.join(candidate, 'mcp.jsonc')) || existsSync(path.join(candidate, 'mcp.json'))) {
      return candidate;
    }
  }

  for (const candidate of candidates) {
    if (existsSync(path.join(candidate, 'pnpm-workspace.yaml'))) {
      return candidate;
    }
  }

  return path.resolve(process.cwd(), '..', '..');
}

const REPO_ROOT = resolveRepoRoot();

function getCompatConfigPaths(configDir: string): { jsoncPath: string; jsonPath: string } {
  return {
    jsoncPath: path.join(configDir, 'mcp.jsonc'),
    jsonPath: path.join(configDir, 'mcp.json'),
  };
}

async function readOptionalFile(filePath: string): Promise<string | null> {
  try {
    return await fs.readFile(filePath, 'utf-8');
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === 'ENOENT') {
      return null;
    }

    throw error;
  }
}

describe('resolveUpstreamBases', () => {
  const originalUpstream = process.env.TORMENTNEXUS_TRPC_UPSTREAM;

  afterEach(() => {
    if (originalUpstream === undefined) {
      delete process.env.TORMENTNEXUS_TRPC_UPSTREAM;
    } else {
      process.env.TORMENTNEXUS_TRPC_UPSTREAM = originalUpstream;
    }
  });

  it('includes TormentNexus core\'s default tRPC port before legacy fallbacks', () => {
    delete process.env.TORMENTNEXUS_TRPC_UPSTREAM;

    expect(resolveUpstreamBases()).toEqual([
      'http://127.0.0.1:3100/trpc',
      'http://127.0.0.1:4000/trpc',
      'http://127.0.0.1:3001/trpc',
    ]);
  });

  it('prepends a configured upstream while deduplicating defaults', () => {
    process.env.TORMENTNEXUS_TRPC_UPSTREAM = 'http://127.0.0.1:4000/trpc';

    expect(resolveUpstreamBases()).toEqual([
      'http://127.0.0.1:4000/trpc',
      'http://127.0.0.1:3100/trpc',
      'http://127.0.0.1:3001/trpc',
    ]);
  });
});

describe('legacy MCP dashboard compatibility bridge', () => {
  const originalFetch = global.fetch;
  const originalUpstream = process.env.TORMENTNEXUS_TRPC_UPSTREAM;
  const originalTormentNexusConfigDir = process.env.TORMENTNEXUS_CONFIG_DIR;
  let compatConfigDir = '';

  beforeEach(async () => {
    compatConfigDir = await fs.mkdtemp(path.join(os.tmpdir(), 'tormentnexus-trpc-compat-'));
    process.env.TORMENTNEXUS_CONFIG_DIR = compatConfigDir;
  });

  afterEach(async () => {
    global.fetch = originalFetch;

    if (originalUpstream === undefined) {
      delete process.env.TORMENTNEXUS_TRPC_UPSTREAM;
    } else {
      process.env.TORMENTNEXUS_TRPC_UPSTREAM = originalUpstream;
    }

    if (originalTormentNexusConfigDir === undefined) {
      delete process.env.TORMENTNEXUS_CONFIG_DIR;
    } else {
      process.env.TORMENTNEXUS_CONFIG_DIR = originalTormentNexusConfigDir;
    }

    if (compatConfigDir) {
      await fs.rm(compatConfigDir, { recursive: true, force: true });
      compatConfigDir = '';
    }
  });

  it('returns compatibility data for modern MCP procedure batches when upstreams are unavailable', async () => {
    process.env.TORMENTNEXUS_TRPC_UPSTREAM = 'http://127.0.0.1:59999/trpc';
    global.fetch = vi.fn(async () => {
      throw new Error('connect ECONNREFUSED');
    }) as typeof fetch;

    const request = new Request(
      'http://localhost:3010/api/trpc/mcp.listServers,mcp.listTools,mcp.getStatus?batch=1',
      {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ 0: { json: null }, 1: { json: null }, 2: { json: null } }),
      },
    );

    const response = await POST(request);
    const payload = await response.json();

    expect(response.status).toBe(200);
    expect(response.headers.get('x-tormentnexus-trpc-compat')).toBe('legacy-mcp-dashboard-bridge');
    expect(Array.isArray(payload)).toBe(true);
    expect(payload).toHaveLength(3);
    expect(Array.isArray(payload[0]?.result?.data)).toBe(true);
    expect(Array.isArray(payload[1]?.result?.data)).toBe(true);
    expect(payload[2]).toEqual({
      result: {
        data: {
          initialized: true,
          serverCount: payload[0].result.data.length,
          toolCount: 0,
          connectedCount: expect.any(Number),
        },
      },
    });
  });

  it('supports mixed legacy and modern MCP procedure aliases in the same batch', async () => {
    process.env.TORMENTNEXUS_TRPC_UPSTREAM = 'http://127.0.0.1:59999/trpc';
    global.fetch = vi.fn(async () => {
      throw new Error('connect ECONNREFUSED');
    }) as typeof fetch;

    const request = new Request(
      'http://localhost:3010/api/trpc/mcpServers.list,mcp.listTools,mcp.getStatus?batch=1',
      {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ 0: { json: null }, 1: { json: null }, 2: { json: null } }),
      },
    );

    const response = await POST(request);
    const payload = await response.json();

    expect(response.status).toBe(200);
    expect(response.headers.get('x-tormentnexus-trpc-compat')).toBe('legacy-mcp-dashboard-bridge');
    expect(Array.isArray(payload)).toBe(true);
    expect(payload).toHaveLength(3);
    expect(Array.isArray(payload[0]?.result?.data)).toBe(true);
    expect(Array.isArray(payload[1]?.result?.data)).toBe(true);
    expect(payload[2]).toEqual({
      result: {
        data: {
          initialized: true,
          serverCount: payload[0].result.data.length,
          toolCount: 0,
          connectedCount: expect.any(Number),
        },
      },
    });
  });

  it('probes top-level mcpServers.list when bridging legacy MCP batches', async () => {
    process.env.TORMENTNEXUS_TRPC_UPSTREAM = 'http://127.0.0.1:4100/trpc';
    const upstreamServers = [
      {
        uuid: 'server-1',
        name: 'upstream-memory',
        status: 'configured',
        toolCount: 2,
        _meta: {
          uuid: 'server-1',
          status: 'ready',
          metadataSource: 'db-cache',
          toolCount: 2,
          lastSuccessfulBinaryLoadAt: '2026-03-11T00:00:00.000Z',
          crashCount: 0,
          maxAttempts: 0,
        },
      },
    ];

    global.fetch = vi.fn(async (input) => {
      const url = String(input);

      if (url === 'http://127.0.0.1:4100/trpc/mcp.listServers,mcp.listTools,mcp.getStatus?batch=1') {
        return new Response('not found', { status: 404 });
      }

      if (url === 'http://127.0.0.1:4100/trpc/mcpServers.list?input=%7B%7D') {
        return new Response(JSON.stringify({ result: { data: upstreamServers } }), {
          status: 200,
          headers: { 'content-type': 'application/json' },
        });
      }

      throw new Error(`Unexpected fetch: ${url}`);
    }) as typeof fetch;

    const request = new Request(
      'http://localhost:3010/api/trpc/mcp.listServers,mcp.listTools,mcp.getStatus?batch=1',
      {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ 0: { json: null }, 1: { json: null }, 2: { json: null } }),
      },
    );

    const response = await POST(request);
    const payload = await response.json();

    expect(response.status).toBe(200);
    expect(response.headers.get('x-tormentnexus-trpc-compat')).toBe('legacy-mcp-dashboard-bridge');
    expect(payload[0]?.result?.data).toEqual(upstreamServers);
    expect(payload[2]?.result?.data).toEqual({
      initialized: true,
      serverCount: 1,
      toolCount: 0,
      connectedCount: 0,
    });
    expect(
      (global.fetch as ReturnType<typeof vi.fn>).mock.calls.some(
        ([url, init]) => String(url) === 'http://127.0.0.1:4100/trpc/mcpServers.list?input=%7B%7D'
          && typeof init === 'object'
          && init !== null
          && 'method' in init
          && init.method === 'GET',
      ),
    ).toBe(true);
  });

  it('returns local dashboard fallback data for richer MCP pages when upstreams are unavailable', async () => {
    process.env.TORMENTNEXUS_TRPC_UPSTREAM = 'http://127.0.0.1:59999/trpc';
    global.fetch = vi.fn(async () => {
      throw new Error('connect ECONNREFUSED');
    }) as typeof fetch;

    const request = new Request(
      'http://localhost:3010/api/trpc/startupStatus,mcp.getWorkingSet,mcp.searchTools,mcp.getJsoncEditor,serverHealth.check?batch=1',
      {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          0: { json: null },
          1: { json: null },
          2: { json: null },
          3: { json: null },
          4: { json: null },
        }),
      },
    );

    const response = await POST(request);
    const payload = await response.json();

    expect(response.status).toBe(200);
    expect(response.headers.get('x-tormentnexus-trpc-compat')).toBe('local-dashboard-fallback');
    expect(Array.isArray(payload)).toBe(true);
    expect(payload).toHaveLength(5);

    expect(payload[0]?.result?.data).toEqual(expect.objectContaining({
      status: expect.stringMatching(/^(starting|degraded)$/),
      ready: false,
      checks: expect.objectContaining({
        mcpAggregator: expect.objectContaining({
          initialization: 'compat-fallback',
          persistedServerCount: expect.any(Number),
        }),
        configSync: expect.objectContaining({
          status: expect.objectContaining({
            lastServerCount: expect.any(Number),
          }),
        }),
        extensionBridge: expect.objectContaining({
          ready: false,
          clientCount: 0,
        }),
      }),
    }));
    expect(payload[0].result.data.checks.configSync.status.lastServerCount)
      .toBe(payload[0].result.data.checks.mcpAggregator.persistedServerCount);

    expect(payload[1]?.result?.data).toEqual({
      tools: [],
      limits: {
        maxLoadedTools: 24,
        maxHydratedSchemas: 8,
      },
    });
    expect(payload[2]?.result?.data).toEqual([]);
    expect(payload[3]?.result?.data).toEqual(expect.objectContaining({
      path: expect.stringMatching(/mcp\.jsonc?$|mcp\.json$/),
      content: expect.stringContaining('mcpServers'),
    }));
    expect(payload[4]?.result?.data).toEqual({
      status: 'unavailable',
      crashCount: 0,
      maxAttempts: 0,
    });
  });

  it('normalizes batched bulk import payloads before proxying them upstream', async () => {
    process.env.TORMENTNEXUS_TRPC_UPSTREAM = 'http://127.0.0.1:3100/trpc';
    const upstreamResponse = [
      {
        result: {
          data: {
            imported: 1,
            errors: [],
          },
        },
      },
    ];

    global.fetch = vi.fn(async (_input, init) => {
      expect(String(_input)).toBe('http://127.0.0.1:3100/trpc/mcpServers.bulkImport');
      expect(init?.body).toBe(JSON.stringify([
        {
          name: 'test_stdio_import',
          type: 'STDIO',
          command: 'npx',
          args: ['-y', '@modelcontextprotocol/server-memory'],
          metadataStrategy: 'auto',
        },
      ]));

      return new Response(JSON.stringify(upstreamResponse), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      });
    }) as typeof fetch;

    const request = new Request(
      'http://localhost:3010/api/trpc/mcpServers.bulkImport?batch=1',
      {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          0: {
            json: [
              {
                name: 'test_stdio_import',
                type: 'STDIO',
                command: 'npx',
                args: ['-y', '@modelcontextprotocol/server-memory'],
                metadataStrategy: 'auto',
              },
            ],
          },
        }),
      },
    );

    const response = await POST(request);
    const payload = await response.json();

    expect(response.status).toBe(200);
    expect(payload).toEqual(upstreamResponse);
    expect(global.fetch).toHaveBeenCalledTimes(1);
  });

  it('supports local pseudo-managed MCP server actions when upstreams are unavailable', async () => {
    process.env.TORMENTNEXUS_TRPC_UPSTREAM = 'http://127.0.0.1:59999/trpc';
    global.fetch = vi.fn(async () => {
      throw new Error('connect ECONNREFUSED');
    }) as typeof fetch;

    const { jsoncPath: MCP_JSONC_PATH, jsonPath: MCP_JSON_PATH } = getCompatConfigPaths(compatConfigDir);

    const originalJsonc = await readOptionalFile(MCP_JSONC_PATH);
    const originalJson = await readOptionalFile(MCP_JSON_PATH);
    const testServerName = 'compat_action_test';

    try {
      const seededConfig = {
        mcpServers: {
          [testServerName]: {
            command: 'npx',
            args: ['-y', '@modelcontextprotocol/server-memory'],
          },
        },
      };

      await fs.writeFile(MCP_JSONC_PATH, `${JSON.stringify(seededConfig, null, 2)}\n`, 'utf-8');
      await fs.writeFile(MCP_JSON_PATH, `${JSON.stringify(seededConfig, null, 2)}\n`, 'utf-8');

      const listResponse = await POST(new Request('http://localhost:3010/api/trpc/mcpServers.list', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ json: null }),
      }));
      const listPayload = await listResponse.json();
      const listData = listPayload?.result?.data as Array<{ uuid: string; name: string; _meta?: { status?: string } }>;

      expect(listResponse.status).toBe(200);
      expect(listResponse.headers.get('x-tormentnexus-trpc-compat')).toBe('legacy-mcp-dashboard-bridge');
      expect(Array.isArray(listData)).toBe(true);
      expect(listData[0]?.uuid).toEqual(expect.stringMatching(/^local-/));
      expect(listData[0]?.name).toBe(testServerName);

      const serverUuid = listData[0]?.uuid;
      expect(serverUuid).toBeTruthy();

      const getResponse = await POST(new Request('http://localhost:3010/api/trpc/mcpServers.get', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ json: { uuid: serverUuid } }),
      }));
      const getPayload = await getResponse.json();

      expect(getResponse.status).toBe(200);
      expect(getResponse.headers.get('x-tormentnexus-trpc-compat')).toBe('local-dashboard-fallback');
      expect(getPayload?.result?.data).toEqual(expect.objectContaining({
        uuid: serverUuid,
        name: testServerName,
      }));

      const getViaQueryResponse = await GET(new Request(
        `http://localhost:3010/api/trpc/mcpServers.get?input=${encodeURIComponent(JSON.stringify({ json: { uuid: serverUuid } }))}`,
        { method: 'GET' },
      ));
      const getViaQueryPayload = await getViaQueryResponse.json();

      expect(getViaQueryResponse.status).toBe(200);
      expect(getViaQueryResponse.headers.get('x-tormentnexus-trpc-compat')).toBe('local-dashboard-fallback');
      expect(getViaQueryPayload?.result?.data).toEqual(expect.objectContaining({
        uuid: serverUuid,
        name: testServerName,
      }));

      const reloadResponse = await POST(new Request('http://localhost:3010/api/trpc/mcpServers.reloadMetadata', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ json: { uuid: serverUuid, mode: 'binary' } }),
      }));
      const reloadPayload = await reloadResponse.json();

      expect(reloadResponse.status).toBe(200);
      expect(reloadResponse.headers.get('x-tormentnexus-trpc-compat')).toBe('local-mcp-managed-action');
      expect(reloadPayload?.result?.data?.metadata).toEqual(expect.objectContaining({
        uuid: serverUuid,
        status: 'ready',
        metadataSource: 'local-binary',
      }));

      const healthResponse = await POST(new Request('http://localhost:3010/api/trpc/serverHealth.check', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ json: { serverUuid } }),
      }));
      const healthPayload = await healthResponse.json();

      expect(healthResponse.status).toBe(200);
      expect(healthPayload?.result?.data).toEqual({
        status: 'ready',
        crashCount: 0,
        maxAttempts: 0,
      });

      const clearResponse = await POST(new Request('http://localhost:3010/api/trpc/mcpServers.clearMetadataCache', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ json: { uuid: serverUuid } }),
      }));
      const clearPayload = await clearResponse.json();

      expect(clearResponse.status).toBe(200);
      expect(clearPayload?.result?.data?.metadata).toEqual(expect.objectContaining({
        uuid: serverUuid,
        status: 'pending',
        metadataSource: 'local-config-fallback',
      }));

      const updateResponse = await POST(new Request('http://localhost:3010/api/trpc/mcpServers.update', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({
          json: {
            uuid: serverUuid,
            description: 'Compat fallback server',
          },
        }),
      }));
      const updatePayload = await updateResponse.json();

      expect(updateResponse.status).toBe(200);
      expect(updatePayload?.result?.data).toEqual(expect.objectContaining({
        uuid: serverUuid,
        name: testServerName,
        description: 'Compat fallback server',
      }));

      const deleteResponse = await POST(new Request('http://localhost:3010/api/trpc/mcpServers.delete', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ json: { uuid: serverUuid } }),
      }));
      const deletePayload = await deleteResponse.json();

      expect(deleteResponse.status).toBe(200);
      expect(deletePayload?.result?.data).toEqual(expect.objectContaining({
        uuid: serverUuid,
        name: testServerName,
      }));

      const finalListResponse = await POST(new Request('http://localhost:3010/api/trpc/mcpServers.list', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ json: null }),
      }));
      const finalListPayload = await finalListResponse.json();

      expect(finalListPayload?.result?.data).toEqual([]);
    } finally {
      if (originalJsonc === null) {
        await fs.rm(MCP_JSONC_PATH, { force: true });
      } else {
        await fs.writeFile(MCP_JSONC_PATH, originalJsonc, 'utf-8');
      }

      if (originalJson === null) {
        await fs.rm(MCP_JSON_PATH, { force: true });
      } else {
        await fs.writeFile(MCP_JSON_PATH, originalJson, 'utf-8');
      }
    }
  });
});
