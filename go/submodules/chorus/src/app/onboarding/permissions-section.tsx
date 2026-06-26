"use client";

import { Check } from "lucide-react";
import type { SandboxProfile } from "@/lib/api/settings";
import { cn } from "@/lib/utils";

const PROFILES = [
  {
    id: "strict",
    label: "Strict",
    hint: "Read-only. Reviewers can inspect code but can't write files, exec shell, or hit the network.",
  },
  {
    id: "workspace",
    label: "Workspace (recommended)",
    hint: "Read+write inside the chat dir, scoped shell, no network. Default for most teams.",
  },
  {
    id: "full",
    label: "Full access",
    hint: "No sandbox at all. Only on a personal machine you fully trust.",
  },
] as const;

interface PermissionsSectionProps {
  sandboxProfile: SandboxProfile;
  setSandboxProfile: (p: SandboxProfile) => void;
  autoApprovePrompts: boolean;
  setAutoApprovePrompts: (v: boolean) => void;
  networkAccess: boolean;
  setNetworkAccess: (v: boolean) => void;
}

export function PermissionsSection({
  sandboxProfile,
  setSandboxProfile,
  autoApprovePrompts,
  setAutoApprovePrompts,
  networkAccess,
  setNetworkAccess,
}: PermissionsSectionProps) {
  return (
    <section className="mb-8">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Permissions &amp; sandbox
      </h2>
      <p className="mb-4 text-xs leading-relaxed text-muted-foreground">
        Controls what reviewers can do on your machine. You can change this
        anytime in Settings &rarr; Permissions.
      </p>

      <div className="space-y-2">
        {PROFILES.map((p) => {
          const checked = sandboxProfile === p.id;
          return (
            <button
              key={p.id}
              type="button"
              onClick={() => setSandboxProfile(p.id)}
              className={cn(
                "flex w-full items-start gap-3 rounded-lg border p-4 text-left transition",
                checked
                  ? "border-primary/50 bg-primary/10"
                  : "border-border bg-card hover:border-muted-foreground/30",
              )}
            >
              <div
                className={cn(
                  "mt-0.5 grid h-5 w-5 shrink-0 place-items-center rounded-full border transition",
                  checked
                    ? "border-primary bg-primary text-primary-foreground"
                    : "border-border",
                )}
              >
                {checked && <Check className="h-3 w-3" />}
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-sm font-medium">{p.label}</div>
                <div className="mt-1 text-xs leading-relaxed text-muted-foreground">
                  {p.hint}
                </div>
              </div>
            </button>
          );
        })}
      </div>

      <div className="mt-3 space-y-2">
        <label className="flex cursor-pointer items-start gap-3 rounded-lg border border-border bg-card p-3">
          <input
            type="checkbox"
            checked={autoApprovePrompts}
            onChange={(e) => setAutoApprovePrompts(e.target.checked)}
            className="mt-1 h-4 w-4 cursor-pointer accent-primary"
          />
          <div className="min-w-0 flex-1 text-xs leading-relaxed">
            <div className="text-sm font-medium">
              Skip in-CLI permission prompts
            </div>
            <div className="mt-0.5 text-muted-foreground">
              Passes <code className="rounded bg-muted px-1">--afk</code> /{" "}
              <code className="rounded bg-muted px-1">auto_edit</code> to
              spawned reviewers so they don&apos;t hang on per-tool prompts. Off
              = every action requires explicit consent in the CLI&apos;s TUI.
            </div>
          </div>
        </label>

        <label className="flex cursor-pointer items-start gap-3 rounded-lg border border-border bg-card p-3">
          <input
            type="checkbox"
            checked={networkAccess}
            onChange={(e) => setNetworkAccess(e.target.checked)}
            className="mt-1 h-4 w-4 cursor-pointer accent-primary"
          />
          <div className="min-w-0 flex-1 text-xs leading-relaxed">
            <div className="text-sm font-medium">
              Allow outbound network from reviewers
            </div>
            <div className="mt-0.5 text-muted-foreground">
              Off by default. Templates that explicitly need network override
              per phase.
            </div>
          </div>
        </label>
      </div>
    </section>
  );
}
