---
id: t-fdf7da17
title: B1: wire selected profile into ServerContext; rebuild tab subtree on change
status: done
added: 2026-06-25
priority: high
---

## Description

Parent: t-873f7df9 (#120 Design B). Makes all direct-file/DB iOS surfaces (dashboard, memory, cron, sessions, gateway_state, scarf/) follow the selected profile via the existing remoteHome→paths.home seam.

## Plan

## Plan
1. In ScarfGoTabRoot (and any other place building context from `config`), compute resolved remoteHome from B0 selection BEFORE `config.toServerContext(id:)`; inject into the SSHConfig/remoteHome.
2. Key the tab subtree on the selected profile (`.id(selectedProfile)`) so a change tears down + rebuilds VMs/capability store cleanly (the iOS analogue of Mac's relaunch-to-flush-state).
3. Ensure capability store + per-tab context ids rebuild with the new context.

## Tests
- With a selection set, `context.paths.stateDB/memoriesDir/...` resolve under `profiles/<name>`.
- Changing selection changes the derived paths; default selection → base paths.

## Audit
- Fresh-eyes: no stale context captured by long-lived VMs; soft-disconnect/caches keyed by serverID not polluted across profiles; UserHomeCache (keyed by serverID) still correct (home is $HOME, not profile — unaffected).

## Commit
- `feat(ios,profiles): scope direct-file reads to selected profile via remoteHome (#120)`

## Artifacts

Commit 084731f.
- ScarfGoCoordinator: owns per-server selectedProfile (loads from injected store; setSelectedProfile normalizes+dedupes+persists). serverID now required (no default — avoids cross-server bleed).
- ScarfGoTabRoot: effectiveConfig (remoteHome → profile via HermesProfileScope) threaded to all 5 tabs + SystemTab; `.id(selectedProfile)` rebuilds the tab subtree on switch; lifecycle modifiers on stable wrapper (profileScopedTabs extraction).
- HermesProfileScope: double-slash guard for "/" base.

Tests: ScarfCore swift-test 13 pass incl. new end-to-end "profile-scoped remoteHome drives every HermesPathSet path" + default-stays-root. iOS app BUILD SUCCEEDED on iOS 26.5 simulator (real runtime installed mid-phase).

Fresh-eyes audit (subagent): no Critical. Fixed now — required serverID (#5), double-slash "/" base (#6, +test), Curator scoping-rides-on-paths comment (#4), UserDefaults RMW main-actor note (#7), store valid→invalid-clears test (#8). DEFERRED to B3 (ProfilesView rework): badge/header must show the SELECTED profile, not host active_profile (#2); read active_profile from the BASE/root home, not the scoped home (#3). Plumbing is inert until B3 wires the UI (#1) — noted in commit.

Note: app-target unit tests have no runnable scheme in this project (only the Mac app has a shared test scheme; iOS coverage lives in the SwiftPM packages). B1 logic is covered by ScarfCore tests + iOS compile; coordinator glue is thin.

