import Foundation

struct TerminalWorkspaceContext {
    let activeConnection: ConnectionProfile?
    let activeWorkspaceScopeFingerprint: String?
    let isTerminalSectionActive: Bool
    let terminalTheme: TerminalThemePreference
    let terminalFontSize: Double
    let terminalFontFamily: TerminalFontFamilyPreference
    let backgroundImageActive: Bool
}
