/**
 * Codex (`codex exec`).
 *
 * Codex `exec` writes plain stdout — no stream-json. Nothing is emitted
 * during the run (heartbeat keeps the UI alive); on exit emit one
 * `message_done` with the full stdout. Some Codex versions interleave
 * thinking/tool-use markers in stdout; we don't try to parse those.
 *
 * Quota / failure handling: when the user's ChatGPT-subscription Codex
 * account is rate-limited, codex prints "ERROR: You've hit your usage
 * limit" to STDERR (not stdout) and exits 1. Without detection we
 * silently wrote a 0-byte answer.md and the reviewer phase looked like
 * it produced nothing. Now: detect the quota line + non-zero exit and
 * surface a `quota_exhausted` error so the runner emits cli_error.
 */
import type { AgentEvent } from '../types.js';

// Anchored to the literal `ERROR:` prefix codex emits. The loose
// alternation /upgrade to plus/i / /try again at/i without an anchor
// would false-match prompts that legitimately echo those phrases (codex
// `exec` echoes the user prompt back into stderr, so a code review brief
// mentioning "try again at midnight" was a real hazard). Round-1
// review-only dogfood (PR #9) flagged this.
const CODEX_QUOTA_LINE = /ERROR:[^\n]*(usage limit|upgrade to plus|try again at)/i;

function looksLikeCodexQuota(text: string): boolean {
  return CODEX_QUOTA_LINE.test(text);
}

export function parseCodex(_line: string): AgentEvent[] {
  return [];
}

export function parseCodexExit(
  fullStdout: string,
  fullStderr = '',
  code: number | null = 0,
): AgentEvent[] {
  const stdoutTrimmed = fullStdout.trim();

  if (code === 0 && stdoutTrimmed.length > 0) {
    return [{ type: 'message_done', finalText: fullStdout }];
  }

  if (looksLikeCodexQuota(fullStderr) || looksLikeCodexQuota(fullStdout)) {
    // Pull the literal ERROR line for a usable message; fall back to a
    // truncated tail so we never lose the signal.
    const errorLine =
      [fullStderr, fullStdout]
        .flatMap((s) => s.split('\n'))
        .find((l) => /ERROR:.*usage limit/i.test(l))
        ?.trim() ?? 'codex usage limit reached';
    return [
      {
        type: 'error',
        kind: 'quota_exhausted',
        message: errorLine,
      },
    ];
  }

  if (code !== null && code !== 0) {
    const tail = (fullStderr.trim() || fullStdout.trim()).slice(-300);
    return [
      {
        type: 'error',
        kind: 'cli_error',
        message: tail.length > 0 ? tail : `codex exited ${code} with no output`,
      },
    ];
  }

  // code===0 + empty stdout — preserve old "emit nothing" behavior.
  return [];
}
