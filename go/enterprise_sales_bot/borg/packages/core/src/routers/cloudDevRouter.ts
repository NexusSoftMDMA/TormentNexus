/**
 * cloudDevRouter.ts – tRPC router for managing cloud development environments
 *
 * Provides a unified interface for managing cloud dev agents across multiple
/**
 * cloudDevRouter.ts – tRPC router for managing cloud development environments
 *
 * Provides a unified interface for managing cloud dev agents across multiple
 * providers: Jules (Google), Codex (OpenAI), GitHub Copilot Workspace, Devin,
 * and any future cloud coding agents.
 *
 * Features:
 * - Session lifecycle management (create, pause, resume, complete, cancel)
 * - Per-session chat history (full message log with role attribution)
 * - Per-session activity log (structured log entries)
 * - sendMessage: deliver a message to a single session; `force:true` allows
 *   delivery even when the session is completed or failed
 * - broadcastMessage: send a message to ALL sessions (or a filtered subset);
 *   `force:true` overrides status-based delivery gates
 * - autoAcceptPlan: flag that causes the server to auto-confirm any
 *   awaiting_approval → active transition when a "plan" log entry arrives
 * - acceptPlan: manually accept the current plan for a session in
 *   awaiting_approval state
 */

import { z } from 'zod';
import { t, publicProcedure } from '../lib/trpc-core.js';

// ---- Schema Definitions ----

const CloudDevProviderSchema = z.enum([
    'jules',
    'codex',
    'copilot-workspace',
    'devin',
    'custom',
]);

const CloudDevSessionStatusSchema = z.enum([
    'pending',
    'active',
    'paused',
    'completed',
    'failed',
    'awaiting_approval',
    'cancelled',
]);

// Roles for chat messages within a session
const ChatMessageRoleSchema = z.enum(['user', 'agent', 'system', 'plan']);

// Log severity levels
const LogLevelSchema = z.enum(['debug', 'info', 'warn', 'error']);

// ---- Internal Types (exported so the tRPC router inference can name them) ----

export interface ChatMessage {
    id: string;
    role: z.infer<typeof ChatMessageRoleSchema>;
    content: string;
    timestamp: string;
    forceSent?: boolean; // true when sent to a completed/failed session via force
}

export interface LogEntry {
    id: string;
    level: z.infer<typeof LogLevelSchema>;
    message: string;
    timestamp: string;
    meta?: Record<string, unknown>;
}

export interface CloudDevSession {
    id: string;
    provider: z.infer<typeof CloudDevProviderSchema>;
    projectName: string;
    task: string;
    status: z.infer<typeof CloudDevSessionStatusSchema>;
    createdAt: string;
    updatedAt: string;
    metadata?: Record<string, unknown>;
    messages: ChatMessage[];
    logs: LogEntry[];
    autoAcceptPlan: boolean;  // when true, awaiting_approval is auto-transitioned to active
}

// ---- In-memory session store (persisted via DB in production) ----

const sessions: CloudDevSession[] = [];

// ---- Helper ----

/**
 * Statuses that normally block message delivery.
 * `force:true` bypasses this check.
 */
const TERMINAL_STATUSES = new Set<CloudDevSession['status']>(['completed', 'failed', 'cancelled']);
const BROADCAST_SKIPPED_SAMPLE_LIMIT = 25;

function selectBroadcastTargets(options: {
    force: boolean;
    statusFilter?: CloudDevSession['status'][];
    sessionIdFilter?: string[];
}) {
    const { force, statusFilter, sessionIdFilter } = options;
    const sessionIdSet = sessionIdFilter && sessionIdFilter.length > 0
        ? new Set(sessionIdFilter)
        : null;

    return sessions.filter((session) => {
        if (sessionIdSet && !sessionIdSet.has(session.id)) {
            return false;
        }

        if (statusFilter && statusFilter.length > 0) {
            return statusFilter.includes(session.status);
        }
        // Default: target non-terminal sessions unless force is enabled.
        if (!force && TERMINAL_STATUSES.has(session.status)) return false;
        return true;
    });
}

function classifyBroadcastSkipReason(options: {
    force: boolean;
    statusFilter?: CloudDevSession['status'][];
    sessionIdFilter?: string[];
    session: CloudDevSession;
}) {
    const { force, statusFilter, sessionIdFilter, session } = options;

    if (sessionIdFilter && sessionIdFilter.length > 0 && !sessionIdFilter.includes(session.id)) {
        return 'session_filter_mismatch' as const;
    }

    if (statusFilter && statusFilter.length > 0 && !statusFilter.includes(session.status)) {
        return 'status_filter_mismatch' as const;
    }

    if (!force && TERMINAL_STATUSES.has(session.status)) {
        return 'terminal_requires_force' as const;
    }

    return 'other' as const;
}

function mkId(prefix: string) {
    return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function nowIso() {
    return new Date().toISOString();
}

/**
 * Append a structured log entry to the session and return it.
 */
function appendLog(
    session: CloudDevSession,
    level: LogEntry['level'],
    message: string,
    meta?: Record<string, unknown>
): LogEntry {
    const entry: LogEntry = { id: mkId('log'), level, message, timestamp: nowIso(), meta };
    session.logs.push(entry);
    // Keep at most 2000 log lines to avoid unbounded memory growth while
    // preserving substantially more operator-visible history.
    if (session.logs.length > 2000) session.logs.splice(0, session.logs.length - 2000);
    return entry;
}

/**
 * Append a chat message to the session and return it.
 */
function appendMessage(
    session: CloudDevSession,
    role: ChatMessage['role'],
    content: string,
    forceSent = false
): ChatMessage {
    const msg: ChatMessage = { id: mkId('msg'), role, content, timestamp: nowIso(), forceSent: forceSent || undefined };
    session.messages.push(msg);
    // Keep at most 1000 messages per session so long-running sessions don't
    // silently hide too much chat history.
    if (session.messages.length > 1000) session.messages.splice(0, session.messages.length - 1000);
    session.updatedAt = nowIso();
    return msg;
}

/**
 * If the session has autoAcceptPlan enabled and is in awaiting_approval,
 * automatically transition it to active and log the event.
 */
function maybeAutoAccept(session: CloudDevSession): void {
    if (session.autoAcceptPlan && session.status === 'awaiting_approval') {
        session.status = 'active';
        session.updatedAt = nowIso();
        appendLog(session, 'info', 'Plan auto-accepted (autoAcceptPlan=true)');
        appendMessage(session, 'system', '[Auto-accept] Plan approved automatically.');
    }
}

// ---- Provider Configuration ----

interface ProviderConfig {
    name: string;
    provider: z.infer<typeof CloudDevProviderSchema>;
    apiKeyEnvVar: string;
    baseUrl: string;
    enabled: boolean;
}

const providerConfigs: ProviderConfig[] = [
    {
        name: 'Jules (Google)',
        provider: 'jules',
        apiKeyEnvVar: 'JULES_API_KEY',
        baseUrl: 'https://jules.googleapis.com/v1alpha',
        enabled: true,
    },
    {
        name: 'Codex (OpenAI)',
        provider: 'codex',
        apiKeyEnvVar: 'OPENAI_API_KEY',
        baseUrl: 'https://api.openai.com/v1/codex',
        enabled: false,
    },
    {
        name: 'GitHub Copilot Workspace',
        provider: 'copilot-workspace',
        apiKeyEnvVar: 'GITHUB_TOKEN',
        baseUrl: 'https://api.github.com/copilot/workspace',
        enabled: false,
    },
    {
        name: 'Devin',
        provider: 'devin',
        apiKeyEnvVar: 'DEVIN_API_KEY',
        baseUrl: 'https://api.devin.ai/v1',
        enabled: false,
    },
];

// ---- Router ----

export const cloudDevRouter = t.router({
    /**
     * List all registered cloud dev providers and their status.
     */
    listProviders: publicProcedure.query(() => {
        return providerConfigs.map(p => ({
            ...p,
            hasApiKey: !!process.env[p.apiKeyEnvVar],
        }));
    }),

    /**
     * Create a new cloud dev session (starts in pending status).
     */
    createSession: publicProcedure
        .input(
            z.object({
                provider: CloudDevProviderSchema,
                projectName: z.string().min(1),
                task: z.string().min(1),
                metadata: z.record(z.any()).optional(),
                autoAcceptPlan: z.boolean().default(false),
            })
        )
        .mutation(({ input }) => {
            const session: CloudDevSession = {
                id: mkId('cds'),
                provider: input.provider,
                projectName: input.projectName,
                task: input.task,
                status: 'pending',
                createdAt: nowIso(),
                updatedAt: nowIso(),
                metadata: input.metadata,
                messages: [],
                logs: [],
                autoAcceptPlan: input.autoAcceptPlan,
            };
            appendLog(session, 'info', `Session created for task: ${input.task}`);
            sessions.push(session);
            return session;
        }),

    /**
     * List all cloud dev sessions, optionally filtered by provider or status.
     * Returns sessions without full message/log payloads for efficiency.
     * Use getSession to fetch the full record.
     */
    listSessions: publicProcedure
        .input(
            z
                .object({
                    provider: CloudDevProviderSchema.optional(),
                    status: CloudDevSessionStatusSchema.optional(),
                })
                .optional()
        )
        .query(({ input }) => {
            let result = sessions as CloudDevSession[];
            if (input?.provider) result = result.filter(s => s.provider === input.provider);
            if (input?.status) result = result.filter(s => s.status === input.status);
            // Sort newest-updated first
            return [...result]
                .sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime())
                .map(s => ({
                    id: s.id,
                    provider: s.provider,
                    projectName: s.projectName,
                    task: s.task,
                    status: s.status,
                    createdAt: s.createdAt,
                    updatedAt: s.updatedAt,
                    metadata: s.metadata,
                    autoAcceptPlan: s.autoAcceptPlan,
                    messageCount: s.messages.length,
                    logCount: s.logs.length,
                }));
        }),

    /**
     * Get a single session including its full message history and log entries.
     */
    getSession: publicProcedure
        .input(z.object({ sessionId: z.string() }))
        .query(({ input }) => {
            return sessions.find(s => s.id === input.sessionId) ?? null;
        }),

    /**
     * Update a session's status.
     * Triggers auto-accept if transitioning to awaiting_approval with autoAcceptPlan=true.
     */
    updateSessionStatus: publicProcedure
        .input(
            z.object({
                sessionId: z.string(),
                status: CloudDevSessionStatusSchema,
            })
        )
        .mutation(({ input }) => {
            const session = sessions.find(s => s.id === input.sessionId);
            if (!session) throw new Error(`Session ${input.sessionId} not found`);
            const prev = session.status;
            session.status = input.status;
            session.updatedAt = nowIso();
            appendLog(session, 'info', `Status changed: ${prev} → ${input.status}`);
            maybeAutoAccept(session);
            return session;
        }),

    /**
     * Delete a session permanently.
     */
    deleteSession: publicProcedure
        .input(z.object({ sessionId: z.string() }))
        .mutation(({ input }) => {
            const idx = sessions.findIndex(s => s.id === input.sessionId);
            if (idx === -1) throw new Error(`Session ${input.sessionId} not found`);
            sessions.splice(idx, 1);
            return { success: true };
        }),

    /**
     * Send a chat message to a single session.
     *
     * - `force: true` allows delivery even when the session is in a terminal
     *   state (completed, failed, cancelled). Without force, sending to a
     *   terminal session throws an error.
     * - The session's updatedAt is bumped on delivery.
     */
    sendMessage: publicProcedure
        .input(
            z.object({
                sessionId: z.string(),
                content: z.string().min(1),
                force: z.boolean().default(false),
            })
        )
        .mutation(({ input }) => {
            const session = sessions.find(s => s.id === input.sessionId);
            if (!session) throw new Error(`Session ${input.sessionId} not found`);

            if (!input.force && TERMINAL_STATUSES.has(session.status)) {
                throw new Error(
                    `Session ${input.sessionId} is ${session.status}. Use force:true to send anyway.`
                );
            }

            const msg = appendMessage(session, 'user', input.content, input.force && TERMINAL_STATUSES.has(session.status));
            appendLog(session, 'info', `User message sent${input.force ? ' (forced)' : ''}: ${input.content.slice(0, 80)}`);
            return { message: msg, session: { id: session.id, status: session.status } };
        }),

    /**
     * Broadcast a message to ALL sessions (or a filtered subset by status).
     *
     * - `force: true` delivers to terminal sessions (completed/failed/cancelled).
     * - `statusFilter` limits delivery to sessions whose status is in the list.
     * - `sessionIds` limits delivery to explicit session IDs.
     * - Returns a per-session delivery report.
     */
    broadcastMessage: publicProcedure
        .input(
            z.object({
                content: z.string().min(1),
                force: z.boolean().default(false),
                // If provided, only sessions with one of these statuses receive the message
                statusFilter: z.array(CloudDevSessionStatusSchema).optional(),
                // If provided, only these session IDs are eligible to receive the message
                sessionIds: z.array(z.string()).optional(),
            })
        )
        .mutation(({ input }) => {
            const targets = selectBroadcastTargets({
                force: input.force,
                statusFilter: input.statusFilter,
                sessionIdFilter: input.sessionIds,
            });
            const targetIds = new Set(targets.map((session) => session.id));
            const skippedSessions = sessions.filter((session) => !targetIds.has(session.id));
            const skippedByReason: Record<string, number> = {};

            skippedSessions.forEach((session) => {
                const reason = classifyBroadcastSkipReason({
                    force: input.force,
                    statusFilter: input.statusFilter,
                    sessionIdFilter: input.sessionIds,
                    session,
                });
                skippedByReason[reason] = (skippedByReason[reason] ?? 0) + 1;
            });

            const results = targets.map(session => {
                const msg = appendMessage(session, 'user', input.content, input.force && TERMINAL_STATUSES.has(session.status));
                appendLog(session, 'info', `Broadcast message received${input.force ? ' (forced)' : ''}`);
                return { sessionId: session.id, messageId: msg.id, status: session.status };
            });

            return {
                delivered: results.length,
                skipped: sessions.length - results.length,
                results,
                skippedByReason,
                skippedSessionIds: skippedSessions.map((session) => session.id),
                skippedSessions: skippedSessions.slice(0, BROADCAST_SKIPPED_SAMPLE_LIMIT).map((session) => ({
                    id: session.id,
                    provider: session.provider,
                    projectName: session.projectName,
                    status: session.status,
                    reason: classifyBroadcastSkipReason({
                        force: input.force,
                        statusFilter: input.statusFilter,
                        sessionIdFilter: input.sessionIds,
                        session,
                    }),
                })),
                skippedSessionsSampled: skippedSessions.length > BROADCAST_SKIPPED_SAMPLE_LIMIT,
            };
        }),

    /**
     * Dry-run preview for broadcast targeting.
     * Uses the exact same targeting rules as broadcastMessage.
     */
    previewBroadcastRecipients: publicProcedure
        .input(
            z.object({
                force: z.boolean().default(false),
                statusFilter: z.array(CloudDevSessionStatusSchema).optional(),
                sessionIds: z.array(z.string()).optional(),
            })
        )
        .query(({ input }) => {
            const targets = selectBroadcastTargets({
                force: input.force,
                statusFilter: input.statusFilter,
                sessionIdFilter: input.sessionIds,
            });
            const targetIds = new Set(targets.map((session) => session.id));
            const skippedSessions = sessions.filter((session) => !targetIds.has(session.id));
            const skippedByReason: Record<string, number> = {};

            skippedSessions.forEach((session) => {
                const reason = classifyBroadcastSkipReason({
                    force: input.force,
                    statusFilter: input.statusFilter,
                    sessionIdFilter: input.sessionIds,
                    session,
                });
                skippedByReason[reason] = (skippedByReason[reason] ?? 0) + 1;
            });

            const byStatus: Partial<Record<CloudDevSession['status'], number>> = {};
            targets.forEach((session) => {
                byStatus[session.status] = (byStatus[session.status] ?? 0) + 1;
            });

            return {
                totalSessions: sessions.length,
                targeted: targets.length,
                skipped: sessions.length - targets.length,
                byStatus,
                sessionIds: targets.map((session) => session.id),
                recipients: targets.map((session) => ({
                    id: session.id,
                    provider: session.provider,
                    projectName: session.projectName,
                    status: session.status,
                    updatedAt: session.updatedAt,
                })),
                skippedByReason,
                skippedSessionIds: skippedSessions.map((session) => session.id),
                skippedSessions: skippedSessions.slice(0, BROADCAST_SKIPPED_SAMPLE_LIMIT).map((session) => ({
                    id: session.id,
                    provider: session.provider,
                    projectName: session.projectName,
                    status: session.status,
                    reason: classifyBroadcastSkipReason({
                        force: input.force,
                        statusFilter: input.statusFilter,
                        sessionIdFilter: input.sessionIds,
                        session,
                    }),
                })),
                skippedSessionsSampled: skippedSessions.length > BROADCAST_SKIPPED_SAMPLE_LIMIT,
            };
        }),

    /**
     * Accept the current plan for a session that is awaiting_approval.
     * Transitions the session to active and appends a system message.
     */
    acceptPlan: publicProcedure
        .input(z.object({ sessionId: z.string() }))
        .mutation(({ input }) => {
            const session = sessions.find(s => s.id === input.sessionId);
            if (!session) throw new Error(`Session ${input.sessionId} not found`);
            if (session.status !== 'awaiting_approval') {
                throw new Error(
                    `Session ${input.sessionId} is not awaiting approval (current: ${session.status})`
                );
            }
            session.status = 'active';
            session.updatedAt = nowIso();
            appendLog(session, 'info', 'Plan accepted manually');
            appendMessage(session, 'system', '[Plan accepted] Session resumed.');
            return session;
        }),

    /**
     * Toggle autoAcceptPlan for a session.
     * When enabled, any future transition to awaiting_approval is immediately
     * resolved to active without user intervention.
     */
    setAutoAcceptPlan: publicProcedure
        .input(z.object({ sessionId: z.string(), enabled: z.boolean() }))
        .mutation(({ input }) => {
            const session = sessions.find(s => s.id === input.sessionId);
            if (!session) throw new Error(`Session ${input.sessionId} not found`);
            session.autoAcceptPlan = input.enabled;
            session.updatedAt = nowIso();
            appendLog(session, 'info', `autoAcceptPlan set to ${input.enabled}`);
            // Immediately apply if already waiting
            maybeAutoAccept(session);
            return { id: session.id, autoAcceptPlan: session.autoAcceptPlan, status: session.status };
        }),

    /**
     * Retrieve chat messages for a session, newest-first.
     * `limit` defaults to 50; max 200.
     */
    getMessages: publicProcedure
        .input(z.object({ sessionId: z.string(), limit: z.number().int().min(1).max(1000).default(100) }))
        .query(({ input }) => {
            const session = sessions.find(s => s.id === input.sessionId);
            if (!session) return [];
            return [...session.messages].reverse().slice(0, input.limit).reverse();
        }),

    /**
     * Retrieve log entries for a session, newest-first.
     * `limit` defaults to 100; max 500.
     */
    getLogs: publicProcedure
        .input(z.object({ sessionId: z.string(), limit: z.number().int().min(1).max(2000).default(200) }))
        .query(({ input }) => {
            const session = sessions.find(s => s.id === input.sessionId);
            if (!session) return [];
            return [...session.logs].reverse().slice(0, input.limit).reverse();
        }),

    /**
     * Aggregate stats across all providers.
     */
    stats: publicProcedure.query(() => {
        const byProvider: Record<string, number> = {};
        const byStatus: Record<string, number> = {};
        let totalMessages = 0;
        let totalLogs = 0;

        for (const s of sessions) {
            byProvider[s.provider] = (byProvider[s.provider] || 0) + 1;
            byStatus[s.status] = (byStatus[s.status] || 0) + 1;
            totalMessages += s.messages.length;
            totalLogs += s.logs.length;
        }

        return {
            totalSessions: sessions.length,
            byProvider,
            byStatus,
            totalMessages,
            totalLogs,
            providers: providerConfigs.length,
            enabledProviders: providerConfigs.filter(p => p.enabled).length,
        };
    }),
});
