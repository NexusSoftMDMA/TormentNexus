import { z } from 'zod';
import { t, publicProcedure, getAutoDevService } from '../lib/trpc-core.js';

export const autoDevRouter = t.router({
    startLoop: publicProcedure.input(z.object({
        type: z.enum(['test', 'lint', 'build']),
        maxAttempts: z.number().min(1).max(10).default(5),
        target: z.string().optional(),
        command: z.string().optional()
    })).mutation(async ({ input }) => {
        const service = getAutoDevService();
        if (service) {
            const id = await service.startLoop(input);
            return { success: true, loopId: id };
        }
        throw new Error('AutoDevService not initialized');
    }),

    cancelLoop: publicProcedure.input(z.object({
        loopId: z.string()
    })).mutation(({ input }) => {
        return getAutoDevService()?.cancelLoop(input.loopId) ?? false;
    }),

    getLoops: publicProcedure.query(() => {
        return getAutoDevService()?.getLoops() ?? [];
    }),

    getLoop: publicProcedure.input(z.object({
        loopId: z.string()
    })).query(({ input }) => {
        return getAutoDevService()?.getLoop(input.loopId) ?? null;
    }),

    clearCompleted: publicProcedure.mutation(() => {
        return getAutoDevService()?.clearCompleted() ?? 0;
    }),
});

