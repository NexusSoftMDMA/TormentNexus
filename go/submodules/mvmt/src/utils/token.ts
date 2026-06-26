import crypto from 'crypto';
import fs from 'fs';
import os from 'os';
import path from 'path';

export const TOKEN_PATH = path.join(os.homedir(), '.mvmt', '.session-token');
export const SIGNING_KEY_PATH = path.join(os.homedir(), '.mvmt', '.signing-key');
export const CLIENTS_PATH = path.join(os.homedir(), '.mvmt', '.clients.json');
export const REFRESH_TOKENS_PATH = path.join(os.homedir(), '.mvmt', '.refresh-tokens.json');

// Per-process random key; HMACing both sides produces fixed-length
// digests so the subsequent timing-safe compare never short-circuits
// on length and cannot leak the expected token length via timing.
const COMPARE_KEY = crypto.randomBytes(32);

export function generateSessionToken(tokenPath = TOKEN_PATH): string {
  const token = crypto.randomBytes(32).toString('base64url');

  writeSecretFile(token, tokenPath);
  return token;
}

export function ensureSessionToken(tokenPath = TOKEN_PATH): string {
  const existingToken = readSessionToken(tokenPath);
  if (existingToken) return existingToken;
  return generateSessionToken(tokenPath);
}

export function readSessionToken(tokenPath = TOKEN_PATH): string | undefined {
  try {
    const token = fs.readFileSync(tokenPath, 'utf-8').trim();
    return token.length > 0 ? token : undefined;
  } catch {
    return undefined;
  }
}

// Signing key used as the HMAC secret for self-validating OAuth access
// tokens. Kept separate from the session token so that phishing or
// accidental disclosure of the session token (e.g. pasted into a
// malicious authorize page) does not reveal the key needed to forge
// access tokens.
export function ensureSigningKey(keyPath = SIGNING_KEY_PATH): string {
  const existing = readSigningKey(keyPath);
  if (existing) return existing;
  const key = crypto.randomBytes(32).toString('base64url');
  writeSecretFile(key, keyPath);
  return key;
}

export function readSigningKey(keyPath = SIGNING_KEY_PATH): string | undefined {
  try {
    const key = fs.readFileSync(keyPath, 'utf-8').trim();
    return key.length > 0 ? key : undefined;
  } catch {
    return undefined;
  }
}

export function rotateSigningKey(keyPath = SIGNING_KEY_PATH): string {
  const key = crypto.randomBytes(32).toString('base64url');
  writeSecretFile(key, keyPath);
  return key;
}

export function defaultSigningKeyPath(tokenPath: string): string {
  if (tokenPath === TOKEN_PATH) return SIGNING_KEY_PATH;
  return path.join(path.dirname(tokenPath), '.signing-key');
}

export function defaultClientsPath(tokenPath: string): string {
  if (tokenPath === TOKEN_PATH) return CLIENTS_PATH;
  return path.join(path.dirname(tokenPath), '.clients.json');
}

export function defaultRefreshTokensPath(tokenPath: string): string {
  if (tokenPath === TOKEN_PATH) return REFRESH_TOKENS_PATH;
  return path.join(path.dirname(tokenPath), '.refresh-tokens.json');
}

export function validateSessionToken(authHeader: string | undefined, tokenPath = TOKEN_PATH): boolean {
  const expectedToken = readSessionToken(tokenPath);
  if (!expectedToken) return false;
  return validateToken(authHeader, expectedToken);
}

export function verifySessionTokenValue(candidate: unknown, tokenPath = TOKEN_PATH): boolean {
  if (typeof candidate !== 'string' || candidate.length === 0) return false;
  const expectedToken = readSessionToken(tokenPath);
  if (!expectedToken) return false;
  const provided = crypto.createHmac('sha256', COMPARE_KEY).update(candidate, 'utf8').digest();
  const expected = crypto.createHmac('sha256', COMPARE_KEY).update(expectedToken, 'utf8').digest();
  return crypto.timingSafeEqual(provided, expected);
}

function writeSecretFile(value: string, filePath: string): void {
  fs.mkdirSync(path.dirname(filePath), { recursive: true });
  fs.writeFileSync(filePath, value, { mode: 0o600 });
  if (process.platform !== 'win32') {
    fs.chmodSync(filePath, 0o600);
  }
}

export function validateToken(authHeader: string | undefined, expectedToken: string): boolean {
  if (!authHeader) return false;

  const parts = authHeader.split(' ');
  if (parts.length !== 2 || parts[0] !== 'Bearer') return false;

  const provided = crypto.createHmac('sha256', COMPARE_KEY).update(parts[1], 'utf8').digest();
  const expected = crypto.createHmac('sha256', COMPARE_KEY).update(expectedToken, 'utf8').digest();
  return crypto.timingSafeEqual(provided, expected);
}
