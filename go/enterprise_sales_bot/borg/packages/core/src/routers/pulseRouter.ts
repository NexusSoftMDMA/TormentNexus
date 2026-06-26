
import { z } from 'zod';
import { t, publicProcedure, getMcpServer, getEventBus } from '../lib/trpc-core.js';

export const pulseRouter = t.router({
    getLatestEvents: publicProcedure
        .input(z.object({
            limit: z.number().default(20),
            afterTimestamp: z.number().optional()
        }))
        .query(async ({ input }) => {
            const eventBus = getEventBus();
            if (!eventBus) return [];

            const history = eventBus.getHistory(input.limit);

            if (input.afterTimestamp) {
                return history.filter(e => e.timestamp > input.afterTimestamp!);
            }

            return history;
        }),

    getSystemStatus: publicProcedure.query(async () => {
        const mcp = getMcpServer();
        if (!mcp) return { status: 'offline' };

        return {
        status: 'online',
        uptime: process.uptime(),
        agents: Array.from(mcp.activeAgentsMap?.keys() || []),
        memoryInitialized: mcp.isMemoryInitialized,
        a2aActive: true
        };
        }),

        checkLocalProviders: publicProcedure.query(async () => {
        const mcp = getMcpServer();
        if (!mcp) return { ollama: false, lmstudio: false };
        return await mcp.llmService.probeLocalProviders();
        })
        });
