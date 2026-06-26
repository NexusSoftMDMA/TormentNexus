import { describe, expect, it } from 'vitest';
import {
  addMountToConfig,
  editMountInConfig,
  removeMountFromConfig,
} from '../src/cli/mounts.js';
import { parseConfig } from '../src/config/loader.js';
import { DEFAULT_MOUNT_EXCLUDE_PATTERNS, DEFAULT_MOUNT_PROTECT_PATTERNS } from '../src/config/schema.js';

const EXISTING_TOKEN_VERIFIER = 'scrypt:v1:AAAAAAAAAAAAAAAAAAAAAA:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';

describe('mount config helpers', () => {
  it('adds mounts with default protection', () => {
    const config = parseConfig({ version: 1 });
    const next = addMountToConfig(config, {
      name: 'workspace',
      root: '~/code/mvmt',
      writeAccess: true,
      description: 'Project mount',
      guidance: 'Repo-specific instructions.',
    });

    expect(next.mounts).toEqual([
      expect.objectContaining({
        name: 'workspace',
        type: 'local_folder',
        path: '/workspace',
        root: expect.stringContaining('/code/mvmt'),
        description: 'Project mount',
        guidance: 'Repo-specific instructions.',
        exclude: [...DEFAULT_MOUNT_EXCLUDE_PATTERNS],
        protect: [...DEFAULT_MOUNT_PROTECT_PATTERNS],
        writeAccess: true,
        enabled: true,
      }),
    ]);
  });

  it('edits mount root, path, write access, patterns, and enabled state', () => {
    const config = addMountToConfig(parseConfig({ version: 1 }), {
      name: 'workspace',
      root: '/old',
      writeAccess: false,
    });

    const next = editMountInConfig(config, 'workspace', {
      root: '/new',
      path: '/repo',
      writeAccess: true,
      description: 'New description',
      guidance: 'New guidance',
      exclude: ['dist/**'],
      protect: ['secrets/**'],
      enabled: false,
    });

    expect(next.mounts[0]).toMatchObject({
      name: 'workspace',
      root: '/new',
      path: '/repo',
      writeAccess: true,
      description: 'New description',
      guidance: 'New guidance',
      exclude: ['dist/**'],
      protect: ['secrets/**'],
      enabled: false,
    });
  });

  it('removes unreferenced mounts', () => {
    const config = addMountToConfig(parseConfig({ version: 1 }), {
      name: 'workspace',
      root: '/workspace',
      writeAccess: false,
    });

    expect(removeMountFromConfig(config, 'workspace').mounts).toEqual([]);
  });

  it('rejects duplicate mount names, paths, and referenced removals', () => {
    const config = addMountToConfig(parseConfig({ version: 1 }), {
      name: 'workspace',
      root: '/workspace',
      writeAccess: false,
    });

    expect(() => addMountToConfig(config, {
      name: 'workspace',
      root: '/other',
      writeAccess: false,
    })).toThrow('Mount already exists');

    expect(() => addMountToConfig(config, {
      name: 'other',
      path: '/workspace',
      root: '/other',
      writeAccess: false,
    })).toThrow('Mount path already exists');

    const referenced = parseConfig({
      version: 1,
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: '/workspace' }],
      clients: [
        {
          id: 'codex',
          name: 'Codex',
          auth: { type: 'token', tokenHash: EXISTING_TOKEN_VERIFIER },
          permissions: [{ path: '/workspace/**', actions: ['read'] }],
        },
      ],
    });

    expect(() => removeMountFromConfig(referenced, 'workspace')).toThrow('referenced by client codex');
  });
});
