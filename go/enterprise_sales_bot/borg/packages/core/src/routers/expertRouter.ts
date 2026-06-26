import { t, publicProcedure, getMcpServer } from '../lib/trpc-core.js';
import { z } from 'zod';
import { TRPCError } from '@trpc/server';

export const expertRouter = t.router({
    research: publicProcedure
        .input(z.object({
            query: z.string(),
            depth: z.number().default(2),
            breadth: z.number().default(3),
        }))
        .mutation(async ({ input }) => {
            const mcp = getMcpServer();
            const agent = mcp.researcherAgent;
            if (!agent) {
                throw new TRPCError({
                    code: 'INTERNAL_SERVER_ERROR',
                    message: 'Researcher Agent not initialized',
                });
            }

            console.log(`[ExpertRouter] 🕵️ Dispatching research task: ${input.query}`);
            const result = await agent.handleTask({
                task: input.query,
                options: { depth: input.depth, breadth: input.breadth }
            });

            return result;
        }),

    code: publicProcedure
        .input(z.object({
            task: z.string(),
        }))
        .mutation(async ({ input }) => {
            const mcp = getMcpServer();
            const agent = mcp.coderAgent;
            if (!agent) {
                throw new TRPCError({
                    code: 'INTERNAL_SERVER_ERROR',
                    message: 'Coder Agent not initialized',
                });
            }

            console.log(`[ExpertRouter] 👨‍💻 Dispatching coding task: ${input.task}`);
            const result = await agent.handleTask({ task: input.task });
            return result;
        }),

    getStatus: publicProcedure
        .query(async () => {
            const mcp = getMcpServer();
            return {
                researcher: mcp.researcherAgent ? 'active' : 'offline',
                coder: mcp.coderAgent ? 'active' : 'offline',
            };
        }),
});
