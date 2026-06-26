import type { Context, Next } from 'hono';
import { metrics } from '../services/metrics.js';

export async function metricsMiddleware(c: Context, next: Next): Promise<Response | void> {
  const start = Date.now();
  
  await next();
  
  const latencyMs = Date.now() - start;
  const statusCode = c.res.status;
  const method = c.req.method;
  const path = c.req.path;
  
  metrics.recordHttpRequest(method, path, statusCode, latencyMs);
}
