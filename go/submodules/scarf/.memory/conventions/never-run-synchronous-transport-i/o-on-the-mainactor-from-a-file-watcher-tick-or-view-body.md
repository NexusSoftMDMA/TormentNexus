---
title: Never run synchronous transport I/O on the MainActor from a file-watcher tick or view body
type: note
permalink: scarf/conventions/never-run-synchronous-transport-i/o-on-the-mainactor-from-a-file-watcher-tick-or-view-body
---

## Observations
- [gotcha] On a REMOTE (SSH) context, these all do SYNCHRONOUS scp/SSH round-trips: `HermesFileService.readFile/loadConfig/loadGatewayState`, `ServerContext.runHermes/readText/readData`, `HermesEnvService.load()`. Calling any of them on the MainActor from a HOT/REPEATED path — a `.onChange(of: fileWatcher.lastChangeDate)` handler (fires per persisted message during an ACP stream) or a view `body`/computed property — stalls the main thread → typing lag / UI jank. This was the gh#102 *typing-lag* follow-up (distinct from the original gh#102 Dashboard FSEvent close/reopen fix, v2.10.3). #gotcha #perf #remote
- [pattern] Fix: do the read off-main (`Task.detached`, or mark the reader `nonisolated`) and commit `@Observable` state back on the MainActor. For watcher-driven loads ALSO add cancel-prior + a recency guard: store the `Task`, `cancel()` the prior on each tick, `if Task.isCancelled { return }` before committing, and advance any freshness/change token ONLY on a committed read. The synchronous loads these replace couldn't interleave, so a naive async port introduces out-of-order-completion stale-clobber (a fresh-eyes audit caught exactly this). Mirrors `ChatViewModel.loadRecentSessions` (inFlightSessionLoad coalescing) + `HealthViewModel` (t-aud11) + `PlatformsViewModel.load`. #pattern
- [render-hot] Worst variant = synchronous remote reads in a view `body`/computed (e.g. `PlatformsViewModel.connectivity` → `hasConfigBlock` read config.yaml + .env per platform per render). Fix = compute once off-main in `load()`, cache the result (e.g. `configuredPlatforms: Set<String>`), body reads the cache. NOTE `context.readText` is NOT cached — it's `makeTransport().fileExists` + `readFile` (two live round-trips) every call. #render
- [sweep] 2026-06-21 sweep of every app `.onChange(fileWatcher.lastChangeDate)` handler: Gateway/Cron/Memory/CredentialPools/Sessions/Activity/Insights/Dashboard/RichChat were already off-main/debounced; Chat/Platforms/Projects were not — fixed on branch `fix/remote-sync-io-on-main`. When adding a watcher-driven load, follow the off-main + cancel-prior pattern.

## Relations
- relates_to [[Scarf Architecture Rules]]
- relates_to [[Hermes Integration]]
- relates_to [[macOS must mirror iOS scene-phase pause and resume for background work]]



## Audit-found refinements (2026-06-21)
- [sweep-gotcha] Checking "is load() async?" is NOT enough: SessionsViewModel.load() was async but called a synchronous computeStats() that did a stat() on main. The first sweep cleared it; a second audit caught it. Trace INTO the async load body for nested synchronous transport calls (stat/readText/readFile/runHermes), not just the method signature. So Sessions belongs in the "was NOT off-main" set alongside Chat/Platforms/Projects.
- [recency-pattern] Across MULTIPLE suspension points, Task.isCancelled only orders the FIRST commit. If a load writes observable state both before AND after a later await (e.g. registry then dashboard), use a monotonic generation token: gen += 1 at start, capture it, and guard gen == currentGen before EACH commit. isCancelled cannot order a write that sits behind a second suspension. See ProjectsViewModel.reload/reloadDashboard.
