import { Hono } from 'hono';
import { pluginManager } from '../services/plugin-manager.js';

export const pluginRoutes = new Hono();

pluginRoutes.get('/status', (c) => {
  const stats = pluginManager.getStats();
  return c.json({
    success: true,
    ...stats,
  });
});

pluginRoutes.get('/config', (c) => {
  return c.json({
    success: true,
    config: pluginManager.getConfig(),
  });
});

pluginRoutes.post('/config', async (c) => {
  const body = await c.req.json();
  pluginManager.updateConfig(body);
  return c.json({
    success: true,
    config: pluginManager.getConfig(),
  });
});

pluginRoutes.get('/list', (c) => {
  const plugins = pluginManager.getAllPlugins().map(p => ({
    name: p.manifest.name,
    version: p.manifest.version,
    description: p.manifest.description,
    author: p.manifest.author,
    status: p.status,
    loadedAt: p.loadedAt,
    supervisorCount: p.supervisors.length,
    supervisors: p.supervisors.map(s => s.name),
  }));
  
  return c.json({
    success: true,
    plugins,
    count: plugins.length,
  });
});

pluginRoutes.get('/:name', (c) => {
  const name = c.req.param('name');
  const plugin = pluginManager.getPlugin(name);
  
  if (!plugin) {
    return c.json({ success: false, error: `Plugin '${name}' not found` }, 404);
  }
  
  return c.json({
    success: true,
    plugin: {
      name: plugin.manifest.name,
      version: plugin.manifest.version,
      description: plugin.manifest.description,
      author: plugin.manifest.author,
      status: plugin.status,
      loadedAt: plugin.loadedAt,
      errorMessage: plugin.errorMessage,
      manifest: plugin.manifest,
      supervisors: plugin.supervisors.map(s => s.name),
    },
  });
});

pluginRoutes.post('/load', async (c) => {
  const body = await c.req.json();
  
  if (body.path) {
    try {
      const plugin = await pluginManager.loadPlugin(body.path);
      return c.json({
        success: true,
        message: `Plugin '${plugin.manifest.name}' loaded`,
        plugin: {
          name: plugin.manifest.name,
          version: plugin.manifest.version,
          supervisorCount: plugin.supervisors.length,
        },
      });
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      return c.json({ success: false, error: msg }, 400);
    }
  }
  
  if (body.manifest) {
    const validation = pluginManager.validateManifest(body.manifest);
    if (!validation.valid) {
      return c.json({
        success: false,
        error: 'Invalid manifest',
        validationErrors: validation.errors,
      }, 400);
    }
    
    try {
      const plugin = await pluginManager.loadPluginFromManifest(body.manifest);
      return c.json({
        success: true,
        message: `Plugin '${plugin.manifest.name}' loaded from manifest`,
        plugin: {
          name: plugin.manifest.name,
          version: plugin.manifest.version,
          supervisorCount: plugin.supervisors.length,
        },
      });
    } catch (error) {
      const msg = error instanceof Error ? error.message : String(error);
      return c.json({ success: false, error: msg }, 400);
    }
  }
  
  return c.json({
    success: false,
    error: 'Provide either "path" or "manifest" in request body',
  }, 400);
});

pluginRoutes.post('/load-all', async (c) => {
  const result = await pluginManager.loadAllPlugins();
  return c.json({
    success: true,
    ...result,
  });
});

pluginRoutes.delete('/:name', async (c) => {
  const name = c.req.param('name');
  const success = await pluginManager.unloadPlugin(name);
  
  if (!success) {
    return c.json({ success: false, error: `Plugin '${name}' not found` }, 404);
  }
  
  return c.json({
    success: true,
    message: `Plugin '${name}' unloaded`,
  });
});

pluginRoutes.post('/:name/reload', async (c) => {
  const name = c.req.param('name');
  
  try {
    const plugin = await pluginManager.reloadPlugin(name);
    if (!plugin) {
      return c.json({ success: false, error: `Plugin '${name}' not found` }, 404);
    }
    
    return c.json({
      success: true,
      message: `Plugin '${name}' reloaded`,
      plugin: {
        name: plugin.manifest.name,
        version: plugin.manifest.version,
        supervisorCount: plugin.supervisors.length,
      },
    });
  } catch (error) {
    const msg = error instanceof Error ? error.message : String(error);
    return c.json({ success: false, error: msg }, 400);
  }
});

pluginRoutes.post('/:name/enable', (c) => {
  const name = c.req.param('name');
  const success = pluginManager.setPluginStatus(name, 'active');
  
  if (!success) {
    return c.json({ success: false, error: `Plugin '${name}' not found` }, 404);
  }
  
  return c.json({
    success: true,
    message: `Plugin '${name}' enabled`,
  });
});

pluginRoutes.post('/:name/disable', (c) => {
  const name = c.req.param('name');
  const success = pluginManager.setPluginStatus(name, 'disabled');
  
  if (!success) {
    return c.json({ success: false, error: `Plugin '${name}' not found` }, 404);
  }
  
  return c.json({
    success: true,
    message: `Plugin '${name}' disabled`,
  });
});

pluginRoutes.get('/supervisors/all', (c) => {
  const supervisors = pluginManager.getAllSupervisors().map(s => s.name);
  return c.json({
    success: true,
    supervisors,
    count: supervisors.length,
  });
});

pluginRoutes.get('/:name/supervisors', (c) => {
  const name = c.req.param('name');
  const plugin = pluginManager.getPlugin(name);
  
  if (!plugin) {
    return c.json({ success: false, error: `Plugin '${name}' not found` }, 404);
  }
  
  const supervisors = pluginManager.getSupervisorsByPlugin(name).map(s => s.name);
  return c.json({
    success: true,
    supervisors,
    count: supervisors.length,
  });
});

pluginRoutes.post('/validate', async (c) => {
  const body = await c.req.json();
  const validation = pluginManager.validateManifest(body);
  return c.json({
    success: validation.valid,
    ...validation,
  });
});

pluginRoutes.get('/sample-manifest', (c) => {
  return c.json({
    success: true,
    manifest: pluginManager.createSampleManifest(),
  });
});
