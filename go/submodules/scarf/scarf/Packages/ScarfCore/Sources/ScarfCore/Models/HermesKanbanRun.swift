import Foundation

/// One attempt to execute a kanban task — `hermes kanban runs <id> --json`
/// returns an array of these per task. Each run records the worker
/// profile that claimed the task, the outcome, and a structured
/// metadata blob the worker handed back.
public struct HermesKanbanRun: Sendable, Equatable, Identifiable, Codable {
    public let id: Int
    public let taskId: String
    public let profile: String?
    public let stepKey: String?
    public let status: String           // running | done | blocked | crashed | timed_out | failed | released
    public let claimLock: String?       // "host:pid" at spawn time
    public let claimExpires: Int?
    public let workerPid: Int?
    public let maxRuntimeSeconds: Int?
    public let lastHeartbeatAt: String?
    public let startedAt: String
    public let endedAt: String?
    public let outcome: String?         // completed | blocked | crashed | timed_out | spawn_failed | gave_up | reclaimed
    public let summary: String?
    public let error: String?
    /// `metadata` is an opaque JSON dict from the worker. Carried as a
    /// raw string so we don't lock the typed shape.
    public let metadataJSON: String?

    // v0.13 (v2026.5.7) fields. Both Optional / empty-default so a v0.12
    // host's run row decodes without error.
    /// Per-attempt distress signals. Cross-run signals (retry cap hit,
    /// etc.) hang off `HermesKanbanTask.diagnostics`; in-flight signals
    /// (heartbeat stalled, darwin zombie detected) attach here.
    public let diagnostics: [HermesKanbanDiagnostic]
    /// Server-side unified failure counter (renamed from three separate
    /// spawn / timeout / crash counters in v0.13). Optional — when nil,
    /// callers fall back to counting failed runs in the runs array.
    public let failureCount: Int?

    public init(
        id: Int,
        taskId: String,
        profile: String? = nil,
        stepKey: String? = nil,
        status: String,
        claimLock: String? = nil,
        claimExpires: Int? = nil,
        workerPid: Int? = nil,
        maxRuntimeSeconds: Int? = nil,
        lastHeartbeatAt: String? = nil,
        startedAt: String,
        endedAt: String? = nil,
        outcome: String? = nil,
        summary: String? = nil,
        error: String? = nil,
        metadataJSON: String? = nil,
        diagnostics: [HermesKanbanDiagnostic] = [],
        failureCount: Int? = nil
    ) {
        self.id = id
        self.taskId = taskId
        self.profile = profile
        self.stepKey = stepKey
        self.status = status
        self.claimLock = claimLock
        self.claimExpires = claimExpires
        self.workerPid = workerPid
        self.maxRuntimeSeconds = maxRuntimeSeconds
        self.lastHeartbeatAt = lastHeartbeatAt
        self.startedAt = startedAt
        self.endedAt = endedAt
        self.outcome = outcome
        self.summary = summary
        self.error = error
        self.metadataJSON = metadataJSON
        self.diagnostics = diagnostics
        self.failureCount = failureCount
    }

    enum CodingKeys: String, CodingKey {
        case id
        case taskId = "task_id"
        case profile
        case stepKey = "step_key"
        case status
        case claimLock = "claim_lock"
        case claimExpires = "claim_expires"
        case workerPid = "worker_pid"
        case maxRuntimeSeconds = "max_runtime_seconds"
        case lastHeartbeatAt = "last_heartbeat_at"
        case startedAt = "started_at"
        case endedAt = "ended_at"
        case outcome
        case summary
        case error
        case metadata
        case diagnostics
        case failureCount = "failure_count"
    }

    public init(from decoder: any Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = try c.decodeIfPresent(Int.self, forKey: .id) ?? 0
        self.taskId = try c.decodeIfPresent(String.self, forKey: .taskId) ?? ""
        self.profile = try c.decodeIfPresent(String.self, forKey: .profile)
        self.stepKey = try c.decodeIfPresent(String.self, forKey: .stepKey)
        self.status = try c.decodeIfPresent(String.self, forKey: .status) ?? "unknown"
        self.claimLock = try c.decodeIfPresent(String.self, forKey: .claimLock)
        self.claimExpires = try c.decodeIfPresent(Int.self, forKey: .claimExpires)
        self.workerPid = try c.decodeIfPresent(Int.self, forKey: .workerPid)
        self.maxRuntimeSeconds = try c.decodeIfPresent(Int.self, forKey: .maxRuntimeSeconds)
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        if let unix = try? c.decodeIfPresent(Double.self, forKey: .lastHeartbeatAt) {
            self.lastHeartbeatAt = f.string(from: Date(timeIntervalSince1970: unix))
        } else {
            self.lastHeartbeatAt = try c.decodeIfPresent(String.self, forKey: .lastHeartbeatAt)
        }
        if let unix = try? c.decodeIfPresent(Double.self, forKey: .startedAt) {
            self.startedAt = f.string(from: Date(timeIntervalSince1970: unix))
        } else {
            self.startedAt = (try? c.decodeIfPresent(String.self, forKey: .startedAt)) ?? ""
        }
        if let unix = try? c.decodeIfPresent(Double.self, forKey: .endedAt) {
            self.endedAt = f.string(from: Date(timeIntervalSince1970: unix))
        } else {
            self.endedAt = try c.decodeIfPresent(String.self, forKey: .endedAt)
        }
        self.outcome = try c.decodeIfPresent(String.self, forKey: .outcome)
        self.summary = try c.decodeIfPresent(String.self, forKey: .summary)
        self.error = try c.decodeIfPresent(String.self, forKey: .error)

        if let raw = try? c.decodeIfPresent(String.self, forKey: .metadata) {
            self.metadataJSON = raw
        } else if c.contains(.metadata) {
            let nested = try c.decode(JSONAny.self, forKey: .metadata)
            let data = try JSONEncoder().encode(nested)
            self.metadataJSON = String(data: data, encoding: .utf8)
        } else {
            self.metadataJSON = nil
        }

        // v0.13 diagnostics array — `try?` so a malformed entry doesn't
        // poison the whole run row. Empty default for pre-v0.13 hosts.
        self.diagnostics = (try? c.decodeIfPresent([HermesKanbanDiagnostic].self, forKey: .diagnostics)) ?? []
        self.failureCount = try c.decodeIfPresent(Int.self, forKey: .failureCount)
    }

    public func encode(to encoder: any Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encode(id, forKey: .id)
        try c.encode(taskId, forKey: .taskId)
        try c.encodeIfPresent(profile, forKey: .profile)
        try c.encodeIfPresent(stepKey, forKey: .stepKey)
        try c.encode(status, forKey: .status)
        try c.encodeIfPresent(claimLock, forKey: .claimLock)
        try c.encodeIfPresent(claimExpires, forKey: .claimExpires)
        try c.encodeIfPresent(workerPid, forKey: .workerPid)
        try c.encodeIfPresent(maxRuntimeSeconds, forKey: .maxRuntimeSeconds)
        try c.encodeIfPresent(lastHeartbeatAt, forKey: .lastHeartbeatAt)
        try c.encode(startedAt, forKey: .startedAt)
        try c.encodeIfPresent(endedAt, forKey: .endedAt)
        try c.encodeIfPresent(outcome, forKey: .outcome)
        try c.encodeIfPresent(summary, forKey: .summary)
        try c.encodeIfPresent(error, forKey: .error)
        try c.encodeIfPresent(metadataJSON, forKey: .metadata)
        try c.encode(diagnostics, forKey: .diagnostics)
        try c.encodeIfPresent(failureCount, forKey: .failureCount)
    }
}
