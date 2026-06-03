#!/usr/bin/env node

import { spawnSync } from 'node:child_process';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, '..');
const pnpmCommand = process.platform === 'win32' ? 'pnpm.cmd' : 'pnpm';
const goCommand = process.platform === 'win32' ? 'go.exe' : 'go';

const args = new Set(process.argv.slice(2));
const profile = args.has('--profile=workspace') || args.has('--workspace')
  ? 'workspace'
  : 'go-primary';

function printStep(message) {
  console.log(`\n[startup-build] ${message}`);
}

function fail(message, result) {
  const details = result
    ? [
        result.error ? `error=${String(result.error)}` : null,
        typeof result.status === 'number' ? `status=${result.status}` : null,
        result.signal ? `signal=${result.signal}` : null,
      ].filter(Boolean).join(' ')
    : '';
  throw new Error(details ? `${message} (${details})` : message);
}

function run(command, commandArgs, options = {}) {
  return spawnSync(command, commandArgs, {
    cwd: repoRoot,
    stdio: 'inherit',
    shell: false,
    env: process.env,
    ...options,
  });
}

function runPnpm(commandArgs, options = {}) {
  const direct = run(pnpmCommand, commandArgs, options);
  if (!direct.error) {
    return direct;
  }

  return spawnSync(`pnpm ${commandArgs.join(' ')}`, [], {
    cwd: repoRoot,
    stdio: 'inherit',
    shell: true,
    env: process.env,
    ...options,
  });
}

function runWorkspaceBuild() {
  printStep('Running full workspace build for Node compatibility surfaces...');
  const result = runPnpm(['run', 'build:workspace']);
  if ((result.status ?? 1) !== 0) {
    fail('Workspace startup build failed', result);
  }
}

function runGoPrimaryBuild() {
  printStep('Running Go-primary startup build (Go control plane + CLI)...');

  // Build core first since CLI depends on its dist output
  const coreBuild = runPnpm(['-C', 'packages/core', 'run', 'build']);
  if ((coreBuild.status ?? 1) !== 0) {
    fail('Core startup build failed', coreBuild);
  }

  const cliBuild = runPnpm(['-C', 'packages/cli', 'run', 'build']);
  if ((cliBuild.status ?? 1) !== 0) {
    fail('CLI startup build failed', cliBuild);
  }

  const goBuild = run(goCommand, ['build', '-buildvcs=false', './cmd/tormentnexus'], {
    cwd: path.join(repoRoot, 'go'),
  });
  if ((goBuild.status ?? 1) !== 0) {
    fail('Go control-plane build failed', goBuild);
  }

  if (process.env.TORMENTNEXUS_STARTUP_BUILD_WEB === '1') {
    printStep('TORMENTNEXUS_STARTUP_BUILD_WEB=1 set; validating web dashboard build too...');
    const webBuild = runPnpm(['-C', 'apps/web', 'run', 'build']);
    if ((webBuild.status ?? 1) !== 0) {
      fail('Web dashboard startup build failed', webBuild);
    }
  } else {
    printStep('Skipping dashboard/web build in Go-primary startup mode. Set TORMENTNEXUS_STARTUP_BUILD_WEB=1 to include it.');
  }
}

function checkBetterSqlite3Bindings() {
    // On Node 24, better-sqlite3 native bindings must be rebuilt after pnpm install.
    // This check ensures the .node file exists before startup to prevent cascading failures.
    printStep('Checking better-sqlite3 native bindings...');
    const check = run(process.execPath, [
        '-e',
        "try { require('better-sqlite3')(':memory:'); console.log('OK') } catch(e) { console.error('FAIL:' + e.message); process.exit(1) }",
    ], { stdio: 'pipe' });

    if ((check.status ?? 1) === 0) {
        printStep('better-sqlite3 native bindings are functional.');
        return;
    }

    printStep('better-sqlite3 bindings missing or broken. Running pnpm rebuild...');
    const rebuild = runPnpm(['rebuild', 'better-sqlite3']);
    if ((rebuild.status ?? 1) !== 0) {
        console.warn('[startup-build] WARNING: better-sqlite3 rebuild failed. SQLite-backed features will be unavailable.');
        console.warn('[startup-build] You can try manually: pnpm rebuild better-sqlite3');
    } else {
        printStep('better-sqlite3 rebuilt successfully.');
    }
}

try {
  checkBetterSqlite3Bindings();

  if (profile === 'workspace') {
    runWorkspaceBuild();
  } else {
    runGoPrimaryBuild();
  }

  printStep(`Startup build completed successfully for profile: ${profile}`);
} catch (error) {
  console.error(`\n[startup-build] ${error instanceof Error ? error.message : String(error)}`);
  process.exit(1);
}
