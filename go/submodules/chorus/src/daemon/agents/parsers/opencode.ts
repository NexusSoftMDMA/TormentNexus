/**
 * OpenCode (`opencode run --format json`).
 *
 * OpenCode v1.14+ emits JSON Lines — one event per line: step_start,
 * text, tool calls, step_finish. The `text` events carry the LLM output
 * under `part.text`; `step_finish` carries token counts under
 * `part.tokens`.
 *
 * step_finish shape (verified live 2026-05-03 against deepseek-v4-pro):
 *   { "type": "step_finish",
 *     "part": { "tokens": { "total": <n>, "input": <n>, "output": <n>,
 *                           "reasoning": <n>,
 *                           "cache": { "write": <n>, "read": <n> } },
 *               "cost": <usd> } }
 *
 * Why parseOpencode does NOT emit message_done on step_finish:
 * retroactive PR #25 review (gemini + opencode-deepseek + opencode-kimi)
 * caught that opencode can emit MULTIPLE step_finish events per session
 * (tool-call agents, multi-turn). Per-step message_done made the runner
 * overwrite finalText to ``, fire participant_done multiple times, and
 * replace (not accumulate) usage. The fix lives in parseOpencodeExit,
 * which sees the full stdout once and aggregates every step_finish into
 * a single message_done with summed tokens.
 *
 * Mirrors AgentEvent.message_done.usage:
 *   inputTokens         <- sum of tokens.input
 *   outputTokens        <- sum of tokens.output
 *   cachedInputTokens   <- sum of tokens.cache.read
 * Reasoning + cache.write are dropped today (don't render on the chip).
 */
import type { AgentEvent } from '../types.js';
import { tryJson } from './shared.js';

export function parseOpencode(line: string): AgentEvent[] {
  const obj = tryJson(line) as Record<string, unknown> | undefined;
  if (!obj || obj.type !== 'text') return [];
  const part = obj.part as Record<string, unknown> | undefined;
  const text = part && typeof part.text === 'string' ? part.text : '';
  if (text.length === 0) return [];
  return [{ type: 'text_delta', text }];
}

interface OpencodeUsage {
  inputTokens?: number;
  outputTokens?: number;
  cachedInputTokens?: number;
  costUsd?: number;
}

/**
 * Walk every line of opencode JSON-Lines stdout, sum tokens + USD cost
 * from every `step_finish`, return undefined when no step_finish carried
 * any usable counts.
 *
 * Cost summing rationale: opencode's step_finish carries a per-step
 * `cost` (USD) computed against opencode-go's published per-token
 * pricing. Multi-step sessions (tool calls) accrue cost per step; summing
 * yields total session cost.
 */
function aggregateOpencodeUsage(fullStdout: string): OpencodeUsage | undefined {
  const acc: OpencodeUsage = {};
  let any = false;
  for (const line of fullStdout.split('\n')) {
    const obj = tryJson(line) as Record<string, unknown> | undefined;
    if (!obj || obj.type !== 'step_finish') continue;
    const part = obj.part as Record<string, unknown> | undefined;
    const tokens = part?.tokens as
      | { input?: number; output?: number; cache?: { read?: number } }
      | undefined;
    if (tokens) {
      if (typeof tokens.input === 'number') {
        acc.inputTokens = (acc.inputTokens ?? 0) + tokens.input;
        any = true;
      }
      if (typeof tokens.output === 'number') {
        acc.outputTokens = (acc.outputTokens ?? 0) + tokens.output;
        any = true;
      }
      if (typeof tokens.cache?.read === 'number') {
        acc.cachedInputTokens = (acc.cachedInputTokens ?? 0) + tokens.cache.read;
        any = true;
      }
    }
    // Cost is present on every step_finish opencode emits, independent
    // of whether tokens.* fields were populated. A malformed-tokens-but-
    // known-cost step still represents real spend.
    if (typeof part?.cost === 'number') {
      acc.costUsd = (acc.costUsd ?? 0) + (part.cost as number);
      any = true;
    }
  }
  return any ? acc : undefined;
}

/**
 * OpenCode on-exit handler. Two responsibilities:
 *
 * 1. JSON-Lines path (modern `opencode run --format json`):
 *    parseOpencode already emitted text_delta events; the runner
 *    accumulated them. Emit a single synthetic message_done with
 *    finalText="" (runner falls back to its accumulator) plus the SUM
 *    of step_finish token counts across the whole session.
 *
 * 2. Single-blob path (older opencode builds, fallback shape): parse
 *    the whole stdout as a JSON object, lift `message`/`result`/`output`
 *    as finalText. No usage available.
 *
 * Either way: ONE message_done — never multiple — so the runner's
 * participant_done lifecycle fires exactly once.
 *
 * Always emit message_done when JSON-Lines is detected, even when no
 * step_finish was found. Earlier code returned [] when usage was
 * undefined, which dropped the terminal event entirely and caused the
 * runner's `for await` loop to exit without firing participant_done —
 * the phase then sat in `working` until the watchdog timeout.
 */
export function parseOpencodeExit(fullStdout: string): AgentEvent[] {
  if (fullStdout.trim().length === 0) return [];
  const firstLine = fullStdout.split('\n').find((l) => l.trim().length > 0);
  if (firstLine) {
    const probe = tryJson(firstLine) as Record<string, unknown> | undefined;
    if (probe && typeof probe.type === 'string') {
      const usage = aggregateOpencodeUsage(fullStdout);
      return [
        usage
          ? { type: 'message_done', finalText: '', usage }
          : { type: 'message_done', finalText: '' },
      ];
    }
  }
  const obj = tryJson(fullStdout) as Record<string, unknown> | undefined;
  if (!obj) return [{ type: 'message_done', finalText: fullStdout }];
  const text =
    (typeof obj.message === 'string' && obj.message) ||
    (typeof obj.result === 'string' && obj.result) ||
    (typeof obj.output === 'string' && obj.output) ||
    fullStdout;
  return [{ type: 'message_done', finalText: text }];
}
