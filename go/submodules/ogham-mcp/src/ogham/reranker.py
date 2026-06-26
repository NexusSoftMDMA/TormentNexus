"""Optional cross-encoder reranking for search results.

Supports two backends:
  - flashrank: ms-marco-MiniLM-L-12-v2, 21MB, CPU-only (default)
  - bge: BAAI/bge-reranker-v2-m3, 568M params, multilingual

Enable via RERANK_ENABLED=true, select backend via RERANK_MODEL=flashrank|bge.
"""

import logging

logger = logging.getLogger(__name__)

_ranker = None
_ranker_type = None


def _get_ranker():
    """Lazy-load reranker singleton based on config."""
    global _ranker, _ranker_type
    if _ranker is not None:
        return _ranker, _ranker_type

    from ogham.config import settings

    model = settings.rerank_model

    if model == "bge":
        try:
            from sentence_transformers import CrossEncoder

            _ranker = CrossEncoder("BAAI/bge-reranker-v2-m3")
            _ranker_type = "bge"
            logger.info("BGE reranker loaded (bge-reranker-v2-m3)")
            return _ranker, _ranker_type
        except ImportError:
            logger.debug("sentence-transformers not installed, falling back to flashrank")

    # Default: flashrank
    try:
        from flashrank import Ranker

        _ranker = Ranker(model_name="ms-marco-MiniLM-L-12-v2")
        _ranker_type = "flashrank"
        logger.info("FlashRank reranker loaded")
        return _ranker, _ranker_type
    except ImportError:
        logger.debug("flashrank not installed, reranking disabled")
        return None, None


def rerank_results(
    query: str,
    results: list[dict],
    top_k: int = 10,
    alpha: float = 0.55,
) -> list[dict]:
    """Rerank search results using cross-encoder.

    Blends the original retrieval score with the cross-encoder score:
        final = (1 - alpha) * retrieval_score + alpha * ce_score

    Args:
        query: The search query text.
        results: List of memory dicts with 'content' and 'relevance' keys.
        top_k: Number of results to return after reranking.
        alpha: Weight for cross-encoder score (0 = retrieval only, 1 = CE only).

    Returns:
        Reranked list of memory dicts, truncated to top_k.
    """
    ranker, ranker_type = _get_ranker()
    if ranker is None or not results:
        return results[:top_k]

    if ranker_type == "bge":
        return _rerank_bge(ranker, query, results, top_k, alpha)
    return _rerank_flashrank(ranker, query, results, top_k, alpha)


def _rerank_bge(ranker, query, results, top_k, alpha):
    """Rerank using BGE CrossEncoder via sentence-transformers."""
    pairs = [(query, r.get("content", "")) for r in results]

    try:
        scores = ranker.predict(pairs)
    except Exception:
        logger.exception("BGE reranking failed, returning original order")
        return results[:top_k]

    reranked = []
    for idx, ce_score in enumerate(scores):
        result = results[idx].copy()
        retrieval_score = float(result.get("relevance", 0))
        result["relevance"] = (1 - alpha) * retrieval_score + alpha * float(ce_score)
        result["ce_score"] = float(ce_score)
        reranked.append(result)

    reranked.sort(key=lambda x: x["relevance"], reverse=True)
    return reranked[:top_k]


def _rerank_flashrank(ranker, query, results, top_k, alpha):
    """Rerank using FlashRank."""
    from flashrank import RerankRequest

    passages = [{"id": i, "text": r.get("content", "")} for i, r in enumerate(results)]

    try:
        req = RerankRequest(query=query, passages=passages)
        ranked = ranker.rerank(req)
    except Exception:
        logger.exception("FlashRank reranking failed, returning original order")
        return results[:top_k]

    reranked = []
    for item in ranked:
        idx = item["id"]
        if idx < len(results):
            result = results[idx].copy()
            ce_score = float(item["score"])
            retrieval_score = float(result.get("relevance", 0))
            result["relevance"] = (1 - alpha) * retrieval_score + alpha * ce_score
            result["ce_score"] = ce_score
            reranked.append(result)

    return reranked[:top_k]
