import type { Supervisor, SupervisorConfig } from './types.js';
import { createSupervisor } from '../supervisors/index.js';
import { EventEmitter } from 'events';
import * as fs from 'fs';
import * as path from 'path';

interface PluginManifest {
  name: string;
  version: string;
  description?: string;
  author?: string;
  supervisors: SupervisorConfig[];
  hooks?: PluginHooks;
  dependencies?: string[];
}

interface PluginHooks {
  onLoad?: string;
  onUnload?: string;
  beforeDebate?: string;
  afterDebate?: string;
}

interface LoadedPlugin {
  manifest: PluginManifest;
  supervisors: Supervisor[];
  loadedAt: number;
  status: 'active' | 'disabled' | 'error';
  errorMessage?: string;
}

interface PluginConfig {
  pluginsDir: string;
  autoLoad: boolean;
  allowHotReload: boolean;
}

const DEFAULT_CONFIG: PluginConfig = {
  pluginsDir: './plugins',
  autoLoad: true,
  allowHotReload: true,
};

export class PluginManager extends EventEmitter {
  private config: PluginConfig;
  private plugins: Map<string, LoadedPlugin> = new Map();
  private supervisorRegistry: Map<string, { plugin: string; supervisor: Supervisor }> = new Map();

  constructor(config: Partial<PluginConfig> = {}) {
    super();
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  async initialize(): Promise<void> {
    if (this.config.autoLoad) {
      await this.loadAllPlugins();
    }
  }

  async loadAllPlugins(): Promise<{ loaded: string[]; failed: string[] }> {
    const loaded: string[] = [];
    const failed: string[] = [];

    if (!fs.existsSync(this.config.pluginsDir)) {
      return { loaded, failed };
    }

    const entries = fs.readdirSync(this.config.pluginsDir, { withFileTypes: true });
    
    for (const entry of entries) {
      if (entry.isDirectory()) {
        const pluginPath = path.join(this.config.pluginsDir, entry.name);
        try {
          await this.loadPlugin(pluginPath);
          loaded.push(entry.name);
        } catch (error) {
          failed.push(entry.name);
        }
      }
    }

    return { loaded, failed };
  }

  async loadPlugin(pluginPath: string): Promise<LoadedPlugin> {
    const manifestPath = path.join(pluginPath, 'manifest.json');
    
    if (!fs.existsSync(manifestPath)) {
      throw new Error(`Plugin manifest not found: ${manifestPath}`);
    }

    const manifestContent = fs.readFileSync(manifestPath, 'utf-8');
    const manifest: PluginManifest = JSON.parse(manifestContent);

    if (!manifest.name || !manifest.version || !manifest.supervisors) {
      throw new Error(`Invalid plugin manifest: missing required fields (name, version, supervisors)`);
    }

    if (this.plugins.has(manifest.name)) {
      if (this.config.allowHotReload) {
        await this.unloadPlugin(manifest.name);
      } else {
        throw new Error(`Plugin already loaded: ${manifest.name}`);
      }
    }

    const supervisors: Supervisor[] = [];
    for (const config of manifest.supervisors) {
      try {
        const supervisor = createSupervisor(config);
        supervisors.push(supervisor);
        this.supervisorRegistry.set(config.name, { plugin: manifest.name, supervisor });
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        throw new Error(`Failed to create supervisor ${config.name}: ${errorMsg}`);
      }
    }

    const loadedPlugin: LoadedPlugin = {
      manifest,
      supervisors,
      loadedAt: Date.now(),
      status: 'active',
    };

    this.plugins.set(manifest.name, loadedPlugin);

    this.emit('plugin_loaded', {
      name: manifest.name,
      version: manifest.version,
      supervisorCount: supervisors.length,
    });

    return loadedPlugin;
  }

  async loadPluginFromManifest(manifest: PluginManifest): Promise<LoadedPlugin> {
    if (!manifest.name || !manifest.version || !manifest.supervisors) {
      throw new Error(`Invalid plugin manifest: missing required fields`);
    }

    if (this.plugins.has(manifest.name)) {
      if (this.config.allowHotReload) {
        await this.unloadPlugin(manifest.name);
      } else {
        throw new Error(`Plugin already loaded: ${manifest.name}`);
      }
    }

    const supervisors: Supervisor[] = [];
    for (const config of manifest.supervisors) {
      try {
        const supervisor = createSupervisor(config);
        supervisors.push(supervisor);
        this.supervisorRegistry.set(config.name, { plugin: manifest.name, supervisor });
      } catch (error) {
        const errorMsg = error instanceof Error ? error.message : String(error);
        throw new Error(`Failed to create supervisor ${config.name}: ${errorMsg}`);
      }
    }

    const loadedPlugin: LoadedPlugin = {
      manifest,
      supervisors,
      loadedAt: Date.now(),
      status: 'active',
    };

    this.plugins.set(manifest.name, loadedPlugin);

    this.emit('plugin_loaded', {
      name: manifest.name,
      version: manifest.version,
      supervisorCount: supervisors.length,
    });

    return loadedPlugin;
  }

  async unloadPlugin(name: string): Promise<boolean> {
    const plugin = this.plugins.get(name);
    if (!plugin) {
      return false;
    }

    for (const supervisor of plugin.supervisors) {
      this.supervisorRegistry.delete(supervisor.name);
    }

    this.plugins.delete(name);

    this.emit('plugin_unloaded', {
      name,
      supervisorCount: plugin.supervisors.length,
    });

    return true;
  }

  async reloadPlugin(name: string): Promise<LoadedPlugin | null> {
    const plugin = this.plugins.get(name);
    if (!plugin) {
      return null;
    }

    const manifest = plugin.manifest;
    await this.unloadPlugin(name);
    return this.loadPluginFromManifest(manifest);
  }

  setPluginStatus(name: string, status: 'active' | 'disabled'): boolean {
    const plugin = this.plugins.get(name);
    if (!plugin) {
      return false;
    }

    plugin.status = status;
    
    this.emit('plugin_status_changed', {
      name,
      status,
    });

    return true;
  }

  getPlugin(name: string): LoadedPlugin | undefined {
    return this.plugins.get(name);
  }

  getAllPlugins(): LoadedPlugin[] {
    return Array.from(this.plugins.values());
  }

  getActivePlugins(): LoadedPlugin[] {
    return this.getAllPlugins().filter(p => p.status === 'active');
  }

  getSupervisor(name: string): Supervisor | undefined {
    const entry = this.supervisorRegistry.get(name);
    if (!entry) return undefined;
    
    const plugin = this.plugins.get(entry.plugin);
    if (!plugin || plugin.status !== 'active') return undefined;
    
    return entry.supervisor;
  }

  getAllSupervisors(): Supervisor[] {
    const supervisors: Supervisor[] = [];
    for (const plugin of this.getActivePlugins()) {
      supervisors.push(...plugin.supervisors);
    }
    return supervisors;
  }

  getSupervisorsByPlugin(pluginName: string): Supervisor[] {
    const plugin = this.plugins.get(pluginName);
    if (!plugin || plugin.status !== 'active') {
      return [];
    }
    return plugin.supervisors;
  }

  getStats(): {
    totalPlugins: number;
    activePlugins: number;
    disabledPlugins: number;
    totalSupervisors: number;
    pluginsDir: string;
  } {
    const all = this.getAllPlugins();
    const active = all.filter(p => p.status === 'active');
    
    return {
      totalPlugins: all.length,
      activePlugins: active.length,
      disabledPlugins: all.length - active.length,
      totalSupervisors: this.getAllSupervisors().length,
      pluginsDir: this.config.pluginsDir,
    };
  }

  getConfig(): PluginConfig {
    return { ...this.config };
  }

  updateConfig(updates: Partial<PluginConfig>): void {
    this.config = { ...this.config, ...updates };
  }

  validateManifest(manifest: unknown): { valid: boolean; errors: string[] } {
    const errors: string[] = [];
    
    if (!manifest || typeof manifest !== 'object') {
      return { valid: false, errors: ['Manifest must be an object'] };
    }

    const m = manifest as Record<string, unknown>;

    if (!m.name || typeof m.name !== 'string') {
      errors.push('Missing or invalid "name" field');
    }
    if (!m.version || typeof m.version !== 'string') {
      errors.push('Missing or invalid "version" field');
    }
    if (!m.supervisors || !Array.isArray(m.supervisors)) {
      errors.push('Missing or invalid "supervisors" array');
    } else {
      for (let i = 0; i < m.supervisors.length; i++) {
        const sup = m.supervisors[i];
        if (!sup.name) errors.push(`Supervisor ${i}: missing "name"`);
        if (!sup.provider) errors.push(`Supervisor ${i}: missing "provider"`);
      }
    }

    return { valid: errors.length === 0, errors };
  }

  createSampleManifest(): PluginManifest {
    return {
      name: 'my-custom-plugin',
      version: '1.0.0',
      description: 'A custom supervisor plugin',
      author: 'Your Name',
      supervisors: [
        {
          name: 'CustomGPT',
          provider: 'custom',
          baseURL: 'https://your-api-endpoint.com/v1',
          model: 'your-model-name',
          systemPrompt: 'You are a helpful code reviewer.',
        },
      ],
    };
  }
}

export const pluginManager = new PluginManager();
