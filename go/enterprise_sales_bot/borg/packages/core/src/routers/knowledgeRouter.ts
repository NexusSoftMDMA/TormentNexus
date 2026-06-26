import { z } from 'zod';
import {
    t,
    publicProcedure,
    getKnowledgeService,
    getMemoryManager,
    getDeepResearchService,
} from '../lib/trpc-core.js';

export const knowledgeRouter = t.router({
    getGraph: publicProcedure
        .input(z.object({
            query: z.string().optional(),
            depth: z.number().optional().default(1)
        }))
        .query(async ({ input }) => {
            const result = await getKnowledgeService().getGraph(input.query, input.depth);

            try {
                const content = result.content[0].text;
                const graphData = JSON.parse(content);
                return graphData;
            } catch (e) {
                console.error("Failed to parse graph data", e);
                return { nodes: [], edges: [] };
            }
        }),

    getStats: publicProcedure.query(async () => {
        try {
            const contexts = await getMemoryManager()?.listContexts() ?? [];
            return { count: contexts.length };
        } catch { return { count: 0 }; }
    }),

    ingest: publicProcedure
        .input(z.object({ url: z.string() }))
        .mutation(async ({ input }) => {
            return getDeepResearchService().ingest(input.url);
        }),

    getResources: publicProcedure.query(async () => {
        const fs = await import('fs/promises');
        const path = await import('path');
        const resourcePath = path.join(process.cwd(), 'knowledge', 'resources.json');

        try {
            const content = await fs.readFile(resourcePath, 'utf-8');
            return JSON.parse(content);
        } catch (e) {
            console.warn("Resources file not found or invalid", e);
            return { lastUpdated: "Never", categories: [] };
        }
    })
});
