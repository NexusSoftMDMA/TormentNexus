import Foundation
import Testing

@testable import HermesDesktop

struct LocalHermesIntegrationTests {
    @Test
    func realLocalHermesReadOnlyAndDisposableWriteSmokeTest() async throws {
        guard ProcessInfo.processInfo.environment["HERMES_RUN_LOCAL_INTEGRATION"] == "1" else {
            return
        }

        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let transport = SSHTransport(paths: makeTestAppPaths(root: root))
        let connection = ConnectionProfile(kind: .local, label: "This Mac").updated()
        let remoteService = RemoteHermesService(sshTransport: transport)
        let sessions = SessionBrowserService(sshTransport: transport)
        let files = FileEditorService(sshTransport: transport)
        let skills = SkillBrowserService(sshTransport: transport)
        let usage = UsageBrowserService(sshTransport: transport)
        let cron = CronBrowserService(sshTransport: transport)
        let kanban = KanbanBrowserService(sshTransport: transport)

        struct ConnectionProbe: Decodable {
            let ok: Bool
        }
        let probeScript = try RemotePythonScript.wrap(
            ["probe": true],
            body: #"print(json.dumps({"ok": True}))"#
        )
        let probe = try await transport.executeJSON(
            on: connection,
            pythonScript: probeScript,
            responseType: ConnectionProbe.self
        )
        #expect(probe.ok)

        let discovery = try await remoteService.discover(connection: connection)
        #expect(discovery.ok)
        _ = try await sessions.listSessions(connection: connection, offset: 0, limit: 1, query: "")
        _ = try await skills.listSkills(connection: connection)
        _ = try await usage.loadUsage(connection: connection, hintedSessionStore: discovery.sessionStore)
        _ = try await cron.listJobs(connection: connection)
        _ = try await kanban.loadBoards(connection: connection)

        let disposableURL = FileManager.default.temporaryDirectory
            .appendingPathComponent("hermes-desktop-local-verification-\(UUID().uuidString).txt")
        defer { try? FileManager.default.removeItem(at: disposableURL) }
        let marker = "Hermes Desktop local verification\n"
        _ = try await files.write(
            remotePath: disposableURL.path,
            content: marker,
            expectedContentHash: nil,
            connection: connection
        )
        let snapshot = try await files.read(remotePath: disposableURL.path, connection: connection)
        #expect(snapshot.content == marker)

        let cliResult = try await transport.execute(
            on: connection,
            remoteCommand: connection.remoteServiceCommand(
                connection.remoteHermesCommandLine(arguments: ["--version"])
            ),
            allocateTTY: false
        )
        try transport.validateSuccessfulExit(cliResult, for: connection)

        let terminalLaunch = transport.terminalLaunch(
            for: connection,
            startupCommandLine: "printf local-terminal-ready; exit 0"
        )
        let terminalResult = try runProcessLaunch(
            terminalLaunch,
            environment: [:]
        )
        #expect(terminalResult.exitCode == 0)
        #expect(terminalResult.stdout.contains("local-terminal-ready"))
    }
}

private func runProcessLaunch(
    _ launch: ProcessLaunch,
    environment: [String: String]
) throws -> LocalScriptResult {
    let process = Process()
    let stdoutPipe = Pipe()
    let stderrPipe = Pipe()
    process.executableURL = URL(fileURLWithPath: launch.executablePath)
    process.arguments = launch.arguments
    let launchEnvironment = Dictionary(
        uniqueKeysWithValues: launch.environment.compactMap { entry -> (String, String)? in
            guard let separator = entry.firstIndex(of: "=") else { return nil }
            return (
                String(entry[..<separator]),
                String(entry[entry.index(after: separator)...])
            )
        }
    )
    process.environment = launchEnvironment.merging(environment) { _, newValue in newValue }
    process.standardOutput = stdoutPipe
    process.standardError = stderrPipe
    try process.run()
    process.waitUntilExit()
    return LocalScriptResult(
        stdout: String(decoding: stdoutPipe.fileHandleForReading.readDataToEndOfFile(), as: UTF8.self),
        stderr: String(decoding: stderrPipe.fileHandleForReading.readDataToEndOfFile(), as: UTF8.self),
        exitCode: process.terminationStatus
    )
}
