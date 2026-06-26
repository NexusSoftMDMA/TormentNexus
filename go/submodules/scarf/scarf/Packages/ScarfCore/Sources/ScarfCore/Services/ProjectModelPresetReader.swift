import Foundation

/// Cross-platform read-only helper for `<project>/.scarf/manifest.json`'s
/// `modelPresetID` field. Mirrors `KanbanTenantReader` — the full
/// `ProjectTemplateManifest` Codable type lives in the Mac app target
/// with all the install/export machinery; iOS doesn't link it. This
/// lightweight projection gives both targets a way to read the preset
/// binding without duplicating the entire manifest model.
///
/// **Write side.** The Mac target's `ProjectModelPresetBinding` owns
/// writing to this field. iOS doesn't mutate per-project preset
/// bindings in v1 (read-only display only).
public struct ProjectModelPresetReader: Sendable {
    public let context: ServerContext

    public nonisolated init(context: ServerContext) {
        self.context = context
    }

    /// Read the project's bound model preset UUID, or `nil` if the
    /// manifest doesn't exist or doesn't carry a binding (use the
    /// global default in `config.yaml`).
    public nonisolated func presetID(forProjectPath projectPath: String) -> String? {
        let manifestPath = projectPath + "/.scarf/manifest.json"
        let transport = context.makeTransport()
        guard transport.fileExists(manifestPath),
              let data = try? transport.readFile(manifestPath)
        else {
            return nil
        }
        return Self.presetID(fromManifestData: data)
    }

    /// Pure-input variant for tests + tooling that already have the
    /// JSON bytes in hand. Returns `nil` when the bytes don't decode
    /// or the field isn't present.
    public nonisolated static func presetID(fromManifestData data: Data) -> String? {
        struct Projection: Decodable {
            let modelPresetID: String?
        }
        return (try? JSONDecoder().decode(Projection.self, from: data))?.modelPresetID
    }
}
