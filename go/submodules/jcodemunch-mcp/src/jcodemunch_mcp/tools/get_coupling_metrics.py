"""Compute afferent/efferent coupling and instability for a module."""

import time
from typing import Optional

from ..storage import IndexStore
from ._utils import index_status_to_tool_error, resolve_repo
from .get_dependency_graph import _build_adjacency


def get_coupling_metrics(
    repo: str,
    module_path: str,
    storage_path: Optional[str] = None,
) -> dict:
    """Return coupling metrics for a file/module.

    Metrics:
        Ca (afferent coupling)  — number of files that import this module.
                                  High Ca → many dependents → stable/hard to change.
        Ce (efferent coupling)  — number of files this module imports.
                                  High Ce → many dependencies → unstable/fragile.
        instability (I)         — Ce / (Ca + Ce).  0 = maximally stable, 1 = maximally unstable.

    Args:
        repo: Repository identifier (owner/repo or repo name).
        module_path: File path within the repo (e.g. 'src/utils.py').
        storage_path: Custom storage path.

    Returns:
        {
          "repo": str,
          "module": str,
          "ca": int,
          "ce": int,
          "instability": float,       # 0.0–1.0; null when ca+ce == 0
          "assessment": str,          # "stable" | "neutral" | "unstable"
          "importers": [str, ...],    # files that import this module
          "dependencies": [str, ...], # files this module imports
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
                "Re-index with jcodemunch-mcp >= 1.3.0 to enable coupling analysis."
            )
        }

    if module_path not in index.source_files:
        return {"error": f"File not found in index: {module_path}"}

    source_files = frozenset(index.source_files)
    alias_map = getattr(index, "alias_map", None)
    fwd = _build_adjacency(index.imports, source_files, alias_map, getattr(index, "psr4_map", None))

    # Build reverse adjacency (importers)
    rev: dict[str, list[str]] = {}
    for src, targets in fwd.items():
        for tgt in targets:
            rev.setdefault(tgt, []).append(src)

    importers: list[str] = sorted(rev.get(module_path, []))
    dependencies: list[str] = sorted(fwd.get(module_path, []))

    ca = len(importers)
    ce = len(dependencies)
    total = ca + ce

    if total == 0:
        instability = None
        assessment = "isolated"
    else:
        instability = round(ce / total, 4)
        if instability <= 0.3:
            assessment = "stable"
        elif instability <= 0.7:
            assessment = "neutral"
        else:
            assessment = "unstable"

    elapsed = (time.perf_counter() - start) * 1000
    return {
        "repo": f"{owner}/{name}",
        "module": module_path,
        "ca": ca,
        "ce": ce,
        "instability": instability,
        "assessment": assessment,
        "importers": importers,
        "dependencies": dependencies,
        "_meta": {"timing_ms": round(elapsed, 1)},
    }
