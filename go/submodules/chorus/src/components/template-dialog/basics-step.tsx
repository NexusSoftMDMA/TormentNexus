"use client";

import { cn } from "@/lib/utils";
import { CATEGORIES } from "./constants";
import { Chip, Field } from "./primitives";
import type { FormState } from "./types";

export function BasicsStep({
  form,
  setField,
  showErrors,
}: {
  form: FormState;
  setField: <K extends keyof FormState>(k: K, v: FormState[K]) => void;
  showErrors: boolean;
}) {
  const nameInvalid = showErrors && form.name.trim().length === 0;
  const descInvalid = showErrors && form.description.trim().length === 0;
  return (
    <>
      <Field label="Name *" htmlFor="tpl-name">
        <input
          id="tpl-name"
          value={form.name}
          onChange={(e) => setField("name", e.target.value)}
          placeholder="security-audit"
          aria-invalid={nameInvalid}
          className={cn(
            "h-10 w-full rounded-md border bg-background px-3 text-sm placeholder:text-muted-foreground/60 focus:outline-none focus:ring-2",
            nameInvalid
              ? "border-destructive/60 focus:border-destructive focus:ring-destructive/30"
              : "border-border focus:border-primary/60 focus:ring-primary/20",
          )}
        />
        {nameInvalid && (
          <p className="text-[11px] text-destructive">Name is required.</p>
        )}
      </Field>

      <Field
        label="Description *"
        htmlFor="tpl-desc"
        hint="One sentence on when someone should reach for this template."
      >
        <textarea
          id="tpl-desc"
          value={form.description}
          onChange={(e) => setField("description", e.target.value)}
          placeholder="Independent security audit from 3 model families…"
          rows={2}
          aria-invalid={descInvalid}
          className={cn(
            "w-full resize-none rounded-md border bg-background px-3 py-2 text-sm leading-relaxed placeholder:text-muted-foreground/60 focus:outline-none focus:ring-2",
            descInvalid
              ? "border-destructive/60 focus:border-destructive focus:ring-destructive/30"
              : "border-border focus:border-primary/60 focus:ring-primary/20",
          )}
        />
        {descInvalid && (
          <p className="text-[11px] text-destructive">
            Description is required.
          </p>
        )}
      </Field>

      <Field label="Category">
        <div className="flex flex-wrap gap-2">
          {CATEGORIES.map((c) => (
            <Chip
              key={c.id}
              active={c.id === form.category}
              onClick={() => setField("category", c.id)}
            >
              {c.label}
            </Chip>
          ))}
        </div>
      </Field>
    </>
  );
}
