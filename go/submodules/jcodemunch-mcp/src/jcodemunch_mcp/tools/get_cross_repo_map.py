"""Get the cross-repository dependency map at the package level."""

import time
from typing import Optional

from ..storage import IndexStore
from .package_registry import (
    build_package_registry,
    extract_root_package_from_specifier,
)


def get_cross_repo_map(
    repo: Optional[str] = None,
    storage_path: Optional[str] = None,
) -> dict:
    """Return which indexed repos depend on which other indexed repos at the package level.

    Args:
        repo: Optional repo ID to filter. If given, only show deps for this repo.
              If omitted, show the full cross-repo dependency map.
        storage_path: Custom storage path.

    Returns:
        Dict with repos list and cross_repo_edges list.
    """
    start = time.perf_counter()
    store = IndexStore(base_path=storage_path)
    all_repos_raw = store.list_repos()

    if not all_repos_raw:
        return {
            "repos": [],
            "cross_repo_edges": [],
            "_meta": {"timing_ms": round((time.perf_counter() - start) * 1000, 1)},
        }

    # Filter to requested repo if provided
    if repo:
        # Normalize: try to match by repo ID
        target_repos = [r for r in all_repos_raw if r.get("repo") == repo]
        if not target_repos:
            # Try partial match (bare name)
            target_repos = [r for r in all_repos_raw if r.get("repo", "").endswith("/" + repo)]
        if not target_repos:
            return {"error": f"Repository not found: {repo}"}
    else:
        target_repos = all_repos_raw

    # Build package registry: {package_name -> [repo_id, ...]}
    registry = build_package_registry(all_repos_raw)

    # Invert registry: {repo_id -> [package_name, ...]}
    repo_packages: dict[str, list[str]] = {}
    for pkg_name, repo_ids in registry.items():
        for rid in repo_ids:
            repo_packages.setdefault(rid, []).append(pkg_name)

    # Build cross-repo edges by scanning each repo's imports
    cross_repo_edges: list[dict] = []
    seen_edges: set[tuple] = set()

    # For each target repo, analyze its imports
    repos_to_analyze = target_repos if repo else all_repos_raw

    for repo_entry in repos_to_analyze:
        repo_id = repo_entry.get("repo", "")
        if not repo_id or "/" not in repo_id:
            continue
        owner, name = repo_id.split("/", 1)
        index = store.load_index(owner, name)
        if not index or not index.imports:
            continue

        for src_file, file_imports in index.imports.items():
            for imp in file_imports:
                specifier = imp.get("specifier", "")
                lang = index.file_languages.get(src_file, "")
                root_pkg = extract_root_package_from_specifier(specifier, lang)
                if not root_pkg:
                    continue
                # Look up which repo provides this package
                providing_repos = registry.get(root_pkg, [])
                for providing_repo_id in providing_repos:
                    if providing_repo_id == repo_id:
                        continue  # skip self-dependency
                    edge_key = (repo_id, providing_repo_id, root_pkg)
                    if edge_key not in seen_edges:
                        seen_edges.add(edge_key)
                        cross_repo_edges.append({
                            "from_repo": repo_id,
                            "to_repo": providing_repo_id,
                            "package_name": root_pkg,
                        })

    # Build per-repo result objects
    # For each repo in target_repos, compute depends_on and depended_on_by
    result_repos: list[dict] = []

    # Build fast lookup maps from edges
    edges_from: dict[str, list[dict]] = {}  # repo_id -> [edges where from_repo == repo_id]
    edges_to: dict[str, list[dict]] = {}    # repo_id -> [edges where to_repo == repo_id]
    for edge in cross_repo_edges:
        edges_from.setdefault(edge["from_repo"], []).append(edge)
        edges_to.setdefault(edge["to_repo"], []).append(edge)

    for repo_entry in target_repos:
        repo_id = repo_entry.get("repo", "")
        pkg_names = repo_packages.get(repo_id, [])

        depends_on = [
            {"repo": e["to_repo"], "package_name": e["package_name"]}
            for e in edges_from.get(repo_id, [])
        ]
        depended_on_by = [
            {"repo": e["from_repo"], "package_name": e["package_name"]}
            for e in edges_to.get(repo_id, [])
        ]

        result_repos.append({
            "repo": repo_id,
            "package_names": pkg_names,
            "depends_on": depends_on,
            "depended_on_by": depended_on_by,
        })

    elapsed = (time.perf_counter() - start) * 1000
    return {
        "repos": result_repos,
        "cross_repo_edges": cross_repo_edges,
        "_meta": {
            "timing_ms": round(elapsed, 1),
            "total_repos_scanned": len(all_repos_raw),
        },
    }
