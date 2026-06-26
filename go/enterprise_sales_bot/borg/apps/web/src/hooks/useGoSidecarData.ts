/**
 * @file useGoSidecarData.ts
 * @module apps/web/src/hooks
 *
 * WHAT: React hooks that provide dashboard data from the Go sidecar API
 * when the TypeScript tRPC core is unreachable.
 *
 * WHY: Full Assimilation — the dashboard must display real data even when
 * the TS core is down. These hooks query the Go sidecar REST endpoints
 * and return data in the same shape as the tRPC queries, so the
 * DashboardHomeView component can render without changes.
 *
 * ADDED: v1.0.0-alpha.52
 */

'use client';

import { useEffect, useState, useCallback, useRef } from 'react';

const GO_SIDECAR_BASE = process.env.NEXT_PUBLIC_GO_SIDECAR_URL || '';
const GO_PROXY_PREFIX = '/api/go';

/**
 * Build a URL for the Go sidecar.
 *
 * When a direct sidecar URL is configured (NEXT_PUBLIC_GO_SIDECAR_URL),
 * paths are used as-is.  When going through the Next.js catch-all proxy
 * at /api/go/[...path], the proxy passes the path verbatim to the
 * sidecar, so /api/go/health → sidecar /health,
 * /api/go/api/mcp/status → sidecar /api/mcp/status, etc.
 */
function sidecarUrl(path: string): string {
  if (GO_SIDECAR_BASE) {
    return `${GO_SIDECAR_BASE}${path}`;
  }
  return `${GO_PROXY_PREFIX}${path}`;
}

/** Fetch JSON from the Go sidecar with timeout. */
async function fetchSidecar<T>(path: string, init?: RequestInit): Promise<T | null> {
  try {
    const url = sidecarUrl(path);
    const res = await fetch(url, {
      headers: { accept: 'application/json' },
      signal: AbortSignal.timeout(5000),
      ...init,
    });
    if (!res.ok) return null;
    const json = await res.json();
    return json.success ? (json.data as T) : null;
  } catch {
    return null;
  }
}

// ────────────────────────────────────────────────────────────────────────────
// Types — mirror the DashboardHomeView interfaces so we can slot data in
// ────────────────────────────────────────────────────────────────────────────

export interface GoMCPStatus {
  initialized: boolean;
  serverCount: number;
  toolCount: number;
  connectedCount: number;
}

export interface GoServerSummary {
  name: string;
  status: string;
  toolCount: number;
  config?: { command: string; args: string[]; env: string[] };
}

export interface GoStartupStatus {
  status: string;
  ready: boolean;
  uptime: number;
  summary?: string;
  blockingReasons?: Array<{ code: string; detail: string }>;
  checks: {
    mcpAggregator: {
      ready: boolean;
      liveReady?: boolean;
      residentReady?: boolean;
      serverCount: number;
      connectedCount?: number;
      residentConnectedCount?: number;
      warmingServerCount?: number;
      failedWarmupServerCount?: number;
      initialization: {
        inProgress: boolean;
        initialized: boolean;
        connectedClientCount: number;
        configuredServerCount: number;
      } | null;
      persistedServerCount: number;
      persistedToolCount: number;
      configuredServerCount?: number;
      advertisedServerCount?: number;
      advertisedToolCount?: number;
      advertisedAlwaysOnServerCount?: number;
      advertisedAlwaysOnToolCount?: number;
      inventoryReady: boolean;
      inventorySource?: 'database' | 'config' | 'empty';
      warmupInProgress?: boolean;
    };
    configSync: {
      ready: boolean;
      status: {
        inProgress: boolean;
        lastServerCount: number;
        lastToolCount: number;
      } | null;
    };
    memory: {
      ready: boolean;
      initialized: boolean;
      agentMemory: boolean;
      claudeMem?: {
        ready?: boolean;
        enabled?: boolean;
        storeExists?: boolean;
        totalEntries?: number;
        sectionCount?: number;
        defaultSectionCount?: number;
        presentDefaultSectionCount?: number;
        missingSections?: string[];
        lastUpdatedAt?: string | null;
      };
    };
    browser: { ready: boolean; active: boolean; pageCount: number };
    sessionSupervisor: {
      ready: boolean;
      sessionCount: number;
      restore: {
        restoredSessionCount: number;
        autoResumeCount: number;
      } | null;
    };
    extensionBridge: {
      ready: boolean;
      acceptingConnections?: boolean;
      clientCount: number;
      hasConnectedClients?: boolean;
    };
    executionEnvironment: {
      ready: boolean;
      preferredShellId?: string | null;
      preferredShellLabel?: string | null;
      shellCount: number;
      verifiedShellCount: number;
      toolCount: number;
      verifiedToolCount: number;
      harnessCount: number;
      verifiedHarnessCount: number;
      supportsPowerShell: boolean;
      supportsPosixShell: boolean;
      notes?: string[];
    };
  };
}

export interface GoProviderSummary {
  provider: string;
  name: string;
  configured: boolean;
  authenticated?: boolean;
  tier: string;
  limit: number | null;
  used: number;
  remaining: number | null;
  resetDate?: string | null;
  availability?: string;
  lastError?: string | null;
}

export interface GoFallbackSummary {
  priority: number;
  provider: string;
  model?: string;
  reason: string;
}

export interface GoSessionSummary {
  id: string;
  name: string;
  cliType: string;
  workingDirectory: string;
  status: 'created' | 'starting' | 'running' | 'stopping' | 'stopped' | 'restarting' | 'error';
  restartCount: number;
  maxRestartAttempts: number;
  lastActivityAt: number;
  lastError?: string;
  logs: Array<{ timestamp: number; stream: 'stdout' | 'stderr' | 'system'; message: string }>;
}

// ────────────────────────────────────────────────────────────────────────────
// Hook: useGoSidecarDashboard
// ────────────────────────────────────────────────────────────────────────────

export interface GoSidecarDashboardData {
  mcpStatus: GoMCPStatus | null;
  startupStatus: GoStartupStatus | null;
  servers: GoServerSummary[];
  providers: GoProviderSummary[];
  fallbackChain: GoFallbackSummary[];
  sessions: GoSessionSummary[];
  goVersion: string | null;
  connected: boolean;
  lastFetchedAt: number | null;
}

const EMPTY_DASHBOARD: GoSidecarDashboardData = {
  mcpStatus: null,
  startupStatus: null,
  servers: [],
  providers: [],
  fallbackChain: [],
  sessions: [],
  goVersion: null,
  connected: false,
  lastFetchedAt: null,
};

/**
 * useGoSidecarDashboard polls the Go sidecar REST endpoints and returns
 * dashboard data. Intended as a fallback data source when tRPC queries fail.
 */
export function useGoSidecarDashboard(pollIntervalMs = 5000): GoSidecarDashboardData {
  const [data, setData] = useState<GoSidecarDashboardData>(EMPTY_DASHBOARD);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const poll = useCallback(async () => {
    const [
      mcpStatus,
      startupStatus,
      servers,
      billingStatus,
      sessions,
      healthData,
      healthFallback,
    ] = await Promise.all([
      fetchSidecar<GoMCPStatus>('/api/mcp/status'),
      fetchSidecar<GoStartupStatus>('/api/startup/status'),
      fetchSidecar<GoServerSummary[]>('/api/mcp/servers'),
      fetchSidecar<any>('/api/billing/status'),
      fetchSidecar<{ count: number; sessions: GoSessionSummary[] }>('/api/native/session/list'),
      fetchSidecar<any>('/api/health'),
      fetchSidecar<any>('/health'),
    ]);

    // Use /api/health result, fall back to /health (older binaries)
    const effectiveHealth = healthData ?? healthFallback;

    // Extract providers from billing status
    let providers: GoProviderSummary[] = [];
    let fallbackChain: GoFallbackSummary[] = [];
    if (billingStatus) {
      providers = billingStatus.providers ?? billingStatus.providerQuotas ?? [];
      fallbackChain = billingStatus.fallbackChain ?? [];
    }

    const connected = mcpStatus !== null || startupStatus !== null || effectiveHealth !== null;

    setData({
      mcpStatus,
      startupStatus,
      servers: servers ?? [],
      providers,
      fallbackChain,
      sessions: (sessions as any)?.sessions ?? sessions ?? [],
      goVersion: effectiveHealth?.version ?? null,
      connected,
      lastFetchedAt: Date.now(),
    });
  }, []);

  useEffect(() => {
    poll();
    intervalRef.current = setInterval(poll, pollIntervalMs);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [poll, pollIntervalMs]);

  return data;
}

/**
 * useGoSidecarConnectivity returns just the connection status and version.
 * Lightweight — only hits /api/health.
 */
export function useGoSidecarConnectivity(pollIntervalMs = 10000): {
  connected: boolean;
  version: string | null;
  uptime: number | null;
} {
  const [state, setState] = useState({ connected: false, version: null as string | null, uptime: null as number | null });

  useEffect(() => {
    let cancelled = false;

    async function check() {
      try {
        // Try /api/health first (newer binaries), fall back to /health (older)
        let health = await fetchSidecar<any>('/api/health');
        if (!health) {
          health = await fetchSidecar<any>('/health');
        }
        if (!cancelled) {
          setState({
            connected: health !== null,
            version: health?.version ?? null,
            uptime: health?.uptimeSec ?? health?.uptime ?? null,
          });
        }
      } catch {
        if (!cancelled) setState({ connected: false, version: null, uptime: null });
      }
    }

    check();
    const id = setInterval(check, pollIntervalMs);
    return () => { cancelled = true; clearInterval(id); };
  }, [pollIntervalMs]);

  return state;
}
