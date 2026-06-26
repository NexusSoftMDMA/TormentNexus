# Scarf v2.10.3

A reliability and polish release. The headline is a fix for the **100% single-core CPU spin** that hit users with large `state.db` files, plus a broad performance + correctness sweep from a full code audit. Most of the work is under the hood — every build surface is now warning-clean and the test suite is parallel-safe — but several user-visible paper cuts are fixed too.

## 100% CPU on a single core, fixed (gh#102)

On Macs with a large Hermes history, Scarf could peg one core at 100% indefinitely while the Dashboard was open. The root cause was the Dashboard reloading its SQLite handle on **every** file-change tick: each time Hermes touched `state.db` / `state.db-wal`, Scarf would `close()` + re-`open()` the database, forcing SQLite to re-walk the WAL page map. On a 285 MB DB with a 114 MB uncheckpointed WAL — the reporter's real setup — that close+reopen is pure, compounding cost.

The fix keeps the read-only handle open across loads (a read-only SQLite connection sees Hermes's WAL writes transparently, so there's nothing to "refresh" in steady state). Close+reopen is now reserved for the schema-migration escape hatch when you upgrade Hermes mid-session. Each file-change tick now runs the four Dashboard queries against the already-open handle in microseconds instead of reopening a 285 MB file.

> If your WAL has grown large, the upstream driver is Hermes (the only process with write access). A one-time `sqlite3 ~/.hermes/state.db "PRAGMA wal_checkpoint(TRUNCATE); VACUUM;"` while Hermes is stopped will shrink it.

## Menu-bar status stopped flashing every 10 seconds (gh#105)

Remote-server users saw the menu-bar status chrome flash on a 10-second cadence even when nothing changed. `ServerLiveStatus` was republishing its `@Observable` state on every health poll regardless of whether the value actually changed — and SwiftUI invalidates on every setter call, not on real change. A one-line equality guard before each assignment means an unchanging healthy poll no longer re-renders anything; the menu bar stays still through the entire 10s/30s/60s/120s/300s backoff cycle.

(The companion request in gh#105 — a manual "Hermes binary" path field for Docker-fronted hosts — shipped in v2.10.1.)

## Reasoning is visible again on resumed thinking-model chats

When you reopen a chat with a thinking model, the **REASONING** disclosure now appears and loads the model's chain-of-thought on demand — including the newest models that store *only* the rich `reasoning_content` (verified against Hermes v0.16's source). Previously, the two-phase chat loader deliberately skips the heavy reasoning blob for speed, and for these models the lighter legacy `reasoning` column is empty — so the disclosure was hidden entirely on resume. Scarf now carries a cheap "reasoning available" signal in the fast path and lazy-loads the full transcript the first time you open the disclosure, so the speed benefit is preserved and the reasoning is one tap away.

## Snappier sidebar navigation

Switching sidebar sections no longer re-runs each feature's remote `load()` over SSH on every re-entry. Feature panes (Settings, Platforms, Plugins, Quick Commands, Models, MCP Servers, Webhooks, Cron) now keep their loaded data and view state across section switches, refreshing only when the underlying files actually change or you hit Reload. On a remote host that's several fewer SSH round-trips every time you bounce between sections.

## Docker config-save errors are now self-diagnostic (gh#112)

On the iOS app, a failed `hermes config set` during chat pre-flight used to surface only a generic "Couldn't save model+provider to config.yaml." For Hermes-in-Docker setups (where `hermes` is a wrapper around `docker compose exec`), the wrapper's actual diagnostic was being swallowed. The failure banner now shows the real exit code and stderr — missing container, missing TTY, PATH miss, etc. — so you can self-diagnose without capturing a separate log.

## More fixes from the audit

- **Loading overlays** on the Insights, Sessions, and Logs panes (no more blank-then-pop on slow remotes).
- **Standard menu commands**: ⌘, opens Settings, a Help menu links the Hermes docs, and ⌘F focuses the Sessions search.
- **Quieter when backgrounded**: the menu-bar live-status poll floors its cadence at 60s while Scarf isn't frontmost, killing the idle SSH-poll storm without freezing the always-visible status.
- **Cron corruption is surfaced**: a `jobs.json` that can't be parsed now shows a warning instead of a misleading "No cron jobs yet."
- **Chat streaming polish**: block markdown is skipped mid-stream (no half-rendered code fences), and the transcript anchors correctly on session activate.
- A `fatalError` in the SQL layer became a recoverable `throw`; per-render formatter allocations were hoisted to cached statics; in-flight remote loads now cancel when you navigate away.

## Under the hood

- **Warning-clean across the board.** The macOS app, the iOS app, and the ScarfCore package now build with **zero** compiler warnings, including the full Swift 6 strict-concurrency (main-actor isolation, `Sendable`) cleanup.
- **A trustworthy test suite.** 613 tests, parallel-safe, and no longer coupled to your real `~/.hermes` — ScarfCore tests run against isolated temp homes, so a test run can never touch your live Hermes data.
- **Release-signing safeguard.** The release script now preflights the Sparkle EdDSA key against the public key baked into the app, preventing the mis-signed-appcast class of incident.

## Upgrade notes

- Sparkle will offer this update automatically on next launch (or **Scarf → Check for Updates**).
- macOS 14.6+ (Sonoma) deployment target unchanged.
- **iOS testers:** the gh#112 diagnostic and the reasoning-disclosure fix live in the shared core; a ScarfGo TestFlight build carrying them is queued separately on the iOS track.
