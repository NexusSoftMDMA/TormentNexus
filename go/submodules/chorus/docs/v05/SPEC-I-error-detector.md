# Agent I — CLI Failure Detector

You own: `src/daemon/error-detector.ts` (NEW).

**Read first:** `docs/v05/SPEC-llm-shared.md`, then `chorus_cli_failure_modes.md` and `feedback_opencode_session_db_corruption.md`.

## Build

A pollable detector that takes a tmux session pane snapshot and classifies what kind of failure (if any) is happening. Returns a structured event the runner emits as `cli_error`.

### Patterns to detect (observed in production)

```ts
export type CliErrorKind =
  | 'quota_exhausted'       // codex: "You've hit your usage limit. ... try again at <time>."
  | 'token_refresh_lost'    // codex: "access token could not be refreshed because your refresh token was already used"
  | 'mcp_handshake_failed'  // codex: "failed: handshaking with MCP server failed"
  | 'opencode_db_corrupt'   // opencode: "Provider returned error" repeating with no successful response in last N seconds
  | 'cold_start_timeout'    // CLI never showed its prompt within cold_start_timeout_s
  | 'tmux_dead'             // session no longer exists per tmux has-session
  | 'unknown';              // catch-all

export interface CliError {
  kind: CliErrorKind;
  lineage: string;          // 'anthropic' | 'openai' | 'google' | 'xai'
  message: string;          // user-friendly one-liner for the run header
  cta?: string;             // recommended action (e.g. 'Re-authenticate codex')
  detail?: string;          // raw line that triggered the match
  resetAt?: number;         // ms-epoch when retry is plausible (quota only)
}
```

### Pattern definitions

Each is a regex + lineage filter + classifier function.

```ts
const PATTERNS: Array<{
  regex: RegExp;
  appliesTo: (lineage: string) => boolean;
  classify: (match: RegExpExecArray, lineage: string) => CliError;
}> = [
  {
    // Codex quota: parse "try again at Apr 30th, 2026 10:05 PM"
    regex: /You've hit your usage limit\..*?(?:try again at\s+([^\n.]+))?/i,
    appliesTo: l => l === 'openai',
    classify: (m, l) => ({
      kind: 'quota_exhausted',
      lineage: l,
      message: m[1] ? `Codex quota exhausted. Resets ${m[1].trim()}.` : 'Codex quota exhausted.',
      cta: 'Switch to a different codex account, or wait & retry.',
      resetAt: parseResetTime(m[1]),
      detail: m[0],
    }),
  },
  {
    regex: /access token could not be refreshed/i,
    appliesTo: l => l === 'openai',
    classify: (m, l) => ({
      kind: 'token_refresh_lost',
      lineage: l,
      message: 'Codex auth invalidated (token refresh race).',
      cta: 'Re-authenticate codex.',
      detail: m[0],
    }),
  },
  {
    regex: /handshaking with MCP server failed/i,
    appliesTo: l => l === 'openai',
    classify: (m, l) => ({
      kind: 'mcp_handshake_failed',
      lineage: l,
      message: 'Codex MCP startup failed.',
      cta: 'Re-authenticate codex.',
      detail: m[0],
    }),
  },
  // Opencode DB corruption: only flag when "Provider returned error" appears
  // multiple times with no successful response in between. Stateful — see below.
];
```

### Stateful opencode detector

Opencode's "Provider returned error" can be transient. Only flag corruption when:
- Pattern appears ≥ 3 times in last 5 captures (sliding window), AND
- No "## DONE" or successful answer write in same window

Track per-session counters in a small `Map<sessionName, { errCount: number; lastErrAt: number }>`. Reset on any clear success.

When triggered:
```ts
{
  kind: 'opencode_db_corrupt',
  lineage: 'xai',
  message: 'Opencode session DB likely corrupted (Kimi rejecting empty msgs).',
  cta: 'Run `work-fleet-restart kimi deepseek` after wiping ~/.local/share/opencode/opencode.db',
  detail: 'Multiple Provider-returned-error responses in window.',
}
```

### Public API

```ts
export class ErrorDetector {
  /**
   * Feed a fresh capture-pane snapshot. Returns an error if a pattern matched
   * for this session, null otherwise.
   */
  inspect(sessionName: string, lineage: string, paneText: string): CliError | null;

  /** Reset state for a session (call when session is killed/restarted). */
  reset(sessionName: string): void;

  /** Periodic cleanup of stale per-session state. */
  cleanup(maxIdleMs: number): void;
}
```

The runner (Agent H) calls `detector.inspect(...)` periodically (every 2s while waiting on an answer) and forwards any non-null result as a `cli_error` event.

### `parseResetTime` helper

```ts
function parseResetTime(humanTime?: string): number | undefined {
  // "Apr 30th, 2026 10:05 PM" → ms-epoch
  // Best-effort. Returns undefined on parse failure (UI hides the timer).
  if (!humanTime) return undefined;
  // Strip "th"/"st"/"nd"/"rd" ordinal suffixes
  const cleaned = humanTime.replace(/(\d+)(st|nd|rd|th)/i, '$1');
  const t = Date.parse(cleaned);
  return Number.isFinite(t) ? t : undefined;
}
```

## Don't touch

- `src/daemon/tmux.ts` — Agent F
- `src/daemon/agents/` — Agent G
- `src/daemon/runner.ts`, `src/daemon/output-watcher.ts` — Agent H
- `src/daemon/index.ts` route handlers
- UI files

## Acceptance

```bash
pnpm typecheck                       # green
pnpm lint                            # green for your files
```

Bonus: write a small in-memory unit test (Vitest if available, otherwise inline node assertions) feeding sample pane text and asserting the right `CliErrorKind` returns. Sample texts to assert:

```
"You've hit your usage limit. Upgrade to Plus to continue using Codex (https://chatgpt.com/explore/plus), or try again at Apr 30th, 2026 10:05 PM."
→ kind: 'quota_exhausted', resetAt: parseable

"Your access token could not be refreshed because your refresh token was already used. Please log out and sign in again."
→ kind: 'token_refresh_lost'

"failed: handshaking with MCP server failed: Send message error Transport ... Your authentication token has been invalidated"
→ kind: 'mcp_handshake_failed'
```

Document the test cases in your final commit.

**Branch / commit:**
```
feat(daemon): CLI failure detector for quota/token-refresh/MCP/opencode-db (Agent I)
```
