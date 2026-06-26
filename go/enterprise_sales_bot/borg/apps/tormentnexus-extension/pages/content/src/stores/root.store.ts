import { create } from 'zustand';
import { devtools, persist, createJSONStorage } from 'zustand/middleware';
import { createUISlice, type UISlice } from './slices/createUISlice';
import { createConfigSlice, type ConfigSlice } from './slices/createConfigSlice';
import { createConnectionSlice, type ConnectionSlice } from './slices/createConnectionSlice';
import { createToolSlice, type ToolSlice } from './slices/createToolSlice';
import { createResourceSlice, type ResourceSlice } from './slices/createResourceSlice';
import { createAdapterSlice, type AdapterSlice } from './slices/createAdapterSlice';
import { createExtensionStateStorage } from './extension-storage';

export type RootState = UISlice & ConfigSlice & ConnectionSlice & ToolSlice & ResourceSlice & AdapterSlice;

export const useRootStore = create<RootState>()(
  devtools(
    persist(
      (...a) => ({
        ...createUISlice(...a),
        ...createConfigSlice(...a),
        ...createConnectionSlice(...a),
        ...createToolSlice(...a),
        ...createResourceSlice(...a),
        ...createAdapterSlice(...a),
      }),
      {
        name: 'tormentnexus-extension-root-store',
        storage: createJSONStorage(createExtensionStateStorage),
        partialize: (state) => ({
          ui: {
            sidebar: {
              width: state.ui.sidebar.width,
              position: state.ui.sidebar.position,
              isVisible: state.ui.sidebar.isVisible,
              isMinimized: state.ui.sidebar.isMinimized,
            },
            preferences: state.ui.preferences,
            theme: state.ui.theme,
            mcpEnabled: state.ui.mcpEnabled,
            globalSettings: state.ui.globalSettings,
          },
          config: {
            featureFlags: state.config.featureFlags,
            userProperties: state.config.userProperties,
            userSegment: state.config.userSegment,
            notificationConfig: state.config.notificationConfig,
            shownNotifications: state.config.shownNotifications,
            notificationHistory: state.config.notificationHistory,
            lastFetchTime: state.config.lastFetchTime,
          },
          connection: {
            profiles: state.connection.profiles,
            activeProfileIds: state.connection.activeProfileIds,
            connections: {}
          },
          tool: {
            availableTools: state.tool.availableTools,
            toolsByProfile: state.tool.toolsByProfile,
            detectedTools: [],
            toolExecutions: {},
            isExecuting: false,
            lastExecutionId: null,
            enabledTools: state.tool.enabledTools,
            isLoadingEnablement: false,
          },
          resource: {
            resourcesByProfile: state.resource.resourcesByProfile,
            templatesByProfile: state.resource.templatesByProfile,
            allResources: state.resource.allResources,
            allTemplates: state.resource.allTemplates,
            isLoading: false,
            selectedResourceUri: null,
            resourceContent: null,
            error: null,
          },
          // adapter: not persisted — plugin registrations are runtime-only
        }),
      }
    ),
    { name: 'RootStore' }
  )
);
