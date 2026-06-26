/**
 * OpenRouter (`POST /api/v1/chat/completions`, stream=true).
 *
 * OpenAI-compatible streaming: response body is `text/event-stream` with
 * `data: {...}\n\n` framed events. Each `data:` payload is a JSON
 * chat-completion-chunk. With `stream_options: {include_usage: true}`,
 * the terminal chunk carries `usage.{prompt_tokens, completion_tokens,
 * cost}` — `cost` is OpenRouter's own field (USD, not in upstream
 * OpenAI). The stream closes with a literal `data: [DONE]` sentinel.
 *
 * Caller pre-strips: split on `\n\n`, strip the `data: ` prefix, then
 * pass each line. Empty / `[DONE]` / non-JSON lines return [] so the
 * caller can pass everything through uniformly.
 *
 * Real shape (verified 2026-05-03 against openrouter.ai/api/v1):
 *   {"id":"gen-...","choices":[{"index":0,"delta":{"content":"Hi"},
 *                                "finish_reason":null}]}
 *   {"id":"gen-...","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
 *   {"id":"gen-...","choices":[],"usage":{"prompt_tokens":12,
 *      "completion_tokens":48,"total_tokens":60,"cost":0.00012}}
 *
 * Emit `text_delta` for each delta.content. The shim's runHeadless
 * aggregates usage onto the message_done — usage often arrives AFTER
 * the finish chunk, so message_done is emitted on the usage chunk
 * rather than on finish_reason.
 */
import type { AgentEvent } from '../types.js';
import { tryJson } from './shared.js';

export function parseOpenRouterSSE(line: string): AgentEvent[] {
  const trimmed = line.trim();
  if (trimmed.length === 0) return [];
  if (trimmed === '[DONE]') return [];
  // OpenRouter passes through OpenAI-style "OPENROUTER PROCESSING"
  // comments on slow upstreams as `: <message>` — these aren't JSON.
  if (trimmed.startsWith(':')) return [];

  const obj = tryJson(trimmed) as Record<string, unknown> | undefined;
  if (!obj) return [];

  // Error envelope (e.g. invalid model, rate limit). OpenRouter wraps
  // these as `{"error":{"message":"...","code":...,"metadata":...}}`.
  if (obj.error && typeof obj.error === 'object') {
    const err = obj.error as Record<string, unknown>;
    const message =
      typeof err.message === 'string' ? err.message : 'OpenRouter stream error';
    const code =
      typeof err.code === 'string' || typeof err.code === 'number'
        ? String(err.code)
        : 'openrouter_error';
    return [{ type: 'error', kind: code, message }];
  }

  const events: AgentEvent[] = [];

  const choices = obj.choices;
  if (Array.isArray(choices) && choices.length > 0) {
    for (const c of choices) {
      if (!c || typeof c !== 'object') continue;
      const choice = c as Record<string, unknown>;
      const delta = (choice.delta as Record<string, unknown> | undefined) ?? {};
      if (typeof delta.content === 'string' && delta.content.length > 0) {
        events.push({ type: 'text_delta', text: delta.content });
      }
      // finish_reason is set ('stop' | 'length' | 'content_filter' |
      // ...) on the terminal chunk. We don't emit message_done here
      // because the usage-bearing chunk often arrives AFTER the finish
      // chunk; the message_done fires on the usage chunk below.
    }
  }

  // Usage chunk — sent LAST when stream_options.include_usage is set.
  // cost is OpenRouter-specific (USD). Emit a synthetic message_done
  // with empty finalText so the runner's accumulator (which holds the
  // assembled finalText) wins, but usage is attached.
  const usage = obj.usage as Record<string, unknown> | undefined;
  if (usage && typeof usage === 'object') {
    const inputTokens =
      typeof usage.prompt_tokens === 'number' ? usage.prompt_tokens : undefined;
    const outputTokens =
      typeof usage.completion_tokens === 'number'
        ? usage.completion_tokens
        : undefined;
    const cost = typeof usage.cost === 'number' ? usage.cost : undefined;
    const u: { inputTokens?: number; outputTokens?: number; costUsd?: number } = {};
    if (inputTokens !== undefined) u.inputTokens = inputTokens;
    if (outputTokens !== undefined) u.outputTokens = outputTokens;
    if (cost !== undefined) u.costUsd = cost;
    if (Object.keys(u).length > 0) {
      events.push({ type: 'message_done', finalText: '', usage: u });
    }
  }

  return events;
}
