import { z } from 'zod';
import { t, getDarwinService } from '../lib/trpc-core.js';

export const darwinRouter = t.router({
    evolve: t.procedure.input(z.object({ prompt: z.string(), goal: z.string() })).mutation(async ({ input }) => {
        return getDarwinService().proposeMutation(input.prompt, input.goal);
    }),
    experiment: t.procedure.input(z.object({ mutationId: z.string(), task: z.string() })).mutation(async ({ input }) => {
        const exp = await getDarwinService().startExperiment(input.mutationId, input.task);
        return { experimentId: exp.id };
    }),
    getStatus: t.procedure.query(async () => {
        return getDarwinService().getStatus();
    })
});
