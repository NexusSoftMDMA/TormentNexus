"use client";

import { useState } from "react";
import {
  ArrowDown,
  ArrowUp,
  ChevronDown,
  ChevronRight,
  Trash2,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { Persona } from "@/lib/api/personas";
import type { PhaseKind, TemplatePhase } from "@/lib/cockpit-types";
import { KIND_ICON, KINDS, LINEAGES } from "./constants";
import type { ConnectedVoiceMap } from "./hooks";
import {
  ApprovalsSection,
  ArtifactSection,
  DoerSection,
  ExecutionSection,
  GateSection,
  InputsSection,
  IterateSection,
  ReviewersSection,
} from "./phase-card-sections";
import { IconBtn, SubField } from "./primitives";

interface PhaseCardProps {
  phase: TemplatePhase;
  index: number;
  total: number;
  allPhaseIds: string[];
  connectedVoices: ConnectedVoiceMap;
  personas: Persona[];
  onUpdate: (patch: Partial<TemplatePhase>) => void;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onDelete: () => void;
}

export function PhaseCard({
  phase,
  index,
  total,
  allPhaseIds,
  connectedVoices,
  personas,
  onUpdate,
  onMoveUp,
  onMoveDown,
  onDelete,
}: PhaseCardProps) {
  const [expanded, setExpanded] = useState(index === 0);
  const KindIcon = KIND_ICON[phase.kind];
  const priorPhases = allPhaseIds.slice(0, index);

  return (
    <div className="rounded-lg border border-border bg-card/40">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex w-full items-center gap-3 px-3 py-2.5 text-left"
      >
        <span className="font-mono text-[10px] text-muted-foreground/60">
          {String(index + 1).padStart(2, "0")}
        </span>
        <span className="grid h-7 w-7 shrink-0 place-items-center rounded-md bg-primary/15 text-primary">
          <KindIcon className="h-3.5 w-3.5" />
        </span>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-foreground truncate">
              {phase.name || "Untitled phase"}
            </span>
            <span className="font-mono text-[10px] uppercase text-muted-foreground/70">
              {phase.kind}
            </span>
          </div>
          <PhaseSummary phase={phase} />
        </div>
        <div className="flex shrink-0 items-center gap-0.5">
          <IconBtn disabled={index === 0} onClick={onMoveUp} title="Move up">
            <ArrowUp className="h-3 w-3" />
          </IconBtn>
          <IconBtn
            disabled={index === total - 1}
            onClick={onMoveDown}
            title="Move down"
          >
            <ArrowDown className="h-3 w-3" />
          </IconBtn>
          <IconBtn
            onClick={onDelete}
            title="Delete phase"
            className="hover:text-rose-400"
          >
            <Trash2 className="h-3 w-3" />
          </IconBtn>
          <span className="ml-1 text-muted-foreground/60">
            {expanded ? (
              <ChevronDown className="h-3.5 w-3.5" />
            ) : (
              <ChevronRight className="h-3.5 w-3.5" />
            )}
          </span>
        </div>
      </button>

      {expanded && (
        <div className="space-y-4 border-t border-border bg-background/40 px-3 py-3">
          <div className="grid grid-cols-[1fr_140px] gap-2">
            <SubField label="Name">
              <input
                type="text"
                value={phase.name}
                onChange={(e) => onUpdate({ name: e.target.value })}
                className="h-8 w-full rounded-md border border-border bg-background px-2.5 text-xs text-foreground focus:border-primary/60 focus:outline-none"
              />
            </SubField>
            <SubField label="Kind">
              <select
                value={phase.kind}
                onChange={(e) => onUpdate({ kind: e.target.value as PhaseKind })}
                className="h-8 w-full rounded-md border border-border bg-background px-2 text-xs text-foreground focus:border-primary/60 focus:outline-none"
              >
                {KINDS.map((k) => (
                  <option key={k.id} value={k.id}>
                    {k.label}
                  </option>
                ))}
              </select>
            </SubField>
          </div>

          <SubField label="Description">
            <input
              type="text"
              value={phase.description}
              onChange={(e) => onUpdate({ description: e.target.value })}
              placeholder="What this phase produces and why it exists."
              className="h-8 w-full rounded-md border border-border bg-background px-2.5 text-xs text-foreground placeholder:text-muted-foreground/60 focus:border-primary/60 focus:outline-none"
            />
          </SubField>

          {phase.kind !== "review_only" && (
            <DoerSection
              phase={phase}
              onUpdate={onUpdate}
              connectedVoices={connectedVoices}
              personas={personas}
            />
          )}

          {phase.kind === "review_only" && (
            <ArtifactSection phase={phase} onUpdate={onUpdate} />
          )}

          <ReviewersSection
            phase={phase}
            onUpdate={onUpdate}
            connectedVoices={connectedVoices}
            personas={personas}
          />

          <ApprovalsSection phase={phase} onUpdate={onUpdate} />

          <ExecutionSection phase={phase} onUpdate={onUpdate} />

          {priorPhases.length > 0 && (
            <InputsSection
              phase={phase}
              onUpdate={onUpdate}
              priorPhases={priorPhases}
            />
          )}

          {/* Iterate + Gate are runner-loop concerns; review_only is
              single-pass per the schema (no doer to revise). */}
          {phase.kind !== "review_only" && (
            <>
              <IterateSection
                phase={phase}
                onUpdate={onUpdate}
                priorPhases={priorPhases}
              />
              <GateSection phase={phase} onUpdate={onUpdate} />
            </>
          )}
        </div>
      )}
    </div>
  );
}

/** Header subtitle: doer + reviewer dots, execution mode, blind-to chips. */
function PhaseSummary({ phase }: { phase: TemplatePhase }) {
  return (
    <div className="mt-0.5 flex flex-wrap items-center gap-1.5 text-[10px] text-muted-foreground">
      {phase.kind !== "review_only" && (
        <span className="flex items-center gap-1">
          <span
            className={cn(
              "h-1.5 w-1.5 rounded-full",
              LINEAGES.find((l) => l.id === phase.doer.lineage)?.dot,
            )}
          />
          doer: {phase.doer.lineage}
        </span>
      )}
      {phase.reviewer.candidates.length > 0 && (
        <>
          {phase.kind !== "review_only" && (
            <span className="text-muted-foreground/40">·</span>
          )}
          <span className="flex items-center gap-1">
            reviewers:
            {phase.reviewer.candidates.map((l) => (
              <span
                key={l}
                className={cn(
                  "h-1.5 w-1.5 rounded-full",
                  LINEAGES.find((x) => x.id === l)?.dot,
                )}
                title={l}
              />
            ))}
          </span>
        </>
      )}
      {phase.reviewer.crossLineage && (
        <>
          <span className="text-muted-foreground/40">·</span>
          <span className="text-amber-300/90 font-mono text-[9px] uppercase">
            cross-lineage
          </span>
        </>
      )}
      <span className="text-muted-foreground/40">·</span>
      <span
        className={cn(
          "font-mono text-[9px] uppercase",
          phase.execution === "sequential"
            ? "text-amber-300/90"
            : "text-emerald-300/80",
        )}
        title={
          phase.execution === "sequential"
            ? "Sequential hostile — reviewers chain"
            : "Parallel independent — reviewers vote"
        }
      >
        {phase.execution === "sequential" ? "sequential" : "parallel"}
      </span>
      {phase.inputs.exclude.length > 0 && (
        <>
          <span className="text-muted-foreground/40">·</span>
          <span className="text-rose-300/90 font-mono text-[9px]">
            blind-to: {phase.inputs.exclude.join(",")}
          </span>
        </>
      )}
    </div>
  );
}
