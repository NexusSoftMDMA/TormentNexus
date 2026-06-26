/**
 * Gemini CLI (`gemini -p --output-format stream-json`).
 *
 * Real format captured 2026-04-30 from gemini-cli with model
 * gemini-3.1-pro-preview:
 *   {"type":"init", "session_id", "model"}
 *   {"type":"message", "role":"user", "content":"..."}
 *   {"type":"message", "role":"assistant", "content":"<chunk>", "delta":true}
 *   {"type":"result", "status":"success", "stats":{...}}
 *
 * The `result` line carries only stats — final text is the concatenation
 * of all `delta:true` chunks. Runner accumulates from text_delta events
 * and uses that on `message_done` (which we emit with finalText="" so
 * the runner's fallback to `accumulated` kicks in).
 */
import type { AgentEvent } from '../types.js';
import { tryJson } from './shared.js';

export function parseGemini(line: string): AgentEvent[] {
  const obj = tryJson(line) as Record<string, unknown> | undefined;
  if (!obj) return [];

  const t = obj.type;

  if (t === 'message' && obj.role === 'assistant' && obj.delta === true) {
    if (typeof obj.content === 'string' && obj.content.length > 0) {
      return [{ type: 'text_delta', text: obj.content }];
    }
    return [];
  }

  // Tool calls (functionCall). Best-effort detection on common shape variants.
  if (t === 'message' && obj.functionCall) {
    const fc = obj.functionCall as Record<string, unknown>;
    return [
      {
        type: 'tool_call_start',
        tool: typeof fc.name === 'string' ? fc.name : 'unknown',
        input: fc.args,
      },
    ];
  }

  if (t === 'result') {
    const status = obj.status as string | undefined;
    if (status === 'success') {
      // Lift token counts from gemini-cli's `stats` block so the run-page
      // card can render "8.2k tok" alongside duration. Gemini-cli's
      // headless stream uses { stats: { total_tokens, ... } } at top
      // level on the success result line. Best-effort: a missing or
      // malformed stats block falls through to no usage attached.
      const stats = (obj as Record<string, unknown>).stats;
      let usage:
        | { inputTokens?: number; outputTokens?: number }
        | undefined;
      if (stats && typeof stats === 'object') {
        const s = stats as Record<string, unknown>;
        const total = typeof s.total_tokens === 'number' ? s.total_tokens : undefined;
        const input =
          typeof s.input_tokens === 'number'
            ? s.input_tokens
            : typeof s.prompt_tokens === 'number'
              ? s.prompt_tokens
              : undefined;
        const output =
          typeof s.output_tokens === 'number'
            ? s.output_tokens
            : typeof s.completion_tokens === 'number'
              ? s.completion_tokens
              : undefined;
        if (input !== undefined || output !== undefined) {
          usage = {
            ...(input !== undefined ? { inputTokens: input } : {}),
            ...(output !== undefined ? { outputTokens: output } : {}),
          };
        } else if (typeof total === 'number') {
          // No split available — surface total as outputTokens so it
          // still shows up as "Nk tok" on the card.
          usage = { outputTokens: total };
        }
      }
      return [
        {
          type: 'message_done',
          finalText: '',
          ...(usage ? { usage } : {}),
        },
      ];
    }
    const message = extractGeminiErrorMessage(obj, status);
    // Quota exhaustion is the single most common gemini failure on a
    // free or low-tier account (we hit it ourselves during dogfood).
    // Promote it from the generic "gemini_result_error" to a dedicated
    // kind so the cockpit can render a useful card — including the
    // "resets in Nh Nm" hint we extract below.
    if (looksLikeQuotaExhausted(message) || looksLikeQuotaExhausted(JSON.stringify(obj))) {
      const reset = extractResetWindow(message) ?? extractResetWindow(JSON.stringify(obj));
      return [
        {
          type: 'error',
          kind: 'quota_exhausted',
          message: reset
            ? `Gemini quota exhausted — resets in ${reset}.`
            : `Gemini quota exhausted.${message ? ` ${message}` : ''}`,
        },
      ];
    }
    return [
      {
        type: 'error',
        kind: 'gemini_result_error',
        message,
      },
    ];
  }

  // init, user-echo message, anything else — silently ignore.
  return [];
}

/**
 * Pull a human-readable error message out of gemini's `result` line.
 *
 * The CLI nests the actual upstream API error a few layers deep:
 *   { type: 'result', status: 'error',
 *     error: { message: '...', cause: { message: '...', code: 429 } } }
 * — but older builds put it at `obj.error` (string) or `obj.message`
 * (string). Walk the common shapes in order of specificity.
 *
 * Pre-fix this function returned the bare "Gemini result status=error"
 * string for every quota / 5xx / 4xx — useless to the user. Now it
 * returns the actual upstream message when present.
 */
function extractGeminiErrorMessage(
  obj: Record<string, unknown>,
  status: string | undefined,
): string {
  if (typeof obj.error === 'string') return obj.error;
  if (typeof obj.message === 'string') return obj.message;

  const err = obj.error;
  if (err && typeof err === 'object') {
    const e = err as Record<string, unknown>;
    if (typeof e.message === 'string' && e.message.length > 0) return e.message;
    const cause = e.cause;
    if (cause && typeof cause === 'object') {
      const c = cause as Record<string, unknown>;
      if (typeof c.message === 'string' && c.message.length > 0) return c.message;
    }
  }
  return `Gemini result status=${status ?? 'unknown'}`;
}

function looksLikeQuotaExhausted(s: string): boolean {
  if (!s) return false;
  return /quota|exhausted|429|capacity|rate.?limit|QUOTA_EXHAUSTED/i.test(s);
}

/**
 * Extract a "Nh Nm Ns" reset window from gemini's quota error string
 * (formats seen in practice: "resets after 8h14m16s", "in 23m", "1d
 * 2h"). Returns the matched substring or null. Best-effort — when the
 * upstream message format changes we fall through to the generic
 * "quota exhausted" message.
 */
function extractResetWindow(s: string): string | null {
  if (!s) return null;
  const m = s.match(/(?:reset|resets|in)\s*(?:after|in)?\s*(\d+\s*(?:d|h|m|s)\s*)+/i);
  if (!m) return null;
  return m[0]
    .replace(/^(?:reset|resets|in)\s*(?:after|in)?\s*/i, '')
    .trim();
}

/**
 * On-exit hook for the gemini shim. Scans stderr for quota messages
 * the JSON result line doesn't carry — gemini-cli logs the upstream
 * 429 to stderr only when its underlying SDK throws (verified by the
 * stack trace seen in dogfood). Without this the user sees just
 * "Gemini result status=error" and has no idea the issue is
 * "wait 8h for the quota to reset."
 *
 * Returns at most one event. Falls through to no-op when stderr
 * doesn't match — the existing `cli_failed` branch in headless.ts
 * still surfaces the raw tail so we never silently swallow info.
 */
export function parseGeminiExit(
  _fullStdout: string,
  fullStderr: string,
  _code: number | null,
): AgentEvent[] {
  if (!fullStderr) return [];

  // Auth-not-configured. Gemini CLI surfaces this in two formats we've
  // observed:
  //   1. older: "GEMINI_API_KEY environment variable not found. Add that
  //      to your environment and try again..."
  //   2. newer (gemini-cli 0.40.x): "Please set an Auth method in your
  //      ~/.gemini/settings.json or specify one of the following
  //      environment variables before running: GEMINI_API_KEY,
  //      GOOGLE_GENAI_USE_VERTEXAI, GOOGLE_GENAI_USE_GCA"
  // Without a clear branch the runner emits the generic cli_failed with a
  // noisy stderr tail and the user has no idea the fix is "set an env var
  // or run `gemini` for OAuth".
  const apiKeyNotSet =
    /(GEMINI|GOOGLE(?:_GENAI)?)_API_KEY\s+(?:environment variable\s+)?not\s+(?:found|set)/i.test(
      fullStderr,
    );
  const authMethodMissing =
    /set an Auth method|specify one of the following environment variables/i.test(
      fullStderr,
    );
  if (apiKeyNotSet || authMethodMissing) {
    return [
      {
        type: 'error',
        kind: 'auth_error',
        message:
          `Gemini CLI not authenticated — set GEMINI_API_KEY (or run ` +
          `\`gemini\` for the OAuth flow). Get a key at ` +
          `aistudio.google.com/apikey.`,
      },
    ];
  }

  if (looksLikeQuotaExhausted(fullStderr)) {
    const reset = extractResetWindow(fullStderr);
    return [
      {
        type: 'error',
        kind: 'quota_exhausted',
        message: reset
          ? `Gemini quota exhausted — resets in ${reset}.`
          : `Gemini quota exhausted (Google API returned 429).`,
      },
    ];
  }

  return [];
}
