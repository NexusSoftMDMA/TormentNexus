/**
 * @deprecated This standalone store has been migrated to `createResourceSlice` in the unified Root Store.
 * Import from `@src/stores` instead. This file is retained only for type reference.
 */
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { createLogger } from '@extension/shared/lib/logger';
import { eventBus } from '../events';

const logger = createLogger('useResourceStore');

// Basic MCP Resource Types
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

export interface ResourceState {
    resources: Resource[];
    templates: ResourceTemplate[];
    isLoading: boolean;
    selectedResourceUri: string | null;
    resourceContent: { uri: string; content: any } | null;
    error: string | null;

    // Actions
    setResources: (resources: Resource[], templates?: ResourceTemplate[]) => void;
    setLoading: (loading: boolean) => void;
    setError: (error: string | null) => void;
    selectResource: (uri: string | null) => void;
    setResourceContent: (uri: string, content: any) => void;
    clearResourceContent: () => void;
}

export const useResourceStore = create<ResourceState>()(
    devtools(
        (set, get) => ({
            resources: [],
            templates: [],
            isLoading: false,
            selectedResourceUri: null,
            resourceContent: null,
            error: null,

            setResources: (resources: Resource[], templates: ResourceTemplate[] = []) => {
                set({ resources, templates });
                logger.debug('[ResourceStore] Resources updated:', { resources: resources.length, templates: templates.length });
                eventBus.emit('resource:list-updated', { resources, templates });
            },

            setLoading: (loading: boolean) => {
                set({ isLoading: loading });
            },

            setError: (error: string | null) => {
                set({ error });
                if (error) {
                    logger.error('[ResourceStore] Error:', error);
                }
            },

            selectResource: (uri: string | null) => {
                set({ selectedResourceUri: uri });
                if (uri) {
                    logger.debug(`[ResourceStore] Selected resource: ${uri}`);
                    eventBus.emit('resource:selected', { uri });
                } else {
                    // Clear content when deselecting
                    get().clearResourceContent();
                }
            },

            setResourceContent: (uri: string, content: any) => {
                set({ resourceContent: { uri, content }, error: null });
                logger.debug(`[ResourceStore] Content loaded for: ${uri}`);
                eventBus.emit('resource:content-loaded', { uri });
            },

            clearResourceContent: () => {
                set({ resourceContent: null });
            },
        }),
        { name: 'ResourceStore', store: 'resource' },
    ),
);
