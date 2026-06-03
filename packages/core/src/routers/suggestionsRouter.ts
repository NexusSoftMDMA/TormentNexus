import { z } from 'zod';
import { t, publicProcedure, getMcpServer, getSuggestionService } from '../lib/trpc-core.js';

export const suggestionsRouter = t.router({
    list: publicProcedure.query(() => {
        try { return getSuggestionService()?.getPendingSuggestions() ?? []; } catch { return []; }
    }),
    resolve: publicProcedure.input(z.object({
        id: z.string(),
        status: z.enum(['APPROVED', 'REJECTED'])
    })).mutation(async ({ input }) => {
        const suggestionService = getSuggestionService();
        const suggestion = suggestionService.resolveSuggestion(input.id, input.status);

        // EXECUTION LOGIC FOR APPROVED ACTIONS
        if (suggestion && input.status === 'APPROVED' && suggestion.payload && suggestion.payload.tool) {
            console.log(`[TormentNexus Core] Auto-Executing Approved Suggestion: ${suggestion.title}`);
            await getMcpServer().executeTool(suggestion.payload.tool, suggestion.payload.args || {});
        }
        return suggestion ?? null;
    }),
    clearAll: publicProcedure.mutation(() => {
        getSuggestionService().clearAll?.();
        return true;
    })
});

