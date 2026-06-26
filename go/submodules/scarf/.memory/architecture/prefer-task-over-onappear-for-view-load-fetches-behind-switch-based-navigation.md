---
title: Prefer .task over .onAppear for view-load fetches behind switch-based navigation
type: note
permalink: scarf/architecture/prefer-task-over-onappear-for-view-load-fetches-behind-switch-based-navigation
tags:
- performance
- swiftui
- navigation
- architecture
- audit-2026-06-13
---

## Observations
- [rule] 🚨 Navigation here uses `@ViewBuilder switch` on a selected-section/-tab enum that DESTROYS and recreates subtrees per selection, so `.onAppear { load() }` re-fires multi-call remote fetches on every re-entry. Use `.task` (fires once per view instance, auto-cancels on disappear). #rule
- [pattern] `SessionsView`/`ChatView` use `.task` correctly. For true state persistence across switches, hoist the view or cache the loaded data in the coordinator — `.task` alone reduces redundant fetch frequency but does not preserve state across destruction.
- [check] Quick audit: `grep -rn '.onAppear {' --include="*.swift" scarf/scarf/Features | grep -i load`
- [history] 2026-06-13 Cycle 2: `ContentView.swift:50-84` + 9 feature views (e.g. `SettingsView.swift:108`, `HealthView.swift:124-127`), `ProjectsView.swift:454-481` (project tab switch). #history

## Relations
- relates_to [[Scarf Architecture Rules]]
- relates_to [[macOS must mirror iOS scene-phase pause and resume for background work]]
