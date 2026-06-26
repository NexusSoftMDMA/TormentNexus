import { Hono } from 'hono';
import type { ApiResponse, SessionHealth } from './types.js';
import { sessionManager } from '../services/session-manager.js';
import { healthMonitor } from '../services/health-monitor.js';
import { logRotation } from '../services/log-rotation.js';

const health = new Hono();

health.get('/sessions', (c) => {
  const allHealth = sessionManager.getAllSessionHealth();
  const data = Object.fromEntries(allHealth);
  return c.json<ApiResponse<Record<string, SessionHealth>>>({ success: true, data });
});

health.get('/sessions/:id', async (c) => {
  const id = c.req.param('id');
  const sessionHealth = sessionManager.getSessionHealth(id);
  
  if (!sessionHealth) {
    return c.json<ApiResponse<null>>({ success: false, error: `Session '${id}' not found` }, 404);
  }
  
  return c.json<ApiResponse<SessionHealth>>({ success: true, data: sessionHealth });
});

health.post('/sessions/:id/check', async (c) => {
  const id = c.req.param('id');
  const sessionHealth = await healthMonitor.forceCheck(id);
  
  if (!sessionHealth) {
    return c.json<ApiResponse<null>>({ success: false, error: `Session '${id}' not found` }, 404);
  }
  
  return c.json<ApiResponse<SessionHealth>>({ success: true, data: sessionHealth });
});

health.get('/stats', (c) => {
  const sessionStats = sessionManager.getSessionStats();
  const logStats = logRotation.getStats();
  const allHealth = sessionManager.getAllSessionHealth();
  
  let healthy = 0;
  let degraded = 0;
  let unresponsive = 0;
  let crashed = 0;
  
  for (const health of allHealth.values()) {
    switch (health.status) {
      case 'healthy': healthy++; break;
      case 'degraded': degraded++; break;
      case 'unresponsive': unresponsive++; break;
      case 'crashed': crashed++; break;
    }
  }
  
  return c.json<ApiResponse<{
    sessions: typeof sessionStats;
    logs: typeof logStats;
    health: { healthy: number; degraded: number; unresponsive: number; crashed: number };
  }>>({
    success: true,
    data: {
      sessions: sessionStats,
      logs: logStats,
      health: { healthy, degraded, unresponsive, crashed },
    },
  });
});

health.get('/server', (c) => {
  const uptime = process.uptime();
  const memory = process.memoryUsage();
  
  return c.json<ApiResponse<{
    status: 'healthy';
    uptime: number;
    memory: { heapUsed: number; heapTotal: number; rss: number };
  }>>({
    success: true,
    data: {
      status: 'healthy',
      uptime,
      memory: {
        heapUsed: memory.heapUsed,
        heapTotal: memory.heapTotal,
        rss: memory.rss,
      },
    },
  });
});

export { health as healthRoutes };
