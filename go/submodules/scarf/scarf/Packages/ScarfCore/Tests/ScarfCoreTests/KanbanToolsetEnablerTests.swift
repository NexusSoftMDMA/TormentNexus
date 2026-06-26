import Testing
import Foundation
@testable import ScarfCore

/// Pure plan-logic tests for `KanbanToolsetEnabler`. The actor's I/O
/// path (`enable()` calling read → plan → write → verify) is exercised
/// by manual verification; these tests freeze the YAML mutation logic
/// itself, which is the bit that caused the production bug (silent
/// failure of `hermes tools enable kanban`).
@Suite struct KanbanToolsetEnablerTests {

    // MARK: - planEnable — happy paths

    @Test func enablePlanNoOpWhenKanbanAlreadyInPlatformList() {
        let yaml = """
        platform_toolsets:
          cli:
          - browser
          - kanban
          - web
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        #expect(plan == .alreadyPresent)
    }

    @Test func enablePlanNoOpWhenKanbanAlreadyInTopLevelToolsets() {
        // Top-level `toolsets:` is also a valid path Hermes honors —
        // the gating in `tools/kanban_tools.py` checks both. If the
        // user has it there, the per-platform write would be redundant.
        let yaml = """
        toolsets:
          - hermes-cli
          - kanban
        platform_toolsets:
          cli:
          - browser
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        #expect(plan == .alreadyPresent)
    }

    @Test func enablePlanInsertsKanbanAlphabeticallyIntoSortedCliList() {
        // Mirrors the real user-config shape that produced the bug
        // report. Insertion must land between `image_gen` and `memory`.
        let yaml = """
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
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        guard case let .rewrite(newYaml) = plan else {
            Issue.record("Expected .rewrite, got \(plan)")
            return
        }
        let parsed = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: newYaml, platform: "cli"
        )
        #expect(parsed.contains("kanban"))
        // Alphabetical order preserved.
        if let imageIdx = parsed.firstIndex(of: "image_gen"),
           let kanbanIdx = parsed.firstIndex(of: "kanban"),
           let memoryIdx = parsed.firstIndex(of: "memory") {
            #expect(imageIdx < kanbanIdx)
            #expect(kanbanIdx < memoryIdx)
        } else {
            Issue.record("Expected image_gen / kanban / memory all present")
        }
        // Existing items unchanged.
        #expect(parsed.first == "browser")
        #expect(parsed.last == "web")
        #expect(parsed.count == 18)
    }

    @Test func enablePlanAppendsToEndWhenListIsUnsortedAndKanbanBelongsAtEnd() {
        // If existing items are unsorted, the alphabetical heuristic
        // falls through to the end. Worst-case behaviour — but still
        // produces a valid list, and the detector doesn't care about
        // order.
        let yaml = """
        platform_toolsets:
          cli:
          - file
          - browser
          - terminal
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        guard case let .rewrite(newYaml) = plan else {
            Issue.record("Expected .rewrite, got \(plan)")
            return
        }
        let parsed = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: newYaml, platform: "cli"
        )
        #expect(parsed == ["file", "browser", "kanban", "terminal"])
    }

    @Test func enablePlanPreservesAllNonListContentVerbatim() {
        // Sanity check: lines outside the cli list (top-level keys,
        // sibling platform keys, comments) must come through untouched.
        let yaml = """
        # Top comment
        model:
          provider: anthropic
          default: claude-haiku-4-5

        platform_toolsets:
          cli:
          - browser
          - web
          discord:
          - messaging
        # Trailing comment
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        guard case let .rewrite(newYaml) = plan else {
            Issue.record("Expected .rewrite, got \(plan)")
            return
        }
        // Comments + other blocks preserved.
        #expect(newYaml.contains("# Top comment"))
        #expect(newYaml.contains("# Trailing comment"))
        #expect(newYaml.contains("model:"))
        #expect(newYaml.contains("  provider: anthropic"))
        #expect(newYaml.contains("  discord:"))
        #expect(newYaml.contains("  - messaging"))
        // Kanban added to cli only, not to discord.
        let cliItems = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: newYaml, platform: "cli"
        )
        let discordItems = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: newYaml, platform: "discord"
        )
        #expect(cliItems.contains("kanban"))
        #expect(!discordItems.contains("kanban"))
    }

    // MARK: - planEnable — refuse paths

    @Test func enablePlanRefusesOnScalarValueShape() {
        // This is the post-`hermes config set platform_toolsets.cli
        // kanban` corruption shape. Hermes destroys the list and writes
        // a bare scalar value. Refusing is correct — we'd have to guess
        // the original list to "fix" it.
        let yaml = """
        platform_toolsets:
          cli: kanban
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        guard case let .refuse(reason) = plan else {
            Issue.record("Expected .refuse for scalar shape, got \(plan)")
            return
        }
        #expect(reason.contains("scalar value"))
        #expect(reason.contains("kanban"))
    }

    @Test func enablePlanRefusesWhenPlatformBlockMissing() {
        let yaml = """
        model:
          provider: anthropic
        toolsets:
          - hermes-cli
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        guard case let .refuse(reason) = plan else {
            Issue.record("Expected .refuse, got \(plan)")
            return
        }
        #expect(reason.contains("platform_toolsets"))
    }

    @Test func enablePlanRefusesWhenRequestedPlatformKeyMissing() {
        // Block exists but only has `discord:`, not `cli:`. We don't
        // try to manufacture a new platform — better to be explicit.
        let yaml = """
        platform_toolsets:
          discord:
          - messaging
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        guard case let .refuse(reason) = plan else {
            Issue.record("Expected .refuse, got \(plan)")
            return
        }
        #expect(reason.contains("cli"))
    }

    // MARK: - planDisable

    @Test func disablePlanRemovesKanbanLineWhilePreservingNeighbours() {
        let yaml = """
        platform_toolsets:
          cli:
          - browser
          - kanban
          - web
        """
        let plan = KanbanToolsetEnabler.planDisable(yaml: yaml, platform: "cli")
        guard case let .rewrite(newYaml) = plan else {
            Issue.record("Expected .rewrite, got \(plan)")
            return
        }
        let parsed = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: newYaml, platform: "cli"
        )
        #expect(parsed == ["browser", "web"])
    }

    @Test func disablePlanNoOpWhenKanbanAbsent() {
        // Disable on a list that doesn't have kanban is a success
        // no-op — matches "wasn't there in the first place" semantics.
        let yaml = """
        platform_toolsets:
          cli:
          - browser
          - web
        """
        let plan = KanbanToolsetEnabler.planDisable(yaml: yaml, platform: "cli")
        #expect(plan == .alreadyPresent)
    }

    @Test func disablePlanNoOpWhenPlatformBlockMissing() {
        // Nothing to remove from a config with no platform_toolsets
        // section.
        let yaml = """
        model:
          provider: anthropic
        """
        let plan = KanbanToolsetEnabler.planDisable(yaml: yaml, platform: "cli")
        #expect(plan == .alreadyPresent)
    }

    // MARK: - Round-trip with the detector

    @Test func enableMutationProducesYamlDetectorSeesAsEnabled() {
        // The full contract that broke in production: a successful
        // enable mutation must result in YAML that the detector
        // classifies as `.enabled(via: .platform)`. The actor's
        // post-write verification step depends on this.
        let yaml = """
        platform_toolsets:
          cli:
          - browser
          - web
        """
        let plan = KanbanToolsetEnabler.planEnable(yaml: yaml, platform: "cli")
        guard case let .rewrite(newYaml) = plan else {
            Issue.record("Expected .rewrite, got \(plan)")
            return
        }
        let items = KanbanToolsetDetector.parsePlatformToolsets(
            yaml: newYaml, platform: "cli"
        )
        #expect(items.contains("kanban"),
                Comment(rawValue: "post-rewrite YAML must contain kanban so the detector's round-trip check passes — otherwise the actor surfaces a confusing 'wrote but still disabled' error"))
    }
}
