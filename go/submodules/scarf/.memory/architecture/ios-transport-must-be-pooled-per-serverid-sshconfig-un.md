---
title: iOS transport must be pooled per (ServerID, SSHConfig) — un-pooled makeTransport churns SSH connections
type: note
permalink: scarf/architecture/ios-transport-must-be-pooled-per-serverid-sshconfig-un
created: 2026-06-25
updated: 2026-06-25
tags:
- ios
- ssh
- scarfgo
- transport
- gh112
- performance
---

## Observations

- [root-cause] `ServerContext.makeTransport()` returns a FRESH value per call (ServerContext.swift:163). The iOS `sshTransportFactory` (ScarfIOSApp.swift) news up a new `CitadelServerTransport` each time, each with its OWN `ConnectionHolder` (`sshClient = nil`). So pre-fix, EVERY iOS file read / `runProcess` paid a brand-new `SSHClient.connect()` handshake, and close-on-dealloc was fire-and-forget. A Settings load or chat pre-flight fans out a burst of short-lived handshakes; under the churn `SSHClient.connect()` starts to FAIL. gh#112; reproduced across two unrelated hosts; a single manual `ssh` always works (one connection). #gotcha #ios #ssh

- [diagnosis-key] On iOS, `CitadelServerTransport.runProcess` THROWS only at `connectionHolder.ssh()` → `SSHClient.connect()` (CitadelServerTransport.swift:457). EVERY post-connect failure (channel refused, non-zero exit, mid-stream drop) returns a `ProcessResult` with an exit code, never throws. So the chat banner "Transport refused the command" (the `try?`→nil branch in ChatView.preflightFailureMessage) means THE SSH HANDSHAKE FAILED, not a command rejection. Reads hide the same failure: `fileExists` is `try?`→false (CitadelServerTransport.swift:118) → `readTextThrowing` returns nil = "file absent" → UI shows empty / "no model configured" while the file exists. That's why Platforms (cached after one good read) populate while live model/source reads read empty. #gotcha [[never-run-synchronous-transport-i/o-on-the-mainactor-from-a-file-watcher-tick-or-view-body]]

- [fix] `CitadelTransportPool.shared` (ScarfIOS) memoizes ONE transport per `(ServerID, SSHConfig)` behind the `sshTransportFactory`. Key = `ServerID`; the stored `SSHConfig` (Hashable) is the staleness check. A #120 profile switch changes `SSHConfig.remoteHome` → config differs → pool closes the old connection and opens one for the new profile. Churn we killed was per-OPERATION; replacement is per-SWITCH (rare, deliberate) — bounded either way. NSLock-guarded (`withLock`); `make` runs under the lock (construct-only, connection opens lazily). Mirrors `UserHomeCache.shared`. #pattern

- [lifecycle] Eviction (close + drop) wired into the existing seams: `RootModel.softDisconnect`/`forget(id)`/`disconnect` and `ScarfGoCoordinator.setScenePhase(.background)`. The Mac app never had the churn — `SSHTransport` shells out to `/usr/bin/ssh` with ControlMaster multiplexing; the pool is the iOS equivalent. #pattern

- [boundary] The ACP CHAT channel owns its OWN SSH connection (`CitadelSSHService` / `SSHExecACPChannel`, NOT `makeTransport`) and its own scene-phase lifecycle (`ChatController.handleScenePhase`). The pool does not touch it. #boundary

- [residual-risk] (1) The shared `SFTPClient` now sees wider concurrency. This sharing already existed WITHIN a transport instance (e.g. `watchPaths` concurrent `stat`); pooling widens it across callers. Citadel multiplexes SFTP by request-id, so low risk — but verify on a live host. (2) A superseded/evicted transport still referenced by a torn-down view could reconnect ONCE (orphan) before it deallocs; bounded by profile-switch frequency, self-healing, NOT per-op churn. If it ever bites, add an `isShutdown` flag to `ConnectionHolder` so `ssh()` refuses to reconnect after `closeIfOpen()`. #risk

- [test-enablement] ScarfIOS package now compiles + `swift test`s on macOS (it only indexed before): added a `#if !os(iOS)` `makeProcess` stub to `CitadelServerTransport` (the ONLY non-iOS `ServerTransport` requirement) and bumped `Package.swift` `.macOS(.v14)`→`.v15` (iOS sources call the macOS-15 `withExec(_:environment:perform:)`). macOS is index/test-only for ScarfIOS and is NOT linked into the Mac `scarf` app (only "scarf mobile" links ScarfIOS), so zero shipping impact. Run: `swift test --package-path scarf/Packages/ScarfIOS`. #ops

## Relations
- relates_to [[ScarfGo iOS Companion App]]
- relates_to [[Multi-Server Architecture (Scarf 2.0+)]]
- relates_to [[decision-scarfgo-profile-switching-via-per-connection]]
- relates_to [[never-run-synchronous-transport-i/o-on-the-mainactor-from-a-file-watcher-tick-or-view-body]]


## Verification + second-half fix (2026-06-25, live)

- [critical] Pooling the transport is NECESSARY BUT NOT SUFFICIENT. `ConnectionHolder.ssh()`/`sftp()` had an actor-REENTRANCY race: concurrent first-callers all suspend at `await openSSH()` before any sets `sshClient`, so each opens its OWN connection. A simultaneous cold burst (Settings firing parallel reads) therefore still churned even with one pooled transport. Fixed by COALESCING: store the in-flight open as a `Task<SSHClient,Error>` (`connectTask` / `sftpTask`), published with NO `await` between the nil-check and the assignment, so concurrent callers join the one task instead of racing. #gotcha #fix
- [verified] Live integration test (`CitadelTransportPoolLiveTests` + `scripts/verify-ios-transport-pool.sh`) against an ephemeral localhost sshd (`MaxStartups 2:80:4`, throwaway HERMES_HOME): the exact gh#112 chat-init sequence (version + two `config set` + SFTP readback) passes through the pool, and a 24-way concurrent burst goes from **20/24 connect failures un-pooled (fresh-per-op) to 0/24 pooled+coalesced** (one connection). The unit-test pool-dedup alone did NOT catch the holder race — the live test did. #verification
- [hermes-fact] `hermes config get` does NOT exist (v0.16 `config` subcommands: show / edit / set / path / env-path / check / migrate). [[t-ios-cfg-get]] (F2) proposed routing Settings reads through `hermes config get`. Corrected: the grounded read path is `cat "$(hermes config path)"` through the wrapper — `config show` is human-only (no `--json`, no key arg, Model line is a Python-dict repr) and a poor structured-read substitute. Re-verify against the live target version (reporter was v0.17). #correction

