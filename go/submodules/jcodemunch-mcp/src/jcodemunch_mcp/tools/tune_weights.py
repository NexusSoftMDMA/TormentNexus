"""tune_weights — learn per-repo retrieval weights from the ranking ledger.

Reads the v1.78.0 ``ranking_events`` table populated by
``perf_telemetry_enabled``, computes per-signal confidence correlations,
and writes per-repo overrides to ``~/.code-index/tuning.jsonc``.

When ``repo`` is omitted, learns for every repo present in the ledger.
``dry_run=True`` proposes deltas without writing the file.
"""

from __future__ import annotations

import time
from typing import Optional

from ..retrieval import tuning as _tuning


def tune_weights(
    repo: Optional[str] = None,
    dry_run: bool = False,
    min_events: int = 50,
    explain: bool = False,
    max_age_days: int = 90,
    storage_path: Optional[str] = None,
) -> dict:
    """Run a tuning pass over the ranking ledger.

    Args:
        repo:         If set, learn for one repo. Otherwise, every repo
                      present in the ledger.
        dry_run:      Compute proposals without writing tuning.jsonc.
        min_events:   Skip repos with fewer ledger events than this.
        explain:      Include the per-signal correlations used for the
                      proposal in the output.
        max_age_days: Only learn from ledger events newer than this.
                      Guards against stale events anchoring weights to
                      an old query distribution. 0 = lifetime ledger.
        storage_path: Optional override for ~/.code-index.
    """
    t0 = time.perf_counter()
    tuner = _tuning.WeightTuner(base_path=storage_path)

    targets: list[str]
    if repo:
        targets = [repo]
    else:
        targets = _tuning.list_repos(base_path=storage_path)

    results: list[dict] = []
    for target in targets:
        outcome = tuner.learn(
            target, dry_run=dry_run, min_events=min_events, max_age_days=max_age_days
        )
        if not explain:
            outcome.pop("signals", None)
        results.append(outcome)

    applied = sum(1 for r in results if r.get("applied"))
    elapsed_ms = round((time.perf_counter() - t0) * 1000, 2)
    return {
        "results": results,
        "summary": {
            "repos_examined": len(results),
            "applied": applied,
            "skipped": len(results) - applied,
            "dry_run": dry_run,
            "min_events": min_events,
            "max_age_days": max_age_days,
        },
        "_meta": {"timing_ms": elapsed_ms},
    }
