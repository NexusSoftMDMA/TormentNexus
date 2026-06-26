import Foundation

/// Output of `hermes kanban stats --json`. Drives the toolbar glance
/// ("12 todo · 3 running · 5 blocked"), the per-project Kanban summary
/// widget, and the column-count badges on the board header.
public struct HermesKanbanStats: Sendable, Equatable, Codable {
    public let byStatus: [String: Int]
    public let byAssignee: [String: Int]
    public let byTenant: [String: Int]
    /// Age in seconds of the oldest task currently in the `ready` status.
    /// `nil` when no tasks are ready. Helps surface a stuck dispatcher.
    public let oldestReadyAgeSeconds: Double?

    public init(
        byStatus: [String: Int],
        byAssignee: [String: Int] = [:],
        byTenant: [String: Int] = [:],
        oldestReadyAgeSeconds: Double? = nil
    ) {
        self.byStatus = byStatus
        self.byAssignee = byAssignee
        self.byTenant = byTenant
        self.oldestReadyAgeSeconds = oldestReadyAgeSeconds
    }

    public static let empty = HermesKanbanStats(byStatus: [:])

    enum CodingKeys: String, CodingKey {
        case byStatus = "by_status"
        case byAssignee = "by_assignee"
        case byTenant = "by_tenant"
        case oldestReadyAgeSeconds = "oldest_ready_age_seconds"
    }

    public init(from decoder: any Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.byStatus = try c.decodeIfPresent([String: Int].self, forKey: .byStatus) ?? [:]
        self.byAssignee = try c.decodeIfPresent([String: Int].self, forKey: .byAssignee) ?? [:]
        self.byTenant = try c.decodeIfPresent([String: Int].self, forKey: .byTenant) ?? [:]
        self.oldestReadyAgeSeconds = try c.decodeIfPresent(Double.self, forKey: .oldestReadyAgeSeconds)
    }

    /// "12 todo · 3 running · 5 blocked" formatted glance string. Skips
    /// empty buckets and never includes archived. Returns an empty
    /// string when there's nothing to show so callers can hide chrome.
    public var glanceString: String {
        let order: [(String, String)] = [
            ("todo", "todo"),
            ("ready", "ready"),
            ("running", "running"),
            ("blocked", "blocked"),
            ("done", "done")
        ]
        let parts = order.compactMap { (key, label) -> String? in
            guard let n = byStatus[key], n > 0 else { return nil }
            return "\(n) \(label)"
        }
        return parts.joined(separator: " · ")
    }

    /// Active task count across the board (everything except archived
    /// and done). Used as a badge on the sidebar / project tab.
    public var activeCount: Int {
        ["triage", "todo", "ready", "running", "blocked"]
            .map { byStatus[$0] ?? 0 }
            .reduce(0, +)
    }
}
