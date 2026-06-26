import Foundation
import Testing

@testable import HermesDesktop

struct WorkflowModelsTests {
    @Test
    func draftValidationRequiresOnlyNameAndPrompt() {
        var draft = WorkflowDraft()

        #expect(draft.validationError == "Workflow name is required.")

        draft.name = "Release Audit"
        #expect(draft.validationError == "Workflow prompt is required.")

        draft.prompt = "Audit the release branch"
        #expect(draft.validationError == nil)
    }

    @Test
    func presetNormalizesAndDeduplicatesSkillsBySlug() {
        let workflow = WorkflowPreset(
            workspaceScopeFingerprint: "host|user|22|~/.hermes",
            name: "  Release Audit  ",
            prompt: "  Check the release status  ",
            assignedSkills: [
                WorkflowSkillReference(relativePath: "zeta/path", slug: "zeta", name: "Zeta"),
                WorkflowSkillReference(relativePath: "beta/path", slug: "beta", name: "Beta"),
                WorkflowSkillReference(relativePath: "alpha/path", slug: "alpha", name: "Alpha"),
                WorkflowSkillReference(relativePath: "alpha/path", slug: "alpha", name: "Alpha Override")
            ]
        )

        #expect(workflow.name == "Release Audit")
        #expect(workflow.prompt == "Check the release status")
        #expect(workflow.assignedSkills.map(\.slug) == ["alpha", "beta", "zeta"])
        #expect(workflow.assignedSkills.first?.resolvedName == "Alpha Override")
    }

    @Test
    func launchInvocationUsesRepeatedSkillsAndSeedsInteractiveChat() {
        let workflow = WorkflowPreset(
            workspaceScopeFingerprint: "host|user|22|~/.hermes/profiles/research",
            name: "Investigate",
            prompt: "Inspect \"quoted\" output && keep notes",
            assignedSkills: [
                WorkflowSkillReference(relativePath: "ssh/tools", slug: "tools", name: "SSH Tools"),
                WorkflowSkillReference(relativePath: "deploy-check", slug: "deploy-check", name: "Deploy Check")
            ]
        )
        let connection = ConnectionProfile(
            label: "Research",
            sshAlias: "hermes-home",
            hermesProfile: "research"
        ).updated()

        let invocation = WorkflowLaunchInvocation(workflow: workflow, connection: connection)

        #expect(invocation.arguments == [
            "--profile",
            "research",
            "--skills",
            "deploy-check",
            "--skills",
            "ssh/tools",
            "chat"
        ])
        #expect(invocation.initialInput == "Inspect \"quoted\" output && keep notes")
        #expect(
            invocation.commandLine ==
                "hermes --profile research --skills deploy-check --skills ssh/tools chat"
        )
    }

    @Test
    func launchInvocationFlattensMultilinePromptIntoSingleSubmission() {
        let workflow = WorkflowPreset(
            workspaceScopeFingerprint: "host|user|22|~/.hermes/profiles/research",
            name: "Repo triage",
            prompt: """
            Check this GitHub repository: https://github.com/dodo-reach/hermes-desktop


            inspect and summarize the existing PRs and Issues.
            """,
            assignedSkills: [
                WorkflowSkillReference(relativePath: "github/codebase-inspection", slug: "codebase-inspection", name: "Codebase Inspection")
            ]
        )
        let connection = ConnectionProfile(
            label: "Research",
            sshAlias: "hermes-home",
            hermesProfile: "research"
        ).updated()

        let invocation = WorkflowLaunchInvocation(workflow: workflow, connection: connection)

        #expect(
            invocation.initialInput ==
                "Check this GitHub repository: https://github.com/dodo-reach/hermes-desktop inspect and summarize the existing PRs and Issues."
        )
    }

    @Test
    func launchInvocationPreservesVeryLongPromptWithoutTruncation() {
        let longSections = (0..<2_000).map { index in
            "section-\(index) keeps flowing"
        }
        let prompt = longSections.joined(separator: "\n")
        let workflow = WorkflowPreset(
            workspaceScopeFingerprint: "host|user|22|~/.hermes/profiles/research",
            name: "Long prompt",
            prompt: prompt,
            assignedSkills: [
                WorkflowSkillReference(relativePath: "github/codebase-inspection", slug: "codebase-inspection", name: "Codebase Inspection")
            ]
        )
        let connection = ConnectionProfile(
            label: "Research",
            sshAlias: "hermes-home",
            hermesProfile: "research"
        ).updated()

        let invocation = WorkflowLaunchInvocation(workflow: workflow, connection: connection)
        let expected = longSections.joined(separator: " ")

        #expect(invocation.initialInput == expected)
        #expect(invocation.initialInput.count == expected.count)
        #expect(invocation.initialInput.hasSuffix("section-1999 keeps flowing"))
    }

    @Test
    func launchInvocationSkipsProfileArgumentForCustomHermesHome() {
        let workflow = WorkflowPreset(
            workspaceScopeFingerprint: "host|user|22|~/.hermes-work",
            name: "Investigate",
            prompt: "Inspect this setup",
            assignedSkills: [
                WorkflowSkillReference(relativePath: "ssh/tools", slug: "tools", name: "SSH Tools")
            ]
        )
        let connection = ConnectionProfile(
            label: "Research",
            sshAlias: "hermes-home",
            hermesProfile: "research",
            customHermesHomePath: "~/.hermes-work"
        ).updated()

        let invocation = WorkflowLaunchInvocation(workflow: workflow, connection: connection)

        #expect(invocation.arguments == [
            "--skills",
            "ssh/tools",
            "chat"
        ])
        #expect(invocation.commandLine == "hermes --skills ssh/tools chat")
        #expect(invocation.startupCommandLine.contains(#"$HERMES_HOME/hermes-agent/venv/bin/hermes"#))
        #expect(invocation.startupCommandLine.contains(#""$HERMES_BIN" --skills ssh/tools chat"#))
    }

    @Test
    func chatLaunchInvocationStartsTUIAndSeedsSkillCommands() {
        let workflow = WorkflowPreset(
            workspaceScopeFingerprint: "host|user|22|~/.hermes/profiles/research",
            name: "Investigate",
            prompt: "Inspect this setup",
            assignedSkills: [
                WorkflowSkillReference(relativePath: "ssh/tools", slug: "tools", name: "SSH Tools"),
                WorkflowSkillReference(relativePath: "deploy-check", slug: "deploy-check", name: "Deploy Check")
            ]
        )
        let connection = ConnectionProfile(
            label: "Research",
            sshAlias: "hermes-home",
            hermesProfile: "research"
        ).updated()

        let invocation = WorkflowChatLaunchInvocation(workflow: workflow, connection: connection)

        #expect(invocation.tuiInvocation.arguments == ["--profile", "research", "--tui"])
        #expect(invocation.initialInput == "/deploy-check\n/tools\nInspect this setup")
    }
}
