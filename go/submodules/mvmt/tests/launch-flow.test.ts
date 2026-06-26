import { execFile } from 'child_process';
import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { promisify } from 'util';
import { describe, expect, it } from 'vitest';
import { readConfig } from '../src/config/loader.js';
import { defaultTextIndexPath, TextContextIndex } from '../src/context/text-index.js';
import { startHttpServer } from '../src/server/index.js';
import { ToolRouter } from '../src/apps/mcp/router.js';

const execFileAsync = promisify(execFile);
const root = path.resolve(import.meta.dirname, '..');
const cliArgs = ['--import', 'tsx', 'bin/mvmt.ts', '--no-update-check'];

describe('first 10 minute workflow', () => {
  it('mounts folders, creates a scoped token, indexes, connects, and enforces write gates', async () => {
    const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-launch-flow-'));
    const configPath = path.join(tmp, 'config.yaml');
    const tokenPath = path.join(tmp, '.mvmt', '.session-token');
    const notesRoot = path.join(tmp, 'notes');
    const workspaceRoot = path.join(tmp, 'workspace');
    await fs.mkdir(notesRoot);
    await fs.mkdir(workspaceRoot);

    try {
      await fs.writeFile(path.join(notesRoot, 'launch.md'), 'mvmt launch notes mention scoped mounts.', 'utf-8');
      await fs.writeFile(path.join(workspaceRoot, 'README.md'), '# Workspace\nEditable project docs.', 'utf-8');
      await runCli(['mounts', 'add', 'notes', notesRoot, '--config', configPath, '--mount-path', '/notes', '--read-only']);
      await runCli([
        'mounts',
        'add',
        'workspace',
        workspaceRoot,
        '--config',
        configPath,
        '--mount-path',
        '/workspace',
        '--write',
        '--protect',
        '.env',
        '--protect',
        '.env.*',
      ]);
      const tokenOutput = await runCli([
        'token',
        'add',
        'codex',
        '--config',
        configPath,
        '--read',
        '/notes',
        '--write',
        '/workspace',
        '--ttl',
        '7d',
      ]);
      const apiToken = extractApiToken(tokenOutput.stdout);

      const reindex = await runCli(['reindex', '--config', configPath]);
      expect(reindex.stdout).toContain('Indexed 2 text files');

      const config = readConfig(configPath);
      const index = new TextContextIndex({
        mounts: config.mounts,
        indexPath: defaultTextIndexPath(configPath),
      });
      const router = new ToolRouter(undefined, { contextIndex: index });
      await router.initialize();
      const server = await startHttpServer(router, {
        port: 0,
        tokenPath,
        clients: config.clients,
      });

      try {
        const sessionId = await initializeMcpSession(server.port, apiToken);
        const tools = await mcpJsonRequest(server.port, apiToken, sessionId, 2, 'tools/list', {});
        expect(tools.result.tools.map((tool: { name: string }) => tool.name)).toEqual([
          'search',
          'list',
          'read',
          'write',
          'remove',
        ]);

        const search = await mcpJsonRequest(server.port, apiToken, sessionId, 3, 'tools/call', {
          name: 'search',
          arguments: { query: 'scoped mounts' },
        });
        expect(JSON.parse(search.result.content[0].text).results).toEqual([
          expect.objectContaining({ path: '/notes/launch.md' }),
        ]);

        const read = await mcpJsonRequest(server.port, apiToken, sessionId, 4, 'tools/call', {
          name: 'read',
          arguments: { path: '/workspace/README.md' },
        });
        expect(JSON.parse(read.result.content[0].text)).toMatchObject({
          path: '/workspace/README.md',
          content: '# Workspace\nEditable project docs.',
        });

        const write = await mcpJsonRequest(server.port, apiToken, sessionId, 5, 'tools/call', {
          name: 'write',
          arguments: { path: '/workspace/notes/summary.md', content: 'saved from launch flow' },
        });
        expect(JSON.parse(write.result.content[0].text)).toMatchObject({
          path: '/workspace/notes/summary.md',
          content: 'saved from launch flow',
        });

        const protectedWrite = await mcpJsonRequest(server.port, apiToken, sessionId, 6, 'tools/call', {
          name: 'write',
          arguments: { path: '/workspace/.env', content: 'SECRET=value' },
        });
        expect(protectedWrite.result.isError).toBe(true);
        expect(protectedWrite.result.content[0].text).toContain('protected');
      } finally {
        await server.close();
      }
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  }, 20_000);
});

async function runCli(args: string[]): Promise<{ stdout: string; stderr: string }> {
  return execFileAsync(process.execPath, [...cliArgs, ...args], { cwd: root });
}

function extractApiToken(output: string): string {
  const match = output.match(/Token:\s+(mvmt_t_[^\s]+)/);
  if (!match) throw new Error(`Could not find API token in output:\n${output}`);
  return match[1];
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
        clientInfo: { name: 'mvmt-launch-flow-test', version: '0.0.0' },
      },
    }),
  });
  expect(response.status).toBe(200);
  const sessionId = response.headers.get('mcp-session-id');
  expect(sessionId).toBeTruthy();
  await response.text();
  return sessionId!;
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
