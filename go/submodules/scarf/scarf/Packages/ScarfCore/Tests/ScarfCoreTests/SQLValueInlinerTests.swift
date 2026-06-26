import Testing
import Foundation
@testable import ScarfCore

/// Pure unit tests on `try SQLValueInliner.inline(_:params:)` and
/// `SQLValueInliner.encode(_:)`. No backend, no transport, no actor —
/// these are the lexical-substitution rules that drive the remote
/// SQLite backend's `?` → literal pipeline.
@Suite struct SQLValueInlinerTests {

    // MARK: - encode(_:) per SQLValue case

    @Test func encodeNullProducesNULL() {
        #expect(SQLValueInliner.encode(.null) == "NULL")
    }

    @Test func encodeIntegerProducesUnquotedDigits() {
        #expect(SQLValueInliner.encode(.integer(42)) == "42")
        #expect(SQLValueInliner.encode(.integer(-7)) == "-7")
        #expect(SQLValueInliner.encode(.integer(0)) == "0")
        #expect(SQLValueInliner.encode(.integer(Int64.max)) == "9223372036854775807")
    }

    @Test func encodeRealUsesPercent17gFormat() {
        // %.17g round-trips a Double precisely as decimal. Verify the
        // formatted string parses back to the exact same Double.
        let original: Double = 3.14
        let encoded = SQLValueInliner.encode(.real(original))
        #expect(encoded == String(format: "%.17g", original))
        // Round-trip: encoded value re-parsed must equal the source.
        #expect(Double(encoded) == original)

        // Tricky case: 0.1 + 0.2 has imprecise binary representation.
        let imprecise = 0.1 + 0.2
        let encodedImprecise = SQLValueInliner.encode(.real(imprecise))
        #expect(Double(encodedImprecise) == imprecise)
    }

    @Test func encodeTextWrapsInSingleQuotes() {
        #expect(SQLValueInliner.encode(.text("hi")) == "'hi'")
        #expect(SQLValueInliner.encode(.text("")) == "''")
    }

    @Test func encodeTextDoublesEmbeddedSingleQuotes() {
        // SQL literal escape: `it's` becomes `'it''s'`.
        #expect(SQLValueInliner.encode(.text("it's")) == "'it''s'")
        // Multiple embedded quotes — each one is doubled.
        #expect(SQLValueInliner.encode(.text("a'b'c")) == "'a''b''c'")
        // The classic injection-shaped value gets escaped to harmless.
        #expect(SQLValueInliner.encode(.text("' OR 1=1 --")) == "''' OR 1=1 --'")
    }

    @Test func encodeBlobProducesHexLiteral() {
        // Two-byte blob: `X'dead'`.
        #expect(SQLValueInliner.encode(.blob(Data([0xde, 0xad]))) == "X'dead'")
        // Empty blob: `X''`.
        #expect(SQLValueInliner.encode(.blob(Data())) == "X''")
        // Lowercase hex, full byte range, with leading zero preserved.
        #expect(SQLValueInliner.encode(.blob(Data([0x00, 0x0f, 0xff]))) == "X'000fff'")
    }

    // MARK: - inline(_:params:) substitution rules

    @Test func inlineSubstitutesPlaceholdersInOrder() throws {
        let out = try SQLValueInliner.inline(
            "INSERT INTO t VALUES (?, ?, ?)",
            params: [.integer(1), .text("two"), .real(3.0)]
        )
        // Order is preserved: integer 1, text 'two', real 3.0.
        #expect(out.hasPrefix("INSERT INTO t VALUES ("))
        #expect(out.contains("1"))
        #expect(out.contains("'two'"))
        // Real 3.0 should round-trip via %.17g.
        let real3 = String(format: "%.17g", 3.0)
        #expect(out.contains(real3))
    }

    @Test func inlineSkipsPlaceholderInsideStringLiteral() throws {
        // The `?` inside `'?'` is part of a string and must not be bound.
        // Only the trailing `?` (outside the quotes) consumes the param.
        let out = try SQLValueInliner.inline(
            "WHERE name = '?' AND id = ?",
            params: [.integer(7)]
        )
        #expect(out == "WHERE name = '?' AND id = 7")
    }

    @Test func inlineSkipsPlaceholderInsideDoubleQuotedIdentifier() throws {
        // Double-quoted identifiers (column / table names with special chars)
        // are also a quoted region — `?` inside them is literal.
        let out = try SQLValueInliner.inline(
            "SELECT \"col?\" FROM t WHERE x = ?",
            params: [.integer(1)]
        )
        #expect(out == "SELECT \"col?\" FROM t WHERE x = 1")
    }

    @Test func inlineHandlesDoubledSingleQuoteEscapeInString() throws {
        // `'it''s ?'` is a single SQL string literal containing `it's ?`.
        // The doubled single-quote is the SQL escape for an embedded
        // apostrophe — the scanner must NOT toggle out of string state
        // at the doubled quote, and the trailing `?` is inside the string.
        // No params consumed.
        let out = try SQLValueInliner.inline(
            "WHERE x = 'it''s ?'",
            params: []
        )
        #expect(out == "WHERE x = 'it''s ?'")
    }

    @Test func inlineSelectShapeMatchesDataServicePattern() throws {
        // Sanity check: the SELECT shape `HermesDataService.fetchSessions`
        // generates inlines cleanly for the typical `[.integer(100)]`
        // limit param.
        let sql = "SELECT id, source FROM sessions WHERE parent_session_id IS NULL ORDER BY started_at DESC LIMIT ?"
        let out = try SQLValueInliner.inline(sql, params: [.integer(100)])
        #expect(out == "SELECT id, source FROM sessions WHERE parent_session_id IS NULL ORDER BY started_at DESC LIMIT 100")
    }

    @Test func inlineWithNoPlaceholdersReturnsInputUnchanged() throws {
        let sql = "SELECT COUNT(*) FROM messages"
        #expect(try SQLValueInliner.inline(sql, params: []) == sql)
    }

    @Test func inlinePreservesAllOtherCharacters() throws {
        // Make sure we're not mangling whitespace, semicolons, parens.
        let sql = "  SELECT  *\n  FROM   t  WHERE id = ?  ;  "
        let out = try SQLValueInliner.inline(sql, params: [.integer(5)])
        #expect(out == "  SELECT  *\n  FROM   t  WHERE id = 5  ;  ")
    }

    @Test func inlineSubstitutesNullPlaceholder() throws {
        let out = try SQLValueInliner.inline(
            "UPDATE t SET col = ? WHERE id = ?",
            params: [.null, .integer(1)]
        )
        #expect(out == "UPDATE t SET col = NULL WHERE id = 1")
    }

    @Test func inlineSubstitutesBlobPlaceholder() throws {
        let out = try SQLValueInliner.inline(
            "INSERT INTO t (data) VALUES (?)",
            params: [.blob(Data([0x01, 0x02, 0x03]))]
        )
        #expect(out == "INSERT INTO t (data) VALUES (X'010203')")
    }

    // MARK: - inline(_:params:) error path (t-aud08)

    @Test func inlineThrowsWhenMorePlaceholdersThanParams() {
        // Was a fatalError (whole-app crash); now a recoverable throw.
        #expect(throws: SQLValueInliner.InlineError.self) {
            _ = try SQLValueInliner.inline("WHERE a = ? AND b = ?", params: [.integer(1)])
        }
    }

    @Test func inlineThrowsWhenFewerPlaceholdersThanParams() {
        #expect(throws: SQLValueInliner.InlineError.self) {
            _ = try SQLValueInliner.inline("WHERE a = ?", params: [.integer(1), .integer(2)])
        }
    }
}
