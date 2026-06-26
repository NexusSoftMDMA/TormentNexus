import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { createExtensionStateStorage } from './extension-storage';

export interface ContextCaptureInput {
  content: string;
  name?: string;
  source?: string;
  sourceUrl?: string;
  sourceTitle?: string;
}

export interface ContextCaptureResult {
  item: ContextItem;
  duplicate: boolean;
}

function normalizeContextContent(content: string): string {
  return content.trim();
}

function buildContextName(content: string, name?: string): string {
  const trimmedName = name?.trim();
  if (trimmedName) {
    return trimmedName;
  }

  return content.length > 30 ? `${content.substring(0, 30)}...` : content;
}

export interface ContextItem {
  id: string;
  name: string;
  content: string;
  createdAt: number;
  updatedAt: number;
  source?: string;
  sourceUrl?: string;
  sourceTitle?: string;
}

interface ContextStore {
  contexts: ContextItem[];
  addContext: (content: string, name?: string) => void;
  captureContext: (input: ContextCaptureInput) => ContextCaptureResult | null;
  updateContext: (id: string, updates: Partial<Omit<ContextItem, 'id' | 'createdAt'>>) => void;
  deleteContext: (id: string) => void;
  getContext: (id: string) => ContextItem | undefined;
  clearContexts: () => void;
}

export const useContextStore = create<ContextStore>()(
  persist(
    (set, get) => ({
      contexts: [],

      addContext: (content: string, name?: string) => set((state) => {
        const normalizedContent = normalizeContextContent(content);
        if (!normalizedContent) {
          return state;
        }

        const id = crypto.randomUUID();
        const timestamp = Date.now();
        const finalName = buildContextName(normalizedContent, name);

        return {
          contexts: [
            {
              id,
              name: finalName,
              content: normalizedContent,
              createdAt: timestamp,
              updatedAt: timestamp,
            },
            ...state.contexts,
          ],
        };
      }),

      captureContext: input => {
        const normalizedContent = normalizeContextContent(input.content);
        if (!normalizedContent) {
          return null;
        }

        const existing = get().contexts.find(context => context.content === normalizedContent);
        if (existing) {
          const updatedItem: ContextItem = {
            ...existing,
            name: input.name?.trim() ? input.name.trim() : existing.name,
            source: input.source ?? existing.source,
            sourceUrl: input.sourceUrl ?? existing.sourceUrl,
            sourceTitle: input.sourceTitle ?? existing.sourceTitle,
            updatedAt: Date.now(),
          };

          set(state => ({
            contexts: [updatedItem, ...state.contexts.filter(context => context.id !== existing.id)],
          }));

          return { item: updatedItem, duplicate: true };
        }

        const createdItem: ContextItem = {
          id: crypto.randomUUID(),
          name: buildContextName(normalizedContent, input.name),
          content: normalizedContent,
          createdAt: Date.now(),
          updatedAt: Date.now(),
          source: input.source,
          sourceUrl: input.sourceUrl,
          sourceTitle: input.sourceTitle,
        };

        set(state => ({
          contexts: [createdItem, ...state.contexts],
        }));

        return { item: createdItem, duplicate: false };
      },

      updateContext: (id, updates) => set((state) => ({
        contexts: state.contexts.map((c) =>
          c.id === id
            ? {
                ...c,
                ...updates,
                content: updates.content ? normalizeContextContent(updates.content) : c.content,
                name: updates.content || updates.name ? buildContextName(
                  updates.content ? normalizeContextContent(updates.content) : c.content,
                  updates.name ?? c.name,
                ) : c.name,
                updatedAt: Date.now(),
              }
            : c
        ),
      })),

      deleteContext: (id) => set((state) => ({
        contexts: state.contexts.filter((c) => c.id !== id),
      })),

      getContext: (id) => get().contexts.find((c) => c.id === id),

      clearContexts: () => set({ contexts: [] }),
    }),
    {
      name: 'mcp-context',
      storage: createJSONStorage(createExtensionStateStorage),
    }
  )
);
