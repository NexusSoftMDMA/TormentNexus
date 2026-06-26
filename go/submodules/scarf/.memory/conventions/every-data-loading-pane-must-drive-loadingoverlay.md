---
title: Every data-loading pane must drive .loadingOverlay()
type: note
permalink: scarf/conventions/every-data-loading-pane-must-drive-loadingoverlay
tags:
- hig
- swiftui
- ux
- conventions
- audit-2026-06-13
---

## Observations
- [rule] 🚨 When a view loads data asynchronously (any `.task`/`.onChange` calling a ViewModel `load()`), the ViewModel must expose `isLoading: Bool` and the view must apply `.loadingOverlay(viewModel.isLoading, label: "Loading …", isEmpty: viewModel.<collection>.isEmpty)`. Treat the overlay as a required part of a data pane, not optional polish. #rule
- [pattern] `.loadingOverlay()` lives in `LoadingOverlay.swift` and is used correctly across Dashboard, Activity, Settings, Memory, Plugins, Health, CredentialPools, Cron, MCPServers. Panes that skip it leave users on blank/stale content during SSH fetches; some VMs (Logs, Sessions) don't even expose `isLoading`.
- [check] For each `Features/*/Views/*View.swift` with `.task { await viewModel.load() }`, confirm a matching `.loadingOverlay(` and an `isLoading` on the VM.
- [history] 2026-06-13 Cycle 3: `InsightsView.swift:21-47` (high — stale on period change), `SessionsView.swift:43-77` (VM lacks isLoading), `LogsView.swift:17-28` (VM lacks isLoading). #history

## Relations
- relates_to [[Scarf Design System (ScarfDesign)]]
