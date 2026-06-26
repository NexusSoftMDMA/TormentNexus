/**
 * Fail-closed sandbox enforcement for shims that can't translate the
 * 'strict' sandbox profile to a CLI flag.
 *
 * Audit D1 (pre-launch dogfood) found that opencode + kimi shims
 * accepted but ignored `opts.sandbox`. The "Strict" cockpit label
 * became decorative for ~40% of reviewers — a doer running through
 * an opencode-go gateway could write files and execute shell
 * regardless of the visual setting.
 *
 * Long-term fix is to map 'strict' to upstream CLI flags as those
 * land. Until then, refuse to spawn rather than silently misrepresent
 * the sandbox state. The tmux path throws (the runner converts to a
 * cli_error event); the headless path returns a one-shot async
 * iterable that emits the same error event shape.
 */

import type { AgentEvent } from './types.js';

/** Known sandbox profiles a shim might be asked to enforce. */
export type SandboxProfile = 'strict' | 'workspace' | 'full';

/**
 * Throw on `strict`; otherwise no-op. Use from `buildLaunchCommand`
 * where the shim returns a string (tmux path). The runner catches the
 * throw and surfaces it as a cli_error to the cockpit.
 */
export function assertSandboxSupported(
  sandbox: SandboxProfile | undefined,
  cli: string,
): void {
  if (sandbox === 'strict') {
    throw new Error(
      `${cli} shim cannot enforce sandbox=strict (no upstream read-only flag yet). ` +
        `Switch to a lineage that supports strict (claude / codex / gemini), or relax to 'workspace'.`,
    );
  }
}

/**
 * Return a one-shot AsyncIterable<AgentEvent> that emits a
 * `sandbox_unsupported` error when the shim can't enforce strict.
 * Call from `runHeadless` like:
 *
 *   const sb = sandboxFailClosed(opts.sandbox, 'opencode');
 *   if (sb) return sb;
 *
 * Returns null when the requested sandbox is supported (or undefined),
 * letting the caller proceed with the normal spawn path.
 */
export function sandboxFailClosed(
  sandbox: SandboxProfile | undefined,
  cli: string,
): AsyncIterable<AgentEvent> | null {
  if (sandbox !== 'strict') return null;
  const message =
    `${cli} shim cannot enforce sandbox=strict (no upstream read-only flag yet). ` +
    `Switch to a lineage that supports strict (claude / codex / gemini), or relax to 'workspace'.`;
  return {
    [Symbol.asyncIterator](): AsyncIterator<AgentEvent> {
      let yielded = false;
      return {
        async next(): Promise<IteratorResult<AgentEvent>> {
          if (yielded) return { value: undefined, done: true };
          yielded = true;
          return {
            value: {
              type: 'error',
              kind: 'sandbox_unsupported',
              message,
            },
            done: false,
          };
        },
      };
    },
  };
}
