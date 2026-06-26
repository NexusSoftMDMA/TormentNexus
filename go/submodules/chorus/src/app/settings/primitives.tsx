"use client";

import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Check, Info } from "lucide-react";

export type AutoApprove = "auto" | "ask" | "block";
export type Role = "driver" | "reviewer";
export type PrivacyTier = "local" | "proxied" | "cloud";

export interface ToolDef {
  id: string;
  name: string;
  description: string;
}

export const TOOLS: ToolDef[] = [
  { id: "read", name: "Read files", description: "cat, less, head, file open" },
  { id: "list", name: "List & search", description: "ls, grep, find, ripgrep" },
  { id: "write", name: "Write files", description: "Edit, Write inside allowed dirs" },
  { id: "exec", name: "Run commands", description: "Bash, npm, pnpm, python, go" },
  { id: "net", name: "Network access", description: "curl, fetch, package install" },
  { id: "outside-cwd", name: "Writes outside working dir", description: "Anything outside the allowed paths below" },
];

export const DEFAULT_DRIVER: Record<string, AutoApprove> = {
  read: "auto",
  list: "auto",
  write: "ask",
  exec: "ask",
  net: "ask",
  "outside-cwd": "block",
};

// Reviewers should never write code — defaults reflect that.
export const DEFAULT_REVIEWER: Record<string, AutoApprove> = {
  read: "auto",
  list: "auto",
  write: "block",
  exec: "ask",
  net: "ask",
  "outside-cwd": "block",
};

const POLICY_STYLES: Record<AutoApprove, string> = {
  auto: "bg-emerald-500/15 text-emerald-300 border-emerald-500/30",
  ask: "bg-amber-500/15 text-amber-300 border-amber-500/30",
  block: "bg-rose-500/15 text-rose-300 border-rose-500/30",
};

export function cyclePolicy(p: AutoApprove): AutoApprove {
  return p === "auto" ? "ask" : p === "ask" ? "block" : "auto";
}

export function Section({
  id,
  icon,
  title,
  subtitle,
  children,
}: {
  /** Optional anchor id so docs / CLI hints can deep-link via /settings#<id>. */
  id?: string;
  icon: React.ReactNode;
  title: string;
  subtitle: string;
  children: React.ReactNode;
}) {
  return (
    <Card id={id} className="mt-6 scroll-mt-20 bg-card p-5">
      <div className="mb-4 flex items-start gap-3">
        <span className="rounded-md border border-border bg-card/60 p-1.5 text-foreground/70">
          {icon}
        </span>
        <div className="flex-1">
          <h2 className="text-sm font-medium text-foreground">{title}</h2>
          <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>
        </div>
      </div>
      {children}
    </Card>
  );
}

export function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="mb-1.5 block text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  );
}

export function PolicyChip({
  policy,
  onCycle,
  role,
}: {
  policy: AutoApprove;
  onCycle: () => void;
  role: Role;
}) {
  return (
    <button
      type="button"
      onClick={onCycle}
      className={`justify-self-end rounded-md border px-2.5 py-1 font-mono text-[10px] uppercase tracking-wider transition hover:scale-105 ${POLICY_STYLES[policy]}`}
      title={`${role}: cycle policy`}
    >
      {policy}
    </button>
  );
}

export function PrivacyTierOption({
  kind,
  current,
  onSelect,
  icon,
  title,
  tagline,
  features,
  tradeoff,
}: {
  kind: PrivacyTier;
  current: PrivacyTier;
  onSelect: () => void;
  icon: React.ReactNode;
  title: string;
  tagline: string;
  features: string[];
  tradeoff: string;
}) {
  const isSelected = kind === current;
  return (
    <button
      type="button"
      onClick={onSelect}
      className={`flex w-full items-start gap-3 rounded-lg border p-3.5 text-left transition ${
        isSelected
          ? "border-primary/40 bg-primary/5 ring-2 ring-primary/30 ring-offset-2 ring-offset-background"
          : "border-border bg-card/40 hover:border-foreground/30"
      }`}
    >
      <span
        className={`mt-0.5 grid h-7 w-7 shrink-0 place-items-center rounded-md ${
          isSelected ? "bg-primary/20 text-primary" : "bg-card text-muted-foreground"
        }`}
      >
        {icon}
      </span>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-foreground">{title}</span>
          <Badge
            variant="outline"
            className="border-border font-mono text-[10px]"
          >
            {tagline}
          </Badge>
        </div>
        <ul className="mt-1.5 space-y-0.5 text-[11px] text-muted-foreground">
          {features.map((f) => (
            <li key={f} className="flex items-start gap-1.5">
              <Check className="mt-0.5 h-3 w-3 shrink-0 text-emerald-400/60" />
              <span>{f}</span>
            </li>
          ))}
        </ul>
        <div className="mt-2 flex items-start gap-1.5 text-[10px] text-muted-foreground">
          <Info className="mt-0.5 h-3 w-3 shrink-0 text-muted-foreground/60" />
          <span>
            <span className="font-medium text-foreground/80">Trade-off:</span>{" "}
            {tradeoff}
          </span>
        </div>
      </div>
      {isSelected && <Check className="mt-1 h-4 w-4 shrink-0 text-primary" />}
    </button>
  );
}
