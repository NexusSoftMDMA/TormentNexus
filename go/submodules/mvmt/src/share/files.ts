import fsp from 'fs/promises';
import path from 'path';
import { LocalFolderMountConfig } from '../config/schema.js';
import { MountRegistry } from '../context/mount-registry.js';
import { isGloballyDeniedPath, matchesConfiguredOrGlobalPattern } from '../context/path-policy.js';

export interface ShareFileTarget {
  mountName: string;
  virtualPath: string;
  realPath: string;
  filename: string;
  size: number;
  mtimeMs: number;
}

export async function resolveShareFileTarget(
  mounts: readonly LocalFolderMountConfig[],
  virtualPath: string,
): Promise<ShareFileTarget> {
  const resolved = new MountRegistry([...mounts]).resolve(virtualPath);
  const policyPath = resolved.relativePath || path.basename(resolved.realPath);
  if (matchesConfiguredOrGlobalPattern(policyPath, resolved.mount.config.exclude)) {
    throw new Error(`${resolved.virtualPath} is excluded`);
  }

  const rootRealPath = await fsp.realpath(resolved.mount.root);
  const targetRealPath = await fsp.realpath(resolved.realPath);
  if (!isWithin(rootRealPath, targetRealPath)) {
    throw new Error(`${resolved.virtualPath} escapes mount root`);
  }
  if (isGloballyDeniedPath(policyPath, targetRealPath)) {
    throw new Error(`${resolved.virtualPath} is globally denied`);
  }

  const stat = await fsp.stat(targetRealPath);
  if (!stat.isFile()) throw new Error(`${resolved.virtualPath} is not a file`);
  return {
    mountName: resolved.mount.config.name,
    virtualPath: resolved.virtualPath,
    realPath: targetRealPath,
    filename: path.basename(resolved.virtualPath),
    size: stat.size,
    mtimeMs: stat.mtimeMs,
  };
}

function isWithin(root: string, candidate: string): boolean {
  const relative = path.relative(root, candidate);
  return relative === '' || (!relative.startsWith('..') && !path.isAbsolute(relative));
}
