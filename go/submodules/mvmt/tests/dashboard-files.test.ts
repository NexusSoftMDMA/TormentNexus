import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { parseConfig } from '../src/config/loader.js';
import {
  assertDashboardWriteAllowed,
  normalizeDashboardPath,
  resolveDashboardFileTarget,
} from '../src/apps/dashboard/files.js';

describe('dashboard file access helpers', () => {
  let tmp: string;
  let readRoot: string;
  let writeRoot: string;

  beforeEach(async () => {
    tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-dashboard-files-'));
    readRoot = path.join(tmp, 'read-root');
    writeRoot = path.join(tmp, 'write-root');
    await fs.mkdir(path.join(readRoot, 'nested'), { recursive: true });
    await fs.mkdir(writeRoot, { recursive: true });
    await fs.writeFile(path.join(readRoot, 'safe.txt'), 'safe');
    await fs.writeFile(path.join(readRoot, 'secret.txt'), 'secret');
    await fs.writeFile(path.join(readRoot, 'nested', 'note.md'), 'note');
    await fs.writeFile(path.join(writeRoot, 'draft.txt'), 'draft');
    await fs.writeFile(path.join(writeRoot, 'locked.txt'), 'locked');
  });

  afterEach(async () => {
    await fs.rm(tmp, { recursive: true, force: true });
  });

  it('rejects excluded paths via resolveDashboardFileTarget', async () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        {
          name: 'docs',
          type: 'local_folder',
          path: '/docs',
          root: readRoot,
          exclude: ['secret.txt'],
        },
      ],
    });

    await expect(resolveDashboardFileTarget(config.mounts, '/docs/secret.txt')).rejects.toThrow('excluded');
  });

  it('rejects symlinks that escape the mounted root', async () => {
    const outside = path.join(tmp, 'outside.txt');
    await fs.writeFile(outside, 'outside');
    await fs.symlink(outside, path.join(readRoot, 'outside-link.txt'));
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'docs', type: 'local_folder', path: '/docs', root: readRoot },
      ],
    });

    await expect(resolveDashboardFileTarget(config.mounts, '/docs/outside-link.txt')).rejects.toThrow('escapes mount root');
  });

  it('normalizes browser paths and enforces dashboard write gates', async () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'docs', type: 'local_folder', path: '/docs', root: readRoot },
        {
          name: 'write',
          type: 'local_folder',
          path: '/write',
          root: writeRoot,
          writeAccess: true,
          protect: ['locked.txt'],
        },
      ],
    });

    expect(normalizeDashboardPath('///write\\draft.txt')).toBe('/write/draft.txt');
    await expect(assertDashboardWriteAllowed(config.mounts, '/docs/safe.txt')).rejects.toThrow('read-only');
    await expect(assertDashboardWriteAllowed(config.mounts, '/write/locked.txt')).rejects.toThrow('protected');
    await expect(assertDashboardWriteAllowed(config.mounts, '/write/draft.txt')).resolves.toMatchObject({
      virtualPath: '/write/draft.txt',
      writeAccess: true,
    });
  });
});
