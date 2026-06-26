import Testing
import Foundation
@testable import ScarfCore

/// Pure parser tests for the toolset-gating detector. The actor's
/// `detect()` integration path runs against a real `ServerContext`
/// and is exercised by manual verification — these tests freeze the
/// YAML → list classification logic so future config-syntax drift
/// surfaces here first.
@Suite struct KanbanToolsetDetectorTests {

    // MARK: - Top-level toolsets

    @Test func topLevelToolsetsBlockExtractsItems() {
        let yaml = """
        version: 1
        toolsets:
          - hermes-cli
          - kanban
          - web
        agent:
          max_turns: 90
        """
        let items = KanbanToolsetDetector.parseTopLevelToolsets(yaml: yaml)
        #expect(items == ["hermes-cli", "kanban", "web"])
    }

    @Test func topLevelToolsetsAbsentReturnsEmpty() {
        let yaml = """
        version: 1
        platform_toolsets:
          cli:
            - browser
        """
        let items = KanbanToolsetDetector.parseTopLevelToolsets(yaml: yaml)
        #expect(items.isEmpty)
    }

    @Test func topLevelToolsetsToleratesQuotedItems() {
        let yaml = """
        toolsets:
          - "kanban"
          - 'web'
        """
        let items = KanbanToolsetDetector.parseTopLevelToolsets(yaml: yaml)
        #expect(items == ["kanban", "web"])
    }

    @Test func topLevelToolsetsStopsAtNextTopLevelKey() {
        // Block ends when an unindented non-list line appears. Keeps
        // the parser from absorbing items from a sibling block that
        // happens to follow without a blank line.
        let yaml = """
        toolsets:
          - hermes-cli
        platform_toolsets:
          cli:
            - kanban
        """
        let items = KanbanToolsetDetector.parseTopLevelToolsets(yaml: yaml)
        #expect(items == ["hermes-cli"], "must not absorb the cli list")
    }

    // MARK: - Platform toolsets

    @Test func platformToolsetsExtractsRequestedPlatformList() {
        let yaml = """
        platform_toolsets:
          cli:
            - browser
            - kanban
            - web
          discord:
            - messaging
        """
        let cli = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: yaml, platform: "cli"
        )
        #expect(cli == ["browser", "kanban", "web"])
        let discord = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: yaml, platform: "discord"
        )
        #expect(discord == ["messaging"])
    }

    @Test func platformToolsetsAbsentBlockReturnsEmpty() {
        let yaml = """
        version: 1
        agent:
          max_turns: 90
        """
        let items = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: yaml, platform: "cli"
        )
        #expect(items.isEmpty)
    }

    @Test func platformToolsetsAbsentPlatformInBlockReturnsEmpty() {
        // Block exists but the requested platform isn't there.
        let yaml = """
        platform_toolsets:
          discord:
            - messaging
        """
        let items = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: yaml, platform: "cli"
        )
        #expect(items.isEmpty)
    }

    @Test func platformToolsetsRealUserConfig() {
        // Mirrors the user's actual `~/.hermes/config.yaml` shape that
        // produced the empty-board bug — `cli` has a long list but no
        // `kanban`, so the detector must classify this as disabled.
        let yaml = """
        toolsets:
        - hermes-cli
        platform_toolsets:
          cli:
          - browser
          - clarify
          - code_execution
          - cronjob
          - delegation
          - file
          - homeassistant
          - image_gen
          - memory
          - messaging
          - session_search
          - skills
          - terminal
          - todo
          - tts
          - vision
          - web
        """
        let items = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: yaml, platform: "cli"
        )
        #expect(!items.contains("kanban"))
        #expect(items.count == 17)
        #expect(items.first == "browser")
        #expect(items.last == "web")
    }

    @Test func platformToolsetsHandlesIndentedListMarkerVariant() {
        // Some users write the list-item marker flush with the parent
        // key rather than indented two spaces. Both shapes should
        // parse identically.
        let yaml = """
        platform_toolsets:
          cli:
          - browser
          - kanban
        """
        let items = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: yaml, platform: "cli"
        )
        #expect(items == ["browser", "kanban"])
    }
}
