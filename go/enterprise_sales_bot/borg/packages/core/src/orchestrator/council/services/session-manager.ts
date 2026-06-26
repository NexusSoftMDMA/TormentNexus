import type {
  Session,
  DevelopmentTask,
  Guidance,
  CLIType,
  SessionHealth,
  CLITool,
  PersistedSession,
} from './types.js';
import path from 'path';
import { EventEmitter } from 'events';
import { cliRegistry } from './cli-registry.js';
import { wsManager } from './ws-manager.js';
import { loadConfig } from './config.js';
import { sessionPersistence } from './session-persistence.js';
import { getMcpServer } from '../../../lib/trpc-core.js';
import type { PtySupervisor } from '../../../supervisor/PtySupervisor.js';
import type { SupervisedSessionSnapshot, SupervisedSessionStatus } from '../../../supervisor/types.js';

export interface SessionOptions {
  tags?: string[];
  templateName?: string;
  workingDirectory?: string;
  cliType?: CLIType;
  env?: Record<string, string>;
}

export interface BulkSessionRequest {
  count?: number;
  templateName?: string;
  tags?: string[];
  cliType?: CLIType;
  staggerDelayMs?: number;
}

export interface BulkSessionResponse {
  sessions: Session[];
  failed: Array<{ index: number; error: string }>;
}

interface CouncilSessionState {
  id: string;
  cliType: CLIType;
  task?: DevelopmentTask;
  templateName?: string;
  tags: string[];
}

const ACTIVE_SESSION_STATUSES = new Set<Session['status']>(['starting', 'running']);

class SessionManager extends EventEmitter {
  private readonly sessions = new Map<string, CouncilSessionState>();
  private readonly config = loadConfig();
  private pollTimer: ReturnType<typeof setInterval> | null = null;

  constructor() {
    super();
    sessionPersistence.startAutoSave();
  }

  private getPtySupervisor(): PtySupervisor | null {
    try {
      const mcpServer = getMcpServer();
      if (!mcpServer.ptySupervisor) {
        console.warn('[SessionManager] PtySupervisor not initialized in MCPServer');
        return null;
      }
      return mcpServer.ptySupervisor;
    } catch {
      // Graceful fallback for standalone mode
      return null;
    }
  }

  private async resolveTool(cliType: CLIType): Promise<CLITool> {
    let tool = cliRegistry.getTool(cliType);
    if (!tool) {
      await cliRegistry.detectAll();
      tool = cliRegistry.getTool(cliType);
    }

    if (!tool) {
      throw new Error(`CLI tool '${cliType}' not found or not supported.`);
    }

    return tool;
  }

  private toCouncilStatus(status: SupervisedSessionStatus): Session['status'] {
    switch (status) {
      case 'created':
        return 'idle';
      case 'starting':
      case 'restarting':
        return 'starting';
      case 'running':
        return 'running';
      case 'stopping':
        return 'paused';
      case 'stopped':
        return 'stopped';
      case 'error':
        return 'error';
      default:
        return 'error';
    }
  }

  private syncState(snapshot: SupervisedSessionSnapshot): CouncilSessionState {
    const existing = this.sessions.get(snapshot.id);
    if (existing) {
      return existing;
    }

    const metadata = snapshot.metadata ?? {};
    const synced: CouncilSessionState = {
      id: snapshot.id,
      cliType: (typeof metadata.cliType === 'string' ? metadata.cliType : snapshot.cliType) as CLIType,
      task: this.isDevelopmentTask(metadata.task) ? metadata.task : undefined,
      templateName: typeof metadata.templateName === 'string' ? metadata.templateName : undefined,
      tags: this.normalizeTags(metadata.tags),
    };
    this.sessions.set(snapshot.id, synced);
    return synced;
  }

  private isDevelopmentTask(value: unknown): value is DevelopmentTask {
    if (!value || typeof value !== 'object') {
      return false;
    }

    const task = value as Record<string, unknown>;
    return typeof task.id === 'string'
      && typeof task.description === 'string'
      && typeof task.context === 'string'
      && Array.isArray(task.files);
  }

  private normalizeTags(value: unknown): string[] {
    if (!Array.isArray(value)) {
      return [];
    }

    return value.filter((tag): tag is string => typeof tag === 'string');
  }

  private isTrackedCouncilSession(snapshot: SupervisedSessionSnapshot): boolean {
    return this.sessions.has(snapshot.id) || snapshot.metadata?.councilSession === true;
  }

  private getTrackedSnapshots(): SupervisedSessionSnapshot[] {
    const supervisor = this.getPtySupervisor();
    if (!supervisor) return [];

    const snapshots = supervisor
      .listSessions()
      .filter((snapshot) => this.isTrackedCouncilSession(snapshot));

    for (const snapshot of snapshots) {
      this.syncState(snapshot);
    }

    return snapshots;
  }

  private toSession(snapshot: SupervisedSessionSnapshot): Session {
    const state = this.syncState(snapshot);
    return {
      id: snapshot.id,
      status: this.toCouncilStatus(snapshot.status),
      startedAt: snapshot.startedAt ?? snapshot.createdAt,
      lastActivity: snapshot.lastActivityAt,
      currentTask: state.task?.description,
      logs: snapshot.logs.map((entry) => ({
        timestamp: entry.timestamp,
        level: entry.stream === 'stderr' ? 'error' : entry.stream === 'system' ? 'info' : 'info',
        message: entry.message,
        source: entry.stream,
      })),
      port: typeof snapshot.metadata.port === 'number' ? snapshot.metadata.port : undefined,
      workingDirectory: snapshot.workingDirectory,
      templateName: state.templateName,
      tags: [...state.tags],
    };
  }

  private persistSession(session: Session): void {
    const state = this.sessions.get(session.id);
    const persisted: PersistedSession = {
      id: session.id,
      status: session.status,
      startedAt: session.startedAt,
      lastActivity: session.lastActivity,
      currentTask: session.currentTask,
      port: session.port ?? 0,
      workingDirectory: session.workingDirectory,
      templateName: state?.templateName,
      tags: session.tags,
      metadata: state ? { cliType: state.cliType } : undefined,
    };
    sessionPersistence.persistSession(persisted);
  }

  private async updateMetadata(id: string, patch: Record<string, unknown>): Promise<void> {
    const supervisor = this.getPtySupervisor();
    if (supervisor) {
      supervisor.updateSessionMetadata(id, patch);
    }
  }

  async startSession(task?: DevelopmentTask, options: SessionOptions = {}): Promise<Session> {
    const cliType = options.cliType ?? task?.cliType ?? this.config.sessions.defaultCLI;
    const tool = await this.resolveTool(cliType);
    const workingDirectory = path.resolve(options.workingDirectory ?? process.cwd());
    const env = {
      ...(this.config.environment?.globals ?? {}),
      ...(options.env ?? {}),
    };
    const tags = [...(options.tags ?? [])];
    const supervisor = this.getPtySupervisor();
    if (!supervisor) {
      throw new Error('PtySupervisor unavailable (Core/MCPServer not initialized)');
    }

    const created = await supervisor.createSession({
      name: task?.description || `council-${cliType}`,
      cliType,
      workingDirectory,
      command: tool.command,
      args: tool.args,
      env,
      metadata: {
        councilSession: true,
        cliType,
        task,
        templateName: options.templateName,
        tags,
      },
    });
    const started = await supervisor.startSession(created.id);

    this.sessions.set(started.id, {
      id: started.id,
      cliType,
      task,
      templateName: options.templateName,
      tags,
    });

    const session = this.toSession(started);
    this.persistSession(session);
    this.emit('session_started', session);
    wsManager.notifySessionStarted(session);
    return session;
  }

  async startBulkSessions(count: number, options: BulkSessionRequest = { count: 1 }): Promise<BulkSessionResponse> {
    const sessions: Session[] = [];
    const failed: Array<{ index: number; error: string }> = [];
    const staggerDelayMs = options.staggerDelayMs ?? 0;

    for (let index = 0; index < count; index += 1) {
      try {
        const session = await this.startSession(undefined, {
          tags: options.tags,
          templateName: options.templateName,
          cliType: options.cliType,
        });
        sessions.push(session);

        if (staggerDelayMs > 0 && index < count - 1) {
          await new Promise((resolve) => setTimeout(resolve, staggerDelayMs));
        }
      } catch (error) {
        failed.push({
          index,
          error: error instanceof Error ? error.message : 'Failed to start session',
        });
      }
    }

    return { sessions, failed };
  }

  async stopSession(id: string): Promise<void> {
    const supervisor = this.getPtySupervisor();
    if (!supervisor) throw new Error('PtySupervisor unavailable');

    const snapshot = supervisor.getSession(id);
    if (!snapshot || !this.isTrackedCouncilSession(snapshot)) {
      throw new Error(`Session '${id}' not found`);
    }

    const stopped = await supervisor.stopSession(id);
    const session = this.toSession(stopped);
    this.persistSession(session);
    this.emit('session_stopped', id);
    wsManager.notifySessionStopped(id);
  }

  async stopAllSessions(): Promise<{ stopped: number; failed: number }> {
    let stopped = 0;
    let failed = 0;

    for (const session of this.getAllSessions()) {
      try {
        await this.stopSession(session.id);
        stopped += 1;
      } catch {
        failed += 1;
      }
    }

    return { stopped, failed };
  }

  async resumeSession(id: string): Promise<Session> {
    const supervisor = this.getPtySupervisor();
    if (!supervisor) throw new Error('PtySupervisor unavailable');

    const snapshot = supervisor.getSession(id);
    if (!snapshot || !this.isTrackedCouncilSession(snapshot)) {
      throw new Error(`Session '${id}' not found`);
    }

    const resumed = await supervisor.restartSession(id);
    const session = this.toSession(resumed);
    this.persistSession(session);
    wsManager.notifySessionUpdate(session);
    return session;
  }

  async resumeAllSessions(): Promise<BulkSessionResponse> {
    const sessions: Session[] = [];
    const failed: Array<{ index: number; error: string }> = [];
    const resumable = this.getAllSessions().filter((session) => session.status !== 'running');

    for (const [index, session] of resumable.entries()) {
      try {
        sessions.push(await this.resumeSession(session.id));
      } catch (error) {
        failed.push({
          index,
          error: error instanceof Error ? error.message : 'Failed to resume session',
        });
      }
    }

    return { sessions, failed };
  }

  async deleteSession(id: string): Promise<void> {
    const existing = this.getSession(id);
    if (!existing) {
      throw new Error(`Session '${id}' not found`);
    }

    if (ACTIVE_SESSION_STATUSES.has(existing.status)) {
      await this.stopSession(id);
    }

    this.sessions.delete(id);
    sessionPersistence.removeSession(id);
    this.emit('session_deleted', id);
  }

  async sendGuidance(id: string, guidance: Guidance): Promise<void> {
    const session = this.getSession(id);
    if (!session) {
      throw new Error(`Session '${id}' not found`);
    }

    const message = [
      guidance.approved ? 'Guidance approved.' : 'Guidance requires changes.',
      guidance.feedback.trim(),
      guidance.suggestedNextSteps.length > 0
        ? `Next steps:\n- ${guidance.suggestedNextSteps.join('\n- ')}`
        : '',
    ].filter(Boolean).join('\n\n');

    const supervisor = this.getPtySupervisor();
    if (message && supervisor) {
      await supervisor.sendInput(id, `${message}\n`);
    }

    const now = Date.now();
    await this.updateMetadata(id, {
      lastGuidanceAt: now,
      guidance,
    });

    this.persistSession({
      ...session,
      lastActivity: now,
    });
  }

  getSession(id: string): Session | undefined {
    const supervisor = this.getPtySupervisor();
    if (!supervisor) return undefined;

    const snapshot = supervisor.getSession(id);
    if (!snapshot || !this.isTrackedCouncilSession(snapshot)) {
      return undefined;
    }

    return this.toSession(snapshot);
  }

  getAllSessions(): Session[] {
    return this.getTrackedSnapshots().map((snapshot) => this.toSession(snapshot));
  }

  getActiveSessions(): Session[] {
    return this.getAllSessions().filter((session) => ACTIVE_SESSION_STATUSES.has(session.status));
  }

  getSessionStats() {
    const sessions = this.getAllSessions();
    const byStatus = sessions.reduce<Record<Session['status'], number>>((counts, session) => {
      counts[session.status] += 1;
      return counts;
    }, {
      idle: 0,
      starting: 0,
      running: 0,
      paused: 0,
      stopped: 0,
      error: 0,
      completed: 0,
    });

    return {
      total: sessions.length,
      active: byStatus.running + byStatus.starting,
      byStatus,
    };
  }

  getPersistedSessions(): PersistedSession[] {
    return sessionPersistence.getPersistedSessions();
  }

  getSessionsByTag(tag: string): Session[] {
    return this.getAllSessions().filter((session) => session.tags?.includes(tag));
  }

  getSessionsByTemplate(template: string): Session[] {
    return this.getAllSessions().filter((session) => session.templateName === template);
  }

  getSessionsByCLI(cliType: CLIType): Session[] {
    return this.getAllSessions().filter((session) => this.getSessionCLIType(session.id) === cliType);
  }

  getSessionCLIType(id: string): CLIType | undefined {
    const state = this.sessions.get(id);
    if (state) {
      return state.cliType;
    }

    const supervisor = this.getPtySupervisor();
    if (!supervisor) return undefined;

    const snapshot = supervisor.getSession(id);
    if (!snapshot || !this.isTrackedCouncilSession(snapshot)) {
      return undefined;
    }

    return this.syncState(snapshot).cliType;
  }

  updateSessionTags(id: string, tags: string[]): void {
    const session = this.getSession(id);
    if (!session) {
      throw new Error(`Session '${id}' not found`);
    }

    const supervisor = this.getPtySupervisor();
    if (!supervisor) throw new Error('PtySupervisor unavailable');

    const state = this.sessions.get(id) ?? this.syncState(supervisor.getSession(id)!);
    state.tags = [...tags];
    void this.updateMetadata(id, { tags: state.tags });
    this.persistSession({ ...session, tags: state.tags });
  }

  addSessionTag(id: string, tag: string): void {
    const supervisor = this.getPtySupervisor();
    if (!supervisor) return;

    const state = this.sessions.get(id) ?? this.syncState(supervisor.getSession(id)!);
    if (!state.tags.includes(tag)) {
      this.updateSessionTags(id, [...state.tags, tag]);
    }
  }

  removeSessionTag(id: string, tag: string): void {
    const supervisor = this.getPtySupervisor();
    if (!supervisor) return;

    const state = this.sessions.get(id) ?? this.syncState(supervisor.getSession(id)!);
    this.updateSessionTags(id, state.tags.filter((existing) => existing !== tag));
  }

  getSessionHealth(id: string): SessionHealth | undefined {
    const supervisor = this.getPtySupervisor();
    if (!supervisor) return undefined;

    const snapshot = supervisor.getSession(id);
    if (!snapshot || !this.isTrackedCouncilSession(snapshot)) {
      return undefined;
    }

    const health = supervisor.getSessionHealth(id);
    return {
      status: health.status,
      lastCheck: health.lastCheck,
      consecutiveFailures: health.consecutiveFailures,
      restartCount: health.restartCount,
      lastRestartAt: health.lastRestartAt,
      errorMessage: health.errorMessage,
    };
  }

  getAllSessionHealth(): Map<string, SessionHealth> {
    const result = new Map<string, SessionHealth>();
    for (const session of this.getAllSessions()) {
      const health = this.getSessionHealth(session.id);
      if (health) {
        result.set(session.id, health);
      }
    }
    return result;
  }

  startPolling(): void {
    if (this.pollTimer) {
      return;
    }

    this.pollTimer = setInterval(() => {
      for (const session of this.getAllSessions()) {
        this.persistSession(session);
      }
    }, this.config.sessions.pollInterval);
  }

  async cleanup(): Promise<void> {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
      this.pollTimer = null;
    }
    sessionPersistence.stopAutoSave();
    await this.stopAllSessions();
  }
}

export const sessionManager = new SessionManager();
