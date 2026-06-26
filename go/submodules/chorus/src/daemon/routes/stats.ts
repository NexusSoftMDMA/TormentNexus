/**
 * Aggregate stats for the home page.
 *
 * Two data sources:
 *   - chats DB row counts (total, by status, today, week, top template,
 *     avg duration). Cheap — single sqlite roundtrip.
 *   - per-participant `_stats.json` files on disk under
 *     `~/.chorus/chats/<id>/round-<N>/<role-dir>/`. Walked at request
 *     time; for ~hundreds of chats this is sub-second. The on-disk file
 *     is the source of truth (per PR #16) — successful reviewer runs
 *     never write to phase_events, so DB-only would miss most cost.
 *
 * Endpoint stays additive — does not write anything, so we can cache it
 * later if call volume grows.
 */

import type { FastifyInstance } from 'fastify';
import fs from 'fs';
import os from 'os';
import path from 'path';
import { chats } from '../../lib/db/index.js';
import { voices } from '../../lib/db/voices.js';
import { errorResponse, successResponse, type ApiResponse } from '../api-response.js';

const DAY_MS = 24 * 60 * 60 * 1000;

interface StatsSummary {
  /** Total chat rows. */
  totalRuns: number;
  /** Chats created since 00:00 local today. */
  runsToday: number;
  /** Chats created in the last 7 days. */
  runsWeek: number;
  /** Per-status counts (approved / blocked / failed / cancelled / drafting / etc). */
  byStatus: Record<string, number>;
  /** Approval rate over completed (non-drafting) chats, 0-1. */
  approvalRate: number;
  /** Average run duration in ms (finished - created), only for finished chats. */
  avgDurationMs: number;
  /** Most-used template_id + run count. Null when no chats. */
  topTemplate: { id: string; runs: number } | null;
  /** Sum of usage.costUsd across every _stats.json found.
   *  Includes BOTH actual out-of-pocket (openrouter) AND shadow / list-
   *  price-equivalent (claude/codex/gemini/opencode subscription tiers).
   *  Kept for back-compat; new UI should prefer actualCostUsd. */
  totalCostUsd: number;
  /** Sum of cost over chats created today. Same caveat as totalCostUsd. */
  costTodayUsd: number;
  /** Real out-of-pocket spend (HTTP-shim providers — currently openrouter
   *  only). This is what the user is actually being charged. */
  actualCostUsd: number;
  /** Real out-of-pocket today. */
  actualCostTodayUsd: number;
  /** List-price equivalent for subscription-CLI calls
   *  (claude-code, codex-cli, gemini-cli, opencode-cli, kimi-cli). User
   *  doesn't pay this — it's what each call would cost on the underlying
   *  vendor's API at list price. Surfaced as "plan equivalent" so users
   *  can see what their subscription is saving them. */
  shadowCostUsd: number;
  /** Plan-equivalent for chats created today. */
  shadowCostTodayUsd: number;
  /** Sum of usage.inputTokens / outputTokens across every _stats.json. */
  totalTokensIn: number;
  totalTokensOut: number;
  /** Total enabled voices in the reviewer fleet. */
  enabledVoices: number;
  /** Distinct lineages with at least one enabled voice. */
  activeLineages: number;
}

function startOfTodayMs(): number {
  const now = new Date();
  return new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
}

interface StatsFile {
  durationMs?: number;
  usage?: {
    inputTokens?: number;
    outputTokens?: number;
    costUsd?: number;
  };
}

function readStatsFile(filePath: string): StatsFile | null {
  try {
    return JSON.parse(fs.readFileSync(filePath, 'utf-8')) as StatsFile;
  } catch {
    return null;
  }
}

interface ChatStats {
  costUsd: number;
  /** Out-of-pocket portion of costUsd (openrouter etc). */
  actualCostUsd: number;
  /** Subscription/list-price portion of costUsd
   *  (claude-code/codex/gemini/opencode/kimi). */
  shadowCostUsd: number;
  tokensIn: number;
  tokensOut: number;
}

/**
 * Returns true if a participant directory name belongs to a shim that
 * charges the user out-of-pocket for each call (HTTP-dispatched API
 * shims). Kept as a pattern match on the shim's `name` (which forms the
 * middle slug of `reviewer-<name>-<idx>` / `doer-<name>`) so adding a new
 * paid shim is a one-line change here. Anything not matching is
 * considered "shadow" — subscription-tier CLI where the cost is sunk.
 */
const PAID_SHIMS = new Set(['openrouter']);
function isPaidParticipant(participantDirName: string): boolean {
  // Strip the role prefix and trailing -<idx> for reviewers.
  // Doer dir: `doer-<shim>` → shim
  // Reviewer dir: `reviewer-<shim>-<idx>` → shim
  let core = participantDirName;
  if (core.startsWith('reviewer-')) core = core.slice('reviewer-'.length);
  else if (core.startsWith('doer-')) core = core.slice('doer-'.length);
  // Drop a trailing -<digit+>
  core = core.replace(/-\d+$/, '');
  return PAID_SHIMS.has(core);
}

/**
 * Walks every round/<role-dir>/_stats.json under a single chat dir and
 * sums up usage. Returns zeros if the dir doesn't exist or holds no
 * sidecars yet.
 */
function aggregateChatStats(chatDir: string): ChatStats {
  const out: ChatStats = {
    costUsd: 0,
    actualCostUsd: 0,
    shadowCostUsd: 0,
    tokensIn: 0,
    tokensOut: 0,
  };
  if (!fs.existsSync(chatDir)) return out;
  let rounds: string[];
  try {
    rounds = fs.readdirSync(chatDir).filter((f) => f.startsWith('round-'));
  } catch {
    return out;
  }
  for (const round of rounds) {
    const roundDir = path.join(chatDir, round);
    let participantDirs: string[];
    try {
      participantDirs = fs.readdirSync(roundDir);
    } catch {
      continue;
    }
    for (const pd of participantDirs) {
      const statsPath = path.join(roundDir, pd, '_stats.json');
      if (!fs.existsSync(statsPath)) continue;
      const s = readStatsFile(statsPath);
      if (!s?.usage) continue;
      const cost = s.usage.costUsd ?? 0;
      out.costUsd += cost;
      if (isPaidParticipant(pd)) {
        out.actualCostUsd += cost;
      } else {
        out.shadowCostUsd += cost;
      }
      out.tokensIn += s.usage.inputTokens ?? 0;
      out.tokensOut += s.usage.outputTokens ?? 0;
    }
  }
  return out;
}

export function registerStatsRoutes(fastify: FastifyInstance): void {
  fastify.get<{
    Reply: ApiResponse<StatsSummary>;
  }>('/stats', async () => {
    try {
      const allChats = await chats.list();
      const allVoices = await voices.list({ enabled: true });
      const todayStart = startOfTodayMs();
      const weekStart = Date.now() - 7 * DAY_MS;

      const byStatus: Record<string, number> = {};
      const byTemplate: Record<string, number> = {};
      let runsToday = 0;
      let runsWeek = 0;
      let durationSum = 0;
      let durationCount = 0;
      let approvedCount = 0;
      let completedCount = 0;

      for (const c of allChats) {
        byStatus[c.status] = (byStatus[c.status] ?? 0) + 1;
        byTemplate[c.template_id] = (byTemplate[c.template_id] ?? 0) + 1;
        if (c.created_at >= todayStart) runsToday++;
        if (c.created_at >= weekStart) runsWeek++;
        if (c.finished_at) {
          durationSum += c.finished_at - c.created_at;
          durationCount++;
        }
        if (c.status !== 'drafting' && c.status !== 'reviewing') completedCount++;
        if (c.status === 'approved' || c.status === 'merged') approvedCount++;
      }

      const topEntry = Object.entries(byTemplate).sort((a, b) => b[1] - a[1])[0];
      const topTemplate = topEntry ? { id: topEntry[0], runs: topEntry[1] } : null;

      // Walk every chat's on-disk stats to get cost + tokens. Bounded
      // I/O — runs ~O(chats) and each chat has ≤6 sidecar files.
      const chatsRoot = path.join(os.homedir(), '.chorus', 'chats');
      let totalCostUsd = 0;
      let costTodayUsd = 0;
      let actualCostUsd = 0;
      let actualCostTodayUsd = 0;
      let shadowCostUsd = 0;
      let shadowCostTodayUsd = 0;
      let totalTokensIn = 0;
      let totalTokensOut = 0;
      for (const c of allChats) {
        const dir = path.join(chatsRoot, c.id);
        const s = aggregateChatStats(dir);
        totalCostUsd += s.costUsd;
        actualCostUsd += s.actualCostUsd;
        shadowCostUsd += s.shadowCostUsd;
        totalTokensIn += s.tokensIn;
        totalTokensOut += s.tokensOut;
        if (c.created_at >= todayStart) {
          costTodayUsd += s.costUsd;
          actualCostTodayUsd += s.actualCostUsd;
          shadowCostTodayUsd += s.shadowCostUsd;
        }
      }

      const lineages = new Set<string>();
      for (const v of allVoices) lineages.add(v.lineage);

      const summary: StatsSummary = {
        totalRuns: allChats.length,
        runsToday,
        runsWeek,
        byStatus,
        approvalRate: completedCount > 0 ? approvedCount / completedCount : 0,
        avgDurationMs: durationCount > 0 ? Math.round(durationSum / durationCount) : 0,
        topTemplate,
        totalCostUsd,
        costTodayUsd,
        actualCostUsd,
        actualCostTodayUsd,
        shadowCostUsd,
        shadowCostTodayUsd,
        totalTokensIn,
        totalTokensOut,
        enabledVoices: allVoices.length,
        activeLineages: lineages.size,
      };

      return successResponse(summary);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      return errorResponse('stats_error', message);
    }
  });
}
