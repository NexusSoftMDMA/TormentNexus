/**
 * @deprecated This standalone store has been merged into `createConnectionSlice` in the unified Root Store.
 * Import from `@src/stores` instead. This file is retained only for type reference.
 */
import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { ConnectionType } from '@src/types/stores';
import { createExtensionStateStorage } from './extension-storage';

export interface ConnectionProfile {
  id: string;
  name: string;
  uri: string;
  connectionType: ConnectionType;
}

interface ProfileStore {
  profiles: ConnectionProfile[];
  activeProfileIds: string[];
  addProfile: (profile: Omit<ConnectionProfile, 'id'>) => void;
  removeProfile: (id: string) => void;
  updateProfile: (id: string, updates: Partial<ConnectionProfile>) => void;
  toggleProfileActive: (id: string) => void;
  setProfilesActive: (ids: string[]) => void;
}

export const useProfileStore = create<ProfileStore>()(
  persist(
    set => ({
      profiles: [
        {
          id: 'default-sse',
          name: 'Default (SSE)',
          uri: 'http://localhost:3006/sse',
          connectionType: 'sse',
        },
        {
          id: 'default-ws',
          name: 'Default (WebSocket)',
          uri: 'ws://localhost:3006/message',
          connectionType: 'websocket',
        },
      ],
      activeProfileIds: ['default-sse'],
      addProfile: profile =>
        set(state => ({
          profiles: [...state.profiles, { ...profile, id: crypto.randomUUID() }],
        })),
      removeProfile: id =>
        set(state => ({
          profiles: state.profiles.filter(p => p.id !== id),
          activeProfileIds: state.activeProfileIds.filter(activeId => activeId !== id),
        })),
      updateProfile: (id, updates) =>
        set(state => ({
          profiles: state.profiles.map(p => (p.id === id ? { ...p, ...updates } : p)),
        })),
      toggleProfileActive: id =>
        set(state => ({
          activeProfileIds: state.activeProfileIds.includes(id)
            ? state.activeProfileIds.filter(activeId => activeId !== id)
            : [...state.activeProfileIds, id]
        })),
      setProfilesActive: ids => set({ activeProfileIds: ids }),
    }),
    {
      name: 'mcp-connection-profiles',
      storage: createJSONStorage(createExtensionStateStorage),
    },
  ),
);
