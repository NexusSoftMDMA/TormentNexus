"use client";

/**
 * Inline-expandable OpenRouter card for the home-page reviewer fleet.
 * Mirrors LineageFleetCard's shape but for HTTP-dispatched OpenRouter
 * voices (provider='openrouter', source='api'). No /cli/health probe —
 * OpenRouter is HTTP-only, so health is implicit ("Configured" if any
 * voices exist; the panel filters this card out otherwise).
 */

import { useState } from "react";
import Link from "next/link";
import {
  CheckCircle2,
  ChevronDown,
  Check,
  AlertTriangle,
  ArrowRight,
  Clock,
} from "lucide-react";
import { UI_LINEAGE_BRAND } from "@/lib/lineage-maps";
import { updateVoice, type Voice } from "@/lib/api/voices";
import { DaemonError } from "@/lib/api/client";
import { cn } from "@/lib/utils";

type HealthStatus =
  | "healthy"
  | "quota_exhausted"
  | "auth_invalid"
  | "rate_limited"
  | "unknown";

interface OpenRouterFleetCardProps {
  voices: Voice[];
  health?: { status: HealthStatus; message?: string };
}

export function OpenRouterFleetCard({
  voices: initialVoices,
  health,
}: OpenRouterFleetCardProps) {
  // Auto-expand when auth is broken so the "Fix on Connect" CTA is visible
  // without the user having to discover that the card is clickable.
  const isBroken = health?.status === "auth_invalid";
  const [open, setOpen] = useState(isBroken);
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
      setSaveError(
        err instanceof DaemonError ? err.message : "Couldn't save. Is the daemon running?",
      );
    } finally {
      setSaving(null);
    }
  }

  const enabledCount = voices.filter((v) => v.enabled).length;

  return (
    <div
      className={cn(
        "rounded-lg border border-border transition-colors hover:border-foreground/20",
        UI_LINEAGE_BRAND.openrouter.gradient,
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
            UI_LINEAGE_BRAND.openrouter.dot,
          )}
        />
        <div className="min-w-0 flex-1">
          <div className="text-sm font-medium">OpenRouter</div>
          <div className="mt-0.5 flex items-center gap-2">
            {renderStatusBadge(health?.status ?? "healthy")}
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
        <div className="space-y-2 border-t border-border bg-card/50 p-3">
          {isBroken && (
            <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 p-2.5">
              <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-destructive" />
              <div className="min-w-0 flex-1 space-y-1.5">
                <p className="text-[11px] leading-relaxed text-destructive">
                  OpenRouter rejected the saved API key
                  {health?.message ? `: ${health.message}` : "."}
                </p>
                <Link
                  href="/connect"
                  className="inline-flex items-center gap-1 rounded border border-destructive/50 bg-destructive/10 px-2 py-1 text-[11px] font-medium text-destructive transition hover:bg-destructive/20"
                >
                  Fix on Connect
                  <ArrowRight className="h-3 w-3" />
                </Link>
              </div>
            </div>
          )}
          {saveError && <p className="text-[11px] text-destructive">{saveError}</p>}
          {voices.length === 0 ? (
            <p className="text-[11px] text-muted-foreground">
              No OpenRouter voices yet. Add some on the Connect page.
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-1">
              {voices.map((v) => (
                <button
                  key={v.id}
                  type="button"
                  disabled={saving === v.id}
                  onClick={() => toggleVoice(v)}
                  title={v.model_id}
                  className={cn(
                    "flex items-center gap-2 rounded border px-2 py-1.5 text-left text-[11px] transition disabled:opacity-60",
                    v.enabled
                      ? "border-primary/50 bg-primary/10 text-foreground"
                      : "border-border bg-card text-muted-foreground hover:border-muted-foreground/30",
                  )}
                >
                  <div
                    className={cn(
                      "grid h-3 w-3 shrink-0 place-items-center rounded-sm border transition",
                      v.enabled
                        ? "border-primary bg-primary text-primary-foreground"
                        : "border-border",
                    )}
                  >
                    {v.enabled && <Check className="h-2 w-2" />}
                  </div>
                  <span className="truncate font-mono">{v.model_id}</span>
                </button>
              ))}
            </div>
          )}
          {!isBroken && (
            <p className="text-[11px] leading-relaxed text-muted-foreground/70">
              Add or replace your API key on Connect. Toggles save automatically.
            </p>
          )}
        </div>
      )}
    </div>
  );
}

function renderStatusBadge(status: HealthStatus): React.ReactNode {
  switch (status) {
    case "quota_exhausted":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-medium text-destructive">
          <AlertTriangle className="h-3 w-3" />
          Credits out
        </span>
      );
    case "auth_invalid":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-medium text-destructive">
          <AlertTriangle className="h-3 w-3" />
          Auth broken
        </span>
      );
    case "rate_limited":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-300">
          <Clock className="h-3 w-3" />
          Rate-limited
        </span>
      );
    case "healthy":
    case "unknown":
    default:
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-medium text-emerald-300">
          <CheckCircle2 className="h-3 w-3" />
          Configured
        </span>
      );
  }
}
