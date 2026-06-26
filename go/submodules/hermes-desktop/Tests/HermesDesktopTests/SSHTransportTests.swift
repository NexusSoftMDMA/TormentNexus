import Foundation
import Testing

@testable import HermesDesktop

struct SSHTransportTests {
    @Test
    func localExecuteLaunchesNoninteractiveShellAndForwardsStdin() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let runner = RecordingSSHProcessRunner(
            result: SSHCommandResult(stdout: "ok", stderr: "", exitCode: 0)
        )
        let transport = SSHTransport(
            paths: makeTestAppPaths(root: root),
            processRunner: runner
        )
        let connection = ConnectionProfile(kind: .local, label: "This Mac").updated()
        let input = Data("payload".utf8)
        let command = connection.remoteServiceCommand("python3 -")

        _ = try await transport.execute(
            on: connection,
            remoteCommand: command,
            standardInput: input,
            allocateTTY: false
        )

        let invocation = try #require(await runner.lastInvocation)
        #expect(invocation.executableURL.path == "/bin/sh")
        #expect(invocation.arguments == ["-c", command])
        #expect(invocation.standardInput == input)
    }

    @Test
    func terminalLaunchSupportsLocalNamedAndCustomProfilesWithoutChangingSSHLaunch() {
        let transport = SSHTransport(paths: AppPaths())
        let ssh = ConnectionProfile(label: "Prod", sshHost: "example.com").updated()
        let localNamed = ConnectionProfile(
            kind: .local,
            label: "Research",
            hermesProfile: "research"
        ).updated()
        let localCustom = ConnectionProfile(
            kind: .local,
            label: "Custom",
            customHermesHomePath: "~/.hermes-work"
        ).updated()

        let sshLaunch = transport.terminalLaunch(for: ssh)
        #expect(sshLaunch.executablePath == "/usr/bin/ssh")
        #expect(sshLaunch.arguments == transport.shellArguments(for: ssh))
        #expect(sshLaunch.executableName == "ssh")
        #expect(sshLaunch.environment == ["TERM=xterm-256color", "COLORTERM=truecolor"])

        let namedLaunch = transport.terminalLaunch(for: localNamed, startupCommandLine: "hermes --version")
        #expect(namedLaunch.executablePath == "/bin/sh")
        #expect(namedLaunch.arguments == ["-c", localNamed.remoteShellBootstrapCommand(startupCommandLine: "hermes --version")])
        #expect(namedLaunch.arguments[1].contains("$HOME/.hermes/profiles/research"))

        let customLaunch = transport.terminalLaunch(for: localCustom)
        #expect(customLaunch.arguments[1].contains("$HOME/.hermes-work"))
    }

    @Test
    func sshFailureMessagesAndValidationPathRemainExact() throws {
        let transport = SSHTransport(paths: AppPaths())
        let connection = ConnectionProfile(label: "Prod", sshHost: "example.com").updated()
        let result = SSHCommandResult(
            stdout: "",
            stderr: "Permission denied (publickey).",
            exitCode: 255
        )
        let expected = "SSH authentication failed. Verify the key, SSH agent, and user for this SSH target."

        #expect(transport.describeRemoteFailure(
            stdout: result.stdout,
            stderr: result.stderr,
            exitCode: result.exitCode,
            connection: connection
        ) == expected)

        do {
            try transport.validateSuccessfulExit(result, for: connection)
            Issue.record("Expected SSH validation failure")
        } catch let error as SSHTransportError {
            #expect(error.errorDescription == expected)
        }
    }

    @Test
    func serviceArgumentsUseControlSocketAndExplicitDestination() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let transport = SSHTransport(paths: makeTestAppPaths(root: root))
        let connection = ConnectionProfile(
            label: "Prod",
            sshAlias: "prod-box",
            sshPort: 2222,
            sshUser: "alice"
        ).updated()

        let arguments = transport.serviceArguments(
            for: connection,
            remoteCommand: "python3 -"
        )

        #expect(arguments.contains("-T"))
        #expect(arguments.contains("-p"))
        #expect(arguments.contains("2222"))
        #expect(arguments.contains("--"))
        #expect(arguments.contains("alice@prod-box"))
        #expect(arguments.contains("python3 -"))
        #expect(arguments.contains("ControlMaster=auto"))
        #expect(arguments.contains("ControlPersist=300"))
        #expect(arguments.contains(where: { $0.hasPrefix("ControlPath=") }))
    }

    @Test
    func shellArgumentsKeepInteractiveTabsOffSharedControlMaster() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let transport = SSHTransport(paths: makeTestAppPaths(root: root))
        let connection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com",
            sshUser: "alice",
            hermesProfile: "research"
        ).updated()

        let arguments = transport.shellArguments(for: connection)

        #expect(arguments.contains("-tt"))
        #expect(arguments.contains("ControlMaster=no"))
        #expect(arguments.contains("-S"))
        #expect(arguments.contains("none"))
        #expect(arguments.last == connection.remoteShellBootstrapCommand)
    }

    @Test
    func executeUsesInjectedProcessRunner() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let runner = RecordingSSHProcessRunner(
            result: SSHCommandResult(stdout: "ok", stderr: "", exitCode: 0)
        )
        let transport = SSHTransport(
            paths: makeTestAppPaths(root: root),
            processRunner: runner
        )
        let connection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com",
            sshUser: "alice"
        ).updated()

        let stdin = Data("payload".utf8)
        let result = try await transport.execute(
            on: connection,
            remoteCommand: "printf ok",
            standardInput: stdin,
            allocateTTY: false
        )

        let invocation = try #require(await runner.lastInvocation)
        #expect(invocation.executableURL.path == "/usr/bin/ssh")
        #expect(invocation.arguments.contains("alice@example.com"))
        #expect(invocation.arguments.contains("printf ok"))
        #expect(invocation.standardInput == stdin)
        #expect(result.stdout == "ok")
    }

    @Test
    func executeRetriesReachabilityFailuresWithoutControlMaster() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let runner = SequenceSSHProcessRunner(results: [
            SSHCommandResult(stdout: "", stderr: "ssh: connect to host example.com port 22: No route to host", exitCode: 255),
            SSHCommandResult(stdout: "ok", stderr: "", exitCode: 0)
        ])
        let transport = SSHTransport(
            paths: makeTestAppPaths(root: root),
            processRunner: runner
        )
        let connection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com",
            sshUser: "alice"
        ).updated()

        let result = try await transport.execute(
            on: connection,
            remoteCommand: "printf ok",
            allocateTTY: false
        )

        let invocations = await runner.invocations
        #expect(result.stdout == "ok")
        #expect(invocations.count == 2)
        #expect(invocations[0].arguments.contains("ControlMaster=auto"))
        #expect(invocations[1].arguments.contains("ControlMaster=no"))
        #expect(invocations[1].arguments.contains("-S"))
        #expect(invocations[1].arguments.contains("none"))
    }

    @Test
    func remoteFailureMentionsNonInteractivePythonPath() {
        let transport = SSHTransport(paths: AppPaths())
        let connection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com"
        ).updated()

        let message = transport.describeRemoteFailure(
            stdout: "",
            stderr: "zsh:1: command not found: python3",
            exitCode: 127,
            connection: connection
        )

        #expect(message.contains("non-interactive SSH shell PATH"))
        #expect(message.contains("python3"))
    }

    @Test
    func remoteFailureMentionsLocalNetworkPermissionForUnreachableLANHosts() {
        let transport = SSHTransport(paths: AppPaths())
        let connection = ConnectionProfile(
            label: "Home Pi",
            sshAlias: "hermes-home"
        ).updated()

        let message = transport.describeRemoteFailure(
            stdout: "",
            stderr: "ssh: connect to host 192.168.1.17 port 22: No route to host",
            exitCode: 255,
            connection: connection
        )

        #expect(message.contains("Local Network"))
        #expect(message.contains("Privacy & Security"))
    }

    @Test
    func executeJSONFlagsShellStartupNoiseBeforeJSON() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let runner = RecordingSSHProcessRunner(
            result: SSHCommandResult(
                stdout: "Welcome to staging\n{\"ok\": true}",
                stderr: "",
                exitCode: 0
            )
        )
        let transport = SSHTransport(
            paths: makeTestAppPaths(root: root),
            processRunner: runner
        )
        let connection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com"
        ).updated()

        do {
            let _: OKResponse = try await transport.executeJSON(
                on: connection,
                pythonScript: "print('noop')",
                responseType: OKResponse.self
            )
            Issue.record("Expected JSON decoding to fail")
        } catch let error as SSHTransportError {
            let invocation = try #require(await runner.lastInvocation)
            #expect(invocation.arguments.contains(connection.remoteServiceCommand("python3 -")))

            guard case .invalidResponse(let message) = error else {
                Issue.record("Expected invalidResponse, got \(error)")
                return
            }
            #expect(message.contains("shell startup file printed text"))
            #expect(message.contains("Welcome to staging"))
        }
    }
}

private struct OKResponse: Decodable {
    let ok: Bool
}

private struct SSHProcessInvocation {
    let executableURL: URL
    let arguments: [String]
    let standardInput: Data?
}

private actor RecordingSSHProcessRunner: SSHProcessRunning {
    private let result: SSHCommandResult
    private(set) var lastInvocation: SSHProcessInvocation?

    init(result: SSHCommandResult) {
        self.result = result
    }

    func run(
        executableURL: URL,
        arguments: [String],
        standardInput: Data?
    ) async throws -> SSHCommandResult {
        lastInvocation = SSHProcessInvocation(
            executableURL: executableURL,
            arguments: arguments,
            standardInput: standardInput
        )
        return result
    }
}

private actor SequenceSSHProcessRunner: SSHProcessRunning {
    private var results: [SSHCommandResult]
    private(set) var invocations: [SSHProcessInvocation] = []

    init(results: [SSHCommandResult]) {
        self.results = results
    }

    func run(
        executableURL: URL,
        arguments: [String],
        standardInput: Data?
    ) async throws -> SSHCommandResult {
        invocations.append(SSHProcessInvocation(
            executableURL: executableURL,
            arguments: arguments,
            standardInput: standardInput
        ))
        guard !results.isEmpty else {
            return SSHCommandResult(stdout: "", stderr: "unexpected invocation", exitCode: 1)
        }
        return results.removeFirst()
    }
}
