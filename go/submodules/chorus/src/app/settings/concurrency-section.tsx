"use client";

import { useEffect, useState } from "react";
import { Activity } from "lucide-react";
import {
  getConcurrencySettings,
  updateConcurrencySettings,
  type ConcurrencySettings,
} from "@/lib/api/settings";
import { Section } from "./primitives";

/**
 * Concurrency caps — daemon-wide.
 *
 * Two layers:
 *   - Global: max parallel local-CLI subprocesses across the whole daemon.
 *   - Per-CLI: subset cap per binary family.
 *
 * Both compose: a reviewer must acquire BOTH a global slot AND a per-CLI
 * slot before spawning. Whichever is tighter is the queue. HTTP shims
 * (openrouter) bypass entirely — they're network calls.
 *
 * Save-on-blur: the daemon takes effect on the next reviewer that
 * starts, no restart needed.
 */
export function ConcurrencySection() {
  const [data, setData] = useState<ConcurrencySettings | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getConcurrencySettings()
      .then((res) => {
        if (cancelled) return;
        setData(res);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(
          err instanceof Error
            ? err.message
            : "Could not load concurrency settings.",
        );
      });
    return () => {
      cancelled = true;
    };
  }, []);

  /**
   * Save a single field. Used by both the global input and the per-CLI
   * inputs. Optimistic local update on success — `data` reflects the
   * server's authoritative response, so a server-side clamp (e.g.
   * out-of-range coerced to default) shows up in the input immediately.
   */
  const save = async (
    patch: { maxParallelCli?: number; perCli?: Record<string, number> },
  ): Promise<void> => {
    if (pending) return;
    setPending(true);
    setError(null);
    try {
      const next = await updateConcurrencySettings(patch);
      setData(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed.");
    } finally {
      setPending(false);
    }
  };

  if (!data) {
    return (
      <Section
        id="concurrency"
        icon={<Activity className="h-4 w-4" />}
        title="Concurrency"
        subtitle="Max parallel CLI subprocesses across all chats. Caps memory pressure when stacking heavyweight reviewers."
      >
        <div className="text-xs text-muted-foreground">
          {error ?? "Loading…"}
        </div>
      </Section>
    );
  }

  const { maxParallelCli, perCli, cliLineages = [], defaults } = data;

  return (
    <Section
      id="concurrency"
      icon={<Activity className="h-4 w-4" />}
      title="Concurrency"
      subtitle="Max parallel CLI subprocesses across all chats (daemon-wide). HTTP shims like OpenRouter bypass these caps — they don't consume local resources."
    >
      {error && (
        <div className="mb-3 rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          {error}
        </div>
      )}

      <div className="space-y-4">
        {/* Global */}
        <div className="flex items-center justify-between gap-4">
          <div className="min-w-0">
            <label
              htmlFor="cc-global"
              className="block text-sm font-medium text-foreground"
            >
              Global cap
            </label>
            <p className="mt-0.5 text-xs text-muted-foreground">
              Total local-CLI subprocesses across the whole daemon. Applies
              to reviewers + doer combined.
            </p>
          </div>
          <NumberField
            id="cc-global"
            min={1}
            max={10}
            value={maxParallelCli}
            disabled={pending}
            onCommit={(n) => save({ maxParallelCli: n })}
          />
        </div>

        {/* Per-CLI */}
        <div className="border-t border-border pt-4">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Per-CLI caps
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            Tightening a single binary family — e.g. opencode is heavy
            (~450 MB / proc) so default is 2.
          </p>
          <div className="mt-3 space-y-2">
            {cliLineages.map((lineage) => {
              const current = perCli[lineage] ?? defaults?.perCli[lineage] ?? 2;
              return (
                <div
                  key={lineage}
                  className="flex items-center justify-between gap-4"
                >
                  <label
                    htmlFor={`cc-${lineage}`}
                    className="font-mono text-sm text-foreground"
                  >
                    {lineage}
                  </label>
                  <NumberField
                    id={`cc-${lineage}`}
                    min={1}
                    max={5}
                    value={current}
                    disabled={pending}
                    onCommit={(n) => save({ perCli: { [lineage]: n } })}
                  />
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </Section>
  );
}

interface NumberFieldProps {
  id: string;
  min: number;
  max: number;
  value: number;
  disabled?: boolean;
  onCommit: (n: number) => void;
}

/**
 * Compact +/- stepper. Commit on click; clamp out-of-range silently
 * (server validates anyway, this just keeps the UI honest).
 */
function NumberField({ id, min, max, value, disabled, onCommit }: NumberFieldProps) {
  const dec = () => {
    if (disabled) return;
    const next = Math.max(min, value - 1);
    if (next !== value) onCommit(next);
  };
  const inc = () => {
    if (disabled) return;
    const next = Math.min(max, value + 1);
    if (next !== value) onCommit(next);
  };

  return (
    <div className="flex items-center rounded-md border border-border bg-card">
      <button
        type="button"
        onClick={dec}
        disabled={disabled || value <= min}
        aria-label={`Decrease ${id}`}
        className="px-2.5 py-1.5 text-sm text-muted-foreground transition hover:text-foreground disabled:opacity-30"
      >
        −
      </button>
      <input
        id={id}
        type="text"
        readOnly
        value={value}
        className="w-9 border-x border-border bg-transparent text-center font-mono text-sm tabular-nums focus:outline-none"
      />
      <button
        type="button"
        onClick={inc}
        disabled={disabled || value >= max}
        aria-label={`Increase ${id}`}
        className="px-2.5 py-1.5 text-sm text-muted-foreground transition hover:text-foreground disabled:opacity-30"
      >
        +
      </button>
    </div>
  );
}
