"""find_unused_paths — symbols reachable on paper but never executed (Phase 3 + 4).

Distinct from ``find_dead_code`` (static-only graph reachability) — this
tool only flags code that has *runtime evidence of absence*: zero hits
in ``runtime_calls`` over the configured window. A symbol with no
runtime hits but zero static callers is dead by both definitions; a
symbol with no runtime hits but plenty of static callers is the
interesting "looks reachable, never runs" finding only this tool can
surface.

**Phase 4 addition.** When the index has dbt-style column metadata
(``dbt_columns`` / ``sqlmesh_columns`` in context_metadata) and the
``runtime_columns`` table has at least one row, results gain a
``unused_columns`` list per dbt-model symbol — declared columns that
never appeared in any ingested SQL log. The model itself only counts
as ``unused`` when (a) the model symbol has zero hits in runtime_calls
AND (b) every declared column also has zero hits in runtime_columns,
since a model used solely for one column is still load-bearing.

Excludes test files and entry-point heuristics by default — unused tests
aren't "dead" and main/__init__/wsgi/etc. are entry points the runtime
trace probably doesn't capture.

Returns:
  ``{
      'repo': 'owner/name',
      'since_days': D,
      'cutoff_iso': cutoff date for "recent enough" runtime evidence,
      'results': [
          {
              'symbol_id', 'name', 'kind', 'file', 'line',
              'last_seen': '' if never observed,
              'reason': 'no_runtime_evidence' | 'stale_only' | 'dbt_model_no_column_reads',
              'unused_columns'?: [<col>, ...],   # Phase 4: dbt-only
          },
          ...
      ],
      'total_unused': N,
      '_meta': {timing_ms, total_symbols_scanned, excluded_test_files,
                excluded_entry_points, runtime_data_present,
                runtime_columns_present, ...}
  }``

When no traces have been ingested at all, ``results`` is empty and
``_meta.runtime_data_present`` is False — every symbol would be
trivially "unused" otherwise, which would mislead the agent.
"""

from __future__ import annotations

import json
import sqlite3
import time
from datetime import datetime, timedelta, timezone
from typing import Optional

from ._utils import index_status_to_tool_error, resolve_repo
from .find_dead_code import _is_test_file, _is_entry_point_filename
from ..storage import IndexStore


def _load_declared_columns(conn: sqlite3.Connection) -> dict[str, set[str]]:
    """Return ``{model_name: {col_name, ...}}`` from any ``*_columns``
    provider in context_metadata. Empty dict when no dbt/SQLMesh metadata
    exists. Used by the Phase 4 dbt-aware code path."""
    row = conn.execute(
        "SELECT value FROM meta WHERE key = 'context_metadata'"
    ).fetchone()
    if not row or not row[0]:
        return {}
    try:
        ctx = json.loads(row[0])
    except (TypeError, ValueError):
        return {}
    out: dict[str, set[str]] = {}
    if not isinstance(ctx, dict):
        return out
    for key, value in ctx.items():
        if not key.endswith("_columns") or not isinstance(value, dict):
            continue
        for model_name, cols in value.items():
            if not isinstance(cols, dict):
                continue
            out.setdefault(model_name, set()).update(cols.keys())
    return out


def _load_observed_columns(conn: sqlite3.Connection, cutoff: str) -> dict[str, set[str]]:
    """Return ``{model_name: {col_name, ...}}`` from ``runtime_columns``,
    filtered to rows last seen at or after ``cutoff``."""
    out: dict[str, set[str]] = {}
    for r in conn.execute(
        "SELECT model_name, column_name FROM runtime_columns WHERE last_seen >= ?",
        (cutoff,),
    ):
        if r[0]:
            out.setdefault(r[0], set()).add(r[1])
    return out


def find_unused_paths(
    repo: str,
    since_days: int = 90,
    *,
    include_tests: bool = False,
    include_entry_points: bool = False,
    max_results: int = 200,
    storage_path: Optional[str] = None,
) -> dict:
    """Return symbols with zero (or stale) runtime hits within the window.

    Args:
        repo: Repository identifier.
        since_days: Look-back window. ``>=1``. Symbols last observed
            before ``now - since_days`` days surface as ``stale_only``.
            Symbols never observed surface as ``no_runtime_evidence``.
        include_tests: Include symbols in test files. Default False.
        include_entry_points: Include symbols in entry-point filenames
            (``main.py``, ``__main__.py``, ``wsgi.py``, ``app.py``,
            ``manage.py``, etc.). Default False.
        max_results: Cap on returned rows.
        storage_path: Custom storage path.

    Returns:
        See module docstring.
    """
    start = time.perf_counter()
    since_days = max(1, since_days)
    max_results = max(1, min(max_results, 1000))
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

    cutoff = (datetime.now(timezone.utc) - timedelta(days=since_days)).strftime("%Y-%m-%dT%H:%M:%SZ")

    conn = sqlite3.connect(f"file:{db_path}?mode=ro&immutable=1", uri=True)
    conn.row_factory = sqlite3.Row
    try:
        # If runtime_calls is empty, every symbol would trivially qualify —
        # so refuse to dump the entire symbol set and return runtime_data_present=False.
        runtime_present = (
            conn.execute("SELECT 1 FROM runtime_calls LIMIT 1").fetchone() is not None
        )
        # Phase 4: also detect runtime_columns presence — used for the dbt-aware
        # path. The table is only consulted when at least one row exists.
        try:
            columns_present = (
                conn.execute("SELECT 1 FROM runtime_columns LIMIT 1").fetchone() is not None
            )
        except sqlite3.OperationalError:
            # Index is on a pre-v15 schema (table doesn't exist yet)
            columns_present = False
        if not runtime_present:
            return {
                "repo": f"{owner}/{name}",
                "since_days": since_days,
                "cutoff_iso": cutoff,
                "results": [],
                "total_unused": 0,
                "_meta": {
                    "timing_ms": round((time.perf_counter() - start) * 1000, 1),
                    "runtime_data_present": False,
                    "tip": (
                        "No traces ingested yet. find_unused_paths only fires once at "
                        "least one runtime signal exists; otherwise every symbol would "
                        "be 'unused' and the result would be useless. "
                        "Run `import_runtime_signal` first."
                    ),
                },
            }

        # Symbols never observed — left join + IS NULL is the canonical pattern.
        # We also surface "observed once but not within the window".
        # Note: SQLite's HAVING does not reliably match a COALESCE'd empty
        # string against a literal '', so we keep the raw NULL from MAX()
        # and use IS NULL to catch never-observed symbols. The stale
        # case (observed before cutoff) is the second OR branch.
        rows = conn.execute(
            """
            SELECT
                s.id    AS symbol_id,
                s.name  AS name,
                s.kind  AS kind,
                s.file  AS file,
                s.line  AS line,
                MAX(rc.last_seen) AS last_seen_raw
            FROM symbols s
            LEFT JOIN runtime_calls rc ON rc.symbol_id = s.id
            GROUP BY s.id
            HAVING last_seen_raw IS NULL OR last_seen_raw < ?
            ORDER BY (last_seen_raw IS NOT NULL), last_seen_raw ASC, s.file ASC, s.line ASC
            """,
            (cutoff,),
        ).fetchall()

        # Phase 4: when both dbt-style metadata and runtime_columns data exist,
        # attach unused_columns per .sql model symbol. A model with at least
        # one column observed is *not* unused even if the model symbol itself
        # never appeared in runtime_calls (some pipelines log per-column reads
        # without a wrapping symbol-level signal).
        declared_cols = _load_declared_columns(conn) if columns_present else {}
        observed_cols = _load_observed_columns(conn, cutoff) if columns_present else {}

        excluded_tests = 0
        excluded_entry_points = 0
        rescued_by_column_hit = 0
        results: list[dict] = []
        for r in rows:
            file_path = r["file"] or ""
            if not include_tests and _is_test_file(file_path):
                excluded_tests += 1
                continue
            if not include_entry_points and _is_entry_point_filename(file_path):
                excluded_entry_points += 1
                continue
            last_seen_raw = r["last_seen_raw"]
            sym_name = r["name"] or ""
            sym_kind = r["kind"] or ""

            # Phase 4 dbt-aware rescue: a SQL-file symbol with column reads
            # is *not* unused even if no symbol-level runtime hit landed.
            unused_columns_list: Optional[list[str]] = None
            if (
                columns_present
                and file_path.endswith(".sql")
                and sym_name in declared_cols
            ):
                declared = declared_cols[sym_name]
                observed = observed_cols.get(sym_name, set())
                if observed:
                    rescued_by_column_hit += 1
                    continue  # at least one column was read — model is in use
                # No column reads at all — surface every declared column as unused.
                unused_columns_list = sorted(declared)

            if unused_columns_list:
                reason = "dbt_model_no_column_reads"
            elif last_seen_raw is None:
                reason = "no_runtime_evidence"
            else:
                reason = "stale_only"

            entry: dict = {
                "symbol_id": r["symbol_id"],
                "name": sym_name,
                "kind": sym_kind,
                "file": file_path,
                "line": r["line"],
                "last_seen": last_seen_raw or "",
                "reason": reason,
            }
            if unused_columns_list is not None:
                entry["unused_columns"] = unused_columns_list
            results.append(entry)
            if len(results) >= max_results:
                break

        # Total scanned for the _meta breakdown
        total_scanned = conn.execute("SELECT COUNT(*) AS n FROM symbols").fetchone()["n"]
    finally:
        conn.close()

    elapsed = (time.perf_counter() - start) * 1000
    return {
        "repo": f"{owner}/{name}",
        "since_days": since_days,
        "cutoff_iso": cutoff,
        "results": results,
        "total_unused": len(results),
        "_meta": {
            "timing_ms": round(elapsed, 1),
            "total_symbols_scanned": total_scanned,
            "excluded_test_files": excluded_tests,
            "excluded_entry_points": excluded_entry_points,
            "rescued_by_column_hit": rescued_by_column_hit,
            "runtime_data_present": True,
            "runtime_columns_present": columns_present,
            "truncated": len(results) >= max_results,
            "tip": (
                "no_runtime_evidence = never observed in any ingested trace. "
                "stale_only = observed before --since-days cutoff. "
                "dbt_model_no_column_reads = SQL model whose declared columns "
                "have zero hits in runtime_columns over the window — the "
                "data-layer counterpart of dead code. "
                "Pair with find_dead_code for the static-graph view: symbols here "
                "AND in find_dead_code are dead by both definitions; symbols here "
                "BUT NOT in find_dead_code are reachable on paper but never run — "
                "the most interesting category."
            ),
        },
    }
