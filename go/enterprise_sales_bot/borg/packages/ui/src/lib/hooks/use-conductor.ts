'use client';

import { useState, useEffect, useCallback } from 'react';

export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed';
export type TaskRole = 'architect' | 'developer' | 'reviewer' | 'tester';

export interface ConductorTask {
  id: string;
  name: string;
  status: TaskStatus;
  role: TaskRole;
  progress: number;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
  error?: string;
}

export interface ConductorStatus {
  activeTasks: number;
  queueDepth: number;
  workerStatus: 'idle' | 'busy' | 'overloaded';
  totalCompleted: number;
  totalFailed: number;
}

export interface VibeKanbanStatus {
  running: boolean;
  frontendPort?: number;
  backendPort?: number;
  frontendUrl?: string;
  backendUrl?: string;
  startedAt?: string;
}

export interface VibeKanbanStartParams {
  frontendPort: number;
  backendPort: number;
}

const API_BASE = '';

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  return response.json();
}

export async function fetchConductorTasks(): Promise<ConductorTask[]> {
  return fetchJSON<ConductorTask[]>('/api/conductor/tasks');
}

export async function startConductorTask(role: TaskRole): Promise<ConductorTask> {
  return fetchJSON<ConductorTask>('/api/conductor/start', {
    method: 'POST',
    body: JSON.stringify({ role }),
  });
}

export async function fetchConductorStatus(): Promise<ConductorStatus> {
  return fetchJSON<ConductorStatus>('/api/conductor/status');
}

export async function startVibeKanban(params: VibeKanbanStartParams): Promise<VibeKanbanStatus> {
  return fetchJSON<VibeKanbanStatus>('/api/vibekanban/start', {
    method: 'POST',
    body: JSON.stringify(params),
  });
}

export async function stopVibeKanban(): Promise<VibeKanbanStatus> {
  return fetchJSON<VibeKanbanStatus>('/api/vibekanban/stop', {
    method: 'POST',
  });
}

export async function fetchVibeKanbanStatus(): Promise<VibeKanbanStatus> {
  return fetchJSON<VibeKanbanStatus>('/api/vibekanban/status');
}

export interface UseConductorReturn {
  tasks: ConductorTask[];
  tasksLoading: boolean;
  tasksError: string | null;
  startTask: (role: TaskRole) => Promise<void>;
  refreshTasks: () => Promise<void>;

  conductorStatus: ConductorStatus | null;
  conductorStatusLoading: boolean;
  conductorStatusError: string | null;
  refreshConductorStatus: () => Promise<void>;

  vibeKanbanStatus: VibeKanbanStatus | null;
  vibeKanbanLoading: boolean;
  vibeKanbanError: string | null;
  startVibeKanbanInstance: (params: VibeKanbanStartParams) => Promise<void>;
  stopVibeKanbanInstance: () => Promise<void>;
  refreshVibeKanbanStatus: () => Promise<void>;

  refreshAll: () => Promise<void>;
}

export function useConductor(): UseConductorReturn {
  const [tasks, setTasks] = useState<ConductorTask[]>([]);
  const [tasksLoading, setTasksLoading] = useState(true);
  const [tasksError, setTasksError] = useState<string | null>(null);

  const [conductorStatus, setConductorStatus] = useState<ConductorStatus | null>(null);
  const [conductorStatusLoading, setConductorStatusLoading] = useState(true);
  const [conductorStatusError, setConductorStatusError] = useState<string | null>(null);

  const [vibeKanbanStatus, setVibeKanbanStatus] = useState<VibeKanbanStatus | null>(null);
  const [vibeKanbanLoading, setVibeKanbanLoading] = useState(true);
  const [vibeKanbanError, setVibeKanbanError] = useState<string | null>(null);

  const loadTasks = useCallback(async () => {
    setTasksLoading(true);
    setTasksError(null);
    try {
      const data = await fetchConductorTasks();
      setTasks(data);
    } catch (error) {
      setTasksError(error instanceof Error ? error.message : 'Failed to load tasks');
    } finally {
      setTasksLoading(false);
    }
  }, []);

  const loadConductorStatus = useCallback(async () => {
    setConductorStatusLoading(true);
    setConductorStatusError(null);
    try {
      const data = await fetchConductorStatus();
      setConductorStatus(data);
    } catch (error) {
      setConductorStatusError(error instanceof Error ? error.message : 'Failed to load conductor status');
    } finally {
      setConductorStatusLoading(false);
    }
  }, []);

  const loadVibeKanbanStatus = useCallback(async () => {
    setVibeKanbanLoading(true);
    setVibeKanbanError(null);
    try {
      const data = await fetchVibeKanbanStatus();
      setVibeKanbanStatus(data);
    } catch (error) {
      setVibeKanbanError(error instanceof Error ? error.message : 'Failed to load VibeKanban status');
    } finally {
      setVibeKanbanLoading(false);
    }
  }, []);

  useEffect(() => {
    loadTasks();
    loadConductorStatus();
    loadVibeKanbanStatus();
  }, [loadTasks, loadConductorStatus, loadVibeKanbanStatus]);

  useEffect(() => {
    const hasRunningTasks = tasks.some(t => t.status === 'running');
    if (hasRunningTasks) {
      const interval = setInterval(() => {
        loadTasks();
        loadConductorStatus();
      }, 5000);
      return () => clearInterval(interval);
    }
  }, [tasks, loadTasks, loadConductorStatus]);

  const startTask = useCallback(async (role: TaskRole) => {
    try {
      const newTask = await startConductorTask(role);
      setTasks(prev => [...prev, newTask]);
      await loadConductorStatus();
    } catch (error) {
      setTasksError(error instanceof Error ? error.message : 'Failed to start task');
      throw error;
    }
  }, [loadConductorStatus]);

  const startVibeKanbanInstance = useCallback(async (params: VibeKanbanStartParams) => {
    setVibeKanbanLoading(true);
    setVibeKanbanError(null);
    try {
      const status = await startVibeKanban(params);
      setVibeKanbanStatus(status);
    } catch (error) {
      setVibeKanbanError(error instanceof Error ? error.message : 'Failed to start VibeKanban');
      throw error;
    } finally {
      setVibeKanbanLoading(false);
    }
  }, []);

  const stopVibeKanbanInstance = useCallback(async () => {
    setVibeKanbanLoading(true);
    setVibeKanbanError(null);
    try {
      const status = await stopVibeKanban();
      setVibeKanbanStatus(status);
    } catch (error) {
      setVibeKanbanError(error instanceof Error ? error.message : 'Failed to stop VibeKanban');
      throw error;
    } finally {
      setVibeKanbanLoading(false);
    }
  }, []);

  const refreshAll = useCallback(async () => {
    await Promise.all([
      loadTasks(),
      loadConductorStatus(),
      loadVibeKanbanStatus(),
    ]);
  }, [loadTasks, loadConductorStatus, loadVibeKanbanStatus]);

  return {
    tasks,
    tasksLoading,
    tasksError,
    startTask,
    refreshTasks: loadTasks,

    conductorStatus,
    conductorStatusLoading,
    conductorStatusError,
    refreshConductorStatus: loadConductorStatus,

    vibeKanbanStatus,
    vibeKanbanLoading,
    vibeKanbanError,
    startVibeKanbanInstance,
    stopVibeKanbanInstance,
    refreshVibeKanbanStatus: loadVibeKanbanStatus,

    refreshAll,
  };
}
