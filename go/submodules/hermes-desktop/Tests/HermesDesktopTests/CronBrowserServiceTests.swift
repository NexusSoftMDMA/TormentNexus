import Foundation
import Testing
@testable import HermesDesktop

struct CronBrowserServiceTests {
    @Test
    func listJobsRunsPythonThroughHermesServiceEnvironment() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let runner = CronRecordingSSHProcessRunner(
            result: SSHCommandResult(stdout: #"{"ok":true,"jobs":[]}"#, stderr: "", exitCode: 0)
        )
        let transport = SSHTransport(
            paths: makeTestAppPaths(root: root),
            processRunner: runner
        )
        let service = CronBrowserService(sshTransport: transport)
        let connection = ConnectionProfile(
            label: "Prod",
            sshHost: "example.com",
            sshUser: "alice",
            hermesProfile: "research"
        ).updated()

        let jobs = try await service.listJobs(connection: connection)

        let invocation = try #require(await runner.lastInvocation)
        #expect(jobs.isEmpty)
        #expect(invocation.arguments.contains(connection.remoteServiceCommand("python3 -")))
    }

    @Test
    func listJobsDecodesMicrosecondCronTimestamps() async throws {
        let root = try makeTemporaryDirectory()
        defer { try? FileManager.default.removeItem(at: root) }

        let stdout = """
        {
          "ok": true,
          "jobs": [
            {
              "id": "job-1",
              "name": "ICT Discord Market Alert Monitor",
              "prompt": "Run the ICT Discord market alert script",
              "script": null,
              "workdir": null,
              "no_agent": false,
              "skills": [],
              "model": null,
              "provider": null,
              "base_url": null,
              "schedule": {
                "kind": "interval",
                "expr": null,
                "timezone": null
              },
              "schedule_display": "every 240m",
              "recurrence": null,
              "enabled": false,
              "state": "paused",
              "created_at": "2026-04-16T15:22:08.468950-04:00",
              "next_run_at": "2026-04-26T05:16:45.112532-04:00",
              "last_run_at": "2026-04-26T01:16:45.112532-04:00",
              "last_status": "ok",
              "last_error": null,
              "delivery_target": "origin",
              "origin": null,
              "last_delivery_error": null
            }
          ]
        }
        """

        let service = CronBrowserService(
            sshTransport: SSHTransport(
                paths: makeTestAppPaths(root: root),
                processRunner: CronRecordingSSHProcessRunner(
                    result: SSHCommandResult(stdout: stdout, stderr: "", exitCode: 0)
                )
            )
        )
        let connection = ConnectionProfile(label: "Prod", sshHost: "example.com").updated()

        let jobs = try await service.listJobs(connection: connection)

        let job = try #require(jobs.first)
        #expect(job.createdAt != nil)
        #expect(job.nextRunAt != nil)
        #expect(job.lastRunAt != nil)
        #expect(job.state == .paused)
    }
}

private struct CronSSHProcessInvocation {
    let arguments: [String]
}

private actor CronRecordingSSHProcessRunner: SSHProcessRunning {
    private let result: SSHCommandResult
    private(set) var lastInvocation: CronSSHProcessInvocation?

    init(result: SSHCommandResult) {
        self.result = result
    }

    func run(
        executableURL: URL,
        arguments: [String],
        standardInput: Data?
    ) async throws -> SSHCommandResult {
        lastInvocation = CronSSHProcessInvocation(arguments: arguments)
        return result
    }
}
