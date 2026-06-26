#!/usr/bin/env python3
"""Compare ONNX BGE-M3 embeddings (with HF tokenizer) against FlagEmbedding reference.

Tests that our ONNX + HF tokenizer approach produces equivalent vectors to
FlagEmbedding's PyTorch implementation for dense and sparse outputs.

Usage:
    uv run python3 benchmarks/test-onnx-embedder.py

Prerequisites:
    - ONNX model files in models/bge-m3-onnx/ (bge_m3_model.onnx + bge_m3_model.onnx_data)
    - onnxruntime, transformers, FlagEmbedding installed (dev deps)
"""

import sys
import time
from pathlib import Path

import numpy as np

MODEL_DIR = Path(__file__).parent.parent / "models" / "bge-m3-onnx"
MODEL_PATH = MODEL_DIR / "bge_m3_model.onnx"

TEST_TEXTS = [
    "The quick brown fox jumps over the lazy dog.",
    "BGE-M3 produces dense, sparse, and ColBERT vectors from a single model.",
    "Remember to buy milk, eggs, and bread from the store on Thursday.",
    "The user prefers dark mode and vim keybindings in all editors.",
    "We decided to use PostgreSQL with pgvector for the memory backend "
    "because it supports both dense and sparse vector types natively.",
]


def load_onnx_model():
    """Load the ONNX model with HF tokenizer (our approach)."""
    import onnxruntime as ort
    from transformers import AutoTokenizer

    print("Loading HF tokenizer for BAAI/bge-m3...")
    tokenizer = AutoTokenizer.from_pretrained("BAAI/bge-m3")

    print(f"Loading ONNX model from {MODEL_PATH}...")
    options = ort.SessionOptions()
    options.enable_mem_pattern = True
    options.enable_cpu_mem_arena = True
    options.log_severity_level = 2  # WARNING
    options.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_BASIC

    session = ort.InferenceSession(
        str(MODEL_PATH),
        sess_options=options,
        providers=["CPUExecutionProvider"],
    )

    # Print model input/output info
    print("\nModel inputs:")
    for inp in session.get_inputs():
        print(f"  {inp.name}: {inp.shape} ({inp.type})")
    print("Model outputs:")
    for out in session.get_outputs():
        print(f"  {out.name}: {out.shape} ({out.type})")

    return tokenizer, session


def onnx_encode(tokenizer, session, text: str) -> dict:
    """Encode text using ONNX model + HF tokenizer."""
    special_token_ids = {0, 1, 2, 3}  # [PAD], [UNK], [CLS], [SEP] for XLM-RoBERTa

    # Tokenize with HF
    inputs = tokenizer(text, return_tensors="np", padding=False, truncation=True, max_length=8192)
    input_ids = inputs["input_ids"].astype(np.int64)
    attention_mask = inputs["attention_mask"].astype(np.int64)

    # Run model
    outputs = session.run(None, {"input_ids": input_ids, "attention_mask": attention_mask})
    dense_embeddings, sparse_weights, colbert_vectors = outputs

    # Dense: already L2-normalized by the model export
    dense_vecs = dense_embeddings[0]

    # Sparse: extract per-token weights
    sparse_dict = {}
    for i, token_id in enumerate(input_ids[0]):
        if attention_mask[0, i] == 1 and int(token_id) not in special_token_ids:
            weight = float(np.max(sparse_weights[0, i]))
            if weight > 0:
                tid = str(int(token_id))
                sparse_dict[tid] = max(sparse_dict.get(tid, 0), weight)

    # ColBERT: per-token vectors for non-padding positions
    colbert_list = []
    for i in range(colbert_vectors.shape[1]):
        if attention_mask[0, i] == 1:
            colbert_list.append(colbert_vectors[0, i])

    return {
        "dense_vecs": dense_vecs,
        "lexical_weights": sparse_dict,
        "colbert_vecs": colbert_list,
    }


def load_flagembedding():
    """Load FlagEmbedding reference model."""
    from FlagEmbedding import BGEM3FlagModel

    print("Loading FlagEmbedding BGEM3FlagModel (CPU, fp32)...")
    model = BGEM3FlagModel("BAAI/bge-m3", use_fp16=False)
    return model


def flagembedding_encode(model, text: str) -> dict:
    """Encode text using FlagEmbedding reference."""
    result = model.encode(
        [text],
        return_dense=True,
        return_sparse=True,
        return_colbert_vecs=True,
    )
    dense = result["dense_vecs"][0]
    sparse_raw = result["lexical_weights"][0]
    colbert = result["colbert_vecs"][0]

    # Convert sparse from {token_str: weight} to {token_id_str: weight}
    # FlagEmbedding returns string token IDs as keys
    sparse_dict = {str(k): float(v) for k, v in sparse_raw.items()}

    return {
        "dense_vecs": np.array(dense),
        "lexical_weights": sparse_dict,
        "colbert_vecs": [np.array(v) for v in colbert],
    }


def cosine_similarity(a, b):
    """Compute cosine similarity between two vectors."""
    a, b = np.array(a), np.array(b)
    return float(np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b)))


def compare_sparse(onnx_sparse: dict, flag_sparse: dict) -> dict:
    """Compare two sparse weight dicts."""
    all_keys = set(onnx_sparse.keys()) | set(flag_sparse.keys())
    onnx_only = set(onnx_sparse.keys()) - set(flag_sparse.keys())
    flag_only = set(flag_sparse.keys()) - set(onnx_sparse.keys())
    common = set(onnx_sparse.keys()) & set(flag_sparse.keys())

    weight_diffs = []
    for k in common:
        weight_diffs.append(abs(onnx_sparse[k] - flag_sparse[k]))

    return {
        "total_keys": len(all_keys),
        "onnx_only": len(onnx_only),
        "flag_only": len(flag_only),
        "common": len(common),
        "jaccard": len(common) / len(all_keys) if all_keys else 1.0,
        "mean_weight_diff": float(np.mean(weight_diffs)) if weight_diffs else 0.0,
        "max_weight_diff": float(np.max(weight_diffs)) if weight_diffs else 0.0,
    }


def main():
    if not MODEL_PATH.exists():
        print(f"ERROR: ONNX model not found at {MODEL_PATH}")
        print("Download from: https://github.com/yuniko-software/bge-m3-onnx/releases")
        print(f"Unzip to: {MODEL_DIR}/")
        sys.exit(1)

    # Load both models
    tokenizer, onnx_session = load_onnx_model()

    print()
    flag_model = load_flagembedding()

    print(f"\n{'=' * 70}")
    print(f"Comparing {len(TEST_TEXTS)} test texts")
    print(f"{'=' * 70}\n")

    for i, text in enumerate(TEST_TEXTS):
        print(f"--- Text {i + 1}: {text[:60]}{'...' if len(text) > 60 else ''}")

        # ONNX encode
        t0 = time.time()
        onnx_result = onnx_encode(tokenizer, onnx_session, text)
        onnx_time = time.time() - t0

        # FlagEmbedding encode
        t0 = time.time()
        flag_result = flagembedding_encode(flag_model, text)
        flag_time = time.time() - t0

        # Compare dense
        dense_sim = cosine_similarity(onnx_result["dense_vecs"], flag_result["dense_vecs"])
        print(f"  Dense cosine similarity: {dense_sim:.6f}")

        # Compare sparse
        sparse_cmp = compare_sparse(onnx_result["lexical_weights"], flag_result["lexical_weights"])
        print(
            f"  Sparse: {sparse_cmp['common']}/{sparse_cmp['total_keys']} common tokens "
            f"(jaccard={sparse_cmp['jaccard']:.3f}), "
            f"mean weight diff={sparse_cmp['mean_weight_diff']:.4f}, "
            f"max={sparse_cmp['max_weight_diff']:.4f}"
        )
        if sparse_cmp["onnx_only"]:
            print(f"    ONNX-only tokens: {sparse_cmp['onnx_only']}")
        if sparse_cmp["flag_only"]:
            print(f"    FlagEmbedding-only tokens: {sparse_cmp['flag_only']}")

        # Compare ColBERT (just first few vectors)
        n_colbert = min(len(onnx_result["colbert_vecs"]), len(flag_result["colbert_vecs"]))
        if n_colbert > 0:
            colbert_sims = [
                cosine_similarity(onnx_result["colbert_vecs"][j], flag_result["colbert_vecs"][j])
                for j in range(n_colbert)
            ]
            print(
                f"  ColBERT: {n_colbert} vectors, "
                f"mean cosine={np.mean(colbert_sims):.6f}, "
                f"min={np.min(colbert_sims):.6f}"
            )

        print(f"  Time: ONNX={onnx_time:.3f}s, FlagEmbedding={flag_time:.3f}s")
        print()

    # Quick check: is dense already normalized?
    onnx_result = onnx_encode(tokenizer, onnx_session, TEST_TEXTS[0])
    norm = np.linalg.norm(onnx_result["dense_vecs"])
    print(f"Dense vector L2 norm: {norm:.6f} (should be ~1.0 if pre-normalized)")


if __name__ == "__main__":
    main()
