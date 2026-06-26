import type { CouncilDecision, DevelopmentTask } from './types.js';
import { EventEmitter } from 'events';

type VetoAction = 'approve' | 'reject' | 'redebate';

interface PendingDecision {
  id: string;
  task: DevelopmentTask;
  councilDecision: CouncilDecision;
  createdAt: number;
  expiresAt: number;
  status: 'pending' | 'approved' | 'rejected' | 'redebate' | 'expired';
  vetoReason?: string;
}

interface VetoConfig {
  enabled: boolean;
  timeoutMs: number;
  autoApproveOnTimeout: boolean;
  requireVetoForRejections: boolean;
  minConsensusForAutoApprove: number;
}

interface VetoResult {
  action: VetoAction;
  reason?: string;
  modifiedDecision?: Partial<CouncilDecision>;
}

const DEFAULT_CONFIG: VetoConfig = {
  enabled: false,
  timeoutMs: 300000,
  autoApproveOnTimeout: true,
  requireVetoForRejections: false,
  minConsensusForAutoApprove: 0.9,
};

export class HumanVetoService extends EventEmitter {
  private config: VetoConfig;
  private pendingDecisions: Map<string, PendingDecision> = new Map();
  private decisionHistory: PendingDecision[] = [];
  private timeoutHandles: Map<string, ReturnType<typeof setTimeout>> = new Map();

  constructor(config: Partial<VetoConfig> = {}) {
    super();
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  isEnabled(): boolean {
    return this.config.enabled;
  }

  setEnabled(enabled: boolean): void {
    this.config.enabled = enabled;
    if (!enabled) {
      this.clearAllPending();
    }
  }

  getConfig(): VetoConfig {
    return { ...this.config };
  }

  updateConfig(updates: Partial<VetoConfig>): void {
    this.config = { ...this.config, ...updates };
  }

  async submitForVeto(
    task: DevelopmentTask,
    councilDecision: CouncilDecision
  ): Promise<{ requiresVeto: boolean; decisionId?: string }> {
    if (!this.config.enabled) {
      return { requiresVeto: false };
    }

    if (this.shouldAutoApprove(councilDecision)) {
      return { requiresVeto: false };
    }

    const decisionId = `veto-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
    const now = Date.now();

    const pending: PendingDecision = {
      id: decisionId,
      task,
      councilDecision,
      createdAt: now,
      expiresAt: now + this.config.timeoutMs,
      status: 'pending',
    };

    this.pendingDecisions.set(decisionId, pending);

    this.emit('decision_pending', {
      decisionId,
      task,
      councilDecision,
      expiresAt: pending.expiresAt,
      timeoutMs: this.config.timeoutMs,
    });

    const timeoutHandle = setTimeout(() => {
      this.handleTimeout(decisionId);
    }, this.config.timeoutMs);
    this.timeoutHandles.set(decisionId, timeoutHandle);

    return { requiresVeto: true, decisionId };
  }

  async processVeto(decisionId: string, result: VetoResult): Promise<CouncilDecision | null> {
    const pending = this.pendingDecisions.get(decisionId);
    if (!pending || pending.status !== 'pending') {
      return null;
    }

    const timeoutHandle = this.timeoutHandles.get(decisionId);
    if (timeoutHandle) {
      clearTimeout(timeoutHandle);
      this.timeoutHandles.delete(decisionId);
    }

    pending.status = result.action === 'approve' ? 'approved' : 
                     result.action === 'reject' ? 'rejected' : 'redebate';
    pending.vetoReason = result.reason;

    this.pendingDecisions.delete(decisionId);
    this.decisionHistory.push(pending);
    if (this.decisionHistory.length > 100) {
      this.decisionHistory.shift();
    }

    let finalDecision: CouncilDecision;

    switch (result.action) {
      case 'approve':
        finalDecision = {
          ...pending.councilDecision,
          approved: true,
          reasoning: pending.councilDecision.reasoning + 
            `\n\n**Human Veto Override:** APPROVED` + 
            (result.reason ? ` - ${result.reason}` : ''),
        };
        break;

      case 'reject':
        finalDecision = {
          ...pending.councilDecision,
          approved: false,
          reasoning: pending.councilDecision.reasoning + 
            `\n\n**Human Veto Override:** REJECTED` + 
            (result.reason ? ` - ${result.reason}` : ''),
        };
        break;

      case 'redebate':
        this.emit('redebate_requested', {
          decisionId,
          task: pending.task,
          originalDecision: pending.councilDecision,
          reason: result.reason,
        });
        return null;
    }

    this.emit('decision_finalized', {
      decisionId,
      action: result.action,
      originalDecision: pending.councilDecision,
      finalDecision,
      reason: result.reason,
    });

    return finalDecision;
  }

  getPendingDecision(decisionId: string): PendingDecision | undefined {
    return this.pendingDecisions.get(decisionId);
  }

  getAllPending(): PendingDecision[] {
    return Array.from(this.pendingDecisions.values());
  }

  getHistory(limit: number = 50): PendingDecision[] {
    return this.decisionHistory.slice(-limit);
  }

  getStats(): {
    enabled: boolean;
    pendingCount: number;
    historyCount: number;
    config: VetoConfig;
  } {
    return {
      enabled: this.config.enabled,
      pendingCount: this.pendingDecisions.size,
      historyCount: this.decisionHistory.length,
      config: this.getConfig(),
    };
  }

  private shouldAutoApprove(decision: CouncilDecision): boolean {
    if (this.config.requireVetoForRejections && !decision.approved) {
      return false;
    }

    const consensus = decision.weightedConsensus ?? decision.consensus;
    if (consensus >= this.config.minConsensusForAutoApprove) {
      return true;
    }

    return false;
  }

  private handleTimeout(decisionId: string): void {
    const pending = this.pendingDecisions.get(decisionId);
    if (!pending || pending.status !== 'pending') {
      return;
    }

    pending.status = 'expired';
    this.pendingDecisions.delete(decisionId);
    this.timeoutHandles.delete(decisionId);
    this.decisionHistory.push(pending);

    if (this.config.autoApproveOnTimeout) {
      const finalDecision: CouncilDecision = {
        ...pending.councilDecision,
        reasoning: pending.councilDecision.reasoning + 
          `\n\n**Human Veto:** Timed out after ${this.config.timeoutMs / 1000}s - auto-approved`,
      };

      this.emit('decision_timeout', {
        decisionId,
        task: pending.task,
        originalDecision: pending.councilDecision,
        finalDecision,
        autoApproved: true,
      });
    } else {
      this.emit('decision_timeout', {
        decisionId,
        task: pending.task,
        originalDecision: pending.councilDecision,
        finalDecision: null,
        autoApproved: false,
      });
    }
  }

  private clearAllPending(): void {
    for (const [decisionId, handle] of this.timeoutHandles) {
      clearTimeout(handle);
    }
    this.timeoutHandles.clear();

    for (const pending of this.pendingDecisions.values()) {
      pending.status = 'expired';
      this.decisionHistory.push(pending);
    }
    this.pendingDecisions.clear();
  }

  cleanup(): void {
    this.clearAllPending();
    this.removeAllListeners();
  }
}

export const humanVeto = new HumanVetoService();
