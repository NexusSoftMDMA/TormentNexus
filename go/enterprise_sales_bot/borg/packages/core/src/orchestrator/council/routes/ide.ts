import { Hono } from 'hono';
import { z } from 'zod';
import { smartPilot } from '../services/smart-pilot.js';
import { sessionManager } from '../services/session-manager.js';

export const ideRoutes = new Hono();

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

ideRoutes.get('/status', (c) => {
  return c.json({
    success: true,
    ready: true,
    version: '1.0.12',
    capabilities: ['council', 'smart-pilot', 'quota'],
  });
});

ideRoutes.post('/task', async (c) => {
  const { description, fileContext, sessionId } = taskSchema.parse(await c.req.json());

  // Find or create session
  let session = sessionId ? sessionManager.getSession(sessionId) : sessionManager.getActiveSessions()[0];

  if (!session) {
    // Auto-create session for IDE if none exists
    session = await sessionManager.startSession(undefined, {
        cliType: 'opencode', // Default for now
        tags: ['ide-generated']
    });
  }

  // Inject task into Smart Pilot
  // We need to expose a method in Smart Pilot to manually trigger a task with context
  // For now, we'll simulate it via the session manager or direct processing
  // But SmartPilot polls. Let's add a `triggerTask` method to SmartPilotService.

  const task = {
    id: `ide-${Date.now()}`,
    description,
    context: fileContext ? JSON.stringify(fileContext) : '',
    files: fileContext ? [fileContext.path] : [],
  };

  await smartPilot.triggerTask(session.id, task);

  return c.json({
    success: true,
    taskId: task.id,
    sessionId: session.id,
    message: 'Task submitted to Council',
  });
});

export default ideRoutes;
