import { spawnSync } from 'child_process';
import type { TmuxManager, SessionHandle, AcquireSessionOptions } from './tmux-types.js';

/**
 * TmuxManagerImpl: Session manager for Chorus CLI integrations.
 *
 * Responsibilities:
 * - Acquire sessions respecting share-session policy (across rounds, rarely across phases)
 * - Spawn fresh tmux sessions via the agent shim's buildLaunchCommand
 * - Provide tmux operations (sendKeys, pasteBuffer, capturePane)
 * - Track sessions in memory + reconcile from `tmux ls` on startup
 * - Reap orphans via reapOnce() called by index.ts every 5 min
 *
 * Hard rules:
 * 1. %q-quote all user/template values at substitution time
 * 2. Never reuse across chats (hard rule)
 * 3. Tag every session with chorus-<chatId>-...
 * 4. Per-session CODEX_HOME for codex (caller/shim handles env-var passing)
 */
export class TmuxManagerImpl implements TmuxManager {
  /** In-memory registry: "chatId:phaseId:role:agentName" → SessionHandle */
  private sessions = new Map<string, SessionHandle>();

  constructor() {
    // Reconcile existing chorus sessions on startup
    this.reconcileExisting();
  }

  /**
   * Scan `tmux ls` for existing chorus-* sessions and rebuild the registry.
   * Recovers state after daemon restart.
   */
  private reconcileExisting(): void {
    try {
      const result = spawnSync('tmux', ['list-sessions', '-F', '#{session_name}'], {
        encoding: 'utf-8',
      });

      if (result.status === 0 && result.stdout) {
        const lines = result.stdout.trim().split('\n').filter((line) => line.length > 0);

        for (const sessionName of lines) {
          if (!sessionName.startsWith('chorus-')) continue;

          // Parse session name: chorus-<chatId>-<phaseId>-<role>-<agentName>
          const parts = sessionName.split('-');
          if (parts.length < 6) continue; // chorus + chatId + phaseId + role + agentName = 5 parts after split

          // Rebuild a minimal handle; agent lineage/name come from the parts
          const chatId = parts[1];
          const phaseId = parts[2];
          const role = (parts[3] as 'doer' | 'reviewer') || 'doer';
          const agentName = parts.slice(4).join('-'); // Re-join in case agent name has hyphens

          const key = `${chatId}:${phaseId}:${role}:${agentName}`;

          // Avoid overwriting if we already have this session tracked
          if (this.sessions.has(key)) continue;

          const handle: SessionHandle = {
            name: sessionName,
            chatId,
            phaseId,
            role,
            lineage: 'anthropic', // Safe default; actual lineage is unknown from tmux metadata
            agentName,
            spawnedAt: Date.now(),
            lastActivityAt: Date.now(),
            state: 'active',
          };

          this.sessions.set(key, handle);
        }
      }
    } catch {
      // Silent fail; tmux may not be available yet
    }
  }

  /**
   * Validate a string against the allowed charset for tmux session names.
   * Tmux allows [a-zA-Z0-9_-].
   */
  private validateNameComponent(value: string, field: string): string {
    if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
      throw new Error(`Invalid ${field}: ${value} contains forbidden characters`);
    }
    return value;
  }

  /**
   * Build the session key used in our in-memory registry.
   */
  private makeSessionKey(
    chatId: string,
    phaseId: string,
    role: 'doer' | 'reviewer',
    agentName: string
  ): string {
    return `${chatId}:${phaseId}:${role}:${agentName}`;
  }

  /**
   * Build the tmux session name from components.
   * Format: chorus-<chatId>-<phaseId>-<role>-<agentName>
   */
  private makeSessionName(
    chatId: string,
    phaseId: string,
    role: 'doer' | 'reviewer',
    agentName: string
  ): string {
    return `chorus-${chatId}-${phaseId}-${role}-${agentName}`;
  }

  /**
   * Acquire a session for a phase round, respecting share-session policy.
   *
   * Decision tree:
   * 1. If shareSessionAcrossRounds && session exists for this chat+phase+role+agent → reuse
   * 2. Else if shareSessionAcrossPhases && session exists for this chat+role+agent on ANY phase → reuse + rename
   * 3. Else spawn fresh
   */
  async acquire(opts: AcquireSessionOptions): Promise<SessionHandle> {
    const {
      chatId,
      phaseId,
      role,
      shareSessionAcrossRounds,
      shareSessionAcrossPhases,
      agentName,
    } = opts;

    // Validate input charset
    this.validateNameComponent(chatId, 'chatId');
    this.validateNameComponent(phaseId, 'phaseId');
    this.validateNameComponent(agentName, 'agentName');

    const key = this.makeSessionKey(chatId, phaseId, role, agentName);

    // Rule 1: Across rounds (round N → N+1) within THIS phase
    if (shareSessionAcrossRounds) {
      const existing = this.sessions.get(key);
      if (existing) {
        existing.lastActivityAt = Date.now();
        return existing;
      }
    }

    // Rule 2: Across phases (rare) — reuse from a previous phase
    if (shareSessionAcrossPhases) {
      for (const [registryKey, handle] of this.sessions) {
        // Match on chat+role+agent, different phase
        if (
          handle.chatId === chatId &&
          handle.role === role &&
          handle.agentName === agentName &&
          handle.phaseId !== phaseId
        ) {
          // Found one on a previous phase — rename and reuse
          const newSessionName = this.makeSessionName(chatId, phaseId, role, agentName);

          try {
            spawnSync('tmux', ['rename-session', '-t', handle.name, newSessionName]);
          } catch {
            // If rename fails, fall through to spawn fresh
            this.sessions.delete(registryKey);
            break;
          }

          // Update the handle
          handle.phaseId = phaseId;
          handle.lastActivityAt = Date.now();
          this.sessions.delete(registryKey);
          this.sessions.set(key, handle);
          return handle;
        }
      }
    }

    // Rule 3: Spawn fresh
    return this.spawnFresh(opts, key);
  }

  /**
   * Spawn a fresh tmux session for the given phase round.
   */
  private async spawnFresh(opts: AcquireSessionOptions, key: string): Promise<SessionHandle> {
    const { chatId, phaseId, role, agentName, shim, spawnOpts } = opts;
    const sessionName = this.makeSessionName(chatId, phaseId, role, agentName);

    // Build launch command via the shim
    const launchCommand = shim.buildLaunchCommand(spawnOpts);

    // Spawn tmux session
    try {
      // Use child_process.spawnSync with args array for safety
      const result = spawnSync('tmux', ['new-session', '-d', '-s', sessionName, launchCommand], {
        encoding: 'utf-8',
      });

      if (result.status !== 0) {
        throw new Error(
          `Tmux spawn failed: status=${result.status}, stderr=${result.stderr || 'unknown'}`
        );
      }
    } catch (error) {
      throw new Error(
        `TmuxSpawnError(code=tmux_unavailable): ${error instanceof Error ? error.message : String(error)}`
      );
    }

    // Cold-start poll: wait up to 8s for the session to be ready
    const startTime = Date.now();
    const timeout = 8000;

    while (Date.now() - startTime < timeout) {
      try {
        const hasResult = spawnSync('tmux', ['has-session', '-t', sessionName], {
          encoding: 'utf-8',
        });

        if (hasResult.status === 0) {
          // Session is live
          const handle: SessionHandle = {
            name: sessionName,
            chatId,
            phaseId,
            role,
            lineage: shim.lineage,
            agentName,
            spawnedAt: Date.now(),
            lastActivityAt: Date.now(),
            state: 'active',
          };

          this.sessions.set(key, handle);
          return handle;
        }
      } catch {
        // Still not ready
      }

      // Poll every 200ms
      await new Promise((resolve) => setTimeout(resolve, 200));
    }

    // Cold-start timeout
    throw new Error(
      `TmuxSpawnError(code=cold_start_timeout): Session ${sessionName} did not become ready within 8s`
    );
  }

  /**
   * Send raw keys to a session (pre-nudge cleanup: Escape, Ctrl-C, /clear, etc.)
   * Errors swallowed per spec.
   */
  sendKeys(sessionName: string, keys: string[]): void {
    try {
      const args = ['send-keys', '-t', sessionName, ...keys];
      spawnSync('tmux', args, { stdio: 'ignore' });
    } catch {
      // Swallow errors — session may be dead or unresponsive
    }
  }

  /**
   * Send a prompt to a CLI's TUI input box.
   *
   * Originally implemented via `tmux load-buffer + paste-buffer`. That works
   * for Claude/Codex/OpenCode/Kimi but Gemini's TUI silently drops paste-
   * buffer events (likely due to its custom terminal-escape handling). The
   * `send-keys -l` approach types each character literally and works
   * universally across all the CLIs we ship — at the cost of slightly more
   * shell argv pressure for very long prompts.
   *
   * Method name kept as `pasteBuffer` for caller compatibility, but the
   * implementation no longer uses tmux's paste mechanism.
   */
  pasteBuffer(sessionName: string, content: string): void {
    try {
      spawnSync('tmux', ['send-keys', '-l', '-t', sessionName, content], {
        stdio: 'ignore',
      });
    } catch {
      // Silent fail — session may be dead or unresponsive
    }
  }

  /**
   * Capture the current pane content (last 200 lines).
   * Used by the failure detector to match error patterns.
   */
  capturePane(sessionName: string): string {
    try {
      const result = spawnSync('tmux', ['capture-pane', '-t', sessionName, '-p', '-S', '-200'], {
        encoding: 'utf-8',
      });

      if (result.status === 0) {
        return result.stdout || '';
      }

      return '';
    } catch {
      return '';
    }
  }

  /**
   * List all chorus-* sessions currently tracked in memory.
   */
  list(): SessionHandle[] {
    return Array.from(this.sessions.values());
  }

  /**
   * Force-kill a session by name. Idempotent.
   */
  kill(sessionName: string): void {
    try {
      spawnSync('tmux', ['kill-session', '-t', sessionName], { stdio: 'ignore' });
    } catch {
      // Session may not exist or be already dead
    }

    // Remove from registry
    for (const [key, handle] of this.sessions) {
      if (handle.name === sessionName) {
        this.sessions.delete(key);
        break;
      }
    }
  }

  /**
   * Mark a session as terminal (eligible for reaping).
   * Useful when a phase finishes but the tmux session hasn't been killed yet.
   */
  markTerminal(sessionName: string): void {
    for (const handle of this.sessions.values()) {
      if (handle.name === sessionName) {
        handle.state = 'terminal';
        break;
      }
    }
  }

  /**
   * Reaper sweep: identify and kill orphan / idle sessions.
   *
   * Kill criteria:
   * 1. Session's chatId is NOT in activeChats
   * 2. Chat status is terminal (merged, cancelled, errored, timed-out)
   * 3. Session state is awaiting_user AND idle > idleDestroyMinutes
   */
  reapOnce(opts: {
    activeChats: Map<string, string>;
    idleDestroyMinutes: number;
  }): { killed: string[] } {
    const { activeChats, idleDestroyMinutes } = opts;
    const killed: string[] = [];
    const now = Date.now();
    const idleThresholdMs = idleDestroyMinutes * 60 * 1000;

    for (const handle of this.sessions.values()) {
      let shouldKill = false;
      let killReason = '';

      // Criterion 1: Chat is not in activeChats map
      if (!activeChats.has(handle.chatId)) {
        shouldKill = true;
        killReason = 'chat_not_active';
      }

      // Criterion 2: Chat status is terminal
      const chatStatus = activeChats.get(handle.chatId);
      if (chatStatus && ['merged', 'cancelled', 'errored', 'timed_out'].includes(chatStatus)) {
        shouldKill = true;
        killReason = `chat_${chatStatus}`;
      }

      // Criterion 3: Session state is terminal
      if (handle.state === 'terminal') {
        shouldKill = true;
        killReason = 'session_terminal';
      }

      // Criterion 4: Session is awaiting_user and idle too long
      if (handle.state === 'awaiting_user' && now - handle.lastActivityAt > idleThresholdMs) {
        shouldKill = true;
        killReason = `idle_${idleDestroyMinutes}m`;
      }

      if (shouldKill) {
        this.kill(handle.name);
        killed.push(`${handle.name}(${killReason})`);
      }
    }

    return { killed };
  }
}
