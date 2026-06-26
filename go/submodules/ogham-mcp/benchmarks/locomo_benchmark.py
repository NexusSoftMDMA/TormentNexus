#!/usr/bin/env python3
"""LoCoMo benchmark for Ogham MCP.

Evaluates retrieval quality using the LoCoMo dataset (Maharana et al.)
which tests long-term conversational memory with 1,986 QA pairs across
10 conversations.

Metrics:
- Recall@K: fraction of questions where the ground truth evidence
  appears in the top K retrieved results
- MRR: Mean Reciprocal Rank of the first relevant result

Usage:
    # Download LoCoMo dataset first:
    curl -L -o benchmarks/locomo10.json \
      https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json

    # Run the benchmark:
    uv run python3 benchmarks/locomo_benchmark.py [--top-k 10] [--profile locomo]

    # Clean up after:
    uv run python3 benchmarks/locomo_benchmark.py --cleanup

Requires: Ogham MCP server configured with a working database and embedding provider.
"""

import argparse
import json
import logging
import os
import sys
import time
from pathlib import Path

MAX_RETRIES = 3
RETRY_DELAY = 2.0  # seconds, doubles each retry

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s")
logger = logging.getLogger(__name__)

DATA_FILE = Path(__file__).parent / "locomo10.json"
RESULTS_FILE = Path(__file__).parent / "locomo_results.json"


# LoCoMo category mapping
def _with_retry(fn, *args, **kwargs):
    """Call fn with retry on connection errors (Neon pooler drops)."""
    for attempt in range(MAX_RETRIES):
        try:
            return fn(*args, **kwargs)
        except Exception as e:
            err = str(e).lower()
            if "connection" in err or "closed" in err or "terminated" in err:
                if attempt < MAX_RETRIES - 1:
                    delay = RETRY_DELAY * (2**attempt)
                    logger.warning(
                        "Connection lost, retrying in %.1fs (%d/%d)",
                        delay,
                        attempt + 1,
                        MAX_RETRIES,
                    )
                    time.sleep(delay)
                    # Reset the database backend to force a new connection
                    try:
                        from ogham.database import _reset_backend

                        _reset_backend()
                    except Exception:
                        pass
                    continue
            raise


CATEGORIES = {
    "1": "single-hop",
    "2": "temporal",
    "3": "multi-hop",
    "4": "open-domain",
    "5": "adversarial",
}


def load_dataset() -> list[dict]:
    if not DATA_FILE.exists():
        print(f"Dataset not found at {DATA_FILE}")
        print("Download it:")
        print(
            "  curl -L -o benchmarks/locomo10.json "
            "https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json"
        )
        sys.exit(1)
    return json.loads(DATA_FILE.read_text())


def _extract_sessions(conversation: dict) -> list[tuple[str, list[dict]]]:
    """Extract ordered sessions from the LoCoMo conversation dict."""
    import re

    sessions = []
    for key in conversation:
        m = re.match(r"^session_(\d+)$", key)
        if m and isinstance(conversation[key], list):
            sessions.append((int(m.group(1)), key, conversation[key]))
    sessions.sort(key=lambda x: x[0])
    return [(key, turns) for _, key, turns in sessions]


def ingest_conversations(
    data: list[dict], profile: str, chunk_size: int = 0, chunk_overlap: int = 2
) -> int:
    """Ingest LoCoMo conversations into Ogham as memories.

    chunk_size=0 means full sessions. chunk_size>0 uses sliding windows.
    """
    from ogham.database import store_memory as db_store
    from ogham.embeddings import generate_embedding

    total_stored = 0
    for sample in data:
        conv_id = sample["sample_id"]
        conversation = sample["conversation"]
        sessions = _extract_sessions(conversation)
        total_turns = 0

        for session_key, turns in sessions:
            total_turns += len(turns)
            date = conversation.get(f"{session_key}_date_time", "")

            if chunk_size > 0 and len(turns) > chunk_size:
                # Sliding window chunking
                step = max(1, chunk_size - chunk_overlap)
                for i in range(0, len(turns), step):
                    window = turns[i : i + chunk_size]
                    lines = [f"{t.get('speaker', '?')}: {t.get('text', '')}" for t in window]
                    content = "\n".join(lines)
                    if len(content.strip()) < 20:
                        continue
                    embedding = generate_embedding(content)
                    dia_ids = [t.get("dia_id", "") for t in window if t.get("dia_id")]
                    chunk_tags = [f"conv:{conv_id}", f"session:{session_key}", f"chunk:{i}"]
                    chunk_tags.extend([f"dia:{d}" for d in dia_ids])
                    _with_retry(
                        db_store,
                        content=content,
                        embedding=embedding,
                        profile=profile,
                        source="locomo-benchmark",
                        tags=chunk_tags,
                        metadata={"date": date} if date else None,
                    )
                    total_stored += 1
            else:
                # Full session as one memory
                lines = [f"{t.get('speaker', '?')}: {t.get('text', '')}" for t in turns]
                content = "\n".join(lines)
                if len(content.strip()) < 20:
                    continue
                embedding = generate_embedding(content)
                dia_ids = [t.get("dia_id", "") for t in turns if t.get("dia_id")]
                session_tags = [f"conv:{conv_id}", f"session:{session_key}"]
                session_tags.extend([f"dia:{d}" for d in dia_ids])
                _with_retry(
                    db_store,
                    content=content,
                    embedding=embedding,
                    profile=profile,
                    source="locomo-benchmark",
                    tags=session_tags,
                    metadata={"date": date} if date else None,
                )
                total_stored += 1

        logger.info(
            "Ingested %s (%d sessions, %d turns, %d stored)",
            conv_id,
            len(sessions),
            total_turns,
            total_stored,
        )

    return total_stored


def evaluate(data: list[dict], profile: str, top_k: int = 10, graph_depth: int = 0) -> dict:
    """Run QA evaluation against stored memories."""
    from ogham.database import hybrid_search_memories
    from ogham.embeddings import generate_embedding

    results = {
        "total_questions": 0,
        "recall_at_k": 0,
        "reciprocal_ranks": [],
        "by_category": {},
    }

    for sample in data:
        conv_id = sample["sample_id"]
        for qa in sample["qa"]:
            question = str(qa["question"])
            answer = str(qa.get("answer", qa.get("adversarial_answer", "")))
            raw_evidence = qa.get("evidence", "")
            # Evidence can be a list of strings, a single string, or other types
            if isinstance(raw_evidence, list):
                evidence = " ".join(str(e) for e in raw_evidence)
            else:
                evidence = str(raw_evidence)
            category = str(qa.get("category", "unknown"))
            cat_name = CATEGORIES.get(category, f"cat-{category}")

            # Search Ogham
            query_embedding = generate_embedding(question)
            if graph_depth > 0:
                from ogham.database import graph_augmented_search

                search_results = _with_retry(
                    graph_augmented_search,
                    query_text=question,
                    query_embedding=query_embedding,
                    profile=profile,
                    limit=top_k,
                    graph_depth=graph_depth,
                )
            else:
                search_results = _with_retry(
                    hybrid_search_memories,
                    query_text=question,
                    query_embedding=query_embedding,
                    profile=profile,
                    limit=top_k,
                )

            # Check if evidence or answer appears in any of the top K results
            found = False
            rank = 0
            answer_lower = answer.lower().strip()
            for i, result in enumerate(search_results):
                content_lower = result.get("content", "").lower()

                # Exact substring match
                if answer_lower and answer_lower in content_lower:
                    found = True
                    rank = i + 1
                    break

                # Evidence match (any evidence fragment in content)
                if evidence:
                    evidence_parts = evidence.split()
                    # Check if any substantial evidence words appear
                    if len(evidence_parts) > 2:
                        ev_fragment = " ".join(evidence_parts[:5]).lower()
                        if ev_fragment in content_lower:
                            found = True
                            rank = i + 1
                            break

                # Fuzzy match for short answers: all words present in content
                if answer_lower and len(answer_lower) < 30:
                    words = [w for w in answer_lower.split() if len(w) > 2]
                    if words and all(w in content_lower for w in words):
                        found = True
                        rank = i + 1
                        break

            # Evidence ID match via dialogue ID tags
            if not found and isinstance(raw_evidence, list):
                evidence_ids = [str(e) for e in raw_evidence]
                for i, result in enumerate(search_results):
                    result_tags = result.get("tags", [])
                    for eid in evidence_ids:
                        if f"dia:{eid}" in result_tags:
                            found = True
                            rank = i + 1
                            break
                    if found:
                        break

            results["total_questions"] += 1
            if found:
                results["recall_at_k"] += 1
                results["reciprocal_ranks"].append(1.0 / rank)
            else:
                results["reciprocal_ranks"].append(0.0)

            # Per-category tracking
            if cat_name not in results["by_category"]:
                results["by_category"][cat_name] = {"total": 0, "recall": 0, "rr_sum": 0.0}
            results["by_category"][cat_name]["total"] += 1
            if found:
                results["by_category"][cat_name]["recall"] += 1
                results["by_category"][cat_name]["rr_sum"] += 1.0 / rank

        logger.info("Evaluated %s (%d questions so far)", conv_id, results["total_questions"])

    # Compute final metrics
    total = results["total_questions"]
    recall = results["recall_at_k"] / total if total > 0 else 0.0
    mrr = sum(results["reciprocal_ranks"]) / total if total > 0 else 0.0

    summary = {
        "total_questions": total,
        "top_k": top_k,
        "recall_at_k": round(recall * 100, 1),
        "mrr": round(mrr, 3),
        "by_category": {},
    }

    for cat, stats in results["by_category"].items():
        cat_total = stats["total"]
        summary["by_category"][cat] = {
            "total": cat_total,
            "recall_at_k": round(stats["recall"] / cat_total * 100, 1) if cat_total > 0 else 0.0,
            "mrr": round(stats["rr_sum"] / cat_total, 3) if cat_total > 0 else 0.0,
        }

    return summary


def cleanup(profile: str):
    """Remove all benchmark memories."""
    from ogham.config import settings

    if settings.database_backend == "postgres":
        import psycopg

        conn = psycopg.connect(settings.database_url)
        conn.autocommit = True
        cur = conn.execute(
            "DELETE FROM memories WHERE profile = %s AND source = %s",
            (profile, "locomo-benchmark"),
        )
        count = cur.rowcount
        conn.close()
    else:
        from ogham.database import get_backend

        backend = get_backend()
        client = backend._get_client()
        result = (
            client.table("memories")
            .delete()
            .eq("profile", profile)
            .eq("source", "locomo-benchmark")
            .execute()
        )
        count = len(result.data) if result.data else 0

    logger.info("Deleted %d benchmark memories from profile '%s'", count, profile)


def main():
    parser = argparse.ArgumentParser(description="LoCoMo benchmark for Ogham MCP")
    parser.add_argument("--top-k", type=int, default=10, help="Top K for Recall@K (default: 10)")
    parser.add_argument("--profile", default="locomo", help="Ogham profile for benchmark data")
    parser.add_argument("--cleanup", action="store_true", help="Remove benchmark data and exit")
    parser.add_argument(
        "--skip-ingest", action="store_true", help="Skip ingestion (use existing data)"
    )
    parser.add_argument("--env-file", default=None, help="Path to .env file (override default)")
    parser.add_argument(
        "--chunk-size", type=int, default=0, help="Turn chunk size (0 = full session)"
    )
    parser.add_argument("--chunk-overlap", type=int, default=2, help="Overlap turns between chunks")
    parser.add_argument(
        "--graph-depth", type=int, default=0, help="Graph traversal depth (0 = off)"
    )
    args = parser.parse_args()

    # Load alternate env file if specified (e.g. Neon test bench)
    if args.env_file:
        from dotenv import load_dotenv

        load_dotenv(args.env_file, override=True)
        # Force settings to reload from new env
        from ogham.config import settings

        settings._reset()
        logger.info("Loaded env from %s", args.env_file)

    if args.cleanup:
        cleanup(args.profile)
        return

    data = load_dataset()
    total_qa = sum(len(s["qa"]) for s in data)
    logger.info("Loaded %d conversations with %d QA pairs", len(data), total_qa)

    if not args.skip_ingest:
        logger.info(
            "Ingesting conversations into profile '%s' (chunk_size=%s, overlap=%d)...",
            args.profile,
            args.chunk_size or "full-session",
            args.chunk_overlap,
        )
        start = time.time()
        stored = ingest_conversations(data, args.profile, args.chunk_size, args.chunk_overlap)
        elapsed = time.time() - start
        logger.info("Ingested %d chunks in %.1fs", stored, elapsed)

    logger.info("Running evaluation (Recall@%d)...", args.top_k)
    start = time.time()
    results = evaluate(data, args.profile, args.top_k, args.graph_depth)
    elapsed = time.time() - start

    results["evaluation_time_s"] = round(elapsed, 1)
    results["embedding_provider"] = os.environ.get("EMBEDDING_PROVIDER", "unknown")

    # Print results
    print("\n" + "=" * 50)
    print("LoCoMo Benchmark Results (Ogham MCP)")
    print("=" * 50)
    print(f"Questions:      {results['total_questions']}")
    print(f"Recall@{args.top_k}:      {results['recall_at_k']}%")
    print(f"MRR:            {results['mrr']}")
    print(f"Provider:       {results['embedding_provider']}")
    print(f"Eval time:      {results['evaluation_time_s']}s")
    print()
    print("By category:")
    for cat, stats in sorted(results["by_category"].items()):
        r_at_k = stats["recall_at_k"]
        mrr = stats["mrr"]
        n = stats["total"]
        print(f"  {cat:15s}  Recall@{args.top_k}: {r_at_k:5.1f}%  MRR: {mrr:.3f}  (n={n})")

    # Save results
    RESULTS_FILE.write_text(json.dumps(results, indent=2))
    print(f"\nResults saved to {RESULTS_FILE}")


if __name__ == "__main__":
    main()
