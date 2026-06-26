import crypto from 'crypto';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { hashApiToken, verifyApiToken } from '../../utils/api-token-hash.js';
import { TOKEN_PATH } from '../../utils/token.js';
import { isExpired } from '../../utils/token-ttl.js';

export const LEASES_PATH = path.join(os.homedir(), '.mvmt', '.leases.json');
export const DEFAULT_LEASE_TTL = '24h';

export type LeasePermission = 'read' | 'write' | 'upload';
export type LeaseResourceType = 'file' | 'folder';

export interface LeaseResource {
  path: string;
  sourcePath: string;
  type: LeaseResourceType;
}

export interface LeaseRecord {
  id: string;
  label: string;
  path: string;
  resources: LeaseResource[];
  permissions: LeasePermission[];
  tokenHash: string;
  createdAt: string;
  expiresAt?: string;
  lastUsedAt?: string;
  revokedAt?: string;
  downloadCount: number;
  maxDownloads?: number;
  uploadCount: number;
  // Exposure boundary for the capability/grant model. Absent means the
  // lease predates the publish concept and is grandfathered as published
  // so existing share links keep working — see isGrantPublished().
  published?: boolean;
}

interface LeaseStoreFile {
  version: 1;
  leases: LeaseRecord[];
}

export interface CreatedLease {
  record: LeaseRecord;
  token: string;
}

export function defaultLeasesPath(tokenPath: string | undefined = TOKEN_PATH): string {
  if (tokenPath === TOKEN_PATH) return LEASES_PATH;
  return path.join(path.dirname(tokenPath), '.leases.json');
}

export function createLease(
  storePath: string,
  input: {
    label: string;
    path?: string;
    resources?: LeaseResource[];
    expiresAt?: string;
    permissions?: LeasePermission[];
    maxDownloads?: number;
    // Exposure boundary. Omit to leave the lease grandfathered as
    // published; pass false for a capability-only lease that only local
    // apps can reach. Callers minting through a deliberate "share"
    // gesture (e.g. the dashboard) should pass true.
    published?: boolean;
  },
): CreatedLease {
  const store = readLeaseStore(storePath);
  const token = `mvmt_l_${crypto.randomBytes(32).toString('base64url')}`;
  const resources = normalizeLeaseResources(input.resources, input.path);
  if (resources.length === 0) throw new Error('lease must include at least one resource');
  if (input.maxDownloads !== undefined && !validDownloadLimit(input.maxDownloads)) {
    throw new Error('maxDownloads must be a positive safe integer');
  }
  const path = input.path ?? resources[0]!.sourcePath;
  const record: LeaseRecord = {
    id: uniqueLeaseId(store),
    label: input.label,
    path,
    resources,
    permissions: input.permissions ?? ['read'],
    tokenHash: hashApiToken(token),
    createdAt: new Date().toISOString(),
    ...(input.expiresAt ? { expiresAt: input.expiresAt } : {}),
    downloadCount: 0,
    ...(input.maxDownloads === undefined ? {} : { maxDownloads: input.maxDownloads }),
    uploadCount: 0,
    ...(input.published === undefined ? {} : { published: input.published }),
  };
  store.leases.push(record);
  writeLeaseStore(storePath, store);
  return { record, token };
}

export function listLeases(storePath: string): LeaseRecord[] {
  return readLeaseStore(storePath).leases;
}

export function findLease(storePath: string, id: string): LeaseRecord | undefined {
  return readLeaseStore(storePath).leases.find((lease) => lease.id === id);
}

export function findLeaseByToken(storePath: string, token: string | undefined): LeaseRecord | undefined {
  if (!token) return undefined;
  return readLeaseStore(storePath).leases.find((lease) => validateLeaseToken(lease, token));
}

export function addLeaseResources(storePath: string, id: string, resources: LeaseResource[]): LeaseRecord | undefined {
  const store = readLeaseStore(storePath);
  const index = store.leases.findIndex((lease) => lease.id === id);
  if (index < 0) return undefined;
  const current = store.leases[index]!;
  const nextResources = mergeLeaseResources(leaseResources(current), resources);
  const nextLease = {
    ...current,
    path: current.path || nextResources[0]!.sourcePath,
    resources: nextResources,
  };
  store.leases[index] = nextLease;
  writeLeaseStore(storePath, store);
  return nextLease;
}

export function rotateLeaseToken(storePath: string, id: string): CreatedLease | undefined {
  const store = readLeaseStore(storePath);
  const index = store.leases.findIndex((lease) => lease.id === id);
  if (index < 0) return undefined;
  const token = `mvmt_l_${crypto.randomBytes(32).toString('base64url')}`;
  const next: LeaseRecord = { ...store.leases[index], tokenHash: hashApiToken(token) };
  store.leases[index] = next;
  writeLeaseStore(storePath, store);
  return { record: next, token };
}

// Sets the exposure boundary for a lease. `published: false` makes it
// capability-only (no relay door, still usable by local apps);
// `published: true` gives it a relay door. Distinct from revokeLease,
// which kills the lease entirely.
export function setLeasePublished(storePath: string, id: string, published: boolean): LeaseRecord | undefined {
  const store = readLeaseStore(storePath);
  const index = store.leases.findIndex((lease) => lease.id === id);
  if (index < 0) return undefined;
  const next: LeaseRecord = { ...store.leases[index]!, published };
  store.leases[index] = next;
  writeLeaseStore(storePath, store);
  return next;
}

export function revokeLease(storePath: string, id: string): boolean {
  const store = readLeaseStore(storePath);
  const index = store.leases.findIndex((lease) => lease.id === id);
  if (index < 0) return false;
  store.leases[index] = {
    ...store.leases[index],
    revokedAt: store.leases[index].revokedAt ?? new Date().toISOString(),
  };
  writeLeaseStore(storePath, store);
  return true;
}

export function recordLeaseUse(storePath: string, id: string, options: { downloaded?: boolean; uploaded?: boolean } = {}): void {
  const store = readLeaseStore(storePath);
  const index = store.leases.findIndex((lease) => lease.id === id);
  if (index < 0) return;
  const lease = store.leases[index];
  store.leases[index] = {
    ...lease,
    lastUsedAt: new Date().toISOString(),
    downloadCount: (lease.downloadCount ?? 0) + (options.downloaded ? 1 : 0),
    uploadCount: (lease.uploadCount ?? 0) + (options.uploaded ? 1 : 0),
  };
  writeLeaseStore(storePath, store);
}

export function reserveLeaseDownload(storePath: string, id: string): LeaseRecord | 'download_limit_reached' | undefined {
  const store = readLeaseStore(storePath);
  const index = store.leases.findIndex((lease) => lease.id === id);
  if (index < 0) return undefined;
  const lease = store.leases[index]!;
  if (lease.maxDownloads !== undefined && lease.downloadCount >= lease.maxDownloads) {
    return 'download_limit_reached';
  }
  const next: LeaseRecord = {
    ...lease,
    lastUsedAt: new Date().toISOString(),
    downloadCount: lease.downloadCount + 1,
  };
  store.leases[index] = next;
  writeLeaseStore(storePath, store);
  return next;
}

export function validateLeaseToken(lease: LeaseRecord, token: string | undefined): boolean {
  return Boolean(token && verifyApiToken(token, lease.tokenHash));
}

export function leaseUnavailableReason(lease: LeaseRecord, now = Date.now()): 'expired' | 'revoked' | 'download_limit_reached' | undefined {
  if (lease.revokedAt) return 'revoked';
  if (isExpired(lease.expiresAt, now)) return 'expired';
  if (lease.maxDownloads !== undefined && lease.downloadCount >= lease.maxDownloads) return 'download_limit_reached';
  return undefined;
}

export function leaseAllows(lease: LeaseRecord, permission: LeasePermission): boolean {
  return lease.permissions.includes(permission);
}

export function leaseResources(lease: LeaseRecord): LeaseResource[] {
  return normalizeLeaseResources(lease.resources, lease.path);
}

function readLeaseStore(storePath: string): LeaseStoreFile {
  try {
    const parsed = JSON.parse(fs.readFileSync(storePath, 'utf-8')) as LeaseStoreFile;
    if (parsed.version !== 1 || !Array.isArray(parsed.leases)) return emptyStore();
    return {
      version: 1,
      leases: parsed.leases.map((lease) => ({
        ...lease,
        resources: normalizeLeaseResources((lease as Partial<LeaseRecord>).resources, lease.path),
        permissions: lease.permissions?.filter(isLeasePermission) ?? ['read'],
        downloadCount: lease.downloadCount ?? 0,
        ...normalizeMaxDownloads(lease.maxDownloads),
        uploadCount: lease.uploadCount ?? 0,
      })),
    };
  } catch {
    return emptyStore();
  }
}

function normalizeMaxDownloads(value: unknown): { maxDownloads?: number } {
  return typeof value === 'number' && validDownloadLimit(value)
    ? { maxDownloads: value }
    : {};
}

function validDownloadLimit(value: number): boolean {
  return Number.isSafeInteger(value) && value > 0;
}

function writeLeaseStore(storePath: string, store: LeaseStoreFile): void {
  fs.mkdirSync(path.dirname(storePath), { recursive: true });
  fs.writeFileSync(storePath, JSON.stringify(store, null, 2), { mode: 0o600 });
  if (process.platform !== 'win32') fs.chmodSync(storePath, 0o600);
}

function emptyStore(): LeaseStoreFile {
  return { version: 1, leases: [] };
}

function uniqueLeaseId(store: LeaseStoreFile): string {
  const existing = new Set(store.leases.map((lease) => lease.id));
  for (;;) {
    const id = crypto.randomBytes(12).toString('base64url');
    if (!existing.has(id)) return id;
  }
}

function isLeasePermission(value: unknown): value is LeasePermission {
  return value === 'read' || value === 'write' || value === 'upload';
}

function normalizeLeaseResources(resources: LeaseResource[] | undefined, fallbackPath: string | undefined): LeaseResource[] {
  const normalized = (resources ?? [])
    .map((resource) => ({
      path: normalizeLeasePath(resource.path),
      sourcePath: normalizeLeasePath(resource.sourcePath),
      type: resource.type === 'file' ? 'file' as const : 'folder' as const,
    }))
    .filter((resource) => resource.path !== '/' && resource.sourcePath !== '/');
  if (normalized.length > 0) return normalized;
  if (!fallbackPath) return [];
  const path = normalizeLeasePath(fallbackPath ?? '/lease');
  return [{ path, sourcePath: path, type: 'folder' }];
}

function normalizeLeasePath(inputPath: string): string {
  const normalized = inputPath.trim().replaceAll('\\', '/').split('/').filter(Boolean).join('/');
  return normalized ? `/${normalized}` : '/';
}

function mergeLeaseResources(current: LeaseResource[], added: LeaseResource[]): LeaseResource[] {
  const merged: LeaseResource[] = [];
  const seen = new Set<string>();
  for (const resource of [...current, ...normalizeLeaseResources(added, undefined)]) {
    const key = `${resource.sourcePath}:${resource.type}`;
    if (seen.has(key)) continue;
    seen.add(key);
    merged.push(resource);
  }
  return merged;
}
