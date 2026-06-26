"""Six-axis health radar + diff helper (todo.md item #5).

Compresses the existing risk-profile signals into a normalized 0–100 score
per axis plus a composite + letter grade. Same shape works as a *state*
snapshot at HEAD or as a *delta* between two snapshots — that's how PR-
time diff-grade reports get computed (run radar on base branch, run
radar on PR branch, diff the two payloads).

The radar pulls inputs from the existing tools (no new heavy work):

| Axis            | Source                                   |
|-----------------|------------------------------------------|
| complexity      | `get_repo_health.avg_complexity`         |
| dead_code       | `get_repo_health.dead_code_pct`          |
| cycles          | `get_repo_health.cycle_count`            |
| coupling        | unstable_modules / total_files           |
| test_gap        | (1 - `get_untested_symbols.reached_pct`) |
| churn_surface   | top-1 hotspot score from `get_hotspots`  |

Higher score = healthier. Penalties are linear and conservative —
calibration tuned so a typical "average" codebase lands around C.

Methodology is auditable via `--explain` on `jcodemunch-mcp claude-md`
or by reading the `_PER_AXIS_SCORERS` table below; PR an updated
formula if you have better calibration data.
"""

from __future__ import annotations

import logging
from typing import Optional

logger = logging.getLogger(__name__)

# Axis identifiers (frozen contract — adding a new axis is a v2 concern).
_AXES: tuple[str, ...] = (
    "complexity",
    "dead_code",
    "cycles",
    "coupling",
    "test_gap",
    "churn_surface",
)


def _clamp(v: float, lo: float = 0.0, hi: float = 100.0) -> float:
    return max(lo, min(hi, v))


def _score_complexity(avg_complexity: float) -> float:
    """avg ≤ 3 -> 100; avg = 10 -> 58; avg ≥ 19.7 -> 0."""
    if avg_complexity <= 3:
        return 100.0
    return _clamp(100.0 - 6.0 * (avg_complexity - 3.0))


def _score_dead_code(dead_code_pct: float) -> float:
    """0% -> 100; 10% -> 60; ≥25% -> 0."""
    return _clamp(100.0 - 4.0 * dead_code_pct)


def _score_cycles(cycle_count: int) -> float:
    """0 -> 100; 5 -> 75; 10 -> 50; ≥20 -> 0."""
    if cycle_count <= 0:
        return 100.0
    return _clamp(100.0 - 5.0 * cycle_count)


def _score_coupling(unstable_modules: int, total_files: int) -> float:
    """0% unstable -> 100; 25% -> 50; ≥50% -> 0."""
    if total_files <= 0:
        return 100.0
    ratio = unstable_modules / total_files
    return _clamp(100.0 - 200.0 * ratio)


def _score_test_gap(untested_pct: float) -> float:
    """0% untested -> 100; 50% -> 50; 100% -> 0."""
    return _clamp(100.0 - untested_pct)


def _score_churn_surface(top_hotspot_score: float) -> float:
    """Bucketed: hotspots are inherently long-tail.

    score ≤ 0       -> 100
    score < 100     -> 80
    score < 500     -> 60
    score < 1000    -> 40
    score < 2000    -> 20
    otherwise       -> 0
    """
    if top_hotspot_score <= 0:
        return 100.0
    if top_hotspot_score < 100:
        return 80.0
    if top_hotspot_score < 500:
        return 60.0
    if top_hotspot_score < 1000:
        return 40.0
    if top_hotspot_score < 2000:
        return 20.0
    return 0.0


def _score_runtime_coverage(coverage_pct: float) -> float:
    """Phase 7 axis: runtime coverage as a healthy-by-default axis.

    Direct linear mapping — 100% of declared call edges have runtime
    evidence over the look-back window → 100. 0% → 0. The axis is only
    meaningful when at least one trace has been ingested; callers that
    don't pass a value get the axis omitted entirely (same convention as
    test_gap / churn_surface) so the composite is still comparable
    against pre-Phase-7 baselines.
    """
    return _clamp(coverage_pct)


def _letter_grade(composite: float) -> str:
    """A/B/C/D/F by 10-point bands. Borderline values round up (a 79.99 is C+, 80.00 is B)."""
    if composite >= 90:
        return "A"
    if composite >= 80:
        return "B"
    if composite >= 70:
        return "C"
    if composite >= 60:
        return "D"
    return "F"


def compute_radar(
    *,
    avg_complexity: float,
    dead_code_pct: float,
    cycle_count: int,
    unstable_modules: int,
    total_files: int,
    untested_pct: Optional[float] = None,
    top_hotspot_score: Optional[float] = None,
    runtime_coverage_pct: Optional[float] = None,
) -> dict:
    """Compute the six- (or seven-) axis radar from raw signal inputs.

    Args:
        avg_complexity: Mean cyclomatic complexity across functions/methods.
        dead_code_pct: Percentage of functions/methods classified as dead.
        cycle_count: Number of dependency cycles in the import graph.
        unstable_modules: Count of production files with instability > 0.7.
            Callers should exclude tests/benchmarks/scripts/examples — those
            files are guaranteed to look unstable (Ca=0) and would dominate
            the metric for any well-tested project.
        total_files: Denominator for the coupling axis. Should match the
            scope of `unstable_modules` (i.e. production-only count, same
            exclusions). `get_repo_health` derives both from
            `_count_unstable_modules`, which returns them as a pair.
        untested_pct: Percentage of functions/methods with no test reachability.
            None => the test_gap axis returns score=None (axis omitted from
            composite). Most callers will pass 100.0 - reached_pct.
        top_hotspot_score: Top-1 hotspot score (complexity × log(1 + churn)).
            None => churn_surface axis returns score=None.
        runtime_coverage_pct: Phase 7 — percentage of declared call edges
            with runtime evidence over the window. None => axis omitted
            (preserves bit-for-bit comparability with pre-Phase-7
            baselines on repos that haven't ingested traces). Most
            callers will pass the value computed by
            ``get_runtime_coverage``.

    Returns:
        ``{axes: {axis: {score, raw}}, composite, grade, omitted_axes}``
    """
    axes: dict[str, dict] = {
        "complexity": {
            "score": _score_complexity(avg_complexity),
            "raw": avg_complexity,
        },
        "dead_code": {
            "score": _score_dead_code(dead_code_pct),
            "raw": dead_code_pct,
        },
        "cycles": {
            "score": _score_cycles(cycle_count),
            "raw": cycle_count,
        },
        "coupling": {
            "score": _score_coupling(unstable_modules, total_files),
            "raw_unstable": unstable_modules,
            "raw_total_files": total_files,
        },
    }

    omitted: list[str] = []
    if untested_pct is not None:
        axes["test_gap"] = {
            "score": _score_test_gap(untested_pct),
            "raw": untested_pct,
        }
    else:
        omitted.append("test_gap")

    if top_hotspot_score is not None:
        axes["churn_surface"] = {
            "score": _score_churn_surface(top_hotspot_score),
            "raw": top_hotspot_score,
        }
    else:
        omitted.append("churn_surface")

    if runtime_coverage_pct is not None:
        axes["runtime_coverage"] = {
            "score": _score_runtime_coverage(runtime_coverage_pct),
            "raw": runtime_coverage_pct,
        }
    else:
        omitted.append("runtime_coverage")

    scored_values = [a["score"] for a in axes.values()]
    composite = round(sum(scored_values) / len(scored_values), 1) if scored_values else 0.0

    return {
        "axes": axes,
        "composite": composite,
        "grade": _letter_grade(composite),
        "omitted_axes": omitted,
    }


def diff_radar(baseline: dict, current: dict) -> dict:
    """Compute axis-by-axis deltas between two radar payloads.

    Args:
        baseline: A radar payload (e.g. from base branch).
        current: A radar payload (e.g. from PR branch).

    Returns:
        ``{axis_deltas, composite_delta, grade_change, regressions, improvements}``

        ``axis_deltas[axis]``: ``{from, to, delta, raw_from, raw_to}``
        ``regressions``: axes where the score dropped by ≥3 points (config
            via the threshold constant; small fluctuations don't count).
        ``improvements``: axes where the score rose by ≥3 points.
    """
    threshold = 3.0  # points; smaller deltas are noise
    out_axes: dict[str, dict] = {}
    regressions: list[str] = []
    improvements: list[str] = []

    base_axes = baseline.get("axes", {})
    cur_axes = current.get("axes", {})
    all_axis_names = sorted(set(base_axes.keys()) | set(cur_axes.keys()))

    for axis in all_axis_names:
        b = base_axes.get(axis, {}) or {}
        c = cur_axes.get(axis, {}) or {}
        b_score = b.get("score")
        c_score = c.get("score")
        if b_score is None or c_score is None:
            out_axes[axis] = {
                "from": b_score,
                "to": c_score,
                "delta": None,
                "note": "axis missing from one side",
            }
            continue
        delta = round(c_score - b_score, 1)
        out_axes[axis] = {
            "from": b_score,
            "to": c_score,
            "delta": delta,
            "raw_from": b.get("raw"),
            "raw_to": c.get("raw"),
        }
        if delta <= -threshold:
            regressions.append(axis)
        elif delta >= threshold:
            improvements.append(axis)

    base_composite = baseline.get("composite", 0.0)
    cur_composite = current.get("composite", 0.0)
    composite_delta = round(cur_composite - base_composite, 1)

    base_grade = baseline.get("grade", "?")
    cur_grade = current.get("grade", "?")
    grade_change = (
        f"{base_grade} -> {cur_grade}" if base_grade != cur_grade else f"{cur_grade} (unchanged)"
    )

    return {
        "axis_deltas": out_axes,
        "composite_from": base_composite,
        "composite_to": cur_composite,
        "composite_delta": composite_delta,
        "grade_change": grade_change,
        "regressions": regressions,
        "improvements": improvements,
        "verdict": _verdict(composite_delta, regressions, improvements),
    }


def _verdict(composite_delta: float, regressions: list[str], improvements: list[str]) -> str:
    """One-line summary used in PR comments / CI output."""
    if abs(composite_delta) < 1.0 and not regressions and not improvements:
        return "no meaningful change"
    if regressions and not improvements:
        return f"REGRESSION on {len(regressions)} axis/axes (composite {composite_delta:+.1f})"
    if improvements and not regressions:
        return f"improvement on {len(improvements)} axis/axes (composite {composite_delta:+.1f})"
    if regressions and improvements:
        return f"mixed: -{len(regressions)} / +{len(improvements)} axes (composite {composite_delta:+.1f})"
    return f"composite {composite_delta:+.1f}"


def diff_health_radar(baseline: dict, current: dict) -> dict:
    """MCP tool entry point: takes two radar payloads, returns the diff.

    Both ``baseline`` and ``current`` should be radar payloads as
    produced by ``compute_radar`` — typically the ``radar`` field on a
    ``get_repo_health`` response. Pass them as JSON strings or dicts;
    the MCP server will deserialize JSON inputs before calling this.

    The radar payloads can come from anywhere — saved to disk, posted
    via a CI job, fetched from a release artifact, or computed in two
    consecutive ``get_repo_health`` calls between branch checkouts. This
    function is pure: no I/O, no index access, no git.
    """
    if not isinstance(baseline, dict) or not isinstance(current, dict):
        return {
            "error": (
                "diff_health_radar requires two radar payload dicts. "
                "Pass the `radar` field from get_repo_health responses."
            )
        }
    if "axes" not in baseline or "axes" not in current:
        return {
            "error": (
                "Both inputs must be radar payloads (need an `axes` field). "
                "Did you pass the full get_repo_health response instead of its `radar` sub-field?"
            )
        }
    return diff_radar(baseline, current)
