/**
 * Claude Code (`claude --print --output-format stream-json --verbose`).
 *
 * Real format captured 2026-04-30 from Claude Code 2.1.123:
 *   {type:"system", subtype:"init"|"status"|"hook_started"|...}
 *   {type:"stream_event", event:{type:"message_start", message:{...}}}
 *   {type:"stream_event", event:{type:"content_block_delta",
 *                                delta:{type:"text_delta", text:"..."}}}
 *   {type:"stream_event", event:{type:"content_block_start",
 *                                content_block:{type:"tool_use", name, input}}}
 *   {type:"stream_event", event:{type:"message_stop"}}
 *   {type:"assistant", message:{content:[{type:"text", text:"..."}]}}
 *   {type:"rate_limit_event", rate_limit_info:{...}}
 *   {type:"result", subtype:"success"|"error", result:"...", is_error,
 *                   total_cost_usd, duration_ms, ...}
 */
import type { AgentEvent } from '../types.js';
import { tryJson } from './shared.js';

export function parseClaude(line: string): AgentEvent[] {
  const obj = tryJson(line) as Record<string, unknown> | undefined;
  if (!obj || typeof obj !== 'object') return [];

  const t = obj.type;

  // Final result line — emit message_done with the assembled text +
  // usage. Claude Code reports `total_cost_usd` (list-price equivalent
  // even on a Pro/Max plan) and a `usage` block with input/output token
  // counts. Without lifting these, every Claude reviewer landed in
  // _stats.json with usage:undefined and the home page's plan-equivalent
  // spend silently lost the largest contributor (Claude is most users'
  // primary voice). The user pays $0 for these on subscription, but the
  // shadow column is the whole point of "plan equiv" — show what the
  // sub is saving them.
  if (t === 'result') {
    const subtype = obj.subtype as string | undefined;
    const isError = obj.is_error as boolean | undefined;
    if (subtype === 'success' && !isError) {
      const usageBlock = obj.usage as Record<string, unknown> | undefined;
      const costUsd =
        typeof obj.total_cost_usd === 'number' ? obj.total_cost_usd : undefined;
      const inputTokens =
        usageBlock && typeof usageBlock.input_tokens === 'number'
          ? usageBlock.input_tokens
          : undefined;
      const outputTokens =
        usageBlock && typeof usageBlock.output_tokens === 'number'
          ? usageBlock.output_tokens
          : undefined;
      const cachedInputTokens =
        usageBlock && typeof usageBlock.cache_read_input_tokens === 'number'
          ? usageBlock.cache_read_input_tokens
          : undefined;
      const usage =
        costUsd !== undefined ||
        inputTokens !== undefined ||
        outputTokens !== undefined
          ? {
              ...(inputTokens !== undefined ? { inputTokens } : {}),
              ...(outputTokens !== undefined ? { outputTokens } : {}),
              ...(cachedInputTokens !== undefined ? { cachedInputTokens } : {}),
              ...(costUsd !== undefined ? { costUsd } : {}),
            }
          : undefined;
      return [
        {
          type: 'message_done',
          finalText: String(obj.result ?? ''),
          ...(usage ? { usage } : {}),
        },
      ];
    }
    return [
      {
        type: 'error',
        kind: 'claude_result_error',
        message: String(obj.result ?? obj.api_error_status ?? 'Claude reported error'),
      },
    ];
  }

  if (t === 'stream_event') {
    const event = (obj.event as Record<string, unknown> | undefined) ?? {};
    const eventType = event.type;

    if (eventType === 'content_block_delta') {
      const delta = (event.delta as Record<string, unknown> | undefined) ?? {};
      if (delta.type === 'text_delta' && typeof delta.text === 'string') {
        return [{ type: 'text_delta', text: delta.text }];
      }
      // Tool input deltas (input_json_delta) aren't surfaced — the
      // tool_call_start carries the initial input snapshot which is
      // enough for UI ("called Read on /path/x.ts").
      return [];
    }

    if (eventType === 'content_block_start') {
      const block = (event.content_block as Record<string, unknown> | undefined) ?? {};
      if (block.type === 'tool_use') {
        return [
          {
            type: 'tool_call_start',
            tool: typeof block.name === 'string' ? block.name : 'unknown',
            input: block.input,
          },
        ];
      }
      return [];
    }

    // We don't emit tool_call_end from Claude's stream — Claude emits a
    // tool_result message later that we'd need to track separately.
    // Skipping for now; UI shows tool_call_start in the trace, which is
    // enough for live progress.
    return [];
  }

  // System (init, hook events, status), assistant (assembled message),
  // rate_limit_event — silently ignored. Future: surface
  // rate_limit_event into the cli-health module so cockpit can show
  // "Claude resets at <time>" without waiting for a quota_exhausted.
  return [];
}
