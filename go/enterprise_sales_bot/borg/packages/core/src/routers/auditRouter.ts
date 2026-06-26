import { z } from 'zod';
import { t, getAuditService } from '../lib/trpc-core.js';

export const auditRouter = t.router({
    list: t.procedure.input(z.object({ limit: z.number().optional() })).query(async ({ input }) => {
        return getAuditService().getLogs(input.limit || 50);
    }),
    log: t.procedure.input(z.object({
        level: z.string(), agentId: z.string().optional(), action: z.string(), limit: z.number().optional()
    })).query(async ({ input }) => {
        return getAuditService().query(input as Record<string, unknown>);
    })
});
