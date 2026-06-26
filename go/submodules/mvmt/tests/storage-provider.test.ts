import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { parseConfig } from '../src/config/loader.js';
import { MountRegistry } from '../src/context/mount-registry.js';
import { LocalFolderStorageProvider } from '../src/context/storage-provider.js';

const itUnlessWindows = process.platform === 'win32' ? it.skip : it;

describe('LocalFolderStorageProvider', () => {
  let tmp: string;

  beforeEach(async () => {
    tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-storage-provider-'));
  });

  afterEach(async () => {
    await fs.rm(tmp, { recursive: true, force: true });
  });

  it('lists, reads, writes, and removes text files within a mount', async () => {
    await fs.mkdir(path.join(tmp, 'docs'));
    await fs.writeFile(path.join(tmp, 'docs', 'guide.md'), 'old', 'utf-8');
    const provider = createProvider(tmp, true);

    await expect(provider.list('docs')).resolves.toEqual([
      expect.objectContaining({ path: '/workspace/docs/guide.md', type: 'file' }),
    ]);
    await expect(provider.read('docs/guide.md')).resolves.toMatchObject({
      mount: 'workspace',
      path: '/workspace/docs/guide.md',
      content: 'old',
    });

    await expect(provider.write('docs/guide.md', 'new')).resolves.toMatchObject({
      content: 'new',
    });
    expect(await provider.exists('docs/guide.md')).toBe(true);

    await provider.remove('docs/guide.md');
    expect(await provider.exists('docs/guide.md')).toBe(false);
  });

  it('blocks protected writes, excluded paths, and root escapes', async () => {
    await fs.mkdir(path.join(tmp, 'protected'));
    await fs.writeFile(path.join(tmp, 'protected', 'secret.md'), 'secret', 'utf-8');
    await fs.mkdir(path.join(tmp, '.git'));
    await fs.writeFile(path.join(tmp, '.git', 'config.md'), 'git', 'utf-8');
    const provider = createProvider(tmp, true);

    await expect(provider.write('protected/secret.md', 'new')).rejects.toThrow('protected');
    await expect(provider.read('.git/config.md')).rejects.toThrow('excluded');
    await expect(provider.read('../secret.md')).rejects.toThrow('escapes mount root');
  });

  it('walks only supported, non-excluded text files', async () => {
    await fs.writeFile(path.join(tmp, 'note.md'), 'alpha', 'utf-8');
    await fs.writeFile(path.join(tmp, 'photo.jpg'), 'fake image', 'utf-8');
    await fs.mkdir(path.join(tmp, '.git'));
    await fs.writeFile(path.join(tmp, '.git', 'config.md'), 'git', 'utf-8');
    const provider = createProvider(tmp, true);

    const files: string[] = [];
    for await (const file of provider.walkTextFiles()) files.push(file.path);
    expect(files).toEqual(['/workspace/note.md']);
  });

  it('stats and streams binary files without applying the text size cap', async () => {
    const bytes = Buffer.from([0, 1, 2, 3, 4, 5]);
    await fs.writeFile(path.join(tmp, 'photo.jpg'), bytes);
    const provider = createProvider(tmp, true, 4);

    await expect(provider.list()).resolves.toEqual([
      expect.objectContaining({ path: '/workspace/photo.jpg', type: 'file', size: bytes.length }),
    ]);
    await expect(provider.read('photo.jpg')).rejects.toThrow('not a supported text file');
    await expect(provider.stat('photo.jpg')).resolves.toMatchObject({
      path: '/workspace/photo.jpg',
      type: 'file',
      size: bytes.length,
    });

    const { file, stream } = await provider.openReadStream('photo.jpg', { start: 2, end: 4 });
    expect(file).toMatchObject({ path: '/workspace/photo.jpg', type: 'file' });
    const chunks: Buffer[] = [];
    for await (const chunk of stream) chunks.push(Buffer.from(chunk));
    expect(Buffer.concat(chunks)).toEqual(Buffer.from([2, 3, 4]));
  });

  it('supports a single text file as the mount root', async () => {
    const filePath = path.join(tmp, 'note.md');
    await fs.writeFile(filePath, 'single file mount', 'utf-8');
    const provider = createProvider(filePath, true, 2 * 1024 * 1024, '/note.md');

    await expect(provider.list()).resolves.toEqual([
      expect.objectContaining({ path: '/note.md', type: 'file' }),
    ]);
    await expect(provider.read('')).resolves.toMatchObject({
      mount: 'workspace',
      path: '/note.md',
      content: 'single file mount',
    });
    const files: string[] = [];
    for await (const file of provider.walkTextFiles()) files.push(file.path);
    expect(files).toEqual(['/note.md']);
  });

  it('rejects writes to read-only mounts', async () => {
    const provider = createProvider(tmp, false);

    await expect(provider.write('new.md', 'content')).rejects.toThrow('read-only');
  });

  it('rejects over-size writes before creating the file', async () => {
    const provider = createProvider(tmp, true, 4);

    await expect(provider.write('large.md', '12345')).rejects.toThrow('too large to write');
    expect(await provider.exists('large.md')).toBe(false);
  });

  it('blocks globally sensitive paths even when config omits them', async () => {
    await fs.mkdir(path.join(tmp, '.mvmt'));
    await fs.writeFile(path.join(tmp, '.mvmt', '.session-token'), 'secret', 'utf-8');
    const provider = createProvider(tmp, true);

    await expect(provider.read('.mvmt/.session-token')).rejects.toThrow(/excluded|globally denied/);
  });

  it('blocks mounts rooted inside globally sensitive directories', async () => {
    const sensitiveRoot = path.join(tmp, '.mvmt');
    await fs.mkdir(sensitiveRoot);
    await fs.writeFile(path.join(sensitiveRoot, 'config.yaml'), 'secret', 'utf-8');
    const provider = createProvider(sensitiveRoot, true);

    await expect(provider.read('config.yaml')).rejects.toThrow('globally denied');
  });

  it('blocks single-file mount roots inside globally sensitive directories', async () => {
    const sensitiveDir = path.join(tmp, '.mvmt');
    await fs.mkdir(sensitiveDir);
    const sensitiveFile = path.join(sensitiveDir, 'config.txt');
    await fs.writeFile(sensitiveFile, 'secret', 'utf-8');
    const provider = createProvider(sensitiveFile, true, 2 * 1024 * 1024, '/config.txt');

    await expect(provider.read('')).rejects.toThrow(/excluded|globally denied/);
  });

  itUnlessWindows('blocks symlink escapes for direct and nested paths', async () => {
    const outside = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-storage-provider-outside-'));
    try {
      await fs.writeFile(path.join(outside, 'secret.md'), 'secret', 'utf-8');
      await fs.symlink(path.join(outside, 'secret.md'), path.join(tmp, 'linked-secret.md'));
      await fs.symlink(outside, path.join(tmp, 'linked-dir'));
      const provider = createProvider(tmp, true);

      await expect(provider.read('linked-secret.md')).rejects.toThrow('escapes mount root');
      await expect(provider.list('linked-dir')).rejects.toThrow('escapes mount root');
      await expect(provider.write('linked-dir/new.md', 'new')).rejects.toThrow('escapes mount root');
      await expect(provider.remove('linked-secret.md')).rejects.toThrow('escapes mount root');
    } finally {
      await fs.rm(outside, { recursive: true, force: true });
    }
  });
});

function createProvider(
  root: string,
  writeAccess: boolean,
  maxTextBytes = 2 * 1024 * 1024,
  mountPath = '/workspace',
): LocalFolderStorageProvider {
  const config = parseConfig({
    version: 1,
    mounts: [
      {
        name: 'workspace',
        type: 'local_folder',
        path: mountPath,
        root,
        exclude: ['.git/**'],
        protect: ['protected/**'],
        writeAccess,
      },
    ],
  });
  const mount = new MountRegistry(config.mounts).mounts()[0];
  return new LocalFolderStorageProvider(mount, {
    isTextPath: (inputPath) => ['.md', '.txt'].includes(path.extname(inputPath)),
    maxTextBytes,
  });
}
