# WS-9 Plan: ScarfGo iOS catch-up (read-only mirrors of WS-2 / WS-3 / WS-4 / WS-5)

**Workstream:** WS-9 of Scarf v2.8.0
**Hermes target:** v0.13.0 (v2026.5.7)
**Capability gates consumed (already shipped in WS-1, PR #80):**
- `HermesCapabilities.hasGoals` (`>= 0.13.0`) — drives the Goal pill
- `HermesCapabilities.hasACPQueue` (`>= 0.13.0`) — read-only queue indicator stub
- `HermesCapabilities.hasKanbanDiagnostics` (`>= 0.13.0`) — diagnostics on the iOS Kanban detail sheet
- `HermesCapabilities.hasCuratorArchive` (`>= 0.13.0`) — Archived list section in the iOS Curator surface
- `HermesCapabilities.hasGoogleChatPlatform` / `hasGatewayAllowlists` / `hasGatewayBusyAckToggle` / `hasGatewayRestartNotification` (`>= 0.13.0`) — Settings → Platforms additions

**Builds on:**
- v2.7.5 iOS Kanban (`Scarf iOS/Kanban/ScarfGoKanbanView.swift`, `ScarfGoKanbanDetailSheet.swift`).
- v2.7.5 iOS Curator (`Scarf iOS/Curator/CuratorView.swift`).
- v2.7.5 iOS Settings (`Scarf iOS/Settings/SettingsView.swift`) including `platformsSection`.
- v2.5+ iOS Chat (`Scarf iOS/Chat/ChatView.swift`) including `projectContextBar` and `transientHint`.
- WS-1 capability flags + the `.hermesCapabilities(_:)` env injection at `ScarfGoTabRoot.swift:153`.
- Phase H precedent: iOS catch-up "parity-match the Mac surfaces but skip mutating CLI verbs."

**Owner:** TBD
**Reviewers:** Alan (always); whoever owns iOS during v2.8 cycle.
**Sequencing:** WS-9 lands AFTER WS-2 / WS-3 / WS-4 / WS-5 merge to main, since it consumes their model fields, view-model state, and capability flags.

---

## Goals (read-only mirrors of WS-2 / WS-3 / WS-4 / WS-5)

WS-9 is iOS-only and **strictly read-only**. It mirrors selected Mac surfaces from earlier work-streams without introducing any iOS-side write verb. Per the v2.8.0 release plan, iOS write surfaces (Verify / Reject buttons, iOS create-task, iOS curator-archive button, iOS allowlist editor, etc.) are deferred to v2.8.x.

User-visible additions (all capability-gated, all degrade silently on pre-v0.13 hosts):

1. **Goal pill in iOS chat.** When `caps.hasGoals == true` AND `controller.vm.activeGoal != nil`, surface a "Goal: <text>" pill at the top of the chat view (mounted next to the existing folder/branch chips in `projectContextBar`). Read-only — no `/goal` slash command on iOS in v2.8.0; no clear affordance.
2. **Read-only `/queue` count chip.** When `caps.hasACPQueue == true` AND `controller.vm.queuedPrompts.count > 0`, surface a small "N queued" chip in the same `projectContextBar`. No popover, no mutation. Tap is a no-op (or shows a sheet listing the previews — see Open Question #2).
3. **Kanban v0.13 diagnostics on iOS detail sheet.** Extend `ScarfGoKanbanDetailSheet` to render `max_retries`, `auto_blocked_reason`, `hallucination_gate_status`, and the diagnostics array. NO Verify / Reject buttons; the hallucination state is rendered as a badge with the copy "Worker-created — verify on Mac" (since iOS can't verify in v2.8.0).
4. **iOS Curator Archived section.** Append a read-only "Archived" section to the existing `Scarf iOS/Curator/CuratorView.swift`. Per-row: name, kind, archived-date, optional reason (sized small for thumb scrolling). NO Restore / Prune-this / Prune-all buttons. Empty-state copy points the user to the Mac app for restore.
5. **iOS Settings v0.13 features-active badge.** When `caps.semver >= 0.13.0`, surface a small read-only "v0.13 features active" `ScarfBadge` at the top of `SettingsView` with a "Learn more" tap action that opens an action sheet listing the new features.
6. **iOS Platforms read-only mirror (extension to existing `platformsSection`).** Add a Google Chat read-only row, a "Restart notifications" yes/no row, a "Busy ack" yes/no row, and a per-platform allowlist chip-row ("3 allowed channels: …, 4 allowed chats: …"). No editing — that's a Mac-only surface in v2.8.0.

### Non-goals (explicitly deferred)

- **iOS write surfaces** (Verify / Reject, Create Task, Archive Skill, Prune, Allowlist editor, `/goal`, `/queue` send) — deferred to v2.8.x. Per Phase H precedent.
- **iOS Curator surface from scratch** — out of scope. iOS already has `CuratorView.swift`; WS-9 only adds the Archived list. (See Open Question #1 for what the user prompt anticipated.)
- **iOS Gateway/Platforms surface from scratch** — out of scope. iOS Settings already has `platformsSection` (lines 280-288 of `SettingsView.swift`); WS-9 extends it. There is **no separate iOS Gateway feature module** today and WS-9 does not add one.
- **iOS goal/queue clear affordance** — `/goal --clear` and "Clear all queued" are write verbs; deferred.
- **iOS Kanban verify on tap** — iOS Kanban is read-only and stays read-only in v2.8.0.
- **iOS Curator Run Now blocking + progress (synchronous run)** — that's a write change in scope of WS-4, not WS-9. iOS keeps fire-and-forget `runNow` regardless of v0.13.

---

## Existing iOS surface inventory

(Verified by walking `Scarf iOS/` at plan time.)

| iOS dir | Files | Mac counterpart |
|---|---|---|
| `App/` | `ScarfIOSApp.swift`, `ScarfGoCoordinator.swift`, `ScarfGoTabRoot.swift`, `Theme/` | `scarfApp.swift`, `AppCoordinator.swift`, `SidebarView.swift` |
| `Chat/` | `ChatView.swift`, `ChatContentFormatter.swift`, `ProjectPickerSheet.swift`, `ProjectSlashCommandsBrowser.swift` | `Features/Chat/` |
| `Components/` | `FlowLayout.swift`, `HermesVersionBanner.swift` | (cross-feature shared) |
| `Cron/` | (read-only views) | `Features/Cron/` |
| **`Curator/`** | **`CuratorView.swift` (read-mostly, runNow/pause/resume/pin/unpin/restore wired)** | `Features/Curator/` |
| `Dashboard/` | iOS dashboard views | `Features/Dashboard/` |
| **`Kanban/`** | **`ScarfGoKanbanView.swift`, `ScarfGoKanbanDetailSheet.swift` (5-column horizontal-paged Picker, read-only)** | `Features/Kanban/` |
| `Memory/` | (read-only views) | `Features/Memory/` |
| `Notifications/` | `APNSTokenStore.swift`, `NotificationRouter.swift` | `Core/Services/Notifications*` |
| `Onboarding/` | (first-run wizard) | `Features/Onboarding/` |
| `Plugins/` | `PluginsView.swift` (Phase H read-only) | `Features/Plugins/` |
| `Profiles/` | `ProfilesView.swift` (Phase H read-only) | `Features/Profiles/` |
| `Projects/` | iOS project surfaces (incl. `ProjectDetailView.swift`) | `Features/Projects/` |
| `Servers/` | server-list + connect surfaces | `Features/Servers/` |
| **`Settings/`** | **`SettingsView.swift`, `SettingEditorSheet.swift`, `ScarfMonDiagnosticsView.swift`** | `Features/Settings/` |
| `Skills/` | iOS Skills surface | `Features/Skills/` |
| `Webhooks/` | `WebhooksView.swift` (Phase H read-only) | `Features/Webhooks/` |

**Surfaces that DO NOT exist on iOS today:**
- No standalone `Scarf iOS/Gateway/` or `Scarf iOS/Platforms/` directory. iOS surfaces gateway / platform configuration through `SettingsView.platformsSection`. WS-9 mirror item 6 extends that section; it does NOT spin up a new feature module.
- No iOS goal / queue surface. WS-2 lays the VM-side scaffolding (`activeGoal`, `queuedPrompts` on the shared `RichChatViewModel` in ScarfCore); WS-9 is what surfaces it on iOS.
- No iOS dedicated "What's new in v0.13" feature surface. The "v0.13 features active" badge in mirror item 5 is the only entry point WS-9 adds.

**Capability injection (verified):**
- `ScarfGoTabRoot.swift:52` constructs a `HermesCapabilitiesStore` per server connection.
- `ScarfGoTabRoot.swift:153` calls `.hermesCapabilities(capabilities)` on the tab view.
- All iOS feature views read with `@Environment(\.hermesCapabilities) private var capabilitiesStore` (see `ChatView.swift:30`, `ProjectDetailView.swift:22`, `Components/HermesVersionBanner.swift:14`).
- WS-9 reuses the same env injection — no new plumbing required.

---

## 1. iOS Goal pill (mirror WS-2)

**Source path read.** The goal text lives on `RichChatViewModel.activeGoal: HermesActiveGoal?` (added in WS-2 — see WS-2 plan §3 "Active goal state"). iOS reads the same VM through `ChatController.vm` (the shared ScarfCore VM). No new ScarfCore field is needed; the WS-2 plumbing flows automatically into iOS.

### File: `Scarf iOS/Chat/ChatView.swift`

#### 1a. Read the capability + goal state in `body`

iOS already injects `@Environment(\.hermesCapabilities) private var capabilitiesStore` at line 30. Add a derived flag near the existing `supportsImagePrompts` computed property (lines 44-46):

```swift
private var supportsActiveGoal: Bool {
    capabilitiesStore?.capabilities.hasGoals ?? false
}

private var supportsACPQueue: Bool {
    capabilitiesStore?.capabilities.hasACPQueue ?? false
}
```

#### 1b. Mount the goal pill alongside the project chip

The `projectContextBar` (lines 832-892) currently renders only when there's an active project. Adding the goal pill INSIDE that bar would mean a pill-less goal can't render in non-project chats. Solution: split the conditional. Render `projectContextBar` when `projectName != nil OR supportsActiveGoal && controller.vm.activeGoal != nil OR supportsACPQueue && !controller.vm.queuedPrompts.isEmpty`. The bar's tinted-strip background works for any of these states.

```swift
@ViewBuilder
private var projectContextBar: some View {
    let hasProject = (controller.currentProjectName?.isEmpty == false)
    let hasGoal = supportsActiveGoal && controller.vm.activeGoal != nil
    let hasQueue = supportsACPQueue && !controller.vm.queuedPrompts.isEmpty
    if hasProject || hasGoal || hasQueue {
        HStack(spacing: 8) {
            if hasProject { /* existing project chip */ }
            if hasGoal { goalChip }
            if hasQueue { queueChip }
            Spacer()
            if hasProject && !controller.vm.projectScopedCommands.isEmpty {
                /* existing slash-commands chip */
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 6)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(.tint.opacity(0.1))
    }
}

@ViewBuilder
private var goalChip: some View {
    if let goal = controller.vm.activeGoal {
        Label(truncatedGoalText(goal.text), systemImage: "scope")
            .labelStyle(.titleAndIcon)
            .font(.subheadline)        // semantic — Dynamic Type works
            .foregroundStyle(ScarfColor.info)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(ScarfColor.info.opacity(0.16), in: Capsule())
            .lineLimit(1)
            .accessibilityLabel("Goal locked: \(goal.text)")
    }
}

private func truncatedGoalText(_ text: String) -> String {
    text.count <= 28 ? text : String(text.prefix(25)) + "…"
}
```

**Font choice (per CLAUDE.md iOS rules).** Use semantic `.subheadline` because the goal text IS content (the user reads it to recall what they locked the agent on). Per CLAUDE.md "Decision tree per text element: 'is this read for content?' → semantic token. 'Is this chrome / a label / a badge?' → ScarfFont." If the design review pushes back and prefers a fixed-display chip look, switch the inner `Text` to `ScarfFont.captionStrong`; the surrounding pill chrome stays the same.

**Color choice.** `ScarfColor.info` matches Mac's WS-2 plan (informational state, not warning, not error). Keeps the pill visually distinct from the green "success" branch chip and the orange tinted-strip background of `projectContextBar`.

**Truncation.** 25-char prefix matches the iPhone 14 portrait width budget for a chip beside a project name. The full goal text is in the accessibility label (VoiceOver users get the full string).

#### 1c. NO clear affordance

iOS does not get a "Clear goal" gesture in v2.8.0. The pill is purely informational. Tapping is a no-op. Users running `/goal --clear` from the Mac will see the iOS pill drop on the next polled state refresh (or whenever `controller.vm.activeGoal` updates — most likely on the next ACP event).

---

## 2. iOS Kanban v0.13 diagnostics (mirror WS-3)

**Source paths read.** All four new fields land on `HermesKanbanTask` (WS-3 plan §1):
- `task.maxRetries: Int?`
- `task.autoBlockedReason: String?`
- `task.hallucinationGateStatus: String?` → wrap in `KanbanHallucinationGate.from(_:)`
- `task.diagnostics: [HermesKanbanDiagnostic]`

The per-run shape adds `run.diagnostics: [HermesKanbanDiagnostic]` (WS-3 plan §3). The typed-mirror enums `KanbanHallucinationGate` and `KanbanDiagnosticKind` are added in ScarfCore and consumable from iOS by `import ScarfCore`.

### File: `Scarf iOS/Kanban/ScarfGoKanbanDetailSheet.swift`

#### 2a. Capability gate

Add `@Environment(\.hermesCapabilities) private var capabilitiesStore` at the top of the struct alongside the existing state (line ~17). Compute once in `body`:

```swift
private var diagnosticsAvailable: Bool {
    capabilitiesStore?.capabilities.hasKanbanDiagnostics ?? false
}
```

Defensive default to `false` so a missing capability store (preview, smoke test) renders the v2.7.5 sheet unchanged.

#### 2b. Header chip row — add `max_retries` chip

Update `headerCard(_:)` (lines 91-111). Insert between the workspace-kind badge and the tenant badge, gated on `diagnosticsAvailable`:

```swift
if diagnosticsAvailable, let maxRetries = task.maxRetries {
    ScarfBadge("retries: \(maxRetries)", kind: .neutral)
        .accessibilityLabel("Max retries \(maxRetries)")
}
```

Tooltip on iOS is the accessibility label (no hover). No tap action; this is purely informational.

#### 2c. Header chip row — add hallucination-gate badge

Below the existing badge row, insert a NEW row when `KanbanHallucinationGate.from(task.hallucinationGateStatus) == .pending`:

```swift
if diagnosticsAvailable,
   KanbanHallucinationGate.from(task.hallucinationGateStatus) == .pending {
    HStack(spacing: 6) {
        Image(systemName: "questionmark.diamond.fill")
            .foregroundStyle(ScarfColor.warning)
        Text("Worker-created — verify on Mac")
            .font(.subheadline)        // semantic content text
            .foregroundStyle(ScarfColor.warning)
    }
    .padding(.horizontal, 10)
    .padding(.vertical, 6)
    .background(ScarfColor.warning.opacity(0.10), in: RoundedRectangle(cornerRadius: ScarfRadius.md, style: .continuous))
    .overlay(
        RoundedRectangle(cornerRadius: ScarfRadius.md, style: .continuous)
            .strokeBorder(ScarfColor.warning.opacity(0.4), lineWidth: 1)
    )
    .accessibilityHint("Open this task on the Mac app to verify or reject the worker's claim.")
}
```

**Copy choice.** "Worker-created — verify on Mac" is intentional: it surfaces the gate status AND tells the user where the action lives. This is the read-only iOS substitute for Mac's Verify / Reject buttons (which require write CLI verbs deferred to v2.8.x).

**Render order.** Hallucination badge sits BELOW the chip row but ABOVE the markdown body, so users see the worker-created flag before reading the (potentially hallucinated) body content.

#### 2d. Auto-blocked banner

In `headerCard` after the priority line, when status is `blocked` AND `task.autoBlockedReason` is non-empty:

```swift
if diagnosticsAvailable,
   KanbanStatus.from(task.status) == .blocked,
   let reason = task.autoBlockedReason, !reason.isEmpty {
    HStack(alignment: .top, spacing: 8) {
        Image(systemName: "exclamationmark.octagon.fill")
            .foregroundStyle(ScarfColor.danger)
        VStack(alignment: .leading, spacing: 2) {
            Text("Auto-blocked")
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(ScarfColor.danger)
            Text(reason)
                .font(.subheadline)        // semantic — server-supplied verbatim
                .foregroundStyle(.secondary)
        }
    }
    .padding(10)
    .background(ScarfColor.danger.opacity(0.08), in: RoundedRectangle(cornerRadius: ScarfRadius.md, style: .continuous))
}
```

#### 2e. Task-level diagnostics block

After the markdown body block (before the Picker tab selector), render the task-level diagnostics list when non-empty:

```swift
if diagnosticsAvailable, !detail.task.diagnostics.isEmpty {
    diagnosticsBlock(detail.task.diagnostics, label: "Diagnostics")
}
```

Helper:

```swift
@ViewBuilder
private func diagnosticsBlock(_ diags: [HermesKanbanDiagnostic], label: String) -> some View {
    VStack(alignment: .leading, spacing: 6) {
        Text(label)
            .font(.caption.weight(.semibold))
            .foregroundStyle(.secondary)
        FlowLayout(spacing: 6) {        // existing primitive at Scarf iOS/Components/FlowLayout.swift
            ForEach(diags) { diag in
                let kind = KanbanDiagnosticKind.from(diag.kind)
                ScarfBadge(diag.kind, kind: kind.badgeKind)
                    .accessibilityLabel(diag.message ?? diag.kind)
            }
        }
    }
    .frame(maxWidth: .infinity, alignment: .leading)
}
```

Tap-on-badge → an expandable detail sheet that shows kind + message + timestamp. iPhone-friendly substitute for the Mac `.help()` tooltip:

```swift
ScarfBadge(diag.kind, kind: kind.badgeKind)
    .onTapGesture { selectedDiagnostic = diag }
```

Sheet binding: `.sheet(item: $selectedDiagnostic) { DiagnosticDetailSheet(diagnostic: $0) }`. The detail sheet is a simple `NavigationStack` with name + message + ISO timestamp + a "Done" toolbar button. Lightweight (~30 lines).

`HermesKanbanDiagnostic` is `Identifiable` (per WS-3 plan §2 — synthetic UUID).

#### 2f. Per-run diagnostics in the Runs tab

Update `runsSection` (lines 167-204). Inside each run row, after the optional error text, append a diagnostics block when present:

```swift
if diagnosticsAvailable, !run.diagnostics.isEmpty {
    diagnosticsBlock(run.diagnostics, label: "Run diagnostics")
        .padding(.top, 4)
}
```

Same `diagnosticsBlock` helper.

#### 2g. NO write actions

Per WS-9 contract, iOS does not expose Verify / Reject. The hallucination badge in §2c is informational. Mac's `KanbanInspectorPane.healthBanner.hallucinationBanner` (WS-3 plan §8b) wires Verify/Reject buttons; iOS does not.

---

## 3. iOS Curator Archived list (mirror WS-4) — IF iOS Curator exists

**Confirmed:** iOS Curator surface exists at `Scarf iOS/Curator/CuratorView.swift` (read-mostly, with runNow / pause / resume / pin / unpin actions). **In scope.**

**Source paths read.** WS-4 introduces:
- `HermesCuratorArchivedSkill` model (WS-4 plan "New types / fields")
- `CuratorService.listArchived() async throws -> [HermesCuratorArchivedSkill]` (WS-4 plan §"New files")
- `CuratorViewModel.archivedSkills: [HermesCuratorArchivedSkill]` and `loadArchive() async` (WS-4 plan §"Edited files / CuratorViewModel")

The shared `CuratorViewModel` lives in ScarfCore — iOS reuses it directly. The iOS `CuratorView` already constructs it at line 18. No iOS-side ScarfCore changes required.

### File: `Scarf iOS/Curator/CuratorView.swift`

#### 3a. Capability gate

Add `@Environment(\.hermesCapabilities) private var capabilitiesStore` at the top of the struct. Compute once in `body`:

```swift
private var archiveAvailable: Bool {
    capabilitiesStore?.capabilities.hasCuratorArchive ?? false
}
```

#### 3b. Wire `loadArchive()` into the existing `.task`

Update the existing `.task { await viewModel.load() }` (line 92) to also load the archive when capability allows:

```swift
.task {
    await viewModel.load()
    if archiveAvailable {
        await viewModel.loadArchive()
    }
}
.refreshable {
    await viewModel.load()
    if archiveAvailable {
        await viewModel.loadArchive()
    }
}
```

#### 3c. Add the Archived section

After the "Last report" section (lines 74-80) and before the trailing modifiers, render the new section gated on `archiveAvailable`:

```swift
if archiveAvailable {
    archivedSection
}
```

Helper:

```swift
@ViewBuilder
private var archivedSection: some View {
    Section {
        if viewModel.archivedSkills.isEmpty {
            Text("No archived skills — Curator will move stale skills here after the next review cycle.")
                .font(.callout)
                .foregroundStyle(.secondary)
        } else {
            ForEach(viewModel.archivedSkills) { skill in
                VStack(alignment: .leading, spacing: 4) {
                    HStack {
                        Text(skill.name)
                            .font(.body)        // semantic — content
                            .lineLimit(1)
                        Spacer()
                        if let category = skill.category, !category.isEmpty {
                            ScarfBadge(category, kind: .neutral)
                        }
                    }
                    HStack(spacing: 6) {
                        if let reason = skill.reason, !reason.isEmpty {
                            Text(reason)
                                .font(.caption)        // semantic — content
                                .foregroundStyle(.secondary)
                                .lineLimit(2)
                        }
                        Spacer()
                        Text(skill.archivedAtLabel)
                            .font(.caption2)
                            .foregroundStyle(.tertiary)
                    }
                    if let size = skill.sizeBytes, size > 0 {
                        Text(skill.sizeLabel)
                            .font(.caption2)
                            .foregroundStyle(.tertiary)
                    }
                }
            }
        }
    } header: {
        Text("Archived")
    } footer: {
        if !viewModel.archivedSkills.isEmpty {
            Text("Restore or prune archived skills from the Mac app.")
                .font(.caption)
        }
    }
}
```

**Copy.** Empty-state mirrors Mac's empty-state copy so the wiki / docs only need one phrasing. The "Restore or prune from the Mac app" footer is the read-only signpost.

**Font choice.** Skill name + reason → semantic `.body` / `.caption` (read for content). Category badge stays `ScarfBadge` (chrome). Date and size → `.caption2` (chrome metadata).

#### 3d. NO write actions

No per-row Restore button (WS-4 Mac surface adds this — iOS does not). No Prune All. The `CuratorRestoreSheet` Mac fallback for v0.12 hosts does NOT have an iOS counterpart and WS-9 does not introduce one. iOS users wanting to restore an archived skill use the Mac app — that's documented in the section footer.

---

## 4. iOS Gateway / Platforms read-only mirror (mirror WS-5) — extending existing iOS Settings → Platforms

**Investigation result:** iOS does NOT have a separate `Gateway/` or `Platforms/` directory. Gateway / platform configuration is surfaced through `SettingsView.platformsSection` (lines 280-288). WS-9 extends this section rather than spinning up a new feature module.

**Caveat.** WS-5's plan markdown does not yet exist at `scarf/docs/v2.8/WS-5-gateway-v0.13-plan.md` (verified — the dir contains WS-2/3/4/6/7/8 only). The Mac-side WS-5 plan is forthcoming. WS-9 is forced to make best-inference assumptions about the Mac-side model field names. The capability flags themselves DO exist (`hasGoogleChatPlatform`, `hasGatewayAllowlists`, `hasGatewayBusyAckToggle`, `hasGatewayRestartNotification`, `hasGatewayList`) and the surface contract per the user prompt is:
- Show Google Chat as a new platform entry (read-only)
- Show allowlists as read-only chip-rows ("3 allowed channels: ..., 4 allowed chats: ...")
- Show platform-specific toggles as read-only state badges ("Restart notifications: ON", "Busy ack: OFF")

WS-9 mirrors that contract. Concrete model fields are flagged in Open Questions §3 below — the implementer should sync with the WS-5 author before merging.

### File: `Scarf iOS/Settings/SettingsView.swift`

#### 4a. Capability gate

Add the env-injected capability store (it's not currently read in `SettingsView`):

```swift
@Environment(\.hermesCapabilities) private var capabilitiesStore

private var caps: HermesCapabilities {
    capabilitiesStore?.capabilities ?? .empty
}
```

#### 4b. Extend `platformsSection`

The current section (lines 280-288) renders five rows: Discord require-mention, Discord auto-thread, Telegram require-mention, Slack reply-to-mode, Matrix require-mention. WS-9 appends:

```swift
@ViewBuilder
private var platformsSection: some View {
    Section("Platforms") {
        // Existing rows (lines 282-286) — UNCHANGED.
        yesNoRow("Discord: require mention", vm.config.discord.requireMention)
        yesNoRow("Discord: auto-thread", vm.config.discord.autoThread)
        yesNoRow("Telegram: require mention", vm.config.telegram.requireMention)
        LabeledContent("Slack: reply mode", value: vm.config.slack.replyToMode)
        yesNoRow("Matrix: require mention", vm.config.matrix.requireMention)

        // v0.13 additions (gated).
        if caps.hasGoogleChatPlatform {
            googleChatSubsection
        }
        if caps.hasGatewayBusyAckToggle {
            yesNoRow("Gateway: busy ack", vm.config.gateway.busyAckEnabled)
        }
        if caps.hasGatewayRestartNotification {
            yesNoRow("Gateway: restart notification", vm.config.gateway.restartNotificationEnabled)
        }
        if caps.hasGatewayAllowlists {
            allowlistsSubsection
        }
    }
}
```

**Field-name caveat.** The exact field names on `HermesConfig.gateway.*` and `HermesConfig.googleChat.*` are TBD by WS-5. Provisional field names used above (`busyAckEnabled`, `restartNotificationEnabled`, `googleChat.requireMention`, etc.) MUST be aligned with the WS-5 model definitions before this code lands. See Open Questions §3.

#### 4c. Google Chat subsection

```swift
@ViewBuilder
private var googleChatSubsection: some View {
    yesNoRow("Google Chat: require mention", vm.config.googleChat.requireMention)
    if let space = vm.config.googleChat.defaultSpace, !space.isEmpty {
        LabeledContent("Google Chat: default space", value: space)
    }
}
```

#### 4d. Allowlists subsection — chip-row summaries

Read-only, summarized counts. Per the user prompt: "3 allowed channels: …, 4 allowed chats: …". On iOS the summary is collapsed (full lists are wide and a SwiftUI `List` row is narrow). Shape:

```swift
@ViewBuilder
private var allowlistsSubsection: some View {
    if let channels = vm.config.gateway.allowedChannels, !channels.isEmpty {
        DisclosureGroup {
            ForEach(channels, id: \.self) { ch in
                Text(ch)
                    .font(.callout.monospaced())
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }
        } label: {
            LabeledContent("Allowed channels") {
                Text("\(channels.count)")
                    .font(.callout)
                    .foregroundStyle(.secondary)
            }
        }
    }
    if let chats = vm.config.gateway.allowedChats, !chats.isEmpty {
        DisclosureGroup {
            ForEach(chats, id: \.self) { chat in
                Text(chat)
                    .font(.callout.monospaced())
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }
        } label: {
            LabeledContent("Allowed chats") {
                Text("\(chats.count)")
                    .font(.callout)
                    .foregroundStyle(.secondary)
            }
        }
    }
}
```

**UI choice.** `DisclosureGroup` with the count in the label collapses well on iPhone (default-collapsed; the user can tap to expand). Avoids a wall-of-text in a small-screen list. No tap-to-edit (read-only).

#### 4e. NO write actions on iOS Platforms

No editor sheet for Google Chat. No allowlist editor. No toggle switches that send `hermes config set`. The existing `quickEditsSection` (lines 84-117) does drive `setSetting(key, value)` for "v1Editable" specs — WS-9 does NOT add the v0.13 platform fields to `SettingSpec.v1Editable`. That's a Mac-only concern in v2.8.0.

---

## 5. iOS v0.13 features-active badge (Settings)

### File: `Scarf iOS/Settings/SettingsView.swift`

#### 5a. Capability check — semver, not a single flag

Per the prompt: "Capability-gate on `caps.semver >= 0.13.0`." The `HermesCapabilities` struct (verified at `Packages/ScarfCore/Sources/ScarfCore/Services/HermesCapabilities.swift`) exposes `atLeastSemver(_:_:_:)` — a private helper. The simplest public hook is to use any one of the v0.13-gated flags as the proxy (e.g. `caps.hasGoals`) since they all resolve to the same `>= 0.13.0` threshold; or expose a new `public var isV013OrLater: Bool` on `HermesCapabilities`. Recommend the latter for clarity:

> **Coordination requirement.** WS-9 needs `HermesCapabilities.isV013OrLater: Bool { atLeastSemver(0, 13, 0) }`. If WS-1 didn't ship this, WS-9 adds it as a one-line addition to `HermesCapabilities.swift`. Cheap and keeps the badge gating honest. Alternative: piggy-back on `caps.hasGoals` and accept the semantic drift (the badge says "v0.13 features active" but is gated on the goals flag specifically). Recommend the new helper.

#### 5b. Mount the badge above `quickEditsSection`

```swift
var body: some View {
    List {
        if let err = vm.lastError { /* unchanged */ }

        if caps.isV013OrLater {
            v013ActiveBadgeSection
        }

        if !vm.isLoading || vm.config.model != "unknown" {
            quickEditsSection
            // ... rest unchanged
        }
    }
    // ... unchanged modifiers
}

@ViewBuilder
private var v013ActiveBadgeSection: some View {
    Section {
        Button {
            showV013FeaturesSheet = true
        } label: {
            HStack(spacing: 8) {
                ScarfBadge("v0.13 features active", kind: .success)
                Spacer()
                Text("Learn more")
                    .font(.caption)
                    .foregroundStyle(.tint)
                Image(systemName: "chevron.right")
                    .font(.caption)
                    .foregroundStyle(.tertiary)
            }
        }
        .buttonStyle(.plain)
    }
    .listRowBackground(ScarfColor.success.opacity(0.06))
}
```

**State.** Add `@State private var showV013FeaturesSheet = false` near the top.

**Color.** `.success` (green) — the host has new capabilities, framing as positive. Distinct from the warning-tinted error banner above it.

#### 5c. "Learn more" sheet

```swift
.sheet(isPresented: $showV013FeaturesSheet) {
    V013FeaturesSheet()
}
```

New file `Scarf iOS/Settings/V013FeaturesSheet.swift` (~80 lines):

```swift
import SwiftUI
import ScarfDesign

struct V013FeaturesSheet: View {
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            List {
                Section {
                    featureRow(
                        icon: "scope",
                        title: "Persistent goals",
                        description: "Type /goal <text> in chat to lock the agent on a target across turns. Mac only in v2.8."
                    )
                    featureRow(
                        icon: "tray.full",
                        title: "ACP /queue",
                        description: "Queue prompts to run after the current turn finishes. Mac only in v2.8."
                    )
                    featureRow(
                        icon: "stethoscope",
                        title: "Kanban diagnostics",
                        description: "Worker distress signals (heartbeat stalls, retry caps, zombies) surface on the task detail."
                    )
                    featureRow(
                        icon: "questionmark.diamond.fill",
                        title: "Hallucination gate",
                        description: "Worker-created cards are flagged for verify/reject. Verify on the Mac app."
                    )
                    featureRow(
                        icon: "archivebox",
                        title: "Curator archive",
                        description: "Stale skills move to an Archived list. Restore or prune from the Mac app."
                    )
                    featureRow(
                        icon: "bubble.left.and.bubble.right",
                        title: "Google Chat platform",
                        description: "New gateway target — configure on the Mac app."
                    )
                } header: {
                    Text("What's new in v0.13")
                } footer: {
                    Text("This iOS release surfaces v0.13 features read-only. Editing lives in the Mac app for v2.8.")
                        .font(.caption)
                }
            }
            .navigationTitle("v0.13 features")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Done") { dismiss() }
                }
            }
        }
    }

    private func featureRow(icon: String, title: String, description: String) -> some View {
        HStack(alignment: .top, spacing: 12) {
            Image(systemName: icon)
                .foregroundStyle(.tint)
                .font(.title3)
                .frame(width: 28)
            VStack(alignment: .leading, spacing: 4) {
                Text(title).font(.body.weight(.semibold))
                Text(description)
                    .font(.callout)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(.vertical, 4)
    }
}
```

**Copy is the load-bearing piece.** Each row is one sentence; the read-only-on-iOS framing is in the section footer. No deep links to the relevant tab — that's a v2.8.x polish, not WS-9.

---

## Coordination with WS-2 / WS-3 / WS-4 / WS-5

WS-9 consumes models / fields / capability flags from earlier work-streams. **WS-9 must land AFTER all of them merge to main.**

| Consumed surface | Source WS | Consumed at |
|---|---|---|
| `HermesActiveGoal` model | WS-2 | iOS goal pill (§1) |
| `HermesQueuedPrompt` model | WS-2 | iOS queue chip (§1, no popover) |
| `RichChatViewModel.activeGoal` observable | WS-2 | iOS goal pill (§1) |
| `RichChatViewModel.queuedPrompts` observable | WS-2 | iOS queue chip (§1) |
| `HermesCapabilities.hasGoals` | WS-1 | iOS chat (§1) |
| `HermesCapabilities.hasACPQueue` | WS-1 | iOS chat (§1) |
| `HermesKanbanTask.maxRetries` | WS-3 | iOS Kanban detail (§2b) |
| `HermesKanbanTask.autoBlockedReason` | WS-3 | iOS Kanban detail (§2d) |
| `HermesKanbanTask.hallucinationGateStatus` + `KanbanHallucinationGate` | WS-3 | iOS Kanban detail (§2c) |
| `HermesKanbanTask.diagnostics` + `HermesKanbanDiagnostic` + `KanbanDiagnosticKind` | WS-3 | iOS Kanban detail (§2e–§2f) |
| `HermesKanbanRun.diagnostics` | WS-3 | iOS Kanban detail (§2f) |
| `HermesCapabilities.hasKanbanDiagnostics` | WS-1 | iOS Kanban detail (§2a) |
| `HermesCuratorArchivedSkill` model | WS-4 | iOS Curator (§3) |
| `CuratorViewModel.archivedSkills` + `loadArchive()` | WS-4 | iOS Curator (§3) |
| `CuratorService.listArchived()` | WS-4 | (transitively via VM in §3) |
| `HermesCapabilities.hasCuratorArchive` | WS-1 | iOS Curator (§3) |
| `HermesConfig.gateway.allowedChannels` / `.allowedChats` (TBD field names) | WS-5 | iOS Settings (§4d) |
| `HermesConfig.gateway.busyAckEnabled` / `.restartNotificationEnabled` (TBD) | WS-5 | iOS Settings (§4b–§4c) |
| `HermesConfig.googleChat.*` (TBD shape) | WS-5 | iOS Settings (§4c) |
| `HermesCapabilities.hasGoogleChatPlatform` / `.hasGatewayAllowlists` / `.hasGatewayBusyAckToggle` / `.hasGatewayRestartNotification` | WS-1 | iOS Settings (§4) |
| `HermesCapabilities.isV013OrLater` (NEW — see §5a) | WS-1 (small follow-up) | iOS Settings badge (§5) |

### Sequencing (recommended)

1. WS-2 (Goals + queue VM scaffolding) merges → iOS chat goal pill becomes wireable.
2. WS-3 (Kanban diagnostics models) merges → iOS Kanban detail extension becomes wireable.
3. WS-4 (Curator archive service + VM state) merges → iOS Curator section becomes wireable.
4. WS-5 (Gateway / Platforms config models + capability flags consumed) merges → iOS Settings extension becomes wireable.
5. WS-9 PR opens, builds against the merged baseline, ships all five additions in one PR.

Splitting WS-9 into per-mirror PRs is overkill — each diff is small, all gated, all read-only.

### Acceptable to land WS-9 in stages

If WS-5 slips, WS-9 can ship items 1-3-4-5 first (the WS-2/3/4 mirrors plus the badge) and follow up with item 6 (Gateway/Platforms mirror) once WS-5 lands. The badge is independent of any mirror item — it can ship the moment WS-1 capability flags are in (already done).

---

## Files to change / create

| File | Status | Purpose |
|---|---|---|
| `Scarf iOS/Chat/ChatView.swift` | EDIT | Goal pill + queue chip in `projectContextBar` (§1) |
| `Scarf iOS/Kanban/ScarfGoKanbanDetailSheet.swift` | EDIT | Diagnostics + max_retries + hallucination badge + auto-blocked banner (§2) |
| `Scarf iOS/Kanban/DiagnosticDetailSheet.swift` | NEW | Tap-target sheet showing one diagnostic's full message + timestamp (§2e) |
| `Scarf iOS/Curator/CuratorView.swift` | EDIT | Archived section + capability gate + extra `.task` load (§3) |
| `Scarf iOS/Settings/SettingsView.swift` | EDIT | v0.13 badge section + Platforms section extension (§4, §5) |
| `Scarf iOS/Settings/V013FeaturesSheet.swift` | NEW | "Learn more" sheet for the v0.13-features badge (§5c) |
| `Packages/ScarfCore/Sources/ScarfCore/Services/HermesCapabilities.swift` | EDIT (1 line) | `public var isV013OrLater: Bool` helper if not already present (§5a) |

**Total:** 7 files (2 new), ~350-450 lines. ~80% of the diff is the new sheets and the iOS Kanban detail extension.

---

## Capability gating

Every WS-9 surface is hard-gated. Pre-v0.13 hosts see the v2.7.5 iOS surface unchanged.

| Surface | Gate | Pre-v0.13 behaviour |
|---|---|---|
| iOS goal pill | `caps.hasGoals && vm.activeGoal != nil` | hidden (transitive impossibility — pill goes nil because Mac doesn't write it) |
| iOS queue chip | `caps.hasACPQueue && !vm.queuedPrompts.isEmpty` | hidden |
| iOS Kanban max_retries chip | `caps.hasKanbanDiagnostics && task.maxRetries != nil` | hidden (`if let` belt-and-suspenders even if cap leaks) |
| iOS Kanban hallucination badge | `caps.hasKanbanDiagnostics && KanbanHallucinationGate.from(...) == .pending` | hidden |
| iOS Kanban auto-blocked banner | `caps.hasKanbanDiagnostics && status == .blocked && reason != nil` | hidden |
| iOS Kanban diagnostics blocks (task + run) | `caps.hasKanbanDiagnostics && !diagnostics.isEmpty` | hidden |
| iOS Curator Archived section | `caps.hasCuratorArchive` | section absent; `loadArchive()` not invoked |
| iOS Settings v0.13 badge | `caps.isV013OrLater` | section absent |
| iOS Settings Google Chat row | `caps.hasGoogleChatPlatform` | row absent |
| iOS Settings Busy ack row | `caps.hasGatewayBusyAckToggle` | row absent |
| iOS Settings Restart notification row | `caps.hasGatewayRestartNotification` | row absent |
| iOS Settings Allowlists rows | `caps.hasGatewayAllowlists` | rows absent |

**Defensive default.** Every `capabilitiesStore?.capabilities ?? .empty` resolves the absent-store case to `false` for every flag. WS-1's `.empty` static is the explicit pre-v0.13 sentinel (verified — used elsewhere in iOS already at `HermesVersionBanner.swift:14`).

**No new capability flags.** WS-9 adds at most one helper (`isV013OrLater`) to `HermesCapabilities`. All other flags are already shipped by WS-1.

---

## How to test

Per CLAUDE.md "remote-servers dogfooding" memory: dogfood against the Mardon Mac Mini at 192.168.0.82 (running the v0.13 binary on the `remote-servers` branch).

### iOS simulator scenarios — v0.13 host

1. **Goal pill**
   - Open the iOS chat against a v0.13 host. Switch to the Mac, run `/goal finish v2.8 by Friday` in the same session. Switch back to iOS — within 2-3 polled state refreshes the pill should appear in `projectContextBar` with truncated text "finish v2.8 by Friday".
   - VoiceOver: focus the pill, confirm full text reads as "Goal locked: finish v2.8 by Friday".
   - Run `/goal --clear` from Mac. Confirm pill drops on iOS.
   - Without an active project (chat without `projectContextBar` triggered today), confirm the bar STILL shows when the goal pill is the only chip — i.e. the bar is no longer project-only. Without a goal AND without a project, confirm the bar stays hidden.

2. **Queue chip**
   - Trigger a long-running prompt on Mac, send `/queue summarize` while it's working. Confirm iOS shows "1 queued" chip in the bar.
   - When the Mac turn finishes and the queued prompt fires, confirm the iOS chip count decrements.

3. **Kanban diagnostics**
   - Open the iOS Kanban detail sheet for a task with `max_retries: 3`. Confirm the "retries: 3" chip shows in the header.
   - Open a task in `pending` hallucination state. Confirm the yellow "Worker-created — verify on Mac" badge appears below the chip row.
   - Open a blocked task with `auto_blocked_reason`. Confirm the red "Auto-blocked" banner shows the reason verbatim.
   - Open a task with task-level diagnostics. Confirm the chip-list renders. Tap one — confirm the detail sheet opens with kind + message + timestamp.
   - Open a task whose latest run has `darwin_zombie_detected`. Confirm the per-run diagnostics chip-list renders inside the Runs tab row.

4. **Curator Archived list**
   - On v0.13 host with no archives: confirm Archived section renders with empty-state copy.
   - On v0.13 host with 3 archives: confirm rows show name, category badge, reason, archived-at label, size. No Restore button. Footer hint visible.
   - Pull-to-refresh: confirm `loadArchive()` re-fires.

5. **iOS Settings v0.13 badge**
   - On v0.13 host: confirm the green "v0.13 features active" badge sits above the Quick edits section. Tap "Learn more" — confirm the sheet opens with 6 feature rows.
   - Tap Done — confirm dismissal.

6. **iOS Settings Platforms additions**
   - On v0.13 host with Google Chat configured: confirm the Google Chat rows show. Tap is read-only (no nav).
   - With at least 3 allowed channels and 4 allowed chats configured: confirm both DisclosureGroup rows show with the correct counts. Expand each — confirm the entries render in monospaced font.
   - With Busy ack OFF and Restart notifications ON: confirm both rows show the right yes/no labels.

### iOS simulator scenarios — pre-v0.13 host (regression smoke)

1. Connect to a Hermes v0.12 host (Mardon downgrade or local dev install).
2. Verify:
   - `projectContextBar` looks unchanged from v2.7.5 (no goal pill, no queue chip).
   - Kanban detail sheet: no max_retries chip, no hallucination badge, no auto-blocked banner, no diagnostics blocks. v2.7.5 layout intact.
   - Curator: no Archived section. Existing `runNow` / `pause` / `resume` / `pin` actions work.
   - Settings: no v0.13 badge. Platforms section shows the 5 v2.7.5 rows only.
3. Tap through every existing iOS surface to confirm no regressions.

### Dynamic Type accessibility smoke

Per CLAUDE.md: iOS clamps Dynamic Type at the scene root (`ScarfIOSApp.swift`: `.dynamicTypeSize(.xSmall ... .accessibility2)`). Verify at both extremes:

1. Settings → Accessibility → Display & Text Size → set to AX2.
2. Open chat: confirm goal pill text scales (semantic `.subheadline` should). Confirm pill chrome doesn't blow out — the truncation kicks in.
3. Open Kanban detail: confirm body text + diagnostics chip text scale. Badges (`ScarfBadge`) should NOT scale (they're chrome).
4. Open Curator Archived list: confirm skill name + reason scale. Archived-at label stays small.
5. Open Settings v0.13 sheet: confirm description text scales.
6. Switch to xSmall: confirm nothing collapses in a way that's unreadable.

### Build + test gates

- `xcodebuild -project scarf/scarf.xcodeproj -scheme "scarf mobile" -destination 'platform=iOS Simulator,name=iPhone 15' build` must succeed.
- All existing iOS UI smoke tests (if present in the target) stay green.
- New iOS-side snapshot or UI tests are NOT planned for WS-9 — the surfaces are read-only and visual; manual verification is the right pass for v2.8.0.

---

## Open questions

1. **Does iOS Curator surface exist today?** ✅ Confirmed yes. `Scarf iOS/Curator/CuratorView.swift` exists and is read-mostly with runNow / pause / resume / pin / unpin actions. WS-9 mirror item 4 (Curator Archived list) is in scope. (The user prompt anticipated this might be unknown.)

2. **iOS goal/queue chip — is the queue chip tap a no-op or does it open a previews sheet?** Recommend tap = no-op for v2.8.0 (read-only badge, mirroring the goal pill's no-op tap). A previews sheet is nice-to-have but doesn't cross the bar for v2.8 — the user can see queued prompts from the Mac app. If review pushes back, a 30-line sheet listing previews + queued-at timestamps is cheap to add.

3. **WS-5 plan does not yet exist (`scarf/docs/v2.8/WS-5-gateway-v0.13-plan.md` is missing).** The exact `HermesConfig.gateway.*` and `HermesConfig.googleChat.*` field names are TBD. **Action:** before WS-9 implementation starts, sync with the WS-5 author to align on:
   - Where do the allowlists live? `HermesConfig.gateway.allowedChannels: [String]?` or `HermesConfig.platforms.<each>.allowedChannels`?
   - Are restart-notifications and busy-ack global (one toggle) or per-platform (one per Discord/Slack/Telegram/Matrix/Google-Chat)?
   - Is "busy ack" the right wire name? Hermes might call it `busy_acknowledge` or `busy_indicator`.
   - Does Google Chat use the same `requireMention` shape as Discord/Telegram/Matrix?

   WS-9's Settings extensions (§4) are correct in shape but need the field-name patches once WS-5 confirms. The capability flags are stable.

4. **`HermesCapabilities.isV013OrLater` helper.** WS-1 may or may not have shipped this. If not, WS-9 ships a one-line addition. If `caps.hasGoals` is acceptable as a proxy (since all v0.13 flags resolve to the same threshold), the helper isn't strictly needed — but the badge copy says "v0.13 features active" so semantic alignment matters. Coordinator should pick one.

5. **`projectContextBar` re-render frequency.** Today it renders only when there's a project. After WS-9, it renders when there's a project OR a goal OR a queued prompt. The added re-render churn during streaming (every diff to `vm.activeGoal` / `vm.queuedPrompts`) may matter for ScarfMon's `chatRender` budget. **Action:** add a ScarfMon counter to the bar's body to measure during dogfooding. If churn becomes a hot-path issue, extract `goalChip` and `queueChip` into separately-scoped subviews so they re-render in isolation.

6. **Animation on pill / chip appearance.** Should the goal pill fade in when `vm.activeGoal` becomes non-nil? Recommend yes — `.transition(.opacity.combined(with: .scale(scale: 0.9)))` with a `.spring(response: 0.3, dampingFraction: 0.7)` parent animation. Keeps the bar from feeling like it pops. Apply same to the queue chip and the Kanban hallucination badge.

7. **Tap target for the Kanban hallucination badge.** Currently planned as informational-only. Should tapping it open an alert with explanation copy + a "Open in Mac app" placeholder action? Recommend NO for v2.8.0 — the on-screen "verify on Mac" copy is enough; an alert is unnecessary friction for a read-only surface.

8. **iOS deep links from the v0.13 features sheet.** Tapping a feature row could deep-link to the relevant tab (e.g. tap "Hallucination gate" → switch to Kanban tab). Recommend defer — the v2.8.0 sheet is text-only. v2.8.x can add the routing.

---

## Out of scope (deferred to v2.8.x or later)

- **iOS write surfaces** for everything WS-9 mirrors:
  - `/goal` and `/queue` send from iOS chat composer.
  - Verify / Reject buttons on the iOS Kanban detail sheet.
  - Archive / Restore / Prune on the iOS Curator surface.
  - Allowlist editor / platform toggle editor in iOS Settings.
- **Gateway/Platforms iOS feature module from scratch** (separate `Scarf iOS/Gateway/` or `Scarf iOS/Platforms/` dir). v2.8.0 keeps gateway/platform config as an extension to `SettingsView.platformsSection`.
- **iOS Curator Archive `live` updates** beyond pull-to-refresh + the existing `.task` invocation. Hermes hasn't shipped a curator-watch surface; iOS won't either.
- **iOS Kanban hallucination badge tap-to-explain alert** — recommend not adding (see Open Question #7).
- **iOS Kanban diagnostics history graph** — Mac WS-3 also defers this. iOS follows.
- **iOS deep links from v0.13 features sheet** — see Open Question #8.
- **Snapshot tests for the new iOS sheets** — manual verification is the v2.8.0 pass.
- **Localization** — every new copy string is English-only. Existing iOS surfaces aren't localized either; WS-9 stays consistent.
- **iOS Goal pill custom font / pill chrome migration to a `ScarfDesign` component** — keep inline. If Mac WS-2 lands a reusable `ScarfGoalPill` component in the design package, swap iOS to use it as a follow-up.
- **iOS goal-state persistence across app suspends** — relies on the Mac VM state being authoritative. iOS just renders what it polls. If this matters in dogfooding (user perceives a stale pill after a long suspend), revisit.
- **Telemetry counters** for new iOS surfaces (e.g. ScarfMon counter on goal-pill appearance). Add if dogfooding surfaces a perf signal; otherwise ship without.
- **Per-platform notification re-routing toggles on iOS** (e.g. "send Google Chat alerts to APNS"). Out of scope — APNS routing already lives in `Notifications/NotificationRouter.swift` and is platform-agnostic.

---

## Estimate

**Engineering hours (one engineer, focused), assuming WS-2 / WS-3 / WS-4 / WS-5 are merged to main:**

| Block | Hours |
|---|---|
| iOS chat goal pill + queue chip in `projectContextBar` (§1) | 2 |
| iOS Kanban detail sheet — chips + banners + diagnostics blocks + tap sheet (§2) | 5 |
| iOS Kanban `DiagnosticDetailSheet.swift` (NEW, ~30 LOC) | 1 |
| iOS Curator Archived section (§3) | 2 |
| iOS Settings Platforms extension + capability env injection (§4) | 3 |
| iOS Settings v0.13 badge + sheet (§5, including new sheet file) | 2 |
| `HermesCapabilities.isV013OrLater` helper (if not present) | 0.5 |
| Manual smoke on iPhone simulator (v0.13 + v0.12 hosts) + Dynamic Type pass | 3 |
| Code review + revisions | 2 |
| Buffer for WS-5 field-name alignment (Open Q #3) | 1.5 |
| **Total** | **~22 hours (≈3 working days)** |

**Confidence: medium-high.** All five items are mechanical given the existing iOS surface scaffolding (`projectContextBar`, `ScarfGoKanbanDetailSheet`, `CuratorView`, `SettingsView.platformsSection`). The only real risk is WS-5 field-name drift — captured in Open Question #3 — and it's contained to mirror item 4 (Settings → Platforms extensions). If WS-5 slips, mirror items 1-3-5 ship first; item 6 (Platforms) follows once WS-5 lands.

**Critical-path dependency:** WS-2, WS-3, WS-4, WS-5 must all be on `main` before WS-9 PR opens. WS-9 is the final "iOS catch-up" PR of the v2.8.0 release cycle.

**Risk register:**

- **WS-5 field-name drift.** Mitigated by Open Question #3 sync with the WS-5 author before implementation; Settings extensions stub clearly-named provisional field names that fail-fast at compile if WS-5 ships different names.
- **Dynamic Type churn.** Goal pill and Kanban diagnostics blocks are content-text — they scale. Verify nothing collapses at AX2; truncation strategies in §1b and the FlowLayout primitive in §2e are the v2.7.5 patterns and known-good.
- **`projectContextBar` re-render churn.** Open Question #5 captures this. Add a ScarfMon counter; revisit if dogfooding shows a hot-path issue.
- **iOS Kanban polling cadence** — the existing 5s poll picks up the new fields automatically. No new polling logic required.
- **No iOS test coverage regression.** WS-9 doesn't add tests but doesn't remove any either. The shared `RichChatViewModel` / `CuratorViewModel` / `KanbanService` tests in ScarfCore (extended by WS-2/3/4) cover the model + state-machine layer; iOS-specific UI is verified manually in v2.8.0.
