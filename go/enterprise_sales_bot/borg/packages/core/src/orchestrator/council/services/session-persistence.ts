import { loadConfig, type PersistenceConfig } from './config.js';
import type { PersistedSession } from './types.js';
import * as fs from 'fs';
import * as path from 'path';

interface SessionPersistenceData {
  version: number;
  updatedAt: number;
  sessions: PersistedSession[];
}

class SessionPersistence {
  private config: PersistenceConfig;
  private filePath: string;
  private saveTimer: ReturnType<typeof setInterval> | null = null;
  private dirty = false;
  private data: SessionPersistenceData = { version: 1, updatedAt: 0, sessions: [] };

  constructor() {
    const appConfig = loadConfig();
    this.config = {
      enabled: appConfig.persistence?.enabled ?? true,
      filePath: appConfig.persistence?.filePath ?? '.autopilot/sessions.json',
      autoSaveIntervalMs: appConfig.persistence?.autoSaveIntervalMs ?? 5000,
      autoResumeOnStart: appConfig.persistence?.autoResumeOnStart ?? true,
      maxPersistedSessions: appConfig.persistence?.maxPersistedSessions ?? 100,
    };
    
    this.filePath = path.isAbsolute(this.config.filePath) 
      ? this.config.filePath 
      : path.join(process.cwd(), this.config.filePath);
    
    this.ensureDirectory();
    this.load();
  }

  private ensureDirectory(): void {
    const dir = path.dirname(this.filePath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
  }

  private load(): void {
    try {
      if (fs.existsSync(this.filePath)) {
        const raw = fs.readFileSync(this.filePath, 'utf-8');
        this.data = JSON.parse(raw);
        console.log(`[SessionPersistence] Loaded ${this.data.sessions.length} persisted sessions`);
      }
    } catch (err) {
      console.error('[SessionPersistence] Failed to load sessions:', err);
      this.data = { version: 1, updatedAt: 0, sessions: [] };
    }
  }

  save(): void {
    if (!this.config.enabled) return;
    
    try {
      this.data.updatedAt = Date.now();
      fs.writeFileSync(this.filePath, JSON.stringify(this.data, null, 2));
      this.dirty = false;
    } catch (err) {
      console.error('[SessionPersistence] Failed to save sessions:', err);
    }
  }

  startAutoSave(): void {
    if (this.saveTimer || !this.config.enabled) return;
    
    this.saveTimer = setInterval(() => {
      if (this.dirty) {
        this.save();
      }
    }, this.config.autoSaveIntervalMs);
  }

  stopAutoSave(): void {
    if (this.saveTimer) {
      clearInterval(this.saveTimer);
      this.saveTimer = null;
    }
    if (this.dirty) {
      this.save();
    }
  }

  persistSession(session: PersistedSession): void {
    const idx = this.data.sessions.findIndex(s => s.id === session.id);
    if (idx >= 0) {
      this.data.sessions[idx] = session;
    } else {
      if (this.data.sessions.length >= this.config.maxPersistedSessions) {
        const oldest = this.data.sessions
          .filter(s => s.status !== 'running')
          .sort((a, b) => (a.lastActivity ?? a.startedAt) - (b.lastActivity ?? b.startedAt))[0];
        if (oldest) {
          this.data.sessions = this.data.sessions.filter(s => s.id !== oldest.id);
        }
      }
      this.data.sessions.push(session);
    }
    this.dirty = true;
  }

  updateSessionStatus(id: string, status: PersistedSession['status'], lastActivity?: number): void {
    const session = this.data.sessions.find(s => s.id === id);
    if (session) {
      session.status = status;
      if (lastActivity) session.lastActivity = lastActivity;
      this.dirty = true;
    }
  }

  removeSession(id: string): void {
    this.data.sessions = this.data.sessions.filter(s => s.id !== id);
    this.dirty = true;
  }

  getPersistedSessions(): PersistedSession[] {
    return [...this.data.sessions];
  }

  getResumableSessions(): PersistedSession[] {
    return this.data.sessions.filter(s => 
      s.status === 'running' || s.status === 'paused' || s.status === 'starting'
    );
  }

  getConfig(): PersistenceConfig {
    return { ...this.config };
  }

  isAutoResumeEnabled(): boolean {
    return this.config.autoResumeOnStart;
  }

  setEnabled(enabled: boolean): void {
    this.config.enabled = enabled;
    if (enabled) {
      this.startAutoSave();
    } else {
      this.stopAutoSave();
    }
  }

  clearAll(): void {
    this.data.sessions = [];
    this.dirty = true;
    this.save();
  }
}

export const sessionPersistence = new SessionPersistence();
