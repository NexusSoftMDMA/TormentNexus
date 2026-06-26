"use client";

import { cn } from "@/lib/utils";
import type { TemplatePhase } from "@/lib/cockpit-types";

export function IconBtn({
  children,
  onClick,
  disabled,
  title,
  className,
}: {
  children: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
  title?: string;
  className?: string;
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
      title={title}
      className={cn(
        "grid h-6 w-6 place-items-center rounded text-muted-foreground transition hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-30",
        className,
      )}
    >
      {children}
    </button>
  );
}

export function SubField({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="mb-1.5 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  );
}

export function ToggleChip({
  on,
  tone,
  onClick,
  children,
}: {
  on: boolean;
  tone: "emerald" | "rose";
  onClick: () => void;
  children: React.ReactNode;
}) {
  const tones = {
    emerald: on
      ? "border-emerald-500/40 bg-emerald-500/15 text-emerald-200"
      : "border-border bg-card/40 text-muted-foreground hover:border-emerald-500/30",
    rose: on
      ? "border-rose-500/40 bg-rose-500/15 text-rose-200"
      : "border-border bg-card/40 text-muted-foreground hover:border-rose-500/30",
  };
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "rounded-md border px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider transition",
        tones[tone],
      )}
    >
      {children}
    </button>
  );
}

export function makeDefaultPhase(idx: number): TemplatePhase {
  return {
    id: `phase-${idx + 1}`,
    name: `Phase ${idx + 1}`,
    description: "",
    kind: "review",
    gate: "auto",
    doer: { lineage: "claude", models: ["claude-opus-4-7"] },
    reviewer: {
      require: 1,
      crossLineage: true,
      candidates: ["codex"],
    },
    inputs: { include: [], exclude: [] },
    iterate: { max: 3, onMax: "ask-user" },
    blindSpots: [],
    execution: "sequential",
    builtin: false,
  };
}
