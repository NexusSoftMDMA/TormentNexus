import type { TmuxManager } from './tmux-types.js';
import { clearStaleHealth } from '../lib/cli-health.js';

/**
 * Reaper configuration.
 */
export interface ReaperConfig {
  /** Interval in milliseconds to run reaper sweeps. Default 5 * 60 * 1000 (5 min). */
  intervalMs: number;
  /** Minutes a session can be `awaiting_user` before eligible for reaping. Default 30. */
  idleDestroyMinutes: number;
}

/**
 * Start the reaper service. Returns a stop() function for graceful shutdown.
 *
 * The reaper runs every `intervalMs` and kills:
 * 1. Orphaned sessions (chatId not in activeChats)
 * 2. Sessions whose chat is in a terminal state
 * 3. Sessions in terminal state
 * 4. Sessions in awaiting_user state idle > idleDestroyMinutes
 *
 * Logs one-liner summaries of killed sessions.
 */
export function startReaper(
  mgr: TmuxManager,
  getActiveChats: () => Promise<Map<string, string>>,
  cfg: ReaperConfig
): () => void {
  const interval = setInterval(() => {
    void (async () => {
      try {
        const activeChats = await getActiveChats();
        const result = mgr.reapOnce({
          activeChats,
          idleDestroyMinutes: cfg.idleDestroyMinutes,
        });

        if (result.killed.length > 0) {
          console.log(`[reaper] Killed ${result.killed.length} session(s): ${result.killed.join(', ')}`);
        }

        // Auto-clear stale quota / rate-limit state. Without this, the
        // home-page fleet card stays red after the upstream window passes
        // until the next failed call would otherwise overwrite health.
        const cleared = await clearStaleHealth();
        if (cleared.length > 0) {
          console.log(`[reaper] Cleared stale health: ${cleared.join(', ')}`);
        }
      } catch (error) {
        console.error('[reaper] Error during sweep:', error);
      }
    })();
  }, cfg.intervalMs);

  // Return stop function
  return () => {
    clearInterval(interval);
  };
}
