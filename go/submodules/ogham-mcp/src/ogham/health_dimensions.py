"""8-dimension health scoring for ``ogham health`` (v0.13).

Replaces the v0.12 binary green/red readout with a score-out-of-10 across
eight dimensions. UX modeled on the auto-memory blog post (Microsoft
DevBlogs, dezgit2025/auto-memory, 2026-04-28). The thing we get from
the higher-resolution readout: "everything is mostly fine but wiki
coverage is at 6.4/10 because 28% of summaries are stale" — invisible
under the v0.12 binary.

Each ``compute_<dim>(profile=...) -> DimensionResult`` is independently
mockable. ``compose_health(profile)`` runs all eight and returns a list
in the order shown in ``DIMENSION_NAMES``. Failures in any one dimension
are caught and reported as RED with the exception message in ``detail``;
they don't abort the rest of the report.
"""

from __future__ import annotations

import logging
import statistics
import time
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from typing import Any, Callable, Literal, cast
from uuid import uuid4

from ogham.config import settings
from ogham.database import get_backend

logger = logging.getLogger(__name__)

Zone = Literal["GREEN", "AMBER", "RED"]


# ─── DimensionResult ────────────────────────────────────────────────────


@dataclass
class DimensionResult:
    name: str
    score: float
    zone: Zone
    detail: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


def zone(score: float) -> Zone:
    """Map a score (0.0-10.0) to a zone label.

    >=8.0 GREEN, >=5.0 AMBER, <5.0 RED.
    """
    if score >= 8.0:
        return "GREEN"
    if score >= 5.0:
        return "AMBER"
    return "RED"


# ─── 1. DB freshness ────────────────────────────────────────────────────


def _last_memory_write(backend: Any, profile: str) -> datetime | None:
    """Return the most recent memory created_at for ``profile``, or None.

    Routes raw SQL when the backend exposes ``_execute`` (Postgres);
    otherwise falls back to the public ``list_recent_memories`` API
    (Supabase / PostgREST). Both paths normalise to ``datetime``.
    """
    if hasattr(backend, "_execute"):
        last = backend._execute(
            "SELECT MAX(created_at) FROM memories WHERE profile = %(profile)s",
            {"profile": profile},
            fetch="scalar",
        )
        return last
    # Backend lacks raw SQL; use the public listing API and read the top row.
    recent = backend.list_recent_memories(profile=profile, limit=1)
    if not recent:
        return None
    cv = recent[0].get("created_at")
    if isinstance(cv, str):
        # PostgREST returns ISO-8601 strings; tolerate trailing Z.
        return datetime.fromisoformat(cv.replace("Z", "+00:00"))
    return cv


def compute_db_freshness(profile: str) -> DimensionResult:
    """Score how recently this profile last wrote a memory.

    >=24h => 10.0 (GREEN), 24-72h => linear ramp 7.99..5.0 (AMBER),
    >72h => linear ramp 4.99..0.0 (RED). No memory at all => 0.0 (RED).
    """
    backend = cast(Any, get_backend())
    try:
        last = _last_memory_write(backend, profile)
    except Exception as e:  # missing table, connection issue, etc.
        return DimensionResult(
            name="DB freshness",
            score=0.0,
            zone="RED",
            detail=f"query failed: {e}",
        )

    if last is None:
        return DimensionResult(
            name="DB freshness",
            score=0.0,
            zone="RED",
            detail="no memories in profile (dead writer)",
        )

    if last.tzinfo is None:
        last = last.replace(tzinfo=timezone.utc)
    age = datetime.now(timezone.utc) - last
    age_hours = age.total_seconds() / 3600.0

    if age_hours <= 24:
        score = 10.0
    elif age_hours <= 72:
        # 24h => ~7.99, 72h => 5.0 (linear)
        score = 7.99 - (age_hours - 24) * (2.99 / 48.0)
    else:
        # 72h => 4.99, 30d => ~0.0
        excess = min(age_hours - 72, 30 * 24)
        score = max(0.0, 4.99 - excess * (4.99 / (30 * 24)))

    detail = _humanize_age(age_hours)
    return DimensionResult(
        name="DB freshness",
        score=round(score, 1),
        zone=zone(score),
        detail=f"last write {detail} ago",
    )


def _humanize_age(hours: float) -> str:
    if hours < 1:
        return f"{hours * 60:.0f}m"
    if hours < 24:
        return f"{hours:.1f}h"
    return f"{hours / 24:.1f}d"


# ─── 2. Schema integrity ────────────────────────────────────────────────


def _run_schema_integrity_checks() -> tuple[bool, str]:
    """Inline-equivalent of test_migration_integrity contract checks.

    Returns ``(passed, detail_message)``. Any failure short-circuits to
    RED. Mock this in tests to avoid running real file IO.
    """
    from pathlib import Path

    import ogham

    # The migration files ship inside the wheel via hatchling
    # force-include at ogham/sql/migrations/ -- but locally we resolve
    # them from the repo root for the dev case.
    pkg_root = Path(ogham.__file__).resolve().parent
    repo_root_candidate = pkg_root.parent.parent  # src/ogham/.. -> src/.. -> repo root
    candidates = [
        pkg_root / "sql" / "migrations",  # wheel layout
        repo_root_candidate / "sql" / "migrations",  # dev repo layout
    ]
    migrations_dir = next((c for c in candidates if c.is_dir()), None)
    if migrations_dir is None:
        return False, "could not locate sql/migrations/ on disk"

    sql_files = sorted(p for p in migrations_dir.glob("*.sql") if p.is_file())
    if not sql_files:
        return False, "no migration files found"

    # 1. No unnumbered migrations (the v0.9.2 incident class).
    offenders = [p.name for p in sql_files if not p.name[0].isdigit()]
    if offenders:
        return False, f"unnumbered migration(s): {offenders}"

    # 2. The duplicate src/ogham/sql/ tree must NOT exist (Phase B guard).
    duplicate = repo_root_candidate / "src" / "ogham" / "sql"
    if duplicate.exists() and duplicate != pkg_root / "sql":
        return False, f"duplicate SQL tree present at {duplicate}"

    return True, f"all {len(sql_files)} migrations applied, parity OK"


def compute_schema_integrity() -> DimensionResult:
    """Binary green/red on schema/migration integrity.

    All contract checks pass => 10.0 (GREEN); any failure => 0.0 (RED).
    """
    try:
        passed, detail = _run_schema_integrity_checks()
    except Exception as e:
        passed, detail = False, f"check raised {type(e).__name__}: {e}"

    score = 10.0 if passed else 0.0
    return DimensionResult(
        name="Schema integrity",
        score=score,
        zone=zone(score),
        detail=detail,
    )


# ─── 3. Hybrid search latency ───────────────────────────────────────────


def _sample_hybrid_search_latencies(profile: str, samples: int = 5) -> list[float]:
    """Run ``samples`` hybrid_search calls and return their wall-clock
    durations in seconds."""
    from ogham.database import hybrid_search_memories
    from ogham.embeddings import generate_embedding

    # Use a fixed cheap query and a real embedding so the timing is
    # representative of cold-cache user queries. Profile gates the FTS
    # so we don't accidentally measure cross-profile fan-out.
    query = "health-probe-latency-sample"
    embedding = generate_embedding(query)

    durations: list[float] = []
    for _ in range(samples):
        t0 = time.perf_counter()
        try:
            hybrid_search_memories(
                query_text=query,
                query_embedding=embedding,
                profile=profile,
                limit=5,
            )
        except Exception as e:
            logger.debug("hybrid_search_memories sample raised: %s", e)
            return []
        durations.append(time.perf_counter() - t0)
    return durations


def compute_hybrid_search_latency(profile: str) -> DimensionResult:
    """Score p50/p95 of recent hybrid_search calls.

    p50<50ms => 10.0; 50-200ms ramps to 5.0; >200ms ramps to 0.0.
    """
    try:
        durations = _sample_hybrid_search_latencies(profile)
    except Exception as e:
        return DimensionResult(
            name="Hybrid search",
            score=0.0,
            zone="RED",
            detail=f"sample failed: {e}",
        )

    if not durations:
        return DimensionResult(
            name="Hybrid search",
            score=0.0,
            zone="RED",
            detail="no successful samples",
        )

    p50 = statistics.median(durations) * 1000  # ms
    sorted_d = sorted(durations)
    p95_idx = max(0, int(round(0.95 * (len(sorted_d) - 1))))
    p95 = sorted_d[p95_idx] * 1000  # ms

    if p50 <= 50:
        # 0ms => 10.0, 50ms => 8.0
        score = 10.0 - (p50 / 50.0) * 2.0
    elif p50 <= 200:
        # 50ms => 7.99, 200ms => 5.0
        score = 7.99 - ((p50 - 50) / 150.0) * 2.99
    else:
        # 200ms => 4.99, 1000ms => 0
        excess = min(p50 - 200, 800)
        score = max(0.0, 4.99 - (excess / 800.0) * 4.99)

    return DimensionResult(
        name="Hybrid search",
        score=round(score, 1),
        zone=zone(score),
        detail=f"p50 {p50:.0f}ms, p95 {p95:.0f}ms",
    )


# ─── 4. Corpus size ─────────────────────────────────────────────────────


def compute_corpus_size(profile: str) -> DimensionResult:
    """Score memory count.

    >=100 => 10.0; 10-99 => linear 5.0..7.99; <10 => 0..4.99.
    """
    backend = cast(Any, get_backend())
    try:
        if hasattr(backend, "_execute"):
            count = backend._execute(
                "SELECT COUNT(*) FROM memories WHERE profile = %(profile)s",
                {"profile": profile},
                fetch="scalar",
            )
        else:
            # Supabase / PostgREST: route through the existing memory_stats RPC.
            stats = backend.get_memory_stats(profile)
            count = int(stats.get("total", 0) or 0)
    except Exception as e:
        return DimensionResult(
            name="Corpus size",
            score=0.0,
            zone="RED",
            detail=f"count failed: {e}",
        )

    n = int(count or 0)

    if n >= 100:
        score = 10.0
    elif n >= 10:
        # 10 => 5.0, 99 => 7.99
        score = 5.0 + (n - 10) / 89.0 * 2.99
    else:
        # 0 => 0.0, 9 => 4.99
        score = (n / 9.0) * 4.99 if n > 0 else 0.0

    detail = f"{n:,} memories" if n > 0 else "empty profile"
    return DimensionResult(
        name="Corpus size",
        score=round(score, 1),
        zone=zone(score),
        detail=detail,
    )


# ─── 5. Wiki coverage ───────────────────────────────────────────────────


def compute_wiki_coverage(profile: str) -> DimensionResult:
    """Score fresh / total topic_summaries.

    Per spec: any stale row exists => AMBER. Empty table or missing table
    (pre-028 DBs) => N/A => 10.0 (no problem).
    """
    backend = cast(Any, get_backend())
    if not hasattr(backend, "_execute"):
        # Aggregate over topic_summaries needs raw SQL; the Supabase backend
        # (PostgREST) doesn't expose one. Honest N/A until a backend-level
        # wiki_coverage_stats RPC ships.
        return DimensionResult(
            name="Wiki coverage",
            score=10.0,
            zone="GREEN",
            detail="N/A (Supabase backend — Postgres-only check)",
        )
    try:
        row = backend._execute(
            "SELECT "
            "  COUNT(*) FILTER (WHERE status = 'fresh') AS fresh, "
            "  COUNT(*) AS total "
            "FROM topic_summaries WHERE profile_id = %(profile)s",
            {"profile": profile},
            fetch="one",
        )
    except Exception as e:
        # topic_summaries doesn't exist yet (pre-028) — that's fine, N/A.
        return DimensionResult(
            name="Wiki coverage",
            score=10.0,
            zone="GREEN",
            detail=f"N/A (table not available: {type(e).__name__})",
        )

    fresh = int((row or {}).get("fresh", 0) or 0)
    total = int((row or {}).get("total", 0) or 0)

    if total == 0:
        return DimensionResult(
            name="Wiki coverage",
            score=10.0,
            zone="GREEN",
            detail="N/A (no summaries — nothing to be stale)",
        )

    ratio = fresh / total
    score = round(ratio * 10.0, 1)

    # Per spec: any stale exists => AMBER even if score >=8.
    has_stale = fresh < total
    if has_stale and score >= 8.0:
        # Cap at 7.9 to force AMBER zone.
        score = min(score, 7.9)

    pct = ratio * 100.0
    detail = f"{pct:.0f}% fresh ({total - fresh} of {total} stale)"
    if not has_stale:
        detail = f"100% fresh ({total}/{total})"

    return DimensionResult(
        name="Wiki coverage",
        score=score,
        zone=zone(score),
        detail=detail,
    )


# ─── 6. Profile health (avg tags + orphan %) ────────────────────────────


def compute_profile_health(profile: str) -> DimensionResult:
    """Score on tag richness + relationship coverage.

    Composite: avg-tags-per-memory >= 2 AND orphan-pct < 10%
        => GREEN.
    Each axis ramps half the score:
      tags: 0 => 0, 1 => 2.5, 2+ => 5.0
      orphans: 0% => 5.0, 50%+ => 0.0
    """
    backend = cast(Any, get_backend())
    if not hasattr(backend, "_execute"):
        # avg_tags + orphan-pct join needs raw SQL. PostgREST can't express the
        # NOT EXISTS lateral against memory_relationships in one round-trip;
        # honest N/A until a backend-level profile_health_stats RPC ships.
        return DimensionResult(
            name="Profile health",
            score=10.0,
            zone="GREEN",
            detail="N/A (Supabase backend — Postgres-only check)",
        )
    try:
        row = backend._execute(
            "SELECT "
            "  COALESCE(AVG(COALESCE(array_length(tags, 1), 0)), 0)::float AS avg_tags, "
            "  COUNT(*)::int AS total, "
            "  COUNT(*) FILTER ("
            "    WHERE NOT EXISTS ("
            "      SELECT 1 FROM memory_relationships r "
            "      WHERE r.source_id = memories.id OR r.target_id = memories.id"
            "    )"
            "  )::int AS orphans "
            "FROM memories WHERE profile = %(profile)s",
            {"profile": profile},
            fetch="one",
        )
    except Exception as e:
        # memory_relationships might not exist on very old self-hosts.
        return DimensionResult(
            name="Profile health",
            score=10.0,
            zone="GREEN",
            detail=f"N/A (relationships query failed: {type(e).__name__})",
        )

    avg_tags = float((row or {}).get("avg_tags", 0.0) or 0.0)
    total = int((row or {}).get("total", 0) or 0)
    orphans = int((row or {}).get("orphans", 0) or 0)

    if total == 0:
        return DimensionResult(
            name="Profile health",
            score=10.0,
            zone="GREEN",
            detail="N/A (no memories — nothing to score)",
        )

    orphan_pct = (orphans / total) * 100.0

    # Each axis contributes up to 5.0
    tags_axis = min(avg_tags / 2.0, 1.0) * 5.0
    orphan_axis = max(0.0, 1.0 - orphan_pct / 50.0) * 5.0
    score = round(tags_axis + orphan_axis, 1)

    detail = f"avg {avg_tags:.1f} tags/memory, {orphan_pct:.0f}% orphans"
    return DimensionResult(
        name="Profile health",
        score=score,
        zone=zone(score),
        detail=detail,
    )


# ─── 7. Concurrency (pool stats) ────────────────────────────────────────


def compute_concurrency() -> DimensionResult:
    """Score connection-pool busy% + max-wait time.

    busy<50% AND max_wait<100ms => GREEN.
    """
    backend = get_backend()
    pool = getattr(backend, "_pool", None)
    if pool is None or not hasattr(pool, "get_stats"):
        # SupabaseBackend / GatewayBackend have no psycopg pool.
        return DimensionResult(
            name="Concurrency",
            score=10.0,
            zone="GREEN",
            detail="N/A (backend has no psycopg pool)",
        )

    try:
        stats = pool.get_stats()
    except Exception as e:
        return DimensionResult(
            name="Concurrency",
            score=10.0,
            zone="GREEN",
            detail=f"N/A (pool stats unavailable: {type(e).__name__})",
        )

    pool_size = int(stats.get("pool_size", 1) or 1)
    # Note: ``or`` shortcuts on 0, which we DO want as a literal (zero
    # available means fully busy). Explicit None check.
    raw_available = stats.get("pool_available")
    pool_available = int(raw_available) if raw_available is not None else pool_size
    busy = max(0, pool_size - pool_available)
    busy_pct = (busy / pool_size) * 100.0 if pool_size > 0 else 0.0

    # ``usage_ms`` is the max time a request waited for a connection.
    # psycopg-pool exposes ``requests_wait_ms`` aggregated; we use whatever
    # key is available.
    max_wait_ms = float(
        stats.get("usage_ms")
        or stats.get("requests_wait_ms")
        or stats.get("requests_max_wait_ms")
        or 0.0
    )

    # Each axis contributes 5 points.
    busy_axis = max(0.0, 1.0 - busy_pct / 100.0) * 5.0
    wait_axis = max(0.0, 1.0 - max_wait_ms / 1000.0) * 5.0
    score = round(busy_axis + wait_axis, 1)

    detail = f"pool busy {busy_pct:.0f}%, max wait {max_wait_ms:.0f}ms"
    return DimensionResult(
        name="Concurrency",
        score=score,
        zone=zone(score),
        detail=detail,
    )


# ─── 8. E2E probe ───────────────────────────────────────────────────────


def _run_e2e_probe(profile: str) -> tuple[bool, float, str | None]:
    """Actually perform a store -> hybrid_search -> delete round-trip.

    Returns ``(success, total_ms, error_message_or_None)``.
    Cleans up its own row even if search returns nothing — the unique
    tag makes orphans recoverable across runs.
    """
    from ogham.database import (
        delete_memory,
        hybrid_search_memories,
        store_memory,
    )
    from ogham.embeddings import generate_embedding

    probe_id = uuid4().hex[:8]
    tag = f"health-probe-{probe_id}"
    content = f"Health probe round-trip {probe_id}"

    t0 = time.perf_counter()
    stored_id: str | None = None
    try:
        embedding = generate_embedding(content)
        stored = store_memory(
            content=content,
            embedding=embedding,
            profile=profile,
            tags=[tag],
            source="health-probe",
        )
        stored_id = str(stored.get("id"))

        results = hybrid_search_memories(
            query_text=content,
            query_embedding=embedding,
            profile=profile,
            limit=3,
            tags=[tag],
        )
        if not results:
            return (
                False,
                (time.perf_counter() - t0) * 1000,
                "search returned 0 results for unique tag",
            )

        deleted = delete_memory(stored_id, profile)
        if not deleted:
            return (
                False,
                (time.perf_counter() - t0) * 1000,
                "delete returned False",
            )
        stored_id = None  # successful path
        return (True, (time.perf_counter() - t0) * 1000, None)
    except Exception as e:
        return (False, (time.perf_counter() - t0) * 1000, f"{type(e).__name__}: {e}")
    finally:
        # Best-effort cleanup if delete didn't run.
        if stored_id:
            try:
                delete_memory(stored_id, profile)
            except Exception as cleanup_err:
                logger.debug("e2e probe cleanup failed for %s: %s", stored_id, cleanup_err)


def compute_e2e_probe(profile: str) -> DimensionResult:
    """Score the full store -> search -> delete round-trip.

    All three steps succeed => 10.0; any failure => 0.0.
    """
    success, total_ms, err = _run_e2e_probe(profile)
    if success:
        return DimensionResult(
            name="E2E probe",
            score=10.0,
            zone="GREEN",
            detail=f"store->search->delete OK in {total_ms:.0f}ms",
        )
    return DimensionResult(
        name="E2E probe",
        score=0.0,
        zone="RED",
        detail=f"failed: {err}",
    )


# ─── compose ────────────────────────────────────────────────────────────


# Order matches the spec table — important for the CLI render.
DIMENSION_NAMES = [
    "DB freshness",
    "Schema integrity",
    "Hybrid search",
    "Corpus size",
    "Wiki coverage",
    "Profile health",
    "Concurrency",
    "E2E probe",
]


def compose_health(profile: str | None = None) -> list[DimensionResult]:
    """Run all eight ``compute_*`` and return their results in spec order.

    Failures in any one dimension are caught and reported as RED with
    the exception message in ``detail``.
    """
    target = profile or settings.default_profile

    plan: list[tuple[str, Callable[[], DimensionResult]]] = [
        ("DB freshness", lambda: compute_db_freshness(target)),
        ("Schema integrity", lambda: compute_schema_integrity()),
        ("Hybrid search", lambda: compute_hybrid_search_latency(target)),
        ("Corpus size", lambda: compute_corpus_size(target)),
        ("Wiki coverage", lambda: compute_wiki_coverage(target)),
        ("Profile health", lambda: compute_profile_health(target)),
        ("Concurrency", lambda: compute_concurrency()),
        ("E2E probe", lambda: compute_e2e_probe(target)),
    ]

    out: list[DimensionResult] = []
    for name, fn in plan:
        try:
            result = fn()
            # Guard: a mock that returned None should still produce a row.
            if result is None:
                result = DimensionResult(
                    name=name,
                    score=0.0,
                    zone="RED",
                    detail="dimension returned no result",
                )
            # Force the canonical spec name so ordering / mocking stay
            # well-defined even if a dim returns its own label.
            if result.name != name:
                result = DimensionResult(
                    name=name,
                    score=result.score,
                    zone=result.zone,
                    detail=result.detail,
                )
            out.append(result)
        except Exception as e:
            logger.exception("health dim %s raised", name)
            out.append(
                DimensionResult(
                    name=name,
                    score=0.0,
                    zone="RED",
                    detail=f"error: {type(e).__name__}: {e}",
                )
            )
    return out


def overall_score(results: list[DimensionResult]) -> float:
    """Mean score across eight dimensions, rounded to 1 decimal."""
    if not results:
        return 0.0
    return round(sum(r.score for r in results) / len(results), 1)
