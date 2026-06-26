import { randomUUID } from 'crypto';
import { createReadStream, createWriteStream, ReadStream } from 'fs';
import fsp from 'fs/promises';
import path from 'path';
import { pipeline } from 'stream/promises';
import { joinVirtualPath, normalizePathSeparators, RegisteredMount, toVirtualRelative } from './mount-registry.js';
import { isGloballyDeniedPath, matchesConfiguredOrGlobalPattern } from './path-policy.js';

export interface StorageProviderFile {
  mount: string;
  path: string;
  relativePath: string;
  content: string;
  size: number;
  mtimeMs: number;
}

export interface StorageProviderEntry {
  mount: string;
  path: string;
  relativePath: string;
  type: 'file' | 'directory';
  size: number;
  mtimeMs: number;
}

export interface StorageProviderStat {
  mount: string;
  path: string;
  relativePath: string;
  type: 'file' | 'directory';
  size: number;
  mtimeMs: number;
}

export interface StorageProviderReadStream {
  file: StorageProviderStat & { type: 'file' };
  stream: ReadStream;
}

export interface StorageProvider {
  readonly mount: RegisteredMount;
  list(relativePath?: string): Promise<StorageProviderEntry[]>;
  exists(relativePath: string): Promise<boolean>;
  stat(relativePath: string): Promise<StorageProviderStat>;
  openReadStream(relativePath: string, range?: { start?: number; end?: number }): Promise<StorageProviderReadStream>;
  read(relativePath: string): Promise<StorageProviderFile>;
  write(relativePath: string, content: string): Promise<StorageProviderFile>;
  writeStream(relativePath: string, input: NodeJS.ReadableStream): Promise<StorageProviderStat & { type: 'file' }>;
  remove(relativePath: string): Promise<void>;
  walkTextFiles(): AsyncIterable<StorageProviderFile>;
}

export interface LocalFolderStorageProviderOptions {
  isTextPath(path: string): boolean;
  maxTextBytes: number;
  isIndexIgnoredPath?(path: string): boolean;
}

export class LocalFolderStorageProvider implements StorageProvider {
  constructor(
    readonly mount: RegisteredMount,
    private readonly options: LocalFolderStorageProviderOptions,
  ) {}

  async list(relativePath = ''): Promise<StorageProviderEntry[]> {
    const target = this.resolve(relativePath);
    await this.assertRealPathWithinMount(target);
    const stat = await fsp.stat(target.realPath);
    if (!stat.isDirectory()) {
      return [{
        mount: this.mount.config.name,
        path: target.virtualPath,
        relativePath: target.relativePath,
        type: 'file',
        size: stat.size,
        mtimeMs: stat.mtimeMs,
      }];
    }

    const entries = await fsp.readdir(target.realPath, { withFileTypes: true });
    const result: StorageProviderEntry[] = [];
    for (const entry of entries) {
      const childRelative = joinRelative(target.relativePath, entry.name);
      if (this.isExcluded(childRelative)) continue;
      const realPath = path.join(target.realPath, entry.name);
      if (!await this.realPathWithinMount(realPath)) continue;
      const itemStat = await fsp.stat(realPath);
      if (!itemStat.isDirectory() && !itemStat.isFile()) continue;
      result.push({
        mount: this.mount.config.name,
        path: joinVirtualPath(this.mount.config.path, childRelative),
        relativePath: childRelative,
        type: itemStat.isDirectory() ? 'directory' : 'file',
        size: itemStat.isDirectory() ? 0 : itemStat.size,
        mtimeMs: itemStat.mtimeMs,
      });
    }
    return result.sort((a, b) => a.path.localeCompare(b.path));
  }

  async exists(relativePath: string): Promise<boolean> {
    try {
      await fsp.access(this.resolve(relativePath).realPath);
      return true;
    } catch {
      return false;
    }
  }

  async stat(relativePath: string): Promise<StorageProviderStat> {
    const target = this.resolve(relativePath);
    await this.assertRealPathWithinMount(target);
    const stat = await fsp.stat(target.realPath);
    if (!stat.isDirectory() && !stat.isFile()) throw new Error(`${target.virtualPath} is not a file or directory`);
    return {
      mount: this.mount.config.name,
      path: target.virtualPath,
      relativePath: target.relativePath,
      type: stat.isDirectory() ? 'directory' : 'file',
      size: stat.isDirectory() ? 0 : stat.size,
      mtimeMs: stat.mtimeMs,
    };
  }

  async openReadStream(
    relativePath: string,
    range: { start?: number; end?: number } = {},
  ): Promise<StorageProviderReadStream> {
    const file = await this.stat(relativePath);
    if (file.type !== 'file') throw new Error(`${file.path} is not a file`);
    const fileStat = file as StorageProviderStat & { type: 'file' };
    const target = this.resolve(relativePath);
    const stream = createReadStream(target.realPath, {
      ...(range.start === undefined ? {} : { start: range.start }),
      ...(range.end === undefined ? {} : { end: range.end }),
    });
    return { file: fileStat, stream };
  }

  async read(relativePath: string): Promise<StorageProviderFile> {
    const target = this.resolve(relativePath);
    this.assertTextPath(target.virtualPath);
    await this.assertRealPathWithinMount(target);
    const stat = await fsp.stat(target.realPath);
    if (!stat.isFile()) throw new Error(`${target.virtualPath} is not a file`);
    if (stat.size > this.options.maxTextBytes) throw new Error(`${target.virtualPath} is too large to read as text`);
    const content = await fsp.readFile(target.realPath, 'utf-8');
    return {
      mount: this.mount.config.name,
      path: target.virtualPath,
      relativePath: target.relativePath,
      content,
      size: stat.size,
      mtimeMs: stat.mtimeMs,
    };
  }

  async write(relativePath: string, content: string): Promise<StorageProviderFile> {
    const target = this.resolve(relativePath);
    this.assertWritable(target);
    this.assertTextPath(target.virtualPath);
    if (Buffer.byteLength(content, 'utf-8') > this.options.maxTextBytes) {
      throw new Error(`${target.virtualPath} is too large to write as text`);
    }
    await this.assertRealPathWithinMount(target, { allowMissingLeaf: true });
    await fsp.mkdir(path.dirname(target.realPath), { recursive: true });
    await fsp.writeFile(target.realPath, content, 'utf-8');
    return this.read(relativePath);
  }

  async writeStream(relativePath: string, input: NodeJS.ReadableStream): Promise<StorageProviderStat & { type: 'file' }> {
    const target = this.resolve(relativePath);
    this.assertWritable(target);
    await this.assertRealPathWithinMount(target, { allowMissingLeaf: true });
    await fsp.mkdir(path.dirname(target.realPath), { recursive: true });
    const parentRealPath = await fsp.realpath(path.dirname(target.realPath));
    const rootRealPath = await fsp.realpath(this.mount.root);
    if (!isWithin(rootRealPath, parentRealPath)) {
      throw new Error(`${target.virtualPath} escapes mount root`);
    }
    const tempPath = path.join(parentRealPath, `.mvmt-write-${process.pid}-${Date.now()}-${randomUUID()}.tmp`);
    try {
      await pipeline(input, createWriteStream(tempPath, { flags: 'wx', mode: 0o600 }));
      await fsp.rename(tempPath, target.realPath);
      const written = await this.stat(relativePath);
      if (written.type !== 'file') throw new Error(`${target.virtualPath} is not a file`);
      return written as StorageProviderStat & { type: 'file' };
    } finally {
      await fsp.unlink(tempPath).catch(() => undefined);
    }
  }

  async remove(relativePath: string): Promise<void> {
    const target = this.resolve(relativePath);
    this.assertWritable(target);
    await this.assertRealPathWithinMount(target);
    const stat = await fsp.stat(target.realPath);
    if (!stat.isFile()) throw new Error(`${target.virtualPath} is not a file`);
    await fsp.unlink(target.realPath);
  }

  async *walkTextFiles(): AsyncIterable<StorageProviderFile> {
    const root = this.resolve('');
    try {
      const stat = await fsp.stat(root.realPath);
      if (stat.isFile()) {
        yield await this.read('');
        return;
      }
    } catch {
      return;
    }
    yield* this.walkDirectory('');
  }

  private async *walkDirectory(relativePath: string): AsyncIterable<StorageProviderFile> {
    let entries;
    try {
      entries = await fsp.readdir(this.realPath(relativePath), { withFileTypes: true });
    } catch {
      return;
    }

    for (const entry of entries) {
      const childRelative = joinRelative(relativePath, entry.name);
      if (this.options.isIndexIgnoredPath?.(childRelative)) continue;
      if (this.isExcluded(childRelative)) continue;
      if (entry.isDirectory()) {
        yield* this.walkDirectory(childRelative);
        continue;
      }
      if (!entry.isFile() || !this.options.isTextPath(entry.name)) continue;

      try {
        yield await this.read(childRelative);
      } catch {
        // Unsupported, unreadable, and over-size files are skipped by indexing.
      }
    }
  }

  private resolve(relativePath: string): ResolvedProviderPath {
    const normalized = toVirtualRelative(relativePath);
    const realPath = this.realPath(normalized);
    const policyRelativePath = normalized || path.basename(realPath);
    if (this.isExcluded(policyRelativePath)) {
      throw new Error(`${joinVirtualPath(this.mount.config.path, normalized)} is excluded`);
    }
    if (!isWithin(this.mount.root, realPath)) {
      throw new Error(`${joinVirtualPath(this.mount.config.path, normalized)} escapes mount root`);
    }
    if (isGloballyDeniedPath(policyRelativePath, realPath)) {
      throw new Error(`${joinVirtualPath(this.mount.config.path, normalized)} is globally denied`);
    }
    return {
      relativePath: normalized,
      realPath,
      virtualPath: normalized ? joinVirtualPath(this.mount.config.path, normalized) : this.mount.config.path,
    };
  }

  private realPath(relativePath: string): string {
    return path.resolve(this.mount.root, toVirtualRelative(relativePath));
  }

  private async assertRealPathWithinMount(
    target: ResolvedProviderPath,
    options: { allowMissingLeaf?: boolean } = {},
  ): Promise<void> {
    const rootRealPath = await fsp.realpath(this.mount.root);
    try {
      const targetRealPath = await fsp.realpath(target.realPath);
      if (!isWithin(rootRealPath, targetRealPath)) {
        throw new Error(`${target.virtualPath} escapes mount root`);
      }
      return;
    } catch (err) {
      if (!options.allowMissingLeaf || !isMissingPathError(err)) throw err;
    }

    const ancestorRealPath = await this.nearestExistingAncestorRealPath(target.realPath);
    if (!isWithin(rootRealPath, ancestorRealPath)) {
      throw new Error(`${target.virtualPath} escapes mount root`);
    }
  }

  private async realPathWithinMount(realPath: string): Promise<boolean> {
    try {
      const rootRealPath = await fsp.realpath(this.mount.root);
      const targetRealPath = await fsp.realpath(realPath);
      return isWithin(rootRealPath, targetRealPath);
    } catch {
      return false;
    }
  }

  private async nearestExistingAncestorRealPath(realPath: string): Promise<string> {
    let candidate = path.dirname(realPath);
    while (true) {
      try {
        return await fsp.realpath(candidate);
      } catch (err) {
        if (!isMissingPathError(err)) throw err;
        const parent = path.dirname(candidate);
        if (parent === candidate) throw err;
        candidate = parent;
      }
    }
  }

  private assertWritable(target: ResolvedProviderPath): void {
    if (!this.mount.config.writeAccess) {
      throw new Error(`${this.mount.config.name} is read-only`);
    }
    if (this.isProtected(target.relativePath)) {
      throw new Error(`${target.virtualPath} is protected`);
    }
  }

  private assertTextPath(inputPath: string): void {
    if (!this.options.isTextPath(inputPath)) {
      throw new Error(`${inputPath} is not a supported text file`);
    }
  }

  private isExcluded(relativePath: string): boolean {
    return matchesConfiguredOrGlobalPattern(relativePath, this.mount.config.exclude);
  }

  private isProtected(relativePath: string): boolean {
    return matchesConfiguredOrGlobalPattern(relativePath, this.mount.config.protect);
  }
}

interface ResolvedProviderPath {
  relativePath: string;
  realPath: string;
  virtualPath: string;
}

function joinRelative(base: string, leaf: string): string {
  return normalizePathSeparators([base, leaf].filter(Boolean).join('/'));
}

function isWithin(root: string, candidate: string): boolean {
  const relative = path.relative(root, candidate);
  return relative === '' || (!relative.startsWith('..') && !path.isAbsolute(relative));
}

function isMissingPathError(err: unknown): boolean {
  return err instanceof Error && 'code' in err && err.code === 'ENOENT';
}
