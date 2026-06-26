import { describe, test, expect, beforeEach } from 'vitest';

class TestLogRotationService {
  private sessionLogs: Map<string, {
    entries: any[];
    totalPruned: number;
    lastPruneAt: number;
  }> = new Map();

  private config = {
    maxLogsPerSession: 1000,
    maxLogAgeMs: 24 * 60 * 60 * 1000,
    pruneIntervalMs: 60000,
  };

  configure(config: Partial<typeof this.config>): void {
    this.config = { ...this.config, ...config };
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

  addLog(sessionId: string, entry: any): void {
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

  getLogs(sessionId: string): any[] {
    return this.sessionLogs.get(sessionId)?.entries ?? [];
  }

  getLogsWithPagination(
    sessionId: string,
    offset: number = 0,
    limit: number = 100
  ): { logs: any[]; total: number; hasMore: boolean } {
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

  pruneSessionLogs(sessionId: string): number {
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

describe('LogRotationService', () => {
  let service: TestLogRotationService;

  beforeEach(() => {
    service = new TestLogRotationService();
  });

  describe('session registration', () => {
    test('registers new session', () => {
      service.registerSession('session-1');
      
      expect(service.getLogs('session-1')).toEqual([]);
      expect(service.getStats().sessionCount).toBe(1);
    });

    test('does not duplicate existing session', () => {
      service.registerSession('session-1');
      service.addLog('session-1', { timestamp: Date.now(), message: 'test' });
      service.registerSession('session-1');
      
      expect(service.getLogs('session-1')).toHaveLength(1);
    });

    test('unregisters session', () => {
      service.registerSession('session-1');
      service.unregisterSession('session-1');
      
      expect(service.getStats().sessionCount).toBe(0);
    });
  });

  describe('addLog', () => {
    test('adds log to session', () => {
      const entry = { timestamp: Date.now(), message: 'test log' };
      service.addLog('session-1', entry);
      
      expect(service.getLogs('session-1')).toContain(entry);
    });

    test('auto-creates session if not registered', () => {
      service.addLog('new-session', { timestamp: Date.now(), message: 'test' });
      
      expect(service.getStats().sessionCount).toBe(1);
    });

    test('auto-prunes when exceeding threshold', () => {
      service.configure({ maxLogsPerSession: 10 });
      
      for (let i = 0; i < 13; i++) {
        service.addLog('session-1', { timestamp: Date.now(), message: `log ${i}` });
      }
      
      expect(service.getLogs('session-1').length).toBeLessThanOrEqual(10);
    });
  });

  describe('getLogs', () => {
    test('returns empty array for non-existent session', () => {
      expect(service.getLogs('nonexistent')).toEqual([]);
    });

    test('returns all logs for session', () => {
      service.addLog('session-1', { timestamp: 1, message: 'a' });
      service.addLog('session-1', { timestamp: 2, message: 'b' });
      
      expect(service.getLogs('session-1')).toHaveLength(2);
    });
  });

  describe('getLogsWithPagination', () => {
    beforeEach(() => {
      for (let i = 0; i < 50; i++) {
        service.addLog('session-1', { timestamp: Date.now(), message: `log ${i}` });
      }
    });

    test('returns paginated logs', () => {
      const result = service.getLogsWithPagination('session-1', 0, 10);
      
      expect(result.logs).toHaveLength(10);
      expect(result.total).toBe(50);
      expect(result.hasMore).toBe(true);
    });

    test('returns correct offset', () => {
      const result = service.getLogsWithPagination('session-1', 40, 10);
      
      expect(result.logs).toHaveLength(10);
      expect(result.hasMore).toBe(false);
    });

    test('handles empty session', () => {
      const result = service.getLogsWithPagination('nonexistent', 0, 10);
      
      expect(result.logs).toEqual([]);
      expect(result.total).toBe(0);
      expect(result.hasMore).toBe(false);
    });
  });

  describe('clearSessionLogs', () => {
    test('clears all logs and returns count', () => {
      service.addLog('session-1', { timestamp: Date.now(), message: 'a' });
      service.addLog('session-1', { timestamp: Date.now(), message: 'b' });
      
      const cleared = service.clearSessionLogs('session-1');
      
      expect(cleared).toBe(2);
      expect(service.getLogs('session-1')).toEqual([]);
    });

    test('returns 0 for non-existent session', () => {
      expect(service.clearSessionLogs('nonexistent')).toBe(0);
    });

    test('increments totalPruned stat', () => {
      service.addLog('session-1', { timestamp: Date.now(), message: 'a' });
      service.clearSessionLogs('session-1');
      
      expect(service.getStats().totalPruned).toBe(1);
    });
  });

  describe('pruneSessionLogs', () => {
    test('prunes old logs', () => {
      service.configure({ maxLogAgeMs: 1000 });
      
      service.addLog('session-1', { timestamp: Date.now() - 2000, message: 'old' });
      service.addLog('session-1', { timestamp: Date.now(), message: 'new' });
      
      const pruned = service.pruneSessionLogs('session-1');
      
      expect(pruned).toBe(1);
      expect(service.getLogs('session-1')).toHaveLength(1);
    });

    test('prunes excess logs by count', () => {
      service.configure({ maxLogsPerSession: 5 });
      
      for (let i = 0; i < 10; i++) {
        service.addLog('session-1', { timestamp: Date.now(), message: `log ${i}` });
      }
      
      service.pruneSessionLogs('session-1');
      
      expect(service.getLogs('session-1')).toHaveLength(5);
    });

    test('returns 0 for non-existent session', () => {
      expect(service.pruneSessionLogs('nonexistent')).toBe(0);
    });
  });

  describe('getStats', () => {
    test('returns correct stats', () => {
      service.addLog('session-1', { timestamp: Date.now(), message: 'a' });
      service.addLog('session-1', { timestamp: Date.now(), message: 'b' });
      service.addLog('session-2', { timestamp: Date.now(), message: 'c' });
      
      const stats = service.getStats();
      
      expect(stats.sessionCount).toBe(2);
      expect(stats.totalLogs).toBe(3);
      expect(stats.totalPruned).toBe(0);
    });

    test('tracks pruned count', () => {
      service.addLog('session-1', { timestamp: Date.now(), message: 'a' });
      service.clearSessionLogs('session-1');
      service.addLog('session-1', { timestamp: Date.now(), message: 'b' });
      
      const stats = service.getStats();
      
      expect(stats.totalLogs).toBe(1);
      expect(stats.totalPruned).toBe(1);
    });
  });

  describe('configure', () => {
    test('updates configuration', () => {
      service.configure({ maxLogsPerSession: 50 });
      
      for (let i = 0; i < 60; i++) {
        service.addLog('session-1', { timestamp: Date.now(), message: `log ${i}` });
      }
      
      service.pruneSessionLogs('session-1');
      
      expect(service.getLogs('session-1').length).toBeLessThanOrEqual(50);
    });
  });
});
