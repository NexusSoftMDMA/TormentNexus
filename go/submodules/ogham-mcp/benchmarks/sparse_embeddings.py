"""Sparse vector generation using FlagEmbedding's BGE-M3 model.

Benchmark-only utility — not part of the main ogham package.
Generates sparse (lexical_weights) vectors from BGE-M3's multi-vector output.
"""

import logging
from typing import Any

logger = logging.getLogger(__name__)

_model = None


def _get_model():
    """Lazy-load BGEM3FlagModel (heavy: ~1.2GB model weights)."""
    global _model
    if _model is None:
        from FlagEmbedding import BGEM3FlagModel

        logger.info("Loading BGE-M3 model via FlagEmbedding (CPU, fp32)...")
        _model = BGEM3FlagModel("BAAI/bge-m3", use_fp16=False)
        logger.info("BGE-M3 model loaded.")
    return _model


def generate_sparse_vectors(
    texts: list[str],
    batch_size: int = 5,
    on_progress: Any = None,
) -> list[dict[str, float]]:
    """Generate sparse vectors for a list of texts.

    Returns list of {token_string: weight} dicts (FlagEmbedding's native format).
    """
    model = _get_model()
    all_sparse: list[dict[str, float]] = []

    for i in range(0, len(texts), batch_size):
        batch = texts[i : i + batch_size]
        output = model.encode(
            batch,
            return_dense=False,
            return_sparse=True,
            return_colbert_vecs=False,
        )
        for weights in output["lexical_weights"]:
            # weights is a dict {token_id_str: float_weight}
            # Convert token IDs to ints for sparsevec indexing
            sparse = {}
            for token_id, weight in weights.items():
                idx = int(token_id)
                sparse[idx] = float(weight)
            all_sparse.append(sparse)

        if on_progress:
            on_progress(min(i + batch_size, len(texts)), len(texts))

    return all_sparse


def sparse_to_sparsevec_literal(sparse: dict[int, float], dim: int = 250002) -> str:
    """Convert {index: weight} dict to pgvector sparsevec literal format.

    Format: '{index1:value1,index2:value2,...}/dim'
    BGE-M3 vocabulary size is 250002 tokens.
    """
    if not sparse:
        return "{}/{}".format("", dim)
    entries = ",".join(f"{idx}:{val:.6f}" for idx, val in sorted(sparse.items()))
    return "{" + entries + "}/" + str(dim)


def check_sparsevec_limits(sparse_vectors: list[dict[int, float]]) -> dict[str, Any]:
    """Check whether sparse vectors fit within pgvector's sparsevec limits.

    pgvector HNSW index caps at 1,000 non-zero elements.
    Returns stats about non-zero counts.
    """
    counts = [len(v) for v in sparse_vectors]
    over_limit = sum(1 for c in counts if c > 1000)
    return {
        "total": len(counts),
        "min_nnz": min(counts) if counts else 0,
        "max_nnz": max(counts) if counts else 0,
        "mean_nnz": sum(counts) / len(counts) if counts else 0,
        "over_1000": over_limit,
        "pct_over_1000": (over_limit / len(counts) * 100) if counts else 0,
    }
