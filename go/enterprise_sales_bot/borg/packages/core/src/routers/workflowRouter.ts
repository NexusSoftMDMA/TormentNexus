import { z } from 'zod';
import { TRPCError } from '@trpc/server';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { getWorkflowEngine, getWorkflowDefinitions } from '../lib/trpc-core.js';
import { visualWorkflowsRepo } from '../db/repositories/visual-workflows.repo.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

function requireWorkflowEngine() {
    const engine = getWorkflowEngine();
    if (!engine) {
        throw new TRPCError({
            code: 'INTERNAL_SERVER_ERROR',
            message: 'Workflow engine is unavailable for this run.',
        });
    }
    return engine;
}

export const workflowRouter = t.router({
    // --- Visual Canvas DB State ---
    saveCanvas: adminProcedure
        .input(z.object({
            id: z.string().optional(),
            name: z.string(),
            description: z.string().optional(),
            nodes: z.array(z.any()),
            edges: z.array(z.any()),
        }))
        .mutation(async ({ input }) => {
            try {
                if (input.id) {
                    await visualWorkflowsRepo.updateWorkflow(input.id, {
                        name: input.name,
                        description: input.description,
                        nodes: input.nodes,
                        edges: input.edges,
                    });
                    return { id: input.id };
                } else {
                    // @ts-expect-error -- type mismatch from inferred schema types
                    const res = await visualWorkflowsRepo.createWorkflow(input);
                    return { id: res.id };
                }
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Workflow canvas storage is unavailable', error);
            }
        }),

    loadCanvas: adminProcedure
        .input(z.object({ id: z.string() }))
        .query(async ({ input }) => {
            try {
                return await visualWorkflowsRepo.getWorkflow(input.id);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Workflow canvas storage is unavailable', error);
            }
        }),

    listCanvases: adminProcedure
        .query(async () => {
            try {
                return await visualWorkflowsRepo.listWorkflows();
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Workflow canvas storage is unavailable', error);
            }
        }),

    // --- Workflow Definitions ---

    list: publicProcedure.query(() => {
        const engine = requireWorkflowEngine();
        return getWorkflowDefinitions(engine).map((w) => ({
            id: w.id,
            name: w.name ?? w.id,
            description: w.description ?? '',
            entryPoint: w.entryPoint,
            nodeCount: w.nodes?.size ?? 0,
            edgeCount: w.edges?.length ?? 0,
        }));
    }),

    getGraph: publicProcedure
        .input(z.object({ workflowId: z.string() }))
        .query(({ input }) => {
            const engine = requireWorkflowEngine();
            const graph = engine.getGraph(input.workflowId);
            if (!graph) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: `Workflow "${input.workflowId}" was not found.`,
                });
            }
            return graph;
        }),

    // --- Executions ---

    start: adminProcedure
        .input(z.object({
            workflowId: z.string(),
            initialState: z.record(z.unknown()).optional()
        }))
        .mutation(async ({ input }) => {
            const engine = requireWorkflowEngine();
            const execution = await engine.start(input.workflowId, input.initialState || {});
            return execution;
        }),

    listExecutions: publicProcedure.query(() => {
        const engine = requireWorkflowEngine();
        return engine.listExecutions();
    }),

    getExecution: publicProcedure
        .input(z.object({ executionId: z.string() }))
        .query(({ input }) => {
            const engine = requireWorkflowEngine();
            return engine.getExecution(input.executionId);
        }),

    getHistory: publicProcedure
        .input(z.object({ executionId: z.string() }))
        .query(({ input }) => {
            const engine = requireWorkflowEngine();
            return engine.getHistory(input.executionId);
        }),

    // --- Control Flow ---

    resume: adminProcedure
        .input(z.object({ executionId: z.string() }))
        .mutation(async ({ input }) => {
            const engine = requireWorkflowEngine();
            await engine.resume(input.executionId);
            return { success: true };
        }),

    pause: adminProcedure
        .input(z.object({ executionId: z.string() }))
        .mutation(({ input }) => {
            const engine = requireWorkflowEngine();
            engine.pause(input.executionId);
            return { success: true };
        }),

    approve: adminProcedure
        .input(z.object({ executionId: z.string() }))
        .mutation(async ({ input }) => {
            const engine = requireWorkflowEngine();
            await engine.approve(input.executionId);
            return { success: true };
        }),

    reject: adminProcedure
        .input(z.object({
            executionId: z.string(),
            reason: z.string().optional()
        }))
        .mutation(({ input }) => {
            const engine = requireWorkflowEngine();
            engine.reject(input.executionId, input.reason);
            return { success: true };
        })
});

