import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { t, publicProcedure, getResearchService, getDeepResearchService } from '../lib/trpc-core.js';
import fs from 'node:fs/promises';
import path from 'node:path';

const MASTER_INDEX_PATH = path.join(process.cwd(), 'TORMENTNEXUS_MASTER_INDEX.jsonc');
const INGESTION_STATUS_PATH = path.join(process.cwd(), 'scripts', 'ingestion-status.json');

const isMissingFileError = (error: unknown): boolean =>
    typeof error === 'object'
    && error !== null
    && 'code' in error
    && (error as NodeJS.ErrnoException).code === 'ENOENT';

const stripJsonComments = (content: string) =>
    content.replace(/\\"|"(?:\\"|[^"])*"|(\/\/.*|\/\*[\s\S]*?\*\/)/g, (m, g) => g ? '' : m);

const normalizeUrl = (rawUrl: string): string => {
    try {
        const parsed = new URL(String(rawUrl).trim());
        parsed.hash = '';
        const cleanPath = parsed.pathname.replace(/\/+$/, '');
        parsed.pathname = cleanPath || '/';
        return parsed.toString().replace(/\/$/, '');
    } catch {
        return String(rawUrl).trim();
    }
};

type IngestionStatusDoc = {
    processed: string[];
    pending: string[];
    failed: Array<{
        url: string;
        error?: string;
        source?: string;
        attempts?: number;
        last_attempt_at?: string;
    }>;
};

const readIngestionStatus = async (): Promise<IngestionStatusDoc> => {
    try {
        const content = await fs.readFile(INGESTION_STATUS_PATH, 'utf-8');
        const parsed = JSON.parse(content) as Partial<IngestionStatusDoc>;
        return {
            processed: Array.isArray(parsed.processed) ? parsed.processed : [],
            pending: Array.isArray(parsed.pending) ? parsed.pending : [],
            failed: Array.isArray(parsed.failed) ? parsed.failed : []
        };
    } catch (error) {
        if (isMissingFileError(error)) {
            return { processed: [], pending: [], failed: [] };
        }
        throw error;
    }
};

const loadMasterIndexEntries = async (): Promise<Map<string, { id?: string; name?: string; category?: string }>> => {
    try {
        const content = await fs.readFile(MASTER_INDEX_PATH, 'utf-8');
        const parsed = JSON.parse(stripJsonComments(content)) as {
            categories?: Record<string, Array<{ url: string; id?: string; name?: string }>>;
        };

        const lookup = new Map<string, { id?: string; name?: string; category?: string }>();
        const categories = parsed.categories ?? {};
        for (const [category, items] of Object.entries(categories)) {
            if (!Array.isArray(items)) continue;
            for (const item of items) {
                if (!item?.url) continue;
                lookup.set(normalizeUrl(item.url), { id: item.id, name: item.name, category });
            }
        }

        return lookup;
    } catch (error) {
        if (isMissingFileError(error)) {
            return new Map();
        }
        throw error;
    }
};

const formatUnknownError = (error: unknown): string => error instanceof Error ? error.message : String(error);

const writeIngestionStatus = async (doc: IngestionStatusDoc): Promise<void> => {
    await fs.writeFile(INGESTION_STATUS_PATH, `${JSON.stringify(doc, null, 2)}\n`, 'utf-8');
};

export const researchRouter = t.router({
    /** Conduct deep research on a topic — multi-source search, read, memorize */
    conduct: publicProcedure
        .input(z.object({
            topic: z.string(),
            depth: z.number().min(1).max(10).default(3)
        }))
        .mutation(async ({ input }) => {
            const service = getResearchService();
            if (!service) throw new Error("ResearchService not found");
            const report = await service.research(input.topic, input.depth);
            return { report };
        }),

    /** Directly ingest a URL into memory for later retrieval */
    ingest: publicProcedure
        .input(z.object({
            url: z.string().url(),
        }))
        .mutation(async ({ input }) => {
            const service = getResearchService();
            if (!service) throw new Error("ResearchService not found");
            const result = await service.ingest(input.url);
            return { result };
        }),

    /** Recursive deep research — multi-step topic traversal with sub-topics */
    recursiveResearch: publicProcedure
        .input(z.object({
            topic: z.string(),
            depth: z.number().min(1).max(5).default(2),
            maxBreadth: z.number().min(1).max(10).default(3),
        }))
        .mutation(async ({ input }) => {
            // Use DeepResearchService if available (more advanced), fallback to ResearchService
            const deepService = getDeepResearchService();
            if (deepService?.recursiveResearch) {
                const result = await deepService.recursiveResearch(
                    input.topic, input.depth, input.maxBreadth
                );
                return { result };
            }

            // Fallback: basic research
            const service = getResearchService();
            if (!service) throw new Error("ResearchService not found");
            const report = await service.research(input.topic, input.depth);
            return { result: { topic: input.topic, summary: report, sources: [], relatedTopics: [] } };
        }),

    /** Generate research queries for a topic (planning step) */
    generateQueries: publicProcedure
        .input(z.object({
            topic: z.string(),
        }))
        .query(async ({ input }) => {
            const deepService = getDeepResearchService();
            if (deepService?.generateQueries) {
                const queries = await deepService.generateQueries(input.topic);
                return { queries };
            }
            // Fallback: return the topic itself
            return { queries: [input.topic] };
        }),

    /** Read ingestion queue state from canonical index + status tracker */
    ingestionQueue: publicProcedure.query(async () => {
        let statusDoc: IngestionStatusDoc;
        let masterEntries: Map<string, { id?: string; name?: string; category?: string }>;
        try {
            [statusDoc, masterEntries] = await Promise.all([
                readIngestionStatus(),
                loadMasterIndexEntries()
            ]);
        } catch (error) {
            throw new TRPCError({
                code: 'INTERNAL_SERVER_ERROR',
                message: `Research ingestion queue is unavailable: ${formatUnknownError(error)}`,
            });
        }

        const processed = statusDoc.processed.map((url) => {
            const normalized = normalizeUrl(url);
            const meta = masterEntries.get(normalized);
            return {
                url: normalized,
                name: meta?.name || normalized,
                id: meta?.id || null,
                category: meta?.category || null
            };
        });

        const pending = statusDoc.pending.map((url) => {
            const normalized = normalizeUrl(url);
            const meta = masterEntries.get(normalized);
            return {
                url: normalized,
                name: meta?.name || normalized,
                id: meta?.id || null,
                category: meta?.category || null
            };
        });

        const failed = statusDoc.failed.map((item) => {
            const normalized = normalizeUrl(item.url);
            const meta = masterEntries.get(normalized);
            return {
                url: normalized,
                name: meta?.name || normalized,
                id: meta?.id || null,
                category: meta?.category || null,
                error: item.error || 'Unknown fetch failure',
                attempts: item.attempts || 1,
                lastAttemptAt: item.last_attempt_at || null,
                source: item.source || 'unknown'
            };
        });

        return {
            queue: {
                processed,
                pending,
                failed
            },
            totals: {
                processed: processed.length,
                pending: pending.length,
                failed: failed.length
            },
            updatedAt: new Date().toISOString()
        };
    }),

    /** Move a failed URL back to pending for reprocessing */
    retryFailed: publicProcedure
        .input(z.object({
            url: z.string().url(),
        }))
        .mutation(async ({ input }) => {
            const statusDoc = await readIngestionStatus();
            const normalized = normalizeUrl(input.url);

            const failedItem = statusDoc.failed.find((item) => normalizeUrl(item.url) === normalized);
            if (!failedItem) {
                return { success: false, message: 'URL not found in failed queue.' };
            }

            statusDoc.failed = statusDoc.failed.filter((item) => normalizeUrl(item.url) !== normalized);
            if (!statusDoc.pending.some((u) => normalizeUrl(u) === normalized)) {
                statusDoc.pending.push(normalized);
            }

            await writeIngestionStatus(statusDoc);
            return {
                success: true,
                url: normalized,
                attempts: (failedItem.attempts || 1) + 1,
                message: 'Moved URL from failed to pending queue.'
            };
        }),

    /** Move all failed URLs back to pending for reprocessing */
    retryAllFailed: publicProcedure
        .mutation(async () => {
            const statusDoc = await readIngestionStatus();
            if (statusDoc.failed.length === 0) {
                return {
                    success: true,
                    moved: 0,
                    message: 'No failed URLs to retry.'
                };
            }

            const failedUrls = statusDoc.failed.map((item) => normalizeUrl(item.url));
            const pendingSet = new Set(statusDoc.pending.map((url) => normalizeUrl(url)));

            for (const url of failedUrls) {
                pendingSet.add(url);
            }

            statusDoc.pending = Array.from(pendingSet);
            statusDoc.failed = [];

            await writeIngestionStatus(statusDoc);
            return {
                success: true,
                moved: failedUrls.length,
                message: `Moved ${failedUrls.length} failed URL(s) to pending queue.`
            };
        }),

    /** Add a URL directly into pending ingestion queue */
    enqueuePending: publicProcedure
        .input(z.object({
            url: z.string().url(),
            source: z.string().default('dashboard-reader'),
        }))
        .mutation(async ({ input }) => {
            const statusDoc = await readIngestionStatus();
            const normalized = normalizeUrl(input.url);

            if (statusDoc.processed.some((u) => normalizeUrl(u) === normalized)) {
                return {
                    success: false,
                    queued: false,
                    url: normalized,
                    message: 'URL already processed.'
                };
            }

            const wasPending = statusDoc.pending.some((u) => normalizeUrl(u) === normalized);
            statusDoc.failed = statusDoc.failed.filter((item) => normalizeUrl(item.url) !== normalized);

            if (!wasPending) {
                statusDoc.pending.push(normalized);
                await writeIngestionStatus(statusDoc);
                return {
                    success: true,
                    queued: true,
                    url: normalized,
                    source: input.source,
                    message: 'URL added to pending ingestion queue.'
                };
            }

            await writeIngestionStatus(statusDoc);
            return {
                success: true,
                queued: false,
                url: normalized,
                source: input.source,
                message: 'URL already pending.'
            };
        }),
});
