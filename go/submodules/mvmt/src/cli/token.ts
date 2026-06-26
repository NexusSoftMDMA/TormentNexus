import chalk from 'chalk';
import fs from 'fs';
import { defaultSigningKeyPath, ensureSessionToken, generateSessionToken, readSessionToken, rotateSigningKey, TOKEN_PATH } from '../utils/token.js';

export interface TokenSummary {
  token?: string;
  path: string;
  rotatedAt?: Date;
  ageMs?: number;
}

export async function showToken(tokenPath = TOKEN_PATH): Promise<void> {
  const token = ensureSessionToken(tokenPath);
  console.log(token);
}

export async function showTokenSummary(tokenPath = TOKEN_PATH): Promise<void> {
  ensureSessionToken(tokenPath);
  const summary = readTokenSummary(tokenPath);
  printTokenSummary(summary);
}

export async function rotateToken(tokenPath = TOKEN_PATH): Promise<void> {
  const token = generateSessionToken(tokenPath);
  // Rotate the OAuth signing key too so outstanding access tokens are
  // invalidated. Keeping them valid across a token rotation would
  // defeat the purpose of rotation.
  rotateSigningKey(defaultSigningKeyPath(tokenPath));
  console.log(token);
  console.error(
    `Rotated session token at ${tokenPath}. Update any HTTP MCP clients that store the old token. OAuth access tokens were revoked immediately.`,
  );
}

export function readTokenSummary(tokenPath = TOKEN_PATH): TokenSummary {
  const token = readSessionToken(tokenPath);
  try {
    const stat = fs.statSync(tokenPath);
    const rotatedAt = stat.mtime;
    return {
      token,
      path: tokenPath,
      rotatedAt,
      ageMs: Date.now() - stat.mtimeMs,
    };
  } catch {
    return {
      token,
      path: tokenPath,
    };
  }
}

export function printTokenSummary(summary: TokenSummary): void {
  console.log('mvmt token\n');
  if (!summary.token) {
    console.log(chalk.yellow(`No token found at ${summary.path}`));
    console.log(`Run ${chalk.cyan('mvmt serve')} to create one, or ${chalk.cyan('mvmt token session-rotate')} to replace it.`);
    return;
  }

  console.log('Token');
  console.log(`  value: ${summary.token}`);
  if (summary.ageMs !== undefined) {
    console.log(`  age: ${formatAge(summary.ageMs)}`);
  }
  if (summary.rotatedAt) {
    console.log(`  rotated: ${summary.rotatedAt.toLocaleString()}`);
  }
  console.log(`  path: ${summary.path}`);
  console.log(`  rotate: ${chalk.cyan('mvmt token session-rotate')}`);
}

function formatAge(ageMs: number): string {
  const totalSeconds = Math.max(0, Math.floor(ageMs / 1000));
  const days = Math.floor(totalSeconds / 86_400);
  const hours = Math.floor((totalSeconds % 86_400) / 3_600);
  const minutes = Math.floor((totalSeconds % 3_600) / 60);
  const seconds = totalSeconds % 60;

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes}m ${seconds}s`;
  return `${seconds}s`;
}
