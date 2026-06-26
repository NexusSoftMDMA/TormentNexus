import { Hono } from 'hono';
import { logRotation } from '../services/log-rotation.js';

const app = new Hono();

app.get('/status', (c) => {
  return c.json({
    success: true,
    data: logRotation.getStats(),
  });
});

app.get('/config', (c) => {
  return c.json({
    success: true,
    data: logRotation.getConfig(),
  });
});

app.post('/config', async (c) => {
  const body = await c.req.json();
  logRotation.configure(body);
  return c.json({ success: true, data: logRotation.getConfig() });
});

export default app;
