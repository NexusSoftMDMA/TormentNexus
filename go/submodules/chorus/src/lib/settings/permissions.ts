/**
 * Typed accessor for chorus's permission/sandbox settings.
 *
 * Stored in the `settings` SQLite table as JSON-encoded values keyed by:
 *   - sandbox_profile: 'strict' | 'workspace' | 'full'
 *   - auto_approve_prompts: boolean
 *   - network_access: boolean (overlay on top of sandbox_profile)
 *
 * Defaults are conservative:
 *   - workspace: read+write inside chat dir, no network. Reviewers can't
 *     escape the chat scope or call out to the public internet.
 *   - auto_approve_prompts: true. We pre-approve everything we can so that
 *     headless reviewer spawns don't hang on UI prompts. Users on a shared
 *     box can flip this to false.
 *   - network_access: false unless template/transport specifically needs it.
 *
 * The runner consults these before each spawn and forwards them as
 * AgentSpawnOptions to the shim. Each shim translates the abstract setting
 * into the appropriate CLI flag (e.g. codex `-c sandbox_mode=read-only`,
 * kimi `--afk`, gemini `--approval-mode auto_edit`).
 */

import { settings } from '../db';
import { z } from 'zod';

export type SandboxProfile = 'strict' | 'workspace' | 'full';

const SandboxProfileSchema = z.enum(['strict', 'workspace', 'full']);

export interface PermissionSettings {
  sandboxProfile: SandboxProfile;
  autoApprovePrompts: boolean;
  networkAccess: boolean;
}

export const DEFAULT_PERMISSIONS: PermissionSettings = {
  sandboxProfile: 'workspace',
  autoApprovePrompts: true,
  networkAccess: false,
};

const SANDBOX_KEY = 'sandbox_profile';
const AUTO_APPROVE_KEY = 'auto_approve_prompts';
const NETWORK_KEY = 'network_access';

export async function getPermissions(): Promise<PermissionSettings> {
  const [sandboxRaw, autoApproveRaw, networkRaw] = await Promise.all([
    settings.get(SANDBOX_KEY),
    settings.get(AUTO_APPROVE_KEY),
    settings.get(NETWORK_KEY),
  ]);
  const sandboxProfile = SandboxProfileSchema.safeParse(sandboxRaw);

  return {
    sandboxProfile: sandboxProfile.success ? sandboxProfile.data : DEFAULT_PERMISSIONS.sandboxProfile,
    autoApprovePrompts:
      typeof autoApproveRaw === 'boolean' ? autoApproveRaw : DEFAULT_PERMISSIONS.autoApprovePrompts,
    networkAccess:
      typeof networkRaw === 'boolean' ? networkRaw : DEFAULT_PERMISSIONS.networkAccess,
  };
}

export async function setPermissions(input: Partial<PermissionSettings>): Promise<PermissionSettings> {
  if (input.sandboxProfile !== undefined) {
    SandboxProfileSchema.parse(input.sandboxProfile);
    await settings.set(SANDBOX_KEY, input.sandboxProfile);
  }
  if (input.autoApprovePrompts !== undefined) {
    await settings.set(AUTO_APPROVE_KEY, input.autoApprovePrompts);
  }
  if (input.networkAccess !== undefined) {
    await settings.set(NETWORK_KEY, input.networkAccess);
  }
  return getPermissions();
}

/**
 * Human-readable description of each profile, used by onboarding + /settings UI.
 */
export const PROFILE_DESCRIPTIONS: Record<SandboxProfile, { label: string; description: string }> = {
  strict: {
    label: 'Strict',
    description:
      'Reviewers can only read code. No file writes, no shell exec, no network. Safest, but limits what reviewers can do (e.g. can\'t run the test suite they\'re reviewing).',
  },
  workspace: {
    label: 'Workspace (recommended)',
    description:
      'Reviewers can read+write inside the chat dir and run scoped shell commands. No network. Default — fits most teams.',
  },
  full: {
    label: 'Full access',
    description:
      'Reviewers run with no sandbox: write anywhere, run any command, full network. Only enable on a personal machine you trust.',
  },
};
