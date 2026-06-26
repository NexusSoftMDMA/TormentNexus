import { readFile } from 'node:fs/promises';
import path from 'node:path';

import type { RegisteredBridgeClient } from '../bridge/bridge-manifest.js';
import type { ExecutionEnvironmentSummary } from '../services/execution-environment.js';

let tormentnexusVersionPromise: Promise<string> | null = null;

async function getTormentNexusVersion(): Promise<string> {
    if (!tormentnexusVersionPromise) {
        tormentnexusVersionPromise = (async () => {
            try {
                const version = await readFile(path.join(process.cwd(), 'VERSION'), 'utf-8');
                const trimmed = version.trim();
                if (trimmed) {
                    return trimmed;
                }
            } catch {
                // fall through to package.json
            }

            try {
                const pkg = JSON.parse(await readFile(path.join(process.cwd(), 'package.json'), 'utf-8')) as { version?: string };
                if (pkg.version?.trim()) {
                    return pkg.version.trim();
                }
            } catch {
                // fall through to unknown
            }

            return 'unknown';
        })();
    }

    return tormentnexusVersionPromise;
}

type StartupStatusInput = {
    mcpServer: unknown;
    aggregator?: {
        getInitializationStatus?: () => {
            inProgress: boolean;
            initialized: boolean;
            lastStartedAt?: number;
            lastCompletedAt?: number;
            lastSuccessAt?: number;
            lastError?: string;
            connectedClientCount: number;
            configuredServerCount: number;
        };
    } | null;
    agentMemory?: unknown;
    browserService?: unknown;
    browserStatus: {
        active?: boolean;
        pageCount?: number;
        pageIds?: string[];
    };
    sessionSupervisor?: {
        getRestoreStatus?: () => {
            lastRestoreAt?: number;
            restoredSessionCount: number;
            autoResumeCount: number;
        };
    } | null;
    sessionCount: number;
    mcpConfigService?: {
        getStatus?: () => {
            inProgress: boolean;
            lastStartedAt?: number;
            lastCompletedAt?: number;
            lastSuccessAt?: number;
            lastError?: string;
            lastServerCount: number;
            lastToolCount: number;
        };
    } | null;
    liveServerCount: number;
    residentLiveServerCount?: number;
    warmingServerCount?: number;
    failedWarmupServerCount?: number;
    persistedServerCount: number;
    persistedToolCount: number;
    persistedAlwaysOnServerCount: number;
    persistedAlwaysOnToolCount: number;
	lazyMode?: boolean;
    inventorySource?: 'database' | 'config' | 'empty';
    inventorySnapshotUpdatedAt?: string | null;
    executionEnvironment?: ExecutionEnvironmentSummary | null;
    tormentnexus?: {
        enabled: boolean;
        storePath: string | null;
        storeExists: boolean;
        totalEntries: number;
        sectionCount: number;
        defaultSectionCount: number;
        presentDefaultSectionCount: number;
        missingSections: string[];
        lastUpdatedAt: string | null;
    } | null;
};

type StartupBlockingReasonCode =
    | 'mcp_aggregator_not_initialized'
    | 'mcp_inventory_not_ready'
    | 'mcp_resident_runtime_not_ready'
    | 'mcp_config_sync_pending'
    | 'memory_not_ready'
    | 'tormentnexus_not_ready'
    | 'browser_service_not_ready'
    | 'session_restore_not_ready'
    | 'extension_bridge_not_ready'
    | 'execution_environment_not_ready';

type StartupBlockingReason = {
    code: StartupBlockingReasonCode;
    detail: string;
};

export async function buildStartupStatusSnapshot(input: StartupStatusInput) {
    const {
        mcpServer,
        aggregator,
        agentMemory,
        browserService,
        browserStatus,
        sessionSupervisor,
        sessionCount,
        mcpConfigService,
        liveServerCount,
        residentLiveServerCount = 0,
        warmingServerCount = 0,
        failedWarmupServerCount = 0,
        persistedServerCount,
        persistedToolCount,
        persistedAlwaysOnServerCount,
        persistedAlwaysOnToolCount,
        inventorySource,
        inventorySnapshotUpdatedAt,
        executionEnvironment,
        tormentnexus,
    } = input;
    const mcpServerRuntime = mcpServer as {
        memoryManager?: unknown;
        isMemoryInitialized?: boolean;
        getBridgeStatus?: () => {
            ready?: boolean;
            clientCount?: number;
            clients?: RegisteredBridgeClient[];
            supportedCapabilities?: string[];
            supportedHookPhases?: string[];
        };
    };

    const aggregatorStatus = aggregator?.getInitializationStatus?.();
    const configSyncStatus = mcpConfigService?.getStatus?.();
    const restoreStatus = sessionSupervisor?.getRestoreStatus?.();
    const bridgeStatus = mcpServerRuntime.getBridgeStatus?.() ?? {
        ready: false,
        clientCount: 0,
        clients: [],
        supportedCapabilities: [],
        supportedHookPhases: [],
    };

    const bridgeClientCount = Number(bridgeStatus.clientCount ?? 0);
    const bridgeReady = Boolean(bridgeStatus.ready);
    const hasConnectedBridgeClients = bridgeClientCount > 0;
    const memoryInitialized = Boolean(mcpServerRuntime.isMemoryInitialized);
    const memoryReady = Boolean(mcpServerRuntime.memoryManager && agentMemory && memoryInitialized);
    const sessionReady = Boolean(sessionSupervisor);
    const browserReady = Boolean(browserService);
    const mcpReady = Boolean(aggregator);
    const configSyncReady = Boolean(configSyncStatus?.lastCompletedAt) && !configSyncStatus?.inProgress && !configSyncStatus?.lastError;
    const sessionRestoreReady = Boolean(restoreStatus?.lastRestoreAt);
    const configuredServerCount = Number(
        configSyncStatus?.lastServerCount
        ?? aggregatorStatus?.configuredServerCount
        ?? persistedServerCount,
    );
    const knownServerCount = Math.max(liveServerCount, persistedServerCount, configuredServerCount);
    const advertisedServerCount = Math.max(persistedServerCount, configuredServerCount);
    const advertisedToolCount = Math.max(persistedToolCount, Number(configSyncStatus?.lastToolCount ?? 0));
    const advertisedAlwaysOnServerCount = persistedAlwaysOnServerCount;
    const inventoryKnown = Boolean(configSyncStatus?.lastCompletedAt || aggregatorStatus?.initialized || persistedServerCount > 0);
    const inventoryReady = inventoryKnown && (
        configuredServerCount === 0
        || persistedServerCount > 0
        || persistedToolCount > 0
        || liveServerCount > 0
    );
    const aggregatorReady = Boolean(aggregatorStatus?.initialized);
    const liveReady = mcpReady && aggregatorReady;
    const residentReady = liveReady && (input.lazyMode ? true : residentLiveServerCount >= advertisedAlwaysOnServerCount);
    const mcpWarmupInProgress = Boolean(
        aggregatorStatus?.inProgress
        || configSyncStatus?.inProgress
        || warmingServerCount > 0
        || (!input.lazyMode && inventoryReady && configuredServerCount > 0 && liveServerCount < configuredServerCount),
    );
    const executionReady = Boolean(executionEnvironment?.ready);
    const tormentnexusEnabled = Boolean(tormentnexus?.enabled);
    const tormentnexusStoreExists = Boolean(tormentnexus?.storeExists);
    const tormentnexusDefaultSectionsReady = tormentnexusEnabled
        ? Number(tormentnexus?.presentDefaultSectionCount ?? 0) >= Number(tormentnexus?.defaultSectionCount ?? 0)
        : true;
    const tormentnexusReady = !tormentnexusEnabled || (tormentnexusStoreExists && tormentnexusDefaultSectionsReady);

    const blockingReasons: StartupBlockingReason[] = [];

    if (!aggregatorReady) {
        blockingReasons.push({
            code: 'mcp_aggregator_not_initialized',
            detail: 'MCP aggregator has not finished initialization.',
        });
    }

    if (!inventoryReady) {
        blockingReasons.push({
            code: 'mcp_inventory_not_ready',
            detail: 'Cached MCP inventory has not been populated yet.',
        });
    }

    if (!residentReady) {
        const residentTarget = persistedAlwaysOnServerCount;
        const residentConnected = residentLiveServerCount;
        blockingReasons.push({
            code: 'mcp_resident_runtime_not_ready',
            detail: residentTarget > 0
                ? `Resident MCP runtime is still warming (${residentConnected}/${residentTarget} always-on servers connected).`
                : 'MCP runtime connectivity is not fully ready yet.',
        });
    }

    if (!configSyncReady) {
        blockingReasons.push({
            code: 'mcp_config_sync_pending',
            detail: 'MCP config sync has not completed successfully.',
        });
    }

    if (!memoryReady) {
        blockingReasons.push({
            code: 'memory_not_ready',
            detail: memoryInitialized
                ? 'Memory manager is present but agent memory wiring is still pending.'
                : 'Memory manager initialization is still pending.',
        });
    }

    if (!tormentnexusReady) {
        blockingReasons.push({
            code: 'tormentnexus_not_ready',
            detail: !tormentnexusStoreExists
                ? 'tormentnexus store has not been created yet.'
                : 'tormentnexus default sections are still being seeded.',
        });
    }

    if (!browserReady) {
        blockingReasons.push({
            code: 'browser_service_not_ready',
            detail: 'Browser control bridge is not ready.',
        });
    }

    if (!sessionRestoreReady) {
        blockingReasons.push({
            code: 'session_restore_not_ready',
            detail: 'Session supervisor restore has not completed yet.',
        });
    }

    if (!bridgeReady) {
        blockingReasons.push({
            code: 'extension_bridge_not_ready',
            detail: 'Extension bridge listener is offline.',
        });
    }

    if (!executionReady) {
        blockingReasons.push({
            code: 'execution_environment_not_ready',
            detail: 'Execution environment detection has not verified a preferred shell/toolchain yet.',
        });
    }

    const summary = blockingReasons.length === 0
        ? 'All startup checks passed.'
        : `Startup pending: ${blockingReasons.map((reason) => reason.detail).join(' ')}`;

    const version = await getTormentNexusVersion();

    return {
        status: 'running',
        ready: residentReady
            && memoryReady
            && tormentnexusReady
            && browserReady
            && sessionReady
            && bridgeReady
            && configSyncReady
            && sessionRestoreReady
            && inventoryReady
            && aggregatorReady,
        summary,
        blockingReasons,
        uptime: process.uptime(),
        runtime: {
            nodeEnv: process.env.NODE_ENV ?? null,
            platform: process.platform,
            version,
        },
        checks: {
            mcpAggregator: {
                ready: liveReady,
                liveReady,
                residentReady,
                serverCount: knownServerCount,
                connectedCount: liveServerCount,
                residentConnectedCount: residentLiveServerCount,
                initialization: aggregatorStatus ?? null,
                persistedServerCount,
                persistedToolCount,
                advertisedServerCount,
                advertisedToolCount,
                advertisedAlwaysOnServerCount,
                advertisedAlwaysOnToolCount: persistedAlwaysOnToolCount,
                configuredServerCount,
                warmingServerCount,
                failedWarmupServerCount,
                inventoryReady,
                inventorySource: inventorySource ?? (inventoryReady ? 'database' : 'empty'),
                inventorySnapshotUpdatedAt: inventorySnapshotUpdatedAt ?? null,
                warmupInProgress: mcpWarmupInProgress,
            },
            configSync: {
                ready: configSyncReady,
                status: configSyncStatus ?? null,
            },
            memory: {
                ready: memoryReady,
                initialized: memoryInitialized,
                agentMemory: Boolean(agentMemory),
                tormentnexus: {
                    ready: tormentnexusReady,
                    enabled: tormentnexusEnabled,
                    storeExists: tormentnexusStoreExists,
                    storePath: tormentnexus?.storePath ?? null,
                    totalEntries: Number(tormentnexus?.totalEntries ?? 0),
                    sectionCount: Number(tormentnexus?.sectionCount ?? 0),
                    defaultSectionCount: Number(tormentnexus?.defaultSectionCount ?? 0),
                    presentDefaultSectionCount: Number(tormentnexus?.presentDefaultSectionCount ?? 0),
                    missingSections: tormentnexus?.missingSections ?? [],
                    lastUpdatedAt: tormentnexus?.lastUpdatedAt ?? null,
                },
            },
            tormentnexus: {
                ready: tormentnexusReady,
                enabled: tormentnexusEnabled,
                storeExists: tormentnexusStoreExists,
                storePath: tormentnexus?.storePath ?? null,
                totalEntries: Number(tormentnexus?.totalEntries ?? 0),
                sectionCount: Number(tormentnexus?.sectionCount ?? 0),
                defaultSectionCount: Number(tormentnexus?.defaultSectionCount ?? 0),
                presentDefaultSectionCount: Number(tormentnexus?.presentDefaultSectionCount ?? 0),
                missingSections: tormentnexus?.missingSections ?? [],
                lastUpdatedAt: tormentnexus?.lastUpdatedAt ?? null,
            },
            browser: {
                ready: browserReady,
                active: Boolean(browserStatus.active),
                pageCount: Number(browserStatus.pageCount ?? 0),
            },
            sessionSupervisor: {
                ready: sessionReady && sessionRestoreReady,
                sessionCount,
                restore: restoreStatus ?? null,
            },
            extensionBridge: {
                ready: bridgeReady,
                acceptingConnections: bridgeReady,
                clientCount: bridgeClientCount,
                hasConnectedClients: hasConnectedBridgeClients,
                clients: bridgeStatus.clients ?? [],
                supportedCapabilities: bridgeStatus.supportedCapabilities ?? [],
                supportedHookPhases: bridgeStatus.supportedHookPhases ?? [],
            },
            executionEnvironment: {
                ready: executionReady,
                preferredShellId: executionEnvironment?.preferredShellId ?? null,
                preferredShellLabel: executionEnvironment?.preferredShellLabel ?? null,
                shellCount: Number(executionEnvironment?.shellCount ?? 0),
                verifiedShellCount: Number(executionEnvironment?.verifiedShellCount ?? 0),
                toolCount: Number(executionEnvironment?.toolCount ?? 0),
                verifiedToolCount: Number(executionEnvironment?.verifiedToolCount ?? 0),
                harnessCount: Number(executionEnvironment?.harnessCount ?? 0),
                verifiedHarnessCount: Number(executionEnvironment?.verifiedHarnessCount ?? 0),
                supportsPowerShell: Boolean(executionEnvironment?.supportsPowerShell),
                supportsPosixShell: Boolean(executionEnvironment?.supportsPosixShell),
                notes: executionEnvironment?.notes ?? [],
            },
        },
    };
}