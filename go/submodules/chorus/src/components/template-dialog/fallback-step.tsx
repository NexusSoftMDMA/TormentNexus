"use client";

import { Plus, Shuffle, X } from "lucide-react";
import { useConnectedVoices, type ConnectedVoiceMap } from "@/components/phase-editor";
import type { ReviewerLineage } from "@/lib/cockpit-types";
import { FALLBACK_LINEAGES } from "./constants";
import type { FallbackVoice, FormState } from "./types";

export function FallbackStep({
  form,
  setField,
}: {
  form: FormState;
  setField: <K extends keyof FormState>(k: K, v: FormState[K]) => void;
}) {
  // Same hook PhaseEditor uses — fetch enabled voices once. Lets the
  // model dropdown show ONLY models the user has connected.
  const connectedVoices = useConnectedVoices();
  return (
    <>
      <div className="rounded-md border border-border bg-card/40 px-3 py-2 text-[11px] leading-relaxed text-muted-foreground">
        <div className="mb-1 flex items-center gap-1.5 text-foreground">
          <Shuffle className="h-3 w-3" />
          <span className="text-[12px] font-medium">
            Template-level fallback chains
          </span>
        </div>
        Tried in order whenever a slot exhausts its own per-slot model chain.
        Cross-lineage swap supported — a codex slot can fall through to a
        claude or kimi entry. Strict
        <span className="font-mono"> (lineage, model)</span> dedup keeps
        diversity: a fallback whose model already runs as another reviewer
        in this phase is skipped, so you never review the same code twice
        with the same voice. Order matters — list your most-preferred
        fallback first.
      </div>

      <FallbackList
        title="Doer fallback"
        hint="Engaged when the doer's primary + per-slot models all fail."
        rows={form.fallbackDoer}
        onChange={(rows) => setField("fallbackDoer", rows)}
        connectedVoices={connectedVoices}
      />

      <FallbackList
        title="Reviewer fallback"
        hint="Engaged when any reviewer's primary + per-slot models all fail."
        rows={form.fallbackReviewer}
        onChange={(rows) => setField("fallbackReviewer", rows)}
        connectedVoices={connectedVoices}
      />
    </>
  );
}

function FallbackList({
  title,
  hint,
  rows,
  onChange,
  connectedVoices,
}: {
  title: string;
  hint: string;
  rows: FallbackVoice[];
  onChange: (rows: FallbackVoice[]) => void;
  connectedVoices: ConnectedVoiceMap;
}) {
  const addRow = () => {
    // Default lineage = first one with at least one connected model so
    // the model dropdown isn't empty on first add.
    const firstConnected = FALLBACK_LINEAGES.find(
      (l) => (connectedVoices.byLineage[l] ?? []).length > 0,
    );
    const lineage: ReviewerLineage = firstConnected ?? "claude";
    const firstModel = connectedVoices.byLineage[lineage]?.[0] ?? "";
    onChange([...rows, { lineage, model: firstModel }]);
  };
  const removeRow = (idx: number) => {
    onChange(rows.filter((_, i) => i !== idx));
  };
  const updateRow = (idx: number, patch: Partial<FallbackVoice>) => {
    onChange(rows.map((r, i) => (i === idx ? { ...r, ...patch } : r)));
  };
  // Auto-pick the first model when switching lineage if the current
  // model isn't valid for the new lineage. Without this the row could
  // keep e.g. "claude-opus-4-7" after switching to opencode, which would
  // fail at runtime with "model not enabled".
  const updateLineage = (idx: number, nextLineage: ReviewerLineage) => {
    const models = connectedVoices.byLineage[nextLineage] ?? [];
    const currentModel = rows[idx].model;
    const keepModel = models.includes(currentModel);
    updateRow(idx, {
      lineage: nextLineage,
      model: keepModel ? currentModel : models[0] ?? "",
    });
  };
  return (
    <div className="rounded-lg border border-border bg-card/40 p-4">
      <div className="mb-1 flex items-center justify-between">
        <h4 className="text-[13px] font-semibold tracking-tight">{title}</h4>
        <button
          type="button"
          onClick={addRow}
          className="inline-flex h-7 items-center gap-1 rounded-md border border-border bg-card px-2.5 text-[11px] font-medium text-muted-foreground transition hover:border-primary/40 hover:text-primary"
        >
          <Plus className="h-3 w-3" />
          Add fallback
        </button>
      </div>
      <p className="mb-3 text-[11px] leading-snug text-muted-foreground/80">
        {hint}
      </p>
      {rows.length === 0 ? (
        <p className="rounded-md border border-dashed border-border/60 bg-background/30 px-3 py-4 text-center text-[11px] text-muted-foreground">
          None set — slot will fail its phase if the per-slot chain exhausts.
        </p>
      ) : (
        <ul className="space-y-2">
          {rows.map((row, idx) => {
            const lineageModels = connectedVoices.byLineage[row.lineage] ?? [];
            // Surface a stored model that isn't in the connected list as
            // "(stored)" so editing an existing template doesn't silently
            // lose the value when a model is no longer enabled.
            const modelHasOption =
              row.model.length === 0 || lineageModels.includes(row.model);
            return (
              <li key={idx} className="flex items-center gap-2">
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-border bg-card font-mono text-[10px] text-muted-foreground">
                  {idx + 1}
                </span>
                <select
                  value={row.lineage}
                  onChange={(e) =>
                    updateLineage(idx, e.target.value as ReviewerLineage)
                  }
                  className="h-8 rounded-md border border-border bg-background px-2 text-xs text-foreground focus:border-primary/60 focus:outline-none focus:ring-1 focus:ring-primary/20"
                >
                  {FALLBACK_LINEAGES.map((l) => (
                    <option key={l} value={l}>
                      {l}
                    </option>
                  ))}
                </select>
                <select
                  value={row.model}
                  onChange={(e) => updateRow(idx, { model: e.target.value })}
                  className="h-8 flex-1 rounded-md border border-border bg-background px-2 font-mono text-[11px] text-foreground focus:border-primary/60 focus:outline-none focus:ring-1 focus:ring-primary/20"
                >
                  {lineageModels.length === 0 && (
                    <option value="" disabled>
                      {connectedVoices.loaded
                        ? `No ${row.lineage} models enabled — open Connect → ${row.lineage}`
                        : "loading…"}
                    </option>
                  )}
                  {lineageModels.map((m) => (
                    <option key={m} value={m}>
                      {m}
                    </option>
                  ))}
                  {!modelHasOption && row.model.length > 0 && (
                    <option value={row.model}>{row.model} (stored)</option>
                  )}
                </select>
                <button
                  type="button"
                  onClick={() => removeRow(idx)}
                  aria-label={`Remove fallback ${idx + 1}`}
                  className="grid h-7 w-7 place-items-center rounded-md border border-border bg-card text-muted-foreground transition hover:border-destructive/40 hover:text-destructive"
                >
                  <X className="h-3 w-3" />
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
