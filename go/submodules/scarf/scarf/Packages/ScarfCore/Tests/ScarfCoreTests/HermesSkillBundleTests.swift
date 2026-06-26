import Testing
@testable import ScarfCore

/// Coverage for `HermesSkillBundle.parse(yaml:stem:)` — the tolerant
/// YAML reader for Hermes v0.15 skill-bundle files stored at
/// `~/.hermes/skill-bundles/*.yaml`. Mirrors `SkillFrontmatterParserTests`
/// in shape: parse a literal YAML string and assert the decoded model.
@Suite("HermesSkillBundle")
struct HermesSkillBundleTests {

    @Test func parsesFullBundle() {
        let yaml = """
        name: backend-dev
        description: Backend development helpers
        skills:
          - github-code-review
          - test-driven-development
          - sqlite
        instruction: |
          Prefer migrations over manual schema edits.
          Always write a test first.
        """
        let bundle = HermesSkillBundle.parse(yaml: yaml, stem: "backend-dev")
        #expect(bundle != nil)
        #expect(bundle?.name == "backend-dev")
        #expect(bundle?.id == "backend-dev")
        #expect(bundle?.slug == "backend-dev")
        #expect(bundle?.description == "Backend development helpers")
        #expect(bundle?.skills == ["github-code-review", "test-driven-development", "sqlite"])
        #expect(bundle?.instruction == "Prefer migrations over manual schema edits.\nAlways write a test first.")
    }

    @Test func tolerantMinimalBundle() {
        // Just name + skills — no description, no instruction. The
        // optional fields must decode to nil, not crash or empty-string.
        let yaml = """
        name: quick
        skills:
          - timer
        """
        let bundle = HermesSkillBundle.parse(yaml: yaml, stem: "quick")
        #expect(bundle != nil)
        #expect(bundle?.name == "quick")
        #expect(bundle?.skills == ["timer"])
        #expect(bundle?.description == nil)
        #expect(bundle?.instruction == nil)
    }

    @Test func fallsBackToStemWhenNameMissing() {
        // No `name:` key — the file stem becomes the name (and slug).
        let yaml = """
        skills:
          - alpha
          - beta
        """
        let bundle = HermesSkillBundle.parse(yaml: yaml, stem: "ops-tools")
        #expect(bundle?.name == "ops-tools")
        #expect(bundle?.slug == "ops-tools")
        #expect(bundle?.skills == ["alpha", "beta"])
    }

    @Test func parsesInlineFlowSkillsList() {
        // Hermes-emitted YAML can use the inline `[a, b]` flow form.
        let yaml = """
        name: web
        skills: [search, fetch, scrape]
        """
        let bundle = HermesSkillBundle.parse(yaml: yaml, stem: "web")
        #expect(bundle?.skills == ["search", "fetch", "scrape"])
    }

    @Test func slugCollapsesSpacesAndUnderscores() {
        let yaml = """
        name: Backend Dev_Pro
        skills:
          - one
        """
        let bundle = HermesSkillBundle.parse(yaml: yaml, stem: "x")
        #expect(bundle?.slug == "backend-dev-pro")
    }

    @Test func skipsDegenerateContent() {
        // No name resolvable AND no skills → nil so the loader can
        // compactMap it away rather than surfacing an empty card.
        let bundle = HermesSkillBundle.parse(yaml: "# just a comment\n", stem: "")
        #expect(bundle == nil)
    }
}
