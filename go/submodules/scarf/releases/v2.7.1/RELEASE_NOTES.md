## What's in 2.7.1

A patch release covering three bug reports filed against 2.7.0, plus follow-up cleanups in the same neighborhood. No data migrations, no UI surface changes — drop-in replacement for 2.7.0 on Mac.

### Bug fixes

#### Mac

- **[#77](https://github.com/awizemann/scarf/issues/77) — Sessions screen renders empty even when Dashboard reports sessions exist.** v2.7.0 folded the Sessions tab's two SQL queries (sessions list + previews) into a single batched SSH round-trip for perf. The combined wire payload for any user with ~150+ sessions crossed macOS's 16–64 KB pipe-buffer threshold; without a concurrent reader draining the pipe, the remote `sqlite3 -json` blocked, the script never finished, our 30-second timeout fired, and the call returned an empty result. `SSHScriptRunner` now drains stdout/stderr concurrently with the running process via `FileHandle.readabilityHandler`, so the kernel pipe never fills. Same fix applied to the local-execution path. New regression test pushes 256 KB of synthetic output through the runner and asserts full delivery — would have wedged pre-fix.

- **[#78](https://github.com/awizemann/scarf/issues/78) — Skills "What's New" pill contradicts the Updates sub-tab.** The pill at the top of the Skills page was rendering on every sub-tab, including Updates. It counts **local** file deltas since the user last clicked "Mark as seen" (e.g. "18 new" = 18 skills landed on disk that you haven't acknowledged), while the Updates body runs `hermes skills check` to find skills with newer **upstream** versions available — a different concept. Two surfaces using the word "update" for two different things made the screen contradict itself. Two changes: the pill now renders only on the Installed sub-tab (Mac and ScarfGo), and its label says "X **changed** since you last looked" instead of "X updated" so the local-file vocabulary doesn't collide with upstream-update vocabulary anywhere on the page.

- **[#79](https://github.com/awizemann/scarf/issues/79) — Skills hub search returns nothing for terms visible in Browse.** With the source picker on "All Sources", `hermes skills search <query>` (no `--source` flag) routes through Hermes's centralized index and skips external API sources (skills-sh, github, clawhub, lobehub, well-known) — but Browse still aggregates from those sources, so a skill like `honcho` would show up in Browse and disappear in search. Same picker, same query, contradictory results. Rather than chase Hermes's index gaps, "All Sources" search now means "filter what you can already see": Scarf caches the most recent Browse payload and runs a client-side substring filter (case-insensitive against name, description, and identifier) against it, instantly. Source-specific searches still shell out to `hermes skills search --source <s>` for full upstream search semantics. Five new tests cover the filter behavior.

- **`hermesPIDResult()` — narrow the Hermes "is it running?" probe to the gateway.** Previously `pgrep -f hermes`, which matched any process with "hermes" in its argv: chat sessions Scarf itself spawns, `hermes -z` one-shots, log tails, even the README in an editor. The Dashboard "Hermes is running" badge could read true even when the gateway daemon was down. Tightened to a regex that matches only the gateway shape — `python -m hermes_cli.main gateway run …` and `/path/to/hermes gateway run …`. All callers (DashboardViewModel, HealthViewModel, SettingsViewModel, scarfApp, stopHermes) want the gateway PID specifically. Cherry-picked from [#76](https://github.com/awizemann/scarf/pull/76) — thanks to [@unixwzrd](https://github.com/unixwzrd) for the diagnosis and regex.

- **`HealthViewModel.stopDashboard()` — stop the dashboard by port, not `pkill -f`.** External-instance fallback used to be `pkill -f "hermes dashboard"`, broad enough to match shell history, log tails, README readers — anything with the substring in its argv. Now `lsof -tiTCP:<port> -sTCP:LISTEN` resolves the PID actually bound to the dashboard port and only that one process gets `SIGTERM`. Trusting the port is correct here: Scarf owns the configured port and the user-visible intent is "stop the thing on this port." Direction cherry-picked from [#76](https://github.com/awizemann/scarf/pull/76); the `-c hermes` filter from the original was dropped because Hermes installs as a Python shebang script and the kernel COMM is `python`, not `hermes` — `-c hermes` would silently miss every standard install.

### Documentation + tooling

- **`scripts/local-build.sh` + `BUILDING.md` for contributor builds.** New unsigned single-arch Debug build script for contributors without an Apple Developer account. Detects arm64 / x86_64, verifies xcode-select / xcrun / xcodebuild, probes the Metal toolchain (offers an interactive install on TTY, errors cleanly on CI), resolves Swift packages, builds Debug with signing disabled. Optional one-touch `ditto` to `/Applications/scarf.app` on explicit y/N. The canonical Release universal CLI in `README.md` is unchanged — `local-build.sh` is an alternative for contributors, not a replacement for the shipping build. Cherry-picked from [#76](https://github.com/awizemann/scarf/pull/76).

- **`BUILDING.md` + `CONTRIBUTING.md` — restored Sonoma compatibility messaging.** The runtime min is **macOS 14.6 (Sonoma)** — that's the `MACOSX_DEPLOYMENT_TARGET` on the main `scarf` target and is intentional. Build min is **Xcode 16.0** (needed for Swift 6 strict-concurrency features). The legacy CONTRIBUTING.md line had drifted to "Xcode 26.3+ / macOS 26.2+", which would have steered Sonoma contributors and users away from a build that actually runs on their box. Corrected, with a load-bearing-callout in BUILDING.md so future doc edits don't silently raise the floor again.

### Migrating from 2.7.0

Sparkle will offer the update automatically. No config migration, no schema changes. Existing sessions, skills, and projects are untouched.

If you've been working around #77 by collapsing the sidebar or restarting Scarf to repopulate the Sessions list, you can stop — sessions should load reliably now.

### Acknowledgements

- [@bricelb](https://github.com/bricelb) for the three v2.7.0 bug reports ([#77](https://github.com/awizemann/scarf/issues/77), [#78](https://github.com/awizemann/scarf/issues/78), [#79](https://github.com/awizemann/scarf/issues/79)) — well-instrumented reproductions including screenshots and environment details made the diagnosis straightforward.
- [@unixwzrd](https://github.com/unixwzrd) for [#76](https://github.com/awizemann/scarf/pull/76) — the gateway-pgrep tighten, the `pkill -f "hermes dashboard"` direction, and the `local-build.sh` contributor flow are all cherry-picked from that PR.
