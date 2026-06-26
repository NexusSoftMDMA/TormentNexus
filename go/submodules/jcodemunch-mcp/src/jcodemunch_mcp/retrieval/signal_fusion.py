"""Unified signal fusion pipeline — Weighted Reciprocal Rank (WRR).

Merges four independent scoring channels into one ranked list:

  - **lexical**:    BM25 text relevance
  - **structural**: PageRank / import-graph centrality
  - **similarity**: Embedding cosine distance (optional)
  - **identity**:   Exact / prefix match on symbol name or qualified ID

Each channel produces a ranked list.  The final score for every symbol
that appears in *any* channel is:

    score(s) = sum( weight[c] / (k + rank(c, s)) )
               for c in channels where s appears

Where ``k`` is a smoothing constant (default 60).
"""

from __future__ import annotations

import logging
from dataclasses import dataclass, field
from typing import Optional

logger = logging.getLogger(__name__)

# Default channel weights — tuned for code search.
# Identity is weighted highest because an exact name hit is almost always
# what the user wants; similarity second (semantic intent); lexical third
# (keyword overlap); structural last (tiebreaker).
DEFAULT_WEIGHTS: dict[str, float] = {
    "lexical": 1.0,
    "structural": 0.4,
    "similarity": 0.8,
    "identity": 2.0,
}

DEFAULT_SMOOTHING = 60


@dataclass
class ChannelResult:
    """A single channel's ranked output.

    ``ranked_ids`` is an ordered list of symbol IDs from most to least
    relevant.  Only symbols that *matched* in this channel should appear.
    """

    name: str
    ranked_ids: list[str]
    # None = "use the default weight for this channel name from the weights
    # dict / DEFAULT_WEIGHTS". An explicit float (including 0.0) overrides it.
    weight: Optional[float] = None
    # Optional per-ID raw scores (for debug output)
    raw_scores: dict[str, float] = field(default_factory=dict)


@dataclass
class FusedResult:
    """One symbol's fused score with per-channel diagnostics."""

    symbol_id: str
    score: float
    channel_contributions: dict[str, float] = field(default_factory=dict)
    channel_ranks: dict[str, int] = field(default_factory=dict)


def fuse(
    channels: list[ChannelResult],
    *,
    smoothing: int = DEFAULT_SMOOTHING,
    weights: Optional[dict[str, float]] = None,
) -> list[FusedResult]:
    """Run Weighted Reciprocal Rank fusion across channels.

    Args:
        channels: Ranked results from each scoring channel.
        smoothing: RRF smoothing constant ``k`` (higher → less top-heavy).
        weights: Per-channel weight overrides.  A channel whose own
            ``weight`` is ``None`` (the builders' default) draws its weight
            from this dict by channel name, falling back to
            ``DEFAULT_WEIGHTS`` then ``1.0``.  A channel that sets an
            explicit ``weight`` (including ``0.0``) uses that literal value.

    Returns:
        List of ``FusedResult`` sorted by descending fused score.
    """
    effective_weights = dict(DEFAULT_WEIGHTS)
    if weights:
        effective_weights.update(weights)

    # Build per-symbol accumulator
    accum: dict[str, FusedResult] = {}

    for ch in channels:
        w = ch.weight if ch.weight is not None else effective_weights.get(ch.name, 1.0)

        for rank_0, sid in enumerate(ch.ranked_ids):
            rank_1 = rank_0 + 1  # 1-based rank
            contribution = w / (smoothing + rank_1)

            if sid not in accum:
                accum[sid] = FusedResult(symbol_id=sid, score=0.0)
            entry = accum[sid]
            entry.score += contribution
            entry.channel_contributions[ch.name] = contribution
            entry.channel_ranks[ch.name] = rank_1

    results = sorted(accum.values(), key=lambda r: r.score, reverse=True)
    return results


# ---------------------------------------------------------------------------
# Convenience: build channel results from existing index data
# ---------------------------------------------------------------------------

def build_lexical_channel(
    symbols: list[dict],
    query_terms: list[str],
    idf: dict[str, float],
    avgdl: float,
    centrality: Optional[dict] = None,
    *,
    weight: Optional[float] = None,  # None means "use default from weights dict"
) -> ChannelResult:
    """Score all symbols via BM25 and return as a ranked channel.

    Reuses the existing ``_bm25_score`` function from search_symbols.
    """

    scored: list[tuple[float, str]] = []
    for sym in symbols:
        # BM25 score WITHOUT identity (we have a separate identity channel)
        score = _bm25_score_no_identity(sym, query_terms, idf, avgdl, centrality)
        if score > 0:
            scored.append((score, sym["id"]))

    scored.sort(key=lambda x: x[0], reverse=True)

    return ChannelResult(
        name="lexical",
        ranked_ids=[sid for _, sid in scored],
        weight=weight,
        raw_scores={sid: s for s, sid in scored},
    )


def build_structural_channel(
    symbols: list[dict],
    pagerank_scores: dict[str, float],
    candidate_ids: Optional[set[str]] = None,
    *,
    weight: Optional[float] = None,
) -> ChannelResult:
    """Rank symbols by PageRank of their containing file.

    If ``candidate_ids`` is given, only those symbols participate.
    """
    scored: list[tuple[float, str]] = []
    for sym in symbols:
        sid = sym["id"]
        if candidate_ids and sid not in candidate_ids:
            continue
        pr = pagerank_scores.get(sym.get("file", ""), 0.0)
        if pr > 0:
            scored.append((pr, sid))

    scored.sort(key=lambda x: x[0], reverse=True)

    return ChannelResult(
        name="structural",
        ranked_ids=[sid for _, sid in scored],
        weight=weight,
        raw_scores={sid: s for s, sid in scored},
    )


def build_identity_channel(
    symbols: list[dict],
    query: str,
    *,
    weight: Optional[float] = None,
) -> ChannelResult:
    """Rank symbols by identity match (exact/prefix/segment).

    Reuses ``_identity_score`` from search_symbols.
    """
    from ..tools.search_symbols import _identity_score  # noqa: PLC0415

    query_lower = query.lower()
    scored: list[tuple[float, str]] = []
    for sym in symbols:
        s = _identity_score(sym, query_lower, raw_query=query_lower)
        if s > 0:
            scored.append((s, sym["id"]))

    scored.sort(key=lambda x: x[0], reverse=True)

    return ChannelResult(
        name="identity",
        ranked_ids=[sid for _, sid in scored],
        weight=weight,
        raw_scores={sid: s for s, sid in scored},
    )


def build_similarity_channel(
    query_embedding: list[float],
    symbol_embeddings: dict[str, list[float]],
    *,
    weight: Optional[float] = None,
    min_similarity: float = 0.1,
) -> ChannelResult:
    """Rank symbols by embedding cosine similarity.

    Args:
        query_embedding: The query's embedding vector.
        symbol_embeddings: Mapping of symbol_id -> embedding vector.
        weight: Channel weight override.
        min_similarity: Minimum cosine similarity to include in ranked list.
    """
    from ..tools.search_symbols import _cosine_similarity  # noqa: PLC0415

    scored: list[tuple[float, str]] = []
    for sid, emb in symbol_embeddings.items():
        sim = _cosine_similarity(query_embedding, emb)
        if sim >= min_similarity:
            scored.append((sim, sid))

    scored.sort(key=lambda x: x[0], reverse=True)

    return ChannelResult(
        name="similarity",
        ranked_ids=[sid for _, sid in scored],
        weight=weight,
        raw_scores={sid: s for s, sid in scored},
    )


# ---------------------------------------------------------------------------
# Internal: BM25 without identity boost (to avoid double-counting)
# ---------------------------------------------------------------------------

def _bm25_score_no_identity(
    sym: dict,
    query_terms: list[str],
    idf: dict[str, float],
    avgdl: float,
    centrality: Optional[dict] = None,
) -> float:
    """BM25 score without identity channel (exact/prefix/segment match).

    Used by the fusion pipeline so identity is scored separately.
    """
    from ..tools.search_symbols import _sym_tokens, _BM25_K1, _BM25_B  # noqa: PLC0415

    _sym_tokens(sym)
    tf_raw = sym["_tf"]
    dl = sym["_dl"]

    score = 0.0
    K = _BM25_K1 * (1 - _BM25_B + _BM25_B * dl / max(avgdl, 1.0))
    for term in set(query_terms):
        idf_val = idf.get(term, 0.0)
        if idf_val == 0.0:
            continue
        tf = tf_raw.get(term, 0)
        if tf == 0:
            continue
        score += idf_val * (tf * (_BM25_K1 + 1)) / (tf + K)

    if centrality and score > 0:
        score += centrality.get(sym.get("file", ""), 0.0)

    return score


# ---------------------------------------------------------------------------
# Config-aware weight loader
# ---------------------------------------------------------------------------

def load_fusion_weights() -> tuple[dict[str, float], int]:
    """Load fusion weights and smoothing from config, with defaults.

    Reads from ``config.jsonc`` under ``retrieval.fusion_weights`` and
    ``retrieval.fusion_smoothing``.

    Returns:
        (weights_dict, smoothing_k)
    """
    weights = dict(DEFAULT_WEIGHTS)
    smoothing = DEFAULT_SMOOTHING
    try:
        from .. import config as _cfg  # noqa: PLC0415
        retrieval = _cfg.get("retrieval", {})
        if isinstance(retrieval, dict):
            user_weights = retrieval.get("fusion_weights", {})
            if isinstance(user_weights, dict):
                for k, v in user_weights.items():
                    if isinstance(v, (int, float)) and v >= 0:
                        weights[k] = float(v)
            user_smooth = retrieval.get("fusion_smoothing")
            if isinstance(user_smooth, int) and user_smooth > 0:
                smoothing = user_smooth
    except Exception:
        logger.debug("Failed to load fusion config, using defaults", exc_info=True)
    return weights, smoothing
