"use client";

import { ArrowRight, DollarSign } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import type { BillingMode } from "@/lib/api/settings";
import { uiLineageDot } from "@/lib/lineage-maps";
import type { Template } from "@/lib/types";
import type { CostEstimate } from "./helpers";

interface PromptCardProps {
  template: Template;
  prompt: string;
  setPrompt: (v: string) => void;
  reviewOnly: boolean;
  artifactSpec: { label: string; hint: string; maxBytes: number } | undefined;
  billingMode: BillingMode;
  costEstimate: CostEstimate;
  overCap: boolean;
  isPending: boolean;
  onStart: () => void;
}

export function PromptCard({
  template,
  prompt,
  setPrompt,
  reviewOnly,
  artifactSpec,
  billingMode,
  costEstimate,
  overCap,
  isPending,
  onStart,
}: PromptCardProps) {
  return (
    <>
      {reviewOnly && artifactSpec && (
        <div className="mb-2 flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
          <span className="font-medium text-foreground">{artifactSpec.label}</span>
          <span className="text-muted-foreground/60">·</span>
          <span>cap {(artifactSpec.maxBytes / 1024).toLocaleString()} KB</span>
        </div>
      )}
      <Card className="overflow-hidden p-0 mb-4">
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder={
            reviewOnly && artifactSpec
              ? artifactSpec.hint
              : "Describe what you want chorus to weigh in on. Paste code, errors, design docs — anything the reviewers should see."
          }
          className={`block w-full resize-none border-0 bg-transparent px-5 py-4 text-sm text-foreground placeholder:text-muted-foreground/70 focus:outline-none ${
            reviewOnly ? "font-mono text-[12px] leading-relaxed" : ""
          }`}
          rows={reviewOnly ? 16 : 10}
          spellCheck={!reviewOnly}
        />

        <div className="flex flex-col gap-3 border-t border-border bg-card/40 px-5 py-3 sm:flex-row sm:items-center sm:justify-between">
          <RoleSummary template={template} reviewOnly={reviewOnly} />
          <div className="flex items-center justify-between gap-3 sm:justify-end">
            <CostPreview
              billingMode={billingMode}
              costEstimate={costEstimate}
              overCap={overCap}
              template={template}
            />
            <button
              type="button"
              onClick={onStart}
              disabled={!prompt || overCap || isPending}
              className={`inline-flex shrink-0 items-center gap-2 whitespace-nowrap rounded-md px-4 py-2 text-sm font-medium transition ${
                !prompt || overCap || isPending
                  ? "cursor-not-allowed bg-muted text-muted-foreground"
                  : "bg-primary text-primary-foreground hover:bg-primary/90"
              }`}
            >
              {isPending ? "Starting..." : reviewOnly ? "Send for review" : "Start the run"}
              <ArrowRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      </Card>
    </>
  );
}

function RoleSummary({
  template,
  reviewOnly,
}: {
  template: Template;
  reviewOnly: boolean;
}) {
  return (
    <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
      {/* No doer card for review-only — there's no spawned doer; the
          artifact IS the input. The reviewers row carries all the
          information the user needs about who'll critique it. */}
      {!reviewOnly && (
        <>
          <span className="flex items-center gap-1.5">
            <span
              className={`h-1.5 w-1.5 rounded-full ${uiLineageDot(template.phases[0]?.doer.lineage)}`}
            />
            Doer: {template.phases[0]?.doer.lineage}
          </span>
          <span className="text-muted-foreground/50">·</span>
        </>
      )}
      <ReviewerChips template={template} />
      {template.phases.length > 1 && (
        <>
          <span className="text-muted-foreground/50">·</span>
          <span className="font-mono text-[10px]">
            {template.phases.length} phases
          </span>
        </>
      )}
    </div>
  );
}

function ReviewerChips({ template }: { template: Template }) {
  // Two reviewer rows on the same lineage are indistinguishable at
  // "codex · gemini · opencode · opencode" — caught in the 2026-05-03
  // UX walk. When ≥2 candidates share a lineage, append the model so
  // the user can tell e.g. opencode-go/kimi from opencode-go/deepseek
  // before submitting.
  const slots = template.phases[0]?.reviewer.candidatesWithModels ?? [];
  const lineageCounts = new Map<string, number>();
  for (const slot of slots) {
    lineageCounts.set(slot.lineage, (lineageCounts.get(slot.lineage) ?? 0) + 1);
  }
  return (
    <span className="flex items-center gap-1.5">
      <span className="text-muted-foreground/80">Reviewers:</span>
      {slots.map((slot, i) => {
        const dup = (lineageCounts.get(slot.lineage) ?? 0) > 1;
        const modelLabel = slot.models?.[0]?.split("/").pop() ?? slot.models?.[0];
        const label = dup && modelLabel ? `${slot.lineage} · ${modelLabel}` : slot.lineage;
        return (
          <span
            key={`${slot.lineage}-${slot.models?.[0] ?? "default"}-${i}`}
            className="flex items-center gap-1"
            title={slot.models?.[0] ? `${slot.lineage} · ${slot.models[0]}` : slot.lineage}
          >
            <span
              className={`h-1.5 w-1.5 rounded-full ${uiLineageDot(slot.lineage)}`}
            />
            {label}
          </span>
        );
      })}
      {template.phases[0]?.reviewer.crossLineage && (
        <Badge
          variant="outline"
          className="ml-1 border-border font-mono text-[9px] uppercase"
          title="Reviewer must be a different lineage than the doer"
        >
          cross-lineage
        </Badge>
      )}
    </span>
  );
}

function CostPreview({
  billingMode,
  costEstimate,
  overCap,
  template,
}: {
  billingMode: BillingMode;
  costEstimate: CostEstimate;
  overCap: boolean;
  template: Template;
}) {
  // Subscription users see "Subscription quota" with token count; API
  // users see a low–high range that reflects the maxRounds retry
  // multiplier.
  if (billingMode === "subscription") {
    return (
      <div
        className="flex items-center gap-1.5 font-mono text-[10px] text-muted-foreground"
        title={`~${costEstimate.inputTokens.toLocaleString()} input tokens × ${costEstimate.reviewerCount} reviewers — counts against your CLI subscription quota, not billed per call`}
      >
        Subscription quota · ~{costEstimate.inputTokens.toLocaleString()} tok
      </div>
    );
  }
  return (
    <div
      className={`flex items-center gap-1.5 font-mono text-[10px] ${
        overCap ? "text-rose-300" : "text-muted-foreground"
      }`}
      title={`~${costEstimate.inputTokens.toLocaleString()} input tokens × ${costEstimate.reviewerCount} reviewers; up to ${costEstimate.maxRounds} round${costEstimate.maxRounds > 1 ? "s" : ""} on disagreement`}
    >
      <DollarSign className="h-3 w-3" />
      {costEstimate.maxRounds > 1
        ? `~$${costEstimate.usd.toFixed(3)} – $${costEstimate.usdRangeMax.toFixed(3)} est`
        : `~$${costEstimate.usd.toFixed(3)} est`}
      {template.costCapUsd > 0 && (
        <span className="text-muted-foreground/60">
          / cap ${template.costCapUsd.toFixed(2)}
        </span>
      )}
    </div>
  );
}
