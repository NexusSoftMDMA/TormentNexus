import SwiftUI
import ScarfCore
import ScarfDesign

/// Secrets tab — Bitwarden Secrets Manager bootstrap (`secrets.bitwarden.*`,
/// Hermes v0.15). A single bootstrap access token (whose env-var NAME is set
/// here; the token VALUE lives in `~/.hermes/.env`) lets Hermes resolve
/// per-provider API keys from a Bitwarden Secrets Manager project, replacing
/// per-provider keys scattered across config/.env.
///
/// The whole tab is release-gated in `SettingsView` — pre-v0.15 hosts never
/// see it. Server URL empty = US Cloud; `https://vault.bitwarden.eu` = EU; or
/// a self-hosted vault URL.
struct SecretsTab: View {
    @Bindable var viewModel: SettingsViewModel
    @State private var statusOutput: String = ""
    @State private var showStatus = false

    private var bitwarden: BitwardenSettings { viewModel.config.bitwarden }

    var body: some View {
        SettingsSection(title: "Bitwarden Secrets Manager", icon: "key.horizontal") {
            ToggleRow(label: "Enabled", isOn: bitwarden.enabled) { viewModel.setBitwardenEnabled($0) }
            EditableTextField(label: "Access Token Env Var", value: bitwarden.accessTokenEnv) { viewModel.setBitwardenAccessTokenEnv($0) }
            EditableTextField(label: "Project ID", value: bitwarden.projectID) { viewModel.setBitwardenProjectID($0) }
            ToggleRow(label: "Override Existing", isOn: bitwarden.overrideExisting) { viewModel.setBitwardenOverrideExisting($0) }
            EditableTextField(label: "Server URL", value: bitwarden.serverURL) { viewModel.setBitwardenServerURL($0) }
            StepperRow(label: "Cache TTL (s)", value: bitwarden.cacheTTLSeconds, range: 0...86400, step: 30) { viewModel.setBitwardenCacheTTLSeconds($0) }
            ToggleRow(label: "Auto Install SDK", isOn: bitwarden.autoInstall) { viewModel.setBitwardenAutoInstall($0) }
        }

        Text("The bootstrap access token itself goes in `~/.hermes/.env` as the env var named above (default `BWS_ACCESS_TOKEN`) — never in config.yaml. Leave Server URL empty for US Cloud, use `https://vault.bitwarden.eu` for EU, or a self-hosted vault URL.")
            .scarfStyle(.caption)
            .foregroundStyle(ScarfColor.foregroundMuted)
            .padding(.horizontal, ScarfSpace.s4)

        SettingsSection(title: "Status", icon: "stethoscope") {
            HStack {
                Text("Actions")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .frame(width: 160, alignment: .trailing)
                Button("Check Status") {
                    statusOutput = viewModel.bitwardenStatus()
                    showStatus = true
                }
                .controlSize(.small)
                Spacer()
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 6)
            .background(.quaternary.opacity(0.3))

            if showStatus {
                Text(statusOutput.isEmpty ? "(no output)" : statusOutput)
                    .font(.system(.caption, design: .monospaced))
                    .textSelection(.enabled)
                    .padding(8)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(.quaternary.opacity(0.5))
            }
        }
    }
}
