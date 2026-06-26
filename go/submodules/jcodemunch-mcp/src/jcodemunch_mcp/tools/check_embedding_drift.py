"""check_embedding_drift — re-embed the pinned 16-string canary and report drift.

Captures the canary on first run (or with ``force=True``); subsequent calls
re-embed those strings with the live provider and compare cosine similarity
to the stored vectors. Alarms when max cosine distance exceeds ``threshold``
(default 0.05, i.e. cos sim < 0.95).
"""

from __future__ import annotations

import time
from typing import Optional

from ..retrieval import embed_drift as _ed


def check_embedding_drift(
    capture: bool = False,
    force: bool = False,
    threshold: float = 0.05,
    storage_path: Optional[str] = None,
) -> dict:
    t0 = time.perf_counter()
    if capture:
        out = _ed.capture_canary(base_path=storage_path, force=force)
    else:
        if force:
            captured = _ed.capture_canary(base_path=storage_path, force=True)
            if not captured.get("captured"):
                captured["_meta"] = {
                    "timing_ms": round((time.perf_counter() - t0) * 1000, 2)
                }
                return captured
        out = _ed.check_drift(base_path=storage_path, threshold=threshold)
    out.setdefault("_meta", {})["timing_ms"] = round(
        (time.perf_counter() - t0) * 1000, 2
    )
    return out
