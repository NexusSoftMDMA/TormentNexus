import { TRPCError } from '@trpc/server';
import { z } from 'zod';
import { t, getGitService } from '../lib/trpc-core.js';

export const gitRouter = t.router({
    getModules: t.procedure.query(async () => {
        const fs = await import('fs/promises');
        const path = await import('path');
        try {
            const gitModulesPath = path.join(process.cwd(), '.gitmodules');
            const content = await fs.readFile(gitModulesPath, 'utf-8');
            const modules = [];
            const regex = /\[submodule "(.*?)"\]\s*path = (.*?)\s*url = (.*?)\s/g;
            let match;
            while ((match = regex.exec(content)) !== null) {
                modules.push({
                    name: match[1],
                    path: match[2],
                    url: match[3],
                    status: 'unknown',
                    branch: 'main',
                    lastCommit: 'HEAD',
                    date: new Date().toISOString().split('T')[0],
                    active: false
                });
            }
            if (modules.length === 0 && content.trim().length > 0) {
                throw new Error('Malformed .gitmodules file');
            }
            return modules;
        } catch (e) {
            const errorCode = (e as NodeJS.ErrnoException | undefined)?.code;
            if (errorCode === 'ENOENT') {
                return [];
            }
            const message = e instanceof Error ? e.message : String(e);
            console.error('Failed to read .gitmodules', e);
            throw new TRPCError({
                code: 'INTERNAL_SERVER_ERROR',
                message: `Git modules are unavailable: ${message}`,
            });
        }
    }),
    getLog: t.procedure.input(z.object({ limit: z.number().optional() })).query(async ({ input }) => {
        return getGitService().getLog(input.limit);
    }),
    getStatus: t.procedure.query(async () => {
        return getGitService().getStatus();
    }),
    revert: t.procedure.input(z.object({ hash: z.string() })).mutation(async ({ input }) => {
        return getGitService().revert(input.hash);
    })
});
