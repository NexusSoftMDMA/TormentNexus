import Testing
import Foundation
import ScarfCore
@testable import scarf

/// Tests for `ProjectModelPresetBinding`. Mirrors the shape of
/// `KanbanTenantResolverTests` — each test stages a tmp project dir
/// and exercises the read/write paths against a real
/// `ProjectTemplateManifest` Codable round-trip. No reliance on the
/// user's `~/.hermes`.
@Suite struct ProjectModelPresetBindingTests {

    @Test func bindingMissingWhenManifestAbsent() throws {
        let dir = try Self.makeTempDir()
        defer { try? FileManager.default.removeItem(atPath: dir) }
        let project = ProjectEntry(name: "Bare", path: dir)

        let binding = ProjectModelPresetBinding(context: .local)
        #expect(binding.boundPresetID(for: project) == nil)
    }

    @Test func bindWritesSentinelManifestForBareProject() throws {
        let dir = try Self.makeTempDir()
        defer { try? FileManager.default.removeItem(atPath: dir) }
        let project = ProjectEntry(name: "Bare", path: dir)
        let id = UUID().uuidString

        let binding = ProjectModelPresetBinding(context: .local)
        try binding.bind(presetID: id, to: project)

        // Manifest should exist after first bind and carry the id.
        #expect(FileManager.default.fileExists(atPath: dir + "/.scarf/manifest.json"))
        #expect(binding.boundPresetID(for: project) == id)
    }

    @Test func bindMutatesExistingManifest() throws {
        let dir = try Self.makeTempDir()
        defer { try? FileManager.default.removeItem(atPath: dir) }
        let project = ProjectEntry(name: "Templated", path: dir)
        let scarfDir = dir + "/.scarf"
        try FileManager.default.createDirectory(atPath: scarfDir, withIntermediateDirectories: true)

        // Minimal v3 manifest with kanbanTenant set — verify we
        // preserve sibling fields when we mutate modelPresetID.
        let preExisting = """
        {
          "schemaVersion": 3,
          "id": "author/example",
          "name": "Templated",
          "version": "1.0.0",
          "description": "demo",
          "contents": { "dashboard": true, "agentsMd": true },
          "kanbanTenant": "scarf:demo"
        }
        """
        try preExisting.data(using: .utf8)!.write(
            to: URL(fileURLWithPath: scarfDir + "/manifest.json")
        )

        let id = UUID().uuidString
        let binding = ProjectModelPresetBinding(context: .local)
        try binding.bind(presetID: id, to: project)

        // Re-read raw JSON: both fields must survive the round-trip.
        let raw = try String(contentsOfFile: scarfDir + "/manifest.json", encoding: .utf8)
        #expect(raw.contains("\"kanbanTenant\""))
        #expect(raw.contains("scarf:demo"))
        #expect(raw.contains(id))
    }

    @Test func clearBindingRemovesIt() throws {
        let dir = try Self.makeTempDir()
        defer { try? FileManager.default.removeItem(atPath: dir) }
        let project = ProjectEntry(name: "Bare", path: dir)

        let binding = ProjectModelPresetBinding(context: .local)
        try binding.bind(presetID: UUID().uuidString, to: project)
        #expect(binding.boundPresetID(for: project) != nil)

        try binding.bind(presetID: nil, to: project)
        #expect(binding.boundPresetID(for: project) == nil)
    }

    @Test func emptyStringTreatedAsClear() throws {
        let dir = try Self.makeTempDir()
        defer { try? FileManager.default.removeItem(atPath: dir) }
        let project = ProjectEntry(name: "Bare", path: dir)

        let binding = ProjectModelPresetBinding(context: .local)
        try binding.bind(presetID: UUID().uuidString, to: project)
        try binding.bind(presetID: "", to: project)
        #expect(binding.boundPresetID(for: project) == nil)
    }

    @Test func idempotentRebindDoesntRewriteFile() throws {
        let dir = try Self.makeTempDir()
        defer { try? FileManager.default.removeItem(atPath: dir) }
        let project = ProjectEntry(name: "Bare", path: dir)
        let id = UUID().uuidString

        let binding = ProjectModelPresetBinding(context: .local)
        try binding.bind(presetID: id, to: project)
        let path = dir + "/.scarf/manifest.json"
        let firstMtime = try FileManager.default
            .attributesOfItem(atPath: path)[.modificationDate] as? Date

        // Second bind with the same id — file should not be touched
        // (avoids file-watcher churn).
        Thread.sleep(forTimeInterval: 0.05)  // Make any mtime diff observable.
        try binding.bind(presetID: id, to: project)
        let secondMtime = try FileManager.default
            .attributesOfItem(atPath: path)[.modificationDate] as? Date

        #expect(firstMtime == secondMtime)
    }

    @Test func projectModelPresetReaderRoundTripsRawJSON() {
        // Pure-input variant — verify the cross-platform projection
        // reader picks up the field.
        let json = """
        { "modelPresetID": "deadbeef-1234-5678-9abc-def012345678", "otherField": 42 }
        """
        let data = json.data(using: .utf8)!
        let id = ProjectModelPresetReader.presetID(fromManifestData: data)
        #expect(id == "deadbeef-1234-5678-9abc-def012345678")
    }

    @Test func projectModelPresetReaderReturnsNilForMissingField() {
        let json = """
        { "schemaVersion": 1, "id": "x/y", "name": "Z" }
        """
        let data = json.data(using: .utf8)!
        let id = ProjectModelPresetReader.presetID(fromManifestData: data)
        #expect(id == nil)
    }

    // MARK: - Helpers

    nonisolated static func makeTempDir() throws -> String {
        let dir = NSTemporaryDirectory() + "scarf-project-preset-test-" + UUID().uuidString
        try FileManager.default.createDirectory(atPath: dir, withIntermediateDirectories: true)
        return dir
    }
}
