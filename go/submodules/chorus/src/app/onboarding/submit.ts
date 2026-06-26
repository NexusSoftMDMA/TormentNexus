import {
  DaemonError,
  updateSettings,
  upsertSecret,
} from "@/lib/api";
import type { OpencodeModelsResult } from "@/lib/api/orchestrators";
import {
  updatePermissions,
  type SandboxProfile,
} from "@/lib/api/settings";
import { CLIS, classifyOpencodeClient } from "./helpers";

export interface OnboardingSubmitArgs {
  selectedClis: Set<string>;
  apiKeys: Record<string, string>;
  sandboxProfile: SandboxProfile;
  autoApprovePrompts: boolean;
  networkAccess: boolean;
  opencodeModels: OpencodeModelsResult | null;
  selectedOpencodeModels: Set<string>;
}

/**
 * Persist the onboarding form. Throws DaemonError when the daemon
 * rejects a write; the caller surfaces the message and keeps the form.
 */
export async function submitOnboarding(args: OnboardingSubmitArgs): Promise<void> {
  for (const cliId of args.selectedClis) {
    const cli = CLIS.find((c) => c.id === cliId);
    if (!cli) continue;
    await upsertSecret(cli.provider, {
      kind: "cli_subscription",
      value: cli.id,
      updatedAt: Date.now(),
    });
  }

  for (const [provider, value] of Object.entries(args.apiKeys)) {
    const trimmed = value.trim();
    if (!trimmed) continue;
    await upsertSecret(provider, {
      kind: "api_key",
      value: trimmed,
      updatedAt: Date.now(),
    });
  }

  await updatePermissions({
    sandboxProfile: args.sandboxProfile,
    autoApprovePrompts: args.autoApprovePrompts,
    networkAccess: args.networkAccess,
  });

  // Persist OpenCode model picks (if any) by upserting voice rows. We
  // can't assume the daemon's Phase 2 background seed has completed —
  // three round-1 reviewers caught this race: PUT-only would silently
  // drop the user's selection if the row didn't exist yet. List current
  // voices, PUT existing rows, POST missing ones.
  if (args.selectedClis.has("opencode-cli") && args.opencodeModels) {
    await persistOpencodePicks(args);
  }

  await updateSettings({ onboarded: true });
}

async function persistOpencodePicks(args: OnboardingSubmitArgs): Promise<void> {
  const { listVoices, updateVoice, createVoice } = await import(
    "@/lib/api/voices"
  );
  const existing = new Map(
    (await listVoices({ provider: "opencode-cli" }).catch(() => [])).map(
      (v) => [v.id, v] as const,
    ),
  );

  const models = args.opencodeModels;
  if (!models) return;

  await Promise.all(
    models.flat.map(async (m) => {
      const id = `opencode-cli:${m}`;
      const wantEnabled = args.selectedOpencodeModels.has(m);
      const row = existing.get(id);
      try {
        if (row) {
          if (row.enabled !== wantEnabled) {
            await updateVoice(id, { enabled: wantEnabled });
          }
        } else {
          // Row doesn't exist yet (Phase 2 seed missed it) — create
          // directly.
          const { lineage, vendor_family } = classifyOpencodeClient(m);
          await createVoice({
            provider: "opencode-cli",
            model_id: m,
            label: m,
            source: "cli",
            lineage,
            vendor_family,
            enabled: wantEnabled,
          });
        }
      } catch {
        // Best-effort: a single failed write shouldn't block onboarding.
      }
    }),
  );
}

export function describeError(err: unknown): string {
  return err instanceof DaemonError
    ? err.message
    : "Could not save. Is the Chorus daemon running?";
}
