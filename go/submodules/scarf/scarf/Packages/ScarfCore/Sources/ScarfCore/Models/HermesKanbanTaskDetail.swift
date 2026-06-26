import Foundation

/// Output of `hermes kanban show <id> --json`. Wraps a task with its full
/// audit trail: comments + events + parent results. Loaded on-demand
/// when the user opens the inspector pane; the board itself only carries
/// the lightweight `HermesKanbanTask` rows.
public struct HermesKanbanTaskDetail: Sendable, Equatable, Codable {
    public let task: HermesKanbanTask
    public let comments: [HermesKanbanComment]
    public let events: [HermesKanbanEvent]
    /// Parent-task results keyed by parent task id. Hermes hands these
    /// to the worker as upstream context; surfacing them in the
    /// inspector is useful for understanding why a task started.
    public let parentResults: [String: String]
    /// Envelope-level diagnostics array (sibling to `task`, not nested
    /// inside it). Defensive — Hermes v0.13's wire shape may attach
    /// diagnostics to the task itself OR to the envelope.
    /// `allDiagnostics` dedupes both sources by `(kind, detected_at)`.
    public let envelopeDiagnostics: [HermesKanbanDiagnostic]?

    public init(
        task: HermesKanbanTask,
        comments: [HermesKanbanComment] = [],
        events: [HermesKanbanEvent] = [],
        parentResults: [String: String] = [:],
        envelopeDiagnostics: [HermesKanbanDiagnostic]? = nil
    ) {
        self.task = task
        self.comments = comments
        self.events = events
        self.parentResults = parentResults
        self.envelopeDiagnostics = envelopeDiagnostics
    }

    enum CodingKeys: String, CodingKey {
        case task
        case comments
        case events
        case parentResults = "parent_results"
        case envelopeDiagnostics = "diagnostics"
    }

    public init(from decoder: any Decoder) throws {
        // Hermes emits `kanban show --json` either as a nested
        // {task: {...}, comments: [...], events: [...]} object or
        // as a flat task object with extra `comments`/`events`
        // keys at top level. Try the nested form first; fall
        // back to top-level decode.
        let container = try decoder.container(keyedBy: CodingKeys.self)
        if let nested = try? container.decode(HermesKanbanTask.self, forKey: .task) {
            self.task = nested
        } else {
            let single = try decoder.singleValueContainer()
            self.task = try single.decode(HermesKanbanTask.self)
        }
        self.comments = (try? container.decodeIfPresent([HermesKanbanComment].self, forKey: .comments)) ?? []
        self.events = (try? container.decodeIfPresent([HermesKanbanEvent].self, forKey: .events)) ?? []
        self.parentResults = (try? container.decodeIfPresent([String: String].self, forKey: .parentResults)) ?? [:]
        // Same `try?` shield as the rest — a malformed envelope
        // diagnostics array shouldn't reject the whole show response.
        self.envelopeDiagnostics = try? container.decodeIfPresent([HermesKanbanDiagnostic].self, forKey: .envelopeDiagnostics)
    }

    public func encode(to encoder: any Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encode(task, forKey: .task)
        try c.encode(comments, forKey: .comments)
        try c.encode(events, forKey: .events)
        try c.encode(parentResults, forKey: .parentResults)
        try c.encodeIfPresent(envelopeDiagnostics, forKey: .envelopeDiagnostics)
    }

    /// Unified diagnostics view for the inspector. Combines `task.diagnostics`
    /// with envelope-level diagnostics (when present) and dedupes on the
    /// `(kind, detectedAt)` tuple. Wire-side dupes are unlikely but cheap to
    /// filter. Empty for pre-v0.13 hosts.
    public var allDiagnostics: [HermesKanbanDiagnostic] {
        let onTask = task.diagnostics
        let onEnvelope = envelopeDiagnostics ?? []
        var seen = Set<String>()
        return (onTask + onEnvelope).filter { diag in
            let key = "\(diag.kind)|\(diag.detectedAt ?? "")"
            return seen.insert(key).inserted
        }
    }
}
