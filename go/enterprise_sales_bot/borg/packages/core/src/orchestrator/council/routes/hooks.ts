import { Hono } from 'hono';
import { z } from 'zod';
import { autoContinueHooks, type HookPhase, type HookHandler } from '../services/hooks.js';
import { apiRateLimit } from '../middleware/rate-limit.js';
import { apiKeyAuth } from '../middleware/auth.js';
import { hookRegisterSchema } from '../schemas.js';

export const hooksRoutes = new Hono();

const dynamicHooks: Map<string, HookHandler> = new Map();

const registerBodySchema = z.object({
  phase: z.enum(['pre-debate', 'post-debate', 'pre-guidance', 'post-guidance', 'on-error']),
  webhookUrl: z.string().url(),
  priority: z.number().int().min(0).max(100).optional(),
});

hooksRoutes.get('/', apiRateLimit(), (c) => {
  return c.json({
    hooks: autoContinueHooks.getRegisteredHooks(),
  });
});

hooksRoutes.post('/register', apiRateLimit(), apiKeyAuth, async (c) => {
  const { phase, webhookUrl, priority } = registerBodySchema.parse(await c.req.json());

  const handler: HookHandler = async (context) => {
    try {
      const res = await fetch(webhookUrl, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(context),
      });
      
      if (!res.ok) {
        return { continue: true };
      }
      
      return await res.json();
    } catch {
      return { continue: true };
    }
  };

  const id = autoContinueHooks.register(phase, handler, priority);
  dynamicHooks.set(id, handler);
  
  return c.json({ success: true, hookId: id });
});

hooksRoutes.delete('/:id', apiRateLimit(), apiKeyAuth, (c) => {
  const id = c.req.param('id');
  const removed = autoContinueHooks.unregister(id);
  
  if (removed) {
    dynamicHooks.delete(id);
  }
  
  return c.json({ success: removed });
});

hooksRoutes.delete('/', apiRateLimit(), apiKeyAuth, (c) => {
  autoContinueHooks.clear();
  dynamicHooks.clear();
  return c.json({ success: true });
});
