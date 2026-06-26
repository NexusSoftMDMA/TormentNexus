
import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { t, publicProcedure, getSquadService } from '../lib/trpc-core.js';

export const squadRouter = t.router({
    list: publicProcedure.query(async () => {
        const service = getSquadService();
        if (!service) return [];
        return service.listMembers();
    }),

    spawn: publicProcedure
        .input(z.object({
            branch: z.string(),
            goal: z.string()
        }))
        .mutation(async ({ input }) => {
            const service = getSquadService();
            if (!service) throw new Error("SquadService not initialized");
            return await service.spawnMember(input.branch, input.goal);
        }),

    kill: publicProcedure
        .input(z.object({
            branch: z.string()
        }))
        .mutation(async ({ input }) => {
            const service = getSquadService();
            if (!service) throw new Error("SquadService not initialized");
            return await service.killMember(input.branch);
        }),

    chat: publicProcedure
        .input(z.object({
            branch: z.string(),
            message: z.string()
        }))
        .mutation(async ({ input }) => {
            const service = getSquadService();
            if (!service) throw new Error("SquadService not initialized");
            return await service.messageMember(input.branch, input.message);
        }),

    // --- Indexer ---

    toggleIndexer: publicProcedure
        .input(z.object({ enabled: z.boolean() }))
        .mutation(async ({ input }) => {
            const service = getSquadService();
            if (!service) {
                throw new TRPCError({
                    code: 'SERVICE_UNAVAILABLE',
                    message: 'Squad service is not initialized; cannot toggle the indexer.',
                });
            }
            return await service.toggleIndexer(input.enabled);
        }),

    getIndexerStatus: publicProcedure.query(() => {
        const service = getSquadService();
        if (!service) {
            throw new TRPCError({
                code: 'SERVICE_UNAVAILABLE',
                message: 'Squad service is not initialized.',
            });
        }
        return service.getIndexerStatus();
    })
});

