/**
 * Typed accessor for chorus's transport setting — how shims invoke the
 * underlying CLI. Two modes:
 *
 *   - 'headless' (default): spawn the CLI in `--print` / `exec` / `run` mode,
 *     pipe prompt to stdin or argv, parse stream-json from stdout. No tmux,
 *     no TUI process resident, ~80% lower RAM, no permission prompts (the
 *     CLIs auto-approve in headless mode).
 *
 *   - 'tmux' (opt-in): spawn the CLI in its interactive TUI inside a tmux
 *     session, send the prompt via send-keys, watch answer.md for completion.
 *     Higher RAM, fragile pane-scraping, but humans can attach to the tmux
 *     session for visual debugging — useful if you want to watch the AI
 *     work step-by-step.
 *
 * Settings UI: cockpit /settings exposes this as "Show live terminal sessions"
 * (off=headless, on=tmux). No deprecation timeline — both modes are first-
 * class.
 *
 * The runner reads this once per chat (not per phase) so flipping the
 * setting mid-run is intentionally inert.
 */

import { settings } from '../db';
import { z } from 'zod';
import { platform } from 'os';
import { spawnSync } from 'child_process';

export type Transport = 'headless' | 'tmux';

const TransportSchema = z.enum(['headless', 'tmux']);

const TRANSPORT_KEY = 'transport';

/**
 * tmux availability check.
 *
 *   1. Windows has no tmux equivalent — false unconditionally.
 *   2. On Unix, probe `tmux -V` once at module load. If the binary isn't on
 *      PATH, return false so `setTransport('tmux')` can refuse with an
 *      actionable install hint instead of silently letting the user opt into
 *      a mode whose first chat hangs.
 *
 * Cached at module load — tmux is rarely installed/uninstalled mid-process.
 * Restart the daemon after `apt install tmux` to pick up the new state.
 */
function detectTmuxBinary(): boolean {
  if (platform() === 'win32') return false;
  const result = spawnSync('tmux', ['-V'], { stdio: 'ignore' });
  return result.status === 0;
}

export const TMUX_AVAILABLE: boolean = detectTmuxBinary();

/**
 * v0.5 default is 'headless' — the migration target. Existing users who had
 * tmux working will see the same outputs in the cockpit; only the spawn
 * mechanism differs. To revert, flip the toggle in /settings or set the env
 * override `CHORUS_TRANSPORT=tmux` (env wins over settings — useful for CI
 * and per-user overrides without DB writes).
 */
export const DEFAULT_TRANSPORT: Transport = 'headless';

// Module-level dedup so we warn once per process, not once per chat. The
// daemon reads getTransport() per-chat, and a stuck typo in CHORUS_TRANSPORT
// would otherwise spam stderr for every new chat.
let envWarnFired = false;

export async function getTransport(): Promise<Transport> {
  // Env override takes precedence — operator escape hatch.
  const envOverride = process.env.CHORUS_TRANSPORT;
  let resolved: Transport;
  if (envOverride === 'headless' || envOverride === 'tmux') {
    resolved = envOverride;
  } else {
    if (envOverride !== undefined && envOverride !== '' && !envWarnFired) {
      // Operator-typo aid: silent fallback used to mean a misspelled
      // CHORUS_TRANSPORT (e.g. "headles") just inherited DB/default with no
      // signal. Surface once on stderr; one line, no stack, no rotation.
      console.warn(
        `[chorus] CHORUS_TRANSPORT=${JSON.stringify(envOverride)} is not 'headless' or 'tmux' — ignoring env, falling back to settings/default.`,
      );
      envWarnFired = true;
    }
    const raw = await settings.get(TRANSPORT_KEY);
    const parsed = TransportSchema.safeParse(raw);
    resolved = parsed.success ? parsed.data : DEFAULT_TRANSPORT;
  }
  // Cross-platform safety net — see TMUX_AVAILABLE doc.
  if (resolved === 'tmux' && !TMUX_AVAILABLE) return 'headless';
  return resolved;
}

export async function setTransport(value: Transport): Promise<Transport> {
  TransportSchema.parse(value);
  if (value === 'tmux' && !TMUX_AVAILABLE) {
    if (platform() === 'win32') {
      throw new Error(
        'tmux transport is not available on Windows — headless transport works for everything tmux does plus more.',
      );
    }
    throw new Error(
      'tmux is not installed on this host. Install it first (macOS: `brew install tmux` · Ubuntu/Debian: `sudo apt install tmux` · Fedora/RHEL: `sudo dnf install tmux`), then restart the chorus daemon and try again.',
    );
  }
  await settings.set(TRANSPORT_KEY, value);
  return getTransport();
}

/**
 * Human-readable description for the /settings UI toggle.
 */
export const TRANSPORT_DESCRIPTIONS: Record<Transport, { label: string; description: string }> = {
  headless: {
    label: 'Headless (faster, default)',
    description:
      'Each CLI runs as a one-shot subprocess. ~80% less RAM, faster cold start, no permission dialogs. Recommended for everyday use.',
  },
  tmux: {
    label: 'Tmux — attach & take over',
    description:
      'Each CLI runs in a persistent tmux session you can attach to (`tmux attach -t <name>`) to watch step-by-step or take over mid-run. Uses more RAM. Requires tmux installed on the host (brew/apt/dnf).',
  },
};
