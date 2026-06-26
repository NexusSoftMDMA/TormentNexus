"use client";

import { AlertTriangle } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Field,
  LINEAGE_OPTIONS,
  type FormState,
  type ValidationIssue,
} from "./helpers";

interface FormBodyProps {
  form: FormState;
  setField: <K extends keyof FormState>(k: K, v: FormState[K]) => void;
  setIdDirty: (dirty: boolean) => void;
  isEdit: boolean;
  isBuiltin: boolean;
  issues: ValidationIssue[];
}

export function FormBody({
  form,
  setField,
  setIdDirty,
  isEdit,
  isBuiltin,
  issues,
}: FormBodyProps) {
  return (
    <div className="min-h-0 flex-1 overflow-y-auto px-6 py-5">
      {isEdit && isBuiltin && (
        <div className="mb-4 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-[11px] text-amber-200">
          <div className="flex items-start gap-2">
            <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            <div>
              Built-in persona. Saving demotes it to a user copy so the
              boot-time seed leaves it alone — you&apos;ll no longer
              receive upstream updates for this id.
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field
          label="Label"
          hint="Human-friendly name shown in the picker"
          issue={issues.find((i) => i.field === "label")}
        >
          <input
            type="text"
            value={form.label}
            onChange={(e) => setField("label", e.target.value)}
            placeholder="Security Skeptic"
            className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm outline-none transition focus:border-primary/60"
          />
        </Field>

        <Field
          label="ID"
          hint={isEdit ? "Locked once created" : "Lowercase, dashes — auto-derived from label"}
          issue={issues.find((i) => i.field === "id")}
        >
          <input
            type="text"
            value={form.id}
            disabled={isEdit}
            onChange={(e) => {
              setIdDirty(true);
              setField("id", e.target.value);
            }}
            placeholder="security-skeptic"
            className={cn(
              "h-9 w-full rounded-md border border-border bg-background px-3 font-mono text-sm outline-none transition focus:border-primary/60",
              isEdit && "cursor-not-allowed opacity-60",
            )}
          />
        </Field>

        <Field
          label="One-liner"
          hint="Shown under the label in the persona list"
          issue={issues.find((i) => i.field === "one_liner")}
          span2
        >
          <input
            type="text"
            value={form.one_liner}
            onChange={(e) => setField("one_liner", e.target.value)}
            placeholder="Reads every diff like an attacker."
            className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm outline-none transition focus:border-primary/60"
          />
        </Field>

        <Field
          label="Recommended lineage"
          hint="Optional — pickers can default to this CLI"
          span2
        >
          <select
            value={form.recommended_lineage}
            onChange={(e) => setField("recommended_lineage", e.target.value)}
            className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm outline-none transition focus:border-primary/60"
          >
            {LINEAGE_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </Field>
      </div>

      <div className="mt-5">
        <Field
          label="System prompt"
          hint="Prepended to the doer / reviewer ask at run time. Markdown."
          issue={issues.find((i) => i.field === "system_prompt")}
        >
          <textarea
            value={form.system_prompt}
            onChange={(e) => setField("system_prompt", e.target.value)}
            spellCheck={false}
            className="min-h-[280px] w-full rounded-md border border-border bg-background px-3 py-2 font-mono text-[12px] leading-relaxed outline-none transition focus:border-primary/60"
          />
        </Field>
      </div>
    </div>
  );
}
