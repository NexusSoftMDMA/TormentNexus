/**
 * Claude Code agent shim.
 * Multi-paragraph prompts are fine; no special formatting needed.
 */

import type {
  AgentShim,
  AgentSpawnOptions,
  AgentNudgeOptions,
  HeadlessSpawnOptions,
  AgentEvent,
} from './types.js';
import { quoteValue, quotePath } from './quote.js';
import { preTrustClaudeWorkspace } from './preflight.js';
import { spawnHeadless } from '../headless.js';
import { parseClaude } from './parsers/index.js';

export const claudeShim: AgentShim = {
  lineage: 'anthropic',
  name: 'claude-code',

  buildLaunchCommand(opts: AgentSpawnOptions): string {
    // Pre-trust the chat dir so Claude Code doesn't stop on its first-launch
    // "Trust this folder?" prompt. Idempotent; safe to call every spawn.
    preTrustClaudeWorkspace(opts.cwd);

    const cwd = quotePath(opts.cwd);
    let cmd = `cd ${cwd} && claude`;

    // Sandbox profile mapping. Claude Code itself doesn't expose a
    // workspace-vs-strict toggle — the existing per-tool permission allow-list
    // (settings.local.json `permissions.allow[]`) already implements the
    // "workspace" profile. Full = bypass everything; strict isn't expressible
    // at spawn time, only via tighter allow-list.
    //
    // Root-aware: --dangerously-skip-permissions is refused under root (same
    // policy as the headless permission-mode bypass). Under root we drop the
    // flag — claude will use its default per-tool prompts which the tmux
    // session lets the user approve interactively.
    const runningAsRoot =
      typeof process.getuid === 'function' && process.getuid() === 0;
    if ((opts.unsandboxed || opts.sandbox === 'full') && !runningAsRoot) {
      cmd += ` --dangerously-skip-permissions`;
    }

    if (opts.model) {
      cmd += ` --model ${quoteValue(opts.model)}`;
    }

    return cmd;
  },

  formatPrompt(opts: AgentNudgeOptions): string {
    const sentinel = opts.expectDoneSentinel
      ? '\n\nWhen finished, end your response with: ## DONE'
      : '';

    return (
      `${opts.task}\n\n` +
      `Read the prompt at: ${opts.promptFile}\n\n` +
      `Write your full answer to: ${opts.answerFile}${sentinel}`
    );
  },

  /**
   * Headless mode (`claude --print --output-format stream-json --verbose`).
   *
   * Pipes the full prompt via stdin, parses Anthropic-shape stream-json events,
   * yields AgentEvents to the runner. No tmux session, no pane scraping, no
   * permission dialogs (bypassPermissions is set when autoApprove is on).
   *
   * Verified format 2026-04-30 against Claude Code 2.1.123. See parsers.ts
   * for the parsed shapes and inline tests.
   */
  runHeadless(opts: HeadlessSpawnOptions): AsyncIterable<AgentEvent> {
    preTrustClaudeWorkspace(opts.cwd);

    const args = ['--print', '--output-format', 'stream-json', '--verbose'];

    // Root-aware permission mode. Claude CLI refuses both
    // `--dangerously-skip-permissions` AND `--permission-mode
    // bypassPermissions` when running as root (Anthropic's security
    // policy — common WSL trap since WSL defaults to root). `plan`
    // mode is the read-only escape hatch that doesn't trigger the
    // refusal: reviewer / pure-text-generation use cases work
    // normally, but the doer can't edit files or run bash. For
    // chorus's typical reviewer fan-out (Tri-Review, Review-Only,
    // Code-Review) plan mode is functionally equivalent. Power users
    // running implement-style templates (Red-Green's implement phase)
    // need a non-root user.
    const runningAsRoot =
      typeof process.getuid === 'function' && process.getuid() === 0;

    // Sandbox profile → permission-mode mapping. Headless mode auto-skips the
    // workspace-trust dialog (per `claude -p` docs) so we don't need preflight
    // beyond the trust marker write above.
    if (opts.sandbox === 'strict' || runningAsRoot) {
      // 'plan' = read-only-ish: no Edit, no Bash. The closest Claude has to
      // a strict sandbox in headless. Also the only mode Claude allows under
      // root, so we use it as the fallback when uid 0.
      args.push('--permission-mode', 'plan');
      if (runningAsRoot && opts.sandbox !== 'strict') {
        // One-line stderr nudge so power users notice the downgrade
        // and know to switch off root for write-capable doers.
        process.stderr.write(
          '[chorus] Claude running as root: permission mode downgraded to "plan" ' +
            '(read-only). Run chorus as a non-root user for file-editing doer slots.\n',
        );
      }
    } else if (opts.autoApprove !== false || opts.sandbox === 'full') {
      // Default for headless reviewer spawns — bypass per-tool prompts so the
      // agent doesn't hang waiting on stdin for an approval that'll never come.
      args.push('--permission-mode', 'bypassPermissions');
    }
    // 'workspace' (default) without auto-approve: leave Claude in default
    // permission mode (per-tool prompts). Will hang on the first prompt
    // unless settings.local.json pre-approves what's needed — which is what
    // `chorus init` writes into the user's config.

    if (opts.model) {
      args.push('--model', opts.model);
    }

    // Claude doesn't have a "no network" flag in headless, so networkAccess
    // is implicitly governed by the user's claude config and any tool the
    // agent attempts. Strict sandbox + plan mode is our gate.

    const run = spawnHeadless({
      command: 'claude',
      args,
      cwd: opts.cwd,
      stdinPayload: opts.promptText,
      parseLine: parseClaude,
      cli: 'claude',
      timeoutMs: opts.timeoutMs,
      abortSignal: opts.abortSignal,
      // Streaming CLI — no heartbeat needed; text_delta events provide
      // continuous progress signal.
      heartbeat: false,
    });

    return run.events;
  },

  estimateCostUsd(): number {
    // Claude Code uses subscription, not per-call API billing
    return 0;
  },
};
