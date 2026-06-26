import Foundation

/// Errors thrown by `CuratorService`. Each case carries enough detail
/// to render a user-actionable message — the view model surfaces these
/// inline as a banner above the leaderboard rather than blocking with a
/// modal alert.
public enum CuratorError: Error, LocalizedError, Sendable {
    /// `hermes` binary couldn't be located.
    case cliMissing
    /// Subprocess returned non-zero exit. `stderr` may carry a synthetic
    /// message when the transport itself failed.
    case nonZeroExit(verb: String, code: Int32, stderr: String)
    /// JSON decoding failed. Underlying message wrapped for diagnostics.
    case decoding(verb: String, message: String)
    /// Generic transport error — process couldn't start, IO failed, etc.
    case transport(message: String)

    public var errorDescription: String? {
        switch self {
        case .cliMissing:
            return "Hermes CLI couldn't be found. Install Hermes v0.13+ and ensure it's on your PATH."
        case .nonZeroExit(let verb, let code, let stderr):
            let trimmed = stderr.trimmingCharacters(in: .whitespacesAndNewlines)
            if trimmed.isEmpty {
                return "`hermes curator \(verb)` exited with code \(code)."
            }
            return trimmed
        case .decoding(let verb, let message):
            return "Couldn't decode `hermes curator \(verb)` output: \(message)"
        case .transport(let message):
            return message
        }
    }
}
