"use client";

import {
  AlertTriangle,
  Check,
  Cloud,
  FolderLock,
  Loader2,
  Monitor,
  Plus,
  Send,
  Server,
  Shield,
  Sliders,
  Webhook,
  Workflow,
  X,
} from "lucide-react";
import {
  Field,
  PolicyChip,
  PrivacyTierOption,
  Section,
  TOOLS,
  type AutoApprove,
  type PrivacyTier,
  type Role,
} from "./primitives";

export interface PreviewSectionsProps {
  maxConcurrent: number;
  setMaxConcurrent: (n: number) => void;
  driverPolicies: Record<string, AutoApprove>;
  reviewerPolicies: Record<string, AutoApprove>;
  cyclePolicy: (role: Role, toolId: string) => void;
  allowedDirs: string[];
  setAllowedDirs: (v: string[]) => void;
  newDir: string;
  setNewDir: (v: string) => void;
  addDir: () => void;
  defaultDriver: string;
  setDefaultDriver: (v: string) => void;
  defaultCostCap: string;
  setDefaultCostCap: (v: string) => void;
  defaultThreshold: string;
  setDefaultThreshold: (v: string) => void;
  webhookUrl: string;
  setWebhookUrl: (v: string) => void;
  mcpEnabled: boolean;
  setMcpEnabled: (v: boolean) => void;
  privacyTier: PrivacyTier;
  setPrivacyTier: (v: PrivacyTier) => void;
  webhookTestState: "idle" | "running" | "ok" | "fail";
  testWebhook: () => void;
}

/**
 * v0.8 preview sections. Wrapped in a `pointer-events-none + opacity-60`
 * group so clicks/keystrokes don't escape; the YAML view exposes the
 * planned schema for users who want to peek.
 */
export function PreviewSections(p: PreviewSectionsProps) {
  return (
    <div
      className="mt-2 select-none opacity-60 [&_*]:pointer-events-none"
      aria-disabled="true"
    >
      <Section
        icon={<FolderLock className="h-4 w-4" />}
        title="Per-tool auto-approve"
        subtitle="Drivers write code; reviewers don't. Defaults reflect that. Click any chip to cycle auto / ask / block."
      >
        <div className="grid grid-cols-[1fr_120px_120px] gap-3 border-b border-border pb-2 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          <span>Tool</span>
          <span className="flex items-center justify-end gap-1.5">
            <span className="h-1.5 w-1.5 rounded-full bg-amber-400" />
            Driver
          </span>
          <span className="flex items-center justify-end gap-1.5">
            <span className="h-1.5 w-1.5 rounded-full bg-blue-400" />
            Reviewer
          </span>
        </div>

        <div className="mt-1 divide-y divide-border opacity-70">
          {TOOLS.map((tool) => (
            <div
              key={tool.id}
              className="grid grid-cols-[1fr_120px_120px] items-center gap-3 py-2.5"
            >
              <div className="flex flex-col">
                <span className="text-sm text-foreground">{tool.name}</span>
                <span className="text-[11px] text-muted-foreground">
                  {tool.description}
                </span>
              </div>
              <PolicyChip
                policy={p.driverPolicies[tool.id]}
                onCycle={() => p.cyclePolicy("driver", tool.id)}
                role="driver"
              />
              <PolicyChip
                policy={p.reviewerPolicies[tool.id]}
                onCycle={() => p.cyclePolicy("reviewer", tool.id)}
                role="reviewer"
              />
            </div>
          ))}
        </div>

        <p className="mt-3 flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-xs text-amber-100">
          <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-amber-400" />
          <span>
            <span className="font-semibold">When wired:</span>{" "}
            <code className="rounded bg-amber-400/10 px-1 font-mono text-amber-200">
              ask
            </code>{" "}
            prompts will appear as cards in the chat — you click{" "}
            <span className="font-semibold">Approve</span> /{" "}
            <span className="font-semibold">Deny</span> instead of the agent
            stalling.
          </span>
        </p>
      </Section>

      <Section
        icon={<FolderLock className="h-4 w-4" />}
        title="Working directories"
        subtitle="Each tmux session is started inside one of these paths. Writes outside are blocked."
      >
        <div className="flex flex-wrap gap-2">
          {p.allowedDirs.map((d) => (
            <span
              key={d}
              className="group inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2.5 py-1 font-mono text-[11px] text-foreground"
            >
              {d}
              <button
                type="button"
                onClick={() =>
                  p.setAllowedDirs(p.allowedDirs.filter((x) => x !== d))
                }
                className="text-muted-foreground/50 transition hover:text-rose-400"
                aria-label={`Remove ${d}`}
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
        <div className="mt-3 flex gap-2">
          <input
            type="text"
            value={p.newDir}
            onChange={(e) => p.setNewDir(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && p.addDir()}
            placeholder="~/code/another-project"
            className="flex-1 rounded-md border border-border bg-card px-3 py-1.5 font-mono text-xs text-foreground placeholder:text-muted-foreground/60 focus:border-foreground/30 focus:outline-none"
          />
          <button
            type="button"
            onClick={p.addDir}
            className="flex items-center gap-1.5 rounded-md bg-primary px-3 py-1.5 text-xs font-medium text-primary-foreground transition hover:bg-primary/90"
          >
            <Plus className="h-3.5 w-3.5" />
            Add path
          </button>
        </div>
      </Section>

      <Section
        icon={<Sliders className="h-4 w-4" />}
        title="Template defaults"
        subtitle="Applied to new templates and one-off chats. Per-template values override these."
      >
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Default driver">
            <select
              value={p.defaultDriver}
              onChange={(e) => p.setDefaultDriver(e.target.value)}
              className="w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-foreground/30 focus:outline-none"
            >
              <option value="claude-code">claude-code</option>
              <option value="cursor">cursor</option>
              <option value="codex-cli">codex-cli</option>
              <option value="windsurf">windsurf</option>
              <option value="external">external (ask each run)</option>
            </select>
          </Field>
          <Field label="Default agreement threshold">
            <select
              value={p.defaultThreshold}
              onChange={(e) => p.setDefaultThreshold(e.target.value)}
              className="w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-foreground/30 focus:outline-none"
            >
              <option value="unanimous">unanimous</option>
              <option value="majority">majority</option>
              <option value="any">any</option>
            </select>
          </Field>
          <Field label="Default cost cap (USD per chat)">
            <div className="flex items-center gap-2">
              <span className="font-mono text-sm text-muted-foreground">$</span>
              <input
                type="number"
                step="0.10"
                min={0}
                value={p.defaultCostCap}
                onChange={(e) => p.setDefaultCostCap(e.target.value)}
                className="w-full rounded-md border border-border bg-card px-3 py-2 font-mono text-sm text-foreground focus:border-foreground/30 focus:outline-none"
              />
            </div>
          </Field>
          <Field label="On model error">
            <select
              defaultValue="fallback"
              className="w-full rounded-md border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-foreground/30 focus:outline-none"
            >
              <option value="fallback">try next fallback model</option>
              <option value="fail">mark chat failed</option>
              <option value="ask-user">surface to me</option>
            </select>
          </Field>
        </div>
      </Section>

      <Section
        icon={<Shield className="h-4 w-4" />}
        title="Privacy & data path"
        subtitle="How prompts and code reach each LLM. Affects what leaves your machine."
      >
        <div className="space-y-2">
          <PrivacyTierOption
            current={p.privacyTier}
            kind="local"
            onSelect={() => p.setPrivacyTier("local")}
            icon={<Monitor className="h-4 w-4" />}
            title="Local-only (BYO CLIs)"
            tagline="Strictest"
            features={[
              "Daemon runs on your machine; CLIs run on your machine.",
              "Prompts go directly from your machine to each provider.",
              "Chorus sees nothing — not even chat IDs.",
            ]}
            tradeoff="No cloud dashboard, no team sharing, no fallback when a CLI is down."
          />
          <PrivacyTierOption
            current={p.privacyTier}
            kind="proxied"
            onSelect={() => p.setPrivacyTier("proxied")}
            icon={<Server className="h-4 w-4" />}
            title="Proxied credits"
            tagline="Recommended for credits"
            features={[
              "Stateless edge proxy: prompts transit, never persist.",
              "Stripe-pattern: same payload-handling discipline.",
              "Used only when you spend credits to try a model you don't have.",
            ]}
            tradeoff="One extra hop between you and the provider."
          />
          <PrivacyTierOption
            current={p.privacyTier}
            kind="cloud"
            onSelect={() => p.setPrivacyTier("cloud")}
            icon={<Cloud className="h-4 w-4" />}
            title="Cloud workspace"
            tagline="Team / multi-machine"
            features={[
              "Code is cloned into Chorus's cloud workspace.",
              "Transcripts + verdicts kept for audit & retro.",
              "Encrypted at rest; SOC 2 in progress.",
            ]}
            tradeoff="Your code lives on our infra. Read the privacy whitepaper before flipping this on."
          />
        </div>
      </Section>

      <Section
        icon={<Webhook className="h-4 w-4" />}
        title="Notifications & MCP"
        subtitle="How outer orchestrators (your main Claude / Cursor session) get pinged when a chat finishes or blocks."
      >
        <div className="space-y-4">
          <button
            type="button"
            onClick={() => p.setMcpEnabled(!p.mcpEnabled)}
            className="flex w-full items-center justify-between rounded-md border border-border bg-card/50 px-3 py-3 text-left transition hover:border-foreground/20"
          >
            <div className="flex items-start gap-3">
              <Workflow className="mt-0.5 h-4 w-4 text-primary" />
              <div className="flex flex-col">
                <span className="text-sm text-foreground">
                  Chorus MCP server
                </span>
                <span className="text-[11px] text-muted-foreground">
                  Exposes <span className="font-mono">mm.create_chat</span>,{" "}
                  <span className="font-mono">mm.update_settings</span>,{" "}
                  <span className="font-mono">mm.create_template</span> and 9
                  other tools. Outer Claude / Cursor / Codex can configure
                  Chorus for you.
                </span>
              </div>
            </div>
            <span
              className={`flex h-5 w-9 items-center rounded-full border p-0.5 transition ${
                p.mcpEnabled
                  ? "border-emerald-500/40 bg-emerald-500/20"
                  : "border-border bg-card"
              }`}
            >
              <span
                className={`h-3.5 w-3.5 rounded-full transition-transform ${
                  p.mcpEnabled
                    ? "translate-x-4 bg-emerald-400"
                    : "bg-muted-foreground/50"
                }`}
              />
            </span>
          </button>

          <Field label="Webhook URL (optional)">
            <div className="flex gap-2">
              <input
                type="url"
                value={p.webhookUrl}
                onChange={(e) => p.setWebhookUrl(e.target.value)}
                placeholder="https://hooks.example.com/chorus"
                className="flex-1 rounded-md border border-border bg-card px-3 py-2 font-mono text-xs text-foreground placeholder:text-muted-foreground/60 focus:border-foreground/30 focus:outline-none"
              />
              <button
                type="button"
                onClick={p.testWebhook}
                disabled={!p.webhookUrl || p.webhookTestState === "running"}
                className={`flex items-center gap-1.5 rounded-md border px-3 py-2 text-xs font-medium transition ${
                  p.webhookTestState === "ok"
                    ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
                    : p.webhookTestState === "fail"
                      ? "border-rose-500/30 bg-rose-500/10 text-rose-300"
                      : "border-border bg-card text-muted-foreground hover:text-foreground"
                } disabled:cursor-not-allowed disabled:opacity-50`}
              >
                {p.webhookTestState === "running" ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : p.webhookTestState === "ok" ? (
                  <Check className="h-3.5 w-3.5" />
                ) : p.webhookTestState === "fail" ? (
                  <AlertTriangle className="h-3.5 w-3.5" />
                ) : (
                  <Send className="h-3.5 w-3.5" />
                )}
                {p.webhookTestState === "ok"
                  ? "204 OK"
                  : p.webhookTestState === "fail"
                    ? "Failed"
                    : p.webhookTestState === "running"
                      ? "Posting…"
                      : "Test"}
              </button>
            </div>
            <p className="mt-1.5 text-[11px] text-muted-foreground">
              Posted on chat finished / blocked / errored. Payload includes
              chat_id, status, synthesis.
            </p>
          </Field>
        </div>
      </Section>
    </div>
  );
}
