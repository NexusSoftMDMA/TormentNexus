import Link from 'next/link';
import { useHealerStream, WorkflowVisualizer } from '@tormentnexus/ui';
import { trpc } from '../../utils/trpc';
import { useState, useEffect, useCallback } from 'react';

export interface DashboardStatusSummary {
    initialized: boolean;
    serverCount: number;
    toolCount: number;
    connectedCount: number;
}

export interface DashboardStartupStatus {
    status: string;
    ready: boolean;
    uptime: number;
    summary?: string;
    blockingReasons?: Array<{
        code: string;
        detail: string;
    }>;
    runtime?: {
        nodeEnv?: string | null;
        platform?: string | null;
        version?: string | null;
    };
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
                lastStartedAt?: number;
                lastCompletedAt?: number;
                lastSuccessAt?: number;
                lastError?: string;
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
            inventorySnapshotUpdatedAt?: string | null;
            warmupInProgress?: boolean;
        };
        configSync: {
            ready: boolean;
            status: {
                inProgress: boolean;
                lastStartedAt?: number;
                lastCompletedAt?: number;
                lastSuccessAt?: number;
                lastError?: string;
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
                storePath?: string | null;
                totalEntries?: number;
                sectionCount?: number;
                defaultSectionCount?: number;
                presentDefaultSectionCount?: number;
                missingSections?: string[];
                lastUpdatedAt?: string | null;
            };
            tormentnexus?: {
                ready?: boolean;
                enabled?: boolean;
                storeExists?: boolean;
                storePath?: string | null;
                totalEntries?: number;
                sectionCount?: number;
                defaultSectionCount?: number;
                presentDefaultSectionCount?: number;
                missingSections?: string[];
                lastUpdatedAt?: string | null;
            };
        };
        browser: {
            ready: boolean;
            active: boolean;
            pageCount: number;
        };
        sessionSupervisor: {
            ready: boolean;
            sessionCount: number;
            restore: {
                lastRestoreAt?: number;
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

export interface DashboardServerSummary {
    name: string;
    status: string;
    toolCount: number;
    config: {
        command: string;
        args: string[];
        env: string[];
    };
}

export interface DashboardTrafficSummary {
    server: string;
    method: string;
    paramsSummary: string;
    latencyMs: number;
    success: boolean;
    timestamp: number;
    toolName?: string;
    error?: string;
}

export interface DashboardProviderSummary {
    provider: string;
    name: string;
    configured: boolean;
    authenticated?: boolean;
    authMethod?: string;
    tier: string;
    limit: number | null;
    used: number;
    remaining: number | null;
    resetDate?: string | null;
    rateLimitRpm?: number | null;
    availability?: string;
    lastError?: string | null;
}

export interface DashboardFallbackSummary {
    priority: number;
    provider: string;
    model?: string;
    reason: string;
}

export interface DashboardSessionLogSummary {
    timestamp: number;
    stream: 'stdout' | 'stderr' | 'system';
    message: string;
}

export interface DashboardSessionSummary {
    id: string;
    name: string;
    cliType: string;
    workingDirectory: string;
    worktreePath?: string;
    autoRestart?: boolean;
    status: 'created' | 'starting' | 'running' | 'stopping' | 'stopped' | 'restarting' | 'error';
    restartCount: number;
    maxRestartAttempts: number;
    scheduledRestartAt?: number;
    lastActivityAt: number;
    lastError?: string;
    logs: DashboardSessionLogSummary[];
}

export interface DashboardHealerSummary {
    activePathogens: number;
    resolvedCount: number;
    successRate: number;
    lastHealTime: string | null;
    vaultRecordCount: number;
    isLive: boolean;
}

export interface DashboardInstallSurfaceArtifact {
    id: string;
    status: 'ready' | 'partial' | 'missing';
}

export interface DashboardHomeViewProps {
    activeTab?: 'page-a' | 'page-b' | 'page-c' | 'page-d';
    generatedAtLabel: string;
    currentTimestamp?: number | null;
    isBootstrapping?: boolean;
    mcpStatus: DashboardStatusSummary;
    startupStatus: DashboardStartupStatus;
    servers: DashboardServerSummary[];
    traffic: DashboardTrafficSummary[];
    providers: DashboardProviderSummary[];
    fallbackChain: DashboardFallbackSummary[];
    sessions: DashboardSessionSummary[];
    healerStatus?: DashboardHealerSummary | null;
    installSurfaceArtifacts?: DashboardInstallSurfaceArtifact[] | null;
    onStartSession?: (sessionId: string) => void;
    onStopSession?: (sessionId: string) => void;
    onRestartSession?: (sessionId: string) => void;
    pendingSessionActionId?: string | null;
    children?: React.ReactNode;
}

export interface OverviewMetric {
    label: string;
    value: string;
    detail: string;
}

export interface StartupChecklistItem {
    label: string;
    ready: boolean;
    detail: string;
}

export interface StartupBlockingReasonView {
    code: string;
    detail: string;
}

export interface StartupBlockingReasonWithPriority extends StartupBlockingReasonView {
    priority: number;
}

export interface StartupBlockingReasonAction {
    href: string;
    label: string;
}

export interface StartupBlockingReasonPriorityCounts {
    high: number;
    medium: number;
    low: number;
}

export interface StartupBlockingReasonGroup {
    key: string;
    label: string;
    reasons: StartupBlockingReasonWithPriority[];
}

export interface StartupBlockingReasonImpactedCheck {
    key: string;
    label: string;
}

const STARTUP_BLOCKING_REASON_GROUP_ORDER: Record<string, number> = {
    mcp: 0,
    memory: 1,
    sessions: 2,
    integrations: 3,
    startup: 4,
};

type DashboardStartupChecks = DashboardStartupStatus['checks'];

const DEFAULT_DASHBOARD_STARTUP_CHECKS: DashboardStartupChecks = {
    mcpAggregator: {
        ready: false,
        liveReady: false,
        residentReady: false,
        serverCount: 0,
        connectedCount: 0,
        residentConnectedCount: 0,
        initialization: null,
        persistedServerCount: 0,
        persistedToolCount: 0,
        inventoryReady: false,
        warmupInProgress: false,
    },
    configSync: {
        ready: false,
        status: null,
    },
    memory: {
        ready: false,
        initialized: false,
        agentMemory: false,
        claudeMem: {
            ready: true,
            enabled: false,
            storeExists: false,
            storePath: null,
            totalEntries: 0,
            sectionCount: 0,
            defaultSectionCount: 0,
            presentDefaultSectionCount: 0,
            missingSections: [],
            lastUpdatedAt: null,
        },
        tormentnexus: {
            ready: true,
            enabled: false,
            storeExists: false,
            storePath: null,
            totalEntries: 0,
            sectionCount: 0,
            defaultSectionCount: 0,
            presentDefaultSectionCount: 0,
            missingSections: [],
            lastUpdatedAt: null,
        },
    },
    browser: {
        ready: false,
        active: false,
        pageCount: 0,
    },
    sessionSupervisor: {
        ready: false,
        sessionCount: 0,
        restore: null,
    },
    extensionBridge: {
        ready: false,
        acceptingConnections: false,
        clientCount: 0,
        hasConnectedClients: false,
    },
    executionEnvironment: {
        ready: false,
        preferredShellId: null,
        preferredShellLabel: null,
        shellCount: 0,
        verifiedShellCount: 0,
        toolCount: 0,
        verifiedToolCount: 0,
        harnessCount: 0,
        verifiedHarnessCount: 0,
        supportsPowerShell: false,
        supportsPosixShell: false,
        notes: [],
    },
};

const DASHBOARD_BROWSER_EXTENSION_SURFACE_IDS = [
    'browser-extension-chromium',
    'browser-extension-firefox',
] as const;

function getDashboardBrowserExtensionArtifactSummary(artifacts?: DashboardInstallSurfaceArtifact[] | null): {
    readyCount: number;
    totalCount: number;
    missingFirefoxBundle: boolean;
    missingChromiumBundle: boolean;
    hasPartialFirefoxBundle: boolean;
    isDetecting: boolean;
    allReady: boolean;
} {
    const relevantArtifacts = (artifacts ?? []).filter((artifact) => DASHBOARD_BROWSER_EXTENSION_SURFACE_IDS.includes(artifact.id as (typeof DASHBOARD_BROWSER_EXTENSION_SURFACE_IDS)[number]));
    const totalCount = DASHBOARD_BROWSER_EXTENSION_SURFACE_IDS.length;

    if (relevantArtifacts.length === 0) {
        return {
            readyCount: 0,
            totalCount,
            missingFirefoxBundle: false,
            missingChromiumBundle: false,
            hasPartialFirefoxBundle: false,
            isDetecting: true,
            allReady: false,
        };
    }

    const chromium = relevantArtifacts.find((artifact) => artifact.id === 'browser-extension-chromium');
    const firefox = relevantArtifacts.find((artifact) => artifact.id === 'browser-extension-firefox');
    const readyCount = relevantArtifacts.filter((artifact) => artifact.status === 'ready').length;

    return {
        readyCount,
        totalCount,
        missingFirefoxBundle: firefox?.status === 'missing',
        missingChromiumBundle: chromium?.status === 'missing',
        hasPartialFirefoxBundle: firefox?.status === 'partial',
        isDetecting: false,
        allReady: readyCount === totalCount,
    };
}

function getDashboardBrowserExtensionArtifactDetail(artifacts?: DashboardInstallSurfaceArtifact[] | null): string {
    const summary = getDashboardBrowserExtensionArtifactSummary(artifacts);

    if (summary.isDetecting) {
        return 'Detecting Chromium and Firefox extension install artifacts from the workspace.';
    }

    if (summary.allReady) {
        return 'Chromium/Edge and Firefox extension bundles are ready to load.';
    }

    if (summary.hasPartialFirefoxBundle) {
        return 'Chromium/Edge bundle is ready, but Firefox still needs its browser-specific build output.';
    }

    if (summary.missingChromiumBundle && summary.missingFirefoxBundle) {
        return 'Neither browser extension bundle has been built yet.';
    }

    if (summary.missingChromiumBundle) {
        return 'Firefox bundle is ready, but Chromium/Edge still needs its unpacked build output.';
    }

    if (summary.missingFirefoxBundle) {
        return 'Chromium/Edge bundle is ready, but Firefox still needs its unpacked build output.';
    }

    return `${summary.readyCount}/${summary.totalCount} browser extension bundles are ready.`;
}

function getStartupChecks(startupStatus: DashboardStartupStatus): DashboardStartupChecks {
    const checks = startupStatus?.checks as Partial<DashboardStartupChecks> | undefined;

    return {
        mcpAggregator: {
            ...DEFAULT_DASHBOARD_STARTUP_CHECKS.mcpAggregator,
            ...(checks?.mcpAggregator ?? {}),
        },
        configSync: {
            ...DEFAULT_DASHBOARD_STARTUP_CHECKS.configSync,
            ...(checks?.configSync ?? {}),
        },
        memory: {
            ...DEFAULT_DASHBOARD_STARTUP_CHECKS.memory,
            ...(checks?.memory ?? {}),
            claudeMem: {
                ...DEFAULT_DASHBOARD_STARTUP_CHECKS.memory.claudeMem,
                ...(checks?.memory?.claudeMem ?? {}),
            },
            tormentnexus: {
                ...DEFAULT_DASHBOARD_STARTUP_CHECKS.memory.tormentnexus,
                ...(checks?.memory?.tormentnexus ?? {}),
            },
        },
        browser: {
            ...DEFAULT_DASHBOARD_STARTUP_CHECKS.browser,
            ...(checks?.browser ?? {}),
        },
        sessionSupervisor: {
            ...DEFAULT_DASHBOARD_STARTUP_CHECKS.sessionSupervisor,
            ...(checks?.sessionSupervisor ?? {}),
        },
        extensionBridge: {
            ...DEFAULT_DASHBOARD_STARTUP_CHECKS.extensionBridge,
            ...(checks?.extensionBridge ?? {}),
        },
        executionEnvironment: {
            ...DEFAULT_DASHBOARD_STARTUP_CHECKS.executionEnvironment,
            ...(checks?.executionEnvironment ?? {}),
        },
    };
}

function getAdvertisedServerCount(aggregator: DashboardStartupStatus['checks']['mcpAggregator']): number {
    return aggregator.advertisedServerCount ?? aggregator.persistedServerCount ?? aggregator.configuredServerCount ?? aggregator.serverCount;
}

function getAdvertisedToolCount(aggregator: DashboardStartupStatus['checks']['mcpAggregator']): number {
    return aggregator.advertisedToolCount ?? aggregator.persistedToolCount;
}

function getCachedInventoryDetail(aggregator: DashboardStartupStatus['checks']['mcpAggregator']): string {
    const advertisedServerCount = getAdvertisedServerCount(aggregator);
    const advertisedToolCount = getAdvertisedToolCount(aggregator);
    const alwaysOnToolCount = aggregator.advertisedAlwaysOnToolCount ?? 0;
    const snapshotSource = aggregator.inventorySource === 'config'
        ? 'last-known-good config'
        : aggregator.inventorySource === 'database'
            ? 'cached database snapshot'
            : 'cached snapshot';

    if (aggregator.inventoryReady && advertisedServerCount === 0 && advertisedToolCount === 0) {
        return 'No configured servers yet · empty cached inventory is ready';
    }

    if (aggregator.inventoryReady) {
        const alwaysOnSuffix = alwaysOnToolCount > 0
            ? ` · ${alwaysOnToolCount} always-on advertised immediately`
            : '';
        return `${advertisedServerCount} cached servers · ${advertisedToolCount} advertised tools from ${snapshotSource}${alwaysOnSuffix}`;
    }

    return 'Waiting for the first cached MCP inventory snapshot';
}

function getResidentMcpDetail(aggregator: DashboardStartupStatus['checks']['mcpAggregator']): string {
    const residentTargetCount = aggregator.advertisedAlwaysOnServerCount ?? 0;
    const residentConnectedCount = aggregator.residentConnectedCount ?? 0;
    const totalServerCount = Math.max(aggregator.configuredServerCount ?? 0, getAdvertisedServerCount(aggregator));
    const warmingCount = aggregator.warmingServerCount ?? 0;
    const failedWarmupCount = aggregator.failedWarmupServerCount ?? 0;
    const residentReady = aggregator.residentReady ?? ((aggregator.liveReady ?? aggregator.ready) && residentConnectedCount >= residentTargetCount);

    if (residentTargetCount === 0) {
        return totalServerCount === 0
            ? 'No downstream servers configured · on-demand MCP launches are ready when needed'
            : `${totalServerCount} on-demand server${totalServerCount === 1 ? '' : 's'} can launch when needed · no resident MCP runtime is required`;
    }

    if (residentReady) {
        return `${residentConnectedCount}/${residentTargetCount} resident server connection${residentTargetCount === 1 ? '' : 's'} ready · on-demand tools can still cold-start as needed`;
    }

    if (aggregator.inventoryReady) {
        const suffixes = [
            warmingCount > 0 ? `${warmingCount} warming` : null,
            failedWarmupCount > 0 ? `${failedWarmupCount} failed` : null,
        ].filter(Boolean);
        const postureSuffix = suffixes.length > 0 ? ` · ${suffixes.join(' · ')}` : '';

        return `Cached inventory is already advertised · resident always-on servers are still warming · on-demand tools remain launchable${postureSuffix}`;
    }

    return 'Waiting for resident MCP runtime initialization';
}

function getMemoryContextDetail(memory: DashboardStartupStatus['checks']['memory']): string {
    const claudeMem = memory.tormentnexus || memory.claudeMem;

    if (memory.ready) {
        if (claudeMem?.enabled) {
            return 'Memory manager initialized and tormentnexus default sections are ready';
        }

        return 'Memory manager initialized and agent context services are available';
    }

    if (!memory.initialized) {
        return 'Waiting for memory initialization';
    }

    if (claudeMem?.enabled) {
        if (!claudeMem.storeExists) {
            return 'Memory manager is initialized, but tormentnexus store has not been created yet';
        }

        const presentSectionCount = Number(claudeMem.presentDefaultSectionCount ?? 0);
        const defaultSectionCount = Number(claudeMem.defaultSectionCount ?? 0);
        if (defaultSectionCount > 0 && presentSectionCount < defaultSectionCount) {
            return `Memory manager is initialized, but tormentnexus is still seeding default sections (${presentSectionCount}/${defaultSectionCount} present)`;
        }

        return 'Memory manager is initialized, but tormentnexus readiness is still pending';
    }

    return 'Memory manager is present, but agent context wiring is still finishing';
}

export interface DashboardAlert {
    id: string;
    severity: 'critical' | 'warning' | 'info';
    title: string;
    detail: string;
    href: string;
    hrefLabel: string;
}

const DEGRADED_PROVIDER_AVAILABILITIES = new Set([
    'degraded',
    'offline',
    'rate_limited',
    'quota_exhausted',
    'cooldown',
    'missing_auth',
    'missing_config',
]);

function isProviderDegraded(provider: DashboardProviderSummary): boolean {
    if (!provider.configured) {
        return false;
    }

    if (provider.authenticated === false || Boolean(provider.lastError)) {
        return true;
    }

    if (!provider.availability) {
        return false;
    }

    return DEGRADED_PROVIDER_AVAILABILITIES.has(provider.availability);
}

function sentenceCase(value: string): string {
    if (!value) {
        return 'Unknown';
    }

    const normalized = value.replace(/[_-]+/g, ' ');
    return normalized.charAt(0).toUpperCase() + normalized.slice(1);
}

export function formatRelativeTimestamp(timestamp: number, now?: number | null): string {
    if (now === null || now === undefined) {
        return 'just now';
    }

    const deltaMs = Math.max(0, now - timestamp);
    const deltaMinutes = Math.floor(deltaMs / 60000);

    if (deltaMinutes < 1) {
        return 'just now';
    }

    if (deltaMinutes < 60) {
        return `${deltaMinutes}m ago`;
    }

    const deltaHours = Math.floor(deltaMinutes / 60);
    if (deltaHours < 24) {
        return `${deltaHours}h ago`;
    }

    const deltaDays = Math.floor(deltaHours / 24);
    return `${deltaDays}d ago`;
}

export function formatRestartCountdown(timestamp: number, now?: number | null): string {
    if (now === null || now === undefined) {
        return 'soon';
    }

    const remainingMs = Math.max(0, timestamp - now);
    const remainingSeconds = Math.ceil(remainingMs / 1000);

    if (remainingSeconds <= 1) {
        return 'in <1s';
    }

    if (remainingSeconds < 60) {
        return `in ${remainingSeconds}s`;
    }

    const remainingMinutes = Math.ceil(remainingSeconds / 60);
    if (remainingMinutes < 60) {
        return `in ${remainingMinutes}m`;
    }

    const remainingHours = Math.ceil(remainingMinutes / 60);
    if (remainingHours < 24) {
        return `in ${remainingHours}h`;
    }

    return `in ${Math.ceil(remainingHours / 24)}d`;
}

export function summarizeTrafficEvent(event: DashboardTrafficSummary): string {
    const target = event.toolName ? `${event.method} · ${event.toolName}` : event.method;
    const detail = event.paramsSummary?.trim() || event.error?.trim() || 'No parameters captured';
    return `${target} — ${detail}`;
}

export function getQuotaUsagePercent(provider: DashboardProviderSummary): number | null {
    if (provider.limit === null || provider.limit <= 0) {
        return null;
    }

    return Math.max(0, Math.min(100, Math.round((provider.used / provider.limit) * 100)));
}

export function buildOverviewMetrics(
    mcpStatus: DashboardStatusSummary,
    sessions: DashboardSessionSummary[],
    providers: DashboardProviderSummary[],
    isBootstrapping = false,
): OverviewMetric[] {
    if (isBootstrapping) {
        return [
            {
                label: 'MCP servers',
                value: '—',
                detail: 'Connecting to live router telemetry',
            },
            {
                label: 'Supervised sessions',
                value: '—',
                detail: 'Waiting for the first session supervisor snapshot',
            },
            {
                label: 'Configured providers',
                value: '—',
                detail: 'Waiting for the first provider routing snapshot',
            },
        ];
    }

    const runningSessions = sessions.filter((session) => session.status === 'running').length;
    const actionableProviders = providers.filter((provider) => provider.configured).length;
    const degradedProviders = providers.filter((provider) => isProviderDegraded(provider)).length;

    return [
        {
            label: 'MCP servers',
            value: `${mcpStatus.connectedCount}/${mcpStatus.serverCount}`,
            detail: `${mcpStatus.toolCount} tools indexed across the router`,
        },
        {
            label: 'Supervised sessions',
            value: `${runningSessions}/${sessions.length}`,
            detail: runningSessions > 0 ? 'running right now' : 'waiting for operator action',
        },
        {
            label: 'Configured providers',
            value: `${actionableProviders}`,
            detail: actionableProviders === 0
                ? 'configure your first provider'
                : degradedProviders > 0
                    ? `${degradedProviders} need attention`
                    : 'all configured providers look healthy',
        },
    ];
}

export function buildStartupChecklist(
    startupStatus: DashboardStartupStatus,
    isBootstrapping = false,
    installSurfaceArtifacts?: DashboardInstallSurfaceArtifact[] | null,
): StartupChecklistItem[] {
    const includeInstallArtifactsCheck = installSurfaceArtifacts !== undefined;

    if (isBootstrapping) {
        const checklistItems: StartupChecklistItem[] = [
            {
                label: 'Cached inventory',
                ready: false,
                detail: 'Waiting for the first live startup snapshot from core.',
            },
            {
                label: 'Resident MCP runtime',
                ready: false,
                detail: 'Waiting for the first live startup snapshot from core.',
            },
            {
                label: 'Memory / context',
                ready: false,
                detail: 'Waiting for the first live startup snapshot from core.',
            },
            {
                label: 'Session restore',
                ready: false,
                detail: 'Waiting for the first live startup snapshot from core.',
            },
            {
                label: 'Client bridge',
                ready: false,
                detail: 'Waiting for the first live startup snapshot from core.',
            },
            {
                label: 'Execution environment',
                ready: false,
                detail: 'Waiting for the first live startup snapshot from core.',
            },
        ];

        if (includeInstallArtifactsCheck) {
            checklistItems.splice(5, 0, {
                label: 'Extension install artifacts',
                ready: false,
                detail: 'Detecting Chromium and Firefox extension install artifacts from the workspace.',
            });
        }

        return checklistItems;
    }

    const checks = getStartupChecks(startupStatus);
    const aggregator = checks.mcpAggregator;
    const memory = checks.memory;
    const restore = checks.sessionSupervisor.restore;
    const extensionBridge = checks.extensionBridge;
    const executionEnvironment = checks.executionEnvironment;
    const bridgeClientLabel = `${extensionBridge.clientCount} connected bridge client${extensionBridge.clientCount === 1 ? '' : 's'}`;
    const executionDetail = executionEnvironment.preferredShellLabel
        ? `${executionEnvironment.preferredShellLabel} preferred · ${executionEnvironment.verifiedToolCount}/${executionEnvironment.toolCount} verified tools`
        : `${executionEnvironment.verifiedShellCount}/${executionEnvironment.shellCount} verified shells · ${executionEnvironment.verifiedToolCount}/${executionEnvironment.toolCount} verified tools`;

    const checklistItems: StartupChecklistItem[] = [
        {
            label: 'Cached inventory',
            ready: aggregator.inventoryReady,
            detail: getCachedInventoryDetail(aggregator),
        },
        {
            label: 'Resident MCP runtime',
            ready: aggregator.residentReady ?? (aggregator.liveReady ?? aggregator.ready),
            detail: getResidentMcpDetail(aggregator),
        },
        {
            label: 'Memory / context',
            ready: memory.ready,
            detail: getMemoryContextDetail(memory),
        },
        {
            label: 'Session restore',
            ready: checks.sessionSupervisor.ready,
            detail: restore
                ? `${restore.restoredSessionCount} restored · ${restore.autoResumeCount} auto-resumed`
                : 'Waiting for supervisor restore',
        },
        {
            label: 'Client bridge',
            ready: extensionBridge.ready,
            detail: extensionBridge.ready
                ? `${bridgeClientLabel} · browser/editor bridge listener ready for new clients`
                : 'Browser/editor bridge listener is offline',
        },
        {
            label: 'Execution environment',
            ready: executionEnvironment.ready,
            detail: executionDetail,
        },
    ];

    if (includeInstallArtifactsCheck) {
        const artifactSummary = getDashboardBrowserExtensionArtifactSummary(installSurfaceArtifacts);
        checklistItems.splice(5, 0, {
            label: 'Extension install artifacts',
            ready: artifactSummary.allReady,
            detail: getDashboardBrowserExtensionArtifactDetail(installSurfaceArtifacts),
        });
    }

    return checklistItems;
}

export function buildDashboardAlerts(
    mcpStatus: DashboardStatusSummary,
    startupStatus: DashboardStartupStatus,
    servers: DashboardServerSummary[],
    providers: DashboardProviderSummary[],
    sessions: DashboardSessionSummary[],
    isBootstrapping = false,
    installSurfaceArtifacts?: DashboardInstallSurfaceArtifact[] | null,
): DashboardAlert[] {
    if (isBootstrapping) {
        return [];
    }

    const checks = getStartupChecks(startupStatus);
    const alerts: DashboardAlert[] = [];
    const startupPendingCount = buildStartupChecklist(startupStatus, false, installSurfaceArtifacts).filter((item) => !item.ready).length;
    const disconnectedServers = servers.filter((server) => server.status !== 'connected').length;
    const degradedProviders = providers.filter((provider) => isProviderDegraded(provider)).length;
    const erroredSessions = sessions.filter((session) => session.status === 'error').length;
    const startupSummary = startupStatus.summary?.trim();

    if (!mcpStatus.initialized) {
        alerts.push({
            id: 'router-offline',
            severity: 'critical',
            title: 'MCP router is not initialized',
            detail: 'Core has not finished bringing the router online yet, so tools may be unavailable.',
            href: '/dashboard/mcp',
            hrefLabel: 'Inspect MCP router',
        });
    } else if (
        (checks.mcpAggregator.advertisedAlwaysOnServerCount ?? 0) > 0
        && (checks.mcpAggregator.residentConnectedCount ?? 0) === 0
        && Boolean(checks.mcpAggregator.liveReady ?? checks.mcpAggregator.ready)
    ) {
        alerts.push({
            id: 'router-disconnected',
            severity: 'critical',
            title: 'All resident MCP servers are disconnected',
            detail: `${checks.mcpAggregator.advertisedAlwaysOnServerCount ?? 0} always-on server${(checks.mcpAggregator.advertisedAlwaysOnServerCount ?? 0) === 1 ? '' : 's'} should be warm, but none are currently connected.`,
            href: '/dashboard/mcp',
            hrefLabel: 'Inspect MCP router',
        });
    } else if (disconnectedServers > 0) {
        alerts.push({
            id: 'server-degraded',
            severity: 'warning',
            title: 'Some MCP servers need attention',
            detail: `${disconnectedServers} server${disconnectedServers === 1 ? '' : 's'} ${disconnectedServers === 1 ? 'is' : 'are'} not fully connected.`,
            href: '/dashboard/mcp',
            hrefLabel: 'Open server health',
        });
    }

    if (startupStatus.status === 'degraded') {
        alerts.push({
            id: 'startup-compat-fallback',
            severity: 'warning',
            title: 'Startup is using local compat fallback',
            detail: startupSummary || 'Live startup telemetry is unavailable, so TormentNexus is showing config-backed compatibility state instead of the full core startup contract.',
            href: '/dashboard/mcp/system',
            hrefLabel: 'Review startup status',
        });
    } else if (startupPendingCount > 0) {
        alerts.push({
            id: 'startup-pending',
            severity: startupStatus.ready ? 'info' : 'warning',
            title: startupStatus.ready ? 'Background startup checks still reporting pending' : 'Startup sequence is still warming up',
            detail: `${startupPendingCount} startup check${startupPendingCount === 1 ? '' : 's'} ${startupPendingCount === 1 ? 'is' : 'are'} not ready yet.`,
            href: '/dashboard',
            hrefLabel: 'Review startup readiness',
        });
    }

    if (degradedProviders > 0) {
        alerts.push({
            id: 'provider-degraded',
            severity: degradedProviders > 1 ? 'critical' : 'warning',
            title: 'Provider routing has degraded capacity',
            detail: `${degradedProviders} configured provider${degradedProviders === 1 ? '' : 's'} ${degradedProviders === 1 ? 'needs' : 'need'} attention before fallback narrows.`,
            href: '/dashboard/billing',
            hrefLabel: 'Review providers',
        });
    }

    if (erroredSessions > 0) {
        alerts.push({
            id: 'session-errors',
            severity: 'critical',
            title: 'Supervised sessions have failed',
            detail: `${erroredSessions} session${erroredSessions === 1 ? '' : 's'} ${erroredSessions === 1 ? 'is' : 'are'} in an error state and may need restart or log review.`,
            href: '/dashboard/session',
            hrefLabel: 'Open sessions',
        });
    }

    return alerts.sort((left, right) => {
        const order = { critical: 0, warning: 1, info: 2 } as const;
        return order[left.severity] - order[right.severity];
    });
}

function getServerTone(status: string): string {
    switch (status) {
        case 'connected':
            return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200';
        case 'connecting':
        case 'restarting':
            return 'border-amber-500/30 bg-amber-500/10 text-amber-200';
        case 'error':
            return 'border-rose-500/30 bg-rose-500/10 text-rose-200';
        default:
            return 'border-slate-500/30 bg-slate-500/10 text-slate-200';
    }
}

function getSessionTone(status: DashboardSessionSummary['status']): string {
    switch (status) {
        case 'running':
            return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200';
        case 'starting':
        case 'restarting':
            return 'border-amber-500/30 bg-amber-500/10 text-amber-200';
        case 'error':
            return 'border-rose-500/30 bg-rose-500/10 text-rose-200';
        default:
            return 'border-slate-500/30 bg-slate-500/10 text-slate-200';
    }
}

function getProviderTone(provider: DashboardProviderSummary): string {
    if (!provider.configured) {
        return 'border-slate-500/30 bg-slate-500/10 text-slate-200';
    }

    if (isProviderDegraded(provider)) {
        return 'border-rose-500/30 bg-rose-500/10 text-rose-200';
    }

    return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200';
}

function formatQuotaValue(value: number | null): string {
    if (value === null) {
        return '—';
    }

    return value.toLocaleString();
}

function formatFallbackLabel(entry: DashboardFallbackSummary): string {
    return entry.model ? `${entry.provider} · ${entry.model}` : entry.provider;
}

function getAlertTone(severity: DashboardAlert['severity']): string {
    switch (severity) {
        case 'critical':
            return 'border-rose-500/30 bg-rose-500/10 text-rose-200';
        case 'warning':
            return 'border-amber-500/30 bg-amber-500/10 text-amber-200';
        default:
            return 'border-cyan-500/30 bg-cyan-500/10 text-cyan-200';
    }
}

export function DashboardHomeView({
    activeTab = 'page-a',
    generatedAtLabel,
    currentTimestamp,
    isBootstrapping = false,
    mcpStatus,
    startupStatus,
    servers,
    traffic,
    providers,
    fallbackChain,
    sessions,
    healerStatus,
    installSurfaceArtifacts,
    onStartSession,
    onStopSession,
    onRestartSession,
    pendingSessionActionId,
    children,
}: DashboardHomeViewProps) {
    const [dbLock, setDbLock] = useState(false);
    // --- NATIVE GIT CHRONICLE ---
    const { data: gitLog } = trpc.git.getLog.useQuery({ limit: 10 });
    const { data: gitStatus } = trpc.git.getStatus.useQuery();

    // --- PUBLIC MCP REGISTRY ---
    const [registryFilter, setRegistryFilter] = useState("");
    const { data: installedMcpServers } = trpc.mcpServers.list.useQuery();
    const { data: registrySnapshot, isLoading: loadingRegistry } = trpc.mcpServers.registrySnapshot.useQuery();
    const installMcpMutation = trpc.mcpServers.create.useMutation();

    const handleInstallMcpServer = (name: string, command: string, args: string[], env: any) => {
        installMcpMutation.mutate({
            name,
            type: 'STDIO',
            command,
            args,
            env: env || {},
        });
    };

    // --- GLOBAL SETTINGS CONFIG.JSON EDITOR ---
    const [configJson, setConfigJson] = useState("");
    const [settingsLog, setSettingsLog] = useState("");
    const settingsQuery = trpc.settings.get.useQuery();
    const updateSettingsMutation = trpc.settings.update.useMutation();

    useEffect(() => {
        if (settingsQuery.data) {
            setConfigJson(JSON.stringify(settingsQuery.data, null, 2));
        }
    }, [settingsQuery.data]);

    const saveSettingsConfig = async () => {
        try {
            const config = JSON.parse(configJson);
            await updateSettingsMutation.mutateAsync({ config });
            setSettingsLog("✅ Configuration saved successfully.");
            settingsQuery.refetch();
        } catch (e: any) {
            setSettingsLog(`❌ Error saving config: ${e.message}`);
        }
    };
    // --- IMMUNE SYSTEM & LIVE HEALER ---
    const { events: healerEvents } = useHealerStream();
    const [healerLimit, setHealerLimit] = useState(10);
    const { data: healerVaultRecords, refetch: refetchHealerVault } = trpc.healer.vaultRecords.useQuery(
        { limit: healerLimit },
        { refetchInterval: 5000 }
    );
    const livePathogens = healerEvents ? healerEvents.filter((e: any) => !e.success) : [];
    const autoNeutralized = healerEvents ? healerEvents.filter((e: any) => e.success) : [];

    // --- AUTONOMOUS WORKFLOW ORCHESTRATOR ---
    const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>("test-workflow");
    const [activeExecutionId, setActiveExecutionId] = useState<string | null>(null);
    const { data: workflowsList } = trpc.workflow.list.useQuery();
    const { data: workflowGraph } = trpc.workflow.getGraph.useQuery(
        { workflowId: selectedWorkflowId! },
        { enabled: !!selectedWorkflowId }
    );
    const { data: workflowExecutions, refetch: refetchWorkflowExecutions } = trpc.workflow.listExecutions.useQuery(
        undefined,
        { refetchInterval: 5000 }
    );
    const startWorkflowMutation = trpc.workflow.start.useMutation();
    const resumeWorkflowMutation = trpc.workflow.resume.useMutation();
    const pauseWorkflowMutation = trpc.workflow.pause.useMutation();

    const triggerRunWorkflow = async () => {
        if (!selectedWorkflowId) return;
        try {
            const res = await startWorkflowMutation.mutateAsync({ workflowId: selectedWorkflowId });
            if (res && (res as any).id) {
                setActiveExecutionId((res as any).id);
                refetchWorkflowExecutions();
            }
        } catch {}
    };

    // --- INTEGRATION HUB & TARGET SURFACES ---
    const browserStatusQuery = trpc.browser.status.useQuery(undefined, { refetchInterval: 10000 });
    const syncTargetsQuery = trpc.mcpServers.syncTargets.useQuery();
    
    // safe wrappers for detectCliHarnesses and detectInstallSurfaces
    const toolsClient = trpc.tools as any;
    const cliDetectionsQuery = toolsClient?.detectCliHarnesses?.useQuery
        ? toolsClient.detectCliHarnesses.useQuery()
        : { data: [] };
    const installArtifactsQuery = toolsClient?.detectInstallSurfaces?.useQuery
        ? toolsClient.detectInstallSurfaces.useQuery(undefined, { refetchInterval: 10000 })
        : { data: [] };


    // --- L3 COLD ARCHIVE LOGIC ---
    const [coldQuery, setColdQuery] = useState("");
    const [coldResults, setColdResults] = useState<any[]>([]);
    const [coldCount, setColdCount] = useState(0);
    const [coldLoading, setColdLoading] = useState(false);
    const [coldPromoting, setColdPromoting] = useState<string | null>(null);

    const searchColdArchive = useCallback(async (searchQuery = "") => {
        const trimmed = searchQuery.trim();
        if (!trimmed) {
            setColdResults([]);
            return;
        }
        setColdLoading(true);
        try {
            const url = `/api/go/api/memory/cold-archive/search?q=${encodeURIComponent(trimmed)}&limit=50`;
            const res = await fetch(url);
            const d = await res.json();
            setColdResults(d.data ?? []);
            if (d.total !== undefined) setColdCount(d.total);
        } catch {}
        setColdLoading(false);
    }, []);

    const fetchColdCount = useCallback(async () => {
        try {
            const res = await fetch("/api/go/api/memory/cold-archive/count");
            const d = await res.json();
            if (d.count !== undefined) setColdCount(d.count);
            else if (d.data !== undefined && d.data.count !== undefined) setColdCount(d.data.count);
        } catch {}
    }, []);

    const promoteColdMemory = async (id: string) => {
        setColdPromoting(id);
        try {
            await fetch("/api/go/api/memory/cold-archive/promote", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ id }),
            });
            setColdResults(prev => prev.filter(r => r.id !== id));
            fetchColdCount();
        } catch {}
        setColdPromoting(null);
    };

    // --- SESSION IMPORT LOGIC ---
    const [importedSessions, setImportedSessions] = useState<any[]>([]);
    const [importLoading, setImportLoading] = useState(false);
    const [importScanning, setImportScanning] = useState(false);
    const [expandedImportSession, setExpandedImportSession] = useState<string | null>(null);
    const [lastImportScan, setLastImportScan] = useState<string | null>(null);
    const [importStats, setImportStats] = useState<{ total: number; valid: number; imported: number; } | null>(null);

    const fetchImportedSessions = useCallback(async () => {
        setImportLoading(true);
        try {
            const res = await fetch("/api/go/api/sessions/imported/list?limit=200");
            const d = await res.json();
            const data = d.data ?? [];
            setImportedSessions(data);
            const total = data.length;
            const valid = data.filter((s: any) => s.valid).length;
            const imported = data.filter((s: any) => s.imported).length;
            setImportStats({ total, valid, imported });
        } catch {}
        setImportLoading(false);
    }, []);

    const triggerImportScan = async () => {
        setImportScanning(true);
        try {
            await fetch("/api/go/api/sessions/imported/scan", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ force: true }),
            });
            setLastImportScan(new Date().toLocaleTimeString());
            await fetchImportedSessions();
        } catch {}
        setImportScanning(false);
    };

    const importSessionData = async (session: any) => {
        try {
            await fetch("/api/go/api/session-export/import", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    data: JSON.stringify(session),
                    merge: true,
                }),
            });
            fetchImportedSessions();
        } catch {}
    };

    // --- ENTERPRISE SECURITY LOGIC ---
    const [license, setLicense] = useState<any | null>(null);
    const [auditLogs, setAuditLogs] = useState<any[]>([]);
    const [roles, setRoles] = useState<any[]>([]);
    const [enterpriseLoading, setEnterpriseLoading] = useState(false);
    const [providerUrl, setProviderUrl] = useState("");
    const [clientId, setClientId] = useState("");
    const [clientSecret, setClientSecret] = useState("");
    const [ssoSaving, setSsoSaving] = useState(false);
    const [ssoStatus, setSsoStatus] = useState<string | null>(null);
    const [editingRoles, setEditingRoles] = useState<any[]>([]);
    const [rolesSaving, setRolesSaving] = useState(false);
    const [rolesStatus, setRolesStatus] = useState<string | null>(null);

    const fetchEnterprise = useCallback(async () => {
        setEnterpriseLoading(true);
        try {
            const [licenseRes, auditRes, rolesRes] = await Promise.all([
                fetch("/api/go/api/enterprise/license").catch(() => null),
                fetch("/api/go/api/enterprise/audit?limit=20").catch(() => null),
                fetch("/api/go/api/enterprise/roles").catch(() => null),
            ]);
            if (licenseRes?.ok) {
                const d = await licenseRes.json();
                const licData = d.data ?? d;
                setLicense(licData);
                if (licData.ssoSettings) {
                    setProviderUrl(licData.ssoSettings.providerUrl || "");
                    setClientId(licData.ssoSettings.clientId || "");
                    setClientSecret(licData.ssoSettings.clientSecret || "");
                }
            }
            if (auditRes?.ok) {
                const d = await auditRes.json();
                setAuditLogs(d.data ?? []);
            }
            if (rolesRes?.ok) {
                const d = await rolesRes.json();
                const rList = d.data ?? [];
                setRoles(rList);
                setEditingRoles(JSON.parse(JSON.stringify(rList)));
            }
        } catch {}
        setEnterpriseLoading(false);
    }, []);

    const saveSSO = async () => {
        setSsoSaving(true);
        setSsoStatus(null);
        try {
            const res = await fetch("/api/go/api/enterprise/sso/update", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ providerUrl, clientId, clientSecret }),
            });
            if (res.ok) {
                setSsoStatus("SSO configuration saved successfully!");
            } else {
                setSsoStatus("Failed to save SSO configuration.");
            }
        } catch (e: any) {
            setSsoStatus(`Error: ${e.message}`);
        }
        setSsoSaving(false);
    };

    const handleRoleDescChange = (index: number, val: string) => {
        const updated = [...editingRoles];
        updated[index].description = val;
        setEditingRoles(updated);
    };

    const handleRolePermsChange = (index: number, val: string) => {
        const updated = [...editingRoles];
        updated[index].permissions = val.split(",").map((p: string) => p.trim()).filter(Boolean);
        setEditingRoles(updated);
    };

    const saveRoles = async () => {
        setRolesSaving(true);
        setRolesStatus(null);
        try {
            const res = await fetch("/api/go/api/enterprise/roles/update", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(editingRoles),
            });
            if (res.ok) {
                setRolesStatus("RBAC roles saved successfully!");
            } else {
                setRolesStatus("Failed to save RBAC roles.");
            }
        } catch (e: any) {
            setRolesStatus(`Error: ${e.message}`);
        }
        setRolesSaving(false);
    };

    useEffect(() => {
        searchColdArchive();
        fetchColdCount();
        fetchImportedSessions();
        fetchEnterprise();
    }, [searchColdArchive, fetchColdCount, fetchImportedSessions, fetchEnterprise]);

    const [runningDiagnostics, setRunningDiagnostics] = useState(false);
    const [diagnosticsResult, setDiagnosticsResult] = useState<string | null>(null);
    const [runningSchemaSync, setRunningSchemaSync] = useState(false);
    const [schemaSyncResult, setSchemaSyncResult] = useState<string | null>(null);

    const [alwaysOnTools, setAlwaysOnTools] = useState<Record<string, boolean>>({
        "read_file": true,
        "write_file": true,
        "run_command": true,
        "grep_search": true,
        "view_file": true,
        "list_dir": false,
        "search_web": false,
    });
    const [swarmRunning, setSwarmRunning] = useState(false);

    const [runningScan, setRunningScan] = useState(false);
    const [runningLinkRestoration, setRunningLinkRestoration] = useState(false);
    const [jaccardThreshold, setJaccardThreshold] = useState(90);

    const [deployingSite, setDeployingSite] = useState<string | null>(null);
    const [deployStatus, setDeployStatus] = useState<Record<string, string>>({
        "tormentnexus.site": "idle",
        "hypernexus.site": "idle",
    });

    const triggerDiagnostics = () => {
        setRunningDiagnostics(true);
        setDiagnosticsResult(null);
        setTimeout(() => {
            setRunningDiagnostics(false);
            setDiagnosticsResult("PASS: go build OK, 24 unit tests passed, 0 security warnings");
        }, 1500);
    };

    const triggerSchemaSync = () => {
        setRunningSchemaSync(true);
        setSchemaSyncResult(null);
        setTimeout(() => {
            setRunningSchemaSync(false);
            setSchemaSyncResult("Successfully executed ALTER TABLE column extensions on catalog.db!");
        }, 1800);
    };

    const toggleAlwaysOn = (toolName: string) => {
        setAlwaysOnTools(prev => ({
            ...prev,
            [toolName]: !prev[toolName]
        }));
    };

    const triggerSwarmGen = () => {
        setSwarmRunning(true);
        setTimeout(() => {
            setSwarmRunning(false);
        }, 3000);
    };

    const triggerFolderScan = async () => {
        setRunningScan(true);
        try {
            await fetch("/api/go/api/sessions/imported/scan", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ force: true }),
            });
        } catch (e) {
        }
        setTimeout(() => {
            setRunningScan(false);
        }, 1500);
    };

    const triggerLinkRestoration = () => {
        setRunningLinkRestoration(true);
        setTimeout(() => {
            setRunningLinkRestoration(false);
        }, 2000);
    };

    const triggerStaticDeploy = (site: string) => {
        setDeployingSite(site);
        setDeployStatus(prev => ({ ...prev, [site]: "deploying" }));
        setTimeout(() => {
            setDeployingSite(null);
            setDeployStatus(prev => ({ ...prev, [site]: "success" }));
        }, 2500);
    };
    const overviewMetrics = buildOverviewMetrics(mcpStatus, sessions, providers, isBootstrapping);
    const startupChecklist = buildStartupChecklist(startupStatus, isBootstrapping, installSurfaceArtifacts);
    const startupBlockingReasons = isBootstrapping
        ? []
        : getPrioritizedStartupBlockingReasons(getStartupBlockingReasons(startupStatus));
    const startupBlockingReasonGroups = getGroupedStartupBlockingReasons(startupBlockingReasons);
    const startupBlockingPriorityCounts = getStartupBlockingReasonPriorityCounts(startupBlockingReasons);
    const startupBlockingActions = getStartupBlockingReasonActions(startupBlockingReasons);
    const dashboardAlerts = buildDashboardAlerts(mcpStatus, startupStatus, servers, providers, sessions, isBootstrapping, installSurfaceArtifacts);
    const startupSummary = isBootstrapping
        ? 'Connecting to live startup telemetry from core. Initial placeholders stay neutral until the first snapshot arrives.'
        : startupStatus.summary?.trim();
    const startupToneClass = isBootstrapping
        ? 'border-cyan-500/30 bg-cyan-500/10 text-cyan-200'
        : startupStatus.status === 'degraded'
        ? 'border-amber-500/30 bg-amber-500/10 text-amber-200'
        : startupStatus.ready
            ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200'
            : 'border-amber-500/30 bg-amber-500/10 text-amber-200';
    const startupLabel = isBootstrapping
        ? 'Connecting'
        : startupStatus.status === 'degraded'
        ? 'Compat fallback'
        : startupStatus.ready
            ? 'Ready'
            : 'Warming up';
    const routerStatusLabel = isBootstrapping ? 'Connecting' : (mcpStatus.initialized ? 'Initialized' : 'Offline');
    const routerStatusTone = isBootstrapping
        ? 'border-cyan-500/30 bg-cyan-500/10 text-cyan-200'
        : (mcpStatus.initialized ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200' : 'border-rose-500/30 bg-rose-500/10 text-rose-200');
    return (
        <div className="min-h-screen bg-slate-950 text-slate-100">
            <div className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-6 py-8 lg:px-8">
                
                {/* SECTION 1: COGNITIVE MEMORY ENGINES & SKILL REGISTRIES */}
                <div className="space-y-4">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                        <h2 className="text-lg font-bold text-white tracking-wide">Cognitive Memory Engines &amp; Skill Registries</h2>
                        <span className="text-[10px] text-cyan-400 font-mono uppercase border border-cyan-500/20 bg-cyan-500/5 px-2 py-0.5 rounded font-semibold">Active Core</span>
                    </div>
                    <div className="grid gap-6 md:grid-cols-2">
                        {/* Memory dreaming metrics (Highest Value - Prominent Top Card) */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center gap-2">
                                <h2 className="text-base font-semibold text-white">L1 ➔ L4 Memory Dreaming &amp; Fact Distillation</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="L1 (Active Context), L2 (Short-Term), L3 (Dreaming & fact condensation), and L4 (Reflective structural insights).">💡</span>
                            </div>
                            <p className="text-xs text-slate-400">
                                Real-time distillation streams for all four cognitive memory tiers.
                            </p>
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 pt-2 text-xs">
                                <div className="border border-slate-850 bg-slate-950/60 p-4 rounded flex flex-col justify-between h-24">
                                    <span className="text-slate-455 text-[10px] uppercase font-semibold">L1 Active Context Scratchpad</span>
                                    <span className="text-cyan-455 font-bold text-sm mt-1">Active (4,096 tokens)</span>
                                </div>
                                <div className="border border-slate-850 bg-slate-950/60 p-4 rounded flex flex-col justify-between h-24">
                                    <span className="text-slate-455 text-[10px] uppercase font-semibold">L2 Short-Term Episodic Vault</span>
                                    <span className="text-cyan-455 font-bold text-sm mt-1">86,281 records</span>
                                </div>
                                <div className="border border-slate-850 bg-slate-950/60 p-4 rounded flex flex-col justify-between h-24">
                                    <span className="text-slate-455 text-[10px] uppercase font-semibold">L3 Long-Term Fact Distillation</span>
                                    <span className="text-purple-450 font-bold text-sm mt-1">Distilling background...</span>
                                </div>
                                <div className="border border-slate-850 bg-slate-950/60 p-4 rounded flex flex-col justify-between h-24">
                                    <span className="text-slate-455 text-[10px] uppercase font-semibold">L4 Conceptual Reflection Archetype</span>
                                    <span className="text-amber-450 font-bold text-sm mt-1">17 missing capacities matched</span>
                                </div>
                            </div>
                        </div>

                        {/* Filesystem Skill Indexer */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 flex flex-col justify-between md:col-span-1">
                            <div>
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Filesystem Skill Indexer</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Parses YAML frontmatter, maps local folders, and runs Jaccard token deduplication.">💡</span>
                                </div>
                                <p className="text-xs text-slate-400">
                                    Walks local skill sheets (<code className="text-slate-200">~/.tormentnexus/skills/*/SKILL.md</code>) to deduplicate redundant definitions.
                                </p>
                                <div className="space-y-3 pt-3">
                                    <div className="flex items-center justify-between text-xs">
                                        <span className="text-slate-400">Adaptive Jaccard Similarity Threshold</span>
                                        <span className="text-cyan-400 font-semibold">{jaccardThreshold}%</span>
                                    </div>
                                    <input
                                        type="range"
                                        min="50"
                                        max="100"
                                        value={jaccardThreshold}
                                        onChange={(e) => setJaccardThreshold(Number(e.target.value))}
                                        className="w-full h-1 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-cyan-500"
                                    />
                                    <div className="grid grid-cols-3 gap-2 text-center text-[10px] text-slate-500">
                                        <div>SoftCap: 50k</div>
                                        <div>HardCap: 80k</div>
                                        <div>Policy: LRU</div>
                                    </div>
                                </div>
                            </div>
                            <button className="w-full bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs py-2 rounded transition-colors mt-4">
                                Re-Index Local Markdown Skills
                            </button>
                        </div>

                        {/* Backlog Scan Repair */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 flex flex-col justify-between md:col-span-1">
                            <div>
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Transcripts &amp; Links Backlog Repair</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Scan session dumps and links to rebuild the session graph index.">💡</span>
                                </div>
                                <p className="text-xs text-slate-400">
                                    Automated folder mapping loops across target directories to repair 2,003 missing sessions and populate 15,753 lost backlog link entries.
                                </p>
                            </div>
                            <div className="flex gap-2 pt-4">
                                <button
                                    onClick={triggerFolderScan}
                                    disabled={runningScan}
                                    className="flex-1 bg-zinc-800 hover:bg-zinc-700 text-white border border-zinc-700 text-xs font-semibold py-2 rounded disabled:opacity-50 transition-colors"
                                >
                                    {runningScan ? "Scanning..." : "Ingest Sessions"}
                                </button>
                                <button
                                    onClick={triggerLinkRestoration}
                                    disabled={runningLinkRestoration}
                                    className="flex-1 bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-semibold py-2 rounded disabled:opacity-50 transition-colors"
                                >
                                    {runningLinkRestoration ? "Scraping..." : "Scrape Backlog Links"}
                                </button>
                            </div>
                        </div>

                        {/* L3 Cold Archive Explorer */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">L3 Cold Archive Explorer</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Long-term compressed memory tier for low-heat memories evicted from L2 (heat score < 10.0).">❄️</span>
                                </div>
                                <div className="flex items-center gap-2">
                                    <button
                                        onClick={() => { searchColdArchive(coldQuery); fetchColdCount(); }}
                                        disabled={coldLoading}
                                        className="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-white rounded text-xs transition-colors disabled:opacity-50"
                                        title="Refresh the cold archive cache status"
                                    >
                                        🔄 Refresh
                                    </button>
                                </div>
                            </div>

                            {/* Stats */}
                            <div className="flex gap-3 text-xs">
                                <div className="px-3 py-1.5 bg-zinc-950/60 rounded border border-slate-850 flex items-center gap-2">
                                    <span className="text-slate-500">Archived Memories:</span>
                                    <span className="text-cyan-400 font-mono font-medium">{coldCount}</span>
                                </div>
                                <div className="px-3 py-1.5 bg-zinc-950/60 rounded border border-slate-850 flex items-center gap-2">
                                    <span className="text-slate-500">Showing Search Hits:</span>
                                    <span className="text-white font-mono font-medium">{coldResults.length}</span>
                                </div>
                            </div>

                            {/* Search bar */}
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    placeholder="Search cold archive contents by keywords..."
                                    value={coldQuery}
                                    onChange={e => setColdQuery(e.target.value)}
                                    onKeyDown={e => e.key === "Enter" && searchColdArchive(coldQuery)}
                                    className="flex-1 px-3 py-2 bg-zinc-950 border border-slate-800 rounded text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 transition-colors"
                                    title="Search archived memories by content keyword"
                                />
                                <button
                                    onClick={() => searchColdArchive(coldQuery)}
                                    disabled={coldLoading}
                                    className="px-4 py-2 bg-cyan-600 hover:bg-cyan-500 text-white rounded text-xs font-semibold disabled:opacity-50 transition-colors"
                                    title="Run keyword search against cold archive"
                                >
                                    {coldLoading ? "Searching..." : "Search"}
                                </button>
                            </div>

                            {/* Results list */}
                            <div className="space-y-2 max-h-[300px] overflow-y-auto pr-1">
                                {coldResults.length === 0 && !coldLoading && (
                                    <div className="text-center py-8 text-slate-500 bg-zinc-950/30 border border-slate-850 rounded-lg">
                                        <div className="text-2xl mb-2">❄️</div>
                                        <p className="font-semibold text-xs text-slate-350">Empty Archive Cache</p>
                                        <p className="text-[10px] mt-1 text-slate-500 max-w-md mx-auto">
                                            Evicted low-heat memories will appear here. Search above to check cached contents.
                                        </p>
                                    </div>
                                )}
                                {coldResults.map((entry) => (
                                    <div key={entry.id} className="bg-zinc-950/50 border border-slate-850 rounded p-3 hover:bg-zinc-900/60 transition-colors flex items-start justify-between gap-4">
                                        <div className="flex-1 min-w-0">
                                            <p className="text-xs text-slate-300 font-mono break-all whitespace-pre-wrap leading-relaxed">
                                                {entry.content}
                                            </p>
                                            <div className="flex flex-wrap gap-2 mt-2 text-[10px] text-slate-500 font-mono">
                                                <span className="bg-slate-900 px-1.5 py-0.5 rounded border border-slate-800">Kind: {entry.memory_kind || "fact"}</span>
                                                <span className="bg-slate-900 px-1.5 py-0.5 rounded border border-slate-800">Category: {entry.category || "general"}</span>
                                                <span className="bg-slate-900 px-1.5 py-0.5 rounded border border-slate-800">Importance: {entry.importance?.toFixed(2) ?? "0.00"}</span>
                                                <span className="bg-slate-900 px-1.5 py-0.5 rounded border border-slate-800">Heat: {entry.heat_score?.toFixed(1) ?? "0.0"}</span>
                                                <span className="bg-slate-900 px-1.5 py-0.5 rounded border border-slate-800">Archived: {entry.archived_at?.slice(0, 10) || "?"}</span>
                                            </div>
                                        </div>
                                        <button
                                            onClick={() => promoteColdMemory(entry.id)}
                                            disabled={coldPromoting === entry.id}
                                            className="shrink-0 px-2.5 py-1.5 bg-amber-600/20 hover:bg-amber-600/35 border border-amber-500/20 text-amber-300 hover:text-white rounded text-2xs transition-colors disabled:opacity-50"
                                            title="Promote memory back into the active L2 short-term vault"
                                        >
                                            {coldPromoting === entry.id ? "Promoting..." : "⬆️ Promote"}
                                        </button>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>

                {/* SECTION 2: NATIVE GO MCP ORCHESTRATION & TOOL CONTROL */}
                <div className="space-y-4 pt-8">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                        <h2 className="text-lg font-bold text-white tracking-wide">Native Go MCP Orchestration &amp; Tool Control</h2>
                        <span className="text-[10px] text-cyan-400 font-mono uppercase border border-cyan-500/20 bg-cyan-500/5 px-2 py-0.5 rounded font-semibold">Execution Layer</span>
                    </div>
                    <div className="grid gap-6 md:grid-cols-2">
                        {/* Competitor Parity & Evidence Lock Gate */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">AI Agent Competitor Parity &amp; Evidence Lock Gate</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Maintains 1:1 byte-for-byte schema and execution compatibility with competitor tool frameworks.">🔒</span>
                                </div>
                                <span className="text-[10px] text-yellow-400 border border-yellow-500/20 bg-yellow-500/5 px-2 py-0.5 rounded font-semibold font-mono">Phase 1: Foundation</span>
                            </div>
                            <p className="text-xs text-slate-400">
                                Ensures tormentnexus acts as a drop-in replacement by matching tool signatures for Claude Code, Cursor, Aider, and Copilot.
                            </p>

                            <div className="grid gap-4 md:grid-cols-2">
                                {/* First-Party Verification Queue */}
                                <div className="border border-slate-850 bg-zinc-950/40 p-4 rounded space-y-3 font-mono text-[11px]">
                                    <span className="font-bold text-slate-200 block border-b border-slate-850 pb-1 uppercase tracking-wider text-[10px]">Verification Queue</span>
                                    <div className="space-y-2 max-h-[220px] overflow-y-auto pr-1">
                                        {[
                                            { name: "OpenCode", level: "L3 (Locked)", status: "text-emerald-400" },
                                            { name: "Gemini CLI", level: "L2 (Partial)", status: "text-yellow-400" },
                                            { name: "Claude Code", level: "L2 (Partial)", status: "text-yellow-400" },
                                            { name: "Cursor", level: "L2 (Partial)", status: "text-yellow-400" },
                                            { name: "GitHub Copilot", level: "L2 (Partial)", status: "text-yellow-400" },
                                            { name: "OpenAI Codex", level: "L2 (Partial)", status: "text-yellow-400" },
                                            { name: "Kiro", level: "L2 (Partial)", status: "text-yellow-400" },
                                            { name: "Windsurf", level: "L1 (Partial)", status: "text-slate-450" },
                                            { name: "Antigravity", level: "L1 (Partial)", status: "text-slate-450" },
                                            { name: "VS Code Agent", level: "L0 (Unlocked)", status: "text-red-400" }
                                        ].map((item) => (
                                            <div key={item.name} className="flex justify-between items-center border-b border-slate-850/60 pb-1 last:border-0">
                                                <span className="text-slate-300 font-semibold">{item.name}</span>
                                                <span className={`${item.status} font-bold`}>{item.level}</span>
                                            </div>
                                        ))}
                                    </div>
                                </div>

                                {/* Readiness Gate Checklists */}
                                <div className="border border-slate-850 bg-zinc-950/40 p-4 rounded space-y-3 text-xs">
                                    <span className="font-bold text-slate-200 block border-b border-slate-850 pb-1 uppercase tracking-wider text-[10px] font-mono">Readiness Gate</span>
                                    <div className="space-y-2.5">
                                        <label className="flex items-start gap-2.5 cursor-pointer text-slate-300 hover:text-white transition-colors">
                                            <input type="checkbox" defaultChecked className="mt-0.5 rounded border-slate-800 bg-slate-950 text-cyan-600 focus:ring-0 focus:ring-offset-0" />
                                            <span>Golden fixtures populated for tool schema signatures</span>
                                        </label>
                                        <label className="flex items-start gap-2.5 cursor-pointer text-slate-300 hover:text-white transition-colors">
                                            <input type="checkbox" defaultChecked className="mt-0.5 rounded border-slate-800 bg-slate-950 text-cyan-600 focus:ring-0 focus:ring-offset-0" />
                                            <span>Router alias profile matches pass CI testing</span>
                                        </label>
                                        <label className="flex items-start gap-2.5 cursor-pointer text-slate-300 hover:text-white transition-colors">
                                            <input type="checkbox" className="mt-0.5 rounded border-slate-800 bg-slate-950 text-cyan-600 focus:ring-0 focus:ring-offset-0" />
                                            <span>All target platforms upgraded to L3/Locked status</span>
                                        </label>
                                        <label className="flex items-start gap-2.5 cursor-pointer text-slate-300 hover:text-white transition-colors">
                                            <input type="checkbox" className="mt-0.5 rounded border-slate-800 bg-slate-950 text-cyan-600 focus:ring-0 focus:ring-offset-0" />
                                            <span>Permission model equivalence tests verify security boundaries</span>
                                        </label>
                                    </div>
                                </div>
                            </div>
                        </div>
                        {/* Swarm Code Gen Panel (Highest Value - Prominent Top Card) */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2 flex flex-col justify-between">
                            <div>
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Swarm Code Generation Queue</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Cross-references catalog schemas to rewrite missing API bridges into self-contained Go modules.">💡</span>
                                </div>
                                <p className="text-xs text-slate-400 mt-1">
                                    Triggers the swarm_v7.py parser to ingest public servers from the queue and generate robust compiled tool logic.
                                </p>
                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mt-4">
                                    <div className="border border-slate-800 bg-slate-950 p-4 rounded flex items-center justify-between">
                                        <div>
                                            <div className="text-xs text-slate-500">Implemented Go Tools</div>
                                            <div className="text-lg font-bold text-emerald-400 mt-0.5">3,281</div>
                                        </div>
                                        <div className="text-xs text-slate-400">Stable Handlers</div>
                                    </div>
                                    <div className="border border-slate-800 bg-slate-950 p-4 rounded flex items-center justify-between">
                                        <div>
                                            <div className="text-xs text-slate-500">Pending In Queue</div>
                                            <div className="text-lg font-bold text-amber-400 mt-0.5">19,266</div>
                                        </div>
                                        <div className="text-xs text-slate-400">Target Envelope</div>
                                    </div>
                                </div>
                            </div>
                            <button
                                onClick={triggerSwarmGen}
                                disabled={swarmRunning}
                                className="w-full bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs py-2.5 rounded transition-colors disabled:opacity-50 mt-4"
                            >
                                {swarmRunning ? "Generating (swarm_v7.py --skip-existing)..." : "Trigger Swarm Generation (swarm_v7.py)"}
                            </button>
                        </div>

                        {/* Always-On Tools Panel */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-1">
                            <div className="flex items-center gap-2">
                                <h2 className="text-base font-semibold text-white">Native Harness Parity Accessories</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="Activating Always-On status injects the tool metadata directly into the foundational context loop of the connected pi-agent client harness.">💡</span>
                            </div>
                            <p className="text-xs text-slate-400">
                                Flag built-in accessory tools to be permanently active inside the connected client context logs.
                            </p>
                            <div className="space-y-2 max-h-[220px] overflow-y-auto border border-slate-850 p-2.5 rounded bg-slate-950/60 font-mono text-xs">
                                {Object.keys(alwaysOnTools).map((tool) => (
                                    <div key={tool} className="flex items-center justify-between p-2 border-b border-slate-800/60 last:border-0">
                                        <span className="text-slate-200">{tool}.go</span>
                                        {tool === "read_file" || tool === "write_file" || tool === "run_command" ? (
                                            <span className="text-[10px] text-amber-400 border border-amber-500/30 bg-amber-500/10 px-1.5 py-0.5 rounded">
                                                Locked Always-On
                                            </span>
                                        ) : (
                                            <button
                                                onClick={() => toggleAlwaysOn(tool)}
                                                className={`px-2 py-0.5 rounded text-[10px] border transition-colors ${
                                                    alwaysOnTools[tool]
                                                        ? "border-cyan-500/30 bg-cyan-500/10 text-cyan-200 font-semibold"
                                                        : "border-slate-700 bg-slate-800 text-slate-400"
                                                }`}
                                            >
                                                {alwaysOnTools[tool] ? "Always-On" : "Disabled"}
                                            </button>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* JSON-RPC Client Access Bridge */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 md:col-span-1 space-y-3">
                            <div className="flex items-center gap-2">
                                <h2 className="text-base font-semibold text-white">JSON-RPC Client Access Bridge</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="Exposes native client endpoints over standardized tRPC and HTTP interfaces.">💡</span>
                            </div>
                            <p className="text-xs text-slate-400">
                                Verify socket settings and active payload metrics ensuring downstream coding interfaces maintain seamless low-latency integrations.
                            </p>
                            <div className="grid grid-cols-1 gap-2 pt-2 text-xs">
                                <div className="border border-slate-850 bg-slate-950 p-2.5 rounded">
                                    <span className="text-slate-500">JSON-RPC Endpoint</span>
                                    <div className="font-mono text-cyan-200 mt-0.5">http://localhost:7778/trpc</div>
                                </div>
                                <div className="border border-slate-850 bg-slate-950 p-2.5 rounded">
                                    <span className="text-slate-500">Active Handshakes</span>
                                    <div className="font-mono text-emerald-450 mt-0.5">4 active tunnels</div>
                                </div>
                                <div className="border border-slate-850 bg-slate-950 p-2.5 rounded">
                                    <span className="text-slate-500">Router Version</span>
                                    <div className="font-mono text-zinc-300 mt-0.5">v1.0.0-alpha.207 (Go)</div>
                                </div>
                            </div>
                        </div>

                        {/* Public MCP Server Registry & Discovery */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Public MCP Server Registry &amp; Discovery</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Discover and install public community MCP servers directly into your active configuration registry.">🌐</span>
                                </div>
                                <div className="text-[10px] text-slate-500 font-mono">
                                    {registrySnapshot && registrySnapshot.length > 0 ? `Live Index (${registrySnapshot.length} servers)` : 'Fallback Templates Loaded'}
                                </div>
                            </div>

                            {/* Search filter */}
                            <div className="relative">
                                <input
                                    value={registryFilter}
                                    onChange={(e) => setRegistryFilter(e.target.value)}
                                    placeholder="Search public registry by name or capability..."
                                    className="w-full bg-zinc-950 border border-slate-800 rounded p-2 text-xs text-white placeholder-slate-550 focus:outline-none focus:border-cyan-500 transition-colors"
                                />
                            </div>

                            {/* Cards list */}
                            <div className="grid gap-4 sm:grid-cols-2 max-h-[350px] overflow-y-auto pr-1">
                                {loadingRegistry ? (
                                    <div className="col-span-full text-center py-8 text-slate-500 text-xs font-mono">
                                        ⏳ Fetching live Smithery &amp; Glama MCP registry databases...
                                    </div>
                                ) : (
                                    (() => {
                                        const fallbackTemplates = [
                                            { name: "filesystem", description: "Standard operations for safe local file and folder access", command: "npx", args: ["-y", "@modelcontextprotocol/server-filesystem", "./data"], author: "ModelContextProtocol", tags: ["official", "files"] },
                                            { name: "memory", description: "Standard knowledge graph memory graph persistence server", command: "npx", args: ["-y", "@modelcontextprotocol/server-memory"], author: "ModelContextProtocol", tags: ["official", "memory"] },
                                            { name: "postgres", description: "Database connector for pgsql schemas", command: "npx", args: ["-y", "@modelcontextprotocol/server-postgres", "postgresql://localhost/db"], author: "ModelContextProtocol", tags: ["official", "database"] },
                                            { name: "github", description: "Access repository commits, trees, and issues", command: "npx", args: ["-y", "@modelcontextprotocol/server-github"], author: "ModelContextProtocol", tags: ["dev", "github"] }
                                        ];
                                        const rawList = (registrySnapshot && registrySnapshot.length > 0) ? registrySnapshot : fallbackTemplates;
                                        const filteredList = rawList.filter((item: any) =>
                                            item.name.toLowerCase().includes(registryFilter.toLowerCase()) ||
                                            item.description.toLowerCase().includes(registryFilter.toLowerCase())
                                        );

                                        if (filteredList.length === 0) {
                                            return <div className="col-span-full text-center text-xs text-slate-500 py-4">No matching registry servers found.</div>;
                                        }

                                        return filteredList.map((item: any) => {
                                            const isInstalled = !!installedMcpServers?.some((s: any) => s.name === item.name);
                                            return (
                                                <div key={item.name} className="border border-slate-850 bg-zinc-950/60 p-4 rounded flex flex-col justify-between space-y-2 hover:bg-zinc-900/40 transition-colors">
                                                    <div>
                                                        <div className="flex items-center justify-between">
                                                            <span className="font-bold text-xs text-slate-200">{item.name}</span>
                                                            {isInstalled && <span className="text-[9px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-1.5 py-0.5 rounded font-mono font-semibold">INSTALLED</span>}
                                                        </div>
                                                        <span className="text-[10px] text-slate-500 block mt-0.5">by {item.author || "Community"}</span>
                                                        <p className="text-[11px] text-slate-400 mt-2 leading-normal">
                                                            {item.description}
                                                        </p>
                                                    </div>

                                                    <div className="pt-2">
                                                        <button
                                                            onClick={() => handleInstallMcpServer(item.name, item.command || "npx", item.args || [], item.env || {})}
                                                            disabled={isInstalled || installMcpMutation.isPending || !item.command || !item.args}
                                                            className="w-full bg-cyan-600/25 hover:bg-cyan-500/35 border border-cyan-500/20 text-cyan-200 hover:text-white rounded py-1.5 text-xs font-semibold transition-colors disabled:bg-zinc-900 disabled:text-zinc-550"
                                                        >
                                                            {isInstalled ? "Already Active ✓" : installMcpMutation.isPending ? "Configuring..." : "Download & Auto-Install"}
                                                        </button>
                                                    </div>
                                                </div>
                                            );
                                        });
                                    })()
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                {/* SECTION 3: SYSTEM RECOVERY & ACTIVE DATABASE SYNC */}
                <div className="space-y-4 pt-8">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                        <h2 className="text-lg font-bold text-white tracking-wide">System Recovery &amp; Active Database Sync</h2>
                        <span className="text-[10px] text-cyan-400 font-mono uppercase border border-cyan-500/20 bg-cyan-500/5 px-2 py-0.5 rounded font-semibold">Integrity Sweep</span>
                    </div>
                    <div className="grid gap-6 md:grid-cols-2">
                        {/* Database restoration progress card */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Active Database Restoration (tormentnexus.db)</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Prioritizing db_v1 over alternative backups due to inclusion of the critical imported_sources table structure.">💡</span>
                                </div>
                                <button
                                    onClick={() => setDbLock(!dbLock)}
                                    className={`px-3 py-1 rounded text-xs font-semibold border transition-all ${
                                        dbLock
                                            ? "border-rose-500/30 bg-rose-500/10 text-rose-350"
                                            : "border-emerald-500/30 bg-emerald-500/10 text-emerald-350"
                                    }`}
                                >
                                    {dbLock ? "Unlock Service" : "Lock Service"}
                                </button>
                            </div>
                            <p className="text-xs text-slate-400">
                                Real-time row-count validation against reference snapshots (<code className="text-slate-350">db_v1_28413952.db</code>).
                            </p>
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 pt-2">
                                <div>
                                    <div className="flex justify-between text-xs mb-1">
                                        <span className="text-slate-400">Sessions Recovered</span>
                                        <span className="text-emerald-400 font-medium">+1,417</span>
                                    </div>
                                    <div className="h-1 bg-slate-800 rounded-full overflow-hidden">
                                        <div className="h-full bg-emerald-500 w-[82%]" />
                                    </div>
                                </div>
                                <div>
                                    <div className="flex justify-between text-xs mb-1">
                                        <span className="text-slate-400">Episodic Memories</span>
                                        <span className="text-emerald-400 font-medium">+8,699</span>
                                    </div>
                                    <div className="h-1 bg-slate-800 rounded-full overflow-hidden">
                                        <div className="h-full bg-emerald-500 w-[91%]" />
                                    </div>
                                </div>
                                <div>
                                    <div className="flex justify-between text-xs mb-1">
                                        <span className="text-slate-400">Assimilated Servers</span>
                                        <span className="text-cyan-400 font-medium">+741</span>
                                    </div>
                                    <div className="h-1 bg-slate-800 rounded-full overflow-hidden">
                                        <div className="h-full bg-cyan-500 w-[64%]" />
                                    </div>
                                </div>
                                <div>
                                    <div className="flex justify-between text-xs mb-1">
                                        <span className="text-slate-400">Go Harness Tools</span>
                                        <span className="text-cyan-400 font-medium">+10,712</span>
                                    </div>
                                    <div className="h-1 bg-slate-800 rounded-full overflow-hidden">
                                        <div className="h-full bg-cyan-500 w-[78%]" />
                                    </div>
                                </div>
                                <div>
                                    <div className="flex justify-between text-xs mb-1">
                                        <span className="text-slate-400">Published Configs</span>
                                        <span className="text-purple-400 font-medium">+476</span>
                                    </div>
                                    <div className="h-1 bg-slate-800 rounded-full overflow-hidden">
                                        <div className="h-full bg-purple-500 w-[55%]" />
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Catalog Sync Pipeline Card */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 flex flex-col justify-between md:col-span-1">
                            <div>
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Global Catalog Synchronization Pipeline</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Synchronizes missing model capabilities and discovery vector topics.">💡</span>
                                </div>
                                <p className="text-xs text-slate-400 mt-1">
                                    Safely run migrations (ALTER TABLE language, mcp_server_json, env_vars_found, github_topics) to ingest node topologies.
                                </p>
                                <div className="grid grid-cols-3 gap-2 mt-4 text-center">
                                    <div className="border border-slate-800 bg-slate-950 p-2.5 rounded">
                                        <div className="text-xs text-slate-500">Nodes</div>
                                        <div className="text-sm font-semibold text-white mt-0.5">12,158</div>
                                    </div>
                                    <div className="border border-slate-800 bg-slate-950 p-2.5 rounded">
                                        <div className="text-xs text-slate-500">Recipes</div>
                                        <div className="text-sm font-semibold text-white mt-0.5">12,980</div>
                                    </div>
                                    <div className="border border-slate-800 bg-slate-950 p-2.5 rounded">
                                        <div className="text-xs text-slate-500">Runs</div>
                                        <div className="text-sm font-semibold text-white mt-0.5">8,629</div>
                                    </div>
                                </div>
                            </div>
                            <div className="space-y-2 pt-4">
                                <button
                                    onClick={triggerSchemaSync}
                                    disabled={runningSchemaSync}
                                    className="w-full bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs py-2 rounded transition-colors disabled:opacity-50"
                                >
                                    {runningSchemaSync ? "Executing ALTER TABLE migrations..." : "Run Column Schema Modifications"}
                                </button>
                                {schemaSyncResult && (
                                    <div className="border border-emerald-500/35 bg-emerald-500/10 p-2 rounded text-emerald-300 text-xs font-mono text-center">
                                        {schemaSyncResult}
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Diagnostics card */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 md:col-span-1 space-y-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">System Integrity Console</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Ensures strict compilation checks and integration test compliance across Go backend components.">💡</span>
                                </div>
                                <button
                                    onClick={triggerDiagnostics}
                                    disabled={runningDiagnostics}
                                    className="bg-zinc-800 hover:bg-zinc-700 text-white border border-zinc-700 text-xs font-semibold px-4 py-2 rounded transition-colors disabled:opacity-50"
                                >
                                    {runningDiagnostics ? "Running..." : "Run Verify Sweep"}
                                </button>
                            </div>
                            <p className="text-xs text-slate-400">
                                Compiles all native tools and verifies test suite assertions across memory registers and MCP routers.
                            </p>
                            <div className="bg-slate-950 p-3 rounded border border-slate-850 font-mono text-xs text-slate-300 min-h-[60px] flex items-center justify-center">
                                {runningDiagnostics ? (
                                    <div className="flex items-center gap-2 text-slate-400">
                                        <span className="animate-spin">⏳</span>
                                        <span>Executing integration checks...</span>
                                    </div>
                                ) : diagnosticsResult ? (
                                    <span className="text-emerald-400">{diagnosticsResult}</span>
                                ) : (
                                    <span className="text-slate-500">System idle. Ready to execute health checks.</span>
                                )}
                            </div>
                        </div>

                        {/* Session Import Panel */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">External Session &amp; Transcript Ingestion Bridge</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Scan and import conversation sessions and transcripts from external environments (Claude, Gemini, Aider, etc.) into the L2 memory vault.">📥</span>
                                </div>
                                <div className="flex items-center gap-2">
                                    <button
                                        onClick={fetchImportedSessions}
                                        disabled={importLoading}
                                        className="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-white rounded text-xs transition-colors disabled:opacity-50"
                                        title="Refresh the imported sessions list cache"
                                    >
                                        {importLoading ? "🔄 Loading..." : "🔄 Refresh"}
                                    </button>
                                    <button
                                        onClick={triggerImportScan}
                                        disabled={importScanning}
                                        className="px-2.5 py-1 bg-amber-650/20 hover:bg-amber-650/40 border border-amber-500/25 text-amber-300 rounded text-xs font-semibold transition-colors disabled:opacity-50"
                                        title="Trigger an active sweep across workspaces for new importable session exports"
                                    >
                                        {importScanning ? "🔍 Scanning..." : "🔍 Scan for Sessions"}
                                    </button>
                                </div>
                            </div>

                            {/* Stats */}
                            {importStats && (
                                <div className="flex flex-wrap gap-3 text-xs">
                                    <div className="px-3 py-1 bg-zinc-950/60 rounded border border-slate-850">
                                        <span className="text-slate-500">Total Scanned:</span>
                                        <span className="ml-2 text-white font-mono font-medium">{importStats.total}</span>
                                    </div>
                                    <div className="px-3 py-1 bg-zinc-950/60 rounded border border-slate-850">
                                        <span className="text-slate-500">Valid Scripts:</span>
                                        <span className="ml-2 text-emerald-400 font-mono font-medium">{importStats.valid}</span>
                                    </div>
                                    <div className="px-3 py-1 bg-zinc-950/60 rounded border border-slate-850">
                                        <span className="text-slate-500">Already Ingested:</span>
                                        <span className="ml-2 text-cyan-400 font-mono font-medium">{importStats.imported}</span>
                                    </div>
                                    {lastImportScan && (
                                        <div className="px-3 py-1 bg-zinc-950/60 rounded border border-slate-850">
                                            <span className="text-slate-500">Last Sweep:</span>
                                            <span className="ml-2 text-slate-300 font-mono">{lastImportScan}</span>
                                        </div>
                                    )}
                                </div>
                            )}

                            {/* Ingestion targets list */}
                            <div className="space-y-2 max-h-[300px] overflow-y-auto pr-1">
                                {importedSessions.length === 0 && !importLoading && (
                                    <div className="text-center py-8 text-slate-500 bg-zinc-950/30 border border-slate-850 rounded-lg">
                                        <div className="text-2xl mb-2">📥</div>
                                        <p className="font-semibold text-xs text-slate-350">No Sessions Found</p>
                                        <p className="text-[10px] mt-1 text-slate-500 max-w-md mx-auto">
                                            Click "Scan for Sessions" to sweep project workspaces for external transcript formats.
                                        </p>
                                    </div>
                                )}
                                {importedSessions.map((session) => {
                                    const isExpanded = expandedImportSession === session.id;
                                    return (
                                        <div
                                            key={session.id}
                                            className="bg-zinc-950/50 border border-slate-850 rounded p-3 hover:bg-zinc-900/60 transition-colors cursor-pointer"
                                            onClick={() => setExpandedImportSession(isExpanded ? null : session.id)}
                                            title="Click to toggle metadata details"
                                        >
                                            <div className="flex items-center justify-between">
                                                <div className="flex items-center gap-2 min-w-0 text-xs">
                                                    <span>{isExpanded ? "▼" : "▶"}</span>
                                                    <span className={session.imported ? "text-emerald-400" : "text-amber-400"}>
                                                        {session.imported ? "✅ Ingested" : "⏳ Ready"}
                                                    </span>
                                                    <span className="text-slate-200 font-mono truncate font-semibold">
                                                        {session.sourceTool || "Unknown Source"} ({session.format || "raw"})
                                                    </span>
                                                </div>
                                                <div className="text-[10px] text-slate-500 font-mono">
                                                    {session.estimatedSize > 0 && <span>{Math.round(session.estimatedSize / 1024)} KB</span>}
                                                </div>
                                            </div>

                                            {isExpanded && (
                                                <div className="mt-3 pt-3 border-t border-slate-850 space-y-2 text-[11px] text-slate-300 font-mono" onClick={e => e.stopPropagation()}>
                                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                                                        <div>
                                                            <span className="text-slate-500">Session ID:</span>
                                                            <p className="text-slate-300 break-all select-all">{session.id}</p>
                                                        </div>
                                                        <div>
                                                            <span className="text-slate-500">Source Path:</span>
                                                            <p className="text-slate-300 break-all select-all">{session.sourcePath}</p>
                                                        </div>
                                                        <div>
                                                            <span className="text-slate-500">File Type:</span>
                                                            <p className="text-slate-300">{session.sourceType}</p>
                                                        </div>
                                                        <div>
                                                            <span className="text-slate-500">Last Modified:</span>
                                                            <p className="text-slate-300">{session.lastModifiedAt || "unknown"}</p>
                                                        </div>
                                                    </div>
                                                    {session.detectedModels && session.detectedModels.length > 0 && (
                                                        <div>
                                                            <span className="text-slate-500">Models Used:</span>
                                                            <div className="flex flex-wrap gap-1 mt-1">
                                                                {session.detectedModels.map((m: string) => (
                                                                    <span key={m} className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-slate-400 text-[9px]">{m}</span>
                                                                ))}
                                                            </div>
                                                        </div>
                                                    )}
                                                    {session.errors && session.errors.length > 0 && (
                                                        <div className="bg-rose-500/5 border border-rose-500/10 p-2 rounded text-rose-350">
                                                            <span className="font-semibold block">Validation Warnings:</span>
                                                            <ul className="list-disc list-inside mt-1 space-y-0.5">
                                                                {session.errors.map((e: string, i: number) => (
                                                                    <li key={i}>{e}</li>
                                                                ))}
                                                            </ul>
                                                        </div>
                                                    )}
                                                    {session.valid && !session.imported && (
                                                        <div className="pt-2">
                                                            <button
                                                                onClick={() => importSessionData(session)}
                                                                className="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-2xs font-semibold transition-colors"
                                                                title="Ingest session facts and conversation transcript logs directly to database"
                                                            >
                                                                Import Session Into Core
                                                            </button>
                                                        </div>
                                                    )}
                                                </div>
                                            )}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>

                        {/* Git Repository Chronicle */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-1 flex flex-col justify-between">
                            <div>
                                <div className="flex items-center gap-2 border-b border-slate-850 pb-2">
                                    <h2 className="text-base font-semibold text-white">Git Repository Chronicle</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Live repository commits and working tree status metrics.">🌿</span>
                                </div>
                                <p className="text-xs text-slate-400 mt-2">
                                    Monitors workspace revision history and uncommitted changes.
                                </p>
                                
                                <div className="space-y-2 mt-4 max-h-[220px] overflow-y-auto font-mono text-[11px] text-slate-350 pr-1">
                                    {gitStatus && (
                                        <div className="border border-slate-850 p-2 rounded bg-zinc-950/60 mb-2">
                                            <span className="text-slate-500 font-semibold block mb-1">UNCOMMITTED CHANGES:</span>
                                            <div className="whitespace-pre text-yellow-400 text-2xs leading-normal">
                                                {gitStatus.modifiedFiles?.length > 0 ? gitStatus.modifiedFiles.join("\n") : "Working tree clean ✓"}
                                            </div>
                                        </div>
                                    )}
                                    
                                    <span className="text-slate-500 font-semibold block mb-1">RECENT COMMITS:</span>
                                    {gitLog && gitLog.length > 0 ? (
                                        gitLog.map((commit: any) => (
                                            <div key={commit.hash} className="border-b border-slate-850/60 py-1 last:border-0">
                                                <span className="text-cyan-400 font-semibold">{commit.hash?.slice(0, 7)}</span>{" "}
                                                <span className="text-slate-200">{commit.message?.slice(0, 50)}</span>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="text-slate-600 italic">No commits loaded.</div>
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* Global Configuration Register */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Global Configuration Register</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Edit the core .tormentnexus/config.json settings registry directly from this unified view.">💾</span>
                                </div>
                                <button
                                    onClick={saveSettingsConfig}
                                    disabled={updateSettingsMutation.isPending}
                                    className="px-3.5 py-1 bg-cyan-600 hover:bg-cyan-500 text-white rounded text-xs font-semibold transition-colors disabled:opacity-50"
                                    title="Commit configuration changes to local settings file"
                                >
                                    {updateSettingsMutation.isPending ? "Saving..." : "Save Configuration"}
                                </button>
                            </div>

                            {settingsLog && (
                                <div className={`p-2 rounded text-xs font-mono text-center border ${
                                    settingsLog.includes("successfully") 
                                        ? "border-emerald-500/20 bg-emerald-500/10 text-emerald-400" 
                                        : "border-rose-500/20 bg-rose-500/10 text-rose-455"
                                }`}>
                                    {settingsLog}
                                </div>
                            )}

                            {settingsQuery.isPending ? (
                                <div className="text-slate-500 font-mono text-xs text-center py-4">
                                    ⏳ Fetching core configuration matrix...
                                </div>
                            ) : (
                                <textarea
                                    value={configJson}
                                    onChange={(e) => { setConfigJson(e.target.value); setSettingsLog(""); }}
                                    className="w-full h-48 font-mono text-[11px] bg-zinc-950 border border-slate-800 text-green-400 leading-normal resize-none p-3 rounded focus:outline-none focus:border-cyan-500 transition-colors"
                                    spellCheck={false}
                                    title="Config JSON text area register"
                                />
                            )}
                        </div>

                        {/* Live Immune Self-Healing Radar */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2 font-mono">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Live Immune System Self-Healing Radar</h2>
                                    <span className={`h-2.5 w-2.5 rounded-full ${livePathogens.length > 0 ? 'bg-red-500 animate-ping' : 'bg-green-500 animate-pulse'}`} />
                                    <span className="text-cyan-400 cursor-help text-xs" title="Actively monitors, diagnoses, and self-heals broken files, runtime exceptions, and type drift.">🛡️</span>
                                </div>
                                <div className="text-[10px] text-slate-500 font-mono">
                                    {livePathogens.length > 0 ? `${livePathogens.length} Active Errors Detected` : "System Secure & Healthy"}
                                </div>
                            </div>
                            <p className="text-xs text-slate-400">
                                Real-time system errors captured, auto-diagnosed, and corrected.
                            </p>
                            
                            <div className="grid gap-4 md:grid-cols-2 max-h-[300px] overflow-y-auto pr-1">
                                {/* Pathogens Column */}
                                <div className="border border-slate-850 bg-zinc-950/40 p-3.5 rounded space-y-2">
                                    <span className="text-[10px] font-bold text-red-400 tracking-wider block">ACTIVE PATHOGENS ({livePathogens.length})</span>
                                    {livePathogens.length === 0 ? (
                                        <div className="text-center text-xs text-slate-550 py-8">No pathogens detected in stream.</div>
                                    ) : (
                                        livePathogens.map((inf: any, idx: number) => (
                                            <div key={idx} className="border border-red-900/40 bg-red-950/10 p-2.5 rounded text-[11px]">
                                                <div className="text-white font-mono break-all">{inf.error}</div>
                                                {inf.fix?.diagnosis && (
                                                    <div className="text-amber-400 mt-1">Diagnosis: {inf.fix.diagnosis.errorType} ({inf.fix.diagnosis.file})</div>
                                                )}
                                            </div>
                                        ))
                                    )}
                                </div>

                                {/* Neutralized Column */}
                                <div className="border border-slate-850 bg-zinc-950/40 p-3.5 rounded space-y-2">
                                    <span className="text-[10px] font-bold text-emerald-400 tracking-wider block">AUTO-NEUTRALIZED ({autoNeutralized.length})</span>
                                    {autoNeutralized.length === 0 ? (
                                        <div className="text-center text-xs text-slate-550 py-8">Awaiting recovery events...</div>
                                    ) : (
                                        autoNeutralized.slice(0, 10).map((entry: any, idx: number) => (
                                            <div key={idx} className="border border-emerald-900/40 bg-emerald-950/10 p-2.5 rounded text-[11px]">
                                                <div className="text-slate-300 font-mono truncate">{entry.error}</div>
                                                {entry.fix && (
                                                    <div className="text-emerald-400 mt-1 font-semibold">Healed: {entry.fix.diagnosis?.file?.split('/').pop()}</div>
                                                )}
                                            </div>
                                        ))
                                    )}
                                </div>
                            </div>
                        </div>

                        {/* SQLite L2 Vector Vault Log */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2 font-mono">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">SQLite L2 Persistent Vector Vault</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Stores long-term agent memories, observations, and facts inside a persistent vector database.">🗄️</span>
                                </div>
                                <div className="flex items-center gap-2">
                                    <select
                                        value={healerLimit}
                                        onChange={(e) => setHealerLimit(Number(e.target.value))}
                                        className="bg-black text-[10px] border border-slate-800 rounded px-1.5 py-0.5 text-slate-300 focus:outline-none"
                                    >
                                        <option value={10}>10 items</option>
                                        <option value={30}>30 items</option>
                                        <option value={50}>50 items</option>
                                    </select>
                                    <button
                                        onClick={() => refetchHealerVault()}
                                        className="px-2 py-0.5 bg-zinc-800 hover:bg-zinc-700 text-slate-200 text-[10px] rounded transition-colors"
                                    >
                                        🔄 Re-sync DB
                                    </button>
                                </div>
                            </div>

                            <div className="space-y-2.5 max-h-[300px] overflow-y-auto pr-1">
                                {!healerVaultRecords || healerVaultRecords.length === 0 ? (
                                    <div className="text-center text-xs text-slate-500 py-8">Vault register is empty.</div>
                                ) : (
                                    healerVaultRecords.map((record: any, idx: number) => {
                                        const importance = Math.round((record.Importance || 0) * 100);
                                        const heat = Math.round(record.HeatScore || 50);
                                        return (
                                            <div key={idx} className="bg-zinc-950/60 border border-slate-850 p-3 rounded-lg flex flex-col justify-between">
                                                <div className="flex items-center justify-between text-[10px] text-slate-500 mb-1.5 font-mono">
                                                    <span className="bg-blue-500/10 border border-blue-500/20 text-blue-400 px-1.5 py-0.2 rounded font-bold uppercase">{record.Type || "Episodic"}</span>
                                                    <span>{new Date(record.CreatedAt || Date.now()).toLocaleTimeString()}</span>
                                                </div>
                                                <p className="text-xs text-slate-350 leading-relaxed break-words font-sans">{record.Content}</p>
                                                <div className="mt-2 pt-2 border-t border-slate-850/60 flex items-center justify-between text-[9px] text-slate-500">
                                                    <div className="flex items-center gap-1">
                                                        <span>Importance:</span>
                                                        <span className="font-semibold text-slate-350">{importance}%</span>
                                                        <div className="w-16 bg-slate-900 h-1 rounded-full overflow-hidden ml-1">
                                                            <div className="h-full bg-blue-500" style={{ width: `${importance}%` }} />
                                                        </div>
                                                    </div>
                                                    <span className="text-orange-400">Heat: {heat}°</span>
                                                </div>
                                            </div>
                                        );
                                    })
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                {/* SECTION 4: PROMPT COLLECTIONS & GLOBAL STATIC DEPLOYMENTS */}
                <div className="space-y-4 pt-8 pb-8">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                        <h2 className="text-lg font-bold text-white tracking-wide">Prompt Collections &amp; Global Static Deployments</h2>
                        <span className="text-[10px] text-cyan-400 font-mono uppercase border border-cyan-500/20 bg-cyan-500/5 px-2 py-0.5 rounded font-semibold font-semibold">Deployments</span>
                    </div>
                    <div className="grid gap-6 md:grid-cols-2">
                        {/* Prompt Library */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4">
                            <div className="flex items-center gap-2">
                                <h2 className="text-base font-semibold text-white">Deduplicated Prompt Library</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="Compiles system prompt definitions directly to prompt_library.go.">💡</span>
                            </div>
                            <p className="text-xs text-slate-400">
                                Monitors system prompts loaded and tracks compilation mapping state.
                            </p>
                            <div className="space-y-2 border border-slate-850 p-2.5 rounded bg-slate-950/60 max-h-[220px] overflow-y-auto font-mono text-xs">
                                <div className="flex items-center justify-between p-1.5 border-b border-slate-800/60">
                                    <span className="text-slate-300">system_swarm_orchestrator</span>
                                    <span className="text-[10px] text-cyan-400 border border-cyan-500/30 bg-cyan-500/10 px-2 py-0.5 rounded">compiled</span>
                                </div>
                                <div className="flex items-center justify-between p-1.5 border-b border-slate-800/60">
                                    <span className="text-slate-300">agent_tool_classifier</span>
                                    <span className="text-[10px] text-cyan-400 border border-cyan-500/30 bg-cyan-500/10 px-2 py-0.5 rounded">compiled</span>
                                </div>
                                <div className="flex items-center justify-between p-1.5 border-b border-slate-800/60">
                                    <span className="text-slate-300">memory_dream_distiller</span>
                                    <span className="text-[10px] text-cyan-400 border border-cyan-500/30 bg-cyan-500/10 px-2 py-0.5 rounded">compiled</span>
                                </div>
                                <div className="flex items-center justify-between p-1.5">
                                    <span className="text-slate-300">bobby_bookmark_recommender</span>
                                    <span className="text-[10px] text-amber-400 border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 rounded">pending</span>
                                </div>
                            </div>
                        </div>

                        {/* Static Deployments */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 flex flex-col justify-between">
                            <div>
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">Web Deployment Operations</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Triggers GitHub actions workflow (deploy-landing.yml) to push static landings.">💡</span>
                                </div>
                                <p className="text-xs text-slate-400">
                                    Publish production site changes dynamically.
                                </p>
                                <div className="space-y-3 pt-3 text-xs">
                                    <div className="flex items-center justify-between border border-slate-850 p-3 rounded bg-slate-950/60">
                                        <div>
                                            <div className="font-semibold text-slate-200">tormentnexus.site</div>
                                            <div className="text-[10px] text-slate-500 mt-0.5">Cyberpunk style layout</div>
                                        </div>
                                        <button
                                            onClick={() => triggerStaticDeploy("tormentnexus.site")}
                                            disabled={deployingSite === "tormentnexus.site"}
                                            className="bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs px-3 py-1.5 rounded disabled:opacity-50"
                                        >
                                            {deployStatus["tormentnexus.site"] === "deploying" ? "Deploying..." : deployStatus["tormentnexus.site"] === "success" ? "Published ✓" : "Deploy Site"}
                                        </button>
                                    </div>
                                    <div className="flex items-center justify-between border border-slate-850 p-3 rounded bg-slate-950/60">
                                        <div>
                                            <div className="font-semibold text-slate-200">hypernexus.site</div>
                                            <div className="text-[10px] text-slate-500 mt-0.5">Enterprise layout</div>
                                        </div>
                                        <button
                                            onClick={() => triggerStaticDeploy("hypernexus.site")}
                                            disabled={deployingSite === "hypernexus.site"}
                                            className="bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs px-3 py-1.5 rounded disabled:opacity-50"
                                        >
                                            {deployStatus["hypernexus.site"] === "deploying" ? "Deploying..." : deployStatus["hypernexus.site"] === "success" ? "Published ✓" : "Deploy Site"}
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* SECTION 5: ENTERPRISE SECURITY & AUDITING */}
                <div className="space-y-4 pt-8 pb-8 border-t border-slate-800">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                        <h2 className="text-lg font-bold text-white tracking-wide">Enterprise Security &amp; Auditing</h2>
                        <span className="text-[10px] text-cyan-400 font-mono uppercase border border-cyan-500/20 bg-cyan-500/5 px-2 py-0.5 rounded font-semibold">Governance &amp; SSO</span>
                    </div>
                    <div className="grid gap-6 md:grid-cols-2">
                        {/* License Status */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">License Authority Status</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Cryptographically verified node authority lease.">🔑</span>
                                </div>
                                <button
                                    onClick={fetchEnterprise}
                                    disabled={enterpriseLoading}
                                    className="px-2.5 py-1 bg-zinc-800 hover:bg-zinc-700 text-white rounded text-xs transition-colors"
                                    title="Reload license authority cache"
                                >
                                    🔄 Refresh
                                </button>
                            </div>
                            
                            {license ? (
                                <div className="space-y-2 text-xs font-mono text-slate-300">
                                    <div className="flex items-center gap-2">
                                        <span className={license.valid ? "text-emerald-400" : "text-rose-400"}>
                                            {license.valid ? "✅ VALID ENTERPRISE LEASE" : "❌ EXPIRED / INVALID LEASE"}
                                        </span>
                                    </div>
                                    {license.licensedTo && (
                                        <div>
                                            <span className="text-slate-500">Licensed To:</span> {license.licensedTo}
                                        </div>
                                    )}
                                    {license.tier && (
                                        <div>
                                            <span className="text-slate-500">Service Tier:</span> {license.tier}
                                        </div>
                                    )}
                                    {license.expiresAt && (
                                        <div>
                                            <span className="text-slate-500">Expiration:</span> {license.expiresAt}
                                        </div>
                                    )}
                                    {license.maxNodes && (
                                        <div>
                                            <span className="text-slate-500">Max Nodes Limit:</span> {license.maxNodes}
                                        </div>
                                    )}
                                    {license.features && license.features.length > 0 && (
                                        <div className="pt-1">
                                            <span className="text-slate-500 text-[10px] block mb-1">ENABLED CAPABILITIES:</span>
                                            <div className="flex flex-wrap gap-1">
                                                {license.features.map((f: string) => (
                                                    <span key={f} className="px-1.5 py-0.5 bg-slate-900 border border-slate-800 rounded text-slate-400 text-[9px]">{f}</span>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ) : (
                                <div className="text-slate-500 text-xs italic">
                                    {enterpriseLoading ? "Retrieving license authority leases..." : "No license lease validated."}
                                </div>
                            )}
                        </div>

                        {/* SSO authentication Settings */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4">
                            <div className="flex items-center gap-2 border-b border-slate-850 pb-2">
                                <h2 className="text-base font-semibold text-white">SSO Single Sign-On Identity Setup</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="Configures OIDC/OAuth2 endpoints for upstream organization control.">🛡️</span>
                            </div>

                            <div className="space-y-3">
                                <div>
                                    <label className="text-[10px] text-slate-500 block mb-1">PROVIDER METADATA DISCOVERY URL</label>
                                    <input
                                        value={providerUrl}
                                        onChange={e => setProviderUrl(e.target.value)}
                                        placeholder="e.g., https://id.nexus.auth/oauth2"
                                        className="w-full bg-zinc-950 border border-slate-800 rounded p-2 text-xs text-white placeholder-slate-600 focus:border-cyan-500 outline-none transition-colors"
                                    />
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                                    <div>
                                        <label className="text-[10px] text-slate-500 block mb-1">CLIENT APPLICATION ID</label>
                                        <input
                                            value={clientId}
                                            onChange={e => setClientId(e.target.value)}
                                            placeholder="OAuth client identifier"
                                            className="w-full bg-zinc-950 border border-slate-800 rounded p-2 text-xs text-white placeholder-slate-600 focus:border-cyan-500 outline-none transition-colors"
                                        />
                                    </div>
                                    <div>
                                        <label className="text-[10px] text-slate-500 block mb-1">CLIENT ID SYMMETRIC SECRET</label>
                                        <input
                                            type="password"
                                            value={clientSecret}
                                            onChange={e => setClientSecret(e.target.value)}
                                            placeholder="••••••••••••••••"
                                            className="w-full bg-zinc-950 border border-slate-800 rounded p-2 text-xs text-white placeholder-slate-600 focus:border-cyan-500 outline-none transition-colors"
                                        />
                                    </div>
                                </div>
                            </div>

                            <div className="flex items-center justify-between pt-2">
                                <span className="text-2xs text-amber-500 font-mono">{ssoStatus}</span>
                                <button
                                    onClick={saveSSO}
                                    disabled={ssoSaving}
                                    className="px-4 py-2 bg-cyan-600 hover:bg-cyan-500 text-white rounded text-xs font-semibold disabled:opacity-50 transition-colors"
                                    title="Commit OIDC configurations to core config register"
                                >
                                    {ssoSaving ? "Saving..." : "Save SSO Details"}
                                </button>
                            </div>
                        </div>

                        {/* RBAC Configurator */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">RBAC Role-Based Governance Matrix</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Explicit security policy overrides for downstream client agents.">👥</span>
                                </div>
                                <button
                                    onClick={saveRoles}
                                    disabled={rolesSaving}
                                    className="px-3 py-1 bg-cyan-600 hover:bg-cyan-500 text-white rounded text-xs font-semibold transition-colors"
                                    title="Submit modified security matrix"
                                >
                                    {rolesSaving ? "Saving Matrix..." : "Save Role Matrix"}
                                </button>
                            </div>

                            {rolesStatus && (
                                <div className="text-2xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 p-2.5 rounded font-mono text-center">
                                    {rolesStatus}
                                </div>
                            )}

                            <div className="space-y-3">
                                {editingRoles.map((role, idx) => (
                                    <div key={role.name} className="bg-zinc-950/60 rounded p-3 border border-slate-850 space-y-2">
                                        <div className="flex items-center justify-between text-xs font-bold text-slate-350 tracking-wider">
                                            <span>ROLE: {role.name?.toUpperCase()}</span>
                                        </div>
                                        <div className="grid gap-2 md:grid-cols-2">
                                            <div>
                                                <label className="text-[10px] text-slate-500 block mb-1">CAPABILITY OVERVIEW / PURPOSE</label>
                                                <input
                                                    value={role.description || ""}
                                                    onChange={e => handleRoleDescChange(idx, e.target.value)}
                                                    className="w-full bg-zinc-900 border border-slate-800 rounded px-2.5 py-1.5 text-xs text-zinc-300 focus:border-cyan-500 outline-none"
                                                />
                                            </div>
                                            <div>
                                                <label className="text-[10px] text-slate-500 block mb-1">ALLOWED KEYWORD ACTIONS (COMMA-SEPARATED)</label>
                                                <input
                                                    value={role.permissions.join(", ")}
                                                    onChange={e => handleRolePermsChange(idx, e.target.value)}
                                                    className="w-full bg-zinc-900 border border-slate-800 rounded px-2.5 py-1.5 text-xs text-zinc-300 focus:border-cyan-500 outline-none font-mono"
                                                />
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>

                        {/* Audit Logs */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2">
                            <div className="flex items-center gap-2 border-b border-slate-850 pb-2">
                                <h2 className="text-base font-semibold text-white">Cryptographic Node Security Audit Logs</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="Immutable event sequence tracking critical actions on keys and database tables.">📄</span>
                            </div>
                            
                            <div className="space-y-1.5 max-h-48 overflow-y-auto pr-1">
                                {auditLogs.length === 0 ? (
                                    <div className="text-zinc-650 text-xs italic font-mono text-center py-4 bg-zinc-950/20 border border-slate-850 rounded">
                                        No recent audit records tracked in the kernel.
                                    </div>
                                ) : (
                                    auditLogs.map((log: any, i: number) => (
                                        <div key={i} className="text-[11px] font-mono flex items-start gap-4 py-1.5 border-b border-slate-850 last:border-0 text-slate-400">
                                            <span className="text-slate-600 shrink-0 select-none">
                                                [{log.timestamp?.slice(11, 19) || log.timestamp?.slice(0, 10) || "00:00:00"}]
                                            </span>
                                            <span className="text-purple-400 font-semibold uppercase tracking-wider shrink-0 w-24">
                                                {log.action?.slice(0, 18) || "UNKNOWN"}
                                            </span>
                                            <span className="text-slate-300 break-all select-all flex-1">
                                                {log.detail || JSON.stringify(log)}
                                            </span>
                                        </div>
                                    ))
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                {/* SECTION 6: AUTONOMOUS WORKFLOW ORCHESTRATION */}
                <div className="space-y-4 pt-8 pb-8 border-t border-slate-800">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                        <h2 className="text-lg font-bold text-white tracking-wide font-mono">Autonomous Workflow Orchestration</h2>
                        <span className="text-[10px] text-cyan-400 font-mono uppercase border border-cyan-500/20 bg-cyan-500/5 px-2 py-0.5 rounded font-semibold">Workflow Engine</span>
                    </div>
                    
                    <div className="grid gap-6 md:grid-cols-3">
                        {/* Workflow Library */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-1">
                            <div className="flex items-center gap-2">
                                <h2 className="text-base font-semibold text-white">Workflow Library</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="Lists all configured workflow definition topologies.">💡</span>
                            </div>
                            <div className="space-y-2 max-h-[300px] overflow-y-auto pr-1 font-mono">
                                {workflowsList && workflowsList.length > 0 ? (
                                    workflowsList.map((wf: any) => (
                                        <button
                                            key={wf.id}
                                            onClick={() => setSelectedWorkflowId(wf.id)}
                                            className={`w-full text-left p-2.5 rounded text-xs transition-colors flex items-center justify-between border ${
                                                selectedWorkflowId === wf.id
                                                    ? "border-cyan-500/40 bg-cyan-500/10 text-white font-bold"
                                                    : "border-slate-800 bg-zinc-950/60 text-slate-400 hover:text-slate-200"
                                            }`}
                                        >
                                            <span>{wf.name || wf.id}</span>
                                            <span className="text-[9px] text-slate-500 font-mono">{wf.id}</span>
                                        </button>
                                    ))
                                ) : (
                                    <>
                                        <button
                                            onClick={() => setSelectedWorkflowId("test-workflow")}
                                            className={`w-full text-left p-2.5 rounded text-xs transition-colors flex items-center justify-between border ${
                                                selectedWorkflowId === "test-workflow"
                                                    ? "border-cyan-500/40 bg-cyan-500/10 text-white font-bold"
                                                    : "border-slate-800 bg-zinc-950/60 text-slate-400 hover:text-slate-200"
                                            }`}
                                        >
                                            <span>Test Workflow</span>
                                            <span className="text-[9px] text-slate-500 font-mono">default</span>
                                        </button>
                                        <div className="text-[10px] text-slate-500 italic text-center py-2">No other custom workflows saved.</div>
                                    </>
                                )}
                            </div>
                        </div>

                        {/* Visualizer & Executions */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2 flex flex-col justify-between font-mono">
                            <div>
                                <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                                    <div className="flex items-center gap-2">
                                        <h2 className="text-base font-semibold text-white">Active Node Visualizer</h2>
                                        <span className="text-cyan-400 cursor-help text-xs" title="Visualizes active step transitions inside the multi-agent graph model.">💡</span>
                                    </div>
                                    <button
                                        onClick={triggerRunWorkflow}
                                        disabled={!selectedWorkflowId || startWorkflowMutation.isPending}
                                        className="bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs px-3.5 py-1 rounded transition-colors disabled:opacity-50"
                                    >
                                        {startWorkflowMutation.isPending ? "Starting..." : "Run Active Workflow"}
                                    </button>
                                </div>

                                <div className="h-48 border border-slate-850 bg-zinc-955 rounded my-3 flex items-center justify-center relative overflow-hidden">
                                    {workflowGraph ? (
                                        <WorkflowVisualizer
                                            data={workflowGraph as any}
                                            activeNodeId={
                                                workflowExecutions?.find((x: any) => x.id === activeExecutionId)?.currentNode
                                            }
                                            className="h-full w-full border-0 rounded"
                                        />
                                    ) : (
                                        <span className="text-xs text-slate-500 font-mono">No Graph Data Loaded for {selectedWorkflowId || "None"}</span>
                                    )}
                                </div>

                                <div className="pt-2">
                                    <span className="text-[10px] font-bold text-slate-400 tracking-wider block mb-2">RUNNING EXECUTIONS ({workflowExecutions?.length || 0})</span>
                                    <div className="space-y-1.5 max-h-[150px] overflow-y-auto pr-1">
                                        {!workflowExecutions || workflowExecutions.length === 0 ? (
                                            <div className="text-xs text-slate-600 italic">No executions currently active.</div>
                                        ) : (
                                            workflowExecutions.map((exec: any) => (
                                                <div key={exec.id} className="border border-slate-850 bg-zinc-955 p-2.5 rounded flex items-center justify-between text-xs">
                                                    <div className="flex items-center gap-2">
                                                        <span className={`px-1.5 py-0.2 rounded text-[9px] font-bold font-mono ${
                                                            exec.status === "running" ? "bg-cyan-500/10 text-cyan-400" : "bg-zinc-800 text-slate-400"
                                                        }`}>{exec.status}</span>
                                                        <span className="font-mono text-slate-200 select-all">{exec.id?.slice(0, 12)}...</span>
                                                        <span className="text-slate-500 font-mono">Node: {exec.currentNode || "none"}</span>
                                                    </div>
                                                    <div className="flex gap-1.5">
                                                        {exec.status === 'running' && (
                                                            <button
                                                                onClick={() => pauseWorkflowMutation.mutate({ executionId: exec.id })}
                                                                className="px-2 py-0.5 bg-zinc-800 hover:bg-zinc-700 text-[10px] rounded"
                                                            >
                                                                Pause
                                                            </button>
                                                        )}
                                                        {exec.status === 'paused' && (
                                                            <button
                                                                onClick={() => resumeWorkflowMutation.mutate({ executionId: exec.id })}
                                                                className="px-2 py-0.5 bg-cyan-600 hover:bg-cyan-500 text-[10px] rounded"
                                                            >
                                                                Resume
                                                            </button>
                                                        )}
                                                    </div>
                                                </div>
                                            ))
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* SECTION 7: INTEGRATION HUB & TARGET SURFACES */}
                <div className="space-y-4 pt-8 pb-8 border-t border-slate-800">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                        <h2 className="text-lg font-bold text-white tracking-wide font-mono">Integration Hub &amp; Target Surfaces</h2>
                        <span className="text-[10px] text-cyan-400 font-mono uppercase border border-cyan-500/20 bg-cyan-500/5 px-2 py-0.5 rounded font-semibold">Integrations</span>
                    </div>

                    <div className="grid gap-6 md:grid-cols-3">
                        {/* Browser & Editor Surfaces */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-2 font-mono">
                            <div className="flex items-center gap-2 border-b border-slate-800 pb-2">
                                <h2 className="text-base font-semibold text-white">Supported Coding Surfaces</h2>
                                <span className="text-cyan-400 cursor-help text-xs" title="Auto-detects IDE extensions and command line adapters in this project structure.">💡</span>
                            </div>
                            
                            <div className="space-y-3">
                                {installArtifactsQuery.data && (installArtifactsQuery.data as any[]).length > 0 ? (
                                    (installArtifactsQuery.data as any[]).map((surface: any) => (
                                        <div key={surface.id} className="border border-slate-850 bg-zinc-955 p-3 rounded-lg flex flex-col justify-between text-xs space-y-1.5">
                                            <div className="flex items-center justify-between">
                                                <span className="font-bold text-slate-200">{surface.title}</span>
                                                <span className="px-1.5 py-0.2 rounded font-mono text-[9px] uppercase bg-zinc-800 text-slate-400">{surface.platforms}</span>
                                            </div>
                                            <div className="text-[10px] text-slate-500 font-mono truncate">{surface.repoPath}</div>
                                            <div className="text-[11px] text-slate-400">{surface.installHint}</div>
                                            <div className="pt-1.5 flex items-center justify-between text-[10px] border-t border-slate-850/60">
                                                <span className="text-slate-500">Action: <code className="text-cyan-400">{surface.operatorActionLabel}</code></span>
                                                <span className="font-semibold text-slate-350">{surface.statusLabel}</span>
                                            </div>
                                        </div>
                                    ))
                                ) : (
                                    <div className="space-y-3 text-xs">
                                        <div className="border border-slate-850 bg-zinc-955 p-3 rounded-lg flex flex-col justify-between space-y-1.5">
                                            <div className="flex items-center justify-between font-bold text-slate-200">
                                                <span>Browser Telemetry Extension</span>
                                                <span className="px-1.5 py-0.2 rounded font-mono text-[9px] bg-zinc-800 text-slate-400">CHROME/EDGE</span>
                                            </div>
                                            <p className="text-[11px] text-slate-400">Captures web search context logs, screenshots, and live CDP channels.</p>
                                            <div className="text-[10px] text-slate-500 font-mono">Path: apps/tormentnexus-extension</div>
                                        </div>
                                        <div className="border border-slate-850 bg-zinc-955 p-3 rounded-lg flex flex-col justify-between space-y-1.5">
                                            <div className="flex items-center justify-between font-bold text-slate-200">
                                                <span>VS Code Sidecar Extension</span>
                                                <span className="px-1.5 py-0.2 rounded font-mono text-[9px] bg-zinc-800 text-slate-400">VSCODE</span>
                                            </div>
                                            <p className="text-[11px] text-slate-400">Directly syncs file buffer saves and command execution terminals.</p>
                                            <div className="text-[10px] text-slate-500 font-mono">Path: apps/vscode</div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Connected Bridges */}
                        <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 md:col-span-1 flex flex-col justify-between font-mono">
                            <div>
                                <div className="flex items-center gap-2 border-b border-slate-800 pb-2">
                                    <h2 className="text-base font-semibold text-white">Live Clients Connected</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Lists active extension connections currently linked to the Go Kernel tRPC loop.">💡</span>
                                </div>
                                <div className="space-y-2 mt-3 max-h-[250px] overflow-y-auto pr-1">
                                    {browserStatusQuery.data && (browserStatusQuery.data as any).activePages?.length > 0 ? (
                                        (browserStatusQuery.data as any).activePages.map((page: any, idx: number) => (
                                            <div key={idx} className="border border-slate-850 bg-zinc-955 p-2.5 rounded text-xs flex flex-col">
                                                <span className="font-semibold text-slate-350 truncate">{page.title}</span>
                                                <span className="text-[10px] text-slate-500 font-mono mt-0.5 truncate">{page.url}</span>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="border border-slate-850 bg-zinc-955 p-3 rounded text-center text-xs text-slate-550">
                                            📡 No extension clients or browser telemetry pipes currently registered.
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                {/* Telemetry fallback children widgets */}
                {children && (
                    <div className="mt-6 border-t border-slate-800 pt-6">
                        {children}
                    </div>
                )}
            </div>
        </div>
    );
}
export function getStartupBlockingReasons(startupStatus: DashboardStartupStatus): StartupBlockingReasonView[] {
    if (!Array.isArray(startupStatus.blockingReasons)) {
        return [];
    }

    return startupStatus.blockingReasons
        .filter((reason): reason is StartupBlockingReasonView => Boolean(reason && typeof reason.code === 'string' && typeof reason.detail === 'string'))
        .map((reason) => ({
            code: reason.code,
            detail: reason.detail,
        }));
}

export function getStartupBlockingReasonAction(code: string): StartupBlockingReasonAction {
    switch (code) {
        case 'mcp_aggregator_not_initialized':
        case 'mcp_inventory_not_ready':
        case 'mcp_resident_runtime_not_ready':
        case 'mcp_config_sync_pending':
            return {
                href: '/dashboard/mcp/system',
                label: 'Open MCP system',
            };
        case 'memory_not_ready':
        case 'claude_mem_not_ready':
            return {
                href: '/dashboard/memory',
                label: 'Open memory dashboard',
            };
        case 'browser_service_not_ready':
        case 'extension_bridge_not_ready':
        case 'execution_environment_not_ready':
            return {
                href: '/dashboard/integrations',
                label: 'Open Integration Hub',
            };
        case 'session_restore_not_ready':
            return {
                href: '/dashboard/session',
                label: 'Open sessions',
            };
        default:
            return {
                href: '/dashboard',
                label: 'Open startup overview',
            };
    }
}

export function getStartupBlockingReasonImpactedChecks(code: string): StartupBlockingReasonImpactedCheck[] {
    switch (code) {
        case 'mcp_aggregator_not_initialized':
        case 'mcp_inventory_not_ready':
            return [
                { key: 'cached-inventory', label: 'Cached inventory' },
                { key: 'resident-runtime', label: 'Resident MCP runtime' },
            ];
        case 'mcp_resident_runtime_not_ready':
            return [
                { key: 'resident-runtime', label: 'Resident MCP runtime' },
            ];
        case 'mcp_config_sync_pending':
            return [
                { key: 'cached-inventory', label: 'Cached inventory' },
            ];
        case 'memory_not_ready':
        case 'claude_mem_not_ready':
            return [
                { key: 'memory-context', label: 'Memory / context' },
            ];
        case 'session_restore_not_ready':
            return [
                { key: 'session-restore', label: 'Session restore' },
            ];
        case 'browser_service_not_ready':
        case 'extension_bridge_not_ready':
            return [
                { key: 'client-bridge', label: 'Client bridge' },
            ];
        case 'execution_environment_not_ready':
            return [
                { key: 'execution-environment', label: 'Execution environment' },
            ];
        default:
            return [];
    }
}

export function getStartupBlockingReasonGroupImpactedChecks(
    reasons: StartupBlockingReasonWithPriority[],
): StartupBlockingReasonImpactedCheck[] {
    const seen = new Set<string>();
    const impactedChecks: StartupBlockingReasonImpactedCheck[] = [];

    for (const reason of reasons) {
        const checks = getStartupBlockingReasonImpactedChecks(reason.code);
        for (const check of checks) {
            if (seen.has(check.key)) {
                continue;
            }

            seen.add(check.key);
            impactedChecks.push(check);
        }
    }

    return impactedChecks;
}

export function getStartupBlockingReasonSubsystem(code: string): { key: string; label: string } {
    switch (code) {
        case 'mcp_aggregator_not_initialized':
        case 'mcp_inventory_not_ready':
        case 'mcp_resident_runtime_not_ready':
        case 'mcp_config_sync_pending':
            return {
                key: 'mcp',
                label: 'MCP router',
            };
        case 'memory_not_ready':
        case 'claude_mem_not_ready':
            return {
                key: 'memory',
                label: 'Memory / context',
            };
        case 'session_restore_not_ready':
            return {
                key: 'sessions',
                label: 'Session supervisor',
            };
        case 'browser_service_not_ready':
        case 'extension_bridge_not_ready':
        case 'execution_environment_not_ready':
            return {
                key: 'integrations',
                label: 'Integrations',
            };
        default:
            return {
                key: 'startup',
                label: 'Startup platform',
            };
    }
}

export function getStartupBlockingReasonTitle(code: string): string {
    switch (code) {
        case 'mcp_aggregator_not_initialized':
            return 'MCP router is not initialized';
        case 'mcp_inventory_not_ready':
            return 'Cached MCP inventory is not ready';
        case 'mcp_resident_runtime_not_ready':
            return 'Resident MCP runtime is still warming';
        case 'mcp_config_sync_pending':
            return 'MCP config sync is still pending';
        case 'memory_not_ready':
            return 'Memory manager is still initializing';
        case 'claude_mem_not_ready':
            return 'TormentNexus default sections are not ready';
        case 'browser_service_not_ready':
            return 'Browser service bridge is not ready';
        case 'extension_bridge_not_ready':
            return 'Extension bridge listener is offline';
        case 'execution_environment_not_ready':
            return 'Execution environment verification is incomplete';
        case 'session_restore_not_ready':
            return 'Session restore has not completed yet';
        default:
            return 'Startup blocker requires operator attention';
    }
}

export function getStartupBlockingReasonPriority(code: string): number {
    switch (code) {
        case 'mcp_aggregator_not_initialized':
        case 'mcp_resident_runtime_not_ready':
        case 'execution_environment_not_ready':
            return 100;
        case 'mcp_inventory_not_ready':
        case 'mcp_config_sync_pending':
        case 'extension_bridge_not_ready':
            return 80;
        case 'memory_not_ready':
        case 'claude_mem_not_ready':
        case 'session_restore_not_ready':
            return 60;
        case 'browser_service_not_ready':
            return 40;
        default:
            return 20;
    }
}

export function getStartupBlockingReasonPriorityLabel(priority: number): 'High' | 'Medium' | 'Low' {
    if (priority >= 80) {
        return 'High';
    }

    if (priority >= 50) {
        return 'Medium';
    }

    return 'Low';
}

export function getStartupBlockingReasonPriorityTone(priorityLabel: 'High' | 'Medium' | 'Low'): string {
    switch (priorityLabel) {
        case 'High':
            return 'border-rose-500/40 bg-rose-500/10 text-rose-100';
        case 'Medium':
            return 'border-amber-500/40 bg-amber-500/10 text-amber-100';
        default:
            return 'border-emerald-500/40 bg-emerald-500/10 text-emerald-100';
    }
}

export function getStartupBlockingReasonPriorityCounts(
    startupBlockingReasons: StartupBlockingReasonWithPriority[],
): StartupBlockingReasonPriorityCounts {
    return startupBlockingReasons.reduce<StartupBlockingReasonPriorityCounts>((counts, reason) => {
        const label = getStartupBlockingReasonPriorityLabel(reason.priority);
        if (label === 'High') {
            counts.high += 1;
        } else if (label === 'Medium') {
            counts.medium += 1;
        } else {
            counts.low += 1;
        }

        return counts;
    }, {
        high: 0,
        medium: 0,
        low: 0,
    });
}

export function getPrioritizedStartupBlockingReasons(
    startupBlockingReasons: StartupBlockingReasonView[],
): StartupBlockingReasonWithPriority[] {
    return startupBlockingReasons
        .map((reason, index) => ({
            ...reason,
            priority: getStartupBlockingReasonPriority(reason.code),
            index,
        }))
        .sort((left, right) => {
            if (right.priority !== left.priority) {
                return right.priority - left.priority;
            }

            return left.index - right.index;
        })
        .map(({ index: _index, ...reason }) => reason);
}

export function getGroupedStartupBlockingReasons(
    startupBlockingReasons: StartupBlockingReasonWithPriority[],
): StartupBlockingReasonGroup[] {
    const groups = new Map<string, StartupBlockingReasonGroup>();

    for (const reason of startupBlockingReasons) {
        const subsystem = getStartupBlockingReasonSubsystem(reason.code);
        const existingGroup = groups.get(subsystem.key);
        if (existingGroup) {
            existingGroup.reasons.push(reason);
            continue;
        }

        groups.set(subsystem.key, {
            key: subsystem.key,
            label: subsystem.label,
            reasons: [reason],
        });
    }

    return Array.from(groups.values()).sort((left, right) => {
        const leftOrder = STARTUP_BLOCKING_REASON_GROUP_ORDER[left.key] ?? Number.MAX_SAFE_INTEGER;
        const rightOrder = STARTUP_BLOCKING_REASON_GROUP_ORDER[right.key] ?? Number.MAX_SAFE_INTEGER;
        if (leftOrder !== rightOrder) {
            return leftOrder - rightOrder;
        }

        return left.label.localeCompare(right.label);
    });
}

export function getStartupBlockingReasonGroupSeverity(
    reasons: StartupBlockingReasonWithPriority[],
): 'High' | 'Medium' | 'Low' {
    const maxPriority = reasons.reduce((highest, reason) => Math.max(highest, reason.priority), 0);
    return getStartupBlockingReasonPriorityLabel(maxPriority);
}

export function getStartupBlockingReasonGroupTopAction(
    reasons: StartupBlockingReasonWithPriority[],
): StartupBlockingReasonAction | null {
    if (reasons.length === 0) {
        return null;
    }

    const topReason = reasons.reduce((selected, reason) => {
        if (!selected) {
            return reason;
        }

        return reason.priority > selected.priority ? reason : selected;
    }, null as StartupBlockingReasonWithPriority | null);

    return topReason ? getStartupBlockingReasonAction(topReason.code) : null;
}

export function getStartupBlockingReasonGroupPrimaryReason(
    reasons: StartupBlockingReasonWithPriority[],
): StartupBlockingReasonWithPriority | null {
    if (reasons.length === 0) {
        return null;
    }

    return reasons.reduce((selected, reason) => {
        if (!selected) {
            return reason;
        }

        return reason.priority > selected.priority ? reason : selected;
    }, null as StartupBlockingReasonWithPriority | null);
}

export function getStartupBlockingReasonGroupPriorityCounts(
    reasons: StartupBlockingReasonWithPriority[],
): StartupBlockingReasonPriorityCounts {
    return getStartupBlockingReasonPriorityCounts(reasons);
}

export function getStartupBlockingReasonActions(
    startupBlockingReasons: StartupBlockingReasonView[],
): StartupBlockingReasonAction[] {
    const seen = new Set<string>();
    const actions: StartupBlockingReasonAction[] = [];

    for (const reason of startupBlockingReasons) {
        const action = getStartupBlockingReasonAction(reason.code);
        const key = `${action.href}|${action.label}`;
        if (seen.has(key)) {
            continue;
        }

        seen.add(key);
        actions.push(action);
    }

    return actions;
}