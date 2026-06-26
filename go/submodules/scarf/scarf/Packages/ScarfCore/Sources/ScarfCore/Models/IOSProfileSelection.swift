import Foundation

/// Per-server selected Hermes profile for ScarfGo (issue #120, Design B).
///
/// Stores only a profile *name* per `ServerID` (or nothing = the default /
/// root profile). The name is combined with the server's base home by
/// `HermesProfileScope.resolveHome` to scope every read/write, and passed
/// to `hermes -p <name>` to scope chat/CLI — WITHOUT mutating the host's
/// `active_profile`.
///
/// Sync (not async) on purpose: reads happen while building the view tree,
/// the payload is a tiny string map, and there is no migration to perform —
/// the same low-ceremony shape as SwiftUI's `@AppStorage`. The protocol and
/// the in-memory implementation live in ScarfCore (Linux-testable, mirroring
/// `IOSServerConfigStore`); the `UserDefaults` concrete implementation lives
/// in ScarfIOS.
public protocol IOSProfileSelectionStore: Sendable {
    /// Selected profile name for a server, or `nil` for the default
    /// (root) profile. Always returns a normalized value.
    func selectedProfile(for id: ServerID) -> String?

    /// Set (a valid name) or clear (`nil`/`"default"`/invalid → default)
    /// the selected profile for a server.
    func setSelectedProfile(_ name: String?, for id: ServerID)
}

/// In-memory store for tests and previews. Thread-safe via a lock so it
/// satisfies `Sendable` and matches production call patterns.
public final class InMemoryProfileSelectionStore: IOSProfileSelectionStore, @unchecked Sendable {
    private let lock = NSLock()
    private var storage: [ServerID: String] = [:]

    public init() {}

    public func selectedProfile(for id: ServerID) -> String? {
        lock.withLock { HermesProfileScope.normalize(storage[id]) }
    }

    public func setSelectedProfile(_ name: String?, for id: ServerID) {
        lock.withLock {
            if let normalized = HermesProfileScope.normalize(name) {
                storage[id] = normalized
            } else {
                storage.removeValue(forKey: id)
            }
        }
    }
}
