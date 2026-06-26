import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { t, publicProcedure, getBrowserService, getMcpServer } from '../lib/trpc-core.js';

type BrowserHistoryItem = {
    title: string;
    url: string;
    visitCount: number;
};

type BrowserDebugResponse = {
    raw: string;
    data: unknown;
};

type BrowserScrapeResponse = {
    raw: string;
    url: string | null;
    title: string | null;
    content: string;
};

type BrowserScreenshotResponse = {
    message: string;
    imageDataUrl: string | null;
    mimeType: string | null;
};

type BrowserProxyFetchResponse = {
    raw: string;
    data: {
        status?: number;
        statusText?: string;
        body?: string;
        error?: string;
        [key: string]: unknown;
    } | null;
};

function extractToolText(result: any): string {
    return Array.isArray(result?.content)
        ? result.content.map((item: any) => item?.text ?? '').join('\n')
        : '';
}

function parseJsonText(text: string): unknown {
    try {
        return JSON.parse(text);
    } catch {
        return text;
    }
}

function extractImageContent(result: any): { imageDataUrl: string | null; mimeType: string | null } {
    const imagePart = Array.isArray(result?.content)
        ? result.content.find((item: any) => item?.type === 'image' && typeof item?.data === 'string')
        : null;

    if (!imagePart) {
        return { imageDataUrl: null, mimeType: null };
    }

    const mimeType = typeof imagePart.mimeType === 'string' ? imagePart.mimeType : 'image/png';
    return {
        imageDataUrl: `data:${mimeType};base64,${imagePart.data}`,
        mimeType,
    };
}

export const browserRouter = t.router({
    status: publicProcedure.query(() => {
        const service = getBrowserService();
        if (!service) {
            return {
                available: false,
                active: false,
                pageCount: 0,
                pageIds: [] as string[],
            };
        }

        const status = service.getStatus();
        return {
            available: true,
            ...status,
        };
    }),

    closePage: publicProcedure
        .input(z.object({ pageId: z.string().min(1) }))
        .mutation(async ({ input }) => {
            const service = getBrowserService();
            if (!service) {
                throw new TRPCError({
                    code: 'SERVICE_UNAVAILABLE',
                    message: 'Browser service unavailable',
                });
            }

            await service.close(input.pageId);
            return { success: true };
        }),

    closeAll: publicProcedure.mutation(async () => {
        const service = getBrowserService();
        if (!service) {
            throw new TRPCError({
                code: 'SERVICE_UNAVAILABLE',
                message: 'Browser service unavailable',
            });
        }

        await service.closeAll();
        return { success: true };
    }),

    searchHistory: publicProcedure
        .input(z.object({
            query: z.string().trim().min(1),
            maxResults: z.number().int().min(1).max(50).default(10),
        }))
        .query(async ({ input }) => {
            const result = await getMcpServer().executeTool('browser_get_history', {
                query: input.query,
                maxResults: input.maxResults,
            });

            const text = extractToolText(result);

            const items = text
                .split(/\r?\n/)
                .map((line: string) => line.trim())
                .filter(Boolean)
                .map((line: string) => {
                    const match = line.match(/^- \[(.*?)\]\((.*?)\) \(Visits: (\d+)\)$/);
                    if (!match) {
                        return null;
                    }

                    return {
                        title: match[1] || 'Untitled',
                        url: match[2] || '',
                        visitCount: Number(match[3] || 0),
                    };
                })
                .filter((item: BrowserHistoryItem | null): item is BrowserHistoryItem => item !== null);

            return {
                query: input.query,
                items,
                raw: text,
            };
        }),

    scrapePage: publicProcedure
        .query(async (): Promise<BrowserScrapeResponse> => {
            const result = await getMcpServer().executeTool('browser_scrape', {});
            const text = extractToolText(result);
            const urlMatch = text.match(/^URL:\s*(.*)$/m);
            const titleMatch = text.match(/^Title:\s*(.*)$/m);
            const content = text
                .replace(/^URL:\s*.*$/m, '')
                .replace(/^Title:\s*.*$/m, '')
                .trim();

            return {
                raw: text,
                url: urlMatch?.[1]?.trim() || null,
                title: titleMatch?.[1]?.trim() || null,
                content,
            };
        }),

    screenshot: publicProcedure
        .mutation(async (): Promise<BrowserScreenshotResponse> => {
            const result = await getMcpServer().executeTool('browser_screenshot', {});
            const text = extractToolText(result) || 'Screenshot captured.';
            const image = extractImageContent(result);

            return {
                message: text,
                imageDataUrl: image.imageDataUrl,
                mimeType: image.mimeType,
            };
        }),

    debug: publicProcedure
        .input(z.object({
            action: z.enum(['attach', 'detach', 'command']),
            method: z.string().trim().optional(),
            params: z.record(z.string(), z.unknown()).optional(),
        }).superRefine((value, ctx) => {
            if (value.action === 'command' && !value.method) {
                ctx.addIssue({
                    code: z.ZodIssueCode.custom,
                    message: 'CDP command method is required when action is command.',
                    path: ['method'],
                });
            }
        }))
        .mutation(async ({ input }): Promise<BrowserDebugResponse> => {
            const result = await getMcpServer().executeTool('browser_debug', input);
            const text = extractToolText(result);

            return {
                raw: text,
                data: parseJsonText(text),
            };
        }),

    proxyFetch: publicProcedure
        .input(z.object({
            url: z.string().trim().min(1),
            method: z.string().trim().min(1).default('GET'),
            headers: z.record(z.string(), z.string()).default({}),
            body: z.string().optional(),
        }))
        .mutation(async ({ input }): Promise<BrowserProxyFetchResponse> => {
            const result = await getMcpServer().executeTool('browser_proxy_fetch', {
                url: input.url,
                options: {
                    method: input.method,
                    headers: input.headers,
                    ...(input.body ? { body: input.body } : {}),
                },
            });
            const text = extractToolText(result);
            const parsed = parseJsonText(text);

            return {
                raw: text,
                data: typeof parsed === 'object' && parsed !== null
                    ? parsed as BrowserProxyFetchResponse['data']
                    : null,
            };
        }),
});
