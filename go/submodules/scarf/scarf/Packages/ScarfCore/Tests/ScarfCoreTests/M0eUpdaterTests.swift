import Testing
import Foundation
@testable import ScarfCore

/// Pure-function matrix for `HermesUpdaterCommandBuilder.updateArgv`. The
/// builder degrades flags silently when the connected host can't honor
/// them, so the "is the right flag emitted on the right version?" matrix
/// is the meaningful test surface.
@Suite struct M0eUpdaterTests {

    // MARK: - Helpers

    private func caps(_ versionLine: String?) -> HermesCapabilities {
        guard let line = versionLine else { return .empty }
        return HermesCapabilities.parseLine(line)
    }

    // MARK: - Pre-v0.12 (no flags supported)

    @Test func preV012_returnsBareUpdateRegardlessOfFlags() {
        let pre = caps("Hermes Agent v0.11.0 (2026.4.23)")
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: pre, unattended: false, checkOnly: false
        ) == ["update"])
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: pre, unattended: true, checkOnly: false
        ) == ["update"])
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: pre, unattended: true, checkOnly: true
        ) == ["update"])
    }

    @Test func unknownVersion_returnsBareUpdate() {
        // No detected version means we can't guarantee any flag is
        // honored; defensively emit the bare verb.
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: .empty, unattended: true, checkOnly: true
        ) == ["update"])
    }

    // MARK: - v0.12 (--check supported, --yes is not)

    @Test func v012_checkOnly_emitsCheckFlag() {
        let v012 = caps("Hermes Agent v0.12.0 (2026.4.30)")
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: v012, unattended: false, checkOnly: true
        ) == ["update", "--check"])
    }

    @Test func v012_unattended_dropsYesFlag() {
        // v0.12 doesn't honor --yes; the helper degrades silently.
        let v012 = caps("Hermes Agent v0.12.0 (2026.4.30)")
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: v012, unattended: true, checkOnly: false
        ) == ["update"])
    }

    @Test func v012_checkOnlyAndUnattended_emitsOnlyCheck() {
        let v012 = caps("Hermes Agent v0.12.0 (2026.4.30)")
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: v012, unattended: true, checkOnly: true
        ) == ["update", "--check"])
    }

    // MARK: - v0.13 (full flag support)

    @Test func v013_unattended_emitsYesFlag() {
        let v013 = caps("Hermes Agent v0.13.0 (2026.5.7)")
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: v013, unattended: true, checkOnly: false
        ) == ["update", "--yes"])
    }

    @Test func v013_checkOnlyAndUnattended_emitsBothFlags() {
        let v013 = caps("Hermes Agent v0.13.0 (2026.5.7)")
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: v013, unattended: true, checkOnly: true
        ) == ["update", "--check", "--yes"])
    }

    @Test func v013_neither_emitsBareUpdate() {
        let v013 = caps("Hermes Agent v0.13.0 (2026.5.7)")
        #expect(HermesUpdaterCommandBuilder.updateArgv(
            capabilities: v013, unattended: false, checkOnly: false
        ) == ["update"])
    }
}
