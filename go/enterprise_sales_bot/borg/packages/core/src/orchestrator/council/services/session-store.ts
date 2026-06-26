import type { Session, LogEntry } from './types.js';
import { wsManager } from './ws-manager.js';

class SessionStore {
  private sessions = new Map<string, Session>();

  get(id: string): Session | undefined {
    return this.sessions.get(id);
  }

  getAll(): Session[] {
    return Array.from(this.sessions.values());
  }

  create(id: string): Session {
    const session: Session = {
      id,
      status: 'idle',
      startedAt: Date.now(),
      logs: [],
    };
    this.sessions.set(id, session);
    wsManager.notifySessionUpdate(session);
    return session;
  }

  update(id: string, updates: Partial<Session>): Session | undefined {
    const session = this.sessions.get(id);
    if (!session) return undefined;
    
    Object.assign(session, updates, { lastActivity: Date.now() });
    wsManager.notifySessionUpdate(session);
    return session;
  }

  addLog(id: string, level: LogEntry['level'], message: string, source?: string): void {
    const session = this.sessions.get(id);
    if (!session) return;
    
    const log: LogEntry = {
      timestamp: Date.now(),
      level,
      message,
      source,
    };
    
    session.logs.push(log);
    session.lastActivity = Date.now();
    wsManager.notifyLog(id, log);
  }

  delete(id: string): boolean {
    return this.sessions.delete(id);
  }
}

export const sessionStore = new SessionStore();
