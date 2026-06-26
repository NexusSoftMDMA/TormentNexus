import Foundation
import MetricKit
import os

/// MetricKit subscriber that persists crash + hang diagnostic payloads
/// to the app's Documents directory so the user can share them on the
/// next launch.
///
/// **Why this exists.** TestFlight feedback entries arrive without
/// stack traces — five reports landed (2026-05-12 → 2026-05-15) with
/// only one-line comments and `appUptimeMillis: 4-5s`. Without a
/// symbolicated trace we can guess at root causes (LazyVStack mid-
/// scroll, background watchdog, low-memory resume OOM), but we can't
/// confirm. MetricKit gives us the on-device crash log Apple already
/// captures, persisted before the user re-launches, with a ~24-hour
/// delivery cadence after the next launch.
///
/// **Lifecycle.** Apple delivers payloads via `didReceive(
/// [MXDiagnosticPayload])` shortly after app start. We write each
/// payload's `jsonRepresentation()` to
/// `Documents/ScarfDiagnostics/<timestamp>-<kind>.json` and let the
/// Settings → "Share Latest Diagnostic" affordance pick the most
/// recent file.
///
/// **Privacy.** `MXDiagnosticPayload` contains a stack trace (function
/// names + offsets), thread state, and process metadata. No user
/// content (chat text, paths, credentials). The persisted file lives
/// inside the app's Documents container and never leaves the device
/// unless the user actively shares it via the Settings row.
///
/// **Actor isolation.** Not `@MainActor`-isolated: the protocol's
/// callbacks are `nonisolated` (MetricKit may deliver from any
/// queue), and everything this type does is file IO + a Logger
/// write. Keeping it isolation-free avoids a maze of `Sendable`
/// crossings.
final class MetricKitSubscriber: NSObject, MXMetricManagerSubscriber, @unchecked Sendable {
    private static let logger = Logger(
        subsystem: "com.scarf.ios",
        category: "MetricKit"
    )

    static let shared = MetricKitSubscriber()

    private override init() {
        super.init()
        MXMetricManager.shared.add(self)
        Self.logger.info("MetricKit subscriber registered")
    }

    // MARK: - MXMetricManagerSubscriber

    func didReceive(_ payloads: [MXMetricPayload]) {
        // Metric payloads (CPU/memory/launch-time aggregates) are
        // useful for fleet-level perf work, not individual triage.
        // Drop them on the floor; the subscription itself costs
        // nothing.
        Self.logger.debug("Received \(payloads.count) MXMetricPayload(s); ignored")
    }

    func didReceive(_ payloads: [MXDiagnosticPayload]) {
        Self.logger.info("Received \(payloads.count) MXDiagnosticPayload(s)")
        for payload in payloads {
            Self.persist(payload: payload)
        }
    }

    // MARK: - Persistence

    /// Directory containing persisted diagnostics. Created lazily.
    static var diagnosticsDirectory: URL {
        let docs = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first
            ?? URL(fileURLWithPath: NSHomeDirectory())
        return docs.appendingPathComponent("ScarfDiagnostics", isDirectory: true)
    }

    /// Most-recently-modified diagnostic file on disk, or nil when
    /// none has been received yet. Drives the Settings affordance —
    /// nothing to share = hide the row.
    static func mostRecentDiagnosticFile() -> URL? {
        let dir = diagnosticsDirectory
        guard let entries = try? FileManager.default.contentsOfDirectory(
            at: dir,
            includingPropertiesForKeys: [.contentModificationDateKey],
            options: [.skipsHiddenFiles]
        ) else {
            return nil
        }
        return entries
            .filter { $0.pathExtension == "json" }
            .map { url -> (URL, Date) in
                let mtime = (try? url.resourceValues(forKeys: [.contentModificationDateKey])
                    .contentModificationDate) ?? .distantPast
                return (url, mtime)
            }
            .max(by: { $0.1 < $1.1 })?
            .0
    }

    /// Bounded cleanup: keep at most this many on-disk payloads so a
    /// chronic crash loop doesn't pin disk space. MetricKit delivers
    /// once per day at most, so 30 files = a month of history.
    private static let maxPersistedFiles = 30

    private static func persist(payload: MXDiagnosticPayload) {
        // Off-MainActor write — JSON encoding + file IO.
        let json = payload.jsonRepresentation()
        let dir = diagnosticsDirectory
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [
            .withInternetDateTime,
            .withDashSeparatorInDate,
            .withColonSeparatorInTime,
        ]
        let stamp = formatter.string(from: payload.timeStampEnd)
            .replacingOccurrences(of: ":", with: "-")
        let crashCount = payload.crashDiagnostics?.count ?? 0
        let hangCount = payload.hangDiagnostics?.count ?? 0
        let kind: String
        if crashCount > 0 { kind = "crash" }
        else if hangCount > 0 { kind = "hang" }
        else { kind = "diagnostic" }
        let filename = "\(stamp)-\(kind).json"

        do {
            try FileManager.default.createDirectory(
                at: dir,
                withIntermediateDirectories: true
            )
            let url = dir.appendingPathComponent(filename)
            try json.write(to: url, options: .atomic)
            logger.info(
                "Persisted \(kind, privacy: .public) diagnostic (\(json.count) bytes) to \(url.lastPathComponent, privacy: .public)"
            )
            prune(in: dir)
        } catch {
            logger.error(
                "Failed to persist diagnostic: \(error.localizedDescription, privacy: .public)"
            )
        }
    }

    /// Trim the persisted-file set down to `maxPersistedFiles`.
    /// Deletes the oldest by modification time. Idempotent; safe to
    /// call on every payload arrival.
    private static func prune(in dir: URL) {
        guard let entries = try? FileManager.default.contentsOfDirectory(
            at: dir,
            includingPropertiesForKeys: [.contentModificationDateKey],
            options: [.skipsHiddenFiles]
        ) else { return }
        let jsonFiles = entries.filter { $0.pathExtension == "json" }
        guard jsonFiles.count > maxPersistedFiles else { return }
        let dated = jsonFiles.map { url -> (URL, Date) in
            let mtime = (try? url.resourceValues(forKeys: [.contentModificationDateKey])
                .contentModificationDate) ?? .distantPast
            return (url, mtime)
        }
        let sortedByAge = dated.sorted { $0.1 < $1.1 }
        let toDelete = sortedByAge.prefix(jsonFiles.count - maxPersistedFiles)
        for (url, _) in toDelete {
            try? FileManager.default.removeItem(at: url)
        }
    }
}
