import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { serveStatic } from 'hono/bun';
import { HTTPException } from 'hono/http-exception';
import { sessionRoutes } from './routes/sessions.js';
import { councilRoutes } from './routes/council.js';
import { wsRoutes } from './routes/ws.js';
import { smartPilotRoutes } from './routes/smart-pilot.js';
import { hooksRoutes } from './routes/hooks.js';
import { cliRoutes } from './routes/cli.js';
import { healthRoutes } from './routes/health.js';
import { envRoutes } from './routes/env.js';
import { dynamicSelectionRoutes } from './routes/dynamic-selection.js';
import { vetoRoutes } from './routes/veto.js';
import { pluginRoutes } from './routes/plugins.js';
import { debateHistoryRoutes } from './routes/debate-history.js';
import quotaRoutes from './routes/quota.js';
import supervisorAnalyticsRoutes from './routes/supervisor-analytics.js';
import debateSimulatorRoutes from './routes/debate-simulator.js';
import workspaceRoutes from './routes/workspace.js';
import templateRoutes from './routes/templates.js';
import fineTunedModelsRoutes from './routes/fine-tuned-models.js';
import collaborativeDebatesRoutes from './routes/collaborative-debates.js';
import systemRoutes from './routes/system.js';
import visualRoutes from './routes/visual.js';
import memoryRoutes from './routes/memory.js';
import logsRoutes from './routes/logs.js';
import ideRoutes from './routes/ide.js';
import { loadConfig } from './services/config.js';
import { council } from './services/council.js';
import { createSupervisors } from './supervisors/index.js';
import { sessionManager } from './services/session-manager.js';
import { smartPilot } from './services/smart-pilot.js';
import { autoContinueHooks } from './services/hooks.js';
import { metrics } from './services/metrics.js';
import { metricsMiddleware } from './middleware/metrics.js';
import { healthMonitor } from './services/health-monitor.js';
import { logRotation } from './services/log-rotation.js';
import { debateHistory } from './services/debate-history.js';
import { quotaManager } from './services/quota-manager.js';
import { selfEvolution } from './services/self-evolution.js';
import { checkpointService } from './services/checkpoint-service.js';
import { councilHierarchy } from './services/council-hierarchy.js';
import { autonomousMaintenance } from './services/system-maintenance.js';
import { collectiveMemory } from './services/collective-memory.js';

const startTime = Date.now();
const config = loadConfig();

if (config.council.supervisors.length > 0) {
  const supervisors = createSupervisors(config.council.supervisors);
  for (const supervisor of supervisors) {
    council.addSupervisor(supervisor);
  }
  console.log(`Loaded ${supervisors.length} supervisor(s) from config`);
}

await councilHierarchy.initialize(council);

council.setDebateRounds(config.council.debateRounds || 2);
council.setConsensusThreshold(config.council.consensusThreshold || 0.7);

const app = new Hono();

app.use('*', logger());
app.use('*', metricsMiddleware);

app.use('*', cors({
  origin: config.server.corsOrigins || '*',
  allowMethods: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
  allowHeaders: ['Content-Type', 'Authorization', 'X-Requested-With'],
  exposeHeaders: ['X-RateLimit-Limit', 'X-RateLimit-Remaining', 'X-RateLimit-Reset'],
  maxAge: 86400,
  credentials: true,
}));

app.onError((err, c) => {
  console.error(`[ERROR] ${c.req.method} ${c.req.path}:`, err.message);
  
  if (err instanceof HTTPException) {
    return c.json({ 
      success: false, 
      error: err.message 
    }, err.status);
  }
  
  if (err.message?.includes('validation')) {
    return c.json({ 
      success: false, 
      error: 'Validation error', 
      details: err.message 
    }, 400);
  }
  
  return c.json({ 
    success: false, 
    error: 'Internal server error' 
  }, 500);
});

app.get('/', (c) => c.json({ 
  name: 'cli-orchestrator', 
  version: '0.2.0',
  config: {
    supervisors: council.getSupervisors().map(s => s.name),
    debateRounds: config.council.debateRounds,
    consensusThreshold: config.council.consensusThreshold,
    enabled: config.council.enabled,
    smartPilot: smartPilot.isEnabled(),
  },
  endpoints: {
    sessions: '/api/sessions',
    council: '/api/council',
    smartPilot: '/api/smart-pilot',
    hooks: '/api/hooks',
    cli: '/api/cli',
    sessionHealth: '/api/health',
    env: '/api/env',
    dynamicSelection: '/api/dynamic-selection',
    veto: '/api/veto',
    plugins: '/api/plugins',
    debateHistory: '/api/debate-history',
    quota: '/api/quota',
    analytics: '/api/analytics',
    simulator: '/api/simulator',
    workspaces: '/api/workspaces',
    templates: '/api/templates',
    fineTunedModels: '/api/fine-tuned-models',
    collaborativeDebates: '/api/collaborative-debates',
    websocket: '/ws',
    health: '/health',
    dashboard: '/dashboard',
    analyticsDashboard: '/analytics',
  }
}));

app.get('/health', async (c) => {
  const supervisors = council.getSupervisors();
  const availableSupervisors = await council.getAvailableSupervisors();
  const sessions = sessionManager.getAllSessions();
  const activeSessions = sessions.filter((s: { status: string }) => s.status === 'running');
  const hooks = autoContinueHooks.getRegisteredHooks();
  
  const uptimeSeconds = Math.floor((Date.now() - startTime) / 1000);
  const memoryUsage = process.memoryUsage();
  const metricsSummary = metrics.getSummary();
  
  const status = availableSupervisors.length > 0 || supervisors.length === 0 ? 'healthy' : 'degraded';
  
  return c.json({
    status,
    timestamp: Date.now(),
    uptime: {
      seconds: uptimeSeconds,
      formatted: formatUptime(uptimeSeconds),
    },
    supervisors: {
      total: supervisors.length,
      available: availableSupervisors.length,
      names: supervisors.map(s => s.name),
    },
    sessions: {
      total: sessions.length,
      active: activeSessions.length,
    },
    council: {
      enabled: config.council.enabled,
      debateRounds: config.council.debateRounds,
      consensusThreshold: config.council.consensusThreshold,
      weightedVoting: config.council.weightedVoting,
    },
    smartPilot: {
      enabled: smartPilot.isEnabled(),
      config: smartPilot.getConfig(),
    },
    hooks: {
      registered: hooks.length,
    },
    memory: {
      heapUsedMB: Math.round(memoryUsage.heapUsed / 1024 / 1024),
      heapTotalMB: Math.round(memoryUsage.heapTotal / 1024 / 1024),
      rssMB: Math.round(memoryUsage.rss / 1024 / 1024),
    },
    metrics: {
      http: {
        totalRequests: metricsSummary.http.totalRequests,
        totalErrors: metricsSummary.http.totalErrors,
        avgLatencyMs: metricsSummary.http.avgLatencyMs,
      },
      supervisors: {
        totalCalls: metricsSummary.supervisors.totalCalls,
        totalErrors: metricsSummary.supervisors.totalErrors,
        avgLatencyMs: metricsSummary.supervisors.avgLatencyMs,
      },
      debates: {
        count: metricsSummary.debates.count,
        consensusRate: metricsSummary.debates.consensusRate,
      },
    },
  });
});

app.get('/metrics', (c) => {
  const format = c.req.query('format');
  
  if (format === 'prometheus' || c.req.header('Accept')?.includes('text/plain')) {
    c.header('Content-Type', 'text/plain; version=0.0.4');
    return c.text(metrics.getPrometheusFormat());
  }
  
  return c.json(metrics.getSummary());
});

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;
  
  const parts: string[] = [];
  if (days > 0) parts.push(`${days}d`);
  if (hours > 0) parts.push(`${hours}h`);
  if (minutes > 0) parts.push(`${minutes}m`);
  parts.push(`${secs}s`);
  
  return parts.join(' ');
}

app.get('/ready', async (c) => {
  const supervisors = council.getSupervisors();
  
  if (supervisors.length === 0) {
    return c.json({ 
      ready: true, 
      message: 'No supervisors configured - accepting requests' 
    });
  }
  
  const available = await council.getAvailableSupervisors();
  const ready = available.length > 0;
  
  return c.json({ 
    ready, 
    availableSupervisors: available.length,
    totalSupervisors: supervisors.length,
  }, ready ? 200 : 503);
});

app.route('/api/sessions', sessionRoutes);
app.route('/api/council', councilRoutes);
app.route('/api/smart-pilot', smartPilotRoutes);
app.route('/api/hooks', hooksRoutes);
app.route('/api/cli', cliRoutes);
app.route('/api/health', healthRoutes);
app.route('/api/env', envRoutes);
app.route('/api/dynamic-selection', dynamicSelectionRoutes);
app.route('/api/veto', vetoRoutes);
app.route('/api/plugins', pluginRoutes);
app.route('/api/debate-history', debateHistoryRoutes);
app.route('/api/quota', quotaRoutes);
app.route('/api/analytics', supervisorAnalyticsRoutes);
app.route('/api/simulator', debateSimulatorRoutes);
app.route('/api/workspaces', workspaceRoutes);
app.route('/api/templates', templateRoutes);
app.route('/api/fine-tuned-models', fineTunedModelsRoutes);
app.route('/api/collaborative-debates', collaborativeDebatesRoutes);
app.route('/api/system', systemRoutes);
app.route('/api/visual', visualRoutes);
app.route('/api/memory', memoryRoutes);
app.route('/api/logs', logsRoutes);
app.route('/api/ide', ideRoutes);
app.route('/ws', wsRoutes);

app.use('/dashboard/*', serveStatic({ root: '../../public' }));
app.get('/dashboard', serveStatic({ path: '../../public/index.html' }));
app.get('/analytics', serveStatic({ path: '../../public/analytics.html' }));

app.notFound((c) => c.json({ 
  success: false, 
  error: 'Not found',
  path: c.req.path,
}, 404));

const port = config.server.port;
const host = config.server.host;

sessionManager.startPolling();
healthMonitor.start();
logRotation.start();
await debateHistory.initialize();
selfEvolution.start();
checkpointService.start();
autonomousMaintenance.start();

if (config.council.smartPilot) {
  smartPilot.setEnabled(true);
}

process.on('SIGINT', async () => {
  console.log('\nShutting down gracefully...');
  smartPilot.cleanup();
  healthMonitor.stop();
  logRotation.stop();
  selfEvolution.stop();
  checkpointService.stop();
  autonomousMaintenance.stop();
  await sessionManager.cleanup();
  process.exit(0);
});

process.on('SIGTERM', async () => {
  console.log('Received SIGTERM, shutting down...');
  smartPilot.cleanup();
  healthMonitor.stop();
  logRotation.stop();
  selfEvolution.stop();
  checkpointService.stop();
  autonomousMaintenance.stop();
  await sessionManager.cleanup();
  process.exit(0);
});

export default {
  port,
  hostname: host,
  fetch: app.fetch,
};

console.log(`Server running on http://${host}:${port}`);
console.log(`WebSocket available at ws://${host}:${port}/ws`);
console.log(`Health check: http://${host}:${port}/health`);
if (council.getSupervisors().length > 0) {
  console.log(`Council supervisors: ${council.getSupervisors().map(s => s.name).join(', ')}`);
} else {
  console.log('No supervisors configured. Add via API or set API keys in environment.');
}
