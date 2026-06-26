import { z } from 'zod';
import { t, publicProcedure } from '../../lib/trpc-core.js';
import { smartPilot } from '../../orchestrator/council/services/smart-pilot.js';
import { sessionManager } from '../../orchestrator/council/services/session-manager.js';

const taskSchema = z.object({
  description: z.string(),
  fileContext: z.object({
    path: z.string(),
    content: z.string(),
    selection: z.object({
      start: z.number(),
      end: z.number(),
    }).optional(),
    cursor: z.number().optional(),
  }).optional(),
  sessionId: z.string().optional(),
});

export const ideRouter = t.router({
  status: publicProcedure.query(async () => {
    return {
      success: true,
      ready: true,
      version: '2.1.0',
      capabilities: ['council', 'smart-pilot', 'quota', 'trpc'],
    };
  }),

  submitTask: publicProcedure.input(taskSchema).mutation(async ({ input }) => {
    let session = input.sessionId ? sessionManager.getSession(input.sessionId) : sessionManager.getActiveSessions()[0];

    if (!session) {
      session = await sessionManager.startSession(undefined, {
          cliType: 'claude-code',
          tags: ['ide-generated']
      });
    }

    const task = {
      id: `ide-${Date.now()}`,
      description: input.description,
      context: input.fileContext ? JSON.stringify(input.fileContext) : '',
      files: input.fileContext ? [input.fileContext.path] : [],
    };

    await smartPilot.triggerTask(session.id, task);

    return {
      success: true,
      taskId: task.id,
      sessionId: session.id,
      message: 'Task submitted to Council',
    };
  }),
});
