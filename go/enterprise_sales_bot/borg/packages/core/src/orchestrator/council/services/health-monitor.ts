import type { SessionHealth, SessionHealthStatus, HealthCheckConfig, CrashRecoveryConfig } from './types.js';
import { wsManager } from './ws-manager.js';

interface HealthMonitoredSession {
  id: string;
  port: number;
  health: SessionHealth;
  healthEndpoint: string;
}

type RestartCallback = (sessionId: string, reason: string) => Promise<boolean>;

class SessionHealthMonitor {
  private sessions: Map<string, HealthMonitoredSession> = new Map();
  private checkInterval: ReturnType<typeof setInterval> | null = null;
  private restartCallback: RestartCallback | null = null;

  private config: HealthCheckConfig = {
    enabled: true,
    intervalMs: 10000,
    timeoutMs: 5000,
    maxFailures: 3,
  };

  private recoveryConfig: CrashRecoveryConfig = {
    enabled: true,
    maxRestartAttempts: 3,
    restartDelayMs: 2000,
    backoffMultiplier: 2,
    maxBackoffMs: 30000,
  };

  configure(healthConfig?: Partial<HealthCheckConfig>, recoveryConfig?: Partial<CrashRecoveryConfig>): void {
    if (healthConfig) {
      this.config = { ...this.config, ...healthConfig };
    }
    if (recoveryConfig) {
      this.recoveryConfig = { ...this.recoveryConfig, ...recoveryConfig };
    }
  }

  setRestartCallback(callback: RestartCallback): void {
    this.restartCallback = callback;
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

  start(): void {
    if (!this.config.enabled || this.checkInterval) return;

    this.checkInterval = setInterval(() => {
      this.checkAllSessions();
    }, this.config.intervalMs);

    console.log(`[HealthMonitor] Started with ${this.config.intervalMs}ms interval`);
  }

  stop(): void {
    if (this.checkInterval) {
      clearInterval(this.checkInterval);
      this.checkInterval = null;
    }
  }

  private async checkAllSessions(): Promise<void> {
    const checkPromises = Array.from(this.sessions.values()).map(session => 
      this.checkSession(session)
    );
    await Promise.allSettled(checkPromises);
  }

  private async checkSession(session: HealthMonitoredSession): Promise<void> {
    const previousStatus = session.health.status;

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.config.timeoutMs);

      const response = await fetch(`http://localhost:${session.port}${session.healthEndpoint}`, {
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (response.ok) {
        session.health.consecutiveFailures = 0;
        session.health.status = 'healthy';
        session.health.errorMessage = undefined;
      } else {
        await this.handleFailure(session, `HTTP ${response.status}`);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      await this.handleFailure(session, message);
    }

    session.health.lastCheck = Date.now();

    if (previousStatus !== session.health.status) {
      this.notifyStatusChange(session.id, previousStatus, session.health.status);
    }
  }

  private async handleFailure(session: HealthMonitoredSession, errorMessage: string): Promise<void> {
    session.health.consecutiveFailures++;
    session.health.errorMessage = errorMessage;

    if (session.health.consecutiveFailures >= this.config.maxFailures) {
      const wasUnresponsive = session.health.status === 'unresponsive';
      session.health.status = 'unresponsive';

      if (!wasUnresponsive && this.recoveryConfig.enabled) {
        await this.attemptRecovery(session);
      }
    } else {
      session.health.status = 'degraded';
    }
  }

  private async attemptRecovery(session: HealthMonitoredSession): Promise<void> {
    if (!this.restartCallback) return;

    if (session.health.restartCount >= this.recoveryConfig.maxRestartAttempts) {
      session.health.status = 'crashed';
      wsManager.broadcast({
        type: 'error',
        payload: {
          sessionId: session.id,
          error: `Session crashed after ${session.health.restartCount} restart attempts`,
          recoverable: false,
        },
        timestamp: Date.now(),
      });
      return;
    }

    const delay = Math.min(
      this.recoveryConfig.restartDelayMs * Math.pow(this.recoveryConfig.backoffMultiplier, session.health.restartCount),
      this.recoveryConfig.maxBackoffMs
    );

    console.log(`[HealthMonitor] Attempting recovery for session ${session.id} in ${delay}ms (attempt ${session.health.restartCount + 1})`);

    await new Promise(resolve => setTimeout(resolve, delay));

    const success = await this.restartCallback(session.id, session.health.errorMessage || 'Unresponsive');

    if (success) {
      session.health.restartCount++;
      session.health.lastRestartAt = Date.now();
      session.health.consecutiveFailures = 0;
      session.health.status = 'healthy';
    } else {
      session.health.restartCount++;
    }
  }

  private notifyStatusChange(sessionId: string, from: SessionHealthStatus, to: SessionHealthStatus): void {
    wsManager.broadcast({
      type: 'session_update',
      payload: {
        id: sessionId,
        healthStatus: to,
        healthChange: { from, to },
      },
      timestamp: Date.now(),
    });

    if (to === 'unresponsive' || to === 'crashed') {
      wsManager.broadcast({
        type: 'error',
        payload: {
          sessionId,
          error: `Session became ${to}`,
          recoverable: to === 'unresponsive',
        },
        timestamp: Date.now(),
      });
    }
  }

  async forceCheck(sessionId: string): Promise<SessionHealth | undefined> {
    const session = this.sessions.get(sessionId);
    if (!session) return undefined;

    await this.checkSession(session);
    return session.health;
  }

  markCrashed(sessionId: string, reason: string): void {
    const session = this.sessions.get(sessionId);
    if (!session) return;

    session.health.status = 'crashed';
    session.health.errorMessage = reason;
    session.health.lastCheck = Date.now();

    this.notifyStatusChange(sessionId, session.health.status, 'crashed');
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
}

export const healthMonitor = new SessionHealthMonitor();
