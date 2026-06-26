import { z } from 'zod';
import { t, publicProcedure, getHealerService } from '../lib/trpc-core.js';
import { observable } from '@trpc/server/observable';

import type { AnyTRPCRouter } from '@trpc/server';
import type * as _TRPCCore from '@trpc/server/unstable-core-do-not-import';

const SIDECAR_URL = process.env.TORMENTNEXUS_SIDECAR_URL || 'http://127.0.0.1:4300';

export const healerRouter = t.router({
    diagnose: t.procedure.input(z.object({ error: z.string(), context: z.string().optional() })).mutation(async ({ input }) => {
        try {
            const resp = await fetch(`${SIDECAR_URL}/api/native/healer/diagnose`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ error: input.error, context: input.context || "" })
            });
            return await resp.json();
        } catch (e) {
            console.error("[HealerRouter] ❌ Failed to call native healer diagnose:", e);
            return getHealerService().analyzeError(input.error, input.context || "");
        }
    }),
    heal: t.procedure.input(z.object({ error: z.string(), context: z.string().optional() })).mutation(async ({ input }) => {
        try {
            const resp = await fetch(`${SIDECAR_URL}/api/native/healer/heal`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ error: input.error, context: input.context || "" })
            });
            const data = await resp.json() as any;
            return { success: data.success || false };
        } catch (e) {
            console.error("[HealerRouter] ❌ Failed to call native healer heal:", e);
            // Fallback to TS (which might be broken as noted during research, but keeping original path as last resort)
            const success = await (getHealerService() as any).heal?.(input.error, input.context || "") ?? false;
            return { success };
        }
    }),
    getHistory: t.procedure.query(async () => {
        try {
            const resp = await fetch(`${SIDECAR_URL}/api/native/healer/history`);
            const data = await resp.json() as any;
            return data.history || [];
        } catch (e) {
            console.error("[HealerRouter] ❌ Failed to fetch native healer history:", e);
            try { return getHealerService()?.getHistory() ?? []; } catch { return []; }
        }
    }),
    vaultRecords: publicProcedure.input(z.object({ limit: z.number().optional() }).optional()).query(async ({ input }) => {
        const limit = input?.limit || 50;
        try {
            const res = await fetch(`${SIDECAR_URL}/api/native/healer/vault?limit=${limit}`);
            if (res.ok) {
                const json = await res.json();
                if (json.records && Array.isArray(json.records)) {
                    // Map Go snake_case to TS UI PascalCase
                    return json.records.map((r: any) => ({
                        id: r.id,
                        SessionID: r.session_id,
                        Type: r.memory_type,
                        Content: r.content,
                        Importance: r.importance,
                        HeatScore: r.heat_score,
                        LastAccessedAt: r.last_accessed_at,
                        CreatedAt: r.created_at
                    }));
                }
            }
        } catch (e) {
            console.warn('[healerRouter] Failed to fetch vault records from sidecar, trying fallback:', e);
        }

        try {
            // Fallback for getting records if go backend isn't mapped
            const service = getHealerService() as any;
            if (service && typeof service.getVaultRecords === 'function') {
                return service.getVaultRecords();
            }
        } catch { }
        return [];
    }),
    vaultRecordCount: t.procedure.query(async () => {
        try {
            const resp = await fetch(`${SIDECAR_URL}/api/native/healer/vault?limit=0`);
            const data = await resp.json() as any;
            return data.totalCount || 0;
        } catch (e) {
            console.error("[HealerRouter] ❌ Failed to fetch native healer vault count:", e);
            return 0;
        }
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
            // In TormentNexus, either the server or generic system holds event bus. We use MCPServer fallback.
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
