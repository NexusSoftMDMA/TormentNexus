#!/usr/bin/env node
import { existsSync, mkdirSync, writeFileSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawn } from 'node:child_process';

const __dirname = dirname(fileURLToPath(import.meta.url));
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
const host = readOption(['--host', '--hostname', '-H'], process.env.HOSTNAME || '0.0.0.0');
const standaloneServer = resolve(webDir, '.next', 'standalone', 'apps', 'web', 'server.js');
const portMarkerPath = resolve(webDir, '.tormentnexus-dev-port.json');

function writePortMarker() {
  mkdirSync(dirname(portMarkerPath), { recursive: true });
  writeFileSync(portMarkerPath, JSON.stringify({ port: Number(port), host, mode: 'standalone', updatedAt: new Date().toISOString() }, null, 2));
}

if (!existsSync(standaloneServer)) {
  console.error('[web:start] standalone server not found. Run `pnpm -C apps/web build` first.');
  process.exit(1);
}

writePortMarker();

const child = spawn(process.execPath, [standaloneServer], {
  cwd: webDir,
  stdio: 'inherit',
  env: {
    ...process.env,
    PORT: String(port),
    HOSTNAME: host,
    TORMENTNEXUS_TRPC_UPSTREAM: process.env.TORMENTNEXUS_TRPC_UPSTREAM || 'http://127.0.0.1:4100/trpc',
  },
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 0);
});
