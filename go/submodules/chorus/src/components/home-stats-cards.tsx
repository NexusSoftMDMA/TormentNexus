/**
 * Six stats cards for the home page. Server component — fetches /stats
 * directly from the daemon. Replaces the "Recent runs" list when there's
 * any chat data to show; falls back to nothing rendered when totalRuns=0
 * so the EmptyHero takes over.
 */

import {
  Activity,
  CheckCircle2,
  Clock,
  Coins,
  Layers,
  Sparkles,
} from "lucide-react";
import type { StatsSummary } from "@/lib/api/stats";

interface Props {
  stats: StatsSummary;
}

export function HomeStatsCards({ stats }: Props) {
  const approvalPct = Math.round(stats.approvalRate * 100);

  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-3">
      <StatCard
        icon={<Activity className="h-3.5 w-3.5" />}
        label="Runs"
        value={stats.runsToday.toString()}
        sub={`${stats.totalRuns} total · ${stats.runsWeek} this week`}
      />
      <StatCard
        icon={<CheckCircle2 className="h-3.5 w-3.5" />}
        label="Approval rate"
        value={`${approvalPct}%`}
        sub={statusBreakdown(stats.byStatus)}
      />
      <StatCard
        icon={<Coins className="h-3.5 w-3.5" />}
        label="Spend today"
        value={formatUsd(stats.actualCostTodayUsd)}
        sub={
          <>
            <div>
              {formatUsd(stats.actualCostUsd)} actual all-time ·{" "}
              {formatTokens(stats.totalTokensIn + stats.totalTokensOut)} tok
            </div>
            <div
              className="text-muted-foreground/70"
              title="What these calls would have cost on each provider's API at list price. You don't pay this — your subscriptions cover it."
            >
              plan equiv: {formatUsd(stats.shadowCostTodayUsd)} today ·{" "}
              {formatUsd(stats.shadowCostUsd)} all-time
            </div>
          </>
        }
      />
      <StatCard
        icon={<Clock className="h-3.5 w-3.5" />}
        label="Avg duration"
        value={formatDuration(stats.avgDurationMs)}
        sub="across finished runs"
      />
      <StatCard
        icon={<Layers className="h-3.5 w-3.5" />}
        label="Reviewer fleet"
        value={`${stats.enabledVoices}`}
        sub={`across ${stats.activeLineages} lineage${stats.activeLineages === 1 ? "" : "s"}`}
      />
      <StatCard
        icon={<Sparkles className="h-3.5 w-3.5" />}
        label="Top template"
        value={stats.topTemplate?.id ?? "—"}
        sub={
          stats.topTemplate
            ? `${stats.topTemplate.runs} run${stats.topTemplate.runs === 1 ? "" : "s"}`
            : "no runs yet"
        }
        mono
      />
    </div>
  );
}

interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  sub: React.ReactNode;
  mono?: boolean;
}

function StatCard({ icon, label, value, sub, mono }: StatCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        {icon}
        {label}
      </div>
      <div
        className={
          mono
            ? "mt-2 truncate font-mono text-base font-semibold"
            : "mt-2 truncate text-2xl font-semibold tabular-nums"
        }
        title={value}
      >
        {value}
      </div>
      <div className="mt-1 text-[11px] text-muted-foreground">{sub}</div>
    </div>
  );
}

function statusBreakdown(byStatus: Record<string, number>): string {
  const order: ReadonlyArray<keyof typeof byStatus> = [
    "approved",
    "merged",
    "blocked",
    "failed",
    "no_review",
    "cancelled",
    "drafting",
    "reviewing",
  ];
  const parts: string[] = [];
  for (const k of order) {
    const n = byStatus[k];
    if (n) parts.push(`${n} ${k}`);
  }
  return parts.length === 0 ? "no runs yet" : parts.join(" · ");
}

function formatUsd(amount: number): string {
  if (amount === 0) return "$0";
  // Sub-cent spend rounds to "$0.00" or "$0.0000" — both look like
  // "we spent literally nothing". "<$0.01" reads as "real but tiny",
  // matching the run-page card convention.
  if (amount < 0.01) return "<$0.01";
  if (amount < 1) return `$${amount.toFixed(3)}`;
  return `$${amount.toFixed(2)}`;
}

function formatTokens(n: number): string {
  if (n < 1000) return `${n}`;
  if (n < 1_000_000) return `${(n / 1000).toFixed(1)}k`;
  return `${(n / 1_000_000).toFixed(1)}M`;
}

function formatDuration(ms: number): string {
  if (ms === 0) return "—";
  if (ms < 1000) return `${ms}ms`;
  const sec = ms / 1000;
  if (sec < 60) return `${sec.toFixed(1)}s`;
  const min = sec / 60;
  if (min < 60) return `${min.toFixed(1)}m`;
  return `${(min / 60).toFixed(1)}h`;
}
