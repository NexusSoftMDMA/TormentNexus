#!/usr/bin/env python3
"""LongMemEval benchmark for Ogham MCP.

Evaluates long-term memory retrieval using the LongMemEval dataset
(Wu et al., ICLR 2025) -- 500 questions testing 5 memory abilities:
  - Information Extraction (user/assistant/preference)
  - Multi-Session Reasoning
  - Temporal Reasoning
  - Knowledge Updates
  - Abstention (false premises)

Metrics:
  - Recall@K: fraction of gold sessions found in top K results
  - NDCG@K: normalized discounted cumulative gain
  - MRR: mean reciprocal rank of first relevant session

Usage:
    # Download dataset:
    uv run python3 benchmarks/longmemeval_benchmark.py --download

    # Run retrieval-only evaluation (free, no LLM calls):
    uv run python3 benchmarks/longmemeval_benchmark.py --top-k 10

    # Run single question for debugging:
    uv run python3 benchmarks/longmemeval_benchmark.py --question-id 42

    # Clean up profiles after:
    uv run python3 benchmarks/longmemeval_benchmark.py --cleanup

Requires: Ogham MCP server configured with a working database and embedding provider.
Cost: Retrieval-only is free (embedding costs only). Full QA adds ~$4 in LLM calls.
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

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s")
logger = logging.getLogger(__name__)

DATA_DIR = Path(__file__).parent
RESULTS_FILE = DATA_DIR / "longmemeval_results.json"
DATASET_URL = "https://huggingface.co/datasets/xiaowu0162/longmemeval-cleaned/resolve/main"


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


def download_dataset(variant: str = "S"):
    """Download LongMemEval dataset from HuggingFace."""
    import urllib.request

    filename = f"longmemeval_{variant}.json"
    filepath = DATA_DIR / filename

    if filepath.exists():
        logger.info("Dataset already exists: %s", filepath)
        return filepath

    url = f"{DATASET_URL}/{filename}"
    logger.info("Downloading %s ...", url)
    urllib.request.urlretrieve(url, filepath)
    logger.info("Saved to %s", filepath)
    return filepath


def load_dataset(variant: str = "S") -> list[dict]:
    """Load the LongMemEval dataset."""
    filepath = DATA_DIR / f"longmemeval_{variant}.json"
    if not filepath.exists():
        raise FileNotFoundError(f"Dataset not found: {filepath}. Run with --download first.")
    with open(filepath) as f:
        return json.load(f)


def _expected_memory_count(question: dict) -> int:
    """Approximate round count the ingest would produce for this question.

    Each session turn maps to one memory. A small tolerance is OK -- some
    sessions have trailing assistant-only turns that get dropped. We use
    this only to warn on profile pollution, not to force a count match.
    """
    total = 0
    for session in question.get("haystack_sessions", []):
        total += len(session) if isinstance(session, list) else 0
    return total


def _existing_profile_mem_map(backend, profile: str) -> dict[str, str]:
    """Build memory_id -> session_id map from memories already in the profile.

    Used when a previous run already ingested this profile so we can skip
    re-embedding (expensive + hits Gemini / OpenAI rate limits for no gain).
    Session IDs are stored verbatim as "session:<id>" tags on each memory.
    """
    try:
        rows = backend.get_all_memories_full(profile)
    except Exception as exc:
        logger.debug("Could not read existing profile %s: %s", profile, exc)
        return {}

    mem_map: dict[str, str] = {}
    for row in rows:
        mem_id = row.get("id")
        if not mem_id:
            continue
        for tag in row.get("tags") or []:
            if isinstance(tag, str) and tag.startswith("session:"):
                mem_map[str(mem_id)] = tag[len("session:") :]
                break
    return mem_map


def ingest_question(question: dict, profile: str):
    """Ingest all sessions for a single question into Ogham.

    Uses batch embedding (up to 1000 at a time) for efficiency.
    Each session is decomposed into user-assistant round pairs.
    Temporal metadata is prepended to each round.

    Idempotent: if the profile already has memories for this question, the
    existing memory_id -> session_id map is returned and ingestion is skipped.
    Skips any round where the embedding provider returned None (previously
    serialised "None" into the SQL INSERT and broke the whole batch).
    """
    from ogham.database import get_backend
    from ogham.embeddings import generate_embeddings_batch

    backend = get_backend()

    # If --fresh was requested, wipe any prior state before checking
    # idempotency so we get a genuinely clean run. Prior pollution from
    # interrupted benchmarks otherwise silently inflates retrieval-count
    # metrics (a question's profile ends up with 2x the intended memories,
    # and gold sessions get outranked by duplicates).
    if os.environ.get("OGHAM_LME_FRESH") == "1":
        backend._execute(
            "DELETE FROM memories WHERE profile = %(p)s",
            {"p": profile},
            fetch="none",
        )

    # Idempotency: if profile already populated, skip re-embed / re-insert.
    # Warn (not fail) if the existing count differs from what this question
    # expects -- pollution detection without forcing a re-ingest unless the
    # operator opts in via --fresh.
    existing_map = _existing_profile_mem_map(backend, profile)
    if existing_map:
        expected = _expected_memory_count(question)
        if expected and abs(len(existing_map) - expected) > max(5, expected // 10):
            logger.warning(
                "profile %s has %d memories but question expects ~%d -- "
                "rerun with --fresh to wipe and re-ingest",
                profile,
                len(existing_map),
                expected,
            )
        return len(existing_map), existing_map

    sessions = question.get("haystack_sessions", [])
    session_ids = question.get("haystack_session_ids", [])
    session_dates = question.get("haystack_dates", [])

    # Phase 1: Build all round content + metadata
    all_rows = []  # (content, tags, metadata, session_id)

    for i, session in enumerate(sessions):
        session_id = session_ids[i] if i < len(session_ids) else f"session_{i}"
        session_date = session_dates[i] if i < len(session_dates) else "unknown"

        rounds = []
        current_round = []
        for turn in session:
            current_round.append(turn)
            if turn["role"] == "assistant":
                rounds.append(current_round)
                current_round = []
        if current_round:
            rounds.append(current_round)

        for j, round_turns in enumerate(rounds):
            content_parts = [f"[Date: {session_date}]"]
            for turn in round_turns:
                role = "User" if turn["role"] == "user" else "Assistant"
                content_parts.append(f"{role}: {turn['content']}")
            content = "\n".join(content_parts)

            if len(content) > 10000:
                content = content[:10000] + "..."

            tags = [
                f"session:{session_id}",
                f"date:{session_date}",
                f"round:{j}",
            ]
            meta = {"session_id": session_id, "date": session_date}
            all_rows.append((content, tags, meta, session_id))

    if not all_rows:
        return 0, {}

    # Phase 2: Batch embed all content
    all_texts = [r[0] for r in all_rows]
    embeddings = _with_retry(generate_embeddings_batch, all_texts)

    # Drop rows where embedding generation returned None (Gemini occasionally
    # fails on pathological content -- empty strings, specific unicode, rate
    # limits that retry gave up on). Previously `str(None)` was serialised
    # into the vector column and SQL rejected the entire batch.
    valid_indices = [idx for idx, emb in enumerate(embeddings) if emb is not None]
    skipped = len(embeddings) - len(valid_indices)
    if skipped:
        logger.warning(
            "  Skipped %d/%d rounds where embedding generation returned None",
            skipped,
            len(embeddings),
        )

    # Phase 3: Batch insert into database
    memory_to_session = {}
    batch_size = 100
    stored_count = 0

    for start in range(0, len(valid_indices), batch_size):
        end = min(start + batch_size, len(valid_indices))
        batch_rows = []
        batch_source_rows = []
        for vidx in valid_indices[start:end]:
            content, tags, meta, session_id = all_rows[vidx]
            batch_rows.append(
                {
                    "content": content,
                    "embedding": str(embeddings[vidx]),
                    "profile": profile,
                    "source": "longmemeval",
                    "tags": tags,
                    "metadata": meta,
                }
            )
            batch_source_rows.append(all_rows[vidx])

        try:
            results = _with_retry(backend.store_memories_batch, batch_rows)
            for r, row_data in zip(results, batch_source_rows):
                mem_id = r.get("id", "")
                memory_to_session[mem_id] = row_data[3]  # session_id
            stored_count += len(results)
        except Exception as e:
            logger.warning("Batch insert failed at %d-%d: %s", start, end, e)

    return stored_count, memory_to_session


def search_question(
    question: dict,
    profile: str,
    top_k: int = 10,
    query_embedding: list[float] | None = None,
) -> list[dict]:
    """Search for the answer to a question using the enriched search pipeline.

    Uses search_memories_enriched which includes temporal re-ranking
    when queries have temporal intent (e.g. "four months ago").
    If query_embedding is provided, skip the embedding call (pre-batched).
    """
    from ogham.service import search_memories_enriched

    query = question["question"]

    results = _with_retry(
        search_memories_enriched,
        query=query,
        profile=profile,
        limit=top_k,
        embedding=query_embedding,
    )

    return results


def compute_retrieval_metrics(
    results: list[dict],
    gold_session_ids: list[str],
    memory_to_session: dict[str, str],
    top_k_values: list[int] = [5, 10, 50],
) -> dict:
    """Compute session-level retrieval metrics."""
    # Map retrieved memories back to session IDs
    retrieved_sessions = []
    for r in results:
        mem_id = r.get("id", "")
        # Check tags for session ID
        tags = r.get("tags", [])
        session_id = None
        for tag in tags:
            if tag.startswith("session:"):
                session_id = tag[8:]
                break
        if session_id is None:
            session_id = memory_to_session.get(mem_id, "unknown")
        retrieved_sessions.append(session_id)

    # Deduplicate while preserving order
    seen = set()
    unique_sessions = []
    for s in retrieved_sessions:
        if s not in seen:
            seen.add(s)
            unique_sessions.append(s)

    gold_set = set(gold_session_ids)
    metrics = {}

    # Recall@K
    for k in top_k_values:
        top_k_sessions = set(unique_sessions[:k])
        if gold_set:
            recall = len(gold_set & top_k_sessions) / len(gold_set)
        else:
            recall = 0.0
        metrics[f"recall@{k}"] = recall

    # MRR
    mrr = 0.0
    for i, session_id in enumerate(unique_sessions):
        if session_id in gold_set:
            mrr = 1.0 / (i + 1)
            break
    metrics["mrr"] = mrr

    # NDCG@10
    k = 10
    dcg = 0.0
    for i, session_id in enumerate(unique_sessions[:k]):
        if session_id in gold_set:
            dcg += 1.0 / math.log2(i + 2)
    idcg = sum(1.0 / math.log2(i + 2) for i in range(min(len(gold_set), k)))
    metrics["ndcg@10"] = dcg / idcg if idcg > 0 else 0.0

    return metrics


def cleanup_profiles(prefix: str = "lme_"):
    """Delete all benchmark profiles."""
    from ogham.database import get_backend

    backend = get_backend()
    profiles = backend.list_profiles()
    deleted = 0
    for p in profiles:
        name = p.get("profile", "")
        if name.startswith(prefix):
            # Delete all memories in the profile
            memories = backend.get_all_memories_content(profile=name)
            for mem in memories:
                try:
                    backend.delete_memory(mem["id"], name)
                except Exception:
                    pass
            deleted += 1
            logger.info("Cleaned up profile: %s", name)
    logger.info("Deleted %d benchmark profiles", deleted)


def run_benchmark(
    variant: str = "S",
    top_k: int = 10,
    max_questions: int | None = None,
    question_id: int | None = None,
    question_type: str | None = None,
):
    """Run the LongMemEval retrieval benchmark."""
    dataset = load_dataset(variant)
    logger.info("Loaded %d questions (variant %s)", len(dataset), variant)

    if question_id is not None:
        dataset = [q for q in dataset if q.get("question_id") == question_id]
        if not dataset:
            logger.error("Question ID %d not found", question_id)
            return
        logger.info("Running single question: %d", question_id)

    if question_type:
        dataset = [q for q in dataset if q.get("question_type") == question_type]
        logger.info("Filtered to %d questions of type '%s'", len(dataset), question_type)

    if max_questions:
        dataset = dataset[:max_questions]
        logger.info("Limited to %d questions", max_questions)

    # Skip abstention questions for retrieval eval (no gold sessions)
    retrieval_questions = [
        q for q in dataset if q.get("question_type") != "abstention" and q.get("answer_session_ids")
    ]
    logger.info("Evaluating %d retrieval questions (skipped abstention)", len(retrieval_questions))

    # Pre-batch all query embeddings in one API call
    from ogham.embeddings import generate_embeddings_batch

    query_texts = [q["question"] for q in retrieval_questions]
    logger.info("Pre-embedding %d queries in batch...", len(query_texts))
    query_embeddings = _with_retry(generate_embeddings_batch, query_texts)
    logger.info("Query embeddings ready (%d vectors)", len(query_embeddings))

    all_metrics = []
    category_metrics: dict[str, list[dict]] = {}
    total_ingested = 0
    start_time = time.time()

    for i, question in enumerate(retrieval_questions):
        qid = question.get("question_id", i)
        qtype = question.get("question_type", "unknown")
        profile = f"lme_{qid}"

        logger.info(
            "[%d/%d] Q%s (%s): %s",
            i + 1,
            len(retrieval_questions),
            qid,
            qtype,
            question["question"][:60],
        )

        # Ingest sessions for this question
        try:
            count, mem_map = ingest_question(question, profile)
            total_ingested += count
            logger.info("  Ingested %d memories", count)
        except Exception as e:
            logger.error("  Ingestion failed: %s", e)
            continue

        # Search (using pre-batched query embedding)
        try:
            results = search_question(
                question,
                profile,
                top_k=max(top_k, 50),
                query_embedding=query_embeddings[i],
            )
        except Exception as e:
            logger.error("  Search failed: %s", e)
            continue

        # Compute metrics
        gold_sessions = question.get("answer_session_ids", [])
        metrics = compute_retrieval_metrics(
            results, gold_sessions, mem_map, top_k_values=[5, 10, 50]
        )
        metrics["question_id"] = qid
        metrics["question_type"] = qtype
        all_metrics.append(metrics)

        if qtype not in category_metrics:
            category_metrics[qtype] = []
        category_metrics[qtype].append(metrics)

        logger.info(
            "  R@5=%.2f R@10=%.2f MRR=%.2f",
            metrics["recall@5"],
            metrics["recall@10"],
            metrics["mrr"],
        )

    elapsed = time.time() - start_time

    # Aggregate
    if all_metrics:
        avg = {
            "recall@5": sum(m["recall@5"] for m in all_metrics) / len(all_metrics),
            "recall@10": sum(m["recall@10"] for m in all_metrics) / len(all_metrics),
            "recall@50": sum(m["recall@50"] for m in all_metrics) / len(all_metrics),
            "ndcg@10": sum(m["ndcg@10"] for m in all_metrics) / len(all_metrics),
            "mrr": sum(m["mrr"] for m in all_metrics) / len(all_metrics),
        }

        # Per-category averages
        cat_avgs = {}
        for cat, mlist in category_metrics.items():
            cat_avgs[cat] = {
                "recall@10": sum(m["recall@10"] for m in mlist) / len(mlist),
                "mrr": sum(m["mrr"] for m in mlist) / len(mlist),
                "count": len(mlist),
            }

        results_summary = {
            "variant": variant,
            "questions_evaluated": len(all_metrics),
            "total_ingested": total_ingested,
            "elapsed_seconds": round(elapsed, 1),
            "overall": avg,
            "per_category": cat_avgs,
            "per_question": all_metrics,
        }

        # Print summary
        print("\n" + "=" * 60)
        print("LongMemEval Results")
        print("=" * 60)
        print(f"Questions: {len(all_metrics)}")
        print(f"Total memories ingested: {total_ingested}")
        print(f"Time: {elapsed:.0f}s")
        print()
        print(f"  Recall@5:  {avg['recall@5']:.4f}")
        print(f"  Recall@10: {avg['recall@10']:.4f}")
        print(f"  Recall@50: {avg['recall@50']:.4f}")
        print(f"  NDCG@10:   {avg['ndcg@10']:.4f}")
        print(f"  MRR:       {avg['mrr']:.4f}")
        print()
        print("Per category:")
        for cat, cavg in sorted(cat_avgs.items()):
            print(
                f"  {cat:30s}  R@10={cavg['recall@10']:.4f}"
                f"  MRR={cavg['mrr']:.4f}  (n={cavg['count']})"
            )

        # Save results
        with open(RESULTS_FILE, "w") as f:
            json.dump(results_summary, f, indent=2)
        logger.info("Results saved to %s", RESULTS_FILE)
    else:
        print("No results to report.")


def main():
    parser = argparse.ArgumentParser(description="LongMemEval benchmark for Ogham MCP")
    parser.add_argument("--download", action="store_true", help="Download the dataset")
    parser.add_argument(
        "--variant", default="S", choices=["oracle", "S", "M"], help="Dataset variant"
    )
    parser.add_argument("--top-k", type=int, default=10, help="Top K for retrieval")
    parser.add_argument("--max-questions", type=int, default=None, help="Limit number of questions")
    parser.add_argument("--question-id", type=int, default=None, help="Run single question by ID")
    parser.add_argument(
        "--question-type",
        type=str,
        default=None,
        help="Filter by question type (e.g. temporal-reasoning, multi-session)",
    )
    parser.add_argument("--cleanup", action="store_true", help="Delete benchmark profiles")
    parser.add_argument(
        "--fresh",
        action="store_true",
        help="Wipe each profile before ingesting (forces clean re-embed; slower "
        "but the only way to get trustworthy retrieval metrics after a crashed run)",
    )
    args = parser.parse_args()

    if args.fresh:
        os.environ["OGHAM_LME_FRESH"] = "1"

    # Ensure ogham is importable
    src_dir = Path(__file__).parent.parent / "src"
    if str(src_dir) not in sys.path:
        sys.path.insert(0, str(src_dir))

    # Load benchmark env (local Postgres, not Supabase!)
    env_file = DATA_DIR / ".env.local"
    if env_file.exists():
        logger.info("Loading env from %s", env_file)
        with open(env_file) as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith("#") and "=" in line:
                    key, _, val = line.partition("=")
                    os.environ[key.strip()] = val.strip()
    else:
        logger.warning(
            "No .env.local found -- using default env. "
            "Create benchmarks/.env.local with DATABASE_BACKEND=postgres "
            "to avoid exhausting Supabase free tier!"
        )

    if args.download:
        download_dataset(args.variant)
        return

    if args.cleanup:
        cleanup_profiles()
        return

    run_benchmark(
        variant=args.variant,
        top_k=args.top_k,
        max_questions=args.max_questions,
        question_id=args.question_id,
        question_type=args.question_type,
    )


if __name__ == "__main__":
    main()
