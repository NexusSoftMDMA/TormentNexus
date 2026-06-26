#!/usr/bin/env node
/**
 * Sync all workspace package.json versions to match the root VERSION file.
 * Run via: node scripts/sync-versions.mjs
 *
 * This ensures every internal @tormentnexus/*, @extension/*, and app package
 * stays at the same version as the monorepo VERSION file.
 */

import { readFileSync, writeFileSync, existsSync } from 'fs';
import { globSync } from 'glob';

import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = join(__dirname, '..');

// Read the canonical version
const versionFile = `${rootDir}/VERSION`;
if (!existsSync(versionFile)) {
  console.error('VERSION file not found at', versionFile);
  process.exit(1);
}
const version = readFileSync(versionFile, 'utf8').trim();

// All package.json paths that should be synced
const patterns = [
  'packages/*/package.json',
  'apps/web/package.json',
  'apps/tormentnexus-extension/package.json',
  'apps/vscode/package.json',
  'apps/tormentnexus-extension/packages/*/package.json',
  'cli/mcp-router-cli/package.json',
];

let updated = 0;
let checked = 0;

for (const pattern of patterns) {
  const files = globSync(pattern, { cwd: rootDir });
  for (const file of files) {
    const fullPath = `${rootDir}/${file}`;
    const pkg = JSON.parse(readFileSync(fullPath, 'utf8'));
    checked++;
    if (pkg.version !== version) {
      console.log(`  ${pkg.name || file}: ${pkg.version} -> ${version}`);
      pkg.version = version;
      writeFileSync(fullPath, JSON.stringify(pkg, null, 2) + '\n');
      updated++;
    }
  }
}

// Also update root package.json
const rootPkgPath = `${rootDir}/package.json`;
const rootPkg = JSON.parse(readFileSync(rootPkgPath, 'utf8'));
if (rootPkg.version !== version) {
  console.log(`  ${rootPkg.name}: ${rootPkg.version} -> ${version}`);
  rootPkg.version = version;
  writeFileSync(rootPkgPath, JSON.stringify(rootPkg, null, 2) + '\n');
  updated++;
}
checked++;

// Also update Go buildinfo default version
const buildinfoPath = `${rootDir}/go/internal/buildinfo/buildinfo.go`;
if (existsSync(buildinfoPath)) {
  const content = readFileSync(buildinfoPath, 'utf8');
  const updated_content = content.replace(
    /var Version = "[^"]*"/,
    `var Version = "${version}"`
  );
  if (content !== updated_content) {
    writeFileSync(buildinfoPath, updated_content);
    console.log(`  Go buildinfo: -> ${version}`);
    updated++;
  }
  checked++;
}

console.log(`\nChecked: ${checked} | Updated: ${updated} | Target: ${version}`);
if (updated === 0) {
  console.log('All versions already in sync.');
}
