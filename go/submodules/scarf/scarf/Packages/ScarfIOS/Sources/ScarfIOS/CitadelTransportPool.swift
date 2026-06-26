// Gated on `canImport(Citadel)` for the same reason as
// `CitadelServerTransport` — it stores that concrete type so it can close
// the underlying connection. Linux CI (no Citadel) skips the file.
#if canImport(Citadel)

import Foundation
import ScarfCore
#if canImport(os)
import os
#endif

/// Process-wide pool of one long-lived `CitadelServerTransport` per
/// `(ServerID, SSHConfig)`.
///
/// **Why this exists.** `ServerContext.makeTransport()` returns a FRESH
/// transport on every call, and each `CitadelServerTransport` owns its own
/// lazily-opened SSH connection. Without pooling, every file read and every
/// `runProcess` on iOS paid a brand-new `SSHClient.connect()` handshake, and
/// close-on-dealloc is fire-and-forget. A Settings load or a chat pre-flight
/// fans out a burst of those short-lived handshakes; under the churn
/// `SSHClient.connect()` starts to fail — the write path surfaces it as
/// "Transport refused the command" and reads silently degrade to "empty"
/// (gh#112). A single manual `ssh` never reproduces it because it opens one
/// connection.
///
/// Reusing one warm connection per server collapses the churn. The Mac app
/// never had this problem: its `SSHTransport` shells out to `/usr/bin/ssh`
/// with ControlMaster multiplexing. This pool is the iOS equivalent.
///
/// **Keying.** The key is `ServerID`; the stored `SSHConfig` is a staleness
/// check. A #120 profile switch re-points the phone at
/// `<base>/profiles/<name>` by changing `SSHConfig.remoteHome`, so the config
/// no longer matches and the pool transparently closes the old connection and
/// opens one for the new profile. Profile scoping is per-operation
/// (`HERMES_HOME` + `remoteHome`), so one connection per profile is correct;
/// the churn we're killing was per-*operation*, this replacement is per-*switch*
/// (rare, deliberate) — bounded either way.
///
/// **Concurrency.** `transport(for:config:make:)` is synchronous because the
/// `sshTransportFactory` it backs is synchronous; it guards the dictionary
/// with an `NSLock` (the `make` closure only constructs the transport — no I/O,
/// the connection opens lazily on first use — so it's safe under the lock).
/// `evict` / `evictAll` are async so they can await the connection close.
/// Mirrors the `UserHomeCache.shared` singleton + `ResultBox` lock patterns
/// already in this target.
public final class CitadelTransportPool: @unchecked Sendable {
    #if canImport(os)
    private static let logger = Logger(subsystem: "com.scarf", category: "CitadelTransportPool")
    #endif

    public static let shared = CitadelTransportPool()

    private struct Entry {
        let config: SSHConfig
        let transport: CitadelServerTransport
    }

    private let lock = NSLock()
    private var entries: [ServerID: Entry] = [:]

    /// `internal` (not `private`) so `@testable import ScarfIOS` can spin up
    /// throwaway pools without sharing the `.shared` singleton's state across
    /// tests. External modules can only reach `.shared`.
    init() {}

    /// Return the pooled transport for `id`, reusing the warm connection when
    /// the `SSHConfig` is unchanged. Creates one via `make` when absent, or
    /// when the stored config no longer matches (host / port / user /
    /// `remoteHome` change — including a #120 profile switch); the superseded
    /// transport's connection is closed off the caller's thread.
    ///
    /// `make` must construct (not connect) — the returned transport opens its
    /// connection lazily, so calling `make` under the lock costs nothing.
    public func transport(
        for id: ServerID,
        config: SSHConfig,
        make: () -> CitadelServerTransport
    ) -> any ServerTransport {
        // `make` only constructs (the connection opens lazily), so it's safe
        // to call inside the lock; the async `close()` of any superseded
        // transport happens after we release it.
        let (transport, superseded): (CitadelServerTransport, CitadelServerTransport?) = lock.withLock {
            if let existing = entries[id], existing.config == config {
                return (existing.transport, nil)
            }
            let superseded = entries[id]?.transport
            let fresh = make()
            entries[id] = Entry(config: config, transport: fresh)
            return (fresh, superseded)
        }

        if let superseded {
            #if canImport(os)
            Self.logger.debug("Replacing pooled transport for \(id, privacy: .public) (config changed)")
            #endif
            Task.detached { await superseded.close() }
        }
        return transport
    }

    /// Close and drop the pooled transport for one server. Call when leaving a
    /// server (soft disconnect) or forgetting it, so its connection doesn't
    /// linger after the user moved on.
    public func evict(_ id: ServerID) async {
        let entry = lock.withLock { entries.removeValue(forKey: id) }
        if let entry {
            await entry.transport.close()
        }
    }

    /// Close and drop every pooled transport. Call on full sign-out and when
    /// the app enters the background (iOS suspends the sockets anyway; this
    /// frees them deterministically and guarantees a clean reconnect on
    /// return). The ACP chat channel owns its own connection and handles its
    /// own scene-phase lifecycle, so it is unaffected.
    public func evictAll() async {
        let drained = lock.withLock { () -> [ServerID: Entry] in
            let copy = entries
            entries.removeAll()
            return copy
        }
        for (_, entry) in drained {
            await entry.transport.close()
        }
    }

    /// Test seam: number of live pooled entries.
    var pooledCount: Int {
        lock.withLock { entries.count }
    }
}

#endif // canImport(Citadel)
