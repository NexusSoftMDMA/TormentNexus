"""Ranking-quality metrics: nDCG@k, MRR@k, Recall@k.

Pure-Python, dependency-free. Each metric takes ``predicted`` (the ranked
list of IDs returned by the system under test) and ``relevant`` (the set
or list of IDs known to be correct for the query). Metrics are normalized
to the 0..1 range so they can be averaged across queries and compared
across releases.
"""

from __future__ import annotations

import math
from typing import Iterable


def _to_set(ids: Iterable[str]) -> set[str]:
    return set(ids) if not isinstance(ids, set) else ids


def recall_at_k(predicted: list[str], relevant: Iterable[str], k: int = 10) -> float:
    """Fraction of relevant items present in the top-k predictions."""
    rel = _to_set(relevant)
    if not rel:
        return 0.0
    top_k = predicted[:k]
    hits = sum(1 for p in top_k if p in rel)
    return hits / len(rel)


def mrr_at_k(predicted: list[str], relevant: Iterable[str], k: int = 10) -> float:
    """Reciprocal rank of the first relevant item in the top-k.

    Returns 0.0 if no relevant item appears within the top-k.
    """
    rel = _to_set(relevant)
    for i, p in enumerate(predicted[:k]):
        if p in rel:
            return 1.0 / (i + 1)
    return 0.0


def dcg(predicted: list[str], relevant: Iterable[str], k: int = 10) -> float:
    """Discounted Cumulative Gain (binary relevance)."""
    rel = _to_set(relevant)
    score = 0.0
    for i, p in enumerate(predicted[:k]):
        if p in rel:
            score += 1.0 / math.log2(i + 2)  # +2 because i is 0-indexed
    return score


def ndcg_at_k(predicted: list[str], relevant: Iterable[str], k: int = 10) -> float:
    """Normalized Discounted Cumulative Gain.

    Divides DCG by the ideal DCG (all relevant items at the top of the
    ranking). Range: 0..1.
    """
    rel = _to_set(relevant)
    if not rel:
        return 0.0
    actual = dcg(predicted, rel, k)
    ideal_count = min(len(rel), k)
    ideal = sum(1.0 / math.log2(i + 2) for i in range(ideal_count))
    if ideal == 0.0:
        return 0.0
    return actual / ideal


def aggregate(metrics: list[dict]) -> dict:
    """Mean of each numeric field across a list of per-query metric dicts."""
    if not metrics:
        return {}
    numeric_keys = {k for m in metrics for k, v in m.items() if isinstance(v, (int, float))}
    return {
        k: round(sum(m.get(k, 0.0) for m in metrics) / len(metrics), 4)
        for k in numeric_keys
    }
