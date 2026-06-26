import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { toolSetsRepository } from '../db/repositories/index.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

const CreateToolSetInput = z.object({
    name: z.string(),
    description: z.string().nullable().optional(),
    tools: z.array(z.string()),
    user_id: z.string().nullable().optional(),
});

const UpdateToolSetInput = z.object({
    uuid: z.string(),
    name: z.string().optional(),
    description: z.string().nullable().optional(),
    tools: z.array(z.string()).optional(),
    user_id: z.string().nullable().optional(),
});

export const toolSetsRouter = t.router({
    list: publicProcedure.query(async () => {
        try {
            return await toolSetsRepository.findAll();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Tool sets are unavailable', error);
        }
    }),

    get: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .query(async ({ input }) => {
            try {
                return await toolSetsRepository.findByUuid(input.uuid);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool sets are unavailable', error);
            }
        }),

    create: adminProcedure
        .input(CreateToolSetInput)
        .mutation(async ({ input }) => {
            try {
                // @ts-expect-error -- type mismatch from inferred schema types
                return await toolSetsRepository.create(input);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool sets are unavailable', error);
            }
        }),

    update: adminProcedure
        .input(UpdateToolSetInput)
        .mutation(async ({ input }) => {
            try {
                // @ts-expect-error -- type mismatch from inferred schema types
                const updated = await toolSetsRepository.update(input);
                if (!updated) {
                    throw new Error("Tool set not found");
                }
                return updated;
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool sets are unavailable', error);
            }
        }),

    delete: adminProcedure
        .input(z.object({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            try {
                await toolSetsRepository.deleteByUuid(input.uuid);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Tool sets are unavailable', error);
            }
        }),
});
