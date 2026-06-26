import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { quotaManager } from '../../orchestrator/council/services/quota-manager.js';

export const quotaRouter = t.router({
  status: publicProcedure.query(async () => {
    return quotaManager.getStatus();
  }),

  getConfig: publicProcedure.query(async () => {
    return quotaManager.getConfig();
  }),

  updateConfig: adminProcedure.input(z.any()).mutation(async ({ input }) => {
    quotaManager.setConfig(input);
    return { success: true, config: quotaManager.getConfig() };
  }),

  enable: adminProcedure.mutation(async () => {
    quotaManager.setEnabled(true);
    return { success: true, enabled: true };
  }),

  disable: adminProcedure.mutation(async () => {
    quotaManager.setEnabled(false);
    return { success: true, enabled: false };
  }),

  check: publicProcedure.input(z.object({ provider: z.string() })).query(async ({ input }) => {
    return quotaManager.checkQuota(input.provider);
  }),

  allStats: publicProcedure.query(async () => {
    return quotaManager.getAllStats();
  }),

  providerStats: publicProcedure.input(z.object({ provider: z.string() })).query(async ({ input }) => {
    return quotaManager.getProviderStats(input.provider);
  }),

  setLimits: adminProcedure.input(z.object({
    provider: z.string(),
    limits: z.any(),
  })).mutation(async ({ input }) => {
    quotaManager.setProviderLimits(input.provider, input.limits);
    return { success: true, limits: quotaManager.getLimits(input.provider) };
  }),

  getLimits: publicProcedure.input(z.object({ provider: z.string() })).query(async ({ input }) => {
    return quotaManager.getLimits(input.provider);
  }),

  recordRequest: adminProcedure.input(z.object({
    provider: z.string(),
    tokensUsed: z.number().optional(),
    latencyMs: z.number().optional(),
    success: z.boolean().optional(),
  })).mutation(async ({ input }) => {
    quotaManager.recordRequest(input.provider, input.tokensUsed || 0, input.latencyMs || 0, input.success ?? true);
    return { success: true };
  }),

  recordRateLimitError: adminProcedure.input(z.object({ provider: z.string() })).mutation(async ({ input }) => {
    quotaManager.recordRateLimitError(input.provider);
    return { success: true };
  }),

  unthrottle: adminProcedure.input(z.object({ provider: z.string() })).mutation(async ({ input }) => {
    quotaManager.unthrottleProvider(input.provider);
    return { success: true };
  }),

  resetProvider: adminProcedure.input(z.object({ provider: z.string() })).mutation(async ({ input }) => {
    quotaManager.resetProviderUsage(input.provider);
    return { success: true };
  }),

  resetAll: adminProcedure.mutation(async () => {
    quotaManager.resetAllUsage();
    return { success: true };
  }),
});
