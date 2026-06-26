#!/usr/bin/env python3
"""BEAM benchmark for Ogham MCP.

Evaluates long-term memory retrieval using the BEAM dataset
(Tavakoli et al., ICLR 2026) -- 2,000 probing questions across
10 memory abilities tested against conversations up to 10M tokens.

Categories:
  - abstention             - contradiction_resolution
  - event_ordering         - information_extraction
  - instruction_following  - knowledge_update
  - multi_session_reasoning - preference_following
  - summarization          - temporal_reasoning

Metrics:
  - Recall@K: fraction of gold source chat IDs found in top K results
  - NDCG@K: normalized discounted cumulative gain
  - MRR: mean reciprocal rank of first relevant result

Usage:
    # Ingest a single chat (128K bucket, chat 1):
    uv run python3 benchmarks/beam_benchmark.py --ingest --bucket 100K --chat 1

    # Ingest all 128K chats:
    uv run python3 benchmarks/beam_benchmark.py --ingest --bucket 100K

    # Run evaluation on all ingested chats:
    uv run python3 benchmarks/beam_benchmark.py --eval --bucket 100K

    # Run only temporal_reasoning questions:
    uv run python3 benchmarks/beam_benchmark.py --eval --bucket 100K --category temporal_reasoning

    # Run a single chat's questions:
    uv run python3 benchmarks/beam_benchmark.py --eval --bucket 100K --chat 3

    # Clean up benchmark data:
    uv run python3 benchmarks/beam_benchmark.py --cleanup --bucket 100K

Requires:
  - BEAM repo cloned to /tmp/BEAM (or set --beam-dir)
  - Local Postgres with pgvector (benchmarks/.env.local)
  - Voyage API key configured
"""

import argparse
import json
import logging
import math
import os
import sys
import time
from pathlib import Path

MAX_RETRIES = 3
RETRY_DELAY = 2.0
# Per-provider batch size caps (avoids API "too many requests" errors).
# Voyage supports 1000, Gemini caps at 100, OpenAI 500.
# None = use provider default from settings.
EMBEDDING_BATCH_SIZE: int | None = None
# Graph augmentation depth. 0 = pure hybrid search. 1 = follow one hop of
# memory_relationships edges to pull connected memories into top-K.
# Default 0 -- experiments on 2026-04-06 showed graph_depth=1 with the current
# memory-similarity graph hurts retrieval significantly across all categories
# (-13 to -44pp) because similarity edges duplicate vector-search hits and
# displace diverse evidence in the top-K. The win requires a real entity-
# relationship graph (memory→entity edges), not memory→memory similarity.
# Override via BEAM_GRAPH_DEPTH env var to experiment.
BEAM_GRAPH_DEPTH: int = int(os.environ.get("BEAM_GRAPH_DEPTH", "0"))

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s")
logger = logging.getLogger(__name__)

DATA_DIR = Path(__file__).parent
RESULTS_DIR = DATA_DIR / "beam_results"
DEFAULT_BEAM_DIR = Path("/Users/kevinburns/Developer/web-projects/ogham-beam-dataset")

ALL_CATEGORIES = [
    "abstention",
    "contradiction_resolution",
    "event_ordering",
    "information_extraction",
    "instruction_following",
    "knowledge_update",
    "multi_session_reasoning",
    "preference_following",
    "summarization",
    "temporal_reasoning",
]

ALL_BUCKETS = ["100K", "500K", "1M", "10M"]


def _with_retry(fn, *args, **kwargs):
    """Call fn with retry on connection errors."""
    for attempt in range(MAX_RETRIES):
        try:
            return fn(*args, **kwargs)
        except Exception as e:
            if attempt == MAX_RETRIES - 1:
                raise
            wait = RETRY_DELAY * (2**attempt)
            logger.warning("Attempt %d failed: %s. Retrying in %.1fs", attempt + 1, e, wait)
            time.sleep(wait)


# ---------------------------------------------------------------------------
# Dataset loading
# ---------------------------------------------------------------------------


def list_chats(beam_dir: Path, bucket: str) -> list[int]:
    """List available chat numbers for a bucket."""
    bucket_dir = beam_dir / "chats" / bucket
    if not bucket_dir.exists():
        raise FileNotFoundError(f"Bucket dir not found: {bucket_dir}")
    chat_ids = []
    for entry in bucket_dir.iterdir():
        if entry.is_dir() and entry.name.isdigit():
            chat_ids.append(int(entry.name))
    return sorted(chat_ids)


def load_chat(beam_dir: Path, bucket: str, chat_id: int) -> list[dict]:
    """Load a chat's conversation turns as a flat list of messages."""
    chat_path = beam_dir / "chats" / bucket / str(chat_id) / "chat.json"
    with open(chat_path) as f:
        batches = json.load(f)

    messages = []
    for batch in batches:
        batch_anchor = batch.get("time_anchor", "")
        for turn in batch["turns"]:
            for msg in turn:
                # Prefer message-level time_anchor over batch-level.
                # BEAM messages have per-message anchors like "March-03-2024"
                # while batch-level is often "None" (string).
                msg_anchor = msg.get("time_anchor", "")
                if msg_anchor and msg_anchor != "None" and msg_anchor != "null":
                    msg["_time_anchor"] = msg_anchor
                elif batch_anchor and batch_anchor != "None" and batch_anchor != "null":
                    msg["_time_anchor"] = batch_anchor
                else:
                    msg["_time_anchor"] = ""
                msg["_chat_id"] = chat_id
                messages.append(msg)
    return messages


def load_topic(beam_dir: Path, bucket: str, chat_id: int) -> dict:
    """Load a chat's topic metadata."""
    topic_path = beam_dir / "chats" / bucket / str(chat_id) / "topic.json"
    with open(topic_path) as f:
        return json.load(f)


def load_probing_questions(beam_dir: Path, bucket: str, chat_id: int) -> dict:
    """Load probing questions keyed by category."""
    q_path = (
        beam_dir / "chats" / bucket / str(chat_id) / "probing_questions" / "probing_questions.json"
    )
    with open(q_path) as f:
        return json.load(f)


# ---------------------------------------------------------------------------
# Ingestion -- conversation turns into Ogham memories
# ---------------------------------------------------------------------------


def ingest_chat(
    beam_dir: Path,
    bucket: str,
    chat_id: int,
    profile: str,
) -> tuple[int, dict[str, int]]:
    """Ingest a single BEAM chat into Ogham.

    Groups consecutive user+assistant turns into rounds (like LongMemEval).
    Each round becomes one memory with temporal metadata prepended.
    Uses batch embedding at 1000 per call (Voyage max).

    Returns (stored_count, memory_id_to_chat_id mapping).
    """
    from ogham.database import get_backend
    from ogham.embeddings import generate_embeddings_batch

    messages = load_chat(beam_dir, bucket, chat_id)
    topic = load_topic(beam_dir, bucket, chat_id)

    # Group into user-assistant rounds
    rounds = []
    current_round = []
    current_anchor = ""
    for msg in messages:
        if msg["role"] == "user" and current_round:
            # Previous round complete, start new one
            rounds.append((current_anchor, current_round))
            current_round = []
        current_anchor = msg.get("_time_anchor") or current_anchor
        current_round.append(msg)
    if current_round:
        rounds.append((current_anchor, current_round))

    if not rounds:
        return 0, {}

    # Build content for each round
    all_rows = []
    for time_anchor, round_msgs in rounds:
        parts = []
        if time_anchor:
            parts.append(f"[Date: {time_anchor}]")
        for msg in round_msgs:
            role = "User" if msg["role"] == "user" else "Assistant"
            parts.append(f"{role}: {msg['content']}")
        content = "\n".join(parts)

        # Truncate very long rounds (some assistant responses are huge)
        if len(content) > 10000:
            content = content[:10000] + "..."

        msg_ids = [m.get("id", -1) for m in round_msgs]
        tags = [
            f"chat:{chat_id}",
            f"bucket:{bucket}",
            f"topic:{topic.get('category', 'unknown')}",
        ]
        if time_anchor:
            tags.append(f"date:{time_anchor}")
        meta = {
            "chat_id": chat_id,
            "bucket": bucket,
            "time_anchor": time_anchor,
            "msg_ids": msg_ids,
            "topic": topic.get("title", ""),
        }
        all_rows.append((content, tags, meta))

    # Batch embed all rounds (uses provider default batch size if None)
    all_texts = [r[0] for r in all_rows]
    batch_label = EMBEDDING_BATCH_SIZE if EMBEDDING_BATCH_SIZE else "provider default"
    logger.info("  Embedding %d rounds (batch=%s)...", len(all_texts), batch_label)
    embeddings = _with_retry(generate_embeddings_batch, all_texts, batch_size=EMBEDDING_BATCH_SIZE)

    # Batch insert into database
    backend = get_backend()
    memory_to_chat: dict[str, int] = {}
    stored = 0
    insert_batch = 100

    for start in range(0, len(all_rows), insert_batch):
        end = min(start + insert_batch, len(all_rows))
        batch_rows = []
        for idx in range(start, end):
            content, tags, meta = all_rows[idx]
            batch_rows.append(
                {
                    "content": content,
                    "embedding": str(embeddings[idx]),
                    "profile": profile,
                    "source": "beam",
                    "tags": tags,
                    "metadata": meta,
                }
            )

        try:
            results = _with_retry(backend.store_memories_batch, batch_rows)
            for r in results:
                mem_id = r.get("id", "")
                memory_to_chat[mem_id] = chat_id
            stored += len(results)
        except Exception as e:
            logger.warning("  Batch insert failed at %d-%d: %s", start, end, e)

    return stored, memory_to_chat


def ingest_bucket(
    beam_dir: Path,
    bucket: str,
    chat_ids: list[int] | None = None,
) -> dict:
    """Ingest all (or selected) chats in a bucket.

    Each chat gets its own profile (beam_100K_1, beam_100K_2, etc.)
    so questions can be evaluated per-chat with isolated retrieval.
    """
    available = list_chats(beam_dir, bucket)
    if chat_ids:
        targets = [c for c in chat_ids if c in available]
    else:
        targets = available

    logger.info("Ingesting %d chats from bucket %s", len(targets), bucket)
    stats = {"bucket": bucket, "chats": {}}
    total_stored = 0
    start = time.time()

    for chat_id in targets:
        profile = f"beam_{bucket}_{chat_id}"
        logger.info("Chat %d → profile %s", chat_id, profile)
        try:
            count, mem_map = ingest_chat(beam_dir, bucket, chat_id, profile)
            stats["chats"][chat_id] = {"stored": count, "profile": profile}
            total_stored += count
            logger.info("  Stored %d memories", count)
        except Exception as e:
            logger.error("  Failed: %s", e)
            stats["chats"][chat_id] = {"error": str(e)}

    elapsed = time.time() - start
    stats["total_stored"] = total_stored
    stats["elapsed_seconds"] = round(elapsed, 1)
    logger.info("Ingestion complete: %d memories in %.0fs", total_stored, elapsed)

    # Save ingestion stats
    RESULTS_DIR.mkdir(exist_ok=True)
    stats_file = RESULTS_DIR / f"ingest_{bucket}.json"
    with open(stats_file, "w") as f:
        json.dump(stats, f, indent=2)
    logger.info("Stats saved to %s", stats_file)

    return stats


# ---------------------------------------------------------------------------
# Evaluation -- probing questions against ingested memories
# ---------------------------------------------------------------------------


def evaluate_question(
    question: dict,
    category: str,
    chat_id: int,
    profile: str,
    query_embedding: list[float],
    top_k: int = 10,
) -> dict:
    """Evaluate a single probing question against stored memories.

    Returns metrics dict with recall, MRR, NDCG, plus question metadata.
    """
    from ogham.service import search_memories_enriched

    query = question["question"]
    results = _with_retry(
        search_memories_enriched,
        query=query,
        profile=profile,
        limit=max(top_k, 50),
        embedding=query_embedding,
        graph_depth=BEAM_GRAPH_DEPTH,
    )

    # Gold answer: source_chat_ids field tells us which chat IDs contain the answer
    # For single-chat eval, the gold is the current chat_id if it has source_chat_ids
    gold_chat_ids = set()
    source_ids = question.get("source_chat_ids", {})
    if isinstance(source_ids, dict):
        for key, ids in source_ids.items():
            if isinstance(ids, list):
                gold_chat_ids.update(str(i) for i in ids)
    elif isinstance(source_ids, list):
        gold_chat_ids.update(str(i) for i in source_ids)

    # Map results back to message IDs from source_chat_ids
    # For single-chat evaluation, check if retrieved memories contain the gold msg IDs
    retrieved_msg_ids = []
    for r in results:
        meta = r.get("metadata", {})
        msg_ids = meta.get("msg_ids", [])
        retrieved_msg_ids.extend(str(m) for m in msg_ids)

    # Compute hit: did we find any memory whose msg_ids overlap with gold?
    gold_msg_set = gold_chat_ids  # These are actually message IDs in BEAM

    # Also check chat-level: did we get results from the right chat?
    retrieved_chats = set()
    for r in results:
        tags = r.get("tags", [])
        for tag in tags:
            if tag.startswith("chat:"):
                retrieved_chats.add(tag[5:])

    # Compute metrics at message-ID level
    hit_at = {}
    for k in [5, 10, 20, 30, 50]:
        top_k_msgs = set()
        for r in results[:k]:
            meta = r.get("metadata", {})
            top_k_msgs.update(str(m) for m in meta.get("msg_ids", []))
        if gold_msg_set:
            hit_at[k] = len(gold_msg_set & top_k_msgs) / len(gold_msg_set)
        else:
            hit_at[k] = 1.0 if category in ("abstention", "summarization") else 0.0

    # For categories where empty gold = correct by definition (abstention,
    # summarization with no source), treat all metrics as perfect.
    no_gold_is_correct = not gold_msg_set and category in ("abstention", "summarization")

    # MRR: rank of first result containing a gold msg ID
    if no_gold_is_correct:
        mrr = 1.0
    else:
        mrr = 0.0
        for i, r in enumerate(results):
            meta = r.get("metadata", {})
            r_msgs = set(str(m) for m in meta.get("msg_ids", []))
            if r_msgs & gold_msg_set:
                mrr = 1.0 / (i + 1)
                break

    # NDCG@10
    if no_gold_is_correct:
        ndcg = 1.0
    else:
        dcg = 0.0
        for i, r in enumerate(results[:10]):
            meta = r.get("metadata", {})
            r_msgs = set(str(m) for m in meta.get("msg_ids", []))
            if r_msgs & gold_msg_set:
                dcg += 1.0 / math.log2(i + 2)
        idcg = sum(1.0 / math.log2(i + 2) for i in range(min(len(gold_msg_set), 10)))
        ndcg = dcg / idcg if idcg > 0 else 0.0

    # Kendall tau-b for event_ordering (BEAM paper specifies this metric).
    # For all other categories, tau_b is None (not applicable).
    tau_b = None
    if category == "event_ordering":
        from beam_scoring import score_event_ordering

        ordering_tested = question.get("ordering_tested", [])
        if ordering_tested:
            tau_result = score_event_ordering(ordering_tested, results)
            tau_b = tau_result["kendall_tau_b"]

    metrics = {
        "chat_id": chat_id,
        "category": category,
        "question": query[:120],
        "difficulty": question.get("difficulty", "unknown"),
        "recall@5": hit_at.get(5, 0.0),
        "recall@10": hit_at.get(10, 0.0),
        "recall@20": hit_at.get(20, 0.0),
        "recall@30": hit_at.get(30, 0.0),
        "recall@50": hit_at.get(50, 0.0),
        "ndcg@10": ndcg,
        "mrr": mrr,
        "gold_ids": sorted(gold_msg_set),
        "gold_count": len(gold_msg_set),
        "retrieved_count": len(results),
    }
    if tau_b is not None:
        metrics["kendall_tau_b"] = tau_b
    return metrics


def evaluate_bucket(
    beam_dir: Path,
    bucket: str,
    chat_ids: list[int] | None = None,
    categories: list[str] | None = None,
    top_k: int = 10,
) -> dict:
    """Run evaluation across all chats and categories in a bucket.

    Pre-batches all query embeddings for efficiency.
    """
    from ogham.embeddings import generate_embeddings_batch

    available = list_chats(beam_dir, bucket)
    if chat_ids:
        targets = [c for c in chat_ids if c in available]
    else:
        targets = available

    if categories:
        eval_categories = [c for c in categories if c in ALL_CATEGORIES]
    else:
        eval_categories = ALL_CATEGORIES

    # Collect all questions across chats
    all_questions = []  # (chat_id, category, question_dict)
    for chat_id in targets:
        questions = load_probing_questions(beam_dir, bucket, chat_id)
        for cat in eval_categories:
            for q in questions.get(cat, []):
                all_questions.append((chat_id, cat, q))

    logger.info(
        "Evaluating %d questions across %d chats, categories: %s",
        len(all_questions),
        len(targets),
        eval_categories,
    )

    if not all_questions:
        logger.warning("No questions found")
        return {}

    # Pre-batch all query embeddings (provider default batch size)
    query_texts = [q[2]["question"] for q in all_questions]
    batch_label = EMBEDDING_BATCH_SIZE if EMBEDDING_BATCH_SIZE else "provider default"
    logger.info("Pre-embedding %d queries (batch=%s)...", len(query_texts), batch_label)
    query_embeddings = _with_retry(
        generate_embeddings_batch, query_texts, batch_size=EMBEDDING_BATCH_SIZE
    )
    logger.info("Query embeddings ready")

    # Evaluate each question
    all_metrics = []
    category_metrics: dict[str, list[dict]] = {}
    start = time.time()

    for i, (chat_id, cat, question) in enumerate(all_questions):
        profile = f"beam_{bucket}_{chat_id}"
        logger.info(
            "[%d/%d] Chat %d %s: %s",
            i + 1,
            len(all_questions),
            chat_id,
            cat,
            question["question"][:60],
        )

        try:
            metrics = evaluate_question(question, cat, chat_id, profile, query_embeddings[i], top_k)
            all_metrics.append(metrics)
            category_metrics.setdefault(cat, []).append(metrics)
            logger.info(
                "  R@5=%.2f R@10=%.2f MRR=%.2f",
                metrics["recall@5"],
                metrics["recall@10"],
                metrics["mrr"],
            )
        except Exception as e:
            logger.error("  Failed: %s", e)

    elapsed = time.time() - start

    # Aggregate results
    if not all_metrics:
        print("No results to report.")
        return {}

    avg = {
        "recall@5": sum(m["recall@5"] for m in all_metrics) / len(all_metrics),
        "recall@10": sum(m["recall@10"] for m in all_metrics) / len(all_metrics),
        "recall@20": sum(m["recall@20"] for m in all_metrics) / len(all_metrics),
        "recall@30": sum(m["recall@30"] for m in all_metrics) / len(all_metrics),
        "recall@50": sum(m["recall@50"] for m in all_metrics) / len(all_metrics),
        "ndcg@10": sum(m["ndcg@10"] for m in all_metrics) / len(all_metrics),
        "mrr": sum(m["mrr"] for m in all_metrics) / len(all_metrics),
    }

    cat_avgs = {}
    for cat, mlist in sorted(category_metrics.items()):
        cat_avgs[cat] = {
            "recall@5": sum(m["recall@5"] for m in mlist) / len(mlist),
            "recall@10": sum(m["recall@10"] for m in mlist) / len(mlist),
            "mrr": sum(m["mrr"] for m in mlist) / len(mlist),
            "ndcg@10": sum(m["ndcg@10"] for m in mlist) / len(mlist),
            "count": len(mlist),
        }

    # Difficulty breakdown
    diff_metrics: dict[str, list[dict]] = {}
    for m in all_metrics:
        diff_metrics.setdefault(m["difficulty"], []).append(m)
    diff_avgs = {}
    for diff, mlist in sorted(diff_metrics.items()):
        diff_avgs[diff] = {
            "recall@10": sum(m["recall@10"] for m in mlist) / len(mlist),
            "mrr": sum(m["mrr"] for m in mlist) / len(mlist),
            "count": len(mlist),
        }

    results_summary = {
        "bucket": bucket,
        "categories_evaluated": eval_categories,
        "questions_evaluated": len(all_metrics),
        "elapsed_seconds": round(elapsed, 1),
        "overall": avg,
        "per_category": cat_avgs,
        "per_difficulty": diff_avgs,
        "per_question": all_metrics,
    }

    # Print summary
    print("\n" + "=" * 70)
    print(f"BEAM Benchmark Results -- {bucket}")
    print("=" * 70)
    print(f"Questions: {len(all_metrics)}")
    print(f"Time: {elapsed:.0f}s")
    print()
    print(f"  Recall@5:  {avg['recall@5']:.4f}")
    print(f"  Recall@10: {avg['recall@10']:.4f}")
    print(f"  Recall@20: {avg['recall@20']:.4f}")
    print(f"  Recall@30: {avg['recall@30']:.4f}")
    print(f"  Recall@50: {avg['recall@50']:.4f}")
    print(f"  NDCG@10:   {avg['ndcg@10']:.4f}")
    print(f"  MRR:       {avg['mrr']:.4f}")
    print()
    print("Per category:")
    for cat, cavg in sorted(cat_avgs.items()):
        print(
            f"  {cat:30s}  R@10={cavg['recall@10']:.4f}  MRR={cavg['mrr']:.4f}  (n={cavg['count']})"
        )
    print()
    print("Per difficulty:")
    for diff, davg in sorted(diff_avgs.items()):
        print(
            f"  {diff:15s}  R@10={davg['recall@10']:.4f}"
            f"  MRR={davg['mrr']:.4f}  (n={davg['count']})"
        )

    # Save results
    RESULTS_DIR.mkdir(exist_ok=True)
    cats_suffix = "_".join(eval_categories) if categories else "all"
    results_file = RESULTS_DIR / f"eval_{bucket}_{cats_suffix}.json"
    with open(results_file, "w") as f:
        json.dump(results_summary, f, indent=2)
    logger.info("Results saved to %s", results_file)

    return results_summary


# ---------------------------------------------------------------------------
# Cleanup
# ---------------------------------------------------------------------------


def cleanup_bucket(bucket: str):
    """Delete all benchmark profiles for a bucket."""
    from ogham.database import get_backend

    prefix = f"beam_{bucket}_"
    backend = get_backend()
    profiles = backend.list_profiles()
    deleted = 0
    for p in profiles:
        name = p.get("profile", "")
        if name.startswith(prefix):
            memories = backend.get_all_memories_content(profile=name)
            for mem in memories:
                try:
                    backend.delete_memory(mem["id"], name)
                except Exception:
                    pass
            deleted += 1
            logger.info("Cleaned up profile: %s", name)
    logger.info("Deleted %d benchmark profiles", deleted)


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def main():
    parser = argparse.ArgumentParser(
        description="BEAM benchmark for Ogham MCP",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Ingest first chat only (quick test):
  %(prog)s --ingest --bucket 100K --chat 1

  # Ingest all 128K chats:
  %(prog)s --ingest --bucket 100K

  # Evaluate all categories:
  %(prog)s --eval --bucket 100K

  # Focus on weak categories:
  %(prog)s --eval --bucket 100K --category temporal_reasoning event_ordering

  # Single chat evaluation:
  %(prog)s --eval --bucket 100K --chat 3

  # List available chats/categories:
  %(prog)s --info --bucket 100K
""",
    )

    # Mode
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--ingest", action="store_true", help="Ingest conversations into Ogham")
    mode.add_argument("--eval", action="store_true", help="Run evaluation on ingested data")
    mode.add_argument("--cleanup", action="store_true", help="Delete benchmark profiles")
    mode.add_argument("--info", action="store_true", help="Show dataset info")

    # Targeting
    parser.add_argument(
        "--bucket",
        default="100K",
        choices=ALL_BUCKETS,
        help="Size bucket (default: 100K)",
    )
    parser.add_argument(
        "--chat",
        type=int,
        nargs="*",
        default=None,
        help="Specific chat ID(s) to process",
    )
    parser.add_argument(
        "--category",
        nargs="*",
        default=None,
        choices=ALL_CATEGORIES,
        help="Filter to specific question categories",
    )
    parser.add_argument("--top-k", type=int, default=10, help="Top K for retrieval (default: 10)")
    parser.add_argument(
        "--beam-dir",
        type=Path,
        default=DEFAULT_BEAM_DIR,
        help="Path to BEAM repo (default: /tmp/BEAM)",
    )

    args = parser.parse_args()

    # Ensure ogham is importable
    src_dir = Path(__file__).parent.parent / "src"
    if str(src_dir) not in sys.path:
        sys.path.insert(0, str(src_dir))

    # Load benchmark env
    env_file = DATA_DIR / ".env.local"
    if env_file.exists():
        logger.info("Loading env from %s", env_file)
        with open(env_file) as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith("#") and "=" in line:
                    key, _, val = line.partition("=")
                    os.environ.setdefault(key.strip(), val.strip())

    # Verify BEAM dir exists
    if not args.beam_dir.exists():
        logger.error(
            "BEAM repo not found at %s. Clone it:\n"
            "  cd /Users/kevinburns/Developer/web-projects && "
            "git clone --depth 1 https://github.com/mohammadtavakoli78/BEAM.git "
            "ogham-beam-dataset",
            args.beam_dir,
        )
        sys.exit(1)

    if args.info:
        chats = list_chats(args.beam_dir, args.bucket)
        print(f"\nBucket: {args.bucket}")
        print(f"Chats: {len(chats)} ({min(chats)}-{max(chats)})")
        total_qs = 0
        for chat_id in chats:
            qs = load_probing_questions(args.beam_dir, args.bucket, chat_id)
            q_count = sum(len(v) for v in qs.values())
            total_qs += q_count
            cats = {k: len(v) for k, v in qs.items() if v}
            print(f"  Chat {chat_id:3d}: {q_count} questions  {cats}")
        print(f"\nTotal: {total_qs} questions across {len(ALL_CATEGORIES)} categories")
        return

    if args.ingest:
        ingest_bucket(args.beam_dir, args.bucket, args.chat)
        return

    if args.eval:
        evaluate_bucket(args.beam_dir, args.bucket, args.chat, args.category, args.top_k)
        return

    if args.cleanup:
        cleanup_bucket(args.bucket)
        return


if __name__ == "__main__":
    main()
