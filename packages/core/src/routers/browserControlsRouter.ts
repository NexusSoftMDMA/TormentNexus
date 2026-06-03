import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { browserDataRepository } from '../db/repositories/browser-data.repo.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

/**
 * Browser Controls Router (Phase J)
 *
 * Enables TormentNexus to interact with browser content:
 * - Scrape web pages (via fetch or headless browser)
 * - Read browser history (from extension bridge)
 * - Intercept console/debug logs (from extension bridge)
 */

const ScrapeInputSchema = z.object({
    url: z.string().url(),
    selector: z.string().optional(),
    waitMs: z.number().min(0).max(30_000).default(0),
    extractLinks: z.boolean().default(false),
    extractImages: z.boolean().default(false),
    maxContentLength: z.number().min(100).max(500_000).default(100_000),
    userAgent: z.string().optional(),
});

const HistoryQuerySchema = z.object({
    query: z.string().optional(),
    limit: z.number().min(1).max(500).default(50),
    since: z.number().optional(), // Unix timestamp
    domain: z.string().optional(),
});

export interface ScrapedPage {
    url: string;
    title: string;
    description: string;
    content: string;
    links: Array<{ href: string; text: string }>;
    images: Array<{ src: string; alt: string }>;
    statusCode: number;
    contentType: string;
    scrapedAt: number;
    wordCount: number;
}

export interface BrowserHistoryEntry {
    url: string;
    title: string;
    visitedAt: number;
    visitCount: number;
    domain: string;
}

export interface ConsoleLogEntry {
    level: 'log' | 'warn' | 'error' | 'info' | 'debug';
    message: string;
    source: string;
    timestamp: number;
    url?: string;
    lineNumber?: number;
}

function stripHtml(html: string): string {
    return html
        .replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '')
        .replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
        .replace(/<[^>]+>/g, ' ')
        .replace(/\s+/g, ' ')
        .trim();
}

function extractTitle(html: string): string {
    const match = html.match(/<title[^>]*>([\s\S]*?)<\/title>/i);
    return match ? stripHtml(match[1]) : '';
}

function extractMeta(html: string, name: string): string {
    const regex = new RegExp(`<meta[^>]*(?:name|property)=["']${name}["'][^>]*content=["']([^"']+)["']`, 'i');
    const match = html.match(regex);
    return match ? match[1] : '';
}

function extractLinks(html: string): Array<{ href: string; text: string }> {
    const links: Array<{ href: string; text: string }> = [];
    const regex = /<a\s[^>]*href=["']([^"']+)["'][^>]*>([\s\S]*?)<\/a>/gi;
    let match;
    while ((match = regex.exec(html)) !== null && links.length < 200) {
        links.push({ href: match[1], text: stripHtml(match[2]).substring(0, 200) });
    }
    return links;
}

function extractImages(html: string): Array<{ src: string; alt: string }> {
    const images: Array<{ src: string; alt: string }> = [];
    const regex = /<img\s[^>]*src=["']([^"']+)["'][^>]*(?:alt=["']([^"']*)["'])?/gi;
    let match;
    while ((match = regex.exec(html)) !== null && images.length < 100) {
        images.push({ src: match[1], alt: match[2] || '' });
    }
    return images;
}

export const browserControlsRouter = t.router({
    /**
     * Scrape a web page by URL (server-side fetch).
     */
    scrape: adminProcedure
        .input(ScrapeInputSchema)
        .mutation(async ({ input }) => {
            const headers: Record<string, string> = {
                Accept: 'text/html,application/xhtml+xml',
                'User-Agent': input.userAgent || 'TormentNexus/BrowserControls (compatible)',
            };

            const response = await fetch(input.url, { headers });
            const html = await response.text();

            const content = stripHtml(html).substring(0, input.maxContentLength);
            const result: ScrapedPage = {
                url: input.url,
                title: extractTitle(html),
                description: extractMeta(html, 'description') || extractMeta(html, 'og:description'),
                content,
                links: input.extractLinks ? extractLinks(html) : [],
                images: input.extractImages ? extractImages(html) : [],
                statusCode: response.status,
                contentType: response.headers.get('content-type') || '',
                scrapedAt: Date.now(),
                wordCount: content.split(/\s+/).filter(Boolean).length,
            };

            return result;
        }),

    /**
     * Push browser history entries (from extension bridge).
     */
    pushHistory: adminProcedure
        .input(z.object({
            entries: z.array(z.object({
                url: z.string().url(),
                title: z.string(),
                visitedAt: z.number(),
                visitCount: z.number().default(1),
            })),
        }))
        .mutation(async ({ input }) => {
            try {
                for (const entry of input.entries) {
                    const domain = new URL(entry.url).hostname;
                    await browserDataRepository.addHistoryEntry({
                        id: `hist_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
                        url: entry.url,
                        title: entry.title,
                        domain,
                        visitedAt: new Date(entry.visitedAt),
                        visitCount: entry.visitCount,
                    });
                }
                return { stored: input.entries.length };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Browser data store is unavailable', error);
            }
        }),

    /**
     * Query browser history.
     */
    queryHistory: publicProcedure
        .input(HistoryQuerySchema)
        .query(async ({ input }) => {
            try {
                const allHistory = await browserDataRepository.getHistory();
                let filtered = [...allHistory];

                if (input.query) {
                    const q = input.query.toLowerCase();
                    filtered = filtered.filter(e =>
                        e.title.toLowerCase().includes(q) || e.url.toLowerCase().includes(q)
                    );
                }
                if (input.domain) {
                    filtered = filtered.filter(e => e.domain === input.domain);
                }
                if (input.since) {
                    filtered = filtered.filter(e => e.visitedAt.getTime() >= input.since!);
                }

                filtered.sort((a, b) => b.visitedAt.getTime() - a.visitedAt.getTime());
                return {
                    entries: filtered.slice(0, input.limit).map(e => ({
                        ...e,
                        visitedAt: e.visitedAt.getTime()
                    })),
                    total: filtered.length
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Browser data store is unavailable', error);
            }
        }),

    /**
     * Push console/debug logs from browser extension.
     */
    pushConsoleLogs: adminProcedure
        .input(z.object({
            logs: z.array(z.object({
                level: z.enum(['log', 'warn', 'error', 'info', 'debug']),
                message: z.string(),
                source: z.string(),
                timestamp: z.number(),
                url: z.string().optional(),
                lineNumber: z.number().optional(),
            })),
        }))
        .mutation(async ({ input }) => {
            try {
                for (const log of input.logs) {
                    await browserDataRepository.addConsoleLog({
                        id: `log_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`,
                        level: log.level,
                        message: log.message,
                        source: log.source,
                        url: log.url || null,
                        lineNumber: log.lineNumber || null,
                        timestamp: new Date(log.timestamp),
                    });
                }
                return { stored: input.logs.length };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Browser data store is unavailable', error);
            }
        }),

    /**
     * Query intercepted console logs.
     */
    queryConsoleLogs: publicProcedure
        .input(z.object({
            level: z.enum(['log', 'warn', 'error', 'info', 'debug']).optional(),
            search: z.string().optional(),
            limit: z.number().min(1).max(500).default(100),
        }))
        .query(async ({ input }) => {
            try {
                const allLogs = await browserDataRepository.getConsoleLogs({ level: input.level });
                let filtered = [...allLogs];
                
                if (input.search) {
                    const q = input.search.toLowerCase();
                    filtered = filtered.filter(l => l.message.toLowerCase().includes(q));
                }
                
                filtered.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
                return {
                    logs: filtered.slice(0, input.limit).map(l => ({
                        ...l,
                        timestamp: l.timestamp.getTime()
                    })),
                    total: filtered.length
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Browser data store is unavailable', error);
            }
        }),

    /**
     * Stats on collected browser data.
     */
    stats: publicProcedure.query(async () => {
        try {
            const history = await browserDataRepository.getHistory();
            const logs = await browserDataRepository.getConsoleLogs();
            
            const domains = new Set(history.map(e => e.domain));
            const errorCount = logs.filter(l => l.level === 'error').length;
            
            return {
                historyCount: history.length,
                uniqueDomains: domains.size,
                consoleLogCount: logs.length,
                consoleErrors: errorCount,
            };
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Browser data store is unavailable', error);
        }
    }),
});
