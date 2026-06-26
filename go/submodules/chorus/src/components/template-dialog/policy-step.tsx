"use client";

import { cn } from "@/lib/utils";
import { ACTIONS, THRESHOLDS } from "./constants";
import type { FormState } from "./types";

export function PolicyStep({
  form,
  setField,
}: {
  form: FormState;
  setField: <K extends keyof FormState>(k: K, v: FormState[K]) => void;
}) {
  return (
    <div className="rounded-lg border border-border bg-card/40 p-4">
      <h3 className="mb-3 text-[13px] font-semibold tracking-tight">
        Across all phases
      </h3>

      <div className="mb-4">
        <div className="mb-1.5 text-[11px] font-medium text-muted-foreground">
          Agreement threshold (per review-style phase)
        </div>
        <div className="grid grid-cols-3 gap-2">
          {THRESHOLDS.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => setField("threshold", t.id)}
              className={cn(
                "rounded-md border px-3 py-2 text-left transition",
                form.threshold === t.id
                  ? "border-primary/60 bg-primary/10 ring-1 ring-primary/40"
                  : "border-border bg-card hover:border-muted-foreground/30 hover:bg-accent/40",
              )}
            >
              <div className="text-xs font-medium">{t.label}</div>
            </button>
          ))}
        </div>
        <p className="mt-1.5 text-[11px] leading-snug text-muted-foreground/80">
          {THRESHOLDS.find((t) => t.id === form.threshold)?.hint}
        </p>
      </div>

      <div className="mb-4">
        <div className="mb-1.5 text-[11px] font-medium text-muted-foreground">
          When threshold is met
        </div>
        <div className="grid grid-cols-2 gap-2">
          {ACTIONS.map((a) => (
            <button
              key={a.id}
              type="button"
              onClick={() => setField("onThresholdMet", a.id)}
              className={cn(
                "rounded-md border px-3 py-2 text-left transition",
                form.onThresholdMet === a.id
                  ? "border-primary/60 bg-primary/10 ring-1 ring-primary/40"
                  : "border-border bg-card hover:border-muted-foreground/30 hover:bg-accent/40",
              )}
            >
              <div className="text-xs font-medium">{a.label}</div>
            </button>
          ))}
        </div>
      </div>

      <div className="mb-4">
        <div className="mb-1.5 flex items-center justify-between">
          <span className="text-[11px] font-medium text-muted-foreground">
            Max revise rounds (per phase)
          </span>
          <span className="font-mono text-xs text-foreground">
            {form.maxRounds}
          </span>
        </div>
        <input
          type="range"
          min={1}
          max={5}
          step={1}
          value={form.maxRounds}
          onChange={(e) => setField("maxRounds", parseInt(e.target.value, 10))}
          className="h-1 w-full cursor-pointer appearance-none rounded-full bg-muted accent-primary"
        />
      </div>

      <button
        type="button"
        onClick={() => setField("yoloDefault", !form.yoloDefault)}
        className={cn(
          "flex w-full items-center justify-between rounded-md border px-3 py-2 text-left transition",
          form.yoloDefault
            ? "border-rose-500/40 bg-rose-500/5"
            : "border-border bg-card hover:border-foreground/30",
        )}
      >
        <div>
          <div className="text-xs font-medium text-foreground">
            🚀 Yolo by default
          </div>
          <div className="text-[10px] text-muted-foreground">
            Auto-approve every gate. Only flip on for trusted templates or
            trivial fixes. Cost cap still enforced.
          </div>
        </div>
        <span
          className={cn(
            "flex h-5 w-9 shrink-0 items-center rounded-full border p-0.5 transition",
            form.yoloDefault
              ? "border-rose-500/40 bg-rose-500/20"
              : "border-border bg-card",
          )}
        >
          <span
            className={cn(
              "h-3.5 w-3.5 rounded-full transition-transform",
              form.yoloDefault
                ? "translate-x-4 bg-rose-400"
                : "bg-muted-foreground/50",
            )}
          />
        </span>
      </button>
    </div>
  );
}
