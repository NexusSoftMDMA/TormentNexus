import Foundation
import Testing

@testable import HermesDesktop

struct FileEditorServiceTests {
    @Test
    func readScriptRejectsInvalidUTF8Files() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let fileURL = root.appendingPathComponent("invalid.txt")
        try Data([0xFF, 0xFE, 0xFD]).write(to: fileURL)

        let script = try FileEditorService.makeReadScript(
            remotePath: fileURL.path,
            maxEditableBytes: 1024
        )
        let result = try runPythonScript(script)

        #expect(result.exitCode == 1)
        #expect(result.stderr.isEmpty)
        #expect(result.stdout.contains("is not valid UTF-8"))
    }

    @Test
    func readScriptEnforcesEditableSizeLimit() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let fileURL = root.appendingPathComponent("large.txt")
        try Data("abcd".utf8).write(to: fileURL)

        let script = try FileEditorService.makeReadScript(
            remotePath: fileURL.path,
            maxEditableBytes: 2
        )
        let result = try runPythonScript(script)

        #expect(result.exitCode == 1)
        #expect(result.stdout.contains("Hermes Desktop can edit remote text files up to"))
    }

    @Test
    func writeScriptRejectsStaleContentHashConflicts() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let fileURL = root.appendingPathComponent("notes.txt")
        let original = Data("before".utf8)
        try original.write(to: fileURL)
        let expectedHash = sha256Hex(original)

        try Data("after".utf8).write(to: fileURL)

        let script = try FileEditorService.makeWriteScript(
            remotePath: fileURL.path,
            content: "replacement",
            expectedContentHash: expectedHash
        )
        let result = try runPythonScript(script)

        #expect(result.exitCode == 1)
        #expect(result.stdout.contains("changed on the active host after it was loaded"))
    }

    @Test
    func writeScriptRejectsDanglingSymlinks() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let targetURL = root.appendingPathComponent("missing.txt")
        let symlinkURL = root.appendingPathComponent("link.txt")
        try FileManager.default.createSymbolicLink(
            at: symlinkURL,
            withDestinationURL: targetURL
        )

        let script = try FileEditorService.makeWriteScript(
            remotePath: symlinkURL.path,
            content: "replacement",
            expectedContentHash: nil
        )
        let result = try runPythonScript(script)

        #expect(result.exitCode == 1)
        #expect(result.stdout.contains("is a dangling symlink"))
    }
}
