import Testing
import Foundation
@testable import ScarfCore

/// Coverage for the per-(server, profile) scoping of the Skills "What's New"
/// snapshot baseline (issue #120 B4 follow-up). Before this, the baseline was
/// keyed by serverID alone (and on iOS by a FIXED UUID), so switching Hermes
/// profiles — each of which has its own `HERMES_HOME/skills` set — diffed the
/// new profile's skills against the previous profile's baseline and the pill
/// reported a bogus "N new / M changed".
///
/// All tests drive `InMemorySnapshotBackend`, which keys by the SAME composed
/// `(server, scope)` storage key as the production Mac/iOS backends — so the
/// namespacing behavior is exercised exactly, and the suite stays Linux-safe.
@Suite struct SkillSnapshotServiceTests {

    /// A skill whose snapshot signature depends only on `(id, files)` —
    /// the two fields `SkillSnapshotService.signatures(for:)` hashes.
    private func skill(_ id: String, _ files: [String] = ["SKILL.md"]) -> HermesSkill {
        HermesSkill(id: id, name: id, category: "test", path: "/s/\(id)", files: files, requiredConfig: [])
    }

    /// Mirrors both `SkillsView`s' render gate: the "What's New" pill shows
    /// ONLY for a real change against a NON-empty prior baseline. A
    /// first-ever (empty-baseline) diff primes silently — every skill reads
    /// as "new" there (so `hasChanges` is true), but the view suppresses the
    /// pill on `previousSnapshotEmpty`. Testing this predicate ties the
    /// assertions to the actual user-visible behavior.
    private func pillWouldShow(_ diff: SkillSnapshotDiff) -> Bool {
        diff.hasChanges && !diff.previousSnapshotEmpty
    }

    // MARK: - storageKey contract (the "no migration for default" guarantee)

    @Test func storageKeyIsBareUUIDForDefaultAndNamespacedForProfiles() {
        let id = ServerID()
        // nil / empty collapse to the bare per-server key — byte-for-byte the
        // pre-#120 key, so existing default-profile baselines keep resolving.
        #expect(SkillSnapshotService.storageKey(for: id, scope: nil) == id.uuidString)
        #expect(SkillSnapshotService.storageKey(for: id, scope: "") == id.uuidString)
        // A named profile gets a dot-suffixed namespace.
        #expect(SkillSnapshotService.storageKey(for: id, scope: "gateway") == "\(id.uuidString).gateway")
    }

    /// The service normalizes its `profile` argument, so `nil` / `"default"` /
    /// whitespace / an invalid name ALL resolve to the bare per-server
    /// baseline — i.e. they're interchangeable and backward-compatible.
    @Test func serviceCollapsesDefaultAndInvalidProfilesToTheSharedBaseline() {
        let backend = InMemorySnapshotBackend()
        let id = ServerID()
        SkillSnapshotService(serverID: id, profile: nil, backend: .inMemory(backend))
            .markSeen([skill("a")])

        for sentinel in [nil, "default", "  ", "Bad Name", "../escape"] as [String?] {
            let svc = SkillSnapshotService(serverID: id, profile: sentinel, backend: .inMemory(backend))
            #expect(svc.profile == nil, "profile \(String(describing: sentinel)) should normalize to nil")
            let diff = svc.diff(against: [skill("a")])
            #expect(!diff.previousSnapshotEmpty, "should read the shared default baseline")
            #expect(!diff.hasChanges)
        }
    }

    // MARK: - profile isolation (the core bug)

    @Test func differentProfilesKeepIndependentBaselines() {
        let backend = InMemorySnapshotBackend()
        let id = ServerID()
        let def = SkillSnapshotService(serverID: id, profile: nil, backend: .inMemory(backend))
        let gateway = SkillSnapshotService(serverID: id, profile: "gateway", backend: .inMemory(backend))

        // Default profile records its skills as seen.
        def.markSeen([skill("a"), skill("b")])

        // Switching to the gateway profile must NOT inherit the default's
        // baseline: its own is still empty, so the first diff primes silently
        // (previousSnapshotEmpty) instead of flashing a bogus "1 new" pill.
        // Had it shared the key, last would be {a,b} → previousSnapshotEmpty
        // false and the pill would render "1 new" for skill c.
        let gatewayDiff = gateway.diff(against: [skill("c")])
        #expect(gatewayDiff.previousSnapshotEmpty)
        #expect(!pillWouldShow(gatewayDiff))
    }

    @Test func switchingProfilesProducesNoFalseDiffAndSwitchBackPreservesBaseline() {
        let backend = InMemorySnapshotBackend()
        let id = ServerID()
        let def = SkillSnapshotService(serverID: id, profile: nil, backend: .inMemory(backend))
        let gateway = SkillSnapshotService(serverID: id, profile: "gateway", backend: .inMemory(backend))

        def.markSeen([skill("a"), skill("b")])      // default baseline = {a, b}
        gateway.markSeen([skill("c")])              // gateway baseline = {c}

        // Back on default with the SAME skills: no changes, baseline intact —
        // the gateway markSeen did not clobber it.
        let backToDefault = def.diff(against: [skill("a"), skill("b")])
        #expect(!backToDefault.previousSnapshotEmpty)
        #expect(!backToDefault.hasChanges)

        // Gateway likewise unchanged against its own {c}.
        let gatewayAgain = gateway.diff(against: [skill("c")])
        #expect(!gatewayAgain.previousSnapshotEmpty)
        #expect(!gatewayAgain.hasChanges)
    }

    // MARK: - cross-server isolation (the fixed-UUID bleed)

    @Test func differentServersKeepIndependentBaselines() {
        let backend = InMemorySnapshotBackend()
        let serverA = ServerID()
        let serverB = ServerID()

        SkillSnapshotService(serverID: serverA, backend: .inMemory(backend))
            .markSeen([skill("a")])

        // A different server's first diff must read its OWN (empty) baseline,
        // not server A's — the bug the fixed `...A1` context id caused on iOS.
        let bDiff = SkillSnapshotService(serverID: serverB, backend: .inMemory(backend))
            .diff(against: [skill("b")])
        #expect(bDiff.previousSnapshotEmpty)
        #expect(!pillWouldShow(bDiff))
    }

    // MARK: - the mechanism still detects real deltas within one profile

    @Test func sameProfileStillDetectsNewAndChangedSkills() {
        let backend = InMemorySnapshotBackend()
        let svc = SkillSnapshotService(serverID: ServerID(), profile: "gateway", backend: .inMemory(backend))

        svc.markSeen([skill("a", ["SKILL.md"])])
        // 'a' gained a file (signature changes → "changed") and 'b' is new.
        let diff = svc.diff(against: [skill("a", ["SKILL.md", "extra.py"]), skill("b")])
        #expect(!diff.previousSnapshotEmpty)
        #expect(diff.newCount == 1)
        #expect(diff.updatedCount == 1)
        #expect(diff.changedSkillIds == ["a", "b"])
    }
}
