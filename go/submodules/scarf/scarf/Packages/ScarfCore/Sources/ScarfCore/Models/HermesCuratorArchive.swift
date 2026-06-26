import Foundation

/// One entry in the `hermes curator list-archived` output. Decoded
/// tolerantly via `decodeIfPresent` so a stripped-down host (or a future
/// Hermes that drops one of the optional columns) doesn't crash the view.
///
/// Only `name` is required — every other field is optional and the
/// computed `*Label` accessors render `"—"` for missing values.
public struct HermesCuratorArchivedSkill: Sendable, Equatable, Identifiable, Codable {
    public var id: String { name }
    public let name: String
    public let category: String?
    public let archivedAt: String?
    public let reason: String?
    public let sizeBytes: Int?
    public let path: String?

    public init(
        name: String,
        category: String? = nil,
        archivedAt: String? = nil,
        reason: String? = nil,
        sizeBytes: Int? = nil,
        path: String? = nil
    ) {
        self.name = name
        self.category = category
        self.archivedAt = archivedAt
        self.reason = reason
        self.sizeBytes = sizeBytes
        self.path = path
    }

    private enum CodingKeys: String, CodingKey {
        case name
        case category
        case archivedAt = "archived_at"
        case reason
        case sizeBytes = "size_bytes"
        case path
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.name = try c.decode(String.self, forKey: .name)
        self.category = try c.decodeIfPresent(String.self, forKey: .category)
        self.archivedAt = try c.decodeIfPresent(String.self, forKey: .archivedAt)
        self.reason = try c.decodeIfPresent(String.self, forKey: .reason)
        self.sizeBytes = try c.decodeIfPresent(Int.self, forKey: .sizeBytes)
        self.path = try c.decodeIfPresent(String.self, forKey: .path)
    }

    public func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encode(name, forKey: .name)
        try c.encodeIfPresent(category, forKey: .category)
        try c.encodeIfPresent(archivedAt, forKey: .archivedAt)
        try c.encodeIfPresent(reason, forKey: .reason)
        try c.encodeIfPresent(sizeBytes, forKey: .sizeBytes)
        try c.encodeIfPresent(path, forKey: .path)
    }

    /// "4.4 KB" / "1.2 MB" / "—" for nil. Uses the SI byte formatter so
    /// the labels match what Finder shows.
    public var sizeLabel: String {
        guard let bytes = sizeBytes else { return "—" }
        let formatter = ByteCountFormatter()
        formatter.allowedUnits = [.useAll]
        formatter.countStyle = .file
        return formatter.string(fromByteCount: Int64(bytes))
    }

    /// `2026-04-22` (ISO date prefix) / "—". Hermes returns full ISO
    /// timestamps with seconds + Z; the date prefix is what the user
    /// actually wants in the archived list.
    public var archivedAtLabel: String {
        guard let iso = archivedAt, !iso.isEmpty else { return "—" }
        // Trim to date prefix if it looks like a full ISO timestamp.
        if let tIdx = iso.firstIndex(of: "T") {
            return String(iso[..<tIdx])
        }
        return iso
    }
}

/// One skill a `hermes curator prune` run would bulk-archive — an
/// agent-created skill idle for at least the chosen threshold. Archiving is
/// reversible (Restore), so this is a tidy-up, not a deletion.
public struct CuratorPruneCandidate: Sendable, Equatable, Identifiable {
    public var id: String { name }
    public let name: String
    public let idleDays: Int

    public init(name: String, idleDays: Int) {
        self.name = name
        self.idleDays = idleDays
    }

    /// "idle 412d" — compact right-aligned column for the confirm sheet.
    public var idleLabel: String { "idle \(idleDays)d" }
}

/// Result of `hermes curator prune [--days N] --dry-run` — the agent-created
/// skills idle ≥ `days` that a real prune would **bulk-archive** (reversibly;
/// it is NOT a disk deletion). Parsed from the CLI's text output
/// (`curator: N skill(s) idle >= Nd:` then `  <name> idle Nd` rows); Hermes
/// has no `--json` for this verb. `days` is threaded through from the request.
public struct CuratorPruneSummary: Sendable, Equatable {
    public let candidates: [CuratorPruneCandidate]
    public let days: Int
    public var count: Int { candidates.count }

    public init(candidates: [CuratorPruneCandidate], days: Int) {
        self.candidates = candidates
        self.days = days
    }
}
