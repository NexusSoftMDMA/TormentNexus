"use client";

import { Check, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Persona } from "@/lib/api/personas";
import type { TemplatePhase } from "@/lib/cockpit-types";
import { DEFAULT_MODELS, LINEAGES } from "./constants";
import type { ConnectedVoiceMap } from "./hooks";
import { SubField, ToggleChip } from "./primitives";
import { ModelSelect, PersonaSelect } from "./selects";
import { ReviewerSlotsEditor, reviewerToRows } from "./reviewer-slots";

interface SectionProps {
  phase: TemplatePhase;
  onUpdate: (patch: Partial<TemplatePhase>) => void;
}

interface DoerSectionProps extends SectionProps {
  connectedVoices: ConnectedVoiceMap;
  personas: Persona[];
}

export function DoerSection({
  phase,
  onUpdate,
  connectedVoices,
  personas,
}: DoerSectionProps) {
  return (
    <SubField label="Doer · the agent that writes this phase's output">
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
        {/* Falls through to the full lineage list on first-run with zero
            voices so the dialog stays usable. */}
        {LINEAGES.filter((l) =>
          connectedVoices.connectedLineages.size === 0
            ? true
            : connectedVoices.connectedLineages.has(l.id),
        ).map((l) => (
          <button
            key={l.id}
            type="button"
            onClick={() => {
              // Prefer an actually-enabled model so the template doesn't
              // reference one the user can't reach.
              const enabledForLineage = connectedVoices.byLineage[l.id] ?? [];
              const fallback = enabledForLineage[0] ?? DEFAULT_MODELS[l.id];
              onUpdate({
                doer: {
                  lineage: l.id,
                  models: fallback ? [fallback] : [],
                },
              });
            }}
            className={cn(
              "flex items-center gap-1.5 rounded-md border px-2 py-1.5 text-left transition",
              phase.doer.lineage === l.id
                ? "border-primary/40 bg-primary/10"
                : "border-border bg-card/40 hover:border-foreground/30",
            )}
          >
            <span className={cn("h-1.5 w-1.5 rounded-full", l.dot)} />
            <span className="text-xs">{l.label}</span>
            {phase.doer.lineage === l.id && (
              <Check className="ml-auto h-3 w-3 text-primary" />
            )}
          </button>
        ))}
      </div>
      <div className="mt-2">
        <ModelSelect
          lineage={phase.doer.lineage}
          value={phase.doer.models[0] ?? ""}
          options={connectedVoices.byLineage[phase.doer.lineage] ?? []}
          defaultModel={DEFAULT_MODELS[phase.doer.lineage]}
          onChange={(next) =>
            onUpdate({
              doer: { ...phase.doer, models: next ? [next] : [] },
            })
          }
        />
      </div>
      <div className="mt-2">
        <label className="mb-1 block text-[10px] text-muted-foreground">
          Persona (optional · prepends a worldview to the doer&apos;s prompt)
        </label>
        <PersonaSelect
          value={phase.doer.persona}
          personas={personas}
          onChange={(next) =>
            onUpdate({
              doer: next
                ? { ...phase.doer, persona: next }
                : (() => {
                    // Strip `persona` cleanly when (none) — preserves the
                    // omit-when-undefined YAML emission contract.
                    const { persona: _drop, ...rest } = phase.doer;
                    void _drop;
                    return rest;
                  })(),
            })
          }
          ariaLabel="Doer persona"
        />
      </div>
    </SubField>
  );
}

export function ArtifactSection({ phase, onUpdate }: SectionProps) {
  const fallback = {
    label: phase.artifact?.label ?? "Artifact to review",
    hint:
      phase.artifact?.hint ??
      "Paste a unified diff, a markdown draft, code, or any text blob.",
    maxBytes: phase.artifact?.maxBytes ?? 1024 * 1024,
  };
  return (
    <SubField label="Artifact · what the user pastes when starting a chat">
      <div className="space-y-2">
        <div>
          <label className="text-[10px] text-muted-foreground">Label</label>
          <input
            type="text"
            value={fallback.label}
            onChange={(e) =>
              onUpdate({
                artifact: { ...fallback, label: e.target.value },
              })
            }
            placeholder="Artifact to review"
            className="mt-1 h-7 w-full rounded-md border border-border bg-background px-2.5 text-[11px] placeholder:text-muted-foreground/60 focus:border-primary/60 focus:outline-none"
          />
        </div>
        <div>
          <label className="text-[10px] text-muted-foreground">
            Placeholder hint
          </label>
          <textarea
            value={phase.artifact?.hint ?? ""}
            onChange={(e) =>
              onUpdate({
                artifact: { ...fallback, hint: e.target.value },
              })
            }
            placeholder="Paste a unified diff, a markdown draft, code, or any text blob."
            rows={2}
            className="mt-1 w-full resize-none rounded-md border border-border bg-background px-2.5 py-1.5 text-[11px] placeholder:text-muted-foreground/60 focus:border-primary/60 focus:outline-none"
          />
        </div>
        <div>
          <label className="text-[10px] text-muted-foreground">
            Max bytes (server-side cap)
          </label>
          <input
            type="number"
            min={1024}
            step={1024}
            value={fallback.maxBytes}
            onChange={(e) =>
              onUpdate({
                artifact: {
                  ...fallback,
                  maxBytes: Math.max(
                    1024,
                    parseInt(e.target.value, 10) || 1024 * 1024,
                  ),
                },
              })
            }
            className="mt-1 h-7 w-full rounded-md border border-border bg-background px-2.5 font-mono text-[11px] focus:border-primary/60 focus:outline-none"
          />
          <p className="mt-1 text-[10px] text-muted-foreground/80">
            1 MiB = 1048576. Default if unset.
          </p>
        </div>
      </div>
    </SubField>
  );
}

interface ReviewersSectionProps extends SectionProps {
  connectedVoices: ConnectedVoiceMap;
  personas: Persona[];
}

export function ReviewersSection({
  phase,
  onUpdate,
  connectedVoices,
  personas,
}: ReviewersSectionProps) {
  return (
    <SubField label="Reviewers · who gates this phase">
      <ReviewerSlotsEditor
        phase={phase}
        connectedVoices={connectedVoices}
        personas={personas}
        onUpdate={onUpdate}
      />
    </SubField>
  );
}

export function ApprovalsSection({ phase, onUpdate }: SectionProps) {
  return (
    <SubField label="Approvals">
      <div className="flex items-center gap-4 flex-wrap">
        <label className="flex items-center gap-2 text-[11px] text-muted-foreground">
          Require:
          <input
            type="number"
            min={0}
            // Cap at the actual expanded reviewer-slot count — multi-model
            // rows can produce >4 slots, and a stale 4 made the input
            // refuse a quorum that targets all of them.
            max={Math.max(1, reviewerToRows(phase.reviewer).length)}
            value={phase.reviewer.require}
            onChange={(e) =>
              onUpdate({
                reviewer: {
                  ...phase.reviewer,
                  require: Math.max(0, parseInt(e.target.value, 10) || 0),
                },
              })
            }
            className="h-6 w-12 rounded border border-border bg-background px-1.5 text-center font-mono text-[11px]"
          />
          approval{phase.reviewer.require === 1 ? "" : "s"}
        </label>
        {phase.kind !== "review_only" && (
          <label className="flex items-center gap-2 text-[11px] text-muted-foreground">
            <input
              type="checkbox"
              checked={phase.reviewer.crossLineage}
              onChange={(e) =>
                onUpdate({
                  reviewer: {
                    ...phase.reviewer,
                    crossLineage: e.target.checked,
                  },
                })
              }
              className="accent-primary"
            />
            Cross-lineage required
          </label>
        )}
      </div>
    </SubField>
  );
}

export function ExecutionSection({ phase, onUpdate }: SectionProps) {
  return (
    <SubField label="Reviewer execution · how multiple reviewers interact">
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        <button
          type="button"
          onClick={() => onUpdate({ execution: "parallel" })}
          className={cn(
            "rounded-md border px-3 py-2.5 text-left transition",
            phase.execution === "parallel"
              ? "border-emerald-500/40 bg-emerald-500/5"
              : "border-border bg-card/40 hover:border-foreground/30",
          )}
        >
          <div className="flex items-center gap-1.5 text-xs font-medium">
            Parallel independent
            <span className="font-mono text-[9px] uppercase text-emerald-300/80">
              fast
            </span>
          </div>
          <p className="mt-0.5 text-[10px] leading-snug text-muted-foreground">
            All reviewers see the same pack, work in isolation, vote.
            Catches single-reviewer mistakes via consensus. Risk: shared
            blind spots pass anyway.
          </p>
        </button>
        <button
          type="button"
          onClick={() => onUpdate({ execution: "sequential" })}
          className={cn(
            "rounded-md border px-3 py-2.5 text-left transition",
            phase.execution === "sequential"
              ? "border-amber-500/40 bg-amber-500/5"
              : "border-border bg-card/40 hover:border-foreground/30",
          )}
        >
          <div className="flex items-center gap-1.5 text-xs font-medium">
            Sequential hostile
            <span className="font-mono text-[9px] uppercase text-amber-300/80">
              thorough
            </span>
          </div>
          <p className="mt-0.5 text-[10px] leading-snug text-muted-foreground">
            Reviewers run in order. Each receives the already-hardened
            output and hunts what the prior reviewer missed. Slower; only
            works if the first reviewer&apos;s solution is sound.
          </p>
        </button>
      </div>
    </SubField>
  );
}

interface InputsSectionProps extends SectionProps {
  priorPhases: string[];
}

export function InputsSection({
  phase,
  onUpdate,
  priorPhases,
}: InputsSectionProps) {
  return (
    <SubField label="Inputs · which prior phases this doer can read">
      <div className="space-y-1.5">
        {priorPhases.map((id) => {
          const inc = phase.inputs.include.includes(id);
          const exc = phase.inputs.exclude.includes(id);
          const state: "none" | "include" | "exclude" = exc
            ? "exclude"
            : inc
              ? "include"
              : "none";
          return (
            <div
              key={id}
              className="flex items-center justify-between rounded-md border border-border bg-card/30 px-2.5 py-1.5"
            >
              <code className="font-mono text-[11px] text-foreground">{id}</code>
              <div className="flex items-center gap-1">
                <ToggleChip
                  on={state === "include"}
                  tone="emerald"
                  onClick={() =>
                    onUpdate({
                      inputs: {
                        include: inc
                          ? phase.inputs.include.filter((x) => x !== id)
                          : [...phase.inputs.include, id],
                        exclude: phase.inputs.exclude.filter((x) => x !== id),
                      },
                    })
                  }
                >
                  include
                </ToggleChip>
                <ToggleChip
                  on={state === "exclude"}
                  tone="rose"
                  onClick={() =>
                    onUpdate({
                      inputs: {
                        include: phase.inputs.include.filter((x) => x !== id),
                        exclude: exc
                          ? phase.inputs.exclude.filter((x) => x !== id)
                          : [...phase.inputs.exclude, id],
                      },
                    })
                  }
                >
                  exclude (blind)
                </ToggleChip>
              </div>
            </div>
          );
        })}
      </div>
      <p className="mt-2 flex items-start gap-1.5 text-[10px] text-muted-foreground">
        <Info className="mt-0.5 h-2.5 w-2.5 shrink-0 text-muted-foreground/60" />
        <span>
          <span className="font-mono text-rose-300">exclude</span> prevents
          this doer from seeing that phase&apos;s output. Used for info
          asymmetry — e.g. the implementer shouldn&apos;t see the tests.
        </span>
      </p>
    </SubField>
  );
}

interface IterateSectionProps extends SectionProps {
  priorPhases: string[];
}

export function IterateSection({
  phase,
  onUpdate,
  priorPhases,
}: IterateSectionProps) {
  return (
    <SubField label="Iterate · what happens when reviewer rejects">
      <div className="grid grid-cols-3 gap-2">
        <div>
          <label className="text-[10px] text-muted-foreground">Max revise</label>
          <input
            type="number"
            min={1}
            max={10}
            value={phase.iterate.max}
            onChange={(e) =>
              onUpdate({
                iterate: {
                  ...phase.iterate,
                  max: Math.max(1, parseInt(e.target.value, 10) || 1),
                },
              })
            }
            className="mt-1 h-7 w-full rounded-md border border-border bg-background px-2 font-mono text-[11px]"
          />
        </div>
        <div>
          <label className="text-[10px] text-muted-foreground">On max</label>
          <select
            value={phase.iterate.onMax}
            onChange={(e) =>
              onUpdate({
                iterate: {
                  ...phase.iterate,
                  onMax: e.target.value as "ask-user" | "loopback" | "fail",
                },
              })
            }
            className="mt-1 h-7 w-full rounded-md border border-border bg-background px-1.5 text-[11px]"
          >
            <option value="ask-user">ask me</option>
            <option value="loopback">loop back</option>
            <option value="fail">fail</option>
          </select>
        </div>
        <div>
          <label className="text-[10px] text-muted-foreground">
            Loop back to
          </label>
          <select
            disabled={phase.iterate.onMax !== "loopback"}
            value={phase.iterate.loopbackTo ?? ""}
            onChange={(e) =>
              onUpdate({
                iterate: {
                  ...phase.iterate,
                  loopbackTo: e.target.value || undefined,
                },
              })
            }
            className="mt-1 h-7 w-full rounded-md border border-border bg-background px-1.5 text-[11px] disabled:opacity-40"
          >
            <option value="">—</option>
            {priorPhases.map((id) => (
              <option key={id} value={id}>
                {id}
              </option>
            ))}
          </select>
        </div>
      </div>
    </SubField>
  );
}

export function GateSection({ phase, onUpdate }: SectionProps) {
  return (
    <SubField label="Gate · what happens when this phase finishes">
      <div className="grid grid-cols-2 gap-2">
        <button
          type="button"
          onClick={() => onUpdate({ gate: "auto" })}
          className={cn(
            "rounded-md border px-2.5 py-1.5 text-left transition",
            phase.gate === "auto"
              ? "border-emerald-500/40 bg-emerald-500/10"
              : "border-border bg-card/40 hover:border-foreground/30",
          )}
        >
          <div className="text-xs font-medium">Auto-proceed</div>
          <div className="text-[10px] text-muted-foreground">
            Continue to next phase without asking.
          </div>
        </button>
        <button
          type="button"
          onClick={() => onUpdate({ gate: "ask-user" })}
          className={cn(
            "rounded-md border px-2.5 py-1.5 text-left transition",
            phase.gate === "ask-user"
              ? "border-amber-500/40 bg-amber-500/10"
              : "border-border bg-card/40 hover:border-foreground/30",
          )}
        >
          <div className="text-xs font-medium">Checkpoint (ask me)</div>
          <div className="text-[10px] text-muted-foreground">
            Stop here, surface verdict, wait for click.
          </div>
        </button>
      </div>
    </SubField>
  );
}
