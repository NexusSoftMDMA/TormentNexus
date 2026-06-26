import { Hono } from 'hono';
import { supervisorAnalytics } from '../services/supervisor-analytics.js';

const app = new Hono();

app.get('/summary', (c) => {
  return c.json(supervisorAnalytics.getSummary());
});

app.get('/supervisor/:name', (c) => {
  const name = c.req.param('name');
  const from = c.req.query('from');
  const to = c.req.query('to');
  
  const timeRange = from && to ? { start: parseInt(from), end: parseInt(to) } : undefined;
  return c.json(supervisorAnalytics.getSupervisorMetrics(name, timeRange));
});

app.get('/supervisor/:name/patterns', (c) => {
  const name = c.req.param('name');
  const from = c.req.query('from');
  const to = c.req.query('to');
  
  const timeRange = from && to ? { start: parseInt(from), end: parseInt(to) } : undefined;
  return c.json(supervisorAnalytics.getVotingPatterns(name, timeRange));
});

app.get('/all', (c) => {
  const from = c.req.query('from');
  const to = c.req.query('to');
  
  const timeRange = from && to ? { start: parseInt(from), end: parseInt(to) } : undefined;
  const metrics = supervisorAnalytics.getAllSupervisorMetrics(timeRange);
  return c.json(Object.fromEntries(metrics));
});

app.get('/compare', (c) => {
  const from = c.req.query('from');
  const to = c.req.query('to');
  
  const timeRange = from && to ? { start: parseInt(from), end: parseInt(to) } : undefined;
  return c.json(supervisorAnalytics.compareSupervisors(timeRange));
});

app.get('/insights', (c) => {
  const from = c.req.query('from');
  const to = c.req.query('to');
  
  const timeRange = from && to ? { start: parseInt(from), end: parseInt(to) } : undefined;
  return c.json(supervisorAnalytics.generateInsights(timeRange));
});

app.get('/trends', (c) => {
  const periodDays = parseInt(c.req.query('periodDays') || '7');
  const bucketCount = parseInt(c.req.query('bucketCount') || '7');
  return c.json(supervisorAnalytics.getTrends(periodDays, bucketCount));
});

app.get('/debates', (c) => {
  const from = c.req.query('from');
  const to = c.req.query('to');
  
  const timeRange = from && to ? { start: parseInt(from), end: parseInt(to) } : undefined;
  return c.json(supervisorAnalytics.getDebateStatistics(timeRange));
});

app.post('/record-vote', async (c) => {
  const { supervisor, debateId, vote, confidence, responseTimeMs, tokensUsed, consensusVote } = await c.req.json();
  supervisorAnalytics.recordVote(supervisor, debateId, vote, confidence, responseTimeMs, tokensUsed, consensusVote);
  return c.json({ success: true });
});

app.post('/record-outcome', async (c) => {
  const { debateId, topic, consensusMode, outcome, participatingSupervisors, roundCount, durationMs } = await c.req.json();
  supervisorAnalytics.recordDebateOutcome(debateId, topic, consensusMode, outcome, participatingSupervisors, roundCount, durationMs);
  return c.json({ success: true });
});

app.post('/clear', (c) => {
  supervisorAnalytics.clearHistory();
  return c.json({ success: true });
});

export default app;
