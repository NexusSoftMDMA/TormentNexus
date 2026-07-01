import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';

import { describe, expect, it } from 'vitest';

import {
    chooseStaleCoreRefreshTarget,
    getTormentNexusStartLockPath,
    getPendingStartupChecks,
    getWaitingReasons,
    isCompatibleStartupStatusContract,
    isLikelyTormentNexusCoreCommand,
    isDirectExecution,
    isHttpProbeResponsive,
    parseListeningPidFromLsof,
    parseListeningPidFromNetstat,
    readTormentNexusStartLockRecord,
    resolveTormentNexusDataDir,
    summarizeBrowserExtensionArtifacts,
    waitForCoreBridgeShutdown,
} from './dev_tabby_ready_helpers.mjs';

const tempDirs: string[] = [];

function createTempDir() {
    const dir = mkdtempSync(path.join(tmpdir(), 'tormentnexus-dev-ready-'));
    tempDirs.push(dir);
    return dir;
}

describe('isDirectExecution', () => {
    it('matches the current script path when invoked directly', () => {
        const scriptPath = path.join(process.cwd(), 'scripts', 'dev_tabby_ready.mjs');
        const scriptUrl = pathToFileURL(scriptPath).href;

        expect(isDirectExecution(scriptUrl, scriptPath)).toBe(true);
    });

    it('returns false when argv1 points at a different script', () => {
        const scriptPath = path.join(process.cwd(), 'scripts', 'dev_tabby_ready.mjs');
        const scriptUrl = pathToFileURL(scriptPath).href;
        const otherPath = path.join(process.cwd(), 'scripts', 'verify_dev_readiness.mjs');

        expect(isDirectExecution(scriptUrl, otherPath)).toBe(false);
    });

    it('handles drive-letter case differences on Windows-style paths', () => {
        const scriptPath = path.join(process.cwd(), 'scripts', 'dev_tabby_ready.mjs');
        const scriptUrl = pathToFileURL(scriptPath).href;

        if (process.platform !== 'win32' || !/^[a-z]:/i.test(scriptPath)) {
            expect(isDirectExecution(scriptUrl, scriptPath)).toBe(true);
            return;
        }

        const flippedDrivePath = `${scriptPath[0] === scriptPath[0].toLowerCase() ? scriptPath[0].toUpperCase() : scriptPath[0].toLowerCase()}${scriptPath.slice(1)}`;
        expect(isDirectExecution(scriptUrl, flippedDrivePath)).toBe(true);
    });

    it('treats any HTTP response as proof that a probe target is alive', () => {
        expect(isHttpProbeResponsive({ ok: true, status: 200 })).toBe(true);
        expect(isHttpProbeResponsive({ ok: false, status: 500 })).toBe(true);
        expect(isHttpProbeResponsive({ ok: false, status: 404 })).toBe(true);
        expect(isHttpProbeResponsive({ ok: false, status: null })).toBe(false);
        expect(isHttpProbeResponsive(null)).toBe(false);
    });

    it('includes MCP warmup posture in pending startup checks', () => {
        expect(getPendingStartupChecks({
            checks: {
                mcpAggregator: {
                    ready: true,
                    liveReady: true,
                    residentReady: false,
                    inventoryReady: true,
                    advertisedAlwaysOnServerCount: 1,
                    residentConnectedCount: 0,
                    warmingServerCount: 2,
                    failedWarmupServerCount: 1,
                },
                configSync: { ready: true },
                sessionSupervisor: { ready: true },
                browser: { ready: true },
                memory: { ready: true },
                extensionBridge: { ready: true },
            },
        })).toContain('resident MCP runtime (2 warming, 1 failed)');
    });

    it('propagates MCP warmup posture into launcher waiting reasons', () => {
        expect(getWaitingReasons({
            web: { port: 3000 },
            coreBridge: { ok: true },
            startupStatus: {
                ok: true,
                compatible: true,
                data: {
                    ready: false,
                    checks: {
                        mcpAggregator: {
                            ready: true,
                            liveReady: true,
                            residentReady: false,
                            inventoryReady: true,
                            advertisedAlwaysOnServerCount: 1,
                            residentConnectedCount: 0,
                            warmingServerCount: 3,
                            failedWarmupServerCount: 0,
                        },
                        configSync: { ready: true },
                        sessionSupervisor: { ready: true },
                        browser: { ready: true },
                        memory: { ready: true },
                        extensionBridge: { ready: true },
                    },
                },
            },
            mcpStatus: { ok: true },
            memoryStatus: { ok: true },
            browserStatus: { ok: true },
            sessionStatus: { ok: true },
            extensions: [],
        })).toContain('resident MCP runtime (3 warming)');
    });

    it('rejects older startup payloads that do not expose the new readiness contract fields', () => {
        expect(isCompatibleStartupStatusContract({
            checks: {
                mcpAggregator: {
                    ready: true,
                    liveReady: true,
                },
                memory: {
                    ready: true,
                    initialized: true,
                },
                executionEnvironment: {
                    ready: true,
                },
            },
        })).toBe(false);
    });

    it('accepts startup payloads that expose the current readiness contract fields', () => {
        expect(isCompatibleStartupStatusContract({
            checks: {
                mcpAggregator: {
                    ready: true,
                    liveReady: true,
                    residentReady: true,
                    residentConnectedCount: 0,
                    inventorySource: 'database',
                },
                memory: {
                    ready: true,
                    initialized: true,
                    tormentnexus: {
                        ready: true,
                    },
                },
                executionEnvironment: {
                    ready: true,
                    harnessCount: 2,
                    verifiedHarnessCount: 2,
                },
            },
        })).toBe(true);
    });

    it('surfaces startup-contract drift as a launcher waiting reason', () => {
        expect(getWaitingReasons({
            web: { port: 3000 },
            coreBridge: { ok: true },
            startupStatus: {
                ok: true,
                compatible: false,
                data: {
                    ready: true,
                    checks: {
                        mcpAggregator: {
                            ready: true,
                            liveReady: true,
                        },
                        configSync: { ready: true },
                        sessionSupervisor: { ready: true },
                        browser: { ready: true },
                        memory: { ready: true },
                        extensionBridge: { ready: true },
                    },
                },
            },
            mcpStatus: { ok: true },
            memoryStatus: { ok: true },
            browserStatus: { ok: true },
            sessionStatus: { ok: true },
            extensions: [],
        })).toContain('core bridge startup contract refresh');
    });

    it('summarizes browser extension artifacts with ready counts and missing bundles', () => {
        expect(summarizeBrowserExtensionArtifacts([
            {
                id: 'browser-extension-chromium',
                label: 'browser extension Chromium bundle',
                artifactPath: 'apps/tormentnexus-extension/dist-chromium',
                ready: true,
                missingFiles: [],
                requiredFiles: ['background.js', 'manifest.json'],
            },
            {
                id: 'browser-extension-firefox',
                label: 'browser extension Firefox bundle',
                artifactPath: 'apps/tormentnexus-extension/dist-firefox',
                ready: false,
                missingFiles: ['background.js'],
                requiredFiles: ['background.js', 'manifest.json'],
            },
        ])).toMatchObject({
            ready: false,
            readyCount: 1,
            totalCount: 2,
            summary: '1/2 ready · missing Firefox bundle',
        });
    });

    it('expands the TormentNexus data dir shorthand and derives the lock path', () => {
        const resolved = resolveTormentNexusDataDir('~/.tormentnexus');

        expect(resolved.toLowerCase()).toContain(path.join('.tormentnexus').toLowerCase());
        expect(getTormentNexusStartLockPath('~/.tormentnexus').toLowerCase()).toContain(path.join('.tormentnexus', 'lock').toLowerCase());
    });

    it('reads a valid TormentNexus startup lock record from disk', () => {
        const dataDir = createTempDir();
        const lockPath = path.join(dataDir, 'lock');
        writeFileSync(lockPath, JSON.stringify({
            instanceId: 'tormentnexus-123',
            pid: 123,
            port: 4000,
            host: '127.0.0.1',
            createdAt: '2026-03-13T00:00:00.000Z',
        }), 'utf8');

        expect(readTormentNexusStartLockRecord(lockPath)).toEqual({
            instanceId: 'tormentnexus-123',
            pid: 123,
            port: 4000,
            host: '127.0.0.1',
            createdAt: '2026-03-13T00:00:00.000Z',
        });
    });

    it('waits for both core bridge probes to become unresponsive before reporting shutdown', async () => {
        const probeImpl = async () => {
            const attempt = probeCalls;
            probeCalls += 1;

            if (attempt < 2) {
                return { ok: true, status: 200 };
            }

            return { ok: false, status: null };
        };
        let probeCalls = 0;

        await expect(waitForCoreBridgeShutdown(
            ['http://127.0.0.1:4300/health', 'http://127.0.0.1:4300/api/sse'],
            {
                timeoutMs: 50,
                pollIntervalMs: 1,
            },
            {
                probeImpl,
                waitImpl: async () => undefined,
            },
        )).resolves.toBe(true);
        expect(probeCalls).toBeGreaterThanOrEqual(4);
    });

    it('parses a listening PID from Windows netstat output', () => {
        expect(parseListeningPidFromNetstat(`
  Proto  Local Address          Foreign Address        State           PID
  TCP    127.0.0.1:4300         0.0.0.0:0              LISTENING       4242
`, 4300)).toBe(4242);
    });

    it('parses a listening PID from lsof output', () => {
        expect(parseListeningPidFromLsof('4242\n')).toBe(4242);
    });

    it('treats TormentNexus CLI command lines as safe stale-core owners', () => {
        expect(isLikelyTormentNexusCoreCommand('node C:\\repo\\tormentnexus\\node_modules\\tsx\\dist\\cli.mjs src/index.ts start --port 3100')).toBe(true);
        expect(isLikelyTormentNexusCoreCommand('node /workspace/tormentnexus/packages/core/dist/server-stdio.js')).toBe(true);
    });

    it('rejects unrelated port owners for stale-core termination', () => {
        expect(isLikelyTormentNexusCoreCommand('node C:\\other-app\\server.js --port 4300')).toBe(false);
        expect(isLikelyTormentNexusCoreCommand('python -m http.server 4300')).toBe(false);
    });

    it('prefers the TormentNexus startup lock over port-owner fallback when both exist', () => {
        expect(chooseStaleCoreRefreshTarget({
            lockRecord: { pid: 111, instanceId: 'tormentnexus', port: 4300, host: '127.0.0.1', createdAt: '2026-03-13T00:00:00.000Z' },
            owner: { pid: 222, trusted: true, commandLine: 'node tormentnexus' },
            currentPid: 999,
        })).toEqual({
            kind: 'lock',
            pid: 111,
            sourceLabel: 'lock',
            trusted: true,
        });
    });

    it('uses a trusted port owner when no valid TormentNexus lock exists', () => {
        expect(chooseStaleCoreRefreshTarget({
            lockRecord: null,
            owner: { pid: 222, trusted: true, commandLine: 'node tormentnexus' },
            currentPid: 999,
        })).toEqual({
            kind: 'owner',
            pid: 222,
            sourceLabel: 'port 4300',
            trusted: true,
        });
    });

    it('skips automatic refresh when the port owner is not TormentNexus-owned', () => {
        expect(chooseStaleCoreRefreshTarget({
            lockRecord: null,
            owner: { pid: 333, trusted: false, commandLine: 'python -m http.server 4300' },
            currentPid: 999,
        })).toEqual({
            kind: 'skip-untrusted-owner',
            pid: 333,
            sourceLabel: 'port 4300',
            trusted: false,
        });
    });

    it('never selects the current process as a stale-core refresh target', () => {
        expect(chooseStaleCoreRefreshTarget({
            lockRecord: { pid: 999, instanceId: 'tormentnexus', port: 4300, host: '127.0.0.1', createdAt: '2026-03-13T00:00:00.000Z' },
            owner: { pid: 999, trusted: true, commandLine: 'node tormentnexus' },
            currentPid: 999,
        })).toEqual({
            kind: 'skip-untrusted-owner',
            pid: 999,
            sourceLabel: 'port 4300',
            trusted: false,
        });
    });
});

afterEach(() => {
    while (tempDirs.length > 0) {
        const dir = tempDirs.pop();
        if (dir) {
            rmSync(dir, { recursive: true, force: true });
        }
    }
});