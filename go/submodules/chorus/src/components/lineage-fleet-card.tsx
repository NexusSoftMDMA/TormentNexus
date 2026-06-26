"use client";

/**
 * Inline-expandable fleet card for single-subscription CLIs (Claude Code,
 * Codex CLI, Gemini CLI, Kimi CLI). Shares the toggle UX of
 * OpencodeFleetCard but skips the gateway grouping — these CLIs back a
 * single subscription with a flat list of models.
 *
 * Data source: voices table via /voices?provider=<id>. Each toggle calls
 * PUT /voices/:id immediately so changes are live without a save button.
 *
 * Shows ALL voices (enabled + disabled) so users can re-enable from the
 * fleet card without going through onboarding.
 */

import { useState } from "react";
import {
  CheckCircle2,
  AlertTriangle,
  Clock,
  ChevronDown,
  Check,
} from "lucide-react";
import { UI_LINEAGE_BRAND, type UILineage } from "@/lib/lineage-maps";
import { updateVoice, type Voice } from "@/lib/api/voices";
import { DaemonError } from "@/lib/api/client";
import { cn } from "@/lib/utils";

const DAEMON_TO_UI_LINEAGE: Record<string, UILineage> = {
  anthropic: "claude",
  openai: "codex",
  google: "gemini",
  opencode: "opencode",
  moonshot: "kimi",
};

const NEUTRAL_BRAND = {
  dot: "bg-muted-foreground/40",
  gradient: "bg-gradient-to-br from-muted/30 via-card to-card",
} as const;

function brandForDaemonLineage(lineage: string) {
  const ui = DAEMON_TO_UI_LINEAGE[lineage];
  return ui ? UI_LINEAGE_BRAND[ui] : NEUTRAL_BRAND;
}

interface LineageFleetCardProps {
  /** Daemon-side lineage name — "anthropic", "openai", "google", "moonshot". */
  lineage: string;
  /** Display label — "Claude Code", "Codex CLI", etc. */
  label: string;
  /** Voices for this provider — both enabled and disabled. */
  voices: Voice[];
  health: {
    status: "healthy" | "quota_exhausted" | "auth_invalid" | "rate_limited" | "unknown";
    message?: string;
  };
}

export function LineageFleetCard({
  lineage,
  label,
  voices: initialVoices,
  health,
}: LineageFleetCardProps) {
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

  const enabledCount = voices.filter((v) => v.enabled).length;
  const brand = brandForDaemonLineage(lineage);

  return (
    <div
      className={cn(
        "rounded-lg border border-border transition-colors hover:border-foreground/20",
        brand.gradient,
      )}
    >
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-3 p-3 text-left"
      >
        <span className={cn("h-2 w-2 shrink-0 rounded-full", brand.dot)} />
        <div className="min-w-0 flex-1">
          <div className="text-sm font-medium">{label}</div>
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
        <div className="space-y-2 border-t border-border bg-card/50 p-3">
          {saveError && <p className="text-[11px] text-destructive">{saveError}</p>}
          {voices.length === 0 ? (
            <p className="text-[11px] text-muted-foreground">
              No voices detected for this provider yet.
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-1">
              {voices.map((v) => {
                const sel = v.enabled;
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
                    <span className="truncate font-mono">{v.model_id}</span>
                  </button>
                );
              })}
            </div>
          )}
          <p className="text-[11px] leading-relaxed text-muted-foreground/70">
            Toggles save automatically. Voice list is curated per chorus release —
            new models appear after upgrades.
          </p>
        </div>
      )}
    </div>
  );
}

function StatusBadge({ status }: { status: LineageFleetCardProps["health"]["status"] }) {
  // Card is only rendered for already-connected CLIs (panel filters on
  // orchestrator.connected), so the baseline state is "Connected" rather
  // than "Untested" — the latter never made sense to surface here. Real
  // failure states (auth/quota/rate-limit) still override and show their
  // amber/red badge.
  switch (status) {
    case "auth_invalid":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-medium text-destructive">
          <AlertTriangle className="h-3 w-3" />
          Auth broken
        </span>
      );
    case "quota_exhausted":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-destructive/10 px-2 py-0.5 text-[10px] font-medium text-destructive">
          <AlertTriangle className="h-3 w-3" />
          Quota out
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
    default:
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-medium text-emerald-300">
          <CheckCircle2 className="h-3 w-3" />
          Connected
        </span>
      );
  }
}
