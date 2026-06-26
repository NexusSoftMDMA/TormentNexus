import { Hono } from 'hono';
import { z } from 'zod';
import { diagramService } from '../services/diagram-service.js';
import { smartPilot } from '../services/smart-pilot.js';
import { sessionManager } from '../services/session-manager.js';
import { cliRegistry } from '../services/cli-registry.js';
import { apiRateLimit } from '../middleware/rate-limit.js';

const app = new Hono();

app.get('/swarm/:sessionId', apiRateLimit(), (c) => {
  const sessionId = c.req.param('sessionId');
  const plans = smartPilot.getActivePlans();
  const plan = plans.get(sessionId);

  if (!plan) {
    return c.json({ success: false, error: 'No active plan for session' }, 404);
  }

  const mermaid = diagramService.generateSwarmMermaid(plan);
  return c.json({ success: true, mermaid });
});

app.get('/system', apiRateLimit(), (c) => {
  const availableTools = cliRegistry.getAvailableTools().map(t => t.type);
  const activeSessions = sessionManager.getActiveSessions().map(s => ({
    id: s.id,
    cliType: sessionManager.getSessionCLIType(s.id) || 'unknown'
  }));

  const mermaid = diagramService.generateSystemMermaid(availableTools, activeSessions);
  return c.json({ success: true, mermaid });
});

app.post('/plan-from-mermaid', async (c) => {
  const { mermaid } = mermaidSchema.parse(await c.req.json());
  const plan = diagramService.parseMermaidToPlan(mermaid);
  return c.json({ success: true, plan });
});

export default app;
const mermaidSchema = z.object({
  mermaid: z.string().min(10)
});
