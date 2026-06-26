import type { Session, DevelopmentTask, Guidance, CouncilDecision, TaskPlan } from './types.js';
import { sessionManager } from './session-manager.js';
import { council } from './council.js';
import { wsManager } from './ws-manager.js';
import { loadConfig } from './config.js';
import { autoContinueHooks } from './hooks.js';
import { councilHierarchy } from './council-hierarchy.js';

interface TaskCheckpoint {
  sessionId: string;
  lastCheckedAt: number;
  lastTaskId?: string;
  pendingDebate: boolean;
}

export interface SmartPilotConfig {
  enabled: boolean;
  pollIntervalMs: number;
  autoApproveThreshold: number;
  requireUnanimous: boolean;
  maxAutoApprovals: number;
}

class SmartPilot {
  private config: SmartPilotConfig;
  private checkpoints: Map<string, TaskCheckpoint> = new Map();
  private pollTimer: ReturnType<typeof setInterval> | null = null;
  private autoApprovalCount: Map<string, number> = new Map();
  private activePlans: Map<string, TaskPlan> = new Map();

  constructor() {
    const appConfig = loadConfig();
    this.config = {
      enabled: appConfig.council.smartPilot || false,
      pollIntervalMs: 3000,
      autoApproveThreshold: 0.8,
      requireUnanimous: false,
      maxAutoApprovals: 10,
    };
  }

  setEnabled(enabled: boolean): void {
    this.config.enabled = enabled;
    if (enabled) {
      this.start();
    } else {
      this.stop();
    }
  }

  setAutoApproveThreshold(threshold: number): void {
    this.config.autoApproveThreshold = Math.max(0.5, Math.min(1.0, threshold));
  }

  setRequireUnanimous(require: boolean): void {
    this.config.requireUnanimous = require;
  }

  setMaxAutoApprovals(max: number): void {
    this.config.maxAutoApprovals = Math.max(1, max);
  }

  isEnabled(): boolean {
    return this.config.enabled;
  }

  getConfig(): SmartPilotConfig {
    return { ...this.config };
  }

  async triggerTask(sessionId: string, task: DevelopmentTask): Promise<void> {
    const session = sessionManager.getSession(sessionId);
    if (!session) throw new Error(`Session ${sessionId} not found`);

    // Check if task needs decomposition (simple heuristic: description length or keyword)
    const isComplex = task.description.length > 50 ||
                      task.description.toLowerCase().includes('implement') ||
                      task.description.toLowerCase().includes('create') ||
                      task.description.toLowerCase().includes('refactor');

    if (isComplex) {
      wsManager.notifyLog(sessionId, {
        timestamp: Date.now(),
        level: 'info',
        message: `[SmartPilot] Analyzing task complexity... initiating Swarm decomposition.`,
        source: 'smart-pilot-swarm',
      });

      try {
        const plan = await council.planTask(task);

        wsManager.notifyLog(sessionId, {
          timestamp: Date.now(),
          level: 'info',
          message: `[SmartPilot] Swarm Plan Generated: ${plan.subtasks.length} subtasks. Reasoning: ${plan.reasoning}`,
          source: 'smart-pilot-swarm',
        });

        this.activePlans.set(sessionId, plan);
        await this.executeSwarmPlan(sessionId, task, plan);
        this.activePlans.delete(sessionId);
        return;
      } catch (e) {
        this.activePlans.delete(sessionId);
        console.error('Swarm decomposition failed, falling back to standard execution', e);
      }
    }

    // Process immediately without polling (or fallback)
    await this.runDebateAndRespond(session, task);
  }

  getActivePlans(): Map<string, TaskPlan> {
    return new Map(this.activePlans);
  }

  private async executeSwarmPlan(originalSessionId: string, originalTask: DevelopmentTask, plan: TaskPlan): Promise<void> {
    const completedTasks = new Set<string>();
    const inProgressTasks = new Set<string>();
    const subtasks = [...plan.subtasks];
    
    // Map to store tool-specific sessions to avoid over-spawning
    const toolSessions = new Map<string, string>();
    toolSessions.set(sessionManager.getSessionCLIType(originalSessionId) || 'opencode', originalSessionId);

    while (completedTasks.size < subtasks.length) {
      const readyTasks = subtasks.filter(t => 
        t.status === 'pending' && 
        !inProgressTasks.has(t.id) &&
        t.dependencies.every((depId: string) => completedTasks.has(depId))
      );

      if (readyTasks.length === 0 && inProgressTasks.size === 0) {
        throw new Error('Deadlock detected in swarm plan dependencies');
      }

      if (readyTasks.length === 0) {
        // Wait for some tasks to complete
        await new Promise(resolve => setTimeout(resolve, 1000));
        continue;
      }

      // Execute ready tasks in parallel
      await Promise.all(readyTasks.map(async (subtask) => {
        inProgressTasks.add(subtask.id);
        subtask.status = 'in_progress';

        const cliType = subtask.preferredCLI || 'opencode';
        let targetSessionId = toolSessions.get(cliType);

        if (!targetSessionId) {
          wsManager.notifyLog(originalSessionId, {
            timestamp: Date.now(),
            level: 'info',
            message: `[SmartPilot] Spawning new ${cliType} session for subtask: ${subtask.title}`,
            source: 'smart-pilot-swarm',
          });

          const newSession = await sessionManager.startSession(undefined, {
            cliType,
            tags: ['swarm-worker', `parent:${originalSessionId}`],
            workingDirectory: originalTask.files.length > 0 ? undefined : process.cwd(),
          });
          targetSessionId = newSession.id;
          toolSessions.set(cliType, targetSessionId);
        }

        const session = sessionManager.getSession(targetSessionId)!;

        wsManager.notifyLog(originalSessionId, {
          timestamp: Date.now(),
          level: 'info',
          message: `[SmartPilot] Swarm executing subtask: ${subtask.title} on ${cliType} (${targetSessionId})`,
          source: 'smart-pilot-swarm',
        });

        try {
          await this.runDebateAndRespond(session, {
            id: subtask.id,
            description: subtask.description,
            context: `Parent Task: ${originalTask.description}\n\nSubtask Context: ${subtask.context || ''}`,
            files: originalTask.files,
            cliType,
          });
          subtask.status = 'completed';
          completedTasks.add(subtask.id);
        } catch (error) {
          console.error(`Subtask ${subtask.id} failed:`, error);
          subtask.status = 'failed';
          // For now, we'll stop the whole plan if a subtask fails
          throw error;
        } finally {
          inProgressTasks.delete(subtask.id);
        }
      }));
    }

    wsManager.notifyLog(originalSessionId, {
      timestamp: Date.now(),
      level: 'info',
      message: `[SmartPilot] Swarm Plan Successfully Completed: ${subtasks.length} tasks finished.`,
      source: 'smart-pilot-swarm',
    });
  }

  start(): void {
    if (this.pollTimer) return;
    if (!this.config.enabled) return;

    console.log('[SmartPilot] Starting auto-pilot mode');
    
    this.pollTimer = setInterval(() => {
      this.checkAllSessions();
    }, this.config.pollIntervalMs);
  }

  stop(): void {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
      this.pollTimer = null;
      console.log('[SmartPilot] Stopped');
    }
  }

  private async checkAllSessions(): Promise<void> {
    const activeSessions = sessionManager.getActiveSessions();
    
    for (const session of activeSessions) {
      await this.checkSession(session);
    }
  }

  private async checkSession(session: Session): Promise<void> {
    let checkpoint = this.checkpoints.get(session.id);
    
    if (!checkpoint) {
      checkpoint = {
        sessionId: session.id,
        lastCheckedAt: Date.now(),
        pendingDebate: false,
      };
      this.checkpoints.set(session.id, checkpoint);
    }

    if (checkpoint.pendingDebate) return;

    const task = await this.fetchCurrentTask(session);
    if (!task) return;

    if (task.id === checkpoint.lastTaskId) return;

    checkpoint.lastTaskId = task.id;
    checkpoint.pendingDebate = true;

    try {
      await this.runDebateAndRespond(session, task);
    } finally {
      checkpoint.pendingDebate = false;
      checkpoint.lastCheckedAt = Date.now();
    }
  }

  private async fetchCurrentTask(session: Session): Promise<DevelopmentTask | null> {
    const managed = this.getManagedSession(session.id);
    if (!managed) return null;

    try {
      const res = await fetch(`http://localhost:${managed.port}/current-task`);
      if (!res.ok) return null;
      return await res.json();
    } catch {
      return null;
    }
  }

  private getManagedSession(sessionId: string): { port: number } | undefined {
    const allSessions = sessionManager.getAllSessions();
    const session = allSessions.find(s => s.id === sessionId);
    if (!session) return undefined;

    const basePort = loadConfig().sessions.basePort;
    const idx = allSessions.indexOf(session);
    return { port: basePort + idx };
  }

  private async runDebateAndRespond(session: Session, task: DevelopmentTask): Promise<void> {
    console.log(`[SmartPilot] New task detected in ${session.id}: ${task.description}`);
    
    wsManager.notifyLog(session.id, {
      timestamp: Date.now(),
      level: 'info',
      message: `[SmartPilot] Initiating council debate for: ${task.description}`,
      source: 'smart-pilot',
    });

    const preDebateResult = await autoContinueHooks.execute({
      phase: 'pre-debate',
      session,
      task,
    });
    
    if (!preDebateResult.continue) {
      wsManager.notifyLog(session.id, {
        timestamp: Date.now(),
        level: 'warn',
        message: `[SmartPilot] Debate blocked by hook: ${preDebateResult.reason || 'No reason provided'}`,
        source: 'smart-pilot',
      });
      return;
    }

    let decision = await councilHierarchy.routeTask(task);

    const postDebateResult = await autoContinueHooks.execute({
      phase: 'post-debate',
      session,
      task,
      decision,
    });
    
    if (!postDebateResult.continue) {
      wsManager.notifyLog(session.id, {
        timestamp: Date.now(),
        level: 'warn',
        message: `[SmartPilot] Post-debate hook stopped flow: ${postDebateResult.reason || 'No reason provided'}`,
        source: 'smart-pilot',
      });
      return;
    }
    
    if (postDebateResult.modifiedDecision) {
      decision = postDebateResult.modifiedDecision;
    }

    wsManager.notifyCouncilDecision(session.id, decision);

    let guidance = this.decisionToGuidance(decision, session.id);

    if (guidance.approved) {
      const count = (this.autoApprovalCount.get(session.id) || 0) + 1;
      this.autoApprovalCount.set(session.id, count);

      if (count > this.config.maxAutoApprovals) {
        console.log(`[SmartPilot] Max auto-approvals (${this.config.maxAutoApprovals}) reached for ${session.id}`);
        guidance.approved = false;
        guidance.feedback = `Auto-approval limit reached (${this.config.maxAutoApprovals}). Manual review required.`;
      }
    }

    const preGuidanceResult = await autoContinueHooks.execute({
      phase: 'pre-guidance',
      session,
      task,
      decision,
      guidance,
    });
    
    if (!preGuidanceResult.continue) {
      wsManager.notifyLog(session.id, {
        timestamp: Date.now(),
        level: 'warn',
        message: `[SmartPilot] Guidance blocked by hook: ${preGuidanceResult.reason || 'No reason provided'}`,
        source: 'smart-pilot',
      });
      return;
    }
    
    if (preGuidanceResult.modifiedGuidance) {
      guidance = preGuidanceResult.modifiedGuidance;
    }

    wsManager.notifyLog(session.id, {
      timestamp: Date.now(),
      level: guidance.approved ? 'info' : 'warn',
      message: `[SmartPilot] Decision: ${guidance.approved ? 'AUTO-APPROVED' : 'REQUIRES REVIEW'} (consensus: ${(decision.weightedConsensus || decision.consensus * 100).toFixed(1)}%)`,
      source: 'smart-pilot',
    });

    try {
      await sessionManager.sendGuidance(session.id, guidance);
      
      await autoContinueHooks.execute({
        phase: 'post-guidance',
        session,
        task,
        decision,
        guidance,
      });
    } catch (err) {
      console.error(`[SmartPilot] Failed to send guidance to ${session.id}:`, err);
      
      await autoContinueHooks.execute({
        phase: 'on-error',
        session,
        task,
        decision,
        guidance,
        error: err instanceof Error ? err : new Error(String(err)),
      });
    }
  }

  private decisionToGuidance(decision: CouncilDecision, sessionId: string): Guidance {
    const effectiveConsensus = decision.weightedConsensus ?? decision.consensus;
    
    let canAutoApprove = decision.approved && effectiveConsensus >= this.config.autoApproveThreshold;

    if (this.config.requireUnanimous && decision.votes.some(v => !v.approved)) {
      canAutoApprove = false;
    }

    if (decision.dissent && decision.dissent.length > 0) {
      canAutoApprove = false;
    }

    const suggestedNextSteps: string[] = [];
    
    if (canAutoApprove) {
      suggestedNextSteps.push('Continue with implementation');
    } else {
      suggestedNextSteps.push('Review council feedback');
      if (decision.dissent && decision.dissent.length > 0) {
        suggestedNextSteps.push('Address dissenting concerns');
      }
      suggestedNextSteps.push('Request manual approval if appropriate');
    }

    return {
      approved: canAutoApprove,
      feedback: decision.reasoning,
      suggestedNextSteps,
    };
  }

  resetApprovalCount(sessionId: string): void {
    this.autoApprovalCount.delete(sessionId);
  }

  resetAllApprovalCounts(): void {
    this.autoApprovalCount.clear();
  }

  cleanup(): void {
    this.stop();
    this.checkpoints.clear();
    this.autoApprovalCount.clear();
  }
}

export const smartPilot = new SmartPilot();
