import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { createExtensionStateStorage } from './extension-storage';

export type LogType = 'tool_execution' | 'connection' | 'error' | 'info';
export type LogStatus = 'success' | 'error' | 'pending' | 'info';

export interface ActivityLogItem {
  id: string;
  timestamp: number;
  type: LogType;
  title: string;
  detail?: string;
  status: LogStatus;
  metadata?: any;
}

interface ActivityStore {
  logs: ActivityLogItem[];
  addLog: (log: Omit<ActivityLogItem, 'id' | 'timestamp'>) => void;
  clearLogs: () => void;
  removeLog: (id: string) => void;
}

export const useActivityStore = create<ActivityStore>()(
  persist(
    set => ({
      logs: [],
      addLog: log =>
        set(state => {
          const newLog: ActivityLogItem = {
            ...log,
            id: crypto.randomUUID(),
            timestamp: Date.now(),
          };
          // Keep only the last 50 logs
          const updatedLogs = [newLog, ...state.logs].slice(0, 50);
          return { logs: updatedLogs };
        }),
      clearLogs: () => set({ logs: [] }),
      removeLog: id =>
        set(state => ({
          logs: state.logs.filter(log => log.id !== id),
        })),
    }),
    {
      name: 'mcp-activity-logs',
      storage: createJSONStorage(createExtensionStateStorage),
    },
  ),
);
