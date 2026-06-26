import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { createLiveClientsResolver } from '../src/cli/start.js';
import { parseConfig, saveConfig } from '../src/config/loader.js';
import { TextContextIndex } from '../src/context/text-index.js';
import { hashApiToken } from '../src/utils/api-token-hash.js';

describe('start helpers', () => {
  let tmp: string;
  let configPath: string;

  beforeEach(async () => {
    tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-start-'));
    configPath = path.join(tmp, 'config.yaml');
  });

  afterEach(async () => {
    await fs.rm(tmp, { recursive: true, force: true });
  });

  it('reloads client policy from config when resolving HTTP auth clients', async () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'notes', type: 'local_folder', path: '/notes', root: tmp },
      ],
      clients: [
        {
          id: 'old-token',
          name: 'Old token',
          auth: { type: 'token', tokenHash: hashApiToken('old') },
          permissions: [{ path: '/notes/**', actions: ['search', 'read'] }],
        },
      ],
    });
    await saveConfig(configPath, config);
    const resolveClients = createLiveClientsResolver(configPath, config);

    expect(resolveClients()?.map((client) => client.id)).toEqual(['old-token']);

    const nextConfig = parseConfig({
      ...config,
      clients: [
        {
          id: 'new-token',
          name: 'New token',
          auth: { type: 'token', tokenHash: hashApiToken('new') },
          permissions: [{ path: '/notes/**', actions: ['search', 'read'] }],
        },
      ],
    });
    await saveConfig(configPath, nextConfig);

    expect(resolveClients()?.map((client) => client.id)).toEqual(['new-token']);
    expect(config.clients?.map((client) => client.id)).toEqual(['new-token']);
  });

  it('keeps the previous client policy when live config reload fails', async () => {
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'notes', type: 'local_folder', path: '/notes', root: tmp },
      ],
      clients: [
        {
          id: 'known-good',
          name: 'Known good',
          auth: { type: 'token', tokenHash: hashApiToken('known-good') },
          permissions: [{ path: '/notes/**', actions: ['search', 'read'] }],
        },
      ],
    });
    await saveConfig(configPath, config);
    const resolveClients = createLiveClientsResolver(configPath, config);

    expect(resolveClients()?.map((client) => client.id)).toEqual(['known-good']);

    await fs.writeFile(configPath, 'version: [', 'utf-8');

    expect(resolveClients()?.map((client) => client.id)).toEqual(['known-good']);
    expect(config.clients?.map((client) => client.id)).toEqual(['known-good']);
  });

  it('reloads mount write access into the running text index', async () => {
    const root = path.join(tmp, 'notes');
    await fs.mkdir(root);
    const config = parseConfig({
      version: 1,
      mounts: [
        { name: 'notes', type: 'local_folder', path: '/notes', root, writeAccess: false },
      ],
      clients: [
        {
          id: 'writer',
          name: 'Writer',
          auth: { type: 'token', tokenHash: hashApiToken('writer') },
          permissions: [{ path: '/notes/**', actions: ['search', 'read', 'write'] }],
        },
      ],
    });
    await saveConfig(configPath, config);
    const index = new TextContextIndex({
      mounts: config.mounts,
      indexPath: path.join(tmp, 'index.json'),
    });
    const resolveClients = createLiveClientsResolver(configPath, config, index);

    await expect(index.write('/notes/new.md', 'content')).rejects.toThrow('read-only');

    const nextConfig = parseConfig({
      ...config,
      mounts: [{ ...config.mounts[0], writeAccess: true }],
    });
    await saveConfig(configPath, nextConfig);
    resolveClients();

    await expect(index.write('/notes/new.md', 'content')).resolves.toMatchObject({
      path: '/notes/new.md',
      content: 'content',
    });
    expect(config.mounts[0].writeAccess).toBe(true);
  });
});
