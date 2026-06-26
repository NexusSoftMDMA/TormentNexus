import type { LogEntry, LogRotationConfig } from './types.js';

interface SessionLogs {
  entries: LogEntry[];
  totalPruned: number;
  lastPruneAt: number;
}

class LogRotationService {
  private sessionLogs: Map<string, SessionLogs> = new Map();
  private pruneInterval: ReturnType<typeof setInterval> | null = null;

  private config: LogRotationConfig = {
    maxLogsPerSession: 1000,
    maxLogAgeMs: 24 * 60 * 60 * 1000,
    pruneIntervalMs: 60000,
  };

  configure(config: Partial<LogRotationConfig>): void {
    this.config = { ...this.config, ...config };
  }

  getConfig(): LogRotationConfig {
    return { ...this.config };
  }

  registerSession(sessionId: string): void {
    if (!this.sessionLogs.has(sessionId)) {
      this.sessionLogs.set(sessionId, {
        entries: [],
        totalPruned: 0,
        lastPruneAt: Date.now(),
      });
    }
  }

  unregisterSession(sessionId: string): void {
    this.sessionLogs.delete(sessionId);
  }

  addLog(sessionId: string, entry: LogEntry): void {
    let logs = this.sessionLogs.get(sessionId);
    if (!logs) {
      logs = { entries: [], totalPruned: 0, lastPruneAt: Date.now() };
      this.sessionLogs.set(sessionId, logs);
    }

    logs.entries.push(entry);

    if (logs.entries.length > this.config.maxLogsPerSession * 1.2) {
      this.pruneSessionLogs(sessionId);
    }
  }

  getLogs(sessionId: string): LogEntry[] {
    return this.sessionLogs.get(sessionId)?.entries ?? [];
  }

  getLogsWithPagination(
    sessionId: string,
    offset: number = 0,
    limit: number = 100
  ): { logs: LogEntry[]; total: number; hasMore: boolean } {
    const logs = this.sessionLogs.get(sessionId);
    if (!logs) {
      return { logs: [], total: 0, hasMore: false };
    }

    const sliced = logs.entries.slice(offset, offset + limit);
    return {
      logs: sliced,
      total: logs.entries.length,
      hasMore: offset + limit < logs.entries.length,
    };
  }

  clearSessionLogs(sessionId: string): number {
    const logs = this.sessionLogs.get(sessionId);
    if (!logs) return 0;

    const count = logs.entries.length;
    logs.entries = [];
    logs.totalPruned += count;
    return count;
  }

  private pruneSessionLogs(sessionId: string): number {
    const logs = this.sessionLogs.get(sessionId);
    if (!logs) return 0;

    const now = Date.now();
    const originalLength = logs.entries.length;

    logs.entries = logs.entries.filter(entry => {
      const age = now - entry.timestamp;
      return age < this.config.maxLogAgeMs;
    });

    if (logs.entries.length > this.config.maxLogsPerSession) {
      const overflow = logs.entries.length - this.config.maxLogsPerSession;
      logs.entries = logs.entries.slice(overflow);
    }

    const pruned = originalLength - logs.entries.length;
    logs.totalPruned += pruned;
    logs.lastPruneAt = now;

    return pruned;
  }

  start(): void {
    if (this.pruneInterval) return;

    this.pruneInterval = setInterval(() => {
      this.pruneAllSessions();
    }, this.config.pruneIntervalMs);

    console.log(`[LogRotation] Started with ${this.config.pruneIntervalMs}ms interval, max ${this.config.maxLogsPerSession} logs/session`);
  }

  stop(): void {
    if (this.pruneInterval) {
      clearInterval(this.pruneInterval);
      this.pruneInterval = null;
    }
  }

  private pruneAllSessions(): void {
    let totalPruned = 0;
    for (const sessionId of this.sessionLogs.keys()) {
      totalPruned += this.pruneSessionLogs(sessionId);
    }
    if (totalPruned > 0) {
      console.log(`[LogRotation] Pruned ${totalPruned} log entries across ${this.sessionLogs.size} sessions`);
    }
  }

  getStats(): {
    sessionCount: number;
    totalLogs: number;
    totalPruned: number;
  } {
    let totalLogs = 0;
    let totalPruned = 0;

    for (const logs of this.sessionLogs.values()) {
      totalLogs += logs.entries.length;
      totalPruned += logs.totalPruned;
    }

    return {
      sessionCount: this.sessionLogs.size,
      totalLogs,
      totalPruned,
    };
  }
}

export const logRotation = new LogRotationService();
