"""Phase 4 SQL query log parser — pg_stat_statements + generic SQL log.

Pure parsing layer (no DB writes). Yields ``SqlQueryRecord`` instances
with a redaction-friendly *normalized* query plus extracted table and
column references.

Two file shapes are supported:

1. **pg_stat_statements CSV export.** Header row required. Recognised
   columns (case-insensitive): ``query`` / ``calls`` / ``total_time`` /
   ``total_exec_time`` / ``mean_time`` / ``mean_exec_time``. The
   columns missing from a particular Postgres version are silently
   tolerated; we only need ``query`` and ``calls``.

2. **Generic SQL JSON-Lines log.** Each line: ``{"sql": "...",
   "duration_ms": 12.3, "ts": "2026-05-09T12:00:00Z"}``. ``sql`` is
   required; ``duration_ms`` and ``ts`` are optional. Application
   loggers (Datadog, OpenSearch, vector.dev, custom shims) all converge
   on this shape after their SQL pipeline is mapped through.

Reference extraction is deliberately lightweight regex — full SQL
parsing is the wrong tool for a counts-only rollup. Heuristics:

* **Tables** — token after ``FROM`` / ``JOIN`` / ``INSERT INTO`` /
  ``UPDATE`` / ``DELETE FROM`` / ``MERGE INTO``. Schema-qualified
  names (``analytics.fact_orders``) keep only the trailing identifier
  (``fact_orders``) since that's how dbt model names land in the index.
* **Columns** — identifiers that appear in ``SELECT ... FROM`` lists,
  ``WHERE`` clauses, ``GROUP BY`` / ``ORDER BY`` / ``ON`` / ``HAVING``.
  Filtered against a permissive set of SQL keywords. Never reach the
  index unless they later match a real (model, column) pair via the
  ``dbt_columns`` metadata.

The parser's contract: ``parse_sql_log_file(path)`` is iterable; one
``SqlQueryRecord`` per logical query. Empty / comment-only files yield
nothing.
"""

from __future__ import annotations

import csv
import gzip
import io
import json
import logging
import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Iterator, Optional

logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Public types
# ---------------------------------------------------------------------------


@dataclass
class SqlQueryRecord:
    """One logical query observed in a SQL log."""

    sql: str
    """Raw SQL text (already pg_stat_statements-normalized in the CSV path)."""

    calls: int = 1
    """Execution count for this row. pg_stat_statements rolls these up;
    generic JSON-Lines defaults to 1 per record."""

    total_ms: Optional[float] = None
    """Total exec time in ms for the row, if known. None = no timing data."""

    mean_ms: Optional[float] = None
    """Mean per-call exec time in ms for the row, if known."""

    timestamp: Optional[str] = None
    """ISO-8601 timestamp string if the log line carried one."""

    tables: list[str] = field(default_factory=list)
    """Trailing identifiers extracted from FROM/JOIN/INTO/UPDATE/DELETE."""

    columns: list[tuple[str, str]] = field(default_factory=list)
    """``(table_name, column_name)`` pairs, best-effort. ``table_name`` is
    the empty string when the column was unqualified — the resolver will
    try to match against any table referenced in the query."""


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def parse_sql_log_file(path: str) -> Iterator[SqlQueryRecord]:
    """Yield SqlQueryRecord for every query in the file.

    Format dispatch:
      * ``.csv`` → pg_stat_statements parser
      * ``.json`` / ``.jsonl`` / ``.log`` → JSON-Lines parser
      * ``.gz`` → strip the suffix and re-dispatch on the inner extension
      * anything else → JSON-Lines fallback (most flexible)
    """
    p = Path(path)
    if not p.exists():
        raise FileNotFoundError(f"SQL log file not found: {path}")

    inner_suffix = p.suffix.lower()
    is_gz = inner_suffix == ".gz"
    if is_gz:
        # ``foo.csv.gz`` → take the inner ``.csv``
        inner_suffix = Path(p.stem).suffix.lower()

    opener = (lambda: gzip.open(p, "rt", encoding="utf-8")) if is_gz else (
        lambda: open(p, "r", encoding="utf-8")
    )

    if inner_suffix == ".csv":
        yield from _parse_pg_stat_statements_csv(opener)
    else:
        # Default to JSON-Lines for .jsonl / .json / .log / unknown.
        yield from _parse_json_lines(opener)


def iter_sql_from_text(text: str, *, fmt: str = "auto") -> Iterator[SqlQueryRecord]:
    """Yield SqlQueryRecord from an in-memory SQL log payload.

    Used by the Phase 6 HTTP route. ``fmt`` selects the parser:
      * ``'auto'`` (default) — heuristic: if the first non-whitespace
        line contains ``,`` and looks header-shaped (no obvious SQL
        keyword) treat it as CSV; otherwise JSON-Lines.
      * ``'csv'`` — force pg_stat_statements parsing.
      * ``'jsonl'`` — force JSON-Lines parsing.
    """
    if not text.strip():
        return

    if fmt == "auto":
        head = text.lstrip()
        # JSON-Lines / arrays start with `{` or `[`. Anything else is CSV.
        fmt = "jsonl" if head[:1] in ("{", "[") else "csv"

    def _opener_str() -> "io.StringIO":
        return io.StringIO(text)

    if fmt == "csv":
        yield from _parse_pg_stat_statements_csv(_opener_str)
    else:
        yield from _parse_json_lines(_opener_str)


# ---------------------------------------------------------------------------
# CSV (pg_stat_statements)
# ---------------------------------------------------------------------------


def _parse_pg_stat_statements_csv(opener) -> Iterator[SqlQueryRecord]:
    """Parse a pg_stat_statements CSV export.

    Required column: ``query``. Optional: ``calls``, ``total_time`` /
    ``total_exec_time``, ``mean_time`` / ``mean_exec_time``. Header row
    must be present — pg_stat_statements exports include one by default
    via ``\\copy ... CSV HEADER``.
    """
    with opener() as f:
        reader = csv.DictReader(f)
        if reader.fieldnames is None:
            return
        # Build case-insensitive header map
        header_map = {h.lower().strip(): h for h in reader.fieldnames if h}
        query_col = header_map.get("query")
        if not query_col:
            logger.warning("pg_stat_statements CSV missing 'query' column; skipping")
            return
        calls_col = header_map.get("calls")
        total_col = header_map.get("total_exec_time") or header_map.get("total_time")
        mean_col = header_map.get("mean_exec_time") or header_map.get("mean_time")

        for row in reader:
            sql = (row.get(query_col) or "").strip()
            if not sql:
                continue
            calls = _safe_int(row.get(calls_col) if calls_col else None, default=1)
            total_ms = _safe_float(row.get(total_col)) if total_col else None
            mean_ms = _safe_float(row.get(mean_col)) if mean_col else None
            yield _build_record(sql=sql, calls=calls, total_ms=total_ms, mean_ms=mean_ms)


# ---------------------------------------------------------------------------
# JSON-Lines (generic SQL log)
# ---------------------------------------------------------------------------


def _parse_json_lines(opener) -> Iterator[SqlQueryRecord]:
    """Parse a JSON-Lines SQL log. One ``{"sql": "..."}`` object per line.

    Tolerates blank lines and comment lines (``# ...`` or ``// ...``).
    Tolerates a top-level JSON array as a fallback (whole-file ``[{...}, ...]``)
    so users can hand us either shape.
    """
    with opener() as f:
        first_chunk = f.read(4096)
        # If the file starts with ``[`` it's a top-level array — slurp it.
        stripped = first_chunk.lstrip()
        if stripped.startswith("["):
            rest = f.read()
            try:
                payload = json.loads(first_chunk + rest)
            except json.JSONDecodeError as exc:
                logger.warning("SQL log: top-level array decode failed: %s", exc)
                return
            if not isinstance(payload, list):
                return
            for obj in payload:
                rec = _record_from_json_obj(obj)
                if rec is not None:
                    yield rec
            return

        # Stream the rest as JSON-Lines
        line_iter = io.StringIO(first_chunk + f.read()).readlines()

    for raw in line_iter:
        line = raw.strip()
        if not line or line.startswith("#") or line.startswith("//"):
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            logger.debug("SQL log: skipping malformed JSONL line", exc_info=True)
            continue
        rec = _record_from_json_obj(obj)
        if rec is not None:
            yield rec


def _record_from_json_obj(obj) -> Optional[SqlQueryRecord]:
    if not isinstance(obj, dict):
        return None
    sql = obj.get("sql") or obj.get("query") or obj.get("statement")
    if not isinstance(sql, str) or not sql.strip():
        return None
    calls = _safe_int(obj.get("calls") or obj.get("count"), default=1)
    duration = obj.get("duration_ms")
    if duration is None:
        duration = obj.get("mean_ms") or obj.get("elapsed_ms")
    duration_ms = _safe_float(duration)
    ts = obj.get("ts") or obj.get("timestamp")
    return _build_record(
        sql=sql.strip(),
        calls=calls,
        total_ms=None,
        mean_ms=duration_ms,
        timestamp=ts if isinstance(ts, str) else None,
    )


# ---------------------------------------------------------------------------
# Reference extraction
# ---------------------------------------------------------------------------

# Identifier: A-Za-z_ followed by A-Za-z0-9_ (dot allowed for schema.table).
# Quoted identifiers ("foo"."bar") get their quotes stripped before matching.
_IDENT = r'(?:"[^"]+"|[A-Za-z_][A-Za-z0-9_]*)'
_QUALIFIED = rf'(?:{_IDENT}\.)*{_IDENT}'

_TABLE_RE = re.compile(
    rf'(?is)\b(?:FROM|JOIN|INTO|UPDATE|DELETE\s+FROM|MERGE\s+INTO|TABLE)\s+(?P<ref>{_QUALIFIED})'
)

# SELECT list grabs everything between SELECT and FROM (non-greedy across newlines)
_SELECT_LIST_RE = re.compile(r'(?is)\bSELECT\b\s+(?:DISTINCT\s+)?(?P<list>.+?)\s+\bFROM\b')

# WHERE/ON/HAVING/GROUP BY/ORDER BY identifier scan
_PREDICATE_BLOCK_RE = re.compile(
    r'(?is)\b(?:WHERE|ON|HAVING|GROUP\s+BY|ORDER\s+BY)\b(?P<body>.+?)(?=(?:\b(?:WHERE|ON|HAVING|GROUP\s+BY|ORDER\s+BY|UNION|INTERSECT|EXCEPT|LIMIT|OFFSET|RETURNING)\b|;|$))'
)

# Qualified column ref: ``alias.col`` — first half = table-or-alias, second = column.
_QUALIFIED_COL_RE = re.compile(
    rf'(?P<table>{_IDENT})\.(?P<col>{_IDENT})'
)

# Plain column refs in a select list: skip ``*``, function calls, literals.
_BARE_IDENT_RE = re.compile(rf'(?<![A-Za-z_."]){_IDENT}(?!\s*\()')

# Tokens that look like identifiers but are SQL reserved words — never columns.
_SQL_KEYWORDS = frozenset({
    "select", "from", "where", "and", "or", "not", "null", "true", "false",
    "as", "on", "in", "is", "by", "group", "order", "having", "limit",
    "offset", "union", "intersect", "except", "with", "case", "when", "then",
    "else", "end", "between", "like", "ilike", "exists", "any", "all", "some",
    "join", "left", "right", "inner", "outer", "full", "cross", "lateral",
    "using", "natural", "distinct", "asc", "desc", "nulls", "first", "last",
    "into", "values", "insert", "update", "delete", "set", "returning",
    "create", "drop", "alter", "table", "view", "index", "primary", "key",
    "foreign", "references", "constraint", "default", "unique", "check",
    "cast", "extract", "interval", "current_date", "current_time",
    "current_timestamp", "current_user", "session_user", "user", "now",
    "true", "false", "null", "unknown", "merge", "matched",
})


def _build_record(
    *,
    sql: str,
    calls: int = 1,
    total_ms: Optional[float] = None,
    mean_ms: Optional[float] = None,
    timestamp: Optional[str] = None,
) -> SqlQueryRecord:
    tables = _extract_tables(sql)
    columns = _extract_columns(sql, tables)
    return SqlQueryRecord(
        sql=sql,
        calls=calls,
        total_ms=total_ms,
        mean_ms=mean_ms,
        timestamp=timestamp,
        tables=tables,
        columns=columns,
    )


def _extract_tables(sql: str) -> list[str]:
    """Return distinct trailing identifiers referenced in FROM/JOIN/INTO/etc.

    ``analytics.fact_orders`` → ``fact_orders``.
    Quoted identifiers have their quotes stripped.
    Order is preserved (first-seen wins) so the resolver picks the
    fact table over a joined dimension when ranking.
    """
    out: list[str] = []
    seen: set[str] = set()
    for m in _TABLE_RE.finditer(sql):
        ref = m.group("ref")
        # Take the trailing component of a dotted name
        last = ref.rsplit(".", 1)[-1].strip().strip('"')
        if not last or last.lower() in _SQL_KEYWORDS:
            continue
        if last not in seen:
            seen.add(last)
            out.append(last)
    return out


def _extract_columns(sql: str, tables: list[str]) -> list[tuple[str, str]]:
    """Return distinct ``(table_name, column_name)`` pairs.

    Strategy:
      1. Walk every qualified ref ``alias.col`` across the entire query.
         Aliases get resolved against the table list later; for now we
         pass the alias through as ``table_name``.
      2. Walk the SELECT list and predicate blocks for bare identifiers.
         Bare identifiers get table_name="" — the upstream resolver
         tries them against every table referenced in the query.
    """
    out: list[tuple[str, str]] = []
    seen: set[tuple[str, str]] = set()

    # Qualified refs: alias.col
    for m in _QUALIFIED_COL_RE.finditer(sql):
        table = m.group("table").strip('"')
        col = m.group("col").strip('"')
        if col == "*" or col.lower() in _SQL_KEYWORDS:
            continue
        key = (table, col)
        if key not in seen:
            seen.add(key)
            out.append(key)

    # Bare identifiers in SELECT list
    select_match = _SELECT_LIST_RE.search(sql)
    if select_match:
        body = select_match.group("list")
        # Strip qualified refs we already captured; mask them out so the
        # bare-ident regex doesn't double-count the column part.
        masked = _QUALIFIED_COL_RE.sub(" ", body)
        for m in _BARE_IDENT_RE.finditer(masked):
            ident = m.group(0).strip('"')
            if ident == "*" or ident.lower() in _SQL_KEYWORDS:
                continue
            if ident.isdigit():
                continue
            key = ("", ident)
            if key not in seen:
                seen.add(key)
                out.append(key)

    # Bare identifiers in predicate blocks (WHERE/ON/GROUP BY/ORDER BY/HAVING)
    for m in _PREDICATE_BLOCK_RE.finditer(sql):
        body = m.group("body")
        masked = _QUALIFIED_COL_RE.sub(" ", body)
        for tok_match in _BARE_IDENT_RE.finditer(masked):
            ident = tok_match.group(0).strip('"')
            if ident == "*" or ident.lower() in _SQL_KEYWORDS:
                continue
            if ident.isdigit():
                continue
            # Single-letter aliases (a/b/x/y) are noise here.
            if len(ident) == 1:
                continue
            key = ("", ident)
            if key not in seen:
                seen.add(key)
                out.append(key)

    return out


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _safe_int(v, *, default: int = 0) -> int:
    if v is None:
        return default
    try:
        return int(float(v))
    except (TypeError, ValueError):
        return default


def _safe_float(v) -> Optional[float]:
    if v is None or v == "":
        return None
    try:
        return float(v)
    except (TypeError, ValueError):
        return None
