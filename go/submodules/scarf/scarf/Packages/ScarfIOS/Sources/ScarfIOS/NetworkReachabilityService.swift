import Foundation
import Network
import Observation
#if canImport(os)
import os
#endif

/// Process-wide reachability monitor wrapping `NWPathMonitor`. Used by
/// `ChatController` to decide when to attempt a reconnect (on
/// `.satisfied`) vs. mark the chat offline (on `.unsatisfied`).
///
/// Singleton because `NWPathMonitor` is per-process by design — there's
/// no benefit to instantiating multiple monitors and the cost (a small
/// background queue per instance) accumulates if every controller
/// spawns its own.
///
/// ## Usage
///
/// Don't read the published state from a SwiftUI view body — the
/// runtime samples through `NWPathMonitor`'s queue, but a `body`
/// re-evaluation that touches `currentPath` directly would block. Read
/// `isSatisfied` / observe `transitionTick` instead. Tests and
/// non-iOS callers can use the no-op default behavior (`isSatisfied`
/// reports `true`).
@Observable
@MainActor
public final class NetworkReachabilityService {
    public static let shared = NetworkReachabilityService()

    /// `true` when the OS reports a usable network path (any
    /// interface). Inverted via `!isSatisfied` for "we're offline."
    public private(set) var isSatisfied: Bool = true

    /// Mirrors `NWPath.isExpensive`. Useful as a hint to UI for not
    /// auto-fetching big payloads on cellular. Not consumed yet —
    /// reserved so callers don't have to add another property later.
    public private(set) var isExpensive: Bool = false

    /// Monotonic counter that bumps every time `isSatisfied` changes.
    /// Views observe `transitionTick` rather than `isSatisfied` to
    /// kick a `.onChange` even if the value is the same as before
    /// (rare but possible during rapid network flapping).
    public private(set) var transitionTick: Int = 0

    private let monitor = NWPathMonitor()
    private let queue = DispatchQueue(label: "com.scarf.ios.reachability")

    #if canImport(os)
    private static let logger = Logger(subsystem: "com.scarf.ios", category: "NetworkReachability")
    #endif

    private init() {
        // Seed from the current path synchronously so first reads on
        // launch don't show "satisfied" while the OS reports otherwise.
        // `currentPath` is safe here at init (the monitor hasn't been
        // started yet, no queue handler is firing).
        let initial = monitor.currentPath
        self.isSatisfied = (initial.status == .satisfied)
        self.isExpensive = initial.isExpensive

        monitor.pathUpdateHandler = { [weak self] path in
            // Bounce back through MainActor — the `Observable`
            // protocol's published-property invariants require main-
            // thread mutation. The pathUpdateHandler is invoked on
            // `queue`, which is a private background queue.
            Task { @MainActor in
                guard let self else { return }
                let satisfied = (path.status == .satisfied)
                if self.isSatisfied != satisfied {
                    self.isSatisfied = satisfied
                    self.transitionTick &+= 1
                    #if canImport(os)
                    Self.logger.info(
                        "Reachability transition: \(satisfied ? "satisfied" : "unsatisfied", privacy: .public)"
                    )
                    #endif
                }
                self.isExpensive = path.isExpensive
            }
        }
        monitor.start(queue: queue)
    }

    deinit {
        // Singleton is process-lifetime; this only runs on shutdown.
        monitor.cancel()
    }
}
