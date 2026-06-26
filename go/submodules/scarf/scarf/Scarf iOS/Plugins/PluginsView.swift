import SwiftUI
import ScarfCore
import ScarfDesign

/// iOS read-only Plugins view (v2.6).
///
/// Walks `~/.hermes/plugins/` (each subdirectory is one plugin) and
/// reads the optional `plugin.json` / `plugin.yaml` manifest for each.
/// Mirrors the Mac PluginsViewModel's filesystem-first source-of-truth
/// approach — `hermes plugins list`'s box-drawn output is fragile to
/// parse from a phone form-factor.
///
/// Install / update / remove / enable / disable verbs stay on Mac for
/// v2.6 — installing a plugin from a phone is an unusual flow.
struct PluginsView: View {
    let config: IOSServerConfig

    @State private var plugins: [PluginRow] = []
    @State private var isLoading = true
    @State private var lastError: String?
    @Environment(\.serverContext) private var contextFromEnv

    private var context: ServerContext {
        config.toServerContext(id: contextFromEnv.id)
    }

    var body: some View {
        List {
            if let err = lastError {
                Section {
                    Label(err, systemImage: "exclamationmark.triangle.fill")
                        .foregroundStyle(ScarfColor.warning)
                }
            }

            if plugins.isEmpty && !isLoading {
                Section {
                    ContentUnavailableView(
                        "No plugins installed",
                        systemImage: "app.badge.checkmark",
                        description: Text("Hermes plugins live under `~/.hermes/plugins/<name>/`. Install one with `hermes plugins install <repo>` from the Mac app.")
                    )
                }
            } else {
                ForEach(plugins) { plugin in
                    Section(plugin.name) {
                        HStack {
                            statusBadge(plugin.enabled)
                            if !plugin.version.isEmpty {
                                Text("v\(plugin.version)")
                                    .font(ScarfFont.monoSmall)
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                        }
                        if !plugin.source.isEmpty {
                            LabeledContent("Source", value: plugin.source)
                                .font(.caption.monospaced())
                        }
                        Text(plugin.path)
                            .font(.caption.monospaced())
                            .foregroundStyle(.secondary)
                            .textSelection(.enabled)
                    }
                }
            }
        }
        .navigationTitle("Plugins")
        .navigationBarTitleDisplayMode(.large)
        .refreshable { await load() }
        .task { await load() }
    }

    private func statusBadge(_ enabled: Bool) -> some View {
        ScarfBadge(enabled ? "Enabled" : "Disabled", kind: enabled ? .success : .neutral)
    }

    private func load() async {
        isLoading = true
        defer { isLoading = false }
        let ctx = context
        let entries = await Task.detached {
            Self.scan(context: ctx)
        }.value
        self.plugins = entries
    }

    nonisolated private static func scan(context: ServerContext) -> [PluginRow] {
        let transport = context.makeTransport()
        let dir = context.paths.pluginsDir
        guard let entries = try? transport.listDirectory(dir) else { return [] }
        var results: [PluginRow] = []
        for entry in entries.sorted() where !entry.hasPrefix(".") {
            let path = dir + "/" + entry
            guard transport.stat(path)?.isDirectory == true else { continue }
            let manifest = readManifest(path: path, context: context)
            let disabled = transport.fileExists(path + "/.disabled")
            results.append(PluginRow(
                name: entry,
                version: manifest.version,
                source: manifest.source,
                path: path,
                enabled: !disabled
            ))
        }
        return results
    }

    /// Read `plugin.json` first; fall back to `plugin.yaml` for plugins
    /// that author manifest in YAML. Same shape as the Mac VM so
    /// parsing stays consistent across targets.
    nonisolated private static func readManifest(path: String, context: ServerContext) -> (source: String, version: String) {
        if let data = context.readData(path + "/plugin.json"),
           let obj = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
            let source = (obj["source"] as? String) ?? (obj["repository"] as? String) ?? (obj["url"] as? String) ?? ""
            let version = (obj["version"] as? String) ?? ""
            return (source, version)
        }
        if let yaml = context.readText(path + "/plugin.yaml") {
            let parsed = HermesYAML.parseNestedYAML(yaml)
            let source = HermesYAML.stripYAMLQuotes(parsed.values["source"] ?? parsed.values["repository"] ?? parsed.values["url"] ?? "")
            let version = HermesYAML.stripYAMLQuotes(parsed.values["version"] ?? "")
            return (source, version)
        }
        return ("", "")
    }

    private struct PluginRow: Identifiable {
        var id: String { name }
        let name: String
        let version: String
        let source: String
        let path: String
        let enabled: Bool
    }
}
