"""Validate inter-module imports against declared architectural layers."""

import time
from typing import Optional

from ..storage import IndexStore
from ._utils import index_status_to_tool_error, resolve_repo
from .get_dependency_graph import _build_adjacency


def _file_to_layer(file_path: str, layers: list[dict]) -> Optional[str]:
    """Return the layer name for *file_path*, or None if unassigned.

    Matching uses path prefix (``paths`` entries are treated as directory
    prefixes; a file matches if its path starts with the prefix).  First
    matching layer wins.
    """
    for layer in layers:
        for prefix in layer.get("paths", []):
            # Normalise to forward slashes for consistent comparison
            norm_file = file_path.replace("\\", "/")
            norm_prefix = prefix.rstrip("/")
            if norm_file.startswith(norm_prefix + "/") or norm_file == norm_prefix:
                return layer["name"]
    return None


def _resolve_layers(
    rules: Optional[list[dict]],
    repo: str,
    index_source_root: Optional[str],
) -> list[dict]:
    """Return the layer definitions to use.

    Priority:
    1. Explicit ``rules`` argument (caller-supplied).
    2. ``architecture.layers`` from the project's .jcodemunch.jsonc.
    3. Empty list (no rules).
    """
    if rules is not None:
        return rules

    # Try project config
    try:
        from .. import config as _cfg
        arch = _cfg.get("architecture", {}, repo=repo)
        if isinstance(arch, dict):
            layers = arch.get("layers", [])
            if layers:
                return layers
    except Exception:
        import logging as _logging
        _logging.getLogger(__name__).debug("Failed to read architecture config", exc_info=True)

    return []


def get_layer_violations(
    repo: str,
    rules: Optional[list[dict]] = None,
    storage_path: Optional[str] = None,
) -> dict:
    """Check whether imports respect declared architectural layer boundaries.

    Layer rules define which layers may not import from which other layers.
    Files are assigned to layers by path prefix; unassigned files are skipped.

    Args:
        repo: Repository identifier (owner/repo or repo name).
        rules: Layer definitions (overrides project config).  Each entry:
            {
              "name": str,              # layer identifier
              "paths": [str, ...],      # path prefixes belonging to this layer
              "may_not_import": [str, ...]  # layers this layer must not import
            }
            If omitted, reads ``architecture.layers`` from .jcodemunch.jsonc.
        storage_path: Custom storage path.

    Returns:
        {
          "repo": str,
          "layer_count": int,
          "violation_count": int,
          "violations": [
            {
              "file": str,
              "file_layer": str,
              "import_target": str,
              "target_layer": str,
              "rule_violated": str   # e.g. "api may_not_import db"
            }, ...
          ],
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
                "Re-index with jcodemunch-mcp >= 1.3.0 to enable layer analysis."
            )
        }

    source_root = getattr(index, "source_root", None)
    layers = _resolve_layers(rules, f"{owner}/{name}", source_root)

    if not layers:
        return {
            "repo": f"{owner}/{name}",
            "layer_count": 0,
            "violation_count": 0,
            "violations": [],
            "note": (
                "No layer rules defined. Pass 'rules' or add 'architecture.layers' "
                "to .jcodemunch.jsonc."
            ),
            "_meta": {"timing_ms": round((time.perf_counter() - start) * 1000, 1)},
        }

    # Build a fast lookup: layer_name -> set of forbidden target layers
    forbidden: dict[str, set[str]] = {}
    for layer in layers:
        lname = layer.get("name", "")
        mni = layer.get("may_not_import", [])
        if lname and mni:
            forbidden[lname] = set(mni)

    source_files = frozenset(index.source_files)
    alias_map = getattr(index, "alias_map", None)
    fwd = _build_adjacency(index.imports, source_files, alias_map, getattr(index, "psr4_map", None))

    violations: list[dict] = []

    for src_file, targets in fwd.items():
        src_layer = _file_to_layer(src_file, layers)
        if not src_layer or src_layer not in forbidden:
            continue
        disallowed = forbidden[src_layer]
        for tgt_file in targets:
            tgt_layer = _file_to_layer(tgt_file, layers)
            if tgt_layer and tgt_layer in disallowed:
                violations.append({
                    "file": src_file,
                    "file_layer": src_layer,
                    "import_target": tgt_file,
                    "target_layer": tgt_layer,
                    "rule_violated": f"{src_layer} may_not_import {tgt_layer}",
                })

    elapsed = (time.perf_counter() - start) * 1000
    return {
        "repo": f"{owner}/{name}",
        "layer_count": len(layers),
        "violation_count": len(violations),
        "violations": violations,
        "_meta": {"timing_ms": round(elapsed, 1)},
    }
