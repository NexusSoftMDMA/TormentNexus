import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { policiesRepository } from '../db/repositories/index.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';
import {
    CreatePolicySchema,
    UpdatePolicySchema,
    DeletePolicySchema
} from '../types/tormentnexus/policies.zod.js';

export const policiesRouter = t.router({
    list: publicProcedure.query(async () => {
        try {
            return await policiesRepository.findAll();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Policies are unavailable', error);
        }
    }),

    get: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .query(async ({ input }) => {
            try {
                return await policiesRepository.findByUuid(input.uuid);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Policies are unavailable', error);
            }
        }),

    create: adminProcedure
        .input(CreatePolicySchema)
        .mutation(async ({ input }) => {
            try {
                return await policiesRepository.create(input);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Policies are unavailable', error);
            }
        }),

    update: adminProcedure
        .input(UpdatePolicySchema)
        .mutation(async ({ input }) => {
            try {
                return await policiesRepository.update(input);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Policies are unavailable', error);
            }
        }),

    delete: adminProcedure
        .input(DeletePolicySchema)
        .mutation(async ({ input }) => {
            try {
                await policiesRepository.delete(input.uuid);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Policies are unavailable', error);
            }
        }),
});
