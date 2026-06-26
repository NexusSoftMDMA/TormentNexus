import Foundation

/// Optimistic local mirror of the agent's currently-locked goal (set via
/// the `/goal <text>` slash command, Hermes v0.13+). Scarf records this
/// the moment the user sends `/goal …` so the chat header pill appears
/// synchronously, without waiting for a server round-trip. There is no
/// authoritative read-back path in v2.8.0 — see WS-2 plan Q1.
///
/// Plain value type, no mutation API. Drives the goal pill in
/// `SessionInfoBar` and the inspector contextual menu.
public struct HermesActiveGoal: Sendable, Equatable, Identifiable {
    /// The user's verbatim goal text (post-trim).
    public let text: String
    /// When Scarf observed the `/goal` send. Local clock — not the
    /// server's authoritative timestamp.
    public let setAt: Date

    public var id: String {
        text + "@" + ISO8601DateFormatter().string(from: setAt)
    }

    public init(text: String, setAt: Date) {
        self.text = text
        self.setAt = setAt
    }
}
