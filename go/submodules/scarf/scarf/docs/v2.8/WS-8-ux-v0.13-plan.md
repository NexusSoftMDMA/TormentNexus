# WS-8 Plan: UX polish (v0.13 small-surface additions)

Branch suggestion: `ws-8-ux-v0.13`. Depends on WS-1 (`ws-1-capabilities-v0.13`, PR #80) for the v0.13 capability flags consumed below — every change here is a leaf surface that reads from `HermesCapabilities` and degrades silently on pre-v0.13 hosts.

## Goals (what this PR ships)

Six small, mostly-independent UX additions tracking the v0.13 release notes' "everything else" bucket:

1. **Context compression count chip** in the chat status bar — `🗜 ×N` rendered alongside the existing token counter when Hermes' status feed surfaces a non-zero compression count.
2. **`/new <name>` argument hint** on the slash menu — extends `argumentHint` for the `/new` entry on v0.13+ hosts so users discover the optional name.
3. **`hermes update --yes` plumbing** — purely forward-compatible. v2.7.5 has no in-app "Update Hermes" affordance (Sparkle handles Scarf-self-update, and `hermes update` is invoked by users in their terminal). This WS adds a stub helper on `UpdaterService` (or a new `HermesUpdaterCommandBuilder` static) that the future affordance will call; the helper takes a `HermesCapabilities` and decides whether to append `--yes`. No user-visible change ships in v2.8 from this item alone — see [Out of scope](#out-of-scope).
4. **Redaction default-flip awareness** — the existing "Redact secrets in patches" toggle in `Settings → Advanced → Caching & Redaction` gets a hint footnote whose copy depends on the connected host's version (server default flipped from OFF in v0.12 → ON in v0.13).
5. **`display.language` picker** in Settings → General → Locale — 8-option enum (`en` / `zh` / `ja` / `de` / `es` / `fr` / `uk` / `tr`), persisted via `hermes config set display.language <code>`.
6. **xAI Custom Voices badge** next to the xAI TTS provider entry in Settings → Voice → Text-to-Speech (and `xai` added to the provider list — it's not currently there).

Out-of-scope items captured in [Out of scope](#out-of-scope).

## 1. Context compression count

### What v0.13 emits

Hermes v0.13 adds a context compression count to the status feed shown in the CLI / TUI. The release notes phrase it as "Show context compression count in status bar" — they don't pin the wire field name. See [Open question Q1](#open-questions) — the plan below assumes it lands on the existing `usage` blob in `session/prompt`'s response and that it's a monotonically-incrementing integer counting how many auto-compactions have run on the active session. This matches the structure of the existing token counters (also on `usage`) and means a single small extension to `ACPPromptResult` covers it.

### Files to change

#### [scarf/Packages/ScarfCore/Sources/ScarfCore/Models/ACPMessages.swift](../../Packages/ScarfCore/Sources/ScarfCore/Models/ACPMessages.swift)

`ACPPromptResult` (around line 240) gains one optional field:

```swift
public struct ACPPromptResult: Sendable {
    public let stopReason: String
    public let inputTokens: Int
    public let outputTokens: Int
    public let thoughtTokens: Int
    public let cachedReadTokens: Int
    /// Number of automatic context compactions Hermes has performed on
    /// this session so far. v0.13+ — older Hermes hosts always return 0,
    /// which the chat status bar treats as "hide chip". Optional in the
    /// wire payload; folded into a non-optional `Int` here with a 0
    /// default so the rest of the pipeline doesn't need to nil-check.
    public let compressionCount: Int

    public init(
        stopReason: String,
        inputTokens: Int,
        outputTokens: Int,
        thoughtTokens: Int,
        cachedReadTokens: Int,
        compressionCount: Int = 0
    ) { … }
}
```

Default-zero on the initializer keeps existing call sites compiling; the only mutator is `ACPClient.sendPrompt`.

#### [scarf/Packages/ScarfCore/Sources/ScarfCore/ACP/ACPClient.swift](../../Packages/ScarfCore/Sources/ScarfCore/ACP/ACPClient.swift)

`sendPrompt` (around line 311–322) gains one decode line. The exact key is the open question — encode tolerantly:

```swift
let usage = dict["usage"] as? [String: Any] ?? [:]
// Tolerate either snake_case or camelCase per the rest of the ACP
// payload's mixed conventions; whichever Hermes ships, we read.
let compression = (usage["compressionCount"] as? Int)
                ?? (usage["compression_count"] as? Int)
                ?? 0
```

Pass `compressionCount: compression` into the `ACPPromptResult` initializer.

#### [scarf/Packages/ScarfCore/Sources/ScarfCore/ViewModels/RichChatViewModel.swift](../../Packages/ScarfCore/Sources/ScarfCore/ViewModels/RichChatViewModel.swift)

Add an observable counter alongside the existing token counters (around line 228–231):

```swift
public private(set) var acpCompressionCount = 0
```

Reset to 0 in `reset()` (around line 464–470) alongside the token counters.

In `handlePromptComplete` (around line 810–813) — the same place that aggregates ACP token counts — overwrite (don't add) with the latest server value:

```swift
acpInputTokens += response.inputTokens
acpOutputTokens += response.outputTokens
acpThoughtTokens += response.thoughtTokens
acpCachedReadTokens += response.cachedReadTokens
// Compression count is a session-wide running total emitted by Hermes;
// each prompt response carries the latest value, so we replace rather
// than accumulate. Treat 0 as "no compactions yet" — the view hides
// the chip in that case.
acpCompressionCount = max(acpCompressionCount, response.compressionCount)
```

The `max(...)` guard tolerates pre-v0.13 hosts that return `0` mid-session: if the agent is upgraded server-side without restarting Scarf, the count will resume at the higher value the next time `usage` carries a real number.

#### [scarf/scarf/Features/Chat/Views/SessionInfoBar.swift](../../scarf/Features/Chat/Views/SessionInfoBar.swift)

Add one more pass-through prop alongside the existing `acpInputTokens` / `acpOutputTokens` / `acpThoughtTokens` (lines 9–11):

```swift
var acpCompressionCount: Int = 0
/// Capability snapshot for v0.13 surfaces. Defaulted so previews and
/// pre-v0.13 hosts render the v2.7.5 layout unchanged.
var capabilities: HermesCapabilities = .empty
```

Inside the `body` `HStack`, after the reasoning-tokens label and before the cost label, render the compression chip:

```swift
if capabilities.hasContextCompressionCount && acpCompressionCount > 0 {
    Label("×\(acpCompressionCount)", systemImage: "arrow.down.right.and.arrow.up.left")
        .scarfStyle(.caption)
        .foregroundStyle(ScarfColor.foregroundMuted)
        .help("Hermes auto-compacted this session's context \(acpCompressionCount) time\(acpCompressionCount == 1 ? "" : "s")")
}
```

Notes on the visual: stick to existing `Label` + `scarfStyle(.caption)` + `ScarfColor.foregroundMuted` so the chip blends with the other counters. **Don't** invent a new `ScarfBadge` style — the row's already badge-like via the surrounding `.padding(.horizontal, ScarfSpace.s4)` background, and ScarfBadge would visually overpower a passive count. Icon: `arrow.down.right.and.arrow.up.left` (the SF Symbol for compaction). If the symbol doesn't render on macOS 14.6 — which we deploy to — fall back to a Unicode box-drawing glyph or `archivebox.fill`; flag as a follow-up rather than picking now.

#### [scarf/scarf/Features/Chat/Views/ChatTranscriptPane.swift](../../scarf/Features/Chat/Views/ChatTranscriptPane.swift)

Plumb the new field plus the env-resolved capabilities through to `SessionInfoBar`:

```swift
SessionInfoBar(
    session: richChat.currentSession,
    isWorking: richChat.isGenerating,
    acpInputTokens: richChat.acpInputTokens,
    acpOutputTokens: richChat.acpOutputTokens,
    acpThoughtTokens: richChat.acpThoughtTokens,
    acpCompressionCount: richChat.acpCompressionCount,
    projectName: chatViewModel.currentProjectName,
    gitBranch: chatViewModel.currentGitBranch,
    capabilities: capabilities?.capabilities ?? .empty
)
```

Pull the capabilities from the existing `@Environment(\.hermesCapabilities)` (declared on the parent view tree per [HermesCapabilities.swift:411](../../Packages/ScarfCore/Sources/ScarfCore/Services/HermesCapabilities.swift)). If the pane doesn't currently observe it, add `@Environment(\.hermesCapabilities) private var capabilities` at the top.

#### iOS

`Scarf iOS` doesn't have a `SessionInfoBar` mirror today; the iOS chat tab uses a different header. Skip iOS in this WS — capture under [Out of scope](#out-of-scope) for follow-up. Reasoning: iOS users are read-only consumers of compression count, the data model already flows through `RichChatViewModel`, and an iOS surface isn't gated on this WS.

### Coordination with WS-2

WS-2 mounts a "Goal locked" pill into `SessionInfoBar` between the project / branch chips and the working dot. The compression chip lives on the **right** half of the bar (next to tokens / cost), not the left, so the two changes don't collide spatially. They both add `var capabilities: HermesCapabilities = .empty` to `SessionInfoBar`, however — pick the same parameter name and order so whichever WS lands first establishes the prop and the second WS just reads it. WS-2 is presumed to land first (WS-2 is a flagship feature, this is polish); if not, both WSs need to add the prop and the merger should keep one declaration.

## 2. `/new <name>` slash command argument

### Current state

`/new` already appears in the slash menu — it's advertised by the ACP server via `available_commands_update` (handled in [RichChatViewModel:234](../../Packages/ScarfCore/Sources/ScarfCore/ViewModels/RichChatViewModel.swift) into `acpCommands`). The argumentHint comes from whatever the server sends. That means the v0.13 server will *automatically* surface a hint update because Hermes will send `"argument_hint": "[name]"` (or similar) once the new flag lands. We don't need to hardcode a Scarf-side override.

### What we change

The user-visible work here is mostly verification / smoke-testing. The mechanical changes are minor, mostly defensive:

#### [scarf/scarf/Features/Chat/Views/SlashCommandMenu.swift](../../scarf/Features/Chat/Views/SlashCommandMenu.swift)

The argument hint renderer at line 89–93 wraps the hint in `<…>` literally. Hermes v0.13 likely emits the optional argument as `[name]` (square-bracket convention for "optional"). If we leave the wrapper in place we'd render `<[name]>`. Replace the wrapper with a smarter join:

```swift
if let hint = command.argumentHint {
    let display = hint.hasPrefix("<") || hint.hasPrefix("[")
        ? hint
        : "<\(hint)>"
    Text(display)
        .font(ScarfFont.monoSmall)
        .foregroundStyle(ScarfColor.foregroundFaint)
}
```

This way the server's chosen brackets pass through, and existing entries that send `guidance` (without brackets) still render `<guidance>`.

#### Capability gate (none required, but a help-text override is allowed)

We *could* gate the rendering behind `hasNewWithSessionName` and override the hint only on v0.13+ — but the agent is the source of truth for the hint, and pre-v0.13 will send no hint at all (or the old hint). Leaving the renderer un-gated and trusting the agent's value is simpler and forward-compatible. **No flag check at this site.**

The flag exists for one place: a small banner in the slash menu that says "Tip: `/new <name>` lets you label the next session" on v0.13+ if the user hovers `/new` for >1s. **Defer the tip — over-engineering for one slash command.** Capture under [Out of scope](#out-of-scope).

### Coordination with WS-2

WS-2 also touches the slash menu (adds `/goal` and `/queue` to `nonInterruptiveCommands`), but only at the `RichChatViewModel.nonInterruptiveCommands` array site. This WS doesn't touch that array — only the renderer. Independent.

## 3. `hermes update --yes` plumbing

### Current state

There is **no in-app `hermes update` affordance** in v2.7.5. `UpdaterService` ([scarf/Core/Services/UpdaterService.swift](../../scarf/Core/Services/UpdaterService.swift)) wraps Sparkle for Scarf-self-update — that's a separate concern from updating the Hermes binary. The `hermes update` subcommand (added in v0.12 with `--check`, extended in v0.13 with `--yes`) is currently invoked by users in their terminal. The comment at [scarfApp.swift:281](../../scarf/scarfApp.swift) ("explicit refresh after `hermes update`") is aspirational — there's no UI that invokes `hermes update`.

### What this WS adds

A small forward-compatible utility so the future "Update Hermes" affordance (queued for a later release) doesn't have to re-derive flag selection. Add a single static helper on either `HermesUpdaterCommandBuilder` (new, in ScarfCore) or as a static on `UpdaterService` (Mac-only). Picking ScarfCore so iOS gets it for free, even though iOS won't ship the affordance soon either:

#### [scarf/Packages/ScarfCore/Sources/ScarfCore/Services/HermesUpdaterCommandBuilder.swift](../../Packages/ScarfCore/Sources/ScarfCore/Services/HermesUpdaterCommandBuilder.swift) (NEW)

```swift
import Foundation

/// Pure helpers that build argv arrays for `hermes update` invocations.
/// Lives here so the eventual UI surface (Mac / iOS / remote) shares
/// flag selection. Each helper is a `nonisolated static` pure function
/// — no transport, no MainActor, no mocking surface required.
public enum HermesUpdaterCommandBuilder {
    /// Argv for an interactive update. Pre-v0.12 hosts only had `update`;
    /// v0.12+ accepts `--check` for preflight; v0.13+ accepts `--yes` /
    /// `-y` for unattended runs.
    public static func updateArgv(
        capabilities: HermesCapabilities,
        unattended: Bool,
        checkOnly: Bool
    ) -> [String] {
        var args: [String] = ["update"]
        if checkOnly && capabilities.hasUpdateCheck {
            args.append("--check")
        }
        if unattended && capabilities.hasUpdateNonInteractive {
            args.append("--yes")
        }
        return args
    }
}
```

Test target: a small `M0eUpdaterTests` suite (new file under `ScarfCoreTests`) covering the matrix:

- pre-v0.12 → `["update"]` regardless of flags
- v0.12 + checkOnly → `["update", "--check"]`
- v0.12 + unattended → `["update"]` (flag absent — host can't honor it)
- v0.13 + unattended → `["update", "--yes"]`
- v0.13 + checkOnly + unattended → `["update", "--check", "--yes"]`

### What this WS does NOT add

No UI surface. No menu item, no Settings row, no command-palette entry. The plumbing exists so when v2.9 / v3.0 adds the affordance it doesn't need to derive flag logic from scratch. Per the WS-8 prompt: "If no such surface exists in v2.7.5, the v0.13 flag is forward-compat plumbing only — note that and don't over-build."

### Coordination with WS-2

None. Different files.

## 4. Redaction default-flip awareness

### Current state

The toggle lives in [scarf/Features/Settings/Views/Tabs/AdvancedTab.swift:129–133](../../scarf/Features/Settings/Views/Tabs/AdvancedTab.swift), inside the `Caching & Redaction` section. It's wired through `viewModel.config.redactionEnabled` ↔ `redaction.enabled`. The default for the *Scarf-side* `bool("redaction.enabled", default: false)` at [HermesFileService.swift:315](../../scarf/Core/Services/HermesFileService.swift) is `false` — meaning when the YAML key is absent, Scarf reads the toggle as off. That matches v0.12 server behavior.

In v0.13 the *server-side* default flips to ON (Hermes treats absence-of-key as redaction-enabled). Scarf's read default at the line above stays `false` because that's what we display when the user hasn't explicitly set the key — but the *meaning* of "off-with-no-key" diverges:

- pre-v0.13 host + no key → Scarf shows OFF, server treats as OFF. Honest.
- v0.13 host + no key → Scarf shows OFF, server treats as ON. **Confusing.**

### What we change — option A (recommended): hint copy only

Smallest possible surface. Don't change the parsing default; the file ground-truth is "key absent". Add a one-line hint below the toggle whose copy depends on `capabilities.hasContextCompressionCount` (any v0.13 flag works as a discriminant; reuse one rather than adding `hasV013` to `HermesCapabilities`). Pick `hasGoals` as the marker since it's the most central v0.13 flag — but that's an aesthetic choice; any of the v0.13 flags discriminate the same set of hosts.

#### [scarf/scarf/Features/Settings/Views/Tabs/AdvancedTab.swift](../../scarf/Features/Settings/Views/Tabs/AdvancedTab.swift)

Inside `v012CachingSection`'s `SettingsSection` (around line 122–139), after the `ToggleRow` for `redaction.enabled`, append a `HintRow` (or whatever the existing inline-hint pattern in that file is — likely just a `Text` wrapped in a styled `HStack` matching the `credentialsHint` shape from `GeneralTab`):

```swift
ToggleRow(
    label: "Redact secrets in patches",
    isOn: viewModel.config.redactionEnabled
) { viewModel.setSetting("redaction.enabled", value: $0 ? "true" : "false") }

redactionDefaultsHint
```

…and add the computed view:

```swift
@Environment(\.hermesCapabilities) private var capabilitiesStore

@ViewBuilder
private var redactionDefaultsHint: some View {
    let v013 = capabilitiesStore?.capabilities.hasGoals == true
    HStack {
        Text("")
            .frame(width: 160, alignment: .trailing)
        Text(v013
             ? "Recommended: ON. Hermes v0.13+ defaults to redacting secrets unless you opt out."
             : "Default OFF in Hermes v0.12. Toggle ON to redact secrets in logs and shares.")
            .scarfStyle(.caption)
            .foregroundStyle(ScarfColor.foregroundFaint)
        Spacer()
    }
    .padding(.horizontal, 12)
    .padding(.vertical, 4)
}
```

The aligned-right empty `Text` mimics the label-column gutter so the hint tucks under the toggle's value column rather than aligning with the section's left edge — matches the existing visual rhythm in this tab.

### Why option A and not option B (changing the parsing default)

Option B would be: read `bool("redaction.enabled", default: capabilities.hasGoals)`. That sounds nicer but wires capabilities into `HermesFileService.parseConfig`, which is currently `nonisolated` and pure. Threading the store through would touch a dozen call sites. Not worth it for a hint that's already accurate via option A.

### Coordination with WS-2

None. Different file, different section.

## 5. `display.language` picker

### What v0.13 adds

Hermes v0.13 honors `display.language` in `config.yaml` for static-message translations. Supported values: `en` (default), `zh`, `ja`, `de`, `es`, `fr`, `uk`, `tr`. Users can already write the YAML by hand; this WS adds an in-app picker so it's discoverable.

### Files to change

#### [scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift](../../Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift)

`DisplaySettings` (around line 30) gains one field:

```swift
public struct DisplaySettings: Sendable, Equatable {
    public var skin: String
    public var compact: Bool
    public var resumeDisplay: String
    public var bellOnComplete: Bool
    public var inlineDiffs: Bool
    public var toolProgressCommand: Bool
    public var toolPreviewLength: Int
    public var busyInputMode: String
    /// Static-message translation language. v0.13+. Empty string means
    /// "follow Hermes default" — we display this as `en` in the picker.
    /// Persisted via `hermes config set display.language <code>`.
    public var language: String
    …
}
```

Add to the initializer (with a default empty-string value and a fall-through assignment) and to the `.empty` static. **Don't** default to `"en"` here — empty string means "config key absent", which is semantically distinct from "user explicitly chose en". The picker collapses both to "English" in display, but the writer only writes a value when the user picks something.

#### [scarf/scarf/Core/Services/HermesFileService.swift](../../scarf/Core/Services/HermesFileService.swift)

Inside the `display` block construction (around line 79–84), add:

```swift
let display = DisplaySettings(
    skin: str("display.skin", default: "default"),
    compact: bool("display.compact", default: false),
    resumeDisplay: str("display.resume_display", default: "full"),
    bellOnComplete: bool("display.bell_on_complete", default: false),
    inlineDiffs: bool("display.inline_diffs", default: true),
    toolProgressCommand: bool("display.tool_progress_command", default: false),
    toolPreviewLength: int("display.tool_preview_length", default: 0),
    busyInputMode: str("display.busy_input_mode", default: "interrupt"),
    language: str("display.language", default: "")
)
```

#### [scarf/scarf/Features/Settings/ViewModels/SettingsViewModel.swift](../../scarf/Features/Settings/ViewModels/SettingsViewModel.swift)

Add a setter alongside the existing `setSkin` (line 99):

```swift
func setDisplayLanguage(_ value: String) {
    setSetting("display.language", value: value)
}
```

And expose the option list (8 entries; mirror the v0.13 release notes):

```swift
var displayLanguages: [(code: String, label: String)] = [
    ("",   "English (default)"),
    ("en", "English"),
    ("zh", "中文 (Chinese)"),
    ("ja", "日本語 (Japanese)"),
    ("de", "Deutsch (German)"),
    ("es", "Español (Spanish)"),
    ("fr", "Français (French)"),
    ("uk", "Українська (Ukrainian)"),
    ("tr", "Türkçe (Turkish)"),
]
```

Two "English" entries (empty string + `en`) is intentional: the empty string means "no key" — picking `en` writes the key explicitly. UX-wise that's fine — the picker shows "English (default)" while the value-stored is still empty, and switching to `en` writes a key. Most users will move between languages, not toggle the key's presence.

#### [scarf/scarf/Features/Settings/Views/Tabs/GeneralTab.swift](../../scarf/Features/Settings/Views/Tabs/GeneralTab.swift)

Inside the existing `Locale` section (line 40–42), add a picker row gated on `hasDisplayLanguage`:

```swift
SettingsSection(title: "Locale", icon: "globe.americas") {
    EditableTextField(label: "Timezone (IANA)", value: viewModel.config.timezone) {
        viewModel.setTimezone($0)
    }
    if capabilitiesStore?.capabilities.hasDisplayLanguage == true {
        PickerRow(
            label: "Display language",
            selection: viewModel.config.display.language.isEmpty
                ? "" : viewModel.config.display.language,
            options: viewModel.displayLanguages.map(\.code),
            optionLabel: { code in
                viewModel.displayLanguages.first { $0.code == code }?.label ?? code
            }
        ) { viewModel.setDisplayLanguage($0) }
    }
}
```

Add `@Environment(\.hermesCapabilities) private var capabilitiesStore` at the top of `GeneralTab`.

The `PickerRow` overload that takes a `optionLabel:` mapper may not exist today — check at implementation time, and if it doesn't, either (a) add the overload to `PickerRow.swift` (a simple closure parameter), or (b) inline a SwiftUI `Picker` directly rather than `PickerRow` for this one row. Option (a) is preferred so the rest of Settings can use it.

#### iOS

`Scarf iOS` settings are read-mostly (config writes are deferred to the Mac per the existing pattern). Skip iOS for the picker; iOS just shows the value as-is wherever Settings displays it. No iOS work in this WS.

### Capability gate

`hasDisplayLanguage` is checked at the picker site. Pre-v0.13 hosts hide the row entirely — the field would be silently ignored by the agent if written. **Don't** half-render with a "requires v0.13" label; the row should be invisible on older hosts so the user doesn't think the surface is broken.

### Coordination with WS-2

None. Different file.

## 6. xAI Custom Voices badge

### Current state

The xAI provider is **not in `ttsProviders` today** (verify at [SettingsViewModel.swift:32](../../scarf/Features/Settings/ViewModels/SettingsViewModel.swift) — the array reads `["edge", "elevenlabs", "openai", "minimax", "mistral", "neutts", "piper"]`, no `xai`). Hermes v0.13 adds xAI as a TTS provider (it was added earlier in fact, v0.12 — the v0.13 surface is just the *Custom Voices* / cloning support on top). This WS does both at once: add `xai` to the picker and surface the cloning-supported badge.

### Files to change

#### [scarf/scarf/Features/Settings/ViewModels/SettingsViewModel.swift](../../scarf/Features/Settings/ViewModels/SettingsViewModel.swift)

Extend the provider list:

```swift
var ttsProviders = ["edge", "elevenlabs", "openai", "minimax", "mistral", "neutts", "piper", "xai"]
```

Add setter(s) for whichever xAI-specific config keys Hermes uses. Per [Open question Q2](#open-questions) the exact keys — likely `tts.xai.voice_id` (or similar) and possibly `tts.xai.model` — need confirmation. Conservative shape mirroring elevenlabs:

```swift
func setTTSXAIVoiceID(_ value: String) { setSetting("tts.xai.voice_id", value: value) }
func setTTSXAIModel(_ value: String)   { setSetting("tts.xai.model", value: value) }
```

#### [scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift](../../Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift)

`VoiceSettings` (around line 178) gains two fields next to the existing TTS provider blobs:

```swift
public var ttsXAIVoiceID: String
public var ttsXAIModel: String
```

Initializer + `.empty` updates. Defaults to empty string.

#### [scarf/scarf/Core/Services/HermesFileService.swift](../../scarf/Core/Services/HermesFileService.swift)

Add the YAML reads inside the voice block construction (mirror the elevenlabs / openai shape).

#### [scarf/scarf/Features/Settings/Views/Tabs/VoiceTab.swift](../../scarf/Features/Settings/Views/Tabs/VoiceTab.swift)

Inside the `switch viewModel.config.voice.ttsProvider` (line 19), add a `case "xai":` arm:

```swift
case "xai":
    EditableTextField(label: "Voice ID", value: viewModel.config.voice.ttsXAIVoiceID) {
        viewModel.setTTSXAIVoiceID($0)
    }
    EditableTextField(label: "Model", value: viewModel.config.voice.ttsXAIModel) {
        viewModel.setTTSXAIModel($0)
    }
    if capabilitiesStore?.capabilities.hasXAIVoiceCloning == true {
        xaiCloningBadge
    }
```

Add `@Environment(\.hermesCapabilities) private var capabilitiesStore` at the top.

The badge view, using `ScarfBadge` (kind `.info`):

```swift
@ViewBuilder
private var xaiCloningBadge: some View {
    HStack {
        Text("")
            .frame(width: 160, alignment: .trailing)
        ScarfBadge("Cloning supported", kind: .info)
        Text("Manage cloned voices in your terminal: `hermes voice` (xAI subcommands).")
            .scarfStyle(.caption)
            .foregroundStyle(ScarfColor.foregroundMuted)
        Spacer()
    }
    .padding(.horizontal, 12)
    .padding(.vertical, 4)
}
```

The hint text references `hermes voice` because Scarf doesn't manage cloned voices — Hermes does, and v2.7.5 has no in-app voice-cloning UI. Capture under [Out of scope](#out-of-scope) for follow-up.

### Capability gate

- `xai` in the provider picker: **not gated**. The provider exists pre-v0.13 (TTS only); cloning is the v0.13 add-on. Listing it always means pre-v0.13 users with xAI keys can still pick it.
- Cloning badge: gated on `hasXAIVoiceCloning`. Pre-v0.13: badge hidden, EditableTextField rows still visible.

### Coordination with WS-2

None.

## Files to change (combined)

New:

- `scarf/Packages/ScarfCore/Sources/ScarfCore/Services/HermesUpdaterCommandBuilder.swift` (item 3)
- `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/M0eUpdaterTests.swift` (item 3 tests)

Modified:

- `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/ACPMessages.swift` (item 1: `compressionCount` field)
- `scarf/Packages/ScarfCore/Sources/ScarfCore/ACP/ACPClient.swift` (item 1: decode)
- `scarf/Packages/ScarfCore/Sources/ScarfCore/ViewModels/RichChatViewModel.swift` (item 1: counter + reset + `handlePromptComplete`)
- `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift` (items 5 + 6: `display.language`, xAI voice/model fields)
- `scarf/scarf/Features/Chat/Views/SessionInfoBar.swift` (item 1: chip + props)
- `scarf/scarf/Features/Chat/Views/ChatTranscriptPane.swift` (item 1: pass-through)
- `scarf/scarf/Features/Chat/Views/SlashCommandMenu.swift` (item 2: bracket-aware hint)
- `scarf/scarf/Features/Settings/Views/Tabs/AdvancedTab.swift` (item 4: redaction hint)
- `scarf/scarf/Features/Settings/Views/Tabs/GeneralTab.swift` (item 5: language picker)
- `scarf/scarf/Features/Settings/Views/Tabs/VoiceTab.swift` (item 6: xai case + badge)
- `scarf/scarf/Features/Settings/ViewModels/SettingsViewModel.swift` (items 5 + 6: setters + lists)
- `scarf/scarf/Core/Services/HermesFileService.swift` (items 5 + 6: YAML reads)
- (possibly) `scarf/scarf/Features/Settings/Views/Components/PickerRow.swift` — add a `optionLabel:` overload (item 5, if the existing API doesn't carry one)

That's roughly **3 ScarfCore files + 7 Mac app files + 1 new file + 1 test file = ~12 files**, most edits being a few lines each.

## Capability gating (combined)

| Item | Flag | Behavior on pre-v0.13 |
|------|------|------------------------|
| 1. Compression chip | `hasContextCompressionCount` + `acpCompressionCount > 0` | Chip hidden (counter stays 0) |
| 2. `/new <name>` hint | none — driven by ACP server payload | Hint is whatever pre-v0.13 server sends (probably empty) |
| 3. `--yes` plumbing | `hasUpdateNonInteractive` (used inside the helper) | Helper omits the flag |
| 4. Redaction hint copy | discriminator on any v0.13 flag (use `hasGoals`) | Shows the v0.12 copy |
| 5. Language picker | `hasDisplayLanguage` | Picker row hidden |
| 6. xAI cloning badge | `hasXAIVoiceCloning` | Badge hidden, xai picker option still visible |

Six surfaces, six independent fall-back paths. None of them break the existing layout if every flag returns false.

## How to test

### Unit (ScarfCoreTests)

- `M0eUpdaterTests` — five-case matrix for `HermesUpdaterCommandBuilder.updateArgv` covering every combination listed in item 3.
- Extend `M0dViewModelsTests` with one test that sets `acpCompressionCount = 5` via a mocked `handlePromptComplete` and asserts the value via the public getter; assert `reset()` clears it.
- Extend the existing `ACPMessages` tests (or add one if there isn't one) with: a `usage` blob carrying `"compressionCount": 3` parses into `ACPPromptResult.compressionCount == 3`; same with `"compression_count": 3`; missing key parses as 0.

### UI smoke (manual against real Hermes)

1. **Pre-v0.13 host**: launch Scarf with a Hermes v0.12 binary on PATH. Verify:
   - No compression chip in `SessionInfoBar` even after long sessions.
   - Settings → General → Locale shows only the Timezone field; no language picker.
   - Settings → Advanced → Caching & Redaction shows the v0.12 hint copy.
   - Settings → Voice → Text-to-Speech with provider `xai` shows Voice ID + Model fields, **no** "Cloning supported" badge.

2. **v0.13 host**: launch Scarf against the v0.13 dev branch. Verify:
   - Long enough chat to trigger compaction → chip appears in `SessionInfoBar` with the count.
   - Settings → General → Locale → "Display language" picker visible, switching writes `display.language` in `config.yaml`.
   - Settings → Advanced shows the v0.13 hint copy.
   - Settings → Voice → xai provider shows the "Cloning supported" badge.
   - `/new Foo Bar Baz` from the slash menu starts a session named "Foo Bar Baz" (no Scarf-side validation; Hermes handles it).
   - Slash menu shows `/new` with whatever hint v0.13 server sends — bracket-aware renderer doesn't double-wrap if hint is `[name]`.

3. **`HermesUpdaterCommandBuilder` smoke** (no UI): once integrated, write a one-shot script (or a `#Preview`-only call) that prints `updateArgv` for each capability snapshot and pastes the matrix into the PR description.

### Visual / accessibility

- Compression chip uses `ScarfColor.foregroundMuted` — verify in light + dark; ensure contrast ratio ≥ 4.5:1 against `backgroundSecondary`.
- Picker on Locale section honors keyboard navigation (Tab in / Space to open / Arrows / Return / Esc).
- "Cloning supported" badge uses `ScarfBadge(... kind: .info)` — verify color resolves correctly in both modes; not green (that's `.success`), not yellow (that's `.warning`).

## Open questions

**Q1. Wire field name for compression count.** v0.13 release notes say "Show context compression count in status bar" without naming the field. The plan assumes `usage.compressionCount` (or `usage.compression_count`) on the `session/prompt` response. If Hermes instead emits it as a `session/update` notification on a status feed (separate path from `usage`), the plumbing is bigger: `RichChatViewModel.handleStatusUpdate` (or equivalent) needs a new branch, and `ACPClient.startReadLoop` needs a new event type. **Resolution path**: read `~/.hermes/hermes-agent/hermes_cli/acp/server.py` (or wherever the v0.13 status emission lives) before merging. If the field is on a notification, swap items 1's `ACPPromptResult` extension for a new `ACPEvent.compressionCountChanged(sessionId:count:)` case in `ACPMessages.swift` and a corresponding branch in `RichChatViewModel.handleEvent`.

**Q2. xAI TTS config keys.** The plan assumes `tts.xai.voice_id` / `tts.xai.model` mirroring elevenlabs. v0.13 source might use different names (`tts.xai.voice`, `tts.xai.model_id`, or a top-level `tts.xai_voice`). **Resolution path**: grep `~/.hermes/hermes-agent/hermes_cli/voice/tts.py` for the xAI config block before merging. If keys differ, just rename the setter functions and `VoiceSettings` fields — no architectural change.

**Q3. Empty-string vs `"en"` for `display.language` default.** The plan uses an empty string in `DisplaySettings.language` to represent "key absent" and surfaces the picker entry as "English (default)". Whether the picker should *also* offer `en` as a separate explicit value is a UX call. The plan keeps both for now; v2.8.1 can collapse if it's confusing.

**Q4. iOS coverage.** The plan defers iOS for items 1 (compression chip) and 5 (language picker) — iOS doesn't have a `SessionInfoBar` mirror, and iOS Settings is read-mostly. For v2.8 this is acceptable; for v2.9 we should mirror both surfaces in `Scarf iOS/`. Tracking under [Out of scope](#out-of-scope) below.

**Q5. Redaction hint discriminator.** Using `hasGoals` as a stand-in for "is this a v0.13 host" feels indirect. Consider adding a small convenience `var isV013OrLater: Bool { atLeastSemver(0, 13, 0) }` on `HermesCapabilities` so the call site reads more honestly. Trivial change; either lands in WS-1 (preferred — that's the capabilities home) or here. Flag for WS-1 owner.

## Out of scope (deferred)

- **iOS compression chip** — iOS chat header doesn't currently render any token counter; adding the chip there means designing a header bar, not just inserting one element. Track for v2.9.
- **iOS `display.language` picker** — iOS Settings is read-mostly; full pickers wait until iOS Settings becomes a write surface.
- **In-app "Update Hermes" affordance** — a Sparkle-style auto-updater for the Hermes binary, with the `--yes` flag plumbed through. Long-term feature, probably v3.0. The helper added in item 3 paves the runway.
- **`/new <name>` hover tooltip** — extra discoverability for the optional argument. v0.13 server sends the hint via `available_commands_update`; that's enough for v2.8.
- **xAI Custom Voices management UI** — the badge points users at `hermes voice`. Building cloned-voice management in-app is a feature on its own. Track separately.
- **Schema sync to `tools/build-catalog.py`** — none of this WS adds new widget types or template manifest fields, so the catalog validator doesn't need an update. Verify at PR time.

## Estimate

- ScarfCore changes: ~30 LOC across 3 files + 1 new file + 1 test file ≈ **~120 LOC**.
- Mac app changes: ~15-20 LOC per item 1, 4, 5, 6 + 5 LOC for items 2 = **~80 LOC** spread over 7 files.
- Tests: ~80 LOC for `M0eUpdaterTests` + ~40 LOC for compression decode tests = **~120 LOC**.

Total ≈ **300-350 LOC**, ~12 files. Each item is independently revertable and capability-gated. Implementation: 1 dev-day; review + smoke against v0.13 host: 0.5 day. **1.5 dev-days end-to-end.**

Confidence: **high** that items 2 / 3 / 4 / 5 / 6 land cleanly. **Medium** for item 1 (compression chip) — pinned to Q1's wire-field resolution. If Q1 surfaces an event-stream shape rather than a `usage` blob, item 1's plumbing roughly doubles in size but the *view* is unchanged.
