---
id: t-aud32
title: **[flaky-test · NEW]** `RemoteSQLiteBackendTests.openWithDefaultTildeHomeExpands` (un-skipped in t-aud25) flakes ~1-in-many under the full parallel `swift test`: `backend.open()` returns false with `lastOpenError = "unable to open database \"2026-04-22 12:00:00,001 INFO hermes.agent: starting…\""` — i.e. sqlite3 received a HermesLogService LOG FIXTURE (`M5FeatureVMTests.swift:468`) as its DB-path/preflight output. PASSES 5/5 in isolation → it's a cross-suite race: concurrent `Foundation.Process` spawns crossing stdout/stderr under heavy parallel load (the test transport's `streamScript` pipe read another suite's subprocess output). NOT a production bug. Fix options: serialize the subprocess-spawning DB tests, or harden `LocalSQLite3Transport.streamScript`'s pipe handling. Risk: LOW (test-infra only).
status: todo
added: 2026-06-13, source: t-aud31 4x run
---

## Description



## Plan



## Artifacts



