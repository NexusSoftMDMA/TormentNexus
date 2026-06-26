import Foundation
#if canImport(os)
import os
#endif

/// `os.Logger`-backed sink. Off by default — opt-in via the Diagnostics
/// settings toggle. Writes one `.debug` line per sample at the
/// `com.scarf.mon` subsystem, so users can stream the output via
/// `log stream --predicate 'subsystem == "com.scarf.mon"'` without
/// enabling private-data redaction overrides.
///
/// Only meaningful for users running their own debug build or with the
/// "verbose performance logging" toggle on.
public final class ScarfMonLoggerBackend: ScarfMonBackend, @unchecked Sendable {
    #if canImport(os)
    private let logger: Logger

    public init(category: String = "perf") {
        self.logger = Logger(subsystem: "com.scarf.mon", category: category)
    }

    public func record(_ sample: ScarfMon.Sample) {
        switch sample.kind {
        case .interval:
            // `\(static:)` interpolation keeps the StaticString out of the
            // private-data redaction path — names are public, durations
            // are public, the user's content never touches this channel.
            logger.debug(
                "\(sample.category.rawValue, privacy: .public) \(sample.name.description, privacy: .public) ms=\(Double(sample.durationNanos) / 1_000_000.0, privacy: .public)"
            )
        case .event:
            logger.debug(
                "\(sample.category.rawValue, privacy: .public) \(sample.name.description, privacy: .public) count=\(sample.count, privacy: .public) bytes=\(sample.bytes ?? -1, privacy: .public)"
            )
        }
    }
    #else
    public init(category: String = "perf") {}
    public func record(_ sample: ScarfMon.Sample) { /* no-op off-Apple */ }
    #endif
}
