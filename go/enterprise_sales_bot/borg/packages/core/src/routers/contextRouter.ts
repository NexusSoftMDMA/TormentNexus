import { z } from 'zod';
import { t, publicProcedure, getContextManager } from '../lib/trpc-core.js';
import { contextHarvester, type ContextChunk, type HarvestReport } from '../services/ContextHarvester.js';

export const contextRouter = t.router({
    list: publicProcedure
        .output(z.array(z.string()))
        .query((): string[] => {
            return (getContextManager()?.list() ?? []) as string[];
        }),

    add: publicProcedure.input(z.object({
        filePath: z.string()
    })).mutation(({ input }) => {
        const contextManager = getContextManager();
        if (contextManager) return contextManager.add(input.filePath);
        return 'ContextManager not initialized';
    }),

    remove: publicProcedure.input(z.object({
        filePath: z.string()
    })).mutation(({ input }) => {
        const contextManager = getContextManager();
        if (contextManager) return contextManager.remove(input.filePath);
        return 'ContextManager not initialized';
    }),

    clear: publicProcedure.mutation(() => {
        const contextManager = getContextManager();
        if (contextManager) return contextManager.clear();
        return 'ContextManager not initialized';
    }),

    getPrompt: publicProcedure.query(() => {
        return getContextManager()?.getContextPrompt() ?? '';
    }),

    getHarvestedContext: publicProcedure.query((): ContextChunk[] => {
        return contextHarvester.getChunks();
    }),

    getHarvestStats: publicProcedure.query((): HarvestReport => {
        return contextHarvester.getReport();
    }),

    compact: publicProcedure.mutation((): HarvestReport => {
        contextHarvester.compact();
        return contextHarvester.getReport();
    }),

    prune: publicProcedure.mutation((): HarvestReport => {
        contextHarvester.prune();
        return contextHarvester.getReport();
    }),
});

