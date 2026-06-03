import fs from 'node:fs';
import { homedir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const WEB_DEV_PORT_MARKER = ['apps', 'web', '.tormentnexus-dev-port.json'];

const STARTUP_CHECK_LABELS = {
  configSync: 'MCP config sync',
  sessionSupervisor: 'session restore',
  browser: 'browser runtime',
  memory: 'memory initialization',
  extensionBridge: 'extension bridge listener',
};

export function resolveTormentNexusDataDir(dataDir = process.env.TORMENTNEXUS_DEV_READY_DATA_DIR ?? process.env.TORMENTNEXUS_DATA_DIR ?? '~/.tormentnexus') {
  if (typeof dataDir !== 'string' || dataDir.length === 0) {
    return path.join(homedir(), '.tormentnexus');
  }

  if (dataDir === '~') {
    return homedir();
  }

  if (dataDir.startsWith('~/') || dataDir.startsWith('~\\')) {
    return path.resolve(homedir(), dataDir.slice(2));
  }

  return path.resolve(dataDir);
}

export function getTormentNexusStartLockPath(dataDir) {
  return path.join(resolveTormentNexusDataDir(dataDir), 'lock');
}

export function readTormentNexusStartLockRecord(lockPath = getTormentNexusStartLockPath()) {
  try {
    if (!fs.existsSync(lockPath)) {
      return null;
    }

    const parsed = JSON.parse(fs.readFileSync(lockPath, 'utf8'));
    if (
      !parsed
      || typeof parsed !== 'object'
      || typeof parsed.instanceId !== 'string'
      || typeof parsed.pid !== 'number'
      || typeof parsed.port !== 'number'
      || typeof parsed.host !== 'string'
      || typeof parsed.createdAt !== 'string'
    ) {
      return null;
    }

    return parsed;
  } catch {
    return null;
  }
}

export function parseListeningPidFromNetstat(output, port) {
  if (typeof output !== 'string' || !Number.isInteger(port) || port <= 0) {
    return null;
  }

  for (const rawLine of output.split(/\r?\n/u)) {
    const line = rawLine.trim();
    if (!line) {
      continue;
    }

    const columns = line.split(/\s+/u);
    if (columns.length < 5) {
      continue;
    }

    const [protocol, localAddress, , stateOrPid, pidOrEmpty] = columns;
    if (protocol.toUpperCase() !== 'TCP') {
      continue;
    }

    const isListening = stateOrPid.toUpperCase() === 'LISTENING';
    const pidText = isListening ? pidOrEmpty : stateOrPid;
    if (!localAddress.endsWith(`:${port}`) || !/^\d+$/u.test(pidText ?? '')) {
      continue;
    }

    return Number(pidText);
  }

  return null;
}

export function parseListeningPidFromLsof(output) {
  if (typeof output !== 'string') {
    return null;
  }

  const line = output
    .split(/\r?\n/u)
    .map((entry) => entry.trim())
    .find((entry) => /^\d+$/u.test(entry));

  return line ? Number(line) : null;
}

export function isLikelyTormentNexusCoreCommand(commandLine) {
  if (typeof commandLine !== 'string') {
    return false;
  }

  const normalized = commandLine.trim().toLowerCase();
  if (!normalized) {
    return false;
  }

  const tormentnexusMarkers = [
    '@tormentnexus/',
    'packages/core',
    'packages\\core',
    'packages/cli',
    'packages\\cli',
    '/tormentnexus/',
    '\\tormentnexus\\',
    'tsx src/index.ts start',
    'tsx src/server-stdio.ts',
    'backgroundcorebootstrap',
    'server-stdio',
  ];

  return tormentnexusMarkers.some((marker) => normalized.includes(marker));
}

export function chooseStaleCoreRefreshTarget({
  lockRecord = null,
  owner = null,
  currentPid = process.pid,
} = {}) {
  const lockPid = typeof lockRecord?.pid === 'number' ? lockRecord.pid : null;
  if (lockPid && lockPid > 0 && lockPid !== currentPid) {
    return {
      kind: 'lock',
      pid: lockPid,
      sourceLabel: 'lock',
      trusted: true,
    };
  }

  const ownerPid = typeof owner?.pid === 'number' ? owner.pid : null;
  if (ownerPid && ownerPid > 0 && ownerPid !== currentPid && owner?.trusted === true) {
    return {
      kind: 'owner',
      pid: ownerPid,
      sourceLabel: 'port 3001',
      trusted: true,
    };
  }

  if (ownerPid && ownerPid > 0) {
    return {
      kind: 'skip-untrusted-owner',
      pid: ownerPid,
      sourceLabel: 'port 3001',
      trusted: false,
    };
  }

  return {
    kind: 'skip-missing',
    pid: null,
    sourceLabel: null,
    trusted: false,
  };
}

export function isCompatibleStartupStatusContract(startupStatusData) {
  if (!startupStatusData || typeof startupStatusData !== 'object') {
    return false;
  }

  const checks = startupStatusData.checks;
  if (!checks || typeof checks !== 'object') {
    return false;
  }

  const aggregator = checks.mcpAggregator;
  const memory = checks.memory;
  const executionEnvironment = checks.executionEnvironment;

  return Boolean(
    aggregator
      && typeof aggregator === 'object'
      && Object.prototype.hasOwnProperty.call(aggregator, 'residentReady')
      && Object.prototype.hasOwnProperty.call(aggregator, 'residentConnectedCount')
      && Object.prototype.hasOwnProperty.call(aggregator, 'inventorySource')
      && memory
      && typeof memory === 'object'
      && Object.prototype.hasOwnProperty.call(memory, 'claudeMem')
      && executionEnvironment
      && typeof executionEnvironment === 'object'
      && Object.prototype.hasOwnProperty.call(executionEnvironment, 'harnessCount')
      && Object.prototype.hasOwnProperty.call(executionEnvironment, 'verifiedHarnessCount')
  );
}

function getPendingMcpStartupChecks(startupStatusData) {
  const aggregator = startupStatusData?.checks?.mcpAggregator;
  if (!aggregator || typeof aggregator !== 'object') {
    return [];
  }

  const pending = [];
  if (aggregator.inventoryReady === false) {
    pending.push('cached MCP inventory');
  }

  if (aggregator.inventoryReady === undefined && aggregator.ready === false) {
    pending.push('cached MCP inventory');
  }

  const residentTargetCount = Number(aggregator.advertisedAlwaysOnServerCount ?? 0);
  const residentConnectedCount = Number(aggregator.residentConnectedCount ?? 0);
  const residentReady = aggregator.residentReady
    ?? ((aggregator.liveReady ?? aggregator.ready) && residentConnectedCount >= residentTargetCount);

  if (residentReady === false) {
    const warmingCount = Number(aggregator.warmingServerCount ?? 0);
    const failedWarmupCount = Number(aggregator.failedWarmupServerCount ?? 0);
    const posture = [
      warmingCount > 0 ? `${warmingCount} warming` : null,
      failedWarmupCount > 0 ? `${failedWarmupCount} failed` : null,
    ].filter(Boolean).join(', ');

    pending.push(posture ? `resident MCP runtime (${posture})` : 'resident MCP runtime');
  }

  return pending;
}

const BROWSER_EXTENSION_ARTIFACTS = [
  {
    id: 'browser-extension-chromium',
    label: 'browser extension Chromium bundle',
    candidates: [
      ['apps', 'tormentnexus-extension', 'dist-chromium'],
      ['apps', 'tormentnexus-extension', 'dist'],
      ['apps', 'extension', 'dist'],
    ],
    requiredFiles: ['background.js', 'manifest.json'],
  },
  {
    id: 'browser-extension-firefox',
    label: 'browser extension Firefox bundle',
    candidates: [
      ['apps', 'tormentnexus-extension', 'dist-firefox'],
    ],
    requiredFiles: ['background.js', 'manifest.json'],
  },
];

function normalizePathForComparison(filePath) {
  const resolved = path.resolve(filePath);
  return process.platform === 'win32'
    ? resolved.toLowerCase()
    : resolved;
}

export function isDirectExecution(importMetaUrl, argv1 = process.argv[1]) {
  if (!argv1) {
    return false;
  }

  try {
    const importPath = normalizePathForComparison(fileURLToPath(importMetaUrl));
    const argvPath = normalizePathForComparison(argv1);
    return importPath === argvPath;
  } catch {
    return false;
  }
}

export function detectBrowserExtensionArtifacts(repoRoot) {
  return BROWSER_EXTENSION_ARTIFACTS.map((artifact) => {
    const resolvedCandidates = artifact.candidates.map((segments) => path.join(repoRoot, ...segments));
    const existingCandidate = resolvedCandidates.find((candidatePath) => fs.existsSync(candidatePath));
    const artifactPath = existingCandidate ?? resolvedCandidates[0] ?? null;
    const missingFiles = artifact.requiredFiles.filter((fileName) => !artifactPath || !fs.existsSync(path.join(artifactPath, fileName)));

    return {
      id: artifact.id,
      label: artifact.label,
      artifactPath,
      ready: missingFiles.length === 0,
      missingFiles,
      requiredFiles: [...artifact.requiredFiles],
    };
  });
}

export function getWebDevPortMarkerPath(repoRoot) {
  return path.join(repoRoot, ...WEB_DEV_PORT_MARKER);
}

export function readWebDevPortMarker(repoRoot) {
  try {
    const markerPath = getWebDevPortMarkerPath(repoRoot);
    if (!fs.existsSync(markerPath)) {
      return null;
    }

    const raw = fs.readFileSync(markerPath, 'utf8');
    const parsed = JSON.parse(raw);
    const port = Number(parsed?.port);

    if (!Number.isInteger(port) || port <= 0) {
      return null;
    }

    return {
      ...parsed,
      port,
      markerPath,
    };
  } catch {
    return null;
  }
}

export function getPreferredWebPorts(repoRoot, fallbackPorts) {
  const preferred = readWebDevPortMarker(repoRoot)?.port;
  const normalizedFallbacks = Array.isArray(fallbackPorts) ? fallbackPorts : [];

  if (!preferred) {
    return normalizedFallbacks;
  }

  return [preferred, ...normalizedFallbacks.filter((port) => Number(port) !== preferred)];
}

export function summarizeBrowserExtensionArtifacts(artifacts) {
  if (!Array.isArray(artifacts) || artifacts.length === 0) {
    return {
      ready: false,
      items: [],
      readyCount: 0,
      totalCount: 0,
      summary: 'no browser extension artifacts detected',
    };
  }

  const readyCount = artifacts.filter((artifact) => artifact.ready).length;
  const totalCount = artifacts.length;
  const readyLabels = artifacts
    .filter((artifact) => artifact.ready)
    .map((artifact) => artifact.label.replace('browser extension ', ''));
  const missingLabels = artifacts
    .filter((artifact) => !artifact.ready)
    .map((artifact) => artifact.label.replace('browser extension ', ''));

  return {
    ready: artifacts.every((artifact) => artifact.ready),
    items: artifacts,
    readyCount,
    totalCount,
    summary: missingLabels.length === 0
      ? `${readyCount}/${totalCount} ready (${readyLabels.join(', ')})`
      : `${readyCount}/${totalCount} ready · missing ${missingLabels.join(', ')}`,
  };
}

export function isHttpProbeResponsive(status) {
  if (!status || typeof status !== 'object') {
    return false;
  }

  if (status.ok === true) {
    return true;
  }

  return Number.isInteger(status.status);
}

export async function waitForCoreBridgeShutdown(
  probeUrls,
  {
    timeoutMs = 15000,
    pollIntervalMs = 500,
  } = {},
  {
    probeImpl,
    waitImpl = (ms) => new Promise((resolve) => setTimeout(resolve, ms)),
  } = {},
) {
  if (!Array.isArray(probeUrls) || probeUrls.length === 0 || typeof probeImpl !== 'function') {
    return false;
  }

  const deadline = Date.now() + timeoutMs;

  do {
    const statuses = await Promise.all(probeUrls.map((url) => probeImpl(url)));
    if (statuses.every((status) => !isHttpProbeResponsive(status))) {
      return true;
    }

    if (Date.now() >= deadline) {
      break;
    }

    await waitImpl(pollIntervalMs);
  } while (true);

  return false;
}

function getMissingExtensionReasons(artifacts) {
  return artifacts
    .filter((artifact) => !artifact.ready)
    .map((artifact) => {
      const missing = artifact.missingFiles.length > 0
        ? ` (${artifact.missingFiles.join(', ')})`
        : '';

      return `${artifact.label}${missing}`;
    });
}

export function getPendingStartupChecks(startupStatusData) {
  if (!startupStatusData || typeof startupStatusData !== 'object') {
    return [];
  }

  const checks = startupStatusData.checks;
  if (!checks || typeof checks !== 'object') {
    return [];
  }

  return [
    ...getPendingMcpStartupChecks(startupStatusData),
    ...Object.entries(STARTUP_CHECK_LABELS)
    .filter(([key]) => {
      const check = checks[key];
      return Boolean(check) && typeof check === 'object' && check.ready === false;
    })
    .map(([, label]) => label),
  ];
}

export function getWaitingReasons(state) {
  const missing = [];

  if (!state.web) {
    missing.push('dashboard web server');
  }

  if (!state.coreBridge.ok) {
    missing.push('core extension bridge (/api/mesh/stream)');
  }

  if (!state.startupStatus.ok) {
    missing.push('startup telemetry query');
  }

  if (state.startupStatus.ok && state.startupStatus.compatible === false) {
    missing.push('core bridge startup contract refresh');
  }

  const pendingStartupChecks = getPendingStartupChecks(state.startupStatus.data);
  if (pendingStartupChecks.length > 0) {
    missing.push(...pendingStartupChecks);
  } else {
    if (!state.startupStatus.data?.ready && !state.mcpStatus.ok) {
      missing.push('MCP telemetry query');
    }

    if (!state.startupStatus.data?.ready && !state.memoryStatus.ok) {
      missing.push('memory telemetry query');
    }

    if (!state.startupStatus.data?.ready && !state.browserStatus.ok) {
      missing.push('browser telemetry query');
    }

    if (!state.startupStatus.data?.ready && !state.sessionStatus.ok) {
      missing.push('session telemetry query');
    }
  }

  if (Array.isArray(state.extensions)) {
    missing.push(...getMissingExtensionReasons(state.extensions));
  } else if (state.extension && !state.extension.ready) {
    missing.push(`extension dist files (${state.extension.missing.join(', ')})`);
  }

  return missing;
}