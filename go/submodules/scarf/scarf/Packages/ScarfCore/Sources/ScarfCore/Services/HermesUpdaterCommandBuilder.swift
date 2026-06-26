import Foundation

/// Pure helpers that build argv arrays for `hermes update` invocations.
///
/// Lives in ScarfCore so the eventual UI surface (Mac / iOS / remote)
/// shares flag selection. There is no in-app "Update Hermes" affordance
/// in v2.7.5 — Sparkle handles Scarf-self-update and `hermes update` is
/// invoked by users in their terminal — but capability-gated flag logic
/// is forward-compat plumbing that the future affordance will call. Each
/// helper is a `nonisolated static` pure function: no transport, no
/// MainActor, no mocking surface required.
public enum HermesUpdaterCommandBuilder {
    /// Argv for an `hermes update` invocation, capability-gated.
    ///
    /// Pre-v0.12 hosts only had `update` (no flags). v0.12+ accepts
    /// `--check` for preflight. v0.13+ accepts `--yes` / `-y` for
    /// unattended runs (skips the interactive confirmation prompt).
    /// Flags are silently dropped when the connected host can't honor
    /// them so callers don't need to branch on capabilities themselves.
    public static func updateArgv(
        capabilities: HermesCapabilities,
        unattended: Bool,
        checkOnly: Bool
    ) -> [String] {
        var args: [String] = ["update"]
        if checkOnly && capabilities.hasUpdateCheck {
            args.append("--check")
        }
        if unattended && capabilities.hasUpdateNonInteractive {
            args.append("--yes")
        }
        return args
    }
}
