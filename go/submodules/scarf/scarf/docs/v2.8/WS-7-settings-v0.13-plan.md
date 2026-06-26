# WS-7 Plan: Settings tab additions

**Workstream:** WS-7 of Scarf v2.8.0
**Hermes target:** v0.13.0 (v2026.5.7)
**Capability gates (already shipped in WS-1):**
- `HermesCapabilities.hasMCPSSETransport` (`>= 0.13.0`)
- `HermesCapabilities.hasCronNoAgent` (`>= 0.13.0`)
- `HermesCapabilities.hasWebToolsBackendSplit` (`>= 0.13.0`)
- `HermesCapabilities.hasProfileNoSkills` (`>= 0.13.0`)

**Builds on:**
- v2.7.5 MCP Servers feature (`Features/MCPServers/`) — list + detail + add (preset / custom) + edit + per-server delete + OAuth token surface.
- v2.7.5 Cron feature (`Features/Cron/`) — `--workdir` already plumbed through `CronJobEditor` + `CronViewModel.createJob` / `updateJob`. Provides the precedent for v0.13 capability-gated form fields.
- v2.7.5 Settings feature (`Features/Settings/`) — 10 tabs, single `SettingsViewModel` write surface routing through `setSetting(key, value)` → `hermes config set <key> <value>`.
- v2.7.5 Profiles feature (`Features/Profiles/`) — Mac (read/write) + iOS (read-only); Mac create-sheet has `--clone` / `--clone-all` toggles today.

**Owner:** TBD
**Reviewers:** Alan; whoever rides Settings/Profiles during v2.8.

---

## Goals

Four small, independent additions, each gated on its own v0.13 capability flag. Each lands as its own commit inside the WS-7 PR so reviewers can scan them as four self-contained changes.

1. **MCP SSE transport** — third transport option alongside `stdio` and `http` (which Hermes calls "pipe" when it means stdin/stdout JSON-RPC; "http" in our code is the HTTP transport — see Open Questions). Adds `URL` + `sse_read_timeout` fields to the add-server flow and the editor; surfaces the "SSE" segment only on v0.13+ hosts.
2. **Cron `--no-agent`** — script-only watchdog jobs. New toggle in `CronJobEditor`; when ON, the prompt + skills fields collapse with a hint. Maps to `--no-agent` on `hermes cron create / edit`. Read-side adds `noAgent: Bool?` to `HermesCronJob` for round-trip tolerance.
3. **Web Tools backend split** — `web_search` and `web_extract` config keys gain distinct backends. Net-new tab "Web Tools" in `SettingsView` with two backend pickers. Pre-v0.13 hosts see a legacy combined picker (single `web_tools.backend` key) rendered inside the same tab so the chrome stays consistent.
4. **Profiles `--no-skills`** — Mac create-profile sheet gains an "Empty profile (no skills)" toggle that appends `--no-skills` to `hermes profile create`. iOS is read-only and out of scope.

### Non-goals

- **Live MCP SSE wire-format probing.** WS-7 only writes the YAML + surfaces the field. Hermes owns the runtime connect; Scarf trusts `hermes mcp test <name>` to verify.
- **MCP `pipe` transport surface.** v0.13 release notes mention "Retry stale pipe transport failures as session-expired" — pipe is Hermes-internal jargon for the existing stdio transport (per parser logic at `HermesFileService.parseMCPServersBlock` and `MCPTransport` enum cases). No new user-facing transport option for "pipe".
- **`web_tools.search.<backend>.<api_key>` deep settings.** Backend-specific tuning (e.g. SearXNG host URL, Tavily API key) stays in raw YAML editor for v2.8. Per-backend config sheets are a follow-up — the "split" is the v0.13 wire change WS-7 must ship.
- **iOS `--no-skills`.** iOS Profiles is read-only (per CLAUDE.md "v0.12 iOS catch-up (Phase H)" and `Scarf iOS/Profiles/ProfilesView.swift`). No new toggles on iOS.
- **Cron `--no-agent` retroactive flagging.** A v0.13 host whose `~/.hermes/cron/jobs.json` already has `no_agent: true` jobs gets the badge for free via the new `noAgent` field; no migration UX.

---

## 1. MCP SSE transport

### Files / changes

#### 1a. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesMCPServer.swift`

**Why:** `MCPTransport` is currently a 2-case enum (`stdio`, `http`). Adding `sse` keeps SwiftUI Picker code paths simple — the existing `Picker(selection: $transport) { ForEach(MCPTransport.allCases) { ... } }` in `MCPServerAddCustomView` then iterates three cases automatically.

**Edits:**

- Extend `MCPTransport`:
  ```swift
  public enum MCPTransport: String, Sendable, Equatable, CaseIterable, Identifiable {
      case stdio
      case http
      case sse   // v0.13+
      ...
  }
  ```
- Add `displayName` case for `.sse`: `"Remote (SSE)"`.
- Add a single new stored property to `HermesMCPServer`:
  - `public let sseReadTimeout: Int?` — seconds. `nil` when the YAML doesn't specify `sse_read_timeout`.
- Append `sseReadTimeout: Int? = nil` to the memberwise initializer's tail (defaulted) so existing call sites compile unchanged. Mirrors how `connectTimeout` lives next to `timeout`.
- Update `summary` so `.sse` returns `url ?? ""` (same shape as `.http`).

**Tolerance contract:** A pre-v0.13 server entry with no `url` and no `sse_read_timeout` parses as `.stdio`. A v0.13 entry with `url` + `sse_read_timeout` parses as `.sse` — see parser change below.

#### 1b. `scarf/scarf/Core/Services/HermesFileService.swift`

**Why:** YAML parser at `parseMCPServersBlock` (line 796) currently distinguishes stdio vs http with `let transport: MCPTransport = fields["url"] != nil ? .http : .stdio`. SSE also has a `url`, so we need a second discriminator.

**Edits:**

- Inside the `flush()` closure (around line 815), replace the binary discriminator with a 3-way one:
  ```swift
  let transport: MCPTransport = {
      if fields["transport"]?.lowercased() == "sse" { return .sse }
      if fields["url"] != nil { return .http }
      return .stdio
  }()
  ```
  Hermes v0.13's `mcp add --url <https://...> --transport sse` writes a `transport: sse` scalar into the YAML entry; older hosts emit no `transport` key, defaulting to `.http` for url-based entries and `.stdio` otherwise. This preserves byte-for-byte round-trip on existing files.
- Read `sse_read_timeout` from `fields["sse_read_timeout"]`, parse as `Int?`, pass into `HermesMCPServer` initializer.
- New writer method:
  ```swift
  @discardableResult
  nonisolated func addMCPServerSSE(name: String, url: String, sseReadTimeout: Int?) -> (exitCode: Int32, output: String) {
      var args = ["mcp", "add", name, "--url", url, "--transport", "sse"]
      if let t = sseReadTimeout { args += ["--sse-read-timeout", String(t)] }
      return runHermesCLI(args: args, timeout: 45, stdinInput: "y\ny\ny\n")
  }
  ```
  Verify the exact CLI flag name during integration — `--sse-read-timeout` is the natural form but Hermes may have shipped it as `--sse-read-timeout-seconds` or merged it under `--timeout`. See Open Questions.
- New writer for changing `sse_read_timeout` post-create:
  ```swift
  @discardableResult
  nonisolated func setMCPServerSSETimeout(name: String, sseReadTimeout: Int?) -> Bool {
      patchMCPServerField(name: name) { entryLines in
          if let t = sseReadTimeout {
              Self.replaceOrInsertScalar(key: "sse_read_timeout", value: String(t), in: &entryLines)
          } else {
              Self.removeScalar(key: "sse_read_timeout", in: &entryLines)
          }
      }
  }
  ```
  Mirrors `setMCPServerTimeouts` line-for-line.

**Round-trip invariant:** Adding an SSE server through `addMCPServerSSE`, then editing its `sse_read_timeout` through `setMCPServerSSETimeout`, then re-loading, must produce the same `HermesMCPServer.sseReadTimeout` value. Test fixture below.

#### 1c. `scarf/scarf/Features/MCPServers/Views/MCPServerAddCustomView.swift`

**Why:** This is the add-server form. It currently has a 2-segment transport picker.

**Edits:**

- Add `@Environment(\.hermesCapabilities) private var capabilitiesStore`.
- Add `@State private var sseReadTimeout: String = ""`.
- Replace the static `Picker { ForEach(MCPTransport.allCases) }` segmented control with a filtered list that drops `.sse` when capability is off:
  ```swift
  private var availableTransports: [MCPTransport] {
      var t: [MCPTransport] = [.stdio, .http]
      if capabilitiesStore?.capabilities.hasMCPSSETransport ?? false { t.append(.sse) }
      return t
  }
  ```
  Render with `ForEach(availableTransports) { ... }`. Iterating `MCPTransport.allCases` would render the SSE option even on pre-v0.13 hosts, which Hermes argparse would reject.
- Branch the body: when `transport == .sse`, render an `sseSection` next to (not replacing) the existing `httpSection`. Shape:
  ```swift
  private var sseSection: some View {
      sectionBox(title: "Endpoint (SSE)") {
          VStack(alignment: .leading, spacing: 8) {
              VStack(alignment: .leading, spacing: 4) {
                  Text("URL").font(.caption.bold())
                  TextField("https://.../sse", text: $url)
                      .textFieldStyle(.roundedBorder)
                      .font(.system(.body, design: .monospaced))
              }
              VStack(alignment: .leading, spacing: 4) {
                  Text("SSE Read Timeout (seconds)").font(.caption.bold())
                  TextField("default 300", text: $sseReadTimeout)
                      .textFieldStyle(.roundedBorder)
                      .frame(maxWidth: 140)
                  Text("Hermes-side keepalive interval. Leave blank to use the default.")
                      .font(.caption2)
                      .foregroundStyle(.secondary)
              }
          }
      }
  }
  ```
  Default placeholder reads `default 300` since Hermes v0.13's `sse_read_timeout` defaults to 300s (verify against `~/.hermes/hermes-agent/hermes_cli/mcp.py` during integration; if it's 60s or 600s adjust the placeholder copy).
- Adjust `canSubmit` + `submit()`:
  - `case .sse: return !url.trimmingCharacters(in: .whitespaces).isEmpty`
  - In `submit()`, dispatch based on `transport`:
    ```swift
    switch transport {
    case .stdio: viewModel.addCustom(...)         // existing
    case .http:  viewModel.addCustom(...)         // existing
    case .sse:   viewModel.addCustomSSE(name: trimmedName, url: ..., sseReadTimeout: Int(sseReadTimeout))
    }
    ```

#### 1d. `scarf/scarf/Features/MCPServers/ViewModels/MCPServersViewModel.swift`

**Edits:**

- New method:
  ```swift
  func addCustomSSE(name: String, url: String, sseReadTimeout: Int?) {
      let fileService = self.fileService
      Task.detached {
          let result = fileService.addMCPServerSSE(name: name, url: url, sseReadTimeout: sseReadTimeout)
          await MainActor.run {
              if result.exitCode == 0 {
                  self.flashStatus("Added \(name)")
                  self.load()
                  self.selectedServerName = name
                  self.showRestartBanner = true
                  self.showAddCustom = false
              } else {
                  self.activeError = "Add failed: \(result.output)"
              }
          }
      }
  }
  ```
- Optional cosmetic: add a third filtered list `sseServers: [HermesMCPServer]` matching the `stdioServers` / `httpServers` pattern, plus a third `Section("Remote (SSE)")` in `MCPServersView.serversList`. Keeping the two existing sections + a new one mirrors the existing UX better than collapsing all remote into one section.

#### 1e. `scarf/scarf/Features/MCPServers/Views/MCPServersView.swift`

**Edits:**

- Add a third `if !viewModel.sseServers.isEmpty { Section("Remote (SSE)") { ... } }` block in `serversList`. The icon for the row stays `network` (same as http) — the "(SSE)" label in the section header is the differentiator.
- No capability gate inside `MCPServersView` — pre-v0.13 hosts simply have no `.sse` entries to render.

#### 1f. `scarf/scarf/Features/MCPServers/Views/MCPServerEditorView.swift`

**Why:** Edit existing server's `sse_read_timeout`. The editor today exposes `timeout` + `connect_timeout` in `timeoutsSection`; SSE servers want a third numeric.

**Edits:**

- Add `@Environment(\.hermesCapabilities)` so the editor can know whether the field is editable.
- Branch `timeoutsSection` on `viewModel.server.transport`:
  - For `.stdio` and `.http`: render the existing connect/call timeouts.
  - For `.sse`: render the existing connect/call timeouts AND add a third "SSE Read Timeout" field bound to `viewModel.sseReadTimeoutDraft`.
- Update `MCPServerEditorViewModel`:
  - Add `var sseReadTimeoutDraft: String` initialized from `server.sseReadTimeout.map(String.init) ?? ""`.
  - Inside `save()`, when `transport == .sse`, call `service.setMCPServerSSETimeout(name: name, sseReadTimeout: Int(sseReadTimeoutDraft))` alongside the existing `setMCPServerTimeouts` call. A failure flips `ok = false` like the others.

#### 1g. `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/HermesMCPServerYAMLTests.swift` (NEW or extension to existing)

**Tests:**

1. `parseMCPServersBlock_v013_sseEntry_decodesAsSSE` — fixture YAML with `transport: sse` + `url: https://...` + `sse_read_timeout: 300` parses to `.sse` transport with the right `sseReadTimeout` value.
2. `parseMCPServersBlock_v012_httpEntry_stillDecodesAsHTTP` — pre-v0.13 entry without `transport:` still resolves to `.http` when `url` is present.
3. `parseMCPServersBlock_v012_stdioEntry_stillDecodesAsStdio` — entry with no `url` and no `transport:` resolves to `.stdio`.
4. `setMCPServerSSETimeout_writesAndClears` — round-trip integration test using a temp YAML: write `300`, re-read, assert; write `nil`, re-read, assert key removed.

### Capability gating

- **Add-server form:** `availableTransports` filter drops `.sse` when `hasMCPSSETransport` is false. Pre-v0.13 hosts see only "stdio | http" segments. The toolbar add button stays unconditional — the gate lives inside the form.
- **Editor:** `sse_read_timeout` field renders only for servers whose `transport == .sse`. Since pre-v0.13 hosts can't write SSE servers, the field never appears for those users. (Defensive: if a v0.13 server is somehow viewed on a pre-v0.13 host — e.g. user downgraded Hermes — the editor still reads + writes the field. Hermes will ignore it. Acceptable.)
- **List rendering:** `Section("Remote (SSE)")` only renders when `sseServers` is non-empty, so pre-v0.13 hosts don't see an empty section.

### Tests

- ScarfCore: 4 YAML-parser tests above + 2 model tests (`MCPTransport.allCases.count == 3`, `sseReadTimeout` round-trips through memberwise init).
- ScarfTests (Mac app): `MCPServersViewModelTests.testAddCustomSSE` mock-fileservice test verifying the `--transport sse --sse-read-timeout` flag shape.

### Rollout

- Feature-gate behind `hasMCPSSETransport` so a pre-v0.13 host never sees the SSE option.
- No migration: existing stdio/http servers are unaffected.
- One commit. Should land at ~250-350 LOC additions across 6 files.

---

## 2. Cron `--no-agent` toggle

### Files / changes

#### 2a. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesCronJob.swift`

**Why:** Read-side support so `loadCronJobs()` can round-trip `no_agent: true` from `~/.hermes/cron/jobs.json`. Pre-v0.13 jobs.json files don't have the field — the existing `decodeIfPresent` pattern (line 113 for `workdir`) handles that.

**Edits:**

- Add `public nonisolated let noAgent: Bool?` between `workdir` and `contextFrom`.
- Extend `enum CodingKeys` with `case noAgent = "no_agent"`.
- Extend the public memberwise initializer's tail with `noAgent: Bool? = nil`.
- Extend `init(from decoder:)`: `self.noAgent = try c.decodeIfPresent(Bool.self, forKey: .noAgent)`.
- Extend `encode(to encoder:)`: `try c.encodeIfPresent(noAgent, forKey: .noAgent)`.

**Tolerance contract:** A pre-v0.13 jobs.json with no `no_agent` field decodes with `noAgent == nil`. A v0.13 jobs.json with explicit `no_agent: false` decodes with `noAgent == false`. The "render the badge?" check is `job.noAgent == true` (treats `nil` and `false` identically — a script-only job must opt in).

#### 2b. `scarf/scarf/Features/Cron/Views/CronView.swift`

**Edits:**

- Extend `CronJobEditor.FormState` with `var noAgent: Bool = false`.
- Add `let supportsNoAgent: Bool` next to the existing `let supportsWorkdir: Bool`.
- Inside `body`, add a Toggle row near the bottom of the form (after `Workdir`, before `availableSkills`):
  ```swift
  if supportsNoAgent {
      Toggle("Run script only (no agent call)", isOn: $form.noAgent)
          .scarfStyle(.body)
          .tint(ScarfColor.accent)
      if form.noAgent {
          Text("Watchdog mode — Hermes runs the pre-run script and skips the AI turn. Prompt + skills are ignored.")
              .scarfStyle(.caption)
              .foregroundStyle(ScarfColor.foregroundMuted)
              .padding(.leading, ScarfSpace.s3)
      }
  }
  ```
- Conditionally collapse the prompt + skills sections when `form.noAgent` is true. Don't *remove* them from the view tree — keep them rendered but visually muted (and perhaps disabled). This avoids the layout shift surprise of fields disappearing mid-edit:
  ```swift
  // around the existing Prompt TextEditor
  .opacity(form.noAgent ? 0.4 : 1.0)
  .disabled(form.noAgent)
  .accessibilityHint(form.noAgent ? Text("Disabled — Run script only is on") : Text(""))
  ```
  Apply the same to the Skills picker. Script field stays fully active — it's the load-bearing thing in `--no-agent` mode.
- On entering edit mode (the existing `.onAppear` handler), hydrate `form.noAgent = job.noAgent ?? false`.
- Wire through to the parent: pass `form.noAgent` in the `onSave(form)` callback. The parent's `viewModel.createJob` / `updateJob` then knows the flag.

#### 2c. `scarf/scarf/Features/Cron/Views/CronView.swift` — owner site

**Edits:**

- Add a private capability accessor next to `hasCronWorkdir`:
  ```swift
  private var hasCronNoAgent: Bool {
      capabilitiesStore?.capabilities.hasCronNoAgent ?? false
  }
  ```
- Plumb `supportsNoAgent: hasCronNoAgent` into `CronJobEditor` instantiations (both the create and edit sheet paths, mirroring how `supportsWorkdir` is wired).
- Update the create + edit `.sheet` closures to pass `noAgent: form.noAgent` into `viewModel.createJob` / `updateJob`. Mirror the `workdir` strip-on-pre-v0.12 pattern: pass `hasCronNoAgent ? form.noAgent : false`. (For the update path, pass `hasCronNoAgent ? form.noAgent : nil` if the underlying VM signature distinguishes "don't touch" from "set false" — see VM section below.)

#### 2d. `scarf/scarf/Features/Cron/ViewModels/CronViewModel.swift`

**Edits:**

- Extend `createJob` signature with `noAgent: Bool = false` at the tail:
  ```swift
  func createJob(schedule: String, prompt: String, name: String, deliver: String, skills: [String], script: String, repeatCount: String, workdir: String = "", noAgent: Bool = false) {
      var args = ["cron", "create"]
      ...
      if noAgent { args.append("--no-agent") }
      args.append(schedule)
      // When --no-agent is set Hermes ignores the prompt arg, but argparse still
      // wants positional args to line up with the schedule. Pass an empty string
      // explicitly so the positional parser doesn't treat the prompt as missing.
      if noAgent {
          args.append("")
      } else if !prompt.isEmpty {
          args.append(prompt)
      }
      runAndReload(args, success: "Job created")
  }
  ```
  Verify Hermes's argparse behavior during integration — if `cron create --no-agent <schedule>` rejects the trailing empty positional, drop the empty-string append.
- Extend `updateJob` signature with `noAgent: Bool? = nil`:
  ```swift
  func updateJob(id: String, ..., workdir: String? = nil, noAgent: Bool? = nil) {
      var args = ["cron", "edit", id]
      ...
      if let noAgent {
          // Hermes documents `--no-agent` as a flag on `cron edit` for v0.13+.
          // Verify exact toggle-off shape (likely `--no-agent=false` or
          // `--agent` to flip back). See Open Questions.
          if noAgent { args.append("--no-agent") }
          else { args.append("--agent") }
      }
      runAndReload(args, success: "Updated")
  }
  ```

#### 2e. `scarf/scarf/Features/Cron/Views/CronView.swift` — detail rendering

**Edits (cosmetic, optional but high-value):** When the selected job has `noAgent == true`, render a small `ScarfBadge("script-only", kind: .info)` in `detailHeader` next to the existing `paused` / `running…` badges so the user can tell at a glance which jobs are watchdogs. Same in the `cronRow` list — append a `ScarfBadge("no-agent", kind: .neutral)` when the flag is on, similar to the existing `paused` badge.

### Capability gating

- **Editor toggle:** rendered only when `supportsNoAgent` is true. Pre-v0.13 hosts never see the field.
- **Defensive write-strip:** `CronView` passes `hasCronNoAgent ? form.noAgent : false` on create and `hasCronNoAgent ? form.noAgent : nil` on edit. Mirrors the `workdir` strip from v0.12 (`workdir: hasCronWorkdir ? form.workdir : ""` on create, `nil` on edit).
- **Read-side rendering:** badges + collapsed-fields visual cue render unconditionally when `job.noAgent == true`. A user who downgraded Hermes after creating a `no_agent` job still sees it labeled correctly, even though they can no longer create new ones.

### Tests

- `M6ConfigCronTests` extension: add `decodes_v013_jobs_json_with_no_agent` — fixture jobs.json with one job carrying `no_agent: true`. Assert `job.noAgent == true`.
- `M6ConfigCronTests`: `decodes_v012_jobs_json_no_no_agent_field` — pre-v0.13 fixture, assert `job.noAgent == nil`.
- `CronViewModelNoAgentTests` (new): mock-fileservice test asserting `createJob(..., noAgent: true)` produces `["cron", "create", "--no-agent", schedule, ""]` (or whatever argparse shape we converge on after integration).
- Manual: pre-v0.13 host — toggle absent in editor. v0.13 host — toggle present, creating a script-only job with no AGENTS.md context completes without an LLM call (verify in `~/.hermes/logs/`).

### Rollout

- One commit. ~150-200 LOC across 4 files (model + view + editor + VM).

---

## 3. Web Tools backend split

### Files / changes

A net-new Settings tab. Today there is no Web Tools tab — `web_extract`'s **provider** lives in Aux Models, but `web_tools.search.backend` / `web_tools.extract.backend` (the backend-not-provider keys) are not surfaced by Scarf today (verified: `grep web_tools = ` returns no Scarf hits). v0.13 makes "split per capability" the wire model, so introducing the tab here gives us a clean substrate to add backend-specific rows on later.

Layout shape:

- Pre-v0.13: a single row "Combined backend" → `web_tools.backend` key (legacy v0.12 shape).
- v0.13+: two rows — "Search backend" → `web_tools.search.backend`, "Extract backend" → `web_tools.extract.backend`. SearXNG appears in the Search picker only.

Both shapes coexist in the same tab; the gate decides which renders.

#### 3a. `scarf/scarf/Features/Settings/Views/SettingsView.swift`

**Edits:**

- Add a new case to `SettingsTab`:
  ```swift
  case webTools = "Web Tools"
  ```
  Position: between `.browser` and `.voice` (browser-adjacent in the user's mental model). Update `displayName`, `icon` (`"globe.americas"`), and `tabContent` switch.
- `tabContent` adds: `case .webTools: WebToolsTab(viewModel: viewModel)`.

#### 3b. `scarf/scarf/Features/Settings/Views/Tabs/WebToolsTab.swift` (NEW)

**Why:** Self-contained tab file matching the existing pattern (`BrowserTab.swift`, `TerminalTab.swift`, etc.). Pre-v0.13 + v0.13+ shapes both live here behind a capability check.

**Shape:**

```swift
import SwiftUI
import ScarfCore
import ScarfDesign

struct WebToolsTab: View {
    @Bindable var viewModel: SettingsViewModel
    @Environment(\.hermesCapabilities) private var capabilitiesStore

    private var split: Bool {
        capabilitiesStore?.capabilities.hasWebToolsBackendSplit ?? false
    }

    private static let searchBackends: [String] = [
        "duckduckgo", "tavily", "brave", "exa", "you", "searxng"
    ]
    private static let extractBackends: [String] = [
        "reader", "browserless", "trafilatura", "firecrawl"
    ]
    private static let combinedBackends: [String] = [
        "duckduckgo", "tavily", "brave", "exa", "you", "reader", "browserless", "trafilatura", "firecrawl"
    ]

    var body: some View {
        VStack(alignment: .leading, spacing: ScarfSpace.s5) {
            SettingsSection(title: "Web Tools", icon: "globe.americas") {
                if split {
                    Picker("Search backend", selection: Binding(
                        get: { viewModel.config.webToolsSearchBackend },
                        set: { viewModel.setWebToolsSearchBackend($0) }
                    )) {
                        ForEach(Self.searchBackends, id: \.self) { Text($0).tag($0) }
                    }
                    Text("SearXNG joined v0.13 as a search-only backend.")
                        .scarfStyle(.caption)
                        .foregroundStyle(ScarfColor.foregroundMuted)
                    Picker("Extract backend", selection: Binding(
                        get: { viewModel.config.webToolsExtractBackend },
                        set: { viewModel.setWebToolsExtractBackend($0) }
                    )) {
                        ForEach(Self.extractBackends, id: \.self) { Text($0).tag($0) }
                    }
                } else {
                    Picker("Backend", selection: Binding(
                        get: { viewModel.config.webToolsBackend },
                        set: { viewModel.setWebToolsBackend($0) }
                    )) {
                        ForEach(Self.combinedBackends, id: \.self) { Text($0).tag($0) }
                    }
                    Text("Hermes v0.13 splits search and extract into separate backends. Update Hermes to access the per-capability picker.")
                        .scarfStyle(.caption)
                        .foregroundStyle(ScarfColor.foregroundFaint)
                }
            }
        }
    }
}
```

The backend lists are intentionally small + curated. **The exact set must be reconciled against `~/.hermes/hermes-agent/hermes_cli/web_tools.py` (or wherever Hermes registers the dispatch table)** during integration. See Open Questions.

#### 3c. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift`

**Edits:**

- Add three new top-level fields to `HermesConfig` (next to `redactionEnabled` near line 663, since they share the v0.12+ migration tail comment):
  ```swift
  /// Pre-v0.13: single combined backend at `web_tools.backend`. v0.13
  /// flipped to per-capability split (see below). Kept for round-trip
  /// on hosts that never migrated.
  public var webToolsBackend: String          // default "duckduckgo"
  /// v0.13+: `web_tools.search.backend`. SearXNG can land here.
  public var webToolsSearchBackend: String    // default "duckduckgo"
  /// v0.13+: `web_tools.extract.backend`.
  public var webToolsExtractBackend: String   // default "reader"
  ```
- Add to the memberwise initializer at the tail with defaults so v2.7.5 call sites still compile.
- Extend `.empty` with `"duckduckgo"` / `"duckduckgo"` / `"reader"` defaults.

#### 3d. `scarf/Packages/ScarfCore/Sources/ScarfCore/Parsing/HermesConfig+YAML.swift`

**Edits:** Read three new keys via the existing `str(...)` helper:
- `webToolsBackend: str("web_tools.backend", default: "duckduckgo")`
- `webToolsSearchBackend: str("web_tools.search.backend", default: "duckduckgo")`
- `webToolsExtractBackend: str("web_tools.extract.backend", default: "reader")`

Pre-v0.13 YAML has only `web_tools.backend`; the two split keys default to the same value. v0.13 YAML may have `web_tools.search.backend` set and `web_tools.backend` absent — the legacy field falls back to its default but is unused on v0.13 hosts (the tab gates on `hasWebToolsBackendSplit`).

#### 3e. `scarf/scarf/Features/Settings/ViewModels/SettingsViewModel.swift`

**Edits:** Three new setters:
```swift
func setWebToolsBackend(_ value: String) { setSetting("web_tools.backend", value: value) }
func setWebToolsSearchBackend(_ value: String) { setSetting("web_tools.search.backend", value: value) }
func setWebToolsExtractBackend(_ value: String) { setSetting("web_tools.extract.backend", value: value) }
```
All three route through `hermes config set <key> <value>` — the v0.13 CLI accepts the dotted path keys as written. Hermes config-set rejects unknown keys, so on a pre-v0.13 host `setWebToolsSearchBackend` would fail; we don't expose the call site there (the picker isn't rendered).

### Capability gating

- **Tab itself:** the tab is always shown — pre-v0.13 hosts see the legacy combined picker so they're not blocked from configuring Web Tools at all. Removing the tab entirely on pre-v0.13 hosts would create a feature regression for users on v0.12.
- **Picker shape:** `split` flag inside `WebToolsTab` chooses between the two shapes.
- **SearXNG visibility:** appears only in `searchBackends` (the v0.13 split case). Never in `combinedBackends`. This matches Hermes — pre-v0.13 doesn't dispatch SearXNG at all.

### Tests

- `HermesConfigYAMLTests`:
  1. `parses_v012_combined_backend` — fixture with `web_tools.backend: tavily`, no split keys → `webToolsBackend == "tavily"`, split keys == defaults.
  2. `parses_v013_split_backend` — fixture with both `web_tools.search.backend: searxng` + `web_tools.extract.backend: reader` → both split keys populated.
  3. `parses_v013_partial` — fixture with only `web_tools.search.backend` set (extract uses default) → search populated, extract == default.
- Manual: load v0.12 host → see combined picker. Load v0.13 host → see split. Confirm SearXNG only in Search.

### Rollout

- One commit. ~200-260 LOC: 1 new file (~80 LOC), edits to 4 existing files. New tab makes this the largest of the four additions.
- Add an entry to the Settings tab strip — verify horizontal scroll still fits 11 tabs comfortably (it should; the strip is `.scrollView(.horizontal)` already).

---

## 4. Profiles `--no-skills` toggle

### Files / changes

#### 4a. `scarf/scarf/Features/Profiles/Views/ProfilesView.swift`

**Edits:**

- Add `@Environment(\.hermesCapabilities) private var capabilitiesStore` next to the existing state.
- Add `@State private var createNoSkills: Bool = false` next to `createCloneConfig` / `createCloneAll`.
- Inside `createSheet`, add a new toggle row between the existing toggles:
  ```swift
  if capabilitiesStore?.capabilities.hasProfileNoSkills ?? false {
      Toggle("Empty profile (no skills)", isOn: $createNoSkills)
          .disabled(createCloneAll)  // mutually exclusive with full clone
  }
  ```
  Why disabled when `createCloneAll`: a full clone copies skills wholesale — `--no-skills` would be a contradiction. Hermes likely rejects the combination but the UX is cleaner if we don't let the user reach it.
- Reset on sheet open: in the existing reset (line 126: `createName = ""; createCloneConfig = true; createCloneAll = false`), add `createNoSkills = false`.
- Wire to the VM:
  ```swift
  Button("Create") {
      viewModel.create(name: createName, cloneConfig: createCloneConfig, cloneAll: createCloneAll, noSkills: createNoSkills)
      showCreate = false
  }
  ```

#### 4b. `scarf/scarf/Features/Profiles/ViewModels/ProfilesViewModel.swift`

**Edits:** Extend `create` signature with `noSkills: Bool = false`:

```swift
func create(name: String, cloneConfig: Bool, cloneAll: Bool, noSkills: Bool = false) {
    var args = ["profile", "create", name]
    if cloneAll { args.append("--clone-all") }
    else if cloneConfig { args.append("--clone") }
    if noSkills { args.append("--no-skills") }
    runAndReload(args, success: "Profile '\(name)' created")
}
```

The `--no-skills` flag is independent of `--clone` / `--clone-all` per the v0.13 release notes ("`--no-skills` flag for empty profile creation"). The UX disables the toggle under `--clone-all` for clarity, but the wire is unconditional — the user can stack `--clone --no-skills` to clone config but skip skills, which is a plausible workflow.

### Capability gating

- **Toggle visibility:** wrapped in `capabilitiesStore?.capabilities.hasProfileNoSkills ?? false`. Pre-v0.13 hosts never see it.
- **Defensive write-strip:** the VM always reads `noSkills` as the default `false` if the form didn't surface the toggle. No need for a `?? false` strip at the call site — the parameter has a default in the VM signature.

### Tests

- `ProfilesViewModelTests` (new or extension): `create_emitsNoSkillsFlagWhenSet` — mock-fileservice asserting `["profile", "create", "name", "--no-skills"]` for `noSkills: true`.
- `create_combinesCloneAndNoSkills` — `["profile", "create", "name", "--clone", "--no-skills"]`.
- `create_omitsNoSkillsByDefault` — verifies the v2.7.5 signature still produces the v2.7.5 args.
- Manual: pre-v0.13 host — toggle absent. v0.13 host — toggle creates an empty `~/.hermes/profiles/<name>/skills/` (verify on disk).

### Rollout

- One commit. ~30-50 LOC across 2 files. Smallest of the four additions.

---

## Open questions

1. **MCP transport names.** The release notes say "SSE transport" and reference "stale pipe transport failures." Scarf's `MCPTransport` enum has `stdio` and `http`; Hermes internally calls those `stdio` and `streamable-http` (or just `http`), and the "pipe" callsite likely refers to internal stdio process pipes — not a third user-facing transport. We're proceeding on that assumption. **Verify:** read `~/.hermes/hermes-agent/hermes_cli/mcp.py` (or equivalent) during integration to confirm `pipe` is internal-only and not a fourth user-selectable transport.

2. **`sse_read_timeout` default value.** The plan uses 300s as the placeholder ("default 300"). Hermes v0.13's `_wait_for_lifecycle_event` keepalive cadence may have a different default — could be 60s, could be 600s. Verify in code; the placeholder copy is the only impact.

3. **`hermes mcp add --transport sse` flag spelling.** The plan assumes `--transport sse` and `--sse-read-timeout <int>`. If Hermes shipped them as `--sse` (boolean) + `--read-timeout`, or merged into `--timeout`, adjust `addMCPServerSSE` accordingly. Test by running `hermes mcp add --help` against a v0.13 install.

4. **Cron `--no-agent` toggle-off shape on edit.** The plan assumes `hermes cron edit <id> --agent` flips the flag back. Possible Hermes ships only `--no-agent` (one-way) and you must `hermes cron remove` + `cron create` without the flag to undo. If so, the edit-mode toggle should be disabled or render a tooltip "Toggling off requires recreating the job." Verify against `hermes cron edit --help`.

5. **Cron `--no-agent` + positional prompt argparse.** The plan passes an empty-string positional after `--no-agent <schedule>` to satisfy argparse. Verify whether Hermes's `cron create` parser tolerates a missing prompt positional when `--no-agent` is set.

6. **Web Tools backend lists.** The plan curates a backend list inline based on the v0.13 release notes mentioning "SearXNG joined search-only." The exact dispatch table (which backends Hermes registers for search vs extract) lives in Hermes source. **Verify** during integration; the Picker contents are the only source of drift, and a wrong entry just produces a `hermes config set` failure on save (recoverable, but ugly).

7. **`web_tools.backend` legacy key on v0.13 hosts.** Hermes v0.13 may *also* honor the legacy `web_tools.backend` key as a fallback when neither split key is set, or may *only* honor it on the rare combined-capability backends. The plan keeps the field readable but only writes the split keys when `hasWebToolsBackendSplit` is true. Verify Hermes' fallback semantics — if `web_tools.backend` is silently ignored on v0.13, a user upgrading from v0.12 with `web_tools.backend: tavily` would suddenly see DuckDuckGo on both capabilities. We may want to add a one-time migration ("we noticed your config has the legacy `web_tools.backend` — promote to split keys?") in a follow-up.

8. **Profile `--no-skills` interaction with `--clone-all`.** Plan disables the `noSkills` toggle when `cloneAll` is on. Verify Hermes's behavior when both flags are passed: argparse may reject as mutually exclusive (good — argparse is the source of truth); may take last-flag-wins; or may produce a profile with everything-but-skills cloned (most useful). The disabled-toggle UX is conservative until we know.

---

## Out of scope

- **MCP per-server SSE auth selection** (Bearer vs OAuth vs none for SSE endpoints). The existing `auth` field on `HermesMCPServer` may or may not carry through to SSE; left untouched. Users can edit the YAML directly via "Open in Editor."
- **Cron `--no-agent` health surface.** A watchdog cron that fails silently (script returns non-zero, no LLM to recover) is a meaningful failure mode but the existing `lastError` rendering covers it. No new health check.
- **Web Tools per-backend config sheets.** SearXNG host URL, Tavily API key, Brave key — all stay in raw YAML for v2.8. The two backend pickers are the v0.13 wire-format change WS-7 ships; the deeper config UI is a follow-up (plausible v2.9).
- **Profiles `--no-skills` post-create surface.** No UI to list a profile's skill count, no "convert to skill-less" verb. Profiles stay create-time-only for skill scoping.
- **iOS surfaces.** All four additions are Mac-only:
  - MCP SSE: Scarf has no iOS MCP servers UI today.
  - Cron `--no-agent`: iOS Cron is read-only (`Scarf iOS/Cron/CronListView.swift`); no editor.
  - Web Tools: iOS Settings doesn't currently surface Web Tools.
  - Profiles `--no-skills`: iOS Profiles is read-only (`Scarf iOS/Profiles/ProfilesView.swift`).
  iOS catch-up is WS-9 territory.
- **Wiki updates.** Per CLAUDE.md, wiki updates land alongside the release once the feature is shipped — not pre-merge. WS-7 PR notes the wiki pages that will need updating in `Scarf-Settings.md`, `Scarf-Cron.md`, `Scarf-MCP-Servers.md`, `Scarf-Profiles.md`, and `Hermes-Version-Compatibility.md`. The wiki PR is its own commit on `gh-pages` after v2.8.0 ships.

---

## Estimate

| Section | LOC est. | Files | Risk |
|---------|----------|-------|------|
| 1. MCP SSE | 250-350 | 6 (model + parser + view × 2 + VM + editor) | Medium — YAML parser change is the riskiest |
| 2. Cron `--no-agent` | 150-200 | 4 (model + view + editor + VM) | Low — mirrors v0.12 `--workdir` pattern |
| 3. Web Tools split | 200-260 | 5 (1 new tab + config model + parser + VM + tabs enum) | Medium — backend lists need verification against Hermes source |
| 4. Profiles `--no-skills` | 30-50 | 2 (view + VM) | Trivial |
| **Total** | **~700-900** | **~17 unique files** | |

**Time estimate (single dev, focused):** 2-3 days of implementation + 1 day of integration verification (the Open Questions section is mostly small empirical checks against a v0.13 Hermes install). Ten files have no overlap between the four additions, so two devs could parallelize after the model-layer work in §1 + §2 + §3 lands.

**Commit shape inside the WS-7 PR (one PR, four commits):**

1. `feat(mcp): add SSE transport support gated on hasMCPSSETransport`
2. `feat(cron): add --no-agent watchdog toggle gated on hasCronNoAgent`
3. `feat(settings): add Web Tools tab with v0.13 search/extract split`
4. `feat(profiles): add --no-skills toggle to create-profile sheet`

Reviewer can scan one commit at a time, and each can be reverted independently if a v0.13 wire-format surprise lands during integration.
