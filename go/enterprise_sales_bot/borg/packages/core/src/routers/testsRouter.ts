import { z } from 'zod';
import { t, publicProcedure, getAutoTestService } from '../lib/trpc-core.js';

export const testsRouter = t.router({
    status: publicProcedure.query(() => {
        const service = getAutoTestService();
        const results: Record<string, { status: string; timestamp: number; output?: string }> = {};
        for (const [file, result] of service.testResults.entries()) {
            results[file] = result;
        }
        return {
            isRunning: service.isRunning,
            results
        };
    }),

    start: publicProcedure.mutation(async () => {
        await getAutoTestService().start();
        return { success: true };
    }),

    stop: publicProcedure.mutation(() => {
        getAutoTestService().stop();
        return { success: true };
    }),

    run: publicProcedure.input(z.object({
        filePath: z.string()
    })).mutation(async ({ input }) => {
        const service = getAutoTestService();
        // Manually trigger test runner
        const testFile = service.findTestFile?.(input.filePath);
        if (testFile) {
            service.runTest?.(testFile);
            return { success: true, testFile };
        }
        return { success: false, error: 'No test file found' };
    }),

    results: publicProcedure.query(() => {
        const service = getAutoTestService();
        const results: Array<{ file: string; status: string; timestamp: number; output?: string }> = [];
        for (const [file, result] of service.testResults.entries()) {
            results.push({ file, ...result });
        }
        return results.sort((a, b) => b.timestamp - a.timestamp);
    }),
});
