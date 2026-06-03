import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { t, publicProcedure, getMcpServer } from '../lib/trpc-core.js';
import fs from 'fs/promises';
import path from 'path';

export const projectRouter = t.router({
    getContext: publicProcedure.query(async () => {
        const contextPath = path.join(process.cwd(), '.tormentnexus', 'project_context.md');
        try {
            if (!(await fs.access(contextPath).then(() => true).catch(() => false))) {
                return "# Project Context\n\nDefine your repository rules and architectural vision here.";
            }
            return await fs.readFile(contextPath, 'utf-8');
        } catch (e) {
            const errorCode = (e as NodeJS.ErrnoException | undefined)?.code;
            if (errorCode === 'ENOENT') {
                return "# Project Context\n\nDefine your repository rules and architectural vision here.";
            }
            const message = e instanceof Error ? e.message : String(e);
            throw new TRPCError({
                code: 'INTERNAL_SERVER_ERROR',
                message: `Project context is unavailable: ${message}`,
            });
        }
    }),

    updateContext: publicProcedure
        .input(z.object({ content: z.string() }))
        .mutation(async ({ input }) => {
            const contextPath = path.join(process.cwd(), '.tormentnexus', 'project_context.md');
            const tormentnexusDir = path.dirname(contextPath);
            
            try {
                await fs.mkdir(tormentnexusDir, { recursive: true });
                await fs.writeFile(contextPath, input.content, 'utf-8');
                return { success: true };
            } catch (e: any) {
                throw new Error(`Failed to save project context: ${e.message}`);
            }
        }),
        
    getHandoffs: publicProcedure.query(async () => {
        const handoffDir = path.join(process.cwd(), '.tormentnexus', 'handoffs');
        try {
            if (!(await fs.access(handoffDir).then(() => true).catch(() => false))) {
                return [];
            }
            const files = await fs.readdir(handoffDir);
            return files
                .filter(f => f.startsWith('handoff_'))
                .sort((a, b) => b.localeCompare(a))
                .map(f => ({
                    id: f,
                    timestamp: parseInt(f.replace('handoff_', '').replace('.json', '')),
                    path: f
                }));
        } catch (e) {
            const errorCode = (e as NodeJS.ErrnoException | undefined)?.code;
            if (errorCode === 'ENOENT') {
                return [];
            }
            const message = e instanceof Error ? e.message : String(e);
            throw new TRPCError({
                code: 'INTERNAL_SERVER_ERROR',
                message: `Project handoffs are unavailable: ${message}`,
            });
        }
    })
});
