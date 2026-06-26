# Scarf v2.10.0

Hermes v0.15.0 (v2026.5.28) — "The Velocity Release" — catch-up. Adopts the GUI-relevant surface of v0.15 across providers, web search, messaging platforms, voice, the 104-PR Kanban maturation wave, and five larger surfaces (Bitwarden secrets, supply-chain audit, MCP mTLS, skill bundles, per-session ACP edit-approval modes). Every new affordance is capability-gated on `HermesCapabilities` v0.15 flags — pre-v0.15 hosts render the v2.9.x surface unchanged. All flag/config/wire shapes were verified against the `v2026.5.28` Hermes source before implementation.

## Providers & models

- **OpenAI API as a first-class provider** — wire ID `openai-api`, distinct from the existing OpenAI Codex runtime. Surfaces in the Models picker. (Bare `openai` stays a Hermes alias to OpenRouter, so it is intentionally not registered.)
- **Krea image generation** — `krea-2-medium` ($0.03) and `krea-2-large` ($0.06) added to the image-generation model list (Settings → Auxiliary).
- **xAI May-15 model retirement** — retired Grok IDs (`grok-4-0709`, `grok-4-fast-*`, `grok-4-1-fast-*`, `grok-code-fast-1`, `grok-3`) resolve to `grok-4.3` in the picker, and `grok-imagine-image-pro` → `grok-imagine-image-quality`, so a stored retired model still works. The Health view warns when your configured model is retired and offers one-click `hermes migrate xai`.
- **Vercel removed** — Vercel AI Gateway (provider) and Vercel Sandbox (terminal backend) were deleted from Hermes in v0.15 and are dropped from Scarf's picker and terminal-backend list.

## Web Tools

- **xAI Web Search** — new `web_tools.search.backend` option (`xai`), reusing your Grok OAuth / `XAI_API_KEY` credentials with no new env var.

## Platforms

- **ntfy** — 23rd gateway platform. Push notifications via a topic URL, no account required. New setup form (topic, server, optional publish-topic / token / markdown) under the Platforms tab.
- **Per-platform config** — Telegram `disable_topic_auto_rename` + `ignore_root_dm`, Discord `allow_any_attachment`, Signal group-only `require_mention`, surfaced in each platform's setup form.

## Voice

- **xAI TTS `auto_speech_tags`** — opt-in toggle (default off) that inserts light `[pause]` tags for more natural xAI voice replies.

## Kanban — the v0.15 maturation wave

- **Server-side sort** — a sort picker on the board header (`--sort` by priority, created, status, assignee, title, updated).
- **New lifecycle actions** — Promote (`todo`/`blocked` → `ready`), Schedule / Park, and Delete-permanently (hard-delete archived tasks via `archive --rm`) as card context actions.
- **New columns** — Scheduled (parked tasks) and Review (post-work verification), collapsed when empty.
- **Worktree + model surfacing** — per-task worktree `--branch` on create; the inspector shows a task's branch and (read-only) model override.
- **Precise chat-scoped board** — the chat-header Kanban chip and the board it opens now filter by the originating ACP `session_id` (stamped automatically by Hermes), replacing the old tenant + time-window approximation. A "This chat ⇄ All tasks" scope toggle lets you widen the view.
- Multi-board `--board` plumbing is wired through the service layer (a board switcher UI and the `swarm` create-sheet are follow-ups).

## Secrets

- **Bitwarden Secrets Manager** — new Settings → Secrets tab for the `secrets.bitwarden` block: enable it, point Hermes at a bootstrap token (env var, default `BWS_ACCESS_TOKEN`), set project ID / server URL (US / EU / self-hosted) / cache TTL / auto-install / override-existing. The bootstrap token itself lives in `~/.hermes/.env`, not in this form.

## MCP

- **mTLS client certificates** — client cert / key paths + an SSL-verify control (with optional custom CA bundle) for HTTP and SSE MCP servers, in the server editor.
- **Nous MCP catalog** — a read-only "Browse catalog" sheet showing `hermes mcp catalog` output.

## Skills

- **Skill bundles** — a read-only Bundles tab listing the named skill groups in `~/.hermes/skill-bundles/*.yaml` (each loadable in Hermes via one `/<name>` slash command), with each bundle's member skills and instruction.

## Chat

- **Per-session edit-approval modes** — a chat-header chip to switch a live session between Default (ask before edits), Accept Edits (auto-allow workspace + /tmp edits), and Don't Ask (auto-allow except sensitive paths) via ACP `session/set_mode`. Separate from the global `approvals.mode` / YOLO surface; sensitive paths always still prompt.

## Health

- **Supply-chain audit** — a "Run supply-chain audit" button that runs `hermes audit` (OSV.dev) and shows the result inline.

## Fixes & polish (from the pre-release review)

- **Kanban session scope (correctness)** — the chat-scoped board no longer drops session tasks the agent created without tagging the project tenant; the precise `--session` filter is now authoritative.
- **Kanban scope pill** — shown for global chats too, so a session-scoped board with no project tenant isn't locked to the session filter.
- **MCP SSL-verify** — the verify toggle and the CA-bundle path are independent controls; toggling verification off no longer wipes a typed CA path.
- Removed dead code left behind when the Kanban time-window heuristic was replaced by session scoping.

## Hermes compatibility

Targets Hermes v0.15.0 (v2026.5.28). Pre-v0.15 hosts continue to work — every v0.15 surface is capability-gated on `HermesCapabilities.has*` flags so older Hermes installs gracefully hide the new affordances. The bulk of v0.15 (the `run_agent.py` refactor, cold-start performance, promptware defense, `session_search` rebuild, the Ink TUI orchestrator, the web dashboard, Docker s6 supervision, the API-server session-control REST) is server-side or other-frontend and benefits Scarf transparently with no Scarf change.

## Upgrade notes

- The Sparkle appcast at `https://awizemann.github.io/scarf/appcast.xml` will offer this update automatically on next launch.
- macOS 14.6+ (Sonoma) deployment target unchanged.
- No breaking changes to `~/.hermes/` state. New config keys (`secrets.bitwarden.*`, `platforms.ntfy.*`, `tts.xai.auto_speech_tags`, MCP `client_cert`/`client_key`/`ssl_verify`, the per-platform flags) are written only when you set them.
