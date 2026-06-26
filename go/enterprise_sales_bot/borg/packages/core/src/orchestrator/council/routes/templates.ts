import { Hono } from 'hono';
import { debateTemplateManager } from '../services/debate-template-manager.js';
import type { TemplateCategory, TemplateConfig, TemplatePrompts, ScoringConfig } from '../services/debate-template-manager.js';
import type { TaskType } from './types.js';

const app = new Hono();

app.get('/', (c) => {
  const category = c.req.query('category') as TemplateCategory | undefined;
  const builtInOnly = c.req.query('builtIn') === 'true';
  const customOnly = c.req.query('custom') === 'true';

  let templates = debateTemplateManager.getAllTemplates();

  if (category) {
    templates = templates.filter(t => t.category === category);
  }
  if (builtInOnly) {
    templates = templates.filter(t => t.isBuiltIn);
  }
  if (customOnly) {
    templates = templates.filter(t => !t.isBuiltIn);
  }

  return c.json(templates);
});

app.get('/stats', (c) => {
  const stats = debateTemplateManager.getTemplateStats();
  return c.json(stats);
});

app.get('/categories', (c) => {
  const categories: TemplateCategory[] = [
    'code-review',
    'security-audit',
    'architecture-review',
    'performance-review',
    'api-design',
    'testing-strategy',
    'documentation',
    'refactoring',
    'custom',
  ];
  return c.json(categories);
});

app.post('/suggest', async (c) => {
  const body = await c.req.json<{ taskDescription: string; limit?: number }>();

  if (!body.taskDescription) {
    return c.json({ error: 'taskDescription is required' }, 400);
  }

  const suggestions = debateTemplateManager.suggestTemplates(
    body.taskDescription,
    body.limit || 3
  );

  return c.json(suggestions);
});

app.post('/find-best', async (c) => {
  const body = await c.req.json<{ taskDescription: string; taskType?: TaskType }>();

  if (!body.taskDescription) {
    return c.json({ error: 'taskDescription is required' }, 400);
  }

  const template = debateTemplateManager.findBestTemplate(
    body.taskDescription,
    body.taskType
  );

  if (!template) {
    return c.json({ error: 'No matching template found' }, 404);
  }

  return c.json(template);
});

app.post('/', async (c) => {
  const body = await c.req.json<{
    name: string;
    description: string;
    category: TemplateCategory;
    config: TemplateConfig;
    prompts: TemplatePrompts;
    scoring: ScoringConfig;
  }>();

  if (!body.name || !body.category || !body.config || !body.prompts || !body.scoring) {
    return c.json({ error: 'name, category, config, prompts, and scoring are required' }, 400);
  }

  const template = debateTemplateManager.createTemplate(body);
  return c.json(template, 201);
});

app.get('/by-name/:name', (c) => {
  const name = c.req.param('name');
  const template = debateTemplateManager.getTemplateByName(decodeURIComponent(name));

  if (!template) {
    return c.json({ error: 'Template not found' }, 404);
  }

  return c.json(template);
});

app.get('/:id', (c) => {
  const id = c.req.param('id');
  const template = debateTemplateManager.getTemplate(id);

  if (!template) {
    return c.json({ error: 'Template not found' }, 404);
  }

  return c.json(template);
});

app.patch('/:id', async (c) => {
  const id = c.req.param('id');
  const updates = await c.req.json();

  const template = debateTemplateManager.updateTemplate(id, updates);

  if (!template) {
    return c.json({ error: 'Template not found or is built-in (cannot modify)' }, 400);
  }

  return c.json(template);
});

app.delete('/:id', (c) => {
  const id = c.req.param('id');
  const success = debateTemplateManager.deleteTemplate(id);

  if (!success) {
    return c.json({ error: 'Template not found or is built-in (cannot delete)' }, 400);
  }

  return c.json({ success: true });
});

app.post('/:id/apply', async (c) => {
  const id = c.req.param('id');
  const body = await c.req.json<{ taskDescription: string }>();

  if (!body.taskDescription) {
    return c.json({ error: 'taskDescription is required' }, 400);
  }

  const application = debateTemplateManager.applyTemplate(id, body.taskDescription);

  if (!application) {
    return c.json({ error: 'Template not found' }, 404);
  }

  return c.json(application);
});

app.post('/:id/clone', async (c) => {
  const id = c.req.param('id');
  const body = await c.req.json<{ newName: string }>();

  if (!body.newName) {
    return c.json({ error: 'newName is required' }, 400);
  }

  const template = debateTemplateManager.cloneTemplate(id, body.newName);

  if (!template) {
    return c.json({ error: 'Source template not found' }, 404);
  }

  return c.json(template, 201);
});

app.get('/:id/export', (c) => {
  const id = c.req.param('id');
  const template = debateTemplateManager.exportTemplate(id);

  if (!template) {
    return c.json({ error: 'Template not found' }, 404);
  }

  return c.json(template);
});

app.get('/export/custom', (c) => {
  const templates = debateTemplateManager.exportAllCustomTemplates();
  return c.json(templates);
});

app.post('/import', async (c) => {
  const template = await c.req.json();
  const imported = debateTemplateManager.importTemplate(template);
  return c.json(imported, 201);
});

app.post('/reset', (c) => {
  debateTemplateManager.resetToBuiltIn();
  return c.json({ success: true, message: 'Reset to built-in templates only' });
});

export default app;
