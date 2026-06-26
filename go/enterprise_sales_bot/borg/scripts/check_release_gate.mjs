#!/usr/bin/env node

import { spawnSync } from 'node:child_process';

const usePnpm = process.platform === 'win32' ? 'pnpm.cmd' : 'pnpm';
const useNode = process.execPath;
const args = new Set(process.argv.slice(2));
const skipReadiness = args.has('--skip-readiness');
const includeTurboLint = args.has('--with-turbo-lint');
const includeDashboardSmoke = args.has('--with-dashboard-smoke');

function run(name, command, commandArgs, options = {}) {
  const result = spawnSync(command, commandArgs, {
    cwd: process.cwd(),
    encoding: 'utf-8',
    ...options,
  });

  return { name, command, args: commandArgs, ...result };
}

function runPnpm(name, commandArgs, options = {}) {
  const direct = run(name, usePnpm, commandArgs, {
    shell: false,
    ...options,
  });

  if (!direct.error) {
    return direct;
  }

  return run(name, `pnpm ${commandArgs.join(' ')}`, [], {
    shell: true,
    ...options,
  });
}

function formatFailure(result) {
  return [
    result.error ? `error=${String(result.error)}` : null,
    typeof result.status === 'number' ? `status=${result.status}` : null,
    result.signal ? `signal=${result.signal}` : null,
  ].filter(Boolean).join(' ');
}

function fail(message, details) {
  console.error(`\n[release-gate] ${message}`);
  if (details) {
    console.error(details);
  }
  process.exit(1);
}

function printStep(message) {
  console.log(`\n[release-gate] ${message}`);
}

async function main() {
  if (!skipReadiness) {
    printStep('Running strict readiness probe...');
    const readiness = run('readiness', useNode, ['scripts/verify_dev_readiness.mjs', '--strict-json'], {
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    if (readiness.error) {
      fail('Readiness probe failed to execute.', String(readiness.error));
    }

    if ((readiness.status ?? 1) !== 0) {
      fail('Readiness probe reported failure.', [readiness.stdout, readiness.stderr].filter(Boolean).join('\n'));
    }

    let readinessPayload;
    try {
      readinessPayload = JSON.parse(readiness.stdout);
    } catch {
      fail('Readiness output was not valid JSON.', readiness.stdout || readiness.stderr);
    }

    if (!readinessPayload?.passed) {
      fail('Readiness JSON indicates failed critical services.', JSON.stringify(readinessPayload, null, 2));
    }

    console.log('[release-gate] Readiness OK');
  } else {
    printStep('Skipping readiness probe (--skip-readiness).');
  }

  printStep('Running placeholder regression check...');
  const placeholder = runPnpm('placeholder-check', ['run', 'check:placeholders'], { stdio: 'inherit' });
  if ((placeholder.status ?? 1) !== 0) {
    fail(`Placeholder regression check failed. ${formatFailure(placeholder)}`);
  }

  printStep('Running core typecheck...');
  const coreTypecheck = runPnpm('core-typecheck', ['-C', 'packages/core', 'exec', 'tsc', '--noEmit'], { stdio: 'inherit' });
  if ((coreTypecheck.status ?? 1) !== 0) {
    fail(`Core typecheck failed. ${formatFailure(coreTypecheck)}`);
  }

  printStep('Running CLI typecheck...');
  const cliTypecheck = runPnpm('cli-typecheck', ['-C', 'packages/cli', 'exec', 'tsc', '--noEmit'], { stdio: 'inherit' });
  if ((cliTypecheck.status ?? 1) !== 0) {
    fail(`CLI typecheck failed. ${formatFailure(cliTypecheck)}`);
  }

  printStep('Running dashboard production build...');
  const webBuild = runPnpm('web-build', ['-C', 'apps/web', 'build'], { stdio: 'inherit' });
  if ((webBuild.status ?? 1) !== 0) {
    fail(`Dashboard build failed. ${formatFailure(webBuild)}`);
  }

  if (includeDashboardSmoke) {
    printStep('Running dashboard smoke test...');
    const smoke = run(useNode, useNode, ['scripts/dashboard-smoke.cjs'], { stdio: 'inherit' });
    if ((smoke.status ?? 1) !== 0) {
      fail(`Dashboard smoke test failed. ${formatFailure(smoke)}`);
    }
  } else {
    printStep('Skipping dashboard smoke test (pass --with-dashboard-smoke to enable).');
  }

  if (includeTurboLint) {
    printStep('Running scoped Turbo lint...');
    const turboLint = runPnpm('turbo-lint', ['run', 'lint:turbo'], { stdio: 'inherit' });
    if ((turboLint.status ?? 1) !== 0) {
      fail(`Scoped Turbo lint failed. ${formatFailure(turboLint)}`);
    }
  }

  printStep('All release gate checks passed ✅');
}

main().catch((error) => {
  fail('Unexpected release gate error.', String(error));
});
