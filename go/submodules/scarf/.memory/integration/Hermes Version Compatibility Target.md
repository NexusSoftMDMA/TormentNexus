---
title: Hermes Version Compatibility Target
type: note
permalink: scarf/integration/hermes-version-compatibility-target
tags:
- hermes
- compatibility
- versioning
source_sha: 427321d742d63298100f9e444f96fd1524d7a46c
source_paths: README.md, CLAUDE.md, scarf/Packages/ScarfCore/Sources/ScarfCore/Services/HermesCapabilities.swift, scarf/scarf/Core/Services/HermesLogService.swift
---

## Observations
- [target] Current Scarf (v2.10.2) targets Hermes v0.16.0 / v2026.6.5 (was v0.15.2 / v2026.5.29 through Scarf v2.10.x; that line stays fully supported) — the latest patch on the v0.15 "Velocity Release" line. v0.15.0/.1/.2 are all back-compat (v0.15.1 is the hotfix wave: dashboard 401, Docker --insecure opt-in, MCP bare-command resolution, Kanban SIGTERM, skills.sh catalog, /yolo mid-session, /model parity; v0.15.2 is a packaging-only fix for plugin.yaml). #current
- [compatibility] Minimum supported Hermes: v0.6.0 (2026-03-30). All versions v0.6.0 through v0.15.2 are verified. Older Hermes versions degrade gracefully — new behavior is capability-gated. #minimum
- [no-v016] v0.16.0 (v2026.6.5) shipped 2026-06-05 and is now Scarf's target — see [[Hermes v0.16 Compatibility Decisions]]. v0.15.1/.2 introduce no new ACP wire formats, no new slash commands, no new tool kinds, no new permission flows — Scarf's v0.15 capability flags remain authoritative; no source-side gating work is needed for the patches. #status
- [schema] Scarf reads Hermes's SQLite state.db and parses CLI output from `hermes status`, `hermes doctor`, `hermes tools`, `hermes sessions`, `hermes gateway`, `hermes pairing`. Automatic schema detection provides backward compatibility. #schema
- [parsing] Log lines may carry an optional `[session_id]` tag between level and logger name; `HermesLogService.parseLine` treats the session tag as an optional capture group so older untagged lines still parse. #logs
- [sync-checklist] On each Hermes bump, keep in sync: `overlayOnlyProviders` / `modelAliases` / `demotedProviders` / `imageGenModels` (vs hermes_cli/providers.py + models.py + xai_retirement.py), the platform roster (vs plugins/platforms/ + gateway/platforms/), and the search/TTS backend lists. #maintenance

## Relations
- implements [[Hermes Capability Gating Pattern]]
- supersedes [[Hermes v0.15 Capability Gating Decisions]]