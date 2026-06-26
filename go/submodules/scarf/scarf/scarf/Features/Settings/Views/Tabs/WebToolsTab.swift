import SwiftUI
import ScarfCore
import ScarfDesign

/// Web Tools tab — search + extract backend pickers. Pre-v0.13 hosts
/// see a single "Combined backend" row writing to the legacy
/// `web_tools.backend` key. v0.13+ hosts see two rows writing to the
/// per-capability split keys (`web_tools.search.backend` +
/// `web_tools.extract.backend`); SearXNG appears in the search picker
/// only because Hermes registers it as a search-only backend.
struct WebToolsTab: View {
    @Bindable var viewModel: SettingsViewModel
    @Environment(\.hermesCapabilities) private var capabilitiesStore

    private var split: Bool {
        capabilitiesStore?.capabilities.hasWebToolsBackendSplit ?? false
    }

    // Wire-accurate against `tools/web_tools.py` in Hermes v2026.5.16
    // (line 143/1364 in v0.14: the canonical set is exa, parallel,
    // firecrawl, tavily, searxng, brave-free, ddgs). `brave-free` and
    // `ddgs` are v0.14 additions — gated below so pre-v0.14 hosts only
    // see the older five. Search-only entries (searxng / brave-free /
    // ddgs) don't appear in the extract picker.
    private static let v013SearchBackends: [String] = [
        "exa", "parallel", "firecrawl", "tavily", "searxng"
    ]
    private static let v014SearchAdditions: [String] = [
        "brave-free", "ddgs"
    ]
    /// v0.15 search-only addition — xAI Web Search (`plugins/web/xai`),
    /// reuses Grok OAuth / `XAI_API_KEY`. Search-only, so it isn't in the
    /// extract picker.
    private static let v015SearchAdditions: [String] = [
        "xai"
    ]
    private static let extractBackends: [String] = [
        "exa", "parallel", "firecrawl", "tavily"
    ]
    /// v0.12 combined-backend list — pre-v0.13 hosts that haven't yet
    /// split search/extract into per-capability keys. Conservative
    /// superset: every backend that handles either capability.
    private static let combinedBackends: [String] = [
        "exa", "parallel", "firecrawl", "tavily", "searxng"
    ]

    private var searchBackends: [String] {
        let caps = capabilitiesStore?.capabilities ?? .empty
        var list = Self.v013SearchBackends
        if caps.hasBraveFreeSearchBackend { list.append("brave-free") }
        if caps.hasDDGSearchBackend { list.append("ddgs") }
        if caps.hasXAIWebSearchBackend { list.append("xai") }
        return list
    }

    var body: some View {
        if split {
            SettingsSection(title: "Web Tools", icon: "globe.americas") {
                PickerRow(
                    label: "Search backend",
                    selection: viewModel.config.webToolsSearchBackend,
                    options: searchBackends
                ) { viewModel.setWebToolsSearchBackend($0) }
                PickerRow(
                    label: "Extract backend",
                    selection: viewModel.config.webToolsExtractBackend,
                    options: Self.extractBackends
                ) { viewModel.setWebToolsExtractBackend($0) }
            }
            // Footer copy adapts to the connected host — v0.14 adds the
            // two new free-tier search backends; older hosts see the
            // SearXNG-joined-search-only line.
            let caps = capabilitiesStore?.capabilities ?? .empty
            let footerCopy: String = {
                if caps.hasXAIWebSearchBackend {
                    return "v0.15 added xAI Web Search (reuses your Grok OAuth / XAI_API_KEY). v0.14 added Brave Search (free tier; honors BRAVE_SEARCH_API_KEY) and DuckDuckGo (DDGS). All three are search-only. Backend-specific tuning lives in the raw YAML editor for now."
                }
                if caps.hasBraveFreeSearchBackend || caps.hasDDGSearchBackend {
                    return "v0.14 added Brave Search (free tier; honors BRAVE_SEARCH_API_KEY) and DuckDuckGo (DDGS) as search-only backends. Backend-specific tuning lives in the raw YAML editor for now."
                }
                return "SearXNG is search-only. Backend-specific tuning (host URLs, API keys) lives in the raw YAML editor for now."
            }()
            Text(footerCopy)
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundMuted)
                .padding(.horizontal, ScarfSpace.s4)
        } else {
            SettingsSection(title: "Web Tools", icon: "globe.americas") {
                PickerRow(
                    label: "Backend",
                    selection: viewModel.config.webToolsBackend,
                    options: Self.combinedBackends
                ) { viewModel.setWebToolsBackend($0) }
            }
            Text("Hermes v0.13 splits search and extract into separate backends. Update Hermes to access the per-capability picker.")
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundFaint)
                .padding(.horizontal, ScarfSpace.s4)
        }
    }
}
