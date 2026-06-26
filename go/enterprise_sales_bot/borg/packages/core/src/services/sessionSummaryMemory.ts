export interface SessionSummaryInput {
    sessionId: string;
    name?: string;
    cliType?: string;
    workingDirectory?: string;
    status?: string;
    restartCount?: number;
    startedAt?: number;
    stoppedAt?: number;
    lastActivityAt?: number;
    lastError?: string;
    lastExitCode?: number;
    activeGoal?: string | null;
    lastObjective?: string | null;
    logTail?: string[];
    metadata?: Record<string, unknown>;
}

export interface StructuredSessionSummary {
    sessionId: string;
    name?: string;
    cliType?: string;
    status: string;
    workingDirectory?: string;
    restartCount: number;
    startedAt?: number;
    stoppedAt?: number;
    lastActivityAt?: number;
    lastError?: string;
    lastExitCode?: number;
    activeGoal?: string | null;
    lastObjective?: string | null;
    logTail: string[];
    contentHash: string;
    recordedAt: number;
}

function sanitizeSentence(value: string, maxLength: number = 140): string {
    return value.replace(/\s+/g, ' ').trim().slice(0, maxLength);
}

function safeJsonStringify(value: unknown): string {
    return JSON.stringify(value, null, 2);
}

export function createSessionSummaryContentHash(summary: Omit<StructuredSessionSummary, 'contentHash'>): string {
    const payload = safeJsonStringify({
        sessionId: summary.sessionId,
        status: summary.status,
        stoppedAt: summary.stoppedAt ?? null,
        lastExitCode: summary.lastExitCode ?? null,
        lastError: summary.lastError ?? null,
        activeGoal: summary.activeGoal ?? null,
        lastObjective: summary.lastObjective ?? null,
        logTail: summary.logTail,
    });

    let hash = 0;
    for (let index = 0; index < payload.length; index += 1) {
        hash = ((hash << 5) - hash) + payload.charCodeAt(index);
        hash |= 0;
    }

    return `session-${Math.abs(hash)}`;
}

export function buildStructuredSessionSummary(input: SessionSummaryInput, recordedAt: number, logTail: string[]): StructuredSessionSummary {
    const baseSummary = {
        sessionId: input.sessionId,
        name: input.name ? sanitizeSentence(input.name, 120) : undefined,
        cliType: input.cliType ? sanitizeSentence(input.cliType, 80) : undefined,
        status: sanitizeSentence(input.status ?? 'stopped', 40),
        workingDirectory: input.workingDirectory,
        restartCount: input.restartCount ?? 0,
        startedAt: input.startedAt,
        stoppedAt: input.stoppedAt,
        lastActivityAt: input.lastActivityAt,
        lastError: input.lastError ? sanitizeSentence(input.lastError, 240) : undefined,
        lastExitCode: input.lastExitCode,
        activeGoal: typeof input.activeGoal === 'string' ? sanitizeSentence(input.activeGoal, 180) : input.activeGoal,
        lastObjective: typeof input.lastObjective === 'string' ? sanitizeSentence(input.lastObjective, 180) : input.lastObjective,
        logTail,
        recordedAt,
    };

    return {
        ...baseSummary,
        contentHash: createSessionSummaryContentHash(baseSummary),
    };
}

export function getStructuredSessionSummary(metadata: Record<string, unknown> | undefined): StructuredSessionSummary | null {
    const value = metadata?.structuredSessionSummary;
    if (!value || typeof value !== 'object') {
        return null;
    }

    const summary = value as Partial<StructuredSessionSummary>;
    if (typeof summary.sessionId !== 'string' || typeof summary.status !== 'string') {
        return null;
    }

    return {
        sessionId: summary.sessionId,
        name: typeof summary.name === 'string' ? summary.name : undefined,
        cliType: typeof summary.cliType === 'string' ? summary.cliType : undefined,
        status: summary.status,
        workingDirectory: typeof summary.workingDirectory === 'string' ? summary.workingDirectory : undefined,
        restartCount: typeof summary.restartCount === 'number' ? summary.restartCount : 0,
        startedAt: typeof summary.startedAt === 'number' ? summary.startedAt : undefined,
        stoppedAt: typeof summary.stoppedAt === 'number' ? summary.stoppedAt : undefined,
        lastActivityAt: typeof summary.lastActivityAt === 'number' ? summary.lastActivityAt : undefined,
        lastError: typeof summary.lastError === 'string' ? summary.lastError : undefined,
        lastExitCode: typeof summary.lastExitCode === 'number' ? summary.lastExitCode : undefined,
        activeGoal: typeof summary.activeGoal === 'string' || summary.activeGoal === null ? summary.activeGoal : undefined,
        lastObjective: typeof summary.lastObjective === 'string' || summary.lastObjective === null ? summary.lastObjective : undefined,
        logTail: Array.isArray(summary.logTail) ? summary.logTail.filter((item): item is string => typeof item === 'string') : [],
        contentHash: typeof summary.contentHash === 'string' ? summary.contentHash : '',
        recordedAt: typeof summary.recordedAt === 'number' ? summary.recordedAt : 0,
    };
}

export function buildSessionSummaryContent(summary: StructuredSessionSummary): string {
    const headline = `${summary.name ?? summary.sessionId} (${summary.cliType ?? 'unknown'}) ended with status ${summary.status}.`;
    const details = [
        summary.workingDirectory ? `Working directory: ${summary.workingDirectory}` : null,
        summary.activeGoal ? `Goal: ${summary.activeGoal}` : null,
        summary.lastObjective ? `Last objective: ${summary.lastObjective}` : null,
        summary.restartCount > 0 ? `Restart count: ${summary.restartCount}` : null,
        typeof summary.lastExitCode === 'number' ? `Exit code: ${summary.lastExitCode}` : null,
        summary.lastError ? `Last error: ${summary.lastError}` : null,
        summary.logTail.length > 0 ? `Log tail: ${summary.logTail.join(' | ')}` : null,
    ].filter((item): item is string => Boolean(item));

    return [headline, ...details].join('\n');
}
