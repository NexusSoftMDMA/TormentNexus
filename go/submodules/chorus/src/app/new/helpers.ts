import type { Template } from "@/lib/types";

export interface Attachment {
  id: string;
  name: string;
  kind: "file" | "diff" | "url";
  size?: string;
}

/**
 * Pick the first meaningful line from a review-only artifact so the chat
 * title reflects what the user pasted instead of a static framing prompt.
 * Skips fence markers (``` / ~~~) and pure-whitespace lines, then truncates
 * to ~80 chars on a word boundary so it slugs cleanly. Falls back to the
 * static brief when nothing usable is found.
 */
export function deriveReviewOnlyTitle(artifact: string): string {
  const fallback = "Review the supplied artifact independently.";
  if (!artifact) return fallback;
  const lines = artifact.split("\n");
  for (const raw of lines) {
    const line = raw.trim();
    if (line.length === 0) continue;
    if (line.startsWith("```") || line.startsWith("~~~")) continue;
    const max = 80;
    if (line.length <= max) return line;
    const cut = line.slice(0, max);
    const lastSpace = cut.lastIndexOf(" ");
    return (lastSpace > 40 ? cut.slice(0, lastSpace) : cut) + "…";
  }
  return fallback;
}

export interface CostEstimate {
  usd: number;
  usdRangeMax: number;
  inputTokens: number;
  reviewerCount: number;
  maxRounds: number;
}

/**
 * Fallback baseline when a template doesn't declare its own
 * `estimatedBaselineTokens`. Approximates the prompt scaffolding
 * (system prompt + persona block + ask framing) the daemon prepends
 * before the user's artifact/work — a per-template constant that lives
 * on the template now; this is the floor when it's missing.
 */
export const DEFAULT_BASELINE_TOKENS = 800;

/**
 * Rough cost heuristic. Two refinements over the v0.6 version:
 *   1. Multiplies by template.maxRounds so users see the worst-case
 *      cost when reviewers disagree and trigger retries.
 *   2. Returns `usdRangeMax` for the upper bound so the UI can render a
 *      range like "$0.30 – $0.90 (with retries)" instead of a single
 *      misleading number.
 *
 * Subscription mode is applied at render time, not here — keep the
 * dollar math pure so it stays correct for users on API mode.
 */
export function estimateCost(args: {
  template: Template | undefined;
  prompt: string;
  attachments: Attachment[];
}): CostEstimate {
  const { template, prompt, attachments } = args;
  const reviewerCount =
    template?.phases?.[0]?.reviewer?.candidates?.length ?? 3;
  const maxRounds = Math.max(1, template?.maxRounds ?? 1);
  const promptTokens = Math.ceil(prompt.length / 4);
  const attachTokens = attachments.length * 1500;
  const baseTokens =
    template?.estimatedBaselineTokens ?? DEFAULT_BASELINE_TOKENS;
  const inputTokens = promptTokens + attachTokens + baseTokens;
  const outputTokens = 1200; // estimate per reviewer
  const perReviewerUsd = inputTokens * 0.000003 + outputTokens * 0.000015;
  const single = perReviewerUsd * reviewerCount;
  return {
    usd: single,
    usdRangeMax: single * maxRounds,
    inputTokens,
    reviewerCount,
    maxRounds,
  };
}
