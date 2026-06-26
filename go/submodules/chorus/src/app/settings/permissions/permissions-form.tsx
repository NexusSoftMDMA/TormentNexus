"use client";

import { useCallback, useState } from "react";
import { Loader2, Check, AlertTriangle, ShieldCheck, FolderLock, Globe } from "lucide-react";
import {
  type PermissionSettings,
  type SandboxProfile,
  updatePermissions,
} from "@/lib/api/settings";
import { DaemonError } from "@/lib/api";

interface Props {
  initial: PermissionSettings;
}

const PROFILE_ORDER: SandboxProfile[] = ["strict", "workspace", "full"];

const PROFILE_ICON: Record<SandboxProfile, typeof ShieldCheck> = {
  strict: ShieldCheck,
  workspace: FolderLock,
  full: Globe,
};

export function PermissionsForm({ initial }: Props) {
  const [state, setState] = useState<PermissionSettings>(initial);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [savedAt, setSavedAt] = useState<number | null>(null);

  const descriptions: Record<SandboxProfile, { label: string; description: string }> =
    state.profileDescriptions ?? initial.profileDescriptions ?? {
      strict: { label: "Strict", description: "" },
      workspace: { label: "Workspace", description: "" },
      full: { label: "Full access", description: "" },
    };

  // useCallback so the React Compiler treats the body as deferred-
  // execution (event-handler time, not render time). Without this,
  // `Date.now()` is flagged as an impure call during render even
  // though it only fires when the user toggles a switch.
  const persist = useCallback(async (patch: Partial<PermissionSettings>) => {
    setBusy(true);
    setError(null);
    try {
      const next = await updatePermissions(patch);
      setState(next);
      setSavedAt(Date.now());
    } catch (err) {
      setError(
        err instanceof DaemonError
          ? err.message
          : "Failed to save — is the daemon running?",
      );
    } finally {
      setBusy(false);
    }
  }, []);

  return (
    <div className="space-y-6">
      <section className="space-y-3">
        <h2 className="text-sm font-semibold">Sandbox profile</h2>
        <div className="grid gap-3">
          {PROFILE_ORDER.map((profile) => {
            const Icon = PROFILE_ICON[profile];
            const meta = descriptions[profile] ?? {
              label: profile,
              description: "",
            };
            const selected = state.sandboxProfile === profile;
            return (
              <button
                key={profile}
                type="button"
                disabled={busy}
                onClick={() => persist({ sandboxProfile: profile })}
                className={`group flex items-start gap-3 rounded-lg border p-4 text-left transition disabled:cursor-not-allowed ${
                  selected
                    ? "border-primary bg-primary/5"
                    : "border-border bg-card hover:border-primary/50"
                }`}
              >
                <Icon
                  className={`mt-0.5 h-5 w-5 shrink-0 ${
                    selected ? "text-primary" : "text-muted-foreground"
                  }`}
                />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{meta.label}</span>
                    {selected && (
                      <span className="inline-flex items-center gap-1 rounded-full bg-emerald-500/10 px-2 py-0.5 text-[10px] font-medium text-emerald-300">
                        <Check className="h-3 w-3" /> Active
                      </span>
                    )}
                  </div>
                  <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                    {meta.description}
                  </p>
                </div>
              </button>
            );
          })}
        </div>
      </section>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold">Auto-approve prompts</h2>
        <label className="flex cursor-pointer items-start gap-3 rounded-lg border border-border bg-card p-4">
          <input
            type="checkbox"
            checked={state.autoApprovePrompts}
            disabled={busy}
            onChange={(e) => persist({ autoApprovePrompts: e.target.checked })}
            className="mt-1 h-4 w-4 cursor-pointer accent-primary"
          />
          <div className="min-w-0 flex-1">
            <div className="text-sm font-medium">
              Skip permission prompts inside the spawned CLI
            </div>
            <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
              When on, chorus passes <code className="rounded bg-muted px-1">--afk</code> to
              kimi, <code className="rounded bg-muted px-1">--approval-mode auto_edit</code> to
              gemini, and equivalent flags to other CLIs so reviewers don&apos;t hang on per-tool
              permission prompts. Turn off if you want every action to require an explicit
              &ldquo;yes&rdquo; inside the CLI&apos;s TUI (slower, but fully observable).
            </p>
          </div>
        </label>
      </section>

      <section className="space-y-3">
        <h2 className="text-sm font-semibold">Network access</h2>
        <label className="flex cursor-pointer items-start gap-3 rounded-lg border border-border bg-card p-4">
          <input
            type="checkbox"
            checked={state.networkAccess}
            disabled={busy}
            onChange={(e) => persist({ networkAccess: e.target.checked })}
            className="mt-1 h-4 w-4 cursor-pointer accent-primary"
          />
          <div className="min-w-0 flex-1">
            <div className="text-sm font-medium">Allow outbound network from reviewers</div>
            <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
              Off by default. Reviewers can&apos;t curl, npm install, or call external APIs.
              Templates that explicitly need network (e.g. fetching docs, gh CLI) override this
              per phase. Enable globally only if you trust every template you run.
            </p>
          </div>
        </label>
      </section>

      <div className="flex h-6 items-center text-xs">
        {busy && (
          <span className="inline-flex items-center gap-1 text-muted-foreground">
            <Loader2 className="h-3 w-3 animate-spin" /> Saving…
          </span>
        )}
        {!busy && error && (
          <span className="inline-flex items-center gap-1 text-destructive">
            <AlertTriangle className="h-3 w-3" /> {error}
          </span>
        )}
        {!busy && !error && savedAt && (
          <span className="inline-flex items-center gap-1 text-emerald-400">
            <Check className="h-3 w-3" /> Saved
          </span>
        )}
      </div>
    </div>
  );
}
