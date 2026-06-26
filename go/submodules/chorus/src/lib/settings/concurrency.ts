/**
 * Typed accessor for chorus's concurrency settings — caps on parallel
 * CLI subprocesses, daemon-wide.
 *
 * Two knobs:
 *
 *   - `maxParallelCli` (1..10, default 3): GLOBAL cap on the total
 *     number of local-CLI shim subprocesses (reviewers + doer combined)
 *     in flight across the whole daemon. HTTP-dispatched shims
 *     (openrouter and friends) don't count — they're network calls and
 *     consume zero local CPU/RAM.
 *
 *   - `perCli` (1..5, defaults below): per-binary cap, also daemon-wide.
 *     Lets the user say "max 2 opencode" even when 4+ chats are in
 *     flight — opencode subprocesses are ~450 MB each and a 4-stack hits
 *     swap. Composes with the global cap as `min(global, perCli)`: a
 *     reviewer slot must acquire BOTH a global slot AND a per-CLI slot
 *     before spawning, whichever is tighter is the queue.
 *
 * Why a single setting object instead of two: keeps the YAML/Form view
 * tidy, and lets `mm.update_settings` apply both atomically. Empty
 * `perCli` values fall through to defaults so the user only stores
 * deltas from default.
 *
 * The daemon reads this dynamically each acquire (not per chat or per
 * boot) so settings changes take effect on the next reviewer to start —
 * no daemon restart needed.
 */

import { z } from 'zod';
import { settings } from '../db';

/** CLIs we cap individually. Mirrors the keys in `cli-detect.ts`. */
export const CLI_LINEAGES = [
  'claude-code',
  'codex-cli',
  'gemini-cli',
  'opencode-cli',
  'kimi-cli',
] as const;
export type CliLineageKey = (typeof CLI_LINEAGES)[number];

/**
 * Defaults reflect what we learned during the May 2026 OOM incident:
 * opencode is heaviest (~450 MB / proc), gemini parser leaks listeners
 * under churn so we keep it tight, claude/codex are lighter and can run
 * 3-wide. Adjustable in /settings.
 */
const DEFAULT_PER_CLI: Record<CliLineageKey, number> = {
  'claude-code': 3,
  'codex-cli': 3,
  'gemini-cli': 2,
  'opencode-cli': 2,
  'kimi-cli': 2,
};

const DEFAULT_MAX_PARALLEL_CLI = 3;

const PER_CLI_KEY_SCHEMA = z.enum(CLI_LINEAGES);

export const ConcurrencySchema = z.object({
  maxParallelCli: z.number().int().min(1).max(10).default(DEFAULT_MAX_PARALLEL_CLI),
  perCli: z
    .record(PER_CLI_KEY_SCHEMA, z.number().int().min(1).max(5))
    .default({}),
});

export type ConcurrencyConfig = z.infer<typeof ConcurrencySchema>;

const SETTINGS_KEY = 'concurrency';

/**
 * Resolve the per-CLI cap for a given lineage with default fallback.
 * Centralized so callers don't sprinkle `?? DEFAULT_PER_CLI[k]` checks.
 */
export function resolvePerCliCap(
  config: ConcurrencyConfig,
  lineage: CliLineageKey,
): number {
  return config.perCli[lineage] ?? DEFAULT_PER_CLI[lineage];
}

export async function getConcurrency(): Promise<ConcurrencyConfig> {
  const raw = await settings.get(SETTINGS_KEY);
  if (raw === null) {
    return ConcurrencySchema.parse({});
  }
  // safeParse so a hand-edited bogus value never crashes the runner —
  // fall back to defaults, the cockpit will surface the broken state on
  // next save anyway.
  const result = ConcurrencySchema.safeParse(raw);
  if (!result.success) {
    return ConcurrencySchema.parse({});
  }
  return result.data;
}

export async function setConcurrency(config: ConcurrencyConfig): Promise<void> {
  const validated = ConcurrencySchema.parse(config);
  await settings.set(SETTINGS_KEY, validated);
}

/**
 * Defaults exposed for the cockpit so the form can pre-populate
 * placeholder values matching what the daemon would use if a row is
 * left blank.
 */
export const _defaults = {
  maxParallelCli: DEFAULT_MAX_PARALLEL_CLI,
  perCli: DEFAULT_PER_CLI,
};
