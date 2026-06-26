import Foundation
import ScarfCore
import os

/// Manages a Scarf-owned cron job that keeps OAuth refresh tokens
/// alive by booting a trivial Hermes session on a daily cadence.
///
/// **Why this exists.** Hermes refreshes OAuth access tokens on
/// agent startup (via `resolve_nous_runtime_credentials()` and
/// equivalents), but never proactively. If the user goes longer than
/// the *refresh*-token lifetime without starting a session, the
/// refresh token itself expires and only a full re-auth recovers it.
/// Refresh-token lifetimes are typically ~30 days; a 24-hour
/// heartbeat keeps the window from closing for users who go quiet.
///
/// **What it runs.** A single cron job with a stable name
/// (`Self.jobName`) and a minimal one-token prompt. Executing the
/// job boots `hermes acp` end-to-end, which is what triggers the
/// refresh. There is no public Hermes CLI verb to refresh a token in
/// isolation today (no `hermes auth refresh <provider>`), so booting
/// a session is the only mechanism we have. When Hermes adds a
/// dedicated refresh verb, swap the prompt for a `--script` that
/// invokes it and the surrounding wiring stays unchanged.
///
/// **Identification.** The job is found by exact-match on
/// `Self.jobName`. Users can edit the schedule from the Cron tab
/// without breaking detection — only the name is load-bearing here.
@MainActor
final class OAuthKeepaliveCronService {
    /// Stable job name. The leading `[scarf:oauth-keepalive]` prefix
    /// follows the convention `ProjectTemplateInstaller` uses for
    /// template-installed cron jobs (`[tmpl:<id>] …`) so future
    /// inspection tools can distinguish Scarf-owned schedules from
    /// user-authored ones at a glance.
    nonisolated static let jobName = "[scarf:oauth-keepalive] OAuth token refresh"

    /// 4am local daily. Off-peak avoids contending with interactive
    /// usage and is a reasonable default; users can reschedule from
    /// the Cron tab if they prefer a different cadence. The cron
    /// window must stay <= the shortest refresh-token lifetime among
    /// the user's configured OAuth providers (~30d for Nous).
    ///
    /// `nonisolated` so the detached `enable()` closure can read it
    /// without an await-hop to the main actor — matches `jobName`.
    nonisolated static let defaultSchedule = "0 4 * * *"

    /// Minimal prompt. The point is to boot a session — not to do
    /// useful work — so we want the LLM call to terminate fast. A
    /// one-word prompt + a one-word reply is the cheapest end-to-end
    /// turn. Subscription-routed providers (Nous) bear zero
    /// per-call cost; for API-key users, a single trivial turn per
    /// day is negligible compared to the alternative of full re-auth
    /// every month.
    ///
    /// `nonisolated` so the detached `enable()` closure can read it
    /// without an await-hop to the main actor — matches `jobName`.
    nonisolated static let defaultPrompt = "Reply with the single word 'ok'."

    private let logger = Logger(subsystem: "com.scarf", category: "OAuthKeepaliveCronService")
    let context: ServerContext
    private let fileService: HermesFileService

    init(context: ServerContext = .local) {
        self.context = context
        self.fileService = HermesFileService(context: context)
    }

    // MARK: - Read

    /// Returns the keepalive job if one is currently registered, nil
    /// otherwise. Reads `~/.hermes/cron/jobs.json` synchronously via
    /// the existing `loadCronJobs()` path.
    nonisolated func currentJob() -> HermesCronJob? {
        fileService.loadCronJobs().first { $0.name == Self.jobName }
    }

    nonisolated func isEnabled() -> Bool {
        currentJob() != nil
    }

    // MARK: - Mutate

    /// Register the keepalive job via `hermes cron create`. No-op when
    /// a job with the same name already exists — toggle semantics
    /// stay idempotent so a double-tap doesn't duplicate the entry.
    /// Returns true on success or no-op, false on CLI failure.
    @discardableResult
    nonisolated func enable() async -> Bool {
        if isEnabled() { return true }
        // `hermes cron create` only accepts: --name, --deliver,
        // --repeat, --skill, --script, --workdir. The `silent: Bool?`
        // field on HermesCronJob is JSON-only (Hermes can write it,
        // but the CLI's create verb doesn't expose a flag for it).
        // Pass any unknown flag and argparse rejects the whole
        // command, so stick to the supported surface and let Hermes
        // pick its default delivery target — the side effect we care
        // about (token refresh during session boot) fires regardless.
        let result = await Task.detached { [fileService] in
            fileService.runHermesCLI(
                args: [
                    "cron", "create",
                    "--name", Self.jobName,
                    Self.defaultSchedule,
                    Self.defaultPrompt,
                ],
                timeout: 60
            )
        }.value
        if result.exitCode != 0 {
            logger.warning("oauth-keepalive enable failed: exit=\(result.exitCode) output=\(result.output, privacy: .public)")
            return false
        }
        return true
    }

    /// Remove the keepalive job. Idempotent — when no job exists
    /// today, the call is a no-op success. Returns true on success
    /// or no-op, false on CLI failure.
    @discardableResult
    nonisolated func disable() async -> Bool {
        guard let job = currentJob() else { return true }
        let result = await Task.detached { [fileService] in
            fileService.runHermesCLI(
                args: ["cron", "remove", job.id],
                timeout: 30
            )
        }.value
        if result.exitCode != 0 {
            logger.warning("oauth-keepalive disable failed: exit=\(result.exitCode) output=\(result.output, privacy: .public)")
            return false
        }
        return true
    }
}
