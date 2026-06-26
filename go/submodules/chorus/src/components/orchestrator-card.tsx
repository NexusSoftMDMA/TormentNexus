"use client";

import { useState } from "react";
import { Check, Loader2, Plug, AlertTriangle } from "lucide-react";
import {
  connectOrchestrator,
  type OrchestratorStatus,
  type OrchestratorName,
  DaemonError,
} from "@/lib/api";
import { updateVoice, type Voice } from "@/lib/api/voices";
import { UI_LINEAGE_BRAND, type UILineage } from "@/lib/lineage-maps";
import { cn } from "@/lib/utils";

/**
 * One unified card per CLI on the /connect page. Combines MCP wiring
 * status (connect button) and inline voice picker. Header is always
 * visible; body shows inline whenever the CLI is actually usable so
 * users don't have to click-to-reveal model toggles they actually need.
 */
interface Props {
  initial: OrchestratorStatus;
  /** Voices for this orchestrator's provider — both enabled and disabled. */
  voices: Voice[];
}

const ORCHESTRATOR_TO_PROVIDER: Record<string, string> = {
  claude: "claude-code",
  codex: "codex-cli",
  gemini: "gemini-cli",
  opencode: "opencode-cli",
  kimi: "kimi-cli",
};

const NEUTRAL_BRAND = {
  dot: "bg-muted-foreground/40",
  gradient: "bg-gradient-to-br from-muted/30 via-card to-card",
} as const;

function brandForOrchestrator(name: string) {
  if (name in UI_LINEAGE_BRAND) {
    return UI_LINEAGE_BRAND[name as UILineage];
  }
  return NEUTRAL_BRAND;
}

export function OrchestratorCard({ initial, voices: initialVoices }: Props) {
  const [status, setStatus] = useState<OrchestratorStatus>(initial);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [justConnected, setJustConnected] = useState(false);

  const provider = ORCHESTRATOR_TO_PROVIDER[initial.name];
  const [voices, setVoices] = useState<Voice[]>(initialVoices);
  const [saving, setSaving] = useState<string | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);

  const supportsModels = !!provider && voices.length > 0;
  const enabledCount = voices.filter((v) => v.enabled).length;
  const isOpencode = provider === "opencode-cli";
  const isConnected = status.connected;
  const partial = status.approvedTools > 0 && !isConnected;
  const brand = brandForOrchestrator(initial.name);

  // Body always renders so every card has the same height. For
  // Coming-soon CLIs we show a placeholder note so the body isn't empty.

  const connect = async () => {
    setBusy(true);
    setError(null);
    try {
      const result = await connectOrchestrator(status.name as OrchestratorName);
      setStatus(result.status);
      setJustConnected(result.added.length > 0);
    } catch (err) {
      setError(
        err instanceof DaemonError
          ? err.message
          : "Failed to connect — is the daemon running?",
      );
    } finally {
      setBusy(false);
    }
  };

  const toggleVoice = async (v: Voice) => {
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
  };

  // Group OpenCode voices by gateway prefix in model_id.
  const opencodeGroups = new Map<string, Voice[]>();
  if (isOpencode) {
    for (const v of voices) {
      const slash = v.model_id.indexOf("/");
      const gw = slash > 0 ? v.model_id.slice(0, slash) : "other";
      const list = opencodeGroups.get(gw) ?? [];
      list.push(v);
      opencodeGroups.set(gw, list);
    }
  }
  const sortedGateways = Array.from(opencodeGroups.keys()).sort();

  return (
    <div
      className={cn(
        "flex h-72 flex-col rounded-lg border border-border",
        brand.gradient,
      )}
    >
      <div className="flex shrink-0 items-center gap-3 px-4 py-3">
        <div className="flex min-w-0 flex-1 items-center gap-2">
          <span className={cn("h-2 w-2 shrink-0 rounded-full", brand.dot)} />
          <h3 className="whitespace-nowrap text-sm font-semibold">{status.label}</h3>
          {isConnected ? (
            <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-medium text-emerald-300">
              <Check className="h-3 w-3" /> Connected
            </span>
          ) : partial ? (
            <span className="inline-flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-300">
              {status.approvedTools}/{status.totalTools} tools
            </span>
          ) : status.supported ? (
            <span className="inline-flex items-center rounded-full border border-border bg-card px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
              Not connected
            </span>
          ) : (
            <span className="inline-flex items-center rounded-full border border-border bg-card px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
              Coming soon
            </span>
          )}
          {supportsModels && (
            <span className="whitespace-nowrap text-[10px] text-muted-foreground">
              · {enabledCount} model{enabledCount === 1 ? "" : "s"}
            </span>
          )}
        </div>
      </div>

      <div className="flex flex-1 flex-col gap-3 overflow-y-auto border-t border-border bg-card/30 p-4">
        <p className="text-[11px] leading-relaxed text-muted-foreground">
          {status.note}
        </p>

        {!status.supported && (
          <p className="text-[11px] italic text-muted-foreground/70">
            Wiring will land in a future release.
          </p>
        )}

        {status.supported && (
          <div className="flex flex-wrap items-center gap-2">
            <button
              type="button"
              onClick={connect}
              disabled={busy || (isConnected && !partial)}
              className="inline-flex h-8 items-center gap-1.5 rounded-md bg-primary px-3 text-xs font-medium text-primary-foreground transition hover:bg-primary/90 disabled:cursor-not-allowed disabled:bg-muted disabled:text-muted-foreground"
            >
              {busy ? (
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Plug className="h-3.5 w-3.5" />
              )}
              {isConnected
                ? "Already connected"
                : partial
                  ? "Approve remaining tools"
                  : `Connect ${status.label}`}
            </button>
            {justConnected && !error && (
              <p className="text-[11px] text-emerald-400">
                ✓ Done. Restart {status.label} for the change to take effect.
              </p>
            )}
            {error && (
              <p className="flex items-start gap-1 text-[11px] text-destructive">
                <AlertTriangle className="mt-0.5 h-3 w-3 shrink-0" /> {error}
              </p>
            )}
          </div>
        )}

        {isConnected && status.firstCallBehavior === "prompts_once" && (
          <p className="text-[11px] text-amber-300/90">
            ⚠ First chorus.* call will show a one-time prompt — click &quot;Always allow&quot;.
          </p>
        )}

        {supportsModels && (
          <div className="space-y-2">
            <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
              Models
            </p>
            {saveError && <p className="text-[11px] text-destructive">{saveError}</p>}

            {isOpencode ? (
              <div className="space-y-2">
                {sortedGateways.map((gateway) => {
                  const list = opencodeGroups.get(gateway) ?? [];
                  return (
                    <div key={gateway} className="space-y-1">
                      <p className="text-[10px] font-mono text-muted-foreground/80">
                        {gateway}/
                      </p>
                      <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
                        {list.map((v) => (
                          <ModelToggle
                            key={v.id}
                            label={v.model_id.slice(gateway.length + 1)}
                            value={v.model_id}
                            selected={v.enabled}
                            disabled={saving === v.id}
                            onClick={() => toggleVoice(v)}
                          />
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-1 sm:grid-cols-2">
                {voices.map((v) => (
                  <ModelToggle
                    key={v.id}
                    label={v.model_id}
                    value={v.model_id}
                    selected={v.enabled}
                    disabled={saving === v.id}
                    onClick={() => toggleVoice(v)}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

interface ModelToggleProps {
  label: string;
  value: string;
  selected: boolean;
  disabled: boolean;
  onClick: () => void;
}

function ModelToggle({ label, value, selected, disabled, onClick }: ModelToggleProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      title={value}
      className={cn(
        "flex items-center gap-2 rounded border px-2 py-1.5 text-left text-[11px] transition disabled:opacity-60",
        selected
          ? "border-primary/50 bg-primary/10 text-foreground"
          : "border-border bg-card hover:border-muted-foreground/30 text-muted-foreground",
      )}
    >
      <div
        className={cn(
          "grid h-3 w-3 shrink-0 place-items-center rounded-sm border transition",
          selected
            ? "border-primary bg-primary text-primary-foreground"
            : "border-border",
        )}
      >
        {selected && <Check className="h-2 w-2" />}
      </div>
      <span className="truncate font-mono">{label}</span>
    </button>
  );
}
