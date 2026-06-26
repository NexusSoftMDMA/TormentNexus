"""analyze_perf — surface tool latency and cache-hit telemetry.

Reads in-memory latency rings (always populated when call_tool fires) and,
if enabled, persisted rows from telemetry.db. No-op safe when no calls have
been recorded yet.

Optional ``compare_release`` parameter loads a baseline snapshot from
``benchmarks/token_baselines/v{X}.json`` (created by
``capture_token_baseline.py``) and reports per-tool deltas in tokens_saved
and latency vs the current session.
"""

from __future__ import annotations

import json
import time
from pathlib import Path
from typing import Optional

from ..storage import token_tracker as _tt


_DEFAULT_TOP = 20


def _baseline_path(version: str) -> Path:
    """Resolve ``benchmarks/token_baselines/v{version}.json`` from repo root."""
    here = Path(__file__).resolve()
    repo_root = here.parents[3]  # tools/.. /jcodemunch_mcp/.. /src/.. /<root>
    # Walk up until we find a sibling 'benchmarks' dir (works in both
    # editable installs and a checked-out clone).
    for ancestor in [repo_root, *repo_root.parents]:
        candidate = ancestor / "benchmarks" / "token_baselines" / f"v{version}.json"
        if candidate.exists():
            return candidate
    return repo_root / "benchmarks" / "token_baselines" / f"v{version}.json"


def _diff_baseline(
    baseline: dict,
    current_latency: dict,
    current_breakdown: dict,
) -> dict:
    """Compute per-tool deltas between baseline snapshot and live session."""
    out: dict = {}
    base_tools = baseline.get("tools", {})
    all_tools = set(base_tools) | set(current_latency) | set(current_breakdown)
    for tool in sorted(all_tools):
        b = base_tools.get(tool, {})
        cur_lat = current_latency.get(tool, {})
        cur_tokens = int(current_breakdown.get(tool, 0))
        out[tool] = {
            "tokens_saved_delta": cur_tokens - int(b.get("tokens_saved", 0)),
            "p50_delta_ms": round(float(cur_lat.get("p50_ms", 0.0)) - float(b.get("p50_ms", 0.0)), 2),
            "p95_delta_ms": round(float(cur_lat.get("p95_ms", 0.0)) - float(b.get("p95_ms", 0.0)), 2),
            "calls_delta": int(cur_lat.get("count", 0)) - int(b.get("calls", 0)),
        }
    return out


def _percentile(sorted_vals: list[float], pct: float) -> float:
    if not sorted_vals:
        return 0.0
    idx = max(0, min(len(sorted_vals) - 1, int(pct * len(sorted_vals))))
    return sorted_vals[idx]


def _ledger_summary(rows: list[tuple], top: int) -> dict:
    """Aggregate ranking_events rows by repo and by tool."""
    by_repo: dict = {}
    by_tool: dict = {}
    for ts, repo, tool, qh, query, returned_ids, top1, top2, conf, sem, ident, stale in rows:
        rb = by_repo.setdefault(repo or "<no-repo>", {
            "events": 0,
            "avg_confidence": 0.0,
            "_conf_total": 0.0,
            "_conf_count": 0,
            "stale_events": 0,
            "identity_hits": 0,
            "semantic_used": 0,
        })
        rb["events"] += 1
        if conf is not None:
            rb["_conf_total"] += float(conf)
            rb["_conf_count"] += 1
        if stale:
            rb["stale_events"] += 1
        if ident:
            rb["identity_hits"] += 1
        if sem:
            rb["semantic_used"] += 1
        tb = by_tool.setdefault(tool, {"events": 0})
        tb["events"] += 1
    for repo_name, rb in by_repo.items():
        ct = rb.pop("_conf_count", 0)
        total = rb.pop("_conf_total", 0.0)
        rb["avg_confidence"] = round(total / ct, 3) if ct else 0.0
    repo_ranked = sorted(by_repo.items(), key=lambda kv: kv[1]["events"], reverse=True)[:top]
    tool_ranked = sorted(by_tool.items(), key=lambda kv: kv[1]["events"], reverse=True)
    return {
        "total_events": len(rows),
        "by_repo": [{"repo": r, **stats} for r, stats in repo_ranked],
        "by_tool": [{"tool": t, **stats} for t, stats in tool_ranked],
    }


def analyze_perf(
    window: str = "session",
    top: int = _DEFAULT_TOP,
    tool: Optional[str] = None,
    storage_path: Optional[str] = None,
    compare_release: Optional[str] = None,
    ledger: bool = False,
) -> dict:
    """Return per-tool latency + cache-hit telemetry for the current session
    (and the persisted perf db if perf_telemetry_enabled is set).

    Args:
        window: ``session`` (in-memory ring), ``1h``, ``24h``, ``7d``, or ``all``.
                Anything other than ``session`` reads the perf SQLite db.
        top:    Cap on how many slowest tools to return (default 20).
        tool:   Restrict the analysis to a single tool name.
        storage_path: Optional override for the index storage root.
    """
    t0 = time.perf_counter()

    cache_stats = _tt.result_cache_stats()
    in_memory = _tt.latency_stats()
    if tool:
        in_memory = {k: v for k, v in in_memory.items() if k == tool}

    persisted: dict = {}
    persisted_meta: dict = {"source": "in_memory_only", "rows": 0}
    if window != "session":
        seconds_map = {
            "1h": 3600.0,
            "24h": 86_400.0,
            "7d": 7 * 86_400.0,
            "all": None,
        }
        if window not in seconds_map:
            return {
                "error": (
                    f"Invalid window {window!r}. Use one of: session, 1h, 24h, 7d, all."
                )
            }
        rows = _tt.perf_db_query(
            base_path=storage_path,
            window_seconds=seconds_map[window],
            tool=tool,
        )
        persisted_meta = {"source": "telemetry.db", "rows": len(rows), "window": window}
        # Aggregate by tool
        by_tool: dict[str, list[float]] = {}
        errors: dict[str, int] = {}
        for ts, t_name, dur, ok, _repo in rows:
            by_tool.setdefault(t_name, []).append(float(dur))
            if not ok:
                errors[t_name] = errors.get(t_name, 0) + 1
        for t_name, durs in by_tool.items():
            durs.sort()
            n = len(durs)
            persisted[t_name] = {
                "count": n,
                "p50_ms": round(_percentile(durs, 0.5), 2),
                "p95_ms": round(_percentile(durs, 0.95), 2),
                "max_ms": round(durs[-1], 2),
                "errors": errors.get(t_name, 0),
                "error_rate": round(errors.get(t_name, 0) / n, 3) if n else 0.0,
            }
        if not _tt._state and persisted_meta["rows"] == 0:  # type: ignore[attr-defined]
            persisted_meta["note"] = (
                "No persisted rows. Set config 'perf_telemetry_enabled': true "
                "or env JCODEMUNCH_PERF_TELEMETRY=1 to enable the SQLite sink."
            )

    # Pick the dataset to rank
    ranked_source = persisted if window != "session" else in_memory
    slowest = sorted(
        ranked_source.items(),
        key=lambda kv: kv[1].get("p95_ms", 0.0),
        reverse=True,
    )[:top]

    # Cache hit-rate ranked low → high (low rates point to cold caches)
    by_tool_cache = cache_stats.get("by_tool", {})
    coldest_caches = sorted(
        by_tool_cache.items(),
        key=lambda kv: kv[1].get("hit_rate", 0.0),
    )[:top]

    baseline_diff: Optional[dict] = None
    baseline_meta: Optional[dict] = None
    if compare_release:
        baseline_path = _baseline_path(compare_release)
        if not baseline_path.exists():
            baseline_meta = {
                "version": compare_release,
                "found": False,
                "looked_at": str(baseline_path),
            }
        else:
            try:
                baseline = json.loads(baseline_path.read_text())
                breakdown = _tt.get_session_stats(base_path=storage_path).get(
                    "tool_breakdown", {}
                )
                baseline_diff = _diff_baseline(baseline, in_memory, breakdown)
                baseline_meta = {
                    "version": baseline.get("version", compare_release),
                    "captured_at": baseline.get("captured_at"),
                    "found": True,
                    "tools_in_baseline": len(baseline.get("tools", {})),
                }
            except Exception as exc:
                baseline_meta = {
                    "version": compare_release,
                    "found": True,
                    "error": f"Failed to parse baseline: {type(exc).__name__}: {exc}",
                }

    elapsed_ms = round((time.perf_counter() - t0) * 1000, 2)
    out = {
        "window": window,
        "tool": tool,
        "in_memory_session": in_memory,
        "persisted": persisted,
        "persisted_meta": persisted_meta,
        "slowest_by_p95": [
            {"tool": name, **stats} for name, stats in slowest
        ],
        "cache": {
            "totals": {
                "hits": cache_stats.get("total_hits", 0),
                "misses": cache_stats.get("total_misses", 0),
                "hit_rate": cache_stats.get("hit_rate", 0.0),
                "cached_entries": cache_stats.get("cached_entries", 0),
            },
            "coldest_by_tool": [
                {"tool": name, **stats} for name, stats in coldest_caches
            ],
        },
        "_meta": {"timing_ms": elapsed_ms},
    }
    if baseline_meta is not None:
        out["baseline_meta"] = baseline_meta
    if baseline_diff is not None:
        out["baseline_diff"] = baseline_diff

    if ledger:
        seconds_map_l = {"1h": 3600.0, "24h": 86_400.0, "7d": 7 * 86_400.0}
        window_seconds = seconds_map_l.get(window)  # None for session/all
        rows = _tt.ranking_db_query(
            base_path=storage_path,
            window_seconds=window_seconds,
            tool=tool,
            limit=10_000,
        )
        out["ranking_ledger"] = _ledger_summary(rows, top=top)

    return out
