"""Hebbian edge strengthening for co-retrieved memories.

Source: MuninnDB's activation_test.go / docs/how-memory-works.md.
eta=0.01 chosen to match their cautious default -- edges grow slowly.

Stored as relationship='related' in memory_relationships because the
enum doesn't (yet) have a 'co-retrieved' value. Semantically close;
upgrade to a dedicated value if we ever need to distinguish.

Pairs are canonicalized to ``(min(a, b), max(a, b))`` before UPSERT for two
reasons:

1. **Deadlock prevention** -- concurrent searches sharing memory IDs can
   deadlock if row locks are acquired in inconsistent orders. Canonical
   ordering guarantees a single global lock order.
2. **Fragmentation prevention** -- without canonicalization,
   ``strengthen_edges([a, b])`` and ``strengthen_edges([b, a])`` would each
   create a separate row against the ``(source_id, target_id, relationship)``
   unique constraint. Canonical ordering makes the write idempotent.

The INSERT is a single round-trip regardless of pair count (UNNEST of two
parallel arrays), replacing the previous C(n, 2) per-pair ON CONFLICT loop.
"""

from __future__ import annotations

from itertools import combinations

from ogham.database import get_backend

HEBBIAN_RATE = 0.01
BOOTSTRAP_STRENGTH = 0.1


def strengthen_edges(memory_ids: list[str]) -> int:
    """Update or insert memory_relationships for pairs in memory_ids.

    Pairs are canonicalized via ``sorted()`` (lexicographic UUID string order)
    before insertion for deadlock + fragmentation safety. The write is a
    single UNNEST-driven INSERT regardless of how many pairs are involved.

    Returns count of edges touched.

    v0.13.1: dispatches through the backend facade
    `hebbian_strengthen_edges`. Postgres keeps the inline UNNEST; Supabase
    calls the matching RPC defined in migration 035.
    """
    if len(memory_ids) < 2:
        return 0

    pairs = [tuple(sorted(p)) for p in combinations(memory_ids, 2)]
    pairs.sort()
    sources = [p[0] for p in pairs]
    targets = [p[1] for p in pairs]

    return get_backend().hebbian_strengthen_edges(
        sources=sources,
        targets=targets,
        bootstrap=BOOTSTRAP_STRENGTH,
        rate=HEBBIAN_RATE,
    )
