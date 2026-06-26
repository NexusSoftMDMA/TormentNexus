// Tmux session manager interface. Implementation in src/daemon/tmux.ts.

import type { AgentSpawnOptions, AgentShim } from './agents/types.js';

/**
 * Lifecycle states a session can be in.
 * 1. spawning — tmux created, CLI booting (2-5s cold start)
 * 2. active — CLI running a phase
 * 3. awaiting_user — phase blocked on a question/permission
 * 4. terminal — merged/cancelled/errored/idle-timed-out → destroy
 */
export type SessionState = 'spawning' | 'active' | 'awaiting_user' | 'terminal';

export interface SessionHandle {
  /** Tmux session name. ALWAYS prefixed `chorus-<chatId>-...`. */
  name: string;
  /** Owning chat. Reaper uses this to cross-check against DB state. */
  chatId: string;
  /** Phase id this session belongs to. */
  phaseId: string;
  /** Doer / reviewer / etc. */
  role: 'doer' | 'reviewer';
  /** CLI lineage running in the session (for failure-mode pattern matching). */
  lineage: string;
  /** Stable agent name. */
  agentName: string;
  /** ms-since-epoch timestamps. */
  spawnedAt: number;
  lastActivityAt: number;
  state: SessionState;
}

export interface AcquireSessionOptions {
  chatId: string;
  phaseId: string;
  role: 'doer' | 'reviewer';
  /** Round number within the phase, 1-indexed. */
  round: number;
  /** Template-level policy from `phase.iterate`. */
  shareSessionAcrossRounds: boolean;
  shareSessionAcrossPhases: boolean;
  /** The agent shim that will build the launch command. */
  shim: AgentShim;
  /** Spawn options if a fresh session is needed. */
  spawnOpts: AgentSpawnOptions;
  /** Stable agent name (extracted from the shim and phase). */
  agentName: string;
}

export interface TmuxManager {
  /**
   * Acquire a session for a phase round. Honors share-session policy:
   *   - Across rounds: reuse existing session for THIS chat+phase+role+agent if shareSessionAcrossRounds
   *   - Across phases: rare, only when shareSessionAcrossPhases (and the same agent+role)
   *   - Across chats: NEVER. Hard rule per chorus_tmux_session_lifecycle.md
   *
   * If reuse hits, returns the existing handle and updates lastActivityAt.
   * If no reuse, builds the launch command via the shim, runs `tmux new-session`,
   * waits for cold-start, and returns a new handle.
   */
  acquire(opts: AcquireSessionOptions): Promise<SessionHandle>;

  /** Send keys to the session (pre-nudge cleanup). Errors swallowed. */
  sendKeys(sessionName: string, keys: string[]): void;

  /**
   * Paste a multi-byte buffer via tmux load-buffer + paste-buffer.
   * Per-session buffer name to avoid races between parallel chats.
   */
  pasteBuffer(sessionName: string, content: string): void;

  /** Capture the current pane content (for failure-mode tail). */
  capturePane(sessionName: string): string;

  /** List all chorus-* sessions on the host. */
  list(): SessionHandle[];

  /** Force-kill a session. Idempotent. */
  kill(sessionName: string): void;

  /** Mark a session terminal — eligible for reaping next sweep. */
  markTerminal(sessionName: string): void;

  /** Reaper sweep. Run every ~5 min by index.ts. */
  reapOnce(opts: {
    /** chatId → status, used to find orphans. */
    activeChats: Map<string, string>;
    /** Min minutes a session can be `awaiting_user` before reaping. */
    idleDestroyMinutes: number;
  }): { killed: string[] };
}
