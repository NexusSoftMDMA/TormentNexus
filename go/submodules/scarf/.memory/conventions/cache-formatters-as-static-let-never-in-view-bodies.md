---
title: Cache formatters as static let, never in view bodies
type: note
permalink: scarf/conventions/cache-formatters-as-static-let-never-in-view-bodies
tags:
- performance
- swiftui
- conventions
- rule
- audit-2026-06-13
---

## Observations
- [rule] 🚨 Never allocate `DateFormatter`, `ISO8601DateFormatter`, `NumberFormatter`, `RelativeDateTimeFormatter`, or `NSRegularExpression` inside a SwiftUI `body`, computed property, or per-row helper — SwiftUI re-evaluates these per render and per-row state change (e.g. hover), creating locale-aware allocations on hot paths. Declare one `private static let` per format. #rule
- [pattern] Reference implementation: `KanbanCardView.swift:407-411` (`relativeFormatter`). Swift's `FormatStyle` API is internally cached and also acceptable.
- [check] Quick audit: `grep -rn 'Formatter()' --include="*.swift" scarf/scarf "scarf/Scarf iOS"`
- [history] 2026-06-13 Cycle 2: `KanbanCardView.swift:395-405` (per-card per-render in a LazyVStack re-rendered on the 10s poll), `SessionsView.swift:604-606` (per-row, re-fired by hover @State), `ProjectSessionsView.swift:186-189`, `ModelPickerSheet.swift:415-417`. #history

## Relations
- relates_to [[Scarf Architecture Rules]]
