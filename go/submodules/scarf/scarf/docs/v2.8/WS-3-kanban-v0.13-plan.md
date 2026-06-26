# WS-3 Plan: Kanban v0.13 diagnostics + recovery UX

**Workstream:** WS-3 of Scarf v2.8.0
**Hermes target:** v0.13.0 (v2026.5.7)
**Capability gate:** `HermesCapabilities.hasKanbanDiagnostics` (already shipped in WS-1, PR #80; resolves to `>= 0.13.0`)
**Builds on:** v2.7.5 Kanban v3 (drag-and-drop board, per-project tenants, optimistic-merge VM, inspector pane). The existing surface stays intact; this WS layers v0.13 reliability + recovery affordances on top.
**Owner:** TBD
**Reviewers:** Alan (always); whoever is on Kanban duty during v2.8 cycle.

---

## Goals

The Hermes v0.13.0 release notes list eight Kanban-shaped items in scope for Scarf:

1. **Hallucination gate + recovery UX** for worker-created cards — workers now emit a "I created a follow-up card" claim that Hermes flags as `hallucination_gate_status=pending` until something verifies the underlying card exists. Scarf needs to render the flag and offer Verify / Reject so the user is the verification gate.
2. **Generic diagnostics engine** for task distress signals — Hermes now emits a structured diagnostics array on a task / run when it observes distress (heartbeat-stalled, repeated tool errors, unbounded retry loop, OOM proxy, etc.). Scarf needs to render those diagnostics in the inspector so the user can act before the auto-block fires.
3. **Per-task `max_retries` override** — `hermes kanban create --max-retries N` (write-once at create) and the field shows up on `kanban show --json`. Surface on the create sheet + inspector header.
4. **Multiline textarea for inline-create title** — v0.13 server tolerates multi-line titles. The Scarf create sheet's title is currently a single-line `ScarfTextField`; convert to a multi-line input so a long title doesn't get clipped on hover-truncate.
5. **Heartbeat / reclaim / zombie / retry-cap reliability fixes** — mostly server-side, but Scarf's run-row + log-tab phrasing ("stale_lock") becomes user-hostile when v0.13 emits a richer outcome ("zombied — reclaimed by reaper"). Render the new outcome string verbatim and add a glossary tooltip.
6. **Auto-block workers that exit without completing** + `auto_blocked_reason` — currently Scarf renders a generic "Last run: blocked" banner; v0.13 attaches a structured reason ("worker exited without `kanban complete`"). Replace the generic banner with the reason when present.
7. **Detect darwin zombie workers** — when a card is reclaimed because the worker zombied (process exited but didn't release the lock), the diagnostics engine emits a `darwin_zombie_detected` kind. Render with a specific glyph + tooltip rather than the generic stale-lock banner.
8. **Unify failure counter across spawn / timeout / crash outcomes** — server-side counter rename; Scarf's run-row outcome label rendering may need to absorb a new normalized counter (`failure_count` rather than three separate counters). Verify the run row still renders all outcomes.

The two release-notes items NOT in WS-3 scope:

- **Multi-project boards** — already shipped in v2.7.5 via per-project tenants. Hermes v0.13's "one install, many kanbans" framing is the server's catch-up to what Scarf already solved client-side via the `scarf:<slug>` tenant convention. No change here.
- **Shared board, workspaces, and worker logs across profiles** — entirely server-side; Scarf already shows whichever assignee owns a row.
- **Dashboard: workspace kind + path inputs, per-platform home-channel notification toggles** — workspace kind/path already shipped in v2.7.5 (`KanbanCreateSheet.workspaceField`); home-channel toggles are in WS-5 (gateway / messaging) not Kanban.
- **Worker task-ownership enforcement on destructive tool calls** — server-side; Scarf observes the failure mode (a run ends with `permission_denied`) but doesn't need new UI.

### Non-goals (explicitly deferred)

- **Within-column reorder.** Hermes still has no `update --priority` verb. CLAUDE.md "Kanban v3" section explicitly forbids client-side ordering sidecars.
- **Drag from Done.** Done is terminal; the WS-2.7.5 transition planner already throws `forbiddenTransition`. No change.
- **Mutating `priority` / `title` / `body` post-create.** No CLI verb exists. We surface `max_retries` on the inspector header in read-only form.
- **iOS read-only counterpart.** WS-9 picks up iOS catch-up. Scope here is Mac.
- **Live `watch` streaming.** v2.7.5 polls every 5s. v0.13 hasn't added a stable `watch --json` shape Scarf can rely on; deferred until a future flag (`hasKanbanWatch`).

---

## Files to change

The plan is intentionally minimal-touch. Most of the lift is in the Mac inspector + card view + create sheet; the model layer adds a handful of `Codable` fields with `nil` defaults so pre-v0.13 hosts decode without error.

### 1. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesKanbanTask.swift`

**Why:** v0.13 adds four task-level fields the inspector / card need (`max_retries`, `auto_blocked_reason`, `hallucination_gate_status`, optional `diagnostics`). All four must be `Optional` with `nil` decoded for pre-v0.13 hosts.

**Edits:**

- Add four new stored properties between `currentRunId` and the end of the property list (preserve existing initializer ordering — append at the tail of the parameter list with nil defaults so call sites in `KanbanModelsTests`, etc. don't break):
  - `public let maxRetries: Int?`
  - `public let autoBlockedReason: String?`
  - `public let hallucinationGateStatus: String?` — wire enum: `pending` / `verified` / `rejected` / nil. Stays a `String` for the same forward-compat reason `status: String` does (Hermes might add `quarantined`).
  - `public let diagnostics: [HermesKanbanDiagnostic]` — defaults to `[]` when absent, matching the existing `skills` pattern (line 115).
- Extend `enum CodingKeys` with:
  - `case maxRetries = "max_retries"`
  - `case autoBlockedReason = "auto_blocked_reason"`
  - `case hallucinationGateStatus = "hallucination_gate_status"`
  - `case diagnostics`
- Extend the custom `init(from:)` with four `decodeIfPresent` calls. The `[HermesKanbanDiagnostic]` decode mirrors the `skills` decode: `(try? c.decodeIfPresent([HermesKanbanDiagnostic].self, forKey: .diagnostics)) ?? []`. Wrapping in `try?` matters — a single malformed diagnostic shouldn't poison the whole row.
- Extend the public memberwise initializer (the explicit one starting line 37) — add the four parameters at the tail with nil defaults so v2.7.5 callers compile unchanged.
- Add a typed-mirror enum `KanbanHallucinationGate` next to `KanbanStatus` so views don't string-compare:
  ```swift
  public enum KanbanHallucinationGate: String, Sendable, CaseIterable {
      case pending, verified, rejected
      public static func from(_ raw: String?) -> KanbanHallucinationGate? {
          guard let raw, !raw.isEmpty else { return nil }
          return KanbanHallucinationGate(rawValue: raw.lowercased())
      }
  }
  ```

**Tolerance contract:** A v0.12 row missing all four fields decodes successfully and renders with no v0.13 chrome. A v0.13 row with all four fields decodes and lights up the new chrome.

### 2. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesKanbanDiagnostic.swift` (NEW)

**Why:** Diagnostics are a fresh wire shape. They're attached in two places (per-task `diagnostics: [...]` and per-run `diagnostics: [...]`), but the Swift type is shared between the two sites.

**Shape (best inference from release notes — verify against live JSON during integration):**

```swift
public struct HermesKanbanDiagnostic: Sendable, Equatable, Identifiable, Codable {
    public let id: UUID = UUID()  // synthetic; not on wire
    public let kind: String       // heartbeat_stalled | tool_error_loop | retry_cap_hit |
                                  // unbounded_retry | darwin_zombie_detected | spawn_failure |
                                  // worker_exit_no_complete | …
    public let message: String?   // human-friendly elaboration
    public let detectedAt: String? // ISO-8601 (decode flexible — Unix int or string)

    enum CodingKeys: String, CodingKey {
        case kind
        case message
        case detectedAt = "detected_at"
    }
    // custom init(from:) for flexible timestamp decode, mirroring HermesKanbanTask.decodeFlexibleTimestamp
}
```

Plus a typed-mirror enum `KanbanDiagnosticKind` for known kinds (default `.unknown` for forward compat — matches the `KanbanStatus` / `KanbanEventKind` pattern). Glyph + color helpers live alongside it so views don't switch on raw strings.

**Cases for the typed-mirror enum (initial set; add as Hermes ships more):**

- `.heartbeatStalled` — heartbeat older than `max_runtime_seconds / 4`, glyph `waveform.path.badge.minus`, tint `.warning`
- `.toolErrorLoop` — same tool errored ≥ 3 times in a row, glyph `arrow.triangle.2.circlepath.exclamationmark`, tint `.warning`
- `.retryCapHit` — `failure_count >= max_retries`, glyph `nosign`, tint `.danger`
- `.unboundedRetry` — worker is retrying without backoff (was a v0.12 bug class), glyph `arrow.clockwise.circle.fill`, tint `.warning`
- `.darwinZombieDetected` — process zombied without releasing lock, glyph `apple.logo`, tint `.danger`
- `.spawnFailure` — `os/exec` returned non-zero spawning the worker, glyph `bolt.slash`, tint `.danger`
- `.workerExitNoComplete` — worker exited 0 without calling `kanban complete`, glyph `figure.walk.departure`, tint `.warning` (pairs with `auto_blocked_reason`)
- `.unknown` — fallback for any kind Hermes adds we don't recognize; render kind raw

### 3. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesKanbanRun.swift`

**Why:** Per-run diagnostics share the same type. The run row in the inspector renders them under the run.

**Edits:**

- Add `public let diagnostics: [HermesKanbanDiagnostic]` (defaults to `[]`).
- Extend `enum CodingKeys` with `case diagnostics`.
- Extend `init(from:)` with the same `decodeIfPresent` + `?? []` pattern.
- Extend the public memberwise initializer with the parameter (default `[]`).
- Extend `encode(to:)` with `try c.encode(diagnostics, forKey: .diagnostics)` (encoding round-trip matters for tests).
- Optional v0.13 housekeeping: `failure_count: Int?` if v0.13's unified counter is exposed on the run shape (unify failure counter across spawn / timeout / crash). If it appears as a top-level key on the run, decode it; if not, this stays a server-internal field and Scarf doesn't need it.

### 4. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesKanbanTaskDetail.swift`

**Why:** No structural change required if `diagnostics` is on the inner `HermesKanbanTask`. But verify the JSON shape: in some Hermes verbs the diagnostics array hangs off the *envelope* (`{task: {…}, comments: […], events: […], diagnostics: […]}`) rather than the task. If it's on the envelope, add an optional sibling field here and surface `detail.task.diagnostics ?? detail.diagnostics ?? []` from the inspector.

**Edits (defensive):** add `public let envelopeDiagnostics: [HermesKanbanDiagnostic]?` decoded from `case envelopeDiagnostics = "diagnostics"`. UI source of truth becomes a computed helper on the detail:

```swift
public var allDiagnostics: [HermesKanbanDiagnostic] {
    let onTask = task.diagnostics
    let onEnvelope = envelopeDiagnostics ?? []
    // Dedup by (kind, detectedAt). Wire-side dupes are unlikely but cheap to filter.
    var seen = Set<String>()
    return (onTask + onEnvelope).filter {
        let key = "\($0.kind)|\($0.detectedAt ?? "")"
        return seen.insert(key).inserted
    }
}
```

### 5. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/KanbanCreateRequest.swift`

**Why:** The create sheet needs a `--max-retries N` flag.

**Edits:**

- Add `public var maxRetries: Int?` to the struct.
- Add the parameter to the public initializer (tail position, default nil).
- Extend `argv()` between `maxRuntimeSeconds` and `createdBy` (line 80-ish):
  ```swift
  if let maxRetries {
      args.append(contentsOf: ["--max-retries", String(maxRetries)])
  }
  ```
- Argv ordering is purely cosmetic from Hermes's perspective (it re-parses), but keep deterministic order so test fixtures stay stable.

### 6. `scarf/Packages/ScarfCore/Sources/ScarfCore/Services/KanbanService.swift`

**Why:** Hallucination-gate verify / reject. Best inference from the release notes is that Hermes added a verb like `kanban verify <id>` or expanded `kanban show` with a sibling write-verb. **This needs verification** — see Open Questions #1.

**Edits (proposed; mark TODO until verified against Hermes v0.13 source):**

- Add a `verify(taskId:)` method that runs `hermes kanban verify <id>`. Returns Void; the polling loop picks up the new `hallucination_gate_status=verified`. If the verb is named differently (`hallucination verify`, `confirm`, `accept`), rename the Swift method to track. **Do not invent a CLI verb name without a real CLI to call against** — gate this behind a guarded TODO and pull from the live binary first.
- Add a `rejectHallucinated(taskId:)` method. Most likely path: the user "rejects" by archiving (since the worker's claim was a hallucination, the right resolution is to archive the bogus card). If Hermes ships a dedicated reject verb, wire it; otherwise route through `archive(taskIds:)` with a comment ("Rejected as hallucinated by Scarf user").
- **Do NOT** add a `setMaxRetries(taskId:)` post-create mutation method. Hermes pattern is write-once. Setting `max_retries` after create has no CLI verb in v0.13. Document this as a Limitation in inspector tooltips.

### 7. `scarf/scarf/Features/Kanban/Views/KanbanCreateSheet.swift`

**Why:** Multi-line title + new `Max retries` numeric field, both gated on `hasKanbanDiagnostics`.

**Edits:**

- Replace the single-line `titleField` (lines 116-122):
  ```swift
  ScarfTextField("What needs doing?", text: $title)
  ```
  with a multi-line variant. Two acceptable approaches:
  - **Preferred:** SwiftUI `TextField` with `axis: .vertical` and `lineLimit(1...4)`. Wraps cleanly inside the existing `ScarfTextField` chrome on macOS 14.6+. Pre-existing `ScarfTextField` is a wrapper — extend the wrapper to take an optional `axis` parameter or add a new `ScarfTextEditor` sibling component to `ScarfDesign`. Touch the design package only if the multi-line variant doesn't already exist there. (Audit `Packages/ScarfDesign/` first; if `ScarfTextEditor` exists, use it.)
  - **Fallback:** A bare `TextEditor` mirroring the `descriptionField` chrome, with a smaller `minHeight: 36, maxHeight: 96` so single-line titles still feel right.
- Gating: Since macOS 14.6 has no plumbing problem with multi-line text, keep the multi-line input on for **all** versions of Hermes — pre-v0.13 will simply receive a single-line title at the wire (`\n` stripped client-side before submit if Hermes < 0.13 truncates on newlines). Use the `hasKanbanDiagnostics` flag to **decide whether to strip newlines** at submit time, not whether to render the multi-line input. Read the capability via the existing `@Environment` injection pattern (look up how other create sheets read it; if not yet wired here, accept it as a `let capabilities: HermesCapabilitiesStore` init parameter).
- Add a new section between `priorityField` and `skillsField`:
  ```
  ┌─────────────────────────────┐
  │ Max retries                 │
  │ subtitle: "0 = no retries.  │
  │   Defaults to 3."           │
  │ ┌───────────────────────┐   │
  │ │ Stepper: [3] [- +]    │   │
  │ └───────────────────────┘   │
  └─────────────────────────────┘
  ```
- New `@State` storage: `@State private var maxRetries: Int = 3` and `@State private var maxRetriesEnabled: Bool = false`. Toggle gates whether `maxRetries` is sent at all (so we can preserve "let server pick the default" by leaving the flag absent).
- Show this section only when `capabilities.hasKanbanDiagnostics` is true. Pre-v0.13 hosts get the v2.7.5 sheet unchanged (no new field).
- Wire into `makeRequest()` (line 309-347): pass `maxRetries: maxRetriesEnabled ? maxRetries : nil`.
- Strip newlines in title pre-submit when `!capabilities.hasKanbanDiagnostics` to defend against pre-v0.13 hosts: `let titleForSubmit = trimmedTitle.replacingOccurrences(of: "\n", with: " ")` only on the pre-v0.13 path.

### 8. `scarf/scarf/Features/Kanban/Views/KanbanInspectorPane.swift`

**Why:** This is the biggest delta — diagnostics rendering, hallucination Verify/Reject banner, max_retries header chip, expanded auto_blocked_reason banner.

**Edits:**

#### 8a. Header chip row (lines 152-167)

Add a chip for `max_retries` when present (gated on `hasKanbanDiagnostics`):

```swift
if let maxRetries = task.maxRetries {
    ScarfBadge("retries: \(maxRetries)", kind: .neutral)
        .fixedSize()
        .help("Max retries set at create time. Hermes has no update verb — re-create the task to change this.")
}
```

Inserted in the chip-row HStack between `workspaceKind` and `tenant`.

#### 8b. Hallucination-gate banner (NEW, in `healthBanner(for:)`)

Insert above the existing `needsAssignee` / `hadFailedEndedRun` checks. Render only when `KanbanHallucinationGate.from(task.hallucinationGateStatus) == .pending`:

```swift
@ViewBuilder
private func hallucinationBanner(for task: HermesKanbanTask) -> some View {
    HStack(alignment: .top, spacing: ScarfSpace.s2) {
        Image(systemName: "questionmark.diamond.fill")
            .foregroundStyle(ScarfColor.warning)
            .font(.system(size: 13, weight: .semibold))
        VStack(alignment: .leading, spacing: 4) {
            Text("Created by a worker — verify before running")
                .scarfStyle(.captionStrong)
                .foregroundStyle(ScarfColor.foregroundPrimary)
            Text("A worker claimed it created this card; Hermes hasn't confirmed the underlying work exists. Verify the card matches a real follow-up, or reject if it's a hallucinated reference.")
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundMuted)
            HStack(spacing: ScarfSpace.s2) {
                Button("Verify") { onVerifyHallucination() }
                    .buttonStyle(ScarfPrimaryButton())
                Button("Reject") { onRejectHallucination() }
                    .buttonStyle(ScarfDestructiveButton())
            }
            .padding(.top, 2)
        }
        Spacer(minLength: 0)
    }
    .padding(ScarfSpace.s2)
    .background(
        RoundedRectangle(cornerRadius: ScarfRadius.md, style: .continuous)
            .fill(ScarfColor.warning.opacity(0.10))
    )
    .overlay(
        RoundedRectangle(cornerRadius: ScarfRadius.md, style: .continuous)
            .strokeBorder(ScarfColor.warning.opacity(0.4), lineWidth: 1)
    )
}
```

Two new closure parameters on the inspector init: `onVerifyHallucination: () -> Void`, `onRejectHallucination: () -> Void`. They're called from the buttons; `KanbanBoardView` wires them to `viewModel.verify(taskId:)` / `viewModel.rejectHallucinated(taskId:)`.

Render order in `healthBanner`: hallucination first (the user must resolve this before anything else makes sense), then unassigned, then last-failed-run. Stack vertically inside a `VStack(alignment: .leading, spacing: ScarfSpace.s2)` rather than the current `if/else if`.

#### 8c. Auto-blocked reason banner (extension of existing red banner)

Currently `healthBanner` renders a generic "Last run: blocked" message. v0.13 ships `auto_blocked_reason` on the task itself. Update logic:

```swift
if KanbanStatus.from(task.status) == .blocked,
   let reason = task.autoBlockedReason, !reason.isEmpty {
    bannerRow(
        icon: "exclamationmark.octagon.fill",
        tint: ScarfColor.danger,
        title: "Auto-blocked",
        message: reason  // verbatim — Hermes-side message is the source of truth
    )
}
```

This banner takes precedence over the existing `lastEndedRun.outcome == "blocked"` rendering (server-side reason is more specific than client-side derived).

#### 8d. Diagnostics rendering on Runs tab

Below each `runRow(_:)` (lines 562-594), insert a `diagnosticsRow(for:)` when the run has any:

```swift
if !run.diagnostics.isEmpty {
    diagnosticsBlock(run.diagnostics)
}
```

```swift
@ViewBuilder
private func diagnosticsBlock(_ diags: [HermesKanbanDiagnostic]) -> some View {
    VStack(alignment: .leading, spacing: 4) {
        Text("Diagnostics")
            .scarfStyle(.captionUppercase)
            .foregroundStyle(ScarfColor.foregroundFaint)
        FlowLayout(spacing: 4) {  // reuse existing layout primitive if present; otherwise HStack with wrapping
            ForEach(diags) { diag in
                let kind = KanbanDiagnosticKind.from(diag.kind)
                ScarfBadge(diag.kind, kind: kind.badgeKind)
                    .help(diag.message ?? diag.kind)
            }
        }
    }
    .padding(.top, 4)
}
```

If a `FlowLayout` primitive doesn't exist in the codebase, fall back to a single-line `ScrollView(.horizontal, showsIndicators: false)` so a long diag list doesn't blow out card width.

#### 8e. Diagnostics on the task header

Top-level diagnostics (the `task.diagnostics ?? []`, NOT the per-run ones) are about the task, not a specific attempt. Render under the chip row in the header:

```swift
if !task.diagnostics.isEmpty {
    diagnosticsBlock(task.diagnostics)
        .padding(.top, 4)
}
```

#### 8f. Action bar update

When `hallucination_gate_status == .pending`, suppress the "Start" button (Verify-or-Reject is the gate). The existing `primaryAction` switch already keys on `KanbanStatus.from(task.status)`; add a guard at the top of `@ViewBuilder primaryAction`:

```swift
if KanbanHallucinationGate.from(task.hallucinationGateStatus) == .pending {
    EmptyView()  // banner provides the actions
} else {
    // existing switch
}
```

### 9. `scarf/scarf/Features/Kanban/Views/KanbanCardView.swift`

**Why:** Card-level signals — hallucination dim + glyph, auto-block sub-line, diagnostics indicator.

**Edits:**

- New computed `private var hallucinationGate: KanbanHallucinationGate?` reading off the task.
- In `body`, apply 0.6 opacity when `hallucinationGate == .pending`:
  ```swift
  .opacity(task.isDone ? doneOpacity : (hallucinationGate == .pending ? 0.6 : 1.0))
  ```
- In `titleRow`, add a yellow ⚠ glyph when `hallucinationGate == .pending`. It overlaps semantically with the existing `needsAssignmentWarning` glyph, so:
  - If both are true, prefer the hallucination glyph (more specific).
  - Render at the same right-side slot.
  ```swift
  if hallucinationGate == .pending {
      Image(systemName: "questionmark.diamond.fill")
          .foregroundStyle(ScarfColor.warning)
          .font(.system(size: 11, weight: .semibold))
          .help("Worker-created — verify before running")
  } else if needsAssignmentWarning {
      Image(systemName: "exclamationmark.triangle.fill")
          .foregroundStyle(ScarfColor.warning)
          .font(.system(size: 11, weight: .semibold))
          .help("Unassigned — Hermes's dispatcher silently skips tasks with no assignee, …")
  }
  ```
- Auto-block sub-line: in the blocked branch of `relativeTimeLabel` (line 254-260), if `task.autoBlockedReason` is present, append the first 30 chars truncated:
  - Easier path: don't shoehorn into `relativeTimeLabel`. Add a separate sub-line in the footer above the existing `relativeTimeLabel` when `KanbanStatus.from(status) == .blocked && task.autoBlockedReason != nil`:
    ```swift
    if KanbanStatus.from(task.status) == .blocked,
       let reason = task.autoBlockedReason, !reason.isEmpty {
        Text(reason.prefix(60))
            .scarfStyle(.caption)
            .foregroundStyle(ScarfColor.danger)
            .lineLimit(1)
            .truncationMode(.tail)
            .help(reason)
    }
    ```
- Diagnostics indicator (subtle): if `!task.diagnostics.isEmpty`, render a small dot in the footer right side next to the priority indicator:
  ```swift
  if !task.diagnostics.isEmpty {
      Image(systemName: "stethoscope")
          .font(.system(size: 9))
          .foregroundStyle(ScarfColor.warning)
          .help("\(task.diagnostics.count) diagnostic signal\(task.diagnostics.count == 1 ? "" : "s")")
  }
  ```
- Done dim: leave alone; v0.13 darwin-zombie fix doesn't change Done semantics.

### 10. `scarf/scarf/Features/Kanban/Views/KanbanBoardView.swift`

**Why:** Wire the new inspector callbacks (`onVerifyHallucination`, `onRejectHallucination`) into the VM.

**Edits:**

- In the inspector instantiation, pass two new closures:
  ```swift
  KanbanInspectorPane(
      service: viewModel.service,
      taskId: id,
      ...,
      onVerifyHallucination: { viewModel.verifyHallucination(taskId: id) },
      onRejectHallucination: { viewModel.rejectHallucination(taskId: id) }
  )
  ```
- Capability gate ambient via the `HermesCapabilitiesStore` `.environment(_:)` injection from `ContextBoundRoot` (already in place per CLAUDE.md). Read with `@Environment(HermesCapabilitiesStore.self)` and pass the relevant flag down to `KanbanCreateSheet` for the max-retries field.

### 11. `scarf/scarf/Features/Kanban/ViewModels/KanbanBoardViewModel.swift`

**Why:** Add `verifyHallucination(taskId:)` and `rejectHallucination(taskId:)` methods. Also extend the optimistic-override mechanism to cover hallucination-gate transitions so the banner disappears immediately on Verify (and the card un-dims).

**Edits:**

- Add a sibling override map for hallucination state:
  ```swift
  /// Mirrors `optimisticOverrides` but for hallucination-gate transitions.
  /// Cleared when the polled response confirms the new gate status.
  private var optimisticHallucinationOverrides: [String: KanbanHallucinationGate] = [:]
  ```
- Or simpler: extend `optimisticOverrides` to a richer struct
  ```swift
  private struct OptimisticOverride {
      var status: String?
      var hallucinationGate: KanbanHallucinationGate?
  }
  private var optimisticOverrides: [String: OptimisticOverride] = [:]
  ```
  This is cleaner long-term; touches more existing code (~10 lines). Recommend the struct approach.
- Add `verifyHallucination(taskId:)`:
  ```swift
  func verifyHallucination(taskId: String) {
      // Optimistic — flip to verified locally so banner disappears.
      optimisticOverrides[taskId, default: .init()].hallucinationGate = .verified
      Task {
          do {
              try await service.verify(taskId: taskId)  // pending CLI verb confirmation; see Open Questions
              await refresh()
          } catch let err as KanbanError {
              optimisticOverrides[taskId]?.hallucinationGate = nil
              lastError = err.errorDescription
          } catch {
              optimisticOverrides[taskId]?.hallucinationGate = nil
              lastError = error.localizedDescription
          }
      }
  }
  ```
- Add `rejectHallucination(taskId:)`:
  ```swift
  func rejectHallucination(taskId: String) {
      // Treat as archive + comment for clarity in the audit trail.
      Task {
          do {
              try await service.comment(taskId: taskId, text: "Rejected as hallucinated (no underlying work).", author: nil)
              try await service.archive(taskIds: [taskId])
              await refresh()
          } catch let err as KanbanError {
              lastError = err.errorDescription
          } catch {
              lastError = error.localizedDescription
          }
      }
  }
  ```
  **Note:** if Hermes v0.13 adds a dedicated `kanban reject` or `kanban hallucination reject` verb, swap the body to call it. Either way, the VM API stays stable — the surface for views is "reject" returning Void.
- Update `mergePolledTasks` to clear `optimisticHallucinationOverrides` entries when the polled task's `hallucination_gate_status` matches:
  ```swift
  for (id, override) in optimisticOverrides {
      guard let row = filtered.first(where: { $0.id == id }) else {
          if !presentIds.contains(id) {
              optimisticOverrides.removeValue(forKey: id)
          }
          continue
      }
      // Status side (existing).
      if let optStatus = override.status,
         columnFromStatus(optStatus) == columnFromStatus(row.status) {
          optimisticOverrides[id]?.status = nil
      }
      // Hallucination gate side (new).
      if let optGate = override.hallucinationGate,
         KanbanHallucinationGate.from(row.hallucinationGateStatus) == optGate {
          optimisticOverrides[id]?.hallucinationGate = nil
      }
      // Empty override — drop entirely.
      if optimisticOverrides[id]?.status == nil,
         optimisticOverrides[id]?.hallucinationGate == nil {
          optimisticOverrides.removeValue(forKey: id)
      }
  }
  ```
- Update `effectiveColumn` and a new `effectiveHallucinationGate(_:)` to consult the override.

### 12. `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/KanbanModelsTests.swift`

**Why:** The tolerant-decode contract is the single most important invariant. Tests must cover both shapes.

**Edits:**

#### 12a. New test — v0.13 task shape decodes with all new fields populated:

```swift
@Test func decodeV013TaskFields() throws {
    let json = """
    {
      "id": "t_v013",
      "title": "v0.13 task",
      "status": "blocked",
      "max_retries": 5,
      "auto_blocked_reason": "worker exited without `kanban complete`",
      "hallucination_gate_status": "pending",
      "diagnostics": [
        {"kind": "worker_exit_no_complete", "message": "exit code 0 with no complete call", "detected_at": 1778160614},
        {"kind": "darwin_zombie_detected", "detected_at": "2026-05-09T12:00:00Z"}
      ]
    }
    """
    let task = try JSONDecoder().decode(HermesKanbanTask.self, from: Data(json.utf8))
    #expect(task.maxRetries == 5)
    #expect(task.autoBlockedReason?.contains("kanban complete") == true)
    #expect(task.hallucinationGateStatus == "pending")
    #expect(task.diagnostics.count == 2)
    #expect(task.diagnostics.first?.kind == "worker_exit_no_complete")
    #expect(task.diagnostics.last?.detectedAt?.contains("2026") == true)
}
```

#### 12b. New test — v0.12 (legacy) task shape decodes with new fields = nil/empty:

```swift
@Test func decodeV012TaskHasNoNewFields() throws {
    let json = """
    {"id": "t_legacy", "title": "v0.12 task", "status": "ready"}
    """
    let task = try JSONDecoder().decode(HermesKanbanTask.self, from: Data(json.utf8))
    #expect(task.maxRetries == nil)
    #expect(task.autoBlockedReason == nil)
    #expect(task.hallucinationGateStatus == nil)
    #expect(task.diagnostics.isEmpty)
}
```

#### 12c. New test — diagnostics with malformed entry doesn't poison the array:

```swift
@Test func decodeMalformedDiagnosticTolerated() throws {
    // If Hermes emits a malformed diagnostic, the rest of the task should
    // still decode. We use try? on the diagnostics decode so a single
    // bad entry doesn't reject the whole row.
    let json = """
    {
      "id": "t_x",
      "title": "x",
      "status": "ready",
      "diagnostics": "not-an-array"
    }
    """
    let task = try JSONDecoder().decode(HermesKanbanTask.self, from: Data(json.utf8))
    #expect(task.id == "t_x")
    // Diagnostics field couldn't decode — treat as empty.
    #expect(task.diagnostics.isEmpty)
}
```

#### 12d. New test — `KanbanHallucinationGate.from(_:)` mirror:

```swift
@Test func hallucinationGateMirrorMapsKnownValues() {
    #expect(KanbanHallucinationGate.from("pending") == .pending)
    #expect(KanbanHallucinationGate.from("verified") == .verified)
    #expect(KanbanHallucinationGate.from("REJECTED") == .rejected)  // case-insensitive
    #expect(KanbanHallucinationGate.from(nil) == nil)
    #expect(KanbanHallucinationGate.from("") == nil)
    #expect(KanbanHallucinationGate.from("quarantined") == nil)  // unknown returns nil
}
```

#### 12e. New test — KanbanCreateRequest argv carries `--max-retries`:

```swift
@Test func createRequestArgvIncludesMaxRetries() {
    let req = KanbanCreateRequest(title: "t", maxRetries: 5)
    let argv = req.argv()
    #expect(argv.contains("--max-retries"))
    #expect(argv.contains("5"))
}

@Test func createRequestArgvOmitsMaxRetriesWhenAbsent() {
    let req = KanbanCreateRequest(title: "t")
    let argv = req.argv()
    #expect(!argv.contains("--max-retries"))
}
```

#### 12f. New test — Run with diagnostics decodes:

```swift
@Test func decodeRunWithDiagnostics() throws {
    let json = """
    {
      "id": 1,
      "task_id": "t_x",
      "status": "failed",
      "started_at": 1778160000,
      "ended_at": 1778160300,
      "outcome": "crashed",
      "error": "OOM",
      "diagnostics": [
        {"kind": "retry_cap_hit", "message": "3/3 retries exhausted"}
      ]
    }
    """
    let run = try JSONDecoder().decode(HermesKanbanRun.self, from: Data(json.utf8))
    #expect(run.diagnostics.count == 1)
    #expect(run.diagnostics.first?.kind == "retry_cap_hit")
}

@Test func decodeRunWithoutDiagnostics() throws {
    let json = """
    {"id": 1, "task_id": "t_x", "status": "running", "started_at": 1778160000}
    """
    let run = try JSONDecoder().decode(HermesKanbanRun.self, from: Data(json.utf8))
    #expect(run.diagnostics.isEmpty)
}
```

These tests pin the tolerant-decode contract on both sides (with new fields, without new fields). Pre-v0.13 hosts running v2.8 Scarf must keep decoding cleanly — without these tests we'd ship a regression that bites every customer not yet on Hermes v0.13.

### 13. `scarf/Packages/ScarfDesign/` — IF a multi-line text component is missing

**Why:** If `ScarfTextField` doesn't already accept an `axis: .vertical` parameter (likely the case in v2.7.5), add one OR add a `ScarfTextEditor` component to the design package so the create sheet can use the design-system token.

**Conservative approach:** Use `TextField` with `axis: .vertical` directly inside `KanbanCreateSheet`, styled to match `ScarfTextField` chrome (background, border, padding from `ScarfColor`/`ScarfRadius`/`ScarfSpace`). Defer adding a new design-system component to a follow-up — design-system additions deserve their own review pass and aren't on this WS's critical path.

---

## Capability gating

All of the new Mac surface gates on `HermesCapabilities.hasKanbanDiagnostics` (already shipped in WS-1, semver `>= 0.13.0`).

### Gating decisions per surface

| Surface | Gated? | Rationale |
| --- | --- | --- |
| `max_retries` field on create sheet | Yes | Pre-v0.13 Hermes rejects `--max-retries` flag with non-zero exit. Hide the field; don't pass the flag. |
| Multi-line title input rendering | No | Multi-line input is harmless on v0.12 (the ScarfTextField is just visually taller). |
| Multi-line title submitted with `\n` | Yes | Pre-v0.13 may truncate at the first `\n`. Strip newlines client-side when `!hasKanbanDiagnostics`. |
| `max_retries` chip on inspector header | Yes | Pre-v0.13 task rows never carry `max_retries`, so `task.maxRetries` is nil — `if let` already hides it. Belt-and-suspenders: also gate on the flag. |
| Hallucination-gate banner | Yes | Pre-v0.13 task rows never carry `hallucination_gate_status`. Same `if let` belt-and-suspenders. |
| Diagnostics rendering on inspector | Yes | Pre-v0.13 tasks carry empty `diagnostics`, so the rendering would no-op. Gate explicitly anyway so a future server-side change doesn't accidentally surface partial UX on a pre-v0.13 host. |
| Auto-blocked banner with reason | Yes | Pre-v0.13 may write a similar string in a different place. Gate so we don't double-render. |
| Card hallucination dim/glyph | Yes | Same. |
| Card diagnostics dot | Yes | Same. |
| Verify / Reject buttons | Yes (hard gate) | The `kanban verify` verb (or whatever Hermes ships) doesn't exist pre-v0.13. The buttons MUST be hidden, not just disabled — a disabled button conveys "this might work later in this session" which is wrong for a capability-gated feature. |

### Plumbing

`HermesCapabilitiesStore` is already injected via `.environment(_:)` on `ContextBoundRoot` (Mac) — see CLAUDE.md "Capability gating" section. Read in views with `@Environment(HermesCapabilitiesStore.self) private var capabilities` (or whatever key is currently used; verify with the existing `Curator` feature gating).

**No new HermesCapabilities flag.** WS-1 already shipped `hasKanbanDiagnostics` covering all eight v0.13 Kanban additions in a single boolean. Resist the urge to split into `hasHallucinationGate` / `hasDiagnostics` / `hasMaxRetries` — Hermes shipped them together, and finer gating is YAGNI per the CLAUDE.md "Kanban v3" pattern.

---

## How to test

### Unit tests (KanbanModelsTests)

The test additions are listed above (§12.a–§12.f). Run:

```bash
xcodebuild -project scarf/scarf.xcodeproj \
  -scheme ScarfCore \
  -destination 'platform=macOS' \
  test
```

All v0.13 fixtures should decode AND all v0.12 fixtures should continue to decode. The two-shape pair is the critical contract.

### Manual smoke (against a real Hermes v0.13 host)

Per CLAUDE.md "remote-servers dogfooding" memory: dogfood against the Mardon Mac Mini at 192.168.0.82 — set context to that server (or run against local v0.13 install).

1. **Hallucination gate end-to-end**
   - Trigger a worker that creates a follow-up card via the agent's tooling. Server flips it to `pending`.
   - Card on board: 0.6 opacity, yellow ⚠ glyph in title row.
   - Inspector: yellow banner above body with Verify / Reject buttons.
   - Click Verify: optimistic flip — banner disappears immediately, card un-dims. Within 5s, polled state confirms `verified`. No regressions in optimistic-override clearing.
   - Click Reject (on a different pending task): comment + archive sequence runs; card disappears from active board (visible only with "Show archived").

2. **Diagnostics**
   - Trigger a worker that hits a heartbeat stall (e.g. Sleep > heartbeat interval). Verify `heartbeat_stalled` diagnostic appears under the run row in the inspector Runs tab.
   - Trigger a tool-error loop (force a tool to error 3+ times). Verify `tool_error_loop` diagnostic shows up.
   - Verify the dot-indicator on the card lights up.

3. **`max_retries`**
   - Create a task via the create sheet with Max retries = 1.
   - Verify the inspector header shows `retries: 1`.
   - Force a failure; verify the worker is auto-blocked after 1 retry (server-side behavior).
   - The chip is read-only — verify there's no edit affordance.

4. **Auto-blocked reason**
   - Trigger a worker that exits 0 without calling `kanban complete`.
   - Verify the inspector banner says "Auto-blocked" with the server's `auto_blocked_reason` verbatim.
   - Verify the card footer shows the truncated reason in red.

5. **Multi-line title**
   - In the create sheet, type a 3-line title.
   - Verify the field grows.
   - Submit. Verify on the Hermes v0.13 host the title is preserved with newlines (`hermes kanban show` JSON should round-trip them).

6. **Pre-v0.13 host (regression smoke)**
   - Switch context to a Hermes v0.12 host.
   - Verify: max-retries field hidden in create sheet; max-retries chip absent in inspector; no hallucination banner; no diagnostics rendering; create still works; existing v2.7.5 chrome unchanged.
   - Title field: type a multi-line title — submit. Verify newlines were stripped client-side (no exception on the server).

### Integration smoke

Build the app and run the existing Kanban smoke flow from `docs/PRD.md` to verify drag-drop, optimistic merge, and the per-project tenant flow are unaffected. The new code paths should not change v2.7.5 behavior on a v0.13 host that happens to have no diagnostics / hallucination signals (the dominant case in normal use).

### Example v0.13 JSON fixtures (use as test inputs and as documentation)

Drop these into `KanbanModelsTests` as inline fixtures. They're our wire-shape claim until we can validate against real CLI output during integration.

#### Task with all v0.13 fields

```json
{
  "id": "t_v013_full",
  "title": "Investigate flaky test\nReproduces only on CI",
  "body": "Repro: run the integration suite 10x.",
  "assignee": "researcher",
  "status": "blocked",
  "priority": 75,
  "tenant": "scarf:demo",
  "workspace_kind": "scratch",
  "workspace_path": "/Users/alan/.hermes/kanban/workspaces/t_v013_full",
  "created_by": "agent:claude-sonnet-4-7",
  "created_at": 1778160614,
  "skills": ["debugging"],
  "max_runtime_seconds": 1800,
  "max_retries": 3,
  "auto_blocked_reason": "worker exited (code 0) without calling `kanban complete`",
  "hallucination_gate_status": "pending",
  "diagnostics": [
    {
      "kind": "worker_exit_no_complete",
      "message": "exit code 0 with no complete call",
      "detected_at": 1778161000
    },
    {
      "kind": "heartbeat_stalled",
      "message": "no heartbeat for 4m20s (max_runtime/4 = 7m30s, slack budget exceeded)",
      "detected_at": 1778161200
    }
  ]
}
```

#### Task with no v0.13 fields (legacy v0.12 host)

```json
{
  "id": "t_v012_legacy",
  "title": "Translate doc",
  "status": "ready",
  "priority": 50,
  "skills": []
}
```

#### Run with diagnostics

```json
{
  "id": 7,
  "task_id": "t_v013_full",
  "profile": "researcher",
  "status": "failed",
  "started_at": 1778160614,
  "ended_at": 1778160914,
  "outcome": "crashed",
  "error": "subprocess died with SIGKILL",
  "summary": null,
  "diagnostics": [
    {"kind": "darwin_zombie_detected", "message": "PID 9842 left as zombie", "detected_at": 1778160916},
    {"kind": "retry_cap_hit", "message": "3/3 retries exhausted"}
  ]
}
```

---

## Open questions

1. **What's the exact CLI verb name for hallucination-gate verify / reject?** Release notes say "hallucination gate + recovery UX" but don't enumerate the verb. Best inference is `hermes kanban verify <id>` or `hermes kanban gate verify <id>`. **Action:** before implementation, run `hermes kanban --help` against a v0.13 binary and confirm. If absent (and the gate is server-flipped automatically once a worker tries to dispatch a hallucinated card), the Reject path still works (archive + comment), but Verify becomes "do nothing" and the card waits for server-side detection. Document in code comment.

2. **Where do diagnostics live on the wire — task envelope, run envelope, or both?** Release notes: "Generic diagnostics engine for task distress signals." This implies task-level. But heartbeat-stalled is a per-run signal. Best inference: per-run for in-flight signals, per-task for cross-run signals (retry cap hit). **Action:** plan handles both via `HermesKanbanTaskDetail.allDiagnostics` and per-run `run.diagnostics`. Verify against real JSON during integration.

3. **Does Hermes v0.13 expose a `set_max_retries` verb post-create?** Release notes say "Per-task `max_retries` override configuration" — ambiguous. If it's create-only (write-once like `priority`), we surface the chip read-only and document the limitation. If it's a settable field, we add an inspector edit affordance. **Action:** confirm at integration time. Plan assumes write-once (matches Hermes pattern).

4. **Failure-counter unification — does the run row need a new field?** Release notes: "Unify failure counter across spawn / timeout / crash outcomes." Best inference: server-side, the `failure_count` is a single column rather than three columns. From Scarf's view, this changes nothing — we render `outcome` (already present), and the count is implicit (count of failed runs in `runs` array). **Action:** verify at integration. If a `failure_count: Int` field shows up, decode it on `HermesKanbanRun` (already in §3) and surface in the run row label as "x/N retries" when `max_retries` is set.

5. **How does v0.13 distinguish darwin zombie from generic stale_lock?** Release notes: "Detect darwin zombie workers." Best inference: the diagnostics array includes a `darwin_zombie_detected` kind on the run. **Action:** plan renders it via the typed-mirror enum. Verify the kind string at integration.

6. **What's the default `max_retries` value?** Plan defaults the create-sheet field to 3 with a "0 = no retries. Defaults to 3." subtitle. Confirm against `hermes kanban stats --json` defaults block (or `hermes kanban --help` text) at integration. If Hermes config exposes a global default, read it and use that as the field's pre-fill.

7. **Are there sub-commands like `hermes kanban diagnose <id>`?** Release notes don't mention, but generic-diagnostics-engine framing leaves room. If such a verb exists, the inspector's diagnostics block could grow a "Run diagnostics" button to manually trigger a fresh check. **Action:** ship without; revisit when verb existence is confirmed.

---

## Out of scope (deferred — likely v2.8.x or v2.9)

- **iOS read-only counterpart** — covered by WS-9 (iOS catch-up). Render hallucination dim, max_retries chip, and auto_blocked_reason banner on the iOS detail sheet read-only. No buttons.
- **`watch` streaming** — when Hermes ships a stable `kanban watch --json` shape, replace the 5s polling loop. New flag `hasKanbanWatch` will gate the surface.
- **Within-column reorder** — still no `update --priority` verb. If Hermes ships one in a future minor, revisit.
- **In-place title / body edit** — same constraint. CLAUDE.md "Don't" list applies unchanged.
- **Cross-column drag from Done** — terminal state.
- **Diagnostics filter on the board** — could imagine "show only tasks with active diagnostics" toggle in the toolbar. Defer until we see how often the dot indicator fires in real use.
- **Bulk verify / reject** — multi-select card → verify all. Defer; the hallucination gate is rare enough that one-at-a-time UX is fine in v2.8.0.
- **Diagnostics history graph** — over time, "this task had heartbeat-stalled 3 times in 6 attempts" is a valuable signal. Defer to a v2.9 dashboard widget on top of the v0.13 stats endpoint.
- **Worker log → diagnostics correlation** — when a diagnostic fires at time T, scroll the log tab to that timestamp. Nice-to-have; defer.

---

## Estimate

**Engineering hours (one engineer, focused):**

| Block | Hours |
| --- | --- |
| Model additions (§1, §2, §3, §4, §5) — fields + tolerant decode | 3 |
| KanbanService verb additions (§6) — verify + reject (with TODO until CLI confirmed) | 2 |
| KanbanCreateSheet edits (§7) — multi-line title + max_retries field | 3 |
| KanbanInspectorPane edits (§8) — banners + diagnostics + header chip + action-bar gate | 5 |
| KanbanCardView edits (§9) — hallucination dim/glyph + auto-block sub-line + diagnostics dot | 2 |
| KanbanBoardView wiring (§10) | 1 |
| KanbanBoardViewModel edits (§11) — extended optimistic override + verify/reject methods | 3 |
| KanbanModelsTests additions (§12) | 2 |
| Capability gating audit / plumbing | 1 |
| Manual smoke (§How to test) — both v0.13 host and v0.12 host | 2 |
| Code review + revisions | 3 |
| **Total** | **~27 hours (≈3.5 working days)** |

**Confidence:** medium-high. The model additions and view edits are mechanical given v2.7.5's existing scaffolding (the optimistic-override pattern, the inspector pane structure, the tolerant-decode tests). The single biggest risk is the hallucination-gate CLI verb name (Open Question #1) — if Hermes shipped a verb name we can't infer, the Verify path is a stub until we see the binary's `--help`. The Reject path always works (archive + comment) so the recovery UX is functional even with #1 unresolved.

**Critical-path dependency:** none. WS-1 already shipped `hasKanbanDiagnostics`. WS-3 has no other workstream dependency.

**Risk register:**

- **Wire-shape mismatch.** If our inferred JSON shape is wrong (e.g. `diagnostics` is keyed `signals` on the wire), the model code is wrong. Mitigation: tolerant decode + integration smoke against a real v0.13 host before merging. Add a fixture-from-real-output test once we have stdout from `hermes kanban show --json` on a v0.13 host.
- **Verb-name uncertainty.** See Open Question #1. Mitigation: stub method with TODO + comment-only archive flow for Reject; ship Verify behind a feature gate in the inspector if needed.
- **Optimistic-override regressions.** Extending the override mechanism to cover hallucination state could destabilize the existing drag-drop optimistic flow. Mitigation: write the struct refactor as a single commit, run the existing transition-planner tests, then write the new tests.
- **Pre-v0.13 silent regression.** The most damaging failure mode is a v0.12 user upgrading Scarf and seeing the board stop loading. Mitigation: §12 tests pin the v0.12 contract; the gating audit table covers each surface; manual smoke against a v0.12 host is a P0 step.

---

## Appendix A — File-touch summary

| File | Purpose | Lines changed (estimate) |
| --- | --- | --- |
| `Models/HermesKanbanTask.swift` | +4 fields, init/decoder updates, +1 enum | ~50 |
| `Models/HermesKanbanDiagnostic.swift` | NEW model + enum mirror | ~80 (new file) |
| `Models/HermesKanbanRun.swift` | +1 field, init/decoder/encoder updates | ~15 |
| `Models/HermesKanbanTaskDetail.swift` | +1 envelope-level diagnostics field, +1 helper | ~20 |
| `Models/KanbanCreateRequest.swift` | +1 field, +1 argv branch | ~10 |
| `Services/KanbanService.swift` | +2 verb methods (verify, reject) | ~30 |
| `Tests/KanbanModelsTests.swift` | +6 tests | ~120 |
| `Features/Kanban/Views/KanbanCreateSheet.swift` | multi-line title + max-retries field + capability plumbing | ~80 |
| `Features/Kanban/Views/KanbanInspectorPane.swift` | hallucination banner + diagnostics + header chip + auto-block reason + action-bar gate | ~150 |
| `Features/Kanban/Views/KanbanCardView.swift` | hallucination dim/glyph + auto-block sub-line + diagnostics dot | ~50 |
| `Features/Kanban/Views/KanbanBoardView.swift` | wire new closures | ~10 |
| `Features/Kanban/ViewModels/KanbanBoardViewModel.swift` | struct override refactor + verify/reject methods + merge update | ~80 |

**Total: 12 files (1 new), roughly 690 lines changed.**

---

## Appendix B — Wiring diagram

```
  Hermes v0.13 binary
         │
         │ hermes kanban show --json
         ▼
  KanbanService.show ─┐
                      │
  hermes kanban runs  │
         │            │
         ▼            ▼
  HermesKanbanRun  HermesKanbanTaskDetail
  + diagnostics    + task.diagnostics
                   + envelope.diagnostics
                   + task.maxRetries
                   + task.autoBlockedReason
                   + task.hallucinationGateStatus
         │
         │ KanbanBoardViewModel polls every 5s
         ▼
  optimisticOverrides (struct, not String)
   { taskId: { status?, hallucinationGate? } }
         │
         ▼
  KanbanBoardView ─── KanbanCardView (dim/glyph/dot/sub-line)
         │
         └── KanbanInspectorPane
               ├── headerChips (+ retries chip)
               ├── hallucinationBanner (Verify / Reject)
               ├── autoBlockedBanner
               ├── failureBanner (existing)
               ├── unassignedBanner (existing)
               ├── runsTab (+ per-run diagnostics)
               └── actionBar (suppressed when hallucination=pending)
```

---

## Appendix C — UX copy register

Centralizing the user-facing strings here so a copy review pass can run before implementation.

| Surface | Copy |
| --- | --- |
| Hallucination banner title | "Created by a worker — verify before running" |
| Hallucination banner body | "A worker claimed it created this card; Hermes hasn't confirmed the underlying work exists. Verify the card matches a real follow-up, or reject if it's a hallucinated reference." |
| Hallucination banner Verify button | "Verify" |
| Hallucination banner Reject button | "Reject" |
| Card hallucination glyph tooltip | "Worker-created — verify before running" |
| Auto-blocked banner title | "Auto-blocked" |
| Auto-blocked banner body | (server-supplied verbatim from `auto_blocked_reason`) |
| Max retries chip | `retries: N` |
| Max retries chip tooltip | "Max retries set at create time. Hermes has no update verb — re-create the task to change this." |
| Diagnostics block label | "Diagnostics" (uppercase caption style) |
| Card diagnostics dot tooltip | "N diagnostic signal(s)" |
| Create sheet max-retries section header | "Max retries" |
| Create sheet max-retries subtitle | "0 = no retries. Defaults to 3." |
| Reject confirm-comment text | "Rejected as hallucinated (no underlying work)." |

---

## Appendix D — Why no dedicated `set_max_retries` verb is right

Hermes's design pattern is consistent: anything that affects how a worker is dispatched is set at `create` time and immutable afterward. `priority`, `title`, `body`, `tenant`, `max_runtime_seconds`, and now `max_retries` all follow this pattern.

The reasoning is dispatcher-correctness: a worker spawning at moment T captures the configuration at moment T. Mutating `max_retries` post-spawn would either:
- Apply only to *future* retry attempts (confusing because the user thinks they raised the cap), OR
- Apply retroactively (confusing because the dispatcher's internal counter mid-stream needs a flush).

Hermes resolves this by making the question moot — the field is write-once. Scarf's posture should be: surface the value clearly, explain the limitation, and make re-create-with-new-value cheap. We already meet the third bar (the create sheet pre-fills sensible defaults). For v2.8.0 we surface the value (max_retries chip in inspector header) and document the limitation in tooltip copy. If there's user demand for "raise the cap on this stuck task," the right move is a "Re-create with bumped retries" inspector action that reads the existing task body / assignee / etc., archives the original, and creates a sibling — a pattern v0.12 already supports without any new verbs. Defer until v2.8.x.
