import { z } from 'zod';
import { t, publicProcedure, getSymbolPinService, getMemoryManager } from '../lib/trpc-core.js';

export const symbolsRouter = t.router({
    list: publicProcedure.query(() => {
        return getSymbolPinService()?.list() ?? [];
    }),

    find: publicProcedure.input(z.object({
        query: z.string(),
        limit: z.number().default(10)
    })).query(async ({ input }) => {
        return await getMemoryManager().searchSymbols?.(input.query, input.limit) ?? [];
    }),

    pin: publicProcedure.input(z.object({
        name: z.string(),
        file: z.string(),
        type: z.enum(['function', 'class', 'method', 'variable', 'interface']),
        lineStart: z.number().optional(),
        lineEnd: z.number().optional(),
        notes: z.string().optional()
    })).mutation(({ input }) => {
        const service = getSymbolPinService();
        if (service) return service.pin(input);
        throw new Error('SymbolPinService not initialized');
    }),

    unpin: publicProcedure.input(z.object({
        id: z.string()
    })).mutation(({ input }) => {
        const service = getSymbolPinService();
        if (service) return service.unpin(input.id);
        return false;
    }),

    updatePriority: publicProcedure.input(z.object({
        id: z.string(),
        priority: z.number()
    })).mutation(({ input }) => {
        const service = getSymbolPinService();
        if (service) return service.updatePriority(input.id, input.priority);
        return false;
    }),

    addNotes: publicProcedure.input(z.object({
        id: z.string(),
        notes: z.string()
    })).mutation(({ input }) => {
        const service = getSymbolPinService();
        if (service) return service.addNotes(input.id, input.notes);
        return false;
    }),

    clear: publicProcedure.mutation(() => {
        const service = getSymbolPinService();
        if (service) return service.clear();
        return 0;
    }),

    forFile: publicProcedure.input(z.object({
        filePath: z.string()
    })).query(({ input }) => {
        return getSymbolPinService()?.forFile(input.filePath) ?? [];
    }),
});

