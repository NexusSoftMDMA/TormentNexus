import { z } from 'zod';
import { t, publicProcedure, getMcpServer } from '../lib/trpc-core.js';
import type { ToolDefinition } from '../services/CodeModeService.js';

export const codeModeRouter = t.router({
    /** Returns current Code Mode status, registered tool count, and context reduction stats. */
    getStatus: publicProcedure.query(() => {
        const server = getMcpServer();
        const svc = server.codeModeService;
        const enabled = svc.isEnabled();
        const registry = svc.getRegistry();
        const reduction = svc.calculateContextReduction();
        const tools = registry.getAll().map((tool: ToolDefinition) => ({
            name: tool.name,
            description: tool.description,
        }));
        return {
            enabled,
            toolCount: tools.length,
            tools,
            reduction: {
                traditional: reduction.traditional,
                codeMode: reduction.codeMode,
                reductionPct: reduction.reduction,
            },
        };
    }),

    /** Enable Code Mode — allows LLMs to call tools via executable code rather than JSON schemas. */
    enable: publicProcedure.mutation(() => {
        getMcpServer().codeModeService.enable();
        return { ok: true, enabled: true };
    }),

    /** Disable Code Mode — reverts to standard JSON schema tool calling. */
    disable: publicProcedure.mutation(() => {
        getMcpServer().codeModeService.disable();
        return { ok: true, enabled: false };
    }),

    /** Execute arbitrary code against the registered Code Mode SDK (escape hatch). */
    execute: publicProcedure
        .input(z.object({ code: z.string().min(1), context: z.record(z.unknown()).optional() }))
        .mutation(async ({ input }) => {
            const svc = getMcpServer().codeModeService;
            if (!svc.isEnabled()) {
                return { success: false, error: 'Code Mode is not enabled. Enable it first.' } as const;
            }
            const result = await svc.executeCode(input.code, input.context ?? {});
            return result;
        }),
});
