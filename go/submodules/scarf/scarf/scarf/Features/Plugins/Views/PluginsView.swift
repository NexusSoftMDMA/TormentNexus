import SwiftUI
import ScarfCore
import ScarfDesign

struct PluginsView: View {
    // Coordinator-cached (t-aud24) so it survives section switches; still
    // observed via Observation (property reads in `body`).
    let viewModel: PluginsViewModel
    @State private var installIdentifier = ""
    @State private var showInstall = false
    @State private var pendingRemove: HermesPlugin?
    /// v0.16 Spotify sign-in sheet state. Only rendered when the spotify
    /// plugin is present and isV016OrLater is true.
    @State private var showSpotifySignIn = false
    @Environment(\.hermesCapabilities) private var capabilitiesStore

    init(viewModel: PluginsViewModel) {
        self.viewModel = viewModel
    }


    var body: some View {
        VStack(spacing: 0) {
            header
            if viewModel.isLoading && viewModel.plugins.isEmpty {
                ProgressView().padding()
            } else if viewModel.plugins.isEmpty {
                emptyState
            } else {
                list
            }
        }
        .background(ScarfColor.backgroundPrimary)
        .navigationTitle("Plugins")
        .loadingOverlay(
            viewModel.isLoading,
            label: "Loading plugins…",
            isEmpty: viewModel.plugins.isEmpty
        )
        .onAppear { viewModel.load() }
        .sheet(isPresented: $showInstall) { installSheet }
        .sheet(isPresented: $showSpotifySignIn) {
            SpotifySignInSheet(onSignedIn: {
                // No state to refresh in this view yet — chat picks
                // up the new token on next session start. Keep the
                // hook so a future "auth status" indicator can rebind.
            })
        }
        .confirmationDialog(
            pendingRemove.map { "Remove \($0.name)?" } ?? "",
            isPresented: Binding(get: { pendingRemove != nil }, set: { if !$0 { pendingRemove = nil } })
        ) {
            Button("Remove", role: .destructive) {
                if let plugin = pendingRemove { viewModel.remove(plugin) }
                pendingRemove = nil
            }
            Button("Cancel", role: .cancel) { pendingRemove = nil }
        }
    }

    private var header: some View {
        ScarfPageHeader(
            "Plugins",
            subtitle: "Hermes plugins discovered from `~/.hermes/plugins/`."
        ) {
            HStack(spacing: ScarfSpace.s2) {
                if let msg = viewModel.message {
                    Label(msg, systemImage: "info.circle.fill")
                        .scarfStyle(.caption)
                        .foregroundStyle(ScarfColor.success)
                }
                Button("Reload") { viewModel.load(force: true) }
                    .buttonStyle(ScarfGhostButton())
                Button {
                    installIdentifier = ""
                    showInstall = true
                } label: {
                    Label("Install", systemImage: "plus")
                }
                .buttonStyle(ScarfPrimaryButton())
            }
            .fixedSize(horizontal: true, vertical: false)
        }
    }

    private var emptyState: some View {
        VStack(spacing: 8) {
            Image(systemName: "app.badge.checkmark")
                .font(.largeTitle)
                .foregroundStyle(.secondary)
            Text("No plugins installed")
                .foregroundStyle(.secondary)
            Text("Plugins extend hermes with custom tools, providers, or memory backends.")
                .font(.caption)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
                .frame(maxWidth: 400)
            Button("Install a Plugin") {
                installIdentifier = ""
                showInstall = true
            }
            .buttonStyle(.borderedProminent)
            .controlSize(.small)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding()
    }

    private var list: some View {
        ScrollView {
            LazyVStack(spacing: 1) {
                // v0.16 Spotify sign-in affordance: surface when the
                // spotify plugin is present and we're on v0.16+. Reuses
                // the same SpotifySignInSheet and SpotifyAuthFlow as the
                // SkillsView placement (pre-v0.16 only).
                if capabilitiesStore?.capabilities.isV016OrLater == true,
                   viewModel.plugins.contains(where: { $0.name == "spotify" }) {
                    spotifyAuthRow
                        .padding()
                }
                ForEach(viewModel.plugins) { plugin in
                    row(plugin)
                }
            }
            .padding()
        }
    }

    private func row(_ plugin: HermesPlugin) -> some View {
        HStack(spacing: 12) {
            Image(systemName: plugin.enabled ? "app.badge.checkmark.fill" : "app.badge")
                .foregroundStyle(plugin.enabled ? .green : .secondary)
            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 6) {
                    Text(plugin.name)
                        .font(.system(.body, design: .monospaced, weight: .medium))
                    if !plugin.version.isEmpty {
                        Text(plugin.version)
                            .font(.caption2.monospaced())
                            .foregroundStyle(.secondary)
                    }
                    // v0.14 — surface plugins that replace a built-in
                    // tool with a visible badge so users notice
                    // overridden behavior. The flag comes from the
                    // plugin's manifest (`tool_override: true`).
                    if plugin.toolOverride {
                        ScarfBadge("tool-override", kind: .info)
                    }
                }
                if !plugin.source.isEmpty {
                    Text(plugin.source)
                        .font(.caption.monospaced())
                        .foregroundStyle(.secondary)
                        .textSelection(.enabled)
                }
            }
            Spacer()
            Button(plugin.enabled ? "Disable" : "Enable") {
                if plugin.enabled { viewModel.disable(plugin) } else { viewModel.enable(plugin) }
            }
            .controlSize(.small)
            Button("Update") { viewModel.update(plugin) }
                .controlSize(.small)
            Button("Remove", role: .destructive) { pendingRemove = plugin }
                .controlSize(.small)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(.quaternary.opacity(0.3))
    }

    /// Renders the v0.16 Spotify auth row in the plugins list when the
    /// spotify plugin is discovered. Tapping opens `SpotifySignInSheet`
    /// which drives `hermes auth spotify` end-to-end in-app.
    private var spotifyAuthRow: some View {
        HStack(spacing: 10) {
            Image(systemName: "music.note")
                .foregroundStyle(.green)
            VStack(alignment: .leading, spacing: 2) {
                Text("Sign in to Spotify")
                    .font(.callout.weight(.medium))
                Text("Authorise Hermes to control playback, search, and library actions.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Button("Sign In") { showSpotifySignIn = true }
                .buttonStyle(.borderedProminent)
                .controlSize(.small)
        }
        .padding(10)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color.green.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private var installSheet: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Install Plugin")
                .font(.headline)
            Text("Provide a Git URL (https://github.com/...) or a shorthand like `owner/repo`.")
                .font(.caption)
                .foregroundStyle(.secondary)
            TextField("github.com/owner/plugin-repo  or  owner/repo", text: $installIdentifier)
                .textFieldStyle(.roundedBorder)
                .font(.system(.caption, design: .monospaced))
            HStack {
                Spacer()
                Button("Cancel") { showInstall = false }
                Button("Install") {
                    viewModel.install(installIdentifier)
                    showInstall = false
                }
                .buttonStyle(.borderedProminent)
                .disabled(installIdentifier.trimmingCharacters(in: .whitespaces).isEmpty)
            }
        }
        .padding()
        .frame(minWidth: 500, minHeight: 200)
    }
}
