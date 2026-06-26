import fsp from 'fs/promises';
import path from 'path';
import { LocalFolderMountConfig } from '../../config/schema.js';
import { MountRegistry } from '../../context/mount-registry.js';
import { isGloballyDeniedPath, matchesConfiguredOrGlobalPattern, matchesPathPatterns } from '../../context/path-policy.js';

export interface DashboardFileTarget {
  virtualPath: string;
  realPath: string;
  parentRealPath: string;
  filename: string;
  type: 'directory' | 'file';
  size: number;
  mtimeMs: number;
  writeAccess: boolean;
}

export async function resolveDashboardFileTarget(
  mounts: readonly LocalFolderMountConfig[],
  requestPath: string,
): Promise<DashboardFileTarget> {
  const registry = new MountRegistry([...mounts]);
  const normalizedPath = normalizeDashboardPath(requestPath);
  const resolved = registry.resolve(normalizedPath);
  const stat = await fsp.stat(resolved.realPath);
  if (!stat.isDirectory() && !stat.isFile()) throw new Error(`${normalizedPath} is not a file or directory`);
  const policyPath = resolved.relativePath || path.basename(resolved.realPath);
  if (matchesConfiguredOrGlobalPattern(policyPath, resolved.mount.config.exclude)) {
    throw new Error(`${normalizedPath} is excluded`);
  }
  const mountRootRealPath = await fsp.realpath(resolved.mount.root);
  const realPath = await fsp.realpath(resolved.realPath);
  if (!isWithin(mountRootRealPath, realPath)) {
    throw new Error(`${normalizedPath} escapes mount root`);
  }
  if (isGloballyDeniedPath(policyPath, realPath)) {
    throw new Error(`${normalizedPath} is globally denied`);
  }
  return {
    virtualPath: resolved.virtualPath,
    realPath,
    parentRealPath: await fsp.realpath(path.dirname(resolved.realPath)),
    filename: path.basename(resolved.realPath),
    type: stat.isDirectory() ? 'directory' : 'file',
    size: stat.isDirectory() ? 0 : stat.size,
    mtimeMs: stat.mtimeMs,
    writeAccess: Boolean(resolved.mount.config.writeAccess),
  };
}

export async function assertDashboardWriteAllowed(
  mounts: readonly LocalFolderMountConfig[],
  requestPath: string,
): Promise<DashboardFileTarget> {
  const target = await resolveDashboardFileTarget(mounts, requestPath);
  if (!target.writeAccess) throw new Error(`${target.virtualPath} is read-only`);
  const registry = new MountRegistry([...mounts]);
  const resolved = registry.resolve(target.virtualPath);
  const policyPath = resolved.relativePath || target.filename;
  if (matchesPathPatterns(policyPath, resolved.mount.config.protect)) {
    throw new Error(`${target.virtualPath} is protected`);
  }
  return target;
}

export function normalizeDashboardPath(inputPath: string): string {
  const normalized = inputPath.trim().replaceAll('\\', '/').split('/').filter(Boolean).join('/');
  return normalized ? `/${normalized}` : '/';
}

function isWithin(root: string, candidate: string): boolean {
  const relative = path.relative(root, candidate);
  return relative === '' || (!relative.startsWith('..') && !path.isAbsolute(relative));
}
