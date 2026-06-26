import Testing
import Foundation
@testable import ScarfCore

/// Verifies the lenient `ListItemStatus(raw:)` parser. Real dashboards on
/// disk use a mix of canonical names + synonyms (`done`, `info`, `ok`,
/// `pending`, `up` are seen on the dev's machine today) — the parser must
/// fold those onto the canonical case set without throwing or returning nil
/// for the common synonyms. Unknown strings → nil so the renderer can fall
/// back to plain text without losing the original.
@Suite struct ListItemStatusTests {
    @Test func canonicalNamesParse() {
        for c in ListItemStatus.allCases {
            #expect(ListItemStatus(raw: c.rawValue) == c)
        }
    }

    @Test func synonymsCollapseToCanonical() {
        #expect(ListItemStatus(raw: "ok") == .success)
        #expect(ListItemStatus(raw: "OK") == .success)        // case-insensitive
        #expect(ListItemStatus(raw: " up ") == .success)      // whitespace trim
        #expect(ListItemStatus(raw: "down") == .danger)
        #expect(ListItemStatus(raw: "error") == .danger)
        #expect(ListItemStatus(raw: "failed") == .danger)
        #expect(ListItemStatus(raw: "warn") == .warning)
        #expect(ListItemStatus(raw: "degraded") == .warning)
        #expect(ListItemStatus(raw: "active") == .info)
        #expect(ListItemStatus(raw: "queued") == .pending)
        #expect(ListItemStatus(raw: "complete") == .done)
    }

    @Test func unknownReturnsNilNotThrows() {
        #expect(ListItemStatus(raw: "hologram") == nil)
        #expect(ListItemStatus(raw: "") == nil)
        #expect(ListItemStatus(raw: nil) == nil)
        #expect(ListItemStatus(raw: "   ") == nil)
    }

    @Test func listItemStillDecodesUnknownStatusString() throws {
        // Backwards-compat invariant: `ListItem.status` stays a free String? on
        // the wire. Decoding a v2.6 dashboard with a non-canonical status must
        // succeed and preserve the original string (renderer falls back).
        let json = #"{"text":"foo","status":"weird"}"#.data(using: .utf8)!
        let item = try JSONDecoder().decode(ListItem.self, from: json)
        #expect(item.status == "weird")
        #expect(ListItemStatus(raw: item.status) == nil)
    }
}
