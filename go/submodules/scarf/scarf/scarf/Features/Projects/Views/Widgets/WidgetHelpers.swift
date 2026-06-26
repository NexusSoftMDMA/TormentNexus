import SwiftUI
import Foundation

/// Strips CSI ANSI escape sequences (`ESC [ ... letter`) so log output
/// pasted into the dashboard renders cleanly. Single regex, fast enough
/// for the small windows the log_tail / cron_status widgets work with.
/// Lightweight result type for file-reading widgets — failure is just a
/// human-readable string the widget surfaces in its error card. `Result<_, String>`
/// won't compile because `String` doesn't conform to `Error`; this alias
/// uses a typed wrapper so the rest of the call sites stay readable.
typealias WidgetIOResult<T> = Result<T, WidgetIOError>

struct WidgetIOError: Error, Sendable {
    let message: String
    nonisolated init(_ m: String) { self.message = m }
}

extension Result where Failure == WidgetIOError {
    /// Convenience constructor — `.failure("…")` instead of
    /// `.failure(WidgetIOError("…"))`. Marked nonisolated so detached
    /// tasks can call it from outside the main actor.
    nonisolated static func failure(_ message: String) -> Self {
        .failure(WidgetIOError(message))
    }
}

enum AnsiStripper {
    /// Single-call regex strip. Compiles per call — log windows are small,
    /// the cost is negligible, and skipping a `static let` cache means
    /// callers from `Task.detached` don't fight the Swift 6 actor checker.
    nonisolated static func strip(_ s: String) -> String {
        // ESC = \u{1B}; CSI = ESC [ ; final byte is in 0x40..0x7E.
        guard let pattern = try? NSRegularExpression(
            pattern: "\u{1B}\\[[0-?]*[ -/]*[@-~]", options: []
        ) else { return s }
        let range = NSRange(s.startIndex..., in: s)
        return pattern.stringByReplacingMatches(
            in: s, options: [], range: range, withTemplate: ""
        )
    }
}

func parseColor(_ name: String?) -> Color {
    switch name?.lowercased() {
    case "red": return .red
    case "orange": return .orange
    case "yellow": return .yellow
    case "green": return .green
    case "blue": return .blue
    case "purple": return .purple
    case "pink": return .pink
    case "teal", "cyan": return .teal
    case "indigo": return .indigo
    case "mint": return .mint
    case "brown": return .brown
    case "gray", "grey": return .gray
    default: return .blue
    }
}
