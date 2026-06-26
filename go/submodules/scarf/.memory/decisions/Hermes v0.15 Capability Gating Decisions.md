---
title: Hermes v0.15 Capability Gating Decisions
type: note
permalink: scarf/decisions/hermes-v0.15-capability-gating-decisions
tags:
- hermes
- capabilities
- v015
---

## Observations
- [decision] Vercel AI Gateway + Sandbox removed from Hermes in v0.15 — Scarf drops `vercel` from demotedProviders, modelAliases, and terminalBackends (unconditional, no flag) #providers
- [decision] OpenAI is now a first-class provider with wire ID `openai-api` (distinct from `openai-codex`). Bare `openai` is a Hermes alias for `openrouter` so Scarf does not register it #providers
- [decision] xAI May-15 retired Grok model IDs (grok-4-0709, grok-4-fast-*, grok-3, grok-code-fast-1, grok-imagine-image-pro, etc.) resolve forward to grok-4.3 / grok-imagine-image-quality in modelAliases — mirrors hermes_cli/xai_retirement.py #providers
- [decision] Kanban chat-scope: Hermes now stamps ACP session_id on tasks via HERMES_SESSION_ID env around run_conversation — `hermes kanban list --session <id>` filters server-side. Removed the old client-side sessionStartedAt/filterBySessionStart approximation; chat chip + handoff now gated on hasKanbanSessionFilter (>= 0.15). Global Kanban sidebar + per-project tab stay on hasKanban (v0.12+) #kanban
- [decision] /handoff is NOT a model handoff — it's cli_only=True and hands a session to a messaging platform (Telegram, Discord). Scarf intentionally does not add it to the ACP slash menu. Mid-chat model switching uses session/set_model RPC under hasACPSetSessionModel (v0.13) #chat
- [decision] Per-session edit-approval modes (Default/Accept Edits/Don't Ask) via ACP session/set_mode are distinct from global approvals.mode/YOLO — sensitive paths always still prompt regardless #safety
- [decision] Hermes Proxy (v0.14+) is local-only in v1 — SSH remote contexts show explanatory notice (would need port-forward wiring) #proxy
- [sync-target] Keep these in sync on each Hermes bump: overlayOnlyProviders / modelAliases / demotedProviders / imageGenModels (vs hermes_cli/providers.py + models.py + xai_retirement.py); platform roster (vs plugins/platforms/ + gateway/platforms/); search/TTS backend lists #maintenance

## Relations
- implements [[Hermes Integration]]