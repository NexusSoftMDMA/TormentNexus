import Foundation

/// A structured signal Hermes emits when it observes worker / task
/// distress. Hermes v0.13 introduced a generic diagnostics engine that
/// attaches these to a task (cross-run signals) and/or a run (per-attempt
/// signals). Pre-v0.13 hosts never emit diagnostics so the array decodes
/// empty and downstream UI no-ops.
///
/// **Wire shape (best inference from release notes — verify against live
/// JSON during integration):** an array of objects with `kind`, optional
/// `message`, optional `detected_at` (ISO-8601 string OR Unix integer,
/// matching the rest of `HermesKanbanTask`'s timestamp tolerance).
///
/// **Forward compat:** `kind` stays a `String` so a future Hermes can
/// add new diagnostic kinds without a Scarf release. `KanbanDiagnosticKind`
/// is the typed mirror — it falls back to `.unknown` for unrecognized
/// kinds and renders the raw string verbatim.
public struct HermesKanbanDiagnostic: Sendable, Equatable, Identifiable, Codable {
    /// Synthetic id — not on the wire. Lets SwiftUI `ForEach` over a
    /// diagnostic array without forcing a deterministic id from the
    /// server (Hermes doesn't currently mint one).
    public let id: UUID
    /// Wire-side `kind` string. Compared case-insensitively via
    /// `KanbanDiagnosticKind.from(_:)`.
    public let kind: String
    /// Human-friendly elaboration ("no heartbeat for 4m20s", "exit code
    /// 0 with no complete call", etc.). May be nil; render the raw
    /// `kind` then.
    public let message: String?
    /// ISO-8601 string. Decoder accepts Unix integer seconds (Hermes's
    /// SQLite-backed shape) and converts to ISO-8601 so consumers see
    /// one type — same pattern as `HermesKanbanTask.decodeFlexibleTimestamp`.
    public let detectedAt: String?

    public init(
        kind: String,
        message: String? = nil,
        detectedAt: String? = nil
    ) {
        self.id = UUID()
        self.kind = kind
        self.message = message
        self.detectedAt = detectedAt
    }

    enum CodingKeys: String, CodingKey {
        case kind
        case message
        case detectedAt = "detected_at"
    }

    public init(from decoder: any Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.id = UUID()
        self.kind = try c.decodeIfPresent(String.self, forKey: .kind) ?? "unknown"
        self.message = try c.decodeIfPresent(String.self, forKey: .message)
        // Flexible timestamp decode mirrors HermesKanbanTask's pattern.
        if !c.contains(.detectedAt) {
            self.detectedAt = nil
        } else if let unix = try? c.decodeIfPresent(Double.self, forKey: .detectedAt) {
            let date = Date(timeIntervalSince1970: unix)
            self.detectedAt = Self.isoFormatter.string(from: date)
        } else {
            self.detectedAt = try c.decodeIfPresent(String.self, forKey: .detectedAt)
        }
    }

    public func encode(to encoder: any Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encode(kind, forKey: .kind)
        try c.encodeIfPresent(message, forKey: .message)
        try c.encodeIfPresent(detectedAt, forKey: .detectedAt)
    }

    public static func == (lhs: HermesKanbanDiagnostic, rhs: HermesKanbanDiagnostic) -> Bool {
        // Compare on wire fields, not synthetic id — round-trip decoding
        // mints fresh ids.
        lhs.kind == rhs.kind
            && lhs.message == rhs.message
            && lhs.detectedAt == rhs.detectedAt
    }

    private static let isoFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        return f
    }()
}

// MARK: - Typed mirror

/// Typed view of `HermesKanbanDiagnostic.kind`. Models keep the raw
/// string for forward compatibility; UI helpers read this enum to pick
/// the right glyph + tint without string-matching at every callsite.
///
/// `unknown` is the fallback for any kind a future Hermes adds that
/// Scarf doesn't recognize. Views render the raw string verbatim in
/// that case so the user still sees what Hermes flagged.
public enum KanbanDiagnosticKind: String, Sendable, CaseIterable {
    case heartbeatStalled = "heartbeat_stalled"
    case toolErrorLoop = "tool_error_loop"
    case retryCapHit = "retry_cap_hit"
    case unboundedRetry = "unbounded_retry"
    case darwinZombieDetected = "darwin_zombie_detected"
    case spawnFailure = "spawn_failure"
    case workerExitNoComplete = "worker_exit_no_complete"
    case unknown

    /// Map a wire string (case-insensitive) to a typed kind. Unknown
    /// values fall through to `.unknown` so callers can still surface
    /// the raw string.
    public static func from(_ raw: String) -> KanbanDiagnosticKind {
        KanbanDiagnosticKind(rawValue: raw.lowercased()) ?? .unknown
    }

    /// SF Symbol name to render alongside the diagnostic. View code
    /// reaches through the typed enum so glyph choices live in one
    /// place.
    public var glyphName: String {
        switch self {
        case .heartbeatStalled:     return "waveform.path.badge.minus"
        case .toolErrorLoop:        return "arrow.triangle.2.circlepath.exclamationmark"
        case .retryCapHit:          return "nosign"
        case .unboundedRetry:       return "arrow.clockwise.circle.fill"
        case .darwinZombieDetected: return "apple.logo"
        case .spawnFailure:         return "bolt.slash"
        case .workerExitNoComplete: return "figure.walk.departure"
        case .unknown:              return "stethoscope"
        }
    }

    /// Severity tier for this kind — drives badge tint. `.danger` for
    /// terminal-class signals (retry cap hit, zombie, spawn failure);
    /// `.warning` for recoverable signals (heartbeat stalled, tool
    /// error loop); `.neutral` only for unknown / forward-compat kinds.
    public var severity: DiagnosticSeverity {
        switch self {
        case .retryCapHit, .darwinZombieDetected, .spawnFailure:
            return .danger
        case .heartbeatStalled, .toolErrorLoop, .unboundedRetry, .workerExitNoComplete:
            return .warning
        case .unknown:
            return .neutral
        }
    }

    public enum DiagnosticSeverity: Sendable {
        case warning
        case danger
        case neutral
    }
}
