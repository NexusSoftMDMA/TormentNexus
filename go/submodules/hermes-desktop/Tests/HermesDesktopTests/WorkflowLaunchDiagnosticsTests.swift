import Foundation
import Testing

@testable import HermesDesktop

struct WorkflowLaunchDiagnosticsTests {
    @Test
    func diagnosticsWriteReadableLaunchEventsToLatestLogFile() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let paths = makeTestAppPaths(root: root)
        let diagnostics = WorkflowLaunchDiagnostics(
            logFileURL: paths.applicationSupportURL
                .appendingPathComponent("Diagnostics", isDirectory: true)
                .appendingPathComponent("workflow-launch-latest.log")
        )
        let workflow = WorkflowPreset(
            workspaceScopeFingerprint: "host|user|22|~/.hermes/profiles/research",
            name: "Long prompt",
            prompt: "Line 1\nLine 2\nLine 3",
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
        let context = WorkflowLaunchDiagnosticsContext(
            workflow: workflow,
            invocation: invocation,
            connection: connection
        )

        await diagnostics.recordWorkflowRunRequested(context)
        await diagnostics.recordInitialInputWaitStarted(context, deadlineMilliseconds: 8_000)
        await diagnostics.recordInitialInputSent(
            context,
            deliveryMode: .standardSubmit,
            reason: "deadline_reached_before_bracketed_paste_mode",
            bracketedPasteModeAtSend: false
        )

        let logContents = try String(contentsOf: diagnostics.logFileURL, encoding: .utf8)
        #expect(logContents.contains("event=workflow_run_requested"))
        #expect(logContents.contains("event=initial_input_wait_started"))
        #expect(logContents.contains("event=initial_input_sent"))
        #expect(logContents.contains("delivery_mode=\"standard_submit\""))
        #expect(logContents.contains("workflow_name=\"Long prompt\""))
    }
}
