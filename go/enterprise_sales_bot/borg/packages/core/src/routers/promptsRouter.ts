import { z } from 'zod';
import { t, publicProcedure, getMcpServer } from '../lib/trpc-core.js';
import type { PromptTemplate } from '../prompts/PromptRegistry.js';

export const promptsRouter = t.router({
    list: publicProcedure.query(async (): Promise<PromptTemplate[]> => {
        const server = getMcpServer();
        if (!server.promptRegistry) return [];
        return server.promptRegistry.list();
    }),

    get: publicProcedure.input(z.object({ id: z.string() })).query(async ({ input }): Promise<PromptTemplate | null> => {
        const server = getMcpServer();
        if (!server.promptRegistry) return null;
        return server.promptRegistry.get(input.id) ?? null;
    }),

    save: publicProcedure.input(z.object({
        id: z.string(),
        description: z.string(),
        template: z.string(),
    })).mutation(async ({ input }): Promise<{ success: boolean }> => {
        const server = getMcpServer();
        if (!server.promptRegistry) throw new Error("Prompt registry not initialized");

        const extractTemplateVariables = (template: string): string[] => {
            const names = new Set<string>();
            const doubleBracePattern = /\{\{\s*([a-zA-Z_][a-zA-Z0-9_.-]*)\s*\}\}/g;
            for (const match of template.matchAll(doubleBracePattern)) {
                if (match[1]) names.add(match[1].trim());
            }
            const dollarBracePattern = /\$\{\s*([a-zA-Z_][a-zA-Z0-9_.-]*)\s*\}/g;
            for (const match of template.matchAll(dollarBracePattern)) {
                if (match[1]) names.add(match[1].trim());
            }
            return Array.from(names).sort((a, b) => a.localeCompare(b));
        };

        const existing = server.promptRegistry.get(input.id);
        const version = existing ? existing.version + 1 : 1;
        const variables = extractTemplateVariables(input.template);

        await server.promptRegistry.save({
            id: input.id,
            version,
            description: input.description,
            template: input.template,
            variables,
            updatedAt: new Date().toISOString()
        });

        return { success: true };
    }),

    delete: publicProcedure.input(z.object({ id: z.string() })).mutation(async ({ input }): Promise<{ success: boolean }> => {
        const server = getMcpServer();
        if (!server.promptRegistry) throw new Error("Prompt registry not initialized");
        await server.promptRegistry.delete(input.id);
        return { success: true };
    }),
});
