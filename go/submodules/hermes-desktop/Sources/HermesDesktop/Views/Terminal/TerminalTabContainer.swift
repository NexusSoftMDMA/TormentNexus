import SwiftUI

struct TerminalTabContainer: View {
    @ObservedObject var session: TerminalSession
    let appearance: TerminalThemeAppearance
    let fontSize: Double
    let fontFamily: TerminalFontFamilyPreference
    let isActive: Bool
    let activeWorkspaceScopeFingerprint: String?
    let backgroundImageActive: Bool

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    HStack(spacing: 8) {
                        Text(session.connection.resolvedHermesProfileName)
                            .font(.headline)

                        if isDifferentFromActiveWorkspace {
                            HermesBadge(text: "Other Profile", tint: .orange)
                        }
                    }

                    Text(session.connection.localizedDisplayDestination)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                if let currentDirectory = session.currentDirectory {
                    Text(currentDirectory)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }

                Spacer()

                if let exitCode = session.exitCode {
                    Text(exitCode == 0 ? L10n.string("Shell exited") : L10n.string("Connection ended (%@)", "\(exitCode)"))
                        .font(.caption)
                        .foregroundStyle(exitCode == 0 ? Color.secondary : Color.orange)

                    Button(L10n.string("Reconnect")) {
                        session.requestReconnect()
                    }
                    .buttonStyle(.bordered)
                }
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 10)
            .background(backgroundImageActive ? AnyShapeStyle(.thinMaterial) : AnyShapeStyle(Color.secondary.opacity(0.08)))

            SwiftTermTerminalView(
                session: session,
                appearance: appearance,
                fontSize: fontSize,
                fontFamily: fontFamily,
                isActive: isActive,
                backgroundImageActive: backgroundImageActive
            )
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .background(backgroundImageActive ? Color.clear : appearance.backgroundColor.swiftUIColor)
                .clipped()
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }

    private var isDifferentFromActiveWorkspace: Bool {
        guard let activeWorkspaceScopeFingerprint else { return false }
        return activeWorkspaceScopeFingerprint != session.connection.workspaceScopeFingerprint
    }
}
