import Foundation
#if canImport(os)
import os
import os.signpost
#endif

/// Always-on signpost backend. Emits an `os_signpost` event per sample so
/// users can attach Instruments and see Scarf's instrumentation in the
/// Points of Interest track without a debug build.
///
/// `os_signpost` is elided by the runtime when no Instruments session is
/// recording the relevant subsystem — the backend pays the cost of one
/// `OSLog` lookup per emit and nothing else.
public final class ScarfMonSignpostBackend: ScarfMonBackend, @unchecked Sendable {
    #if canImport(os)
    private let log: OSLog

    public init(subsystem: String = "com.scarf.mon") {
        self.log = OSLog(subsystem: subsystem, category: .pointsOfInterest)
    }

    public func record(_ sample: ScarfMon.Sample) {
        // Signposts want a `StaticString` name — we already require
        // exactly that on the API. Format string is also static; the
        // dynamic values flow as printf-style args, so no allocations
        // for the event name itself.
        switch sample.kind {
        case .interval:
            os_signpost(
                .event,
                log: log,
                name: sample.name,
                "category=%{public}@ ms=%{public}.3f count=%d",
                sample.category.rawValue,
                Double(sample.durationNanos) / 1_000_000.0,
                sample.count
            )
        case .event:
            os_signpost(
                .event,
                log: log,
                name: sample.name,
                "category=%{public}@ count=%d bytes=%d",
                sample.category.rawValue,
                sample.count,
                sample.bytes ?? -1
            )
        }
    }
    #else
    public init(subsystem: String = "com.scarf.mon") {}
    public func record(_ sample: ScarfMon.Sample) { /* no-op off-Apple */ }
    #endif
}
