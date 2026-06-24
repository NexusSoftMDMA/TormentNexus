import { randomUUID } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { spawn } from 'node:child_process';

import {
    detectLocalExecutionEnvironment,
    type LocalExecutionEnvironment,
} from '../services/execution-environment.js';
import {
    buildExecutionPolicyEnv,
    selectSessionExecutionPolicy,
} from '../services/session-execution-policy.js';
import {
    type CreateSupervisedSessionInput,
    type PersistedSessionSupervisorState,
    type SchedulerLike,
    type SessionAttachInfo,
    type SessionSupervisorLogEntry,
    type SessionSupervisorOptions,
    type SpawnSupervisedProcess,
    type SupervisedProcessHandle,
    type SupervisedSessionHealth,
    type SupervisedSessionSnapshot,
    type WorktreeManagerLike,
} from './types.js';
import { DEFAULT_SUPERVISED_COMMANDS } from './cliHarnessCatalog.js';

interface SessionRuntime {
    process?: SupervisedProcessHandle;
    restartTimer?: unknown;
    manualStop: boolean;
    restartAfterStop: boolean;
    health: SupervisedSessionHealth;
}

const ACTIVE_SESSION_STATUSES = new Set([
    'created',
    'starting',
    'running',
    'restarting',
] as SupervisedSessionSnapshot['status'][]);

function defaultSpawnProcess(command: string, args: string[], options: { cwd: string; env: NodeJS.ProcessEnv }): SupervisedProcessHandle {
    return spawn(command, args, {
        cwd: options.cwd,
        env: options.env,
        stdio: ['ignore', 'pipe', 'pipe'],
    });
}

function defaultScheduler(): SchedulerLike {
    return {
        setTimeout: (callback, delayMs) => setTimeout(callback, delayMs),
        clearTimeout: (handle) => clearTimeout(handle as NodeJS.Timeout),
    };
}

export class SessionSupervisor {
    private readonly rootDir: string;
    private readonly persistencePath: string;
    private readonly maxPersistedSessions: number;
    private readonly maxLogEntries: number;
    private readonly autoResumeOnStart: boolean;
    private readonly restartDelayMs: number;
    private readonly backoffMultiplier: number;
    private readonly maxBackoffMs: number;
    private readonly defaultMaxRestartAttempts: number;
    private readonly now: () => number;
    private readonly spawnProcess: SpawnSupervisedProcess;
    private readonly scheduler: SchedulerLike;
    private readonly worktreeManager?: WorktreeManagerLike;
    private readonly detectExecutionEnvironment: () => Promise<LocalExecutionEnvironment>;

    private readonly sessions = new Map<string, SupervisedSessionSnapshot>();
    private readonly runtimes = new Map<string, SessionRuntime>();
    private restoreState: {
        lastRestoreAt?: number;
        restoredSessionCount: number;
        autoResumeCount: number;
    } = {
        restoredSessionCount: 0,
        autoResumeCount: 0,
    };

    constructor(options: SessionSupervisorOptions = {}) {
        this.rootDir = options.rootDir ?? process.cwd();
        this.persistencePath = options.persistencePath ?? path.join(this.rootDir, '.tormentnexus', 'session-supervisor.json');
        this.maxPersistedSessions = options.maxPersistedSessions ?? 100;
        this.maxLogEntries = options.maxLogEntries ?? 200;
        this.autoResumeOnStart = options.autoResumeOnStart ?? true;
        this.restartDelayMs = options.restartDelayMs ?? 1_000;
        this.backoffMultiplier = options.backoffMultiplier ?? 2;
        this.maxBackoffMs = options.maxBackoffMs ?? 10_000;
        this.defaultMaxRestartAttempts = options.maxRestartAttempts ?? 3;
        this.now = options.now ?? (() => Date.now());
        this.spawnProcess = options.spawnProcess ?? defaultSpawnProcess;
        this.scheduler = options.scheduler ?? defaultScheduler();
        this.worktreeManager = options.worktreeManager;
        this.detectExecutionEnvironment = options.detectExecutionEnvironment ?? (() => detectLocalExecutionEnvironment());

        this.restoreSessions();
    }

    public listSessions(): SupervisedSessionSnapshot[] {
        return [...this.sessions.values()]
            .map((session) => this.cloneSession(session))
            .sort((left, right) => left.createdAt - right.createdAt);
    }

    public getSession(id: string): SupervisedSessionSnapshot | undefined {
        const session = this.sessions.get(id);
        return session ? this.cloneSession(session) : undefined;
    }

    public async createSession(input: CreateSupervisedSessionInput): Promise<SupervisedSessionSnapshot> {
        const id = randomUUID();
        const requestedWorkingDirectory = path.resolve(input.workingDirectory);
        const usesWorktree = await this.shouldUseWorktree(requestedWorkingDirectory, input.isolateWorktree ?? true);
        const worktreePath = usesWorktree
            ? await this.createWorktree(id)
            : undefined;
        const resolvedCommand = this.resolveCommand(input.cliType, input.command);
        const resolvedArgs = input.args ?? this.resolveDefaultArgs(input.cliType);
        const requestedExecutionProfile = input.executionProfile ?? 'auto';
        let executionPolicy = null;
        try {
            executionPolicy = selectSessionExecutionPolicy(
                await this.detectExecutionEnvironment(),
                requestedExecutionProfile,
            );
        } catch (error) {
            console.warn('[SessionSupervisor] Failed to detect execution environment:', error);
        }
        const now = this.now();

        const session: SupervisedSessionSnapshot = {
            id,
            name: input.name?.trim() || `${input.cliType}-${id.slice(0, 8)}`,
            cliType: input.cliType,
            command: resolvedCommand,
            args: [...resolvedArgs],
            env: {
                ...(input.env ?? {}),
                ...buildExecutionPolicyEnv(executionPolicy),
            },
            executionProfile: requestedExecutionProfile,
            executionPolicy,
            requestedWorkingDirectory,
            workingDirectory: worktreePath ?? requestedWorkingDirectory,
            worktreePath,
            autoRestart: input.autoRestart ?? true,
            isolateWorktree: usesWorktree,
            status: 'created',
            createdAt: now,
            lastActivityAt: now,
            restartCount: 0,
            maxRestartAttempts: input.maxRestartAttempts ?? this.defaultMaxRestartAttempts,
            metadata: { ...(input.metadata ?? {}) },
            logs: [],
        };

        this.sessions.set(id, session);
        this.runtimes.set(id, this.createRuntimeState());
        this.appendLog(id, 'system', `Session created for ${session.cliType} in ${session.workingDirectory}`);
        if (executionPolicy?.shellLabel) {
            this.appendLog(
                id,
                'system',
                `Execution policy ${executionPolicy.requestedProfile} selected ${executionPolicy.shellLabel} (${executionPolicy.reason})`,
            );
        }
        this.persist();
        return this.cloneSession(session);
    }

    public async startSession(id: string): Promise<SupervisedSessionSnapshot> {
        const session = this.requireSession(id);
        const runtime = this.requireRuntime(id);

        if (runtime.process && (session.status === 'running' || session.status === 'starting')) {
            return this.cloneSession(session);
        }

        if (runtime.restartTimer) {
            this.scheduler.clearTimeout(runtime.restartTimer);
            runtime.restartTimer = undefined;
        }

        runtime.manualStop = false;
        runtime.restartAfterStop = false;
        runtime.health.nextRestartAt = undefined;
        session.status = 'starting';
        session.scheduledRestartAt = undefined;
        session.lastError = undefined;
        session.startedAt = this.now();
        session.lastActivityAt = this.now();
        this.appendLog(id, 'system', `Starting ${session.command} ${session.args.join(' ')}`.trim());

        const env: NodeJS.ProcessEnv = {
            ...process.env,
            ...session.env,
        };
        let child;
        try {
            child = this.spawnProcess(session.command, session.args, {
                cwd: session.workingDirectory,
                env,
            });
        } catch (spawnErr) {
            const err = spawnErr as any;
            session.status = 'error';
            session.lastError = err?.message || String(spawnErr);
            this.appendLog(id, 'system', `Spawn failed: ${session.lastError}`);
            runtime.health.status = 'degraded';
            runtime.health.errorMessage = session.lastError;
            this.persist();
            return this.cloneSession(session);
        }

        runtime.process = child;
        session.status = 'running';
        runtime.health.status = 'healthy';
        runtime.health.lastCheck = this.now();
        runtime.health.consecutiveFailures = 0;
        runtime.health.errorMessage = undefined;
        session.lastActivityAt = this.now();

        if (typeof child.pid === 'number') {
            this.appendLog(id, 'system', `Spawned process ${child.pid}`);
        }

        child.stdout?.on('data', (chunk) => {
            this.recordOutput(id, 'stdout', chunk);
        });
        child.stderr?.on('data', (chunk) => {
            this.recordOutput(id, 'stderr', chunk);
        });
        child.on('exit', (code, signal) => {
            void this.handleExit(id, code, signal);
        });

        this.persist();
        return this.cloneSession(session);
    }

    public async stopSession(id: string, options: { force?: boolean } = {}): Promise<SupervisedSessionSnapshot> {
        const session = this.requireSession(id);
        const runtime = this.requireRuntime(id);

        runtime.manualStop = true;
        runtime.restartAfterStop = false;

        if (runtime.restartTimer) {
            this.scheduler.clearTimeout(runtime.restartTimer);
            runtime.restartTimer = undefined;
        }

        if (!runtime.process) {
            session.status = 'stopped';
            session.stoppedAt = this.now();
            session.scheduledRestartAt = undefined;
            runtime.health.status = 'degraded';
            runtime.health.lastCheck = this.now();
            runtime.health.nextRestartAt = undefined;
            this.appendLog(id, 'system', 'Stop requested while no process was running.');
            this.persist();
            return this.cloneSession(session);
        }

        session.status = 'stopping';
        this.appendLog(id, 'system', `Stopping process${options.force ? ' forcefully' : ''}.`);
        runtime.process.kill(options.force ? 'SIGKILL' : 'SIGTERM');
        this.persist();
        return this.cloneSession(session);
    }

    public async restartSession(id: string): Promise<SupervisedSessionSnapshot> {
        const session = this.requireSession(id);
        const runtime = this.requireRuntime(id);

        if (!runtime.process) {
            session.status = 'restarting';
            this.appendLog(id, 'system', 'Restart requested while idle; starting session.');
            this.persist();
            return this.startSession(id);
        }

        runtime.manualStop = true;
        runtime.restartAfterStop = true;
        session.status = 'restarting';
        this.appendLog(id, 'system', 'Manual restart requested.');
        runtime.process.kill('SIGTERM');
        this.persist();
        return this.cloneSession(session);
    }

    public updateSessionMetadata(id: string, metadataPatch: Record<string, unknown>): SupervisedSessionSnapshot {
        const session = this.requireSession(id);
        session.metadata = {
            ...session.metadata,
            ...metadataPatch,
        };
        session.lastActivityAt = this.now();
        this.persist();
        return this.cloneSession(session);
    }

    public getSessionLogs(id: string, limit = this.maxLogEntries): SessionSupervisorLogEntry[] {
        const session = this.requireSession(id);
        return session.logs.slice(-limit).map((entry) => ({ ...entry }));
    }

    public getAttachInfo(id: string): SessionAttachInfo {
        const session = this.requireSession(id);
        const runtime = this.requireRuntime(id);
        
        const hasPid = typeof runtime.process?.pid === 'number';
        
        // Determine attach readiness based on status and process availability
        let attachReadiness: 'ready' | 'pending' | 'unavailable';
        let attachReadinessReason: 'running-with-pid' | 'starting' | 'restarting' | 'stopping' | 'stopped' | 'created' | 'no-pid' | 'error';
        
        if (session.status === 'running' && hasPid) {
            attachReadiness = 'ready';
            attachReadinessReason = 'running-with-pid';
        } else if (session.status === 'starting') {
            attachReadiness = 'pending';
            attachReadinessReason = 'starting';
        } else if (session.status === 'restarting') {
            attachReadiness = 'pending';
            attachReadinessReason = 'restarting';
        } else if (session.status === 'stopping') {
            attachReadiness = 'pending';
            attachReadinessReason = 'stopping';
        } else if (session.status === 'running' && !hasPid) {
            attachReadiness = 'unavailable';
            attachReadinessReason = 'no-pid';
        } else if (session.status === 'stopped') {
            attachReadiness = 'unavailable';
            attachReadinessReason = 'stopped';
        } else if (session.status === 'created') {
            attachReadiness = 'unavailable';
            attachReadinessReason = 'created';
        } else if (session.status === 'error') {
            attachReadiness = 'unavailable';
            attachReadinessReason = 'error';
        } else {
            // Defensive fallback
            attachReadiness = 'unavailable';
            attachReadinessReason = 'error';
        }
        
        return {
            id: session.id,
            pid: runtime.process?.pid,
            command: session.command,
            args: [...session.args],
            cwd: session.workingDirectory,
            status: session.status,
            attachable: session.status === 'running' && hasPid,
            attachReadiness,
            attachReadinessReason,
        };
    }

    public getSessionHealth(id: string): SupervisedSessionHealth {
        const runtime = this.requireRuntime(id);
        return { ...runtime.health };
    }

    public restoreSessions(): SupervisedSessionSnapshot[] {
        const persisted = this.readPersistedState();
        this.sessions.clear();
        this.runtimes.clear();
        let autoResumeCount = 0;

        for (const session of persisted.sessions.slice(-this.maxPersistedSessions)) {
            const normalized = this.normalizeRestoredSession(session);
            this.sessions.set(normalized.id, normalized);
            this.runtimes.set(normalized.id, {
                process: undefined,
                restartTimer: undefined,
                manualStop: false,
                restartAfterStop: false,
                health: {
                    status: normalized.status === 'error' ? 'crashed' : 'degraded',
                    lastCheck: this.now(),
                    consecutiveFailures: normalized.status === 'error' ? 1 : 0,
                    restartCount: normalized.restartCount,
                    lastExitCode: normalized.lastExitCode,
                    lastExitSignal: normalized.lastExitSignal,
                    errorMessage: normalized.lastError,
                },
            });
        }

        if (this.autoResumeOnStart) {
            for (const session of this.sessions.values()) {
                if (session.status === 'restarting') {
                    autoResumeCount += 1;
                    void this.startSession(session.id);
                }
            }
        }

        this.restoreState = {
            lastRestoreAt: this.now(),
            restoredSessionCount: this.sessions.size,
            autoResumeCount,
        };

        return this.listSessions();
    }

    public getRestoreStatus() {
        return { ...this.restoreState };
    }

    public async shutdown(): Promise<void> {
        for (const runtime of this.runtimes.values()) {
            if (runtime.restartTimer) {
                this.scheduler.clearTimeout(runtime.restartTimer);
                runtime.restartTimer = undefined;
            }
            if (runtime.process) {
                runtime.manualStop = true;
                runtime.process.kill('SIGTERM');
                runtime.process = undefined;
            }
        }
        this.persist();
    }

    private createRuntimeState(): SessionRuntime {
        return {
            process: undefined,
            restartTimer: undefined,
            manualStop: false,
            restartAfterStop: false,
            health: {
                status: 'degraded',
                lastCheck: this.now(),
                consecutiveFailures: 0,
                restartCount: 0,
            },
        };
    }

    private resolveCommand(cliType: string, explicitCommand?: string): string {
        if (explicitCommand?.trim()) {
            return explicitCommand.trim();
        }
        return DEFAULT_SUPERVISED_COMMANDS[cliType]?.command ?? cliType;
    }

    private resolveDefaultArgs(cliType: string): string[] {
        return [...(DEFAULT_SUPERVISED_COMMANDS[cliType]?.args ?? [])];
    }

    private async shouldUseWorktree(workingDirectory: string, requestedIsolation: boolean): Promise<boolean> {
        if (!requestedIsolation || !this.worktreeManager) {
            return false;
        }

        // Only allocate a worktree when another active session already occupies
        // the same working directory. The first session claims the directory directly;
        // subsequent parallel sessions get isolated worktrees to avoid clobbering.
        const resolvedDir = path.resolve(workingDirectory);
        const conflict = [...this.sessions.values()].some(
            (session) =>
                ACTIVE_SESSION_STATUSES.has(session.status) &&
                path.resolve(session.workingDirectory) === resolvedDir,
        );
        return conflict;
    }

    private async createWorktree(sessionId: string): Promise<string> {
        if (!this.worktreeManager) {
            throw new Error('Worktree manager unavailable.');
        }
        return this.worktreeManager.createTaskEnvironment(sessionId);
    }

    private async handleExit(id: string, code: number | null, signal: NodeJS.Signals | null) {
        const session = this.sessions.get(id);
        const runtime = this.runtimes.get(id);
        if (!session || !runtime) {
            return;
        }

        runtime.process = undefined;
        session.lastExitCode = code ?? undefined;
        session.lastExitSignal = signal ?? undefined;
        session.scheduledRestartAt = undefined;
        session.lastActivityAt = this.now();
        session.stoppedAt = this.now();
        runtime.health.lastCheck = this.now();
        runtime.health.nextRestartAt = undefined;
        runtime.health.lastExitCode = code ?? undefined;
        runtime.health.lastExitSignal = signal ?? undefined;

        if (runtime.restartAfterStop) {
            runtime.restartAfterStop = false;
            runtime.manualStop = false;
            session.status = 'restarting';
            this.appendLog(id, 'system', 'Restarting session after manual stop.');
            this.persist();
            await this.startSession(id);
            return;
        }

        if (runtime.manualStop || session.status === 'stopping') {
            runtime.manualStop = false;
            session.status = 'stopped';
            runtime.health.status = 'degraded';
            this.appendLog(id, 'system', `Process exited${code !== null ? ` with code ${code}` : ''}${signal ? ` (${signal})` : ''}.`);
            this.persist();
            return;
        }

        runtime.health.consecutiveFailures += 1;
        runtime.health.status = 'crashed';
        runtime.health.errorMessage = `Process exited unexpectedly${code !== null ? ` with code ${code}` : ''}${signal ? ` (${signal})` : ''}.`;
        session.lastError = runtime.health.errorMessage;
        this.appendLog(id, 'system', runtime.health.errorMessage);

        if (session.autoRestart && session.restartCount < session.maxRestartAttempts) {
            session.restartCount += 1;
            runtime.health.restartCount = session.restartCount;
            runtime.health.status = 'degraded';
            runtime.health.lastRestartAt = this.now();
            const delayMs = this.calculateRestartDelay(session.restartCount);
            session.status = 'restarting';
            session.scheduledRestartAt = this.now() + delayMs;
            runtime.health.nextRestartAt = session.scheduledRestartAt;
            this.appendLog(id, 'system', `Scheduling restart attempt ${session.restartCount} in ${delayMs}ms.`);
            runtime.restartTimer = this.scheduler.setTimeout(() => {
                runtime.restartTimer = undefined;
                void this.startSession(id);
            }, delayMs);
            this.persist();
            return;
        }

        session.status = 'error';
        this.persist();
    }

    private calculateRestartDelay(restartCount: number): number {
        const computed = this.restartDelayMs * (this.backoffMultiplier ** Math.max(0, restartCount - 1));
        return Math.min(computed, this.maxBackoffMs);
    }

    private recordOutput(id: string, stream: 'stdout' | 'stderr', chunk: Buffer | string) {
        const message = Buffer.isBuffer(chunk) ? chunk.toString('utf8') : chunk;
        const trimmed = message.replace(/\r/g, '');
        if (!trimmed) {
            return;
        }

        const session = this.sessions.get(id);
        const runtime = this.runtimes.get(id);
        if (!session || !runtime) {
            return;
        }

        session.lastActivityAt = this.now();
        runtime.health.lastCheck = this.now();
        runtime.health.status = 'healthy';
        this.appendLog(id, stream, trimmed);
        this.persist();
    }

    private appendLog(id: string, stream: 'stdout' | 'stderr' | 'system', message: string) {
        const session = this.sessions.get(id);
        if (!session) {
            return;
        }

        const entry: SessionSupervisorLogEntry = {
            timestamp: this.now(),
            stream,
            message,
        };
        session.logs = [...session.logs, entry].slice(-this.maxLogEntries);
        session.lastActivityAt = entry.timestamp;
    }

    private readPersistedState(): PersistedSessionSupervisorState {
        try {
            if (!fs.existsSync(this.persistencePath)) {
                return { sessions: [], savedAt: this.now() };
            }
            const raw = fs.readFileSync(this.persistencePath, 'utf8');
            const parsed = JSON.parse(raw) as PersistedSessionSupervisorState;
            return {
                sessions: Array.isArray(parsed.sessions) ? parsed.sessions : [],
                savedAt: typeof parsed.savedAt === 'number' ? parsed.savedAt : this.now(),
            };
        } catch (error) {
            console.warn('[SessionSupervisor] Failed to read persistence state:', error);
            return { sessions: [], savedAt: this.now() };
        }
    }

    private persist() {
        const state: PersistedSessionSupervisorState = {
            sessions: this.listSessions().slice(-this.maxPersistedSessions),
            savedAt: this.now(),
        };

        fs.mkdirSync(path.dirname(this.persistencePath), { recursive: true });
        fs.writeFileSync(this.persistencePath, JSON.stringify(state, null, 2));
    }

    private normalizeRestoredSession(session: SupervisedSessionSnapshot): SupervisedSessionSnapshot {
        const status = session.status === 'running' || session.status === 'starting' || session.status === 'stopping'
            ? (this.autoResumeOnStart ? 'restarting' : 'stopped')
            : session.status;

        return {
            ...session,
            args: [...(session.args ?? [])],
            env: { ...(session.env ?? {}) },
            executionProfile: session.executionProfile ?? 'auto',
            executionPolicy: session.executionPolicy ?? null,
            metadata: { ...(session.metadata ?? {}) },
            logs: [...(session.logs ?? [])].slice(-this.maxLogEntries),
            status,
            scheduledRestartAt: undefined,
            lastActivityAt: typeof session.lastActivityAt === 'number' ? session.lastActivityAt : this.now(),
            restartCount: typeof session.restartCount === 'number' ? session.restartCount : 0,
            maxRestartAttempts: typeof session.maxRestartAttempts === 'number'
                ? session.maxRestartAttempts
                : this.defaultMaxRestartAttempts,
        };
    }

    private requireSession(id: string): SupervisedSessionSnapshot {
        const session = this.sessions.get(id);
        if (!session) {
            throw new Error(`Unknown supervised session '${id}'.`);
        }
        return session;
    }

    private requireRuntime(id: string): SessionRuntime {
        const runtime = this.runtimes.get(id);
        if (!runtime) {
            throw new Error(`Missing runtime state for session '${id}'.`);
        }
        return runtime;
    }

    private cloneSession(session: SupervisedSessionSnapshot): SupervisedSessionSnapshot {
        return {
            ...session,
            args: [...session.args],
            env: { ...session.env },
            executionPolicy: session.executionPolicy ? { ...session.executionPolicy } : null,
            metadata: { ...session.metadata },
            logs: session.logs.map((entry) => ({ ...entry })),
        };
    }
}
