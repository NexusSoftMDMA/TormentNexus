import { Hono } from 'hono';
import type { Session, ApiResponse, DevelopmentTask, Guidance, BulkSessionRequest, BulkSessionResponse, SessionTemplate, CLIType } from './types.js';
import { sessionManager } from '../services/session-manager.js';
import { wsManager } from '../services/ws-manager.js';
import { loadConfig } from '../services/config.js';

const sessions = new Hono();

sessions.get('/', (c) => {
  const list = sessionManager.getAllSessions();
  return c.json<ApiResponse<Session[]>>({ success: true, data: list });
});

sessions.get('/active', (c) => {
  const list = sessionManager.getActiveSessions();
  return c.json<ApiResponse<Session[]>>({ success: true, data: list });
});

sessions.get('/stats', (c) => {
  const stats = sessionManager.getSessionStats();
  return c.json<ApiResponse<typeof stats>>({ success: true, data: stats });
});

sessions.get('/:id', (c) => {
  const session = sessionManager.getSession(c.req.param('id'));
  if (!session) {
    return c.json<ApiResponse<null>>({ success: false, error: 'Session not found' }, 404);
  }
  return c.json<ApiResponse<Session>>({ success: true, data: session });
});

sessions.post('/start', async (c) => {
  try {
    const body = await c.req.json().catch(() => ({}));
    const task = body.task as DevelopmentTask | undefined;
    const options = {
      tags: body.tags as string[] | undefined,
      templateName: body.template as string | undefined,
      workingDirectory: body.workingDirectory as string | undefined,
      cliType: body.cliType as CLIType | undefined,
      env: body.env as Record<string, string> | undefined,
    };
    const session = await sessionManager.startSession(task, options);
    return c.json<ApiResponse<Session>>({ success: true, data: session });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to start session';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 500);
  }
});

sessions.post('/bulk/start', async (c) => {
  try {
    const body = await c.req.json() as BulkSessionRequest & { cliType?: CLIType };
    const result = await sessionManager.startBulkSessions(body.count, {
      tags: body.tags,
      templateName: body.template,
      cliType: body.cliType,
      staggerDelayMs: body.staggerDelayMs ?? 500,
    });
    
    wsManager.broadcast({
      type: 'bulk_update',
      payload: { action: 'start', count: result.sessions.length, failed: result.failed.length },
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<BulkSessionResponse>>({ success: true, data: result });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to start bulk sessions';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 500);
  }
});

sessions.post('/bulk/stop', async (c) => {
  try {
    const result = await sessionManager.stopAllSessions();
    
    wsManager.broadcast({
      type: 'bulk_update',
      payload: { action: 'stop', stopped: result.stopped, failed: result.failed },
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<{ stopped: number; failed: number }>>({ success: true, data: result });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to stop all sessions';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 500);
  }
});

sessions.post('/bulk/resume', async (c) => {
  try {
    const result = await sessionManager.resumeAllSessions();
    
    wsManager.broadcast({
      type: 'bulk_update',
      payload: { action: 'resume', count: result.sessions.length, failed: result.failed.length },
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<BulkSessionResponse>>({ success: true, data: result });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to resume sessions';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 500);
  }
});

sessions.get('/persisted', (c) => {
  const list = sessionManager.getPersistedSessions();
  return c.json<ApiResponse<typeof list>>({ success: true, data: list });
});

sessions.get('/by-tag/:tag', (c) => {
  const tag = c.req.param('tag');
  const list = sessionManager.getSessionsByTag(tag);
  return c.json<ApiResponse<Session[]>>({ success: true, data: list });
});

sessions.get('/by-template/:template', (c) => {
  const template = c.req.param('template');
  const list = sessionManager.getSessionsByTemplate(template);
  return c.json<ApiResponse<Session[]>>({ success: true, data: list });
});

sessions.post('/:id/stop', async (c) => {
  try {
    await sessionManager.stopSession(c.req.param('id'));
    const session = sessionManager.getSession(c.req.param('id'));
    return c.json<ApiResponse<Session | null>>({ success: true, data: session || null });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to stop session';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 404);
  }
});

sessions.post('/:id/resume', async (c) => {
  try {
    const session = await sessionManager.resumeSession(c.req.param('id'));
    return c.json<ApiResponse<Session>>({ success: true, data: session });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to resume session';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 400);
  }
});

sessions.delete('/:id', async (c) => {
  try {
    await sessionManager.deleteSession(c.req.param('id'));
    return c.json<ApiResponse<{ deleted: boolean }>>({ success: true, data: { deleted: true } });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to delete session';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 400);
  }
});

sessions.post('/:id/guidance', async (c) => {
  try {
    const body = await c.req.json();
    const guidance: Guidance = {
      approved: body.approved ?? true,
      feedback: body.feedback || body.comment || '',
      suggestedNextSteps: body.suggestedNextSteps || body.suggestions || [],
    };
    await sessionManager.sendGuidance(c.req.param('id'), guidance);
    const session = sessionManager.getSession(c.req.param('id'));
    return c.json<ApiResponse<Session | null>>({ success: true, data: session || null });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to send guidance';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 400);
  }
});

sessions.get('/:id/logs', (c) => {
  const session = sessionManager.getSession(c.req.param('id'));
  if (!session) {
    return c.json<ApiResponse<null>>({ success: false, error: 'Session not found' }, 404);
  }
  return c.json<ApiResponse<typeof session.logs>>({ success: true, data: session.logs });
});

sessions.get('/:id/logs/export', (c) => {
  const session = sessionManager.getSession(c.req.param('id'));
  if (!session) {
    return c.json<ApiResponse<null>>({ success: false, error: 'Session not found' }, 404);
  }
  
  const format = c.req.query('format') || 'json';
  
  if (format === 'csv') {
    const csv = ['timestamp,level,message,source']
      .concat(session.logs.map(log => 
        `${new Date(log.timestamp).toISOString()},${log.level},"${(log.message || '').replace(/"/g, '""')}",${log.source || ''}`
      ))
      .join('\n');
    
    c.header('Content-Type', 'text/csv');
    c.header('Content-Disposition', `attachment; filename="session-${session.id}-logs.csv"`);
    return c.text(csv);
  }
  
  if (format === 'text') {
    const text = session.logs
      .map(log => `[${new Date(log.timestamp).toISOString()}] [${log.level.toUpperCase()}] ${log.message}`)
      .join('\n');
    
    c.header('Content-Type', 'text/plain');
    c.header('Content-Disposition', `attachment; filename="session-${session.id}-logs.txt"`);
    return c.text(text);
  }
  
  c.header('Content-Disposition', `attachment; filename="session-${session.id}-logs.json"`);
  return c.json({
    sessionId: session.id,
    exportedAt: new Date().toISOString(),
    logCount: session.logs.length,
    logs: session.logs,
  });
});

sessions.get('/logs/export-all', (c) => {
  const allSessions = sessionManager.getAllSessions();
  const format = c.req.query('format') || 'json';
  
  const allLogs = allSessions.flatMap(session => 
    session.logs.map(log => ({ ...log, sessionId: session.id }))
  ).sort((a, b) => a.timestamp - b.timestamp);
  
  if (format === 'csv') {
    const csv = ['timestamp,sessionId,level,message,source']
      .concat(allLogs.map(log => 
        `${new Date(log.timestamp).toISOString()},${log.sessionId},${log.level},"${(log.message || '').replace(/"/g, '""')}",${log.source || ''}`
      ))
      .join('\n');
    
    c.header('Content-Type', 'text/csv');
    c.header('Content-Disposition', `attachment; filename="all-sessions-logs.csv"`);
    return c.text(csv);
  }
  
  c.header('Content-Disposition', `attachment; filename="all-sessions-logs.json"`);
  return c.json({
    exportedAt: new Date().toISOString(),
    sessionCount: allSessions.length,
    logCount: allLogs.length,
    logs: allLogs,
  });
});

sessions.get('/templates', (c) => {
  const config = loadConfig();
  return c.json<ApiResponse<SessionTemplate[]>>({ success: true, data: config.templates });
});

sessions.post('/from-template/:name', async (c) => {
  try {
    const templateName = c.req.param('name');
    const config = loadConfig();
    const template = config.templates.find(t => t.name === templateName);
    
    if (!template) {
      return c.json<ApiResponse<null>>({ success: false, error: `Template '${templateName}' not found` }, 404);
    }
    
    const session = await sessionManager.startSession(undefined, {
      templateName: template.name,
      tags: template.tags,
    });
    
    return c.json<ApiResponse<Session>>({ success: true, data: session });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to start session from template';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 500);
  }
});

sessions.get('/by-cli/:cliType', (c) => {
  const cliType = c.req.param('cliType') as CLIType;
  const list = sessionManager.getSessionsByCLI(cliType);
  return c.json<ApiResponse<Session[]>>({ success: true, data: list });
});

sessions.put('/:id/tags', async (c) => {
  try {
    const id = c.req.param('id');
    const body = await c.req.json();
    const tags = body.tags as string[];
    
    if (!Array.isArray(tags)) {
      return c.json<ApiResponse<null>>({ success: false, error: 'tags must be an array' }, 400);
    }
    
    sessionManager.updateSessionTags(id, tags);
    const session = sessionManager.getSession(id);
    
    if (!session) {
      return c.json<ApiResponse<null>>({ success: false, error: 'Session not found' }, 404);
    }
    
    wsManager.broadcast({
      type: 'session_update',
      payload: { sessionId: id, action: 'tags_updated', tags },
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<Session>>({ success: true, data: session });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to update tags';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 400);
  }
});

sessions.post('/:id/tags', async (c) => {
  try {
    const id = c.req.param('id');
    const body = await c.req.json();
    const tag = body.tag as string;
    
    if (!tag || typeof tag !== 'string') {
      return c.json<ApiResponse<null>>({ success: false, error: 'tag must be a string' }, 400);
    }
    
    sessionManager.addSessionTag(id, tag);
    const session = sessionManager.getSession(id);
    
    if (!session) {
      return c.json<ApiResponse<null>>({ success: false, error: 'Session not found' }, 404);
    }
    
    wsManager.broadcast({
      type: 'session_update',
      payload: { sessionId: id, action: 'tag_added', tag },
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<Session>>({ success: true, data: session });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to add tag';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 400);
  }
});

sessions.delete('/:id/tags/:tag', (c) => {
  try {
    const id = c.req.param('id');
    const tag = c.req.param('tag');
    
    sessionManager.removeSessionTag(id, tag);
    const session = sessionManager.getSession(id);
    
    if (!session) {
      return c.json<ApiResponse<null>>({ success: false, error: 'Session not found' }, 404);
    }
    
    wsManager.broadcast({
      type: 'session_update',
      payload: { sessionId: id, action: 'tag_removed', tag },
      timestamp: Date.now(),
    });
    
    return c.json<ApiResponse<Session>>({ success: true, data: session });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to remove tag';
    return c.json<ApiResponse<null>>({ success: false, error: message }, 400);
  }
});

export { sessions as sessionRoutes };
