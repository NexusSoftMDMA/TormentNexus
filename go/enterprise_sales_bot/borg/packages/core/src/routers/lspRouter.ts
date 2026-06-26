import { z } from 'zod';
import { t, getLspService } from '../lib/trpc-core.js';

function requireLspService() {
    const lsp = getLspService();
    if (!lsp) {
        throw new Error('LSP service unavailable');
    }
    return lsp;
}

export const lspRouter = t.router({
    findSymbol: t.procedure.input(z.object({
        filePath: z.string(),
        symbolName: z.string()
    })).query(async ({ input }) => {
        return requireLspService().findSymbol(input.filePath, input.symbolName);
    }),

    findReferences: t.procedure.input(z.object({
        filePath: z.string(),
        line: z.number(),
        character: z.number()
    })).query(async ({ input }) => {
        return requireLspService().findReferences(input.filePath, input.line, input.character);
    }),

    getSymbols: t.procedure.input(z.object({
        filePath: z.string()
    })).query(async ({ input }) => {
        return requireLspService().getSymbols(input.filePath);
    }),

    searchSymbols: t.procedure.input(z.object({
        query: z.string()
    })).query(async ({ input }) => {
        return requireLspService().searchSymbols(input.query);
    }),

    indexProject: t.procedure.mutation(async () => {
        await requireLspService().indexProject();
        return { success: true };
    })
});
