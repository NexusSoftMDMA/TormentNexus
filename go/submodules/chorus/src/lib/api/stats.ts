// Stats client — talks to /stats on the daemon. Aggregates everything
// the home page needs in one round-trip. See routes/stats.ts for shape.

import { fetchFromDaemon } from "./client";

export interface StatsSummary {
  totalRuns: number;
  runsToday: number;
  runsWeek: number;
  byStatus: Record<string, number>;
  approvalRate: number;
  avgDurationMs: number;
  topTemplate: { id: string; runs: number } | null;
  /** Combined actual + shadow. Kept for back-compat. */
  totalCostUsd: number;
  /** Combined actual + shadow for chats created today. */
  costTodayUsd: number;
  /** Real out-of-pocket spend (HTTP-shim providers — currently
   *  openrouter only). What the user is actually being charged. */
  actualCostUsd: number;
  actualCostTodayUsd: number;
  /** What subscription-CLI calls would cost at the underlying vendor's
   *  API list price. The user doesn't pay this — subscription covers it.
   *  Surfaced as "plan equivalent" so users can see what they're saving. */
  shadowCostUsd: number;
  shadowCostTodayUsd: number;
  totalTokensIn: number;
  totalTokensOut: number;
  enabledVoices: number;
  activeLineages: number;
}

export async function getStats(): Promise<StatsSummary> {
  return fetchFromDaemon<StatsSummary>("/stats");
}
