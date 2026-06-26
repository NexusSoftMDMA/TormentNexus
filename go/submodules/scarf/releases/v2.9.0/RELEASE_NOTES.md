# Scarf v2.9.0

Hermes v0.14.0 (v2026.5.16) — "Foundation Release" — catch-up. Adds support for the new ACP slash commands, two new providers, two new gateway platforms, two new web-search backends, the local OpenAI-compatible proxy, ACP browser-tools setup, the per-turn file-mutation verifier surface, a YOLO mode warning, and a long tail of settings and config additions. Bundles 9 pending fixes and feature commits that landed since v2.8.0.

## Hermes v0.14 catch-up

### Chat
- **`/subgoal` slash command** — append extra success criteria to an active `/goal` loop. Surfaced as a trailing line on the goal pill. Forms: `<text>`, `remove N`, `clear`.
- **`/yolo` slash command** — toggle YOLO mode (skip dangerous-command approvals) without leaving chat. Pairs with a new chat-header warning badge when YOLO is on.
- **`/sessions` slash command** — browse and resume prior sessions from inside an active chat.
- **`/codex-runtime` slash command** — toggle the Codex app-server runtime for OpenAI/Codex models.
- **Per-turn file-mutation verifier** — v0.14 Hermes appends a verifier footer to assistant turns that wrote files (configurable via `file_mutation_verifier`). The footer flows through Scarf chat as plain agent output; bespoke card-style rendering is deferred until the upstream format stabilizes.
- **YOLO mode warning banner** — chat-header badge when `agent.approval_mode = yolo`. Mirrors the warning Hermes itself surfaces in its banner.

### Providers
- **xAI Grok OAuth (SuperGrok)** — sign in with your xAI account, talk to Grok models from Scarf. Subscription-gated; surfaces in the Models picker with a "Subscription" pill.
- **NovitaAI** — new API-key inference provider.
- **Alibaba Cloud → Qwen Cloud** — display rename to match Hermes's picker. Provider wire ID stays `alibaba` so existing config keys keep working.

### Platforms
- **LINE Messaging API** — 21st gateway platform.
- **SimpleX Chat** — 22nd gateway platform. Requires a local `simplex-chat` daemon running in WebSocket mode (setup instructions in the Discovery panel).

### Web Tools
- **Brave Search (free tier)** + **DuckDuckGo (DDGS)** — two new web-search backends. Brave honors a `BRAVE_SEARCH_API_KEY` env var for higher quotas; DDGS works anonymously.

### MCP, plugins, and config
- **MCP `supports_parallel_tool_calls`** — new optional flag on MCP server entries; the agent batches concurrent tool calls instead of serializing.
- **OpenRouter `min_coding_score`** — knob for the Pareto Code router (`openrouter/pareto-code`). Routes to the cheapest model meeting your quality bar (0.0–1.0, default 0.65).
- **Custom provider `api_mode`** — explicit wire-protocol selection (`chat_completions` / `anthropic_messages` / `codex_responses`).
- **Plugin `tool_override` badge** — plugins that replace a built-in tool now show a badge in PluginsView.
- **`docker_extra_args`** — extra flags passed verbatim to `docker run` for the docker-backed terminal backend.
- **`display.timestamps`** — per-message timestamps toggle.
- **Cron `deliver=all`** — fan-out cron job delivery to every connected channel.
- **Discord channel-history backfill** — toggle the v0.14 default-on backfill behavior per profile.

### New feature surfaces
- **Hermes Proxy** — new sidebar destination under Configure. Runs `hermes proxy start` with the user's OAuth-authenticated upstream (Nous Portal in v0.14, more adapters coming). Surfaces the endpoint URL (`http://127.0.0.1:8645/v1`), routed-provider status, and a tail of the startup log. Use it to point Codex / Aider / Cline / VS Code Continue at your Hermes subscription.
- **ACP browser tools setup** — Health view gains a "Run setup" button that calls `hermes acp --setup-browser` to install Chromium and provision Playwright in one shot.

## Performance + reliability fixes (post-v2.8.0)

- **Per-project model presets + mid-chat switcher** — save named model+provider presets, bind one to a project, or switch mid-chat via a popover from the chat header.
- **Kanban toolset enable** — direct YAML write with detector-based verification (the CLI flow had a regression that didn't persist correctly).
- **MetricKit crash + hang diagnostics** — Scarf-iOS persists MetricKit reports across launches so post-mortems work after a TestFlight crash.
- **JSON read caps** — session-attribution and project-dashboard JSON reads now bound their input to defend against pathological files.
- **Scroll crash + background-lifecycle hardening** — eliminates a scroll-view crash and tightens behavior when the app suspends mid-chat.
- **Selectable text across paragraphs** in chat transcript (#93).
- **Process pipe draining** for large Kanban outputs (#95).
- **Transcript render perf** — chat transcript scrolls faster on long conversations, plus a cron-filter and two-stage load-earlier.
- **iOS slash command parity** — iOS chat now sees the same slash menu as Mac, plus four TestFlight-reported fixes.

## Hermes compatibility

Targets Hermes v0.14.0 (v2026.5.16). Pre-v0.14 hosts continue to work — every v0.14 surface is capability-gated on `HermesCapabilities.has*` flags so older Hermes installs gracefully hide new affordances rather than throwing.

Note: the new `/handoff` slash command in Hermes v0.14 is messaging-platform handoff (CLI-only). Mid-chat model switching in Scarf continues to use the existing `session/set_model` RPC path via the chat-header model badge.

## Upgrade notes

- The Sparkle appcast at `https://awizemann.github.io/scarf/appcast.xml` will offer this update automatically on next launch.
- macOS 14.6+ (Sonoma) deployment target unchanged.
- No breaking changes to `~/.hermes/` state. The Scarf sidecar at `~/.hermes/scarf/` gains a new optional file for per-project Hermes proxy bindings (created lazily on first use).
