import { sessionManager } from './session-manager.js';
import { dbService } from './db.js';
import * as fs from 'fs';
import * as path from 'path';
import type { Session } from './types.js';

interface SessionCheckpoint {
  sessionId: string;
  timestamp: number;
  terminalBuffer: string;
  currentTask?: string;
  workingDirectory: string;
  cliType: string;
}

class CheckpointService {
  private checkpointDir: string;
  private interval: number = 60000; // 1 minute
  private timer: ReturnType<typeof setInterval> | null = null;

  constructor() {
    this.checkpointDir = process.env.TORMENTNEXUS_CHECKPOINT_DIR || path.resolve(process.cwd(), '../../.autopilot/checkpoints');
    this.ensureDirectory();
  }

  private ensureDirectory(): void {
    if (!fs.existsSync(this.checkpointDir)) {
      fs.mkdirSync(this.checkpointDir, { recursive: true });
    }
  }

  start(): void {
    if (this.timer) return;
    this.timer = setInterval(() => this.captureAll(), this.interval);
    console.log('[CheckpointService] Started periodic checkpointing');
  }

  stop(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  async captureAll(): Promise<void> {
    const sessions = sessionManager.getActiveSessions();
    for (const session of sessions) {
      await this.captureSession(session.id);
    }
  }

  async captureSession(sessionId: string): Promise<void> {
    const session = sessionManager.getSession(sessionId);
    if (!session) return;

    // Capture last 50 lines of logs as a pseudo-terminal buffer
    const terminalBuffer = session.logs
      .slice(-50)
      .map(l => l.message)
      .join('\n');

    const checkpoint: SessionCheckpoint = {
      sessionId,
      timestamp: Date.now(),
      terminalBuffer,
      currentTask: session.currentTask,
      workingDirectory: session.workingDirectory || process.cwd(),
      cliType: sessionManager.getSessionCLIType(sessionId) || 'unknown',
    };

    const filePath = path.join(this.checkpointDir, `${sessionId}.json`);
    fs.writeFileSync(filePath, JSON.stringify(checkpoint, null, 2));
  }

  getLatestCheckpoint(sessionId: string): SessionCheckpoint | null {
    const filePath = path.join(this.checkpointDir, `${sessionId}.json`);
    if (!fs.existsSync(filePath)) return null;
    
    try {
      return JSON.parse(fs.readFileSync(filePath, 'utf-8'));
    } catch {
      return null;
    }
  }

  async restoreSessionContext(sessionId: string): Promise<string | null> {
    const checkpoint = this.getLatestCheckpoint(sessionId);
    if (!checkpoint) return null;

    return `RESUMING FROM CHECKPOINT (${new Date(checkpoint.timestamp).toISOString()})\n\n` +
           `Last Task: ${checkpoint.currentTask || 'None'}\n\n` +
           `Recent Terminal Output:\n---\n${checkpoint.terminalBuffer}\n---\n`;
  }
}

export const checkpointService = new CheckpointService();
