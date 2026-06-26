import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { smartPilot } from '../../orchestrator/council/services/smart-pilot.js';

const smartPilotConfigSchema = z.object({
  enabled: z.boolean().optional(),
  pollIntervalMs: z.number().optional(),
  autoApproveThreshold: z.number().optional(),
  requireUnanimous: z.boolean().optional(),
  maxAutoApprovals: z.number().optional(),
});

export const smartPilotRouter = t.router({
  status: publicProcedure.query(async () => {
    return {
      enabled: smartPilot.isEnabled(),
      config: smartPilot.getConfig(),
      activePlans: Array.from(smartPilot.getActivePlans().entries()).map(([sessionId, plan]) => ({
        sessionId,
        plan,
      })),
    };
  }),

  getConfig: publicProcedure.query(async () => {
    return smartPilot.getConfig();
  }),

  updateConfig: adminProcedure.input(smartPilotConfigSchema).mutation(async ({ input }) => {
    if (input.enabled !== undefined) smartPilot.setEnabled(input.enabled);
    if (input.autoApproveThreshold !== undefined) smartPilot.setAutoApproveThreshold(input.autoApproveThreshold);
    if (input.requireUnanimous !== undefined) smartPilot.setRequireUnanimous(input.requireUnanimous);
    if (input.maxAutoApprovals !== undefined) smartPilot.setMaxAutoApprovals(input.maxAutoApprovals);
    
    return smartPilot.getConfig();
  }),

  trigger: adminProcedure.input(z.object({
    sessionId: z.string(),
    task: z.any(),
  })).mutation(async ({ input }) => {
    await smartPilot.triggerTask(input.sessionId, input.task);
    return { success: true };
  }),

  resetCount: adminProcedure.input(z.object({ sessionId: z.string() })).mutation(async ({ input }) => {
    smartPilot.resetApprovalCount(input.sessionId);
    return { success: true };
  }),

  resetAllCounts: adminProcedure.mutation(async () => {
    smartPilot.resetAllApprovalCounts();
    return { success: true };
  }),
});
