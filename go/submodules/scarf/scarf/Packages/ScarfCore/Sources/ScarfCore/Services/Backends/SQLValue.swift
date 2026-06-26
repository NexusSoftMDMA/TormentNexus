import Foundation

/// Typed SQLite column value. Mirrors SQLite's storage classes
/// (`SQLITE_NULL`, `SQLITE_INTEGER`, `SQLITE_FLOAT`, `SQLITE_TEXT`,
/// `SQLITE_BLOB`) so both backends — libsqlite3 (`LocalSQLiteBackend`)
/// and remote `sqlite3 -json` parsing (`RemoteSQLiteBackend`) — can
/// produce and consume the same `Row` shape.
///
/// Used in two places:
///
/// 1. **Bound parameters**: callers hand `[SQLValue]` to
///    `HermesQueryBackend.query(_:params:)`. The local backend feeds
///    them into `sqlite3_bind_*`; the remote backend inlines them as
///    SQLite literals via `SQLValueInliner.inline(_:into:)`.
/// 2. **Result columns**: each `Row.values` entry is one of these.
///    Parsers (`sessionFromRow`, `messageFromRow` in HermesDataService)
///    read positional accessors like `row.string(at: 3)` to get the
///    typed value.
public enum SQLValue: Sendable, Equatable {
    case null
    case integer(Int64)
    case real(Double)
    case text(String)
    case blob(Data)
}

/// One result row from a query. Indexable both by position (matching the
/// libsqlite3 `sqlite3_column_*` ergonomics that `HermesDataService`'s
/// existing parsers expect) and by name (more readable for new code).
///
/// `columnIndex` is built once per result-set, not per row, so the
/// per-row overhead is just the `[SQLValue]` allocation.
public struct Row: Sendable {
    /// Ordered column values, indexable by their position in the
    /// underlying SELECT.
    public let values: [SQLValue]

    /// Column-name → position map. Built once per result-set by the
    /// backend, then shared (by reference) across every row in the
    /// set. Lookups are case-sensitive — match SQLite's default.
    public let columnIndex: [String: Int]

    public init(values: [SQLValue], columnIndex: [String: Int]) {
        self.values = values
        self.columnIndex = columnIndex
    }

    public subscript(_ position: Int) -> SQLValue {
        guard position >= 0, position < values.count else { return .null }
        return values[position]
    }

    public subscript(_ name: String) -> SQLValue {
        guard let i = columnIndex[name] else { return .null }
        return values[i]
    }

    // MARK: - Typed positional accessors
    //
    // These mirror the `columnText(stmt, i)` / `columnDate(stmt, i)`
    // helpers that lived in HermesDataService so the row-parser
    // migrations from `OpaquePointer` to `Row` are line-for-line.

    public func string(at i: Int) -> String {
        if case .text(let s) = self[i] { return s }
        return ""
    }

    public func optionalString(at i: Int) -> String? {
        switch self[i] {
        case .text(let s): return s
        case .null: return nil
        default: return nil
        }
    }

    public func int(at i: Int) -> Int {
        switch self[i] {
        case .integer(let n): return Int(n)
        case .real(let d): return Int(d)
        case .text(let s): return Int(s) ?? 0
        default: return 0
        }
    }

    public func optionalInt(at i: Int) -> Int? {
        switch self[i] {
        case .integer(let n): return Int(n)
        case .real(let d): return Int(d)
        case .text(let s): return Int(s)
        case .null: return nil
        default: return nil
        }
    }

    public func int64(at i: Int) -> Int64 {
        switch self[i] {
        case .integer(let n): return n
        case .real(let d): return Int64(d)
        case .text(let s): return Int64(s) ?? 0
        default: return 0
        }
    }

    public func double(at i: Int) -> Double {
        switch self[i] {
        case .real(let d): return d
        case .integer(let n): return Double(n)
        case .text(let s): return Double(s) ?? 0
        default: return 0
        }
    }

    public func optionalDouble(at i: Int) -> Double? {
        switch self[i] {
        case .real(let d): return d
        case .integer(let n): return Double(n)
        case .text(let s): return Double(s)
        case .null: return nil
        default: return nil
        }
    }

    /// Interpret the column as a Unix-epoch timestamp (seconds, fractional
    /// allowed). Returns `nil` when the column is NULL or unparseable.
    /// Mirrors the existing `columnDate` helper exactly.
    public func date(at i: Int) -> Date? {
        guard let secs = optionalDouble(at: i) else { return nil }
        return Date(timeIntervalSince1970: secs)
    }

    public func isNull(at i: Int) -> Bool {
        if case .null = self[i] { return true }
        return false
    }
}
