import { Hono, type Context } from 'hono';
import { z } from 'zod';
import { collectiveMemory } from '../services/collective-memory.js';
import { apiRateLimit } from '../middleware/rate-limit.js';

const app = new Hono();

function handleMemoryRouteError(c: Context, error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  if (message.includes('Collective memory is unavailable:')) {
    return c.json({ success: false, error: message }, 503);
  }

  throw error;
}

app.get('/facts', apiRateLimit(), async (c) => {
  try {
    const query = c.req.query('q');
    let facts;
    
    if (query) {
      facts = await collectiveMemory.searchFacts(query);
    } else {
      facts = await collectiveMemory.getAllFacts();
    }
    
    return c.json({ success: true, data: facts });
  } catch (error) {
    return handleMemoryRouteError(c, error);
  }
});

app.post('/facts', apiRateLimit(), async (c) => {
  try {
    const body = factSchema.parse(await c.req.json());
    // @ts-expect-error -- type mismatch from inferred schema types
    const fact = await collectiveMemory.storeFact({
      ...body,
      tags: body.tags || []
    });
    return c.json({ success: true, data: fact });
  } catch (error) {
    return handleMemoryRouteError(c, error);
  }
});

app.get('/recall/:key', apiRateLimit(), async (c) => {
  try {
    const key = c.req.param('key');
    const facts = await collectiveMemory.recallFact(key);
    return c.json({ success: true, data: facts });
  } catch (error) {
    return handleMemoryRouteError(c, error);
  }
});

export default app;
const factSchema = z.object({
  key: z.string(),
  value: z.string(),
  sourceSession: z.string(),
  confidence: z.number().min(0).max(1),
  tags: z.array(z.string()).optional(),
});
