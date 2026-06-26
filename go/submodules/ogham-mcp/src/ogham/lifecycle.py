"""Memory lifecycle state machine.

Blended design -- see docs/plans/2026-04-22-consolidation-cascade.md
for provenance (MuninnDB, Shodh, Cortex).

Stages:
    fresh    -- just written, not yet passed the importance gate
    stable   -- normal retrieval weight, hybrid decay
    editing  -- retrieved recently, window open for in-place update

This module is the pure logic. Hooks integration lives in
hooks.py (SessionStart wiring) and service.py (search-triggered window
opens + co-retrieval Hebbian updates).

v0.13.1: stage transitions now go through the backend facade
(`lifecycle_advance_stages`, `lifecycle_close_editing_windows`, etc.)
instead of `backend._execute(...)`. Postgres uses inline SQL via the
facade; Supabase calls the matching RPC defined in migration 035. See
the v0.13 health-dimensions refactor for the precedent.
"""

from __future__ import annotations

import math
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone

from ogham.database import get_backend

DEFAULT_DWELL_HOURS = 1.0
DEFAULT_SURPRISE_GATE = 0.3
DEFAULT_IMPORTANCE_GATE = 0.5
DEFAULT_EDITING_WINDOW_MINUTES = 30


@dataclass
class StageReport:
    fresh_to_stable: int = 0
    editing_closed: int = 0


def advance_stages(profile: str) -> StageReport:
    """Run the sweep for a profile.

    Advances FRESH memories to STABLE when both (a) they've dwelled at
    least DEFAULT_DWELL_HOURS and (b) their surprise or importance
    clears the dual-signal gate. Also closes EDITING windows that have
    expired.

    Updates run against ``memory_lifecycle`` (not ``memories``) so the
    HNSW index on memories is not disturbed. The gates still come from
    memories, via a PK join -- see migration 026 for the rationale.
    """
    backend = get_backend()
    report = StageReport()

    cutoff = (datetime.now(tz=timezone.utc) - timedelta(hours=DEFAULT_DWELL_HOURS)).isoformat()
    report.fresh_to_stable = backend.lifecycle_advance_stages(
        profile=profile,
        cutoff_iso=cutoff,
        surprise_gate=DEFAULT_SURPRISE_GATE,
        importance_gate=DEFAULT_IMPORTANCE_GATE,
    )
    report.editing_closed = close_editing_windows(profile)
    return report


def close_editing_windows(profile: str) -> int:
    """Close EDITING windows older than DEFAULT_EDITING_WINDOW_MINUTES."""
    backend = get_backend()
    cutoff = (
        datetime.now(tz=timezone.utc) - timedelta(minutes=DEFAULT_EDITING_WINDOW_MINUTES)
    ).isoformat()
    return backend.lifecycle_close_editing_windows(profile=profile, cutoff_iso=cutoff)


def open_editing_window(memory_ids: list[str]) -> None:
    """Mark retrieved memories as EDITING with a 30min window.

    Idempotent -- calling twice on the same IDs just resets the window.
    Only STABLE memories can be opened; FRESH stays FRESH (too new to
    edit). Runs against ``memory_lifecycle`` so the HNSW index is not
    disturbed at retrieval time.
    """
    if not memory_ids:
        return
    get_backend().lifecycle_open_editing_window(memory_ids)


def lifecycle_pipeline_counts(profile: str) -> dict[str, int]:
    """Return {stage: count} for the dashboard pipeline card."""
    return get_backend().lifecycle_pipeline_counts(profile)


def hybrid_decay_factor(
    age_days: float, decay_lambda: float = 0.1, decay_beta: float = 0.4
) -> float:
    """Shodh-derived hybrid decay curve.

    0-3 days: exponential decay
    3+ days: power-law
    Used at retrieval time to weight the base relevance score.
    """
    if age_days < 3:
        return math.exp(-decay_lambda * age_days)
    return (age_days / 3) ** (-decay_beta)
