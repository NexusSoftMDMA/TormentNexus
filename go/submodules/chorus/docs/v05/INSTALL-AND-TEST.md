# Install Chorus and run a real chat

This is the **dev-mode** install — the npm package isn't published yet, so we
point each orchestrator at the absolute path of `bin/chorus.mjs`.

## 1. Build + start the daemon

```bash
cd /home/ubuntu/dev/chorus
pnpm install
pnpm build:server
pm2 start ecosystem.config.cjs           # daemon on :7707, cockpit on :5050
```

Verify:

```bash
curl -s http://127.0.0.1:7707/health | jq
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:5050/   # 200
```

## 2. Connect every CLI on this box in one pass

```bash
pnpm exec tsx src/cli/index.ts connect
```

You should see roughly:

```
✓ Claude Code: MCP server registered · all 7 tools already approved · /chorus command installed
✓ Codex CLI: MCP server registered
✓ Gemini CLI: MCP server registered
✓ OpenCode: MCP server registered
✓ Kimi CLI: MCP server registered
✗ Cursor: not detected on this machine
✗ Windsurf: not detected on this machine

Heads up: Kimi CLI will show a one-time permission prompt on your first
chorus.* tool call. Click "Always allow" to make it stick.
```

To pick a subset:

```bash
pnpm exec tsx src/cli/index.ts connect claude,gemini
```

## 3. Where chorus is registered (per CLI)

| CLI | Config file edited | Pre-approval |
|---|---|---|
| Claude Code | `~/.claude.json` + `~/.claude/settings.local.json` | 7 tools auto-allowed |
| Codex | `~/.codex/config.toml` `[mcp_servers.chorus]` | inherits `approval_policy` |
| Gemini | `~/.gemini/settings.json` `mcpServers.chorus` | `--trust` set |
| OpenCode | `~/.config/opencode/opencode.json` `mcp.chorus` | `enabled: true` |
| Kimi | `~/.kimi/mcp.json` `mcpServers.chorus` | one-time TUI prompt |
| Cursor | `~/.cursor/mcp.json` `mcpServers.chorus` | first-call IDE prompt |
| Windsurf | `~/.codeium/windsurf/mcp_config.json` | first-call IDE prompt |

## 4. Fire a chat from the cockpit

Open <http://localhost:5050>. Click **New chat** → pick `bug-diagnose` →
paste a small snippet → Submit. You should see:

- Phase stepper at the top with one phase ("Diagnose Bug")
- Round 1 grid: doer card (Claude) and reviewer card (Gemini)
- Doer streams to "DONE" first, reviewer card flips to "WORKING", then "DONE"
- Status banner flips from `DRAFTING` → `REVIEWING` → `APPROVED` (or
  `NO REVIEW` if all reviewers fail)

## 5. Or fire from inside Claude Code

```
/chorus bug-diagnose Spot the bug in: function divide(a,b){return a/b}
```

The Claude Code instance calls `mcp__chorus__create_chat` (silent — all 7
tools pre-approved by step 2), then `mcp__chorus__wait_for_chat`, then
summarises the verdict.

## 6. Or fire from the daemon HTTP API directly

```bash
RESP=$(curl -s -X POST http://127.0.0.1:7707/chats \
  -H 'content-type: application/json' \
  -d '{"work":"Spot the bug in: const a=[1]; for(let i=0;i<=a.length;i++) console.log(a[i]);","templateId":"bug-diagnose"}')
CHAT_ID=$(echo "$RESP" | jq -r '.data.id')

# The SSE stream drives the runner — keep it open for the run's lifetime
curl -sN "http://127.0.0.1:7707/chats/$CHAT_ID/stream"
```

Open `http://localhost:5050/runs/$CHAT_ID` in another tab to watch live.

## 7. Permissions

Cockpit `/settings/permissions` controls reviewer sandbox. Three profiles
(strict / workspace / full), plus toggles for auto-approve-prompts and
network access. Onboarding asks the same questions on first run.

## 8. Troubleshooting

| Symptom | Fix |
|---|---|
| Cockpit can't reach daemon | `pm2 status` — both `chorus-web` and `chorus-daemon` should be online |
| Reviewer pane shows OAuth login URL | CLI needs auth here. `codex login`, `gemini --auth`, `kimi /login`. |
| Reviewer never writes answer.md | Attach with `tmux attach -t <session>`. Common causes: quota (cockpit "Reviewer fleet" panel shows reset time), or a new first-launch popup we haven't pre-suppressed yet. |
| Chat ends with `no_review` | All reviewers failed (typically quota). Use a template targeting a different lineage, or wait for reset. |
