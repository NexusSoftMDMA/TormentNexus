import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import { afterEach, describe, expect, it } from 'vitest';

import { SessionSupervisor } from './SessionSupervisor.js';
import type { SchedulerLike, SpawnSupervisedProcess, SupervisedProcessHandle } from './types.js';

class FakeProcess implements SupervisedProcessHandle {
    public pid = 4242;

    public stdout = {
        on: () => undefined,
    };

    public stderr = {
        on: () => undefined,
    };

    private exitListener?: (code: number | null, signal: NodeJS.Signals | null) => void;

    public on(event: 'exit', listener: (code: number | null, signal: NodeJS.Signals | null) => void) {
        if (event === 'exit') {
            this.exitListener = listener;
        }
        return undefined;
    }

    public kill() {
        return true;
    }

    public emitExit(code: number | null, signal: NodeJS.Signals | null = null) {
        this.exitListener?.(code, signal);
    }
}

describe('SessionSupervisor', () => {
    const tempDirs: string[] = [];

    afterEach(() => {
        while (tempDirs.length > 0) {
            const dir = tempDirs.pop();
            if (dir && fs.existsSync(dir)) {
                fs.rmSync(dir, { recursive: true, force: true });
            }
        }
    });

    it('tracks queued restart posture and last exit details after an unexpected crash', async () => {
        let now = 1_700_000_000_000;
        const createdProcesses: FakeProcess[] = [];
        const scheduled: Array<{ callback: () => void; delayMs: number }> = [];

        const spawnProcess: SpawnSupervisedProcess = () => {
            const child = new FakeProcess();
            createdProcesses.push(child);
            return child;
        };

        const scheduler: SchedulerLike = {
            setTimeout: (callback, delayMs) => {
                scheduled.push({ callback, delayMs });
                return callback;
            },
            clearTimeout: () => undefined,
        };

        const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'tormentnexus-session-supervisor-'));
        tempDirs.push(tempDir);

        const supervisor = new SessionSupervisor({
            rootDir: tempDir,
            persistencePath: path.join(tempDir, 'session-supervisor.json'),
            now: () => now,
            spawnProcess,
            scheduler,
            restartDelayMs: 1_000,
            backoffMultiplier: 2,
            maxBackoffMs: 10_000,
            detectExecutionEnvironment: async () => ({
                os: 'win32',
                summary: {
                    ready: true,
                    preferredShellId: 'pwsh',
                    preferredShellLabel: 'PowerShell 7',
                    shellCount: 1,
                    verifiedShellCount: 1,
                    toolCount: 0,
                    verifiedToolCount: 0,
                    harnessCount: 0,
                    verifiedHarnessCount: 0,
                    supportsPowerShell: true,
                    supportsPosixShell: false,
                    notes: [],
                },
                shells: [],
                tools: [],
                harnesses: [],
            }),
        });

        const session = await supervisor.createSession({
            cliType: 'node',
            workingDirectory: tempDir,
            autoRestart: true,
            maxRestartAttempts: 3,
        });

        await supervisor.startSession(session.id);
        createdProcesses[0]?.emitExit(17);

        const scheduledSession = supervisor.getSession(session.id);
        const scheduledHealth = supervisor.getSessionHealth(session.id);

        expect(scheduledSession).toEqual(expect.objectContaining({
            status: 'restarting',
            restartCount: 1,
            lastExitCode: 17,
            scheduledRestartAt: now + 1_000,
        }));
        expect(scheduledHealth).toEqual(expect.objectContaining({
            status: 'degraded',
            restartCount: 1,
            nextRestartAt: now + 1_000,
            lastExitCode: 17,
        }));
        expect(scheduled).toHaveLength(1);
        expect(scheduled[0]?.delayMs).toBe(1_000);

        now += 1_000;
        scheduled[0]?.callback();

        const restartedSession = supervisor.getSession(session.id);
        const restartedHealth = supervisor.getSessionHealth(session.id);

        expect(restartedSession).toEqual(expect.objectContaining({
            status: 'running',
            scheduledRestartAt: undefined,
        }));
        expect(restartedHealth).toEqual(expect.objectContaining({
            status: 'healthy',
            nextRestartAt: undefined,
        }));
    });

    it('captures execution shell policy and exports it into the supervised environment', async () => {
        const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'tormentnexus-session-supervisor-'));
        tempDirs.push(tempDir);

        const supervisor = new SessionSupervisor({
            rootDir: tempDir,
            persistencePath: path.join(tempDir, 'session-supervisor.json'),
            spawnProcess: () => new FakeProcess(),
            detectExecutionEnvironment: async () => ({
                os: 'win32',
                summary: {
                    ready: true,
                    preferredShellId: 'pwsh',
                    preferredShellLabel: 'PowerShell 7',
                    shellCount: 2,
                    verifiedShellCount: 2,
                    toolCount: 0,
                    verifiedToolCount: 0,
                    harnessCount: 0,
                    verifiedHarnessCount: 0,
                    supportsPowerShell: true,
                    supportsPosixShell: true,
                    notes: [],
                },
                shells: [
                    {
                        id: 'pwsh',
                        name: 'PowerShell 7',
                        family: 'powershell',
                        installed: true,
                        verified: true,
                        resolvedPath: 'C:\\Program Files\\PowerShell\\7\\pwsh.exe',
                        version: '7.5.0',
                        preferred: true,
                        notes: [],
                    },
                    {
                        id: 'cygwin-bash',
                        name: 'Cygwin Bash',
                        family: 'posix',
                        installed: true,
                        verified: true,
                        resolvedPath: 'C:\\cygwin64\\bin\\bash.exe',
                        version: '3.5.4',
                        preferred: false,
                        notes: [],
                    },
                ],
                tools: [],
                harnesses: [],
            }),
        });

        const session = await supervisor.createSession({
            cliType: 'node',
            workingDirectory: tempDir,
            executionProfile: 'posix',
        });

        expect(session.executionProfile).toBe('posix');
        expect(session.executionPolicy).toEqual(expect.objectContaining({
            requestedProfile: 'posix',
            effectiveProfile: 'posix',
            shellId: 'cygwin-bash',
            shellLabel: 'Cygwin Bash',
        }));
        expect(session.env).toEqual(expect.objectContaining({
            TORMENTNEXUS_EXECUTION_PROFILE_REQUESTED: 'posix',
            TORMENTNEXUS_EXECUTION_SHELL_ID: 'cygwin-bash',
            TORMENTNEXUS_SUPPORTS_POWERSHELL: '1',
            TORMENTNEXUS_SUPPORTS_POSIX_SHELL: '1',
        }));
    });

    it('persists metadata patches for a running session', async () => {
        const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'tormentnexus-session-supervisor-'));
        tempDirs.push(tempDir);

        const supervisor = new SessionSupervisor({
            rootDir: tempDir,
            persistencePath: path.join(tempDir, 'session-supervisor.json'),
            spawnProcess: () => new FakeProcess(),
            detectExecutionEnvironment: async () => ({
                os: 'win32',
                summary: {
                    ready: true,
                    preferredShellId: 'pwsh',
                    preferredShellLabel: 'PowerShell 7',
                    shellCount: 1,
                    verifiedShellCount: 1,
                    toolCount: 0,
                    verifiedToolCount: 0,
                    harnessCount: 0,
                    verifiedHarnessCount: 0,
                    supportsPowerShell: true,
                    supportsPosixShell: false,
                    notes: [],
                },
                shells: [],
                tools: [],
                harnesses: [],
            }),
        });

        const session = await supervisor.createSession({
            cliType: 'node',
            workingDirectory: tempDir,
        });

        const updated = supervisor.updateSessionMetadata(session.id, {
            memoryBootstrapGeneratedAt: 123,
            memoryBootstrap: {
                prompt: 'Memory bootstrap: recent work available',
            },
        });

        expect(updated.metadata).toEqual(expect.objectContaining({
            memoryBootstrapGeneratedAt: 123,
            memoryBootstrap: {
                prompt: 'Memory bootstrap: recent work available',
            },
        }));
    });
});
