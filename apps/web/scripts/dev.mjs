#!/usr/bin/env node
import { spawn } from 'child_process';
import { existsSync, mkdirSync, readdirSync, rmSync, writeFileSync } from 'node:fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const workspaceRoot = resolve(__dirname, '..', '..', '..');
const webDir = resolve(__dirname, '..');
const args = process.argv.slice(2);

function readOption(flagNames, fallback) {
  for (let index = 0; index < args.length; index += 1) {
    if (flagNames.includes(args[index])) {
      return args[index + 1] ?? fallback;
    }
  }
  return fallback;
}

const port = readOption(['--port', '-p'], process.env.PORT || '3000');
const host = readOption(['--host', '--hostname', '-H'], process.env.HOSTNAME || 'localhost');
const lockPath = resolve(webDir, '.next', 'dev', 'lock');
const portMarkerPath = resolve(webDir, '.tormentnexus-dev-port.json');

function writePortMarker() {
  mkdirSync(dirname(portMarkerPath), { recursive: true });
  writeFileSync(portMarkerPath, JSON.stringify({
    port: Number(port), host, mode: 'dev', updatedAt: new Date().toISOString()
  }, null, 2));
}

async function isDashboardReachable() {
  try {
    const response = await fetch(`http://${host}:${port}/dashboard`, {
      signal: AbortSignal.timeout(1200),
    });
    return response.ok;
  } catch {
    return false;
  }
}

// Resolve next binary — prefer workspace link, then scan pnpm store
let nextBin = resolve(webDir, 'node_modules/next/dist/bin/next');
if (!existsSync(nextBin)) {
  nextBin = resolve(workspaceRoot, 'node_modules/next/dist/bin/next');
}
if (!existsSync(nextBin)) {
  const pnpmDir = resolve(workspaceRoot, 'node_modules/.pnpm');
  if (existsSync(pnpmDir)) {
    const nextDir = readdirSync(pnpmDir).find(d => d.startsWith('next@'));
    if (nextDir) nextBin = resolve(pnpmDir, nextDir, 'node_modules/next/dist/bin/next');
  }
}
if (!existsSync(nextBin)) {
  console.error('[web:dev] Could not find next binary. Run pnpm install.');
  process.exit(1);
}
console.log('[web:dev] Using next binary:', nextBin);

if (existsSync(lockPath)) {
  if (await isDashboardReachable()) {
    writePortMarker();
    console.log(`[web:dev] Reusing existing Next dev server at http://${host}:${port}`);
    process.exit(0);
  }
  try {
    rmSync(lockPath, { force: true });
    console.log(`[web:dev] Cleared stale Next dev lock: ${lockPath}`);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    console.warn(`[web:dev] Could not clear Next dev lock (${message}). Another dev instance may still be running.`);
  }
}

writePortMarker();

const child = spawn(process.execPath, [nextBin, 'dev', '--port', port], {
  stdio: 'inherit',
  cwd: webDir,
  env: {
    ...process.env,
    NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE: process.env.NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE ?? '1'
  },
});

child.on('exit', async (code) => {
  if (code === 1 && await isDashboardReachable()) {
    console.log(`[web:dev] Next dev server is already available at http://${host}:${port}`);
    process.exit(0);
    return;
  }
  process.exit(code ?? 0);
});
