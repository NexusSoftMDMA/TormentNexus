import Foundation

/// One row from `hermes kanban assignees --json`. The output is the
/// union of profiles configured on the host (`~/.hermes/profiles/`)
/// and any names appearing in the live board's `assignee` column —
/// covers the case where a profile was renamed but historical tasks
/// still reference the old name.
public struct HermesKanbanAssignee: Sendable, Equatable, Identifiable, Codable {
    public var id: String { profile }
    public let profile: String
    public let activeCount: Int
    public let totalCount: Int

    public init(profile: String, activeCount: Int = 0, totalCount: Int = 0) {
        self.profile = profile
        self.activeCount = activeCount
        self.totalCount = totalCount
    }

    enum CodingKeys: String, CodingKey {
        case profile
        case activeCount = "active"
        case totalCount = "total"
    }

    public init(from decoder: any Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.profile = try c.decode(String.self, forKey: .profile)
        self.activeCount = try c.decodeIfPresent(Int.self, forKey: .activeCount) ?? 0
        self.totalCount = try c.decodeIfPresent(Int.self, forKey: .totalCount) ?? 0
    }
}
