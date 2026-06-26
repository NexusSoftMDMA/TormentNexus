# WS-4 Plan: Curator archive + prune + list-archived (v2.8.0 / Hermes v0.13)

> **Scope.** Catch Scarf's Curator surface up to Hermes v0.13's new write-side
> verbs: `archive <skill>`, `prune`, `list-archived`, and the synchronous flavor
> of `run`. WS-4 owns Mac UX end-to-end + the ScarfCore parser/service work that
> backs it. iOS catches up read-only in WS-9 (deferred — note at the end).

---

## Goals

1. **Wire all four new v0.13 curator verbs** (`archive`, `prune`, `list-archived`,
   synchronous `run`) into ScarfCore through a typed actor surface so the view
   model stops shelling out via `runHermes` ad-hoc.
2. **Replace the v0.12 placeholder restore sheet** (free-form text field that
   prompted the user to remember archived skill names) with an actual list
   of archived skills returned by `hermes curator list-archived`, each row with
   per-row Restore + Prune-this-one actions.
3. **Add an "Archive" affordance** to every active-skill row in the leaderboard
   so users can manually archive a skill the curator didn't auto-archive.
4. **Add a destructive "Prune all archived" toolbar button** that opens a
   confirm sheet enumerating exactly which archived skills are about to be
   deleted forever.
5. **Make the "Run Now" button block-with-progress on v0.13+** since the verb is
   now synchronous; preserve fire-and-forget on pre-v0.13 hosts.
6. **Pre-v0.13 hosts must see the v2.7.x curator surface unchanged** — no
   "Archive" buttons, no Archived section, no Prune button. The legacy
   `CuratorRestoreSheet` stays accessible (it's all the v0.12 host has).
7. **Keep parsing pure & testable**: list-archived / prune-summary parse paths
   live in `HermesCuratorStatusParser` (or a sibling) with synthetic-fixture
   coverage in `HermesCuratorParserTests`.

Non-goals: iOS surface (WS-9), curator config knobs (out of scope — config tab
already covers `auxiliary.curator`), exporting reports.

---

## CLI integration — wire shape per verb

> **Investigation note.** Hermes v0.13 ships these verbs but neither the release
> notes nor the CLI man-page in our repo capture the exact stdout format. Plan
> assumes both human-text and `--json` are available since that's the v0.12
> Kanban convention; first task at implementation time is to run each verb
> against a real v0.13 install and capture stdout into `Tests/Fixtures/`. If
> `--json` doesn't exist for one of these verbs, fall back to a defensive
> text parser and add a `// TODO upstream` flag. **All assumed CLI flags below
> must be confirmed before wiring the parser.**

### `hermes curator list-archived [--json]`

- **Wire shape:** prefer `--json` and decode to `[HermesCuratorArchivedSkill]`.
  Fall back to text parse if the flag isn't present (mirrors `kanban runs` JSON
  envelope handling).
- **Assumed JSON shape (verify on first run):**

  ```json
  [
    {
      "name": "legacy-helper",
      "category": "templates",
      "archived_at": "2026-04-22T03:14:09Z",
      "reason": "stale: 91d unused",
      "size_bytes": 4521,
      "path": "/Users/u/.hermes/skills/.archived/legacy-helper"
    }
  ]
  ```

- **New model:** `HermesCuratorArchivedSkill` in
  `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesCuratorReport.swift`
  with `name: String`, `category: String?`, `archivedAt: String?`,
  `reason: String?`, `sizeBytes: Int?`, `path: String?`. All optional except
  `name` so a stripped-down host doesn't crash the view. Identifiable on `name`.
- **Empty-state sentinel:** Hermes may print `"no archived skills"` instead of
  `[]` (parallel to `"no matching tasks"` in Kanban). Treat as empty — same
  defensive fold KanbanService does at line ~45 today.

### `hermes curator archive <skill-name>`

- **Wire shape:** non-destructive (skill is moved, not deleted). No `--json`
  needed — exit code is the success channel; stdout is human-readable.
- **Argv:** `["curator", "archive", name]`. No flags in v0.13.
- **Side effects we surface to the user:** the active count drops by 1, the
  archived count rises by 1 — both visible after the next `status` reload.

### `hermes curator prune [--dry-run]`

- **Wire shape:** destructive. Removes everything currently archived. Open
  question 1 (below): does Hermes v0.13 ship `--dry-run`? Plan **two code paths**:
  1. **If `--dry-run` exists:** Scarf's prune confirm sheet calls
     `hermes curator prune --dry-run` first, parses the "would remove N skills"
     output, and renders the list. Final confirmation calls
     `hermes curator prune` (no flag). This is the preferred path.
  2. **If no `--dry-run`:** Scarf calls `hermes curator list-archived` to
     enumerate what's about to be deleted, shows that list in the confirm
     sheet, then calls `hermes curator prune` once the user confirms.
- **Assumed `--dry-run` JSON output (verify):**

  ```json
  { "would_remove": [{ "name": "...", "size_bytes": 4521 }, ...], "total_bytes": 12345 }
  ```

- **Optional per-skill prune:** if Hermes accepts
  `hermes curator prune <name>` (single-skill prune), wire it as a per-row
  action in the Archived list. **Verify before implementing** — release notes
  describe `prune` only in the bulk sense. If single-skill is unavailable, the
  per-row "Prune" button on the Archived list is dropped from the v2.8
  scope and only the bulk "Prune all archived" toolbar button ships.

### `hermes curator run` (now synchronous)

- **Wire shape:** unchanged argv. Behavior changes from fire-and-forget to
  blocking on v0.13+. Plan: bump the `runProcess(timeout:)` value from the
  current 30 s default to 600 s on v0.13+ hosts. Surface a `ProgressView` next
  to the "Run Now" button while the call is in flight, and disable the button
  until completion.
- **Capability branch:** `if caps.hasCuratorArchive { /* blocking with
  progress */ } else { /* fire-and-forget, immediate toast */ }`.
- **Cancel UX:** for v0.13+ blocking runs, plan a "Cancel" button that calls
  `transport.cancel()` on the running process (existing TransportError path).
  If transport-level cancel isn't reliable (Local vs Citadel parity), the
  cancel button is dropped and we just show indeterminate progress.

---

## Files to change (with specific edits)

### New files

- **`scarf/Packages/ScarfCore/Sources/ScarfCore/Services/CuratorService.swift`**
  — new `public actor CuratorService`. Mirrors `KanbanService` shape exactly:
  pure I/O, no UI state, every public method dispatches the CLI invocation
  through `Task.detached(priority: .utility)` inside the actor. Exposes:

  ```swift
  public actor CuratorService {
      public init(context: ServerContext)

      // Reads
      public func status() async -> HermesCuratorStatus  // moves logic out of VM
      public func listArchived() async throws -> [HermesCuratorArchivedSkill]

      // Writes — already-wired verbs (refactored from VM helpers)
      public func runNow(synchronous: Bool, timeout: TimeInterval) async throws
      public func pause() async throws
      public func resume() async throws
      public func pin(_ name: String) async throws
      public func unpin(_ name: String) async throws
      public func restore(_ name: String) async throws

      // Writes — new in v0.13 (WS-4)
      public func archive(_ name: String) async throws
      public func prune(dryRun: Bool) async throws -> CuratorPruneSummary

      // Pure helpers
      public nonisolated static func parseListArchived(stdout: String) throws -> [HermesCuratorArchivedSkill]
      public nonisolated static func parsePruneDryRun(stdout: String) throws -> CuratorPruneSummary
  }
  ```

  - Errors land in a new `CuratorError` enum (Sendable, LocalizedError) —
    `transport(message:)`, `nonZeroExit(verb:code:stderr:)`,
    `decoding(verb:message:)`. Identical shape to `KanbanError`.
  - `runNow(synchronous:timeout:)` takes the capability-decided sync flag from
    the call site; the service itself stays version-agnostic (only the timeout
    differs in practice).

- **`scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesCuratorArchive.swift`**
  — new file holding `HermesCuratorArchivedSkill` and `CuratorPruneSummary`
  structs. Both `Sendable, Equatable, Identifiable, Codable`.

  ```swift
  public struct HermesCuratorArchivedSkill: Sendable, Equatable, Identifiable, Codable {
      public var id: String { name }
      public let name: String
      public let category: String?
      public let archivedAt: String?
      public let reason: String?
      public let sizeBytes: Int?
      public let path: String?

      // Computed for UI — never persisted.
      public var sizeLabel: String { /* "4.4 KB" / "—" */ }
      public var archivedAtLabel: String { /* "2026-04-22" / "—" */ }
  }

  public struct CuratorPruneSummary: Sendable, Equatable, Codable {
      public let wouldRemove: [HermesCuratorArchivedSkill]
      public let totalBytes: Int
      public var totalCount: Int { wouldRemove.count }
  }
  ```

- **`scarf/scarf/Features/Curator/Views/CuratorArchivedSection.swift`** — new
  Mac sub-view used by `CuratorView`. Renders a `ScarfCard` containing the
  Archived list. Inputs: `[HermesCuratorArchivedSkill]`,
  `onRestore(name:)`, `onPruneOne(name:)?`, `onPruneAll()`. Empty-state path
  renders an "No archived skills" `ScarfCard` with copy explaining what archive
  does (helpful since Curator hasn't run yet on a fresh install).

- **`scarf/scarf/Features/Curator/Views/CuratorPruneConfirmSheet.swift`** —
  new destructive-confirm sheet. Presents the about-to-be-removed list, total
  count, total bytes, and a final "Prune permanently" red button.

### Edited files

- **`scarf/Packages/ScarfCore/Sources/ScarfCore/ViewModels/CuratorViewModel.swift`**
  - Replace inline `runAndReload(args:successMessage:)` helpers with
    `service.<verb>()` calls. Keep the toast + reload pattern inside the VM.
  - Add new `@Observable` state:
    - `archivedSkills: [HermesCuratorArchivedSkill] = []`
    - `isLoadingArchive = false`
    - `isPruning = false`
    - `pruneSummary: CuratorPruneSummary?`
    - `pendingArchiveName: String?` (track which skill is currently being
      archived so the row can show a small spinner without blocking the rest)
    - `errorMessage: String?` (replace transient-toast-only failure path with
      an inline-banner state, mirroring KanbanBoardViewModel)
  - Add new methods:
    - `func loadArchive() async`
    - `func archive(_ name: String) async`
    - `func planPrune() async` — calls `service.prune(dryRun: true)`, populates
      `pruneSummary`, opens the confirm sheet (sheet binding sits in the View)
    - `func confirmPrune() async` — calls `service.prune(dryRun: false)`
    - `func pruneOne(_ name: String) async` — only wired if upstream supports
      single-skill prune; otherwise method elided
  - Update `runNow()` to accept a `caps: HermesCapabilities` argument (passed
    from the View) and switch between sync/async invocations:
    - On v0.13+: `await service.runNow(synchronous: true, timeout: 600)` and
      poll `viewModel.isLoading` for a progress spinner.
    - On pre-v0.13: existing fire-and-forget; toast says "Curator run started".
  - Construct service lazily: `private lazy var service = CuratorService(context: context)`.

- **`scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesCuratorReport.swift`**
  - No edits to existing models. Add archive-related types in the new
    `HermesCuratorArchive.swift` to keep the diff scoped. (Decision: keep one
    file per logical surface.)

- **`scarf/scarf/Features/Curator/Views/CuratorView.swift`**
  - Inject `@Environment(\.hermesCapabilities)` and read
    `caps?.hasCuratorArchive ?? false` once into a local `let archiveAvailable`.
  - Header toolbar additions (only when `archiveAvailable`):
    - "Prune Archived…" `ScarfDestructiveButton` in the overflow `Menu`,
      disabled when `archivedSkills.isEmpty && !isLoadingArchive`.
  - Replace "Restore Archived…" menu item with a deep-link to scroll to the
    new Archived section (when `archiveAvailable`); leave the existing
    `CuratorRestoreSheet` reachable from the same menu **only on pre-v0.13** as
    the legacy fallback. On v0.13+ the menu shows just "Prune Archived…" and
    the section becomes the restore entry point.
  - Add `archiveAvailable` to `activityTables` rendering: each row in the three
    leaderboards gains an "Archive" pin-style button (small `Image(systemName:
    "archivebox")`) next to the existing pin button. Tooltip "Archive (move
    out of active set)". Hidden on pre-v0.13.
  - Append `CuratorArchivedSection` between `activityTables` and
    `lastReportSection` whenever `archiveAvailable`. Loaded by an additional
    `viewModel.loadArchive()` call inside `.task { … }`.
  - Wire confirm sheets:
    - `.sheet(isPresented: $showPruneSheet) { CuratorPruneConfirmSheet(...) }`
    - Existing `$showRestoreSheet` stays — only shown on pre-v0.13.
  - Run Now button: while `viewModel.isLoading && archiveAvailable`, show a
    `ProgressView()` next to the button label and disable the button. Tooltip:
    "Curator running — usually 10-90s. Hermes v0.13 runs synchronously."
  - Inline error banner: render `viewModel.errorMessage` as a yellow
    `ScarfCard` above `statusSummary` with an "x" dismiss. (Use existing
    `ScarfColor.warning` background; inspect the Kanban inline banner for
    pattern.)

- **`scarf/scarf/Features/Curator/Views/CuratorRestoreSheet.swift`**
  - **No code changes.** Sheet stays as v0.12 fallback. Add a doc-comment
    update at the top noting it's legacy-only on v0.13+ — the new
    `CuratorArchivedSection` is the preferred path. Don't delete this file
    even after WS-4 ships; pre-v0.13 hosts still need it.

- **`scarf/Scarf iOS/Curator/CuratorView.swift`**
  - **No code changes in WS-4.** WS-9 will add a read-only "Archived" section
    that mirrors the Mac one without per-row write actions. Leave a
    `// TODO(WS-9):` marker.

- **`scarf/Packages/ScarfCore/Tests/ScarfCoreTests/HermesCuratorParserTests.swift`**
  - Add tests (see "How to test" below).

---

## New types / fields

### `HermesCuratorArchivedSkill` (new)

In `HermesCuratorArchive.swift`. Codable directly from the assumed
`list-archived --json` shape. All fields except `name` optional so a
stripped-down host doesn't crash decoding. Computed `sizeLabel` /
`archivedAtLabel` for the view layer; never persisted.

### `CuratorPruneSummary` (new)

Lists what `prune --dry-run` would remove, plus aggregated `totalBytes`. The
view derives `totalCount` from `wouldRemove.count` so the wire shape stays
flat.

### `CuratorError` (new)

```swift
public enum CuratorError: Error, Sendable, LocalizedError {
    case transport(message: String)
    case nonZeroExit(verb: String, code: Int32, stderr: String)
    case decoding(verb: String, message: String)
}
```

Identical shape to `KanbanError`. View model maps these to inline-banner copy.

### `CuratorViewModel` additions

Already enumerated above. Note: the existing `transientMessage: String?` stays
for happy-path success ("Pinned X", "Resumed", "Archived legacy-helper");
failures route through the new `errorMessage: String?` so dismissals don't
cross-contaminate.

---

## Capability gating

All branches keyed on `caps?.hasCuratorArchive ?? false` (already defined in
`HermesCapabilities.swift:138` per the WS-1 inventory).

| Surface | Pre-v0.13 (`hasCurator && !hasCuratorArchive`) | v0.13+ (`hasCuratorArchive`) |
|---|---|---|
| Sidebar item | Visible (gated on `hasCurator`) | Visible |
| Status summary, leaderboards, pinned section | Identical | Identical |
| Per-row "Archive" button | **Hidden** | Visible |
| "Archived" section in CuratorView | **Hidden** | Visible (renders empty-state if no archives) |
| "Prune Archived…" menu item | **Hidden** | Visible |
| Existing "Restore Archived…" menu item | Visible (legacy text-prompt sheet) | **Hidden** (replaced by per-row Restore in Archived section) |
| `Run Now` blocking + progress | **No** (fire-and-forget) | **Yes** (synchronous w/ progress + 600s timeout) |
| `CuratorRestoreSheet.swift` | Used | Dead code path but file kept |

The View reads `caps` once at the top of `body` and threads
`archiveAvailable: Bool` down. Don't sprinkle `caps?.hasCuratorArchive` checks
across every sub-view — single source of truth at the entry point.

**Defensive default.** If `caps` is `nil` (preview / smoke test) or detection
hasn't completed yet, `archiveAvailable` resolves to `false` and the surface
behaves like a pre-v0.13 host. Same defensive shape as the Goals / Kanban-watch
gates.

---

## How to test

### CLI fixtures (capture once, commit to repo)

Create `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/Fixtures/Curator/`:

- `list-archived-empty.json` — `[]`
- `list-archived-three.json` — three skills with varied optional fields
- `list-archived-no-json-flag.txt` — text fallback (one row per line)
- `prune-dry-run.json` — `{ wouldRemove: [...], totalBytes: 12345 }`
- `status-with-archived.txt` — pre-existing fixture but with the
  `archived 4` count populated (drives the badge-count test)

These are captured by running the verbs against a real Hermes v0.13 install
on the dogfooding Mardon Mac Mini (per the "remote-servers dogfooding" memory)
during implementation. **Do not commit fabricated fixtures** — every fixture
must come from a real CLI invocation; otherwise the tests lock in a parser
that doesn't match production.

### Parser tests (`HermesCuratorParserTests.swift`)

Add to the existing `@Suite struct HermesCuratorParserTests`:

- `listArchivedEmpty()` — empty array decodes to `[]`.
- `listArchivedThreeSkills()` — happy path, asserts each field including
  optional `category` / `reason`.
- `listArchivedNoJSONFallback()` — text parser on the .txt fixture.
- `listArchivedNoArchivedSkillsSentinel()` — `"no archived skills"` literal in
  stdout folds to `[]` (parallel to KanbanService's `"no matching tasks"`).
- `listArchivedMissingOptionalsStaysSafe()` — JSON with only `name` populated
  decodes; size/date labels render `"—"`.
- `pruneDryRunHappyPath()` — `CuratorPruneSummary` decodes `wouldRemove` list
  and `totalBytes`.
- `pruneDryRunZeroSkills()` — `wouldRemove: [], totalBytes: 0` is valid.

### View-model tests (new file `CuratorViewModelTests.swift` — optional)

If a `MockCuratorService` protocol is plausible (the actor pattern allows
swapping via a protocol), add:

- `archiveCallSucceedsAndReloads()` — verifies `viewModel.transientMessage`
  flips to "Archived X" and `loadArchive()` is re-invoked.
- `archiveCallFailsRoutesToErrorBanner()` — failure path populates
  `errorMessage` (not `transientMessage`).
- `pruneTwoStepFlow()` — `planPrune()` populates `pruneSummary` then
  `confirmPrune()` clears it.
- `runNowIsSynchronousOnV013()` — VM passes `synchronous: true` to the service.

If extracting a protocol is too much yak-shave, plan only the parser tests.

### UI scenarios (manual verification on Mardon)

1. **Pre-v0.13 host (Mac Mini paused at v0.12):** sidebar shows Curator;
   page renders unchanged from v2.7.5; "Restore Archived…" menu item present;
   no Archive section, no Prune button; `Run Now` returns immediately.
2. **v0.13 host with no archives:** Archived section shows empty-state copy
   ("No archived skills — Curator will move stale skills here after the next
   review cycle."); "Prune Archived…" menu item disabled.
3. **v0.13 host with 3 archives:** Archived rows render with size + date;
   per-row Restore moves the skill back to active (verified by status reload);
   "Prune Archived…" opens confirm sheet listing all 3 with sizes; confirming
   removes them.
4. **v0.13 host: archive an active skill:** click Archive on a leaderboard
   row → row disappears from active list, appears in Archived section, active
   count drops by 1, archived count rises by 1.
5. **v0.13 host: blocking `Run Now`:** spinner appears, button stays disabled
   for the full duration; on completion the toast fires and the leaderboard
   reflects the new pass.
6. **v0.13 host: prune failure mid-flight:** simulate by SIGKILL'ing the
   curator process; verify error banner appears with stderr excerpt and the
   archived list isn't optimistically wiped.
7. **Restore sheet legacy fallback (pre-v0.13):** unchanged — verify the
   existing free-form text sheet still works.

---

## Open questions (must resolve at implementation start)

1. **Does `hermes curator prune` ship a `--dry-run` flag in v0.13?** If yes,
   the prune confirm sheet uses it for accurate "will remove these" copy. If
   no, the sheet falls back to displaying the current `list-archived` output
   and assumes prune removes exactly that set. This is the **biggest unknown**
   in the plan — the entire prune confirm UX shape pivots on this answer.
   _Resolution path: run `hermes curator prune --help` against v0.13 install
   on Mardon as the very first WS-4 implementation step._

2. **Does any curator verb support `--json`?** Plan assumes yes for
   `list-archived` and `prune --dry-run` since v0.12 Kanban set the precedent.
   If neither does, parser fixtures shift to text-only and decode logic moves
   into `HermesCuratorStatusParser`. Resolution: same as Q1.

3. **Is `hermes curator prune <name>` (single-skill prune) supported?** If so,
   per-row "Prune permanently" buttons in the Archived section are easy to
   add. If not, the only prune affordance is the bulk one. Plan accommodates
   both; per-row prune is dropped if upstream doesn't support it. Resolution:
   `hermes curator prune --help`.

4. **What's the exact synchronous-`run` timeout?** The release notes say
   "synchronous" but don't specify duration. 600 s (10 min) is a defensible
   default since curator runs are O(skill-count × LLM RTT). Long-running
   timeouts are acceptable here since the spinner is honest. Open: should
   Scarf surface a Cancel button? Probably not in v0.13 — transport-level
   process cancel isn't reliable across LocalTransport / CitadelServerTransport
   parity. Defer cancel to a later release if users complain.

5. **Confirm UX: typed-name confirmation, multi-tap, or destructive-button
   confirm sheet?** Scarf precedent (see "Constraints"):
   - **Memory reset** (`MemoryView.swift:56-65`) uses a single-step
     `.confirmationDialog` with `Button("Reset", role: .destructive)`. One
     click after the dialog opens.
   - **Template uninstall** (`TemplateUninstallSheet.swift:79-96`) uses a
     custom modal sheet listing every file/skill/cron/memory entry that will
     be removed, then a `ScarfPrimaryButton` tinted red labeled "Remove".
     One click after the sheet opens.
   - **Recommendation for prune:** match template-uninstall's shape. Prune is
     bulkier than memory-reset (multiple skills enumerated) and the user
     benefits from seeing the list. Custom sheet > confirmation dialog. The
     confirm button is `ScarfDestructiveButton` labeled "Prune permanently"
     with `keyboardShortcut(.defaultAction)` reserved for Cancel (not the
     destructive action — flipping it reduces accidental Enter-key prunes).
     Cancel is `ScarfGhostButton`, "Cancel". No typed-name confirmation; the
     enumerated list + the asymmetric keyboard shortcut is enough friction
     for a v0.13 surface that's already gated on a destructive intent ("I
     opened the prune sheet on purpose"). Single-tap on the destructive
     button is fine.

6. **Should the `lastReportPath` JSON field on `HermesCuratorStatus` get
   populated from a v0.13 path under `logs/curator/`?** v0.12 already populates
   it via the state file. v0.13 might point at a different directory after
   archive/prune runs (a separate `archive_report_path`?). Out of scope unless
   v0.13 introduces a new field — plan only handles existing
   `lastReportPath`. Defer to dogfooding.

---

## Out of scope (deferred)

- **iOS archive surface (WS-9).** Read-only Archived list mirroring the Mac
  one — no Archive / Prune actions. iOS users still get value (visibility
  into what the curator pruned). Scoped to a separate work-stream.
- **Curator scheduling knobs.** Already lives in Settings → Auxiliary; no
  changes for v2.8.
- **Per-skill curator-config flags** (e.g. "exclude this skill from auto-archive
  forever" — distinct from pin which already prevents auto-archive). Hermes
  doesn't ship this verb in v0.13. If the user wants permanent exclusion, pin.
- **Bulk-archive multi-select on active skills.** A future v0.14 verb might
  enable this; for v2.8 each archive is one CLI call.
- **Archive history / undo.** Hermes doesn't track archive history beyond the
  archived state itself. Restore is the undo for archive; once pruned, there's
  no recovery.
- **Curator report rendering for archive/prune events.** v0.12's
  `lastReportMarkdown` covers run reports; whether v0.13's archive/prune
  events land in a separate report is an open question. Stick with
  current rendering; revisit if dogfooding shows a gap.
- **`hermes curator pause/resume` on the synchronous run.** The new sync `run`
  doesn't interact with the autonomous schedule; pause/resume still work as
  before. No UX change.
- **Telemetry on prune.** No ScarfMon event for prune — measure if a user
  reports a slow prune. Easy follow-up.

---

## Risk + rollback

- **Highest risk:** parser drift between assumed JSON shape and Hermes v0.13's
  actual output. Mitigation: capture real fixtures at implementation start
  (see Open Q1 + Q2). Don't commit synthetic fixtures.
- **Second risk:** synchronous `run` timing out on `runProcess(timeout: 600)`.
  Mitigation: 10 min is generous; if a real run exceeds 10 min, that's a
  Hermes regression worth surfacing. Falls back to inline error banner.
- **Rollback path:** every WS-4 surface is gated on `hasCuratorArchive`. If a
  late-cycle bug shows up, a single-line revert in `HermesCapabilities.swift`
  (`atLeastSemver(0, 13, 0)` → `atLeastSemver(99, 0, 0)`) hides every WS-4
  surface from production hosts without ripping the code out. Same rollback
  shape as Kanban v3 used during v2.7.5 dogfooding.

---

## Estimate

| Bucket | Effort |
|---|---|
| `CuratorService` actor + models + errors | 0.5 day |
| Parser tests (with real fixtures captured from Mardon) | 0.5 day |
| `CuratorViewModel` refactor + new state + new methods | 0.5 day |
| `CuratorView` edits (header, per-row archive, archived section, prune sheet, error banner) | 1 day |
| `CuratorPruneConfirmSheet` + `CuratorArchivedSection` views | 0.5 day |
| Capability-gating audit + manual UI scenarios on pre-v0.13 + v0.13 hosts | 0.5 day |
| Unknown-buffer (CLI shape surprises, single-skill prune verification) | 0.5 day |

**Total: ~4 days of focused work** for one engineer, assuming a v0.13 install
is already running on Mardon and accessible for fixture capture. If `--json`
turns out to be missing on either of the two read verbs, add a 0.5-day
buffer for text-parser hardening.

---

## Sequencing inside WS-4

1. Capture real-world stdout fixtures by running every new v0.13 curator verb
   against the dogfooding Mardon install. Commit to
   `Tests/ScarfCoreTests/Fixtures/Curator/`. _(Resolves Open Q1 + Q2 + Q3.)_
2. Land `HermesCuratorArchive.swift` (models) + `CuratorService` actor with
   parser tests. No UI yet.
3. Refactor `CuratorViewModel` to use the service. Existing v0.12 surface
   should still work after this step — verify by rebuilding and clicking
   through every existing button.
4. Add the Mac Archived section + per-row Archive button + Prune confirm sheet
   behind the `archiveAvailable` flag.
5. Bump `Run Now` to synchronous-with-progress on v0.13+.
6. Pre-v0.13 regression pass on a v0.12 install.
7. v0.13 dogfood pass on Mardon — full UI tour + error injection.
8. Update relevant wiki pages (`Core-Services.md` adds `CuratorService`;
   sidebar / Curator user-guide page documents the new actions). Per
   CLAUDE.md the wiki update is part of the WS, not a follow-up.
