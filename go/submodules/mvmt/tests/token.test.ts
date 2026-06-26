import fs from 'fs';
import os from 'os';
import path from 'path';
import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  ensureSessionToken,
  generateSessionToken,
  readSessionToken,
  validateSessionToken,
  validateToken,
} from '../src/utils/token.js';
import { rotateToken, showToken, showTokenSummary } from '../src/cli/token.js';

const tempDirs: string[] = [];

afterEach(() => {
  process.exitCode = undefined;
  vi.restoreAllMocks();
  while (tempDirs.length > 0) {
    fs.rmSync(tempDirs.pop()!, { recursive: true, force: true });
  }
});

function createTokenPath(): string {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'mvmt-token-test-'));
  tempDirs.push(dir);
  return path.join(dir, '.mvmt', '.session-token');
}

describe('session token utilities', () => {
  it('generates a token, writes it to disk, and validates bearer headers', () => {
    const tokenPath = createTokenPath();
    const token = generateSessionToken(tokenPath);

    expect(token).toHaveLength(43);
    expect(fs.readFileSync(tokenPath, 'utf-8')).toBe(token);
    expect(readSessionToken(tokenPath)).toBe(token);
    expect(validateToken(`Bearer ${token}`, token)).toBe(true);
    expect(validateSessionToken(`Bearer ${token}`, tokenPath)).toBe(true);
    expect(validateToken(undefined, token)).toBe(false);
    expect(validateToken(`Token ${token}`, token)).toBe(false);
    expect(validateToken('Bearer wrong', token)).toBe(false);
    expect(validateSessionToken('Bearer wrong', tokenPath)).toBe(false);
  });

  it('reuses an existing token without rotating it', () => {
    const tokenPath = createTokenPath();
    const token = generateSessionToken(tokenPath);

    const reused = ensureSessionToken(tokenPath);

    expect(reused).toBe(token);
    expect(readSessionToken(tokenPath)).toBe(token);
  });

  it('creates a token when one does not already exist', () => {
    const tokenPath = createTokenPath();
    fs.rmSync(tokenPath, { force: true });

    const token = ensureSessionToken(tokenPath);

    expect(token).toHaveLength(43);
    expect(readSessionToken(tokenPath)).toBe(token);
  });
});

describe('token CLI helpers', () => {
  it('shows the current token without rotating it', async () => {
    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    const tokenPath = createTokenPath();
    const token = generateSessionToken(tokenPath);

    await showToken(tokenPath);

    expect(output).toHaveBeenCalledWith(token);
    expect(readSessionToken(tokenPath)).toBe(token);
  });

  it('creates a missing token before showing it', async () => {
    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    const tokenPath = createTokenPath();
    fs.rmSync(tokenPath, { force: true });

    await showToken(tokenPath);

    const token = readSessionToken(tokenPath);
    expect(token).toHaveLength(43);
    expect(output).toHaveBeenCalledWith(token);
  });

  it('creates a missing token before printing the summary', async () => {
    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    const tokenPath = createTokenPath();
    fs.rmSync(tokenPath, { force: true });

    await showTokenSummary(tokenPath);

    expect(readSessionToken(tokenPath)).toHaveLength(43);
    expect(output).toHaveBeenCalledWith('mvmt token\n');
  });

  it('rotates and prints a new token', async () => {
    const output = vi.spyOn(console, 'log').mockImplementation(() => undefined);
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    const tokenPath = createTokenPath();
    const oldToken = generateSessionToken(tokenPath);

    await rotateToken(tokenPath);

    const newToken = readSessionToken(tokenPath);
    expect(newToken).toHaveLength(43);
    expect(newToken).not.toBe(oldToken);
    expect(output).toHaveBeenCalledWith(newToken);
  });
});
