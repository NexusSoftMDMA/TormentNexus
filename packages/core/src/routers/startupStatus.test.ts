import { describe, expect, it } from 'vitest';

import { buildStartupStatusSnapshot } from './startupStatus.js';

describe('buildStartupStatusSnapshot', () => {
    it('treats an empty but fully initialized router inventory as ready', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: false,
                    initialized: true,
                    connectedClientCount: 0,
                    configuredServerCount: 0,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 0,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 0,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 0,
                    lastToolCount: 0,
                }),
            },
            liveServerCount: 0,
            persistedServerCount: 0,
            persistedToolCount: 0,
            persistedAlwaysOnServerCount: 0,
            persistedAlwaysOnToolCount: 0,
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 2,
                verifiedShellCount: 2,
                toolCount: 4,
                verifiedToolCount: 4,
                harnessCount: 1,
                verifiedHarnessCount: 1,
                supportsPowerShell: true,
                supportsPosixShell: false,
                notes: ['Prefer PowerShell 7.'],
            },
        });

        expect(snapshot.ready).toBe(true);
        expect(snapshot.checks.mcpAggregator.inventoryReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.liveReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.residentReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.advertisedToolCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.inventorySource).toBe('database');
        expect(snapshot.checks.extensionBridge.ready).toBe(true);
        expect(snapshot.checks.extensionBridge.hasConnectedClients).toBe(false);
        expect(snapshot.runtime).toEqual(expect.objectContaining({
            nodeEnv: process.env.NODE_ENV ?? null,
            platform: process.platform,
            version: expect.any(String),
        }));
        expect(snapshot.checks.memory.tormentnexus).toEqual(expect.objectContaining({
            ready: true,
            enabled: false,
        }));
    });

    it('keeps startup pending when the bridge listener is offline', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: false,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: false,
                    initialized: true,
                    connectedClientCount: 1,
                    configuredServerCount: 1,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 1,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 1,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 1,
                    lastToolCount: 4,
                }),
            },
            liveServerCount: 1,
            persistedServerCount: 1,
            persistedToolCount: 4,
            persistedAlwaysOnServerCount: 0,
            persistedAlwaysOnToolCount: 0,
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 1,
                verifiedShellCount: 1,
                toolCount: 3,
                verifiedToolCount: 3,
                harnessCount: 1,
                verifiedHarnessCount: 1,
                supportsPowerShell: true,
                supportsPosixShell: false,
                notes: ['Prefer PowerShell 7.'],
            },
        });

        expect(snapshot.ready).toBe(false);
        expect(snapshot.checks.extensionBridge.ready).toBe(false);
        expect(snapshot.checks.extensionBridge.acceptingConnections).toBe(false);
        expect(snapshot.summary).toContain('Extension bridge listener is offline.');
        expect(snapshot.blockingReasons).toEqual(
            expect.arrayContaining([
                expect.objectContaining({ code: 'extension_bridge_not_ready' }),
            ]),
        );
    });

    it('reports a clear startup summary with blocking reason codes when checks are pending', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: false,
                getBridgeStatus: () => ({
                    ready: false,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: true,
                    initialized: false,
                    connectedClientCount: 0,
                    configuredServerCount: 2,
                }),
            },
            agentMemory: {},
            browserService: null,
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    restoredSessionCount: 0,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 0,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: true,
                    lastServerCount: 2,
                    lastToolCount: 0,
                }),
            },
            liveServerCount: 0,
            persistedServerCount: 2,
            persistedToolCount: 0,
            persistedAlwaysOnServerCount: 1,
            persistedAlwaysOnToolCount: 0,
            executionEnvironment: {
                ready: false,
                preferredShellId: null,
                preferredShellLabel: null,
                shellCount: 1,
                verifiedShellCount: 0,
                toolCount: 0,
                verifiedToolCount: 0,
                harnessCount: 0,
                verifiedHarnessCount: 0,
                supportsPowerShell: false,
                supportsPosixShell: false,
                notes: ['No verified shell.'],
            },
        });

        expect(snapshot.ready).toBe(false);
        expect(snapshot.summary).toContain('Startup pending:');
        expect(snapshot.blockingReasons).toEqual(
            expect.arrayContaining([
                expect.objectContaining({ code: 'mcp_aggregator_not_initialized' }),
                expect.objectContaining({ code: 'mcp_config_sync_pending' }),
                expect.objectContaining({ code: 'memory_not_ready' }),
                expect.objectContaining({ code: 'extension_bridge_not_ready' }),
                expect.objectContaining({ code: 'execution_environment_not_ready' }),
            ]),
        );
    });

    it('preserves rich bridge client metadata in the startup snapshot', async () => {
        const connectedAt = 1_700_000_000_000;
        const lastSeenAt = connectedAt + 5_000;
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 1,
                    clients: [
                        {
                            clientId: 'client-1',
                            clientName: 'TormentNexus VS Code Bridge',
                            clientType: 'vscode-extension',
                            version: '1.2.3',
                            platform: 'VS Code 1.99',
                            capabilities: ['chat.inject'],
                            hookPhases: ['chat.submit'],
                            connectedAt,
                            lastSeenAt,
                        },
                    ],
                    supportedCapabilities: ['chat.inject'],
                    supportedHookPhases: ['chat.submit'],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: false,
                    initialized: true,
                    connectedClientCount: 1,
                    configuredServerCount: 1,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: connectedAt,
                    restoredSessionCount: 1,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 1,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: connectedAt,
                    lastSuccessAt: connectedAt,
                    lastServerCount: 1,
                    lastToolCount: 1,
                }),
            },
            liveServerCount: 1,
            persistedServerCount: 1,
            persistedToolCount: 1,
            persistedAlwaysOnServerCount: 0,
            persistedAlwaysOnToolCount: 0,
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 1,
                verifiedShellCount: 1,
                toolCount: 1,
                verifiedToolCount: 1,
                harnessCount: 0,
                verifiedHarnessCount: 0,
                supportsPowerShell: true,
                supportsPosixShell: false,
                notes: ['Prefer PowerShell 7.'],
            },
        });

        expect(snapshot.checks.extensionBridge.clients).toEqual([
            expect.objectContaining({
                clientId: 'client-1',
                version: '1.2.3',
                platform: 'VS Code 1.99',
                connectedAt,
                lastSeenAt,
            }),
        ]);
    });

    it('keeps startup pending until memory is actually initialized', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: false,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: false,
                    initialized: true,
                    connectedClientCount: 0,
                    configuredServerCount: 0,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 0,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 0,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 0,
                    lastToolCount: 0,
                }),
            },
            liveServerCount: 0,
            persistedServerCount: 0,
            persistedToolCount: 0,
            persistedAlwaysOnServerCount: 0,
            persistedAlwaysOnToolCount: 0,
            executionEnvironment: {
                ready: false,
                preferredShellId: null,
                preferredShellLabel: null,
                shellCount: 1,
                verifiedShellCount: 0,
                toolCount: 0,
                verifiedToolCount: 0,
                harnessCount: 0,
                verifiedHarnessCount: 0,
                supportsPowerShell: false,
                supportsPosixShell: false,
                notes: ['No verified shell.'],
            },
        });

        expect(snapshot.ready).toBe(false);
        expect(snapshot.checks.memory.ready).toBe(false);
        expect(snapshot.checks.memory.initialized).toBe(false);
    });

    it('reports known router inventory even when no live clients are connected yet', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: false,
                    initialized: true,
                    connectedClientCount: 0,
                    configuredServerCount: 7,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 0,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 0,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 7,
                    lastToolCount: 26,
                }),
            },
            liveServerCount: 0,
            persistedServerCount: 7,
            persistedToolCount: 26,
            persistedAlwaysOnServerCount: 2,
            persistedAlwaysOnToolCount: 5,
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 2,
                verifiedShellCount: 2,
                toolCount: 5,
                verifiedToolCount: 5,
                harnessCount: 2,
                verifiedHarnessCount: 2,
                supportsPowerShell: true,
                supportsPosixShell: true,
                notes: ['Prefer PowerShell 7.', 'Cygwin Bash is available for POSIX pipelines.'],
            },
        });

        expect(snapshot.ready).toBe(false);
        expect(snapshot.checks.mcpAggregator.serverCount).toBe(7);
        expect(snapshot.checks.mcpAggregator.connectedCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.residentConnectedCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.residentReady).toBe(false);
        expect(snapshot.checks.mcpAggregator.warmingServerCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.failedWarmupServerCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.persistedServerCount).toBe(7);
        expect(snapshot.checks.mcpAggregator.advertisedServerCount).toBe(7);
        expect(snapshot.checks.mcpAggregator.advertisedToolCount).toBe(26);
        expect(snapshot.checks.mcpAggregator.advertisedAlwaysOnToolCount).toBe(5);
        expect(snapshot.checks.mcpAggregator.inventorySource).toBe('database');
        expect(snapshot.checks.mcpAggregator.warmupInProgress).toBe(true);
        expect(snapshot.checks.executionEnvironment).toEqual(expect.objectContaining({
            ready: true,
            preferredShellId: 'pwsh',
            verifiedToolCount: 5,
            supportsPosixShell: true,
        }));
    });

    it('advertises always-on cached tools before the live runtime has connected every server', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: true,
                    initialized: true,
                    connectedClientCount: 1,
                    configuredServerCount: 3,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 1,
                    autoResumeCount: 1,
                }),
            },
            sessionCount: 1,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 3,
                    lastToolCount: 12,
                }),
            },
            liveServerCount: 1,
            persistedServerCount: 3,
            persistedToolCount: 12,
            persistedAlwaysOnServerCount: 1,
            persistedAlwaysOnToolCount: 4,
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 2,
                verifiedShellCount: 2,
                toolCount: 4,
                verifiedToolCount: 4,
                harnessCount: 1,
                verifiedHarnessCount: 1,
                supportsPowerShell: true,
                supportsPosixShell: false,
                notes: ['Prefer PowerShell 7.'],
            },
        });

        expect(snapshot.ready).toBe(false);
        expect(snapshot.checks.mcpAggregator.inventoryReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.liveReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.residentReady).toBe(false);
        expect(snapshot.checks.mcpAggregator.connectedCount).toBe(1);
        expect(snapshot.checks.mcpAggregator.residentConnectedCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.warmingServerCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.failedWarmupServerCount).toBe(0);
        expect(snapshot.checks.mcpAggregator.advertisedToolCount).toBe(12);
        expect(snapshot.checks.mcpAggregator.advertisedAlwaysOnServerCount).toBe(1);
        expect(snapshot.checks.mcpAggregator.advertisedAlwaysOnToolCount).toBe(4);
        expect(snapshot.checks.mcpAggregator.inventorySource).toBe('database');
        expect(snapshot.checks.mcpAggregator.warmupInProgress).toBe(true);
    });

    it('marks startup ready once all always-on servers are connected even if on-demand peers are still cold', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: true,
                    initialized: true,
                    connectedClientCount: 1,
                    configuredServerCount: 3,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 1,
                    autoResumeCount: 1,
                }),
            },
            sessionCount: 1,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 3,
                    lastToolCount: 12,
                }),
            },
            liveServerCount: 1,
            residentLiveServerCount: 1,
            persistedServerCount: 3,
            persistedToolCount: 12,
            persistedAlwaysOnServerCount: 1,
            persistedAlwaysOnToolCount: 4,
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 2,
                verifiedShellCount: 2,
                toolCount: 4,
                verifiedToolCount: 4,
                harnessCount: 1,
                verifiedHarnessCount: 1,
                supportsPowerShell: true,
                supportsPosixShell: false,
                notes: ['Prefer PowerShell 7.'],
            },
        });

        expect(snapshot.ready).toBe(true);
        expect(snapshot.checks.mcpAggregator.liveReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.residentReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.connectedCount).toBe(1);
        expect(snapshot.checks.mcpAggregator.residentConnectedCount).toBe(1);
    });

    it('surfaces last-known-good config metadata when that is the only available inventory source', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: true,
                    initialized: true,
                    connectedClientCount: 0,
                    configuredServerCount: 2,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 0,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 0,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 2,
                    lastToolCount: 8,
                }),
            },
            liveServerCount: 0,
            warmingServerCount: 2,
            failedWarmupServerCount: 1,
            persistedServerCount: 2,
            persistedToolCount: 8,
            persistedAlwaysOnServerCount: 1,
            persistedAlwaysOnToolCount: 3,
            inventorySource: 'config',
            inventorySnapshotUpdatedAt: '2026-03-12T17:00:00.000Z',
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 1,
                verifiedShellCount: 1,
                toolCount: 2,
                verifiedToolCount: 2,
                harnessCount: 1,
                verifiedHarnessCount: 1,
                supportsPowerShell: true,
                supportsPosixShell: false,
                notes: ['Prefer PowerShell 7.'],
            },
        });

        expect(snapshot.checks.mcpAggregator.inventorySource).toBe('config');
        expect(snapshot.checks.mcpAggregator.inventorySnapshotUpdatedAt).toBe('2026-03-12T17:00:00.000Z');
        expect(snapshot.checks.mcpAggregator.inventoryReady).toBe(true);
        expect(snapshot.checks.mcpAggregator.warmupInProgress).toBe(true);
        expect(snapshot.checks.mcpAggregator.warmingServerCount).toBe(2);
        expect(snapshot.checks.mcpAggregator.failedWarmupServerCount).toBe(1);
    });

    it('keeps startup pending when tormentnexus is enabled but its store has not been initialized yet', async () => {
        const snapshot = await buildStartupStatusSnapshot({
            mcpServer: {
                memoryManager: {},
                isMemoryInitialized: true,
                getBridgeStatus: () => ({
                    ready: true,
                    clientCount: 0,
                    clients: [],
                    supportedCapabilities: [],
                    supportedHookPhases: [],
                }),
            },
            aggregator: {
                getInitializationStatus: () => ({
                    inProgress: false,
                    initialized: true,
                    connectedClientCount: 0,
                    configuredServerCount: 0,
                }),
            },
            agentMemory: {},
            browserService: {},
            browserStatus: { active: false, pageCount: 0, pageIds: [] },
            sessionSupervisor: {
                getRestoreStatus: () => ({
                    lastRestoreAt: 1_700_000_000_000,
                    restoredSessionCount: 0,
                    autoResumeCount: 0,
                }),
            },
            sessionCount: 0,
            mcpConfigService: {
                getStatus: () => ({
                    inProgress: false,
                    lastCompletedAt: 1_700_000_000_000,
                    lastSuccessAt: 1_700_000_000_000,
                    lastServerCount: 0,
                    lastToolCount: 0,
                }),
            },
            liveServerCount: 0,
            persistedServerCount: 0,
            persistedToolCount: 0,
            persistedAlwaysOnServerCount: 0,
            persistedAlwaysOnToolCount: 0,
            executionEnvironment: {
                ready: true,
                preferredShellId: 'pwsh',
                preferredShellLabel: 'PowerShell 7',
                shellCount: 1,
                verifiedShellCount: 1,
                toolCount: 1,
                verifiedToolCount: 1,
                harnessCount: 0,
                verifiedHarnessCount: 0,
                supportsPowerShell: true,
                supportsPosixShell: false,
                notes: ['Prefer PowerShell 7.'],
            },
            tormentnexus: {
                enabled: true,
                storePath: '.tormentnexus/claude_mem.json',
                storeExists: false,
                totalEntries: 0,
                sectionCount: 0,
                defaultSectionCount: 7,
                presentDefaultSectionCount: 0,
                missingSections: ['project_overview'],
                lastUpdatedAt: null,
            },
        });

        expect(snapshot.ready).toBe(false);
        expect(snapshot.checks.memory.tormentnexus).toEqual(expect.objectContaining({
            enabled: true,
            storeExists: false,
            ready: false,
        }));
    });
});