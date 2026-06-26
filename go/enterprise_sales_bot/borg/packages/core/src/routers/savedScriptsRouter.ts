import { z } from 'zod';
import { t, publicProcedure, adminProcedure } from '../lib/trpc-core.js';
import { jsonConfigProvider } from '../services/config/JsonConfigProvider.js';
import { codeExecutorService } from '../services/CodeExecutorService.js';
import { SavedScriptConfig } from '../interfaces/IConfigProvider.js';

// Define input schemas here if not exported from Zod file, or reuse existing
const CreateScriptInput = z.object({
    name: z.string(),
    description: z.string().nullable().optional(),
    code: z.string(),
    userId: z.string().nullable().optional(),
});

const UpdateScriptInput = z.object({
    uuid: z.string(),
    name: z.string().optional(),
    description: z.string().nullable().optional(),
    code: z.string().optional(),
});

export const savedScriptsRouter = t.router({
    list: publicProcedure.query(async () => {
        return await jsonConfigProvider.loadScripts();
    }),

    get: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .query(async ({ input }) => {
            const scripts = await jsonConfigProvider.loadScripts();
            return scripts.find(s => s.uuid === input.uuid);
        }),

    create: publicProcedure
        .input(CreateScriptInput)
        .mutation(async ({ input }) => {
            const script: SavedScriptConfig = {
                name: input.name,
                code: input.code,
                description: input.description,
                // userId is not stored in local config currently
            };
            return await jsonConfigProvider.saveScript(script);
        }),

    update: publicProcedure
        .input(UpdateScriptInput)
        .mutation(async ({ input }) => {
            const scripts = await jsonConfigProvider.loadScripts();
            const existing = scripts.find(s => s.uuid === input.uuid);
            if (!existing) throw new Error("Script not found");

            const updated: SavedScriptConfig = {
                ...existing,
                name: input.name ?? existing.name,
                description: input.description ?? existing.description,
                code: input.code ?? existing.code,
            };

            return await jsonConfigProvider.saveScript(updated);
        }),

    delete: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            await jsonConfigProvider.deleteScript(input.uuid);
            return { success: true };
        }),

    execute: publicProcedure
        .input(z.object({ uuid: z.string() }))
        .mutation(async ({ input }) => {
            const scripts = await jsonConfigProvider.loadScripts();
            const script = scripts.find(s => s.uuid === input.uuid);

            if (!script) throw new Error("Script not found");

            const startedAt = Date.now();

            try {
                const output = await codeExecutorService.executeCode(script.code);
                const finishedAt = Date.now();

                return {
                    success: true,
                    result: output,
                    execution: {
                        scriptUuid: script.uuid,
                        scriptName: script.name,
                        startedAt: new Date(startedAt).toISOString(),
                        finishedAt: new Date(finishedAt).toISOString(),
                        durationMs: finishedAt - startedAt,
                    },
                };
            } catch (error: unknown) {
                const message = error instanceof Error ? error.message : String(error);
                const finishedAt = Date.now();

                return {
                    success: false,
                    result: null,
                    error: message,
                    execution: {
                        scriptUuid: script.uuid,
                        scriptName: script.name,
                        startedAt: new Date(startedAt).toISOString(),
                        finishedAt: new Date(finishedAt).toISOString(),
                        durationMs: finishedAt - startedAt,
                    },
                };
            }
        }),
});
