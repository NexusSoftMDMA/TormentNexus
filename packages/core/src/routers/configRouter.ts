import { z } from 'zod';
import path from 'node:path';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { configRepo } from '../db/repositories/index.js';
import { configService } from '../services/config.service.js';
import { jsonConfigProvider } from '../services/config/JsonConfigProvider.js';
import { getTormentNexusConfigDir, writeTormentNexusMcpConfig } from '../mcp/mcpJsonConfig.js';
import { rethrowSqliteUnavailableAsTrpc } from './sqliteTrpc.js';

const ConfigValueSchema = z.object({
    id: z.string(),
    value: z.string(),
    description: z.string().optional(),
});

export const configRouter = t.router({
    list: publicProcedure
        .output(z.array(z.object({ key: z.string(), value: z.string() })))
        .query(async () => {
            try {
                const configRecord = await configRepo.findAll();
                return Object.entries(configRecord).map(([key, value]) => ({ key, value }));
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    get: publicProcedure
        .input(z.object({ key: z.string() }))
        .query(async ({ input }) => {
            try {
                return await configRepo.get(input.key);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    upsert: adminProcedure
        .input(ConfigValueSchema)
        .mutation(async ({ input }) => {
            try {
                return await configRepo.set(input.id, input.value);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    delete: adminProcedure
        .input(z.object({ key: z.string() }))
        .mutation(async ({ input }) => {
            try {
                return await configRepo.delete(input.key);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    update: adminProcedure
        .input(z.object({ key: z.string(), value: z.string() }))
        .mutation(async ({ input }) => {
            try {
                return await configRepo.set(input.key, input.value);
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    reset: adminProcedure
        .input(z.object({ section: z.string().optional() }))
        .mutation(async ({ input }) => {
            try {
                const configRecord = await configRepo.findAll();
                const keys = Object.keys(configRecord)
                    .filter((key) => !input.section || key === input.section || key.startsWith(`${input.section}.`));

                for (const key of keys) {
                    await configRepo.delete(key);
                }

                return {
                    success: true,
                    section: input.section ?? null,
                    removed: keys.length,
                    keys,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    init: adminProcedure
        .input(z.object({ scope: z.enum(['global', 'local']).default('local') }))
        .mutation(async ({ input }) => {
            try {
                const configDir = input.scope === 'global'
                    ? getTormentNexusConfigDir()
                    : path.join(process.cwd(), '.tormentnexus');

                await writeTormentNexusMcpConfig({ mcpServers: {} }, configDir);

                return {
                    success: true,
                    scope: input.scope,
                    path: configDir,
                };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getMcpTimeout: publicProcedure.query(async () => {
        try {
            return await configService.getMcpTimeout();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setMcpTimeout: adminProcedure
        .input(z.object({ timeout: z.number().min(1000).max(86400000) }))
        .mutation(async ({ input }) => {
            try {
                await configService.setMcpTimeout(input.timeout);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getMcpMaxAttempts: publicProcedure.query(async () => {
        try {
            return await configService.getMcpMaxAttempts();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setMcpMaxAttempts: adminProcedure
        .input(z.object({ maxAttempts: z.number().min(1).max(10) }))
        .mutation(async ({ input }) => {
            try {
                await configService.setMcpMaxAttempts(input.maxAttempts);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getMcpMaxTotalTimeout: publicProcedure.query(async () => {
        try {
            return await configService.getMcpMaxTotalTimeout();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setMcpMaxTotalTimeout: adminProcedure
        .input(z.object({ timeout: z.number().min(1000).max(86400000) }))
        .mutation(async ({ input }) => {
            try {
                await configService.setMcpMaxTotalTimeout(input.timeout);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getMcpResetTimeoutOnProgress: publicProcedure.query(async () => {
        try {
            return await configService.getMcpResetTimeoutOnProgress();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setMcpResetTimeoutOnProgress: adminProcedure
        .input(z.object({ enabled: z.boolean() }))
        .mutation(async ({ input }) => {
            try {
                await configService.setMcpResetTimeoutOnProgress(input.enabled);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getSessionLifetime: publicProcedure.query(async () => {
        try {
            return await configService.getSessionLifetime();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setSessionLifetime: adminProcedure
        .input(z.object({ lifetime: z.number().min(300000).max(86400000).nullable().optional() }))
        .mutation(async ({ input }) => {
            try {
                await configService.setSessionLifetime(input.lifetime);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getSignupDisabled: publicProcedure.query(async () => {
        try {
            return await configService.isSignupDisabled();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setSignupDisabled: adminProcedure
        .input(z.object({ disabled: z.boolean() }))
        .mutation(async ({ input }) => {
            try {
                await configService.setSignupDisabled(input.disabled);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getSsoSignupDisabled: publicProcedure.query(async () => {
        try {
            return await configService.isSsoSignupDisabled();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setSsoSignupDisabled: adminProcedure
        .input(z.object({ disabled: z.boolean() }))
        .mutation(async ({ input }) => {
            try {
                await configService.setSsoSignupDisabled(input.disabled);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getBasicAuthDisabled: publicProcedure.query(async () => {
        try {
            return await configService.isBasicAuthDisabled();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    setBasicAuthDisabled: adminProcedure
        .input(z.object({ disabled: z.boolean() }))
        .mutation(async ({ input }) => {
            try {
                await configService.setBasicAuthDisabled(input.disabled);
                return { success: true };
            } catch (error) {
                rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
            }
        }),

    getAuthProviders: publicProcedure.query(async () => {
        try {
            return await configService.getAuthProviders();
        } catch (error) {
            rethrowSqliteUnavailableAsTrpc('Configuration store is unavailable', error);
        }
    }),

    getAlwaysVisibleTools: publicProcedure.query(async () => {
        return await jsonConfigProvider.loadAlwaysVisibleTools();
    }),

    setAlwaysVisibleTools: adminProcedure
        .input(z.object({ tools: z.array(z.string()) }))
        .mutation(async ({ input }) => {
            const tools = await jsonConfigProvider.saveAlwaysVisibleTools(input.tools);
            return { success: true, tools };
        }),
});
