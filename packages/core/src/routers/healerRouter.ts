import { z } from 'zod';
import { t, publicProcedure, getHealerService } from '../lib/trpc-core.js';
import { observable } from '@trpc/server/observable';

import type { AnyTRPCRouter } from '@trpc/server';
import type * as _TRPCCore from '@trpc/server/unstable-core-do-not-import';

export const healerRouter = t.router({
    diagnose: t.procedure.input(z.object({ error: z.string(), context: z.string().optional() })).mutation(async ({ input }) => {
        return getHealerService().analyzeError(input.error, input.context || "");
    }),
    heal: t.procedure.input(z.object({ error: z.string(), context: z.string().optional() })).mutation(async ({ input }) => {
        const success = await getHealerService().heal(input.error, input.context || "");
        return { success };
    }),
    getHistory: t.procedure.query(async () => {
        try { return getHealerService()?.getHistory() ?? []; } catch { return []; }
    }),
    vaultRecords: t.procedure.query(async () => {
        try {
            // Fallback for getting records if go backend isn't mapped
            const service = getHealerService() as any;
            if (service && typeof service.getVaultRecords === 'function') {
                return service.getVaultRecords();
            }
            return [];
        } catch { return []; }
    }),
    subscribe: publicProcedure.subscription(() => {
        return observable<unknown>((emit) => {
            const onHeal = (data: unknown) => {
                emit.next(data);
            };
            const service = getHealerService();
            service.on('heal', onHeal);
            return () => {
                service.off('heal', onHeal);
            };
        });
    }),
    subscribeQuotaEvents: publicProcedure.subscription(() => {
        return observable<unknown>((emit) => {
            const onEvent = (eventData: unknown) => {
                emit.next(eventData);
            };
            // Safely fetch internal EventBus from active Orchestrator Context
            const server = getHealerService() as any; 
            // In Borg, either the server or generic system holds event bus. We use MCPServer fallback.
            const mcpServer = (global as any).mcpServerInstance;
            if (mcpServer && mcpServer.eventBus) {
                mcpServer.eventBus.on('system:llm_quota_exhausted', onEvent);
                mcpServer.eventBus.on('system:llm_fallback', onEvent);
                mcpServer.eventBus.on('system:healer_halted', onEvent);
                return () => {
                    mcpServer.eventBus.off('system:llm_quota_exhausted', onEvent);
                    mcpServer.eventBus.off('system:llm_fallback', onEvent);
                    mcpServer.eventBus.off('system:healer_halted', onEvent);
                };
            }
            return () => {}; // No-op teardown if missing context
        });
    })
});
