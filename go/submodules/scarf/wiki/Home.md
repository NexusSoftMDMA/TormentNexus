---
title: Home
type: note
permalink: scarf-wiki/home
---

# Scarf

A native macOS companion app for the [Hermes AI agent](https://github.com/hermes-ai/hermes-agent). Full visibility into what Hermes is doing, when, and what it creates — across one local install or many remote ones.

**Latest release:** [v2.12.0](https://github.com/awizemann/scarf/releases/tag/v2.12.0) — a coordinated catch-up to **Hermes v0.17.0**, bundled with a **remote-chat performance fix**: typing into a session on an SSH host no longer lags or spikes CPU now that watcher-driven reads run off the main thread ([#119](https://github.com/awizemann/scarf/issues/119)). The v0.17 audit found the upstream surface entirely stable (zero mandatory changes), so it's a feature catch-up plus pre-existing-bug fixes — four broken Health/Settings CLI actions and the Curator "Prune" rebuilt as the reversible **"Archive idle skills"** — alongside new **WhatsApp Business Cloud API** + **SimpleX** gateway setup forms, Telegram rich-messages / status toggles, an opt-in curator-consolidation toggle, and a max-concurrent-sessions cap, all capability-gated. Targets Hermes v0.17.0. See [v2.12.0 release notes](https://github.com/awizemann/scarf/blob/main/releases/v2.12.0/RELEASE_NOTES.md).

**Previous release:** [v2.10.0](https://github.com/awizemann/scarf/releases/tag/v2.10.0) — coordinated catch-up to Hermes v0.15.0, "The Velocity Release". Ships **OpenAI as a first-class provider** (wire ID `openai-api`, distinct from OpenAI Codex; bare `openai` stays a Hermes alias to OpenRouter so it's intentionally not registered), **Krea image-gen models** (`krea-2-medium` / `krea-2-large`) + the **xAI May-15 model-retirement aliases** (retired Grok IDs and `grok-imagine-image-pro` resolve forward; Vercel AI Gateway + Vercel Sandbox dropped, deleted upstream), **xAI Web Search** as a `web_tools.search.backend: xai` option, **ntfy** as the 23rd gateway platform (push notifications via a topic URL, no account) plus per-platform flags (Telegram `disable_topic_auto_rename` / `ignore_root_dm`, Discord `allow_any_attachment`, Signal group-only `require_mention`), the **xAI TTS `auto_speech_tags`** opt-in toggle, the **Kanban v0.15 maturation wave** (server-side `--sort`; Promote / Schedule / Delete-permanently card actions; new Scheduled + Review columns; per-task worktree `--branch` + read-only `model_override` in the inspector; a precise chat-scoped board keyed by the originating ACP `session_id` with a "This chat ⇄ All tasks" scope toggle; `--board` multi-board plumbing in the service layer), **Bitwarden Secrets Manager** as a new Settings → Secrets tab (`secrets.bitwarden.*`), **MCP mTLS client certs** (`client_cert` / `client_key` / `ssl_verify`) + a read-only `hermes mcp catalog` browse sheet, a read-only **skill Bundles tab** over `~/.hermes/skill-bundles/*.yaml`, **per-session edit-approval modes** via a chat-header chip (Default / Accept Edits / Don't Ask through ACP `session/set_mode`), and a Health **"Run supply-chain audit"** button (`hermes audit`, OSV.dev) + an xAI retired-model warning with one-click `hermes migrate xai`. New v0.15 capability flags gate every surface; pre-v0.15 hosts render the v2.9.x layout unchanged. See [v2.10.0 release notes](https://github.com/awizemann/scarf/blob/main/releases/v2.10.0/RELEASE_NOTES.md).

**Earlier release:** [v2.9.0](https://github.com/awizemann/scarf/releases/tag/v2.9.0) — coordinated catch-up to Hermes v0.14.0, "The Foundation Release". Ships **`/subgoal` + `/yolo` + `/sessions` + `/codex-runtime`** ACP slash commands (subgoals layer extra success criteria onto an active `/goal` loop with a `+N` count badge inside the goal pill; the YOLO chip warns when dangerous-command auto-approval is on), **two new inference providers** (xAI Grok OAuth / SuperGrok as a subscription-gated overlay + NovitaAI as an API-key overlay; Alibaba → Qwen Cloud display rename mirrors Hermes's picker without changing the wire ID), **two new gateway platforms** (LINE Messaging API as the 21st platform, SimpleX Chat as the 22nd talking to a local `simplex-chat` daemon over WebSocket), **two new web-search backends** (Brave Search free tier + DuckDuckGo via DDGS), **Hermes Proxy** as a new Configure → Hermes Proxy sidebar destination wrapping `hermes proxy start` to give Codex / Aider / Cline / VS Code Continue an OpenAI-compatible local endpoint that attaches your authenticated upstream credentials, **MCP `supports_parallel_tool_calls`** as a tri-state picker in the server editor, **ACP `--setup-browser`** as a one-click chromium + playwright provisioning button on the Health view, plus a stack of smaller settings (`terminal.docker_extra_args`, `display.timestamps`, Cron `deliver=all`, Discord channel-history backfill, plugin `tool_override` manifest badge) and the bundled performance + reliability fixes since v2.8 (per-project model presets + mid-chat switcher via `session/set_model`, Kanban toolset enable YAML write fix, MetricKit crash + hang diagnostics on iOS, JSON read caps in session-attribution + project-dashboard, scroll crash + background-lifecycle hardening, selectable text across paragraphs, process pipe draining for large Kanban output, transcript render perf, iOS slash-command parity, `/steer` honesty pre-session). 24 new capability flags gate every v0.14 surface; pre-v0.14 hosts render the v2.8.0 layout byte-identical. See [v2.9.0 release notes](https://github.com/awizemann/scarf/blob/main/releases/v2.9.0/RELEASE_NOTES.md).
**Latest mobile:** [Join the public TestFlight](https://testflight.apple.com/join/qCrRpcTz). The link is live now but only accepts new beta testers once Apple's Beta Review approves the first build — see [ScarfGo](ScarfGo) for the full feature tour.
**Targets Hermes:** v0.15.0 (v2026.5.28) — OpenAI (`openai-api`) first-class provider, Krea image-gen models, xAI model-retirement aliases (Vercel AI Gateway + Sandbox removed), xAI Web Search backend, ntfy (23rd platform) + per-platform flags, xAI TTS `auto_speech_tags`, Kanban v0.15 wave (server-side `--sort`, Promote / Schedule / Delete-permanently actions, Scheduled + Review columns, worktree `--branch` + `model_override` display, session-scoped board, `--board` plumbing), Bitwarden Secrets Manager, MCP mTLS client certs + `mcp catalog` browse, skill Bundles tab, per-session edit-approval modes (`session/set_mode`), Health supply-chain audit + xAI migrate action. v0.14.0 / v0.13.0 / v0.12.0 / v0.11.0 / v0.10.0 still work for everything that didn't change — Scarf detects the host's Hermes version and hides v0.15-only surfaces gracefully.
**Available in:** English, Simplified Chinese (zh-Hans), German (de), French (fr), Spanish (es), Japanese (ja), Brazilian Portuguese (pt-BR). See [Localization](Localization). _ScarfGo is English-only in v1._

## Quick links

- [Installation](Installation) — download, first launch, system requirements (Mac)
- **[ScarfGo](ScarfGo)** — the iPhone companion (public TestFlight from v2.5)
- **[ScarfGo Onboarding](ScarfGo-Onboarding)** — SSH keys, paste-public-key, connection test
- [Platform Differences](Platform-Differences) — Mac vs iOS feature matrix
- [First Run](First-Run) — what Scarf expects in `~/.hermes/`
- [Project Templates](Project-Templates) — `.scarftemplate` bundles, install / export / author
- **[Slash Commands](Slash-Commands)** — author project-scoped slash commands (v2.5+)
- **[Hermes Proxy](Hermes-Proxy)** — OpenAI-compatible local server for Codex / Aider / Cline / VS Code Continue (v2.9+, Hermes v0.14+)
- **[Design System](Design-System)** — ScarfColor / ScarfFont / components reference
- [Architecture Overview](Architecture-Overview) — MVVM-F, services, transport, ScarfCore
- [Performance Monitoring](Performance-Monitoring) — ScarfMon: opt-in perf instrumentation, how to capture a baseline
- [Servers & Remote](Servers-and-Remote) — adding remote Hermes hosts over SSH
- [Localization](Localization) — supported languages + how to contribute a new one
- [Release Notes Index](Release-Notes-Index) — every version's notes
- [Troubleshooting: Update "improperly signed"](Troubleshooting-Sparkle-Update) — recovery if Sparkle rejects an update
- [Privacy Policy](Privacy-Policy) · [Support](Support) — what data the apps access; how to get help
- [Wiki Maintenance](Wiki-Maintenance) — how this wiki is edited and kept in sync

## What Scarf does

Scarf mirrors Hermes's surface area through a sidebar-based UI grouped into four sections:

- **Monitor** — Dashboard, Insights, Sessions, Activity. See what Hermes is doing.
- **Interact** — Chat, Memory, Skills. Talk to Hermes and shape what it knows.
- **Configure** — Platforms, Personalities, Quick Commands, Credential Pools, Plugins, Webhooks, Profiles, Servers. Set Hermes up.
- **Manage** — Tools, MCP Servers, Gateway, Cron, Health, Logs, Settings. Operate Hermes.

Scarf 2.0 is a multi-window app — one window per Hermes server, local or remote. Remote hosts are reached over plain SSH using your existing `~/.ssh/config`, agent, ProxyJump, and ControlMaster.

## Project status

Open-source (MIT), 160+ stars, actively maintained. See [Roadmap](Roadmap) for what's coming.

---
_Last updated: 2026-05-28 — Scarf v2.10.0 (Hermes v0.15 catch-up: OpenAI `openai-api` first-class provider + Krea image-gen + xAI model-retirement aliases (Vercel removed), xAI Web Search backend, ntfy (23rd platform) + per-platform flags, xAI TTS `auto_speech_tags`, Kanban v0.15 wave (`--sort` / Promote / Schedule / Delete-permanently / Scheduled + Review columns / worktree `--branch` + `model_override` / session-scoped board / `--board` plumbing), Bitwarden Secrets Manager, MCP mTLS client certs + `mcp catalog`, skill Bundles tab, per-session edit-approval modes, Health supply-chain audit + xAI migrate)_