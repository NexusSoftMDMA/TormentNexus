import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { apiKeysRepository } from '../db/repositories/index.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';
import {
    ApiKeyCreateInputSchema,
    ApiKeyUpdateInputSchema
} from '../types/tormentnexus/api-keys.zod.js';

export const apiKeysRouter = t.router({
    list: publicProcedure.query(async () => {
        try {
            return await apiKeysRepository.findAll();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('API key store is unavailable', error);
        }
    }),

    get: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .query(async ({ input }) => {
            try {
                return await apiKeysRepository.findByUuidWithAccess(input.uuid, 'system');
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('API key store is unavailable', error);
            }
        }),

    create: adminProcedure
        .input(ApiKeyCreateInputSchema)
        .mutation(async ({ input }) => {
            try {
                return await apiKeysRepository.create({ ...input, user_id: 'system' });
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('API key store is unavailable', error);
            }
        }),

    update: adminProcedure
        .input(ApiKeyUpdateInputSchema.extend({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            try {
                return await apiKeysRepository.update(input.uuid, 'system', input);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('API key store is unavailable', error);
            }
        }),

    delete: adminProcedure
        .input(z.object({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            try {
                return await apiKeysRepository.delete(input.uuid, 'system');
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('API key store is unavailable', error);
            }
        }),

    validate: publicProcedure
        .input(z.object({ key: z.string() }))
        .mutation(async ({ input }) => {
            try {
                return await apiKeysRepository.validateApiKey(input.key);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('API key store is unavailable', error);
            }
        }),
});
