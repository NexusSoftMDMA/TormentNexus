import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { serverErrorTracker } from '../services/server-error-tracker.service.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

export const serverHealthRouter = t.router({
    check: publicProcedure
        .input(z.object({ serverUuid: z.string() }))
        .query(async ({ input }) => {
            try {
                const hasError = await serverErrorTracker.isServerInErrorState(input.serverUuid);
                const crashCount = serverErrorTracker.getServerAttempts(input.serverUuid);
                const maxAttempts = await serverErrorTracker.getServerMaxAttempts(input.serverUuid);

                return {
                    status: hasError ? "ERROR" : "HEALTHY",
                    crashCount,
                    maxAttempts,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Server health status is unavailable', error);
            }
        }),

    reset: adminProcedure
        .input(z.object({ serverUuid: z.string() }))
        .mutation(async ({ input }) => {
            await serverErrorTracker.resetServerErrorState(input.serverUuid);
            return { success: true };
        }),
});
