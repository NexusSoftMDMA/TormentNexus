import SwiftUI

struct ConnectionEditorSheet: View {
    @Environment(\.dismiss) private var dismiss

    private enum Field: Hashable {
        case label
        case alias
        case host
        case user
        case port
        case hermesProfile
        case customHermesHome
    }

    @State private var draft: ConnectionProfile
    @State private var portText: String
    @State private var showsCustomHermesHomeOptions: Bool
    @FocusState private var focusedField: Field?
    let isEditing: Bool
    let onSave: (ConnectionProfile) -> Void

    init(connection: ConnectionProfile, isEditing: Bool, onSave: @escaping (ConnectionProfile) -> Void) {
        _draft = State(initialValue: connection)
        _portText = State(initialValue: connection.sshPort.map(String.init) ?? "")
        _showsCustomHermesHomeOptions = State(initialValue: connection.customHermesHomePath?.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false)
        self.isEditing = isEditing
        self.onSave = onSave
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 18) {
                    HermesPageHeader(
                        title: isEditing ? "Edit Connection" : "New Connection",
                        subtitle: "Connect directly to Hermes on this Mac or use an SSH host."
                    )

                    HermesSurfacePanel(
                        title: "Connection Details",
                        subtitle: "Choose where Hermes runs, then give the connection a clear name."
                    ) {
                        VStack(alignment: .leading, spacing: 14) {
                            Picker(L10n.string("Connection Type"), selection: $draft.kind) {
                                Text(L10n.string("This Mac")).tag(ConnectionKind.local)
                                Text(L10n.string("SSH Host")).tag(ConnectionKind.ssh)
                            }
                            .pickerStyle(.segmented)

                            EditorField(label: "Name") {
                                TextField(L10n.string("Home Pi, Studio Mac, Prod VPS"), text: $draft.label)
                                    .focused($focusedField, equals: .label)
                                    .textFieldStyle(.roundedBorder)
                            }

                            if draft.kind == .ssh {
                                EditorField(label: "SSH alias") {
                                    TextField(L10n.string("hermes-home"), text: $draft.sshAlias)
                                        .focused($focusedField, equals: .alias)
                                        .textFieldStyle(.roundedBorder)
                                }

                                EditorField(label: "Host or IP address") {
                                    TextField(L10n.string("mac-studio.local, 203.0.113.10, localhost"), text: $draft.sshHost)
                                        .focused($focusedField, equals: .host)
                                        .textFieldStyle(.roundedBorder)
                                }

                                ViewThatFits(in: .horizontal) {
                                    HStack(alignment: .top, spacing: 14) {
                                        EditorField(label: "SSH user") {
                                            TextField("alex", text: $draft.sshUser)
                                                .focused($focusedField, equals: .user)
                                                .textFieldStyle(.roundedBorder)
                                        }

                                        EditorField(label: "SSH port") {
                                            TextField("22", text: $portText)
                                                .focused($focusedField, equals: .port)
                                                .textFieldStyle(.roundedBorder)
                                        }
                                    }

                                    VStack(alignment: .leading, spacing: 14) {
                                        EditorField(label: "SSH user") {
                                            TextField("alex", text: $draft.sshUser)
                                                .focused($focusedField, equals: .user)
                                                .textFieldStyle(.roundedBorder)
                                        }

                                        EditorField(label: "SSH port") {
                                            TextField("22", text: $portText)
                                                .focused($focusedField, equals: .port)
                                                .textFieldStyle(.roundedBorder)
                                        }
                                    }
                                }
                            } else {
                                HermesInsetSurface {
                                    Text(L10n.string("Local changes use your current macOS account and affect this Mac’s real Hermes files, profiles, Cron jobs, Kanban data, and sessions."))
                                        .font(.subheadline)
                                        .foregroundStyle(.orange)
                                        .fixedSize(horizontal: false, vertical: true)
                                }
                            }

                            EditorField(label: "Hermes profile") {
                                TextField(L10n.string("default or researcher"), text: hermesProfileBinding)
                                    .focused($focusedField, equals: .hermesProfile)
                                    .textFieldStyle(.roundedBorder)
                            }

                            DisclosureGroup(
                                isExpanded: $showsCustomHermesHomeOptions
                            ) {
                                VStack(alignment: .leading, spacing: 10) {
                                    EditorField(label: draft.kind == .local ? "Hermes home path" : "Remote Hermes home path") {
                                        TextField(L10n.string("~/.hermes-work or /opt/data"), text: customHermesHomeBinding)
                                            .focused($focusedField, equals: .customHermesHome)
                                            .textFieldStyle(.roundedBorder)
                                    }

                                    Text(L10n.string(draft.kind == .local
                                        ? "Only use this if Hermes lives outside the standard `~/.hermes` or `~/.hermes/profiles/<name>` layout on this Mac."
                                        : "Only use this if Hermes lives outside the standard `~/.hermes` or `~/.hermes/profiles/<name>` layout on the remote host."))
                                        .font(.subheadline)
                                        .foregroundStyle(.secondary)
                                        .fixedSize(horizontal: false, vertical: true)

                                    Text(L10n.string("Leave this empty for the normal setup. When set, Hermes Desktop uses this path as `HERMES_HOME` for Sessions, Files, Skills, Usage, Cron, chat, and Terminal."))
                                        .font(.subheadline)
                                        .foregroundStyle(.secondary)
                                        .fixedSize(horizontal: false, vertical: true)
                                }
                                .padding(.top, 8)
                            } label: {
                                Text(L10n.string("Custom Hermes home (advanced)"))
                                    .font(.headline)
                            }

                            if let validationMessage {
                                HermesValidationMessage(text: validationMessage)
                            }
                        }
                    }

                    if draft.kind == .ssh {
                        HermesSurfacePanel(
                        title: "Connection Tips",
                        subtitle: "Use the same SSH setup that works in Terminal."
                    ) {
                        VStack(alignment: .leading, spacing: 10) {
                            ConnectionHintRow(
                                title: "SSH alias",
                                detail: "Use an alias when you have one."
                            )

                            ConnectionHintRow(
                                title: "Same Mac",
                                detail: "Use localhost or a local SSH alias."
                            )

                            ConnectionHintRow(
                                title: "Authentication",
                                detail: "SSH should connect without prompts."
                            )

                            ConnectionHintRow(
                                title: "Hermes profile",
                                detail: "Leave empty for the default profile."
                            )

                            if draft.trimmedAlias != nil && draft.trimmedHost != nil {
                                HermesInsetSurface {
                                    Text(L10n.string("SSH alias is used before Host."))
                                        .font(.subheadline)
                                        .foregroundStyle(.orange)
                                        .fixedSize(horizontal: false, vertical: true)
                                }
                            }
                        }
                        }
                    }

                    HermesSurfacePanel(
                        title: "Examples",
                        subtitle: "Common setups."
                    ) {
                        VStack(alignment: .leading, spacing: 12) {
                            ExampleValueRow(label: "Raspberry Pi", value: "Alias `hermes-home` or host `raspberrypi.local`")
                            ExampleValueRow(label: "Remote Mac", value: "Host `mac-studio.local`")
                            ExampleValueRow(label: "VPS", value: "Host `vps.example.com` or `203.0.113.10`")
                            ExampleValueRow(label: "Same Mac", value: "Host `localhost` or a local SSH alias")
                        }
                    }
                }
                .frame(maxWidth: 760, alignment: .leading)
                .padding(.horizontal, 24)
                .padding(.vertical, 20)
                .frame(maxWidth: .infinity, alignment: .top)
            }
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(L10n.string("Cancel")) {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(L10n.string("Save")) {
                        var updatedDraft = draft
                        updatedDraft.sshPort = parsedPort
                        onSave(updatedDraft)
                        dismiss()
                    }
                    .disabled(!isDraftValid)
                }
            }
        }
        .frame(minWidth: 620, minHeight: 560)
        .onAppear {
            NSApp.activate(ignoringOtherApps: true)
            DispatchQueue.main.async {
                focusedField = .label
            }
        }
    }

    private var parsedPort: Int? {
        let trimmed = portText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }
        guard let value = Int(trimmed), (1...65_535).contains(value) else { return nil }
        return value
    }

    private var isDraftValid: Bool {
        validationMessage == nil
    }

    private var validationMessage: String? {
        let hasValidPort = portText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || parsedPort != nil
        var candidate = draft
        candidate.sshPort = parsedPort

        if candidate.label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            return "Name is required."
        }

        if candidate.kind == .ssh && candidate.trimmedAlias == nil && candidate.trimmedHost == nil {
            return "Add an SSH alias or host."
        }

        if candidate.kind == .ssh && !hasValidPort {
            return "Enter a valid SSH port from 1 to 65535."
        }

        return candidate.validationError
    }

    private var hermesProfileBinding: Binding<String> {
        Binding {
            draft.hermesProfile ?? ""
        } set: { newValue in
            draft.hermesProfile = newValue
        }
    }

    private var customHermesHomeBinding: Binding<String> {
        Binding {
            draft.customHermesHomePath ?? ""
        } set: { newValue in
            draft.customHermesHomePath = newValue
        }
    }
}

private struct EditorField<Content: View>: View {
    let label: String
    let content: Content

    init(label: String, @ViewBuilder content: () -> Content) {
        self.label = label
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(L10n.string(label))
                .font(.caption)
                .foregroundStyle(.secondary)

            content
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

private struct ConnectionHintRow: View {
    let title: String
    let detail: String

    var body: some View {
        HermesInsetSurface {
            VStack(alignment: .leading, spacing: 6) {
                Text(L10n.string(title))
                    .font(.headline)

                Text(L10n.string(detail))
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }
}

private struct ExampleValueRow: View {
    let label: String
    let value: String

    var body: some View {
        HermesInsetSurface {
            VStack(alignment: .leading, spacing: 4) {
                Text(L10n.string(label))
                    .font(.headline)

                Text(value)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
        }
    }
}
