import Foundation

/// Cross-platform read-only helper for `<project>/.scarf/manifest.json`'s
/// `kanbanTenant` field. The full `ProjectTemplateManifest` Codable
/// type lives in the Mac app target (with all the install/export
/// machinery); iOS doesn't link it, so this lightweight projection
/// gives both targets a way to read just the tenant slug without
/// duplicating the entire manifest model.
public struct KanbanTenantReader: Sendable {
    public let context: ServerContext

    public nonisolated init(context: ServerContext) {
        self.context = context
    }

    /// Read the project's Kanban tenant slug, or `nil` if the manifest
    /// doesn't exist or doesn't carry one. Cheap — single JSON parse
    /// of a tiny projection.
    public nonisolated func tenant(forProjectPath projectPath: String) -> String? {
        let manifestPath = projectPath + "/.scarf/manifest.json"
        let transport = context.makeTransport()
        guard transport.fileExists(manifestPath),
              let data = try? transport.readFile(manifestPath)
        else {
            return nil
        }
        return Self.tenant(fromManifestData: data)
    }

    /// Pure-input variant for tests + tooling that already have the
    /// JSON bytes in hand. Returns `nil` when the bytes don't decode
    /// or the field isn't present.
    public nonisolated static func tenant(fromManifestData data: Data) -> String? {
        struct Projection: Decodable {
            let kanbanTenant: String?
        }
        return (try? JSONDecoder().decode(Projection.self, from: data))?.kanbanTenant
    }
}
