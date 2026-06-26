import Foundation

/// Errors thrown by `KanbanService`. Each case carries enough detail
/// to render a user-actionable message — VMs surface these inline in
/// the board's error banner rather than blocking with alerts, since
/// kanban interactions are high-frequency.
public enum KanbanError: Error, LocalizedError, Sendable {
    /// `hermes` binary couldn't be located (local) or the remote
    /// `hermesBinaryHint` is unset (SSH).
    case cliMissing
    /// Subprocess returned non-zero exit. `stderr` may be empty if the
    /// transport itself failed; carries a synthetic message in that case.
    case nonZeroExit(code: Int32, stderr: String)
    /// JSON decoding failed. Underlying `Error` is wrapped for
    /// diagnostics; the user-facing message is generic.
    case decoding(message: String)
    /// `hermes kanban list --json` printed the literal string
    /// "no matching tasks" instead of `[]`. Treated as a successful
    /// empty result by callers but exposed here so VMs can distinguish
    /// it from "transport error" if they want to.
    case noMatchingTasks
    /// Verb is not supported by this Hermes version (gated upstream
    /// by `HermesCapabilities.hasKanban` + reasoned-about feature
    /// drift). Carries the verb name + a hint.
    case notSupported(verb: String, reason: String)
    /// Disallowed transition the UI tried to perform (e.g. dragging a
    /// `done` card back to `todo`). Caller surfaces a tooltip; this is
    /// thrown only when a programmatic transition is requested instead
    /// of being filtered out at the drag-target gate.
    case forbiddenTransition(from: String, to: String, reason: String)

    public var errorDescription: String? {
        switch self {
        case .cliMissing:
            return "Hermes CLI couldn't be found. Install Hermes v0.12+ and ensure it's on your PATH."
        case .nonZeroExit(let code, let stderr):
            let trimmed = stderr.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty {
                return "Hermes exited with code \(code)."
            }
            return trimmed
        case .decoding(let message):
            return "Couldn't decode Hermes output: \(message)"
        case .noMatchingTasks:
            return "No matching tasks."
        case .notSupported(let verb, let reason):
            return "`hermes kanban \(verb)` isn't available: \(reason)"
        case .forbiddenTransition(let from, let to, let reason):
            return "Can't move a \(from) task to \(to): \(reason)"
        }
    }
}
