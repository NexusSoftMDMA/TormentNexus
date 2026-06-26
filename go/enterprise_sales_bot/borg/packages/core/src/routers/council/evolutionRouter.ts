import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../../lib/trpc-core.js';
import { selfEvolution } from '../../orchestrator/council/services/self-evolution.js';

export const evolutionRouter = t.router({
  start: adminProcedure.mutation(async () => {
    selfEvolution.start();
    return { success: true, message: 'Continuous learning started' };
  }),

  stop: adminProcedure.mutation(async () => {
    selfEvolution.stop();
    return { success: true, message: 'Continuous learning stopped' };
  }),

  optimize: adminProcedure.mutation(async () => {
    selfEvolution.optimizeWeights();
    return { success: true };
  }),

  evolve: adminProcedure.input(z.object({ description: z.string() })).mutation(async ({ input }) => {
    const sessionId = await selfEvolution.evolveSystem(input.description);
    return { success: true, sessionId };
  }),

  test: adminProcedure.query(async () => {
    const passed = await selfEvolution.runSelfTest();
    return { success: true, passed };
  }),
});
