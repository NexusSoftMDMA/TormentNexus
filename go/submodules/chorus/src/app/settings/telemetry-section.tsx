"use client";

import { useEffect, useState } from "react";
import { Eye } from "lucide-react";
import {
  getTelemetryStatus,
  updateTelemetryEnabled,
  type TelemetryStatus,
} from "@/lib/api/settings";
import { Section } from "./primitives";

/**
 * Anonymous telemetry opt-out. Three independent paths disable telemetry
 * (env var, touch-file, settings DB); this section drives the third while
 * surfacing the other two so a toggle that secretly does nothing is never
 * shown — env / file always trump the toggle and we say so inline.
 */
export function TelemetrySection() {
  const [status, setStatus] = useState<TelemetryStatus | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getTelemetryStatus()
      .then((res) => {
        if (!cancelled) setStatus(res);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Could not load telemetry status.");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const overridden = status?.envOverride || status?.fileOverride;

  const onToggle = (next: boolean): void => {
    if (pending || !status || overridden) return;
    setPending(true);
    setError(null);
    updateTelemetryEnabled(next)
      .then(setStatus)
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Save failed.");
      })
      .finally(() => setPending(false));
  };

  return (
    <Section
      icon={<Eye className="h-4 w-4" />}
      title="Anonymous telemetry"
      subtitle="A small heartbeat sent to chorus.codes once per boot + every 24h. Helps us see what versions / OSes are in the wild. No code, no prompts, no chat content."
    >
      {error && (
        <div className="mb-3 rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          {error}
        </div>
      )}

      <div className="rounded-lg border border-border bg-card p-3">
        <label className="flex items-start gap-3">
          <input
            type="checkbox"
            checked={status?.enabled ?? false}
            onChange={(e) => onToggle(e.target.checked)}
            disabled={pending || !status || overridden}
            className="mt-1 h-4 w-4 cursor-pointer accent-primary disabled:cursor-not-allowed"
          />
          <span className="flex-1">
            <span className="block text-sm font-medium">
              Send anonymous heartbeat
              {pending && (
                <span className="ml-2 rounded bg-primary/15 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider text-primary">
                  saving…
                </span>
              )}
            </span>
            <span className="mt-1 block text-xs leading-relaxed text-muted-foreground">
              Sends: install UUID, chorus version, OS, arch, node major,
              daemon uptime, count of chats in the last 24h. That&apos;s it.
            </span>
          </span>
        </label>
      </div>

      {overridden && (
        <div className="mt-3 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
          Telemetry is currently disabled by{" "}
          {status?.envOverride && (
            <code className="rounded bg-muted px-1">CHORUS_TELEMETRY</code>
          )}
          {status?.envOverride && status?.fileOverride && " and "}
          {status?.fileOverride && (
            <code className="rounded bg-muted px-1">~/.chorus/no-telemetry</code>
          )}
          . The toggle above won&apos;t take effect until you remove the override.
        </div>
      )}

      <details className="mt-3 text-xs text-muted-foreground">
        <summary className="cursor-pointer hover:text-foreground">
          Other ways to opt out
        </summary>
        <ul className="mt-2 list-inside list-disc space-y-1 leading-relaxed">
          <li>
            Set <code className="rounded bg-muted px-1">CHORUS_TELEMETRY=0</code>{" "}
            in your shell — also accepts <code>false</code>, <code>no</code>,{" "}
            <code>off</code>.
          </li>
          <li>
            <code className="rounded bg-muted px-1">touch ~/.chorus/no-telemetry</code>{" "}
            (cargo / brew convention).
          </li>
          <li>
            Reset your install ID:{" "}
            <code className="rounded bg-muted px-1">rm ~/.chorus/install-id</code>
            {" "}— a fresh UUID is minted on the next heartbeat.
          </li>
        </ul>
        {status?.endpoint && (
          <p className="mt-2 leading-relaxed">
            Endpoint: <code className="rounded bg-muted px-1">{status.endpoint}</code>
          </p>
        )}
      </details>
    </Section>
  );
}
