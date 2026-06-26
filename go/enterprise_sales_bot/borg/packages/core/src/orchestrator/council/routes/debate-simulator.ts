import { Hono } from 'hono';
import { debateSimulator } from '../services/debate-simulator.js';

const app = new Hono();

app.get('/summary', (c) => {
  return c.json(debateSimulator.getSummary());
});

app.get('/debates', (c) => {
  const limit = c.req.query('limit');
  return c.json(debateSimulator.listStoredDebates(limit ? parseInt(limit) : undefined));
});

app.get('/debates/:id', (c) => {
  const id = c.req.param('id');
  const debate = debateSimulator.getStoredDebate(id);
  if (!debate) {
    return c.json({ error: 'Debate not found' }, 404);
  }
  return c.json(debate);
});

app.post('/debates', async (c) => {
  const debate = await c.req.json();
  debateSimulator.storeDebate(debate);
  return c.json({ success: true, id: debate.id });
});

app.post('/replay/:id', async (c) => {
  const id = c.req.param('id');
  const config = await c.req.json();
  
  try {
    const result = debateSimulator.replayDebate(id, config);
    return c.json(result);
  } catch (error) {
    return c.json({ error: (error as Error).message }, 400);
  }
});

app.post('/simulate', async (c) => {
  const config = await c.req.json();
  const result = debateSimulator.simulateDebate(config);
  return c.json(result);
});

app.get('/simulations', (c) => {
  const limit = c.req.query('limit');
  return c.json(debateSimulator.listSimulations(limit ? parseInt(limit) : undefined));
});

app.get('/simulations/:id', (c) => {
  const id = c.req.param('id');
  const simulation = debateSimulator.getSimulation(id);
  if (!simulation) {
    return c.json({ error: 'Simulation not found' }, 404);
  }
  return c.json(simulation);
});

app.post('/what-if/:id', async (c) => {
  const id = c.req.param('id');
  const { scenarios } = await c.req.json();
  
  try {
    const results = debateSimulator.runWhatIfAnalysis(id, scenarios);
    return c.json(results);
  } catch (error) {
    return c.json({ error: (error as Error).message }, 400);
  }
});

app.get('/compare-modes/:id', (c) => {
  const id = c.req.param('id');
  
  try {
    const results = debateSimulator.compareConsensusModes(id);
    return c.json(results);
  } catch (error) {
    return c.json({ error: (error as Error).message }, 400);
  }
});

app.post('/find-optimal-team/:id', async (c) => {
  const id = c.req.param('id');
  const { targetOutcome, minTeamSize } = await c.req.json();
  
  try {
    const result = debateSimulator.findOptimalTeam(id, targetOutcome, minTeamSize);
    if (!result) {
      return c.json({ error: 'No optimal team found for target outcome' }, 404);
    }
    return c.json(result);
  } catch (error) {
    return c.json({ error: (error as Error).message }, 400);
  }
});

app.delete('/debates', (c) => {
  debateSimulator.clearStoredDebates();
  return c.json({ success: true });
});

app.delete('/simulations', (c) => {
  debateSimulator.clearSimulations();
  return c.json({ success: true });
});

export default app;
