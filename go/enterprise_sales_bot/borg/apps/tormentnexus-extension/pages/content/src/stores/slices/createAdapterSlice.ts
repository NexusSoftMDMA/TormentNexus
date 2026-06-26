import { StateCreator } from 'zustand';
import { eventBus } from '../../events';
import type { AdapterPlugin, PluginRegistration, AdapterCapability } from '../../types/plugins';
import { createLogger } from '@extension/shared/lib/logger';
import type { RootState } from '../root.store';

const logger = createLogger('AdapterSlice');

export interface AdapterSlice {
  adapter: {
    registeredPlugins: Record<string, PluginRegistration>;
    activeAdapterName: string | null;
    currentCapabilities: AdapterCapability[];
    lastAdapterError: { name: string; error: string | Error } | null;
  };

  // Actions
  registerPlugin: (plugin: AdapterPlugin, config: PluginRegistration['config']) => Promise<boolean>;
  unregisterPlugin: (name: string) => Promise<void>;
  activateAdapter: (name: string) => Promise<boolean>;
  deactivateAdapter: (name: string, reason?: string) => Promise<void>;
  getPlugin: (name: string) => PluginRegistration | undefined;
  getActiveAdapter: () => PluginRegistration | undefined;
  updatePluginConfig: (name: string, config: Partial<PluginRegistration['config']>) => void;
  setPluginError: (name: string, error: string | Error) => void;
}

const initialAdapterState = {
  registeredPlugins: {} as Record<string, PluginRegistration>,
  activeAdapterName: null as string | null,
  currentCapabilities: [] as AdapterCapability[],
  lastAdapterError: null as { name: string; error: string | Error } | null,
};

export const createAdapterSlice: StateCreator<RootState, [], [], AdapterSlice> = (set, get) => ({
  adapter: initialAdapterState,

  registerPlugin: async (plugin: AdapterPlugin, config: PluginRegistration['config']): Promise<boolean> => {
    if (get().adapter.registeredPlugins[plugin.name]) {
      logger.warn(`Plugin "${plugin.name}" already registered.`);
      return false;
    }
    const registration: PluginRegistration = {
      plugin,
      config,
      registeredAt: Date.now(),
      status: 'registered',
    };
    set((state: RootState) => ({
      adapter: {
        ...state.adapter,
        registeredPlugins: { ...state.adapter.registeredPlugins, [plugin.name]: registration },
      }
    }));
    logger.debug(`Plugin "${plugin.name}" registered.`);
    eventBus.emit('plugin:registered', { name: plugin.name, version: plugin.version });
    return true;
  },

  unregisterPlugin: async (name: string): Promise<void> => {
    const pluginReg = get().adapter.registeredPlugins[name];
    if (!pluginReg) {
      logger.warn(`Plugin "${name}" not found for unregistration.`);
      return;
    }
    if (get().adapter.activeAdapterName === name && pluginReg.instance) {
      try {
        await pluginReg.instance.deactivate();
        await pluginReg.instance.cleanup();
      } catch (e) {
        logger.error(`Error deactivating/cleaning up plugin "${name}" during unregistration:`, e);
      }
    }
    const { [name]: _, ...remainingPlugins } = get().adapter.registeredPlugins;
    set((state: RootState) => ({
      adapter: {
        ...state.adapter,
        registeredPlugins: remainingPlugins,
        activeAdapterName: state.adapter.activeAdapterName === name ? null : state.adapter.activeAdapterName,
        currentCapabilities: state.adapter.activeAdapterName === name ? [] : state.adapter.currentCapabilities,
      }
    }));
    logger.debug(`Plugin "${name}" unregistered.`);
    eventBus.emit('plugin:unregistered', { name });
  },

  activateAdapter: async (name: string): Promise<boolean> => {
    const pluginReg = get().adapter.registeredPlugins[name];
    if (!pluginReg) {
      logger.error(`Cannot activate: Plugin "${name}" not registered.`);
      get().setPluginError(name, `Plugin "${name}" not registered.`);
      return false;
    }
    if (!pluginReg.config.enabled) {
      logger.warn(`Cannot activate: Plugin "${name}" is disabled by config.`);
      get().setPluginError(name, `Plugin "${name}" is disabled.`);
      return false;
    }

    const currentActiveAdapter = get().getActiveAdapter();
    if (currentActiveAdapter && currentActiveAdapter.plugin.name !== name) {
      if (currentActiveAdapter.plugin.name !== 'sidebar-plugin') {
        try {
          logger.debug(`Deactivating current adapter "${currentActiveAdapter.plugin.name}".`);
          await currentActiveAdapter.instance?.deactivate();
          eventBus.emit('adapter:deactivated', {
            pluginName: currentActiveAdapter.plugin.name,
            reason: 'switching adapter',
            timestamp: Date.now(),
          });
        } catch (e) {
          logger.error(`Error deactivating previous adapter "${currentActiveAdapter.plugin.name}":`, e);
        }
      } else {
        logger.debug(`Skipping deactivation of sidebar-plugin - it persists alongside site adapters.`);
      }
    }

    try {
      if (!pluginReg.instance) {
        const pluginContext = {
          eventBus,
          stores: {},
          utils: {},
          chrome,
          logger: console,
        };
        await pluginReg.plugin.initialize(pluginContext as any);
        pluginReg.instance = pluginReg.plugin;
        pluginReg.status = 'initialized';
        eventBus.emit('plugin:initialization-complete', { name });
      }

      logger.debug(`Activating adapter "${name}".`);
      await pluginReg.instance!.activate();
      pluginReg.status = 'active';
      pluginReg.lastUsedAt = Date.now();

      set((state: RootState) => ({
        adapter: {
          ...state.adapter,
          activeAdapterName: name !== 'sidebar-plugin' ? name : state.adapter.activeAdapterName,
          currentCapabilities: pluginReg.plugin.capabilities,
          lastAdapterError: null,
          registeredPlugins: { ...state.adapter.registeredPlugins, [name]: pluginReg },
        }
      }));
      logger.debug(`Adapter "${name}" activated with capabilities:`, pluginReg.plugin.capabilities);
      eventBus.emit('adapter:activated', { pluginName: name, timestamp: Date.now() });
      eventBus.emit('adapter:capability-changed', { name, capabilities: pluginReg.plugin.capabilities });
      return true;
    } catch (error: any) {
      logger.error(`Error activating adapter "${name}":`, error);
      pluginReg.status = 'error';
      pluginReg.error = error;
      set((state: RootState) => ({
        adapter: {
          ...state.adapter,
          lastAdapterError: { name, error },
          registeredPlugins: { ...state.adapter.registeredPlugins, [name]: pluginReg },
        }
      }));
      eventBus.emit('plugin:activation-failed', { name, error });
      eventBus.emit('adapter:error', { name, error });
      return false;
    }
  },

  deactivateAdapter: async (name: string, reason?: string): Promise<void> => {
    const pluginReg = get().adapter.registeredPlugins[name];
    if (!pluginReg || get().adapter.activeAdapterName !== name) {
      logger.warn(`Adapter "${name}" is not active or not registered.`);
      return;
    }
    try {
      await pluginReg.instance?.deactivate();
      pluginReg.status = 'inactive';
      set((state: RootState) => ({
        adapter: {
          ...state.adapter,
          activeAdapterName: null,
          currentCapabilities: [],
          registeredPlugins: { ...state.adapter.registeredPlugins, [name]: pluginReg },
        }
      }));
      logger.debug(`Adapter "${name}" deactivated. Reason: ${reason || 'user action'}`);
      eventBus.emit('adapter:deactivated', {
        pluginName: name,
        reason: reason || 'user action',
        timestamp: Date.now(),
      });
    } catch (error: any) {
      logger.error(`Error deactivating adapter "${name}":`, error);
      pluginReg.status = 'error';
      pluginReg.error = error;
      set((state: RootState) => ({
        adapter: {
          ...state.adapter,
          lastAdapterError: { name, error },
          registeredPlugins: { ...state.adapter.registeredPlugins, [name]: pluginReg },
        }
      }));
      eventBus.emit('adapter:error', { name, error });
    }
  },

  getPlugin: (name: string): PluginRegistration | undefined => {
    return get().adapter.registeredPlugins[name];
  },

  getActiveAdapter: (): PluginRegistration | undefined => {
    const activeName = get().adapter.activeAdapterName;
    return activeName ? get().adapter.registeredPlugins[activeName] : undefined;
  },

  updatePluginConfig: (name: string, configUpdate: Partial<PluginRegistration['config']>) => {
    const pluginReg = get().adapter.registeredPlugins[name];
    if (pluginReg) {
      const updatedReg = { ...pluginReg, config: { ...pluginReg.config, ...configUpdate } };
      set((state: RootState) => ({
        adapter: {
          ...state.adapter,
          registeredPlugins: { ...state.adapter.registeredPlugins, [name]: updatedReg },
        }
      }));
      logger.debug(`Config updated for plugin "${name}":`, updatedReg.config);
      if (name === get().adapter.activeAdapterName && updatedReg.config.enabled === false) {
        get().deactivateAdapter(name, 'disabled by config update');
      }
    } else {
      logger.warn(`Cannot update config: Plugin "${name}" not found.`);
    }
  },

  setPluginError: (name: string, error: string | Error) => {
    const pluginReg = get().adapter.registeredPlugins[name];
    if (pluginReg) {
      const updatedReg = { ...pluginReg, status: 'error' as const, error };
      set((state: RootState) => ({
        adapter: {
          ...state.adapter,
          registeredPlugins: { ...state.adapter.registeredPlugins, [name]: updatedReg },
          lastAdapterError: { name, error },
        }
      }));
    } else {
      set((state: RootState) => ({
        adapter: { ...state.adapter, lastAdapterError: { name, error } }
      }));
    }
    logger.error(`Error set for plugin/adapter "${name}":`, error);
    eventBus.emit('adapter:error', { name, error });
  },
});
