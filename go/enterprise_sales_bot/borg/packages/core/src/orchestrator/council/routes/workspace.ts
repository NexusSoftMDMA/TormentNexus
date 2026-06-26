import { Hono, type Context } from 'hono';
import { workspaceManager } from '../services/workspace-manager.js';
import type { WorkspaceConfig } from '../services/workspace-manager.js';
import { formatOptionalSqliteFailure, isSqliteUnavailableError } from '../../../db/sqliteAvailability.js';

const app = new Hono();

function handleWorkspaceRouteError(c: Context, error: unknown) {
  if (isSqliteUnavailableError(error)) {
    return c.json({
      error: formatOptionalSqliteFailure('Workspace manager is unavailable', error),
    }, 503);
  }

  throw error;
}

app.post('/', async (c) => {
  try {
    const body = await c.req.json<{
      name: string;
      path: string;
      description?: string;
      config?: Partial<WorkspaceConfig>;
    }>();

    if (!body.name || !body.path) {
      return c.json({ error: 'name and path are required' }, 400);
    }

    const workspace = await workspaceManager.createWorkspace(
      body.name,
      body.path,
      body.config,
      body.description
    );

    return c.json(workspace, 201);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.get('/', async (c) => {
  try {
    const status = c.req.query('status');
    const tag = c.req.query('tag');

    let workspaces = await workspaceManager.getAllWorkspaces();

    if (status) {
      workspaces = workspaces.filter(w => w.status === status);
    }
    if (tag) {
      workspaces = workspaces.filter(w => w.config.tags.includes(tag));
    }

    return c.json(workspaces);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.get('/active', async (c) => {
  try {
    const workspace = await workspaceManager.getActiveWorkspace();
    if (!workspace) {
      return c.json({ error: 'No active workspace' }, 404);
    }
    return c.json(workspace);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/active/:id', async (c) => {
  try {
    const id = c.req.param('id');
    const success = await workspaceManager.setActiveWorkspace(id);

    if (!success) {
      return c.json({ error: 'Workspace not found or not active' }, 404);
    }

    return c.json({ success: true, activeWorkspaceId: id });
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.delete('/active', (c) => {
  workspaceManager.clearActiveWorkspace();
  return c.json({ success: true });
});

app.get('/debates/active', async (c) => {
  try {
    const debates = await workspaceManager.getAllActiveDebates();
    return c.json(debates);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.get('/:id', async (c) => {
  try {
    const id = c.req.param('id');
    const workspace = await workspaceManager.getWorkspace(id);

    if (!workspace) {
      return c.json({ error: 'Workspace not found' }, 404);
    }

    return c.json(workspace);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.patch('/:id', async (c) => {
  try {
    const id = c.req.param('id');
    const updates = await c.req.json();

    const workspace = await workspaceManager.updateWorkspace(id, updates);

    if (!workspace) {
      return c.json({ error: 'Workspace not found' }, 404);
    }

    return c.json(workspace);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.patch('/:id/config', async (c) => {
  try {
    const id = c.req.param('id');
    const config = await c.req.json<Partial<WorkspaceConfig>>();

    const workspace = await workspaceManager.updateWorkspaceConfig(id, config);

    if (!workspace) {
      return c.json({ error: 'Workspace not found' }, 404);
    }

    return c.json(workspace);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.delete('/:id', async (c) => {
  try {
    const id = c.req.param('id');
    const success = await workspaceManager.deleteWorkspace(id);

    if (!success) {
      return c.json({ error: 'Workspace not found' }, 404);
    }

    return c.json({ success: true });
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/:id/archive', async (c) => {
  try {
    const id = c.req.param('id');
    const workspace = await workspaceManager.archiveWorkspace(id);

    if (!workspace) {
      return c.json({ error: 'Workspace not found' }, 404);
    }

    return c.json(workspace);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.get('/:id/stats', async (c) => {
  try {
    const id = c.req.param('id');
    const periodDays = parseInt(c.req.query('periodDays') || '30', 10);

    const stats = await workspaceManager.getWorkspaceStats(id, periodDays);

    if (!stats) {
      return c.json({ error: 'Workspace not found' }, 404);
    }

    return c.json(stats);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.get('/:id/debates', async (c) => {
  try {
    const id = c.req.param('id');
    const limit = c.req.query('limit') ? parseInt(c.req.query('limit')!, 10) : undefined;

    const debates = await workspaceManager.getWorkspaceDebates(id, limit);
    return c.json(debates);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.get('/:id/debates/active', async (c) => {
  try {
    const id = c.req.param('id');
    const debates = await workspaceManager.getActiveDebates(id);
    return c.json(debates);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/:id/debates', async (c) => {
  try {
    const id = c.req.param('id');
    const task = await c.req.json();

    const debate = await workspaceManager.startDebate(id, task);

    if (!debate) {
      return c.json({ error: 'Failed to start debate - workspace not found or concurrent limit reached' }, 400);
    }

    return c.json(debate, 201);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/:id/debates/:debateId/complete', async (c) => {
  try {
    const id = c.req.param('id');
    const debateId = c.req.param('debateId');
    const body = await c.req.json<{
      decision: any;
      tokensUsed?: number;
      cost?: number;
    }>();

    const debate = await workspaceManager.completeDebate(
      id,
      debateId,
      body.decision,
      body.tokensUsed,
      body.cost
    );

    if (!debate) {
      return c.json({ error: 'Debate not found' }, 404);
    }

    return c.json(debate);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/:id/debates/:debateId/fail', async (c) => {
  try {
    const id = c.req.param('id');
    const debateId = c.req.param('debateId');
    const body = await c.req.json<{ error: string }>();

    const debate = await workspaceManager.failDebate(id, debateId, body.error);

    if (!debate) {
      return c.json({ error: 'Debate not found' }, 404);
    }

    return c.json(debate);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/compare', async (c) => {
  try {
    const body = await c.req.json<{ workspaceIds: string[] }>();

    if (!body.workspaceIds || body.workspaceIds.length < 2) {
      return c.json({ error: 'At least 2 workspace IDs required' }, 400);
    }

    const comparison = await workspaceManager.compareWorkspaces(body.workspaceIds);
    return c.json(comparison);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/pause-all', async (c) => {
  try {
    const count = await workspaceManager.pauseAllWorkspaces();
    return c.json({ success: true, pausedCount: count });
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/resume-all', async (c) => {
  try {
    const count = await workspaceManager.resumeAllWorkspaces();
    return c.json({ success: true, resumedCount: count });
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/:targetId/clone-config/:sourceId', async (c) => {
  try {
    const sourceId = c.req.param('sourceId');
    const targetId = c.req.param('targetId');

    const success = await workspaceManager.cloneWorkspaceConfig(sourceId, targetId);

    if (!success) {
      return c.json({ error: 'Source or target workspace not found' }, 404);
    }

    return c.json({ success: true });
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.get('/:id/export', async (c) => {
  try {
    const id = c.req.param('id');
    const data = await workspaceManager.exportWorkspace(id);

    if (!data) {
      return c.json({ error: 'Workspace not found' }, 404);
    }

    return c.json(data);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

app.post('/import', async (c) => {
  try {
    const data = await c.req.json();
    const workspace = await workspaceManager.importWorkspace(data);
    return c.json(workspace, 201);
  } catch (error) {
    return handleWorkspaceRouteError(c, error);
  }
});

export default app;
