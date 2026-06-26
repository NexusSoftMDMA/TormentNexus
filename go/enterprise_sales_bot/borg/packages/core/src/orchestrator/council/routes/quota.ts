import { Hono } from 'hono';
import { quotaManager } from '../services/quota-manager.js';

const app = new Hono();

app.get('/status', (c) => {
  return c.json(quotaManager.getStatus());
});

app.post('/config', async (c) => {
  const body = await c.req.json();
  quotaManager.setConfig(body);
  return c.json({ success: true, config: quotaManager.getConfig() });
});

app.get('/config', (c) => {
  return c.json(quotaManager.getConfig());
});

app.post('/enable', (c) => {
  quotaManager.setEnabled(true);
  return c.json({ success: true, enabled: true });
});

app.post('/disable', (c) => {
  quotaManager.setEnabled(false);
  return c.json({ success: true, enabled: false });
});

app.get('/check/:provider', (c) => {
  const provider = c.req.param('provider');
  const result = quotaManager.checkQuota(provider);
  return c.json(result);
});

app.get('/stats', (c) => {
  return c.json(quotaManager.getAllStats());
});

app.get('/stats/:provider', (c) => {
  const provider = c.req.param('provider');
  return c.json(quotaManager.getProviderStats(provider));
});

app.post('/limits/:provider', async (c) => {
  const provider = c.req.param('provider');
  const limits = await c.req.json();
  quotaManager.setProviderLimits(provider, limits);
  return c.json({ success: true, limits: quotaManager.getLimits(provider) });
});

app.get('/limits/:provider', (c) => {
  const provider = c.req.param('provider');
  return c.json(quotaManager.getLimits(provider));
});

app.post('/record/:provider', async (c) => {
  const provider = c.req.param('provider');
  const { tokensUsed, latencyMs, success } = await c.req.json();
  quotaManager.recordRequest(provider, tokensUsed || 0, latencyMs || 0, success ?? true);
  return c.json({ success: true });
});

app.post('/rate-limit-error/:provider', (c) => {
  const provider = c.req.param('provider');
  quotaManager.recordRateLimitError(provider);
  return c.json({ success: true });
});

app.post('/unthrottle/:provider', (c) => {
  const provider = c.req.param('provider');
  quotaManager.unthrottleProvider(provider);
  return c.json({ success: true });
});

app.post('/reset/:provider', (c) => {
  const provider = c.req.param('provider');
  quotaManager.resetProviderUsage(provider);
  return c.json({ success: true });
});

app.post('/reset-all', (c) => {
  quotaManager.resetAllUsage();
  return c.json({ success: true });
});

export default app;
