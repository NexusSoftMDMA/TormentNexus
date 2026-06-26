'use client';

import { useState, useEffect, useCallback } from 'react';

export interface CLITool {
  id: string;
  name: string;
  version: string;
  path: string;
  status: 'available' | 'unavailable';
  description?: string;
}

export interface CLISession {
  id: string;
  cliType: string;
  workingDirectory: string;
  status: 'active' | 'stopped' | 'error';
  uptime: number;
  startedAt: string;
  pid?: number;
}

export interface SmartPilotConfig {
  autoApprovalLimit: number;
  maxConcurrentSessions: number;
  decisionTimeout: number;
}

export interface SmartPilotStatus {
  enabled: boolean;
  state: 'running' | 'paused' | 'stopped';
  remainingApprovals: number;
  config: SmartPilotConfig;
}

export type RiskLevel = 'low' | 'medium' | 'high';

export interface VetoRequest {
  id: string;
  action: string;
  description: string;
  requestingAgent: string;
  sessionId?: string;
  riskLevel: RiskLevel;
  createdAt: string;
  expiresAt: string;
  timeRemaining: number;
}

export interface DebateDecision {
  id: string;
  timestamp: string;
  action: string;
  decision: 'approved' | 'rejected' | 'timeout';
  participants: string[];
  duration: number;
  confidence: number;
}

export interface DebateAnalytics {
  totalDecisions: number;
  approvalRate: number;
  avgDecisionTime: number;
  todayDecisions: number;
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

export async function fetchCLITools(): Promise<CLITool[]> {
  const data = await fetchJSON<{ tools: CLITool[] }>('/api/autopilot/cli/tools');
  return data.tools;
}


export async function refreshCLIDetection(): Promise<CLITool[]> {
  return fetchJSON<CLITool[]>('/api/autopilot/cli/tools/refresh', { method: 'POST' });
}

export async function registerCustomTool(tool: Partial<CLITool>): Promise<CLITool> {
  return fetchJSON<CLITool>('/api/autopilot/cli/tools/custom', {
    method: 'POST',
    body: JSON.stringify(tool),
  });
}


export async function fetchSessions(): Promise<CLISession[]> {
  const data = await fetchJSON<{ sessions: CLISession[] }>('/api/autopilot/sessions');
  return data.sessions;
}


export async function createSession(data: { cliType: string; workingDirectory: string }): Promise<CLISession> {
  return fetchJSON<CLISession>('/api/autopilot/sessions', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function startSession(id: string): Promise<CLISession> {
  return fetchJSON<CLISession>(`/api/autopilot/sessions/${id}/start`, { method: 'POST' });
}

export async function stopSession(id: string): Promise<CLISession> {
  return fetchJSON<CLISession>(`/api/autopilot/sessions/${id}/stop`, { method: 'POST' });
}

export async function restartSession(id: string): Promise<CLISession> {
  return fetchJSON<CLISession>(`/api/autopilot/sessions/${id}/restart`, { method: 'POST' });
}

export async function deleteSession(id: string): Promise<void> {
  await fetchJSON<void>(`/api/autopilot/sessions/${id}`, { method: 'DELETE' });
}

export async function startAllSessions(): Promise<void> {
  // No "start all" endpoint in core right now.
  // UI will fall back to starting sessions one-by-one.
  return;
}

export async function stopAllSessions(): Promise<void> {
  // No "stop all" endpoint in core right now.
  // UI will fall back to stopping sessions one-by-one.
  return;
}


export async function fetchSmartPilotStatus(): Promise<SmartPilotStatus> {
  return fetchJSON<SmartPilotStatus>('/api/autopilot/smart-pilot/status');
}

export async function updateSmartPilotStatus(enabled: boolean): Promise<SmartPilotStatus> {
  return fetchJSON<SmartPilotStatus>('/api/autopilot/smart-pilot/status', {
    method: 'PUT',
    body: JSON.stringify({ enabled }),
  });
}

export async function updateSmartPilotConfig(config: Partial<SmartPilotConfig>): Promise<SmartPilotStatus> {
  return fetchJSON<SmartPilotStatus>('/api/autopilot/smart-pilot/config', {
    method: 'PUT',
    body: JSON.stringify(config),
  });
}

export async function pauseSmartPilot(): Promise<SmartPilotStatus> {
  return fetchJSON<SmartPilotStatus>('/api/autopilot/smart-pilot/pause', { method: 'POST' });
}

export async function resumeSmartPilot(): Promise<SmartPilotStatus> {
  return fetchJSON<SmartPilotStatus>('/api/autopilot/smart-pilot/resume', { method: 'POST' });
}

export async function resetApprovals(): Promise<SmartPilotStatus> {
  return fetchJSON<SmartPilotStatus>('/api/autopilot/smart-pilot/reset-approvals', { method: 'POST' });
}

export async function fetchPendingVetos(): Promise<VetoRequest[]> {
  return fetchJSON<VetoRequest[]>('/api/autopilot/veto/pending');
}

export async function approveVeto(id: string): Promise<void> {
  await fetchJSON<void>(`/api/autopilot/veto/${id}/approve`, { method: 'POST' });
}

export async function rejectVeto(id: string): Promise<void> {
  await fetchJSON<void>(`/api/autopilot/veto/${id}/reject`, { method: 'POST' });
}

export async function extendVetoTimeout(id: string, seconds: number): Promise<VetoRequest> {
  return fetchJSON<VetoRequest>(`/api/autopilot/veto/${id}/extend`, {
    method: 'POST',
    body: JSON.stringify({ seconds }),
  });
}

export async function fetchDebateHistory(params?: {
  limit?: number;
  offset?: number;
  search?: string;
}): Promise<DebateDecision[]> {
  const queryParams = new URLSearchParams();
  if (params?.limit) queryParams.set('limit', String(params.limit));
  if (params?.offset) queryParams.set('offset', String(params.offset));
  if (params?.search) queryParams.set('search', params.search);
  
  const query = queryParams.toString();
  return fetchJSON<DebateDecision[]>(`/api/autopilot/debate-history${query ? `?${query}` : ''}`);
}

export async function fetchDebateAnalytics(): Promise<DebateAnalytics> {
  return fetchJSON<DebateAnalytics>('/api/autopilot/debate-history/analytics');
}

export async function exportDebateHistory(format: 'json' | 'csv'): Promise<Blob> {
  const response = await fetch(`${API_BASE}/api/autopilot/debate-history/export?format=${format}`);
  if (!response.ok) {
    throw new Error(`Export failed: ${response.status}`);
  }
  return response.blob();
}

export interface UseAutopilotReturn {
  cliTools: CLITool[];
  cliToolsLoading: boolean;
  cliToolsError: string | null;
  refreshTools: () => Promise<void>;
  registerTool: (tool: Partial<CLITool>) => Promise<void>;

  sessions: CLISession[];
  sessionsLoading: boolean;
  sessionsError: string | null;
  createNewSession: (cliType: string, workingDirectory: string) => Promise<void>;
  startSessionById: (id: string) => Promise<void>;
  stopSessionById: (id: string) => Promise<void>;
  restartSessionById: (id: string) => Promise<void>;
  deleteSessionById: (id: string) => Promise<void>;
  startAll: () => Promise<void>;
  stopAll: () => Promise<void>;

  smartPilotStatus: SmartPilotStatus | null;
  smartPilotLoading: boolean;
  smartPilotError: string | null;
  toggleSmartPilot: (enabled: boolean) => Promise<void>;
  updateConfig: (config: Partial<SmartPilotConfig>) => Promise<void>;
  pause: () => Promise<void>;
  resume: () => Promise<void>;
  resetApprovalCount: () => Promise<void>;

  vetoQueue: VetoRequest[];
  vetoLoading: boolean;
  vetoError: string | null;
  approve: (id: string) => Promise<void>;
  reject: (id: string) => Promise<void>;
  extendTimeout: (id: string, seconds: number) => Promise<void>;

  debateHistory: DebateDecision[];
  debateAnalytics: DebateAnalytics | null;
  debateLoading: boolean;
  debateError: string | null;
  searchHistory: (query: string) => Promise<void>;
  exportHistory: (format: 'json' | 'csv') => Promise<void>;

  refreshAll: () => Promise<void>;
}

export function useAutopilot(): UseAutopilotReturn {
  const [cliTools, setCLITools] = useState<CLITool[]>([]);
  const [cliToolsLoading, setCLIToolsLoading] = useState(true);
  const [cliToolsError, setCLIToolsError] = useState<string | null>(null);

  const [sessions, setSessions] = useState<CLISession[]>([]);
  const [sessionsLoading, setSessionsLoading] = useState(true);
  const [sessionsError, setSessionsError] = useState<string | null>(null);

  const [smartPilotStatus, setSmartPilotStatus] = useState<SmartPilotStatus | null>(null);
  const [smartPilotLoading, setSmartPilotLoading] = useState(true);
  const [smartPilotError, setSmartPilotError] = useState<string | null>(null);

  const [vetoQueue, setVetoQueue] = useState<VetoRequest[]>([]);
  const [vetoLoading, setVetoLoading] = useState(true);
  const [vetoError, setVetoError] = useState<string | null>(null);

  const [debateHistory, setDebateHistory] = useState<DebateDecision[]>([]);
  const [debateAnalytics, setDebateAnalytics] = useState<DebateAnalytics | null>(null);
  const [debateLoading, setDebateLoading] = useState(true);
  const [debateError, setDebateError] = useState<string | null>(null);

  const loadCLITools = useCallback(async () => {
    setCLIToolsLoading(true);
    setCLIToolsError(null);
    try {
      const tools = await fetchCLITools();
      setCLITools(tools);
    } catch (error) {
      setCLIToolsError(error instanceof Error ? error.message : 'Failed to load CLI tools');
    } finally {
      setCLIToolsLoading(false);
    }
  }, []);

  const loadSessions = useCallback(async () => {
    setSessionsLoading(true);
    setSessionsError(null);
    try {
      const data = await fetchSessions();
      setSessions(data);
    } catch (error) {
      setSessionsError(error instanceof Error ? error.message : 'Failed to load sessions');
    } finally {
      setSessionsLoading(false);
    }
  }, []);

  const loadSmartPilotStatus = useCallback(async () => {
    setSmartPilotLoading(true);
    setSmartPilotError(null);
    try {
      const status = await fetchSmartPilotStatus();
      setSmartPilotStatus(status);
    } catch (error) {
      setSmartPilotError(error instanceof Error ? error.message : 'Failed to load smart pilot status');
    } finally {
      setSmartPilotLoading(false);
    }
  }, []);

  const loadVetoQueue = useCallback(async () => {
    setVetoLoading(true);
    setVetoError(null);
    try {
      const data = await fetchPendingVetos();
      setVetoQueue(data);
    } catch (error) {
      setVetoError(error instanceof Error ? error.message : 'Failed to load veto queue');
    } finally {
      setVetoLoading(false);
    }
  }, []);

  const loadDebateHistory = useCallback(async () => {
    setDebateLoading(true);
    setDebateError(null);
    try {
      const [history, analytics] = await Promise.all([
        fetchDebateHistory({ limit: 50 }),
        fetchDebateAnalytics(),
      ]);
      setDebateHistory(history);
      setDebateAnalytics(analytics);
    } catch (error) {
      setDebateError(error instanceof Error ? error.message : 'Failed to load debate history');
    } finally {
      setDebateLoading(false);
    }
  }, []);

  useEffect(() => {
    loadCLITools();
    loadSessions();
    loadSmartPilotStatus();
    loadVetoQueue();
    loadDebateHistory();
  }, [loadCLITools, loadSessions, loadSmartPilotStatus, loadVetoQueue, loadDebateHistory]);

  useEffect(() => {
    const interval = setInterval(() => {
      loadVetoQueue();
    }, 5000);
    return () => clearInterval(interval);
  }, [loadVetoQueue]);

  const refreshTools = useCallback(async () => {
    setCLIToolsLoading(true);
    try {
      const tools = await refreshCLIDetection();
      setCLITools(tools);
    } catch (error) {
      setCLIToolsError(error instanceof Error ? error.message : 'Failed to refresh tools');
    } finally {
      setCLIToolsLoading(false);
    }
  }, []);

  const registerTool = useCallback(async (tool: Partial<CLITool>) => {
    const newTool = await registerCustomTool(tool);
    setCLITools(prev => [...prev, newTool]);
  }, []);

  const createNewSession = useCallback(async (cliType: string, workingDirectory: string) => {
    const session = await createSession({ cliType, workingDirectory });
    setSessions(prev => [...prev, session]);
  }, []);

  const startSessionById = useCallback(async (id: string) => {
    const updated = await startSession(id);
    setSessions(prev => prev.map(s => s.id === id ? updated : s));
  }, []);

  const stopSessionById = useCallback(async (id: string) => {
    const updated = await stopSession(id);
    setSessions(prev => prev.map(s => s.id === id ? updated : s));
  }, []);

  const restartSessionById = useCallback(async (id: string) => {
    const updated = await restartSession(id);
    setSessions(prev => prev.map(s => s.id === id ? updated : s));
  }, []);

  const deleteSessionById = useCallback(async (id: string) => {
    await deleteSession(id);
    setSessions(prev => prev.filter(s => s.id !== id));
  }, []);

  const startAll = useCallback(async () => {
    // Core does not currently provide a bulk start endpoint.
    // Start sequentially to avoid overwhelming the machine.
    for (const session of sessions) {
      if (session.status !== 'active') {
        try {
          await startSession(session.id);
        } catch {
          // ignore individual failures
        }
      }
    }
    await loadSessions();
  }, [loadSessions, sessions]);
 
  const stopAll = useCallback(async () => {
    // Core does not currently provide a bulk stop endpoint.
    for (const session of sessions) {
      if (session.status === 'active') {
        try {
          await stopSession(session.id);
        } catch {
          // ignore individual failures
        }
      }
    }
    await loadSessions();
  }, [loadSessions, sessions]);


  const toggleSmartPilot = useCallback(async (enabled: boolean) => {
    const status = await updateSmartPilotStatus(enabled);
    setSmartPilotStatus(status);
  }, []);

  const updateConfig = useCallback(async (config: Partial<SmartPilotConfig>) => {
    const status = await updateSmartPilotConfig(config);
    setSmartPilotStatus(status);
  }, []);

  const pause = useCallback(async () => {
    const status = await pauseSmartPilot();
    setSmartPilotStatus(status);
  }, []);

  const resume = useCallback(async () => {
    const status = await resumeSmartPilot();
    setSmartPilotStatus(status);
  }, []);

  const resetApprovalCount = useCallback(async () => {
    const status = await resetApprovals();
    setSmartPilotStatus(status);
  }, []);

  const approve = useCallback(async (id: string) => {
    await approveVeto(id);
    setVetoQueue(prev => prev.filter(v => v.id !== id));
    await loadDebateHistory();
  }, [loadDebateHistory]);

  const reject = useCallback(async (id: string) => {
    await rejectVeto(id);
    setVetoQueue(prev => prev.filter(v => v.id !== id));
    await loadDebateHistory();
  }, [loadDebateHistory]);

  const extendTimeout = useCallback(async (id: string, seconds: number) => {
    const updated = await extendVetoTimeout(id, seconds);
    setVetoQueue(prev => prev.map(v => v.id === id ? updated : v));
  }, []);

  const searchHistory = useCallback(async (query: string) => {
    setDebateLoading(true);
    try {
      const history = await fetchDebateHistory({ search: query, limit: 50 });
      setDebateHistory(history);
    } catch (error) {
      setDebateError(error instanceof Error ? error.message : 'Search failed');
    } finally {
      setDebateLoading(false);
    }
  }, []);

  const exportHistory = useCallback(async (format: 'json' | 'csv') => {
    const blob = await exportDebateHistory(format);
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `debate-history.${format}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  }, []);

  const refreshAll = useCallback(async () => {
    await Promise.all([
      loadCLITools(),
      loadSessions(),
      loadSmartPilotStatus(),
      loadVetoQueue(),
      loadDebateHistory(),
    ]);
  }, [loadCLITools, loadSessions, loadSmartPilotStatus, loadVetoQueue, loadDebateHistory]);

  return {
    cliTools,
    cliToolsLoading,
    cliToolsError,
    refreshTools,
    registerTool,

    sessions,
    sessionsLoading,
    sessionsError,
    createNewSession,
    startSessionById,
    stopSessionById,
    restartSessionById,
    deleteSessionById,
    startAll,
    stopAll,

    smartPilotStatus,
    smartPilotLoading,
    smartPilotError,
    toggleSmartPilot,
    updateConfig,
    pause,
    resume,
    resetApprovalCount,

    vetoQueue,
    vetoLoading,
    vetoError,
    approve,
    reject,
    extendTimeout,

    debateHistory,
    debateAnalytics,
    debateLoading,
    debateError,
    searchHistory,
    exportHistory,

    refreshAll,
  };
}
