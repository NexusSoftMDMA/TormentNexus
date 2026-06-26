# Plan — t-aud25: Make the local Hermes home injectable (test isolation)

> Status: NOT STARTED · Created 2026-06-13 (from the Swift audit follow-up loop) ·
> Owner: next session · Risk: MEDIUM (ScarfCore API seam; pervasive `ServerContext.local`) ·
> Source: t-aud22 verification

## Problem

ScarfCore + app tests are coupled to the developer's **real `~/.hermes`** because
`HermesPathSet.defaultLocalHome` is derived from `NSHomeDirectory()` and is **not
test-injectable**. Under Swift Testing's parallel suites this causes cross-suite races and
real-data risk:

- `ModelPresetServiceDiskTests` backs up/restores the real
  `~/.hermes/scarf/model_presets.json` (its own comment admits this is a compromise) and
  flaked under the full parallel run.
- `RemoteSQLiteBackendTests.openWithDefaultTildeHomeExpands` moved aside / symlinked the
  real `~/.hermes/state.db` (now **skipped** on machines with a real `~/.hermes` — t-aud22).
- **Still failing deterministically:** `M0dViewModelsTests.richChatViewModelInitsEmpty` —
  `RichChatViewModel.init` loads global slash commands from the real `~/.hermes`, so on any
  machine where Scarf has bootstrapped the `/scarf-*` commands, `availableCommands.count > 1`
  and `hasBroaderCommandMenu`/`supportsCompress` are true while the test expects false.

t-aud22 fixed the *parallelism* flakiness (ScarfMon global backend; tilde test skip). This
ticket fixes the *root*: the non-injectable home.

## Goal

Make the local home **per-instance injectable** (NO process-global override — that would
just move the race) so each test constructs a `ServerContext` pointing at a unique temp
home. Then: remove the backup/restore hack, un-skip the tilde test, and fix
`richChatViewModelInitsEmpty` — full `swift test` green (610/610), zero real-`~/.hermes`
contact, zero behavior change in production.

## Approach

1. Add an injection seam to the path layer. `HermesPathSet` (ScarfCore) currently computes
   the local home from `NSHomeDirectory()`. Add an initializer / factory that accepts an
   explicit local-home path, threaded through to all derived paths
   (`stateDB`, `configYAML`, `scarf/…`, `skills/…`, `cron/…`, `memories/…`, etc.).
   Production keeps the `NSHomeDirectory()` default.
2. Give `ServerContext` a test-friendly constructor that builds a `.local`-shaped context
   from an explicit home (e.g. `ServerContext.local(home: URL)` or
   `ServerContext(localHome: …)`). It must be **additive** — `ServerContext.local`
   (the production singleton) and all current call sites stay unchanged.
   - Note the precedent: `ServerContext.sshTransportFactory` is already a test seam
     (see `M5FeatureVMTests`); mirror that style.
3. Migrate the coupled tests to the new seam (each with its own `FileManager
   .temporaryDirectory/<uuid>` home), and delete the workarounds:
   - `ModelPresetServiceDiskTests.sandboxed` → construct a temp-home context; drop the
     real-file backup/restore.
   - `RemoteSQLiteBackendTests.openWithDefaultTildeHomeExpands` → use a temp `$HOME`-style
     context; remove the `.enabled(if: !exists(~/.hermes))` skip and the real-`~/.hermes`
     move/symlink.
   - `M0dViewModelsTests.richChatViewModelInitsEmpty` → init `RichChatViewModel` with a
     temp-home context so `availableCommands` is genuinely empty.
   - Grep for other suites touching real `~/.hermes` and migrate them too.

## Affected files

- `scarf/Packages/ScarfCore/Sources/ScarfCore/…` — `HermesPathSet` (+ wherever
  `ServerContext` / `.local` / `.paths` are constructed). Find with
  `memophant code find HermesPathSet` / `grep -rn "defaultLocalHome\|ServerContext.local"`.
- Tests: `ModelPresetServiceTests.swift`, `RemoteSQLiteBackendTests.swift`,
  `M0dViewModelsTests.swift` (+ any others surfaced by grep).

## Risks / gotchas

- `ServerContext.local` is pervasive — the seam MUST be additive; do not change the
  production default or the resolved production paths.
- Don't introduce a global mutable home override (e.g. a static var / setenv HOME) — that
  reintroduces the exact parallel race this ticket removes. Injection must be per-instance.
- `RichChatViewModel.init` does synchronous home reads (slash commands) — verify the temp
  home makes those genuinely empty; may reveal further init-time real-home reads to thread.

## Acceptance / verification

- `swift test --package-path scarf/Packages/ScarfCore` → **610/610 green**, including
  `M0dViewModelsTests`, the un-skipped `openWithDefaultTildeHomeExpands`, and
  `ModelPresetServiceDiskTests` with no backup/restore.
- No test reads or writes the real `~/.hermes` (audit the diffs / add a guard).
- Production app still resolves `~/.hermes` exactly as before (smoke: launch, Settings,
  Dashboard read real data).
- Pairs with t-aud22 (done) to make the whole suite parallel-safe.
