"use client";

import { useState } from "react";
import { CheckCircle2, ChevronLeft, ChevronRight } from "lucide-react";
import { PhaseEditor } from "@/components/phase-editor";
import { cn } from "@/lib/utils";
import { BasicsStep } from "./basics-step";
import { WIZARD_STEPS } from "./constants";
import { FallbackStep } from "./fallback-step";
import { PolicyStep } from "./policy-step";
import type { FormState } from "./types";

function validateStep(step: number, form: FormState): string[] {
  const issues: string[] = [];
  if (step === 1) {
    if (!form.name || form.name.trim().length === 0) {
      issues.push("Name is required.");
    } else if (form.name.length > 80) {
      issues.push("Name is too long (max 80 chars).");
    }
    if (!form.description || form.description.trim().length === 0) {
      issues.push(
        "Description is required — one line on when to use this template.",
      );
    }
  } else if (step === 2) {
    if (form.phases.length === 0) {
      issues.push("At least one phase is required.");
    }
    for (const phase of form.phases) {
      if (!phase.id || phase.id.trim().length === 0) {
        issues.push(`Phase "${phase.name || "(unnamed)"}" needs an id.`);
      }
      if (
        phase.kind !== "review_only" &&
        (!phase.doer.lineage || phase.doer.models.length === 0)
      ) {
        issues.push(
          `Phase "${phase.name || phase.id}" needs a doer lineage + model.`,
        );
      }
      if (phase.reviewer.candidates.length === 0) {
        issues.push(
          `Phase "${phase.name || phase.id}" needs at least one reviewer.`,
        );
      }
    }
  }
  // Steps 3 (Fallback) + 4 (Policy) are optional.
  return issues;
}

/**
 * 4-step wizard. Walking the user through {basics → phases → fallback →
 * policy} keeps each screen short and lets us gate Next on per-step
 * validation. Save (in the parent dialog footer) is gated by global YAML
 * validation; the wizard's Next/Back is independent — power users can
 * land on the last step and click Save without walking every screen.
 */
export function FormPanel({
  form,
  setField,
}: {
  form: FormState;
  setField: <K extends keyof FormState>(k: K, v: FormState[K]) => void;
}) {
  const [step, setStep] = useState<1 | 2 | 3 | 4>(1);
  // Errors only show when the user TRIED to advance past a step with
  // issues — keeps the form quiet for power users who fill top-down.
  const [showErrors, setShowErrors] = useState(false);

  const issues = validateStep(step, form);
  const canAdvance = issues.length === 0;

  const goNext = () => {
    if (!canAdvance) {
      setShowErrors(true);
      return;
    }
    setShowErrors(false);
    if (step < 4) setStep((step + 1) as 1 | 2 | 3 | 4);
  };
  const goBack = () => {
    setShowErrors(false);
    if (step > 1) setStep((step - 1) as 1 | 2 | 3 | 4);
  };
  const goTo = (target: 1 | 2 | 3 | 4) => {
    // Backward jumps freely; forward jumps require all intermediate
    // steps to validate.
    if (target <= step) {
      setShowErrors(false);
      setStep(target);
      return;
    }
    for (let s = step; s < target; s++) {
      if (validateStep(s, form).length > 0) {
        setShowErrors(true);
        return;
      }
    }
    setShowErrors(false);
    setStep(target);
  };

  return (
    <div className="flex h-full flex-col px-6 py-6">
      <StepIndicator currentStep={step} onJump={goTo} form={form} />

      <div className="mt-6 flex-1 space-y-6">
        {step === 1 && (
          <BasicsStep form={form} setField={setField} showErrors={showErrors} />
        )}
        {step === 2 && (
          <PhasesStep form={form} setField={setField} showErrors={showErrors} />
        )}
        {step === 3 && <FallbackStep form={form} setField={setField} />}
        {step === 4 && <PolicyStep form={form} setField={setField} />}
      </div>

      {showErrors && issues.length > 0 && (
        <div className="mt-6 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-[11px] text-destructive">
          <div className="mb-1 font-medium">Resolve these to continue:</div>
          <ul className="space-y-0.5 pl-4">
            {issues.map((i) => (
              <li key={i} className="list-disc">
                {i}
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="mt-6 flex items-center justify-between border-t border-border pt-4">
        <button
          type="button"
          onClick={goBack}
          disabled={step === 1}
          className="inline-flex h-9 items-center gap-1.5 rounded-md border border-border bg-card px-3 text-sm font-medium text-muted-foreground transition hover:border-muted-foreground/30 hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
        >
          <ChevronLeft className="h-3.5 w-3.5" />
          Back
        </button>
        <span className="text-[11px] text-muted-foreground">
          Step {step} of 4 · {WIZARD_STEPS[step - 1].label}
        </span>
        {step < 4 ? (
          <button
            type="button"
            onClick={goNext}
            className={cn(
              "inline-flex h-9 items-center gap-1.5 rounded-md px-3 text-sm font-medium shadow-sm transition",
              canAdvance
                ? "bg-primary text-primary-foreground hover:bg-primary/90"
                : "bg-muted text-muted-foreground/60",
            )}
          >
            Next
            <ChevronRight className="h-3.5 w-3.5" />
          </button>
        ) : (
          <span className="text-[11px] text-emerald-300">
            <CheckCircle2 className="mr-1 inline-block h-3 w-3" />
            Save in the footer below
          </span>
        )}
      </div>
    </div>
  );
}

function StepIndicator({
  currentStep,
  onJump,
  form,
}: {
  currentStep: 1 | 2 | 3 | 4;
  onJump: (n: 1 | 2 | 3 | 4) => void;
  form: FormState;
}) {
  return (
    <div className="flex items-stretch gap-2">
      {WIZARD_STEPS.map((s) => {
        const Icon = s.icon;
        const isActive = s.id === currentStep;
        const stepIssues = validateStep(s.id, form);
        const valid = stepIssues.length === 0;
        const isPast = s.id < currentStep;
        return (
          <button
            key={s.id}
            type="button"
            onClick={() => onJump(s.id)}
            className={cn(
              "flex flex-1 items-center gap-2 rounded-md border px-3 py-2 text-left transition",
              isActive
                ? "border-primary/60 bg-primary/10"
                : isPast && !valid
                  ? "border-destructive/40 bg-destructive/5"
                  : "border-border bg-card hover:border-muted-foreground/30",
            )}
          >
            <span
              className={cn(
                "flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[10px] font-mono",
                isActive
                  ? "bg-primary text-primary-foreground"
                  : isPast && valid
                    ? "bg-emerald-500/20 text-emerald-300"
                    : isPast && !valid
                      ? "bg-destructive/20 text-destructive"
                      : "bg-muted text-muted-foreground",
              )}
            >
              {isPast && valid ? "✓" : s.id}
            </span>
            <Icon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <span className="hidden text-xs font-medium sm:inline">
              {s.label}
            </span>
          </button>
        );
      })}
    </div>
  );
}

function PhasesStep({
  form,
  setField,
  showErrors,
}: {
  form: FormState;
  setField: <K extends keyof FormState>(k: K, v: FormState[K]) => void;
  showErrors: boolean;
}) {
  // PhaseEditor surfaces its own per-phase errors — on validation failure
  // a small banner reminds the user the issue is in here.
  const issues = validateStep(2, form);
  return (
    <>
      {showErrors && issues.length > 0 && (
        <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-[11px] text-destructive">
          {issues[0]}
        </div>
      )}
      <PhaseEditor
        phases={form.phases}
        onChange={(phases) => setField("phases", phases)}
      />
    </>
  );
}
