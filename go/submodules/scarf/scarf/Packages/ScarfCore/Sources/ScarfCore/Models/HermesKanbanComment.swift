import Foundation

/// One comment from `hermes kanban show <id> --json` or appended via
/// `hermes kanban comment <id> <text>`. Comments are append-only — there's
/// no edit/delete verb.
public struct HermesKanbanComment: Sendable, Equatable, Identifiable, Codable {
    public let id: Int
    public let taskId: String
    public let author: String
    public let body: String
    public let createdAt: String

    public init(
        id: Int,
        taskId: String,
        author: String,
        body: String,
        createdAt: String
    ) {
        self.id = id
        self.taskId = taskId
        self.author = author
        self.body = body
        self.createdAt = createdAt
    }

    enum CodingKeys: String, CodingKey {
        case id
        case taskId = "task_id"
        case author
        case body
        case createdAt = "created_at"
    }

    public init(from decoder: any Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decode(Int.self, forKey: .id)
        self.taskId = try c.decodeIfPresent(String.self, forKey: .taskId) ?? ""
        self.author = try c.decodeIfPresent(String.self, forKey: .author) ?? ""
        self.body = try c.decodeIfPresent(String.self, forKey: .body) ?? ""
        // Hermes emits Unix integer timestamps from its SQLite columns;
        // accept both ints and ISO strings.
        if let unix = try? c.decodeIfPresent(Double.self, forKey: .createdAt) {
            let f = ISO8601DateFormatter()
            f.formatOptions = [.withInternetDateTime]
            self.createdAt = f.string(from: Date(timeIntervalSince1970: unix))
        } else {
            self.createdAt = (try? c.decodeIfPresent(String.self, forKey: .createdAt)) ?? ""
        }
    }
}
