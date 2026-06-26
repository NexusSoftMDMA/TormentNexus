import crypto from 'crypto';
import fsp from 'fs/promises';
import os from 'os';
import path from 'path';
import { LocalFolderMountConfig } from '../config/schema.js';
import {
  MountRegistry,
  normalizePathSeparators,
  ResolvedMountPath,
} from './mount-registry.js';
import {
  LocalFolderStorageProvider,
  StorageProvider,
  StorageProviderFile,
} from './storage-provider.js';

const TEXT_EXTENSIONS = new Set([
  '.md',
  '.markdown',
  '.mdx',
  '.txt',
  '.text',
  '.log',
  '.json',
  '.jsonl',
  '.yaml',
  '.yml',
  '.toml',
  '.csv',
  '.ts',
  '.tsx',
  '.js',
  '.jsx',
  '.mjs',
  '.cjs',
  '.py',
  '.rb',
  '.go',
  '.rs',
  '.java',
  '.c',
  '.h',
  '.cpp',
  '.hpp',
  '.css',
  '.scss',
  '.html',
  '.xml',
  '.sh',
  '.bash',
  '.zsh',
]);

export const MAX_TEXT_BYTES = 2 * 1024 * 1024;
const CHUNK_SIZE = 1600;
const CHUNK_OVERLAP = 200;
const DEFAULT_MAX_INDEXED_FILES = 5_000;
const DEFAULT_MAX_INDEXED_CHUNKS = 20_000;
const DEFAULT_MAX_CHUNKS_PER_FILE = 24;
const DEFAULT_INDEX_IGNORED_SEGMENTS = new Set([
  '.cache',
  '.expo',
  '.git',
  '.gradle',
  '.mypy_cache',
  '.next',
  '.nuxt',
  '.pnpm-store',
  '.pytest_cache',
  '.ruff_cache',
  '.svelte-kit',
  '.tox',
  '.turbo',
  '.venv',
  '.yarn',
  '__pycache__',
  'build',
  'coverage',
  'dist',
  'node_modules',
  'out',
  'target',
  'venv',
]);
const DEFAULT_INDEX_IGNORED_FILENAMES = new Set([
  'bun.lock',
  'bun.lockb',
  'cargo.lock',
  'gemfile.lock',
  'go.sum',
  'package-lock.json',
  'pnpm-lock.yaml',
  'poetry.lock',
  'yarn.lock',
]);

export interface TextContextIndexOptions {
  mounts: LocalFolderMountConfig[];
  indexPath: string;
  maxIndexedFiles?: number;
  maxIndexedChunks?: number;
  maxChunksPerFile?: number;
}

export interface TextIndexStats {
  files: number;
  chunks: number;
  truncated?: boolean;
}

export interface TextSearchResult {
  mount: string;
  path: string;
  chunk_id: string;
  score: number;
  snippet: string;
  hash: string;
  mtime_ms: number;
}

export interface TextListEntry {
  mount: string;
  path: string;
  type: 'file' | 'directory';
  size: number;
  mtime_ms: number;
  description?: string;
  guidance?: string;
  write_access?: boolean;
}

export interface TextReadResult {
  mount: string;
  path: string;
  content: string;
  mime_type: 'text/plain';
  size: number;
  hash: string;
  mtime_ms: number;
}

interface IndexedFile {
  mount: string;
  path: string;
  hash: string;
  size: number;
  mtime_ms: number;
}

interface IndexedChunk {
  mount: string;
  path: string;
  chunk_id: string;
  text: string;
  hash: string;
  mtime_ms: number;
}

interface TextIndexSnapshot {
  version: 1;
  indexed_at: string;
  files: IndexedFile[];
  chunks: IndexedChunk[];
}

export function defaultTextIndexPath(configPath: string): string {
  return path.join(path.dirname(configPath), 'text-index.json');
}

export class TextContextIndex {
  private registry: MountRegistry;
  private providers: Map<string, StorageProvider>;

  constructor(private readonly options: TextContextIndexOptions) {
    this.registry = new MountRegistry(options.mounts);
    this.providers = this.providersForRegistry(this.registry);
  }

  updateMounts(mounts: LocalFolderMountConfig[]): void {
    this.options.mounts = mounts;
    this.registry = new MountRegistry(mounts);
    this.providers = this.providersForRegistry(this.registry);
  }

  private providersForRegistry(registry: MountRegistry): Map<string, StorageProvider> {
    return new Map(
      registry.mounts().map((mount) => [
        mount.config.name,
        new LocalFolderStorageProvider(mount, {
          isTextPath,
          isIndexIgnoredPath,
          maxTextBytes: MAX_TEXT_BYTES,
        }),
      ]),
    );
  }

  mountNames(): string[] {
    return this.registry.mountNames();
  }

  writableMountNames(): string[] {
    return this.registry.writableMountNames();
  }

  mountNameForPath(inputPath: string): string | undefined {
    return this.registry.mountNameForPath(inputPath);
  }

  mountPathForName(name: string): string | undefined {
    return this.registry.mountPathForName(name);
  }

  async rebuild(): Promise<TextIndexStats> {
    const files: IndexedFile[] = [];
    const chunks: IndexedChunk[] = [];
    let truncated = false;
    const maxFiles = this.options.maxIndexedFiles ?? DEFAULT_MAX_INDEXED_FILES;
    const maxChunks = this.options.maxIndexedChunks ?? DEFAULT_MAX_INDEXED_CHUNKS;
    const maxChunksPerFile = this.options.maxChunksPerFile ?? DEFAULT_MAX_CHUNKS_PER_FILE;

    mountLoop:
    for (const mount of this.registry.mounts()) {
      const provider = this.providerForMount(mount.config.name);
      for await (const file of provider.walkTextFiles()) {
        if (files.length >= maxFiles || chunks.length >= maxChunks) {
          truncated = true;
          break mountLoop;
        }

        const indexed = this.indexFile(file, maxChunksPerFile);
        if (indexed.truncated) truncated = true;
        if (files.length + 1 > maxFiles || chunks.length + indexed.chunks.length > maxChunks) {
          truncated = true;
          break mountLoop;
        }

        files.push(indexed.file);
        chunks.push(...indexed.chunks);
      }
    }

    await this.writeSnapshot({ version: 1, indexed_at: new Date().toISOString(), files, chunks });
    return { files: files.length, chunks: chunks.length, ...(truncated ? { truncated } : {}) };
  }

  async search(query: string, mountNames?: string[], limit = 8): Promise<TextSearchResult[]> {
    const terms = tokenize(query);
    if (terms.length === 0) return [];

    const allowedMounts = new Set(mountNames ?? this.mountNames());
    const snapshot = await this.readSnapshot();
    const ranked = snapshot.chunks
      .filter((chunk) => allowedMounts.has(chunk.mount))
      .map((chunk) => ({ chunk, score: scoreText(chunk.text, terms) }))
      .filter((entry) => entry.score > 0)
      .sort((a, b) => b.score - a.score || a.chunk.path.localeCompare(b.chunk.path))
      .slice(0, Math.max(1, Math.min(20, limit)));

    return ranked.map(({ chunk, score }) => ({
      mount: chunk.mount,
      path: chunk.path,
      chunk_id: chunk.chunk_id,
      score,
      snippet: snippetFor(chunk.text, terms),
      hash: chunk.hash,
      mtime_ms: chunk.mtime_ms,
    }));
  }

  async list(inputPath = '/'): Promise<TextListEntry[]> {
    if (inputPath === '/' || inputPath.trim() === '') {
      return Promise.all(this.registry.mounts().map(async (mount) => {
        let type: 'file' | 'directory' = 'directory';
        let size = 0;
        let mtimeMs = 0;
        try {
          const stat = await fsp.stat(mount.root);
          type = stat.isFile() ? 'file' : 'directory';
          size = stat.isFile() ? stat.size : 0;
          mtimeMs = stat.mtimeMs;
        } catch {
          // Keep root listing resilient; direct access will report the real error.
        }
        return {
          mount: mount.config.name,
          path: mount.config.path,
          type,
          size,
          mtime_ms: mtimeMs,
          description: mount.config.description,
          guidance: mount.config.guidance,
          write_access: mount.config.writeAccess,
        };
      }));
    }

    const resolved = this.resolvePath(inputPath);
    return (await this.providerForResolved(resolved).list(resolved.relativePath)).map((entry) => ({
      mount: entry.mount,
      path: entry.path,
      type: entry.type,
      size: entry.size,
      mtime_ms: entry.mtimeMs,
    }));
  }

  async read(inputPath: string): Promise<TextReadResult> {
    const resolved = this.resolvePath(inputPath);
    return this.toReadResult(await this.providerForResolved(resolved).read(resolved.relativePath));
  }

  async write(inputPath: string, content: string, expectedHash?: string): Promise<TextReadResult> {
    const resolved = this.resolvePath(inputPath);
    const provider = this.providerForResolved(resolved);
    if (expectedHash) {
      let current: TextReadResult;
      try {
        current = await this.read(inputPath);
      } catch (err) {
        if (isNodeError(err) && err.code === 'ENOENT') {
          throw new Error(`hash mismatch for ${resolved.virtualPath}`);
        }
        throw err;
      }
      if (current.hash !== expectedHash) {
        throw new Error(`hash mismatch for ${resolved.virtualPath}`);
      }
    }
    const read = this.toReadResult(await provider.write(resolved.relativePath, content));
    await this.rebuild();
    return read;
  }

  async remove(inputPath: string): Promise<{ mount: string; path: string; removed: true }> {
    const resolved = this.resolvePath(inputPath);
    await this.providerForResolved(resolved).remove(resolved.relativePath);
    await this.rebuild();
    return { mount: resolved.mount.config.name, path: resolved.virtualPath, removed: true };
  }

  private indexFile(
    file: StorageProviderFile,
    maxChunksPerFile: number,
  ): { file: IndexedFile; chunks: IndexedChunk[]; truncated: boolean } {
    const hash = sha256(file.content);
    const { chunks, truncated } = chunkText(file.mount, file.path, file.content, hash, file.mtimeMs, maxChunksPerFile);
    return {
      file: {
        mount: file.mount,
        path: file.path,
        hash,
        size: file.size,
        mtime_ms: file.mtimeMs,
      },
      chunks,
      truncated,
    };
  }

  private resolvePath(inputPath: string): ResolvedMountPath {
    return this.registry.resolve(inputPath);
  }

  private providerForResolved(resolved: ResolvedMountPath): StorageProvider {
    return this.providerForMount(resolved.mount.config.name);
  }

  private providerForMount(mountName: string): StorageProvider {
    const provider = this.providers.get(mountName);
    if (!provider) throw new Error(`no provider registered for mount: ${mountName}`);
    return provider;
  }

  private toReadResult(file: StorageProviderFile): TextReadResult {
    return {
      mount: file.mount,
      path: file.path,
      content: file.content,
      mime_type: 'text/plain',
      size: file.size,
      hash: sha256(file.content),
      mtime_ms: file.mtimeMs,
    };
  }

  private async readSnapshot(): Promise<TextIndexSnapshot> {
    try {
      return JSON.parse(await fsp.readFile(this.options.indexPath, 'utf-8')) as TextIndexSnapshot;
    } catch (err) {
      if (isNodeError(err) && err.code === 'ENOENT') {
        return emptySnapshot();
      }
      throw new Error(`failed to read text index snapshot: ${err instanceof Error ? err.message : 'unknown error'}`);
    }
  }

  private async writeSnapshot(snapshot: TextIndexSnapshot): Promise<void> {
    const indexDir = path.dirname(this.options.indexPath);
    const tempPath = path.join(
      indexDir,
      `.${path.basename(this.options.indexPath)}.${process.pid}.${Date.now()}.tmp`,
    );
    await fsp.mkdir(indexDir, { recursive: true });
    try {
      await fsp.writeFile(tempPath, JSON.stringify(snapshot, null, 2), { encoding: 'utf-8', mode: 0o600 });
      if (process.platform !== 'win32') {
        await fsp.chmod(tempPath, 0o600);
      }
      await fsp.rename(tempPath, this.options.indexPath);
    } catch (err) {
      await fsp.rm(tempPath, { force: true });
      throw err;
    }
    if (process.platform !== 'win32') {
      await fsp.chmod(this.options.indexPath, 0o600);
    }
  }
}

function chunkText(
  mountName: string,
  virtualPath: string,
  text: string,
  hash: string,
  mtimeMs: number,
  maxChunks: number,
): { chunks: IndexedChunk[]; truncated: boolean } {
  const chunks: IndexedChunk[] = [];
  let truncated = false;
  for (let offset = 0; offset < text.length; offset += CHUNK_SIZE - CHUNK_OVERLAP) {
    const chunk = text.slice(offset, offset + CHUNK_SIZE);
    if (chunk.trim().length === 0) continue;
    if (chunks.length >= maxChunks) {
      truncated = true;
      break;
    }
    chunks.push({
      mount: mountName,
      path: virtualPath,
      chunk_id: `${virtualPath}#${offset}`,
      text: chunk,
      hash,
      mtime_ms: mtimeMs,
    });
  }
  return { chunks, truncated };
}

export function isTextPath(inputPath: string): boolean {
  return TEXT_EXTENSIONS.has(path.extname(inputPath).toLowerCase());
}

function isIndexIgnoredPath(inputPath: string): boolean {
  const segments = normalizePathSeparators(inputPath)
    .split('/')
    .filter(Boolean);
  const leaf = segments.at(-1)?.toLowerCase();
  return segments.some((segment) => DEFAULT_INDEX_IGNORED_SEGMENTS.has(segment))
    || Boolean(leaf && DEFAULT_INDEX_IGNORED_FILENAMES.has(leaf));
}

function tokenize(query: string): string[] {
  return query
    .toLowerCase()
    .split(/[^a-z0-9_]+/i)
    .filter((term) => term.length > 1);
}

function scoreText(text: string, terms: string[]): number {
  const lower = text.toLowerCase();
  return terms.reduce((score, term) => {
    const matches = lower.match(new RegExp(escapeRegExp(term), 'g'));
    return score + (matches?.length ?? 0);
  }, 0);
}

function snippetFor(text: string, terms: string[]): string {
  const lower = text.toLowerCase();
  const first = terms
    .map((term) => lower.indexOf(term))
    .filter((index) => index >= 0)
    .sort((a, b) => a - b)[0] ?? 0;
  const start = Math.max(0, first - 160);
  return text.slice(start, start + 500).replace(/\s+/g, ' ').trim();
}

function sha256(value: string): string {
  return crypto.createHash('sha256').update(value, 'utf-8').digest('hex');
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

export function defaultTextIndexRoot(): string {
  return path.join(os.homedir(), '.mvmt');
}

function emptySnapshot(): TextIndexSnapshot {
  return { version: 1, indexed_at: new Date(0).toISOString(), files: [], chunks: [] };
}

function isNodeError(err: unknown): err is NodeJS.ErrnoException {
  return err instanceof Error && 'code' in err;
}
