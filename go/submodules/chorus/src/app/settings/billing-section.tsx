"use client";

import { useEffect, useState } from "react";
import { CreditCard } from "lucide-react";
import {
  getBillingMode,
  updateBillingMode,
  type BillingMode,
} from "@/lib/api/settings";
import { Section } from "./primitives";

/**
 * Billing-mode toggle. Tells chorus how the user pays for the underlying
 * CLIs so the cost preview on /new can render honestly:
 *   api          → show $ estimate at spot rates
 *   subscription → show "Subscription quota" + token count, no $
 *   mixed        → show worst-case $ with caveat
 */
export function BillingModeSection() {
  const [current, setCurrent] = useState<BillingMode | null>(null);
  const [descriptions, setDescriptions] = useState<
    Record<BillingMode, { label: string; description: string }> | undefined
  >(undefined);
  const [pending, setPending] = useState<BillingMode | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getBillingMode()
      .then((res) => {
        if (cancelled) return;
        setCurrent(res.mode);
        setDescriptions(res.descriptions);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Could not load billing mode.");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const onPick = (m: BillingMode): void => {
    if (current === m || pending) return;
    setPending(m);
    setError(null);
    updateBillingMode({ mode: m })
      .then((res) => setCurrent(res.mode))
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Save failed.");
      })
      .finally(() => setPending(null));
  };

  return (
    <Section
      icon={<CreditCard className="h-4 w-4" />}
      title="Billing mode"
      subtitle="How you pay for the AI CLIs chorus drives. Affects what the cost preview shows on the new-chat page — defaults to API rates so the number stays honest until you tell us otherwise."
    >
      {error && (
        <div className="mb-3 rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          {error}
        </div>
      )}
      <div className="grid gap-2">
        {(["api", "subscription", "mixed"] as const).map((m) => {
          const active = current === m;
          const isPending = pending === m;
          const meta = descriptions?.[m];
          return (
            <button
              key={m}
              type="button"
              onClick={() => onPick(m)}
              disabled={pending !== null || current === null}
              className={`flex flex-col gap-1 rounded-lg border p-3 text-left transition ${
                active
                  ? "border-primary/50 bg-primary/10"
                  : "border-border bg-card hover:border-muted-foreground/30"
              } ${pending !== null ? "opacity-60" : ""}`}
            >
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">
                  {meta?.label ?? m}
                </span>
                {active && (
                  <span className="rounded bg-primary/15 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-primary">
                    {isPending ? "saving…" : "active"}
                  </span>
                )}
              </div>
              <p className="text-xs leading-relaxed text-muted-foreground">
                {meta?.description ?? ""}
              </p>
            </button>
          );
        })}
      </div>
    </Section>
  );
}
