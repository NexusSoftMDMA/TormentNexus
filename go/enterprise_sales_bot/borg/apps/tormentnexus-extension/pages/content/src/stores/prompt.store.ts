/**
 * Prompt Templates Store
 *
 * Persists reusable prompt templates via guarded extension storage.
 * Supports CRUD operations, search, and clipboard insertion.
 */

import { create } from 'zustand';
import { devtools, persist, createJSONStorage } from 'zustand/middleware';
import { createLogger } from '@extension/shared/lib/logger';
import { createExtensionStateStorage } from './extension-storage';

const logger = createLogger('PromptStore');

export interface PromptTemplate {
    id: string;
    title: string;
    content: string;
    tags: string[];
    createdAt: number;
    updatedAt: number;
}

interface PromptState {
    templates: PromptTemplate[];

    // Actions
    addTemplate: (title: string, content: string, tags?: string[]) => PromptTemplate;
    updateTemplate: (id: string, updates: Partial<Pick<PromptTemplate, 'title' | 'content' | 'tags'>>) => void;
    deleteTemplate: (id: string) => void;
    reorderTemplates: (fromIndex: number, toIndex: number) => void;
}

export const usePromptStore = create<PromptState>()(
    devtools(
        persist(
            (set, get) => ({
                templates: [],

                addTemplate: (title: string, content: string, tags: string[] = []) => {
                    const newTemplate: PromptTemplate = {
                        id: `prompt_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
                        title,
                        content,
                        tags,
                        createdAt: Date.now(),
                        updatedAt: Date.now(),
                    };
                    set(state => ({ templates: [...state.templates, newTemplate] }));
                    logger.debug(`[PromptStore] Template added: ${newTemplate.title}`);
                    return newTemplate;
                },

                updateTemplate: (id, updates) => {
                    set(state => ({
                        templates: state.templates.map(t =>
                            t.id === id ? { ...t, ...updates, updatedAt: Date.now() } : t,
                        ),
                    }));
                    logger.debug(`[PromptStore] Template updated: ${id}`);
                },

                deleteTemplate: (id) => {
                    set(state => ({ templates: state.templates.filter(t => t.id !== id) }));
                    logger.debug(`[PromptStore] Template deleted: ${id}`);
                },

                reorderTemplates: (fromIndex, toIndex) => {
                    set(state => {
                        const newTemplates = [...state.templates];
                        const [moved] = newTemplates.splice(fromIndex, 1);
                        newTemplates.splice(toIndex, 0, moved);
                        return { templates: newTemplates };
                    });
                },
            }),
            {
                name: 'mcp-prompt-templates-store',
                storage: createJSONStorage(createExtensionStateStorage),
            },
        ),
        { name: 'PromptStore' },
    ),
);
