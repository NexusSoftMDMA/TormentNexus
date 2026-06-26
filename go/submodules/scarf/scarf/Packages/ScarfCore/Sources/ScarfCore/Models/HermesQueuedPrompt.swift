import Foundation

/// One queued prompt the user has staged via `/queue <text>` (Hermes
/// v0.13+ ACP `/queue` slash command). Hermes is the authoritative owner
/// of the actual queue server-side — Scarf maintains this mirror so the
/// chat header chip + popover can show "what's pending" without an
/// extra round-trip. The mirror drains best-effort when a turn
/// completes (`RichChatViewModel.popQueuedPrompt`).
///
/// `id` is a Scarf-side UUID minted at queue-time — Hermes' wire
/// protocol does not expose a per-queue-entry id, so we never round-trip
/// an entry-level identifier. See WS-2 plan Q5.
public struct HermesQueuedPrompt: Sendable, Equatable, Identifiable {
    public let id: UUID
    public let text: String
    public let queuedAt: Date

    public init(id: UUID = UUID(), text: String, queuedAt: Date = Date()) {
        self.id = id
        self.text = text
        self.queuedAt = queuedAt
    }
}
