import { z } from 'zod';
import { t, publicProcedure, getCommandRegistry } from '../lib/trpc-core.js';

export const commandsRouter = t.router({
    execute: publicProcedure.input(z.object({
        input: z.string()
    })).mutation(async ({ input }) => {
        const result = await getCommandRegistry().execute(input.input);
        if (result) {
            return {
                handled: result.handled,
                output: result.output,
                error: (result as { error?: unknown }).error
            };
        }
        return { handled: false, output: 'Not a command', error: null };
    }),

    list: publicProcedure.query(() => {
        try {
        const commands = getCommandRegistry()?.getCommands?.() ?? [];
        return commands.map((cmd) => ({
                name: cmd.name,
                description: cmd.description
        }));
        } catch { return []; }
    }),
});

