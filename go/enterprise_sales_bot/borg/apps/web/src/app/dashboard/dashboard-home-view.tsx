import Link from 'next/link';

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
    const claudeMem = memory.claudeMem;

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
                <header className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-2xl shadow-slate-950/30">
                    <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
                        <div className="space-y-3">
                            <span className="inline-flex items-center rounded-full border border-cyan-500/30 bg-cyan-500/10 px-3 py-1 text-xs font-semibold uppercase tracking-[0.24em] text-cyan-200">
                                TormentNexus 1.0 control plane
                            </span>
                            <div className="space-y-2">
                                <h1 className="text-3xl font-semibold tracking-tight text-white">Operator dashboard</h1>
                                <p className="max-w-3xl text-sm text-slate-300">
                                    One page to see whether the MCP router is alive, which providers can take traffic, and what your supervised CLI sessions are doing.
                                </p>
                            </div>
                        </div>
                        <div className="flex flex-wrap items-center gap-3 text-sm text-slate-300">
                            <span className="rounded-full border border-slate-700 bg-slate-950/70 px-3 py-1.5">
                                Refreshed {generatedAtLabel}
                            </span>
                            <Link
                                href="/dashboard/integrations"
                                title="Open the Integration Hub for browser extension installs, VS Code packaging, and MCP client sync"
                                aria-label="Open Integration Hub"
                                className="rounded-full border border-slate-700 bg-slate-950/70 px-3 py-1.5 font-medium text-cyan-200 transition hover:border-slate-600 hover:text-cyan-100"
                            >
                                Integration Hub →
                            </Link>
                        </div>
                    </div>

                    <div className="mt-6 grid gap-4 md:grid-cols-3">
                        {overviewMetrics.map((metric) => (
                            <div key={metric.label} className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">{metric.label}</p>
                                <p className="mt-2 text-3xl font-semibold text-white">{metric.value}</p>
                                <p className="mt-2 text-sm text-slate-400">{metric.detail}</p>
                            </div>
                        ))}
                    </div>

                    <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                        <div className="flex items-start justify-between gap-4">
                            <div>
                                <h2 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-300">Operator alerts</h2>
                                <p className="mt-1 text-sm text-slate-500">Cross-panel issues that deserve attention before you start driving the swarm.</p>
                            </div>
                            <span className="rounded-full border border-slate-700 bg-slate-900/70 px-3 py-1 text-xs font-medium text-slate-200">
                                {isBootstrapping ? 'Connecting' : (dashboardAlerts.length === 0 ? 'All clear' : `${dashboardAlerts.length} active`)}
                            </span>
                        </div>

                        {isBootstrapping ? (
                            <div className="mt-4 rounded-2xl border border-cyan-500/30 bg-cyan-500/10 p-4 text-sm text-cyan-100">
                                Connecting to live core telemetry. TormentNexus will replace these neutral placeholders as soon as the first startup snapshot arrives.
                            </div>
                        ) : dashboardAlerts.length === 0 ? (
                            <div className="mt-4 rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-4 text-sm text-emerald-100">
                                All major systems look healthy. Router, providers, and supervised sessions are not reporting any cross-panel alerts.
                            </div>
                        ) : (
                            <div className="mt-4 grid gap-3 lg:grid-cols-2">
                                {dashboardAlerts.map((alert) => (
                                    <div key={alert.id} className="rounded-2xl border border-slate-800 bg-slate-900/70 p-4">
                                        <div className="flex items-start justify-between gap-3">
                                            <div>
                                                <div className={`inline-flex rounded-full border px-2.5 py-1 text-xs font-medium uppercase tracking-[0.2em] ${getAlertTone(alert.severity)}`}>
                                                    {alert.severity}
                                                </div>
                                                <h3 className="mt-3 text-base font-semibold text-white">{alert.title}</h3>
                                                <p className="mt-2 text-sm text-slate-300">{alert.detail}</p>
                                            </div>
                                        </div>
                                        <Link
                                            href={alert.href}
                                            title={alert.hrefLabel}
                                            aria-label={alert.hrefLabel}
                                            className="mt-4 inline-flex text-sm font-medium text-cyan-200 transition hover:text-cyan-100"
                                        >
                                            {alert.hrefLabel} →
                                        </Link>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </header>

                <div className="grid gap-6 xl:grid-cols-2">
                    <section className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/20">
                        <div className="flex items-start justify-between gap-4">
                            <div>
                                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-200">Overview</p>
                                <h2 className="mt-2 text-xl font-semibold text-white">Router posture</h2>
                                <p className="mt-2 text-sm text-slate-400">Quick health readout for first-time operators.</p>
                            </div>
                            <div className={`rounded-full border px-3 py-1 text-xs font-medium ${routerStatusTone}`}>
                                {routerStatusLabel}
                            </div>
                        </div>


                        <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                            <div className="flex items-start justify-between gap-4">
                                <div>
                                    <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-300">Startup readiness</h3>
                                    <p className="mt-1 text-sm text-slate-500">{startupSummary || 'Boot checks reported directly from core startup state.'}</p>
                                </div>
                                <span className={`rounded-full border px-3 py-1 text-xs font-medium ${startupToneClass}`}>
                                    {startupLabel}
                                </span>
                            </div>

                            <div className="mt-4 grid gap-3 sm:grid-cols-2">
                                {startupChecklist.map((item) => (
                                    <div key={item.label} className="rounded-2xl border border-slate-800 bg-slate-900/70 p-3 text-sm">
                                        <div className="flex items-center justify-between gap-3">
                                            <span className="font-medium text-white">{item.label}</span>
                                            <span className={`rounded-full border px-2.5 py-1 text-xs font-medium ${item.ready ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-200' : 'border-amber-500/30 bg-amber-500/10 text-amber-200'}`}>
                                                {item.ready ? 'Ready' : 'Pending'}
                                            </span>
                                        </div>
                                        <p className="mt-2 text-slate-400">{item.detail}</p>
                                    </div>
                                ))}
                            </div>

                            {startupBlockingReasons.length > 0 ? (
                                <div className="mt-4 rounded-2xl border border-amber-500/30 bg-amber-500/10 p-4">
                                    <div className="flex items-start justify-between gap-3">
                                        <div>
                                            <h4 className="text-sm font-semibold uppercase tracking-[0.2em] text-amber-200">Blocking reasons</h4>
                                            <p className="mt-1 text-xs text-amber-100/80">Live reasons reported by core startup checks.</p>
                                        </div>
                                        <span className="rounded-full border border-amber-500/40 px-2.5 py-1 text-xs font-medium text-amber-200">
                                            {startupBlockingReasons.length} pending
                                        </span>
                                    </div>

                                    {startupBlockingActions.length > 0 ? (
                                        <div className="mt-3 flex flex-wrap gap-2">
                                            <span className="text-xs uppercase tracking-[0.16em] text-amber-200">Suggested actions:</span>
                                            {startupBlockingActions.map((action) => (
                                                <Link
                                                    key={`${action.href}-${action.label}`}
                                                    href={action.href}
                                                    className="inline-flex rounded-full border border-cyan-500/30 bg-cyan-500/10 px-2.5 py-1 text-xs font-medium text-cyan-200 transition hover:border-cyan-400 hover:bg-cyan-500/20"
                                                >
                                                    {action.label}
                                                </Link>
                                            ))}
                                        </div>
                                    ) : null}

                                    <p className="mt-2 text-xs text-amber-100/80">
                                        Priority mix: {startupBlockingPriorityCounts.high} high · {startupBlockingPriorityCounts.medium} medium · {startupBlockingPriorityCounts.low} low
                                    </p>

                                    <div className="mt-3 space-y-3">
                                        {startupBlockingReasonGroups.map((group) => {
                                            const groupSeverity = getStartupBlockingReasonGroupSeverity(group.reasons);
                                            const groupSeverityTone = getStartupBlockingReasonPriorityTone(groupSeverity);
                                            const groupTopAction = getStartupBlockingReasonGroupTopAction(group.reasons);
                                            const groupImpactedChecks = getStartupBlockingReasonGroupImpactedChecks(group.reasons);
                                            const groupPrimaryReason = getStartupBlockingReasonGroupPrimaryReason(group.reasons);
                                            const groupPrimaryReasonTitle = groupPrimaryReason
                                                ? getStartupBlockingReasonTitle(groupPrimaryReason.code)
                                                : null;
                                            const groupPriorityCounts = getStartupBlockingReasonGroupPriorityCounts(group.reasons);

                                            return (
                                            <section key={group.key} className="rounded-xl border border-amber-500/20 bg-slate-950/30 p-3">
                                                <div className="flex items-center justify-between gap-2">
                                                    <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-amber-200">{group.label}</div>
                                                    <div className="flex items-center gap-2">
                                                        <span className={`rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] ${groupSeverityTone}`}>
                                                            {groupSeverity} group
                                                        </span>
                                                        <span className="rounded-full border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-amber-100">
                                                            {group.reasons.length} item{group.reasons.length === 1 ? '' : 's'}
                                                        </span>
                                                    </div>
                                                </div>
                                                {groupTopAction ? (
                                                    <Link
                                                        href={groupTopAction.href}
                                                        className="mt-2 inline-flex text-xs font-medium text-cyan-300 underline transition hover:text-cyan-200"
                                                    >
                                                        Top action: {groupTopAction.label} →
                                                    </Link>
                                                ) : null}
                                                {groupImpactedChecks.length > 0 ? (
                                                    <div className="mt-2 text-xs text-amber-100/80">
                                                        Impacts: {groupImpactedChecks.map((check) => check.label).join(' · ')}
                                                    </div>
                                                ) : null}
                                                {groupPrimaryReasonTitle ? (
                                                    <div className="mt-1 text-xs text-amber-100/80">
                                                        Primary blocker: {groupPrimaryReasonTitle}
                                                    </div>
                                                ) : null}
                                                <div className="mt-1 text-xs text-amber-100/80">
                                                    Group mix: {groupPriorityCounts.high} high · {groupPriorityCounts.medium} medium · {groupPriorityCounts.low} low
                                                </div>
                                                <ul className="mt-2 space-y-2">
                                                    {group.reasons.map((reason) => {
                                                        const action = getStartupBlockingReasonAction(reason.code);
                                                        const priorityLabel = getStartupBlockingReasonPriorityLabel(reason.priority);
                                                        const priorityTone = getStartupBlockingReasonPriorityTone(priorityLabel);
                                                        const reasonTitle = getStartupBlockingReasonTitle(reason.code);

                                                        return (
                                                            <li key={`${reason.code}-${reason.detail}`} className="rounded-xl border border-amber-500/20 bg-slate-950/40 p-3 text-sm text-amber-50">
                                                                <div className="flex items-center justify-between gap-2">
                                                                    <div className="text-sm font-medium text-amber-50">{reasonTitle}</div>
                                                                    <span className={`rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.12em] ${priorityTone}`}>
                                                                        {priorityLabel} priority
                                                                    </span>
                                                                </div>
                                                                <div className="mt-1 text-xs uppercase tracking-[0.16em] text-amber-200">{reason.code}</div>
                                                                <div className="mt-1">{reason.detail}</div>
                                                                <Link
                                                                    href={action.href}
                                                                    className="mt-2 inline-flex text-xs font-medium text-cyan-300 underline transition hover:text-cyan-200"
                                                                >
                                                                    {action.label} →
                                                                </Link>
                                                            </li>
                                                        );
                                                    })}
                                                </ul>
                                            </section>
                                            );
                                        })}
                                    </div>
                                </div>
                            ) : null}
                        </div>

                        <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                            <div className="flex items-start justify-between gap-4">
                                <div>
                                    <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-300">Install &amp; connect TormentNexus</h3>
                                    <p className="mt-1 text-sm text-slate-500">Fast path for getting browser bridges, editor surfaces, and managed MCP configs into the tools you already use.</p>
                                </div>
                                <Link
                                    href="/dashboard/integrations"
                                    title="Open install surfaces, browser extension artifacts, VS Code packaging, and client sync targets"
                                    aria-label="Open Integration Hub from router posture section"
                                    className="text-sm font-medium text-cyan-200 transition hover:text-cyan-100"
                                >
                                    Open Integration Hub →
                                </Link>
                            </div>

                            <div className="mt-4 grid gap-3 sm:grid-cols-3 text-xs text-slate-400">
                                <div>
                                    <span className="font-semibold text-slate-200">Browser extensions</span>: Load Chromium/Edge and Firefox bundles.
                                </div>
                                <div>
                                    <span className="font-semibold text-slate-200">Editor surfaces</span>: Package and install VS Code extension.
                                </div>
                                <div>
                                    <span className="font-semibold text-slate-200">Client config sync</span>: Push endpoints to Cursor, Claude, and VS Code.
                                </div>
                            </div>
                        </div>
                    </section>

                    <section className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/20">
                        <div className="flex items-start justify-between gap-4">
                            <div>
                                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-200">MCP Router</p>
                                <h2 className="mt-2 text-xl font-semibold text-white">Server health and traffic</h2>
                                <p className="mt-2 text-sm text-slate-400">Live server posture plus the latest router activity.</p>
                            </div>
                            <Link
                                href="/dashboard/mcp"
                                title="Open the full MCP router dashboard with server list, tools, and configuration controls"
                                aria-label="Open detailed MCP dashboard"
                                className="text-sm font-medium text-cyan-200 transition hover:text-cyan-100"
                            >
                                Detailed MCP view →
                            </Link>
                        </div>

                        <div className="mt-6 space-y-3">
                            {servers.length === 0 ? (
                                <div className="rounded-2xl border border-dashed border-slate-700 bg-slate-950/50 p-4 text-sm text-slate-400">
                                    No MCP servers registered yet.
                                </div>
                            ) : servers.map((server, index) => (
                                <div
                                    key={`${server.name}-${server.config.command}-${server.config.args.join(' ')}-${index}`}
                                    className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4"
                                >
                                    <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                                        <div>
                                            <div className="flex items-center gap-3">
                                                <h3 className="text-base font-semibold text-white">{server.name}</h3>
                                                <span className={`rounded-full border px-2.5 py-1 text-xs font-medium ${getServerTone(server.status)}`}>
                                                    {sentenceCase(server.status)}
                                                </span>
                                            </div>
                                            <p className="mt-2 break-all font-mono text-xs text-slate-400">
                                                {server.config.command} {server.config.args.join(' ')}
                                            </p>
                                        </div>
                                        <div className="text-right text-sm text-slate-300">
                                            <div>{server.toolCount} tools</div>
                                            <div className="text-xs text-slate-500">{server.config.env.length} env vars</div>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>

                        <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                            <div className="flex items-center justify-between gap-4">
                                <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-300">Recent traffic</h3>
                                <Link
                                    href="/dashboard/mcp/inspector"
                                    title="Open the live MCP inspector to trace requests, responses, and tool invocations"
                                    aria-label="Open MCP traffic inspector"
                                    className="text-sm font-medium text-cyan-200 transition hover:text-cyan-100"
                                >
                                    Open inspector →
                                </Link>
                            </div>

                            <div className="mt-4 space-y-3">
                                {traffic.length === 0 ? (
                                    <p className="text-sm text-slate-400">No router traffic captured yet.</p>
                                ) : traffic.slice(0, 5).map((event, index) => (
                                    <div key={`${event.server}-${event.method}-${event.timestamp}-${index}`} className="rounded-2xl border border-slate-800 bg-slate-900/70 p-3 text-sm">
                                        <div className="flex items-center justify-between gap-4">
                                            <div className="flex items-center gap-3">
                                                <span className={`inline-flex h-2.5 w-2.5 rounded-full ${event.success ? 'bg-emerald-400' : 'bg-rose-400'}`} />
                                                <span className="font-medium text-white">{event.server}</span>
                                                <span className="rounded-full border border-slate-700 px-2 py-0.5 text-xs text-slate-300">{event.method}</span>
                                            </div>
                                            <span className="text-xs text-slate-500">{formatRelativeTimestamp(event.timestamp, currentTimestamp)}</span>
                                        </div>
                                        <p className="mt-2 text-sm text-slate-300">{summarizeTrafficEvent(event)}</p>
                                        <div className="mt-2 text-xs text-slate-500">Latency {event.latencyMs}ms</div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </section>

                    <section className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/20">
                        <div className="flex items-start justify-between gap-4">
                            <div>
                                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-200">Sessions</p>
                                <h2 className="mt-2 text-xl font-semibold text-white">Supervised CLI runtime</h2>
                                <p className="mt-2 text-sm text-slate-400">Inspect live session state, recent output, and restart posture.</p>
                            </div>
                            <span className="rounded-full border border-slate-700 bg-slate-950/70 px-3 py-1 text-xs font-medium text-slate-200">
                                {sessions.length} total
                            </span>
                        </div>

                        <div className="mt-6 space-y-3">
                            {sessions.length === 0 ? (
                                <div className="rounded-2xl border border-dashed border-slate-700 bg-slate-950/50 p-4 text-sm text-slate-400">
                                    No supervised sessions are registered yet.
                                </div>
                            ) : sessions.map((session) => {
                                const latestLog = session.logs[session.logs.length - 1];
                                const isPending = pendingSessionActionId === session.id;
                                const isRunning = session.status === 'running';
                                const canStart = session.status === 'created' || session.status === 'stopped' || session.status === 'error';
                                const canStop = session.status === 'starting' || session.status === 'running' || session.status === 'restarting';

                                return (
                                    <div key={session.id} className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                                        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                                            <div className="min-w-0">
                                                <div className="flex flex-wrap items-center gap-3">
                                                    <h3 className="text-base font-semibold text-white">{session.name}</h3>
                                                    <span className={`rounded-full border px-2.5 py-1 text-xs font-medium ${getSessionTone(session.status)}`}>
                                                        {sentenceCase(session.status)}
                                                    </span>
                                                    <span className="rounded-full border border-slate-700 px-2.5 py-1 text-xs text-slate-300">
                                                        {session.cliType}
                                                    </span>
                                                </div>
                                                <p className="mt-2 break-all font-mono text-xs text-slate-400">{session.worktreePath ?? session.workingDirectory}</p>
                                                <p className="mt-2 text-xs text-slate-500">
                                                    Last activity {formatRelativeTimestamp(session.lastActivityAt, currentTimestamp)} · Restarted {session.restartCount}/{session.maxRestartAttempts}
                                                </p>
                                                {session.autoRestart === false ? (
                                                    <p className="mt-2 text-xs text-amber-300">
                                                        Manual restart only · TormentNexus will not auto-restart this session after a crash.
                                                    </p>
                                                ) : null}
                                                {session.status === 'restarting' && session.scheduledRestartAt ? (
                                                    <p className="mt-2 text-xs text-amber-300">
                                                        Restart queued {formatRestartCountdown(session.scheduledRestartAt, currentTimestamp)}
                                                    </p>
                                                ) : null}
                                                {latestLog ? (
                                                    <div className="mt-3 rounded-xl border border-slate-800 bg-slate-900/70 p-3 text-sm text-slate-300">
                                                        <div className="mb-2 flex items-center justify-between gap-3 text-xs uppercase tracking-[0.16em] text-slate-500">
                                                            <span>Latest {latestLog.stream}</span>
                                                            <span>{formatRelativeTimestamp(latestLog.timestamp, currentTimestamp)}</span>
                                                        </div>
                                                        <p className="line-clamp-3 whitespace-pre-wrap break-words">{latestLog.message}</p>
                                                    </div>
                                                ) : null}
                                                {session.lastError ? (
                                                    <p className="mt-3 text-sm text-rose-300">{session.lastError}</p>
                                                ) : null}
                                            </div>

                                            <div className="flex flex-wrap gap-2">
                                                <button
                                                    type="button"
                                                    disabled={!onStartSession || !canStart || isPending}
                                                    onClick={() => onStartSession?.(session.id)}
                                                    title={`Start session ${session.name} (${session.cliType})`}
                                                    aria-label={`Start session ${session.name}`}
                                                    className="rounded-full border border-emerald-500/30 bg-emerald-500/10 px-4 py-2 text-sm font-medium text-emerald-200 transition hover:border-emerald-400 hover:bg-emerald-500/20 disabled:cursor-not-allowed disabled:opacity-50"
                                                >
                                                    {isPending && canStart ? 'Starting…' : isRunning ? 'Running' : 'Start'}
                                                </button>
                                                <button
                                                    type="button"
                                                    disabled={!onStopSession || !canStop || isPending}
                                                    onClick={() => onStopSession?.(session.id)}
                                                    title={`Stop session ${session.name}`}
                                                    aria-label={`Stop session ${session.name}`}
                                                    className="rounded-full border border-slate-700 bg-slate-900/70 px-4 py-2 text-sm font-medium text-slate-100 transition hover:border-slate-500 disabled:cursor-not-allowed disabled:opacity-50"
                                                >
                                                    {isPending && canStop ? 'Stopping…' : 'Stop'}
                                                </button>
                                                <button
                                                    type="button"
                                                    disabled={!onRestartSession || isPending}
                                                    onClick={() => onRestartSession?.(session.id)}
                                                    title={`Restart session ${session.name} and reattach supervision`}
                                                    aria-label={`Restart session ${session.name}`}
                                                    className="rounded-full border border-cyan-500/30 bg-cyan-500/10 px-4 py-2 text-sm font-medium text-cyan-200 transition hover:border-cyan-400 hover:bg-cyan-500/20 disabled:cursor-not-allowed disabled:opacity-50"
                                                >
                                                    {isPending ? 'Working…' : 'Restart'}
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    </section>

                    <section className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/20">
                        <div className="flex items-start justify-between gap-4">
                            <div>
                                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-200">Providers</p>
                                <h2 className="mt-2 text-xl font-semibold text-white">Quota and fallback posture</h2>
                                <p className="mt-2 text-sm text-slate-400">Which providers are configured, how much headroom remains, and where fallback will go next.</p>
                            </div>
                            <Link
                                href="/dashboard/billing"
                                title="Open provider billing and quota analytics with fallback chain controls"
                                aria-label="Open detailed provider billing dashboard"
                                className="text-sm font-medium text-cyan-200 transition hover:text-cyan-100"
                            >
                                Detailed provider view →
                            </Link>
                        </div>

                        <div className="mt-6 space-y-3">
                            {providers.length === 0 ? (
                                <div className="rounded-2xl border border-dashed border-slate-700 bg-slate-950/50 p-4 text-sm text-slate-400">
                                    No provider data available yet. Configure an API key or OAuth-backed provider in Billing to unlock fallback routing.
                                </div>
                            ) : providers.map((provider) => {
                                const usagePercent = getQuotaUsagePercent(provider);
                                return (
                                    <div key={provider.provider} className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                                        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                                            <div>
                                                <div className="flex items-center gap-3">
                                                    <h3 className="text-base font-semibold text-white">{provider.name}</h3>
                                                    <span className={`rounded-full border px-2.5 py-1 text-xs font-medium ${getProviderTone(provider)}`}>
                                                        {provider.configured ? sentenceCase(provider.availability ?? (provider.authenticated ? 'healthy' : 'degraded')) : 'Not configured'}
                                                    </span>
                                                </div>
                                                <p className="mt-2 text-sm text-slate-400">
                                                    {provider.authMethod ? `${provider.authMethod} · ` : ''}{provider.tier}
                                                    {provider.resetDate ? ` · resets ${provider.resetDate}` : ''}
                                                </p>
                                                {provider.lastError ? (
                                                    <p className="mt-2 text-sm text-rose-300">{provider.lastError}</p>
                                                ) : null}
                                            </div>
                                            <div className="text-right text-sm text-slate-300">
                                                <div>Used {formatQuotaValue(provider.used)}</div>
                                                <div>Remaining {formatQuotaValue(provider.remaining)}</div>
                                            </div>
                                        </div>
                                        <div className="mt-4">
                                            <div className="h-2 rounded-full bg-slate-800">
                                                <div
                                                    className="h-2 rounded-full bg-cyan-400 transition-all"
                                                    style={{ width: `${usagePercent ?? 100}%` }}
                                                />
                                            </div>
                                            <div className="mt-2 flex items-center justify-between text-xs text-slate-500">
                                                <span>Limit {formatQuotaValue(provider.limit)}</span>
                                                <span>{usagePercent === null ? 'Usage limit unavailable' : `${usagePercent}% used`}</span>
                                            </div>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>

                        <div className="mt-6 rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                            <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-300">Fallback chain</h3>
                            <div className="mt-4 space-y-2">
                                {fallbackChain.length === 0 ? (
                                    <p className="text-sm text-slate-400">No fallback chain is exposed yet. Configure providers to populate the routing order.</p>
                                ) : fallbackChain.map((entry) => (
                                    <div key={`${entry.priority}-${entry.provider}-${entry.model ?? 'default'}`} className="flex items-center justify-between gap-4 rounded-2xl border border-slate-800 bg-slate-900/70 px-4 py-3 text-sm">
                                        <div>
                                            <span className="font-medium text-white">{entry.priority}. {formatFallbackLabel(entry)}</span>
                                            <p className="mt-1 text-xs text-slate-500">{entry.reason}</p>
                                        </div>
                                        <span className="rounded-full border border-slate-700 px-2.5 py-1 text-xs text-slate-300">priority {entry.priority}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </section>
                    <section className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/20">
                        <div className="flex items-start justify-between gap-4">
                            <div>
                                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-200">Immune System</p>
                                <h2 className="mt-2 text-xl font-semibold text-white">Autonomous healer status</h2>
                                <p className="mt-2 text-sm text-slate-400">Real-time pathogen detection, auto-neutralization efficacy, and L2 Vault memory from the Go HealerService.</p>
                            </div>
                            <Link href="/dashboard/healer" title="Open the full Immune System dashboard with live healing stream and persistent vault records" aria-label="Open detailed healer dashboard" className="text-sm font-medium text-cyan-200 transition hover:text-cyan-100">
                                Full Immune System view &rarr;
                            </Link>
                        </div>
                        <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                            <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                                <dt className="text-sm text-slate-400">Active pathogens</dt>
                                <dd className="mt-2 flex items-center gap-2">
                                    <span className="text-2xl font-semibold text-white">{healerStatus ? healerStatus.activePathogens : '—'}</span>
                                    {healerStatus && healerStatus.activePathogens > 0 && (
                                        <span className="h-2.5 w-2.5 rounded-full bg-rose-500 animate-pulse" />
                                    )}
                                </dd>
                                <p className="mt-2 text-sm text-slate-400">{healerStatus ? (healerStatus.activePathogens > 0 ? 'Unresolved errors detected' : 'System is clean') : 'Waiting for healer telemetry'}</p>
                            </div>
                            <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                                <dt className="text-sm text-slate-400">Auto-neutralized</dt>
                                <dd className="mt-2 text-2xl font-semibold text-emerald-300">{healerStatus ? healerStatus.resolvedCount : '—'}</dd>
                                <p className="mt-2 text-sm text-slate-400">{healerStatus && healerStatus.resolvedCount > 0 ? 'Errors healed autonomously' : 'No healing events yet'}</p>
                            </div>
                            <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                                <dt className="text-sm text-slate-400">Immune efficacy</dt>
                                <dd className="mt-2 text-2xl font-semibold text-amber-300">{healerStatus ? `${healerStatus.successRate}%` : '—'}</dd>
                                <p className="mt-2 text-sm text-slate-400">{healerStatus ? (healerStatus.successRate >= 90 ? 'Excellent coverage' : healerStatus.successRate >= 70 ? 'Acceptable coverage' : 'Needs attention') : 'Efficiency metric pending'}</p>
                            </div>
                            <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
                                <dt className="text-sm text-slate-400">Vault records</dt>
                                <dd className="mt-2 text-2xl font-semibold text-blue-300">{healerStatus ? healerStatus.vaultRecordCount : '—'}</dd>
                                <p className="mt-2 text-sm text-slate-400">{healerStatus && healerStatus.vaultRecordCount > 0 ? 'Persistent L2 memories stored' : 'Vault awaiting first commit'}</p>
                            </div>
                        </div>
                        <div className="mt-4 flex items-center gap-3 text-xs text-slate-500">
                            {healerStatus?.isLive ? (
                                <>
                                    <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
                                    <span>Healer service connected &mdash; live telemetry active</span>
                                    {healerStatus.lastHealTime && (
                                        <span className="text-slate-600">&middot; Last heal {healerStatus.lastHealTime}</span>
                                    )}
                                </>
                            ) : (
                                <>
                                    <span className="h-2 w-2 rounded-full bg-slate-600" />
                                    <span>Healer service not connected &mdash; data reflects last known state</span>
                                </>
                            )}
                        </div>
                    </section>
                    {children}
                </div>
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