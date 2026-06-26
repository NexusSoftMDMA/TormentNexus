import SwiftUI
import ScarfCore
import ScarfDesign

/// Two-column model browser sheet. Left column lists providers, right column
/// lists models for the selected provider. Supports filtering and a "Custom…"
/// option for free-form model IDs not in the catalog.
///
/// Overlay-only providers (Nous Portal, OpenAI Codex, Qwen OAuth, …) have no
/// models.dev catalog entry, so their right column renders an overlay detail
/// view: subscription state for Nous, plus a free-form model-ID field for
/// users who know what they want. This is how the picker keeps parity with
/// `hermes model` on the CLI, which can reach these providers natively.
struct ModelPickerSheet: View {
    let initialProvider: String
    let initialModel: String
    let onSelect: (_ modelID: String, _ providerID: String) -> Void
    let onCancel: () -> Void

    @State private var providers: [HermesProviderInfo] = []
    @State private var selectedProviderID: String = ""
    @State private var models: [HermesModelInfo] = []
    @State private var selectedModelID: String = ""
    @State private var searchText: String = ""
    /// True while the initial catalog load (or a per-provider model
    /// reload) is in flight. Drives the loading-overlay placeholder.
    /// Pre-fix this work ran synchronously inside `.onAppear` — issue
    /// #59. The catalog file is multi-MB on remote contexts; sync I/O
    /// on the MainActor froze the picker for 1–2 minutes.
    @State private var isLoadingCatalog: Bool = true

    // Custom model entry — used when the catalog doesn't have the exact model
    // the user needs (e.g., provider-prefixed IDs like "openrouter/some/model").
    @State private var customMode: Bool = false
    @State private var customModelID: String = ""
    @State private var customProviderID: String = ""

    // Overlay-provider model entry — distinct from `customMode` because the
    // provider is pinned; only the model ID is user-editable.
    @State private var overlayModelID: String = ""

    // Subscription state for the Nous Portal row / detail view. Loaded on
    // appear; stays in-memory for the life of the sheet.
    @State private var subscription: NousSubscriptionState = .absent

    /// Drives presentation of the Nous sign-in sheet. Bound to the
    /// "Sign in to Nous Portal" button in the subscription summary.
    @State private var showNousSignIn: Bool = false

    /// Cached + freshly-fetched Nous model list for the picker's
    /// nous-overlay branch. Populated on appear (cache-first) and
    /// refreshed when the user signs in or hits the Refresh button.
    @State private var nousModels: [NousModel] = []
    @State private var nousFetchedAt: Date?
    @State private var nousRefreshError: String?
    @State private var nousIsRefreshing: Bool = false
    /// When true, render the Nous detail with the original free-form
    /// TextField + manual hint instead of the model list. Used when
    /// the user explicitly wants to type a model not in the catalog —
    /// the API list is comprehensive but not infallible, so always
    /// keep the escape hatch reachable.
    @State private var nousManualEntry: Bool = false

    /// Validation failure surfaced on Select when the typed / selected
    /// model isn't in the chosen provider's catalog. Pass-1 M7 #5
    /// cross-platform fix — previously Scarf let you save any string
    /// and the failure only appeared hours later at runtime.
    @State private var validationIssue: ModelValidationIssue?

    @Environment(\.serverContext) private var serverContext
    private var catalog: ModelCatalogService { ModelCatalogService(context: serverContext) }
    private var subscriptionService: NousSubscriptionService { NousSubscriptionService(context: serverContext) }

    var body: some View {
        VStack(spacing: 0) {
            header
            Divider()
            if customMode {
                customEntry
            } else {
                HSplitView {
                    providerColumn.frame(minWidth: 220, idealWidth: 240)
                    modelColumn.frame(minWidth: 340)
                }
            }
            Divider()
            footer
        }
        .frame(minWidth: 720, minHeight: 520)
        .overlay {
            if isLoadingCatalog {
                ProgressView("Loading providers…")
                    .progressViewStyle(.circular)
                    .padding()
                    .background(.regularMaterial)
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            }
        }
        .task {
            // Off-MainActor read of the multi-megabyte models.dev cache
            // (via SSHTransport on remote contexts). Pre-fix this ran
            // sync inside `.onAppear` and froze the picker for 1–2
            // minutes on remote contexts (issue #59).
            isLoadingCatalog = true
            providers = await catalog.loadProvidersAsync()
            selectedProviderID = initialProvider.isEmpty ? (providers.first?.providerID ?? "") : initialProvider
            selectedModelID = initialModel
            overlayModelID = initialModel
            // subscriptionService.loadState() reads auth.json — tiny
            // on local but still SSH-backed on remote, so route it
            // through a detached task too. The result is a small
            // value type; safe to assign back onto MainActor.
            let svc = subscriptionService
            subscription = await Task.detached { svc.loadState() }.value
            await loadModelsForSelectionAsync()
            isLoadingCatalog = false
        }
        .sheet(isPresented: $showNousSignIn) {
            NousSignInSheet {
                // Refresh subscription immediately so the right-column
                // status row flips to "active" without waiting for the
                // picker to be re-opened.
                subscription = subscriptionService.loadState()
                // Sign-in unlocked the bearer token — kick a fresh
                // model-list fetch so the picker populates without the
                // user needing to hit Refresh manually.
                Task { await refreshNousModels(forceRefresh: true) }
            }
        }
        .alert(item: $validationIssue) { issue in
            Alert(
                title: Text("Model not available"),
                message: Text(validationMessage(for: issue)),
                primaryButton: .default(Text("Pick from catalog")) {
                    validationIssue = nil
                    customMode = false
                },
                secondaryButton: .cancel(Text("Edit"))
            )
        }
    }

    private func validationMessage(for issue: ModelValidationIssue) -> String {
        var msg = "\(issue.modelID) isn't in \(issue.providerName)'s catalog."
        if !issue.suggestions.isEmpty {
            msg += " Did you mean one of:\n• " + issue.suggestions.joined(separator: "\n• ")
        } else {
            msg += " Pick one from the catalog or double-check the spelling."
        }
        return msg
    }

    private var header: some View {
        HStack(spacing: 8) {
            Image(systemName: "cpu")
            Text("Select Model")
                .scarfStyle(.headline)
            Spacer()
            if !customMode {
                TextField("Search…", text: $searchText)
                    .textFieldStyle(.roundedBorder)
                    .frame(width: 240)
            }
            Button(customMode ? "Back to Catalog" : "Custom…") {
                customMode.toggle()
                if customMode {
                    customModelID = initialModel
                    customProviderID = initialProvider
                }
            }
            .controlSize(.small)
        }
        .padding()
    }

    private var providerColumn: some View {
        List(selection: Binding(
            get: { selectedProviderID },
            set: { newValue in
                selectedProviderID = newValue
                Task { await loadModelsForSelectionAsync() }
            }
        )) {
            ForEach(filteredProviders) { provider in
                providerRow(provider)
                    .tag(provider.providerID)
            }
        }
        .listStyle(.inset)
    }

    @ViewBuilder
    private func providerRow(_ provider: HermesProviderInfo) -> some View {
        HStack(spacing: 6) {
            Text(provider.providerName)
            if provider.subscriptionGated {
                capsuleTag("Subscription", tint: .accentColor)
            }
            Spacer()
            if !provider.isOverlay {
                Text("\(provider.modelCount)")
                    .font(.caption2.monospaced())
                    .foregroundStyle(.secondary)
            }
        }
    }

    @ViewBuilder
    private var modelColumn: some View {
        if let selected = providers.first(where: { $0.providerID == selectedProviderID }) {
            if selected.providerID == "nous" {
                nousOverlayDetail(selected)
            } else if selected.isOverlay {
                overlayProviderDetail(selected)
            } else {
                cachedModelList
            }
        } else {
            cachedModelList
        }
    }

    private var cachedModelList: some View {
        List(selection: $selectedModelID) {
            ForEach(filteredModels) { model in
                VStack(alignment: .leading, spacing: 2) {
                    HStack {
                        Text(model.modelName)
                            .font(.system(.body, design: .default, weight: .medium))
                        Spacer()
                        if let ctx = model.contextDisplay {
                            Text("\(ctx) ctx")
                                .font(.caption2.monospaced())
                                .foregroundStyle(.secondary)
                        }
                    }
                    HStack(spacing: 6) {
                        Text(model.modelID)
                            .font(.caption2.monospaced())
                            .foregroundStyle(.tertiary)
                        if let cost = model.costDisplay {
                            Text("·")
                                .foregroundStyle(.tertiary)
                            Text(cost)
                                .font(.caption2.monospaced())
                                .foregroundStyle(.secondary)
                        }
                        if model.toolCall {
                            capsuleTag("tools")
                        }
                        if model.reasoning {
                            capsuleTag("reasoning")
                        }
                    }
                }
                .padding(.vertical, 2)
                .tag(model.modelID)
            }
        }
        .listStyle(.inset)
        .overlay {
            if filteredModels.isEmpty {
                ContentUnavailableView("No Models", systemImage: "cpu", description: Text("This provider has no catalogued models."))
            }
        }
    }

    /// Right-column detail for Nous Portal — same overlay shape as
    /// `overlayProviderDetail` but with a live model list fetched from
    /// Nous's OpenAI-compatible `/v1/models` endpoint. The list is
    /// cache-first so opening the sheet feels instant; refresh runs
    /// in the background. Falls back to a hard-coded short list when
    /// the user has no token AND no cache (offline first-run on a
    /// fresh remote install). The "Custom…" button below the list
    /// flips to the original free-form TextField — Nous occasionally
    /// adds a model before our cache hits 24h, and we don't want
    /// users locked out of the latest releases.
    @ViewBuilder
    private func nousOverlayDetail(_ provider: HermesProviderInfo) -> some View {
        let overlay = catalog.overlayMetadata(for: provider.providerID)
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                HStack(alignment: .firstTextBaseline, spacing: 8) {
                    Text(provider.providerName).font(.title3.bold())
                    if provider.subscriptionGated {
                        capsuleTag("Subscription", tint: .accentColor)
                    }
                }
                if provider.subscriptionGated {
                    subscriptionSummary(provider: provider, overlay: overlay)
                }
                Divider()
                if nousManualEntry {
                    nousManualEntryBlock(provider: provider)
                } else {
                    nousModelListBlock
                }
                if let docURL = overlay?.docURL, let url = URL(string: docURL) {
                    Link(destination: url) {
                        Label("Setup documentation", systemImage: "book")
                            .font(.caption)
                    }
                }
                Spacer(minLength: 0)
            }
            .padding()
        }
    }

    @ViewBuilder
    private var nousModelListBlock: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(alignment: .firstTextBaseline) {
                Text("Available models")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                if nousIsRefreshing {
                    HStack(spacing: 4) {
                        ProgressView().controlSize(.mini)
                        Text("Refreshing…").font(.caption2).foregroundStyle(.tertiary)
                    }
                } else {
                    Button {
                        Task { await refreshNousModels(forceRefresh: true) }
                    } label: {
                        Label("Refresh", systemImage: "arrow.clockwise")
                            .labelStyle(.iconOnly)
                    }
                    .buttonStyle(.borderless)
                    .help(nousFetchedAtTooltip)
                }
                Button("Custom…") { nousManualEntry = true }
                    .controlSize(.small)
            }
            if let err = nousRefreshError, !nousIsRefreshing {
                HStack(spacing: 6) {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundStyle(.orange)
                    Text(err)
                        .font(.caption2)
                        .foregroundStyle(.secondary)
                        .fixedSize(horizontal: false, vertical: true)
                }
            }
            List(selection: $overlayModelID) {
                ForEach(filteredNousModels) { model in
                    VStack(alignment: .leading, spacing: 2) {
                        Text(model.id)
                            .font(.system(.body, design: .monospaced))
                        if let owner = model.owned_by, !owner.isEmpty {
                            Text(owner)
                                .font(.caption2)
                                .foregroundStyle(.tertiary)
                        }
                    }
                    .tag(model.id)
                }
            }
            .listStyle(.inset)
            .frame(minHeight: 220)
            .overlay {
                if filteredNousModels.isEmpty && !nousIsRefreshing {
                    if nousModels.isEmpty {
                        ContentUnavailableView(
                            "No models loaded",
                            systemImage: "cpu",
                            description: Text("Sign in to Nous Portal to load the catalog, or enter a model ID manually.")
                        )
                    } else {
                        // Models loaded but the search filtered them all
                        // out. Different message so the user knows the
                        // catalog is fine, just their query didn't match.
                        ContentUnavailableView(
                            "No matches",
                            systemImage: "magnifyingglass",
                            description: Text("No models match \"\(searchText)\".")
                        )
                    }
                }
            }
            if nousFetchedAt == nil && !nousModels.isEmpty {
                Text("Showing built-in fallback list — couldn't reach Nous to refresh.")
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
            }
            Text("Leave blank in config to let Hermes pick the default Nous model. Picking one above writes it explicitly.")
                .font(.caption2)
                .foregroundStyle(.tertiary)
        }
    }

    @ViewBuilder
    private func nousManualEntryBlock(provider: HermesProviderInfo) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text("Model ID").font(.caption).foregroundStyle(.secondary)
                Spacer()
                Button("Use list") { nousManualEntry = false }
                    .controlSize(.small)
            }
            TextField(modelIDPlaceholder(for: provider), text: $overlayModelID)
                .textFieldStyle(.roundedBorder)
                .font(.system(.caption, design: .monospaced))
            Text("Type a model ID exactly as Nous expects it. Leave blank to use Hermes's default.")
                .font(.caption2)
                .foregroundStyle(.tertiary)
        }
    }

    private static let fetchedAtFormatter: RelativeDateTimeFormatter = {
        let f = RelativeDateTimeFormatter()
        f.unitsStyle = .short
        return f
    }()

    private var nousFetchedAtTooltip: String {
        guard let date = nousFetchedAt else {
            return "Fetch the latest model list from Nous."
        }
        return "Last refreshed \(Self.fetchedAtFormatter.localizedString(for: date, relativeTo: Date()))"
    }

    /// Right-column detail for overlay-only providers (Nous Portal, OpenAI
    /// Codex, Qwen OAuth, …). models.dev has no catalog for them, so the user
    /// either trusts Hermes's default (subscription providers) or types a
    /// model ID they know is valid for the provider's API.
    @ViewBuilder
    private func overlayProviderDetail(_ provider: HermesProviderInfo) -> some View {
        let overlay = catalog.overlayMetadata(for: provider.providerID)
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                HStack(alignment: .firstTextBaseline, spacing: 8) {
                    Text(provider.providerName).font(.title3.bold())
                    if provider.subscriptionGated {
                        capsuleTag("Subscription", tint: .accentColor)
                    }
                }
                if provider.subscriptionGated {
                    subscriptionSummary(provider: provider, overlay: overlay)
                } else {
                    Text(overlayInstruction(for: overlay?.authType))
                        .font(.callout)
                        .foregroundStyle(.secondary)
                        .fixedSize(horizontal: false, vertical: true)
                }

                Divider()

                VStack(alignment: .leading, spacing: 4) {
                    Text("Model ID").font(.caption).foregroundStyle(.secondary)
                    TextField(modelIDPlaceholder(for: provider), text: $overlayModelID)
                        .textFieldStyle(.roundedBorder)
                        .font(.system(.caption, design: .monospaced))
                    if provider.subscriptionGated {
                        Text("Leave blank to use Hermes's default Nous model.")
                            .font(.caption2)
                            .foregroundStyle(.tertiary)
                    }
                }

                if let docURL = overlay?.docURL, let url = URL(string: docURL) {
                    Link(destination: url) {
                        Label("Setup documentation", systemImage: "book")
                            .font(.caption)
                    }
                }

                Spacer(minLength: 0)
            }
            .padding()
        }
    }

    @ViewBuilder
    private func subscriptionSummary(provider: HermesProviderInfo, overlay: HermesProviderOverlay?) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Paid Nous Portal subscribers route web search, image generation, TTS, and browser automation through their subscription — no separate API keys needed.")
                .font(.callout)
                .foregroundStyle(.secondary)
                .fixedSize(horizontal: false, vertical: true)

            HStack(spacing: 6) {
                Image(systemName: subscription.subscribed ? "checkmark.circle.fill" : "exclamationmark.circle")
                    .foregroundStyle(subscription.subscribed ? Color.green : Color.secondary)
                if subscription.subscribed {
                    Text("Subscription active — active provider is Nous.")
                } else if subscription.present {
                    Text("Signed in to Nous, but another provider is active.")
                        .foregroundStyle(.secondary)
                } else {
                    Text("Not signed in yet.")
                        .foregroundStyle(.secondary)
                }
            }
            .font(.callout)

            if !subscription.subscribed {
                Button {
                    showNousSignIn = true
                } label: {
                    Label("Sign in to Nous Portal", systemImage: "person.badge.key.fill")
                }
                .buttonStyle(ScarfPrimaryButton())
                .controlSize(.regular)
            }
        }
    }

    private func overlayInstruction(for authType: HermesProviderOverlay.AuthType?) -> String {
        switch authType {
        case .oauthExternal:
            return "Sign in through the provider's OAuth flow — run `hermes auth` from a terminal, then pick the provider to complete sign-in. Back here, set the model ID you want to use."
        case .externalProcess:
            return "Uses an external process (e.g. a local agent bridge). Run `hermes auth` from a terminal to complete the link, then set the model ID you want to use."
        case .oauthDeviceCode:
            return "Sign in via device-code flow — run `hermes auth` from a terminal and follow the printed URL."
        default:
            return "This provider isn't in the models.dev catalog. Enter the model ID you want to use — Hermes will pass it through to the provider verbatim."
        }
    }

    private func modelIDPlaceholder(for provider: HermesProviderInfo) -> String {
        switch provider.providerID {
        case "nous":          return "e.g. hermes-3"
        case "openai-codex":  return "e.g. gpt-5-codex"
        case "qwen-oauth":    return "e.g. qwen3-coder-plus"
        default:              return "e.g. model-name"
        }
    }

    private var customEntry: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Use a model not in the catalog. Hermes accepts any string the provider recognizes, including provider-prefixed forms like \"openrouter/anthropic/claude-opus-4.6\".")
                .font(.caption)
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 4) {
                Text("Model ID").font(.caption).foregroundStyle(.secondary)
                TextField("e.g. openai/gpt-4o", text: $customModelID)
                    .textFieldStyle(.roundedBorder)
                    .font(.system(.caption, design: .monospaced))
            }
            VStack(alignment: .leading, spacing: 4) {
                Text("Provider").font(.caption).foregroundStyle(.secondary)
                TextField("e.g. openai", text: $customProviderID)
                    .textFieldStyle(.roundedBorder)
                    .font(.system(.caption, design: .monospaced))
                Text("Leave blank to infer from the model ID's prefix (\"openai/...\" → openai).")
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
            }
            Spacer()
        }
        .padding()
    }

    private var footer: some View {
        HStack {
            if customMode {
                Text(customProviderPreview)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            } else if let preview = selectedPreview {
                Text(preview)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button("Cancel") { onCancel() }
            Button("Select") { submitSelection() }
                .buttonStyle(ScarfPrimaryButton())
                .disabled(!canSubmit)
        }
        .padding()
    }

    // MARK: - Helpers

    private var filteredProviders: [HermesProviderInfo] {
        guard !searchText.isEmpty else { return providers }
        let q = searchText.lowercased()
        return providers.filter {
            $0.providerName.lowercased().contains(q) || $0.providerID.lowercased().contains(q)
        }
    }

    private var filteredModels: [HermesModelInfo] {
        guard !searchText.isEmpty else { return models }
        let q = searchText.lowercased()
        return models.filter {
            $0.modelName.lowercased().contains(q) || $0.modelID.lowercased().contains(q)
        }
    }

    /// Same shape as `filteredModels` but for the Nous overlay path
    /// (`nousModels` is `[NousModel]`, not `[HermesModelInfo]`).
    /// Nous returned 402 models in the user's capture; without a
    /// filter the picker is a flat unsearchable list. Reuses the
    /// same `searchText` field so the user types once and both
    /// paths respond.
    private var filteredNousModels: [NousModel] {
        guard !searchText.isEmpty else { return nousModels }
        let q = searchText.lowercased()
        return nousModels.filter {
            $0.id.lowercased().contains(q) || ($0.owned_by ?? "").lowercased().contains(q)
        }
    }

    private var isSelectedProviderOverlay: Bool {
        providers.first(where: { $0.providerID == selectedProviderID })?.isOverlay ?? false
    }

    private var isSelectedProviderSubscriptionGated: Bool {
        providers.first(where: { $0.providerID == selectedProviderID })?.subscriptionGated ?? false
    }

    private var canSubmit: Bool {
        if customMode {
            return !customModelID.trimmingCharacters(in: .whitespaces).isEmpty
        }
        if isSelectedProviderOverlay {
            // Subscription-gated providers can submit with an empty model ID
            // (Hermes picks its default). Other overlays require a model ID.
            if isSelectedProviderSubscriptionGated { return true }
            return !overlayModelID.trimmingCharacters(in: .whitespaces).isEmpty
        }
        return !selectedModelID.isEmpty
    }

    private var selectedPreview: String? {
        if isSelectedProviderOverlay {
            let trimmed = overlayModelID.trimmingCharacters(in: .whitespaces)
            if trimmed.isEmpty {
                return selectedProviderID.isEmpty ? nil : "\(selectedProviderID) / (default)"
            }
            return "\(selectedProviderID) / \(trimmed)"
        }
        guard !selectedModelID.isEmpty, !selectedProviderID.isEmpty else { return nil }
        return "\(selectedProviderID) / \(selectedModelID)"
    }

    private var customProviderPreview: String {
        let resolved = resolvedCustomProvider()
        return resolved.isEmpty ? "Provider will not be changed" : "Provider → \(resolved)"
    }

    /// Async variant of the per-provider catalog read. Pre-fix this
    /// was synchronous on the MainActor and froze the picker every
    /// time the user clicked a different provider — same root cause
    /// as the open-sheet freeze (issue #59). Routes through
    /// `loadModelsAsync(for:)` which dispatches the SSHTransport
    /// file read off the main thread.
    private func loadModelsForSelectionAsync() async {
        guard !selectedProviderID.isEmpty else {
            models = []
            return
        }
        models = await catalog.loadModelsAsync(for: selectedProviderID)
        // If the current selection is not in the new list, don't try to keep
        // stale highlight state — clear unless the user originally had this model.
        if !models.contains(where: { $0.modelID == selectedModelID }) {
            selectedModelID = models.first?.modelID ?? ""
        }
        // Cache-first kick for the Nous catalog. Renders from cache
        // immediately, fires a background refresh if stale or empty.
        if selectedProviderID == "nous" {
            Task { await refreshNousModels(forceRefresh: false) }
        }
    }

    /// Cache-first load of the Nous model list. Updates the four
    /// `@State` vars the detail view reads. Force-refresh skips the
    /// TTL check so the user-tapped Refresh button always hits the
    /// network — the cache write keeps the next sheet-open instant.
    private func refreshNousModels(forceRefresh: Bool) async {
        let service = NousModelCatalogService(context: serverContext)
        // PRE-FIX (v2.7): this used to call `service.readCache()`
        // synchronously here for instant first-paint, then call
        // `service.loadModels(...)` which calls `readCache()` AGAIN
        // internally — paying the SSH round-trip TWICE per picker
        // open. On a remote with a corrupt or oversized cache file,
        // the duplicated reads stacked two 60-second timeouts for a
        // 120-second picker stall. ScarfMon perf capture confirmed
        // the duplication.
        //
        // loadModels() already serves cache-first on its happy path
        // (returns `.cache(...)` when fresh), so the inline readCache
        // here is redundant. Drop it; trust loadModels' built-in
        // cache-first behavior. One readCache call per picker open.
        nousIsRefreshing = true
        let result = await service.loadModels(forceRefresh: forceRefresh)
        nousIsRefreshing = false
        switch result {
        case .fresh(let models, let fetchedAt):
            nousModels = models
            nousFetchedAt = fetchedAt
            nousRefreshError = nil
        case .cache(let models, let fetchedAt, let refreshError):
            nousModels = models
            nousFetchedAt = fetchedAt
            nousRefreshError = refreshError
        case .fallback(let models, let reason):
            nousModels = models
            nousFetchedAt = nil
            nousRefreshError = reason
        }
        // Pre-fill `overlayModelID` with the user's previously chosen
        // model when it's in the freshly-loaded list — otherwise the
        // selection state highlights nothing on first paint.
        if !overlayModelID.isEmpty,
           !nousModels.contains(where: { $0.id == overlayModelID }) {
            // Leave overlayModelID alone — it's a user-typed value
            // that may legitimately not be in the catalog.
        }
    }

    /// When the user enters a custom model ID without explicitly naming a
    /// provider, infer from a `provider/model` prefix if present. Otherwise
    /// fall back to whatever is currently selected (we never blank out the
    /// existing provider silently).
    private func resolvedCustomProvider() -> String {
        let explicit = customProviderID.trimmingCharacters(in: .whitespaces)
        if !explicit.isEmpty { return explicit }
        if let slash = customModelID.firstIndex(of: "/") {
            return String(customModelID[customModelID.startIndex..<slash])
        }
        return ""
    }

    private func submitSelection() {
        let (model, provider): (String, String)
        if customMode {
            model = customModelID.trimmingCharacters(in: .whitespaces)
            provider = resolvedCustomProvider()
        } else if isSelectedProviderOverlay {
            model = overlayModelID.trimmingCharacters(in: .whitespaces)
            provider = selectedProviderID
        } else {
            model = selectedModelID
            provider = selectedProviderID
        }

        // Block unknown models before they land in config.yaml.
        // Overlay-only providers short-circuit to .valid inside the
        // validator because their catalogs aren't in models.dev.
        switch catalog.validateModel(model, for: provider) {
        case .valid, .unknownProvider:
            onSelect(model, provider)
        case .invalid(let providerName, let suggestions):
            validationIssue = ModelValidationIssue(
                modelID: model,
                providerName: providerName,
                suggestions: suggestions
            )
        }
    }

    private func capsuleTag(_ text: String, tint: Color = .secondary) -> some View {
        Text(text)
            .font(.caption2)
            .foregroundStyle(tint == .secondary ? AnyShapeStyle(.secondary) : AnyShapeStyle(tint))
            .padding(.horizontal, 5)
            .padding(.vertical, 1)
            .background(tint == .secondary ? AnyShapeStyle(.quaternary) : AnyShapeStyle(tint.opacity(0.15)))
            .clipShape(Capsule())
    }
}

/// Carrier for the catalog-validation alert. Identifiable so SwiftUI's
/// `.alert(item:)` can key off each unique issue.
private struct ModelValidationIssue: Identifiable {
    let id = UUID()
    let modelID: String
    let providerName: String
    let suggestions: [String]
}
