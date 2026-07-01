#!/usr/bin/env node
import fs from 'node:fs/promises';
import path from 'node:path';

const repoRoot = process.cwd();
const targetDirs = [
  path.join(repoRoot, 'apps', 'web', 'src'),
  // path.join(repoRoot, 'packages', 'core', 'src'),
];
const fileExtensions = new Set(['.ts', '.tsx', '.js', '.jsx', '.md']);
const blockedPatterns = [
  /router is currently disabled/i,
  /router disabled/i,
  /router not active/i,
  /search router disabled/i,
  /vscode router disabled/i,
  /config router is not active/i,
  /autotest router not active/i,
  /enable the .* router in trpc\.ts/i,
  /static placeholder/i,
  /using local no-op implementations/i,
];

async function walk(dir) {
  const out = [];
  const entries = await fs.readdir(dir, { withFileTypes: true });
  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      out.push(...await walk(full));
      continue;
    }
    if (entry.isFile() && fileExtensions.has(path.extname(entry.name))) {
      out.push(full);
    }
  }
  return out;
}

function relative(filePath) {
  return path.relative(repoRoot, filePath).replace(/\\/g, '/');
}

async function main() {
  const files = (await Promise.all(targetDirs.map(walk))).flat();
  const violations = [];

  for (const file of files) {
    const content = await fs.readFile(file, 'utf8');
    const lines = content.split(/\r?\n/u);

    lines.forEach((line, index) => {
      for (const pattern of blockedPatterns) {
        if (pattern.test(line)) {
          violations.push({ file: relative(file), line: index + 1, text: line.trim() });
          break;
        }
      }
    });
  }

  if (violations.length > 0) {
    console.error('\n[placeholder-check] Found blocked placeholder/no-op regression markers:\n');
    for (const violation of violations) {
      console.error(`- ${violation.file}:${violation.line}`);
      console.error(`  ${violation.text}`);
    }
    console.error(`\nTotal violations: ${violations.length}`);
    process.exit(1);
  }

  console.log('[placeholder-check] OK: no blocked placeholder/no-op regression markers found.');
}

main().catch((error) => {
  console.error('[placeholder-check] Failed:', error);
  process.exit(1);
});
