import { Hono } from 'hono';
import { smartPilot } from '../services/smart-pilot.js';
import { apiRateLimit } from '../middleware/rate-limit.js';
import { apiKeyAuth } from '../middleware/auth.js';
import { smartPilotConfigSchema } from '../schemas.js';

export const smartPilotRoutes = new Hono();

smartPilotRoutes.get('/status', apiRateLimit(), (c) => {
  return c.json({
    enabled: smartPilot.isEnabled(),
    config: smartPilot.getConfig(),
  });
});

smartPilotRoutes.get('/active-plans', apiRateLimit(), (c) => {
  const plans = smartPilot.getActivePlans();
  return c.json({
    success: true,
    plans: Object.fromEntries(plans),
  });
});

smartPilotRoutes.post('/enable', apiRateLimit(), apiKeyAuth, (c) => {
  smartPilot.setEnabled(true);
  return c.json({ success: true, enabled: true });
});

smartPilotRoutes.post('/disable', apiRateLimit(), apiKeyAuth, (c) => {
  smartPilot.setEnabled(false);
  return c.json({ success: true, enabled: false });
});

smartPilotRoutes.post('/toggle', apiRateLimit(), apiKeyAuth, (c) => {
  const newState = !smartPilot.isEnabled();
  smartPilot.setEnabled(newState);
  return c.json({ success: true, enabled: newState });
});

smartPilotRoutes.post('/config', apiRateLimit(), apiKeyAuth, async (c) => {
  const body = smartPilotConfigSchema.parse(await c.req.json());
  
  if (typeof body.autoApproveThreshold === 'number') {
    smartPilot.setAutoApproveThreshold(body.autoApproveThreshold);
  }
  if (typeof body.requireUnanimous === 'boolean') {
    smartPilot.setRequireUnanimous(body.requireUnanimous);
  }
  if (typeof body.maxAutoApprovals === 'number') {
    smartPilot.setMaxAutoApprovals(body.maxAutoApprovals);
  }
  
  return c.json({ success: true, config: smartPilot.getConfig() });
});

smartPilotRoutes.post('/reset-approvals', apiRateLimit(), apiKeyAuth, async (c) => {
  const body = await c.req.json().catch(() => ({}));
  
  if (body.sessionId) {
    smartPilot.resetApprovalCount(body.sessionId);
  } else {
    smartPilot.resetAllApprovalCounts();
  }
  
  return c.json({ success: true });
});
