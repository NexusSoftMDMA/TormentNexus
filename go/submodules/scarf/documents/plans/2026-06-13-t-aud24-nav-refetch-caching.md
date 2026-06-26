# Plan — t-aud24: Eliminate redundant remote re-fetch on sidebar section switches

> Status: NOT STARTED · Created 2026-06-13 (from the Swift audit follow-up loop) ·
> Owner: next session · Risk: HIGH (navigation + 9 features) · Source: t-aud12 → t-aud24

## Problem

`ContentView` renders the detail pane with a `@ViewBuilder switch` on
`coordinator.selectedSection` (`scarf/scarf/Navigation/ContentView.swift:50-84`).
A `switch` in a `@ViewBuilder` produces a *different view type per case*, so on every
sidebar selection change SwiftUI **tears down the previous section's view and builds the
new one from scratch**. Each feature view owns its view model as `@State` (recreated on
rebuild) and kicks off `load()` from `.onAppear`/`.task`. Result: switching away from a
section and back **re-runs that VM's multi-call remote `load()` over SSH** every time —
several round-trips per re-entry on a remote server.

This was originally filed as "swap `.onAppear` → `.task`," but that is a **no-op here**:
because the views are destroyed+recreated per switch, `.task` fires on every switch just
like `.onAppear` (confirmed by the Cycle-2 audit verifier). The real fix is structural.

## Goals

1. Re-entering a section that was loaded recently does **not** re-fetch over SSH.
2. The 8 feature VMs that still lack it get the **cancellable-load** treatment
   (the t-aud11 pattern: store the `Task` handle, cancel on disappear, check
   `Task.isCancelled` between SSH round-trips). Health (t-aud11) is the reference.

## Approach — evaluate in this order

**Option C (preferred if memory is acceptable): stop destroying the views.**
Replace the `@ViewBuilder switch` with a container that keeps section views alive and
toggles visibility (e.g. a `ZStack` with `.opacity`/`allowsHitTesting` gated on
`selectedSection`, or `.id`-stable conditional rendering that preserves identity). Views +
`@State` + scroll position persist; `.task` fires once per view instance, so no re-fetch.
- Pro: minimal per-feature code; also fixes lost scroll/edit state.
- Con: all ~14 sections instantiated and held simultaneously. Audit the heavy ones
  (Chat, Dashboard, Kanban) for memory/observer cost; some may need lazy-on-first-show
  (build on first visit, then keep). Watch always-running observers (the gh#102 poller,
  HermesFileWatcher) so off-screen sections don't keep working — pair with the
  scene-phase work in [[macOS must mirror iOS scene-phase pause and resume for background work]].

**Option A/B (fallback): persist the data, not the views.**
Move each feature VM (or just its loaded payload + a freshness timestamp) into a
longer-lived owner — `AppCoordinator` (per-window) or a small section-cache keyed by
`SidebarSection`. On `load()`, return early when cached data is fresh (e.g. < N seconds or
no `fileWatcher` change since). Views stay destroy/recreate but read from the cache.
- Pro: bounded memory; no view-lifetime change.
- Con: per-feature cache plumbing; 9 VMs to touch.

Recommendation: prototype Option C behind the existing sidebar first; measure memory with
Instruments. If it regresses, fall back to coordinator-owned VM payloads (A/B).

## Affected files

- `scarf/scarf/Navigation/ContentView.swift` — the section switch (core change).
- `scarf/scarf/Navigation/AppCoordinator.swift` — if caching/VM ownership moves here.
- The 9 feature views + VMs: Settings, Health (done), Platforms, Plugins, QuickCommands,
  Models, MCPServers, Webhooks, Cron — each `…View.swift` + `…ViewModel.swift`.
- Reference patterns: `HealthViewModel.load()`/`cancelLoad()` (t-aud11),
  `DashboardViewModel.inFlightLoad` (overlapping-load guard).

## Step-by-step

1. Baseline: instrument section-switch SSH cost (ScarfMon `…load` events / `log stream`)
   to get a before/after number. Note current memory footprint.
2. Implement Option C in `ContentView` for 2-3 sections behind a flag; verify no re-fetch
   on switch + state preserved; measure memory.
3. If memory OK → roll out to all sections; else pivot to A/B (coordinator cache).
4. Apply the t-aud11 cancellable-load treatment to the 8 remaining feature VMs.
5. Ensure off-screen sections pause their observers/polling (tie to scene-phase note).

## Risks / gotchas

- Navigation regressions: wrong section shown, `@State` bleed between sections, deep-link
  paths (e.g. `coordinator.selectedSection = .settings` from t-aud06's ⌘,) must still work.
- Memory blow-up under Option C (Chat/Kanban are heavy).
- Cancellation correctness (the t-aud11 lesson: sync SSH work doesn't observe
  `Task.cancel` unless you check `isCancelled` between round-trips).
- High blast radius → test each feature screen manually after.

## Acceptance / verification

- Switch A→B→A: section A does NOT re-run its remote `load()` (verify via SSH-call
  counters / logs); state + scroll preserved (Option C).
- All 9 feature VMs cancel in-flight loads on switch-away.
- macOS build clean; manual pass on Chat / Dashboard / Settings / Health / Kanban / the
  rest; Instruments shows no memory regression.
- ⌘, (Settings deep-link), sidebar selection, and window restoration still behave.
