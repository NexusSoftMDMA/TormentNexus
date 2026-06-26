import Foundation

/// A user-saved model selection that can be applied to a project, a
/// cron job, or a live chat session. Lightweight overlay on top of
/// Hermes's own model resolution — applying a preset does **not**
/// touch `~/.hermes/config.yaml`, profiles, skills, or MCP. It only
/// drives the per-session `session/set_model` ACP call or the
/// `-m`/`--provider` flags on `hermes -z` for cron one-shots.
///
/// Identity is the UUID, not the name — renaming a preset doesn't
/// break the references stored on projects or cron jobs.
public struct ModelPreset: Codable, Identifiable, Sendable, Hashable {
    /// Stable identity. Survives renames so per-project / per-cron
    /// bindings keep resolving.
    public let id: UUID

    /// User-facing label. Free text — duplicates are allowed at the
    /// data layer; the CRUD UI nudges users toward unique names but
    /// doesn't enforce.
    public var name: String

    /// Model name as Hermes expects it (e.g. `claude-sonnet-4.6`,
    /// `openrouter/anthropic/claude-3.5-sonnet`). Passed verbatim to
    /// the ACP `session/set_model` call's `modelId` field and to
    /// `hermes -z -m <modelID>` for cron jobs.
    public var modelID: String

    /// Provider slug as Hermes expects (e.g. `anthropic`,
    /// `openrouter`, `nous`). Passed to `hermes -z --provider <id>`
    /// for cron jobs. For ACP sessions the provider is re-derived by
    /// Hermes from `modelID` via `_resolve_model_selection`, but we
    /// still record it so cron and the picker UI have a complete
    /// reference.
    public var providerID: String

    /// Optional free-text rationale shown on the preset row and edit
    /// sheet (e.g. "Best for reasoning-heavy tasks"). Not surfaced to
    /// the agent.
    public var notes: String?

    public let createdAt: Date
    public var updatedAt: Date

    public init(
        id: UUID = UUID(),
        name: String,
        modelID: String,
        providerID: String,
        notes: String? = nil,
        createdAt: Date = Date(),
        updatedAt: Date = Date()
    ) {
        self.id = id
        self.name = name
        self.modelID = modelID
        self.providerID = providerID
        self.notes = notes
        self.createdAt = createdAt
        self.updatedAt = updatedAt
    }
}

/// On-disk envelope for `~/.hermes/scarf/model_presets.json`. Wrapped
/// in a versioned container so we can evolve the schema without
/// breaking older Scarf binaries reading the file.
public struct ModelPresetStore: Codable, Sendable, Equatable {
    public static let currentVersion = 1

    public var version: Int
    public var presets: [ModelPreset]
    public var updatedAt: String?

    public init(
        version: Int = ModelPresetStore.currentVersion,
        presets: [ModelPreset] = [],
        updatedAt: String? = nil
    ) {
        self.version = version
        self.presets = presets
        self.updatedAt = updatedAt
    }

    /// ISO-8601 timestamp helper — same format as `SessionProjectMap`
    /// so cross-file greps see consistent stamps.
    public static func nowISO8601() -> String {
        ISO8601DateFormatter().string(from: Date())
    }
}
