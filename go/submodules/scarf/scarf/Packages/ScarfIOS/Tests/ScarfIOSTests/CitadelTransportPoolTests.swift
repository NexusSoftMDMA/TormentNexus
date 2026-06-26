#if canImport(Citadel)

import Testing
import Foundation
import ScarfCore
@testable import ScarfIOS

/// Build a transport whose `keyProvider` would throw if anyone tried to
/// connect — these tests never do, which is the point. File-scope (not a
/// suite method) so it's callable from the `@Sendable` task-group closures
/// without capturing the suite.
private func makeTestTransport(id: ServerID, config: SSHConfig) -> CitadelServerTransport {
    CitadelServerTransport(
        contextID: id,
        config: config,
        displayName: "test",
        keyProvider: { throw CancellationError() }
    )
}

private func testConfig(remoteHome: String? = nil) -> SSHConfig {
    SSHConfig(host: "example.test", user: "alan", remoteHome: remoteHome)
}

/// Behavioral coverage for `CitadelTransportPool` — the per-`(ServerID,
/// SSHConfig)` reuse layer that stops ScarfGo from opening a brand-new SSH
/// handshake on every file read / exec (gh#112 root cause). These tests
/// exercise the pool's contract WITHOUT a live SSH server: a
/// `CitadelServerTransport` opens its connection lazily, so constructing one
/// and never touching it never dials out, and `close()` on a never-connected
/// transport is a no-op. The `make` closure is the seam we count against.
@Suite struct CitadelTransportPoolTests {

    /// Thread-safe creation counter so the concurrency test can assert the
    /// pool created exactly one transport under contention.
    private final class Counter: @unchecked Sendable {
        private let lock = NSLock()
        private var n = 0
        func bump() { lock.withLock { n += 1 } }
        var value: Int { lock.withLock { n } }
    }

    @Test func reusesOneTransportForSameServerAndConfig() {
        let pool = CitadelTransportPool()
        let id = ServerID()
        let cfg = testConfig()
        let counter = Counter()

        let first = pool.transport(for: id, config: cfg) { counter.bump(); return makeTestTransport(id: id, config: cfg) }
        let second = pool.transport(for: id, config: cfg) { counter.bump(); return makeTestTransport(id: id, config: cfg) }

        #expect(counter.value == 1)                       // created once
        #expect(pool.pooledCount == 1)
        #expect((first as? CitadelServerTransport) === (second as? CitadelServerTransport))
    }

    @Test func profileSwitchReplacesTheConnection() {
        // A #120 profile switch re-points the phone via SSHConfig.remoteHome.
        // The pool must treat the new config as a new connection.
        let pool = CitadelTransportPool()
        let id = ServerID()
        let defaultCfg = testConfig(remoteHome: nil)
        let gatewayCfg = testConfig(remoteHome: "/home/alan/.hermes/profiles/gateway")
        let counter = Counter()

        let onDefault = pool.transport(for: id, config: defaultCfg) { counter.bump(); return makeTestTransport(id: id, config: defaultCfg) }
        let onGateway = pool.transport(for: id, config: gatewayCfg) { counter.bump(); return makeTestTransport(id: id, config: gatewayCfg) }

        #expect(counter.value == 2)                       // re-created on config change
        #expect(pool.pooledCount == 1)                    // but still one entry per server
        #expect((onDefault as? CitadelServerTransport) !== (onGateway as? CitadelServerTransport))
    }

    @Test func distinctServersCoexist() {
        let pool = CitadelTransportPool()
        let a = ServerID()
        let b = ServerID()
        let cfg = testConfig()

        _ = pool.transport(for: a, config: cfg) { makeTestTransport(id: a, config: cfg) }
        _ = pool.transport(for: b, config: cfg) { makeTestTransport(id: b, config: cfg) }

        #expect(pool.pooledCount == 2)
    }

    @Test func evictDropsOneAndForcesRecreate() async {
        let pool = CitadelTransportPool()
        let id = ServerID()
        let cfg = testConfig()
        let counter = Counter()

        _ = pool.transport(for: id, config: cfg) { counter.bump(); return makeTestTransport(id: id, config: cfg) }
        await pool.evict(id)
        #expect(pool.pooledCount == 0)

        _ = pool.transport(for: id, config: cfg) { counter.bump(); return makeTestTransport(id: id, config: cfg) }
        #expect(counter.value == 2)                       // recreated after eviction
        #expect(pool.pooledCount == 1)
    }

    @Test func evictAllClearsEverything() async {
        let pool = CitadelTransportPool()
        let cfg = testConfig()
        for _ in 0..<3 {
            let id = ServerID()
            _ = pool.transport(for: id, config: cfg) { makeTestTransport(id: id, config: cfg) }
        }
        #expect(pool.pooledCount == 3)
        await pool.evictAll()
        #expect(pool.pooledCount == 0)
    }

    /// The load-bearing thread-safety test: many concurrent callers racing on
    /// the same key must still create exactly ONE transport and all receive
    /// the same instance. This is the contract that lets the pool replace the
    /// churn safely.
    @Test func concurrentCallsCreateExactlyOneInstance() async {
        let pool = CitadelTransportPool()
        let id = ServerID()
        let cfg = testConfig()
        let counter = Counter()

        let identities = await withTaskGroup(of: ObjectIdentifier.self) { group -> Set<ObjectIdentifier> in
            for _ in 0..<64 {
                group.addTask {
                    let t = pool.transport(for: id, config: cfg) {
                        counter.bump()
                        return makeTestTransport(id: id, config: cfg)
                    }
                    return ObjectIdentifier(t as! CitadelServerTransport)
                }
            }
            var seen = Set<ObjectIdentifier>()
            for await oid in group { seen.insert(oid) }
            return seen
        }

        #expect(counter.value == 1)                       // no duplicate creation under the race
        #expect(identities.count == 1)                    // everyone got the same transport
        #expect(pool.pooledCount == 1)
    }
}

#endif // canImport(Citadel)
