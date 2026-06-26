import crypto from 'crypto';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { hashApiToken, verifyApiToken } from '../../utils/api-token-hash.js';
import { TOKEN_PATH } from '../../utils/token.js';

export const PRIVILEGED_USERS_PATH = path.join(os.homedir(), '.mvmt', '.privileged-users.json');

export interface PrivilegedUser {
  id: string;
  username: string;
  passwordHash: string;
  createdAt: string;
  lastLoginAt?: string;
  disabled?: boolean;
  // Admins can manage sources through the dashboard. Non-admins can
  // only browse already-mounted folders and
  // create/revoke their own leases. Default false: existing users stay
  // non-admin after a code upgrade until explicitly granted.
  admin?: boolean;
}

interface PrivilegedUserStoreFile {
  version: 1;
  users: PrivilegedUser[];
}

export function defaultPrivilegedUsersPath(tokenPath: string | undefined = TOKEN_PATH): string {
  if (tokenPath === TOKEN_PATH) return PRIVILEGED_USERS_PATH;
  return path.join(path.dirname(tokenPath), '.privileged-users.json');
}

export function listPrivilegedUsers(storePath: string): PrivilegedUser[] {
  return readPrivilegedUserStore(storePath).users;
}

export function findPrivilegedUser(storePath: string, userId: string): PrivilegedUser | undefined {
  return readPrivilegedUserStore(storePath).users.find((user) => user.id === userId);
}

export function createPrivilegedUser(
  storePath: string,
  input: { username: string; password: string; admin?: boolean },
): PrivilegedUser {
  const username = normalizeUsername(input.username);
  validatePassword(input.password);
  const store = readPrivilegedUserStore(storePath);
  if (store.users.some((user) => user.username.toLowerCase() === username.toLowerCase())) {
    throw new Error(`Privileged user already exists: ${username}`);
  }
  const user: PrivilegedUser = {
    id: crypto.randomBytes(12).toString('base64url'),
    username,
    passwordHash: hashApiToken(input.password),
    createdAt: new Date().toISOString(),
    ...(input.admin ? { admin: true } : {}),
  };
  store.users.push(user);
  writePrivilegedUserStore(storePath, store);
  return user;
}

export function setPrivilegedUserAdmin(
  storePath: string,
  usernameInput: string,
  admin: boolean,
): PrivilegedUser {
  const username = normalizeUsername(usernameInput);
  const store = readPrivilegedUserStore(storePath);
  const index = store.users.findIndex((user) => user.username.toLowerCase() === username.toLowerCase());
  if (index < 0) throw new Error(`Unknown privileged user: ${username}`);
  const updated: PrivilegedUser = { ...store.users[index] };
  if (admin) updated.admin = true;
  else delete updated.admin;
  store.users[index] = updated;
  writePrivilegedUserStore(storePath, store);
  return updated;
}

export function removePrivilegedUser(
  storePath: string,
  usernameInput: string,
): PrivilegedUser {
  const username = normalizeUsername(usernameInput);
  const store = readPrivilegedUserStore(storePath);
  const index = store.users.findIndex((user) => user.username.toLowerCase() === username.toLowerCase());
  if (index < 0) throw new Error(`Unknown privileged user: ${username}`);
  const [removed] = store.users.splice(index, 1);
  if (!removed) throw new Error(`Unknown privileged user: ${username}`);
  writePrivilegedUserStore(storePath, store);
  return removed;
}

export function setPrivilegedUserPassword(
  storePath: string,
  usernameInput: string,
  password: string,
): PrivilegedUser {
  const username = normalizeUsername(usernameInput);
  validatePassword(password);
  const store = readPrivilegedUserStore(storePath);
  const index = store.users.findIndex((user) => user.username.toLowerCase() === username.toLowerCase());
  if (index < 0) throw new Error(`Unknown privileged user: ${username}`);
  const updated: PrivilegedUser = {
    ...store.users[index],
    passwordHash: hashApiToken(password),
  };
  store.users[index] = updated;
  writePrivilegedUserStore(storePath, store);
  return updated;
}

export function verifyPrivilegedUserPassword(
  storePath: string,
  usernameInput: string | undefined,
  password: string | undefined,
): PrivilegedUser | undefined {
  if (!usernameInput || !password) return undefined;
  const username = usernameInput.trim();
  if (!username) return undefined;
  const store = readPrivilegedUserStore(storePath);
  const user = store.users.find((entry) => entry.username.toLowerCase() === username.toLowerCase());
  if (!user || user.disabled) return undefined;
  return verifyApiToken(password, user.passwordHash) ? user : undefined;
}

export function recordPrivilegedUserLogin(storePath: string, userId: string): void {
  const store = readPrivilegedUserStore(storePath);
  const index = store.users.findIndex((user) => user.id === userId);
  if (index < 0) return;
  store.users[index] = { ...store.users[index], lastLoginAt: new Date().toISOString() };
  writePrivilegedUserStore(storePath, store);
}

function readPrivilegedUserStore(storePath: string): PrivilegedUserStoreFile {
  try {
    const parsed = JSON.parse(fs.readFileSync(storePath, 'utf-8')) as PrivilegedUserStoreFile;
    if (parsed.version !== 1 || !Array.isArray(parsed.users)) return emptyStore();
    return {
      version: 1,
      users: parsed.users.filter(isPrivilegedUser),
    };
  } catch {
    return emptyStore();
  }
}

function writePrivilegedUserStore(storePath: string, store: PrivilegedUserStoreFile): void {
  fs.mkdirSync(path.dirname(storePath), { recursive: true });
  fs.writeFileSync(storePath, JSON.stringify(store, null, 2), { mode: 0o600 });
  if (process.platform !== 'win32') fs.chmodSync(storePath, 0o600);
}

function emptyStore(): PrivilegedUserStoreFile {
  return { version: 1, users: [] };
}

function normalizeUsername(username: string): string {
  const normalized = username.trim();
  if (!/^[a-zA-Z0-9._-]{1,64}$/.test(normalized)) {
    throw new Error('Username must be 1-64 characters and use letters, numbers, dot, dash, or underscore.');
  }
  return normalized;
}

function validatePassword(password: string): void {
  if (password.length < 8) throw new Error('Password must be at least 8 characters.');
}

function isPrivilegedUser(value: unknown): value is PrivilegedUser {
  if (!value || typeof value !== 'object') return false;
  const user = value as Partial<PrivilegedUser>;
  return typeof user.id === 'string'
    && typeof user.username === 'string'
    && typeof user.passwordHash === 'string'
    && typeof user.createdAt === 'string';
}
