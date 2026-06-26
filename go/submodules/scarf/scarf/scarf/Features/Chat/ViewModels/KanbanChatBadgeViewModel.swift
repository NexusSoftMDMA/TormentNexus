import Foundation
import Observation
import ScarfCore
import os

/// Drives the live count badge on `SessionInfoBar`'s Kanban chip.
/// Polls `KanbanService.list` every 5 seconds while mounted, counts
/// `running + blocked` tasks created by *this chat* (scoped precisely by
/// the originating ACP `session_id`, v0.15+), and exposes the result for
/// the chip to render.
///
/// **Concurrency.** `@MainActor + @Observable`. The polling task runs
/// on `Task.detached(priority: .utility)` per the project's Swift 6
/// rules; only the final `liveCount` write lands on MainActor. Single-
/// flight is enforced so a slow CLI doesn't stack up calls behind it.
///
/// **Lifecycle.** Hosts mount via `.task(id: …)`. The id should
/// include the chat session id, so a session swap restarts polling
/// cleanly. Cancellation is automatic when the view goes off-screen.
@Observable
@MainActor
final class KanbanChatBadgeViewModel {
    private let logger = Logger(
        subsystem: "com.scarf",
        category: "KanbanChatBadgeViewModel"
    )

    /// `nil` while the first poll hasn't returned yet OR while the
    /// host pre-dates kanban. Zero is a real value (idle board).
    private(set) var liveCount: Int?

    /// True if the chip should appear at all — false when polling is
    /// suppressed (e.g. capability-negative host, no chat session).
    private(set) var shouldRender: Bool = false

    private let context: ServerContext
    private let service: KanbanService

    /// Single-flight guard so a slow CLI invocation doesn't stack
    /// next-tick calls. The flag is non-actor-isolated since both
    /// reads and writes happen on the MainActor.
    private var isInflight = false

    /// Interval doubles on error up to this ceiling, then resets to
    /// the floor on first success. Keeps Scarf from hammering a
    /// failing host.
    private let baseInterval: TimeInterval = 5
    private let maxInterval: TimeInterval = 30
    private var currentInterval: TimeInterval = 5

    init(context: ServerContext) {
        self.context = context
        self.service = KanbanService(context: context)
    }

    /// Start the long-running poller. Call from `.task(id: …)` — the
    /// task is automatically cancelled when the view leaves the tree.
    /// `sessionId` is the originating ACP chat session id; the badge
    /// counts only tasks this chat produced.
    func run(
        sessionId: String,
        capabilities: HermesCapabilities
    ) async {
        guard capabilities.hasKanbanSessionFilter else {
            shouldRender = false
            liveCount = nil
            return
        }
        shouldRender = true
        currentInterval = baseInterval
        // Tick immediately so the chip's first render has data, then
        // sleep + tick on the configured interval.
        await poll(sessionId: sessionId)
        while !Task.isCancelled {
            do {
                try await Task.sleep(nanoseconds: UInt64(currentInterval * 1_000_000_000))
            } catch {
                return
            }
            try? Task.checkCancellation()
            await poll(sessionId: sessionId)
        }
    }

    private func poll(sessionId: String) async {
        if isInflight { return }
        isInflight = true
        defer { isInflight = false }

        let filter = KanbanListFilter(session: sessionId)
        do {
            let rows = try await service.list(filter)
            let count = rows.reduce(0) { acc, task in
                let typed = KanbanStatus.from(task.status)
                return acc + ((typed == .running || typed == .blocked) ? 1 : 0)
            }
            liveCount = count
            currentInterval = baseInterval
        } catch {
            logger.debug("kanban badge poll failed: \(error.localizedDescription, privacy: .public)")
            // Don't surface — the chip just goes dark on the next render
            // when the count flips back. Back off so a persistent
            // failure doesn't pin the CPU.
            liveCount = nil
            currentInterval = min(currentInterval * 2, maxInterval)
        }
    }
}
