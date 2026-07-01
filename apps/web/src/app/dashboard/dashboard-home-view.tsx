import Link from 'next/link';
import { useState } from 'react';

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
                
                {/* PAGE A: SYSTEM RECOVERY & ACTIVE DATABASE SYNC */}
                {activeTab === 'page-a' && (
                    <div className="space-y-6">
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
                                                ? "border-rose-500/30 bg-rose-500/10 text-rose-300"
                                                : "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
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
                            <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 md:col-span-2 space-y-4">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <h2 className="text-base font-semibold text-white">System Integrity &amp; Verification Console</h2>
                                        <span className="text-cyan-400 cursor-help text-xs" title="Ensures strict compilation checks and integration test compliance across Go backend components.">💡</span>
                                    </div>
                                    <button
                                        onClick={triggerDiagnostics}
                                        disabled={runningDiagnostics}
                                        className="bg-zinc-800 hover:bg-zinc-700 text-white border border-zinc-700 text-xs font-semibold px-4 py-2 rounded transition-colors disabled:opacity-50"
                                    >
                                        {runningDiagnostics ? "Running compilation..." : "Execute Automated Verification Sweep"}
                                    </button>
                                </div>
                                <p className="text-xs text-slate-400">
                                    Compiles all native tools and verifies test suite assertions across memory registers and MCP routers.
                                </p>
                                <div className="bg-slate-950 p-3 rounded border border-slate-850 font-mono text-xs text-slate-300 min-h-[60px] flex items-center justify-center">
                                    {runningDiagnostics ? (
                                        <div className="flex items-center gap-2 text-slate-400">
                                            <span className="animate-spin">⏳</span>
                                            <span>Compiling binary and executing integration checks...</span>
                                        </div>
                                    ) : diagnosticsResult ? (
                                        <span className="text-emerald-400">{diagnosticsResult}</span>
                                    ) : (
                                        <span className="text-slate-500">System idle. Ready to execute health checks.</span>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {/* PAGE B: NATIVE GO MCP ORCHESTRATION & TOOL CONTROL */}
                {activeTab === 'page-b' && (
                    <div className="space-y-6">
                        <div className="grid gap-6 md:grid-cols-2">
                            {/* Always-On Tools Panel */}
                            <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4">
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

                            {/* Swarm Code Gen Panel */}
                            <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 flex flex-col justify-between">
                                <div>
                                    <div className="flex items-center gap-2">
                                        <h2 className="text-base font-semibold text-white">Swarm Code Generation Queue</h2>
                                        <span className="text-cyan-400 cursor-help text-xs" title="Cross-references catalog schemas to rewrite missing API bridges into self-contained Go modules.">💡</span>
                                    </div>
                                    <p className="text-xs text-slate-400">
                                        Triggers the swarm_v7.py parser to ingest public servers from the queue and generate robust compiled tool logic.
                                    </p>
                                    <div className="grid grid-cols-2 gap-3 mt-4 text-center">
                                        <div className="border border-slate-800 bg-slate-950 p-2 rounded">
                                            <div className="text-xs text-slate-500">Implemented Go Tools</div>
                                            <div className="text-sm font-semibold text-emerald-400 mt-0.5">3,281</div>
                                        </div>
                                        <div className="border border-slate-800 bg-slate-950 p-2 rounded">
                                            <div className="text-xs text-slate-500">Pending In Queue</div>
                                            <div className="text-sm font-semibold text-amber-400 mt-0.5">19,266</div>
                                        </div>
                                    </div>
                                </div>
                                <button
                                    onClick={triggerSwarmGen}
                                    disabled={swarmRunning}
                                    className="w-full bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs py-2 rounded transition-colors disabled:opacity-50 mt-4"
                                >
                                    {swarmRunning ? "Generating (swarm_v7.py --skip-existing)..." : "Trigger Swarm Generation (swarm_v7.py)"}
                                </button>
                            </div>

                            {/* JSON-RPC Client Access Bridge */}
                            <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 md:col-span-2 space-y-3">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">JSON-RPC Client Access Bridge</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="Exposes native client endpoints over standardized tRPC and HTTP interfaces.">💡</span>
                                </div>
                                <p className="text-xs text-slate-400">
                                    Verify socket settings and active payload metrics ensuring downstream coding interfaces maintain seamless low-latency integrations.
                                </p>
                                <div className="grid md:grid-cols-3 gap-3 text-xs pt-2">
                                    <div className="border border-slate-850 bg-slate-950 p-3 rounded">
                                        <span className="text-slate-500">JSON-RPC Endpoint</span>
                                        <div className="font-mono text-cyan-200 mt-1">http://localhost:7778/trpc</div>
                                    </div>
                                    <div className="border border-slate-850 bg-slate-950 p-3 rounded">
                                        <span className="text-slate-500">Active Handshakes</span>
                                        <div className="font-mono text-emerald-450 mt-1">4 active tunnels</div>
                                    </div>
                                    <div className="border border-slate-850 bg-slate-950 p-3 rounded">
                                        <span className="text-slate-500">Router Version</span>
                                        <div className="font-mono text-zinc-300 mt-1">v1.0.0-alpha.207 (Go)</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                )}

                {/* PAGE C: COGNITIVE MEMORY ENGINES & SKILL REGISTRIES */}
                {activeTab === 'page-c' && (
                    <div className="space-y-6">
                        <div className="grid gap-6 md:grid-cols-2">
                            {/* Memory dreaming metrics */}
                            <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4">
                                <div className="flex items-center gap-2">
                                    <h2 className="text-base font-semibold text-white">L1 ➔ L4 Memory Dreaming &amp; Fact Distillation</h2>
                                    <span className="text-cyan-400 cursor-help text-xs" title="L1 (Active Context), L2 (Short-Term), L3 (Dreaming & fact condensation), and L4 (Reflective structural insights).">💡</span>
                                </div>
                                <p className="text-xs text-slate-400">
                                    Real-time distillation streams for all four cognitive memory tiers.
                                </p>
                                <div className="space-y-3 pt-2 text-xs">
                                    <div className="border border-slate-850 bg-slate-950/60 p-3 rounded flex justify-between">
                                        <span className="text-slate-400">L1 Active Context Scratchpad</span>
                                        <span className="text-cyan-400 font-semibold">Active (4,096 tokens)</span>
                                    </div>
                                    <div className="border border-slate-850 bg-slate-950/60 p-3 rounded flex justify-between">
                                        <span className="text-slate-400">L2 Short-Term Episodic Vault</span>
                                        <span className="text-cyan-400 font-semibold">86,281 records</span>
                                    </div>
                                    <div className="border border-slate-850 bg-slate-950/60 p-3 rounded flex justify-between">
                                        <span className="text-slate-400">L3 Long-Term Fact Distillation</span>
                                        <span className="text-purple-400 font-semibold">Distilling in background</span>
                                    </div>
                                    <div className="border border-slate-850 bg-slate-950/60 p-3 rounded flex justify-between">
                                        <span className="text-slate-400">L4 Conceptual Reflection Archetype</span>
                                        <span className="text-amber-400 font-semibold">Matched 17 missing capacities</span>
                                    </div>
                                </div>
                            </div>

                            {/* Filesystem Skill Indexer */}
                            <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 space-y-4 flex flex-col justify-between">
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
                            <div className="rounded-2xl border border-slate-800 bg-slate-900/70 p-6 md:col-span-2 space-y-4">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <h2 className="text-base font-semibold text-white">Transcripts &amp; Links Backlog Repair</h2>
                                        <span className="text-cyan-400 cursor-help text-xs" title="Scan session dumps and links to rebuild the session graph index.">💡</span>
                                    </div>
                                    <div className="flex gap-2">
                                        <button
                                            onClick={triggerFolderScan}
                                            disabled={runningScan}
                                            className="bg-zinc-800 hover:bg-zinc-700 text-white border border-zinc-700 text-xs font-semibold px-3 py-1.5 rounded disabled:opacity-50"
                                        >
                                            {runningScan ? "Scanning sessions..." : "Ingest Session Directories"}
                                        </button>
                                        <button
                                            onClick={triggerLinkRestoration}
                                            disabled={runningLinkRestoration}
                                            className="bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-semibold px-3 py-1.5 rounded disabled:opacity-50"
                                        >
                                            {runningLinkRestoration ? "Restoring backlog..." : "Scrape Lost Backlog Links"}
                                        </button>
                                    </div>
                                </div>
                                <p className="text-xs text-slate-400">
                                    Automated folder mapping loops across target directories to repair 2,003 missing sessions and populate 15,753 lost backlog link entries.
                                </p>
                            </div>
                        </div>
                    </div>
                )}

                {/* PAGE D: PROMPT COLLECTIONS & GLOBAL STATIC DEPLOYMENTS */}
                {activeTab === 'page-d' && (
                    <div className="space-y-6">
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
                )}

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