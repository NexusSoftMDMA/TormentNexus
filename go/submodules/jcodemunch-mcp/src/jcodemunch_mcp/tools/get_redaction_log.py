"""get_redaction_log — forensic accounting of PII redactions (Phase 6).

Reads the ``runtime_redaction_log`` table populated by every ingest
chokepoint (OTel / SQL / stack — Phases 1, 4, 5) and now also by the
HTTP live-ingest endpoint (Phase 6). Operators run this tool to verify
that PII redaction is firing on production traffic — *before* trusting
that "yes, my SQL string literals were stripped before they hit disk."

The redaction patterns are documented in ``runtime/redact.py``. Common
labels seen in practice:

  * ``sql_string_literal``      — quoted SQL literals (``WHERE x = 'foo'``)
  * ``sql_numeric_param``       — bare numbers after ``=`` / ``IN`` / ``BETWEEN``
  * ``json_value_string``       — ``"key": "value"`` blocks in attribute bags
  * ``python_locals_block``     — ``kwargs={...}`` / ``vars={...}`` repr blocks
  * ``ipv4_address``            — IPv4
  * ``email_address``           — RFC-shaped emails
  * any of the secrets-registry labels (``aws_access_key``, ``jwt``,
    ``github_pat``, ``slack_token``, etc.) when those patterns fire

Returns:
  ``{
      'repo': 'owner/name',
      'sources': [<source>, ...],          # sources actually present
      'since_iso': cutoff,
      'patterns': [
          {'source', 'pattern', 'count', 'last_redacted'},
          ...
      ],
      'total_redactions': N,
      '_meta': {timing_ms, ...}
  }``
"""

from __future__ import annotations

import sqlite3
import time
from datetime import datetime, timedelta, timezone
from typing import Optional

from ._utils import index_status_to_tool_error, resolve_repo
from ..storage import IndexStore


_VALID_SOURCES = ("otel", "sql_log", "stack_log", "apm")


def get_redaction_log(
    repo: str,
    source: Optional[str] = None,
    *,
    since_days: int = 30,
    storage_path: Optional[str] = None,
) -> dict:
    """Return per-pattern redaction counts for a repo.

    Args:
        repo: ``owner/name`` (or bare name resolving to ``local``).
        source: Optional filter to a single source label
            (``otel`` / ``sql_log`` / ``stack_log`` / ``apm``).
            Default ``None`` returns all sources.
        since_days: Lookback window for ``last_redacted`` filtering.
            Patterns last fired before the cutoff are omitted. Default 30.
        storage_path: Custom storage path.

    Returns:
        See module docstring.
    """
    start = time.perf_counter()
    if source is not None and source not in _VALID_SOURCES:
        return {"error": f"unknown source {source!r}; valid: {list(_VALID_SOURCES)}"}
    since_days = max(1, int(since_days))

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as exc:
        return {"error": str(exc)}

    store = IndexStore(base_path=storage_path)
    status = store.inspect_index(owner, name)
    if not status.loadable:
        return index_status_to_tool_error(status)
    db_path = store._sqlite._db_path(owner, name)  # type: ignore[attr-defined]
    if not db_path.exists():
        return index_status_to_tool_error(store.inspect_index(owner, name))

    cutoff = (datetime.now(timezone.utc) - timedelta(days=since_days)).strftime(
        "%Y-%m-%dT%H:%M:%SZ"
    )

    # Read-only / immutable connection so the LRU cache isn't evicted by
    # an mtime bump (matches the Phase 2 confidence-probe pattern).
    conn = sqlite3.connect(f"file:{db_path}?mode=ro&immutable=1", uri=True)
    conn.row_factory = sqlite3.Row
    try:
        try:
            base_sql = (
                "SELECT source, pattern, redaction_count, last_redacted "
                "FROM runtime_redaction_log WHERE last_redacted >= ?"
            )
            params: list = [cutoff]
            if source is not None:
                base_sql += " AND source = ?"
                params.append(source)
            base_sql += " ORDER BY redaction_count DESC, source ASC, pattern ASC"
            rows = conn.execute(base_sql, params).fetchall()
        except sqlite3.OperationalError:
            # Pre-Phase-0 DB: no runtime_redaction_log table at all.
            rows = []
    finally:
        conn.close()

    patterns = [
        {
            "source": r["source"],
            "pattern": r["pattern"],
            "count": int(r["redaction_count"] or 0),
            "last_redacted": r["last_redacted"] or "",
        }
        for r in rows
    ]
    sources_seen = sorted({p["source"] for p in patterns})
    total = sum(p["count"] for p in patterns)

    elapsed = (time.perf_counter() - start) * 1000
    return {
        "repo": f"{owner}/{name}",
        "sources": sources_seen,
        "since_iso": cutoff,
        "patterns": patterns,
        "total_redactions": total,
        "_meta": {
            "timing_ms": round(elapsed, 1),
            "filter_source": source or "(all)",
            "since_days": since_days,
            "tip": (
                "Empty patterns list = no redactions recorded in the window. "
                "If you expected hits and don't see any, verify "
                "JCODEMUNCH_RUNTIME_REDACT=1 (default) and that ingests have "
                "actually run. Pair with `get_runtime_coverage` to confirm "
                "data landed in the index at all."
            ),
        },
    }
