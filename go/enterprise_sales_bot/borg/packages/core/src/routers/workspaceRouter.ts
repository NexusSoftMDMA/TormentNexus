import { z } from 'zod';
import { t, publicProcedure } from '../lib/trpc-core.js';
import { workspaceTracker } from '../services/WorkspaceTracker.js';

export const workspaceRouter = t.router({
    list: publicProcedure.query(async () => {
        return await workspaceTracker.listWorkspaces();
    }),
    register: publicProcedure.mutation(async () => {
        await workspaceTracker.registerWorkspace();
        return { success: true };
    }),
});
