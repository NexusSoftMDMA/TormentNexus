import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { addShareLink, listShareLinks, removeShareLink } from '../src/cli/share.js';
import { parseConfig, saveConfig } from '../src/config/loader.js';
import { listShares } from '../src/share/store.js';

describe('share CLI helpers', () => {
  let tmp: string;
  let configPath: string;
  let shareStorePath: string;
  let logSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(async () => {
    tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-share-cli-'));
    configPath = path.join(tmp, 'config.yaml');
    shareStorePath = path.join(tmp, '.shares.json');
    logSpy = vi.spyOn(console, 'log').mockImplementation(() => undefined);
  });

  afterEach(async () => {
    logSpy.mockRestore();
    await fs.rm(tmp, { recursive: true, force: true });
  });

  it('creates a file share with the default 24h expiration', async () => {
    const filePath = path.join(tmp, 'payload.txt');
    await fs.writeFile(filePath, 'payload', 'utf-8');
    await saveConfig(configPath, parseConfig({
      version: 1,
      mounts: [{ name: 'payload', type: 'local_folder', path: '/payload.txt', root: filePath }],
    }));

    await addShareLink('/payload.txt', { config: configPath, shareStorePath });

    const shares = listShares(shareStorePath);
    expect(shares).toHaveLength(1);
    expect(shares[0]).toMatchObject({ path: '/payload.txt', downloadCount: 0 });
    expect(shares[0].expiresAt).toBeTruthy();
    expect(logSpy.mock.calls.flat().join('\n')).toContain('(24h default)');
  });

  it('lists and removes shares from the configured store', async () => {
    const filePath = path.join(tmp, 'payload.txt');
    await fs.writeFile(filePath, 'payload', 'utf-8');
    await saveConfig(configPath, parseConfig({
      version: 1,
      mounts: [{ name: 'payload', type: 'local_folder', path: '/payload.txt', root: filePath }],
    }));
    await addShareLink('/payload.txt', { config: configPath, shareStorePath, expires: '1h' });
    const id = listShares(shareStorePath)[0].id;

    await listShareLinks({ shareStorePath });
    expect(logSpy.mock.calls.flat().join('\n')).toContain('/payload.txt');

    await removeShareLink(id, { shareStorePath });
    expect(listShares(shareStorePath)).toEqual([]);
  });
});
