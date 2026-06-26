import Foundation
import Testing
@testable import HermesDesktop

struct SessionSearchModelsTests {
    @Test
    func sessionSummaryDecodesOptionalSearchMatch() throws {
        let data = Data("""
        {
          "id": "session-123",
          "title": "Release check",
          "model": "gpt-5.2",
          "parent_session_id": "session-root",
          "started_at": 1766800000,
          "last_active": "2026-05-03T18:20:00Z",
          "message_count": 12,
          "preview": "Earlier context",
          "search_match": {
            "match_count": 3,
            "message_id": "msg-9",
            "role": "assistant",
            "timestamp": 1766800100,
            "snippet": "Found the release blocker in the transcript."
          }
        }
        """.utf8)

        let summary = try JSONDecoder().decode(SessionSummary.self, from: data)

        #expect(summary.id == "session-123")
        #expect(summary.parentSessionID == "session-root")
        #expect(summary.searchMatch?.matchCount == 3)
        #expect(summary.searchMatch?.messageID == "msg-9")
        #expect(summary.searchMatch?.role == .assistant)
        #expect(summary.searchMatch?.snippet == "Found the release blocker in the transcript.")
    }

    @Test
    func sessionSummaryStillDecodesWithoutSearchMatch() throws {
        let data = Data("""
        {
          "id": "session-456",
          "title": null,
          "model": null,
          "started_at": null,
          "last_active": null,
          "message_count": null,
          "preview": null
        }
        """.utf8)

        let summary = try JSONDecoder().decode(SessionSummary.self, from: data)

        #expect(summary.id == "session-456")
        #expect(summary.searchMatch == nil)
    }
}
