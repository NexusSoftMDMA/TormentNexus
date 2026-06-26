import Testing
import Foundation
@testable import ScarfCore

/// Issue #79 regression. `searchHub()` with `hubSource == "all"` must
/// filter the cached browse list client-side (instead of shelling out
/// to `hermes skills search`, which routes through Hermes's
/// centralized index and can miss skills that browse aggregates from
/// non-indexed registries — `honcho` was the user-reported example).
///
/// Source-specific searches keep the CLI path; that's not exercised
/// here because it requires a live `hermes` binary — the existing
/// HermesSkillsHubParser tests cover the parser side.
@Suite("SkillsViewModel hub filter")
@MainActor
struct SkillsViewModelHubFilterTests {

    private func makeViewModel() -> SkillsViewModel {
        SkillsViewModel(context: .local)
    }

    private let stubBrowse: [HermesHubSkill] = [
        HermesHubSkill(
            identifier: "honcho",
            name: "honcho",
            description: "Memory provider for chat-scoped facts.",
            source: "github"
        ),
        HermesHubSkill(
            identifier: "1password",
            name: "1password",
            description: "Set up and use 1Password integration.",
            source: "official"
        ),
        HermesHubSkill(
            identifier: "spotify",
            name: "spotify",
            description: "Spotify skill — playback control via OAuth.",
            source: "official"
        ),
    ]

    @Test func allSourcesFilterMatchesByName() {
        let vm = makeViewModel()
        vm.lastBrowseResults = stubBrowse
        vm.hubSource = "all"
        vm.hubQuery = "honcho"
        vm.searchHub()
        #expect(vm.hubResults.count == 1)
        #expect(vm.hubResults.first?.identifier == "honcho")
        #expect(vm.isHubLoading == false)
        #expect(vm.hubMessage == nil)
    }

    @Test func allSourcesFilterMatchesByDescription() {
        let vm = makeViewModel()
        vm.lastBrowseResults = stubBrowse
        vm.hubSource = "all"
        vm.hubQuery = "OAuth"
        vm.searchHub()
        #expect(vm.hubResults.count == 1)
        #expect(vm.hubResults.first?.identifier == "spotify")
    }

    @Test func allSourcesFilterIsCaseInsensitive() {
        let vm = makeViewModel()
        vm.lastBrowseResults = stubBrowse
        vm.hubSource = "all"
        vm.hubQuery = "HONCHO"
        vm.searchHub()
        #expect(vm.hubResults.count == 1)
        #expect(vm.hubResults.first?.identifier == "honcho")
    }

    @Test func allSourcesFilterEmptyMatchSetsMessage() {
        let vm = makeViewModel()
        vm.lastBrowseResults = stubBrowse
        vm.hubSource = "all"
        vm.hubQuery = "ringtone"
        vm.searchHub()
        #expect(vm.hubResults.isEmpty)
        #expect(vm.hubMessage == "No matches")
    }

    /// Empty query should fall through to `browseHub()`, which on
    /// `.local` with no Hermes installed will set isHubLoading=true
    /// and not block the test. We just assert the early-return guard
    /// kicked in by checking the cache was untouched.
    @Test func emptyQueryFallsThroughToBrowse() {
        let vm = makeViewModel()
        vm.lastBrowseResults = stubBrowse
        vm.hubSource = "all"
        vm.hubQuery = ""
        let cacheBefore = vm.lastBrowseResults
        vm.searchHub()
        #expect(vm.lastBrowseResults == cacheBefore)
    }
}
