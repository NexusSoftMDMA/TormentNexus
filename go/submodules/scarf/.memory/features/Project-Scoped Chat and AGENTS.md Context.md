---
title: Project-Scoped Chat and AGENTS.md Context
type: note
permalink: scarf/features/project-scoped-chat-and-agents.md-context
tags:
- projects
- chat
- acp
- agents-md
source_sha: 8d2293330e574b9e3b4ff42f6fcd155af248ab59
source_paths: scarf/scarf/Core/Services/SessionAttributionService.swift, scarf/scarf/Core/Services/ProjectAgentContextService.swift, scarf/scarf/Features/Projects/Views/ProjectSessionsView.swift
---

## Observations
- [feature] v2.3 adds per-project Sessions tab + 'New Chat' button that spawns `hermes acp` with cwd=project.path. ProjectSessionsView lives at Features/Projects/Views/ #projects
- [attribution-sidecar] Session-to-project attribution is persisted in a Scarf-owned sidecar at ~/.hermes/scarf/session_project_map.json. ACP wire protocol has NO project-metadata hook (extra params silently dropped); state.db has NO cwd column. Sidecar is Scarf's source of truth. Managed by SessionAttributionService.swift #attribution
- [context-mechanism] Hermes auto-reads a context file from the session's cwd at startup with priority order: .hermes.md → HERMES.md → AGENTS.md → CLAUDE.md → .cursorrules (first match wins, 20KB cap). Scarf writes a managed block into <project>/AGENTS.md before opening the session via ProjectAgentContextService.swift #mechanism
- [block-shape] Block delimited by `<!-- scarf-project:begin -->` and `<!-- scarf-project:end -->` markers; includes project name, dir, dashboard path, template (if installed), config field NAMES, registered cron jobs, uninstall manifest path, Kanban tenant (if minted). Anything outside markers is preserved #format
- [invariant] SECRET-SAFE: block surfaces field NAMES, never VALUES. Secret fields render as `field_name (secret — name only, value stored in Keychain)`. Keychain ref URI and plaintext value never appear. Auditable by `refreshListsFieldNamesNotValues` in ProjectAgentContextServiceTests #security
- [invariant] IDEMPOTENT: two refreshes with unchanged state produce byte-identical output. Write skipped entirely when no delta, avoiding file-watcher churn #correctness
- [invariant] BOUNDED: everything outside markers is preserved on every refresh. Template-author AGENTS.md content lives safely below the block #correctness
- [invariant] NON-FATAL: ChatViewModel.startACPSession calls refresh with `try?` + log. Failed write doesn't block chat from starting; worst case is session loses project awareness #resilience
- [ordering-rule] Refresh MUST be called BEFORE client.start() so the block lands before Hermes's session-boot context scan. Skipping this ordering = agent sees stale context from previous refresh, or nothing on fresh projects #pitfalls
- [template-contract] Catalog templates should include an AGENTS.md with operational instructions. Authors leave the scarf-project region alone — Scarf populates it at chat-start time. Everything below is template-owned and preserved #templates
- [known-caveat] If any PARENT directory of the project contains .hermes.md or HERMES.md, those SHADOW the project's AGENTS.md (higher in priority order). No fix in v2.3 — deferred pending input on handling authored .hermes.md files #caveats

## Relations
- relates_to [[Project Templates (.scarftemplate)]]
- relates_to [[Kanban Board Architecture (v2.7.5)]]
- relates_to [[Model Presets Feature]]