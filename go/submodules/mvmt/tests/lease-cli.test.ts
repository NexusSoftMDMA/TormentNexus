import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  addPathsToLease,
  createFolderLease,
  listFolderLeases,
  revokeFolderLease,
  setFolderLeasePublished,
} from '../src/cli/lease.js';
import { readConfig } from '../src/config/loader.js';
import { listLeases } from '../src/core/leases/store.js';
import { findLeaseSecret, leaseSecretsPathForLeaseStore } from '../src/core/leases/secrets.js';

describe('lease CLI helpers', () => {
  let tmp: string;
  let configPath: string;
  let leaseStorePath: string;
  let logSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(async () => {
    tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-lease-cli-'));
    configPath = path.join(tmp, 'config.yaml');
    leaseStorePath = path.join(tmp, '.leases.json');
    logSpy = vi.spyOn(console, 'log').mockImplementation(() => undefined);
  });

  afterEach(async () => {
    logSpy.mockRestore();
    await fs.rm(tmp, { recursive: true, force: true });
  });

  it('creates a 24h read-only folder lease and mount for a local folder', async () => {
    const folder = path.join(tmp, 'Taxes');
    await fs.mkdir(folder);

    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Sarah - tax docs',
    });

    const config = readConfig(configPath);
    expect(config.mounts).toHaveLength(1);
    expect(config.mounts[0]).toMatchObject({
      name: 'lease-taxes',
      path: '/Taxes',
      root: folder,
      writeAccess: false,
    });
    const leases = listLeases(leaseStorePath);
    expect(leases).toHaveLength(1);
    expect(leases[0]).toMatchObject({
      label: 'Sarah - tax docs',
      path: '/Taxes',
      resources: [{ path: '/Taxes', sourcePath: '/Taxes', type: 'folder' }],
      permissions: ['read'],
      downloadCount: 0,
      uploadCount: 0,
    });
    expect(leases[0].expiresAt).toBeTruthy();
    const secret = findLeaseSecret(leaseSecretsPathForLeaseStore(leaseStorePath), leases[0].id);
    expect(secret?.token).toMatch(/^mvmt_l_/);
    expect(await fs.readFile(leaseStorePath, 'utf-8')).not.toContain(secret!.token);
    expect(logSpy.mock.calls.flat().join('\n')).toContain('(24h default)');
  });

  it('reuses one internal mount for multiple leases on the same folder', async () => {
    const folder = path.join(tmp, 'Downloads');
    await fs.mkdir(folder);

    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Sarah downloads',
      expires: '1h',
    });
    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Ben downloads',
      expires: '24h',
    });

    const config = readConfig(configPath);
    expect(config.mounts).toHaveLength(1);
    const leases = listLeases(leaseStorePath);
    expect(leases).toHaveLength(2);
    expect(new Set(leases.map((lease) => lease.path))).toEqual(new Set(['/Downloads']));
    expect(leases.map((lease) => lease.label)).toEqual(['Sarah downloads', 'Ben downloads']);
  });

  it('creates one lease for multiple local paths', async () => {
    const folder = path.join(tmp, 'Taxes');
    const file = path.join(tmp, 'w2.pdf');
    await fs.mkdir(folder);
    await fs.writeFile(file, 'pdf');

    await createFolderLease([folder, file], {
      config: configPath,
      leaseStorePath,
      label: 'Sarah files',
      expires: '1h',
    });

    const config = readConfig(configPath);
    expect(config.mounts).toHaveLength(2);
    const lease = listLeases(leaseStorePath)[0]!;
    expect(lease).toMatchObject({
      label: 'Sarah files',
      path: '/Taxes',
      resources: [
        { path: '/Taxes', sourcePath: '/Taxes', type: 'folder' },
        { path: '/w2.pdf', sourcePath: '/w2.pdf', type: 'file' },
      ],
    });
  });

  it('creates one-time download leases', async () => {
    const file = path.join(tmp, 'invite.txt');
    await fs.writeFile(file, 'secret');

    await createFolderLease(file, {
      config: configPath,
      leaseStorePath,
      label: 'Invite',
      expires: '1h',
      downloads: '1',
    });

    const lease = listLeases(leaseStorePath)[0]!;
    expect(lease).toMatchObject({
      label: 'Invite',
      maxDownloads: 1,
      downloadCount: 0,
    });
    expect(logSpy.mock.calls.flat().join('\n')).toContain('Downloads: 1');
  });

  it('treats --downloads -1 as unlimited', async () => {
    const file = path.join(tmp, 'photos.zip');
    await fs.writeFile(file, 'zip');

    await createFolderLease(file, {
      config: configPath,
      leaseStorePath,
      label: 'Photos',
      downloads: '-1',
    });

    expect(listLeases(leaseStorePath)[0].maxDownloads).toBeUndefined();
  });

  it('rejects download limits for upload-only leases', async () => {
    const folder = path.join(tmp, 'dropbox');
    await fs.mkdir(folder);

    await expect(createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Uploads',
      mode: 'upload',
      downloads: '1',
    })).rejects.toThrow(/Download limits require/);
  });

  it('adds paths to an existing read lease without replacing the token record', async () => {
    const folder = path.join(tmp, 'Taxes');
    const receipts = path.join(tmp, 'Receipts');
    await fs.mkdir(folder);
    await fs.mkdir(receipts);

    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Sarah files',
      expires: '1h',
    });
    const before = listLeases(leaseStorePath)[0]!;

    await addPathsToLease(before.id, receipts, { config: configPath, leaseStorePath });

    const config = readConfig(configPath);
    expect(config.mounts).toHaveLength(2);
    const after = listLeases(leaseStorePath)[0]!;
    expect(after.id).toBe(before.id);
    expect(after.tokenHash).toBe(before.tokenHash);
    expect(after.resources).toEqual([
      { path: '/Taxes', sourcePath: '/Taxes', type: 'folder' },
      { path: '/Receipts', sourcePath: '/Receipts', type: 'folder' },
    ]);
    expect(logSpy.mock.calls.flat().join('\n')).toContain('Existing lease token and URL now include these paths.');
  });

  it('creates an upload-only folder lease with a writable internal mount', async () => {
    const folder = path.join(tmp, 'Phone Drop');
    await fs.mkdir(folder);

    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Sarah uploads',
      mode: 'upload',
      expires: '1h',
    });

    const config = readConfig(configPath);
    expect(config.mounts).toHaveLength(1);
    expect(config.mounts[0]).toMatchObject({
      name: 'lease-write-phone-drop',
      path: '/Phone-Drop',
      root: folder,
      writeAccess: true,
    });
    expect(listLeases(leaseStorePath)[0]).toMatchObject({
      label: 'Sarah uploads',
      path: '/Phone-Drop',
      permissions: ['upload'],
      uploadCount: 0,
    });
    expect(logSpy.mock.calls.flat().join('\n')).toContain('Mode: upload only');
  });

  it('creates a two-way folder lease with browse and upload permissions', async () => {
    const folder = path.join(tmp, 'Shared Drop');
    await fs.mkdir(folder);

    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Sarah exchange',
      mode: 'two-way',
      expires: '1h',
    });

    const config = readConfig(configPath);
    expect(config.mounts[0]).toMatchObject({
      name: 'lease-write-shared-drop',
      path: '/Shared-Drop',
      root: folder,
      writeAccess: true,
    });
    expect(listLeases(leaseStorePath)[0]).toMatchObject({
      label: 'Sarah exchange',
      path: '/Shared-Drop',
      permissions: ['read', 'upload'],
    });
    expect(logSpy.mock.calls.flat().join('\n')).toContain('Mode: browse/download/upload');
  });

  it('does not add paths to upload-only leases', async () => {
    const folder = path.join(tmp, 'Phone Drop');
    const other = path.join(tmp, 'Other Drop');
    await fs.mkdir(folder);
    await fs.mkdir(other);
    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Sarah uploads',
      mode: 'upload',
      expires: '1h',
    });
    const id = listLeases(leaseStorePath)[0]!.id;

    await expect(addPathsToLease(id, other, { config: configPath, leaseStorePath }))
      .rejects.toThrow(/Upload leases currently support one folder/);
  });

  it('lists and revokes folder leases', async () => {
    const folder = path.join(tmp, 'photos');
    await fs.mkdir(folder);
    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Family photos',
      expires: 'never',
    });
    const id = listLeases(leaseStorePath)[0].id;

    await listFolderLeases({ leaseStorePath });
    expect(logSpy.mock.calls.flat().join('\n')).toContain('Family photos');

    await revokeFolderLease(id, { leaseStorePath });
    expect(listLeases(leaseStorePath)[0].revokedAt).toBeTruthy();
    expect(findLeaseSecret(leaseSecretsPathForLeaseStore(leaseStorePath), id)).toBeUndefined();
  });

  it('publishes and unpublishes a folder lease', async () => {
    const folder = path.join(tmp, 'shared');
    await fs.mkdir(folder);
    await createFolderLease(folder, {
      config: configPath,
      leaseStorePath,
      label: 'Shared folder',
      expires: 'never',
    });
    const id = listLeases(leaseStorePath)[0].id;

    await setFolderLeasePublished(id, false, { leaseStorePath });
    expect(listLeases(leaseStorePath)[0].published).toBe(false);

    await setFolderLeasePublished(id, true, { leaseStorePath });
    expect(listLeases(leaseStorePath)[0].published).toBe(true);

    await expect(setFolderLeasePublished('missing', true, { leaseStorePath }))
      .rejects.toThrow(/Unknown lease/);
    await expect(setFolderLeasePublished(undefined, true, { leaseStorePath }))
      .rejects.toThrow(/Lease id is required/);
  });
});
