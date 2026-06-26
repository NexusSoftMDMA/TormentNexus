"""find_hot_paths — top-N runtime-confirmed symbols by hit count (Phase 3).

Pairs naturally with ``get_blast_radius`` and ``get_pr_risk_profile``:
the agent learns "this PR touches a function called 4M times/day" before
deciding how aggressively to review.

Returns symbols sorted by their summed runtime ``count`` across all
ingested sources, optionally filtered by a name substring.

Returns:
  ``{
      'repo': 'owner/name',
      'query': '<query or None>',
      'top_n': N,
      'results': [
          {
              'symbol_id', 'name', 'kind', 'file', 'line',
              'runtime_count': total across sources,
              'p50_ms', 'p95_ms',
              'sources': ['otel', ...],
              'last_seen': ISO-8601,
              'first_seen': ISO-8601,
          },
          ...
      ],
      '_meta': {timing_ms, ...},
  }``

Empty ``results`` when no traces have been ingested — same zero-cost
contract as Phase 2.
"""

from __future__ import annotations

import sqlite3
import time
from typing import Optional

from ._utils import index_status_to_tool_error, resolve_repo
from ..storage import IndexStore


def find_hot_paths(
    repo: str,
    query: Optional[str] = None,
    top_n: int = 20,
    *,
    storage_path: Optional[str] = None,
) -> dict:
    """Return the top-N symbols by total runtime hit count.

    Args:
        repo: Repository identifier (``owner/name`` or just ``name``).
        query: Optional case-insensitive substring filter on symbol name
            or qualified_name.
        top_n: Cap on returned rows (default 20, max 200).
        storage_path: Custom storage path.

    Returns:
        See module docstring.
    """
    start = time.perf_counter()
    top_n = max(1, min(top_n, 200))
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

    conn = sqlite3.connect(f"file:{db_path}?mode=ro&immutable=1", uri=True)
    conn.row_factory = sqlite3.Row
    try:
        # Aggregate per symbol_id across sources, then join symbols for
        # display fields. ``MAX(p95_ms)`` and ``MAX(p50_ms)`` are the
        # right-shape-for-now reduction across sources — Phase 4 will
        # replace with a merge-correct streaming-quantile when SQL log /
        # stack log channels land.
        params: list = []
        like_clause = ""
        if query:
            like_clause = "AND (LOWER(s.name) LIKE ? OR LOWER(COALESCE(s.qualified_name, '')) LIKE ?)"
            q = f"%{query.lower()}%"
            params.extend([q, q])
        params.append(top_n)
        rows = conn.execute(
            f"""
            SELECT
                s.id           AS symbol_id,
                s.name         AS name,
                s.kind         AS kind,
                s.file         AS file,
                s.line         AS line,
                SUM(rc.count)  AS runtime_count,
                MAX(rc.p50_ms) AS p50_ms,
                MAX(rc.p95_ms) AS p95_ms,
                MAX(rc.last_seen)  AS last_seen,
                MIN(rc.first_seen) AS first_seen,
                GROUP_CONCAT(DISTINCT rc.source) AS sources
            FROM symbols s
            JOIN runtime_calls rc ON rc.symbol_id = s.id
            WHERE 1 = 1 {like_clause}
            GROUP BY s.id
            ORDER BY runtime_count DESC, last_seen DESC
            LIMIT ?
            """,
            params,
        ).fetchall()
        results = []
        for r in rows:
            sources = sorted(set((r["sources"] or "").split(","))) if r["sources"] else []
            sources = [s for s in sources if s]
            results.append({
                "symbol_id": r["symbol_id"],
                "name": r["name"],
                "kind": r["kind"],
                "file": r["file"],
                "line": r["line"],
                "runtime_count": int(r["runtime_count"] or 0),
                "p50_ms": r["p50_ms"],
                "p95_ms": r["p95_ms"],
                "sources": sources,
                "first_seen": r["first_seen"] or "",
                "last_seen": r["last_seen"] or "",
            })
    finally:
        conn.close()

    elapsed = (time.perf_counter() - start) * 1000
    return {
        "repo": f"{owner}/{name}",
        "query": query,
        "top_n": top_n,
        "results": results,
        "_meta": {
            "timing_ms": round(elapsed, 1),
            "tip": (
                "Empty results means no traces ingested OR no runtime data matches the filter. "
                "Pair with get_blast_radius on a hot path to see how PR-risk the touched code is."
            ),
        },
    }
