"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import type { Persona } from "@/lib/api/personas";
import type { ReviewerLineage } from "@/lib/cockpit-types";

interface PersonaSelectProps {
  value: string | undefined;
  personas: Persona[];
  onChange: (next: string | undefined) => void;
  /** Tight inline render for reviewer rows; default false renders a labeled block. */
  inline?: boolean;
  ariaLabel?: string;
}

/**
 * Compact persona picker. When the current value isn't in the persona
 * list (template authored elsewhere, persona row deleted), keep showing
 * it as "(missing)" so editing doesn't silently swap it.
 */
export function PersonaSelect({
  value,
  personas,
  onChange,
  inline,
  ariaLabel,
}: PersonaSelectProps) {
  const NONE = "__none__";
  const known = personas.some((p) => p.id === value);
  return (
    <select
      value={value ?? NONE}
      onChange={(e) => {
        const v = e.target.value;
        onChange(v === NONE ? undefined : v);
      }}
      className={cn(
        "h-7 rounded-md border border-border bg-background px-2 text-[11px] focus:border-primary/60 focus:outline-none",
        inline ? "w-32 shrink-0" : "w-full",
      )}
      aria-label={ariaLabel ?? "Persona"}
    >
      <option value={NONE}>(no persona)</option>
      {personas.map((p) => (
        <option key={p.id} value={p.id}>
          {p.label}
        </option>
      ))}
      {value && !known && <option value={value}>{value} (missing)</option>}
    </select>
  );
}

interface ModelSelectProps {
  lineage: ReviewerLineage;
  value: string;
  options: string[];
  onChange: (next: string) => void;
  /** Default model used as placeholder when value is empty. */
  defaultModel?: string;
}

/**
 * Dropdown of enabled models for a given lineage, with a final
 * "(custom — type your own)" option that swaps in a freeform input.
 * Preserves a value that isn't in the options list (template authored
 * elsewhere, or model since disabled) by including it as an extra
 * option marked "(not enabled)".
 */
export function ModelSelect({
  lineage,
  value,
  options,
  onChange,
  defaultModel,
}: ModelSelectProps) {
  const CUSTOM = "__custom__";
  const valueInOptions = !value || options.includes(value);
  const [editing, setEditing] = useState(false);

  if (editing) {
    return (
      <div className="flex items-center gap-1">
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={
            defaultModel ? `default: ${defaultModel}` : `${lineage} model id`
          }
          className="h-7 flex-1 rounded-md border border-border bg-background px-2 font-mono text-[11px] text-foreground placeholder:text-muted-foreground/50 focus:border-primary/60 focus:outline-none"
          autoFocus
        />
        <button
          type="button"
          onClick={() => setEditing(false)}
          className="h-7 rounded-md border border-border bg-card/40 px-2 text-[10px] text-muted-foreground hover:text-foreground"
        >
          done
        </button>
      </div>
    );
  }

  if (options.length === 0) {
    return (
      <div className="flex flex-col gap-1">
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={
            defaultModel ? `default: ${defaultModel}` : `${lineage} model id`
          }
          className="h-7 w-full rounded-md border border-border bg-background px-2 font-mono text-[11px] text-foreground placeholder:text-muted-foreground/50 focus:border-primary/60 focus:outline-none"
        />
        <p className="text-[10px] text-amber-400/80">
          No {lineage} voices enabled. Configure in Connect to populate this
          dropdown.
        </p>
      </div>
    );
  }

  // Surface authored-but-not-enabled model so save doesn't silently lose it.
  const allOptions = valueInOptions ? options : [value, ...options];

  return (
    <select
      value={value}
      onChange={(e) => {
        if (e.target.value === CUSTOM) {
          setEditing(true);
          return;
        }
        onChange(e.target.value);
      }}
      className="h-7 w-full rounded-md border border-border bg-background px-2 font-mono text-[11px] text-foreground focus:border-primary/60 focus:outline-none"
    >
      {!value && (
        <option value="">
          {defaultModel ? `default: ${defaultModel}` : "— pick a model —"}
        </option>
      )}
      {allOptions.map((m) => (
        <option key={m} value={m}>
          {m}
          {!options.includes(m) && value === m ? " (not enabled)" : ""}
        </option>
      ))}
      <option value={CUSTOM}>+ custom (type a model id)…</option>
    </select>
  );
}
