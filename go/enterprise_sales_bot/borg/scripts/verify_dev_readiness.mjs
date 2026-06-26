#!/usr/bin/env node

import {
  detectBrowserExtensionArtifacts,
  getPreferredWebPorts,
  readTormentNexusStartLockRecord,
  summarizeBrowserExtensionArtifacts,
} from './dev_tabby_ready_helpers.mjs';

const REPO_ROOT = process.cwd();
const WEB_PORT_CANDIDATES = [3000, 3010, 3020, 3030, 3040];
const REQUEST_TIMEOUT_MS = Number(process.env.READINESS_TIMEOUT_MS || 2000);
const REQUEST_RETRIES = Number(process.env.READINESS_RETRIES || 1);
const RETRY_DELAY_MS = Number(process.env.READINESS_RETRY_DELAY_MS || 500);
const strictJsonMode = process.argv.includes('--strict-json');
const softMode = process.argv.includes('--soft');
const jsonMode = process.argv.includes('--json') || strictJsonMode;
const compactJsonMode = strictJsonMode || process.argv.includes('--json-compact');

function normalizePort(value) {
  if (typeof value !== 'string') return null;
  const trimmed = value.trim();
  if (!/^\d+$/u.test(trimmed)) return null;
  const parsed = Number(trimmed);
  return Number.isInteger(parsed) && parsed > 0 && parsed <= 65535 ? parsed : null;
}

function uniquePorts(values) {
  return [...new Set(values.filter((value) => Number.isInteger(value) && value > 0))];
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function fetchWithTimeout(url, timeoutMs = REQUEST_TIMEOUT_MS) {
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

function formatLine(service, result) {
  if (result.status === 'up') {
    return `✅ ${service.id.padEnd(20)} ${String(result.statusCode).padEnd(3)} ${result.url}`;
  }

  const details = [
    typeof result.statusCode === 'number' ? `lastStatus=${result.statusCode}` : null,
    result.error ? `lastError=${result.error}` : null,
  ].filter(Boolean).join(' ');

  return `❌ ${service.id.padEnd(20)} DOWN checked=[${service.urls.join(', ')}]${details ? ` ${details}` : ''}`;
}

function getFailureHint(serviceId) {
  switch (serviceId) {
    case 'tormentnexus-web':
      return 'Dashboard is unreachable. Start the web runtime with `tormentnexus dashboard` or `pnpm -C apps/web dev`.';
    case 'tormentnexus-core':
      return 'Core control plane is unreachable. Start it with `tormentnexus start --port 4100`.';
    case 'tormentnexus-startup-status':
      return 'Dashboard can load, but startupStatus is not reachable through the web proxy.';
    case 'tormentnexus-go-sidecar':
            return 'Go sidecar is unreachable. Build with go build ./cmd/tormentnexus and run bin/tormentnexus.exe -port 4300.';
        case 'tormentnexus-mcp-status':
      return 'Dashboard can load, but MCP status is not reachable through the web proxy.';
    default:
      return 'Service did not become ready within the retry window.';
  }
}

function buildWebUrls() {
  return getPreferredWebPorts(REPO_ROOT, WEB_PORT_CANDIDATES).map((port) => `http://127.0.0.1:${port}/dashboard`);
}

function buildCoreUrls() {
  const lockRecord = readTormentNexusStartLockRecord();
  return uniquePorts([
    lockRecord?.port,
    normalizePort(process.env.TORMENTNEXUS_PORT),
    4100,
  ]).map((port) => `http://127.0.0.1:${port}/health`);
}

async function detectRunningEndpoint(service) {
  let lastFailure = { statusCode: null, error: null };

  for (let attempt = 0; attempt <= REQUEST_RETRIES; attempt += 1) {
    for (const url of service.urls) {
      const result = await fetchWithTimeout(url);
      if (result.ok) {
        return { status: 'up', url, statusCode: result.status, error: null };
      }

      if (typeof result.status === 'number') {
        lastFailure.statusCode = result.status;
      }
      if (result.error) {
        lastFailure.error = result.error;
      }
    }

    if (attempt < REQUEST_RETRIES) {
      await sleep(RETRY_DELAY_MS);
    }
  }

  return {
    status: 'down',
    url: null,
    statusCode: lastFailure.statusCode,
    error: lastFailure.error,
  };
}

function collectExtensionArtifacts() {
  return detectBrowserExtensionArtifacts(REPO_ROOT).map((artifact) => ({
    id: artifact.id,
    description: artifact.label,
    critical: false,
    status: artifact.ready ? 'up' : 'down',
    artifactPath: artifact.artifactPath,
    checkedFiles: artifact.requiredFiles,
    missingFiles: artifact.missingFiles,
  }));
}

async function main() {
  const goSidecarPort = normalizePort(process.env.TORMENTNEXUS_GO_PORT) || 4300;
const services = [
    {
      id: 'tormentnexus-web',
      description: 'TormentNexus Next.js dashboard',
      critical: !softMode,
      urls: buildWebUrls(),
    },
    {
      id: 'tormentnexus-core',
      description: 'TormentNexus core control plane health',
      critical: true,
      urls: buildCoreUrls(),
    },
    {
        id: 'tormentnexus-go-sidecar',
        description: 'Go sidecar health',
        critical: false,
        urls: [`http://127.0.0.1:${goSidecarPort}/health`],
    },
    {
      id: 'tormentnexus-startup-status',
      description: 'startupStatus through dashboard proxy',
      critical: !softMode,
      urls: getPreferredWebPorts(REPO_ROOT, WEB_PORT_CANDIDATES).map((port) =>
        `http://127.0.0.1:${port}/api/trpc/startupStatus?batch=1&input=%7B%220%22%3A%7B%22json%22%3Anull%7D%7D`
      ),
    },
    {
      id: 'tormentnexus-mcp-status',
      description: 'mcp.getStatus through dashboard proxy',
      critical: !softMode,
      urls: getPreferredWebPorts(REPO_ROOT, WEB_PORT_CANDIDATES).map((port) =>
        `http://127.0.0.1:${port}/api/trpc/mcp.getStatus?batch=1&input=%7B%7D`
      ),
    },
  ];

  if (!jsonMode) {
    console.log(`\n[TormentNexus Dev Readiness] timeout=${REQUEST_TIMEOUT_MS}ms mode=${softMode ? 'soft' : 'strict'}`);
  }

  const serviceResults = await Promise.all(
    services.map(async (service) => ({ service, result: await detectRunningEndpoint(service) })),
  );

  if (!jsonMode) {
    console.log('\nService Status:');
    for (const { service, result } of serviceResults) {
      console.log(formatLine(service, result));
    }
  }

  const artifacts = collectExtensionArtifacts();
  const extensionSummary = summarizeBrowserExtensionArtifacts(artifacts.map((artifact) => ({
    ...artifact,
    ready: artifact.status === 'up',
    label: artifact.description,
  })));

  const failedCritical = serviceResults.filter(({ service, result }) => service.critical && result.status !== 'up');

  const payload = {
    tool: 'verify_dev_readiness',
    mode: softMode ? 'soft' : 'strict',
    timeoutMs: REQUEST_TIMEOUT_MS,
    retries: REQUEST_RETRIES,
    retryDelayMs: RETRY_DELAY_MS,
    checkedAt: new Date().toISOString(),
    passed: failedCritical.length === 0,
    services: serviceResults.map(({ service, result }) => ({
      id: service.id,
      description: service.description,
      critical: service.critical,
      checkedUrls: service.urls,
      status: result.status,
      url: result.url,
      statusCode: result.statusCode,
      error: result.error ?? null,
      hint: result.status === 'up' ? null : getFailureHint(service.id),
    })),
    artifacts,
    browserExtensionReady: extensionSummary.ready,
  };

  if (jsonMode) {
    console.log(JSON.stringify(payload, null, compactJsonMode ? 0 : 2));
  }

  if (failedCritical.length > 0) {
    if (!jsonMode) {
      console.log('\nSummary: ❌ readiness failed');
      for (const { service } of failedCritical) {
        console.log(`- ${service.id}: ${getFailureHint(service.id)}`);
      }
    }

    if (!softMode) {
      process.exit(1);
    }
  } else if (!jsonMode) {
    console.log('\nSummary: ✅ readiness passed');
  }
}

main().catch((error) => {
  console.error('[TormentNexus Dev Readiness] Unexpected error:', error);
  process.exit(1);
});
