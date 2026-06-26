import Foundation
import Testing

@testable import HermesDesktop

@MainActor
struct WorkflowPersistenceTests {
    @Test
    func workflowsPersistAndStayScopedByWorkspaceFingerprint() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let paths = makeTestAppPaths(root: root)
        let store = ConnectionStore(paths: paths)

        let hostWorkflow = WorkflowPreset(
            workspaceScopeFingerprint: "host-a|user|22|~/.hermes",
            name: "Host A",
            prompt: "Inspect host A",
            assignedSkills: [
                WorkflowSkillReference(relativePath: "inspect", slug: "inspect", name: "Inspect")
            ]
        )
        let profileWorkflow = WorkflowPreset(
            workspaceScopeFingerprint: "host-a|user|22|~/.hermes/profiles/research",
            name: "Research",
            prompt: "Inspect research profile",
            assignedSkills: [
                WorkflowSkillReference(relativePath: "research", slug: "research", name: "Research")
            ]
        )

        store.upsertWorkflow(hostWorkflow)
        store.upsertWorkflow(profileWorkflow)

        let reloadedStore = ConnectionStore(paths: paths)

        #expect(reloadedStore.workflows(for: hostWorkflow.workspaceScopeFingerprint).map(\.name) == ["Host A"])
        #expect(reloadedStore.workflows(for: profileWorkflow.workspaceScopeFingerprint).map(\.name) == ["Research"])
    }

    @Test
    func upsertWorkflowReplacesExistingEntryWithoutDuplicatingIt() throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let store = ConnectionStore(paths: makeTestAppPaths(root: root))
        let original = WorkflowPreset(
            id: UUID(uuidString: "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE")!,
            workspaceScopeFingerprint: "host|user|22|~/.hermes",
            name: "Release Audit",
            prompt: "Inspect release",
            assignedSkills: [
                WorkflowSkillReference(relativePath: "inspect", slug: "inspect", name: "Inspect")
            ]
        )

        store.upsertWorkflow(original)
        store.upsertWorkflow(
            original.updated(
                name: "Release Audit Updated",
                prompt: "Inspect release and publish notes",
                assignedSkills: original.assignedSkills
            )
        )

        let items = store.workflows(for: original.workspaceScopeFingerprint)
        #expect(items.count == 1)
        #expect(items.first?.name == "Release Audit Updated")
        #expect(items.first?.prompt == "Inspect release and publish notes")
    }
}
