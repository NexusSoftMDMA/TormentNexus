---
title: ScarfCore tests inject a temp Hermes home via ServerContext.local(home:)
type: note
permalink: scarf/conventions/scarfcore-tests-inject-a-temp-hermes-home-via-servercontext-local-home
tags:
- testing
- scarfcore
- isolation
- servercontext
- swift6
---

ScarfCore unit tests that exercise local-home file I/O MUST point at an isolated temp dir, never the developer's real `~/.hermes`. The seam (added in t-aud25) is a per-instance injectable home on `ServerContext`.

## Observations

- [seam] `ServerContext` has `public private(set) var localHomeOverride: String?` (nil in production). `paths` consults it ONLY for `kind == .local`; production resolves the real `~/.hermes` via `HermesPathSet.defaultLocalHome` unchanged. #servercontext
- [factory] Construct a temp-home context with `ServerContext.local(home: URL)`. It PRESERVES `localID`, so `vm.context.id == ServerContext.local.id` assertions still hold — the only difference from `.local` is `paths.home`. #servercontext
- [pattern] Per-test helper: make a `FileManager.temporaryDirectory/<uuid>` home, `defer { try? FileManager.default.removeItem }`, pass `ServerContext.local(home:)` to the service/VM. `LocalTransport` writes to the real filesystem at that temp path, so reads/writes stay sandboxed. #pattern
- [why-per-instance] Deliberately per-instance, NOT a process-global. The global `SCARF_HERMES_HOME` env override (in `HermesProfileResolver`, gated by a `.scarf-test-home-marker` sentinel) is for SERIAL app-hosted E2E harnesses (`TemplateE2ETests`, `TemplateInstallUITests`) only — it RACES parallel Swift-Testing suites. Do not reach for it in ScarfCore unit tests. #swift6 #isolation
- [remote-tilde] For `.ssh` contexts whose `paths.home` is the unexpanded `~/.hermes`, the seam does not apply (it's `.local`-only). Drive shell `~`/`$HOME` expansion by exporting `HOME` to the subprocess the test transport spawns — see `LocalSQLite3Transport.homeOverride` in `RemoteSQLiteBackendTests`. #pattern
- [gotcha] A non-injected `.local` test reads whatever `~/.hermes` exists on the runner: machine-dependent skips, real-data move/backup hacks, and `.serialized` suites are the symptoms this seam removes. Full `swift test` is 610/610 with zero real-`~/.hermes` contact after t-aud25. #testing

## Relations

- relates_to [[Model Presets Feature]]
- relates_to [[Scarf Architecture Rules]]
- relates_to [[Multi-Server Architecture (Scarf 2.0+)]]
