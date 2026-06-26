import { z } from 'zod';
import { t, getPermissionManager, getDirectorRuntime } from '../lib/trpc-core.js';

export const autonomyRouter = t.router({
    setLevel: t.procedure.input(z.object({ level: z.enum(['low', 'medium', 'high']) })).mutation(async ({ input }) => {
        getPermissionManager().setAutonomyLevel(input.level);
        return input.level;
    }),
    getLevel: t.procedure.query(() => {
        return getPermissionManager().autonomyLevel;
    }),
    activateFullAutonomy: t.procedure.mutation(async () => {
        getPermissionManager().setAutonomyLevel('high');
        const director = getDirectorRuntime();
        director.startChatDaemon?.();
        director.startWatchdog?.(100);
        return "Autonomous Supervisor Activated (High Level + Chat Daemon + Watchdog)";
    })
});
