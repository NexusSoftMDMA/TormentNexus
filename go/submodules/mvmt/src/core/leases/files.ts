import fsp from 'fs/promises';
import type { Stats } from 'fs';
import path from 'path';
import { LocalFolderMountConfig } from '../../config/schema.js';
import { MountRegistry, normalizePathSeparators, toVirtualRelative } from '../../context/mount-registry.js';
import { isGloballyDeniedPath, matchesConfiguredOrGlobalPattern, matchesPathPatterns } from '../../context/path-policy.js';
import { LeaseRecord, LeaseResource, leaseResources } from './store.js';

export interface LeaseFileTarget {
  mountName: string;
  leasePath: string;
  virtualPath: string;
  leaseRelativePath: string;
  realPath: string;
  filename: string;
  size: number;
  mtimeMs: number;
}

export interface LeaseUploadTarget {
  mountName: string;
  leasePath: string;
  virtualPath: string;
  leaseRelativePath: string;
  realPath: string;
  parentRealPath: string;
  filename: string;
}

export interface LeaseDirectoryEntry {
  name: string;
  path: string;
  type: 'directory' | 'file';
  size: number;
  mtimeMs: number;
}

export interface LeaseDirectoryListing {
  label: string;
  leaseId: string;
  path: string;
  expiresAt?: string;
  canUpload?: boolean;
  entries: LeaseDirectoryEntry[];
}

export async function resolveLeaseFileTarget(
  mounts: readonly LocalFolderMountConfig[],
  lease: LeaseRecord,
  requestPath = '',
): Promise<LeaseFileTarget> {
  const target = await resolveLeaseTarget(mounts, lease, requestPath);
  if (target.stat.isDirectory()) throw new Error(`${target.virtualPath} is a directory`);
  if (!target.stat.isFile()) throw new Error(`${target.virtualPath} is not a file`);
  return {
    mountName: target.mountName,
    leasePath: target.leasePath,
    virtualPath: target.virtualPath,
    leaseRelativePath: target.leaseRelativePath,
    realPath: target.realPath,
    filename: path.basename(target.virtualPath),
    size: target.stat.size,
    mtimeMs: target.stat.mtimeMs,
  };
}

export async function listLeaseDirectory(
  mounts: readonly LocalFolderMountConfig[],
  lease: LeaseRecord,
  requestPath = '',
): Promise<LeaseDirectoryListing> {
  const leaseRelativePath = normalizeLeaseRelativePath(requestPath);
  if (usesResourceNamespace(lease) && !leaseRelativePath) {
    return listLeaseResourceRoot(mounts, lease);
  }
  const target = await resolveLeaseTarget(mounts, lease, requestPath);
  if (!target.stat.isDirectory()) throw new Error(`${target.virtualPath} is not a directory`);
  const dirents = await fsp.readdir(target.realPath, { withFileTypes: true });
  const entries: LeaseDirectoryEntry[] = [];
  for (const dirent of dirents) {
    const childPath = joinLeaseRelativePath(target.leaseRelativePath, dirent.name);
    try {
      const child = await resolveLeaseTarget(mounts, lease, childPath);
      if (!child.stat.isDirectory() && !child.stat.isFile()) continue;
      entries.push({
        name: dirent.name,
        path: `/${child.leaseRelativePath}`,
        type: child.stat.isDirectory() ? 'directory' : 'file',
        size: child.stat.isDirectory() ? 0 : child.stat.size,
        mtimeMs: child.stat.mtimeMs,
      });
    } catch {
      // Hidden, denied, broken, or escaping children stay invisible in listings.
    }
  }
  entries.sort((a, b) => {
    if (a.type !== b.type) return a.type === 'directory' ? -1 : 1;
    return a.name.localeCompare(b.name);
  });
  return {
    label: lease.label,
    leaseId: lease.id,
    path: target.leaseRelativePath ? `/${target.leaseRelativePath}` : '/',
    ...(lease.expiresAt ? { expiresAt: lease.expiresAt } : {}),
    entries,
  };
}

export async function resolveLeaseUploadTarget(
  mounts: readonly LocalFolderMountConfig[],
  lease: LeaseRecord,
  requestPath: string,
  options: { allowOverwrite?: boolean; suffixOnCollision?: boolean } = {},
): Promise<LeaseUploadTarget> {
  const leaseRelativePath = normalizeLeaseRelativePath(requestPath);
  if (!leaseRelativePath) throw new Error('upload path is required');
  const resource = uploadResourceForRequest(lease, leaseRelativePath);
  const sourceRelativePath = sourceRelativePathForResource(resource, leaseRelativePath);
  if (!sourceRelativePath && !(options.allowOverwrite && resource.type === 'file')) {
    throw new Error('upload path is required');
  }

  const registry = new MountRegistry([...mounts]);
  const leaseRoot = registry.resolve(resource.sourcePath);
  const leaseRootStat = await fsp.stat(leaseRoot.realPath);
  if (!leaseRoot.mount.config.writeAccess) throw new Error(`${resource.path} is read-only`);

  if (leaseRootStat.isFile()) {
    if (!options.allowOverwrite || resource.type !== 'file' || sourceRelativePath) {
      throw new Error(`${resource.path} is not a directory`);
    }
    const policyPath = leaseRoot.relativePath || path.basename(leaseRoot.realPath);
    if (matchesConfiguredOrGlobalPattern(policyPath, leaseRoot.mount.config.exclude)) {
      throw new Error(`${leaseRoot.virtualPath} is excluded`);
    }
    if (matchesPathPatterns(policyPath, leaseRoot.mount.config.protect)) {
      throw new Error(`${leaseRoot.virtualPath} is protected`);
    }
    const targetRealPath = await fsp.realpath(leaseRoot.realPath);
    if (isGloballyDeniedPath(policyPath, targetRealPath)) {
      throw new Error(`${leaseRoot.virtualPath} is globally denied`);
    }
    return {
      mountName: leaseRoot.mount.config.name,
      leasePath: resource.path,
      virtualPath: leaseRoot.virtualPath,
      leaseRelativePath,
      realPath: targetRealPath,
      parentRealPath: await fsp.realpath(path.dirname(leaseRoot.realPath)),
      filename: path.basename(leaseRoot.realPath),
    };
  }

  if (!leaseRootStat.isDirectory()) throw new Error(`${resource.path} is not a directory`);

  const leaseRootRealPath = await fsp.realpath(leaseRoot.realPath);
  const virtualPath = `${stripTrailingSlash(resource.sourcePath)}/${sourceRelativePath}`;
  const resolved = registry.resolve(virtualPath);
  const filename = path.basename(resolved.realPath);
  if (!filename || filename === '.' || filename === '..') throw new Error('upload filename is required');

  const policyPath = resolved.relativePath || filename;
  if (matchesConfiguredOrGlobalPattern(policyPath, resolved.mount.config.exclude)) {
    throw new Error(`${resolved.virtualPath} is excluded`);
  }
  if (matchesPathPatterns(policyPath, resolved.mount.config.protect)) {
    throw new Error(`${resolved.virtualPath} is protected`);
  }

  const parentRealPath = await fsp.realpath(path.dirname(resolved.realPath));
  if (!isWithin(leaseRootRealPath, parentRealPath)) {
    throw new Error(`${resolved.virtualPath} escapes lease root`);
  }
  let resolvedFilename = filename;
  let targetRealPath = path.resolve(parentRealPath, resolvedFilename);
  if (!isWithin(leaseRootRealPath, targetRealPath)) {
    throw new Error(`${resolved.virtualPath} escapes lease root`);
  }
  if (isGloballyDeniedPath(policyPath, targetRealPath)) {
    throw new Error(`${resolved.virtualPath} is globally denied`);
  }
  try {
    const existing = await fsp.lstat(targetRealPath);
    if (existing.isDirectory()) throw new Error(`${resolved.virtualPath} is a directory`);
    if (options.suffixOnCollision) {
      const next = await findNextAvailableFilename(parentRealPath, filename);
      resolvedFilename = next;
      targetRealPath = path.resolve(parentRealPath, resolvedFilename);
    } else if (!options.allowOverwrite) {
      throw new Error(`${resolved.virtualPath} already exists`);
    }
  } catch (err) {
    if ((err as NodeJS.ErrnoException).code !== 'ENOENT') throw err;
  }

  const finalLeaseRelative = resolvedFilename === filename
    ? leaseRelativePath
    : path.join(path.dirname(leaseRelativePath), resolvedFilename);

  return {
    mountName: resolved.mount.config.name,
    leasePath: resource.path,
    virtualPath: resolved.virtualPath,
    leaseRelativePath: finalLeaseRelative,
    realPath: targetRealPath,
    parentRealPath,
    filename: resolvedFilename,
  };
}

async function findNextAvailableFilename(parentDir: string, filename: string): Promise<string> {
  const ext = path.extname(filename);
  const base = ext ? filename.slice(0, -ext.length) : filename;
  for (let i = 2; i < 10_000; i += 1) {
    const candidate = `${base} (${i})${ext}`;
    try {
      await fsp.lstat(path.join(parentDir, candidate));
    } catch (err) {
      if ((err as NodeJS.ErrnoException).code === 'ENOENT') return candidate;
      throw err;
    }
  }
  throw new Error(`could not find a free filename for ${filename}`);
}

async function resolveLeaseTarget(
  mounts: readonly LocalFolderMountConfig[],
  lease: LeaseRecord,
  requestPath: string,
): Promise<{
  mountName: string;
  leasePath: string;
  virtualPath: string;
  leaseRelativePath: string;
  realPath: string;
  stat: Stats;
}> {
  const resources = leaseResources(lease);
  if (usesResourceNamespace(lease)) {
    const leaseRelativePath = normalizeLeaseRelativePath(requestPath);
    const { resource, resourceRelativePath } = resourceForRequest(resources, leaseRelativePath);
    return resolveLeaseSourceTarget(mounts, resource.sourcePath, resourceRelativePath, {
      leasePath: resource.path,
      leaseRelativePath,
    });
  }

  return resolveLeaseSourceTarget(mounts, resources[0]!.sourcePath, requestPath, {
    leasePath: lease.path,
    leaseRelativePath: normalizeLeaseRelativePath(requestPath),
  });
}

async function resolveLeaseSourceTarget(
  mounts: readonly LocalFolderMountConfig[],
  sourcePath: string,
  requestPath: string,
  output: { leasePath: string; leaseRelativePath: string },
): Promise<{
  mountName: string;
  leasePath: string;
  virtualPath: string;
  leaseRelativePath: string;
  realPath: string;
  stat: Stats;
}> {
  const registry = new MountRegistry([...mounts]);
  const leaseRoot = registry.resolve(sourcePath);
  const leaseRootStat = await fsp.stat(leaseRoot.realPath);
  const leaseRelativePath = normalizeLeaseRelativePath(requestPath);
  if (leaseRootStat.isFile()) {
    if (leaseRelativePath) throw new Error(`${output.leasePath} is a file`);
    const policyPath = leaseRoot.relativePath || path.basename(leaseRoot.realPath);
    if (matchesConfiguredOrGlobalPattern(policyPath, leaseRoot.mount.config.exclude)) {
      throw new Error(`${leaseRoot.virtualPath} is excluded`);
    }
    const targetRealPath = await fsp.realpath(leaseRoot.realPath);
    if (isGloballyDeniedPath(policyPath, targetRealPath)) {
      throw new Error(`${leaseRoot.virtualPath} is globally denied`);
    }
    return {
      mountName: leaseRoot.mount.config.name,
      leasePath: output.leasePath,
      virtualPath: leaseRoot.virtualPath,
      leaseRelativePath: output.leaseRelativePath,
      realPath: targetRealPath,
      stat: await fsp.stat(targetRealPath),
    };
  }
  if (!leaseRootStat.isDirectory()) throw new Error(`${output.leasePath} is not a file or directory`);

  const leaseRootRealPath = await fsp.realpath(leaseRoot.realPath);
  const virtualPath = leaseRelativePath ? `${stripTrailingSlash(sourcePath)}/${leaseRelativePath}` : stripTrailingSlash(sourcePath);
  const resolved = registry.resolve(virtualPath);
  const policyPath = resolved.relativePath || path.basename(resolved.realPath);
  if (matchesConfiguredOrGlobalPattern(policyPath, resolved.mount.config.exclude)) {
    throw new Error(`${resolved.virtualPath} is excluded`);
  }

  const targetRealPath = await fsp.realpath(resolved.realPath);
  if (!isWithin(leaseRootRealPath, targetRealPath)) {
    throw new Error(`${resolved.virtualPath} escapes lease root`);
  }
  if (isGloballyDeniedPath(policyPath, targetRealPath)) {
    throw new Error(`${resolved.virtualPath} is globally denied`);
  }

  return {
    mountName: resolved.mount.config.name,
    leasePath: output.leasePath,
    virtualPath: resolved.virtualPath,
    leaseRelativePath: output.leaseRelativePath,
    realPath: targetRealPath,
    stat: await fsp.stat(targetRealPath),
  };
}

async function listLeaseResourceRoot(
  mounts: readonly LocalFolderMountConfig[],
  lease: LeaseRecord,
): Promise<LeaseDirectoryListing> {
  const entries: LeaseDirectoryEntry[] = [];
  for (const resource of leaseResources(lease)) {
    try {
      const target = await resolveLeaseSourceTarget(mounts, resource.sourcePath, '', {
        leasePath: resource.path,
        leaseRelativePath: toVirtualRelative(resource.path),
      });
      const type = resource.type === 'file' || target.stat.isFile() ? 'file' : 'directory';
      entries.push({
        name: path.basename(resource.path),
        path: resource.path,
        type,
        size: type === 'file' ? target.stat.size : 0,
        mtimeMs: target.stat.mtimeMs,
      });
    } catch {
      // Unavailable resources stay invisible in the lease browser root.
    }
  }
  entries.sort((a, b) => {
    if (a.type !== b.type) return a.type === 'directory' ? -1 : 1;
    return a.name.localeCompare(b.name);
  });
  return {
    label: lease.label,
    leaseId: lease.id,
    path: '/',
    ...(lease.expiresAt ? { expiresAt: lease.expiresAt } : {}),
    entries,
  };
}

function usesResourceNamespace(lease: LeaseRecord): boolean {
  const resources = leaseResources(lease);
  return resources.length > 1 || resources[0]?.type === 'file';
}

function resourceForRequest(
  resources: LeaseResource[],
  leaseRelativePath: string,
): { resource: LeaseResource; resourceRelativePath: string } {
  const segments = leaseRelativePath.split('/').filter(Boolean);
  const requestedRoot = segments[0];
  const resource = resources.find((candidate) => toVirtualRelative(candidate.path) === requestedRoot);
  if (!resource) throw new Error(`unknown lease resource: /${requestedRoot ?? ''}`);
  return { resource, resourceRelativePath: segments.slice(1).join('/') };
}

function uploadResourceForRequest(lease: LeaseRecord, leaseRelativePath: string): LeaseResource {
  const resources = leaseResources(lease);
  if (usesResourceNamespace(lease)) return resourceForRequest(resources, leaseRelativePath).resource;
  return resources[0]!;
}

function sourceRelativePathForResource(resource: LeaseResource, leaseRelativePath: string): string {
  if (toVirtualRelative(resource.path) === leaseRelativePath) return '';
  if (!leaseRelativePath.startsWith(`${toVirtualRelative(resource.path)}/`)) return leaseRelativePath;
  return leaseRelativePath.slice(toVirtualRelative(resource.path).length + 1);
}

function normalizeLeaseRelativePath(inputPath: string): string {
  const segments = normalizePathSeparators(inputPath).split('/').filter(Boolean);
  if (segments.some((segment) => segment === '..')) throw new Error('path escapes lease root');
  const normalized = path.posix.normalize(`/${segments.join('/')}`);
  if (normalized === '/') return '';
  return toVirtualRelative(normalized);
}

function joinLeaseRelativePath(parent: string, child: string): string {
  return [parent, child].filter(Boolean).join('/');
}

function stripTrailingSlash(value: string): string {
  return value.endsWith('/') && value !== '/' ? value.slice(0, -1) : value;
}

function isWithin(root: string, candidate: string): boolean {
  const relative = path.relative(root, candidate);
  return relative === '' || (!relative.startsWith('..') && !path.isAbsolute(relative));
}
