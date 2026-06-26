import { Hono } from 'hono';
import { humanVeto } from '../services/human-veto.js';

export const vetoRoutes = new Hono();

vetoRoutes.get('/status', (c) => {
  const stats = humanVeto.getStats();
  return c.json({
    success: true,
    ...stats,
  });
});

vetoRoutes.post('/toggle', async (c) => {
  const body = await c.req.json().catch(() => ({}));
  const enabled = body.enabled ?? !humanVeto.isEnabled();
  
  humanVeto.setEnabled(enabled);
  
  return c.json({
    success: true,
    enabled: humanVeto.isEnabled(),
    message: `Human veto ${enabled ? 'enabled' : 'disabled'}`,
  });
});

vetoRoutes.get('/config', (c) => {
  return c.json({
    success: true,
    config: humanVeto.getConfig(),
  });
});

vetoRoutes.post('/config', async (c) => {
  const body = await c.req.json();
  
  humanVeto.updateConfig({
    timeoutMs: body.timeoutMs,
    autoApproveOnTimeout: body.autoApproveOnTimeout,
    requireVetoForRejections: body.requireVetoForRejections,
    minConsensusForAutoApprove: body.minConsensusForAutoApprove,
  });
  
  return c.json({
    success: true,
    config: humanVeto.getConfig(),
  });
});

vetoRoutes.get('/pending', (c) => {
  const pending = humanVeto.getAllPending();
  return c.json({
    success: true,
    pending,
    count: pending.length,
  });
});

vetoRoutes.get('/pending/:id', (c) => {
  const id = c.req.param('id');
  const decision = humanVeto.getPendingDecision(id);
  
  if (!decision) {
    return c.json({ success: false, error: `Decision '${id}' not found` }, 404);
  }
  
  return c.json({
    success: true,
    decision,
  });
});

vetoRoutes.post('/pending/:id/approve', async (c) => {
  const id = c.req.param('id');
  const body = await c.req.json().catch(() => ({}));
  
  const result = await humanVeto.processVeto(id, {
    action: 'approve',
    reason: body.reason,
  });
  
  if (result === null) {
    const pending = humanVeto.getPendingDecision(id);
    if (!pending) {
      return c.json({ success: false, error: `Decision '${id}' not found or already processed` }, 404);
    }
  }
  
  return c.json({
    success: true,
    message: 'Decision approved',
    decision: result,
  });
});

vetoRoutes.post('/pending/:id/reject', async (c) => {
  const id = c.req.param('id');
  const body = await c.req.json().catch(() => ({}));
  
  const result = await humanVeto.processVeto(id, {
    action: 'reject',
    reason: body.reason,
  });
  
  if (result === null) {
    const pending = humanVeto.getPendingDecision(id);
    if (!pending) {
      return c.json({ success: false, error: `Decision '${id}' not found or already processed` }, 404);
    }
  }
  
  return c.json({
    success: true,
    message: 'Decision rejected',
    decision: result,
  });
});

vetoRoutes.post('/pending/:id/redebate', async (c) => {
  const id = c.req.param('id');
  const body = await c.req.json().catch(() => ({}));
  
  await humanVeto.processVeto(id, {
    action: 'redebate',
    reason: body.reason,
  });
  
  return c.json({
    success: true,
    message: 'Re-debate requested',
    note: 'Listen for redebate_requested event via WebSocket',
  });
});

vetoRoutes.get('/history', (c) => {
  const limit = parseInt(c.req.query('limit') || '50', 10);
  const history = humanVeto.getHistory(limit);
  
  return c.json({
    success: true,
    history,
    count: history.length,
  });
});
