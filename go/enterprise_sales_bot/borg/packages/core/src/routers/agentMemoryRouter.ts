import { z } from 'zod';
import { t, getMcpServer } from '../lib/trpc-core.js';

export const agentMemoryRouter = t.router({
    search: t.procedure.input(z.object({
        query: z.string(),
        namespace: z.enum(['user', 'agent', 'project']).optional(),
        type: z.enum(['session', 'working', 'long_term']).optional(),
        limit: z.number().optional()
    })).query(async ({ input }) => {
        const server = getMcpServer();
        return server.agentMemoryService.search(input.query, {
            namespace: input.namespace,
            type: input.type,
            limit: input.limit || 20
        });
    }),

    add: t.procedure.input(z.object({
        content: z.string(),
        type: z.enum(['session', 'working', 'long_term']),
        namespace: z.enum(['user', 'agent', 'project']).optional(),
        tags: z.array(z.string()).optional()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return server.agentMemoryService.add(
            input.content,
            input.type,
            input.namespace || 'project',
            { tags: input.tags }
        );
    }),

    getRecent: t.procedure.input(z.object({
        limit: z.number().optional(),
        type: z.enum(['session', 'working', 'long_term']).optional()
    })).query(async ({ input }) => {
        const server = getMcpServer();
        return server.agentMemoryService.getRecent(input.limit || 20, { type: input.type });
    }),

    getByType: t.procedure.input(z.object({
        type: z.enum(['session', 'working', 'long_term'])
    })).query(async ({ input }) => {
        const server = getMcpServer();
        return server.agentMemoryService.getByType(input.type);
    }),

    getByNamespace: t.procedure.input(z.object({
        namespace: z.enum(['user', 'agent', 'project'])
    })).query(async ({ input }) => {
        const server = getMcpServer();
        return server.agentMemoryService.getByNamespace(input.namespace);
    }),

    delete: t.procedure.input(z.object({
        id: z.string()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return server.agentMemoryService.delete(input.id);
    }),

    clearSession: t.procedure.mutation(async () => {
        const server = getMcpServer();
        server.agentMemoryService.clearSession();
        return { success: true };
    }),

    export: t.procedure.query(async () => {
        const server = getMcpServer();
        return server.agentMemoryService.export();
    }),

    handoff: t.procedure.input(z.object({
        notes: z.string().optional()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return await server.agentMemoryService.handoffSession(input);
    }),

    pickup: t.procedure.input(z.object({
        artifact: z.string()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return await server.agentMemoryService.pickupSession(input.artifact);
    }),

    stats: t.procedure.query(async () => {
        const server = getMcpServer();
        const memories = server.agentMemoryService;
        return {
            session: memories.getByType('session').length,
            working: memories.getByType('working').length,
            longTerm: memories.getByType('long_term').length,
            total: memories.getByType('session').length + memories.getByType('working').length + memories.getByType('long_term').length
        };
    })
});
