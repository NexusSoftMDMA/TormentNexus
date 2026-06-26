import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { resolveOrchestratorBase } from '../lib/tormentnexus-orchestrator.js';

/**
 * Session Export/Import Router
 *
 * Provides the ability to:
 * - Export sessions and memories to portable JSON/ZIP formats
 * - Import sessions from other TormentNexus instances or environments
 * - Auto-detect session formats from various CLI tools
 * - Cross-environment memory transfer
 */

const ExportFormatSchema = z.enum(['json', 'jsonl', 'zip']);

const ExportOptionsSchema = z.object({
    format: ExportFormatSchema.default('json'),
    includeMemories: z.boolean().default(true),
    includeLogs: z.boolean().default(true),
    includeMetadata: z.boolean().default(true),
    sessionIds: z.array(z.string()).optional(), // If empty, export all
    dateRange: z.object({
        from: z.number().optional(),
        to: z.number().optional(),
    }).optional(),
});

const ImportOptionsSchema = z.object({
    data: z.string(), // JSON string of exported data
    merge: z.boolean().default(true), // Merge with existing or replace
    dryRun: z.boolean().default(false), // Preview without applying
    sourceEnvironment: z.string().optional(),
});

export interface ExportedSession {
    id: string;
    name: string;
    cliType: string;
    status: string;
    createdAt: number;
    stoppedAt?: number;
    workingDirectory: string;
    metadata: Record<string, unknown>;
    logs: Array<{ timestamp: number; level: string; message: string }>;
    memories: Array<{
        key: string;
        value: string;
        category: string;
        createdAt: number;
    }>;
}

export interface ExportPackage {
    version: '1.0';
    exportedAt: number;
    tormentnexusVersion: string;
    environment: string;
    sessionCount: number;
    sessions: ExportedSession[];
    globalMemories: Array<{
        key: string;
        value: string;
        category: string;
        createdAt: number;
    }>;
}

export interface ImportReport {
    imported: number;
    skipped: number;
    merged: number;
    errors: string[];
    dryRun: boolean;
    details: Array<{
        sessionId: string;
        action: 'imported' | 'merged' | 'skipped' | 'error';
        reason?: string;
    }>;
}

type OrchestratorSessionRecord = {
    id?: string;
    currentTask?: string;
    status?: string;
    startTime?: number;
};

// Session format auto-detection
const SESSION_FORMAT_SIGNATURES: Record<string, { paths: string[]; type: string }> = {
    'claude-code': { paths: ['.claude', '.claude/sessions'], type: 'claude-code' },
    'cursor': { paths: ['.cursor', '.cursor/sessions'], type: 'cursor' },
    'opencode': { paths: ['.docs/ai-logs'], type: 'opencode' },
    'aider': { paths: ['.aider.chat.history.md', '.aider.tags.cache'], type: 'aider' },
    'windsurf': { paths: ['.windsurf', '.docs/ai-logs'], type: 'windsurf' },
    'tormentnexus': { paths: ['.tormentnexus', '.tormentnexus/sessions'], type: 'tormentnexus' },
    'continue': { paths: ['.continue', '.continue/sessions'], type: 'continue' },
    'copilot': { paths: ['.github/copilot'], type: 'copilot' },
};

function detectSessionFormat(data: unknown): string {
    if (!data || typeof data !== 'object') return 'unknown';

    const record = data as Record<string, unknown>;

    // Check for TormentNexus export format
    if (record.version === '1.0' && Array.isArray(record.sessions)) return 'tormentnexus-export';

    // Check for Claude Code format  
    if (record.type === 'conversation' && record.messages) return 'claude-code';

    // Check for Aider format
    if (typeof record === 'string' && (record as string).includes('# aider chat')) return 'aider';

    // Check for generic session array
    if (Array.isArray(record) && record.length > 0 && record[0]?.id) return 'generic-sessions';

    return 'unknown';
}

function parseImportData(rawData: string): { format: string; sessions: ExportedSession[] } {
    let parsed: unknown;
    try {
        parsed = JSON.parse(rawData);
    } catch {
        return { format: 'invalid', sessions: [] };
    }

    const format = detectSessionFormat(parsed);

    if (format === 'tormentnexus-export') {
        const pkg = parsed as ExportPackage;
        return { format, sessions: pkg.sessions };
    }

    if (format === 'generic-sessions') {
        const arr = parsed as Array<Record<string, unknown>>;
        return {
            format,
            sessions: arr.map(item => ({
                id: String(item.id || `imported_${Date.now()}`),
                name: String(item.name || 'Imported Session'),
                cliType: String(item.cliType || 'unknown'),
                status: 'stopped',
                createdAt: Number(item.createdAt || Date.now()),
                workingDirectory: String(item.workingDirectory || process.cwd()),
                metadata: (item.metadata as Record<string, unknown>) || {},
                logs: [],
                memories: [],
            })),
        };
    }

    return { format, sessions: [] };
}

export async function loadExportableOrchestratorSessions(
    fetchImpl: typeof fetch = fetch,
    orchestratorBase: string | null = resolveOrchestratorBase(),
    sessionIds?: string[],
): Promise<ExportedSession[]> {
    if (!orchestratorBase) {
        return [];
    }

    const res = await fetchImpl(`${orchestratorBase}/api/sessions`);
    if (!res.ok) {
        return [];
    }

    const data = await res.json();
    if (!Array.isArray(data)) {
        return [];
    }

    const normalizedSessionIds = sessionIds?.length
        ? new Set(sessionIds.map((id) => id.trim()).filter(Boolean))
        : null;

    return (data as OrchestratorSessionRecord[])
        .filter((record) => {
            if (!normalizedSessionIds) {
                return true;
            }

            return normalizedSessionIds.has(String(record.id ?? '').trim());
        })
        .map((record) => ({
        id: String(record.id ?? `exported_${Date.now()}`),
        name: String(record.currentTask || 'TormentNexus Session'),
        cliType: 'tormentnexus',
        status: String(record.status || 'unknown'),
        createdAt: typeof record.startTime === 'number' ? record.startTime : Date.now(),
        workingDirectory: process.cwd(),
        metadata: {},
        logs: [],
        memories: [],
    }));
}

export async function restoreSessionViaOrchestrator(
    session: ExportedSession,
    fetchImpl: typeof fetch = fetch,
    orchestratorBase: string | null = resolveOrchestratorBase(),
): Promise<{ ok: true } | { ok: false; reason: string }> {
    if (!orchestratorBase) {
        return { ok: false, reason: 'No TormentNexus Orchestrator base configured.' };
    }

    try {
        const res = await fetchImpl(`${orchestratorBase}/api/sessions/start`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                task: { description: `Restore of ${session.name}` },
                workingDirectory: session.workingDirectory || process.cwd()
            })
        });

        if (res.ok) {
            return { ok: true };
        }

        return { ok: false, reason: `API returned ${res.status}` };
    } catch (e: any) {
        return { ok: false, reason: e.message };
    }
}

// In-memory export store
const exportHistory: Array<{ id: string; format: string; sessionCount: number; exportedAt: number }> = [];

export const sessionExportRouter = t.router({
    /**
     * Export sessions to a portable format.
     */
    export: adminProcedure
        .input(ExportOptionsSchema)
        .mutation(async ({ input }) => {
            const pkg: ExportPackage = {
                version: '1.0',
                exportedAt: Date.now(),
                tormentnexusVersion: '0.90.5',
                environment: process.platform,
                sessionCount: 0,
                sessions: [],
                globalMemories: [],
            };

            try {
                pkg.sessions = await loadExportableOrchestratorSessions(fetch, undefined, input.sessionIds);
                pkg.sessionCount = pkg.sessions.length;
            } catch (e: any) {
                console.warn(`[SessionExport] TormentNexus Orchestrator unavailable (${e.message}). Exporting empty sessions list.`);
            }

            const exportId = `export_${Date.now()}`;
            exportHistory.push({
                id: exportId,
                format: input.format,
                sessionCount: pkg.sessionCount,
                exportedAt: pkg.exportedAt,
            });

            return { id: exportId, package: pkg, format: input.format };
        }),

    /**
     * Import sessions from another environment.
     */
    import: adminProcedure
        .input(ImportOptionsSchema)
        .mutation(async ({ input }) => {
            const { format, sessions } = parseImportData(input.data);

            const report: ImportReport = {
                imported: 0,
                skipped: 0,
                merged: 0,
                errors: [],
                dryRun: input.dryRun,
                details: [],
            };

            if (format === 'invalid') {
                report.errors.push('Invalid import data — could not parse JSON');
                return report;
            }

            for (const session of sessions) {
                if (input.dryRun) {
                    report.details.push({ sessionId: session.id, action: 'imported', reason: 'dry-run preview' });
                    report.imported++;
                } else {
                    const result = await restoreSessionViaOrchestrator(session);

                    if (result.ok) {
                        report.details.push({ sessionId: session.id, action: 'imported' });
                        report.imported++;
                    } else {
                        // @ts-expect-error -- type mismatch from inferred schema types
                        report.details.push({ sessionId: session.id, action: 'error', reason: result.reason });
                        // @ts-expect-error -- type mismatch from inferred schema types
                        report.errors.push(`Failed to import session ${session.id}: ${result.reason}`);
                    }
                }
            }

            return report;
        }),

    /**
     * Auto-detect session format from raw data.
     */
    detectFormat: publicProcedure
        .input(z.object({ data: z.string().max(100_000) }))
        .mutation(async ({ input }) => {
            try {
                const parsed = JSON.parse(input.data);
                return { format: detectSessionFormat(parsed), valid: true };
            } catch {
                return { format: 'invalid', valid: false };
            }
        }),

    /**
     * List known session format signatures for auto-detection.
     */
    knownFormats: publicProcedure.query(async () => {
        return Object.entries(SESSION_FORMAT_SIGNATURES).map(([id, sig]) => ({
            id,
            type: sig.type,
            paths: sig.paths,
        }));
    }),

    /**
     * Get export history.
     */
    history: publicProcedure.query(async () => {
        return exportHistory.slice(-50);
    }),
});
