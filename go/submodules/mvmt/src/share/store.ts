import crypto from 'crypto';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { hashApiToken, verifyApiToken } from '../utils/api-token-hash.js';
import { TOKEN_PATH } from '../utils/token.js';
import { isExpired } from '../utils/token-ttl.js';

export const SHARES_PATH = path.join(os.homedir(), '.mvmt', '.shares.json');
export const DEFAULT_SHARE_TTL = '24h';

export interface ShareRecord {
  id: string;
  path: string;
  tokenHash: string;
  createdAt: string;
  expiresAt?: string;
  downloadCount: number;
  revokedAt?: string;
}

interface ShareStoreFile {
  version: 1;
  shares: ShareRecord[];
}

export interface CreatedShare {
  record: ShareRecord;
  token: string;
}

export function defaultSharesPath(tokenPath: string | undefined = TOKEN_PATH): string {
  if (tokenPath === TOKEN_PATH) return SHARES_PATH;
  return path.join(path.dirname(tokenPath), '.shares.json');
}

export function createShare(storePath: string, input: { path: string; expiresAt?: string }): CreatedShare {
  const store = readShareStore(storePath);
  const token = `mvmt_s_${crypto.randomBytes(32).toString('base64url')}`;
  const record: ShareRecord = {
    id: uniqueShareId(store),
    path: input.path,
    tokenHash: hashApiToken(token),
    createdAt: new Date().toISOString(),
    ...(input.expiresAt ? { expiresAt: input.expiresAt } : {}),
    downloadCount: 0,
  };
  store.shares.push(record);
  writeShareStore(storePath, store);
  return { record, token };
}

export function listShares(storePath: string): ShareRecord[] {
  return readShareStore(storePath).shares;
}

export function findShare(storePath: string, id: string): ShareRecord | undefined {
  return readShareStore(storePath).shares.find((share) => share.id === id);
}

export function removeShare(storePath: string, id: string): boolean {
  const store = readShareStore(storePath);
  const next = store.shares.filter((share) => share.id !== id);
  if (next.length === store.shares.length) return false;
  writeShareStore(storePath, { ...store, shares: next });
  return true;
}

export function recordShareDownload(storePath: string, id: string): void {
  const store = readShareStore(storePath);
  const index = store.shares.findIndex((share) => share.id === id);
  if (index < 0) return;
  store.shares[index] = {
    ...store.shares[index],
    downloadCount: store.shares[index].downloadCount + 1,
  };
  writeShareStore(storePath, store);
}

export function validateShareToken(share: ShareRecord, token: string | undefined): boolean {
  return Boolean(token && verifyApiToken(token, share.tokenHash));
}

export function shareUnavailableReason(share: ShareRecord, now = Date.now()): 'expired' | 'revoked' | undefined {
  if (share.revokedAt) return 'revoked';
  if (isExpired(share.expiresAt, now)) return 'expired';
  return undefined;
}

function readShareStore(storePath: string): ShareStoreFile {
  try {
    const parsed = JSON.parse(fs.readFileSync(storePath, 'utf-8')) as ShareStoreFile;
    if (parsed.version !== 1 || !Array.isArray(parsed.shares)) return emptyStore();
    return parsed;
  } catch {
    return emptyStore();
  }
}

function writeShareStore(storePath: string, store: ShareStoreFile): void {
  fs.mkdirSync(path.dirname(storePath), { recursive: true });
  fs.writeFileSync(storePath, JSON.stringify(store, null, 2), { mode: 0o600 });
  if (process.platform !== 'win32') fs.chmodSync(storePath, 0o600);
}

function emptyStore(): ShareStoreFile {
  return { version: 1, shares: [] };
}

function uniqueShareId(store: ShareStoreFile): string {
  const existing = new Set(store.shares.map((share) => share.id));
  for (;;) {
    const id = crypto.randomBytes(12).toString('base64url');
    if (!existing.has(id)) return id;
  }
}
