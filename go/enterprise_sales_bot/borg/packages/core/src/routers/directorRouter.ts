import { z } from 'zod';
import {
    t,
    adminProcedure,
    getMcpServer,
    getDirectorRuntime,
    getMemoryManager,
    getCommandRegistry,
    getSuggestionService,
} from '../lib/trpc-core.js';

export const directorRouter = t.router({
    memorize: t.procedure.input(z.object({ content: z.string(), source: z.string(), title: z.string().optional() })).mutation(async ({ input }) => {
        await getMemoryManager().saveContext(input.content, {
            source: input.source,
            title: input.title || 'Untitled Web Page',
            type: 'web_page'
        });
        return "Memorized.";
    }),
    chat: t.procedure.input(z.object({ message: z.string() })).mutation(async ({ input }) => {
        const server = getMcpServer();
        const director = getDirectorRuntime();

        // 1. Intercept Slash Commands
        if (input.message.trim().startsWith('/')) {
            const commandResult = await getCommandRegistry().execute(input.message);
            if (commandResult && commandResult.handled) {
                return commandResult.output;
            }
        }

        // 2. Intercept "Yes" / "Approve" for Suggestions
        const suggestions = getSuggestionService();
        const pending = suggestions.getPendingSuggestions();
        if (pending.length > 0 && /^(yes|approve|do it|confirm|ok)$/i.test(input.message.trim())) {
            const latest = pending[0];
            const suggestion = suggestions.resolveSuggestion(latest.id, 'APPROVED');

            if (suggestion && suggestion.payload?.tool) {
                director.broadcast?.(`✅ Approved: **${latest.title}**. Executing ${suggestion.payload.tool}...`);
                const result = await server.executeTool(suggestion.payload.tool, suggestion.payload.args ?? {});
                return `✅ Execution Complete.\n\nResult:\n${JSON.stringify(result)?.substring(0, 200)}...`;
            }

            return `✅ Approved suggestion: **${latest.title}**. (No tool attached)`;
        }

        // 3. Default: Director Execution
        const result = await director.executeTask?.(input.message);
        return result;
    }),
    status: t.procedure.query(() => {
        try { return getDirectorRuntime()?.getStatus?.() ?? { status: 'offline' }; } catch { return { status: 'offline' }; }
    }),
    updateConfig: t.procedure.input(z.object({
        defaultTopic: z.string().optional()
    })).mutation(({ input }) => {
        getDirectorRuntime().updateConfig?.(input as Record<string, unknown>);
        return { success: true };
    }),
    stopAutoDrive: adminProcedure.mutation(async () => {
        getDirectorRuntime().stopAutoDrive?.();
        return "Stopped";
    }),
    startAutoDrive: adminProcedure.mutation(async () => {
        getMcpServer().executeTool('start_auto_drive', {});
        return "Started";
    })
});
