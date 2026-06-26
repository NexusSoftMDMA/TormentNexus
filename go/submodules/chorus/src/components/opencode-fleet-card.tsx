"use client";

/**
 * Inline-expandable OpenCode card for the home-page reviewer fleet.
 *
 * Shows count of currently-enabled voices; clicking expands a panel with
 * the gateway-grouped checkbox grid. Toggling a model PUTs /voices/:id
 * immediately so the change is live without a save button.
 *
 * Data source: voices table, filtered to provider='opencode-cli'. The
 * daemon's Phase 2 background warmup populates these from `opencode
 * models`. Includes both enabled + disabled rows (per round 1 cdx-1
 * BLOCKER) so users can re-enable from the fleet card.
 */

import { useState } from "react";
import { CheckCircle2, AlertTriangle, ChevronDown, Check } from "lucide-react";
import { UI_LINEAGE_BRAND } from "@/lib/lineage-maps";
import { updateVoice, type Voice } from "@/lib/api/voices";
import { DaemonError } from "@/lib/api/client";
import { cn } from "@/lib/utils";

interface OpencodeFleetCardProps {
  health: {
    status: "healthy" | "quota_exhausted" | "auth_invalid" | "rate_limited" | "unknown";
    message?: string;
  };
  /** OpenCode voices — both enabled and disabled. */
  voices: Voice[];
}

export function OpencodeFleetCard({ health, voices: initialVoices }: OpencodeFleetCardProps) {
  const [open, setOpen] = useState(false);
  const [voices, setVoices] = useState<Voice[]>(initialVoices);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saving, setSaving] = useState<string | null>(null);

  async function toggleVoice(v: Voice) {
    setSaving(v.id);
    setSaveError(null);
    try {
      const next = await updateVoice(v.id, { enabled: !v.enabled });
      setVoices((prev) => prev.map((p) => (p.id === next.id ? next : p)));
    } catch (err) {
      const message =
        err instanceof DaemonError ? err.message : "Couldn't save. Is the daemon running?";
      setSaveError(message);
    } finally {
      setSaving(null);
    }
  }

  // Group voices by gateway prefix for the card layout. The `model_id`
  // column carries the gateway-prefixed qualified name (e.g.
  // "opencode-go/kimi-k2.6") so the prefix is everything up to the
  // first slash.
  const grouped = new Map<string, Voice[]>();
  for (const v of voices) {
    const slash = v.model_id.indexOf("/");
    const gw = slash > 0 ? v.model_id.slice(0, slash) : "other";
    const list = grouped.get(gw) ?? [];
    list.push(v);
    grouped.set(gw, list);
  }
  const sortedGateways = Array.from(grouped.keys()).sort();

  const enabledCount = voices.filter((v) => v.enabled).length;

  return (
    <div
      className={cn(
        "rounded-lg border border-border transition-colors hover:border-foreground/20",
        UI_LINEAGE_BRAND.opencode.gradient,
      )}
    >
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-3 p-3 text-left"
      >
        <span
          className={cn(
            "h-2 w-2 shrink-0 rounded-full",
            UI_LINEAGE_BRAND.opencode.dot,
          )}
        />
        <div className="min-w-0 flex-1">
          <div className="text-sm font-medium">OpenCode</div>
          <div className="mt-0.5 flex items-center gap-2">
            <StatusBadge status={health.status} />
            <span className="text-[10px] text-muted-foreground">
              {enabledCount} model{enabledCount === 1 ? "" : "s"} enabled
            </span>
          </div>
        </div>
        <ChevronDown
          className={cn(
            "h-4 w-4 shrink-0 text-muted-foreground transition-transform",
            open && "rotate-180",
          )}
        />
      </button>

      {open && (
        <div className="space-y-3 border-t border-border bg-card/50 p-3">
          {saveError && <p className="text-[11px] text-destructive">{saveError}</p>}

          {voices.length === 0 ? (
            <p className="text-[11px] text-muted-foreground">
              No OpenCode voices detected yet. Run <code className="rounded bg-muted px-1">opencode models</code>{" "}
              and restart the daemon.
            </p>
          ) : (
            <div className="space-y-3">
              {sortedGateways.map((gateway) => {
                const list = grouped.get(gateway) ?? [];
                return (
                  <div key={gateway} className="space-y-1">
                    <p className="text-[11px] font-mono text-muted-foreground/80">{gateway}/</p>
                    <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
                      {list.map((v) => {
                        const sel = v.enabled;
                        const shortName = v.model_id.slice(gateway.length + 1);
                        return (
                          <button
                            key={v.id}
                            type="button"
                            disabled={saving === v.id}
                            onClick={() => toggleVoice(v)}
                            title={v.model_id}
                            className={cn(
                              "flex items-center gap-2 rounded border px-2 py-1.5 text-left text-[11px] transition disabled:opacity-60",
                              sel
                                ? "border-primary/50 bg-primary/10 text-foreground"
                                : "border-border bg-card hover:border-muted-foreground/30 text-muted-foreground",
                            )}
                          >
                            <div
                              className={cn(
                                "grid h-3 w-3 shrink-0 place-items-center rounded-sm border transition",
                                sel
                                  ? "border-primary bg-primary text-primary-foreground"
                                  : "border-border",
                              )}
                            >
                              {sel && <Check className="h-2 w-2" />}
                            </div>
                            <span className="truncate font-mono">{shortName}</span>
                          </button>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
              <p className="text-[11px] leading-relaxed text-muted-foreground/70">
                Toggles save automatically. Templates and the New Chat dialog will only offer
                models you&apos;ve enabled here.
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: OpencodeFleetCardProps["health"]["status"] }) {
  // Card only renders for already-connected CLIs — see comment in
  // LineageFleetCard's StatusBadge for the rationale.
  switch (status) {
    case "auth_invalid":
    case "quota_exhausted":
    case "rate_limited":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-medium text-destructive">
          <AlertTriangle className="h-3 w-3" />
          {status === "auth_invalid" ? "Auth broken" : status === "quota_exhausted" ? "Quota out" : "Rate-limited"}
        </span>
      );
    case "healthy":
    default:
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-medium text-emerald-300">
          <CheckCircle2 className="h-3 w-3" />
          Connected
        </span>
      );
  }
}
