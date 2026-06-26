import { describe, expect, it } from 'vitest';

import type { DashboardStartupStatus } from '../dashboard-home-view';
import { buildHealthStartupViewModel } from './health-startup-view-model';

const readyStartupSnapshot: DashboardStartupStatus = {
    status: 'running',
    ready: true,
    uptime: 120,
    checks: {
        mcpAggregator: {
            ready: true,
            liveReady: true,
            residentReady: true,
            serverCount: 2,
            connectedCount: 2,
            residentConnectedCount: 2,
            persistedServerCount: 2,
            persistedToolCount: 12,
            configuredServerCount: 2,
            advertisedServerCount: 2,
            advertisedToolCount: 12,
            advertisedAlwaysOnServerCount: 2,
            advertisedAlwaysOnToolCount: 4,
            inventoryReady: true,
            warmupInProgress: false,
            initialization: {
                inProgress: false,
                initialized: true,
                connectedClientCount: 1,
                configuredServerCount: 2,
            },
        },
        configSync: {
            ready: true,
            status: {
                inProgress: false,
                lastServerCount: 2,
                lastToolCount: 12,
            },
        },
        memory: {
            ready: true,
            initialized: true,
            agentMemory: true,
        },
        browser: {
            ready: true,
            active: true,
            pageCount: 1,
        },
        sessionSupervisor: {
            ready: true,
            sessionCount: 1,
            restore: {
                restoredSessionCount: 1,
                autoResumeCount: 1,
            },
        },
        extensionBridge: {
            ready: true,
            acceptingConnections: true,
            clientCount: 1,
            hasConnectedClients: true,
        },
        executionEnvironment: {
            ready: true,
            preferredShellId: 'pwsh',
            preferredShellLabel: 'PowerShell',
            shellCount: 1,
            verifiedShellCount: 1,
            toolCount: 5,
            verifiedToolCount: 5,
            harnessCount: 2,
            verifiedHarnessCount: 2,
            supportsPowerShell: true,
            supportsPosixShell: false,
            notes: [],
        },
    },
};

describe('health startup view model', () => {
    it('counts extension install artifacts in startup readiness', () => {
        const model = buildHealthStartupViewModel(readyStartupSnapshot, true, [
            { id: 'browser-extension-chromium', status: 'ready' },
            { id: 'browser-extension-firefox', status: 'ready' },
        ]);

        expect(model.statusCards.startupReadiness).toEqual({
            status: 'Ready',
            detail: '7/7 checks ready',
        });

        expect(model.startupChecks).toContainEqual({
            name: 'Extension Install Artifacts',
            status: 'Operational',
            latency: '2/2 ready',
            detail: 'Chromium/Edge and Firefox extension bundles are ready to load.',
        });
    });

    it('keeps extension artifacts in detecting state until install telemetry arrives', () => {
        const model = buildHealthStartupViewModel(readyStartupSnapshot, true, undefined);

        expect(model.statusCards.extensionArtifacts).toEqual({
            status: 'Connecting',
            detail: 'Detecting Chromium and Firefox extension install artifacts from the workspace.',
        });
    });

    it('handles malformed non-array install artifacts payloads without crashing', () => {
        const model = buildHealthStartupViewModel(
            readyStartupSnapshot,
            true,
            { id: 'browser-extension-chromium', status: 'ready' } as any,
        );

        expect(model.statusCards.extensionArtifacts).toEqual({
            status: 'Connecting',
            detail: 'Detecting Chromium and Firefox extension install artifacts from the workspace.',
        });

        expect(model.statusCards.startupReadiness).toEqual({
            status: 'Warming',
            detail: '6/7 checks ready',
        });
    });
});
