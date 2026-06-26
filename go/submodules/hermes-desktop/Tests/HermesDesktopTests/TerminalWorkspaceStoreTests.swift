import Foundation
import Testing

@testable import HermesDesktop

@MainActor
struct TerminalWorkspaceStoreTests {
    @Test
    func commandTabsAlwaysCreateFreshTerminalTabs() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let paths = makeTestAppPaths(root: root)
        let transport = SSHTransport(paths: paths)
        let store = TerminalWorkspaceStore(
            sshTransport: transport,
            workflowLaunchDiagnostics: WorkflowLaunchDiagnostics(
                logFileURL: paths.applicationSupportURL
                    .appendingPathComponent("Diagnostics", isDirectory: true)
                    .appendingPathComponent("workflow-launch-latest.log")
            )
        )
        let connection = ConnectionProfile(
            label: "Pi",
            sshAlias: "hermes-pi",
            hermesProfile: "archivio"
        ).updated()

        store.ensureInitialTab(for: connection)
        #expect(store.tabs.count == 1)

        let seededTab = store.addCommandTab(
            for: connection,
            commandLine: "hermes --skills apple/apple-notes chat",
            initialInput: "hello"
        )
        _ = store.addCommandTab(
            for: connection,
            commandLine: "hermes --skills apple/apple-reminders chat",
            initialInput: "hello again"
        )

        #expect(store.tabs.count == 3)
        #expect(store.selectedTabID == store.tabs.last?.id)
        #expect(seededTab.session.startupInput == "hello")
    }
}
