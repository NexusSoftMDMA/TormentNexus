import { describe, test, expect, beforeEach, afterEach } from 'vitest';
import { PluginManager } from '../plugin-manager.js';
import type { SupervisorConfig } from '../types.js';

describe('PluginManager', () => {
  let manager: PluginManager;

  const validManifest = {
    name: 'test-plugin',
    version: '1.0.0',
    description: 'Test plugin',
    supervisors: [
      {
        name: 'TestSupervisor',
        provider: 'custom' as const,
        baseURL: 'https://api.example.com/v1',
        model: 'test-model',
      },
    ],
  };

  beforeEach(() => {
    manager = new PluginManager({ autoLoad: false });
  });

  describe('getConfig/updateConfig', () => {
    test('returns default config', () => {
      const config = manager.getConfig();
      expect(config.autoLoad).toBe(false);
      expect(config.allowHotReload).toBe(true);
    });

    test('updates config', () => {
      manager.updateConfig({ allowHotReload: false });
      expect(manager.getConfig().allowHotReload).toBe(false);
    });
  });

  describe('loadPluginFromManifest', () => {
    test('loads valid manifest', async () => {
      const plugin = await manager.loadPluginFromManifest(validManifest);
      
      expect(plugin.manifest.name).toBe('test-plugin');
      expect(plugin.status).toBe('active');
      expect(plugin.supervisors.length).toBe(1);
    });

    test('emits plugin_loaded event', async () => {
      let eventData: any = null;
      manager.on('plugin_loaded', (data) => { eventData = data; });
      
      await manager.loadPluginFromManifest(validManifest);
      
      expect(eventData).not.toBeNull();
      expect(eventData.name).toBe('test-plugin');
    });

    test('rejects invalid manifest - missing name', async () => {
      const invalid = { ...validManifest, name: undefined };
      await expect(manager.loadPluginFromManifest(invalid as any)).rejects.toThrow();
    });

    test('rejects invalid manifest - missing supervisors', async () => {
      const invalid = { ...validManifest, supervisors: undefined };
      await expect(manager.loadPluginFromManifest(invalid as any)).rejects.toThrow();
    });

    test('reloads plugin when allowHotReload is true', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      const updated = { ...validManifest, version: '2.0.0' };
      const plugin = await manager.loadPluginFromManifest(updated);
      
      expect(plugin.manifest.version).toBe('2.0.0');
      expect(manager.getAllPlugins().length).toBe(1);
    });

    test('rejects duplicate when allowHotReload is false', async () => {
      manager.updateConfig({ allowHotReload: false });
      await manager.loadPluginFromManifest(validManifest);
      
      await expect(manager.loadPluginFromManifest(validManifest)).rejects.toThrow('already loaded');
    });
  });

  describe('unloadPlugin', () => {
    test('unloads existing plugin', async () => {
      await manager.loadPluginFromManifest(validManifest);
      expect(manager.getAllPlugins().length).toBe(1);
      
      const result = await manager.unloadPlugin('test-plugin');
      
      expect(result).toBe(true);
      expect(manager.getAllPlugins().length).toBe(0);
    });

    test('returns false for non-existent plugin', async () => {
      const result = await manager.unloadPlugin('non-existent');
      expect(result).toBe(false);
    });

    test('emits plugin_unloaded event', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      let eventData: any = null;
      manager.on('plugin_unloaded', (data) => { eventData = data; });
      
      await manager.unloadPlugin('test-plugin');
      
      expect(eventData).not.toBeNull();
      expect(eventData.name).toBe('test-plugin');
    });
  });

  describe('reloadPlugin', () => {
    test('reloads existing plugin', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      const plugin = await manager.reloadPlugin('test-plugin');
      
      expect(plugin).not.toBeNull();
      expect(plugin!.manifest.name).toBe('test-plugin');
    });

    test('returns null for non-existent plugin', async () => {
      const result = await manager.reloadPlugin('non-existent');
      expect(result).toBeNull();
    });
  });

  describe('setPluginStatus', () => {
    test('enables plugin', async () => {
      await manager.loadPluginFromManifest(validManifest);
      manager.setPluginStatus('test-plugin', 'disabled');
      
      const result = manager.setPluginStatus('test-plugin', 'active');
      
      expect(result).toBe(true);
      expect(manager.getPlugin('test-plugin')!.status).toBe('active');
    });

    test('disables plugin', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      manager.setPluginStatus('test-plugin', 'disabled');
      
      expect(manager.getPlugin('test-plugin')!.status).toBe('disabled');
    });

    test('returns false for non-existent plugin', () => {
      const result = manager.setPluginStatus('non-existent', 'active');
      expect(result).toBe(false);
    });
  });

  describe('getPlugin/getAllPlugins', () => {
    test('gets plugin by name', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      const plugin = manager.getPlugin('test-plugin');
      
      expect(plugin).toBeDefined();
      expect(plugin!.manifest.name).toBe('test-plugin');
    });

    test('returns undefined for non-existent plugin', () => {
      expect(manager.getPlugin('non-existent')).toBeUndefined();
    });

    test('gets all plugins', async () => {
      await manager.loadPluginFromManifest(validManifest);
      await manager.loadPluginFromManifest({
        ...validManifest,
        name: 'another-plugin',
        supervisors: [{ name: 'AnotherSupervisor', provider: 'custom' as const, baseURL: 'https://api.test.com' }],
      });
      
      expect(manager.getAllPlugins().length).toBe(2);
    });
  });

  describe('getActivePlugins', () => {
    test('returns only active plugins', async () => {
      await manager.loadPluginFromManifest(validManifest);
      await manager.loadPluginFromManifest({
        ...validManifest,
        name: 'disabled-plugin',
        supervisors: [{ name: 'DisabledSupervisor', provider: 'custom' as const, baseURL: 'https://api.test.com' }],
      });
      
      manager.setPluginStatus('disabled-plugin', 'disabled');
      
      const active = manager.getActivePlugins();
      expect(active.length).toBe(1);
      expect(active[0].manifest.name).toBe('test-plugin');
    });
  });

  describe('supervisor management', () => {
    test('getSupervisor returns supervisor from active plugin', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      const supervisor = manager.getSupervisor('TestSupervisor');
      expect(supervisor).toBeDefined();
    });

    test('getSupervisor returns undefined for disabled plugin', async () => {
      await manager.loadPluginFromManifest(validManifest);
      manager.setPluginStatus('test-plugin', 'disabled');
      
      const supervisor = manager.getSupervisor('TestSupervisor');
      expect(supervisor).toBeUndefined();
    });

    test('getAllSupervisors returns all from active plugins', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      const supervisors = manager.getAllSupervisors();
      expect(supervisors.length).toBe(1);
    });

    test('getSupervisorsByPlugin returns supervisors for specific plugin', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      const supervisors = manager.getSupervisorsByPlugin('test-plugin');
      expect(supervisors.length).toBe(1);
    });
  });

  describe('getStats', () => {
    test('returns correct statistics', async () => {
      await manager.loadPluginFromManifest(validManifest);
      
      const stats = manager.getStats();
      
      expect(stats.totalPlugins).toBe(1);
      expect(stats.activePlugins).toBe(1);
      expect(stats.disabledPlugins).toBe(0);
      expect(stats.totalSupervisors).toBe(1);
    });
  });

  describe('validateManifest', () => {
    test('validates correct manifest', () => {
      const result = manager.validateManifest(validManifest);
      expect(result.valid).toBe(true);
      expect(result.errors.length).toBe(0);
    });

    test('rejects manifest without name', () => {
      const result = manager.validateManifest({ version: '1.0.0', supervisors: [] });
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Missing or invalid "name" field');
    });

    test('rejects manifest without version', () => {
      const result = manager.validateManifest({ name: 'test', supervisors: [] });
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Missing or invalid "version" field');
    });

    test('rejects manifest without supervisors', () => {
      const result = manager.validateManifest({ name: 'test', version: '1.0.0' });
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('Missing or invalid "supervisors" array');
    });

    test('rejects non-object manifest', () => {
      const result = manager.validateManifest('not an object');
      expect(result.valid).toBe(false);
    });
  });

  describe('createSampleManifest', () => {
    test('creates valid sample manifest', () => {
      const sample = manager.createSampleManifest();
      
      expect(sample.name).toBeDefined();
      expect(sample.version).toBeDefined();
      expect(sample.supervisors.length).toBeGreaterThan(0);
      
      const validation = manager.validateManifest(sample);
      expect(validation.valid).toBe(true);
    });
  });
});
