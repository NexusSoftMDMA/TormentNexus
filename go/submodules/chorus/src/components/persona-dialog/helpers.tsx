"use client";

import { LINEAGE_LABEL, type DaemonLineage } from "@/lib/lineage-maps";
import { cn } from "@/lib/utils";

const PERSONA_ID_RE = /^[a-z0-9][a-z0-9-]{1,63}$/;

export const LINEAGE_OPTIONS: Array<{ value: DaemonLineage | ""; label: string }> = [
  { value: "", label: "No preference" },
  { value: "anthropic", label: LINEAGE_LABEL.anthropic },
  { value: "openai", label: LINEAGE_LABEL.openai },
  { value: "google", label: LINEAGE_LABEL.google },
  { value: "opencode", label: LINEAGE_LABEL.opencode },
  { value: "moonshot", label: LINEAGE_LABEL.moonshot },
];

export interface FormState {
  id: string;
  label: string;
  one_liner: string;
  recommended_lineage: string;
  system_prompt: string;
}

export const DEFAULT_FORM: FormState = {
  id: "",
  label: "",
  one_liner: "",
  recommended_lineage: "",
  system_prompt:
    "# Worldview\n\nDescribe the lens this persona uses to read code and findings.\n\n# What you look for\n\n- ...\n- ...\n\n# Tone\n\nHow the persona writes — terse, exhaustive, contrarian, etc.\n",
};

export function slugify(input: string): string {
  return input
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export interface ValidationIssue {
  field: keyof FormState;
  message: string;
}

export function validate(form: FormState, isEdit: boolean): ValidationIssue[] {
  const issues: ValidationIssue[] = [];
  const id = isEdit ? form.id : form.id || slugify(form.label);
  if (!id) issues.push({ field: "id", message: "id is required" });
  else if (!PERSONA_ID_RE.test(id))
    issues.push({
      field: "id",
      message: "id must be lowercase letters, numbers, dashes (2–64 chars)",
    });
  if (!form.label.trim()) issues.push({ field: "label", message: "label is required" });
  if (!form.one_liner.trim())
    issues.push({ field: "one_liner", message: "one-liner is required" });
  if (!form.system_prompt.trim())
    issues.push({ field: "system_prompt", message: "system prompt is required" });
  return issues;
}

export function Field({
  label,
  hint,
  issue,
  span2,
  children,
}: {
  label: string;
  hint?: string;
  issue?: ValidationIssue;
  span2?: boolean;
  children: React.ReactNode;
}) {
  return (
    <label className={cn("flex flex-col gap-1", span2 && "sm:col-span-2")}>
      <span className="flex items-baseline justify-between gap-2">
        <span className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
          {label}
        </span>
        {hint && (
          <span className="text-[10px] text-muted-foreground/70">{hint}</span>
        )}
      </span>
      {children}
      {issue && (
        <span className="text-[10px] text-destructive">{issue.message}</span>
      )}
    </label>
  );
}
