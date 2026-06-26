---
title: Core Engineering Constraints
type: note
permalink: scarf/architecture/core-engineering-constraints
tags:
- rules
- swift
- dependencies
source_sha: 427321d742d63298100f9e444f96fd1524d7a46c
source_paths: CLAUDE.md, README.md, scarf/scarf.xcodeproj/project.pbxproj
---

## Observations
- [rule] No external dependencies for core: system SQLite3, Foundation JSON, AttributedString markdown. Sparkle (updates) and SwiftTerm (terminal/QR scan) are the explicit exceptions; ScarfGo additionally uses Citadel for iOS SSH. #dependencies
- [rule] Read-only DB access: never write to `~/.hermes/state.db`. The only files Scarf writes are memory files (MEMORY.md, USER.md, SOUL.md), config.yaml fragments, and cron jobs. #safety
- [rule] App sandbox is disabled — Scarf reads `~/.hermes/` directly. (See Hermes Integration note.) #sandbox
- [rule] Swift 6 concurrency: `@MainActor` is the default. Services use `nonisolated` + async/await. #concurrency
- [rule] AppCoordinator is a single `@Observable` coordinator owning all navigation state, injected via `.environment()`. #state
- [rule] MVVM-F isolation: features never import sibling features. Cross-feature communication goes through services in `Core/Services/`. #mvvm-f
- [platform] Build targets: macOS 14.6+ (Sonoma) for Scarf, iOS 18.0+ for ScarfGo, Swift 6, Xcode 16.0+. #platform

## Relations
- extends [[Scarf Architecture Rules]]
- relates_to [[Scarf Project Layout]]