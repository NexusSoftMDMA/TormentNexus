import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { browserDataRepository } from '../db/repositories/browser-data.repo.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

/**
 * Memory Browser Extension Router
 *
 * Provides HTTP endpoints for a browser extension to save web memories:
 * - Save the current page (URL, title, text content, selection)
 * - Parse and store DOM snapshots for RAG
 * - Tag and categorize saved memories
 * - Retrieve saved memories by URL or content search
 */

const SaveMemoryInputSchema = z.object({
    url: z.string().url(),
    title: z.string().max(500),
    content: z.string().max(100_000),
    selectedText: z.string().max(10_000).optional(),
    tags: z.array(z.string()).optional(),
    favicon: z.string().url().optional(),
    timestamp: z.number().optional(),
    source: z.literal('browser-extension').default('browser-extension'),
});

const ParseDomInputSchema = z.object({
    url: z.string().url(),
    html: z.string().max(5_000_000),
    extractLinks: z.boolean().optional(),
    extractImages: z.boolean().optional(),
    extractMetadata: z.boolean().optional(),
});

export interface WebMemory {
    id: string;
    url: string;
    normalizedUrl: string;
    title: string;
    content: string;
    selectedText?: string;
    tags: string[];
    favicon?: string;
    savedAt: Date;
    source: string;
    contentHash: string;
}

export interface ParsedDom {
    url: string;
    title: string;
    description: string;
    mainContent: string;
    links: Array<{ href: string; text: string }>;
    images: Array<{ src: string; alt: string }>;
    metadata: Record<string, string>;
    wordCount: number;
}

function normalizeUrl(rawUrl: string): string {
    try {
        const url = new URL(rawUrl);
        const blockedParams = new Set([
            'utm_source', 'utm_medium', 'utm_campaign', 'utm_term', 'utm_content',
            'fbclid', 'gclid', 'ref', 'source',
        ]);
        const kept = Array.from(url.searchParams.entries())
            .filter(([key]) => !blockedParams.has(key.toLowerCase()))
            .sort(([a], [b]) => a.localeCompare(b));
        url.search = '';
        for (const [key, value] of kept) url.searchParams.append(key, value);
        url.hash = '';
        return url.toString();
    } catch {
        return rawUrl.trim();
    }
}

function simpleHash(content: string): string {
    let hash = 0;
    for (let i = 0; i < content.length; i++) {
        const char = content.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash = hash & hash; // Convert to 32-bit integer
    }
    return Math.abs(hash).toString(36);
}

function stripHtmlToText(html: string): string {
    return html
        .replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '')
        .replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
        .replace(/<[^>]+>/g, ' ')
        .replace(/\s+/g, ' ')
        .trim();
}

function extractLinksFromHtml(html: string): Array<{ href: string; text: string }> {
    const links: Array<{ href: string; text: string }> = [];
    const regex = /<a\s[^>]*href=["']([^"']+)["'][^>]*>([\s\S]*?)<\/a>/gi;
    let match;
    while ((match = regex.exec(html)) !== null && links.length < 500) {
        links.push({ href: match[1], text: stripHtmlToText(match[2]).substring(0, 200) });
    }
    return links;
}

function extractImagesFromHtml(html: string): Array<{ src: string; alt: string }> {
    const images: Array<{ src: string; alt: string }> = [];
    const regex = /<img\s[^>]*src=["']([^"']+)["'][^>]*(?:alt=["']([^"']*)["'])?[^>]*\/?>/gi;
    let match;
    while ((match = regex.exec(html)) !== null && images.length < 200) {
        images.push({ src: match[1], alt: match[2] || '' });
    }
    return images;
}

function extractMetaFromHtml(html: string): Record<string, string> {
    const meta: Record<string, string> = {};
    const regex = /<meta\s[^>]*(?:name|property)=["']([^"']+)["'][^>]*content=["']([^"']+)["'][^>]*\/?>/gi;
    let match;
    while ((match = regex.exec(html)) !== null) {
        meta[match[1]] = match[2];
    }
    const titleMatch = html.match(/<title[^>]*>([\s\S]*?)<\/title>/i);
    if (titleMatch) meta['title'] = stripHtmlToText(titleMatch[1]);
    return meta;
}

// In-memory store replaced with SQLite: browserDataRepository

export const browserExtensionRouter = t.router({
    /**
     * Save a web page as a memory. Called by the browser extension when the
     * user clicks "Save to TormentNexus" or highlights text.
     */
    saveMemory: adminProcedure
        .input(SaveMemoryInputSchema)
        .mutation(async ({ input }) => {
            try {
                const normalized = normalizeUrl(input.url);
                const existingAll = await browserDataRepository.getAllMemories();
                const existing = existingAll.find(m => m.normalizedUrl === normalized);
                
                if (existing) {
                    return { id: existing.id, deduplicated: true };
                }

                const memory = await browserDataRepository.saveMemory({
                    id: `mem_${Date.now()}_${simpleHash(input.url)}`,
                    url: input.url,
                    normalizedUrl: normalized,
                    title: input.title,
                    content: input.content.substring(0, 50_000),
                    selectedText: input.selectedText || null,
                    tags: input.tags ?? [],
                    favicon: input.favicon || null,
                    savedAt: new Date(input.timestamp ?? Date.now()),
                    source: input.source,
                    contentHash: simpleHash(input.content),
                });

                return { id: memory.id, deduplicated: false };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Browser memory store is unavailable', error);
            }
        }),

    /**
     * Parse a DOM snapshot from the browser extension.
     * Extracts clean text, links, images, and metadata.
     */
    parseDom: adminProcedure
        .input(ParseDomInputSchema)
        .mutation(async ({ input }) => {
            const mainContent = stripHtmlToText(input.html);
            const metadata = extractMetaFromHtml(input.html);

            const result: ParsedDom = {
                url: input.url,
                title: metadata['title'] || metadata['og:title'] || '',
                description: metadata['description'] || metadata['og:description'] || '',
                mainContent: mainContent.substring(0, 100_000),
                links: input.extractLinks ? extractLinksFromHtml(input.html) : [],
                images: input.extractImages ? extractImagesFromHtml(input.html) : [],
                metadata: input.extractMetadata ? metadata : {},
                wordCount: mainContent.split(/\s+/).filter(Boolean).length,
            };

            return result;
        }),

    /**
     * List saved web memories with optional search.
     */
    listMemories: publicProcedure
        .input(z.object({
            search: z.string().optional(),
            tag: z.string().optional(),
            limit: z.number().min(1).max(100).default(50),
            offset: z.number().min(0).default(0),
        }))
        .query(async ({ input }) => {
            try {
                const allMemories = await browserDataRepository.getAllMemories();
                let filtered = [...allMemories];

                if (input.search) {
                    const q = input.search.toLowerCase();
                    filtered = filtered.filter(m =>
                        m.title.toLowerCase().includes(q) ||
                        m.url.toLowerCase().includes(q) ||
                        m.content.toLowerCase().includes(q)
                    );
                }

                if (input.tag) {
                    filtered = filtered.filter(m => m.tags.includes(input.tag!));
                }

                filtered.sort((a, b) => b.savedAt.getTime() - a.savedAt.getTime());

                return {
                    items: filtered.slice(input.offset, input.offset + input.limit),
                    total: filtered.length,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Browser memory store is unavailable', error);
            }
        }),

    /**
     * Delete a saved memory by ID.
     */
    deleteMemory: adminProcedure
        .input(z.object({ id: z.string() }))
        .mutation(async ({ input }) => {
            try {
                const deleted = await browserDataRepository.deleteMemory(input.id);
                return { deleted };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Browser memory store is unavailable', error);
            }
        }),

    /**
     * Get stats about saved web memories.
     */
    stats: publicProcedure.query(async () => {
        try {
            const allMemories = await browserDataRepository.getAllMemories();
            const tagCounts: Record<string, number> = {};
            for (const memory of allMemories) {
                for (const tag of memory.tags) {
                    tagCounts[tag] = (tagCounts[tag] || 0) + 1;
                }
            }

            return {
                totalMemories: allMemories.length,
                uniqueUrls: new Set(allMemories.map(m => m.normalizedUrl)).size,
                topTags: Object.entries(tagCounts)
                    .sort(([, a], [, b]) => b - a)
                    .slice(0, 20)
                    .map(([tag, count]) => ({ tag, count })),
            };
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Browser memory store is unavailable', error);
        }
    }),
});
