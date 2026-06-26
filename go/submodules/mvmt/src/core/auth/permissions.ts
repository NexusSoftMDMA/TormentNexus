import { PermissionConfig } from '../../config/schema.js';
import { normalizePathSeparators, stripTrailingSlashes } from '../../context/mount-registry.js';
import { ClientIdentity } from './client-identity.js';

export type PermissionAction = PermissionConfig['actions'][number];

export function pathAllowed(inputPath: string, action: PermissionAction, identity?: ClientIdentity): boolean {
  if (!identity || identity.isLegacyDefault) return true;
  const normalized = normalizePermissionPath(inputPath);
  return identity.permissions.some((permission) => (
    permission.actions.includes(action) && pathMatchesPermission(normalized, permission.path)
  ));
}

export function pathMayExposeEntry(inputPath: string, action: PermissionAction, identity?: ClientIdentity): boolean {
  if (!identity || identity.isLegacyDefault) return true;
  const normalized = normalizePermissionPath(inputPath);
  return identity.permissions.some((permission) => {
    if (!permission.actions.includes(action)) return false;
    if (pathMatchesPermission(normalized, permission.path)) return true;
    const base = permission.path.endsWith('/**')
      ? normalizePermissionPath(permission.path.slice(0, -3))
      : normalizePermissionPath(permission.path);
    return base === normalized || base.startsWith(`${normalized}/`);
  });
}

export function actionAvailable(action: PermissionAction, identity?: ClientIdentity): boolean {
  return !identity || identity.isLegacyDefault || identity.permissions.some((permission) => permission.actions.includes(action));
}

function pathMatchesPermission(inputPath: string, pattern: string): boolean {
  const normalizedPattern = normalizePermissionPath(pattern);
  if (normalizedPattern === '/**') return true;
  if (normalizedPattern.endsWith('/**')) {
    const base = normalizedPattern.slice(0, -3);
    return inputPath === base || inputPath.startsWith(`${base}/`);
  }
  return inputPath === normalizedPattern;
}

function normalizePermissionPath(inputPath: string): string {
  const trimmed = stripTrailingSlashes(normalizePathSeparators(inputPath.trim()));
  if (!trimmed || trimmed === '/') return '/';
  return trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
}
