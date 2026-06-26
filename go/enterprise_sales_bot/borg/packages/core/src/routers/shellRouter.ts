import { z } from 'zod';
import { t, publicProcedure, getShellService } from '../lib/trpc-core.js';

export const shellRouter = t.router({
    logCommand: publicProcedure.input(z.object({
        command: z.string(),
        cwd: z.string(),
        exitCode: z.number().optional(),
        duration: z.number().optional(),
        outputSnippet: z.string().optional(),
        session: z.string()
    })).mutation(async ({ input }) => {
        const service = getShellService();
        if (service) {
            const id = await service.logCommand(input);
            return { success: true, id };
        }
        return { success: false, error: 'ShellService not initialized' };
    }),

    queryHistory: publicProcedure.input(z.object({
        query: z.string(),
        limit: z.number().optional()
    })).query(async ({ input }) => {
        return await getShellService()?.queryHistory(input.query, input.limit) ?? [];
    }),

    getSystemHistory: publicProcedure.input(z.object({
        limit: z.number().optional()
    })).query(async ({ input }) => {
        return await getShellService()?.getSystemHistory(input.limit) ?? [];
    })
});

