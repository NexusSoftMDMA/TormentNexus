"""Retrieval-regret extraction over the ranking_events ledger.

We already collect rich retrieval telemetry but feed it to a single consumer
(``WeightTuner``, which nudges two ranking knobs). The same ledger carries a
louder, unread signal: **when retrieval failed and the agent had to re-ask.**
This module mines that regret.

Pure read over the existing ``ranking_events`` ledger via
``token_tracker.ranking_db_query`` — no new tables, no writes. The output is a
list of regret *clusters*; correction synthesis (``tools/suggest_corrections``)
turns clusters into suggested (never applied) config patches.

Ledger tuple layout (matches the SELECT in ``token_tracker.ranking_db_query``):
    0 ts            5 returned_ids (JSON)   10 identity_hit
    1 repo          6 top1_score            11 repo_is_stale
    2 tool          7 top2_score
    3 query_hash    8 confidence
    4 query         9 semantic_used
"""

from __future__ import annotations

import json as _json
from collections import defaultdict
from typing import Any, Optional

from ..storage import token_tracker as _tt
from .. import config as _config

# Column indices into a ranking_events row tuple.
_TS, _REPO, _TOOL, _QH, _QUERY, _RETURNED = 0, 1, 2, 3, 4, 5
_TOP1, _TOP2, _CONF, _SEM, _IDHIT, _STALE = 6, 7, 8, 9, 10, 11

# --- Thresholds (starting points; conservative to avoid noisy suggestions) --- #
DEFAULT_WINDOW_DAYS = 30
REQUERY_LIFETIME = 5          # same query_hash this many times => churn
REQUERY_HIGH = 8             # ... this many => high severity
LOW_CONF = 0.30              # confidence below this on a non-empty result
LOW_CONF_RECUR = 2          # low-confidence events for one query to cluster
THIN_TOP1_FLOOR = 0.10      # top1 below this with <=1 result == thin
AMBIGUOUS_GAP = 0.05        # top1 - top2 below this == couldn't disambiguate
AMBIGUOUS_RECUR = 2
STALE_RATE = 0.20           # >20% of events stale-at-query == freshness problem
STALE_MIN_EVENTS = 5        # ... but only judge the rate over enough events
VOCAB_CONF_FLOOR = 0.30     # identity miss rescued by semantic with >= this conf
VOCAB_RECUR = 2
MAX_EXAMPLES = 3            # example queries carried per cluster
MAX_CLUSTERS_PER_SIGNAL = 5


def _sev(count: int, hi: int, med: int) -> str:
    if count >= hi:
        return "high"
    if count >= med:
        return "medium"
    return "low"


def _decode_ids(raw: Any) -> list:
    if not raw:
        return []
    try:
        v = _json.loads(raw)
        return v if isinstance(v, list) else []
    except (ValueError, TypeError):
        return []


def _examples(rows: list[tuple]) -> list[str]:
    """Distinct example query strings for a cluster, capped."""
    seen: list[str] = []
    for r in rows:
        q = r[_QUERY]
        if q and q not in seen:
            seen.append(q)
        if len(seen) >= MAX_EXAMPLES:
            break
    return seen


def _by_query_hash(events: list[tuple]) -> "dict[str, list[tuple]]":
    groups: dict[str, list[tuple]] = defaultdict(list)
    for e in events:
        groups[e[_QH]].append(e)
    return groups


def _cluster(signal: str, severity: str, rows: list[tuple], **evidence) -> dict:
    tools = sorted({r[_TOOL] for r in rows if r[_TOOL]})
    return {
        "signal": signal,
        "severity": severity,
        "event_count": len(rows),
        "tools": tools,
        "query_examples": _examples(rows),
        "evidence": evidence,
    }


# --- Per-signal detectors --------------------------------------------------- #

def _detect_requery_churn(by_qh: "dict[str, list[tuple]]") -> list[dict]:
    out = []
    for qh, rows in by_qh.items():
        if len(rows) >= REQUERY_LIFETIME:
            out.append(_cluster(
                "requery_churn", _sev(len(rows), REQUERY_HIGH, REQUERY_LIFETIME),
                rows, query_hash=qh, repeats=len(rows),
            ))
    out.sort(key=lambda c: -c["event_count"])
    return out[:MAX_CLUSTERS_PER_SIGNAL]


def _detect_low_confidence(by_qh: "dict[str, list[tuple]]") -> list[dict]:
    out = []
    for qh, rows in by_qh.items():
        hits = [r for r in rows
                if r[_CONF] is not None and r[_CONF] < LOW_CONF and _decode_ids(r[_RETURNED])]
        if len(hits) >= LOW_CONF_RECUR:
            avg = sum(r[_CONF] for r in hits) / len(hits)
            out.append(_cluster(
                "low_confidence", _sev(len(hits), LOW_CONF_RECUR * 3, LOW_CONF_RECUR),
                hits, query_hash=qh, avg_confidence=round(avg, 3),
            ))
    out.sort(key=lambda c: (-c["event_count"], c["evidence"]["avg_confidence"]))
    return out[:MAX_CLUSTERS_PER_SIGNAL]


def _detect_thin_result(by_qh: "dict[str, list[tuple]]") -> list[dict]:
    out = []
    for qh, rows in by_qh.items():
        hits = []
        for r in rows:
            ids = _decode_ids(r[_RETURNED])
            top1 = r[_TOP1]
            if not ids or (len(ids) <= 1 and (top1 is None or top1 < THIN_TOP1_FLOOR)):
                hits.append(r)
        if len(hits) >= 2:
            out.append(_cluster(
                "thin_result", _sev(len(hits), 5, 2),
                hits, query_hash=qh, empty_or_weak=len(hits),
            ))
    out.sort(key=lambda c: -c["event_count"])
    return out[:MAX_CLUSTERS_PER_SIGNAL]


def _detect_ambiguous_top(by_qh: "dict[str, list[tuple]]") -> list[dict]:
    out = []
    for qh, rows in by_qh.items():
        hits = [r for r in rows
                if r[_TOP1] is not None and r[_TOP2] is not None
                and (r[_TOP1] - r[_TOP2]) < AMBIGUOUS_GAP]
        if len(hits) >= AMBIGUOUS_RECUR:
            out.append(_cluster(
                "ambiguous_top", _sev(len(hits), AMBIGUOUS_RECUR * 3, AMBIGUOUS_RECUR),
                hits, query_hash=qh, min_gap=round(
                    min(r[_TOP1] - r[_TOP2] for r in hits), 4),
            ))
    out.sort(key=lambda c: -c["event_count"])
    return out[:MAX_CLUSTERS_PER_SIGNAL]


def _detect_stale_at_query(events: list[tuple]) -> list[dict]:
    if len(events) < STALE_MIN_EVENTS:
        return []
    stale = [e for e in events if e[_STALE]]
    rate = len(stale) / len(events)
    if rate > STALE_RATE:
        return [_cluster(
            "stale_at_query",
            "high" if rate > STALE_RATE * 2 else "medium",
            stale, stale_rate=round(rate, 3), stale_events=len(stale),
            total_events=len(events),
        )]
    return []


def _detect_vocabulary_gap(by_qh: "dict[str, list[tuple]]") -> list[dict]:
    """Identity miss rescued by semantic search => the agent's term doesn't
    match a symbol name but means one. The strongest novelty signal."""
    out = []
    for qh, rows in by_qh.items():
        hits = [r for r in rows
                if not r[_IDHIT] and r[_SEM]
                and r[_CONF] is not None and r[_CONF] >= VOCAB_CONF_FLOOR]
        if len(hits) >= VOCAB_RECUR:
            out.append(_cluster(
                "vocabulary_gap", _sev(len(hits), VOCAB_RECUR * 3, VOCAB_RECUR),
                hits, query_hash=qh, identity_misses=len(hits),
            ))
    out.sort(key=lambda c: -c["event_count"])
    return out[:MAX_CLUSTERS_PER_SIGNAL]


def analyze_regret(
    repo: str,
    *,
    window_days: int = DEFAULT_WINDOW_DAYS,
    storage_path: Optional[str] = None,
    all_time: bool = False,
) -> dict:
    """Mine the ranking_events ledger for retrieval regret for ``repo``.

    Returns a dict with ``telemetry_present``, ``events_analyzed``, and
    ``clusters`` (a flat list of regret clusters across all six signals,
    severity-ranked). Honest no-telemetry / no-events shapes are returned
    rather than fabricated regret. Pure read — never writes.
    """
    telemetry_on = bool(_config.get("perf_telemetry_enabled", False))
    window = None if all_time else float(window_days) * 86_400
    events = _tt.ranking_db_query(
        base_path=storage_path, repo=repo, window_seconds=window, limit=10_000,
    )

    base = {
        "repo": repo,
        "telemetry_present": telemetry_on,
        "window_days": None if all_time else window_days,
        "events_analyzed": len(events),
    }
    if not events:
        base["clusters"] = []
        base["hint"] = (
            "No ranking telemetry for this repo. Enable it with "
            "`perf_telemetry_enabled: true` (or JCODEMUNCH_PERF_TELEMETRY=1) and "
            "run some searches; regret analysis needs a ledger to read."
            if not telemetry_on else
            "Telemetry is on but no ranking events recorded for this repo yet in "
            "the window. Run some searches, or widen the window with all_time."
        )
        return base

    by_qh = _by_query_hash(events)
    clusters: list[dict] = []
    clusters += _detect_requery_churn(by_qh)
    clusters += _detect_low_confidence(by_qh)
    clusters += _detect_thin_result(by_qh)
    clusters += _detect_ambiguous_top(by_qh)
    clusters += _detect_stale_at_query(events)
    clusters += _detect_vocabulary_gap(by_qh)

    _rank = {"high": 0, "medium": 1, "low": 2}
    clusters.sort(key=lambda c: (_rank.get(c["severity"], 9), -c["event_count"]))
    base["clusters"] = clusters
    return base
