"""get_runtime_coverage — runtime coverage histogram + unmapped list (Phase 3).

Pairs with Phase 2's per-result ``_runtime_confidence`` stamping. Where
Phase 2 told you "this one symbol has runtime evidence", this tool tells
you "what fraction of the codebase (or this file) has runtime evidence",
plus the diagnostic list of runtime spans that point at code the AST
extractor missed.

Returns:
  ``{
      'repo': 'owner/name',
      'scope': 'repo' | 'file:<path>',
      'total_symbols': N,
      'confirmed': K,
      'declared_only': N - K,
      'coverage_pct': round(100*K/N),
      'sources': sorted list of ingested sources,
      'last_seen': ISO-8601 most-recent last_seen,
      'unmapped_runtime': list of unresolved span groups
                          (file_path, line_no, function_name, source, count, last_seen),
      '_meta': {timing_ms, ...},
  }``

Returns ``coverage_pct=0`` and an empty ``unmapped_runtime`` when no
traces have been ingested — same zero-cost contract as Phase 2.
"""

from __future__ import annotations

import sqlite3
import time
from typing import Optional

from ._utils import index_status_to_tool_error, resolve_repo
from ..storage import IndexStore


def get_runtime_coverage(
    repo: str,
    file_path: Optional[str] = None,
    *,
    unmapped_limit: int = 50,
    storage_path: Optional[str] = None,
) -> dict:
    """Return runtime coverage stats for a repo or single file.

    Args:
        repo: Repository identifier (``owner/name`` or just ``name``).
        file_path: Optional repo-relative file path. When set, scopes
            the histogram to that file.
        unmapped_limit: Cap on the number of unmapped-runtime entries
            returned for diagnostics. The full list lives in the
            ``runtime_unmapped`` table; this slice surfaces the loudest.
        storage_path: Custom storage path (matches other tools).

    Returns:
        A dict (see module docstring).
    """
    start = time.perf_counter()
    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    status = store.inspect_index(owner, name)
    if not status.loadable:
        return index_status_to_tool_error(status)
    db_path = store._sqlite._db_path(owner, name)  # type: ignore[attr-defined]
    if not db_path.exists():
        return index_status_to_tool_error(store.inspect_index(owner, name))

    scope = f"file:{file_path}" if file_path else "repo"
    response: dict = {
        "repo": f"{owner}/{name}",
        "scope": scope,
        "total_symbols": 0,
        "confirmed": 0,
        "declared_only": 0,
        "coverage_pct": 0,
        "sources": [],
        "last_seen": "",
        "unmapped_runtime": [],
    }

    # Read-only + immutable so we never bump WAL mtime and invalidate the
    # CodeIndex LRU cache (same pattern as runtime/confidence.py).
    conn = sqlite3.connect(f"file:{db_path}?mode=ro&immutable=1", uri=True)
    conn.row_factory = sqlite3.Row
    try:
        # Total symbols (in scope)
        if file_path:
            total = conn.execute(
                "SELECT COUNT(*) AS n FROM symbols WHERE file = ?",
                (file_path,),
            ).fetchone()["n"]
        else:
            total = conn.execute("SELECT COUNT(*) AS n FROM symbols").fetchone()["n"]
        response["total_symbols"] = total

        # Confirmed = symbols with at least one runtime_calls row (in scope)
        if file_path:
            confirmed = conn.execute(
                """
                SELECT COUNT(DISTINCT s.id) AS n
                FROM symbols s
                JOIN runtime_calls rc ON rc.symbol_id = s.id
                WHERE s.file = ?
                """,
                (file_path,),
            ).fetchone()["n"]
        else:
            confirmed = conn.execute(
                """
                SELECT COUNT(DISTINCT symbol_id) AS n
                FROM runtime_calls
                WHERE symbol_id IN (SELECT id FROM symbols)
                """
            ).fetchone()["n"]
        response["confirmed"] = confirmed
        response["declared_only"] = max(0, total - confirmed)
        response["coverage_pct"] = round(100 * confirmed / total) if total else 0

        # Sources + last_seen across the in-scope set
        if file_path:
            rows = conn.execute(
                """
                SELECT rc.source AS source, MAX(rc.last_seen) AS last_seen
                FROM runtime_calls rc
                JOIN symbols s ON s.id = rc.symbol_id
                WHERE s.file = ?
                GROUP BY rc.source
                """,
                (file_path,),
            ).fetchall()
        else:
            rows = conn.execute(
                """
                SELECT source, MAX(last_seen) AS last_seen
                FROM runtime_calls
                GROUP BY source
                """
            ).fetchall()
        sources: list[str] = []
        last_seen = ""
        for r in rows:
            sources.append(r["source"])
            if r["last_seen"] and r["last_seen"] > last_seen:
                last_seen = r["last_seen"]
        response["sources"] = sorted(sources)
        response["last_seen"] = last_seen

        # Unmapped runtime — span groups that didn't resolve to a symbol_id
        if file_path:
            unmapped_rows = conn.execute(
                """
                SELECT file_path, line_no, function_name, source, count, last_seen
                FROM runtime_unmapped
                WHERE file_path = ? OR file_path LIKE ?
                ORDER BY count DESC
                LIMIT ?
                """,
                (file_path, f"%/{file_path}", unmapped_limit),
            ).fetchall()
        else:
            unmapped_rows = conn.execute(
                """
                SELECT file_path, line_no, function_name, source, count, last_seen
                FROM runtime_unmapped
                ORDER BY count DESC
                LIMIT ?
                """,
                (unmapped_limit,),
            ).fetchall()
        response["unmapped_runtime"] = [dict(r) for r in unmapped_rows]
    finally:
        conn.close()

    elapsed = (time.perf_counter() - start) * 1000
    response["_meta"] = {
        "timing_ms": round(elapsed, 1),
        "tip": (
            "coverage_pct=0 with sources=[] means no traces ingested — "
            "run `import_runtime_signal` (or `jcodemunch-mcp import-trace --otel <file>`) first. "
            "unmapped_runtime entries are likely reflective dispatch / dynamic "
            "imports the AST extractor missed; spot-check a few to decide whether "
            "the resolver needs tuning or the codebase has genuine dynamic dispatch."
        ),
    }
    return response
