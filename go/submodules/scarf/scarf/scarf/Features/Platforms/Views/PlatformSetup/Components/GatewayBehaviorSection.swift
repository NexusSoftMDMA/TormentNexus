import SwiftUI
import ScarfCore
import ScarfDesign

/// v0.13 Messaging Gateway behavior subsection composed into each per-
/// platform setup view (Slack, Mattermost, Telegram, WhatsApp, Matrix,
/// Google Chat). Owns its own `@State` view-model so the existing per-
/// platform VMs don't grow another set of fields.
///
/// **Capability gating.** Hides itself entirely on pre-v0.13 hosts
/// (returns `EmptyView` when none of the three v0.13 flags is on). Each
/// internal control gates on its own flag, so a host that gains, say,
/// `hasGatewayAllowlists` but not `hasGatewayBusyAckToggle` still gets
/// the allowlist editor with the toggles hidden.
struct GatewayBehaviorSection: View {
    let platform: String
    let capabilities: HermesCapabilities
    let context: ServerContext

    @State private var viewModel: GatewayBehaviorViewModel

    init(platform: String, capabilities: HermesCapabilities, context: ServerContext) {
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
        // Pre-v0.13 host — hide the entire subsection so the existing
        // platform forms look unchanged. Critical regression invariant
        // per WS-5 plan §"How to test" #1.
        if !capabilities.hasGatewayAllowlists
            && !capabilities.hasGatewayBusyAckToggle
            && !capabilities.hasGatewayRestartNotification {
            EmptyView()
        } else {
            content
        }
    }

    private var content: some View {
        VStack(alignment: .leading, spacing: ScarfSpace.s3) {
            SettingsSection(title: "Gateway behavior (v0.13+)", icon: "dot.radiowaves.left.and.right") {
                if capabilities.hasGatewayAllowlists,
                   let kind = viewModel.kind {
                    AllowlistEditor(
                        items: $viewModel.items,
                        kind: kind
                    )
                }
                if capabilities.hasGatewayBusyAckToggle {
                    ToggleRow(
                        label: "Send 'Agent is working…' ack",
                        isOn: viewModel.busyAckEnabled
                    ) { viewModel.busyAckEnabled = $0 }
                }
                if capabilities.hasGatewayRestartNotification {
                    ToggleRow(
                        label: "Post 'Gateway restarted' notice on boot",
                        isOn: viewModel.gatewayRestartNotification
                    ) { viewModel.gatewayRestartNotification = $0 }
                }
                // TTL field rides on either v0.13 toggle being available
                // — proxy gating per WS-1 Decision F. // TODO(WS-5-Q5)
                if capabilities.hasGatewayBusyAckToggle
                    || capabilities.hasGatewayRestartNotification {
                    StepperRow(
                        label: "Auto-delete slash-command notices (s)",
                        value: viewModel.slashCommandNoticeTTLSeconds,
                        range: 0...3600,
                        step: 5
                    ) { viewModel.slashCommandNoticeTTLSeconds = $0 }
                }
            }

            HStack {
                if let msg = viewModel.message {
                    Label(msg, systemImage: "checkmark.circle.fill")
                        .font(.caption)
                        .foregroundStyle(.green)
                }
                Spacer()
                Button("Save behavior") { viewModel.save() }
                    .buttonStyle(ScarfPrimaryButton())
                    .controlSize(.small)
                    .disabled(viewModel.isSaving)
            }
        }
        .onAppear { viewModel.load() }
    }
}
