import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { performance } from 'perf_hooks';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { parseConfig } from '../src/config/loader.js';
import { TextContextIndex } from '../src/context/text-index.js';

const itUnlessWindows = process.platform === 'win32' ? it.skip : it;

describe('TextContextIndex', () => {
  let tmp: string;

  beforeEach(async () => {
    tmp = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-text-index-'));
  });

  afterEach(async () => {
    await fs.rm(tmp, { recursive: true, force: true });
  });

  it('indexes supported text files and skips binary-looking extensions', async () => {
    await fs.writeFile(path.join(tmp, 'note.md'), 'alpha project note', 'utf-8');
    await fs.writeFile(path.join(tmp, 'todo.txt'), 'beta alpha task', 'utf-8');
    await fs.writeFile(path.join(tmp, 'photo.jpg'), 'alpha in a fake image', 'utf-8');
    await fs.mkdir(path.join(tmp, '.git'));
    await fs.writeFile(path.join(tmp, '.git', 'config'), 'alpha git metadata', 'utf-8');

    const index = createIndex(tmp);
    const stats = await index.rebuild();

    expect(stats.files).toBe(2);
    const results = await index.search('alpha', ['workspace'], 10);
    expect(results.map((result) => result.path).sort()).toEqual([
      '/workspace/note.md',
      '/workspace/todo.txt',
    ]);
    await expect(index.list('/workspace')).resolves.not.toEqual(
      expect.arrayContaining([expect.objectContaining({ path: '/workspace/.git' })]),
    );
  });

  it('skips generated dependency directories during indexing even under broad mounts', async () => {
    await fs.mkdir(path.join(tmp, 'project', 'src'), { recursive: true });
    await fs.mkdir(path.join(tmp, 'project', 'node_modules', 'package'), { recursive: true });
    await fs.mkdir(path.join(tmp, 'project', 'dist'), { recursive: true });
    await fs.writeFile(path.join(tmp, 'project', 'src', 'app.js'), 'alpha local source', 'utf-8');
    await fs.writeFile(path.join(tmp, 'project', 'node_modules', 'package', 'index.js'), 'alpha vendored dependency', 'utf-8');
    await fs.writeFile(path.join(tmp, 'project', 'dist', 'bundle.js'), 'alpha generated bundle', 'utf-8');

    const index = createIndex(tmp);
    const stats = await index.rebuild();

    expect(stats).toMatchObject({ files: 1, chunks: 1 });
    await expect(index.search('alpha', ['workspace'], 10)).resolves.toEqual([
      expect.objectContaining({ path: '/workspace/project/src/app.js' }),
    ]);
  });

  it('truncates oversized indexes instead of growing without bound', async () => {
    await fs.writeFile(path.join(tmp, 'first.md'), 'alpha first', 'utf-8');
    await fs.writeFile(path.join(tmp, 'second.md'), 'alpha second', 'utf-8');
    const index = createIndex(tmp, { maxIndexedFiles: 1 });

    const stats = await index.rebuild();

    expect(stats).toEqual({ files: 1, chunks: 1, truncated: true });
    await expect(index.search('alpha', ['workspace'], 10)).resolves.toHaveLength(1);
  });

  it('truncates very large files instead of letting one file dominate the index', async () => {
    await fs.writeFile(path.join(tmp, 'large.md'), `${'alpha '.repeat(20_000)}omega`, 'utf-8');
    const index = createIndex(tmp, { maxChunksPerFile: 2 });

    const stats = await index.rebuild();

    expect(stats).toEqual({ files: 1, chunks: 2, truncated: true });
    await expect(index.search('omega', ['workspace'], 10)).resolves.toEqual([]);
  });

  it('indexes and searches a generated corpus at integration-test scale', async () => {
    await writeGeneratedCorpus(tmp, 1_200);
    const index = createIndex(tmp);

    const start = performance.now();
    const stats = await index.rebuild();
    const rebuildMs = performance.now() - start;
    const searchStart = performance.now();
    const results = await index.search('needle-latency shard-7', ['workspace'], 10);
    const searchMs = performance.now() - searchStart;

    expect(stats).toEqual({ files: 1_200, chunks: 1_200 });
    expect(results.length).toBeGreaterThan(0);
    expect(results[0].snippet).toContain('needle-latency');
    expect(rebuildMs).toBeLessThan(5_000);
    expect(searchMs).toBeLessThan(1_000);
  });

  it('uses the default index ceilings for broad generated corpora', async () => {
    await writeGeneratedCorpus(tmp, 5_100);
    const index = createIndex(tmp);

    const stats = await index.rebuild();

    expect(stats).toEqual({ files: 5_000, chunks: 5_000, truncated: true });
  }, 20_000);

  it('lists and reads virtual paths without exposing host paths', async () => {
    await fs.mkdir(path.join(tmp, 'docs'));
    await fs.writeFile(path.join(tmp, 'docs', 'guide.md'), 'setup guide', 'utf-8');

    const index = createIndex(tmp);
    await index.rebuild();

    await expect(index.list('/')).resolves.toEqual([
      expect.objectContaining({
        path: '/workspace',
        type: 'directory',
        mount: 'workspace',
        description: 'Workspace mount',
        guidance: 'Use for tests.',
        write_access: true,
      }),
    ]);
    await expect(index.list('/workspace/docs')).resolves.toEqual([
      expect.objectContaining({ path: '/workspace/docs/guide.md', type: 'file' }),
    ]);
    await expect(index.read('/workspace/docs/guide.md')).resolves.toMatchObject({
      mount: 'workspace',
      path: '/workspace/docs/guide.md',
      content: 'setup guide',
    });
    await expect(index.read('/workspace/../secret.md')).rejects.toThrow(/escapes mount root|unknown mount/);
  });

  it('enforces write access, protected paths, and expected hashes', async () => {
    await fs.mkdir(path.join(tmp, 'protected'));
    await fs.writeFile(path.join(tmp, 'draft.md'), 'old', 'utf-8');

    const index = createIndex(tmp);
    const initial = await index.read('/workspace/draft.md');

    await expect(index.write('/workspace/draft.md', 'new', 'wrong')).rejects.toThrow('hash mismatch');
    await expect(index.write('/workspace/protected/secret.md', 'secret')).rejects.toThrow('protected');

    const written = await index.write('/workspace/draft.md', 'new', initial.hash);
    expect(written.content).toBe('new');

    await expect(index.remove('/workspace/draft.md')).resolves.toMatchObject({
      mount: 'workspace',
      path: '/workspace/draft.md',
      removed: true,
    });
    await expect(index.write('/workspace/draft.md', 'recreated', written.hash)).rejects.toThrow('hash mismatch');
    await expect(index.read('/workspace/draft.md')).rejects.toThrow();
  });

  it('rejects writes to read-only mounts', async () => {
    const config = parseConfig({
      version: 1,
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: tmp, writeAccess: false }],
    });
    const index = new TextContextIndex({
      mounts: config.mounts,
      indexPath: path.join(tmp, 'index.json'),
    });

    await expect(index.write('/workspace/new.md', 'content')).rejects.toThrow('read-only');
  });

  it('reloads mount write access in place', async () => {
    const readOnlyConfig = parseConfig({
      version: 1,
      mounts: [{ name: 'workspace', type: 'local_folder', path: '/workspace', root: tmp, writeAccess: false }],
    });
    const index = new TextContextIndex({
      mounts: readOnlyConfig.mounts,
      indexPath: path.join(tmp, 'index.json'),
    });

    await expect(index.write('/workspace/new.md', 'content')).rejects.toThrow('read-only');

    const writableConfig = parseConfig({
      ...readOnlyConfig,
      mounts: [{ ...readOnlyConfig.mounts[0], writeAccess: true }],
    });
    index.updateMounts(writableConfig.mounts);

    await expect(index.write('/workspace/new.md', 'content')).resolves.toMatchObject({
      path: '/workspace/new.md',
      content: 'content',
    });
  });

  itUnlessWindows('does not index or read symlink escapes', async () => {
    const outside = await fs.mkdtemp(path.join(os.tmpdir(), 'mvmt-text-index-outside-'));
    try {
      await fs.writeFile(path.join(outside, 'secret.md'), 'outside secret', 'utf-8');
      await fs.symlink(path.join(outside, 'secret.md'), path.join(tmp, 'linked-secret.md'));
      await fs.symlink(outside, path.join(tmp, 'linked-dir'));

      const index = createIndex(tmp);
      const stats = await index.rebuild();

      expect(stats.files).toBe(0);
      await expect(index.search('secret', ['workspace'], 10)).resolves.toEqual([]);
      await expect(index.read('/workspace/linked-secret.md')).rejects.toThrow('escapes mount root');
      await expect(index.write('/workspace/linked-dir/new.md', 'new')).rejects.toThrow('escapes mount root');
    } finally {
      await fs.rm(outside, { recursive: true, force: true });
    }
  });

  it('returns no search results when the snapshot has not been built yet', async () => {
    const index = createIndex(tmp);

    await expect(index.search('missing')).resolves.toEqual([]);
  });

  it('fails clearly when the snapshot is corrupt', async () => {
    await fs.writeFile(path.join(tmp, 'index.json'), '{not-json', 'utf-8');
    const index = createIndex(tmp);

    await expect(index.search('alpha')).rejects.toThrow(/failed to read text index snapshot/);
  });

  it('writes a complete snapshot without leaving temp files behind', async () => {
    await fs.writeFile(path.join(tmp, 'note.md'), 'alpha project note', 'utf-8');
    const index = createIndex(tmp);

    await index.rebuild();

    const entries = await fs.readdir(tmp);
    expect(entries.filter((entry) => entry.endsWith('.tmp'))).toEqual([]);
    const snapshot = JSON.parse(await fs.readFile(path.join(tmp, 'index.json'), 'utf-8'));
    expect(snapshot).toMatchObject({ version: 1 });
    expect(snapshot.files).toHaveLength(1);
  });
});

function createIndex(
  root: string,
  options: { maxIndexedFiles?: number; maxIndexedChunks?: number; maxChunksPerFile?: number } = {},
): TextContextIndex {
  const config = parseConfig({
    version: 1,
    mounts: [
      {
        name: 'workspace',
        type: 'local_folder',
        path: '/workspace',
        root,
        description: 'Workspace mount',
        guidance: 'Use for tests.',
        exclude: ['.git/**', 'node_modules/**'],
        protect: ['protected/**'],
        writeAccess: true,
      },
    ],
  });
  return new TextContextIndex({
    mounts: config.mounts,
    indexPath: path.join(root, 'index.json'),
    ...options,
  });
}

async function writeGeneratedCorpus(root: string, count: number): Promise<void> {
  const topics = ['latency', 'oauth', 'connector', 'index', 'search', 'policy', 'mount', 'tunnel'];
  for (let i = 0; i < count; i += 1) {
    const topic = topics[i % topics.length];
    const shard = i % 17;
    const dir = path.join(root, `project-${String(i % 25).padStart(2, '0')}`, `shard-${shard}`);
    await fs.mkdir(dir, { recursive: true });
    await fs.writeFile(
      path.join(dir, `doc-${String(i).padStart(5, '0')}.md`),
      [
        `# Generated document ${i}`,
        '',
        `This fixture covers needle-${topic} shard-${shard}.`,
        `The deterministic body mentions mvmt search indexing benchmark corpus ${i}.`,
        `Related terms: local-files mounted-context agent-routing ${topic}.`,
        'Filler text keeps each file below one chunk while still resembling a real note.',
      ].join('\n'),
      'utf-8',
    );
  }
}
