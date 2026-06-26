"use client";

import { Plus, X } from "lucide-react";
import type { Persona } from "@/lib/api/personas";
import type { ReviewerLineage, TemplatePhase } from "@/lib/cockpit-types";
import { DEFAULT_MODELS, LINEAGES } from "./constants";
import type { ConnectedVoiceMap } from "./hooks";
import { ModelSelect, PersonaSelect } from "./selects";

interface ReviewerRow {
  lineage: ReviewerLineage;
  model: string;
  /** Optional persona id — round-trips via reviewer.candidatePersonas. */
  persona?: string;
}

/**
 * One row per reviewer = (lineage, model). Same lineage can repeat with
 * different models; each row emits its own entry in YAML candidates[].
 * Internally the existing candidates+candidateModels shape is preserved
 * so nothing else has to change; the editor flattens for display and
 * re-derives both fields on each edit.
 */
export function reviewerToRows(
  reviewer: TemplatePhase["reviewer"],
): ReviewerRow[] {
  const rows: ReviewerRow[] = [];
  for (const lineage of reviewer.candidates) {
    const models = reviewer.candidateModels?.[lineage];
    const personasByModel = reviewer.candidatePersonas?.[lineage];
    if (!models || models.length === 0) {
      // The "" model slot's persona key is also "" — keeps the lookup
      // consistent with the populated case.
      rows.push({
        lineage,
        model: "",
        ...(personasByModel?.[""] ? { persona: personasByModel[""] } : {}),
      });
    } else {
      for (const m of models) {
        rows.push({
          lineage,
          model: m,
          ...(personasByModel?.[m] ? { persona: personasByModel[m] } : {}),
        });
      }
    }
  }
  return rows;
}

function rowsToReviewer(
  rows: ReviewerRow[],
  base: TemplatePhase["reviewer"],
): TemplatePhase["reviewer"] {
  const candidates: ReviewerLineage[] = [];
  const candidateModels: Partial<Record<ReviewerLineage, string[]>> = {};
  const candidatePersonas: Partial<
    Record<ReviewerLineage, Record<string, string>>
  > = {};
  for (const r of rows) {
    if (!candidates.includes(r.lineage)) candidates.push(r.lineage);
    if (r.model) (candidateModels[r.lineage] ??= []).push(r.model);
    if (r.persona) {
      (candidatePersonas[r.lineage] ??= {})[r.model] = r.persona;
    }
  }
  return {
    ...base,
    candidates,
    candidateModels,
    ...(Object.keys(candidatePersonas).length > 0
      ? { candidatePersonas }
      : { candidatePersonas: undefined }),
  };
}

interface ReviewerSlotsEditorProps {
  phase: TemplatePhase;
  connectedVoices: ConnectedVoiceMap;
  personas: Persona[];
  onUpdate: (patch: Partial<TemplatePhase>) => void;
}

export function ReviewerSlotsEditor({
  phase,
  connectedVoices,
  personas,
  onUpdate,
}: ReviewerSlotsEditorProps) {
  const rows = reviewerToRows(phase.reviewer);
  // Falls through to all lineages on a fresh install with zero voices
  // so the dialog stays usable.
  const availableLineages = LINEAGES.filter((l) =>
    connectedVoices.connectedLineages.size === 0
      ? true
      : connectedVoices.connectedLineages.has(l.id),
  );

  function commit(nextRows: ReviewerRow[]) {
    onUpdate({ reviewer: rowsToReviewer(nextRows, phase.reviewer) });
  }

  function setRow(i: number, patch: Partial<ReviewerRow>) {
    const next = rows.map((r, idx) => (idx === i ? { ...r, ...patch } : r));
    // Lineage change resets the model — old model belonged to a
    // different lineage and won't make sense.
    if (patch.lineage && patch.lineage !== rows[i].lineage) {
      const enabledForNew = connectedVoices.byLineage[patch.lineage] ?? [];
      next[i].model = enabledForNew[0] ?? DEFAULT_MODELS[patch.lineage] ?? "";
    }
    commit(next);
  }

  function removeRow(i: number) {
    commit(rows.filter((_, idx) => idx !== i));
  }

  function addRow() {
    const firstLineage =
      availableLineages[0]?.id ?? ("claude" as ReviewerLineage);
    const enabledForLineage = connectedVoices.byLineage[firstLineage] ?? [];
    const usedModelsForLineage = new Set(
      rows.filter((r) => r.lineage === firstLineage).map((r) => r.model),
    );
    const fresh =
      enabledForLineage.find((m) => !usedModelsForLineage.has(m)) ??
      DEFAULT_MODELS[firstLineage] ??
      "";
    commit([...rows, { lineage: firstLineage, model: fresh }]);
  }

  return (
    <div className="space-y-1.5">
      {rows.length === 0 && (
        <p className="text-[11px] text-muted-foreground">
          No reviewers yet. Click + add reviewer below.
        </p>
      )}
      {rows.map((row, i) => {
        const lineageMeta = LINEAGES.find((x) => x.id === row.lineage);
        return (
          <div key={i} className="flex items-center gap-1.5">
            <select
              value={row.lineage}
              onChange={(e) =>
                setRow(i, { lineage: e.target.value as ReviewerLineage })
              }
              className="h-7 w-32 shrink-0 rounded-md border border-border bg-background px-2 text-[11px] focus:border-primary/60 focus:outline-none"
              aria-label="Reviewer lineage"
            >
              {availableLineages.map((l) => (
                <option key={l.id} value={l.id}>
                  {l.label}
                </option>
              ))}
              {/* Preserve a row that references a lineage NOT in the
                  available list (template authored elsewhere). */}
              {!availableLineages.find((l) => l.id === row.lineage) && (
                <option value={row.lineage}>
                  {lineageMeta?.label ?? row.lineage} (not connected)
                </option>
              )}
            </select>
            <div className="flex-1">
              <ModelSelect
                lineage={row.lineage}
                value={row.model}
                options={connectedVoices.byLineage[row.lineage] ?? []}
                defaultModel={DEFAULT_MODELS[row.lineage]}
                onChange={(next) => setRow(i, { model: next })}
              />
            </div>
            <PersonaSelect
              value={row.persona}
              personas={personas}
              onChange={(next) => setRow(i, { persona: next })}
              inline
              ariaLabel="Reviewer persona"
            />
            <button
              type="button"
              onClick={() => removeRow(i)}
              aria-label="Remove reviewer"
              className="grid h-7 w-7 shrink-0 place-items-center rounded-md border border-border bg-card/40 text-muted-foreground hover:border-destructive/40 hover:text-destructive"
            >
              <X className="h-3 w-3" />
            </button>
          </div>
        );
      })}
      <button
        type="button"
        onClick={addRow}
        className="flex w-full items-center justify-center gap-1.5 rounded-md border border-dashed border-border bg-card/30 px-3 py-2 text-[11px] font-medium text-muted-foreground transition hover:border-primary/40 hover:bg-card/50 hover:text-foreground"
      >
        <Plus className="h-3 w-3" />
        Add another reviewer
      </button>
    </div>
  );
}
