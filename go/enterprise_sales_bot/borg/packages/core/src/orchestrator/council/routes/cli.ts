import { Hono } from 'hono';
import type { ApiResponse, CLITool } from './types.js';
import { cliRegistry } from '../services/cli-registry.js';

const cli = new Hono();

cli.get('/tools', async (c) => {
  const tools = await cliRegistry.detectAll();
  return c.json<ApiResponse<CLITool[]>>({ success: true, data: tools });
});

cli.get('/tools/available', async (c) => {
  await cliRegistry.detectAll();
  const tools = cliRegistry.getAvailableTools();
  return c.json<ApiResponse<CLITool[]>>({ success: true, data: tools });
});

cli.post('/tools/refresh', async (c) => {
  const tools = await cliRegistry.refreshDetection();
  return c.json<ApiResponse<CLITool[]>>({ success: true, data: tools });
});

cli.get('/tools/:type', async (c) => {
  await cliRegistry.detectAll();
  const type = c.req.param('type') as any;
  const tool = cliRegistry.getTool(type);
  
  if (!tool) {
    return c.json<ApiResponse<null>>({ success: false, error: `CLI tool '${type}' not found` }, 404);
  }
  
  return c.json<ApiResponse<CLITool>>({ success: true, data: tool });
});

cli.post('/tools/custom', async (c) => {
  try {
    const body = await c.req.json();
    const tool: CLITool = {
      type: 'custom',
      name: body.name,
      command: body.command,
      args: body.args || [],
      healthEndpoint: body.healthEndpoint || '/health',
      available: true,
      capabilities: body.capabilities || [],
    };
    
    cliRegistry.registerCustomTool(tool);
    return c.json<ApiResponse<CLITool>>({ success: true, data: tool });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to register custom CLI tool';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 400);
  }
});

cli.delete('/tools/custom/:name', (c) => {
  const name = c.req.param('name');
  const removed = cliRegistry.unregisterCustomTool(name);
  
  if (!removed) {
    return c.json<ApiResponse<null>>({ success: false, error: `Custom tool '${name}' not found` }, 404);
  }
  
  return c.json<ApiResponse<{ removed: boolean }>>({ success: true, data: { removed: true } });
});

export { cli as cliRoutes };
