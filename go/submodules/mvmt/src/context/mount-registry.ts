import path from 'path';
import { expandHome } from '../config/loader.js';
import { LocalFolderMountConfig } from '../config/schema.js';

export interface RegisteredMount {
  config: LocalFolderMountConfig;
  root: string;
}

export interface ResolvedMountPath {
  mount: RegisteredMount;
  relativePath: string;
  realPath: string;
  virtualPath: string;
}

export class MountRegistry {
  private readonly registered: RegisteredMount[];

  constructor(mounts: LocalFolderMountConfig[]) {
    this.registered = mounts
      .filter((mount) => mount.enabled !== false)
      .map((mount) => ({
        config: mount,
        root: path.resolve(expandHome(mount.root)),
      }))
      .sort((a, b) => b.config.path.length - a.config.path.length);
  }

  mounts(): RegisteredMount[] {
    return this.registered.map((mount) => ({ ...mount }));
  }

  mountNames(): string[] {
    return this.registered.map((mount) => mount.config.name);
  }

  writableMountNames(): string[] {
    return this.registered
      .filter((mount) => mount.config.writeAccess)
      .map((mount) => mount.config.name);
  }

  mountPathForName(name: string): string | undefined {
    return this.registered.find((mount) => mount.config.name === name)?.config.path;
  }

  mountNameForPath(inputPath: string): string | undefined {
    return this.findMount(normalizeVirtualPath(inputPath))?.config.name;
  }

  resolve(inputPath: string): ResolvedMountPath {
    const normalized = normalizeVirtualPath(inputPath);
    const mount = this.findMount(normalized);
    if (!mount) throw new Error(`unknown mount for path: ${normalized}`);
    const relativePath = toVirtualRelative(normalized.slice(mount.config.path.length));
    const realPath = path.resolve(mount.root, relativePath);
    if (!isWithin(mount.root, realPath)) throw new Error(`${normalized} escapes mount root`);
    return {
      mount,
      relativePath,
      realPath,
      virtualPath: relativePath ? joinVirtualPath(mount.config.path, relativePath) : mount.config.path,
    };
  }

  private findMount(normalizedPath: string): RegisteredMount | undefined {
    return this.registered.find((candidate) => pathMatchesMount(normalizedPath, candidate.config.path));
  }
}

export function normalizeVirtualPath(inputPath: string): string {
  const trimmed = inputPath.trim();
  if (!trimmed || trimmed === '/') return '/';
  return `/${stripLeadingSlashes(normalizePathSeparators(trimmed))}`;
}

export function toVirtualRelative(inputPath: string): string {
  return normalizePathSeparators(inputPath).split('/').filter(Boolean).join('/');
}

export function joinVirtualPath(mountPath: string, relativePath: string): string {
  return [stripTrailingSlashes(mountPath), toVirtualRelative(relativePath)].filter(Boolean).join('/');
}

function pathMatchesMount(inputPath: string, mountPath: string): boolean {
  const normalizedMountPath = stripTrailingSlashes(mountPath) || '/';
  return inputPath === normalizedMountPath || inputPath.startsWith(`${normalizedMountPath}/`);
}

function isWithin(root: string, candidate: string): boolean {
  const relative = path.relative(root, candidate);
  return relative === '' || (!relative.startsWith('..') && !path.isAbsolute(relative));
}

export function normalizePathSeparators(value: string): string {
  return value.replaceAll('\\', '/');
}

export function stripLeadingSlashes(value: string): string {
  let start = 0;
  while (start < value.length && value[start] === '/') start += 1;
  return value.slice(start);
}

export function stripTrailingSlashes(value: string): string {
  let end = value.length;
  while (end > 0 && value[end - 1] === '/') end -= 1;
  return value.slice(0, end);
}
