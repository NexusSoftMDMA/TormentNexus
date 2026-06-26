import { Hono } from 'hono';
import { dynamicSupervisorSelection } from '../services/dynamic-supervisor-selection.js';
import { council } from '../services/council.js';
import type { DevelopmentTask } from './types.js';

export const dynamicSelectionRoutes = new Hono();

dynamicSelectionRoutes.get('/status', (c) => {
  const stats = dynamicSupervisorSelection.getStats();
  return c.json({
    success: true,
    ...stats,
  });
});

dynamicSelectionRoutes.post('/toggle', async (c) => {
  const body = await c.req.json().catch(() => ({}));
  const enabled = body.enabled ?? !dynamicSupervisorSelection.isEnabled();
  
  dynamicSupervisorSelection.setEnabled(enabled);
  
  return c.json({
    success: true,
    enabled: dynamicSupervisorSelection.isEnabled(),
    message: `Dynamic supervisor selection ${enabled ? 'enabled' : 'disabled'}`,
  });
});

dynamicSelectionRoutes.get('/profiles', (c) => {
  const profiles = dynamicSupervisorSelection.getAllProfiles();
  return c.json({
    success: true,
    profiles,
    count: profiles.length,
  });
});

dynamicSelectionRoutes.get('/profiles/:name', (c) => {
  const name = c.req.param('name');
  const profile = dynamicSupervisorSelection.getSupervisorProfile(name);
  
  if (!profile) {
    return c.json({ success: false, error: `Profile '${name}' not found` }, 404);
  }
  
  return c.json({
    success: true,
    profile,
  });
});

dynamicSelectionRoutes.post('/profiles', async (c) => {
  const body = await c.req.json();
  
  if (!body.name || !body.provider || !body.strengths) {
    return c.json({
      success: false,
      error: 'Missing required fields: name, provider, strengths',
    }, 400);
  }
  
  dynamicSupervisorSelection.addSupervisorProfile({
    name: body.name,
    provider: body.provider,
    strengths: body.strengths,
    weaknesses: body.weaknesses || [],
    specializations: body.specializations || [],
    preferredForLeadOn: body.preferredForLeadOn || [],
  });
  
  return c.json({
    success: true,
    message: `Profile '${body.name}' added/updated`,
    profile: dynamicSupervisorSelection.getSupervisorProfile(body.name),
  });
});

dynamicSelectionRoutes.delete('/profiles/:name', (c) => {
  const name = c.req.param('name');
  const profile = dynamicSupervisorSelection.getSupervisorProfile(name);
  
  if (!profile) {
    return c.json({ success: false, error: `Profile '${name}' not found` }, 404);
  }
  
  dynamicSupervisorSelection.removeSupervisorProfile(name);
  
  return c.json({
    success: true,
    message: `Profile '${name}' removed`,
  });
});

dynamicSelectionRoutes.get('/templates', (c) => {
  const templates = dynamicSupervisorSelection.getAllTemplates();
  return c.json({
    success: true,
    templates,
    count: templates.length,
  });
});

dynamicSelectionRoutes.get('/templates/:name', (c) => {
  const name = c.req.param('name');
  const template = dynamicSupervisorSelection.getTeamTemplate(name);
  
  if (!template) {
    return c.json({ success: false, error: `Template '${name}' not found` }, 404);
  }
  
  return c.json({
    success: true,
    template,
  });
});

dynamicSelectionRoutes.post('/templates', async (c) => {
  const body = await c.req.json();
  
  if (!body.name || !body.taskTypes || !body.supervisors) {
    return c.json({
      success: false,
      error: 'Missing required fields: name, taskTypes, supervisors',
    }, 400);
  }
  
  dynamicSupervisorSelection.addTeamTemplate({
    name: body.name,
    description: body.description || '',
    taskTypes: body.taskTypes,
    supervisors: body.supervisors,
    leadSupervisor: body.leadSupervisor,
    consensusMode: body.consensusMode,
    minSupervisors: body.minSupervisors,
  });
  
  return c.json({
    success: true,
    message: `Template '${body.name}' added/updated`,
    template: dynamicSupervisorSelection.getTeamTemplate(body.name),
  });
});

dynamicSelectionRoutes.delete('/templates/:name', (c) => {
  const name = c.req.param('name');
  const template = dynamicSupervisorSelection.getTeamTemplate(name);
  
  if (!template) {
    return c.json({ success: false, error: `Template '${name}' not found` }, 404);
  }
  
  dynamicSupervisorSelection.removeTeamTemplate(name);
  
  return c.json({
    success: true,
    message: `Template '${name}' removed`,
  });
});

dynamicSelectionRoutes.post('/detect', async (c) => {
  const body = await c.req.json();
  
  if (!body.description) {
    return c.json({
      success: false,
      error: 'Missing required field: description',
    }, 400);
  }
  
  const task: DevelopmentTask = {
    id: body.id || `task-${Date.now()}`,
    description: body.description,
    context: body.context || '',
    files: body.files || [],
  };
  
  const councilSupervisors = council.getSupervisors().map(s => s.name);
  dynamicSupervisorSelection.setAvailableSupervisors(councilSupervisors);
  
  const result = dynamicSupervisorSelection.detectTaskType(task);
  
  return c.json({
    success: true,
    ...result,
  });
});

dynamicSelectionRoutes.post('/select', async (c) => {
  const body = await c.req.json();
  
  if (!body.description) {
    return c.json({
      success: false,
      error: 'Missing required field: description',
    }, 400);
  }
  
  const task: DevelopmentTask = {
    id: body.id || `task-${Date.now()}`,
    description: body.description,
    context: body.context || '',
    files: body.files || [],
  };
  
  const councilSupervisors = council.getSupervisors().map(s => s.name);
  dynamicSupervisorSelection.setAvailableSupervisors(councilSupervisors);
  
  const result = dynamicSupervisorSelection.selectTeam(task);
  
  return c.json({
    success: true,
    selection: result,
  });
});

dynamicSelectionRoutes.post('/sync-supervisors', (c) => {
  const councilSupervisors = council.getSupervisors().map(s => s.name);
  dynamicSupervisorSelection.setAvailableSupervisors(councilSupervisors);
  
  return c.json({
    success: true,
    message: 'Synchronized supervisors from council',
    availableSupervisors: councilSupervisors,
  });
});
