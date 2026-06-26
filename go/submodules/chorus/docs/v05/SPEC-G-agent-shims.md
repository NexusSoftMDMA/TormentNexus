# Agent G — Per-CLI Agent Shims

You own: `src/daemon/agents/{claude,codex,gemini,opencode,index}.ts`.
**Don't modify** `src/daemon/agents/types.ts` (the foundation interface).

**Read first:** `docs/v05/SPEC-llm-shared.md`, then the listed memory files there.

Reference the running fleet as ground truth for CLI quirks:
- `~/.local/bin/work` (lines 280-380 for prompt formatters)
- `~/dev/openbridge/lib/agents/codex.sh` and `claude.sh`

## Build

One file per lineage, each implementing `AgentShim`. Plus a registry.

### `src/daemon/agents/claude.ts`

```ts
import type { AgentShim, AgentSpawnOptions, AgentNudgeOptions } from './types.js';

export const claudeShim: AgentShim = {
  lineage: 'anthropic',
  name: 'claude-code',
  buildLaunchCommand(opts) {
    const cwd = quotePath(opts.cwd);
    let cmd = `cd ${cwd} && claude`;
    if (opts.model) cmd += ` --model ${quoteValue(opts.model)}`;
    return cmd;
  },
  formatPrompt(opts) {
    // Claude handles multi-paragraph fine.
    const sentinel = opts.expectDoneSentinel ? '\n\nWhen finished, end your response with: ## DONE' : '';
    return `${opts.task}\n\nRead the prompt at: ${opts.promptFile}\n\nWrite your full answer to: ${opts.answerFile}${sentinel}`;
  },
  estimateCostUsd: () => 0,  // Claude Code uses subscription, not API
};
```

### `src/daemon/agents/codex.ts`

Per-session `CODEX_HOME` is mandatory for parallel safety. `accountId` (e.g. `cdx-1`, `cdx-2`) determines which CODEX_HOME directory to use:

```
~/.codex-<accountId>/
  config.toml          (copied from ~/.codex/config.toml)
  auth.json            (NOT copied — agent runs `codex login` first time)
```

If `CODEX_HOME` doesn't exist, the shim creates the dir + copies `config.toml` from `~/.codex/`. NEVER copies `auth.json` (would clone the same account; defeats the purpose).

Sandbox mode by transport:
- `transport === 'github'` → append `-c sandbox_workspace_write.network_access=true`
- `transport === 'folder' | 'tmux'` (default) → strict workspace-write
- `unsandboxed === true` → append `--dangerously-bypass-approvals-and-sandbox`

```ts
buildLaunchCommand(opts) {
  const codexHome = ensureCodexHome(opts.accountId ?? 'default');
  const flags: string[] = [];
  if (opts.transport === 'github') flags.push('-c', 'sandbox_workspace_write.network_access=true');
  if (opts.unsandboxed) flags.push('--dangerously-bypass-approvals-and-sandbox');
  if (opts.model) flags.push('--model', quoteValue(opts.model));
  return `cd ${quotePath(opts.cwd)} && CODEX_HOME=${quotePath(codexHome)} codex ${flags.join(' ')}`;
}
```

`formatPrompt`: multi-paragraph fine, plain text path references work.

`estimateCostUsd`: 0 (subscription).

### `src/daemon/agents/gemini.ts`

```ts
buildLaunchCommand(opts) {
  // --approval-mode auto_edit (NEVER yolo — file-deletion risk)
  // -m gemini-3.1-pro-preview (default 'gemini-pro' invalid on plan)
  let cmd = `cd ${quotePath(opts.cwd)} && gemini --approval-mode auto_edit`;
  if (opts.model) cmd += ` -m ${quoteValue(opts.model)}`;
  else cmd += ` -m gemini-3.1-pro-preview`;
  return cmd;
}

formatPrompt(opts) {
  // SINGLE LINE. Gemini submits each \n as a separate query.
  // Use @/abs/path inline syntax for file refs.
  // No newlines, no multi-paragraph.
  const sentinel = opts.expectDoneSentinel ? ' End your response with ## DONE.' : '';
  return `@${opts.promptFile} Read this file and follow the <ask> XML block, write your full answer to ${opts.answerFile}.${sentinel}`;
}
```

### `src/daemon/agents/opencode.ts`

Single-line prompts, never starts with `/` (slash-command) or `@` (file-attach popup). Always prepended with "Open the file at this absolute path using your read tool: " — the safe, well-tested form from `~/.local/bin/work`.

```ts
formatPrompt(opts) {
  // Single line. Plain text path. Never `/` or `@` lead.
  const sentinel = opts.expectDoneSentinel ? ' End with ## DONE.' : '';
  return `Open the file at this absolute path using your read tool: ${opts.promptFile} — follow the <ask> block, write your full answer to ${opts.answerFile}.${sentinel}`;
}

preNudge(sessionName) {
  // Always /clear opencode between rounds — see feedback_opencode_clear_always.md.
  // The Tmux manager's sendKeys is what we call here, but since this method
  // is sync, leave a hook the runner can invoke via mgr.sendKeys(...).
  // For this shim, expose a static `clearKeys` array the runner uses.
}

estimateCostUsd: () => 0  // Kimi/DeepSeek via OpenCode Go subscription
```

For the `preNudge` hook: since the shim doesn't get a tmux mgr handle, expose a `clearKeys` constant array `[Escape, Escape, '/clear', Enter]` that the runner pulls from the shim and pipes into `mgr.sendKeys()`. Add `clearKeys` as an optional readonly field to `AgentShim`. (You may add this to `types.ts` if absolutely needed — that's the ONLY sanctioned modification.)

### `src/daemon/agents/index.ts` — registry

```ts
import type { AgentRegistry, AgentShim, Lineage } from './types.js';
import { claudeShim } from './claude.js';
import { codexShim } from './codex.js';
import { geminiShim } from './gemini.js';
import { opencodeShim } from './opencode.js';

const SHIMS: Record<Lineage, AgentShim> = {
  anthropic: claudeShim,
  openai: codexShim,
  google: geminiShim,
  xai: opencodeShim,    // xAI/Grok via OpenCode → kimi or deepseek
  any: claudeShim,      // any → fallback to claude
};

export const registry: AgentRegistry = {
  pickShim(lineage, model) {
    return SHIMS[lineage] ?? SHIMS.any;
  },
  listAvailable() {
    return Object.values(SHIMS);
  },
};
```

### Shared helpers

You'll need a small `quoteValue(s: string): string` and `quotePath(p: string): string`. Implement once (e.g. in `src/daemon/agents/quote.ts`) using the printf %q equivalent:

```ts
export function quoteValue(s: string): string {
  // Bash-safe single-quoting: 'foo'\''bar' style for embedded quotes.
  return `'${s.replace(/'/g, `'\\''`)}'`;
}
export const quotePath = quoteValue;
```

Validate names UPFRONT: reject any `accountId`, `model`, etc. that contains shell metacharacters (`$\`;|&<>()\\"'`). Throw a typed error with a friendly message.

## Don't touch

- Anything outside `src/daemon/agents/`
- `src/daemon/agents/types.ts` (except adding optional `clearKeys` field if you genuinely need it)
- The foundation interfaces

## Acceptance

```bash
pnpm typecheck                       # green
pnpm lint                            # green for your files
```

Unit-test the shims manually — pick claudeShim, call `buildLaunchCommand({...})`, eyeball that it looks right.

**Branch / commit:**
```
feat(agents): per-CLI shims with battle-tested quirks (Agent G)
```
