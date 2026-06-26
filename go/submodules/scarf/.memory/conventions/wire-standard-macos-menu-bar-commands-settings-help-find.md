---
title: Wire standard macOS menu-bar commands (Settings, Help, Find)
type: note
permalink: scarf/conventions/wire-standard-macos-menu-bar-commands-settings-help-find
tags:
- hig
- macos
- conventions
- audit-2026-06-13
---

## Observations
- [rule] 🚨 The macOS `.commands` block must provide the menu-bar affordances users reach for by reflex: Settings via ⌘, (route to `coordinator.selectedSection = .settings`), a Help menu (`CommandGroup(replacing: .help)`) linking to docs, and ⌘F to focus the search field in any searchable list (`@FocusState` + `.keyboardShortcut("f")`). Sidebar-only or click-only access does not satisfy macOS keyboard conventions. #rule
- [pattern] Today the `.commands` block defines only two custom CommandGroups (Check for Updates, Open Server). 114 `.help()` tooltips partially offset the missing Help menu but are not a substitute.
- [check] Read `scarfApp.swift` `.commands { }`; confirm Settings / Help / Find affordances exist.
- [history] 2026-06-13 Cycle 3: `scarfApp.swift:207-218` (no Settings ⌘, no Help menu), `SessionsView.swift:253-280` (no ⌘F to focus search). #history

## Relations
- relates_to [[Scarf Architecture Rules]]
