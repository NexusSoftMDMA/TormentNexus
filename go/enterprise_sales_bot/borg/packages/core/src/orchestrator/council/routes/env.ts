import { Hono } from 'hono';
import type { ApiResponse, CLIType } from './types.js';
import { environmentManager } from '../services/environment-manager.js';

const env = new Hono();

env.get('/sessions/:id', (c) => {
  const id = c.req.param('id');
  const sanitized = environmentManager.getSanitizedEnvironment(id);
  
  if (Object.keys(sanitized).length === 0) {
    return c.json<ApiResponse<null>>({ success: false, error: `Session '${id}' not found` }, 404);
  }
  
  return c.json<ApiResponse<Record<string, string>>>({ success: true, data: sanitized });
});

env.post('/sessions/:id', async (c) => {
  const id = c.req.param('id');
  const body = await c.req.json();
  
  if (body.key && body.value !== undefined) {
    environmentManager.updateSessionVariable(id, body.key, body.value);
    return c.json<ApiResponse<{ updated: boolean }>>({ success: true, data: { updated: true } });
  }
  
  return c.json<ApiResponse<null>>({ success: false, error: 'Missing key or value' }, 400);
});

env.delete('/sessions/:id/:key', (c) => {
  const id = c.req.param('id');
  const key = c.req.param('key');
  
  environmentManager.removeSessionVariable(id, key);
  return c.json<ApiResponse<{ removed: boolean }>>({ success: true, data: { removed: true } });
});

env.get('/required/:cliType', (c) => {
  const cliType = c.req.param('cliType') as CLIType;
  const required = environmentManager.getRequiredVarsForCLI(cliType);
  return c.json<ApiResponse<string[]>>({ success: true, data: required });
});

env.post('/validate/:cliType', async (c) => {
  const cliType = c.req.param('cliType') as CLIType;
  const body = await c.req.json();
  const env = body.env || {};
  const result = environmentManager.validateEnvironmentForCLI(cliType, env);
  return c.json<ApiResponse<{ valid: boolean; missing: string[] }>>({ success: true, data: result });
});

env.post('/global', async (c) => {
  const body = await c.req.json();
  
  if (body.key && body.value !== undefined) {
    environmentManager.setGlobalOverride(body.key, body.value);
    return c.json<ApiResponse<{ set: boolean }>>({ success: true, data: { set: true } });
  }
  
  return c.json<ApiResponse<null>>({ success: false, error: 'Missing key or value' }, 400);
});

env.delete('/global/:key', (c) => {
  const key = c.req.param('key');
  environmentManager.removeGlobalOverride(key);
  return c.json<ApiResponse<{ removed: boolean }>>({ success: true, data: { removed: true } });
});

env.post('/secrets', async (c) => {
  const body = await c.req.json();
  
  if (body.key) {
    environmentManager.addGlobalSecret(body.key);
    return c.json<ApiResponse<{ added: boolean }>>({ success: true, data: { added: true } });
  }
  
  return c.json<ApiResponse<null>>({ success: false, error: 'Missing key' }, 400);
});

export { env as envRoutes };
