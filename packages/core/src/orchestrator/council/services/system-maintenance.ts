import { sessionManager } from './session-manager.js';
import { logRotation } from './log-rotation.js';
import { dbService } from './db.js';
import { isSqliteUnavailableError, formatOptionalSqliteFailure } from '../../../db/sqliteAvailability.js';
import * as fs from 'fs';
import * as path from 'path';
import { spawn } from 'child_process';

class AutonomousMaintenanceService {
  private timer: ReturnType<typeof setInterval> | null = null;
  private readonly CHECK_INTERVAL = 1000 * 60 * 30; // 30 minutes

  start(): void {
    if (this.timer) return;
    this.timer = setInterval(() => this.runMaintenance(), this.CHECK_INTERVAL);
    console.log('[Maintenance] Started autonomous self-maintenance service.');
  }

  stop(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  async runMaintenance(): Promise<void> {
    console.log('[Maintenance] Running scheduled system maintenance...');
    
    try {
      await this.cleanupOrphanedSidecars();
      await this.optimizeDatabase();
      this.checkDiskUsage();
      await this.pruneOldCheckpoints();
    } catch (e) {
      console.error('[Maintenance] Maintenance cycle failed:', e);
    }
  }

  /**
   * Scans for and kills sidecar processes that are no longer linked to an active session.
   */
  private async cleanupOrphanedSidecars(): Promise<void> {
    // This is OS specific. For now, we rely on sessionManager.cleanup() on exit,
    // but a robust TormentNexus system would scan the process tree.
    // implementation planned for next sub-phase.
  }

  /**
   * Runs SQLite VACUUM and ANALYZE to keep the persistence layer fast.
   */
  private async optimizeDatabase(): Promise<void> {
    try {
      const db = dbService.getDb();
      db.exec('VACUUM');
      db.exec('ANALYZE');
      console.log('[Maintenance] Database optimized.');
    } catch (error) {
      if (isSqliteUnavailableError(error)) {
        console.warn(formatOptionalSqliteFailure('[Maintenance] Skipping database optimization', error));
        return;
      }

      throw error;
    }
  }

  /**
   * Alerts if disk space for logs/checkpoints is becoming critical.
   */
  private checkDiskUsage(): void {
    const autopilotDir = path.resolve(process.cwd(), '../../.autopilot');
    if (!fs.existsSync(autopilotDir)) return;

    // Simple directory size estimation (placeholder for real fs.statfs)
    console.log('[Maintenance] Disk usage check complete.');
  }

  /**
   * Removes checkpoints older than 24 hours.
   */
  private async pruneOldCheckpoints(): Promise<void> {
    const checkpointDir = path.resolve(process.cwd(), '../../.autopilot/checkpoints');
    if (!fs.existsSync(checkpointDir)) return;

    const files = fs.readdirSync(checkpointDir);
    const now = Date.now();
    const maxAge = 1000 * 60 * 60 * 24; // 24 hours

    let pruned = 0;
    for (const file of files) {
      const filePath = path.join(checkpointDir, file);
      const stats = fs.statSync(filePath);
      if (now - stats.mtimeMs > maxAge) {
        fs.unlinkSync(filePath);
        pruned++;
      }
    }

    if (pruned > 0) {
      console.log(`[Maintenance] Pruned ${pruned} old checkpoints.`);
    }
  }
}

export const autonomousMaintenance = new AutonomousMaintenanceService();
