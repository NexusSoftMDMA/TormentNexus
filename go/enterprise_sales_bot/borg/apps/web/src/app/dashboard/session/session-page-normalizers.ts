type ExecutionProfile = 'auto' | 'powershell' | 'posix' | 'compatibility';

type SessionStatus = 'created' | 'starting' | 'running' | 'stopping' | 'stopped' | 'restarting' | 'error';

const SESSION_STATUSES = new Set<SessionStatus>([
    'created',
    'starting',
    'running',
    'stopping',
    'stopped',
    'restarting',
    'error',
]);

const EXECUTION_PROFILES = new Set<ExecutionProfile>([
    'auto',
    'powershell',
    'posix',
    'compatibility',
]);

export interface NormalizedSessionState {
    activeGoal: string;
    lastObjective: string;
    isAutoDriveActive: boolean;
}

export interface NormalizedSessionLog {
    stream: 'stdout' | 'stderr' | 'system';
    timestamp: number;
    message: string;
}

export interface NormalizedSessionCatalogEntry {
    id: string;
    name: string;
    command?: string;
    args?: string[];
    homepage?: string;
    docsUrl?: string;
    installHint?: string;
    category?: 'cli' | 'cloud' | 'editor';
    sessionCapable: boolean;
    versionArgs?: string[];
    installed: boolean;
    resolvedPath?: string | null;
    version?: string | null;
    detectionError?: string | null;
}

export interface NormalizedSessionRow {
    id: string;
    name: string;
    cliType: string;
    workingDirectory: string;
    worktreePath?: string;
    executionProfile: ExecutionProfile;
    executionPolicy?: {
        requestedProfile?: 'auto' | 'powershell' | 'posix' | 'compatibility';
        effectiveProfile?: 'powershell' | 'posix' | 'compatibility' | 'fallback';
        shellId?: string | null;
        shellLabel?: string | null;
        shellFamily?: 'powershell' | 'cmd' | 'posix' | 'wsl' | null;
        shellPath?: string | null;
        supportsPowerShell?: boolean;
        supportsPosixShell?: boolean;
        reason?: string;
    } | null;
    autoRestart: boolean;
    status: SessionStatus;
    restartCount: number;
    maxRestartAttempts: number;
    scheduledRestartAt?: number;
    lastActivityAt: number;
    lastError?: string;
    /** Worktree isolation flag surfaced from the supervisor snapshot. */
    isolateWorktree: boolean;
    /** Exit code of the last process run, when available. */
    lastExitCode?: number;
    /** Exit signal of the last process run (e.g. 'SIGTERM'), when available. */
    lastExitSignal?: string;
    logs: NormalizedSessionLog[];
    metadata?: {
        memoryBootstrap?: {
            prompt?: string;
            summaryCount?: number;
            observationCount?: number;
        };
        memoryBootstrapGeneratedAt?: number;
    };
}

const asRecord = (value: unknown): Record<string, unknown> => (
    value && typeof value === 'object' ? (value as Record<string, unknown>) : {}
);

const asTrimmedString = (value: unknown, fallback: string): string => {
    if (typeof value !== 'string') return fallback;
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : fallback;
};

const asOptionalTrimmedString = (value: unknown): string | undefined => {
    if (typeof value !== 'string') return undefined;
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : undefined;
};

const asFiniteNumber = (value: unknown, fallback: number): number => (
    typeof value === 'number' && Number.isFinite(value) ? value : fallback
);

const asNonNegativeNumber = (value: unknown, fallback: number): number => {
    const parsed = asFiniteNumber(value, fallback);
    return parsed >= 0 ? parsed : fallback;
};

const asBoolean = (value: unknown, fallback: boolean): boolean => (
    typeof value === 'boolean' ? value : fallback
);

const normalizeLog = (payload: unknown, index: number): NormalizedSessionLog => {
    const log = asRecord(payload);
    const stream = log.stream === 'stderr' || log.stream === 'system' || log.stream === 'stdout'
        ? log.stream
        : 'system';

    return {
        stream,
        timestamp: asFiniteNumber(log.timestamp, Date.now() + index),
        message: asTrimmedString(log.message, '(empty log line)'),
    };
};

const normalizeSessionStatus = (value: unknown): SessionStatus => {
    return typeof value === 'string' && SESSION_STATUSES.has(value as SessionStatus)
        ? (value as SessionStatus)
        : 'created';
};

const normalizeExecutionProfile = (value: unknown): ExecutionProfile => {
    return typeof value === 'string' && EXECUTION_PROFILES.has(value as ExecutionProfile)
        ? (value as ExecutionProfile)
        : 'auto';
};

export const normalizeSessionState = (payload: unknown): NormalizedSessionState => {
    const state = asRecord(payload);
    return {
        activeGoal: asTrimmedString(state.activeGoal, ''),
        lastObjective: asTrimmedString(state.lastObjective, ''),
        isAutoDriveActive: asBoolean(state.isAutoDriveActive, false),
    };
};

export const normalizeSessionCatalog = (payload: unknown): NormalizedSessionCatalogEntry[] => {
    if (!Array.isArray(payload)) return [];

    return payload.map((raw, index) => {
        const row = asRecord(raw);
        const fallbackId = `unknown-harness-${index + 1}`;
        return {
            id: asTrimmedString(row.id, fallbackId),
            name: asTrimmedString(row.name, 'Unknown harness'),
            command: asOptionalTrimmedString(row.command),
            args: Array.isArray(row.args) ? row.args.filter((value): value is string => typeof value === 'string') : undefined,
            homepage: asOptionalTrimmedString(row.homepage),
            docsUrl: asOptionalTrimmedString(row.docsUrl),
            installHint: asOptionalTrimmedString(row.installHint),
            category: row.category === 'cli' || row.category === 'cloud' || row.category === 'editor' ? row.category : undefined,
            sessionCapable: asBoolean(row.sessionCapable, false),
            versionArgs: Array.isArray(row.versionArgs) ? row.versionArgs.filter((value): value is string => typeof value === 'string') : undefined,
            installed: asBoolean(row.installed, false),
            resolvedPath: typeof row.resolvedPath === 'string' ? row.resolvedPath : null,
            version: typeof row.version === 'string' ? row.version : null,
            detectionError: typeof row.detectionError === 'string' ? row.detectionError : null,
        };
    });
};

export const normalizeSessionList = (payload: unknown): NormalizedSessionRow[] => {
    if (!Array.isArray(payload)) return [];

    return payload.map((raw, index) => {
        const session = asRecord(raw);
        const fallbackId = `unknown-session-${index + 1}`;
        const logs = Array.isArray(session.logs)
            ? session.logs.map((log, logIndex) => normalizeLog(log, logIndex))
            : [];

        return {
            id: asTrimmedString(session.id, fallbackId),
            name: asTrimmedString(session.name, 'Unnamed session'),
            cliType: asTrimmedString(session.cliType, 'cli'),
            workingDirectory: asTrimmedString(session.workingDirectory, ''),
            worktreePath: asOptionalTrimmedString(session.worktreePath),
            executionProfile: normalizeExecutionProfile(session.executionProfile),
            executionPolicy: session.executionPolicy && typeof session.executionPolicy === 'object'
                ? (session.executionPolicy as NormalizedSessionRow['executionPolicy'])
                : null,
            autoRestart: asBoolean(session.autoRestart, true),
            status: normalizeSessionStatus(session.status),
            restartCount: asNonNegativeNumber(session.restartCount, 0),
            maxRestartAttempts: asNonNegativeNumber(session.maxRestartAttempts, 0),
            scheduledRestartAt: typeof session.scheduledRestartAt === 'number' && Number.isFinite(session.scheduledRestartAt)
                ? session.scheduledRestartAt
                : undefined,
            lastActivityAt: asFiniteNumber(session.lastActivityAt, Date.now()),
            lastError: asOptionalTrimmedString(session.lastError),
            isolateWorktree: asBoolean(session.isolateWorktree, false),
            lastExitCode: typeof session.lastExitCode === 'number' && Number.isFinite(session.lastExitCode)
                ? session.lastExitCode
                : undefined,
            lastExitSignal: asOptionalTrimmedString(session.lastExitSignal),
            logs,
            metadata: session.metadata && typeof session.metadata === 'object'
                ? (session.metadata as NormalizedSessionRow['metadata'])
                : undefined,
        };
    });
};
