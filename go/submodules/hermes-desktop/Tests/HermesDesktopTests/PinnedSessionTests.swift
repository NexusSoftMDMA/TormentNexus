import Foundation
import Testing
@testable import HermesDesktop

struct PinnedSessionTests {
    @Test
    func pinnedSessionKeepsLocalSummarySnapshot() {
        let startedAt = SessionTimestamp.unixSeconds(1_766_800_000)
        let lastActive = SessionTimestamp.text("2026-05-03T18:20:00Z")
        let createdAt = Date(timeIntervalSince1970: 1_766_800_100)
        let updatedAt = Date(timeIntervalSince1970: 1_766_800_200)
        let summary = SessionSummary(
            id: "session-123",
            title: "Launch follow-up",
            model: "gpt-5.2",
            parentSessionID: "session-122",
            startedAt: startedAt,
            lastActive: lastActive,
            messageCount: 18,
            preview: "Release checklist"
        )

        let pinnedSession = PinnedSession(
            session: summary,
            workspaceScopeFingerprint: "host-workspace",
            createdAt: createdAt,
            updatedAt: updatedAt
        )

        #expect(pinnedSession.id == "session-123")
        #expect(pinnedSession.workspaceScopeFingerprint == "host-workspace")
        #expect(pinnedSession.createdAt == createdAt)
        #expect(pinnedSession.updatedAt == updatedAt)
        #expect(pinnedSession.parentSessionID == "session-122")
        #expect(pinnedSession.summary == summary)
    }
}
