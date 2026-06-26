import { z } from 'zod';
import { t, getMcpServer } from '../lib/trpc-core.js';

export const planRouter = t.router({
    getMode: t.procedure.query(async () => {
        const server = getMcpServer();
        return { mode: server.planService.getMode() };
    }),

    setMode: t.procedure.input(z.object({
        mode: z.enum(['PLAN', 'BUILD'])
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        if (input.mode === 'PLAN') {
            server.planService.enterPlanMode();
        } else {
            server.planService.enterBuildMode();
        }
        return { mode: input.mode };
    }),

    getDiffs: t.procedure.query(async () => {
        const server = getMcpServer();
        return server.planService.sandbox.getPendingDiffs();
    }),

    approveDiff: t.procedure.input(z.object({
        diffId: z.string()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return server.planService.sandbox.approveDiff(input.diffId);
    }),

    rejectDiff: t.procedure.input(z.object({
        diffId: z.string()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return server.planService.sandbox.rejectDiff(input.diffId);
    }),

    applyAll: t.procedure.mutation(async () => {
        const server = getMcpServer();
        return server.planService.sandbox.applyAllApproved();
    }),

    getSummary: t.procedure.query(async () => {
        const server = getMcpServer();
        return server.planService.sandbox.getSummary();
    }),

    getCheckpoints: t.procedure.query(async () => {
        const server = getMcpServer();
        return server.planService.sandbox.getCheckpoints();
    }),

    createCheckpoint: t.procedure.input(z.object({
        name: z.string(),
        description: z.string().optional()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return server.planService.sandbox.createCheckpoint(input.name, input.description);
    }),

    rollback: t.procedure.input(z.object({
        checkpointId: z.string()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return server.planService.sandbox.rollbackToCheckpoint(input.checkpointId);
    }),

    clear: t.procedure.mutation(async () => {
        const server = getMcpServer();
        server.planService.sandbox.clear();
        return { success: true };
    })
});
