---
id: t-dfc8c0e0
title: B4: docs/wiki/memory sync + fresh-eyes integration audit for #120
status: done
added: 2026-06-25
---

## Description

Parent: t-873f7df9 (#120 Design B). Final integration audit across the whole feature + bring knowledge base and user docs up to date once behavior changes.

## Plan

## Plan
1. Update memory: ScarfGo iOS Companion App note (profiles no longer read-only on iOS; describe per-connection scoping); ensure the verified Hermes HERMES_HOME/-p facts + Design-B decision notes are present and linked.
2. Update wiki (ScarfGo / Profiles page) + README "What's New" entry (release-time) describing phone profile scoping.
3. Full fresh-eyes integration audit (ideally a fresh subagent): end-to-end switch flow, all surfaces consistent (dashboard/memory/cron/sessions/gateway/chat), no host active_profile mutation, no stale state across switches, tests are real (not checkbox).
4. Address audit findings; final verification build.

## Commit
- `docs(ios,profiles): document per-connection profile scoping; memory/wiki sync (#120)`

## Artifacts

Integration audit (fresh subagent, whole-feature diff 8cc64e6..HEAD): VERDICT no Critical, NO unscoped surface. Traced all 16 toServerContext sites + all makeTransport/forIOSApp paths — every iOS surface receives the profile-scoped effectiveConfig (file layer) and routes hermes through asyncRunProcess/ACPClient+iOS (process layer). active_profile is NEVER mutated (only read; only `profile list` runs, read-only). Persistence reloads synchronously → no wrong-profile flash. Capability store correctly host-level (all flags version-gated, not config.yaml). 667 ScarfCore tests pass.

Docs/memory synced:
- Memory: updated "ScarfGo iOS Companion App" note with the #120 profiles capability; Design-B decision + Hermes HERMES_HOME-resolution notes were written up-front in planning.
- Wiki: Projects-and-Profiles.md — added "ScarfGo (iOS) profile switching (#120)" subsection; clarified the manual-remoteHome note is Mac-only.
- README "What's New": deferred to release prep (scarf-release-prep skill; maintainer-controlled).
- New iOS UI strings: English literals pending the i18n pass (release-time).

Deferred follow-ups (non-blocking, cosmetic):
- Skills "What's New" snapshot baseline keyed by serverID only → bogus diff pill after a switch. Spun off as background task task_b4d0ce8d (per-server+per-profile keying).
- Reconnect success tail (ChatView ~2069) lacks a Task.isCancelled check after awaits → rare spurious connect→stop during a switch; largely covered by [weak self] + idempotent stop(). Low severity, pre-existing.

Manual checklist for a real host (no app-target test scheme exists; lifecycle is untestable in unit tests): switch mid-stream → old hermes acp dies + new spawns scoped; host active_profile unchanged after switch (cat before/after); zero-profile host shows footer not error; all surfaces move together; VoiceOver pass.

