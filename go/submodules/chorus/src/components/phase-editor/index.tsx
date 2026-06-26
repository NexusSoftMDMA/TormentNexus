"use client";

import { Plus } from "lucide-react";
import type { TemplatePhase } from "@/lib/cockpit-types";
import { useConnectedVoices, usePersonas } from "./hooks";
import { PhaseCard } from "./phase-card";
import { makeDefaultPhase } from "./primitives";

export type { ConnectedVoiceMap } from "./hooks";
export { useConnectedVoices } from "./hooks";

interface PhaseEditorProps {
  phases: TemplatePhase[];
  onChange: (next: TemplatePhase[]) => void;
}

export function PhaseEditor({ phases, onChange }: PhaseEditorProps) {
  const connectedVoices = useConnectedVoices();
  const { personas } = usePersonas();

  function update(idx: number, patch: Partial<TemplatePhase>) {
    onChange(phases.map((p, i) => (i === idx ? { ...p, ...patch } : p)));
  }

  function move(idx: number, dir: -1 | 1) {
    const nextIdx = idx + dir;
    if (nextIdx < 0 || nextIdx >= phases.length) return;
    const next = [...phases];
    [next[idx], next[nextIdx]] = [next[nextIdx], next[idx]];
    onChange(next);
  }

  function remove(idx: number) {
    onChange(phases.filter((_, i) => i !== idx));
  }

  function add() {
    onChange([...phases, makeDefaultPhase(phases.length)]);
  }

  const phaseIds = phases.map((p) => p.id);

  return (
    <div>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          Phases ({phases.length})
        </span>
        <span className="text-[11px] text-muted-foreground/70">
          Each phase has a doer + reviewer. Adversarial by default.
        </span>
      </div>

      <div className="space-y-2">
        {phases.map((p, i) => (
          <PhaseCard
            key={p.id}
            phase={p}
            index={i}
            total={phases.length}
            allPhaseIds={phaseIds}
            connectedVoices={connectedVoices}
            personas={personas}
            onUpdate={(patch) => update(i, patch)}
            onMoveUp={() => move(i, -1)}
            onMoveDown={() => move(i, 1)}
            onDelete={() => remove(i)}
          />
        ))}
      </div>

      <button
        type="button"
        onClick={add}
        className="mt-3 flex w-full items-center justify-center gap-1.5 rounded-md border border-dashed border-border bg-card/30 px-3 py-2.5 text-xs font-medium text-muted-foreground transition hover:border-primary/40 hover:bg-card/50 hover:text-foreground"
      >
        <Plus className="h-3.5 w-3.5" />
        Add phase
      </button>
    </div>
  );
}
