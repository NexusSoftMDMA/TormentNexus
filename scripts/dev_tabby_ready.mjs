#!/usr/bin/env node

import { execFile, spawn } from 'node:child_process';

import {
  detectBrowserExtensionArtifacts,
  getTormentNexusStartLockPath,
  getPreferredWebPorts,
  getWaitingReasons,
  chooseStaleCoreRefreshTarget,
  isCompatibleStartupStatusContract,
  isLikelyTormentNexusCoreCommand,
  isHttpProbeResponsive,
  isDirectExecution,
  parseListeningPidFromLsof,
  parseListeningPidFromNetstat,
  readTormentNexusStartLockRecord,
  summarizeBrowserExtensionArtifacts,
  waitForCoreBridgeShutdown,
} from './dev_tabby_ready_helpers.mjs';

const WEB_PORT_CANDIDATES = [3000, 3010, 3020, 3030, 3040];
const CORE_HTTP_BASE = 'http://127.0.0.1:3001';
const POLL_INTERVAL_MS = Number(process.env.TORMENTNEXUS_DEV_READY_POLL_MS || 2000);
const READY_TIMEOUT_MS = Number(process.env.TORMENTNEXUS_DEV_READY_TIMEOUT_MS || 600000);
const WEB_DETECT_TIMEOUT_MS = Number(process.env.TORMENTNEXUS_DEV_READY_WEB_TIMEOUT_MS || 15000);
const TRPC_QUERY_TIMEOUT_MS = Number(process.env.TORMENTNEXUS_DEV_READY_TRPC_TIMEOUT_MS || 12000);
const AUTO_OPEN_DASHBOARD = process.env.TORMENTNEXUS_DEV_READY_OPEN_BROWSER !== '0';
const AUTO_REFRESH_STALE_CORE = process.env.TORMENTNEXUS_DEV_READY_RESTART_STALE_CORE !== '0';
const REPO_ROOT = process.cwd();
const CORE_BRIDGE_PROBE_URLS = [
  `${CORE_HTTP_BASE}/api/mesh/stream`,
  `${CORE_HTTP_BASE}/health`,
];

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function fetchStatus(url, timeoutMs = 4000) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      method: 'GET',
      signal: controller.signal,
      redirect: 'follow',
    });
    return { ok: response.ok, status: response.status };
  } catch (error) {
    return {
      ok: false,
      status: null,
      error: error instanceof Error ? error.message : String(error),
    };
  } finally {
    clearTimeout(timeout);
  }
}

async function detectWebPort() {
  for (const port of getPreferredWebPorts(REPO_ROOT, WEB_PORT_CANDIDATES)) {
    const candidates = [
      `http://127.0.0.1:${port}/api/trpc/startupStatus?input=%7B%7D`,
      `http://127.0.0.1:${port}/dashboard`,
      `http://127.0.0.1:${port}/`,
    ];

    for (const url of candidates) {
      const status = await fetchStatus(url, WEB_DETECT_TIMEOUT_MS);
      if (isHttpProbeResponsive(status)) {
        return { port, url };
      }
    }
  }

  return null;
}

function extractTrpcData(payload) {
  if (Array.isArray(payload) && payload.length > 0) {
    return extractTrpcData(payload[0]);
  }

  if (!payload || typeof payload !== 'object') {
    return undefined;
  }

  const result = payload.result;
  if (!result || typeof result !== 'object') {
    return undefined;
  }

  const data = result.data;
  if (!data || typeof data !== 'object') {
    return data;
  }

  if (Object.hasOwn(data, 'json')) {
    return data.json;
  }

  return data;
}

async function fetchTrpcQuery(webPort, procedureName, input = {}) {
  const url = new URL(`http://127.0.0.1:${webPort}/api/trpc/${procedureName}`);
  url.searchParams.set('input', JSON.stringify(input));

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), TRPC_QUERY_TIMEOUT_MS);

  try {
    const response = await fetch(url.toString(), {
      method: 'GET',
      signal: controller.signal,
      redirect: 'follow',
    });

    if (!response.ok) {
      return { ok: false, url: url.toString(), status: response.status, data: null };
    }

    const payload = await response.json();
    return { ok: true, url: url.toString(), status: response.status, data: extractTrpcData(payload) };
  } catch (error) {
    return {
      ok: false,
      url: url.toString(),
      status: null,
      data: null,
      error: error instanceof Error ? error.message : String(error),
    };
  } finally {
    clearTimeout(timeout);
  }
}

async function fetchTrpcMutation(webPort, procedureName, input = {}) {
  const url = `http://127.0.0.1:${webPort}/api/trpc/${procedureName}`;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), TRPC_QUERY_TIMEOUT_MS);

  try {
    const response = await fetch(url, {
      method: 'POST',
      signal: controller.signal,
      redirect: 'follow',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify({ input }),
    });

    if (!response.ok) {
      return { ok: false, url, status: response.status, data: null };
    }

    const payload = await response.json();
    return { ok: true, url, status: response.status, data: extractTrpcData(payload) };
  } catch (error) {
    return {
      ok: false,
      url,
      status: null,
      data: null,
      error: error instanceof Error ? error.message : String(error),
    };
  } finally {
    clearTimeout(timeout);
  }
}

function getBrowserOpenCommand(url) {
  if (process.platform === 'win32') {
    return {
      command: 'cmd.exe',
      args: ['/d', '/s', '/c', 'start', '', url],
    };
  }

  if (process.platform === 'darwin') {
    return {
      command: 'open',
      args: [url],
    };
  }

  return {
    command: 'xdg-open',
    args: [url],
  };
}

function openDashboardInBrowser(url) {
  if (!AUTO_OPEN_DASHBOARD) {
    return;
  }

  const { command, args } = getBrowserOpenCommand(url);
  const child = spawn(command, args, {
    cwd: REPO_ROOT,
    detached: true,
    stdio: 'ignore',
  });

  child.unref();
}

function getPnpmSpawnSpec(commandArgs, cwd = REPO_ROOT) {
  if (process.platform === 'win32') {
    return {
      command: 'cmd.exe',
      args: ['/d', '/s', '/c', 'pnpm', ...commandArgs],
      cwd,
    };
  }

  return {
    command: 'pnpm',
    args: commandArgs,
    cwd,
  };
}

function execFileText(command, args) {
  return new Promise((resolve) => {
    execFile(command, args, { cwd: REPO_ROOT, windowsHide: true }, (error, stdout) => {
      if (error) {
        resolve('');
        return;
      }

      resolve(stdout);
    });
  });
}

function runPnpmCommand(commandArgs, cwd = REPO_ROOT) {
  const { command, args } = getPnpmSpawnSpec(commandArgs, cwd);

  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd,
      stdio: 'inherit',
      env: {
        ...process.env,
        CI: process.env.CI ?? 'true',
      },
    });

    child.on('error', (error) => reject(error));
    child.on('exit', (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }

      reject(new Error(`pnpm ${commandArgs.join(' ')} failed (code=${code ?? 'null'} signal=${signal ?? 'null'})`));
    });
  });
}

async function ensureOfficialBrowserExtensionArtifacts() {
  const extensionRoot = `${REPO_ROOT}${process.platform === 'win32' ? '\\' : '/'}apps${process.platform === 'win32' ? '\\' : '/'}tormentnexus-extension`;

  console.log('[TormentNexus Dev Ready] official browser-extension artifacts missing; building Chromium + Firefox bundles...');

  await runPnpmCommand(['-C', 'apps/tormentnexus-extension', 'run', 'build'], REPO_ROOT);
  await runPnpmCommand(['-C', 'apps/tormentnexus-extension', 'run', 'build:firefox'], REPO_ROOT);

  console.log(`[TormentNexus Dev Ready] browser-extension artifacts refreshed from ${extensionRoot}`);
}

async function warmMcpAndMemory(webPort) {
  await Promise.allSettled([
    fetchTrpcQuery(webPort, 'startupStatus', {}),
    fetchTrpcQuery(webPort, 'mcp.listServers', {}),
    fetchTrpcQuery(webPort, 'mcp.listTools', {}),
    fetchTrpcQuery(webPort, 'mcp.getStatus', {}),
    fetchTrpcQuery(webPort, 'mcp.searchTools', { query: '' }),
    fetchTrpcQuery(webPort, 'memory.getAgentStats', {}),
    fetchTrpcQuery(webPort, 'memory.listContexts', {}),
    fetchTrpcQuery(webPort, 'browser.status', {}),
    fetchTrpcQuery(webPort, 'session.list', {}),
    fetchTrpcQuery(webPort, 'session.getState', {}),
    fetchTrpcMutation(webPort, 'session.restore', {}),
  ]);
}

async function evaluateReadiness() {
  const web = await detectWebPort();
  const coreBridgeChecks = await Promise.all(CORE_BRIDGE_PROBE_URLS.map((url) => fetchStatus(url)));
  const coreBridge = coreBridgeChecks.find((check) => check.ok) ?? coreBridgeChecks[0] ?? { ok: false, status: null };
  const extensions = detectBrowserExtensionArtifacts(REPO_ROOT);
  const extension = summarizeBrowserExtensionArtifacts(extensions);

  let mcpStatus = { ok: false, data: null, url: null };
  let memoryStatus = { ok: false, data: null, url: null };
  let browserStatus = { ok: false, data: null, url: null };
  let sessionStatus = { ok: false, data: null, url: null };
  let startupStatus = { ok: false, compatible: false, data: null, url: null };

  if (web) {
    const [startupResult, mcpResult, memoryResult, browserResult, sessionResult] = await Promise.all([
      fetchTrpcQuery(web.port, 'startupStatus', {}),
      fetchTrpcQuery(web.port, 'mcp.getStatus', {}),
      fetchTrpcQuery(web.port, 'memory.getAgentStats', {}),
      fetchTrpcQuery(web.port, 'browser.status', {}),
      fetchTrpcQuery(web.port, 'session.list', {}),
    ]);

    startupStatus = {
      ok: startupResult.ok,
      compatible: startupResult.ok && isCompatibleStartupStatusContract(startupResult.data),
      data: startupResult.data,
      url: startupResult.url,
    };

    mcpStatus = {
      ok: mcpResult.ok,
      data: mcpResult.data,
      url: mcpResult.url,
    };

    memoryStatus = {
      ok: memoryResult.ok,
      data: memoryResult.data,
      url: memoryResult.url,
    };

    browserStatus = {
      ok: browserResult.ok,
      data: browserResult.data,
      url: browserResult.url,
    };

    sessionStatus = {
      ok: sessionResult.ok,
      data: sessionResult.data,
      url: sessionResult.url,
    };
  }

  const startupSnapshotReady = Boolean(startupStatus.data?.ready);
  const startupContractCompatible = startupStatus.ok && startupStatus.compatible;

  const ready = Boolean(web)
    && coreBridge.ok
    && startupContractCompatible
    && (startupSnapshotReady || (mcpStatus.ok
      && memoryStatus.ok
      && browserStatus.ok
      && sessionStatus.ok))
    && extension.ready;

  return {
    ready,
    web,
    coreBridge,
    startupStatus,
    mcpStatus,
    memoryStatus,
    browserStatus,
    sessionStatus,
    extensions,
    extension,
  };
}

function printReadySummary(state) {
  const webPort = state.web?.port;
  const dashboardUrl = webPort ? `http://127.0.0.1:${webPort}` : 'unavailable';

  console.log('\n[TormentNexus Dev Ready] ✅ stack is ready');
  console.log(`[TormentNexus Dev Ready] Dashboard: ${dashboardUrl}`);
  console.log('[TormentNexus Dev Ready] Core bridge: ws://127.0.0.1:3001 (HTTP probe: /api/mesh/stream or /health)');
  console.log(`[TormentNexus Dev Ready] Startup telemetry API: ${state.startupStatus.url ?? 'unavailable'}`);
  console.log(`[TormentNexus Dev Ready] MCP telemetry API: ${state.mcpStatus.url ?? 'unavailable'}`);
  console.log(`[TormentNexus Dev Ready] Memory telemetry API: ${state.memoryStatus.url ?? 'unavailable'}`);
  console.log(`[TormentNexus Dev Ready] Browser telemetry API: ${state.browserStatus.url ?? 'unavailable'}`);
  console.log(`[TormentNexus Dev Ready] Session telemetry API: ${state.sessionStatus.url ?? 'unavailable'}`);
  console.log(`[TormentNexus Dev Ready] Extension artifacts: ${state.extension.summary ?? 'unavailable'}`);
  for (const artifact of state.extensions) {
    console.log(`[TormentNexus Dev Ready] ${artifact.label}: ${artifact.artifactPath ?? 'unavailable'}`);
  }
}

function printWaitingSummary(state, elapsedMs) {
  const missing = getWaitingReasons(state);

  console.log(`[TormentNexus Dev Ready] connecting ${Math.floor(elapsedMs / 1000)}s: ${missing.join(' | ')}`);
}

function spawnTurboDev() {
  const env = {
    ...process.env,
    CI: 'true',
    TURBO_DAEMON: 'false',
  };

  const command = process.platform === 'win32' ? 'cmd.exe' : 'pnpm';
  const turboArgs = [
    'turbo',
    'run',
    'dev',
    '--concurrency',
    '22',
    '--filter=!@extension/hmr',
    '--filter=!opencode-autopilot',
    '--filter=!@tormentnexus/cli',
    '--filter=!@opencode-autopilot/cli',
    '--filter=!@opencode-autopilot/server',
    '--filter=!@opencode-autopilot/shared',
    '--filter=!backend',
    '--filter=!frontend',
    '--filter=!@repo/*',
  ];
  const args = process.platform === 'win32'
    ? ['/d', '/s', '/c', 'pnpm', ...turboArgs]
    : turboArgs;

  return spawn(command, args, {
    cwd: REPO_ROOT,
    stdio: 'inherit',
    env,
  });
}

function spawnCliDev() {
  const env = {
    ...process.env,
    CI: 'true',
  };

  const { command, args, cwd } = getPnpmSpawnSpec([
    '-C',
    'packages/cli',
    'exec',
    'tsx',
    'src/index.ts',
    'start',
    '--port',
    '4100',
  ]);

  return spawn(command, args, {
    cwd,
    stdio: 'inherit',
    env,
  });
}

async function detectExistingCoreBridge() {
  const checks = await Promise.all(CORE_BRIDGE_PROBE_URLS.map((url) => fetchStatus(url)));

  return checks.some((check) => check.ok);
}

async function detectCoreBridgeOwnerPid() {
  if (process.platform === 'win32') {
    const output = await execFileText('netstat', ['-ano', '-p', 'tcp']);
    return parseListeningPidFromNetstat(output, 3001);
  }

  const output = await execFileText('lsof', ['-nP', '-iTCP:3001', '-sTCP:LISTEN', '-t']);
  return parseListeningPidFromLsof(output);
}

async function readProcessCommandLine(pid) {
  if (typeof pid !== 'number' || pid <= 0) {
    return '';
  }

  if (process.platform === 'win32') {
    const output = await execFileText('powershell.exe', [
      '-NoProfile',
      '-Command',
      `$p = Get-CimInstance Win32_Process -Filter "ProcessId = ${pid}" | Select-Object -ExpandProperty CommandLine; if ($p) { $p }`,
    ]);

    return output.trim();
  }

  const output = await execFileText('ps', ['-p', String(pid), '-o', 'args=']);
  return output.trim();
}

async function detectCoreBridgeOwner() {
  const pid = await detectCoreBridgeOwnerPid();
  if (!pid) {
    return null;
  }

  const commandLine = await readProcessCommandLine(pid);
  return {
    pid,
    commandLine,
    trusted: isLikelyTormentNexusCoreCommand(commandLine),
  };
}

async function stopExistingCoreBridge(pid, sourceLabel) {
  if (typeof pid !== 'number' || pid <= 0 || pid === process.pid) {
    return false;
  }

  try {
    process.kill(pid, 'SIGTERM');
  } catch (error) {
    console.warn(`[TormentNexus Dev Ready] could not terminate stale TormentNexus core PID ${pid} from ${sourceLabel}: ${error instanceof Error ? error.message : String(error)}`);
    return false;
  }

  const stopped = await waitForCoreBridgeShutdown(
    CORE_BRIDGE_PROBE_URLS,
    {
      timeoutMs: TRPC_QUERY_TIMEOUT_MS,
      pollIntervalMs: POLL_INTERVAL_MS,
    },
    {
      probeImpl: fetchStatus,
      waitImpl: sleep,
    },
  );

  if (!stopped) {
    console.warn(`[TormentNexus Dev Ready] stale TormentNexus core PID ${pid} from ${sourceLabel} did not release the bridge within ${Math.floor(TRPC_QUERY_TIMEOUT_MS / 1000)}s.`);
  }

  return stopped;
}

function waitForChildExit(label, child) {
  return new Promise((resolve, reject) => {
    child.once('exit', (code, signal) => {
      if (signal) {
        process.kill(process.pid, signal);
        return;
      }

      if ((code ?? 1) === 0) {
        resolve();
        return;
      }

      reject(new Error(`${label} exited (${code ?? 1})`));
    });
  });
}

async function main() {
  const child = spawnTurboDev();
  let reuseExistingCoreBridge = await detectExistingCoreBridge();
  let cliChild = reuseExistingCoreBridge ? null : spawnCliDev();
  let childExit = null;
  let cliChildExit = null;
  let attemptedStaleCoreRefresh = false;

  if (reuseExistingCoreBridge) {
    console.log('[TormentNexus Dev Ready] reusing existing core bridge on port 3001; skipping duplicate CLI launch.');
  }

  const attachCliChild = (nextCliChild) => {
    cliChild = nextCliChild;
    cliChildExit = null;

    if (!nextCliChild) {
      return;
    }

    nextCliChild.on('error', (error) => {
      console.error(`[TormentNexus Dev Ready] failed to start CLI server: ${error instanceof Error ? error.message : String(error)}`);
      process.exit(1);
    });

    nextCliChild.on('exit', (code, signal) => {
      cliChildExit = { code, signal };
    });
  };

  const terminateChild = (signal) => {
    if (!child.killed) {
      child.kill(signal);
    }

    if (cliChild && !cliChild.killed) {
      cliChild.kill(signal);
    }
  };

  process.on('SIGINT', () => terminateChild('SIGINT'));
  process.on('SIGTERM', () => terminateChild('SIGTERM'));

  child.on('error', (error) => {
    console.error(`[TormentNexus Dev Ready] failed to start dev stack: ${error instanceof Error ? error.message : String(error)}`);
    process.exit(1);
  });

  child.on('exit', (code, signal) => {
    childExit = { code, signal };
  });

  attachCliChild(cliChild);

  const startedAt = Date.now();
  let warmed = false;
  let dashboardOpened = false;
  let extensionBuildPromise = null;

  while (Date.now() - startedAt <= READY_TIMEOUT_MS) {
    if (childExit) {
      const exitDescriptor = childExit.signal
        ? `signal=${childExit.signal}`
        : `code=${childExit.code ?? 1}`;

      throw new Error(`dev stack exited before readiness (${exitDescriptor})`);
    }

    if (cliChildExit) {
      const exitDescriptor = cliChildExit.signal
        ? `signal=${cliChildExit.signal}`
        : `code=${cliChildExit.code ?? 1}`;

      throw new Error(`CLI server exited before readiness (${exitDescriptor})`);
    }

    const state = await evaluateReadiness();

    if (!state.extension.ready && !extensionBuildPromise) {
      extensionBuildPromise = ensureOfficialBrowserExtensionArtifacts().catch((error) => {
        console.warn(`[TormentNexus Dev Ready] browser-extension build failed: ${error instanceof Error ? error.message : String(error)}`);
      });
    }

    if (state.ready) {
      if (state.web && !warmed) {
        warmed = true;
        await warmMcpAndMemory(state.web.port);
      }

      if (state.web && !dashboardOpened) {
        dashboardOpened = true;
        openDashboardInBrowser(`http://127.0.0.1:${state.web.port}/dashboard`);
      }

      printReadySummary(state);
      break;
    }

    if (
      reuseExistingCoreBridge
      && state.startupStatus.ok
      && state.startupStatus.compatible === false
    ) {
      if (!attemptedStaleCoreRefresh && AUTO_REFRESH_STALE_CORE) {
        attemptedStaleCoreRefresh = true;
        const lockPath = getTormentNexusStartLockPath();
        const lockRecord = readTormentNexusStartLockRecord(lockPath);
        const owner = lockRecord ? null : await detectCoreBridgeOwner();
        const refreshTarget = chooseStaleCoreRefreshTarget({
          lockRecord,
          owner,
          currentPid: process.pid,
        });

        if (refreshTarget.kind === 'lock' && lockRecord) {
          console.warn(`[TormentNexus Dev Ready] existing core bridge is healthy but serving an older startup contract; stopping TormentNexus core PID ${lockRecord.pid} from ${lockPath} and starting a fresh CLI instance.`);
          const stopped = await stopExistingCoreBridge(refreshTarget.pid, lockPath);

          if (stopped) {
            reuseExistingCoreBridge = false;
            attachCliChild(spawnCliDev());
            await sleep(POLL_INTERVAL_MS);
            continue;
          }

          console.warn('[TormentNexus Dev Ready] existing core bridge is healthy but serving an older startup contract; the locked TormentNexus core could not be stopped automatically.');
        } else if (refreshTarget.kind === 'owner') {
          console.warn(`[TormentNexus Dev Ready] existing core bridge is healthy but serving an older startup contract; stopping TormentNexus-owned bridge PID ${refreshTarget.pid} discovered from port 3001 and starting a fresh CLI instance.`);
          const stopped = await stopExistingCoreBridge(refreshTarget.pid, refreshTarget.sourceLabel);

          if (stopped) {
            reuseExistingCoreBridge = false;
            attachCliChild(spawnCliDev());
            await sleep(POLL_INTERVAL_MS);
            continue;
          }

          console.warn('[TormentNexus Dev Ready] existing core bridge is healthy but serving an older startup contract; the TormentNexus-owned port listener could not be stopped automatically.');
        } else if (refreshTarget.kind === 'skip-untrusted-owner') {
          console.warn(`[TormentNexus Dev Ready] existing core bridge is healthy but serving an older startup contract; port 3001 is owned by PID ${refreshTarget.pid}, but its command line did not look TormentNexus-owned, so automatic refresh was skipped.`);
        } else {
          console.warn('[TormentNexus Dev Ready] existing core bridge is healthy but serving an older startup contract; no TormentNexus startup lock or port owner PID was found, so automatic refresh was skipped.');
        }
      }

      console.warn('[TormentNexus Dev Ready] existing core bridge is healthy but serving an older startup contract; restart the TormentNexus CLI/core bridge so `pnpm run dev` can validate the current readiness payload.');
    }

    printWaitingSummary(state, Date.now() - startedAt);
    await sleep(POLL_INTERVAL_MS);
  }

  if (!warmed) {
    console.warn(`\n[TormentNexus Dev Ready] ⚠ readiness timeout (${Math.floor(READY_TIMEOUT_MS / 1000)}s). Keeping dev stack running for manual inspection.`);
  }

  await Promise.all([
    waitForChildExit('dev stack', child),
    ...(cliChild ? [waitForChildExit('CLI server', cliChild)] : []),
  ]);
}

if (isDirectExecution(import.meta.url, process.argv[1])) {
  main().catch((error) => {
    console.error(`[TormentNexus Dev Ready] unexpected error: ${error instanceof Error ? error.message : String(error)}`);
    process.exit(1);
  });
}
