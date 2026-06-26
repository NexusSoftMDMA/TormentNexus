import { loadConfig, saveConfig } from './config.js';
import { supervisorAnalytics } from './supervisor-analytics.js';
import { sessionManager } from './session-manager.js';
import { council } from './council.js';
import type { DevelopmentTask } from './types.js';
import { spawn } from 'child_process';

export class SelfEvolutionService {
  private weightOptimizationTimer: ReturnType<typeof setInterval> | null = null;
  private readonly OPTIMIZATION_INTERVAL = 1000 * 60 * 60; // 1 hour

  start(): void {
    if (this.weightOptimizationTimer) return;
    this.weightOptimizationTimer = setInterval(() => {
      this.optimizeWeights();
    }, this.OPTIMIZATION_INTERVAL);
    console.log('[SelfEvolution] Started continuous learning processes.');
  }

  stop(): void {
    if (this.weightOptimizationTimer) {
      clearInterval(this.weightOptimizationTimer);
      this.weightOptimizationTimer = null;
    }
  }

  /**
   * Periodically adjusts supervisor weights based on historical accuracy and consensus agreement.
   */
  optimizeWeights(): void {
    const config = loadConfig();
    const metricsMap = supervisorAnalytics.getAllSupervisorMetrics();

    let configChanged = false;

    for (const supervisor of config.council.supervisors) {
      const metrics = metricsMap.get(supervisor.name);
      if (!metrics || metrics.totalVotes < 5) continue; // Need statistically significant sample size

      let newWeight = supervisor.weight || 1.0;

      if (metrics.consensusAgreementRate >= 0.85) {
        newWeight = Math.min(2.0, newWeight + 0.1);
      } else if (metrics.consensusAgreementRate <= 0.50) {
        newWeight = Math.max(0.1, newWeight - 0.1);
      }

      newWeight = Math.round(newWeight * 10) / 10;

      if (supervisor.weight !== newWeight) {
        console.log(`[SelfEvolution] Adjusting weight for ${supervisor.name}: ${supervisor.weight || 1.0} -> ${newWeight}`);
        supervisor.weight = newWeight;
        configChanged = true;
        
        council.setSupervisorWeight(supervisor.name, newWeight);
      }
    }

    if (configChanged) {
      saveConfig(config);
    }
  }

  /**
   * Spawns an internal meta-session to modify the codebase of TormentNexus itself.
   */
  async evolveSystem(description: string): Promise<string> {
    const taskId = `evolve-${Date.now()}`;
    const task: DevelopmentTask = {
      id: taskId,
      description: `[SELF-EVOLUTION] ${description}\n\nYou are modifying the source code of TormentNexus itself. \n\nIMPORTANT INSTRUCTIONS:\n1. Create a new git branch named 'evolve-${taskId}'.\n2. Make the requested changes.\n3. Run 'npm run build' and 'npm run lint' across the monorepo to ensure integrity.\n4. If tests and builds pass, commit the changes and push the branch.`,
      context: 'Modify TormentNexus source code to implement the requested evolution autonomously.',
      files: [],
      cliType: 'claude-code'
    };

    console.log(`[SelfEvolution] Initiating system evolution: ${description}`);

    const session = await sessionManager.startSession(task, {
      cliType: 'claude-code',
      tags: ['meta', 'self-evolution'],
      workingDirectory: process.cwd()
    });

    const { smartPilot } = await import('./smart-pilot.js');
    smartPilot.triggerTask(session.id, task).catch(err => {
      console.error('[SelfEvolution] Evolution task failed:', err);
    });

    return session.id;
  }

  /**
   * Executes a self-test to verify system integrity after an evolution
   */
  async runSelfTest(): Promise<boolean> {
    return new Promise((resolve) => {
      console.log('[SelfEvolution] Running self-test suite (npm run lint)...');
      const proc = spawn('npm', ['run', 'lint'], {
        cwd: process.cwd(),
        shell: true
      });

      proc.on('close', (code) => {
        const passed = code === 0;
        console.log(`[SelfEvolution] Self-test ${passed ? 'passed' : 'failed'} (code ${code})`);
        resolve(passed);
      });
    });
  }
}

export const selfEvolution = new SelfEvolutionService();
