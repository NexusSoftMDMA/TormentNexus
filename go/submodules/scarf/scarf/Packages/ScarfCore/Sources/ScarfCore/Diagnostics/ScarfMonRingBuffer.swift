import Foundation

/// Fixed-size, lock-protected ring of recent samples. Drives the in-app
/// Diagnostics panel and the export-as-JSON button.
///
/// Capacity is a compile-time choice; 4096 entries × ~80 bytes per sample =
/// ~320 KB resident. That's enough for several minutes of streaming-chat
/// activity at 200 samples/s without overwriting interesting context.
///
/// The hot path takes one `os_unfair_lock` per `record`. Aggregation (the
/// `summary(...)` reader) builds a fresh dictionary each call — only invoked
/// from the panel UI, which polls at a human cadence.
public final class ScarfMonRingBuffer: ScarfMonBackend, @unchecked Sendable {
    public let capacity: Int

    private let lock = ScarfMonLock()
    private var storage: [ScarfMon.Sample?]
    /// Next write index. Wraps around `capacity` so the buffer never grows.
    private var head: Int = 0
    /// True once we've wrapped at least once — switches the read order from
    /// `[0..<head]` to `[head..<capacity] + [0..<head]`.
    private var didWrap: Bool = false

    public init(capacity: Int = 4096) {
        precondition(capacity > 0, "ring buffer needs a positive capacity")
        self.capacity = capacity
        self.storage = Array(repeating: nil, count: capacity)
    }

    public func record(_ sample: ScarfMon.Sample) {
        lock.lock()
        defer { lock.unlock() }
        storage[head] = sample
        head += 1
        if head >= capacity {
            head = 0
            didWrap = true
        }
    }

    /// Snapshot of all currently-resident samples in chronological order.
    public func samples() -> [ScarfMon.Sample] {
        lock.lock()
        defer { lock.unlock() }
        if !didWrap {
            return storage[0..<head].compactMap { $0 }
        }
        let tail = storage[head..<capacity].compactMap { $0 }
        let leading = storage[0..<head].compactMap { $0 }
        return tail + leading
    }

    /// Wipe the buffer. Used by the "Reset" button in the Diagnostics
    /// panel and at the top of every test case.
    public func reset() {
        lock.lock()
        defer { lock.unlock() }
        for i in 0..<capacity { storage[i] = nil }
        head = 0
        didWrap = false
    }

    /// Aggregated stats over the current buffer. Buckets by
    /// `(category, name)`; computes count, total nanos, mean, p50, p95.
    public func summary() -> [ScarfMonStat] {
        let snapshot = samples()
        var buckets: [BucketKey: [UInt64]] = [:]
        var counts: [BucketKey: Int] = [:]
        var byteTotals: [BucketKey: Int] = [:]
        var kinds: [BucketKey: ScarfMon.Sample.Kind] = [:]

        for sample in snapshot {
            let key = BucketKey(category: sample.category, name: sample.name.description)
            kinds[key] = sample.kind
            counts[key, default: 0] += sample.count
            if let b = sample.bytes { byteTotals[key, default: 0] += b }
            if sample.kind == .interval {
                buckets[key, default: []].append(sample.durationNanos)
            }
        }

        var stats: [ScarfMonStat] = []
        for (key, _) in counts {
            let durations = buckets[key] ?? []
            let kind = kinds[key] ?? .event
            stats.append(ScarfMonStat(
                category: key.category,
                name: key.name,
                kind: kind,
                count: counts[key] ?? 0,
                totalNanos: durations.reduce(0, &+),
                p50Nanos: percentile(durations, 0.50),
                p95Nanos: percentile(durations, 0.95),
                maxNanos: durations.max() ?? 0,
                totalBytes: byteTotals[key] ?? 0
            ))
        }
        stats.sort { $0.p95Nanos > $1.p95Nanos }
        return stats
    }

    private struct BucketKey: Hashable {
        let category: ScarfMon.Category
        let name: String
    }

    private func percentile(_ values: [UInt64], _ p: Double) -> UInt64 {
        guard !values.isEmpty else { return 0 }
        let sorted = values.sorted()
        // Nearest-rank percentile — good enough for triage and avoids
        // interpolation edge cases on tiny samples.
        let rank = max(1, min(sorted.count, Int((p * Double(sorted.count)).rounded(.up))))
        return sorted[rank - 1]
    }
}

/// Per-bucket stats surfaced to the in-app panel.
public struct ScarfMonStat: Sendable, Hashable, Codable {
    public let category: ScarfMon.Category
    public let name: String
    public let kind: ScarfMon.Sample.Kind
    public let count: Int
    public let totalNanos: UInt64
    public let p50Nanos: UInt64
    public let p95Nanos: UInt64
    public let maxNanos: UInt64
    public let totalBytes: Int

    public var totalMs: Double { Double(totalNanos) / 1_000_000.0 }
    public var p50Ms: Double { Double(p50Nanos) / 1_000_000.0 }
    public var p95Ms: Double { Double(p95Nanos) / 1_000_000.0 }
    public var maxMs: Double { Double(maxNanos) / 1_000_000.0 }
}

// MARK: - JSON export

extension ScarfMonRingBuffer {
    /// Compact JSON dump for the "Copy as JSON" button. One line per sample
    /// keeps the output greppable when the user pastes it into a feedback
    /// thread.
    public func exportJSON() -> String {
        struct Wire: Codable {
            let category: String
            let name: String
            let kind: String
            let timestampMs: Double
            let durationNanos: UInt64
            let count: Int
            let bytes: Int?
        }
        let snapshot = samples()
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.sortedKeys]
        var lines: [String] = []
        lines.reserveCapacity(snapshot.count + 1)
        lines.append("[")
        for (i, s) in snapshot.enumerated() {
            let wire = Wire(
                category: s.category.rawValue,
                name: s.name.description,
                kind: s.kind.rawValue,
                timestampMs: s.timestamp.timeIntervalSince1970 * 1000,
                durationNanos: s.durationNanos,
                count: s.count,
                bytes: s.bytes
            )
            if let data = try? encoder.encode(wire),
               let line = String(data: data, encoding: .utf8) {
                let suffix = i == snapshot.count - 1 ? "" : ","
                lines.append("  " + line + suffix)
            }
        }
        lines.append("]")
        return lines.joined(separator: "\n")
    }
}
