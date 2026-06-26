import type { Context, Next } from 'hono';

const API_KEY = process.env.API_KEY;

export async function apiKeyAuth(c: Context, next: Next): Promise<Response | void> {
  if (!API_KEY) {
    return next();
  }
  
  const authHeader = c.req.header('Authorization');
  const apiKeyHeader = c.req.header('X-API-Key');
  
  const providedKey = apiKeyHeader || (authHeader?.startsWith('Bearer ') ? authHeader.slice(7) : null);
  
  if (!providedKey) {
    return c.json({ error: 'Missing API key. Provide via Authorization: Bearer <key> or X-API-Key header' }, 401);
  }
  
  if (providedKey !== API_KEY) {
    return c.json({ error: 'Invalid API key' }, 403);
  }
  
  return next();
}
