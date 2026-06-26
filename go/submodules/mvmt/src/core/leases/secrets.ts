import fs from 'fs';
import path from 'path';
import { TOKEN_PATH } from '../../utils/token.js';

export const LEASE_SECRETS_FILENAME = '.lease-secrets.json';

export interface LeaseSecretRecord {
  leaseId: string;
  token: string;
  createdAt: string;
  updatedAt: string;
}

interface LeaseSecretStoreFile {
  version: 1;
  secrets: LeaseSecretRecord[];
}

export function defaultLeaseSecretsPath(tokenPath: string | undefined = TOKEN_PATH): string {
  return path.join(path.dirname(tokenPath ?? TOKEN_PATH), LEASE_SECRETS_FILENAME);
}

export function leaseSecretsPathForLeaseStore(leaseStorePath: string): string {
  return path.join(path.dirname(leaseStorePath), LEASE_SECRETS_FILENAME);
}

export function findLeaseSecret(storePath: string, leaseId: string): LeaseSecretRecord | undefined {
  return readLeaseSecretStore(storePath).secrets.find((secret) => secret.leaseId === leaseId);
}

export function saveLeaseSecret(storePath: string, leaseId: string, token: string): LeaseSecretRecord {
  const store = readLeaseSecretStore(storePath);
  const now = new Date().toISOString();
  const index = store.secrets.findIndex((secret) => secret.leaseId === leaseId);
  const record: LeaseSecretRecord = {
    leaseId,
    token,
    createdAt: index >= 0 ? store.secrets[index]!.createdAt : now,
    updatedAt: now,
  };
  if (index >= 0) store.secrets[index] = record;
  else store.secrets.push(record);
  writeLeaseSecretStore(storePath, store);
  return record;
}

export function removeLeaseSecret(storePath: string, leaseId: string): boolean {
  const store = readLeaseSecretStore(storePath);
  const next = store.secrets.filter((secret) => secret.leaseId !== leaseId);
  if (next.length === store.secrets.length) return false;
  writeLeaseSecretStore(storePath, { ...store, secrets: next });
  return true;
}

function readLeaseSecretStore(storePath: string): LeaseSecretStoreFile {
  try {
    const parsed = JSON.parse(fs.readFileSync(storePath, 'utf-8')) as LeaseSecretStoreFile;
    if (parsed.version !== 1 || !Array.isArray(parsed.secrets)) return emptyStore();
    return {
      version: 1,
      secrets: parsed.secrets
        .filter((secret) => typeof secret.leaseId === 'string' && typeof secret.token === 'string')
        .map((secret) => ({
          leaseId: secret.leaseId,
          token: secret.token,
          createdAt: typeof secret.createdAt === 'string' ? secret.createdAt : new Date(0).toISOString(),
          updatedAt: typeof secret.updatedAt === 'string' ? secret.updatedAt : new Date(0).toISOString(),
        })),
    };
  } catch {
    return emptyStore();
  }
}

function writeLeaseSecretStore(storePath: string, store: LeaseSecretStoreFile): void {
  fs.mkdirSync(path.dirname(storePath), { recursive: true });
  fs.writeFileSync(storePath, JSON.stringify(store, null, 2), { mode: 0o600 });
  if (process.platform !== 'win32') fs.chmodSync(storePath, 0o600);
}

function emptyStore(): LeaseSecretStoreFile {
  return { version: 1, secrets: [] };
}
