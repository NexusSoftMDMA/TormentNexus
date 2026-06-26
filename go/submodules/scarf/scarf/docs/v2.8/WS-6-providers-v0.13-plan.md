# WS-6 Plan: Provider catalog refresh + Auxiliary `image_gen.model` + OpenRouter response caching

**Workstream:** WS-6 of Scarf v2.8.0
**Hermes target:** v0.13.0 (v2026.5.7)
**Capability gates (already shipped in WS-1):**
- `HermesCapabilities.hasImageGenModel` (`>= 0.13.0`) — `image_gen.model` honored from `config.yaml`.
- `HermesCapabilities.hasOpenRouterResponseCache` (`>= 0.13.0`) — OpenRouter response caching toggle.
**Builds on:** v2.7.5 ModelCatalogService overlay table (11 entries: nous, openai-codex, qwen-oauth, google-gemini-cli, copilot-acp, arcee, gmi, azure-foundry, lmstudio, minimax-oauth, tencent-tokenhub) + the existing AuxiliaryTab pattern (Hermes v0.12 catch-up: `curator` aux row, `flush_memories` row inverse-gated).
**Owner:** TBD
**Reviewers:** Alan; whoever has provider-config bandwidth in the v2.8 cycle.

---

## Goals

The Hermes v0.13.0 release notes list four item-clusters in WS-6 scope:

1. **Provider catalog refresh** — five new model IDs (`deepseek/deepseek-v4-pro`, `x-ai/grok-4.3`, `openrouter/owl-alpha`, `tencent/hy3-preview`, Arcee Trinity Large Thinking) plus a rename (`x-ai/grok-4.20-beta` → `x-ai/grok-4.20`). All five new IDs already appear in `models_dev_cache.json` on the local v0.13 dev host (verified: see Appendix A), so the catalog file does the heavy lifting on next `models.dev` cache refresh — Scarf just needs alias-resolution + (sparingly) curated metadata.
2. **Vercel AI Gateway demotion** — Hermes deprioritizes the `vercel` provider (display name `Vercel AI Gateway`) in the picker. Currently Scarf sorts providers `subscriptionGated → alphabetical`; Vercel sits mid-alphabet. We add a `demoted` axis so Vercel sinks to the bottom while keeping all other providers in their alphabetic positions.
3. **`image_gen.model` from `config.yaml`** — Hermes v0.13 honors a top-level `image_gen.model` key. Scarf surfaces a model picker for it on the Auxiliary tab, capability-gated on `hasImageGenModel`.
4. **OpenRouter response caching toggle** — Hermes v0.13 added an OpenRouter response-caching switch in `config.yaml`. Scarf surfaces a `Toggle` next to OpenRouter's other knobs, capability-gated on `hasOpenRouterResponseCache`. **Open Question** on the exact key shape (`openrouter.response_cache.enabled` vs `providers.openrouter.response_cache_enabled` vs nested under `prompt_caching`) — flagged below.

The two release-notes items NOT in WS-6 scope:

- **"Honor runtime default model during delegate provider resolution"** — server-side resolution behavior. Scarf's existing `delegation.model` / `delegation.provider` fields in `DelegationSettings` are unchanged; the picker continues to fill those values straight to `config.yaml`. No Scarf surface change needed. Document in the `Out of scope` section as verified-no-change.
- **"Avoid Bedrock credential probe in provider picker"** — server-side: the `hermes model` CLI no longer probes AWS_ACCESS_KEY_ID at picker open time. Scarf's `ModelPickerSheet` was already not invoking that probe (we read the cached catalog, not `hermes model`). No change needed.
- **`ProviderProfile` ABC + `plugins/model-providers/` + `list_picker_providers`** — these are Hermes-internal pluggability scaffolding. They expand which providers can ship via plugin, but none alter the on-disk shape of `models_dev_cache.json` or the `HERMES_OVERLAYS` table. Scarf's existing read path (cache file + overlay table) reaches them transparently. **Caveat:** the `list_picker_providers` change adds a credential-filter so providers without the right env vars are hidden. Scarf's picker today shows everything regardless of credentials. We choose to **not adopt** the credential filter in the picker (users frequently configure providers in-app and need to see the row before they can fill the secret). Documented in the `Out of scope` section.
- **Shared Hermes dotenv loader / Nous OAuth persistence across profiles** — entirely server-side. Scarf's `NousSubscriptionService` reads `~/.hermes/auth.json`; the new shared dotenv loader doesn't change that file's path or shape. No Scarf surface change.
- **`/provider` alias removal** — server-side CLI cleanup. Scarf already invokes `/model` directly via ACP slash command routing; no Scarf surface used `/provider`. No change.

### Non-goals (explicitly deferred)

- **In-app credential entry sheet** for providers requiring an API key. v2.7.5 surfaces "Set in Terminal: `hermes auth <provider>`" as the path for OAuth providers; for new BYO-key providers (none in this WS — the five new models all flow through OpenRouter / Nous Portal / Arcee already-credentialed) we keep the same convention.
- **Per-model image-gen capability tag** in the catalog. The `models_dev_cache.json` schema doesn't include an `image: true` field today. Filtering the `image_gen.model` picker to "image-capable models only" is therefore not feasible at the catalog level. We pre-populate a small allowlist of well-known image models in Scarf instead (see §New types / fields).
- **iOS surface for new image_gen / openrouter toggles.** ScarfGo's settings is read-mostly; a dedicated iOS tab is deferred to WS-9 (iOS catch-up). The capability flags will work on iOS too once the surface lands.
- **Migration ceremony for the Grok rename.** We resolve the alias at read time (option 1) — no ceremony, no race, lossless. See §Migration.
- **A standalone "Image Gen" Settings tab.** v0.13 has exactly two image-gen-related fields (the model + the existing `image_gen.provider` from v0.12). That's not enough surface to warrant a tab — they belong next to the `vision` row in Auxiliary. If v0.14 adds size/quality/style fields, we revisit and split into its own tab then.

---

## Files to change

The plan is intentionally minimal-touch. The `models_dev_cache.json` refresh handles four of the five new model IDs without any Swift change; the rename + the one new aux field + the toggle are surgical.

### 1. `scarf/Packages/ScarfCore/Sources/ScarfCore/Services/ModelCatalogService.swift`

**Why:** Two changes:
- Adds an alias-resolution path so `x-ai/grok-4.20-beta` keeps working when a user's `config.yaml` references the old name. Lossless, opt-in, zero migration risk.
- Adds a `demoted` axis to provider sort so Vercel AI Gateway sinks to the bottom of the picker.

**Edits:**

- **Alias map.** Add a static table near `overlayOnlyProviders`:
  ```swift
  /// Hermes deprecates model IDs across releases. When a stored config
  /// `model.default` references a deprecated ID, resolve to its
  /// canonical successor. Lossless — we never rewrite the user's
  /// config.yaml; the alias just lets `validateModel` /
  /// `model(providerID:modelID:)` succeed against the new ID.
  ///
  /// Keys are dot-separated `providerID/modelID` to disambiguate
  /// across providers — even if `vercel` later adds a `grok-4.20-beta`
  /// alias on its own, the openrouter resolution shouldn't fire.
  ///
  /// **Schema is Swift-primary.** Mirror new entries into
  /// `tools/build-catalog.py` only if the catalog tool grows a model-ID
  /// validation pass (it doesn't today — see §`tools/build-catalog.py`
  /// mirror below).
  public static let modelAliases: [String: String] = [
      // v0.13: x-ai dropped the `-beta` suffix once Grok 4.20 GA'd.
      // The model is the same one served at the same OpenRouter slot;
      // only the marketing identifier changed.
      "openrouter/x-ai/grok-4.20-beta": "openrouter/x-ai/grok-4.20",
      "xai/grok-4.20-beta": "xai/grok-4.20",
      "vercel/xai/grok-4.20-beta": "vercel/xai/grok-4.20",
  ]

  /// Resolve a stored model identifier through the alias map. Returns
  /// the input unchanged when no alias exists. Pure function — used at
  /// read time everywhere a config'd model ID is rendered, validated,
  /// or sent to Hermes.
  public func resolveModelAlias(providerID: String, modelID: String) -> String {
      let composite = "\(providerID)/\(modelID)"
      return Self.modelAliases[composite].map { resolved -> String in
          // Strip the providerID prefix from the resolved value before
          // returning — callers want the bare model ID.
          let prefix = providerID + "/"
          return resolved.hasPrefix(prefix)
              ? String(resolved.dropFirst(prefix.count))
              : resolved
      } ?? modelID
  }
  ```
  Call sites that need to resolve: `validateModel(_:for:)` resolves the input before lookup; `model(providerID:modelID:)` resolves before `provider.models?[modelID]` indexing; `provider(for:)` resolves the input model ID before scanning. Each is a one-line addition at the top of the function.

- **Demoted-provider axis.** Add a static set:
  ```swift
  /// Provider IDs that Hermes v0.13 explicitly deprioritizes in the
  /// picker. `loadProviders()` sorts these to the tail of the list,
  /// after the alphabetical group, so users who haven't manually
  /// chosen Vercel as their gateway don't end up there by default.
  /// Mirrors Hermes's `DEMOTED_PROVIDERS` list in
  /// `hermes_cli/providers.py`.
  public static let demotedProviders: Set<String> = [
      "vercel",
  ]
  ```
  Update the sort comparator in `loadProviders()`:
  ```swift
  return byID.values.sorted { lhs, rhs in
      // Subscription-gated first (Nous Portal).
      if lhs.subscriptionGated != rhs.subscriptionGated {
          return lhs.subscriptionGated
      }
      // Demoted last (Vercel AI Gateway).
      let lDemoted = Self.demotedProviders.contains(lhs.providerID)
      let rDemoted = Self.demotedProviders.contains(rhs.providerID)
      if lDemoted != rDemoted {
          return !lDemoted
      }
      return lhs.providerName.localizedCaseInsensitiveCompare(rhs.providerName) == .orderedAscending
  }
  ```

- **Image-gen model allowlist.** Add a static curated list of well-known image-gen-capable model IDs (kept short and updated by hand; the catalog file has no `image_capable` flag today):
  ```swift
  /// Known image-generation models, used to pre-populate the
  /// `image_gen.model` picker on the Auxiliary tab. The list is
  /// curated — `models_dev_cache.json` doesn't tag image-capable
  /// models, so we maintain this by hand on Hermes version bumps.
  /// Always free-form-typeable on the picker too, so missing entries
  /// don't block users with non-listed image providers.
  ///
  /// Order: most-likely-to-be-chosen first.
  public static let imageGenModels: [HermesImageGenModel] = [
      .init(modelID: "openai/gpt-image-1", display: "OpenAI · gpt-image-1", providerHint: "openai"),
      .init(modelID: "google/imagen-4", display: "Google · Imagen 4", providerHint: "google-vertex"),
      .init(modelID: "google/imagen-3", display: "Google · Imagen 3", providerHint: "google-vertex"),
      .init(modelID: "stability/stable-image-ultra", display: "Stability · Stable Image Ultra", providerHint: "stability"),
      .init(modelID: "fal-ai/flux-pro-1.1", display: "fal · FLUX 1.1 Pro", providerHint: "fal"),
      .init(modelID: "black-forest-labs/flux-1.1-pro", display: "Black Forest Labs · FLUX 1.1 Pro", providerHint: "openrouter"),
      .init(modelID: "openai/dall-e-3", display: "OpenAI · DALL·E 3", providerHint: "openai"),
  ]

  public struct HermesImageGenModel: Sendable, Identifiable, Hashable {
      public let modelID: String
      public let display: String
      /// Hint at which provider serves this model — surfaced as a
      /// "Configure provider X first" advisory but never enforced.
      public let providerHint: String?
      public var id: String { modelID }
  }
  ```

**Tolerance contract:** When a user has a config with `model.default: x-ai/grok-4.20-beta` and provider `openrouter`, `validateModel("x-ai/grok-4.20-beta", for: "openrouter")` resolves to `"x-ai/grok-4.20"` and validates against the catalog. If the alias isn't present in the map, the function behaves identically to today.

---

### 2. `scarf/Packages/ScarfCore/Sources/ScarfCore/Models/HermesConfig.swift`

**Why:** Add two new top-level config fields:
- `imageGenModel: String` — `image_gen.model` value, default `""` (empty means "use provider default").
- `openrouterResponseCacheEnabled: Bool` — `openrouter.response_cache.enabled` (working name pending Open Question §1), default `false`.

**Edits:**

- Add stored properties next to `cacheTTL` / `redactionEnabled` / `runtimeMetadataFooter`:
  ```swift
  /// `image_gen.model` (v0.13+) — overrides the per-provider default
  /// image-gen model. Empty string means "let Hermes pick the
  /// provider default". Hermes v0.12 advertised this key but ignored
  /// it; Scarf's `AuxiliaryTab` only renders the picker when
  /// `HermesCapabilities.hasImageGenModel` is `true`.
  public var imageGenModel: String

  /// `openrouter.response_cache.enabled` (v0.13+) — when true, Hermes
  /// asks OpenRouter to cache responses for repeat prompts within a
  /// session. **Open Question:** the exact YAML key shape is
  /// unconfirmed. See WS-6 plan §Open Questions #1.
  public var openrouterResponseCacheEnabled: Bool
  ```
- Append `imageGenModel: String = ""` and `openrouterResponseCacheEnabled: Bool = false` to the trailing parameter list in the explicit memberwise `init` (after `runtimeMetadataFooter`). Default values mean every existing call site (`HermesConfig.empty`, `init(yaml:)`) compiles unchanged until updated.
- Update the static `HermesConfig.empty` factory to pass both new defaults explicitly so the empty-config sentinel matches the post-load shape.

**Tolerance contract:** Pre-v0.13 hosts have neither key in `config.yaml`; the parser defaults both to empty / false. UI is gated separately on the capability flag, so the values never reach the screen on pre-v0.13 hosts even if they were somehow non-default.

---

### 3. `scarf/Packages/ScarfCore/Sources/ScarfCore/Parsing/HermesConfig+YAML.swift`

**Why:** Wire the two new keys into the YAML parser.

**Edits:**

- In the trailing `self.init(...)` call, add (next to `cacheTTL` / `redactionEnabled` / `runtimeMetadataFooter`):
  ```swift
  imageGenModel: str("image_gen.model", default: ""),
  openrouterResponseCacheEnabled: bool("openrouter.response_cache.enabled", default: false),
  ```
- The exact key for `openrouter.response_cache.enabled` is **provisional** — see §Open Questions #1. Lock the key only after manual verification on a v0.13 host (`hermes config check` against a sample YAML with the candidate key + a printout of the `Settings`-level key). We may need a fallback: read the legacy key first and fall through to the canonical one, exactly like the `slack.reply_to_mode` ↔ `platforms.slack.reply_to_mode` pattern at line 187.

**Tolerance contract:** A v0.12 host with neither key produces `imageGenModel == ""` and `openrouterResponseCacheEnabled == false`, matching the runtime defaults. A v0.13 host with both keys present round-trips through `init(yaml:)` cleanly.

---

### 4. `scarf/scarf/Features/Settings/ViewModels/SettingsViewModel.swift`

**Why:** Two new setters, one for each new field.

**Edits:**

- Add to the "Auxiliary model sub-tasks" section (since `image_gen` lives logically next to other aux tasks even though the YAML key is at the top level):
  ```swift
  // MARK: - Image generation (v0.13+)

  func setImageGenModel(_ value: String) { setSetting("image_gen.model", value: value) }
  func setOpenRouterResponseCache(_ value: Bool) {
      setSetting("openrouter.response_cache.enabled", value: value ? "true" : "false")
  }
  ```
- Both setters route through `setSetting` → `runHermes(["config", "set", key, value])`, matching the existing pattern. `hermes config set` is forward-compatible — pre-v0.13 hosts accept any key without complaint and write it to YAML; the gate keeps the UI hidden so users on pre-v0.13 never invoke these.

**Tolerance contract:** No new error paths. Existing `setSetting`'s `saveMessage` plumbing handles success/failure surfacing.

---

### 5. `scarf/scarf/Features/Settings/Views/Tabs/AuxiliaryTab.swift`

**Why:** Surface the two new fields. Both belong on Auxiliary because they're per-task / per-provider knobs, not main-model-pickers.

**Edits:**

- **Image-gen model row.** Add a new `SettingsSection(title: "Image Generation", icon: "photo")` between the static base tasks and `unknownTasks`, gated on `capabilitiesStore?.capabilities.hasImageGenModel == true`:
  ```swift
  if capabilitiesStore?.capabilities.hasImageGenModel ?? false {
      SettingsSection(title: "Image Generation", icon: "photo") {
          imageGenRow
      }
  }
  ```
  `imageGenRow` is a small `@ViewBuilder`:
  ```swift
  @ViewBuilder
  private var imageGenRow: some View {
      let value = viewModel.config.imageGenModel
      Picker("Model", selection: Binding(
          get: { value },
          set: { viewModel.setImageGenModel($0) }
      )) {
          Text("Provider default").tag("")
          Divider()
          ForEach(ModelCatalogService.imageGenModels) { model in
              Text(model.display).tag(model.modelID)
          }
          if !value.isEmpty
              && !ModelCatalogService.imageGenModels.contains(where: { $0.modelID == value }) {
              // User has set a custom value; preserve it as a tagged option
              // so the picker renders the actual selection, not "Provider default".
              Divider()
              Text(value + "  (custom)").tag(value)
          }
      }
      .pickerStyle(.menu)
      EditableTextField(label: "Custom model ID", value: value) { newValue in
          viewModel.setImageGenModel(newValue.trimmingCharacters(in: .whitespaces))
      }
      Text("Used for image generation calls. Leave as Provider default unless your provider documents a specific model ID for image-gen.")
          .font(.caption2)
          .foregroundStyle(.tertiary)
  }
  ```
  The `EditableTextField` lets users free-form-type a model ID we haven't curated. Together they cover both the curated allowlist + the long tail.

- **OpenRouter response cache row.** Add a new section (or fold into a future "Providers" section):
  ```swift
  if capabilitiesStore?.capabilities.hasOpenRouterResponseCache ?? false {
      SettingsSection(title: "OpenRouter", icon: "shippingbox") {
          ToggleRow(label: "Response caching",
                    isOn: viewModel.config.openrouterResponseCacheEnabled) { newValue in
              viewModel.setOpenRouterResponseCache(newValue)
          }
          Text("OpenRouter caches identical prompts within a session to reduce token costs. Off by default — enable when your workload has highly repeated prompts.")
              .font(.caption2)
              .foregroundStyle(.tertiary)
              .padding(.horizontal, 12)
              .padding(.bottom, 4)
      }
  }
  ```

**Tolerance contract:** Pre-v0.13 host hides both sections entirely. Capability flag false → guard fails → section never enters the view tree. Dynamic Type clamp on iOS (n/a here, this is Mac-only) preserved on captions.

---

### 6. `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/M0cServicesTests.swift`

**Why:** The existing model-catalog tests freeze the `loadProviders()` sort order + decoding shape. Add three new tests:

**New tests (Swift Testing macros):**

```swift
@Test func vercelAIGatewayDemotedToBottom() throws {
    // Build a minimal catalog with vercel + alphabetically-later providers,
    // then assert vercel sorts after them.
    let json = """
    {
      "anthropic": { "name": "Anthropic", "models": {} },
      "vercel":    { "name": "Vercel AI Gateway", "models": {} },
      "zonk":      { "name": "Zonk Provider", "models": {} }
    }
    """
    let tmp = FileManager.default.temporaryDirectory
        .appendingPathComponent("scarf-models-\(UUID().uuidString).json")
    try json.write(to: tmp, atomically: true, encoding: .utf8)
    defer { try? FileManager.default.removeItem(at: tmp) }
    let svc = ModelCatalogService(path: tmp.path)
    let providers = svc.loadProviders().filter { !$0.isOverlay }
    let names = providers.map(\.providerName)
    // anthropic first (alpha), zonk next (alpha), vercel last (demoted).
    #expect(names.last == "Vercel AI Gateway")
    #expect(names.firstIndex(of: "Vercel AI Gateway")! > names.firstIndex(of: "Zonk Provider")!)
}

@Test func grok420BetaAliasResolvesToGrok420() {
    let svc = ModelCatalogService(path: "/tmp/scarf-nonexistent-\(UUID().uuidString).json")
    #expect(svc.resolveModelAlias(providerID: "openrouter", modelID: "x-ai/grok-4.20-beta")
            == "x-ai/grok-4.20")
    #expect(svc.resolveModelAlias(providerID: "xai", modelID: "grok-4.20-beta")
            == "grok-4.20")
    // Non-aliased ID passes through unchanged.
    #expect(svc.resolveModelAlias(providerID: "anthropic", modelID: "claude-4.7-opus")
            == "claude-4.7-opus")
    // Cross-provider isolation: same modelID on a different provider isn't aliased.
    #expect(svc.resolveModelAlias(providerID: "fictional", modelID: "x-ai/grok-4.20-beta")
            == "x-ai/grok-4.20-beta")
}

@Test func imageGenModelAllowlistShape() {
    // Lock the curated list size + a few sentinel entries so unintentional
    // edits get caught in review.
    let models = ModelCatalogService.imageGenModels
    #expect(models.count >= 5)
    #expect(models.contains(where: { $0.modelID == "openai/gpt-image-1" }))
    #expect(models.contains(where: { $0.modelID == "google/imagen-4" }))
    // Every entry has a non-empty display + a non-empty modelID.
    for m in models {
        #expect(!m.modelID.isEmpty)
        #expect(!m.display.isEmpty)
    }
}
```

**Tolerance contract:** All three are pure-function tests that run without a Hermes binary or models cache file. They survive a `ModelCatalogService(path: nonexistent)` because the alias + allowlist paths don't read the catalog.

---

### 7. `scarf/Packages/ScarfCore/Tests/ScarfCoreTests/M6ConfigCronTests.swift` (or new `WS6ProvidersConfigTests.swift`)

**Why:** Lock the YAML round-trip for the two new keys.

**New test:**

```swift
@Test func imageGenAndOpenRouterCacheRoundTrip() {
    let yaml = """
    image_gen:
      model: openai/gpt-image-1
    openrouter:
      response_cache:
        enabled: true
    """
    let cfg = HermesConfig(yaml: yaml)
    #expect(cfg.imageGenModel == "openai/gpt-image-1")
    #expect(cfg.openrouterResponseCacheEnabled == true)
}

@Test func imageGenDefaultsToEmptyString() {
    let cfg = HermesConfig(yaml: "")
    #expect(cfg.imageGenModel == "")
    #expect(cfg.openrouterResponseCacheEnabled == false)
}
```

**Tolerance contract:** Tracks the exact YAML keys the parser expects. If the Open Question resolves a different key shape, this test pins the change to one place.

---

### 8. `tools/build-catalog.py` mirror

**Why:** Per CLAUDE.md, every new schema-shaped change must mirror into the Python validator. Audit:

| New surface | Mirror needed? | Rationale |
| -- | -- | -- |
| `modelAliases` | **No** | The catalog tool validates `template.json` manifests, not model IDs. Aliases live entirely in Scarf-side ModelCatalogService. |
| `demotedProviders` | **No** | Same — the catalog tool doesn't render the picker. |
| `imageGenModels` (curated) | **No** | Curated list is Scarf UI-only. |
| `HermesConfig.imageGenModel` | **No** | The catalog tool never reads `config.yaml`; it reads `template.json`. |
| `HermesConfig.openrouterResponseCacheEnabled` | **No** | Same. |

**Verdict:** No `tools/build-catalog.py` changes for WS-6. Document the audit explicitly in the WS-6 PR description so future plans know we checked.

If WS-6 ever adds a new `ProjectDashboardWidget.type` (it doesn't — image_gen is in Settings, not a dashboard widget), the mirror would be required. The widget vocabulary is the only Swift-primary schema the catalog tool tracks.

---

### 9. `scarf/CLAUDE.md` — schema-drift line

**Why:** CLAUDE.md says "Keep `ModelCatalogService.overlayOnlyProviders` in sync with `HERMES_OVERLAYS` in … `providers.py`." After this WS, Scarf also needs to keep `modelAliases` in sync with Hermes's deprecation map (currently a small list inside `hermes_cli/providers.py`). Add one bullet in the "Hermes Version" section:

> Keep `ModelCatalogService.modelAliases` in sync with `HERMES_DEPRECATED_MODEL_IDS` (or whatever the upstream module renames to) in `hermes-agent/hermes_cli/providers.py`. Drift here means a user's old model ID stops resolving in the picker even though Hermes still accepts it at runtime.

(Plus the existing demoted-providers bullet — see below.)

> Keep `ModelCatalogService.demotedProviders` in sync with the deprioritized-provider list in `hermes-agent/hermes_cli/providers.py`. Drift means Vercel AI Gateway sorts in the wrong position in Scarf's picker.

**Touchpoint:** the single block at line ~205 of `scarf/CLAUDE.md` (the "Keep `ModelCatalogService.overlayOnlyProviders` in sync" paragraph). Append two more bullets next to it.

---

## New models / overlay entries

| Model ID | Provider | Cache hit (verified) | Overlay change? | Action |
| -- | -- | -- | -- | -- |
| `deepseek/deepseek-v4-pro` | OpenRouter + Nous Portal | **Yes** (openrouter) | No | Auto-shows on next `models_dev_cache.json` refresh; Nous Portal serves it via the Nous overlay's free-form model list. No code change. |
| `x-ai/grok-4.3` | OpenRouter + Nous Portal + xAI direct + Vercel | **Yes** (openrouter, xai, vercel) | No | Auto-shows. No code change. |
| `openrouter/owl-alpha` | OpenRouter only (free tier) | **Yes** | No | Auto-shows. No code change. |
| `tencent/hy3-preview` | OpenRouter only (paid route) | **Yes** | No | Auto-shows. No code change. |
| `arcee-ai/trinity-large-thinking` | Arcee (overlay) + OpenRouter + DigitalOcean + Venice + Kilo | **Yes** (openrouter, etc.) | No | Auto-shows on non-overlay providers. The Arcee overlay's free-form picker remains the path for direct Arcee API users. **No catalog field captures the v0.13 "temperature + compression overrides" — that's a per-call hint Hermes passes through, not a per-model metadata field.** Scarf doesn't need to surface it. |
| `x-ai/grok-4.20-beta` → `x-ai/grok-4.20` | OpenRouter + xAI + Vercel | **Both present** | No | Add to `modelAliases` (see file 1). Resolution at read time means a user's stored config keeps working without a rewrite. |

**Why no overlay changes:** All 11 existing overlay entries (`nous`, `openai-codex`, `qwen-oauth`, `google-gemini-cli`, `copilot-acp`, `arcee`, `gmi`, `azure-foundry`, `lmstudio`, `minimax-oauth`, `tencent-tokenhub`) remain. v0.13's `ProviderProfile` ABC + `plugins/model-providers/` framework adds **internal** Hermes pluggability but does not introduce new overlay-only providers in this release. Verify on Hermes upstream by diffing `hermes_cli/providers.py` against the v0.12 baseline; if the `HERMES_OVERLAYS` dict gained entries, mirror them. Lock in `ToolGatewayTests.v013OverlayProvidersCarryCorrectAuthTypes` (mirror of the existing v0.12 lock-in test).

---

## New types / fields

### `HermesProviderOverlay` — no shape change

The release notes mention `ProviderProfile` ABC, but it's an internal Python abstraction. Nothing in the on-disk overlay contract changes. `HermesProviderOverlay` keeps its current five-field shape (`displayName`, `baseURL`, `authType`, `subscriptionGated`, `docURL`).

### `ModelCatalogService.HermesImageGenModel` — new

Curated image-gen model entry, pre-populated for the picker on Auxiliary tab. Five fields: `modelID`, `display`, `providerHint`. Scope is intentionally tiny — we don't enumerate every provider's image model; users with niche providers free-form-type the model ID instead.

### `ModelCatalogService.modelAliases` — new

`[String: String]` map keyed by composite `providerID/modelID`. Used at read time by `validateModel`, `model(_:_:)`, and `provider(for:)`. **Does not** rewrite stored config.

### `ModelCatalogService.demotedProviders` — new

`Set<String>` of provider IDs to sink to the bottom of the picker. Sort comparator update in `loadProviders()` is the only consumer.

### `HermesConfig.imageGenModel` / `HermesConfig.openrouterResponseCacheEnabled` — new

Top-level config fields, defaults `""` and `false`. Read by `init(yaml:)`, written via `setSetting` → `hermes config set`.

---

## Capability gating

| Capability | Flag | UI surface | Pre-v0.13 host behavior |
| -- | -- | -- | -- |
| `image_gen.model` honored at runtime | `hasImageGenModel` | `AuxiliaryTab` "Image Generation" section | Section never enters the view tree. The model picker would otherwise no-op silently on pre-v0.13 (the value goes to YAML but Hermes ignores it). Hiding spares users a "I set this and nothing happened" trap. |
| OpenRouter response caching | `hasOpenRouterResponseCache` | `AuxiliaryTab` "OpenRouter" section | Section never enters the view tree. Same reasoning — silent no-op on pre-v0.13. |
| `modelAliases` resolution | (none) | `validateModel`, `model(_:_:)`, `provider(for:)` | Always on. The alias is a Scarf-side concept that doesn't depend on Hermes version — even on pre-v0.13 hosts, OpenRouter still serves the model via either the old or new ID. (Verify upstream: if OpenRouter has dropped the `-beta` slot entirely, the alias resolution still helps users on the new ID. If OpenRouter kept the `-beta` slot live, the alias still helps users on the new ID. Win-win.) |
| Vercel demotion | (none) | `loadProviders()` sort | Always on. Vercel's display position is a Scarf-UI choice, not a Hermes-version-gated behavior. |

**Why no flag for the demotion / aliases:** Both are Scarf-UX choices that improve every Hermes version's experience equally. Adding a flag would mean dragging the sort order with the version, which is worse — users on a v0.12 host would see Vercel mid-alphabet, then mysteriously at the bottom after upgrading. Consistency wins.

---

## How to test

### Unit tests (Swift Testing — see file 6 + 7)

- `vercelAIGatewayDemotedToBottom` — locks the new sort axis.
- `grok420BetaAliasResolvesToGrok420` — locks the alias map shape.
- `imageGenModelAllowlistShape` — locks the curated list size + sentinel entries.
- `imageGenAndOpenRouterCacheRoundTrip` — locks the YAML key shape (`image_gen.model` + `openrouter.response_cache.enabled`).
- `imageGenDefaultsToEmptyString` — locks the empty-config default.

### Manual test plan (Mac, against a v0.13 Hermes host)

1. **Picker order.** Open `Settings → General → Model picker`. Confirm Nous Portal (subscription-gated) is at the top, alphabetical group fills the middle, Vercel AI Gateway is the last non-subscription entry. Resize the sheet; the order is stable across re-renders.
2. **Grok rename.** Edit `~/.hermes/config.yaml` directly: set `model.default: x-ai/grok-4.20-beta`, provider `openrouter`. Reload Scarf. The picker should show `x-ai/grok-4.20` selected (the alias resolved). The stored YAML is untouched. Save a new model — confirm Hermes still accepts `x-ai/grok-4.20-beta` at the wire level (it should — OpenRouter keeps the slot live).
3. **Image-gen model picker.** Open `Settings → Auxiliary → Image Generation`. Confirm:
   - Section is visible (you're on v0.13).
   - The picker has "Provider default" + the 7 curated entries.
   - Selecting `openai/gpt-image-1` writes `image_gen.model: openai/gpt-image-1` to `config.yaml` (verify with `grep image_gen ~/.hermes/config.yaml`).
   - Free-form-typing a custom value sets it.
   - Setting it back to "Provider default" (`""`) clears the key from YAML on next save.
4. **OpenRouter response cache toggle.** Same tab, "OpenRouter" section. Confirm:
   - Section is visible.
   - Toggle off → on writes `openrouter.response_cache.enabled: true`.
   - Toggle on → off writes `openrouter.response_cache.enabled: false`.
5. **Pre-v0.13 fallback.** Switch the active server to a v0.12 host (or stash with `HERMES_VERSION_OVERRIDE=0.12.0` env shim). Confirm:
   - Image Generation section is hidden.
   - OpenRouter section is hidden.
   - The picker still shows Vercel AI Gateway at the bottom (sort axis is unconditional).
   - Grok alias resolution still works.
6. **`hermes config set` round-trip.** Set `image_gen.model` from Scarf, then `hermes config check` from Terminal — confirm the new key validates against Hermes's schema.

### Integration / smoke

- `scripts/smoke.sh` (if present) — run the full smoke sweep, verify no provider catalog regressions on the existing 11 overlay entries.
- Build clean: `xcodebuild -project scarf/scarf.xcodeproj -scheme scarf -configuration Debug build`. New Swift Testing tests run via `swift test --package-path scarf/Packages/ScarfCore`.

---

## Open questions

1. **`openrouter.response_cache.enabled` — exact YAML key shape.** The release notes say "OpenRouter response caching support" but don't specify the key. Three plausible shapes:
   - `openrouter.response_cache.enabled: true` (top-level provider block)
   - `providers.openrouter.response_cache_enabled: true` (under the new `providers:` map v0.13 introduces)
   - `prompt_caching.openrouter.enabled: true` (nested under the existing `prompt_caching` block from v0.12)

   **Recommendation:** Verify by inspecting the v0.13 Hermes config schema (`hermes config check` against a sample YAML for each shape, or `grep -r response_cache hermes-agent/hermes_cli/`) before merging WS-6. The first shape is consistent with how Hermes handles other per-provider knobs (`xai.voice_cloning.enabled` from v0.13's xAI Voice Cloning); it's our default until verified. If the shape changes, file 3's parser line + file 4's setter key + file 7's test fixture all update in lockstep.

2. **Default value for OpenRouter response caching.** The release notes don't specify whether v0.13 defaults the toggle on or off. **Recommendation:** Default off in Scarf's parser (`bool("openrouter.response_cache.enabled", default: false)`). Worst case, the user explicitly opts in. If Hermes defaults on server-side, our `false` parse still matches because the key would be present in the YAML.

3. **Arcee Trinity Large Thinking "temperature + compression overrides".** The release notes mention "temperature + compression overrides" for this model. Hermes treats these as per-model invocation hints (not catalog metadata). Scarf has no surface for per-model temperature today — it's set by the user via `hermes ask --temperature` or the per-aux-task config. **Recommendation:** Defer to a future cycle if user feedback asks for per-model temperature picker. v2.8 ships without.

4. **Grok rename — does OpenRouter delete the old slot?** If OpenRouter keeps `x-ai/grok-4.20-beta` live (with a redirect to `x-ai/grok-4.20`), our alias is purely cosmetic — Hermes still accepts the old ID. If OpenRouter deletes the old slot, the alias becomes load-bearing — without it, users on the old config get a 404 at runtime. **Either way, the alias is correct.** Verify before merging by sending a request to OpenRouter for both IDs.

5. **`models_dev_cache.json` refresh timing.** Hermes ships with a snapshot; the user's local cache refreshes via Hermes's own cache-refresh logic (background task or on-demand). Confirm that a v0.13 install ships with all five new models pre-populated (not deferred to a first-run network fetch), so the picker doesn't render an empty list on a fresh `~/.hermes/`. **Verified locally:** the dev host's cache has all five new IDs. Re-verify on a clean `~/.hermes/` after `hermes update` to v0.13.

---

## Out of scope (deferred)

- **In-app Hermes restart** after toggling response caching. Some toggles need a Hermes restart to take effect; the response_cache toggle is unclear. Defer the auto-restart prompt to a future cycle once we know which toggles need it. Scarf already has a "Restart Hermes" button at `Settings → General` for users who hit a stale-toggle case.
- **iOS surface for image_gen.model + OpenRouter cache.** ScarfGo's settings is read-mostly. WS-9 picks up iOS catch-up; the capability flags work cross-platform once the surface lands.
- **Per-image-gen-model metadata** (cost, max resolution, prompt-token-cost). Not in `models_dev_cache.json`; out of scope until the catalog adds a tag.
- **Provider profile MCP plugins (`plugins/model-providers/`).** Server-side framework. Scarf reaches whatever providers Hermes exposes via the cache + overlay — the indirection is transparent.
- **Bedrock credential probe avoidance.** Server-side; Scarf was already not invoking that probe.
- **Honor runtime default model during delegate provider resolution.** Server-side; Scarf's `delegation.model` field is already a free-form string we hand to `hermes config set`.
- **`/provider` alias removal.** Server-side; Scarf already used `/model` directly.
- **Credential filter on picker provider list.** v0.13's `list_picker_providers` filters the CLI picker by available credentials. We deliberately don't adopt this in Scarf — users frequently configure providers in-app and need to see the row before they can fill the secret. If user feedback strongly favors hiding unconfigured providers, revisit in a future WS.
- **Migration to one-shot rewrite for the Grok alias.** Option 2 (rewrite YAML) was rejected; option 1 (read-time alias) wins on safety + simplicity. See §Migration.

---

## Migration

### Grok 4.20-beta → 4.20

**Option 1 — alias-resolve at read time. ✅ Recommended.**

- `ModelCatalogService.modelAliases` maps `openrouter/x-ai/grok-4.20-beta` → `openrouter/x-ai/grok-4.20`.
- `validateModel` resolves the alias before lookup; `model(_:_:)` resolves before indexing; `provider(for:)` resolves before scanning.
- The user's `config.yaml` stays as-is. Scarf treats the alias as an internal display + lookup detail; Hermes (which still accepts both IDs at runtime) handles the wire.

**Pros:**
- Lossless. The user's hand-edits to `config.yaml` are sacred — we never touch them.
- No race. There's no point at which Scarf's "rewrite YAML" path could conflict with the user's editor.
- Trivial to reverse. If a future Hermes brings the old ID back, drop the entry from `modelAliases`.
- Free of edge cases. A user with a custom `model.default` value Hermes never recognized still works.

**Cons:**
- Two IDs in flight on the user's system (one in `config.yaml`, one in the picker's selected state). Cosmetic — the picker shows the resolved name, the YAML keeps the old name.

**Option 2 — one-shot YAML rewrite on next launch.**

Rejected. TOCTOU race (user edits YAML in `vim`, Scarf rewrites mid-edit), no path to undo, and the only "win" (a clean YAML) is invisible to most users.

**Precedent:** No prior model-rename has shipped through Scarf's overlay table. The new alias map is the precedent for this and future renames.

---

## Estimate

- File 1 (`ModelCatalogService.swift`): ~80 lines net add (alias map + helper + curated list + sort axis update).
- File 2 (`HermesConfig.swift`): ~25 lines net add (two stored props + memberwise init params + empty-config update).
- File 3 (`HermesConfig+YAML.swift`): ~5 lines net add (two parser lines).
- File 4 (`SettingsViewModel.swift`): ~5 lines net add (two setters).
- File 5 (`AuxiliaryTab.swift`): ~70 lines net add (two new sections + the image-gen view).
- File 6 (`M0cServicesTests.swift`): ~60 lines net add (three tests).
- File 7 (`M6ConfigCronTests.swift` or new file): ~30 lines net add (two tests).
- File 9 (`scarf/CLAUDE.md`): ~6 lines net add (two new bullets in the schema-drift block).

**Total:** ~280 lines net add across 8 files (Swift + Markdown). No deletes. No file moves. No new package targets.

**Build risk:** Low. All edits are additive; existing call sites use default values. No behavior change for pre-v0.13 hosts (capability flag + alias resolution are both safe).

**Review risk:** Medium-low. The Open Question on the OpenRouter cache key shape is the single highest-risk item; everything else is mechanical. Block the PR until that key is verified.

**Effort:** ~1 day implementation + 0.5 day verification (manual test plan + Open Question verification on a real v0.13 host).

---

## Appendix A — `models_dev_cache.json` verification

Local `~/.hermes/models_dev_cache.json` (v0.13 dev host) confirms:

| Query | Provider | Match |
| -- | -- | -- |
| `deepseek-v4-pro` | openrouter | `deepseek/deepseek-v4-pro` ✅ |
| `grok-4.3` | openrouter, xai, vercel | `x-ai/grok-4.3`, `grok-4.3`, `xai/grok-4.3` ✅ |
| `owl-alpha` | openrouter | `openrouter/owl-alpha` ✅ |
| `hy3-preview` | openrouter | `tencent/hy3-preview` ✅ |
| `trinity-large-thinking` | openrouter, kilo, venice, digitalocean | `arcee-ai/trinity-large-thinking`, etc. ✅ |
| `grok-4.20-beta` | openrouter | `x-ai/grok-4.20-beta` ✅ (live, not yet renamed in cache) |
| `grok-4.20` | openrouter | `x-ai/grok-4.20-multi-agent-beta` (similar but distinct) — the bare `x-ai/grok-4.20` ID is **not yet** in this cache snapshot |

**Implication:** The Grok rename hasn't fully landed in `models_dev_cache.json` on this dev host yet. The alias resolution is therefore **load-bearing** for users who manually update their `model.default` to the new ID before the cache refresh — they'd otherwise get an "unknown model" warning from Scarf's validator. Once the cache catches up, the alias falls back to cosmetic.

`vercel` provider: present, named `Vercel AI Gateway`, 248 models. Demotion target confirmed.

`arcee` overlay: present in Scarf's `overlayOnlyProviders`, NOT in `models_dev_cache.json`. Trinity Large Thinking still reaches users via the Arcee overlay's free-form picker + via OpenRouter / Vercel / DigitalOcean / Venice / Kilo where the cache surfaces it. No code change needed.

---

## Appendix B — schema-drift checklist

Before merging WS-6, verify the following are aligned across Swift and the upstream Hermes Python:

- [ ] `ModelCatalogService.overlayOnlyProviders` matches `HERMES_OVERLAYS` in `hermes_cli/providers.py` (no change in WS-6, but verify nothing drifted since WS-1).
- [ ] `ModelCatalogService.modelAliases` matches Hermes's deprecation map (verify the key location in `hermes_cli/providers.py` or wherever upstream tracks renames).
- [ ] `ModelCatalogService.demotedProviders` matches Hermes's deprioritized-provider list.
- [ ] `HermesConfig.openrouterResponseCacheEnabled` YAML key matches Hermes's config schema (resolve the Open Question).
- [ ] `HermesConfig.imageGenModel` YAML key (`image_gen.model`) matches Hermes's config schema. Currently confident — the release notes name the key explicitly.

---

**End of WS-6 plan.**
