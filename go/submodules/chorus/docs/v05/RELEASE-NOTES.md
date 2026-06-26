# Chorus v0.5 — Release Notes

Pre-release of the multi-LLM peer-review orchestrator. Apache-2.0.

## What ships

### Real-LLM pipeline (end-to-end)

Daemon spawns each doer/reviewer in its own tmux session, writes the prompt
via `tmux send-keys -l`, watches for `## DONE` in `answer.md` via chokidar,
emits SSE events to the cockpit, and aggregates verdicts. Tested on:

- **Doer:** Claude Code (Opus 4.7)
- **Reviewer:** Gemini 3.1 Pro Preview
- **Bug-diagnose template** completes in ~3 minutes per round, doer answer
  + reviewer critique persisted to `~/.chorus/chats/<id>/round-N/`.

### 7 orchestrators wired

`chorus init` detects and connects:

- **Claude Code** — full inbound + outbound (slash command + MCP perms + workspace trust)
- **Codex CLI** — inbound + outbound (uses primary `~/.codex/`, no per-tool gating)
- **Gemini CLI** — inbound + outbound (`--trust` set on MCP server)
- **OpenCode** — inbound + outbound (mcp.chorus block in user config)
- **Kimi CLI (Moonshot)** — inbound + outbound, `--afk` flag for auto-approve
- **Cursor** — inbound only (IDE, not headless-spawnable)
- **Windsurf** — inbound only (IDE, not headless-spawnable)

### Cockpit (Next.js 16, port 5050)

- `/` — home: empty hero or recent chats + "Reviewer fleet" panel showing per-CLI health (quota state + reset time)
- `/onboarding` — 3-section first-run wizard (CLIs / API keys / sandbox profile)
- `/connect` — per-CLI status cards with one-click wire-up
- `/runs/<id>` — live run view with phase stepper, doer + reviewer grid, real artifacts streamed via SSE
- `/templates`, `/new` — template gallery + new-chat composer
- `/settings`, `/settings/permissions` — sandbox profile, auto-approve, network access

### Daemon (Fastify, port 7707)

`GET /health`, `GET|POST /chats`, `GET /chats/:id/stream` (SSE), `GET /templates`,
`GET|PUT /settings`, `GET /orchestrators`, `POST /orchestrators/:name/connect`,
`GET /cli/health`, `GET|PUT /settings/permissions`.

### MCP server (stdio, 7 tools)

`create_chat`, `wait_for_chat`, `get_chat_status`, `list_blocked`,
`resume_chat`, `cancel_chat`, `list_templates`.

### Templates (4 built-in)

`bug-diagnose`, `code-review` (Claude doer + Codex and Gemini reviewers, 2-of-2 quorum),
`architect-review`, `red-green`.

### Sandbox profiles

`strict` (read-only) / `workspace` (default — write inside chat dir, no
network) / `full` (no sandbox). Each CLI shim translates the abstract
profile into the correct flag (`gemini --approval-mode default`,
`codex -c sandbox_mode=read-only`, `claude --dangerously-skip-permissions`,
`kimi --afk`).

### Headless transport (new in v0.5)

Default transport for every CLI is now subprocess + stream-json instead of
tmux + send-keys. Each shim spawns the CLI in `--print`/`exec`/`run` mode,
pipes the prompt via stdin or argv, parses stream-json events from stdout.

- **~80% lower RAM:** no resident TUI process between rounds (~200-500MB
  per CLI saved). 3 reviewers in parallel went from 1-1.5GB tmux overhead
  to near-zero.
- **Permission dialogs disappear:** all 5 CLIs auto-approve in headless mode
  (Claude `--permission-mode bypassPermissions`, Gemini `--approval-mode
  auto_edit`, Kimi `--print`, Codex `exec` honors sandbox config, OpenCode
  `run` is non-interactive).
- **Stream-json parsers verified:** Claude (Anthropic format) and Gemini
  (Google format) tested against real fixtures. Kimi reuses Claude's parser
  (Moonshot intentionally Claude-Code-compatible). OpenCode + Codex are
  one-shot; UI shows a heartbeat every 5s during the silent run.
- **Live streaming UI:** run page subscribes to `phase_progress` SSE events
  and renders text deltas instantly, no 4s polling lag.
- **Stuck-process safety:** every spawn gets a 10-min hard timeout +
  AbortSignal hook, SIGTERM-then-SIGKILL grace, on-disk PID registry,
  daemon-startup orphan reaper. Hung CLIs can't burn API tokens forever.
- **Tmux is still a first-class option:** Settings → Transport toggle, or
  `CHORUS_TRANSPORT=tmux` env override. No deprecation timeline. Use it when
  you want to attach (`tmux attach -t <name>`) and watch agents work
  step-by-step.

Mixed-mode chats are supported: a chat can run Claude doer in headless and
Gemini reviewer in tmux fallback if a shim's `runHeadless` is missing for
the lineage.

### Permission-prompt auto-recovery

Defense-in-depth for the rare cases where Layer 1 (config pre-approval
written by `chorus init`) misses something:

- Error detector watches every CLI pane for "Always allow" / "Allow once"
  dialogs (lineage-agnostic — same regex catches OpenCode bash dialogs,
  Kimi tool prompts, subagent-spawn dialogs).
- Each shim declares `recoverKeys.permission_prompt` (e.g. OpenCode + Kimi
  both use `['Right', 'Enter']` — Right arrow to "Always allow", Enter to
  confirm).
- Runner sends those keys via tmux on detection, emits `cli_warning`
  (not `cli_error`), skips health degradation.
- Plus `connectOpencode` now writes a `permission.bash` allowlist to
  `~/.config/opencode/opencode.json` covering safe read-only ops (`git diff`,
  `cat`, `ls`, `find`, `rg`, etc.) so the dialog never appears for those.

### Status taxonomy

Chat statuses: `drafting | reviewing | approved | no_review | failed | cancelled | merged | blocked`.

`no_review` is new in v0.5: chat ends in this state when every reviewer in a
phase failed (quota / timeout / crash) — the doer output is still on disk
but no actual peer review happened, so we don't claim "approved".

## Deferred to v0.6

The following landed scoped down to make the ship date — full versions in v0.6.

### Slash commands for non-Claude CLIs

Only Claude Code gets a `/chorus` slash command in v0.5
(`~/.claude/commands/chorus.md` shipped via npm asset bundle). For Codex,
Gemini, Kimi, OpenCode, Cursor, Windsurf — invoke chorus via natural
language ("chorus, run bug-diagnose on…"). The MCP server picks it up.
Slash-command equivalents per-CLI in v0.6 once their format is stable.

### Lineage / Provider / CLI split

Current `Lineage` enum (`anthropic | openai | google | opencode | moonshot | any`)
conflates "model family" with "which CLI runs it". v0.6 splits to:

- **Vendor** (anthropic, openai, google, moonshot, deepseek, xai, mistral, …)
- **Channel** (cli vs. api)
- **CLI choice** (claude, codex, gemini, kimi, opencode)

So a user can route Kimi via the dedicated `kimi` CLI _or_ via `opencode`
configured for moonshot _or_ via direct API key, and the template author
just declares "I want a moonshot reviewer."

### Direct API-key reviewers

v0.5 ships CLI-spawn only (headless or tmux). v0.6 adds an alternate spawn
path that calls the vendor SDK directly when the user has provided an API
key, bypassing the CLI subprocess entirely. Faster cold start, no per-CLI
quirks.

### Real-output verification of Gemini/Kimi/OpenCode/Codex headless

Claude's stream-json was verified against a real `claude --print` capture
2026-04-30 with 8 inline tests. Gemini was verified the same way. Kimi
reuses Claude's parser (per Moonshot's stated compatibility) but a captured
fixture lands in v0.6. OpenCode + Codex emit one-shot `message_done` from
full stdout — sufficient for v0.5 but a streaming variant for OpenCode
(`opencode serve`?) is on the v0.6 list.

### Live verdict synthesis

The current run page shows raw doer + reviewer text. v0.6 layers in
parsed-verdict extraction (severity tagging, finding aggregation, decision
panel UX) — the structured pieces from the prototype LiveRunView, fed by
real data.

### Run-page resilience

SSE stream disconnect currently treats abort as cancel (runner.ts:71-80).
v0.6 detaches the runner lifecycle from the SSE consumer — chat keeps
running headless, SSE just tails state from DB.

### Cockpit run page nav link to /settings/permissions

Added a link from the main `/settings` page in v0.5; full sidebar nav
restructure in v0.6.

## Known limitations

- Chrome/Cursor/Windsurf "Always allow" prompts can't be config-pre-approved
  (no public flag exists upstream). One-time click per machine.
- Codex 5h / weekly quota windows aren't reported by the CLI itself; chorus
  shows reset time only when the CLI emits a "try again at <X>" message.
- The legacy mock LiveRunView component is still in the codebase (used by
  no live page). Will be deleted in v0.6.

## Install

See `docs/v05/INSTALL-AND-TEST.md`.
