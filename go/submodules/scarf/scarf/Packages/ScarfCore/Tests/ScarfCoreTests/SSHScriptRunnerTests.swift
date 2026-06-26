import Testing
import Foundation
@testable import ScarfCore

/// Regression tests for `SSHScriptRunner`. Mac-only because the
/// implementation relies on `Foundation.Process`, which doesn't exist
/// on Swift Linux. Drives the `runLocally` path so we don't need an
/// SSH endpoint in CI.
#if os(macOS)
@Suite struct SSHScriptRunnerTests {

    /// Issue #77 regression. Pre-fix the runner read stdout via
    /// `readToEnd()` *after* the subprocess exited; once the script's
    /// output crossed the kernel's pipe buffer (16–64 KB on macOS) the
    /// process wedged because nothing was draining the read end. The
    /// only visible symptom was a 30-second timeout and an empty
    /// result.
    ///
    /// This script writes ~256 KB of bytes — comfortably past every
    /// pipe-buffer threshold. With the readabilityHandler drain in
    /// place the run should complete in well under a second and
    /// return the full payload.
    @Test func drainsLargeStdoutWithoutTimeout() async throws {
        // 256 lines × 1024 bytes/line = 256 KB.
        let script = """
        for i in $(seq 1 256); do
            printf '%04d:' "$i"
            printf '%.0sx' $(seq 1 1018)
            printf '\\n'
        done
        """
        let outcome = await SSHScriptRunner.run(
            script: script,
            context: .local,
            timeout: 10
        )
        switch outcome {
        case .completed(let stdout, _, let exitCode):
            #expect(exitCode == 0)
            // 256 lines + final newline.
            let lines = stdout.split(separator: "\n", omittingEmptySubsequences: false)
            #expect(lines.count >= 256)
            #expect(stdout.utf8.count >= 256 * 1024)
        case .connectFailure(let reason):
            Issue.record("Expected completion, got connectFailure: \(reason)")
        }
    }

    /// Sanity check that small scripts still come back the way they
    /// did before the drain refactor. Guards against an off-by-one in
    /// the readability handler that swallowed trailing bytes.
    @Test func smallScriptPayloadRoundTrips() async throws {
        let outcome = await SSHScriptRunner.run(
            script: "printf 'hello\\n' && printf 'world\\n' >&2 && exit 0",
            context: .local,
            timeout: 5
        )
        switch outcome {
        case .completed(let stdout, let stderr, let exitCode):
            #expect(exitCode == 0)
            #expect(stdout == "hello\n")
            #expect(stderr == "world\n")
        case .connectFailure(let reason):
            Issue.record("Expected completion, got connectFailure: \(reason)")
        }
    }

    /// Non-zero exit codes should still be reported as `.completed`
    /// with the captured stdout/stderr — unchanged contract.
    @Test func nonZeroExitIsReportedAsCompleted() async throws {
        let outcome = await SSHScriptRunner.run(
            script: "echo nope >&2 && exit 7",
            context: .local,
            timeout: 5
        )
        switch outcome {
        case .completed(_, let stderr, let exitCode):
            #expect(exitCode == 7)
            #expect(stderr.contains("nope"))
        case .connectFailure(let reason):
            Issue.record("Expected completion, got connectFailure: \(reason)")
        }
    }
}
#endif
