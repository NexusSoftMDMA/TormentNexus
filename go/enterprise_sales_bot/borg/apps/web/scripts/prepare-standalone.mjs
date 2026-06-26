#!/usr/bin/env node
import { cpSync, existsSync, mkdirSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const webDir = resolve(__dirname, '..');
const nextDir = resolve(webDir, '.next');
const standaloneWebDir = resolve(nextDir, 'standalone', 'apps', 'web');
const staticSourceDir = resolve(nextDir, 'static');
const staticTargetDir = resolve(standaloneWebDir, '.next', 'static');
const publicDir = resolve(webDir, 'public');
const publicTargetDir = resolve(standaloneWebDir, 'public');

function copyRecursiveIfPresent(source, target, label) {
  if (!existsSync(source)) {
    return false;
  }

  mkdirSync(dirname(target), { recursive: true });
  cpSync(source, target, { recursive: true, force: true });
  console.log(`[prepare-standalone] copied ${label}: ${source} -> ${target}`);
  return true;
}

if (!existsSync(standaloneWebDir)) {
  console.warn('[prepare-standalone] standalone output not found; skipping asset copy');
  process.exit(0);
}

copyRecursiveIfPresent(staticSourceDir, staticTargetDir, 'Next static assets');
copyRecursiveIfPresent(publicDir, publicTargetDir, 'public assets');
