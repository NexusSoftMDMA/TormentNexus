import Foundation
import Testing

@testable import HermesDesktop

struct SkillLaunchabilityTests {
    @Test
    func parserExtractsLaunchableIdentifiersFromEnabledSkillsTable() {
        let output = """
                                  Installed Skills (enabled only)
        ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━┳━━━━━━━━━┳━━━━━━━━━┓
        ┃ Name                        ┃ Category             ┃ Source  ┃ Trust   ┃ Status  ┃
        ┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━╇━━━━━━━━━╇━━━━━━━━━┩
        │ camofox                     │                      │ local   │ local   │ enabled │
        │ hermes-cron                 │ devops               │ local   │ local   │ enabled │
        │ subagent-driven-development │ software-development │ builtin │ builtin │ enabled │
        └─────────────────────────────┴──────────────────────┴─────────┴─────────┴─────────┘
        0 hub-installed, 17 builtin, 2 local — 19 enabled shown
        """

        let records = LaunchableSkillInventoryParser.parse(output)

        #expect(records.map(\.launchIdentifier) == [
            "camofox",
            "devops/hermes-cron",
            "software-development/subagent-driven-development"
        ])
    }

    @Test
    func filterKeepsOnlySkillsThatAreActuallyLaunchable() {
        let source = SkillSource(id: "local", kind: .local, rootPath: "~/.hermes/skills", isReadOnly: false)
        let discovered = [
            SkillSummary(
                id: "apple/apple-notes",
                locator: SkillLocator(sourceID: "local", relativePath: "apple/apple-notes"),
                source: source,
                slug: "apple-notes",
                category: "apple",
                relativePath: "apple/apple-notes",
                name: "apple-notes",
                description: nil,
                version: nil,
                platforms: ["macos"],
                tags: [],
                relatedSkills: [],
                hasReferences: false,
                hasScripts: false,
                hasTemplates: false
            ),
            SkillSummary(
                id: "devops/hermes-cron",
                locator: SkillLocator(sourceID: "local", relativePath: "devops/hermes-cron"),
                source: source,
                slug: "hermes-cron",
                category: "devops",
                relativePath: "devops/hermes-cron",
                name: "hermes-cron",
                description: nil,
                version: nil,
                platforms: [],
                tags: [],
                relatedSkills: [],
                hasReferences: false,
                hasScripts: false,
                hasTemplates: false
            )
        ]
        let records = [
            LaunchableSkillRecord(name: "hermes-cron", category: "devops", source: "local", status: "enabled")
        ]

        let filtered = LaunchableSkillInventoryParser.filterDiscoveredSkills(discovered, using: records)

        #expect(filtered.map(\.relativePath) == ["devops/hermes-cron"])
    }
}
