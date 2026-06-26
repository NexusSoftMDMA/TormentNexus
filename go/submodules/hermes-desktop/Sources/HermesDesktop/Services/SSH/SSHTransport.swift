import Foundation

struct SSHCommandResult {
    let stdout: String
    let stderr: String
    let exitCode: Int32
}

struct ProcessLaunch: Equatable {
    let executablePath: String
    let arguments: [String]
    let executableName: String
    let environment: [String]
}

enum SSHTransportError: LocalizedError, Equatable {
    case invalidConnection(String)
    case launchFailure(String)
    case localFailure(String)
    case remoteFailure(String)
    case invalidResponse(String)

    var errorDescription: String? {
        switch self {
        case .invalidConnection(let message),
             .launchFailure(let message),
             .localFailure(let message),
             .remoteFailure(let message),
             .invalidResponse(let message):
            message
        }
    }
}

protocol SSHProcessRunning: Sendable {
    func run(
        executableURL: URL,
        arguments: [String],
        standardInput: Data?
    ) async throws -> SSHCommandResult
}

final class SSHTransport: @unchecked Sendable {
    private let paths: AppPaths
    private let processRunner: any SSHProcessRunning

    private enum ConnectionPurpose {
        case service
        case serviceWithoutMultiplexing
        case terminalShell
    }

    init(
        paths: AppPaths,
        processRunner: any SSHProcessRunning = FoundationSSHProcessRunner()
    ) {
        self.paths = paths
        self.processRunner = processRunner
    }

    func execute(
        on connection: ConnectionProfile,
        remoteCommand: String,
        standardInput: Data? = nil,
        allocateTTY: Bool
    ) async throws -> SSHCommandResult {
        if connection.kind == .local {
            if let validationError = connection.validationError {
                throw SSHTransportError.invalidConnection(validationError)
            }
            return try await processRunner.run(
                executableURL: URL(fileURLWithPath: "/bin/sh"),
                arguments: ["-c", remoteCommand],
                standardInput: standardInput
            )
        }

        guard !connection.effectiveTarget.isEmpty else {
            throw SSHTransportError.invalidConnection("The SSH target is empty.")
        }
        if let validationError = connection.sshValidationError {
            throw SSHTransportError.invalidConnection(validationError)
        }

        let arguments = sshArguments(
            for: connection,
            remoteCommand: remoteCommand,
            allocateTTY: allocateTTY,
            purpose: .service
        )

        let result = try await processRunner.run(
            executableURL: URL(fileURLWithPath: "/usr/bin/ssh"),
            arguments: arguments,
            standardInput: standardInput
        )

        guard result.exitCode != 0,
              shouldRetryWithoutMultiplexing(result) else {
            return result
        }

        let retryArguments = sshArguments(
            for: connection,
            remoteCommand: remoteCommand,
            allocateTTY: allocateTTY,
            purpose: .serviceWithoutMultiplexing
        )

        return try await processRunner.run(
            executableURL: URL(fileURLWithPath: "/usr/bin/ssh"),
            arguments: retryArguments,
            standardInput: standardInput
        )
    }

    func executeJSON<Response: Decodable>(
        on connection: ConnectionProfile,
        pythonScript: String,
        responseType: Response.Type
    ) async throws -> Response {
        let result = try await execute(
            on: connection,
            remoteCommand: connection.remoteServiceCommand("python3 -"),
            standardInput: Data(pythonScript.utf8),
            allocateTTY: false
        )

        try validateSuccessfulExit(result, for: connection)

        guard let data = result.stdout.data(using: .utf8) else {
            throw SSHTransportError.invalidResponse(
                connection.kind == .local
                    ? "Local command output was not valid UTF-8."
                    : "Remote output was not valid UTF-8."
            )
        }

        do {
            return try JSONDecoder().decode(Response.self, from: data)
        } catch {
            throw SSHTransportError.invalidResponse(
                formattedInvalidJSONResponse(
                    stdout: result.stdout,
                    stderr: result.stderr,
                    decodingError: error,
                    isLocal: connection.kind == .local
                )
            )
        }
    }

    func shellArguments(for connection: ConnectionProfile, startupCommandLine: String? = nil) -> [String] {
        sshArguments(
            for: connection,
            remoteCommand: connection.remoteShellBootstrapCommand(startupCommandLine: startupCommandLine),
            allocateTTY: true,
            purpose: .terminalShell
        )
    }

    func terminalLaunch(for connection: ConnectionProfile, startupCommandLine: String? = nil) -> ProcessLaunch {
        if connection.kind == .local {
            return ProcessLaunch(
                executablePath: "/bin/sh",
                arguments: ["-c", connection.remoteShellBootstrapCommand(startupCommandLine: startupCommandLine)],
                executableName: "sh",
                environment: localTerminalEnvironment()
            )
        }

        return ProcessLaunch(
            executablePath: "/usr/bin/ssh",
            arguments: shellArguments(for: connection, startupCommandLine: startupCommandLine),
            executableName: "ssh",
            environment: [
                "TERM=xterm-256color",
                "COLORTERM=truecolor"
            ]
        )
    }

    private func localTerminalEnvironment() -> [String] {
        var environment = ProcessInfo.processInfo.environment
        func nonempty(_ value: String?) -> String? {
            guard let value, !value.isEmpty else { return nil }
            return value
        }
        let home = FileManager.default.homeDirectoryForCurrentUser.path
        environment["HOME"] = nonempty(environment["HOME"]) ?? home
        environment["USER"] = nonempty(environment["USER"]) ?? NSUserName()
        environment["LOGNAME"] = nonempty(environment["LOGNAME"]) ?? environment["USER"]
        environment["SHELL"] = nonempty(environment["SHELL"]) ?? "/bin/zsh"
        environment["PATH"] = nonempty(environment["PATH"])
            ?? "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
        environment["TERM"] = "xterm-256color"
        environment["COLORTERM"] = "truecolor"
        return environment
            .map { "\($0.key)=\($0.value)" }
            .sorted()
    }

    func serviceArguments(
        for connection: ConnectionProfile,
        remoteCommand: String,
        allocateTTY: Bool = false
    ) -> [String] {
        sshArguments(
            for: connection,
            remoteCommand: remoteCommand,
            allocateTTY: allocateTTY,
            purpose: .service
        )
    }

    func validateSuccessfulExit(_ result: SSHCommandResult, for connection: ConnectionProfile? = nil) throws {
        guard result.exitCode == 0 else {
            if connection?.kind == .local {
                throw SSHTransportError.remoteFailure(
                    describeLocalFailure(
                        stdout: result.stdout,
                        stderr: result.stderr,
                        exitCode: result.exitCode
                    )
                )
            }
            throw SSHTransportError.remoteFailure(
                describeRemoteFailure(
                    stdout: result.stdout,
                    stderr: result.stderr,
                    exitCode: result.exitCode,
                    connection: connection
                )
            )
        }
    }

    func describeRemoteFailure(
        stdout: String,
        stderr: String,
        exitCode: Int32,
        connection: ConnectionProfile?
    ) -> String {
        formattedRemoteFailure(
            stdout: stdout,
            stderr: stderr,
            exitCode: exitCode,
            connection: connection
        )
    }

    private func describeLocalFailure(
        stdout: String,
        stderr: String,
        exitCode: Int32
    ) -> String {
        if let structuredError = structuredRemoteError(in: stdout) ?? structuredRemoteError(in: stderr) {
            return structuredError
        }
        let rawMessage = [stderr, stdout]
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .first(where: { !$0.isEmpty })
        return rawMessage ?? "Local command failed with exit code \(exitCode)."
    }

    private func sshArguments(
        for connection: ConnectionProfile,
        remoteCommand: String?,
        allocateTTY: Bool,
        purpose: ConnectionPurpose
    ) -> [String] {
        var arguments = [
            "-o", "BatchMode=yes",
            "-o", "ConnectTimeout=10",
            "-o", "ServerAliveInterval=15",
            "-o", "ServerAliveCountMax=3"
        ]

        switch purpose {
        case .service:
            arguments.append(contentsOf: [
                "-o", "ControlMaster=auto",
                "-o", "ControlPersist=300",
                "-o", "ControlPath=\(paths.controlPath(for: connection))"
            ])
        case .serviceWithoutMultiplexing, .terminalShell:
            // Keep direct sessions isolated when multiplexing could be stale or
            // when an open PTY should not share the background RPC connection.
            arguments.append(contentsOf: [
                "-o", "ControlMaster=no",
                "-S", "none"
            ])
        }

        if allocateTTY {
            arguments.append("-tt")
        } else {
            arguments.append("-T")
        }

        if let port = connection.resolvedPort {
            arguments.append(contentsOf: ["-p", String(port)])
        }

        arguments.append("--")
        arguments.append(destination(for: connection))

        if let remoteCommand {
            arguments.append(remoteCommand)
        }

        return arguments
    }

    private func destination(for connection: ConnectionProfile) -> String {
        let target = connection.effectiveTarget
        guard let user = connection.trimmedUser else {
            return target
        }
        return "\(user)@\(target)"
    }

    private func formattedRemoteFailure(
        stdout: String,
        stderr: String,
        exitCode: Int32,
        connection: ConnectionProfile?
    ) -> String {
        if let structuredError = structuredRemoteError(in: stdout) {
            return structuredError
        }
        if let structuredError = structuredRemoteError(in: stderr) {
            return structuredError
        }

        let rawMessage = [stderr, stdout]
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }
            .first(where: { !$0.isEmpty }) ?? ""

        let lowered = rawMessage.lowercased()
        let target = connection?.effectiveTarget

        if lowered.contains("permission denied") {
            return "SSH authentication failed. Verify the key, SSH agent, and user for this SSH target."
        }
        if lowered.contains("host key verification failed") {
            return "SSH host key verification failed. Connect once in Terminal.app or update known_hosts before retrying."
        }
        if lowered.contains("remote host identification has changed") {
            return "The SSH host key changed for this target. Refresh the entry in known_hosts before retrying."
        }
        if lowered.contains("could not resolve hostname") || lowered.contains("name or service not known") {
            return "The SSH target could not be resolved. Check the alias, hostname, IP address, or SSH config entry in this profile."
        }
        if lowered.contains("connection refused") {
            if isLoopbackTarget(target) {
                return "The SSH server on this Mac refused the connection. If you are connecting to localhost or the same Mac, make sure SSH access is enabled and retry."
            }
            return "The SSH server refused the connection. Confirm that SSH is enabled and reachable on the target host."
        }
        if lowered.contains("operation timed out") || lowered.contains("connection timed out") {
            if isLoopbackTarget(target) {
                return "The SSH connection to this Mac timed out. If you are testing localhost or the same Mac, verify that SSH access is enabled and retry."
            }
            return "The SSH connection timed out. Check that the target host is reachable from this Mac and that your SSH route is correct."
        }
        if lowered.contains("no route to host") || lowered.contains("network is unreachable") {
            return "The SSH target is unreachable from this Mac. Check the hostname, IP address, VPN, or local network path and retry. If this host is on your local Wi-Fi/LAN, allow Hermes Desktop in System Settings > Privacy & Security > Local Network."
        }
        if lowered.contains("python3: command not found") ||
            lowered.contains("command not found: python3") ||
            lowered.contains("python3: not found") ||
            lowered.contains("unknown command: python3") ||
            lowered.contains("env: python3: no such file or directory") {
            if isLoopbackTarget(target) {
                return "SSH succeeded, but python3 is not available in the non-interactive SSH shell PATH for this Mac. Install python3 or expose it in the SSH shell environment before retrying."
            }
            return "SSH succeeded, but python3 is not available in the remote non-interactive SSH shell PATH. Install python3 or expose it in the SSH shell environment before retrying. Hermes Desktop requires python3 for discovery, file editing, and session browsing."
        }

        if !rawMessage.isEmpty {
            return rawMessage
        }

        return "SSH command failed with exit code \(exitCode)."
    }

    private func shouldRetryWithoutMultiplexing(_ result: SSHCommandResult) -> Bool {
        let message = [result.stderr, result.stdout]
            .map { $0.lowercased() }
            .joined(separator: "\n")

        return message.contains("no route to host") ||
            message.contains("network is unreachable") ||
            message.contains("connection timed out") ||
            message.contains("operation timed out") ||
            message.contains("connection refused") ||
            message.contains("connection reset") ||
            message.contains("connection closed") ||
            message.contains("broken pipe") ||
            message.contains("mux_client") ||
            message.contains("control socket")
    }

    private func isLoopbackTarget(_ target: String?) -> Bool {
        guard let target else { return false }
        let normalized = target.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        return normalized == "localhost" ||
            normalized == "127.0.0.1" ||
            normalized == "::1" ||
            normalized.hasPrefix("localhost.")
    }

    private func formattedInvalidJSONResponse(
        stdout: String,
        stderr: String,
        decodingError: Error,
        isLocal: Bool
    ) -> String {
        let trimmedStdout = stdout.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedStderr = stderr.trimmingCharacters(in: .whitespacesAndNewlines)

        if looksLikeNonJSONShellOutput(trimmedStdout) {
            let guidance = isLocal
                ? "Local command returned non-JSON output. This usually means a shell startup file printed text before the expected response. Keep startup files quiet for non-interactive sessions and retry."
                : "Remote command returned non-JSON output. This usually means a shell startup file printed text during a non-interactive SSH command. Keep startup files quiet for non-interactive SSH sessions and retry."
            let preview = shortenedOutputPreview(trimmedStdout)
            if preview.isEmpty {
                return guidance
            }
            return "\(guidance)\n\nPreview:\n\(preview)"
        }

        var message = "Failed to decode remote JSON: \(decodingError.localizedDescription)"
        if !trimmedStdout.isEmpty {
            message += "\n\n\(trimmedStdout)"
        }
        if trimmedStdout.isEmpty, !trimmedStderr.isEmpty {
            message += "\n\nstderr:\n\(trimmedStderr)"
        }
        return message
    }

    private func looksLikeNonJSONShellOutput(_ output: String) -> Bool {
        guard let firstCharacter = output.first else { return false }
        if firstCharacter == "{" || firstCharacter == "[" {
            return false
        }

        let lowered = output.lowercased()
        return output.contains("{") ||
            output.contains("[") ||
            lowered.contains("welcome") ||
            lowered.contains("last login")
    }

    private func shortenedOutputPreview(_ output: String, limit: Int = 240) -> String {
        guard output.count > limit else { return output }
        let endIndex = output.index(output.startIndex, offsetBy: limit)
        return String(output[..<endIndex]).trimmingCharacters(in: .whitespacesAndNewlines) + "..."
    }

    private func structuredRemoteError(in output: String) -> String? {
        guard let data = output.data(using: .utf8),
              let payload = try? JSONDecoder().decode(RemoteErrorPayload.self, from: data),
              let error = payload.trimmedError else {
            return nil
        }

        return error
    }
}

private struct RemoteErrorPayload: Decodable {
    let error: String?

    var trimmedError: String? {
        guard let error else { return nil }
        let trimmed = error.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? nil : trimmed
    }
}

private struct FoundationSSHProcessRunner: SSHProcessRunning {
    func run(
        executableURL: URL,
        arguments: [String],
        standardInput: Data?
    ) async throws -> SSHCommandResult {
        try await withCheckedThrowingContinuation { continuation in
            let process = Process()
            let stdoutPipe = Pipe()
            let stderrPipe = Pipe()
            let stdoutHandle = stdoutPipe.fileHandleForReading
            let stderrHandle = stderrPipe.fileHandleForReading
            let state = ProcessOutputState()

            process.executableURL = executableURL
            process.arguments = arguments
            process.standardOutput = stdoutPipe
            process.standardError = stderrPipe

            if standardInput != nil {
                process.standardInput = Pipe()
            }

            stdoutHandle.readabilityHandler = { handle in
                let chunk = handle.availableData
                if chunk.isEmpty {
                    handle.readabilityHandler = nil
                } else {
                    state.appendStdout(chunk)
                }
            }

            stderrHandle.readabilityHandler = { handle in
                let chunk = handle.availableData
                if chunk.isEmpty {
                    handle.readabilityHandler = nil
                } else {
                    state.appendStderr(chunk)
                }
            }

            process.terminationHandler = { process in
                stdoutHandle.readabilityHandler = nil
                stderrHandle.readabilityHandler = nil

                let remainingStdout = stdoutHandle.readDataToEndOfFile()
                let remainingStderr = stderrHandle.readDataToEndOfFile()
                let result = state.finalResult(
                    exitCode: process.terminationStatus,
                    remainingStdout: remainingStdout,
                    remainingStderr: remainingStderr
                )

                guard state.claimResume() else { return }
                continuation.resume(returning: result)
            }

            do {
                try process.run()
                if let standardInput,
                   let pipe = process.standardInput as? Pipe {
                    pipe.fileHandleForWriting.write(standardInput)
                    do {
                        try pipe.fileHandleForWriting.close()
                    } catch {
                        stdoutHandle.readabilityHandler = nil
                        stderrHandle.readabilityHandler = nil

                        guard state.claimResume() else { return }
                        continuation.resume(
                            throwing: SSHTransportError.localFailure(
                                "Failed to finish sending input to ssh: \(error.localizedDescription)"
                            )
                        )
                    }
                }
            } catch {
                stdoutHandle.readabilityHandler = nil
                stderrHandle.readabilityHandler = nil

                guard state.claimResume() else { return }
                continuation.resume(throwing: SSHTransportError.launchFailure(error.localizedDescription))
            }
        }
    }
}

private final class ProcessOutputState: @unchecked Sendable {
    private let lock = NSLock()
    private var stdoutData = Data()
    private var stderrData = Data()
    private var didResume = false

    func appendStdout(_ data: Data) {
        lock.lock()
        stdoutData.append(data)
        lock.unlock()
    }

    func appendStderr(_ data: Data) {
        lock.lock()
        stderrData.append(data)
        lock.unlock()
    }

    func finalResult(exitCode: Int32, remainingStdout: Data, remainingStderr: Data) -> SSHCommandResult {
        lock.lock()
        stdoutData.append(remainingStdout)
        stderrData.append(remainingStderr)
        let result = SSHCommandResult(
            stdout: String(decoding: stdoutData, as: UTF8.self),
            stderr: String(decoding: stderrData, as: UTF8.self),
            exitCode: exitCode
        )
        lock.unlock()
        return result
    }

    func claimResume() -> Bool {
        lock.lock()
        defer { lock.unlock() }

        if didResume {
            return false
        }

        didResume = true
        return true
    }
}
