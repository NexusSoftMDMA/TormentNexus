"use client";

import { AlertTriangle, CheckCircle2 } from "lucide-react";
import { PhaseStepper, type PhaseState } from "@/components/phase-stepper";
import type { TemplatePhase as MockTemplatePhase } from "@/lib/cockpit-types";
import type { Template } from "@/lib/types";
import type { RoundSnapshot } from "../run-viewer/types.js";

interface PhaseProgressProps {
  template: Template | null;
  status: string;
  totalPhases: number;
  completedPhaseCount: number;
  rounds: RoundSnapshot[];
  enrichedRounds: RoundSnapshot[];
  prUrl: string | undefined;
  shipError: string | undefined;
}

export function PhaseProgress({
  template,
  status,
  totalPhases,
  completedPhaseCount,
  rounds,
  enrichedRounds,
  prUrl,
  shipError,
}: PhaseProgressProps) {
  return (
    <div className="border-b border-border bg-card/20 px-4 py-3 sm:px-8">
      <div className="mx-auto flex w-full flex-col gap-3">
        {prUrl && (
          <div className="rounded-md border border-emerald-500/40 bg-emerald-500/5 px-4 py-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div className="flex items-center gap-2 text-sm font-medium text-emerald-300">
                <CheckCircle2 className="h-4 w-4" />
                Pull request opened
              </div>
              <a
                href={prUrl}
                target="_blank"
                rel="noreferrer"
                className="rounded-md border border-emerald-500/40 bg-emerald-500/10 px-3 py-1.5 text-xs font-medium text-emerald-300 transition hover:bg-emerald-500/20"
              >
                View PR →
              </a>
            </div>
            <p className="mt-1 break-all font-mono text-[11px] text-emerald-200/70">
              {prUrl}
            </p>
          </div>
        )}
        {shipError && !prUrl && (
          <div className="rounded-md border border-amber-500/40 bg-amber-500/5 px-4 py-3">
            <div className="flex items-center gap-2 text-sm font-medium text-amber-300">
              <AlertTriangle className="h-4 w-4" />
              Ship phase blocked
            </div>
            <p className="mt-1 break-words font-mono text-[11px] text-amber-200/80">
              {shipError}
            </p>
            <p className="mt-2 text-[11px] leading-relaxed text-muted-foreground">
              The reviewers approved the doer&apos;s output, but chorus
              couldn&apos;t open a PR. Resolve the issue above and re-run.
            </p>
          </div>
        )}

        {template?.phases && template.phases.length > 0 && (
          <div className="flex justify-center">
            <PhaseStepper
              phases={template.phases as unknown as MockTemplatePhase[]}
              activeIndex={Math.min(completedPhaseCount, totalPhases - 1)}
              states={template.phases.map((_, i): PhaseState => {
                if (status === "approved") return "done";
                if (status === "no_review")
                  return i < completedPhaseCount ? "done" : "blocked";
                if (status === "failed" || status === "cancelled")
                  return i < completedPhaseCount ? "done" : "skipped";
                if (i < completedPhaseCount) return "done";
                if (i === completedPhaseCount) return "active";
                return "pending";
              })}
            />
          </div>
        )}

        <ProgressStrip
          template={template}
          status={status}
          totalPhases={totalPhases}
          completedPhaseCount={completedPhaseCount}
          rounds={rounds}
          enrichedRounds={enrichedRounds}
        />
      </div>
    </div>
  );
}

/**
 * Counts phases for multi-phase templates, rounds for single-phase +
 * multi-round (e.g. bug-diagnose), reviewers for single-phase +
 * single-round + multi-reviewer (tri-review). Hidden when the run is a
 * single shot.
 */
function ProgressStrip({
  template,
  status,
  totalPhases,
  completedPhaseCount,
  rounds,
  enrichedRounds,
}: {
  template: Template | null;
  status: string;
  totalPhases: number;
  completedPhaseCount: number;
  rounds: RoundSnapshot[];
  enrichedRounds: RoundSnapshot[];
}) {
  const maxRounds = template?.maxRounds ?? 1;
  const isTerminal =
    status === "approved" ||
    status === "merged" ||
    status === "no_review" ||
    status === "blocked" ||
    status === "failed" ||
    status === "cancelled";

  const showByPhases = totalPhases > 1;
  const showByRounds = !showByPhases && maxRounds > 1;
  const currentRound = enrichedRounds[enrichedRounds.length - 1];
  // Count reviewers only — the doer (or synthetic `doer-artifact` slot
  // in review-only chats) is its own participant on disk but not part
  // of the "N reviewers complete" mental model. Without this filter a
  // 4-reviewer review-only chat displays "x/5".
  const reviewerParticipants =
    currentRound?.participants.filter((p) => p.role === "reviewer") ?? [];
  const participantTotal = reviewerParticipants.length;
  const participantDone = reviewerParticipants.filter((p) => p.hasAnswer).length;
  const showByParticipants =
    !showByPhases && !showByRounds && participantTotal > 1;
  if (!showByPhases && !showByRounds && !showByParticipants) return null;

  const total = showByPhases
    ? totalPhases
    : showByRounds
      ? maxRounds
      : participantTotal;
  const completed = showByPhases
    ? Math.min(completedPhaseCount, totalPhases)
    : showByRounds
      ? Math.min(Math.max(rounds.length, 1), maxRounds)
      : participantDone;
  const display = showByPhases && isTerminal ? total : completed;
  const label = showByPhases
    ? `${display} / ${total} phases`
    : showByRounds
      ? `Round ${display} / ${total}`
      : `${display} / ${total} complete`;

  return (
    <div className="mx-auto flex w-full max-w-xs items-center gap-2">
      <div className="flex h-1 flex-1 overflow-hidden rounded-full bg-muted">
        <div
          className={`transition-[width] duration-700 ease-out ${
            status === "approved" ? "bg-emerald-400" : "bg-primary"
          }`}
          style={{ width: `${(display / total) * 100}%` }}
        />
      </div>
      <span className="shrink-0 font-mono text-[10px] tabular-nums text-muted-foreground">
        {label}
      </span>
    </div>
  );
}
