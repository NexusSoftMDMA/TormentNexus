import { StateCreator } from 'zustand';
import { createLogger } from '@extension/shared/lib/logger';
import { eventBus } from '../../events';
import type { RootState } from '../root.store';

const logger = createLogger('ResourceSlice');

export interface Resource {
  uri: string;
  name: string;
  description?: string;
  mimeType?: string;
}

export interface ResourceTemplate {
  uriTemplate: string;
  name: string;
  description?: string;
  mimeType?: string;
}

export interface ResourceSlice {
  resource: {
    resourcesByProfile: Record<string, Resource[]>;
    templatesByProfile: Record<string, ResourceTemplate[]>;
    // Flat lists derived from active profiles, or all profiles
    allResources: Resource[];
    allTemplates: ResourceTemplate[];
    isLoading: boolean;
    selectedResourceUri: string | null;
    resourceContent: { uri: string; content: any } | null;
    error: string | null;
  };

  // Actions
  setResources: (profileId: string, resources: Resource[], templates?: ResourceTemplate[]) => void;
  setResourceLoading: (loading: boolean) => void;
  setResourceError: (error: string | null) => void;
  selectResource: (uri: string | null) => void;
  setResourceContent: (uri: string, content: any) => void;
  clearResourceContent: () => void;
}

const initialResourceState = {
  resourcesByProfile: {},
  templatesByProfile: {},
  allResources: [],
  allTemplates: [],
  isLoading: false,
  selectedResourceUri: null,
  resourceContent: null,
  error: null,
};

export const createResourceSlice: StateCreator<RootState, [], [], ResourceSlice> = (set, get) => ({
  resource: initialResourceState,

  setResources: (profileId: string, resources: Resource[], templates: ResourceTemplate[] = []) => {
    set((state: RootState) => {
      const newResourcesByProfile = { ...state.resource.resourcesByProfile, [profileId]: resources };
      const newTemplatesByProfile = { ...state.resource.templatesByProfile, [profileId]: templates };
      
      const allResources = Object.values(newResourcesByProfile).flat() as Resource[];
      const allTemplates = Object.values(newTemplatesByProfile).flat() as ResourceTemplate[];

      logger.debug(`[ResourceSlice] Resources updated for profile ${profileId}:`, { resources: resources.length, templates: templates.length });
      eventBus.emit('resource:list-updated', { resources: allResources, templates: allTemplates });

      return {
        resource: {
          ...state.resource,
          resourcesByProfile: newResourcesByProfile,
          templatesByProfile: newTemplatesByProfile,
          allResources,
          allTemplates,
        }
      };
    });
  },

  setResourceLoading: (loading: boolean) => {
    set((state: RootState) => ({ resource: { ...state.resource, isLoading: loading } }));
  },

  setResourceError: (error: string | null) => {
    set((state: RootState) => ({ resource: { ...state.resource, error } }));
    if (error) {
      logger.error('[ResourceSlice] Error:', error);
    }
  },

  selectResource: (uri: string | null) => {
    set((state: RootState) => ({ resource: { ...state.resource, selectedResourceUri: uri } }));
    if (uri) {
      logger.debug(`[ResourceSlice] Selected resource: ${uri}`);
      eventBus.emit('resource:selected', { uri });
    } else {
      get().clearResourceContent();
    }
  },

  setResourceContent: (uri: string, content: any) => {
    set((state: RootState) => ({
      resource: { ...state.resource, resourceContent: { uri, content }, error: null }
    }));
    logger.debug(`[ResourceSlice] Content loaded for: ${uri}`);
    eventBus.emit('resource:content-loaded', { uri });
  },

  clearResourceContent: () => {
    set((state: RootState) => ({ resource: { ...state.resource, resourceContent: null } }));
  },
});
