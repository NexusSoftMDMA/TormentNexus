
import { z } from 'zod';
import {
    t,
    publicProcedure,
    adminProcedure,
    getMarketplaceService
} from '../lib/trpc-core.js';
// We redefine schema here for API boundary clarity, or import if robust. 
// Redefined to avoid tight coupling if service internals change.
const marketplaceEntrySchema = z.object({
    id: z.string(),
    name: z.string(),
    description: z.string(),
    author: z.string().optional(),
    type: z.enum(['agent', 'tool', 'skill']),
    source: z.enum(['official', 'community', 'local']),
    url: z.string().optional(),
    verified: z.boolean(),
    peerCount: z.number(),
    installed: z.boolean(),
    tags: z.array(z.string())
});

export const marketplaceRouter = t.router({
    list: publicProcedure
        .input(z.object({ filter: z.string().optional() }).optional())
        .output(z.array(marketplaceEntrySchema))
        .query(async ({ input }) => {
            const service = getMarketplaceService();
            if (!service) return []; // Graceful degradation
            return service.list(input?.filter);
        }),

    install: adminProcedure // Install requires permissions
        .input(z.object({ id: z.string() }))
        .mutation(async ({ input }) => {
            const service = getMarketplaceService();
            if (!service) throw new Error("Marketplace Service not available");
            return service.install(input.id);
        }),

    publish: adminProcedure
        .input(z.object({
            name: z.string(),
            description: z.string(),
            url: z.string()
        }))
        .mutation(async ({ input }) => {
            const service = getMarketplaceService();
            if (!service) throw new Error("Marketplace Service not available");
            return service.publish({
                ...input,
                type: 'tool', // Default
                source: 'community'
            });
        })
});
