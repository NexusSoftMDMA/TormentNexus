# WS-5 Plan: Messaging Gateway v0.13 expansion

> **Scope.** Catch the Mac **Messaging Gateway** + **Platforms** surfaces up to
> Hermes v0.13.0. WS-1 already landed the capability flags
> (`hasGoogleChatPlatform`, `hasGatewayAllowlists`, `hasGatewayBusyAckToggle`,
> `hasGatewayRestartNotification`, `hasGatewayList`). This work-stream consumes
> them: 20th platform (Google Chat), per-platform allowlists, three new
> per-platform behavior toggles, a cross-profile status digest in the gateway
> header, and a passing nod to `[[as_document]]` in skill detail. iOS read-only
> mirror is **deferred to WS-9** by stream contract.
>
> **Terminology guard.** Scarf has TWO things called "Gateway" and the
> distinction is load-bearing for users:
>
> 1. **Messaging Gateway** — outbound bridge to chat platforms. This work-stream.
>    Files under `scarf/Features/Gateway/` and `scarf/Features/Platforms/`.
> 2. **Tool Gateway** — Nous Portal subscription routing for web search / image /
>    TTS / browser. v0.10 surface, lives in `scarf/Features/Health/` and
>    `ModelCatalogService`. **DO NOT TOUCH.**
>
> Every label, header, and `// MARK:` introduced in this work-stream that
> contains the word "Gateway" must be prefixed "Messaging" unless the
> surrounding context already rules out the Tool Gateway interpretation
> (e.g. a label nested under `Features/Platforms/` is unambiguously about the
> messaging side).

## Goals

1. **Google Chat** appears as the 20th platform under
   `Settings → Platforms` (Mac), capability-gated on
   `hasGoogleChatPlatform`. Setup is informational ("Run `hermes setup`")
   because the OAuth dance is interactive and lives outside Scarf — same
   shape as the existing `yuanbao` / `microsoft-teams` panels.
2. **Per-platform allowlist editor** for the six platforms Hermes added
   `allowed_channels` / `allowed_chats` / `allowed_rooms` to in v0.13:
   Slack, Mattermost, Google Chat (channels); Telegram, WhatsApp (chats);
   Matrix, DingTalk (rooms). YAML-driven; persists to
   `~/.hermes/config.yaml` under `gateway.platforms.<platform>.allowed_<kind>`.
3. **Per-platform "Gateway behavior" subsection** on each platform's
   setup card with three toggles:
   - `gateway_restart_notification` (bool, default OFF) —
     "Post 'Gateway restarted' notice on boot"
   - `busy_ack_enabled` (bool, default ON) —
     "Send 'Agent is working…' ack" — toggle off to suppress
   - `slash_command_notice_ttl_seconds` (int, default 0=disabled) —
     "Auto-delete slash-command notices after N seconds"
   Each gated separately on the matching capability flag.
4. **Cross-profile status digest** in the `MessagingGatewayView` header —
   one-line summary sourced from `hermes gateway list --json`. Hidden on
   pre-v0.13 hosts and hidden when the verb fails or returns empty.
5. **`[[as_document]]` directive** surfaced as a tooltip on the relevant
   skill detail rows. Informational only.
6. **Tests** — parser tests for `hermes gateway list --json` output and
   round-trip tests for the allowlist YAML editor.

## Non-goals

- iOS read-only mirror of the gateway / platforms surface (WS-9).
- Editing the `allowed_*` lists on remote SSH targets (out of WS-5 because
  remote `config.yaml` writes already round-trip through `writeText` — the
  abstraction should Just Work, but explicit verification is WS-12 / Remote
  Hardening).
- Dingtalk platform card. Dingtalk has `allowed_rooms` but no
  `KnownPlatforms` entry exists yet, and no `<Dingtalk>SetupViewModel.swift`.
  Adding a Dingtalk panel is out of WS-5 scope; the allowlist editor handles
  Dingtalk only if/when a future PR adds the platform card.
- Migrating `Settings → Platforms` to ScarfDesign tokens (separate,
  cross-cutting cleanup; WS-5 stays consistent with neighbors).

## Files to change

### ScarfCore

#### NEW — `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/GatewayPlatformSettings.swift`

A small Sendable model bundling the v0.13 per-platform behavior block. Mirrors
the AuxiliaryModel / DiscordSettings / TelegramSettings pattern.

```
public struct GatewayPlatformSettings: Sendable, Equatable {
    public var allowedChannels: [String]   // Slack / Mattermost / Google Chat
    public var allowedChats: [String]      // Telegram / WhatsApp
    public var allowedRooms: [String]      // Matrix / DingTalk
    public var busyAckEnabled: Bool        // default true
    public var gatewayRestartNotification: Bool   // default false
    public var slashCommandNoticeTTLSeconds: Int  // default 0 = disabled
    // ...empty()/initializer
}
```

The exact `allowedChannels` / `allowedChats` / `allowedRooms` field is
populated based on the platform — for Slack only `allowedChannels` is
meaningful, for Matrix only `allowedRooms`, etc. The model carries all three
so a single struct fits every platform; the editor UX surfaces the right one
based on `GatewayAllowlistKind` (below).

#### NEW — `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/GatewayAllowlistKind.swift`

```
public enum GatewayAllowlistKind: String, Sendable, Equatable {
    case channels   // -> allowed_channels
    case chats      // -> allowed_chats
    case rooms      // -> allowed_rooms

    public var yamlKey: String {
        switch self {
        case .channels: return "allowed_channels"
        case .chats:    return "allowed_chats"
        case .rooms:    return "allowed_rooms"
        }
    }

    public var inputPlaceholder: String {
        switch self {
        case .channels: return "C0123ABCD or #channel-name"
        case .chats:    return "@username or 12345678"
        case .rooms:    return "!RoomId:matrix.org"
        }
    }

    public var noun: String {
        switch self {
        case .channels: return "channel"
        case .chats:    return "chat"
        case .rooms:    return "room"
        }
    }

    /// Map a platform name to the allowlist kind it supports. Returns nil
    /// for platforms without v0.13 allowlist support (cli, signal, email,
    /// imessage, homeassistant, webhook, yuanbao, microsoft-teams).
    public static func kind(for platform: String) -> GatewayAllowlistKind? {
        switch platform {
        case "slack", "mattermost", "google-chat", "googlechat": return .channels
        case "telegram", "whatsapp":                              return .chats
        case "matrix", "dingtalk":                                return .rooms
        default: return nil
        }
    }
}
```

The `googlechat`/`google-chat` dual-spelling guards against Hermes' final
platform identifier landing as either form — verify against Hermes once
v0.13 GA ships. **Open question (Q1).**

#### EDIT — `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesTool.swift`

Add the 20th platform. Mirror the v0.12 yuanbao/teams comment block.

```
HermesToolPlatform(
    name: "google-chat",       // verify against Hermes — see Q1
    displayName: "Google Chat",
    icon: "bubble.left.fill"   // candidate; pick a glyph that doesn't
                               // collide with Discord or Yuanbao
),
```

Update `KnownPlatforms.icon(for:)` switch likewise.

The 20th-platform comment in this file's `KnownPlatforms.all` literal
extends the existing v0.12 marker:

```
// -- v0.13 additions ---------------------------------------------
// Google Chat is the 20th gateway platform. It's a generic
// `env_enablement_fn` / `cron_deliver_env_var`-driven adapter; setup
// runs through `hermes setup` rather than per-field forms because
// the auth dance is OAuth-style and lives outside Scarf.
```

#### EDIT — `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift`

Extend the existing `HermesConfig` to hold a per-platform settings dictionary
keyed by platform name. Two ways to do this; the planner recommends **option
B**:

- **Option A** — bolt seven optional fields onto `HermesConfig` directly
  (`slackAllowedChannels`, `telegramAllowedChats`, ...). Big surface area;
  `HermesConfig` already has 50-odd fields and this would scatter the v0.13
  block across the type. Rejected.
- **Option B** — add a single `gatewayPlatforms: [String: GatewayPlatformSettings]`
  field. The YAML loader populates entries on demand. Editor reads/writes
  through a single lookup key. Recommended.

```
public struct HermesConfig: Sendable, Equatable {
    // ...existing fields...
    public let gatewayPlatforms: [String: GatewayPlatformSettings]
    // ...
}
```

`HermesConfig.empty` initializes `gatewayPlatforms: [:]`. `HermesConfig`
already has a 14-arg init; adding one more positional arg is fine here
because every callsite uses the labeled initializer.

#### EDIT — `scarf/Packages/ScarfCore/Sources/ScarfCore/Parsing/HermesConfig+YAML.swift`

Inside the existing `init(yaml:)`, after the `slack` / `matrix` / `mattermost`
blocks, add a per-platform settings extractor:

```swift
let gatewayAllowlistPlatforms = [
    "slack", "mattermost", "google-chat",
    "telegram", "whatsapp",
    "matrix", "dingtalk",
]
var gatewayPlatforms: [String: GatewayPlatformSettings] = [:]
for platform in gatewayAllowlistPlatforms {
    let prefix = "gateway.platforms.\(platform)."
    let allowedChannels = lists[prefix + "allowed_channels"] ?? []
    let allowedChats    = lists[prefix + "allowed_chats"]    ?? []
    let allowedRooms    = lists[prefix + "allowed_rooms"]    ?? []
    let busy            = bool(prefix + "busy_ack_enabled", default: true)
    let restartNotice   = bool(prefix + "gateway_restart_notification",
                               default: false)
    let ttl             = int(prefix + "slash_command_notice_ttl_seconds",
                              default: 0)
    // Skip platforms with no v0.13 fields present in the file at all.
    let isEmpty = allowedChannels.isEmpty
        && allowedChats.isEmpty
        && allowedRooms.isEmpty
        && values[prefix + "busy_ack_enabled"] == nil
        && values[prefix + "gateway_restart_notification"] == nil
        && values[prefix + "slash_command_notice_ttl_seconds"] == nil
    if !isEmpty {
        gatewayPlatforms[platform] = GatewayPlatformSettings(
            allowedChannels: allowedChannels,
            allowedChats: allowedChats,
            allowedRooms: allowedRooms,
            busyAckEnabled: busy,
            gatewayRestartNotification: restartNotice,
            slashCommandNoticeTTLSeconds: ttl
        )
    }
}
```

The `gateway.platforms.<platform>.allowed_<kind>` YAML path is the
**unverified default**. **Open question (Q2)** — depending on Hermes'
config.yaml layout in v0.13, the actual paths might be
`platforms.<platform>.allowed_<kind>` (sibling to the existing Slack
`platforms.slack.*` namespace) instead of nested under a new `gateway:` key.
Resolution before implementation: read the v0.13 docs / sample config or
ask Hermes maintainers; pick whichever shape Hermes actually emits and
adjust both the parser and the editor in lockstep. The plan below uses
`gateway.platforms.<platform>.*` as the placeholder.

Pass the dict through to the `HermesConfig.init`.

#### NEW — `scarf/Packages/ScarfCore/Sources/ScarfCore/Services/GatewayConfigWriter.swift`

A small helper for writing list-valued YAML keys, since `hermes config set`
stringifies arrays (the same gotcha that forced the Home Assistant editor to
keep its watch lists read-only). Strategy:

- For **scalar** fields (`busy_ack_enabled`, `gateway_restart_notification`,
  `slash_command_notice_ttl_seconds`): use `hermes config set` via
  `PlatformSetupHelpers.runHermesCLI`. Same surface as every other platform
  toggle.
- For **list** fields (`allowed_channels` / `allowed_chats` / `allowed_rooms`):
  load → mutate → save the YAML directly via `ServerContext.writeText`.

The list-write path is a pure function so it can be unit-tested without a
filesystem:

```
public enum GatewayConfigWriter {
    /// Insert or replace `gateway.platforms.<platform>.<key>:` block in the
    /// YAML, preserving everything else byte-for-byte except the targeted
    /// block. Returns the new YAML.
    public static func setList(
        in yaml: String,
        platform: String,
        key: String,         // "allowed_channels" / "allowed_chats" / "allowed_rooms"
        items: [String]
    ) -> String { ... }

    /// Async I/O wrapper that reads, mutates, writes via the given context.
    /// Returns false on read or write failure.
    public static func saveList(
        context: ServerContext,
        platform: String,
        key: String,
        items: [String]
    ) -> Bool { ... }
}
```

Implementation strategy for `setList`: split the YAML into lines, find the
existing block by exact prefix match (`gateway.platforms.<platform>.<key>:`
at any indent), drop the entire bullet-list region until indent regresses,
splice the new block in. If no block exists, append it under a
`gateway:\n  platforms:\n    <platform>:\n` scaffold (creating any missing
ancestors). YAML-quote items that contain special characters (the Slack
channel ID space is alphanumeric so almost never; Telegram chat IDs are
numeric; safest is to single-quote anything containing `:` `#` `@` or
leading/trailing whitespace).

This is the **highest-risk** part of WS-5. The pure round-trip property
("setList(getList(y)) == y modulo whitespace"), plus the behaviour-parity
properties below in the test plan, are what tip the implementation toward
correct vs subtly broken.

#### NEW — `scarf/Packages/ScarfCore/Sources/ScarfCore/Services/HermesGatewayListService.swift`

Thin wrapper around `hermes gateway list --json`. Mirrors the
`HermesCapabilities` style — pure model + nonisolated detection helper,
returning `nil` when the verb fails (pre-v0.13 hosts will exit non-zero).

```
public struct GatewayListSnapshot: Sendable, Equatable {
    public struct ProfileEntry: Sendable, Equatable {
        public let profile: String
        public let isRunning: Bool
        public let pid: Int?
        public let platforms: [String]   // platform names connected/configured
    }
    public let profiles: [ProfileEntry]
    public let detectedAt: Date

    public var headerDigest: String {
        // "3 profiles (2 running) · default: slack, telegram"
        // or "default profile only · slack, telegram, discord"
        ...
    }
}

public enum HermesGatewayListService {
    /// Parse a JSON blob from `hermes gateway list --json` into a snapshot.
    /// Tolerant of unknown keys; returns nil for unparseable / empty input.
    public static func parse(_ json: Data) -> GatewayListSnapshot?

    /// Synchronous fetch helper — call from a `Task.detached`.
    public static func fetch(context: ServerContext) -> GatewayListSnapshot?
}
```

JSON shape is **provisional** — verify against Hermes once v0.13 GA ships.
**Open question (Q3).**

### Mac app — Messaging Gateway feature

#### EDIT — `scarf/Features/Gateway/ViewModels/GatewayViewModel.swift`

Rename the type to `MessagingGatewayViewModel` and the mark-1 `GatewayInfo`
struct to `MessagingGatewayInfo` to match the user-facing copy. **Local
rename only — leave the gateway sidebar enum case (`SidebarSection.gateway`)
untouched** because that string isn't user-facing and renaming it would
churn unrelated callers. (CLAUDE.md spells this out: "The `SidebarSection.gateway`
enum case and `gateway_state.json` / `gateway.log` paths are unchanged
(not user-facing strings).")

Add a `gatewayList: GatewayListSnapshot?` property with a fetch helper that
calls `HermesGatewayListService.fetch` from the existing detached load.
Skip the call when `capabilities.hasGatewayList == false`.

```swift
@Observable
final class MessagingGatewayViewModel {
    let context: ServerContext
    let capabilities: HermesCapabilities  // injected at init time
    // ...existing properties...
    var gatewayList: GatewayListSnapshot?
}
```

The `capabilities` is plumbed in from `MessagingGatewayView.init` via the
environment store.

#### EDIT — `scarf/Features/Gateway/Views/GatewayView.swift`

- Rename type to `MessagingGatewayView` (file name + struct).
- Read `@Environment(\.hermesCapabilities)` at view-init and pass to the VM
  constructor.
- Add a `crossProfileDigest` row above the existing `serviceSection` when
  `capabilities.hasGatewayList && viewModel.gatewayList != nil`. Render
  `viewModel.gatewayList!.headerDigest` with a `dot.radiowaves.left.and.right`
  glyph; clicking opens a popover with the full per-profile breakdown.
- Wrap the existing "Gateway start/stop/restart" buttons in a small toolbar
  on the right of the page header, **using `ScarfPrimaryButton` /
  `ScarfSecondaryButton`** (today they're plain `Button`). Apply
  `ScarfSpace`/`ScarfRadius` tokens — the `cornerRadius: 8` literal in this
  file is a code smell flagged by CLAUDE.md.

The page-header subtitle stays "Outbound channel bridge — Discord, Telegram,
Slack, etc." but adopt the format `"…, Slack, Google Chat, etc."` once
Google Chat is shipped (one-line update).

### Mac app — Platforms feature

#### EDIT — `scarf/Features/Platforms/Views/PlatformsView.swift`

Add a `googleChatPanel` similar to the existing `yuanbaoPanel` /
`microsoftTeamsPanel`:

```swift
private var googleChatPanel: some View {
    SettingsSection(title: "Google Chat",
                    icon: KnownPlatforms.icon(for: "google-chat")) {
        ReadOnlyRow(label: "Type",
                    value: "Generic env-driven gateway adapter (v0.13+)")
        ReadOnlyRow(label: "Setup",
                    value: "Run `hermes setup` and select Google Chat to walk the OAuth flow.")
        ReadOnlyRow(label: "Configured",
                    value: viewModel.hasConfigBlock(for: viewModel.selected)
                        ? "Yes" : "No")
    }
    GatewayBehaviorSection(
        platform: "google-chat",
        capabilities: capabilities,
        context: viewModel.context
    )
}
```

Add a `case "google-chat":` arm to the `platformForm` switch. Wrap the entire
Google Chat list entry in a capability filter so pre-v0.13 hosts don't see
it: filter `KnownPlatforms.all` through the capability flag at the
`platformList` level. New helper:

```swift
private var visiblePlatforms: [HermesToolPlatform] {
    KnownPlatforms.all.filter { p in
        switch p.name {
        case "google-chat":     return capabilities.hasGoogleChatPlatform
        case "yuanbao":         return capabilities.hasYuanbaoPlatform
        case "microsoft-teams": return capabilities.hasTeamsPlatform
        default:                return true
        }
    }
}
```

Today the Yuanbao + Teams entries are unconditionally shown — the
capability check above is the **first time** the platform list is
capability-filtered. **Decision needed (Q4).** Default position: keep
existing platforms unconditionally (avoid changing v0.12-host UX),
only filter Google Chat. Document the divergence in code comments.

#### EDIT — every existing per-platform setup view that owns an allowlist

Six platforms gain a `GatewayBehaviorSection` (allowlist editor + the three
toggles) appended below their existing form:

- `SlackSetupView.swift` — channels
- `MattermostSetupView.swift` — channels
- `TelegramSetupView.swift` — chats
- `WhatsAppSetupView.swift` — chats
- `MatrixSetupView.swift` — rooms
- (new) `GoogleChatPanel` in `PlatformsView.swift` — channels

Slack, Mattermost, Telegram, WhatsApp, Matrix already exist in
`scarf/Features/Platforms/Views/PlatformSetup/`. Each existing view's
`body` gets a single trailing `GatewayBehaviorSection(...)` call. Their
view models gain four `@Observable` properties (or use the
`GatewayBehaviorViewModel` below as a child) and the existing
`save()` runs the new save call after the existing one.

The platform setup VMs already use the `PlatformSetupHelpers.saveForm`
shape; the new behavior block is an **additive save** that runs after
the existing one. Order matters: save existing form first (so
`restartGateway` picks up the env+config edit), then save behavior.

#### NEW — `scarf/Features/Platforms/Views/PlatformSetup/Components/GatewayBehaviorSection.swift`

Reusable SwiftUI section that wraps the four v0.13 controls. Composed into
each platform setup view above its save button. Owns its own `@State`
view-model so the existing per-platform VMs don't grow another set of fields.

```swift
struct GatewayBehaviorSection: View {
    let platform: String
    let capabilities: HermesCapabilities
    let context: ServerContext

    @State private var viewModel: GatewayBehaviorViewModel
    init(platform: String, capabilities: HermesCapabilities,
         context: ServerContext) {
        self.platform = platform
        self.capabilities = capabilities
        self.context = context
        _viewModel = State(initialValue: GatewayBehaviorViewModel(
            platform: platform,
            capabilities: capabilities,
            context: context
        ))
    }

    var body: some View {
        if !capabilities.hasGatewayAllowlists
            && !capabilities.hasGatewayBusyAckToggle
            && !capabilities.hasGatewayRestartNotification {
            EmptyView()  // pre-v0.13 host — hide entire subsection
        } else {
            ScarfCard {
                ScarfSectionHeader(...)
                if capabilities.hasGatewayAllowlists,
                   let kind = GatewayAllowlistKind.kind(for: platform) {
                    AllowlistEditor(viewModel: viewModel, kind: kind)
                }
                if capabilities.hasGatewayBusyAckToggle {
                    Toggle("Send 'Agent is working…' ack",
                           isOn: $viewModel.busyAckEnabled)
                }
                if capabilities.hasGatewayRestartNotification {
                    Toggle("Post 'Gateway restarted' notice on boot",
                           isOn: $viewModel.gatewayRestartNotification)
                }
                Stepper(...)  // slash-command notice TTL
                Button("Save behavior") { viewModel.save() }
                    .buttonStyle(ScarfPrimaryButton())
            }
        }
    }
}
```

#### NEW — `scarf/Features/Platforms/Views/PlatformSetup/Components/AllowlistEditor.swift`

Reusable list-of-strings editor. UI: vertical stack of rows with a delete
glyph each, an "Add row" button that appends an empty entry and focuses
its text field. Behaviour-parity with iOS's existing `EditableStringList`
component if one already exists; otherwise a Mac-only inline component
sized for a single allowlist (typically 0-5 entries).

The editor is a **stateless** component — it binds to the view-model's
`@Binding var items: [String]`. The view-model owns persistence + change
tracking.

#### NEW — `scarf/Features/Platforms/ViewModels/PlatformSetup/GatewayBehaviorViewModel.swift`

```swift
@Observable
@MainActor
final class GatewayBehaviorViewModel {
    let platform: String
    let context: ServerContext
    let capabilities: HermesCapabilities
    let kind: GatewayAllowlistKind?       // nil for platforms w/o allowlist

    // Allowlist
    var items: [String] = []

    // Behavior toggles
    var busyAckEnabled: Bool = true
    var gatewayRestartNotification: Bool = false
    var slashCommandNoticeTTLSeconds: Int = 0

    var message: String?

    init(platform: String, capabilities: HermesCapabilities,
         context: ServerContext) { ... }

    func load() {
        let cfg = HermesFileService(context: context).loadConfig()
        if let block = cfg.gatewayPlatforms[platform] {
            switch kind {
            case .channels: items = block.allowedChannels
            case .chats:    items = block.allowedChats
            case .rooms:    items = block.allowedRooms
            case nil:       break
            }
            busyAckEnabled              = block.busyAckEnabled
            gatewayRestartNotification  = block.gatewayRestartNotification
            slashCommandNoticeTTLSeconds = block.slashCommandNoticeTTLSeconds
        }
    }

    func save() {
        // Step 1 — list write via direct YAML edit. Skip when the platform
        //          has no allowlist support.
        if let kind, capabilities.hasGatewayAllowlists {
            let ok = GatewayConfigWriter.saveList(
                context: context,
                platform: platform,
                key: kind.yamlKey,
                items: items.filter { !$0.isEmpty }
            )
            if !ok {
                message = "Failed to write allowlist to config.yaml"
                return
            }
        }

        // Step 2 — scalar saves via `hermes config set`.
        var configKV: [String: String] = [:]
        let prefix = "gateway.platforms.\(platform)."
        if capabilities.hasGatewayBusyAckToggle {
            configKV[prefix + "busy_ack_enabled"] =
                PlatformSetupHelpers.envBool(busyAckEnabled)
        }
        if capabilities.hasGatewayRestartNotification {
            configKV[prefix + "gateway_restart_notification"] =
                PlatformSetupHelpers.envBool(gatewayRestartNotification)
        }
        // TTL is always saveable on v0.13 hosts; gate via either flag.
        if capabilities.hasGatewayBusyAckToggle
            || capabilities.hasGatewayRestartNotification {
            configKV[prefix + "slash_command_notice_ttl_seconds"] =
                String(slashCommandNoticeTTLSeconds)
        }
        let result = PlatformSetupHelpers.saveForm(
            context: context, envPairs: [:], configKV: configKV
        )
        message = result
        DispatchQueue.main.asyncAfter(deadline: .now() + 3) { [weak self] in
            self?.message = nil
        }
    }
}
```

The view-model's `save()` is **MainActor**-isolated; the YAML write inside
`GatewayConfigWriter.saveList` happens on `ServerContext.writeText`, which
is `nonisolated` and hits the transport layer — safe to call from
MainActor for short writes (config.yaml is typically <50KB). Match the
existing `PlatformSetupHelpers.saveForm` posture.

#### NEW (additive) — extend the six existing per-platform setup views

Each existing per-platform setup view (`SlackSetupView`, `MatrixSetupView`,
etc.) gains one trailing `GatewayBehaviorSection(platform: "slack",
capabilities: caps, context: ctx)` before the closing `}` of its form.
The view does **not** need to know whether the host supports allowlists —
the section self-hides on pre-v0.13.

#### EDIT — `scarf/Features/Skills/Views/SkillDetailView.swift` (lightweight)

When a skill's frontmatter contains `[[as_document]]` markers in any of its
SKILL.md content, render a small info row at the top of the detail view:

```
"Media in this skill marked with `[[as_document]]` is sent as document
attachments instead of inline images on platforms that distinguish."
```

Detection is a substring scan over the SKILL.md body, run once at
SkillDetailViewModel.load. **Do not over-design** — this is informational,
not interactive.

Capability-gated on `hasGoogleChatPlatform` (cheap proxy: the `as_document`
directive shipped in v0.13 alongside Google Chat support; we don't have a
dedicated flag for it, see Q5).

### Tests

#### NEW — `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/GatewayConfigWriterTests.swift`

Round-trip + idempotence tests for the YAML list editor.

```swift
@Suite struct GatewayConfigWriterTests {
    @Test func setListInsertsBlockOnEmpty() {
        let yaml = "model:\n  default: gpt-4o\n"
        let updated = GatewayConfigWriter.setList(
            in: yaml, platform: "slack", key: "allowed_channels",
            items: ["C0123ABCD", "C0456EFGH"]
        )
        // Expect the gateway scaffold + items to be appended.
        // Expect the model.default block to be preserved verbatim.
    }

    @Test func setListReplacesExistingBlock() { ... }
    @Test func setListWithEmptyItemsRemovesBlock() { ... }
    @Test func setListPreservesOtherPlatformsBlocks() { ... }
    @Test func setListPreservesScalarSiblings() { ... }
    @Test func setListIsIdempotent() {
        let yaml = ...
        let once = GatewayConfigWriter.setList(in: yaml, ...)
        let twice = GatewayConfigWriter.setList(in: once, ...)
        #expect(once == twice)
    }
    @Test func setListQuotesItemsContainingColons() { ... }
}
```

These tests **must not** depend on Foundation specifics that differ between
macOS and Linux — keep to plain `String` operations so the suite still
runs in a Linux SwiftPM environment if it ever lands there.

#### NEW — `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/HermesGatewayListServiceTests.swift`

Parser tests for `hermes gateway list --json`.

```swift
@Suite struct HermesGatewayListServiceTests {
    @Test func parsesSingleProfileSinglePlatform() {
        let json = """
        {"profiles":[{"name":"default","running":true,"pid":1234,
        "platforms":["slack","telegram"]}]}
        """.data(using: .utf8)!
        let snap = HermesGatewayListService.parse(json)
        #expect(snap?.profiles.count == 1)
        #expect(snap?.profiles[0].profile == "default")
        #expect(snap?.profiles[0].pid == 1234)
        #expect(snap?.profiles[0].platforms == ["slack", "telegram"])
    }
    @Test func parsesMultipleProfiles() { ... }
    @Test func headerDigestSingleProfileNoneRunning() { ... }
    @Test func headerDigestMultipleProfilesSomeRunning() { ... }
    @Test func returnsNilOnEmptyJSON() { ... }
    @Test func returnsNilOnUnparseableJSON() { ... }
    @Test func toleratesUnknownKeys() { ... }
}
```

#### EDIT — `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/M6ConfigCronTests.swift`

Add coverage for the new `gateway.platforms.<platform>.*` keys in the
`HermesConfig+YAML` parser:

```swift
@Test func parsesGatewayAllowlistsForSlack() {
    let yaml = """
    gateway:
      platforms:
        slack:
          allowed_channels:
            - C01
            - C02
          busy_ack_enabled: false
          gateway_restart_notification: true
          slash_command_notice_ttl_seconds: 120
    """
    let cfg = HermesConfig(yaml: yaml)
    let block = cfg.gatewayPlatforms["slack"]
    #expect(block?.allowedChannels == ["C01", "C02"])
    #expect(block?.busyAckEnabled == false)
    #expect(block?.gatewayRestartNotification == true)
    #expect(block?.slashCommandNoticeTTLSeconds == 120)
}

@Test func gatewayPlatformsEmptyByDefault() {
    let cfg = HermesConfig(yaml: "")
    #expect(cfg.gatewayPlatforms.isEmpty)
}
```

Plus a regression test that the existing per-platform settings (Slack
`platforms.slack.reply_to_mode`, Matrix `matrix.require_mention`) still
parse correctly when the new `gateway:` block is also present —
guarantees no key collision.

#### EDIT — `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/HermesCapabilitiesTests.swift`

WS-1 already added the v0.13 flag assertions. Confirm WS-5 doesn't need to
extend this file by re-reading the suite at WS-5 PR time. **No new
assertions expected from WS-5.**

#### NEW — `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/GatewayAllowlistKindTests.swift`

Quick mapping tests:

```swift
@Test func mapsKnownPlatformsToCorrectKind() {
    #expect(GatewayAllowlistKind.kind(for: "slack") == .channels)
    #expect(GatewayAllowlistKind.kind(for: "telegram") == .chats)
    #expect(GatewayAllowlistKind.kind(for: "matrix") == .rooms)
}
@Test func returnsNilForPlatformsWithoutAllowlist() {
    #expect(GatewayAllowlistKind.kind(for: "cli") == nil)
    #expect(GatewayAllowlistKind.kind(for: "yuanbao") == nil)
}
@Test func yamlKeyMatchesHermesContract() {
    #expect(GatewayAllowlistKind.channels.yamlKey == "allowed_channels")
}
```

## New types / fields

| Type | Where | Notes |
|------|-------|-------|
| `GatewayPlatformSettings` | `ScarfCore/Models/` | All v0.13 per-platform fields in one struct. Sendable, Equatable. |
| `GatewayAllowlistKind` | `ScarfCore/Models/` | `.channels`/`.chats`/`.rooms`. Static `kind(for:)` mapper. |
| `HermesConfig.gatewayPlatforms` | `ScarfCore/Models/HermesConfig.swift` | `[String: GatewayPlatformSettings]`. Empty default. |
| `GatewayConfigWriter` | `ScarfCore/Services/` | YAML list-block read/mutate/write helper. Pure `setList(in:platform:key:items:)` for testability. |
| `GatewayListSnapshot` + `HermesGatewayListService` | `ScarfCore/Services/` | Parser + fetch helper for `hermes gateway list --json`. |
| `MessagingGatewayViewModel` | `Features/Gateway/ViewModels/` | Renamed from `GatewayViewModel`. Adds `gatewayList` snapshot. |
| `GatewayBehaviorViewModel` | `Features/Platforms/ViewModels/PlatformSetup/` | Owns the four v0.13 toggles + allowlist editing for one platform. |
| `GatewayBehaviorSection` (View) | `Features/Platforms/Views/PlatformSetup/Components/` | Reusable subsection composed into every per-platform setup view. |
| `AllowlistEditor` (View) | `Features/Platforms/Views/PlatformSetup/Components/` | Generic list-of-strings editor with `@Binding`. |
| `googleChatPanel` (private View on `PlatformsView`) | `Features/Platforms/Views/PlatformsView.swift` | Static info card mirroring `yuanbaoPanel` / `microsoftTeamsPanel`. |
| `KnownPlatforms.all` extra entry | `ScarfCore/Models/HermesTool.swift` | `name: "google-chat"` (verify Q1). |

## Capability gating

Each control gates on the most specific flag. Composite views show as little
as possible to keep pre-v0.13 hosts on a quiet UI.

| Surface | Flag | Behaviour pre-flag |
|---------|------|---------------------|
| Google Chat platform list entry | `hasGoogleChatPlatform` | Hidden |
| Cross-profile digest in `MessagingGatewayView` | `hasGatewayList` | Hidden (no row at all) |
| Allowlist editor inside `GatewayBehaviorSection` | `hasGatewayAllowlists` | Hidden |
| `busy_ack_enabled` toggle | `hasGatewayBusyAckToggle` | Hidden |
| `gateway_restart_notification` toggle | `hasGatewayRestartNotification` | Hidden |
| `slash_command_notice_ttl_seconds` field | `hasGatewayBusyAckToggle ‖ hasGatewayRestartNotification` | Hidden (proxy gate — see Q5) |
| Entire `GatewayBehaviorSection` | OR of all three | EmptyView (no card chrome) |
| `[[as_document]]` skill detail row | `hasGoogleChatPlatform` | Hidden |

The `GatewayBehaviorSection` short-circuits to `EmptyView` when none of the
three flags is on — this keeps the existing v0.12 platform forms visually
unchanged for pre-v0.13 hosts. **Important regression test (manual):** open
the Slack setup view against a v0.12.x Hermes target and confirm the page
height is identical to v2.7.5 ship.

The `MessagingGatewayView` cross-profile digest is **doubly** guarded —
`hasGatewayList` AND `gatewayList != nil`. The verb can fail on a v0.13
host that hasn't been initialized yet (no profiles registered), and the row
silently hides in that case.

## How to test

### Manual / smoke

1. **Pre-v0.13 host (v0.12.x).** All v0.13 surfaces should be invisible:
   - No Google Chat row in `Settings → Platforms`.
   - No cross-profile digest in `MessagingGatewayView`.
   - Slack / Telegram / Matrix / Mattermost / WhatsApp setup forms should
     look identical to v2.7.5 (no extra "Behavior" card).
   - SkillDetailView shows no `[[as_document]]` info row.
2. **v0.13.0 host with empty config.**
   - Google Chat row appears.
   - Each platform's setup form shows an empty "Gateway behavior" card
     with default toggles and an empty allowlist.
   - Saving with empty allowlists does not write a `gateway:` block (the
     list-write helper short-circuits on empty).
3. **v0.13.0 host with seeded config:**
   ```
   gateway:
     platforms:
       slack:
         allowed_channels: [C01, C02]
         busy_ack_enabled: false
   ```
   - Slack setup view loads with two rows in the channels editor and the
     ack toggle off.
   - Edit channels → save → re-open form → values persist.
   - Toggle ack on → save → confirm `gateway.platforms.slack.busy_ack_enabled`
     in config.yaml is `true`.
   - Run `hermes gateway list --json` manually and confirm the digest in
     the header reflects the running profiles.
4. **Cross-platform allowlists.**
   - Repeat step 3 for Telegram (chats), Matrix (rooms) — confirm the
     editor placeholder text matches the platform's identifier shape.
5. **YAML editor robustness.**
   - Manually edit config.yaml with hand-formatted allowed_channels (mixed
     indents, comments, blank lines) and confirm Scarf's save preserves the
     surrounding content. This is the sharpest edge of WS-5.
6. **Restart gateway after save.**
   - The existing "Restart Gateway" button at the foot of `PlatformsView`
     should pick up the new allowlist edit. No code change needed — it
     already runs `hermes gateway restart` after the parent save.

### Unit (Swift Testing — `scarf/Packages/ScarfCore/Tests/`)

- `GatewayConfigWriterTests` — round-trip, idempotence, quoting, empty-list
  removal, ancestor-creation. ~10 tests; the highest-value safety net in WS-5.
- `HermesGatewayListServiceTests` — parser tolerance + digest formatting.
  ~7 tests.
- `GatewayAllowlistKindTests` — pure mapping. ~3 tests.
- `M6ConfigCronTests` extension — 2-3 new YAML loader tests.

Total new test count: **~22**. Run with
`xcodebuild test -scheme scarf -only-testing ScarfCoreTests/GatewayConfigWriterTests`
etc.

### Integration (Mac app target)

- Existing `M5FeatureVMTests` / `M0dViewModelsTests` cover view-model load
  paths. Add a single test ensuring `GatewayBehaviorViewModel.load()` reads
  values from a seeded config string. Skip a write test in the integration
  suite because the existing harness uses a real `ServerContext` against
  the test fixtures — adding write coverage there risks file-leak between
  tests. The pure `GatewayConfigWriter` tests cover the write path.

## Open questions

**Q1 — Google Chat platform identifier.** Does Hermes name it `google-chat`,
`googlechat`, `gchat`, or something else? This is the wire identifier on
both the `KnownPlatforms` mapping AND the `gateway.platforms.<platform>.*`
YAML path. Implementation blocked on confirming. Default: `google-chat`
(matches the kebab-case used for `microsoft-teams`). Resolution: read
`~/.hermes/config.yaml` after running `hermes setup` against a v0.13 host
and confirm the emitted block name. Owner: implementer at WS-5 start.

**Q2 — YAML key path for allowlists + behavior toggles.** Is it
`gateway.platforms.<platform>.allowed_<kind>` or
`platforms.<platform>.allowed_<kind>` (sibling to the existing Slack
`platforms.slack.*` namespace)? Hermes' v0.13 release notes don't specify.
Implementation blocked on confirming — same resolution path as Q1.
Default in plan: `gateway.platforms.<platform>.*` (which keeps the new
config segregated from the legacy per-platform blocks). Strong contender
for collision: if Hermes uses `platforms.<platform>.allowed_<kind>`, the
allowlist write block lives in the same namespace as `platforms.slack.reply_to_mode`,
which means the YAML editor needs to be careful not to clobber the
existing scalar siblings. The pure-function tests cover this — the test
suite has a dedicated "preserves scalar siblings" assertion.

**Q3 — `hermes gateway list --json` JSON shape.** Best-guess shape used
in the parser:

```json
{"profiles":[{"name":"default","running":true,"pid":1234,
              "platforms":["slack","telegram","discord"]}]}
```

Verify against actual Hermes output once a v0.13 host is reachable. The
parser tolerates unknown keys, but the digest formatter assumes the keys
above; if they differ, the formatter changes alongside.

**Q4 — Should the Mac platform list filter the Yuanbao + Teams entries on
pre-v0.12 hosts?** Today these show unconditionally; the plan adds the
first capability filter (for Google Chat) but leaves Yuanbao + Teams alone.
This is intentional (don't change v0.12 host UX in a v0.13 work-stream),
but worth a one-line comment in the implementation. **Decision: keep
existing behaviour, only filter Google Chat. Document the deliberate
asymmetry.**

**Q5 — Capability flag for `slash_command_notice_ttl_seconds`.** Hermes
v0.13 release notes describe "Auto-delete slash-command system notices
after TTL" but the WS-1 capability matrix doesn't carry a dedicated flag.
WS-5 proxies through `hasGatewayBusyAckToggle ‖ hasGatewayRestartNotification`
because all three landed in v0.13.0 together. If a future patch separates
them, add `hasSlashCommandNoticeAutoDelete` to `HermesCapabilities` and
re-gate. Logged here so the next maintainer doesn't have to rediscover it.

**Q6 — `[[as_document]]` discoverability.** Is the directive only used in
SKILL.md bodies, or does it also appear in skill frontmatter? WS-5 plans a
substring scan; if it's frontmatter-only, the scan is wasted work and
should move to the parser. Resolution: read a v0.13 skill that uses
`[[as_document]]` and check.

**Q7 — Should `hermes gateway list` be polled, file-watched, or fetched
on demand only?** The existing `MessagingGatewayView.onAppear` triggers a
load; that's good enough for the digest because the cross-profile state
changes only when the user runs `hermes gateway start/stop` from another
profile (rare, manual). Plan: refresh on `onAppear` + on
`fileWatcher.lastChangeDate` (`gateway_state.json` writes count as
"profile state changed"). No polling.

**Q8 — Renaming the type to `MessagingGatewayViewModel`.** The view-model
is internal to `Features/Gateway/`; nothing outside that folder
references it by name. Renaming inside one work-stream is safe. The plan
applies the rename. If it triggers more than ~5 callsite churn at
implementation time, fall back to keeping the old name with a top-of-file
comment clarifying the user-facing label is "Messaging Gateway" — the
distinction is on the user surface, not the type name.

## Out of scope

Explicitly **not** part of WS-5:

- iOS Gateway tab and read-only allowlist viewer (WS-9).
- Migrating any platform setup views to ScarfDesign tokens (cross-cutting
  cleanup, separate work-stream).
- Adding a Dingtalk platform card (CLI/setup not yet wired in any Scarf
  surface; Dingtalk is mentioned in `GatewayAllowlistKind` for forward
  compat only).
- IRC plugin migration UX. Hermes v0.13 moved IRC + Teams to the
  platform-plugin architecture; the Mac Plugins tab already lists them
  generically. No change needed.
- Telegram DM user-managed multi-session topics. Server-side; no Scarf
  surface.
- Discord message-deletion action. Skill-author concern, not a
  Scarf-configurable setting.
- WhatsApp env-override home channel UX. Existing WhatsApp setup form
  already exposes the env path; no v0.13 work needed.
- Feishu mention-policy operator config. Hermes-side; surface via the
  existing Feishu setup form unchanged for now (revisit in v2.9 once
  Feishu's exact YAML keys are documented).
- Matrix `/sethome` persistence. Server-side; no Scarf write path.
- Teams sidebar threading + group-chat fallback. Server-side; no Scarf
  knob.
- Weixin content-fingerprint dedupe. Server-side.
- QQBot keyboards / chunked upload / quoted attachments. Server-side.
- ACP `/queue` slash command (separate WS).
- Persistent Goals (`/goal`) chat surface (separate WS).

## Estimate

Rough breakdown of implementation effort, assuming Q1–Q3 resolve cleanly:

| Slice | Hours | Notes |
|-------|-------|-------|
| `GatewayAllowlistKind`, `GatewayPlatformSettings`, `KnownPlatforms` extension | 1 | Pure model layer. |
| `HermesConfig.gatewayPlatforms` + YAML loader extension | 2 | Adjacent to existing extractor logic. |
| `GatewayConfigWriter` pure list-block editor | 4 | Sharp edges around YAML quoting + ancestor creation; majority of WS-5 risk. |
| `GatewayConfigWriter` integration into `ServerContext` (load → mutate → write) | 1 | Just plumbing. |
| `HermesGatewayListService` parser + fetch | 2 | Easy once Q3 resolves. |
| `MessagingGatewayView` rename + digest row | 2 | Includes ScarfDesign cleanup of existing button row. |
| `GatewayBehaviorSection` + `AllowlistEditor` | 4 | Reusable Mac component. |
| `GatewayBehaviorViewModel` | 2 | Mostly orchestrating already-built pieces. |
| Compose into 5 existing platform views + Google Chat panel | 2 | Repetitive. |
| Google Chat platform list filter + entry | 1 | Smallest slice. |
| `[[as_document]]` skill detail row | 1 | Single string-scan + tooltip. |
| Tests (~22 cases across 4 files) | 4 | Bulk concentrated in `GatewayConfigWriterTests`. |
| Manual QA against v0.12.x + v0.13.0 hosts | 2 | Both hosts on Alan's mini. |
| Wiki updates (`Messaging-Gateway.md`, `Platforms.md`, `Hermes-Version-Compatibility.md`) | 1 | Per CLAUDE.md wiki policy. |
| Buffer for Q1–Q3 resolution churn | 3 | Likely small re-key passes. |
| **Total** | **~32h** | ~4 working days. |

The risk concentration is on `GatewayConfigWriter` (sharp YAML editing) and
the unverified Hermes contract details (Q1–Q3). Everything else is wiring
that follows existing patterns in `Features/Platforms/`.

## Notes for the implementer

- **`hermes config set` does not handle list values.** This is a documented
  limitation in the existing Home Assistant view (see header comment in
  `HomeAssistantSetupViewModel.swift`). WS-5 sidesteps it via direct YAML
  editing for the three list keys. Do **not** try to make `hermes config set
  gateway.platforms.slack.allowed_channels '["C01","C02"]'` work; it
  serializes badly and Hermes rejects the result.
- **Always `Task.detached` the YAML round-trip on remote hosts.** On
  `ServerContext.local`, `writeText` is a single FS hit; on remote hosts
  it's an SCP round-trip that should not block MainActor (per
  `~/.claude/CLAUDE.md` Swift 6 rules). The pattern in
  `KanbanService.runHermes` is the canonical reference; the
  `GatewayBehaviorViewModel.save()` plan above must adopt the same posture
  before any remote-host testing.
- **Capability detection cache.** `HermesCapabilitiesStore` is per-server.
  When the user switches servers, the store rebuilds; the
  `GatewayBehaviorSection` is created with the new store's capabilities at
  view-init time. No manual invalidation needed.
- **Sidebar enum case stays.** `SidebarSection.gateway` does not become
  `SidebarSection.messagingGateway`. Rename the user-facing label only.
- **ScarfDesign cleanup on touch.** The existing `GatewayView.swift` uses
  raw `Button("Start") { … }` and `cornerRadius: 8`/`padding(12)` literals.
  When this work-stream re-touches the file, swap to
  `ScarfPrimaryButton` / `ScarfSecondaryButton` and `ScarfRadius.md` /
  `ScarfSpace.s3` to be consistent with neighbors.
- **Error surfaces.** `GatewayBehaviorViewModel.message` mirrors the
  existing per-platform `message` field. Do not introduce a separate
  banner system; the existing inline-toast at 3-second auto-clear is the
  established pattern.
- **Composability with Templates.** Project templates (v2.3) cannot
  populate the new `gateway.platforms.<platform>.*` block — that lives in
  `~/.hermes/config.yaml`, which the v1 installer is forbidden to touch
  per the explicit invariant in CLAUDE.md ("Never let a template write to
  config.yaml"). No interaction surface between WS-5 and Templates work.
