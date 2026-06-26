import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { Node, Edge } from '@xyflow/react';
import { createExtensionStateStorage } from './extension-storage';

export interface Macro {
  id: string;
  name: string;
  description: string;
  nodes: Node[];
  edges: Edge[];
  createdAt: number;
  updatedAt: number;
}

interface MacroStore {
  macros: Macro[];
  addMacro: (macro: Omit<Macro, 'id' | 'createdAt' | 'updatedAt'>) => void;
  updateMacro: (id: string, updates: Partial<Omit<Macro, 'id' | 'createdAt'>>) => void;
  deleteMacro: (id: string) => void;
  getMacro: (id: string) => Macro | undefined;
}

export const useMacroStore = create<MacroStore>()(
  persist(
    (set, get) => ({
      macros: [],

      addMacro: (macroData) => set((state) => ({
        macros: [
          ...state.macros,
          {
            ...macroData,
            id: crypto.randomUUID(),
            createdAt: Date.now(),
            updatedAt: Date.now(),
          },
        ],
      })),

      updateMacro: (id, updates) => set((state) => ({
        macros: state.macros.map((m) =>
          m.id === id ? { ...m, ...updates, updatedAt: Date.now() } : m
        ),
      })),

      deleteMacro: (id) => set((state) => ({
        macros: state.macros.filter((m) => m.id !== id),
      })),

      getMacro: (id) => get().macros.find((m) => m.id === id),
    }),
    {
      name: 'mcp-macros-v2', // Changed storage key to drop old incompatible state
      storage: createJSONStorage(createExtensionStateStorage),
    }
  )
);
