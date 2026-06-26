"use client";

import {
  Check,
  Loader2,
  GitPullRequest,
  Eye,
  Code2,
  FlaskConical,
  Search,
  ClipboardList,
  FileCode2,
  TestTube2,
  Shuffle,
} from "lucide-react";
import type { PhaseKind, TemplatePhase } from "@/lib/cockpit-types";

export type PhaseState = "pending" | "active" | "done" | "blocked" | "skipped";

interface PhaseStepperProps {
  phases: TemplatePhase[];
  /** Index of the active phase (0-based). */
  activeIndex: number;
  /** Per-phase state. Length should match phases. */
  states: PhaseState[];
  onSelect?: (index: number) => void;
}

const KIND_ICON: Record<PhaseKind, React.ComponentType<{ className?: string }>> =
  {
    review: Eye,
    review_only: Eye,
    plan: ClipboardList,
    spec: FileCode2,
    tests: TestTube2,
    implement: Code2,
    verify: FlaskConical,
    pr: GitPullRequest,
    divergence: Shuffle,
    recon: Search,
  };

const STATE_DOT: Record<PhaseState, string> = {
  pending: "bg-muted-foreground/20 text-muted-foreground",
  active: "bg-primary/20 text-primary ring-2 ring-primary/40 ring-offset-2 ring-offset-background",
  done: "bg-emerald-500/20 text-emerald-300",
  blocked: "bg-amber-500/20 text-amber-300",
  skipped: "bg-muted-foreground/10 text-muted-foreground/50 line-through",
};

const STATE_LABEL: Record<PhaseState, string> = {
  pending: "queued",
  active: "active",
  done: "done",
  blocked: "blocked",
  skipped: "skipped",
};

export function PhaseStepper({
  phases,
  activeIndex,
  states,
  onSelect,
}: PhaseStepperProps) {
  return (
    <div className="flex items-center gap-0 overflow-x-auto pb-1">
      {phases.map((phase, i) => {
        const state = states[i] ?? "pending";
        const Icon = KIND_ICON[phase.kind];
        const isActive = i === activeIndex;
        const clickable = onSelect && (state === "done" || state === "active");

        return (
          <div key={phase.id} className="flex items-center">
            <button
              type="button"
              disabled={!clickable}
              onClick={() => clickable && onSelect?.(i)}
              className={`group flex items-center gap-2 rounded-md border px-3 py-1.5 text-left transition ${
                isActive
                  ? "border-primary/40 bg-primary/5"
                  : state === "done"
                    ? "border-emerald-500/30 bg-emerald-500/5"
                    : state === "blocked"
                      ? "border-amber-500/40 bg-amber-500/5"
                      : "border-border bg-card/30"
              } ${clickable ? "hover:border-foreground/30" : "cursor-default"}`}
            >
              <span
                className={`grid h-6 w-6 shrink-0 place-items-center rounded-md ${STATE_DOT[state]}`}
              >
                {state === "done" ? (
                  <Check className="h-3.5 w-3.5" />
                ) : state === "active" ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <Icon className="h-3.5 w-3.5" />
                )}
              </span>
              <span className="flex flex-col leading-tight">
                <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                  {String(i + 1).padStart(2, "0")} · {STATE_LABEL[state]}
                </span>
                <span
                  className={`text-xs font-medium ${
                    state === "skipped"
                      ? "text-muted-foreground/50"
                      : "text-foreground"
                  }`}
                >
                  {phase.name}
                </span>
              </span>
            </button>
            {i < phases.length - 1 && (
              <span
                className={`mx-1.5 h-px w-6 ${
                  states[i] === "done"
                    ? "bg-emerald-500/40"
                    : "bg-border"
                }`}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}
