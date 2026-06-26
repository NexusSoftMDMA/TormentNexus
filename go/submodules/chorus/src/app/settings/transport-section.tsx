"use client";

import { useEffect, useState } from "react";
import { Terminal } from "lucide-react";
import {
  getTransport,
  updateTransport,
  type Transport,
} from "@/lib/api/settings";
import { Section } from "./primitives";

/**
 * Run-mode toggle: subprocess (headless) vs persistent tmux sessions.
 * Tmux is a first-class option, no deprecation timeline — its value is
 * letting users attach to a live agent and take over mid-run for debug
 * or handoff. Headless is the lower-RAM default.
 */
export function TransportSection() {
  const [current, setCurrent] = useState<Transport | null>(null);
  const [descriptions, setDescriptions] = useState<
    Record<Transport, { label: string; description: string }> | undefined
  >(undefined);
  const [tmuxAvailable, setTmuxAvailable] = useState<boolean>(true);
  const [pending, setPending] = useState<Transport | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getTransport()
      .then((res) => {
        if (cancelled) return;
        setCurrent(res.transport);
        setDescriptions(res.descriptions);
        setTmuxAvailable(res.tmuxAvailable !== false);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Could not load transport setting.");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const onPick = (t: Transport): void => {
    if (current === t || pending) return;
    setPending(t);
    setError(null);
    updateTransport({ transport: t })
      .then((res) => {
        setCurrent(res.transport);
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Save failed.");
      })
      .finally(() => setPending(null));
  };

  return (
    <Section
      id="transport"
      icon={<Terminal className="h-4 w-4" />}
      title="Run mode (headless / tmux)"
      subtitle="How chorus runs each CLI. Default is headless (faster, lower RAM). Switch to tmux if you want to attach to a live voice and take over mid-run (debug + handoff)."
    >
      {error && (
        <div className="mb-3 rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          {error}
        </div>
      )}
      <div className="grid gap-2 sm:grid-cols-2">
        {(["headless", "tmux"] as const).map((t) => {
          const active = current === t;
          const isPending = pending === t;
          const meta = descriptions?.[t];
          // Disable tmux on hosts where the binary isn't reachable — greyed
          // state + inline install hint beats a 400 on click.
          const unavailable = t === "tmux" && !tmuxAvailable;
          return (
            <button
              key={t}
              type="button"
              onClick={() => onPick(t)}
              disabled={pending !== null || current === null || unavailable}
              title={
                unavailable
                  ? "tmux is not installed on this host. Install it (brew/apt/dnf) then restart the daemon."
                  : undefined
              }
              className={`flex flex-col gap-1 rounded-lg border p-3 text-left transition ${
                active
                  ? "border-primary/50 bg-primary/10"
                  : "border-border bg-card hover:border-muted-foreground/30"
              } ${pending !== null || unavailable ? "opacity-60 cursor-not-allowed" : ""}`}
            >
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">
                  {meta?.label ?? (t === "headless" ? "Headless" : "Tmux")}
                </span>
                {active ? (
                  <span className="rounded bg-primary/15 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-primary">
                    {isPending ? "saving…" : "active"}
                  </span>
                ) : unavailable ? (
                  <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                    not installed
                  </span>
                ) : null}
              </div>
              <p className="text-xs leading-relaxed text-muted-foreground">
                {meta?.description ??
                  (t === "headless"
                    ? "Subprocess per call. Lower RAM, faster cold start, no permission dialogs."
                    : "Persistent terminal sessions you can attach to for visual debug.")}
              </p>
              {unavailable && (
                <p className="mt-1 text-[10px] leading-relaxed text-amber-300/80">
                  Install: macOS <code>brew install tmux</code> · Ubuntu <code>apt install tmux</code> · Fedora <code>dnf install tmux</code>. Restart the daemon afterwards.
                </p>
              )}
            </button>
          );
        })}
      </div>
      <p className="mt-3 text-[11px] text-muted-foreground">
        Tip: set <code className="rounded bg-muted px-1">CHORUS_TRANSPORT=tmux</code> in
        your environment to override per-shell without changing this setting.
      </p>
    </Section>
  );
}
