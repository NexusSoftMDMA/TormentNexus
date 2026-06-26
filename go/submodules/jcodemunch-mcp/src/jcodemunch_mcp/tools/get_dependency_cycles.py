"""Detect circular import chains in a repository's dependency graph."""

import time
from typing import Optional

from ..storage import IndexStore
from ._utils import index_status_to_tool_error, resolve_repo
from .get_dependency_graph import _build_adjacency


def _find_cycles(adj: dict[str, list[str]]) -> list[list[str]]:
    """Find SCCs of size > 1 using Kosaraju's algorithm (iterative).

    Each returned SCC represents a set of files involved in a circular
    import chain.  The members are sorted for deterministic output.
    """
    # Collect all nodes
    all_nodes: set[str] = set(adj.keys())
    for targets in adj.values():
        all_nodes.update(targets)

    # ------------------------------------------------------------------ #
    # Pass 1: DFS on original graph — record finish order
    # ------------------------------------------------------------------ #
    visited: set[str] = set()
    finish_order: list[str] = []

    for start in all_nodes:
        if start in visited:
            continue
        visited.add(start)
        # Stack entries: (node, neighbor_iterator)
        stack: list[tuple[str, object]] = [(start, iter(adj.get(start, [])))]
        while stack:
            node, it = stack[-1]
            try:
                w = next(it)  # type: ignore[call-overload]
                if w not in visited:
                    visited.add(w)
                    stack.append((w, iter(adj.get(w, []))))
            except StopIteration:
                stack.pop()
                finish_order.append(node)

    # ------------------------------------------------------------------ #
    # Build transpose graph
    # ------------------------------------------------------------------ #
    rev_adj: dict[str, list[str]] = {}
    for src, targets in adj.items():
        for tgt in targets:
            rev_adj.setdefault(tgt, []).append(src)

    # ------------------------------------------------------------------ #
    # Pass 2: DFS on transpose in reverse finish order
    # ------------------------------------------------------------------ #
    visited2: set[str] = set()
    sccs: list[list[str]] = []

    for start in reversed(finish_order):
        if start in visited2:
            continue
        scc: list[str] = []
        work = [start]
        visited2.add(start)
        while work:
            node = work.pop()
            scc.append(node)
            for w in rev_adj.get(node, []):
                if w not in visited2:
                    visited2.add(w)
                    work.append(w)
        if len(scc) > 1:
            sccs.append(sorted(scc))

    return sccs


def get_dependency_cycles(
    repo: str,
    storage_path: Optional[str] = None,
) -> dict:
    """Detect circular import chains in the repository.

    Uses Kosaraju's algorithm on the file-level import graph.  Any set of
    files that mutually import each other (directly or transitively) forms a
    strongly-connected component and is reported as a cycle.

    Args:
        repo: Repository identifier (owner/repo or repo name).
        storage_path: Custom storage path.

    Returns:
        {
          "repo": str,
          "cycle_count": int,
          "cycles": [[file, ...], ...],   # each list = one SCC
          "_meta": {"timing_ms": float}
        }
    """
    start = time.perf_counter()

    try:
        owner, name = resolve_repo(repo, storage_path)
    except ValueError as e:
        return {"error": str(e)}

    store = IndexStore(base_path=storage_path)
    index = store.load_index(owner, name)
    if not index:
        return index_status_to_tool_error(store.inspect_index(owner, name))

    if not index.imports:
        return {
            "error": (
                "No import data available. "
                "Re-index with jcodemunch-mcp >= 1.3.0 to enable dependency analysis."
            )
        }

    source_files = frozenset(index.source_files)
    adj = _build_adjacency(index.imports, source_files, getattr(index, "alias_map", None), getattr(index, "psr4_map", None))
    cycles = _find_cycles(adj)

    elapsed = (time.perf_counter() - start) * 1000
    return {
        "repo": f"{owner}/{name}",
        "cycle_count": len(cycles),
        "cycles": cycles,
        "_meta": {"timing_ms": round(elapsed, 1)},
    }
