import Foundation

enum ConnectionKind: String, Codable, CaseIterable, Identifiable {
    case ssh
    case local

    var id: String { rawValue }
}

struct ConnectionProfile: Codable, Identifiable, Equatable, Hashable {
    var id: UUID
    var kind: ConnectionKind
    var label: String
    var sshAlias: String
    var sshHost: String
    var sshPort: Int?
    var sshUser: String
    var hermesProfile: String?
    var customHermesHomePath: String?
    var createdAt: Date
    var updatedAt: Date
    var lastConnectedAt: Date?

    init(
        id: UUID = UUID(),
        kind: ConnectionKind = .ssh,
        label: String = "",
        sshAlias: String = "",
        sshHost: String = "",
        sshPort: Int? = nil,
        sshUser: String = "",
        hermesProfile: String? = nil,
        customHermesHomePath: String? = nil,
        createdAt: Date = Date(),
        updatedAt: Date = Date(),
        lastConnectedAt: Date? = nil
    ) {
        self.id = id
        self.kind = kind
        self.label = label
        self.sshAlias = sshAlias
        self.sshHost = sshHost
        self.sshPort = sshPort
        self.sshUser = sshUser
        self.hermesProfile = hermesProfile
        self.customHermesHomePath = customHermesHomePath
        self.createdAt = createdAt
        self.updatedAt = updatedAt
        self.lastConnectedAt = lastConnectedAt
    }

    private enum CodingKeys: String, CodingKey {
        case id
        case kind
        case label
        case sshAlias
        case sshHost
        case sshPort
        case sshUser
        case hermesProfile
        case customHermesHomePath
        case createdAt
        case updatedAt
        case lastConnectedAt
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(UUID.self, forKey: .id)
        kind = try container.decodeIfPresent(ConnectionKind.self, forKey: .kind) ?? .ssh
        label = try container.decode(String.self, forKey: .label)
        sshAlias = try container.decodeIfPresent(String.self, forKey: .sshAlias) ?? ""
        sshHost = try container.decodeIfPresent(String.self, forKey: .sshHost) ?? ""
        sshPort = try container.decodeIfPresent(Int.self, forKey: .sshPort)
        sshUser = try container.decodeIfPresent(String.self, forKey: .sshUser) ?? ""
        hermesProfile = try container.decodeIfPresent(String.self, forKey: .hermesProfile)
        customHermesHomePath = try container.decodeIfPresent(String.self, forKey: .customHermesHomePath)
        createdAt = try container.decode(Date.self, forKey: .createdAt)
        updatedAt = try container.decode(Date.self, forKey: .updatedAt)
        lastConnectedAt = try container.decodeIfPresent(Date.self, forKey: .lastConnectedAt)
    }

    var trimmedAlias: String? {
        let value = sshAlias.trimmingCharacters(in: .whitespacesAndNewlines)
        return value.isEmpty ? nil : value
    }

    var trimmedHost: String? {
        let value = sshHost.trimmingCharacters(in: .whitespacesAndNewlines)
        return value.isEmpty ? nil : value
    }

    var trimmedUser: String? {
        let value = sshUser.trimmingCharacters(in: .whitespacesAndNewlines)
        return value.isEmpty ? nil : value
    }

    var trimmedHermesProfile: String? {
        guard let hermesProfile else { return nil }
        let value = hermesProfile.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !value.isEmpty else { return nil }
        guard value.caseInsensitiveCompare("default") != .orderedSame else { return nil }
        return value
    }

    var trimmedCustomHermesHomePath: String? {
        guard let customHermesHomePath else { return nil }
        let value = customHermesHomePath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !value.isEmpty else { return nil }
        return value.normalizedCustomHermesHomePath
    }

    var usesCustomHermesHome: Bool {
        trimmedCustomHermesHomePath != nil
    }

    var resolvedHermesProfileName: String {
        if let trimmedCustomHermesHomePath {
            return trimmedCustomHermesHomePath.displayNameForCustomHermesHomePath
        }
        return trimmedHermesProfile ?? "default"
    }

    var usesDefaultHermesProfile: Bool {
        !usesCustomHermesHome && trimmedHermesProfile == nil
    }

    var cliHermesProfileName: String? {
        guard !usesCustomHermesHome else { return nil }
        return trimmedHermesProfile
    }

    var remoteHermesHomePath: String {
        if let trimmedCustomHermesHomePath {
            return trimmedCustomHermesHomePath
        }
        if let trimmedHermesProfile {
            return "~/.hermes/profiles/\(trimmedHermesProfile)"
        }

        return "~/.hermes"
    }

    var remoteSkillsPath: String {
        "\(remoteHermesHomePath)/skills"
    }

    var remoteCronJobsPath: String {
        "\(remoteHermesHomePath)/cron/jobs.json"
    }

    var remoteKanbanHomePath: String {
        "~/.hermes"
    }

    var remoteKanbanDatabasePath: String {
        "\(remoteKanbanHomePath)/kanban.db"
    }

    func remotePath(for trackedFile: RemoteTrackedFile) -> String {
        "\(remoteHermesHomePath)/\(trackedFile.relativePathFromHermesHome)"
    }

    func applyingHermesProfile(named profileName: String) -> ConnectionProfile {
        var copy = self
        copy.hermesProfile = profileName
        copy.customHermesHomePath = nil
        return copy.updated()
    }

    var remoteHermesHomeShellExpression: String {
        if let trimmedCustomHermesHomePath {
            return trimmedCustomHermesHomePath.customHermesHomeShellExpression
        }
        if let trimmedHermesProfile {
            let escapedProfile = trimmedHermesProfile.escapedForDoubleQuotedShellArgument
            return "$HOME/.hermes/profiles/\(escapedProfile)"
        }

        return "$HOME/.hermes"
    }

    var remoteHermesSearchPathShellExpression: String {
        let entries = [
            "\(remoteHermesHomeShellExpression)/hermes-agent/venv/bin",
            "$HOME/.local/bin",
            "$HOME/.hermes/hermes-agent/venv/bin",
            "$HOME/.cargo/bin",
            "/opt/homebrew/bin",
            "/usr/local/bin",
            "$PATH"
        ]

        var orderedEntries = [String]()
        var seen = Set<String>()
        for entry in entries where seen.insert(entry).inserted {
            orderedEntries.append(entry)
        }
        return orderedEntries.joined(separator: ":")
    }

    var remoteHermesCommandPrefix: String {
        """
        if [ -x "$HERMES_HOME/hermes-agent/venv/bin/hermes" ]; then HERMES_BIN="$HERMES_HOME/hermes-agent/venv/bin/hermes"; elif [ -x "$HOME/.local/bin/hermes" ]; then HERMES_BIN="$HOME/.local/bin/hermes"; elif [ -x "$HOME/.hermes/hermes-agent/venv/bin/hermes" ]; then HERMES_BIN="$HOME/.hermes/hermes-agent/venv/bin/hermes"; elif command -v hermes >/dev/null 2>&1; then HERMES_BIN="$(command -v hermes)"; else printf 'Hermes CLI not found.\\n' >&2; exit 127; fi; "$HERMES_BIN"
        """
    }

    func remoteHermesCommandLine(arguments: [String]) -> String {
        let quotedArguments = arguments.map(\.shellQuotedForTerminalCommand).joined(separator: " ")
        guard !quotedArguments.isEmpty else { return remoteHermesCommandPrefix }
        return "\(remoteHermesCommandPrefix) \(quotedArguments)"
    }

    func remoteServiceCommand(_ commandLine: String) -> String {
        let exportCommand = "export HERMES_HOME=\"\(remoteHermesHomeShellExpression)\""
        let pathCommand = "export PATH=\"\(remoteHermesSearchPathShellExpression)\""
        let escapedCommand = commandLine.escapedForDoubleQuotedShellArgument
        let innerCommand = "\(exportCommand); \(pathCommand); exec /bin/sh -c \"\(escapedCommand)\""
        return "exec /bin/sh -c \"\(innerCommand.escapedForOuterDoubleQuotedShellCommand)\""
    }

    var remoteShellBootstrapCommand: String {
        remoteShellBootstrapCommand()
    }

    func remoteShellBootstrapCommand(startupCommandLine: String? = nil) -> String {
        let exportCommand = "export HERMES_HOME=\"\(remoteHermesHomeShellExpression)\""
        let pathCommand = "export PATH=\"\(remoteHermesSearchPathShellExpression)\""

        let innerCommand: String
        if let startupCommandLine,
           !startupCommandLine.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            let startupSequence = """
\(startupCommandLine); hermes_bootstrap_exit_code=$?; if [ "$hermes_bootstrap_exit_code" -ne 0 ]; then printf '\\n[Hermes Desktop] Startup command exited with status %s.\\n' "$hermes_bootstrap_exit_code"; fi; exec "${SHELL:-/bin/zsh}" -l
"""
            let escapedStartupCommand = startupSequence.escapedForDoubleQuotedShellArgument
            innerCommand = "\(exportCommand); \(pathCommand); exec \"${SHELL:-/bin/zsh}\" -lc \"\(escapedStartupCommand)\""
        } else {
            innerCommand = "\(exportCommand); \(pathCommand); exec \"${SHELL:-/bin/zsh}\" -l"
        }

        return "exec /bin/sh -c \"\(innerCommand.escapedForOuterDoubleQuotedShellCommand)\""
    }

    var workspaceScopeFingerprint: String {
        if kind == .local {
            return "\(localConnectionIdentity)|\(remoteHermesHomePath)"
        }
        return [
            effectiveTarget,
            trimmedUser ?? "",
            resolvedPort.map(String.init) ?? "",
            remoteHermesHomePath
        ].joined(separator: "|")
    }

    var hostConnectionFingerprint: String {
        if kind == .local {
            return localConnectionIdentity
        }
        return [
            effectiveTarget,
            trimmedUser ?? "",
            resolvedPort.map(String.init) ?? ""
        ].joined(separator: "|")
    }

    var effectiveTarget: String {
        if kind == .local {
            return "This Mac"
        }
        return trimmedAlias ?? trimmedHost ?? ""
    }

    var localConnectionIdentity: String {
        "local:v1"
    }

    var usesAliasSourceOfTruth: Bool {
        trimmedAlias != nil && trimmedHost == nil
    }

    var resolvedPort: Int? {
        guard let sshPort, sshPort > 0 else { return nil }
        if usesAliasSourceOfTruth && sshPort == 22 {
            return nil
        }
        return sshPort
    }

    var displayDestination: String {
        if kind == .local {
            return "This Mac"
        }
        guard let user = trimmedUser else {
            return effectiveTarget
        }
        return "\(user)@\(effectiveTarget)"
    }

    var localizedDisplayDestination: String {
        kind == .local ? L10n.string("This Mac") : displayDestination
    }

    var isValid: Bool {
        validationError == nil
    }

    var validationError: String? {
        if label.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            return "Name is required."
        }

        if kind == .ssh, let sshValidationError {
            return sshValidationError
        }

        return hermesWorkspaceValidationError
    }

    var sshValidationError: String? {
        guard !effectiveTarget.isEmpty else {
            return "Add an SSH alias or host."
        }

        if let error = validateSSHArgument(trimmedAlias, fieldName: "SSH alias") {
            return error
        }

        if let error = validateSSHArgument(trimmedHost, fieldName: "Host") {
            return error
        }

        if let error = validateSSHArgument(trimmedUser, fieldName: "SSH user") {
            return error
        }

        return hermesWorkspaceValidationError
    }

    private var hermesWorkspaceValidationError: String? {
        if trimmedHermesProfile != nil && trimmedCustomHermesHomePath != nil {
            return "Choose either a Hermes profile or a custom Hermes home path."
        }

        if let trimmedHermesProfile {
            if trimmedHermesProfile.contains("/") || trimmedHermesProfile == "." || trimmedHermesProfile == ".." {
                return "Hermes profile must be a profile name, not a path."
            }
            if trimmedHermesProfile.containsControlCharacter {
                return "Hermes profile contains unsupported control characters."
            }
        }

        if let trimmedCustomHermesHomePath {
            if trimmedCustomHermesHomePath.containsControlCharacter {
                return "Custom Hermes home contains unsupported control characters."
            }
            if !trimmedCustomHermesHomePath.isValidCustomHermesHomePath {
                return "Custom Hermes home must start with `~/` or `/`."
            }
        }

        return nil
    }

    func updated() -> ConnectionProfile {
        var copy = self
        copy.label = label.trimmingCharacters(in: .whitespacesAndNewlines)
        copy.sshAlias = sshAlias.trimmingCharacters(in: .whitespacesAndNewlines)
        copy.sshHost = sshHost.trimmingCharacters(in: .whitespacesAndNewlines)
        copy.sshUser = sshUser.trimmingCharacters(in: .whitespacesAndNewlines)
        copy.hermesProfile = trimmedHermesProfile
        copy.customHermesHomePath = trimmedCustomHermesHomePath
        if let sshPort = sshPort, sshPort <= 0 {
            copy.sshPort = nil
        }
        copy.updatedAt = Date()
        return copy
    }
}

private extension String {
    var normalizedCustomHermesHomePath: String {
        if self == "/" || self == "~" {
            return self
        }
        if self == "~/" {
            return "~"
        }

        var trimmed = self
        while trimmed.count > 1, trimmed.hasSuffix("/") {
            trimmed.removeLast()
        }
        return trimmed
    }

    var isValidCustomHermesHomePath: Bool {
        self == "~" || hasPrefix("~/") || hasPrefix("/")
    }

    var customHermesHomeShellExpression: String {
        if self == "~" {
            return "$HOME"
        }
        if hasPrefix("~/") {
            let suffix = String(dropFirst(2)).escapedForDoubleQuotedShellArgument
            return "$HOME/\(suffix)"
        }
        return escapedForDoubleQuotedShellArgument
    }

    var displayNameForCustomHermesHomePath: String {
        let trimmed = normalizedCustomHermesHomePath
        if trimmed == "~" || trimmed == "/" {
            return trimmed
        }

        return trimmed.split(separator: "/").last.map(String.init) ?? trimmed
    }

    var escapedForDoubleQuotedShellArgument: String {
        replacingOccurrences(of: "\\", with: "\\\\")
            .replacingOccurrences(of: "\"", with: "\\\"")
            .replacingOccurrences(of: "$", with: "\\$")
            .replacingOccurrences(of: "`", with: "\\`")
    }

    var escapedForOuterDoubleQuotedShellCommand: String {
        replacingOccurrences(of: "\\", with: "\\\\")
            .replacingOccurrences(of: "\"", with: "\\\"")
            .replacingOccurrences(of: "$", with: "\\$")
            .replacingOccurrences(of: "`", with: "\\`")
    }

    var containsControlCharacter: Bool {
        unicodeScalars.contains { CharacterSet.controlCharacters.contains($0) }
    }
}

private func validateSSHArgument(_ value: String?, fieldName: String) -> String? {
    guard let value else { return nil }
    if value.hasPrefix("-") {
        return "\(fieldName) cannot start with a dash."
    }
    if value.unicodeScalars.contains(where: { CharacterSet.whitespacesAndNewlines.contains($0) || CharacterSet.controlCharacters.contains($0) }) {
        return "\(fieldName) cannot contain whitespace or control characters."
    }
    return nil
}
