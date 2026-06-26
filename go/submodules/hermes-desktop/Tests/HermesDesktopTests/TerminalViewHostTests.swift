import AppKit
import Darwin
import Foundation
@preconcurrency import SwiftTerm
import Testing
@testable import HermesDesktop

@MainActor
struct TerminalViewHostTests {
    @Test
    func terminalBackgroundColorIsClearWhenBackgroundImageIsActive() throws {
        let expectedBackground = TerminalThemeColor(hex: 0x112233)
        let preference = TerminalThemePreference(style: .paper)
            .updatingBackgroundColor(expectedBackground)
            .updatingForegroundColor(TerminalThemeColor(hex: 0xEEDFCC))
        let hostView = TerminalHostView(frame: .zero)

        hostView.apply(
            appearance: preference.resolvedAppearance,
            backgroundImageActive: true
        )

        assertColorIsClear(hostView.terminalView.nativeBackgroundColor)
        assertColorIsClear(NSColor(cgColor: try #require(hostView.layer?.backgroundColor)))
    }

    @Test
    func localTerminalLaunchCarriesHomePathAndShellThroughActualPTY() async throws {
        let transport = SSHTransport(paths: AppPaths())
        let connection = ConnectionProfile(kind: .local, label: "This Mac").updated()
        let original = transport.terminalLaunch(
            for: connection,
            startupCommandLine: #"printf 'HOME=%s\nPATH=%s\nSHELL=%s\nPTY_DONE\n' "$HOME" "$PATH" "$SHELL"; exit 0"#
        )
        let delegate = PTYProbeDelegate()
        let process = LocalProcess(delegate: delegate, dispatchQueue: DispatchQueue.global())
        defer { process.terminate() }
        process.startProcess(
            executable: original.executablePath,
            args: original.arguments,
            environment: original.environment,
            execName: original.executableName
        )

        let result = try await delegate.result(timeoutSeconds: 10)
        #expect(result.contains("HOME=\(FileManager.default.homeDirectoryForCurrentUser.path)"))
        #expect(result.contains("PATH="))
        #expect(!result.contains("PATH=\r\n"))
        #expect(result.contains("SHELL="))
        #expect(result.contains("PTY_DONE"))
    }

    private func assertColorIsClear(_ nsColor: NSColor?) {
        let color = nsColor?.usingColorSpace(.deviceRGB)

        #expect(color?.alphaComponent == 0)
    }
}

private final class PTYProbeDelegate: LocalProcessDelegate, @unchecked Sendable {
    private let lock = NSLock()
    private let semaphore = DispatchSemaphore(value: 0)
    private var outputData = Data()

    func processTerminated(_ source: LocalProcess, exitCode: Int32?) {}

    func dataReceived(slice: ArraySlice<UInt8>) {
        lock.lock()
        outputData.append(contentsOf: slice)
        let output = String(decoding: outputData, as: UTF8.self)
        guard output.contains("PTY_DONE") else {
            lock.unlock()
            return
        }
        lock.unlock()
        semaphore.signal()
    }

    func getWindowSize() -> winsize {
        winsize(ws_row: 24, ws_col: 80, ws_xpixel: 0, ws_ypixel: 0)
    }

    func result(timeoutSeconds: Double) async throws -> String {
        let (waitResult, output) = await Task.detached {
            self.blockingResult(timeoutSeconds: timeoutSeconds)
        }.value
        guard waitResult == .success else {
            throw PTYProbeError.timeout
        }
        return output
    }

    private func blockingResult(timeoutSeconds: Double) -> (DispatchTimeoutResult, String) {
        let waitResult = semaphore.wait(timeout: .now() + timeoutSeconds)
        lock.lock()
        let output = String(decoding: outputData, as: UTF8.self)
        lock.unlock()
        return (waitResult, output)
    }
}

private enum PTYProbeError: Error {
    case timeout
}
