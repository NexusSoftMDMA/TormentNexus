---
id: t-1fef0a83
title: iOS: pool CitadelServerTransport per server (fix gh#112 chat-init churn)
status: done
added: 2026-06-25
priority: high
---

## Description

gh#112 / [[t-2c5982]] root cause. ScarfGo chat-init fails with "Couldn't save model.provider … Transport refused the command" AND Settings shows model/source empty while Platforms populate, on hosts where the same `hermes config set` works by hand and reproduces across two unrelated hosts.

PROVEN from code:
- "Transport refused the command" (ChatView.swift:1445) only fires when `runProcess` THROWS, which on iOS happens ONLY at `connectionHolder.ssh()` → `SSHClient.connect()` (CitadelServerTransport.swift:457). Every post-connect failure returns a ProcessResult (exit code), never throws. So the symptom = the SSH handshake itself failed, NOT a command rejection.
- `ServerContext.makeTransport()` returns a FRESH `CitadelServerTransport` per call (ServerContext.swift:163); the iOS factory (ScarfIOSApp.swift:35) news one up each time with its own `ConnectionHolder` (sshClient=nil). So every file read / exec opens a brand-new SSH handshake; close-on-dealloc is fire-and-forget. Settings-load + chat-init churn many short-lived handshakes → connect() starts failing.
- Reads swallow transport failure to "empty": `fileExists` try?→false (CitadelServerTransport.swift:118) → `readTextThrowing` returns nil = "file absent" (ServerContext.swift:330). Why "Pick a model"/"View source" read empty while Platforms (cached after one good read) populate.
- The factory comment (ScarfIOSApp.swift:31-34) FALSELY claims "not a new SSH handshake".

## Plan

Design B-compatible: profile scoping is per-OP (HERMES_HOME from config.remoteHome at asyncRunProcess:476 + SFTP remoteHome paths), so pool key = ServerID with stored SSHConfig (Hashable) as staleness check. Profile switch → remoteHome changes → config differs → pool replaces (close old conn, open new). Bounded per server/profile, not per op.

Steps:
1. New `CitadelTransportPool.shared` (ScarfIOS) — NSLock-guarded [ServerID: Entry(config, transport)]; `transport(for:config:make:)` get-or-create + replace-on-config-change (close superseded off-thread); `evict(id)` + `evictAll()` async close. Mirrors UserHomeCache.shared / ResultBox NSLock pattern.
2. Wire iOS `sshTransportFactory` through the pool; fix the false "not a new SSH handshake" comment.
3. Eviction: RootModel.softDisconnect (evict id), forget(id) (evict id), disconnect (evictAll); ScarfGoCoordinator.setScenePhase(.background) → evictAll (ACP chat channel handles its own scene phase separately, so safe).
4. Diagnostic: chat preflight catches the thrown connect error and passes it into preflightFailureMessage so the else-branch shows the real reason instead of a generic line.

Test (real, not checkbox): pool reuse (same instance for same id+config), replace-on-config-change (new instance + old closed), evict closes, concurrency (N parallel transport(for:) calls → one instance, no race) — via injected counting make-closure, no live SSH. 
Audit: fresh-eyes adversarial pass incl. profile-switch transition races + SFTP-sharing concurrency.

Out of scope (separate): F2/[[t-ios-cfg-get]] Docker read-path via `hermes config get`. Pooling fixes host-1 fully + host-2 write churn.

## Artifacts

Implemented + live-verified + audited (not pushed). Fix is TWO parts:

1. CitadelTransportPool (commit 1324ed4) — one transport per (ServerID, SSHConfig) behind sshTransportFactory; evict on softDisconnect/forget/disconnect + scene-phase background.
2. ConnectionHolder open-coalescing (follow-up commit) — `ssh()`/`sftp()` had an actor-reentrancy race: concurrent first-callers all passed the `sshClient==nil` check during `await openSSH()` and each opened a connection, so a simultaneous cold burst (Settings parallel reads) still churned even with one pooled transport. Fixed with in-flight `Task<SSHClient,Error>`/`Task<SFTPClient,Error>` coalescing. THE LIVE TEST CAUGHT THIS — pooling alone would have shipped incomplete.

Plus diagnostic: chat preflight surfaces the real thrown connect error instead of generic "Transport refused".

LIVE VERIFICATION (scripts/verify-ios-transport-pool.sh + CitadelTransportPoolLiveTests, ephemeral localhost sshd MaxStartups 2:80:4, throwaway HERMES_HOME):
- gh#112 chat-init sequence (version + 2 config set + SFTP readback) PASSES through the pool, value isolated to throwaway home (real ~/.hermes untouched).
- 24-way concurrent burst: un-pooled 20/24 connect failures → pooled+coalesced 0/24.

Tests: ScarfIOS suite 17/17 (15 unit + 2 gated live); iOS "scarf mobile" build SUCCEEDED. Mac app unaffected (doesn't link ScarfIOS).

Side finding: `hermes config get` doesn't exist (v0.16) — F2 [[t-ios-cfg-get]] must use `config show`/file read, not `config get`.

Remaining gate: real reporter-host / TestFlight confirmation (optional — local live test reproduces the mechanism conclusively).

