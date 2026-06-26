'use client';

import { useState, useEffect, useCallback } from 'react';

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
}

export interface Session {
  id: string;
  agentName: string;
  timestamp: string;
  messages: Message[];
  status?: 'active' | 'completed' | 'paused';
}

export interface Handoff {
  id: string;
  timestamp: string;
  description: string;
  context: string;
  status: 'pending' | 'claimed';
  claimedBy?: string;
  claimedAt?: string;
}

export interface CreateHandoffParams {
  description: string;
  context: string;
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

export async function listSessions(): Promise<Session[]> {
  return fetchJSON<Session[]>('/api/sessions');
}

export async function loadSession(id: string): Promise<Session> {
  return fetchJSON<Session>(`/api/sessions/${id}`);
}

export async function deleteSession(id: string): Promise<void> {
  await fetchJSON<void>(`/api/sessions/${id}`, { method: 'DELETE' });
}

export async function resumeSession(id: string): Promise<Session> {
  return fetchJSON<Session>(`/api/sessions/${id}/resume`, { method: 'POST' });
}

export async function getHandoffs(): Promise<Handoff[]> {
  return fetchJSON<Handoff[]>('/api/handoffs');
}

export async function createHandoff(params: CreateHandoffParams): Promise<Handoff> {
  return fetchJSON<Handoff>('/api/handoffs', {
    method: 'POST',
    body: JSON.stringify(params),
  });
}

export async function claimHandoff(id: string): Promise<Handoff> {
  return fetchJSON<Handoff>(`/api/handoffs/${id}/claim`, { method: 'POST' });
}

export interface UseSessionsReturn {
  sessions: Session[];
  sessionsLoading: boolean;
  sessionsError: string | null;
  selectedSession: Session | null;
  refreshSessions: () => Promise<void>;
  selectSession: (id: string) => Promise<void>;
  deleteSessionById: (id: string) => Promise<void>;
  resumeSessionById: (id: string) => Promise<void>;
  clearSelectedSession: () => void;

  handoffs: Handoff[];
  handoffsLoading: boolean;
  handoffsError: string | null;
  refreshHandoffs: () => Promise<void>;
  createNewHandoff: (params: CreateHandoffParams) => Promise<void>;
  claimHandoffById: (id: string) => Promise<void>;

  refreshAll: () => Promise<void>;
}

export function useSessions(): UseSessionsReturn {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [sessionsLoading, setSessionsLoading] = useState(true);
  const [sessionsError, setSessionsError] = useState<string | null>(null);
  const [selectedSession, setSelectedSession] = useState<Session | null>(null);

  const [handoffs, setHandoffs] = useState<Handoff[]>([]);
  const [handoffsLoading, setHandoffsLoading] = useState(true);
  const [handoffsError, setHandoffsError] = useState<string | null>(null);

  const loadSessions = useCallback(async () => {
    setSessionsLoading(true);
    setSessionsError(null);
    try {
      const data = await listSessions();
      setSessions(data);
    } catch (error) {
      setSessionsError(error instanceof Error ? error.message : 'Failed to load sessions');
      setSessions([]);
    } finally {
      setSessionsLoading(false);
    }
  }, []);

  const loadHandoffs = useCallback(async () => {
    setHandoffsLoading(true);
    setHandoffsError(null);
    try {
      const data = await getHandoffs();
      setHandoffs(data);
    } catch (error) {
      setHandoffsError(error instanceof Error ? error.message : 'Failed to load handoffs');
      setHandoffs([]);
    } finally {
      setHandoffsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadSessions();
    loadHandoffs();
  }, [loadSessions, loadHandoffs]);

  const selectSession = useCallback(async (id: string) => {
    try {
      const session = await loadSession(id);
      setSelectedSession(session);
    } catch (error) {
      setSessionsError(error instanceof Error ? error.message : 'Failed to load session');
    }
  }, []);

  const deleteSessionById = useCallback(async (id: string) => {
    try {
      await deleteSession(id);
      setSessions((prev) => prev.filter((s) => s.id !== id));
      if (selectedSession?.id === id) {
        setSelectedSession(null);
      }
    } catch (error) {
      setSessionsError(error instanceof Error ? error.message : 'Failed to delete session');
    }
  }, [selectedSession?.id]);

  const resumeSessionById = useCallback(async (id: string) => {
    try {
      const session = await resumeSession(id);
      setSessions((prev) => prev.map((s) => (s.id === id ? session : s)));
      setSelectedSession(session);
    } catch (error) {
      setSessionsError(error instanceof Error ? error.message : 'Failed to resume session');
    }
  }, []);

  const clearSelectedSession = useCallback(() => {
    setSelectedSession(null);
  }, []);

  const createNewHandoff = useCallback(async (params: CreateHandoffParams) => {
    try {
      const handoff = await createHandoff(params);
      setHandoffs((prev) => [handoff, ...prev]);
    } catch (error) {
      setHandoffsError(error instanceof Error ? error.message : 'Failed to create handoff');
      throw error;
    }
  }, []);

  const claimHandoffById = useCallback(async (id: string) => {
    try {
      const handoff = await claimHandoff(id);
      setHandoffs((prev) => prev.map((h) => (h.id === id ? handoff : h)));
    } catch (error) {
      setHandoffsError(error instanceof Error ? error.message : 'Failed to claim handoff');
    }
  }, []);

  const refreshAll = useCallback(async () => {
    await Promise.all([loadSessions(), loadHandoffs()]);
  }, [loadSessions, loadHandoffs]);

  return {
    sessions,
    sessionsLoading,
    sessionsError,
    selectedSession,
    refreshSessions: loadSessions,
    selectSession,
    deleteSessionById,
    resumeSessionById,
    clearSelectedSession,

    handoffs,
    handoffsLoading,
    handoffsError,
    refreshHandoffs: loadHandoffs,
    createNewHandoff,
    claimHandoffById,

    refreshAll,
  };
}
