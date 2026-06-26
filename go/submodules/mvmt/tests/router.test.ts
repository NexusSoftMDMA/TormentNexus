import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { describe, expect, it, vi } from 'vitest';
import { parseConfig } from '../src/config/loader.js';
import { TextContextIndex } from '../src/context/text-index.js';
import { ClientIdentity } from '../src/server/client-identity.js';
import { ToolRouter } from '../src/apps/mcp/router.js';

describe('ToolRouter', () => {
  it('does not expose tools without a text context index', async () => {
    const router = new ToolRouter();
    await router.initialize();

    expect(router.getAllTools()).toEqual([]);
    await expect(router.callTool('missing', {})).rejects.toThrow('Unknown tool: missing');
  });

  it('exposes text index tools by path/action permission', async () => {
    const { index, tmp } = await createTextIndexFixture();
    try {
      const router = new ToolRouter(undefined, { contextIndex: index });
      await router.initialize();

      const tools = router.getAllTools(client('codex', false, [
        { path: '/workspace/**', actions: ['search', 'read'] },
      ]));

      expect(tools.map((tool) => tool.namespacedName)).toEqual(['search', 'list', 'read']);
      const result = await router.callTool('search', { query: 'alpha' }, client('codex', false, [
        { path: '/workspace/**', actions: ['search', 'read'] },
      ]));
      const parsed = JSON.parse(result.content[0].type === 'text' ? result.content[0].text : '{}');
      expect(parsed.results[0]).toMatchObject({ mount: 'workspace', path: '/workspace/note.md' });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('initializes context tools only once', async () => {
    const { index, tmp } = await createTextIndexFixture();
    try {
      const router = new ToolRouter(undefined, { contextIndex: index });
      await router.initialize();
      await router.initialize();

      expect(router.getAllTools().map((tool) => tool.namespacedName)).toEqual([
        'search',
        'list',
        'read',
        'write',
        'remove',
      ]);
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('normalizes permission paths with trailing slashes without regex backtracking', async () => {
    const { index, tmp } = await createTextIndexFixture();
    try {
      const router = new ToolRouter(undefined, { contextIndex: index });
      await router.initialize();
      const identity = client('codex', false, [{ path: '/workspace/**////', actions: ['read'] }]);

      expect(router.getAllTools(identity).map((tool) => tool.namespacedName)).toEqual(['list', 'read']);
      const result = await router.callTool('read', { path: '/workspace/note.md' }, identity);
      const parsed = JSON.parse(result.content[0].type === 'text' ? result.content[0].text : '{}');

      expect(parsed).toMatchObject({ path: '/workspace/note.md', content: 'alpha note' });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('allows targeted reads with an exact file permission', async () => {
    const { index, tmp } = await createTextIndexFixture();
    try {
      const router = new ToolRouter(undefined, { contextIndex: index });
      await router.initialize();
      const identity = client('codex', false, [{ path: '/workspace/note.md', actions: ['read'] }]);

      expect(router.getAllTools(identity).map((tool) => tool.namespacedName)).toEqual(['list', 'read']);

      const rootList = await router.callTool('list', { path: '/' }, identity);
      const rootParsed = JSON.parse(rootList.content[0].type === 'text' ? rootList.content[0].text : '{}');
      expect(rootParsed.entries).toEqual([
        expect.objectContaining({ path: '/workspace' }),
      ]);

      const result = await router.callTool('read', { path: '/workspace/note.md' }, identity);
      const parsed = JSON.parse(result.content[0].type === 'text' ? result.content[0].text : '{}');
      expect(parsed).toMatchObject({ path: '/workspace/note.md', content: 'alpha note' });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('denies writes without write permission and records the client', async () => {
    const { index, tmp } = await createTextIndexFixture();
    const audit = { record: vi.fn() };
    try {
      const router = new ToolRouter(audit, { contextIndex: index });
      await router.initialize();

      const result = await router.callTool('write', { path: '/workspace/new.md', content: 'new' }, client('chatgpt', false, [
        { path: '/workspace/**', actions: ['search', 'read'] },
      ]));

      expect(result).toMatchObject({ isError: true });
      expect(audit.record).toHaveBeenCalledWith(
        expect.objectContaining({
          connectorId: 'mvmt',
          tool: 'write',
          clientId: 'chatgpt',
          deniedReason: 'missing_permission path=/workspace/new.md action=write',
          isError: true,
        }),
      );
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('allows writes when client and mount permissions allow them', async () => {
    const { index, tmp } = await createTextIndexFixture();
    try {
      const router = new ToolRouter(undefined, { contextIndex: index });
      await router.initialize();
      const identity = client('codex', false, [{ path: '/workspace/**', actions: ['read', 'write'] }]);

      const write = await router.callTool('write', { path: '/workspace/new.md', content: 'new note' }, identity);
      const writeParsed = JSON.parse(write.content[0].type === 'text' ? write.content[0].text : '{}');
      expect(writeParsed).toMatchObject({ path: '/workspace/new.md', content: 'new note' });

      const read = await router.callTool('read', { path: '/workspace/new.md' }, identity);
      const readParsed = JSON.parse(read.content[0].type === 'text' ? read.content[0].text : '{}');
      expect(readParsed).toMatchObject({ content: 'new note' });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

  it('exposes remove as the destructive text index tool', async () => {
    const { index, tmp } = await createTextIndexFixture();
    try {
      const router = new ToolRouter(undefined, { contextIndex: index });
      await router.initialize();
      const identity = client('codex', false, [{ path: '/workspace/**', actions: ['write'] }]);

      expect(router.getAllTools(identity).map((tool) => tool.namespacedName)).toEqual(['write', 'remove']);
      const result = await router.callTool('remove', { path: '/workspace/note.md' }, identity);
      const parsed = JSON.parse(result.content[0].type === 'text' ? result.content[0].text : '{}');
      expect(parsed).toMatchObject({ path: '/workspace/note.md', removed: true });
    } finally {
      await fs.rm(tmp, { recursive: true, force: true });
    }
  });

});

function client(
  id: string,
  rawToolsEnabled: boolean,
  permissions: ClientIdentity['permissions'],
): ClientIdentity {
  return {
    id,
    name: id,
    source: 'token',
    rawToolsEnabled,
    permissions,
  };
}

async function createTextIndexFixture(): Promise<{ index: TextContextIndex; tmp: string }> {
  const tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-router-index-'));
  const sourceRoot = path.join(tmp, 'source');
  await fs.mkdir(sourceRoot);
  await fs.writeFile(path.join(sourceRoot, 'note.md'), 'alpha note', 'utf-8');
  const config = parseConfig({
    version: 1,
    mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: sourceRoot, writeAccess: true }],
  });
  const index = new TextContextIndex({
    mounts: config.mounts,
    indexPath: path.join(tmp, 'index.json'),
  });
  await index.rebuild();
  return { index, tmp };
}
