import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { autoContinueHooks } from '../../orchestrator/council/services/hooks.js';

const hookPhaseSchema = z.enum(['pre-debate', 'post-debate', 'pre-guidance', 'post-guidance', 'on-error']);

export const hooksRouter = t.router({
  list: publicProcedure.query(async () => {
    return {
      hooks: autoContinueHooks.getRegisteredHooks(),
    };
  }),

  register: adminProcedure.input(z.object({
    phase: hookPhaseSchema,
    webhookUrl: z.string().url(),
    priority: z.number().int().min(0).max(100).optional(),
  })).mutation(async ({ input }) => {
    const handler = async (context: any) => {
      try {
        const res = await fetch(input.webhookUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(context),
        });
        
        if (!res.ok) return { continue: true };
        return await res.json();
      } catch {
        return { continue: true };
      }
    };

    const id = autoContinueHooks.register(input.phase as any, handler, input.priority);
    return { success: true, hookId: id };
  }),

  unregister: adminProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }) => {
    const success = autoContinueHooks.unregister(input.id);
    return { success };
  }),

  clear: adminProcedure.mutation(async () => {
    autoContinueHooks.clear();
    return { success: true };
  }),
});
