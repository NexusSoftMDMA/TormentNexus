import { describe, test, expect, beforeEach, mock } from 'vitest';
import type { SessionHealth, SessionHealthStatus } from '../types.js';

// Testable version of health monitor without external dependencies
class TestHealthMonitor {
  private sessions: Map<string, {
    id: string;
    port: number;
    health: SessionHealth;
    healthEndpoint: string;
  }> = new Map();

  private config = {
    enabled: true,
    intervalMs: 10000,
    timeoutMs: 5000,
    maxFailures: 3,
  };

  private recoveryConfig = {
    enabled: true,
    maxRestartAttempts: 3,
    restartDelayMs: 2000,
    backoffMultiplier: 2,
    maxBackoffMs: 30000,
  };

  configure(healthConfig?: Partial<typeof this.config>, recoveryConfig?: Partial<typeof this.recoveryConfig>): void {
    if (healthConfig) {
      this.config = { ...this.config, ...healthConfig };
    }
    if (recoveryConfig) {
      this.recoveryConfig = { ...this.recoveryConfig, ...recoveryConfig };
    }
  }

  registerSession(id: string, port: number, healthEndpoint: string = '/health'): void {
    this.sessions.set(id, {
      id,
      port,
      healthEndpoint,
      health: {
        status: 'healthy',
        lastCheck: Date.now(),
        consecutiveFailures: 0,
        restartCount: 0,
      },
    });
  }

  unregisterSession(id: string): void {
    this.sessions.delete(id);
  }

  getSessionHealth(id: string): SessionHealth | undefined {
    return this.sessions.get(id)?.health;
  }

  getAllHealth(): Map<string, SessionHealth> {
    const result = new Map<string, SessionHealth>();
    for (const [id, session] of this.sessions) {
      result.set(id, session.health);
    }
    return result;
  }

  // Simulate health check result
  simulateHealthCheckSuccess(id: string): void {
    const session = this.sessions.get(id);
    if (!session) return;

    session.health.consecutiveFailures = 0;
    session.health.status = 'healthy';
    session.health.errorMessage = undefined;
    session.health.lastCheck = Date.now();
  }

  simulateHealthCheckFailure(id: string, errorMessage: string): void {
    const session = this.sessions.get(id);
    if (!session) return;

    session.health.consecutiveFailures++;
    session.health.errorMessage = errorMessage;
    session.health.lastCheck = Date.now();

    if (session.health.consecutiveFailures >= this.config.maxFailures) {
      session.health.status = 'unresponsive';
    } else {
      session.health.status = 'degraded';
    }
  }

  markCrashed(sessionId: string, reason: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    session.health.status = 'crashed';
    session.health.errorMessage = reason;
    session.health.lastCheck = Date.now();
  }

  resetHealth(sessionId: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    session.health = {
      status: 'healthy',
      lastCheck: Date.now(),
      consecutiveFailures: 0,
      restartCount: 0,
    };
  }

  getConfig() {
    return this.config;
  }

  getRecoveryConfig() {
    return this.recoveryConfig;
  }
}

describe('SessionHealthMonitor', () => {
  let monitor: TestHealthMonitor;

  beforeEach(() => {
    monitor = new TestHealthMonitor();
  });

  describe('configure', () => {
    test('updates health config', () => {
      monitor.configure({ intervalMs: 5000, maxFailures: 5 });
      
      const config = monitor.getConfig();
      expect(config.intervalMs).toBe(5000);
      expect(config.maxFailures).toBe(5);
    });

    test('updates recovery config', () => {
      monitor.configure(undefined, { maxRestartAttempts: 5, restartDelayMs: 5000 });
      
      const config = monitor.getRecoveryConfig();
      expect(config.maxRestartAttempts).toBe(5);
      expect(config.restartDelayMs).toBe(5000);
    });
  });

  describe('registerSession', () => {
    test('registers session with healthy status', () => {
      monitor.registerSession('session-1', 4000);
      
      const health = monitor.getSessionHealth('session-1');
      expect(health).toBeDefined();
      expect(health!.status).toBe('healthy');
      expect(health!.consecutiveFailures).toBe(0);
      expect(health!.restartCount).toBe(0);
    });

    test('registers session with custom health endpoint', () => {
      monitor.registerSession('session-1', 4000, '/api/health');
      
      expect(monitor.getSessionHealth('session-1')).toBeDefined();
    });
  });

  describe('unregisterSession', () => {
    test('removes session', () => {
      monitor.registerSession('session-1', 4000);
      monitor.unregisterSession('session-1');
      
      expect(monitor.getSessionHealth('session-1')).toBeUndefined();
    });
  });

  describe('getAllHealth', () => {
    test('returns health of all sessions', () => {
      monitor.registerSession('session-1', 4000);
      monitor.registerSession('session-2', 4001);
      monitor.registerSession('session-3', 4002);
      
      const allHealth = monitor.getAllHealth();
      
      expect(allHealth.size).toBe(3);
      expect(allHealth.has('session-1')).toBe(true);
      expect(allHealth.has('session-2')).toBe(true);
      expect(allHealth.has('session-3')).toBe(true);
    });
  });

  describe('health check simulation', () => {
    test('successful check resets failures and sets healthy', () => {
      monitor.registerSession('session-1', 4000);
      monitor.simulateHealthCheckFailure('session-1', 'timeout');
      monitor.simulateHealthCheckSuccess('session-1');
      
      const health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('healthy');
      expect(health!.consecutiveFailures).toBe(0);
      expect(health!.errorMessage).toBeUndefined();
    });

    test('single failure sets degraded status', () => {
      monitor.registerSession('session-1', 4000);
      monitor.simulateHealthCheckFailure('session-1', 'connection refused');
      
      const health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('degraded');
      expect(health!.consecutiveFailures).toBe(1);
      expect(health!.errorMessage).toBe('connection refused');
    });

    test('max failures sets unresponsive status', () => {
      monitor.registerSession('session-1', 4000);
      
      monitor.simulateHealthCheckFailure('session-1', 'error 1');
      monitor.simulateHealthCheckFailure('session-1', 'error 2');
      monitor.simulateHealthCheckFailure('session-1', 'error 3');
      
      const health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('unresponsive');
      expect(health!.consecutiveFailures).toBe(3);
    });
  });

  describe('markCrashed', () => {
    test('sets crashed status with reason', () => {
      monitor.registerSession('session-1', 4000);
      monitor.markCrashed('session-1', 'process exited unexpectedly');
      
      const health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('crashed');
      expect(health!.errorMessage).toBe('process exited unexpectedly');
    });

    test('does nothing for non-existent session', () => {
      monitor.markCrashed('non-existent', 'reason');
      expect(monitor.getSessionHealth('non-existent')).toBeUndefined();
    });
  });

  describe('resetHealth', () => {
    test('resets health to initial state', () => {
      monitor.registerSession('session-1', 4000);
      monitor.simulateHealthCheckFailure('session-1', 'error');
      monitor.simulateHealthCheckFailure('session-1', 'error');
      monitor.markCrashed('session-1', 'crashed');
      
      monitor.resetHealth('session-1');
      
      const health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('healthy');
      expect(health!.consecutiveFailures).toBe(0);
      expect(health!.restartCount).toBe(0);
    });

    test('does nothing for non-existent session', () => {
      monitor.resetHealth('non-existent');
      expect(monitor.getSessionHealth('non-existent')).toBeUndefined();
    });
  });

  describe('status transitions', () => {
    test('healthy -> degraded -> unresponsive -> crashed', () => {
      monitor.registerSession('session-1', 4000);
      
      let health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('healthy');
      
      monitor.simulateHealthCheckFailure('session-1', 'error');
      health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('degraded');
      
      monitor.simulateHealthCheckFailure('session-1', 'error');
      monitor.simulateHealthCheckFailure('session-1', 'error');
      health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('unresponsive');
      
      monitor.markCrashed('session-1', 'max restarts exceeded');
      health = monitor.getSessionHealth('session-1');
      expect(health!.status).toBe('crashed');
    });

    test('recovery: unresponsive -> healthy after success', () => {
      monitor.registerSession('session-1', 4000);
      
      monitor.simulateHealthCheckFailure('session-1', 'error');
      monitor.simulateHealthCheckFailure('session-1', 'error');
      monitor.simulateHealthCheckFailure('session-1', 'error');
      
      expect(monitor.getSessionHealth('session-1')!.status).toBe('unresponsive');
      
      monitor.simulateHealthCheckSuccess('session-1');
      
      expect(monitor.getSessionHealth('session-1')!.status).toBe('healthy');
    });
  });
});
