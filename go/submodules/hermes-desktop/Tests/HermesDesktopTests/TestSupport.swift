import CryptoKit
import Foundation

@testable import HermesDesktop

struct LocalScriptResult {
    let stdout: String
    let stderr: String
    let exitCode: Int32
}

func makeTemporaryDirectory() throws -> URL {
    let url = FileManager.default.temporaryDirectory
        .appendingPathComponent("HermesDesktopTests-\(UUID().uuidString)", isDirectory: true)
    try FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
    return url
}

func makeTestAppPaths(root: URL) -> AppPaths {
    AppPaths(
        fileManager: .default,
        applicationSupportURL: root.appendingPathComponent("ApplicationSupport", isDirectory: true),
        controlSocketDirectoryURL: root.appendingPathComponent("ControlSockets", isDirectory: true)
    )
}

func runPythonScript(
    _ script: String,
    environment: [String: String] = [:]
) throws -> LocalScriptResult {
    let root = try makeTemporaryDirectory()
    defer {
        try? FileManager.default.removeItem(at: root)
    }

    let scriptURL = root.appendingPathComponent("script.py")
    try script.write(to: scriptURL, atomically: true, encoding: .utf8)

    let process = Process()
    let stdoutPipe = Pipe()
    let stderrPipe = Pipe()

    process.executableURL = URL(fileURLWithPath: "/usr/bin/env")
    process.arguments = ["python3", scriptURL.path]
    process.environment = ProcessInfo.processInfo.environment.merging(environment) { _, newValue in
        newValue
    }
    process.standardOutput = stdoutPipe
    process.standardError = stderrPipe

    try process.run()
    process.waitUntilExit()

    let stdout = String(
        data: stdoutPipe.fileHandleForReading.readDataToEndOfFile(),
        encoding: .utf8
    ) ?? ""
    let stderr = String(
        data: stderrPipe.fileHandleForReading.readDataToEndOfFile(),
        encoding: .utf8
    ) ?? ""

    return LocalScriptResult(
        stdout: stdout,
        stderr: stderr,
        exitCode: process.terminationStatus
    )
}

func sha256Hex(_ value: Data) -> String {
    SHA256.hash(data: value).map { String(format: "%02x", $0) }.joined()
}
