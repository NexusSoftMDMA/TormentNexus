import { z } from 'zod';
import { t, getMcpServer } from '../lib/trpc-core.js';
import { resolveOrchestratorBase } from '../lib/tormentnexus-orchestrator.js';

interface SupervisorTaskRuntime {
    status?: 'pending' | 'active' | 'completed' | 'failed' | string;
}

export const supervisorRouter = t.router({
    decompose: t.procedure.input(z.object({
        goal: z.string()
    })).mutation(async ({ input }) => {
        const server = getMcpServer();
        return server.supervisor.decompose(input.goal);
    }),

    supervise: t.procedure.input(z.object({
        goal: z.string(),
        maxSteps: z.number().default(20)
    })).mutation(async ({ input }) => {
        const orchestratorBase = resolveOrchestratorBase();

        if (orchestratorBase) {
            try {
                const res = await fetch(`${orchestratorBase}/api/sessions/start`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        task: { description: input.goal },
                        workingDirectory: process.cwd()
                    })
                });
                if (res.ok) {
                    const sessionData = await res.json();
                    const session = sessionData.data;

                    return {
                        success: true,
                        sessionId: session.id,
                        message: "Goal successfully delegated to TormentNexus Orchestrator."
                    };
                }
            } catch (e: any) {
                console.warn(`[Supervisor] TormentNexus Orchestrator unavailable (${e.message}). Falling back to native supervisor.`);
            }
        }

        const server = getMcpServer();
        return server.supervisor.supervise(input.goal, input.maxSteps);
    }),

    /** Get current supervisor status - active tasks, workers, queue depth */
    status: t.procedure.query(async () => {
        let orchestratorActive = 0;
        const orchestratorBase = resolveOrchestratorBase();
        if (orchestratorBase) {
            try {
                const res = await fetch(`${orchestratorBase}/api/health`);
                if (res.ok) {
                    const data = await res.json();
                    orchestratorActive = data.sessions?.active || 0;
                }
            } catch {
                // Keep native supervisor status truthful even when the external orchestrator is unavailable.
            }
        }

        const server = getMcpServer();
        const sup = server?.supervisor;
        return {
            isActive: sup?.isActive?.() ?? false,
            activeWorkers: (sup?.getActiveWorkers?.() ?? []).concat(orchestratorActive > 0 ? [`Orchestrator (${orchestratorActive})`] : []),
            queueDepth: sup?.getQueueDepth?.() ?? 0,
            lastActivity: sup?.getLastActivity?.() ?? null,
            totalTasksCompleted: sup?.getTotalCompleted?.() ?? 0,
        };
    }),

    /** List all tasks managed by the supervisor - recent history */
    listTasks: t.procedure.input(z.object({
        limit: z.number().min(1).max(100).default(20),
        status: z.enum(['all', 'pending', 'active', 'completed', 'failed']).default('all'),
    }).optional()).query(async ({ input }) => {
        let orchestratorTasks: any[] = [];
        const orchestratorBase = resolveOrchestratorBase();
        if (orchestratorBase) {
            try {
                const res = await fetch(`${orchestratorBase}/api/sessions`);
                if (res.ok) {
                    const data = await res.json();
                    orchestratorTasks = data.map((d: any) => ({
                        id: d.id,
                        description: d.currentTask || 'Orchestrator Session',
                        status: d.status === 'running' ? 'active' : d.status
                    }));
                }
            } catch { }
        }

        const server = getMcpServer();
        const sup = server.supervisor;
        let allTasks = sup.listTasks?.() ?? [];
        allTasks = [...allTasks, ...orchestratorTasks];

        const filtered = input?.status === 'all'
            ? allTasks
            : allTasks.filter((t: unknown) => (t as SupervisorTaskRuntime).status === input?.status);
        return filtered.slice(0, input?.limit ?? 20);
    }),

    /** Cancel a running supervised task */
    cancel: t.procedure.input(z.object({
        taskId: z.string()
    })).mutation(async ({ input }) => {
        const orchestratorBase = resolveOrchestratorBase();
        if (input.taskId.startsWith('session-')) {
            if (orchestratorBase) {
                try {
                    const res = await fetch(`${orchestratorBase}/api/sessions/${input.taskId}/stop`, { method: 'POST' });
                    if (res.ok) return { success: true, taskId: input.taskId };
                } catch { }
            }
        }

        const server = getMcpServer();
        const sup = server.supervisor;
        if (sup.cancelTask) {
            await sup.cancelTask(input.taskId);
            return { success: true, taskId: input.taskId };
        }
        return { success: false, error: 'Cancel not supported' };
    }),
});
