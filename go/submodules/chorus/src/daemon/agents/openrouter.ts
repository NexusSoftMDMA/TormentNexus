/**
 * OpenRouter HTTP shim — dispatches chat completions over `/api/v1/chat/completions`
 * with `stream=true` and parses the resulting SSE into AgentEvents.
 *
 * Why this exists: PR #27 shipped the inline catalog/voice-insert flow but
 * left runtime dispatch as a follow-up. This module is the follow-up. With
 * it, an `openrouter:<model-id>` voice referenced from a template actually
 * runs (instead of falling through to the wrong CLI shim or erroring).
 *
 * Lineage tag is `'any'` because OpenRouter spans every lineage we score
 * (anthropic / openai / google / meta / moonshot / deepseek / mistral / xai).
 * The voice's REAL lineage lives on the voices table row (set by
 * classifyOpencodeModel during voices.upsert) and is what the runner uses
 * for diversity scoring — picking this shim does not collapse diversity.
 *
 * Dispatch: see `pickShimForVoice` in `agents/index.ts`. When the model id
 * starts with `openrouter:`, the runner picks this shim regardless of
 * `phase.doer.lineage`. Precheck (CLI credential file + quota) is skipped
 * — this shim's auth is the OpenRouter key in the secrets table.
 */

import type {
  AgentShim,
  AgentSpawnOptions,
  AgentNudgeOptions,
  HeadlessSpawnOptions,
  AgentEvent,
} from './types.js';
import { secrets } from '../../lib/db/index.js';
import { recordHealth } from '../../lib/cli-health.js';
import { parseOpenRouterSSE } from './parsers/index.js';

const OPENROUTER_BASE = 'https://openrouter.ai/api/v1';
const DEFAULT_TIMEOUT_MS = 10 * 60 * 1000;

/**
 * Strip the `openrouter:` voice-id prefix if present. The voices table stores
 * ids like `openrouter:anthropic/claude-3.5-sonnet`; OpenRouter's API expects
 * the bare model id `anthropic/claude-3.5-sonnet`.
 */
function stripOpenRouterPrefix(model: string): string {
  return model.startsWith('openrouter:') ? model.slice('openrouter:'.length) : model;
}

export const openrouterShim: AgentShim = {
  lineage: 'any',
  name: 'openrouter',

  buildLaunchCommand(_opts: AgentSpawnOptions): string {
    // OpenRouter is HTTP-only; there is no tmux launch path. The runner must
    // route through `runHeadless`. This stub exists only because the shim
    // contract requires it; calling it indicates a routing bug upstream.
    throw new Error(
      'openrouterShim has no tmux launch path — runner must use runHeadless',
    );
  },

  formatPrompt(_opts: AgentNudgeOptions): string {
    // Same — no file-based prompt nudging. The full prompt flows directly
    // through the chat-completions request body.
    throw new Error(
      'openrouterShim does not use file-based prompt nudging — runHeadless ' +
        'passes promptText into the request body directly',
    );
  },

  runHeadless(opts: HeadlessSpawnOptions): AsyncIterable<AgentEvent> {
    return runOpenRouterStream(opts);
  },

  estimateCostUsd(_input: number, _output: number, _model?: string): number {
    // The shim emits actual cost from OpenRouter's `usage.cost` on the final
    // SSE chunk, so estimation here returns 0. Cost preview UX (cost-before-
    // run) should consult voices.input_cost_per_mtok directly rather than
    // this method.
    return 0;
  },
};

/**
 * Generator that wraps the fetch + SSE loop. Honours abortSignal and a
 * hard timeoutMs (defaults to spawnHeadless's 10-min budget). The model id
 * is auto-stripped of any `openrouter:` prefix.
 *
 * Failure modes mapped to AgentEvents:
 *   - missing API key → `error{kind:'auth_missing'}`
 *   - non-2xx HTTP → `error{kind:'openrouter_<status>'}`
 *   - aborted / timed out → `error{kind:'aborted'|'timeout'}`
 *   - upstream stream error envelope → forwarded with its own kind
 */
async function* runOpenRouterStream(
  opts: HeadlessSpawnOptions,
): AsyncIterable<AgentEvent> {
  const stored = await secrets.get('openrouter');
  const apiKey = stored?.value;
  if (!apiKey) {
    yield {
      type: 'error',
      kind: 'auth_missing',
      message:
        'No OpenRouter API key in secrets. Save one via Settings → OpenRouter ' +
        'before using openrouter:* voices.',
    };
    return;
  }

  const rawModel = opts.model;
  if (!rawModel) {
    yield {
      type: 'error',
      kind: 'validation',
      message: 'OpenRouter dispatch requires an explicit model — none supplied.',
    };
    return;
  }
  const model = stripOpenRouterPrefix(rawModel);

  const timeoutMs = opts.timeoutMs ?? DEFAULT_TIMEOUT_MS;
  // Compose the abort source: outer cancel (chat cancel button, daemon
  // shutdown) AND inner timeout. Either firing cancels the fetch + closes
  // the stream reader. Node's AbortSignal.any was added in 20.3 — chorus
  // engines field already requires Node 20+.
  const timeoutCtl = new AbortController();
  const timeoutHandle = setTimeout(() => timeoutCtl.abort('timeout'), timeoutMs);
  const signals: AbortSignal[] = [timeoutCtl.signal];
  if (opts.abortSignal) signals.push(opts.abortSignal);
  const composed = AbortSignal.any(signals);

  let accumulated = '';
  let usage:
    | { inputTokens?: number; outputTokens?: number; costUsd?: number }
    | undefined;
  let finishedNaturally = false;

  try {
    const res = await fetch(`${OPENROUTER_BASE}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${apiKey}`,
        // OpenRouter recommends a referer / X-Title for ranking + analytics.
        // Self-attributing keeps chorus calls grouped on the user's dashboard.
        'HTTP-Referer': 'https://chorus.codes',
        'X-Title': 'Chorus',
      },
      body: JSON.stringify({
        model,
        messages: [{ role: 'user', content: opts.promptText }],
        stream: true,
        // Forces the terminal usage chunk so we get prompt/completion/cost
        // attribution. Without this, OpenRouter omits usage from the stream.
        stream_options: { include_usage: true },
      }),
      signal: composed,
    });

    if (!res.ok) {
      // Try to lift a structured error from the body before falling back to
      // the raw status code. Body is small for error responses.
      let errMessage = `OpenRouter returned ${res.status}`;
      let rawBody = '';
      try {
        rawBody = await res.text();
        const parsed = JSON.parse(rawBody) as { error?: { message?: string } };
        if (parsed.error?.message) errMessage = parsed.error.message;
        else if (rawBody.length > 0 && rawBody.length < 500) errMessage = rawBody;
      } catch {
        /* keep status-code message */
      }
      // Structured daemon-log line — grep-friendly so users can post-mortem
      // an openrouter failure (404 model id, 402 credits, 429 rate-limit)
      // without inspecting per-chat _attempts.jsonl files.
      console.warn(
        `[openrouter] dispatch failed model=${model} status=${res.status} ` +
          `kind=openrouter_${res.status} message=${JSON.stringify(errMessage).slice(0, 300)} ` +
          `body_head=${JSON.stringify(rawBody.slice(0, 200))}`,
      );
      yield {
        type: 'error',
        kind: `openrouter_${res.status}`,
        message: errMessage,
      };
      return;
    }

    if (!res.body) {
      yield {
        type: 'error',
        kind: 'openrouter_no_body',
        message: 'OpenRouter response had no body — cannot stream.',
      };
      return;
    }

    // SSE parsing: split on `\n\n` event boundaries, strip the `data: `
    // prefix, hand each payload to parseOpenRouterSSE. Buffer carries
    // partial events across chunk boundaries.
    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });

      // Drain complete events from buffer. SSE delimiter is a blank line
      // (\n\n). Some servers use \r\n\r\n — handle both.
      let boundary: number;
      while (
        (boundary = findEventBoundary(buffer)) !== -1
      ) {
        const rawEvent = buffer.slice(0, boundary);
        buffer = buffer.slice(boundary).replace(/^[\r\n]+/, '');

        // An event may contain multiple `data:` lines (per SSE spec these
        // concatenate with `\n`). For chat-completions chunks each event is
        // a single data line, but be defensive.
        const dataLines = rawEvent
          .split('\n')
          .filter((l) => l.startsWith('data:'))
          .map((l) => l.slice('data:'.length).replace(/^ /, ''));
        const payload = dataLines.join('\n');
        if (payload.length === 0) continue;

        for (const ev of parseOpenRouterSSE(payload)) {
          if (ev.type === 'text_delta') {
            accumulated += ev.text;
            yield ev;
          } else if (ev.type === 'message_done') {
            // Parser emits this only for usage-bearing chunks. Capture the
            // usage; the real terminal message_done with finalText is
            // emitted at end-of-stream so the runner sees one consolidated
            // event with both finalText AND usage.
            usage = ev.usage;
          } else if (ev.type === 'error') {
            yield ev;
            return;
          } else {
            yield ev;
          }
        }
      }
    }

    finishedNaturally = true;
  } catch (err) {
    const aborted = composed.aborted;
    const reason = composed.aborted ? composed.reason : undefined;
    if (aborted && reason === 'timeout') {
      console.warn(
        `[openrouter] dispatch timed out model=${model} timeout_ms=${timeoutMs}`,
      );
      yield {
        type: 'error',
        kind: 'timeout',
        message: `OpenRouter dispatch exceeded ${Math.round(timeoutMs / 1000)}s.`,
      };
    } else if (aborted) {
      // Cancellation is a user/runner action, not a bug — log at debug
      // level only (still in daemon log via console.log) so a busy
      // daemon's log stays scannable for real failures.
      console.log(`[openrouter] dispatch aborted model=${model}`);
      yield {
        type: 'error',
        kind: 'aborted',
        message: 'OpenRouter dispatch was cancelled.',
      };
    } else {
      const message = err instanceof Error ? err.message : String(err);
      console.warn(
        `[openrouter] dispatch network error model=${model} message=${JSON.stringify(message).slice(0, 300)}`,
      );
      yield {
        type: 'error',
        kind: 'openrouter_fetch_failed',
        message: `Network error: ${message}`,
      };
    }
    return;
  } finally {
    clearTimeout(timeoutHandle);
  }

  if (finishedNaturally) {
    // Clear any stale auth/quota/rate-limit state — the dispatch just
    // succeeded, so any prior badge on the home OpenRouter card no
    // longer reflects reality. Best-effort.
    recordHealth({
      lineage: 'openrouter',
      status: 'healthy',
    }).catch(() => {
      /* health write is informational; ignore */
    });
    yield {
      type: 'message_done',
      finalText: accumulated,
      ...(usage ? { usage } : {}),
    };
  }
}

/**
 * Find the next event boundary in the SSE buffer (`\n\n` or `\r\n\r\n`),
 * returning the index AT the boundary (caller slices [0, boundary] to
 * extract the event, then advances past the boundary's newlines). Returns
 * -1 when no complete event is buffered.
 */
function findEventBoundary(buf: string): number {
  const lf = buf.indexOf('\n\n');
  const crlf = buf.indexOf('\r\n\r\n');
  if (lf === -1) return crlf;
  if (crlf === -1) return lf;
  return Math.min(lf, crlf);
}
