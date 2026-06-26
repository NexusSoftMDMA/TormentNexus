import { z } from 'zod';
import { t, publicProcedure, getAgentMemoryService, getSessionImportService, getSessionManager, getSessionSupervisor, getShellService } from '../lib/trpc-core.js';
import { detectCliHarnesses } from '../services/cli-harness-detection.js';

const sessionExecutionProfileSchema = z.enum(['auto', 'powershell', 'posix', 'compatibility']);

const supervisedSessionStatusSchema = z.enum(['created', 'starting', 'running', 'stopping', 'stopped', 'restarting', 'error']);

const sessionSupervisorLogEntrySchema = z.object({
    timestamp: z.number(),
    stream: z.enum(['stdout', 'stderr', 'system']),
    message: z.string(),
});

const sessionShellExecutionResultSchema = z.object({
    command: z.string(),
    cwd: z.string(),
    shellFamily: z.enum(['powershell', 'cmd', 'posix', 'wsl', 'default']),
    shellPath: z.string().nullable(),
    stdout: z.string(),
    stderr: z.string(),
    output: z.string(),
    exitCode: z.number(),
    durationMs: z.number(),
    succeeded: z.boolean(),
});

const importedSessionMemorySchema = z.object({
    id: z.string(),
    importedSessionId: z.string(),
    kind: z.enum(['memory', 'instruction']),
    content: z.string(),
    tags: z.array(z.string()),
    source: z.enum(['llm', 'heuristic']),
    metadata: z.record(z.unknown()),
    createdAt: z.number(),
});

const importedSessionSchema = z.object({
    id: z.string(),
    sourceTool: z.string(),
    sourcePath: z.string(),
    externalSessionId: z.string().nullable(),
    title: z.string().nullable(),
    sessionFormat: z.string(),
    transcript: z.string(),
    excerpt: z.string().nullable(),
    workingDirectory: z.string().nullable(),
    transcriptHash: z.string(),
    normalizedSession: z.record(z.unknown()),
    metadata: z.record(z.unknown()),
    discoveredAt: z.number(),
    importedAt: z.number(),
    lastModifiedAt: z.number().nullable(),
    createdAt: z.number(),
    updatedAt: z.number(),
    parsedMemories: z.array(importedSessionMemorySchema),
});

const importedInstructionDocSchema = z.object({
    path: z.string(),
    updatedAt: z.number(),
    size: z.number(),
});

const sessionImportSummarySchema = z.object({
    discoveredCount: z.number(),
    importedCount: z.number(),
    skippedCount: z.number(),
    storedMemoryCount: z.number(),
    instructionDocPath: z.string().nullable(),
    tools: z.array(z.string()),
});

const importedMaintenanceStatsSchema = z.object({
    totalSessions: z.number(),
    inlineTranscriptCount: z.number(),
    archivedTranscriptCount: z.number(),
    missingRetentionSummaryCount: z.number(),
});

const supervisedSessionSnapshotSchema = z.object({
    id: z.string(),
    name: z.string(),
    cliType: z.string(),
    command: z.string(),
    args: z.array(z.string()),
    env: z.record(z.string()),
    executionProfile: sessionExecutionProfileSchema,
    executionPolicy: z.object({
        requestedProfile: sessionExecutionProfileSchema,
        effectiveProfile: z.enum(['powershell', 'posix', 'compatibility', 'fallback']),
        shellId: z.string().nullable(),
        shellLabel: z.string().nullable(),
        shellFamily: z.enum(['powershell', 'cmd', 'posix', 'wsl']).nullable(),
        shellPath: z.string().nullable(),
        supportsPowerShell: z.boolean(),
        supportsPosixShell: z.boolean(),
        reason: z.string(),
    }).nullable(),
    requestedWorkingDirectory: z.string(),
    workingDirectory: z.string(),
    worktreePath: z.string().optional(),
    autoRestart: z.boolean(),
    isolateWorktree: z.boolean(),
    status: supervisedSessionStatusSchema,
    createdAt: z.number(),
    startedAt: z.number().optional(),
    stoppedAt: z.number().optional(),
    scheduledRestartAt: z.number().optional(),
    lastActivityAt: z.number(),
    restartCount: z.number(),
    maxRestartAttempts: z.number(),
    lastExitCode: z.number().optional(),
    lastExitSignal: z.string().optional(),
    lastError: z.string().optional(),
    metadata: z.record(z.unknown()),
    logs: z.array(sessionSupervisorLogEntrySchema),
});

export const sessionRouter = t.router({
    catalog: publicProcedure.query(async () => {
        return detectCliHarnesses();
    }),

    list: publicProcedure.output(z.array(supervisedSessionSnapshotSchema)).query(() => {
        try { return getSessionSupervisor()?.listSessions() ?? []; } catch { return []; }
    }),

    get: publicProcedure.input(z.object({
        id: z.string(),
    })).output(supervisedSessionSnapshotSchema.nullable()).query(({ input }) => {
        return getSessionSupervisor().getSession(input.id) ?? null;
    }),

    importedList: publicProcedure.input(z.object({
        limit: z.number().int().min(1).max(200).optional(),
    }).optional().default({ limit: 50 })).output(z.array(importedSessionSchema)).query(({ input }) => {
        return getSessionImportService()?.listImportedSessions(input.limit) as z.infer<typeof importedSessionSchema>[] ?? [];
    }),

    importedGet: publicProcedure.input(z.object({
        id: z.string(),
    })).output(importedSessionSchema.nullable()).query(({ input }) => {
        return getSessionImportService()?.getImportedSession(input.id) as z.infer<typeof importedSessionSchema> | null ?? null;
    }),

    importedScan: publicProcedure.input(z.object({
        force: z.boolean().optional(),
    }).optional().default({ force: false })).output(sessionImportSummarySchema).mutation(async ({ input }) => {
        const service = getSessionImportService();
        if (!service) {
            return {
                discoveredCount: 0,
                importedCount: 0,
                skippedCount: 0,
                storedMemoryCount: 0,
                instructionDocPath: null,
                tools: [],
            };
        }

        return await service.scanAndImport({ force: input.force }) as z.infer<typeof sessionImportSummarySchema>;
    }),

    importedInstructionDocs: publicProcedure.output(z.array(importedInstructionDocSchema)).query(async () => {
        return await getSessionImportService()?.listInstructionDocs?.() as z.infer<typeof importedInstructionDocSchema>[] ?? [];
    }),

    importedMaintenanceStats: publicProcedure.output(importedMaintenanceStatsSchema).query(() => {
        const service = getSessionImportService();
        if (!service?.getImportedMaintenanceStats) {
            return {
                totalSessions: 0,
                inlineTranscriptCount: 0,
                archivedTranscriptCount: 0,
                missingRetentionSummaryCount: 0,
            };
        }

        return service.getImportedMaintenanceStats() as z.infer<typeof importedMaintenanceStatsSchema>;
    }),

    create: publicProcedure.input(z.object({
        name: z.string().min(1).optional(),
        cliType: z.string().min(1),
        workingDirectory: z.string().min(1),
        command: z.string().min(1).optional(),
        args: z.array(z.string()).optional(),
        env: z.record(z.string()).optional(),
        executionProfile: sessionExecutionProfileSchema.optional(),
        autoRestart: z.boolean().optional(),
        isolateWorktree: z.boolean().optional(),
        metadata: z.record(z.unknown()).optional(),
        maxRestartAttempts: z.number().int().min(0).max(20).optional(),
    })).mutation(async ({ input }) => {
        return getSessionSupervisor().createSession(input);
    }),

    start: publicProcedure.input(z.object({
        id: z.string(),
    })).mutation(async ({ input }) => {
        const supervisor = getSessionSupervisor();
        const startedSession = await supervisor.startSession(input.id);
        const sessionState = getSessionManager().getState();
        const agentMemoryService = getAgentMemoryService();

        if (!agentMemoryService?.getSessionBootstrap) {
            return startedSession;
        }

        const bootstrap = agentMemoryService.getSessionBootstrap({
            activeGoal: sessionState.activeGoal,
            lastObjective: sessionState.lastObjective,
        });

        return supervisor.updateSessionMetadata(input.id, {
            activeGoal: sessionState.activeGoal,
            lastObjective: sessionState.lastObjective,
            memoryBootstrap: bootstrap,
            memoryBootstrapGeneratedAt: Date.now(),
        });
    }),

    stop: publicProcedure.input(z.object({
        id: z.string(),
        force: z.boolean().optional(),
    })).mutation(async ({ input }) => {
        const supervisor = getSessionSupervisor();
        const snapshot = supervisor.getSession(input.id);
        const result = await supervisor.stopSession(input.id, { force: input.force });

        const agentMemoryService = getAgentMemoryService();
        if (snapshot && agentMemoryService?.captureSessionSummary) {
            await agentMemoryService.captureSessionSummary({
                sessionId: snapshot.id,
                name: snapshot.name,
                cliType: snapshot.cliType,
                workingDirectory: snapshot.workingDirectory,
                status: 'stopped',
                restartCount: snapshot.restartCount,
                startedAt: snapshot.startedAt,
                stoppedAt: Date.now(),
                lastActivityAt: snapshot.lastActivityAt,
                lastError: snapshot.lastError,
                lastExitCode: snapshot.lastExitCode,
                activeGoal: typeof snapshot.metadata?.activeGoal === 'string' ? snapshot.metadata.activeGoal : null,
                lastObjective: typeof snapshot.metadata?.lastObjective === 'string' ? snapshot.metadata.lastObjective : null,
                logTail: snapshot.logs.slice(-12).map((entry) => `[${entry.stream}] ${entry.message}`),
                metadata: {
                    source: 'session_router_stop',
                    cliType: snapshot.cliType,
                    autoRestart: snapshot.autoRestart,
                    isolateWorktree: snapshot.isolateWorktree,
                    requestedWorkingDirectory: snapshot.requestedWorkingDirectory,
                },
            });
        }

        return result;
    }),

    restart: publicProcedure.input(z.object({
        id: z.string(),
    })).mutation(async ({ input }) => {
        return getSessionSupervisor().restartSession(input.id);
    }),

    executeShell: publicProcedure.input(z.object({
        id: z.string(),
        command: z.string().min(1),
        timeoutMs: z.number().int().min(1_000).max(120_000).optional(),
    })).output(sessionShellExecutionResultSchema).mutation(async ({ input }) => {
        const supervisor = getSessionSupervisor();
        const session = supervisor.getSession(input.id);

        if (!session) {
            throw new Error(`Unknown supervised session '${input.id}'.`);
        }

        const shellService = getShellService();
        if (!shellService?.executeWithContext) {
            throw new Error('ShellService does not support contextual execution.');
        }

        return await shellService.executeWithContext(input.command, {
            cwd: session.workingDirectory,
            env: session.env,
            session: session.id,
            executionPolicy: session.executionPolicy,
            timeoutMs: input.timeoutMs,
        }) as {
            command: string;
            cwd: string;
            shellFamily: 'powershell' | 'cmd' | 'posix' | 'wsl' | 'default';
            shellPath: string | null;
            stdout: string;
            stderr: string;
            output: string;
            exitCode: number;
            durationMs: number;
            succeeded: boolean;
        };
    }),

    logs: publicProcedure.input(z.object({
        id: z.string(),
        limit: z.number().int().min(1).max(500).optional(),
    })).query(({ input }) => {
        return getSessionSupervisor().getSessionLogs(input.id, input.limit);
    }),

    attachInfo: publicProcedure.input(z.object({
        id: z.string(),
    })).query(({ input }) => {
        return getSessionSupervisor().getAttachInfo(input.id);
    }),

    health: publicProcedure.input(z.object({
        id: z.string(),
    })).query(({ input }) => {
        return getSessionSupervisor().getSessionHealth(input.id);
    }),

    getState: publicProcedure.query(() => {
        return getSessionManager().getState();
    }),

    updateState: publicProcedure.input(z.object({
        isAutoDriveActive: z.boolean().optional(),
        activeGoal: z.string().nullable().optional(),
        lastObjective: z.string().nullable().optional(),
    })).mutation(async ({ input }) => {
        const manager = getSessionManager();
        const previousState = manager.getState();
        manager.updateState(input);
        manager.save();

        const nextState = manager.getState();
        const agentMemoryService = getAgentMemoryService();

        if (agentMemoryService?.captureUserPrompt) {
            if (typeof input.activeGoal === 'string' && input.activeGoal.trim() && input.activeGoal !== previousState.activeGoal) {
                await agentMemoryService.captureUserPrompt({
                    content: input.activeGoal,
                    role: 'goal',
                    activeGoal: nextState.activeGoal,
                    lastObjective: nextState.lastObjective,
                    metadata: {
                        source: 'session_router_update_state',
                    },
                });
            }

            if (typeof input.lastObjective === 'string' && input.lastObjective.trim() && input.lastObjective !== previousState.lastObjective) {
                await agentMemoryService.captureUserPrompt({
                    content: input.lastObjective,
                    role: 'objective',
                    activeGoal: nextState.activeGoal,
                    lastObjective: nextState.lastObjective,
                    metadata: {
                        source: 'session_router_update_state',
                    },
                });
            }
        }

        return { success: true };
    }),

    clear: publicProcedure.mutation(() => {
        getSessionManager().clearSession();
        return { success: true };
    }),

    heartbeat: publicProcedure.mutation(() => {
        getSessionManager().touch();
        return { alive: true, timestamp: Date.now() };
    }),

    restore: publicProcedure.mutation(() => {
        return getSessionSupervisor().restoreSessions();
    })
});
