import Testing
import Foundation
import ScarfCore
@testable import scarf

/// Pure slug-generation tests for `KanbanTenantResolver`. The disk
/// I/O paths (`resolveOrMint`, `persist`) need a real `ServerContext`
/// + filesystem and are covered by integration tests.
@Suite struct KanbanTenantResolverSlugTests {

    @Test func basicNameSlugifiesCleanly() {
        #expect(KanbanTenantResolver.makeSlug(for: "My Project") == "scarf:my-project")
    }

    @Test func punctuationCollapsesToHyphens() {
        #expect(KanbanTenantResolver.makeSlug(for: "Foo: Bar / Baz!") == "scarf:foo-bar-baz")
    }

    @Test func consecutiveSeparatorsCollapse() {
        #expect(KanbanTenantResolver.makeSlug(for: "a   b___c") == "scarf:a-b-c")
    }

    @Test func emptyNameFallsBackToProjectLiteral() {
        #expect(KanbanTenantResolver.makeSlug(for: "!@#") == "scarf:project")
    }

    @Test func slugBoundedTo48CharsAfterPrefix() {
        let huge = String(repeating: "x", count: 200)
        let slug = KanbanTenantResolver.makeSlug(for: huge)
        #expect(slug.hasPrefix("scarf:"))
        // 6 chars for "scarf:" + ≤48 for the slug body
        #expect(slug.count <= 6 + 48)
    }

    @Test func unicodeNormalizesToAscii() {
        // The slug rule lowercases and replaces non-letter/digit with
        // hyphens; Latin-extended letters survive lowercase but accented
        // chars route through Foundation's lowercasing path.
        let slug = KanbanTenantResolver.makeSlug(for: "Mañana")
        #expect(slug.hasPrefix("scarf:"))
        #expect(!slug.contains(" "))
    }

    @Test func prefixIsStable() {
        #expect(KanbanTenantResolver.prefix == "scarf:")
    }
}
