# Chorus v0.5 — Real LLM Integration: Shared Context

All agents implementing this batch MUST read this and the relevant memory files before writing code.

## What we're building

Replace the daemon's stub phase events with real CLI invocations:
- Spawn a tmux session per chat/phase/agent (or reuse per the share-session policy).
- Launch the right CLI (claude / codex / gemini / opencode) with per-CLI quirks.
- Drop a prompt file in the chat's filesystem dir.
- Tell the CLI (via tmux paste-buffer) to read the prompt and write its answer.
- Watch the answer file via chokidar; emit SSE phase events on completion.
- Pattern-match capture-pane for known CLI failure modes; emit `cli_error` events.
- Reaper sweeps orphan/idle sessions every 5 min.

## Filesystem layout for chats

```
~/.chorus/chats/<chatId>/
  round-<N>/
    <role>-<agentName>/         # e.g. doer-claude-1, reviewer-codex-2
      ask.md                    # daemon writes the prompt here
      answer.md                 # CLI writes its answer here (we watch this)
      done                      # touch'd by CLI when finished (optional sentinel)
      stderr.log                # captured tmux pane snapshots for diagnostics
```

## Hard rules (from prior incidents — do NOT relearn these)

1. **Shell-injection guard at every layer.** When substituting user-supplied or template-supplied values into tmux launch commands, %q-quote at substitution time AND validate metacharacters at input time. See `feedback_shell_injection_via_tmux.md`.
2. **Per-session `CODEX_HOME` for codex spawns.** Without this, parallel codex sessions race on auth.json refresh and invalidate each other. See `feedback_codex_home_per_account.md`. Never copy `auth.json` between dirs.
3. **Single-line prompts for Gemini.** Multi-paragraph splits at every `\n` and runs each line as a separate query. Use `@/abs/path` inline syntax. See `feedback_gemini_multiline_prompts.md`.
4. **Single-line prompts for Opencode, no leading `/` or `@`.** Slash → slash-command, `@` → file-attach popup. Use plain "at /abs/path". See `feedback_gemini_multiline_prompts.md`.
5. **Always `/clear` opencode between rounds.** Opencode preserves session and drifts conversational. Cross-round context comes from the prompt file, not the session. See `feedback_opencode_clear_always.md`.
6. **Codex sandbox is transport-aware.** workspace-write blocks network (breaks gh transport). Add `-c sandbox_workspace_write.network_access=true` ONLY when transport=github. Default folder/tmux gets strictest. See `feedback_codex_sandbox_modes.md`.
7. **Never use Gemini --approval-mode yolo.** Yolo auto-approves write_file; empty-content writes have wiped repos. Use `--approval-mode auto_edit`. See `feedback_gemini_yolo_dangerous.md`.
8. **Never reuse tmux across chats.** Cross-task contamination is the failure mode. Reuse only across rounds of the same phase, and only if `iterate.shareSessionAcrossRounds` is true (default).
9. **Tag every chorus session with `chorus-<chatId>-...` prefix.** The reaper depends on this tag for reconciliation.
10. **Files ≤ 500 lines, split at 400.** TypeScript strict, no `any`, immutable patterns.

## Memory files to read

- `chorus_tmux_session_lifecycle.md` — fresh-per-task default, lifecycle states
- `agent_fleet_overview.md` — how the existing /work fleet is wired (you're porting this pattern)
- `openbridge_architecture.md` — agent shim model + folder transport
- `openbridge_security_model.md` — sandbox-by-transport, %q quoting layers
- `chorus_cli_failure_modes.md` — quota / token-refresh / MCP-handshake patterns
- `feedback_codex_home_per_account.md` — per-session CODEX_HOME setup
- `feedback_shell_injection_via_tmux.md` — three injection vectors closed
- `feedback_codex_sandbox_modes.md` — transport-aware sandbox flags
- `feedback_gemini_multiline_prompts.md` — single-line + `@/abs/path`
- `feedback_gemini_yolo_dangerous.md` — auto_edit, never yolo
- `feedback_opencode_clear_always.md` — /clear before each nudge
- `feedback_opencode_session_db_corruption.md` — empty content[] poison detection
- `feedback_opencode_external_directory_perm.md` — opencode.json permissions

## Existing reference code (read but don't import — different language)

- `/home/ubuntu/.local/bin/work` — bash-based /work orchestrator. Patterns to port:
  - Per-CLI prompt formatters (lines ~280-380)
  - Pre-nudge cleanup (Escape sequences for opencode/gemini overlays)
  - Capture-pane error detection
- `/home/ubuntu/dev/openbridge/lib/agents/codex.sh` — agent shim pattern reference
- `/home/ubuntu/dev/openbridge/lib/agents/claude.sh` — same

## Foundation interfaces (already defined — DO NOT modify)

- `src/daemon/agents/types.ts` — AgentShim, AgentSpawnOptions, AgentNudgeOptions
- `src/daemon/tmux-types.ts` — TmuxManager, SessionHandle, AcquireSessionOptions
- `src/lib/template-schema.ts` — added `iterate.shareSessionAcrossRounds` (default true) and `iterate.shareSessionAcrossPhases` (default false)

## Your subtree won't overlap

Each agent below owns a non-overlapping path. Stay in your subtree; consume the foundation interfaces only.
