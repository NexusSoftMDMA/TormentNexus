import Testing
@testable import HermesDesktop

struct SkillModelSourceTests {
    @Test
    func localSkillBuildsWritableSourceAndPath() {
        let summary = SkillSummary(
            id: "devops/deploy-k8s",
            locator: SkillLocator(sourceID: "local", relativePath: "devops/deploy-k8s"),
            source: SkillSource(
                id: "local",
                kind: .local,
                rootPath: "~/.hermes/skills",
                isReadOnly: false
            ),
            slug: "deploy-k8s",
            category: "devops",
            relativePath: "devops/deploy-k8s",
            name: "Deploy Kubernetes",
            description: "Ship a manifest safely.",
            version: "1.0.0",
            platforms: ["macos", "linux"],
            tags: ["k8s"],
            relatedSkills: [],
            hasReferences: true,
            hasScripts: false,
            hasTemplates: true
        )

        #expect(summary.id == "devops/deploy-k8s")
        #expect(summary.locator.sourceID == "local")
        #expect(summary.source.kind == .local)
        #expect(summary.source.isReadOnly == false)
        #expect(summary.skillFilePath == "~/.hermes/skills/devops/deploy-k8s/SKILL.md")
        #expect(summary.matchesSearch("linux"))
    }

    @Test
    func externalSkillBuildsReadOnlyDetail() {
        let detail = SkillDetail(
            id: "team-conventions",
            locator: SkillLocator(sourceID: "external:1", relativePath: "team-conventions"),
            source: SkillSource(
                id: "external:1",
                kind: .external,
                rootPath: "~/.agents/skills",
                isReadOnly: true
            ),
            slug: "team-conventions",
            category: nil,
            relativePath: "team-conventions",
            name: "Team Conventions",
            description: "Shared standards.",
            version: nil,
            platforms: [],
            tags: [],
            relatedSkills: [],
            hasReferences: false,
            hasScripts: false,
            hasTemplates: false,
            markdownContent: "# Team Conventions\n",
            contentHash: "abc123"
        )

        #expect(detail.source.kind == .external)
        #expect(detail.isReadOnly)
        #expect(detail.sourceLabel == "External")
        #expect(detail.skillFilePath == "~/.agents/skills/team-conventions/SKILL.md")
    }
}
