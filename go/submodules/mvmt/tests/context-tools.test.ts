import { describe, expect, it } from 'vitest';
import { CONTEXT_TOOLS } from '../src/apps/mcp/tools/index.js';
import { listTool } from '../src/apps/mcp/tools/list.js';

describe('context tool registry', () => {
  it('registers the canonical tool surface in display order', () => {
    expect(CONTEXT_TOOLS.map((tool) => tool.name)).toEqual([
      'search',
      'list',
      'read',
      'write',
      'remove',
    ]);
    expect(CONTEXT_TOOLS.map((tool) => tool.definition.namespacedName)).toEqual([
      'search',
      'list',
      'read',
      'write',
      'remove',
    ]);
  });

  it('describes agent trigger behavior for the built-in tools', () => {
    const descriptions = new Map(CONTEXT_TOOLS.map((tool) => [tool.name, tool.definition.description]));

    expect(descriptions.get('search')).toContain('Use first when the user asks about their own notes');
    expect(descriptions.get('list')).toContain('For topic/content questions, use search first');
    expect(descriptions.get('read')).toContain('Use after search or list');
    expect(descriptions.get('write')).toContain('Use only when the user explicitly asks');
    expect(descriptions.get('remove')).toContain('Use only when the user explicitly asks');
  });

  it('reports write_access from the effective client permission, not only mount capability', async () => {
    const result = await listTool.handle({}, {
      index: {
        mountNameForPath: () => undefined,
        list: async () => [
          {
            mount: 'workspace',
            path: '/workspace',
            type: 'directory',
            size: 0,
            mtime_ms: 0,
            write_access: true,
          },
          {
            mount: 'scratch',
            path: '/scratch',
            type: 'directory',
            size: 0,
            mtime_ms: 0,
            write_access: true,
          },
        ],
      },
      access: {
        allowedMounts: () => [],
        pathMayExposeEntry: () => true,
        pathAllowed: (inputPath, action) => inputPath === '/scratch' && action === 'write',
      },
    } as never);

    const content = result.content[0];
    expect(content.type).toBe('text');
    const payload = JSON.parse(content.type === 'text' ? content.text : '{}');

    expect(payload.entries).toEqual([
      expect.objectContaining({ path: '/workspace', write_access: false }),
      expect.objectContaining({ path: '/scratch', write_access: true }),
    ]);
  });
});
