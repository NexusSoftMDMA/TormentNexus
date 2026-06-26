import Testing
import Foundation
@testable import ScarfCore

/// Parser tests for the `hermes gateway list` text table (Hermes v0.16 has
/// no `--json` flag). Pure — no transport, no process calls.
@Suite struct HermesGatewayListServiceTests {

    @Test func parsesLiveSampleOneRunningTwoStopped() {
        // The exact live v0.16 output: 1 running default + 2 stopped.
        let text = """
        Gateways:
          ✓ default (current)        — PID 44417
          ✗ scarfbox-smoke           — not running
          ✗ scarfbox-test            — not running
        """
        let snap = HermesGatewayListService.parse(text)
        #expect(snap?.profiles.count == 3)

        #expect(snap?.profiles[0].profile == "default")
        #expect(snap?.profiles[0].isRunning == true)
        #expect(snap?.profiles[0].pid == 44417)
        #expect(snap?.profiles[0].platforms.isEmpty == true)

        #expect(snap?.profiles[1].profile == "scarfbox-smoke")
        #expect(snap?.profiles[1].isRunning == false)
        #expect(snap?.profiles[1].pid == nil)

        #expect(snap?.profiles[2].profile == "scarfbox-test")
        #expect(snap?.profiles[2].isRunning == false)
        #expect(snap?.profiles[2].pid == nil)
    }

    @Test func parsesSingleRunningProfile() {
        let text = """
        Gateways:
          ✓ default        — PID 1234
        """
        let snap = HermesGatewayListService.parse(text)
        #expect(snap?.profiles.count == 1)
        #expect(snap?.profiles[0].profile == "default")
        #expect(snap?.profiles[0].pid == 1234)
        #expect(snap?.profiles[0].isRunning == true)
        #expect(snap?.profiles[0].platforms.isEmpty == true)
    }

    @Test func parsesSingleStoppedProfile() {
        let text = """
        Gateways:
          ✗ default        — not running
        """
        let snap = HermesGatewayListService.parse(text)
        #expect(snap?.profiles.count == 1)
        #expect(snap?.profiles[0].profile == "default")
        #expect(snap?.profiles[0].isRunning == false)
        #expect(snap?.profiles[0].pid == nil)
    }

    @Test func stripsCurrentMarkerFromProfileName() {
        // A `(current)` marker after the profile name must not leak into it.
        let text = """
        Gateways:
          ✓ work (current)        — PID 99
        """
        let snap = HermesGatewayListService.parse(text)
        #expect(snap?.profiles[0].profile == "work")
        #expect(snap?.profiles[0].pid == 99)
    }

    @Test func returnsNilOnEmptyString() {
        #expect(HermesGatewayListService.parse("") == nil)
    }

    @Test func returnsNilOnWhitespaceOnly() {
        #expect(HermesGatewayListService.parse("   \n  \n") == nil)
    }

    @Test func returnsNilOnHeaderOnlyNoProfiles() {
        // Just the header, no profile rows → no recognizable entries → nil.
        #expect(HermesGatewayListService.parse("Gateways:\n") == nil)
    }

    @Test func returnsNilOnGarbageInput() {
        #expect(HermesGatewayListService.parse("this is not gateway output") == nil)
    }

    // MARK: - headerDigest

    @Test func headerDigestEmptyProfiles() {
        let snap = GatewayListSnapshot(profiles: [])
        #expect(snap.headerDigest == "no profiles configured")
    }

    @Test func headerDigestSingleProfileRunning() {
        let snap = GatewayListSnapshot(profiles: [
            .init(profile: "default", isRunning: true, pid: 100,
                  platforms: ["slack", "telegram"])
        ])
        #expect(snap.headerDigest == "default profile · running · slack, telegram")
    }

    @Test func headerDigestSingleProfileStopped() {
        let snap = GatewayListSnapshot(profiles: [
            .init(profile: "default", isRunning: false, pid: nil, platforms: [])
        ])
        #expect(snap.headerDigest == "default profile · stopped")
    }

    @Test func headerDigestMultipleProfilesSomeRunning() {
        let snap = GatewayListSnapshot(profiles: [
            .init(profile: "work", isRunning: true, pid: 1, platforms: ["slack"]),
            .init(profile: "home", isRunning: false, pid: nil, platforms: ["matrix"]),
            .init(profile: "extra", isRunning: true, pid: 2, platforms: [])
        ])
        // 3 profiles total, 2 running, surface first running profile's
        // platform list as the highlight.
        #expect(snap.headerDigest == "3 profiles (2 running) · work: slack")
    }

    @Test func headerDigestMultipleProfilesNoneRunning() {
        let snap = GatewayListSnapshot(profiles: [
            .init(profile: "a", isRunning: false, pid: nil, platforms: ["slack"]),
            .init(profile: "b", isRunning: false, pid: nil, platforms: ["matrix"])
        ])
        // No running profile — fall back to the first profile's platforms.
        #expect(snap.headerDigest == "2 profiles (0 running) · a: slack")
    }
}
