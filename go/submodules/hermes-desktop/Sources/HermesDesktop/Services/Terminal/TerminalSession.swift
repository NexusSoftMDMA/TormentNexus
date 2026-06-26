import Foundation

@MainActor
final class TerminalSession: ObservableObject, @unchecked Sendable {
    let connection: ConnectionProfile
    let processLaunch: ProcessLaunch
    let startupInput: String?
    let workflowLaunchDiagnosticsContext: WorkflowLaunchDiagnosticsContext?
    private let workflowLaunchDiagnostics: WorkflowLaunchDiagnostics
    private let viewHost = TerminalViewHost()

    @Published var terminalTitle: String
    @Published var currentDirectory: String?
    @Published var exitCode: Int32?
    @Published var didStart = false
    @Published private(set) var launchToken = UUID()
    @Published private(set) var isRunning = false

    init(
        connection: ConnectionProfile,
        sshTransport: SSHTransport,
        startupCommandLine: String? = nil,
        startupInput: String? = nil,
        workflowLaunchDiagnostics: WorkflowLaunchDiagnostics,
        workflowLaunchDiagnosticsContext: WorkflowLaunchDiagnosticsContext? = nil
    ) {
        self.connection = connection
        self.startupInput = startupInput
        self.workflowLaunchDiagnostics = workflowLaunchDiagnostics
        self.workflowLaunchDiagnosticsContext = workflowLaunchDiagnosticsContext
        self.processLaunch = sshTransport.terminalLaunch(
            for: connection,
            startupCommandLine: startupCommandLine
        )
        self.terminalTitle = "\(connection.label) · \(connection.resolvedHermesProfileName)"
        viewHost.setEventHandlers(
            onProcessStart: { [weak self] in
                self?.markStarted()
            },
            onTitleChange: { [weak self] title in
                self?.updateTitle(title)
            },
            onDirectoryChange: { [weak self] directory in
                self?.currentDirectory = directory
            },
            onProcessExit: { [weak self] exitCode in
                self?.markExited(exitCode)
            }
        )
    }

    deinit {
        viewHost.terminate()
    }

    func markStarted() {
        didStart = true
        isRunning = true
        exitCode = nil
    }

    func updateTitle(_ title: String) {
        guard !title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { return }
        terminalTitle = title
    }

    func markExited(_ code: Int32?) {
        isRunning = false
        exitCode = code
        if let workflowLaunchDiagnosticsContext {
            Task {
                await workflowLaunchDiagnostics.recordTerminalProcessExited(
                    workflowLaunchDiagnosticsContext,
                    exitCode: code
                )
            }
        }
    }

    func requestReconnect() {
        currentDirectory = nil
        exitCode = nil
        launchToken = UUID()
    }

    func mount(
        in container: TerminalMountContainerView,
        appearance: TerminalThemeAppearance,
        fontSize: Double,
        fontFamily: TerminalFontFamilyPreference,
        isActive: Bool,
        backgroundImageActive: Bool
    ) {
        viewHost.mount(
            in: container,
            request: TerminalLaunchRequest(
                processLaunch: processLaunch,
                launchToken: launchToken,
                initialInput: startupInput,
                workflowLaunchDiagnostics: workflowLaunchDiagnostics,
                workflowLaunchDiagnosticsContext: workflowLaunchDiagnosticsContext
            ),
            appearance: appearance,
            fontSize: fontSize,
            fontFamily: fontFamily,
            isActive: isActive,
            backgroundImageActive: backgroundImageActive
        )
    }

    func unmount(from container: TerminalMountContainerView) {
        viewHost.unmount(from: container)
    }

    func stop() {
        viewHost.terminate()
        isRunning = false
        currentDirectory = nil
    }
}
